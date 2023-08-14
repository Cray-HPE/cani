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

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/google/uuid"
)

type CommitResult struct {
	ProviderValidationErrors map[uuid.UUID]provider.HardwareValidationResult
}

func (d *Domain) Commit(ctx context.Context, dryrun bool) (CommitResult, error) {
	inventoryProvider := d.externalInventoryProvider

	// Perform validation integrity of CANI's inventory data
	// TODO handle validation result
	if _, err := d.datastore.Validate(); err != nil {
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
	return CommitResult{}, inventoryProvider.Reconcile(ctx, d.datastore, dryrun)

}
