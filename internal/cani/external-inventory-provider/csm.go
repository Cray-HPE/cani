package external_inventory_provider

import "fmt"

type CSM struct {
}

func NewCSM() (*CSM, error) {
	return &CSM{}, nil
}

// Validate the external services of the inventory provider are correct
func (csm *CSM) ValidateExternal() error {
	return fmt.Errorf("todo")
}

// Validate the respresntation of the inventory data into the destination inventory system
// is consistent.
// TODO perhaps this should just happen during Reconcile
func (csm *CSM) ValidateInternal() error {
	return fmt.Errorf("todo")

}

// Import external inventory data into CANI's inventory format
func (csm *CSM) Import() error {
	return fmt.Errorf("todo")

}

// Reconcile CANI's inventory state with the external inventory state and apply required changes
func (csm *CSM) Reconcile() error {
	return fmt.Errorf("todo")
}
