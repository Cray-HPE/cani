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
package nautobot

import (
	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/internal/provider"
)

// NautobotOpts holds the options for the Nautobot provider
type NautobotOpts struct {
	// URL is the base URL of the Nautobot instance (e.g., http://localhost:8080/api)
	URL string `yaml:"url" json:"url" line_comment:"Base URL of the Nautobot instance"`

	// Token is the API token for authenticating with Nautobot
	Token string `yaml:"token" json:"token" line_comment:"API token for authentication (use --token to override)"`

	// DefaultLocation is the default location to use for devices without a location specified
	DefaultLocation string `yaml:"default_location" json:"default_location" line_comment:"Default location for devices"`

	// DefaultRole is the default role to use for devices without a role specified
	DefaultRole string `yaml:"default_role" json:"default_role" line_comment:"Default role for devices"`

	// DefaultStatus is the default status to use for devices without a status specified
	DefaultStatus string `yaml:"default_status" json:"default_status" line_comment:"Default status for devices"`

	// Import holds import-specific options
	Import *NautobotImportOpts `yaml:"import" json:"import" line_comment:"Import command options"`

	// Export holds export-specific options
	Export *NautobotExportOpts `yaml:"export" json:"export" line_comment:"Export command options"`
}

// NautobotImportOpts holds options specific to the import command
type NautobotImportOpts struct {
	// Reserved for future import-specific options
}

// NautobotExportOpts holds options specific to the export command
type NautobotExportOpts struct {
	CreateDeviceTypes   bool `yaml:"create_device_types" json:"create_device_types" line_comment:"Create missing device types in Nautobot"`
	CreateLocationTypes bool `yaml:"create_location_types" json:"create_location_types" line_comment:"Create missing location types in Nautobot"`
	CreateModuleTypes   bool `yaml:"create_module_types" json:"create_module_types" line_comment:"Create missing module types in Nautobot"`
	CreateLocations     bool `yaml:"create_locations" json:"create_locations" line_comment:"Create missing locations in Nautobot"`
	CreateStatuses      bool `yaml:"create_statuses" json:"create_statuses" line_comment:"Create missing statuses in Nautobot"`
	CreateRoles         bool `yaml:"create_roles" json:"create_roles" line_comment:"Create missing roles in Nautobot"`
	Merge               bool `yaml:"merge" json:"merge" line_comment:"Merge with existing devices instead of skipping conflicts"`
	DryRun              bool `yaml:"dry_run" json:"dry_run" line_comment:"Log planned actions without making API calls"`
}

// --- HasOptions interface implementation ---

// GetDefaultOptions returns the current configuration options for the Nautobot provider
func (p *Nautobot) GetDefaultOptions() map[string]any {
	return provider.StructToMapAll(p.Options)
}

// GetOptionsStruct returns the configuration struct instance for comment extraction
func (p *Nautobot) GetOptionsStruct() interface{} {
	return &NautobotOpts{
		Import: &NautobotImportOpts{},
		Export: &NautobotExportOpts{},
	}
}

// --- HasImportOptions interface implementation ---

// GetImportOptionsStruct returns a pointer to the import options struct
func (p *Nautobot) GetImportOptionsStruct() interface{} {
	return &NautobotImportOpts{}
}

// GetImportDefaults returns the default import configuration options
func (p *Nautobot) GetImportDefaults() map[string]any {
	return provider.StructToMapAll(&NautobotImportOpts{})
}

// BindImportFlags satisfies HasImportOptions.  Import flags are read directly
// from cmd.Flags() in the import path, so no binding is required.
func (p *Nautobot) BindImportFlags(cmd *cli.Command) error {
	return nil
}

// --- HasExportOptions interface implementation ---

// GetExportOptionsStruct returns a pointer to the export options struct
func (p *Nautobot) GetExportOptionsStruct() interface{} {
	return &NautobotExportOpts{}
}

// GetExportDefaults returns the default export configuration options
func (p *Nautobot) GetExportDefaults() map[string]any {
	return provider.StructToMapAll(&NautobotExportOpts{
		CreateDeviceTypes:   true,
		CreateLocationTypes: true,
		CreateModuleTypes:   true,
		CreateLocations:     true,
		CreateStatuses:      true,
		CreateRoles:         true,
		Merge:               false,
		DryRun:              false,
	})
}

// BindExportFlags satisfies HasExportOptions.  Export flags are read directly
// from cmd.Flags() via applyFlagOverrides, so no binding is required.
func (p *Nautobot) BindExportFlags(cmd *cli.Command) error {
	return nil
}

// LoadOptionsFromEnv loads options from environment variables and the config
// file (precedence: env var > config file > defaults).
func (p *Nautobot) LoadOptionsFromEnv() {
	if url := config.LookupString(providerSlug, "url"); url != "" {
		p.Options.URL = url
	}
	if token := config.LookupString(providerSlug, "token"); token != "" {
		p.Options.Token = token
	}

	p.loadDefaultsFromEnv()
	p.loadExportOptsFromEnv()
}

