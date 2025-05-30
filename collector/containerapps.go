package collector

import (
	"context"
	"log"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/DazWilkin/azure-exporter/azure"

	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = (*ContainerAppsCollector)(nil)

// ContainerAppsCollector represents Azure Container Apps
type ContainerAppsCollector struct {
	client *armappcontainers.ContainerAppsClient
	cache  *azure.Cache

	Apps *prometheus.Desc
}

// NewContainerAppsCollector returns a new ContainerAppsCollector
func NewContainerAppsCollector(subscription string, creds *azidentity.DefaultAzureCredential, cache *azure.Cache) *ContainerAppsCollector {
	subsystem := "container_apps"

	clientFactory, err := armappcontainers.NewClientFactory(subscription, creds, nil)
	if err != nil {
		log.Print(err)
	}

	client := clientFactory.NewContainerAppsClient()

	return &ContainerAppsCollector{
		cache:  cache,
		client: client,

		Apps: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "total"),
			"Number of Container Apps deployed",
			[]string{
				"resourcegroup",
			},
			nil,
		),
	}
}

// Collect implements Prometheus' Collector interface and is used to collect metrics
func (c *ContainerAppsCollector) Collect(ch chan<- prometheus.Metric) {
	log.Println("[ContainerAppsCollector:Collect] started")

	ctx := context.Background()

	var wg sync.WaitGroup
	for _, resourcegroup := range c.cache.ResourceGroups {
		wg.Add(1)
		go func(rg *armresources.ResourceGroup) {
			defer wg.Done()
			count := 0
			pager := c.client.NewListByResourceGroupPager(*rg.Name, nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					log.Print(err)
				}

				// To aid clarity, the page.Value is []*armappcontainers.ContainerApp
				containerapps := page.Value
				count += len(containerapps)
			}

			ch <- prometheus.MustNewConstMetric(
				c.Apps,
				prometheus.GaugeValue,
				float64(count),
				[]string{
					*rg.Name,
				}...,
			)
		}(resourcegroup)
	}
	wg.Wait()

	log.Println("[ContainerAppsCollector:Collect] completes")
}

// Describe implements Prometheus' Collector interface and is used to describe metrics
func (c *ContainerAppsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Apps
}
