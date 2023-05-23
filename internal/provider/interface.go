package provider

import (
	"github.com/Cray-HPE/cani/internal/inventory"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

// TODO Need to think about how internal data structures should be supplied to the Inventory Provider
type InventoryProvider interface {
	// Validate the external services of the inventory provider are correct
	ValidateExternal() error

	// Validate the representation of the inventory data into the destination inventory system
	// is consistent.
	// TODO perhaps this should just happen during Reconcile
	ValidateInternal() error

	// Import external inventory data into CANI's inventory format
	Import() error

	// Reconcile CANI's inventory state with the external inventory state and apply required changes
	Reconcile(inventory.Inventory, []sls_client.Hardware) ([]sls_client.Hardware, error)
}
