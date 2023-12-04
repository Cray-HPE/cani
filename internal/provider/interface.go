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
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var ErrDataValidationFailure = fmt.Errorf("data validation failure")

// TODO Need to think about how internal data structures should be supplied to the Inventory Provider
type InventoryProvider interface {
	// Validate the external services of the inventory provider are correct
	ValidateExternal(cmd *cobra.Command, args []string) error

	// Validate the representation of the inventory data into the destination inventory system
	// is consistent. The default set of checks will verify all currently provided data is valid.
	// If enableRequiredDataChecks is set to true, additional checks focusing on missing data will be ran.
	ValidateInternal(ctx context.Context, datastore inventory.Datastore, enableRequiredDataChecks bool) (map[uuid.UUID]HardwareValidationResult, error)

	// Import external inventory data into CANI's inventory format
	Import(ctx context.Context, datastore inventory.Datastore) error

	ExportJson(ctx context.Context, datastore inventory.Datastore, skipValidation bool) ([]byte, error)

	// Reconcile CANI's inventory state with the external inventory state and apply required changes
	Reconcile(ctx context.Context, datastore inventory.Datastore, dryrun bool, ignoreExternalValidation bool) error

	// RecommendHardware returns recommended settings for adding hardware based on the deviceTypeSlug
	RecommendHardware(inv inventory.Inventory, cmd *cobra.Command, args []string, auto bool) (recommended HardwareRecommendations, err error)

	// SetProviderOptions are specific to the Provider. For example, supported Roles and SubRoles
	SetProviderOptions(cmd *cobra.Command, args []string) error

	// SetProviderOptionsInterface passes the options down to the provider as an interface
	// It must be type-asserted at that layer and then set
	SetProviderOptionsInterface(interface{}) error

	// GetProviderOptions gets the options from the provider as an interface
	// It must be type-asserted and then set at the domain layer
	GetProviderOptions() (interface{}, error)

	//
	// Provider Hardware Metadata
	//

	// Build metadata, and add ito the hardware object
	// This function could return the data to put into object
	BuildHardwareMetadata(hw *inventory.Hardware, cmd *cobra.Command, args []string, recommendations HardwareRecommendations) error
	NewHardwareMetadata(hw *inventory.Hardware, cmd *cobra.Command, args []string) error

	// Return values for the given fields from the hardware's metadata
	GetFields(hw *inventory.Hardware, fieldNames []string) (values []string, err error)

	// Set fields in the hardware's metadata
	SetFields(hw *inventory.Hardware, values map[string]string) (result SetFieldsResult, err error)

	// Return metadata about each field
	GetFieldMetadata() ([]FieldMetadata, error)

	// Workflows
	ListCabinetMetadataColumns() (columns []string)
	ListCabinetMetadataRow(inventory.Hardware) (values []string, err error)
}

type HardwareValidationResult struct {
	Hardware inventory.Hardware
	Errors   []string
}

type HardwareRecommendations struct {
	CabinetOrdinal   int
	ChassisOrdinal   int
	BladeOrdinal     int
	ProviderMetadata map[string]interface{}
}

type CsvImportResult struct {
	Total             int
	Modified          int
	ValidationResults map[uuid.UUID]HardwareValidationResult
}

type SetFieldsResult struct {
	ModifiedFields []string
}

type FieldMetadata struct {
	Name         string
	Types        string
	Description  string
	IsModifiable bool
}

func (r HardwareRecommendations) Print() {
	log.Info().Msgf("Suggested cabinet number: %d", r.CabinetOrdinal)
	log.Info().Msgf("Suggested VLAN ID: %d", r.ProviderMetadata["HMNVlan"])
}
