package domain

import (
	"context"
	"errors"
	"fmt"
)

func (d *Domain) Commit(ctx context.Context) error {
	inventoryProvider := d.externalInventoryProvider

	// Perform validation integrity of CANI's inventory data
	if err := d.datastore.Validate(); err != nil {
		return errors.Join(
			fmt.Errorf("failed to validate inventory datastore"),
			err,
		)
	}

	// Validate the current state of CANI's inventory data against the provider plugin
	// for provider specific data
	if _, err := inventoryProvider.ValidateInternal(ctx, d.datastore); err != nil {
		return errors.Join(
			fmt.Errorf("failed to validate inventory against inventory provider plugin"),
			err,
		)
	}

	// Validate the current state of the external inventory
	if err := inventoryProvider.ValidateExternal(ctx); err != nil {
		return errors.Join(
			fmt.Errorf("failed to validate external inventory provider"),
			err,
		)
	}

	// Reconcile our inventory with the external inventory system
	return inventoryProvider.Reconcile(ctx, d.datastore)

}
