package collector

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"github.com/DazWilkin/azure-exporter/azure"

	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = (*ResourceGroupsCollector)(nil)

// ResourceGroupsCollector represents Azure Resource Groups
type ResourceGroupsCollector struct {
	account *azure.Account
	client  *armresources.ResourceGroupsClient

	Count *prometheus.Desc
}

// NewResourceGroupsCollector is a function that creates a new ResourceGroupsCollector
func NewResourceGroupsCollector(account *azure.Account, subscription string, creds *azidentity.DefaultAzureCredential) *ResourceGroupsCollector {
	subsystem := "resource_groups"

	client, err := armresources.NewResourceGroupsClient(subscription, creds, nil)
	if err != nil {
		// TODO(dazwilkin) Should probably return but...
		log.Print(err)
	}

	return &ResourceGroupsCollector{
		account: account,
		client:  client,

		Count: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "total"),
			"Number of Resource Groups",
			nil,
			nil,
		),
	}
}

// Collect implements Prometheus' Collector interface and is used to collect metrics
func (c *ResourceGroupsCollector) Collect(ch chan<- prometheus.Metric) {
	log.Println("[ResourceGroupsCollector] started")

	ctx := context.Background()

	// Side-effect of updating azure.Account
	resourcegroups := []*armresources.ResourceGroup{}

	pager := c.client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Print(err)
		}

		// len(resourcegroups) will be used as the metric's value
		resourcegroups = append(resourcegroups, page.Value...)
	}

	// Update azure.Account with new list of Resource Groups
	c.account.Update(resourcegroups)

	ch <- prometheus.MustNewConstMetric(
		c.Count,
		prometheus.GaugeValue,
		float64(len(resourcegroups)),
		[]string{}...,
	)

	log.Println("[ResourceGroupsCollector] completes")
}

// Describe implements Prometheus' Collector interface and is used to describe metrics
func (c *ResourceGroupsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Count
}
