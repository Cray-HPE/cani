package external_inventory_provider

// TODO Need to think about how internal data structures should be supplied to the Inventory Provider
type InventoryProvider interface {
	// Validate the external services of the inventory provider are correct
	ValidateExternal() error

	// Validate the respresntation of the inventory data into the destination inventory system
	// is consistent.
	// TODO perhaps this should just happen during Reconcile
	ValidateInternal() error

	// Import external inventory data into CANI's inventory format
	Import() error

	// Reconcile CANI's inventory state with the external inventory state and apply required changes
	Reconcile() error
}
