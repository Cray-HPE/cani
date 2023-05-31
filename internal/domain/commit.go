package domain

import (
	"context"
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/google/uuid"
)

type CommitResult struct {
	ProviderValidationErrors map[uuid.UUID]provider.HardwareValidationResult
}

func (d *Domain) Commit(ctx context.Context) (CommitResult, error) {
	inventoryProvider := d.externalInventoryProvider

	// Perform validation integrity of CANI's inventory data
	if err := d.datastore.Validate(); err != nil {
		return CommitResult{}, errors.Join(
			fmt.Errorf("failed to validate inventory datastore"),
			err,
		)
	}

	// Validate the current state of CANI's inventory data against the provider plugin
	// for provider specific data
	if failedValidations, err := inventoryProvider.ValidateInternal(ctx, d.datastore, true); len(failedValidations) > 0 {
		return CommitResult{
			ProviderValidationErrors: failedValidations,
		}, err
	} else if err != nil {
		return CommitResult{}, errors.Join(
			fmt.Errorf("failed to validate inventory against inventory provider plugin"),
			err,
		)
	}

	// Validate the current state of the external inventory
	if err := inventoryProvider.ValidateExternal(ctx); err != nil {
		return CommitResult{}, errors.Join(
			fmt.Errorf("failed to validate external inventory provider"),
			err,
		)
	}

	// Reconcile our inventory with the external inventory system
	return CommitResult{}, inventoryProvider.Reconcile(ctx, d.datastore)

}
