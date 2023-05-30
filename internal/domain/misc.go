package domain

import (
	"context"
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// List returns the inventory
func (d *Domain) List() (inventory.Inventory, error) {
	inv, err := d.datastore.List()
	if err != nil {
		return inventory.Inventory{}, err
	}

	return inv, nil
}

type ValidatePassback struct {
	ProviderValidationErrors map[uuid.UUID]provider.HardwareValidationResult
}

func (d *Domain) Validate(ctx context.Context) (ValidatePassback, error) {
	var passback ValidatePassback

	// Validate the current state of CANI's inventory data against the provider plugin
	// for provider specific data.
	if failedValidations, err := d.externalInventoryProvider.ValidateInternal(ctx, d.datastore, true); len(failedValidations) > 0 {
		passback.ProviderValidationErrors = failedValidations
		return passback, provider.ErrDataValidationFailure
	} else if err != nil {
		return ValidatePassback{}, errors.Join(
			fmt.Errorf("failed to validate inventory against inventory provider plugin"),
			err,
		)
	}
	log.Info().Msg("Validated CANI inventory")

	// Validate external inventory data
	err := d.externalInventoryProvider.ValidateExternal(ctx)
	if err != nil {
		return ValidatePassback{}, err
	}

	log.Info().Msg("Validated external inventory provider")
	return passback, nil
}
