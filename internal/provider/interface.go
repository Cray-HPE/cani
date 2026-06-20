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
	"context"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

type Provider interface {
	// Transform runs after Import
	// It typically does the "Transform" step in ETL to convert the external source data into CANI's format
	// CANI will load the existing inventory from the datastore and pass it to Transform in case the provider needs to
	// check for existing devices or racks.
	// Returns a devicetypes.TransformResult containing devices, racks, and cables (nil maps indicate not applicable)
	// ctx carries cancellation and deadlines from the command layer.
	Transform(ctx context.Context, existing devicetypes.Inventory) (*devicetypes.TransformResult, error)

	// NewProviderCmd returns a new cli.Command for the provider
	// This is used to add provider-specific info the the CLI
	// This is usually a switch statement that returns a command for each command the provider supports
	NewProviderCmd(base *cli.Command) (*cli.Command, error)

	// Slug simply returns the provider's slug, which is used to identify it in the system
	Slug() string
}

// Exporter is an optional interface that providers can implement to support
// syncing the local inventory to an external system (the "Load" step in ETL)
type Exporter interface {
	// Export syncs the local inventory to the external system
	// Returns information about what was created, updated, skipped, or errored
	// ctx carries cancellation and deadlines from the command layer.
	Export(ctx context.Context, cmd *cli.Command, args []string, inventory *devicetypes.Inventory) error
}

// Exporter is an optional interface that providers can implement to support
// syncing the local inventory to an external system (the "Load" step in ETL)
type Importer interface {
	// Import syncs the local inventory from an external system or source
	// This usually imports data into the local inventory from files or APIs
	// and saves it into the provider struct for later processing in Transform()
	// ctx carries cancellation and deadlines from the command layer.
	Import(ctx context.Context, cmd *cli.Command, args []string, existing *devicetypes.Inventory) error
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

// Configurer is an optional interface that providers can implement to receive
// their own section of the configuration file at startup.  The command layer
// calls Configure once per registered provider, passing the provider's
// Cfg.Providers[slug] map, so the provider can decode and validate its
// settings before any command runs.  Returning an error aborts startup.
type Configurer interface {
	Configure(options map[string]any) error
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

	// BindImportFlags registers the import command's CLI flags
	// This enables the precedence: CLI flags > env vars > config file > defaults
	BindImportFlags(cmd *cli.Command) error
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

	// BindExportFlags registers the export command's CLI flags
	// This enables the precedence: CLI flags > env vars > config file > defaults
	BindExportFlags(cmd *cli.Command) error
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

// RackPostAddHook is an optional interface that providers can implement
// to run provider-specific logic after a rack is added to the inventory.
// For example, CSM uses this to assign cabinet numbers and VLANs.
type RackPostAddHook interface {
	// OnRackAdded is called after a rack has been created but before
	// the inventory is saved.  Providers may set provider-specific
	// metadata, status, or name on the rack.
	OnRackAdded(rack *devicetypes.CaniRackType, inventory *devicetypes.Inventory) error
}

// MetadataApplier is an optional interface that providers can implement to
// merge generic --metadata key=value pairs into provider-specific storage.
// The command layer parses the flag into a map and lets each registered
// provider decide where to store it (typically its own Slug() sub-map), so
// no provider name is hard-coded in cmd/.
type MetadataApplier interface {
	// ApplyMetadata merges meta into the provider's section of pm,
	// allocating the map if necessary.
	ApplyMetadata(pm *map[string]any, meta map[string]string)
}

// DeviceUpdateFlagProvider is an optional interface that providers can
// implement to contribute provider-specific flags to the generic
// "update device" command and apply them to a device.  This keeps
// provider-specific flags (e.g. CSM's --nid/--alias) out of cmd/.
type DeviceUpdateFlagProvider interface {
	// RegisterDeviceUpdateFlags adds the provider's flags to cmd.
	RegisterDeviceUpdateFlags(cmd *cli.Command)

	// ApplyDeviceUpdateFlags reads any changed provider flags from cmd
	// and applies them to device.  It is a no-op when no flags changed.
	ApplyDeviceUpdateFlags(cmd *cli.Command, device *devicetypes.CaniDeviceType) error
}

// StagedDeviceDescriber is an optional interface that providers can implement
// to supply human-readable description lines for a staged device (for example,
// CSM's xname-derived Cabinet/Chassis/Blade location).  The command layer logs
// the returned lines verbatim, so xname decoding stays in the provider.
type StagedDeviceDescriber interface {
	// DescribeStagedDevice returns zero or more description lines for the
	// device.  Returns nil when the provider has nothing to add.
	DescribeStagedDevice(device *devicetypes.CaniDeviceType) []string
}
