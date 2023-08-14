/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
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

type ValidateResult struct {
	ProviderValidationErrors map[uuid.UUID]provider.HardwareValidationResult
}

func (d *Domain) Validate(ctx context.Context, checkRequiredData bool) (ValidateResult, error) {
	var result ValidateResult

	// Validate the current state of CANI's inventory data against the provider plugin
	// for provider specific data.
	if failedValidations, err := d.externalInventoryProvider.ValidateInternal(ctx, d.datastore, checkRequiredData); len(failedValidations) > 0 {
		result.ProviderValidationErrors = failedValidations
		return result, provider.ErrDataValidationFailure
	} else if err != nil {
		return ValidateResult{}, errors.Join(
			fmt.Errorf("failed to validate inventory against inventory provider plugin"),
			err,
		)
	}
	log.Info().Msg("Validated CANI inventory")

	// Validate external inventory data
	err := d.externalInventoryProvider.ValidateExternal(ctx, d.configOptions)
	if err != nil {
		return ValidateResult{}, err
	}

	log.Info().Msg("Validated external inventory provider")
	return result, nil
}

func (d *Domain) SetConfigOptions(ctx context.Context, domainOptions *NewOpts) error {
	options, err := d.externalInventoryProvider.ConfigOptions(ctx)
	if err != nil {
		return err
	}
	switch domainOptions.Provider {
	case string(inventory.CSMProvider):
		domainOptions.CsmOptions.ValidRoles = options.ValidRoles
		domainOptions.CsmOptions.ValidSubRoles = options.ValidSubRoles
	}

	return nil
}
