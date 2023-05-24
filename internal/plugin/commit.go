package plugin

import (
	"errors"
	"fmt"
)

// Commit reconciles and validates the current state of the inventory to the external inventory system
func (p *Plugin) Commit() error {
	inventoryProvider := p.provider

	// Perform validation of CANI's inventory data
	if err := p.datastore.Validate(); err != nil {
		return errors.Join(
			fmt.Errorf("failed to validate inventory"),
			err,
		)
	}

	// Validate the current state of the external inventory
	if err := inventoryProvider.ValidateExternal(); err != nil {
		return errors.Join(
			fmt.Errorf("failed to validate external inventory provider"),
			err,
		)
	}

	// Reconcile our inventory with the external inventory system
	return inventoryProvider.Reconcile(p.datastore)

}
