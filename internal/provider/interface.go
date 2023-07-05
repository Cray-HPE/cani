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
package provider

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/google/uuid"
)

var ErrDataValidationFailure = fmt.Errorf("data validation failure")

// TODO Need to think about how internal data structures should be supplied to the Inventory Provider
type InventoryProvider interface {
	// Validate the external services of the inventory provider are correct
	ValidateExternal(ctx context.Context) error

	// Validate the representation of the inventory data into the destination inventory system
	// is consistent. The default set of checks will verify all currently provided data is valid.
	// If enableRequiredDataChecks is set to true, additional checks focusing on missing data will be ran.
	ValidateInternal(ctx context.Context, datastore inventory.Datastore, enableRequiredDataChecks bool) (map[uuid.UUID]HardwareValidationResult, error)

	// Import external inventory data into CANI's inventory format
	Import(ctx context.Context, datastore inventory.Datastore) error

	// Reconcile CANI's inventory state with the external inventory state and apply required changes
	Reconcile(ctx context.Context, datastore inventory.Datastore) error

	// Build metadata, and add ito the hardware object
	// This function could return the data to put into object
	BuildHardwareMetadata(hw *inventory.Hardware, rawProperties map[string]interface{}) error

	// RecommendCabinet returns recommended settings for adding a cabinet
	RecommendCabinet(inv inventory.Inventory, deviceTypeSlug string) (HardwareRecommendations, error)

	ExportCsv(ctx context.Context, datastore inventory.Datastore, writer *csv.Writer, headers []string) error

	ImportCsv(ctx context.Context, datastore inventory.Datastore, reader *csv.Reader) (result CsvImportResult, err error)
}

type HardwareValidationResult struct {
	Hardware inventory.Hardware
	Errors   []string
}

type HardwareRecommendations struct {
	LocationOrdinal  int
	ProviderMetadata map[string]interface{}
}

type CsvImportResult struct {
	Total             int
	Modified          int
	ValidationResults map[uuid.UUID]HardwareValidationResult
}
