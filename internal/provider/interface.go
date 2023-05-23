package provider

import "github.com/Cray-HPE/cani/internal/inventory"

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
	Reconcile(datastore inventory.Datastore) error

	// Build metadata, and add ito the hardware object
	// This function could return the data to put into object
	BuildHardwareMetadata(hw *inventory.Hardware, rawProperties map[string]interface{}) error
}
