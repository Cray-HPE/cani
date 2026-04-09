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
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

type Provider interface {
	// Transform runs after Import
	// It typically does the "Transform" step in ETL to convert the external source data into CANI's format
	// CANI will load the existing inventory from the datastore and pass it to Transform in case the provider needs to
	// check for existing devices or racks.
	// Returns a devicetypes.TransformResult containing devices, racks, and cables (nil maps indicate not applicable)
	Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error)

	// NewProviderCmd returns a new cobra.Command for the provider
	// This is used to add provider-specific info the the CLI
	// This is usually a switch statement that returns a command for each command the provider supports
	NewProviderCmd(base *cobra.Command) (*cobra.Command, error)

	// Slug simply returns the provider's slug, which is used to identify it in the system
	Slug() string
}

// Exporter is an optional interface that providers can implement to support
// syncing the local inventory to an external system (the "Load" step in ETL)
type Exporter interface {
	// Export syncs the local inventory to the external system
	// Returns information about what was created, updated, skipped, or errored
	Export(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error
}

// Exporter is an optional interface that providers can implement to support
// syncing the local inventory to an external system (the "Load" step in ETL)
type Importer interface {
	// Import syncs the local inventory from an external system or source
	// This usually imports data into the local inventory from files or APIs
	// and saves it into the provider struct for later processing in Transform()
	Import(cmd *cobra.Command, args []string, existing *devicetypes.Inventory) error
}

// HasOptions is an optional interface that providers can implement
// to expose their default configuration options
type HasOptions interface {
	// GetDefaultOptions returns the default configuration options for the provider
	GetDefaultOptions() map[string]any

	// GetOptionsStruct returns the configuration struct instance for comment extraction
	// This is used during YAML serialization to preserve field ordering and comments
	GetOptionsStruct() any
}

// HasImportOptions is an optional interface that providers can implement
// to expose import-specific configuration options that correlate to CLI flags
type HasImportOptions interface {
	// GetImportOptionsStruct returns a pointer to the import options struct
	// This enables reflection over struct fields for field ordering and comments
	GetImportOptionsStruct() any

	// GetImportDefaults returns the default import configuration options
	// These are auto-populated in the config file if they do not exist
	GetImportDefaults() map[string]any

	// BindImportFlags binds CLI flags to Viper for the import command
	// This enables the precedence: CLI flags > env vars > config file > defaults
	BindImportFlags(cmd *cobra.Command) error
}

// HasExportOptions is an optional interface that providers can implement
// to expose export-specific configuration options that correlate to CLI flags
type HasExportOptions interface {
	// GetExportOptionsStruct returns a pointer to the export options struct
	// This enables reflection over struct fields for field ordering and comments
	GetExportOptionsStruct() any

	// GetExportDefaults returns the default export configuration options
	// These are auto-populated in the config file if they do not exist
	GetExportDefaults() map[string]any

	// BindExportFlags binds CLI flags to Viper for the export command
	// This enables the precedence: CLI flags > env vars > config file > defaults
	BindExportFlags(cmd *cobra.Command) error
}

// DeviceStager is an optional interface that providers can implement to
// support auto-staging existing devices during the add workflow.
type DeviceStager interface {
	// StageExisting finds the first eligible device matching slug in the
	// inventory, sets it to Staged, and recursively stages children
	// according to device-bay defaults.  Returns true if a device was staged.
	StageExisting(inv *devicetypes.Inventory, slug string) bool
}

// RackStager is an optional interface for creating new devices under
// staged racks when no existing device can be re-staged.
type RackStager interface {
	// StageNewInRack creates a new device hierarchy under the first
	// staged rack matching the slug's class. Returns true if created.
	StageNewInRack(inv *devicetypes.Inventory, slug string) bool
}
