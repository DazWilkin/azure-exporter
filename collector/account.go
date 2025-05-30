package collector

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/billing/armbilling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/consumption/armconsumption"
	"github.com/DazWilkin/azure-exporter/azure"
	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = (*AccountCollector)(nil)

// AccountCollector represents Azure Account
type AccountCollector struct {
	account *azure.Account

	accountClient  *armbilling.AccountsClient
	balancesClient *armconsumption.BalancesClient

	// Metrics
	currentbalance *prometheus.Desc
}

// NewAccountCollector return a new AccountCollector
func NewAccountCollector(account *azure.Account, subscription string, creds *azidentity.DefaultAzureCredential) *AccountCollector {
	subsystem := "account"

	accountClient, err := armbilling.NewAccountsClient(creds, nil)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	balancesClient, err := armconsumption.NewBalancesClient(creds, nil)
	if err != nil {
		log.Fatal(err)
	}

	return &AccountCollector{
		account: account,

		accountClient:  accountClient,
		balancesClient: balancesClient,

		currentbalance: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "current_balance"),
			"Azure current account balance",
			[]string{
				"currency",
			},
			nil,
		),
	}
}

// Collect implements Prometheus' Collector interface and is used to collect metrics
func (c *AccountCollector) Collect(ch chan<- prometheus.Metric) {
	log.Println("[AccountCollector:Collect] started")

	ctx := context.Background()

	pager := c.accountClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println("[AccountCollector:Collect] error getting account list")
			continue
		}

		for _, account := range page.Value {
			// ID is the fully-qualified Billing ID
			// Name is the Billing ID
			accountID := *account.Name

			resp, err := c.balancesClient.GetByBillingAccount(ctx, accountID, nil)
			if err != nil {
				if err := err.(*azcore.ResponseError); err != nil {
					// Error is an Azure ResponseError
					// 404 could mean that there are no account balance records for this quarter (not such an error)
					log.Printf("[AccountCollector:Collect] error getting account balance (%d)", err.StatusCode)
					return
				}

				log.Println("[AccountCollector:Collect] error getting account balance")
				return
			}

			if resp.Balance.Properties == nil {
				log.Println("[AccountCollector:Collect] no balance properties returned")
				return
			}

			p := resp.Balance.Properties

			ch <- prometheus.MustNewConstMetric(
				c.currentbalance,
				prometheus.GaugeValue,
				*p.EndingBalance,
				[]string{
					func(c *string) string {
						if c == nil {
							return "USD"
						}
						return *c
					}(p.Currency),
				}...,
			)
		}
	}
}

// Describe implements Prometheus' Collector interface and is used to describe metrics
func (c *AccountCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.currentbalance
}
