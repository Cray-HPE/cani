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
	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	CreateDeviceTypes bool `yaml:"create_device_types" json:"create_device_types" line_comment:"Create missing device types in Nautobot"`
	CreateLocations   bool `yaml:"create_locations" json:"create_locations" line_comment:"Create missing locations in Nautobot"`
	CreateStatuses    bool `yaml:"create_statuses" json:"create_statuses" line_comment:"Create missing statuses in Nautobot"`
	CreateRoles       bool `yaml:"create_roles" json:"create_roles" line_comment:"Create missing roles in Nautobot"`
	Merge             bool `yaml:"merge" json:"merge" line_comment:"Merge with existing devices instead of skipping conflicts"`
	DryRun            bool `yaml:"dry_run" json:"dry_run" line_comment:"Log planned actions without making API calls"`
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

// BindImportFlags binds CLI flags to Viper for the import command
func (p *Nautobot) BindImportFlags(cmd *cobra.Command) error {
	_ = viper.BindPFlag("nautobot.default_location", cmd.Flags().Lookup("default-location"))
	_ = viper.BindPFlag("nautobot.default_role", cmd.Flags().Lookup("default-role"))
	_ = viper.BindPFlag("nautobot.default_status", cmd.Flags().Lookup("default-status"))
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
		CreateDeviceTypes: true,
		CreateLocations:   true,
		CreateStatuses:    true,
		CreateRoles:       true,
		Merge:             false,
		DryRun:            false,
	})
}

// BindExportFlags binds CLI flags to Viper for the export command
func (p *Nautobot) BindExportFlags(cmd *cobra.Command) error {
	_ = viper.BindPFlag("nautobot.export.create_device_types", cmd.Flags().Lookup("create-device-types"))
	_ = viper.BindPFlag("nautobot.export.create_locations", cmd.Flags().Lookup("create-locations"))
	_ = viper.BindPFlag("nautobot.export.create_statuses", cmd.Flags().Lookup("create-statuses"))
	_ = viper.BindPFlag("nautobot.export.create_roles", cmd.Flags().Lookup("create-roles"))
	_ = viper.BindPFlag("nautobot.export.merge", cmd.Flags().Lookup("merge"))
	_ = viper.BindPFlag("nautobot.export.dry_run", cmd.Flags().Lookup("dry-run"))
	return nil
}

// LoadOptionsFromViper loads options from Viper (handles precedence: CLI > env > config > defaults)
func (p *Nautobot) LoadOptionsFromViper() {
	if url := viper.GetString("nautobot.url"); url != "" {
		p.Options.URL = url
	}
	if token := viper.GetString("nautobot.token"); token != "" {
		p.Options.Token = token
	}

	p.loadDefaultsFromViper()
	p.loadExportOptsFromViper()
}

// loadDefaultsFromViper loads provider-global defaults from Viper with legacy fallback.
func (p *Nautobot) loadDefaultsFromViper() {
	if loc := viper.GetString("nautobot.default_location"); loc != "" {
		p.Options.DefaultLocation = loc
	} else if loc := viper.GetString("nautobot.import.default_location"); loc != "" {
		p.Options.DefaultLocation = loc
	}
	if role := viper.GetString("nautobot.default_role"); role != "" {
		p.Options.DefaultRole = role
	} else if role := viper.GetString("nautobot.import.default_role"); role != "" {
		p.Options.DefaultRole = role
	}
	if status := viper.GetString("nautobot.default_status"); status != "" {
		p.Options.DefaultStatus = status
	} else if status := viper.GetString("nautobot.import.default_status"); status != "" {
		p.Options.DefaultStatus = status
	}

	if p.Options.Import == nil {
		p.Options.Import = &NautobotImportOpts{}
	}
}

// loadExportOptsFromViper loads export-specific options from Viper.
func (p *Nautobot) loadExportOptsFromViper() {
	if p.Options.Export == nil {
		p.Options.Export = &NautobotExportOpts{}
	}
	p.Options.Export.CreateDeviceTypes = viper.GetBool("nautobot.export.create_device_types")
	p.Options.Export.CreateLocations = viper.GetBool("nautobot.export.create_locations")
	p.Options.Export.CreateStatuses = viper.GetBool("nautobot.export.create_statuses")
	p.Options.Export.CreateRoles = viper.GetBool("nautobot.export.create_roles")
	p.Options.Export.Merge = viper.GetBool("nautobot.export.merge")
	p.Options.Export.DryRun = viper.GetBool("nautobot.export.dry_run")
}

// loadOptionsFromConfig loads the Nautobot options from the config file
func (p *Nautobot) loadOptionsFromConfig() error {
	if url, ok := config.GetNestedValue("nautobot", "url"); ok {
		if urlStr, ok := url.(string); ok {
			p.Options.URL = urlStr
		}
	}
	if token, ok := config.GetNestedValue("nautobot", "token"); ok {
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

// loadDefaultsFromConfig loads provider-global defaults from the config file with legacy fallback.
func (p *Nautobot) loadDefaultsFromConfig() {
	if loc, ok := config.GetNestedValue("nautobot", "default_location"); ok {
		if locStr, ok := loc.(string); ok {
			p.Options.DefaultLocation = locStr
		}
	} else if loc, ok := config.GetNestedValue("nautobot", "import", "default_location"); ok {
		if locStr, ok := loc.(string); ok {
			p.Options.DefaultLocation = locStr
		}
	}
	if role, ok := config.GetNestedValue("nautobot", "default_role"); ok {
		if roleStr, ok := role.(string); ok {
			p.Options.DefaultRole = roleStr
		}
	} else if role, ok := config.GetNestedValue("nautobot", "import", "default_role"); ok {
		if roleStr, ok := role.(string); ok {
			p.Options.DefaultRole = roleStr
		}
	}
	if status, ok := config.GetNestedValue("nautobot", "default_status"); ok {
		if statusStr, ok := status.(string); ok {
			p.Options.DefaultStatus = statusStr
		}
	} else if status, ok := config.GetNestedValue("nautobot", "import", "default_status"); ok {
		if statusStr, ok := status.(string); ok {
			p.Options.DefaultStatus = statusStr
		}
	}
}

// loadExportOptsFromConfig loads export-specific create/merge/dry-run flags from the config file.
func (p *Nautobot) loadExportOptsFromConfig() {
	if merge, ok := config.GetNestedValue("nautobot", "export", "merge"); ok {
		if mergeBool, ok := merge.(bool); ok {
			p.Options.Export.Merge = mergeBool
		}
	}
	if dryRun, ok := config.GetNestedValue("nautobot", "export", "dry_run"); ok {
		if dryRunBool, ok := dryRun.(bool); ok {
			p.Options.Export.DryRun = dryRunBool
		}
	}
	if val, ok := config.GetNestedValue("nautobot", "export", "create_device_types"); ok {
		if b, ok := val.(bool); ok {
			p.Options.Export.CreateDeviceTypes = b
		}
	}
	if val, ok := config.GetNestedValue("nautobot", "export", "create_locations"); ok {
		if b, ok := val.(bool); ok {
			p.Options.Export.CreateLocations = b
		}
	}
	if val, ok := config.GetNestedValue("nautobot", "export", "create_statuses"); ok {
		if b, ok := val.(bool); ok {
			p.Options.Export.CreateStatuses = b
		}
	}
	if val, ok := config.GetNestedValue("nautobot", "export", "create_roles"); ok {
		if b, ok := val.(bool); ok {
			p.Options.Export.CreateRoles = b
		}
	}
}