// loadDefaultsFromEnv loads provider-global defaults from env/config with legacy fallback.
func (p *Nautobot) loadDefaultsFromEnv() {
	if loc := config.LookupString(providerSlug, "default_location"); loc != "" {
		p.Options.DefaultLocation = loc
	} else if loc := config.LookupString(providerSlug, "import", "default_location"); loc != "" {
		p.Options.DefaultLocation = loc
	}
	if role := config.LookupString(providerSlug, "default_role"); role != "" {
		p.Options.DefaultRole = role
	} else if role := config.LookupString(providerSlug, "import", "default_role"); role != "" {
		p.Options.DefaultRole = role
	}
	if status := config.LookupString(providerSlug, "default_status"); status != "" {
		p.Options.DefaultStatus = status
	} else if status := config.LookupString(providerSlug, "import", "default_status"); status != "" {
		p.Options.DefaultStatus = status
	}

	if p.Options.Import == nil {
		p.Options.Import = &NautobotImportOpts{}
	}
}

// loadExportOptsFromEnv loads export-specific options from env/config.
func (p *Nautobot) loadExportOptsFromEnv() {
	if p.Options.Export == nil {
		p.Options.Export = &NautobotExportOpts{}
	}
	p.Options.Export.CreateDeviceTypes = config.LookupBool(providerSlug, "export", "create_device_types")
	p.Options.Export.CreateLocationTypes = config.LookupBool(providerSlug, "export", "create_location_types")
	p.Options.Export.CreateModuleTypes = config.LookupBool(providerSlug, "export", "create_module_types")
	p.Options.Export.CreateLocations = config.LookupBool(providerSlug, "export", "create_locations")
	p.Options.Export.CreateStatuses = config.LookupBool(providerSlug, "export", "create_statuses")
	p.Options.Export.CreateRoles = config.LookupBool(providerSlug, "export", "create_roles")
	p.Options.Export.Merge = config.LookupBool(providerSlug, "export", "merge")
	p.Options.Export.DryRun = config.LookupBool(providerSlug, "export", "dry_run")
}

// loadOptionsFromConfig loads the Nautobot options from the config file
func (p *Nautobot) loadOptionsFromConfig() error {
	if url, ok := config.GetNestedValue(providerSlug, "url"); ok {
		if urlStr, ok := url.(string); ok {
			p.Options.URL = urlStr
		}
	}
	if token, ok := config.GetNestedValue(providerSlug, "token"); ok {
		if tokenStr, ok := token.(string); ok {
			p.Options.Token = tokenStr
		}
	}
	if p.Options.Import == nil {
		p.Options.Import = &NautobotImportOpts{}
	}
	if p.Options.Export == nil {
		p.Options.Export = &NautobotExportOpts{}
	}

	p.loadDefaultsFromConfig()
	p.loadExportOptsFromConfig()
	return nil
}

// defaultFromConfig returns a provider-global default string, preferring the
// top-level key and falling back to the legacy "import" subsection.
func (p *Nautobot) defaultFromConfig(key string) (string, bool) {
	if v, ok := config.GetNestedValue(providerSlug, key); ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	if v, ok := config.GetNestedValue(providerSlug, "import", key); ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

// loadDefaultsFromConfig loads provider-global defaults from the config file with legacy fallback.
func (p *Nautobot) loadDefaultsFromConfig() {
	if v, ok := p.defaultFromConfig("default_location"); ok {
		p.Options.DefaultLocation = v
	}
	if v, ok := p.defaultFromConfig("default_role"); ok {
		p.Options.DefaultRole = v
	}
	if v, ok := p.defaultFromConfig("default_status"); ok {
		p.Options.DefaultStatus = v
	}
}

// exportBoolFromConfig returns an export-scoped boolean flag from the config file.
func (p *Nautobot) exportBoolFromConfig(key string) (bool, bool) {
	if v, ok := config.GetNestedValue(providerSlug, "export", key); ok {
		if b, ok := v.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// loadExportOptsFromConfig loads export-specific create/merge/dry-run flags from the config file.
func (p *Nautobot) loadExportOptsFromConfig() {
	if v, ok := p.exportBoolFromConfig("merge"); ok {
		p.Options.Export.Merge = v
	}
	if v, ok := p.exportBoolFromConfig("dry_run"); ok {
		p.Options.Export.DryRun = v
	}
	if v, ok := p.exportBoolFromConfig("create_device_types"); ok {
		p.Options.Export.CreateDeviceTypes = v
	}
	if v, ok := p.exportBoolFromConfig("create_locations"); ok {
		p.Options.Export.CreateLocations = v
	}
	if v, ok := p.exportBoolFromConfig("create_statuses"); ok {
		p.Options.Export.CreateStatuses = v
	}
	if v, ok := p.exportBoolFromConfig("create_roles"); ok {
		p.Options.Export.CreateRoles = v
	}
	if v, ok := p.exportBoolFromConfig("create_location_types"); ok {
		p.Options.Export.CreateLocationTypes = v
	}
	if v, ok := p.exportBoolFromConfig("create_module_types"); ok {
		p.Options.Export.CreateModuleTypes = v
	}
}
