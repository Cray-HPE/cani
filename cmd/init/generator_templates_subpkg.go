package init

// Template definitions for subpackage files

const commandsSubpkgTemplate = `package commands

import "github.com/Cray-HPE/cani/internal/cli"

func NewImportCommand(base *cli.Command) (*cli.Command, error) {
	// TODO: Add import-specific flags or subcommands
	// Example:
	// base.Flags().String("source", "", "Source file to import")
	return &cli.Command{}, nil
}

func NewExportCommand(base *cli.Command) (*cli.Command, error) {
	// TODO: Add export-specific flags or subcommands
	return &cli.Command{}, nil
}

func NewShowCommand(base *cli.Command) (*cli.Command, error) {
	// TODO: Add show-specific flags or subcommands
	return &cli.Command{}, nil
}

func NewAddCommand(base *cli.Command) (*cli.Command, error) {
	// TODO: Add add-specific flags or subcommands
	return &cli.Command{}, nil
}

func NewRemoveCommand(base *cli.Command) (*cli.Command, error) {
	// TODO: Add remove-specific flags or subcommands
	return &cli.Command{}, nil
}

func NewUpdateCommand(base *cli.Command) (*cli.Command, error) {
	// TODO: Add update-specific flags or subcommands
	return &cli.Command{}, nil
}
`

const exportSubpkgTemplate = `package export

import "github.com/Cray-HPE/cani/pkg/devicetypes"

func Export(existing devicetypes.Inventory) error {
	// Common patterns:
	//   - Compare local inventory with external system
	//   - Create/update/delete resources in external system
	//   - Report what was created, updated, skipped, or errored
	return nil
}
`

const importSubpkgTemplate = `package import_

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/internal/cli"
)

func Import(cmd *cli.Command, args []string, inventory *devicetypes.Inventory) error {
	// Common patterns:
	//   - Parse files or query APIs to get data
	//   - Store data in provider struct for later processing in Transform()
	//   - Report what was imported, skipped, or errored
	return nil
}
`

const transformSubpkgTemplate = `package transform

import "github.com/Cray-HPE/cani/pkg/devicetypes"

// Transform converts imported data into CANI's inventory format.
// Returns a TransformResult containing devices, racks, and cables (nil maps indicate not applicable).
func Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	// TODO: Implement Transform
	// Common patterns:
	//   - Convert queued data from Extract to CaniDeviceType
	//   - Check existing inventory for duplicates
	//   - Set parent-child relationships
	//
	// Might accept queues or other data structures as parameters, which were collected during Import
	return &devicetypes.TransformResult{}, nil
}
`
