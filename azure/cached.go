package azure

import (
	"log"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// Cache represents resources cached between Collectors
type Cache struct {
	mu sync.Mutex

	// Azure Resource Groups
	ResourceGroups []*armresources.ResourceGroup
}

// NewCache creates a new Account
func NewCache() *Cache {
	resourcegroups := []*armresources.ResourceGroup{}
	return &Cache{
		ResourceGroups: resourcegroups,
	}
}

// Update is a method that transactionally updates the cache
func (x *Cache) Update(resourcegroups []*armresources.ResourceGroup) {
	log.Print("[Update] replacing Resource Groups")
	x.mu.Lock()
	x.ResourceGroups = resourcegroups
	x.mu.Unlock()
	log.Print("[Update] done")
}
