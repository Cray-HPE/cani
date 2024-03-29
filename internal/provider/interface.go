/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func Init() {
	log.Info().Msgf("%+v", "github.com/Cray-HPE/cani/internal/provider.init")
}

var ErrDataValidationFailure = fmt.Errorf("data validation failure")

// TODO Need to think about how internal data structures should be supplied to the Inventory Provider
type InventoryProvider interface {
	// Validate the external services of the inventory provider are correct
	ValidateExternal(cmd *cobra.Command, args []string) error

	// Validate the representation of the inventory data into the destination inventory system
	// is consistent. The default set of checks will verify all currently provided data is valid.
	// If enableRequiredDataChecks is set to true, additional checks focusing on missing data will be ran.
	ValidateInternal(cmd *cobra.Command, args []string, datastore inventory.Datastore, enableRequiredDataChecks bool) (map[uuid.UUID]HardwareValidationResult, error)

	// Import external inventory data into CANI's inventory format
	// This initializes the data and replaces the existing data
	ImportInit(cmd *cobra.Command, args []string, datastore inventory.Datastore) error

	// Import external inventory data after initialization
	// This import should merge the imported data with the existing data
	Import(cmd *cobra.Command, args []string, datastore inventory.Datastore) error

	// Export inventory in various formats
	Export(cmd *cobra.Command, args []string, datastore inventory.Datastore) error

	// Reconcile CANI's inventory state with the external inventory state and apply required changes
	Reconcile(cmd *cobra.Command, args []string, datastore inventory.Datastore, dryrun bool, ignoreExternalValidation bool) error

	// RecommendHardware returns recommended settings for adding hardware based on the deviceTypeSlug
	RecommendHardware(inv inventory.Inventory, cmd *cobra.Command, args []string, auto bool) (recommended HardwareRecommendations, err error)

	// SetProviderOptions are specific to the Provider. For example, supported Roles and SubRoles
	SetProviderOptions(cmd *cobra.Command, args []string) error

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

	// Print
	PrintHardware(cmd *cobra.Command, args []string, filtered map[uuid.UUID]inventory.Hardware) error
	PrintRecommendations(cmd *cobra.Command, args []string, recommendations HardwareRecommendations) error

	// Provider's name
	Slug() string
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

// ProviderCommands is an interface for the commands that are specific to a provider
// this isn't usually used directly, but is used to generate the commands with 'makeprovdier'
type ProviderCommands interface {
	NewProviderCmd(caniCmd *cobra.Command) (providerCmd *cobra.Command, err error)
	NewSessionInitCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error)
	NewAddCabinetCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error)
	NewUpdateCabinetCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error)
	NewListCabinetCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error)
	NewAddBladeCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error)
	NewUpdateBladeCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error)
	NewListBladeCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error)
	NewAddNodeCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error)
	NewUpdateNodeCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error)
	NewListNodeCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error)
	NewExportCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error)
	NewImportCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error)
}
