package azure

import (
	"log"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// Account represents an Azure account
type Account struct {
	mu sync.Mutex

	// Azure Resource Groups
	ResourceGroups []*armresources.ResourceGroup
}

// NewAccount creates a new Account
func NewAccount() *Account {
	resourcegroups := []*armresources.ResourceGroup{}
	return &Account{
		ResourceGroups: resourcegroups,
	}
}

// Update is a method that transactionally updates the List of Azure Resource Groups
func (x *Account) Update(resourcegroups []*armresources.ResourceGroup) {
	log.Print("[Update] replacing Resource Groups")
	x.mu.Lock()
	x.ResourceGroups = resourcegroups
	x.mu.Unlock()
	log.Print("[Update] done")
}
