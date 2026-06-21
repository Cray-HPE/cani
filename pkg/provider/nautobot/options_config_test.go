/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
	"testing"

	"github.com/Cray-HPE/cani/internal/config"
)

// -----------------------------------------------------------------------------
// These tests exercise the config-file load path (loadOptionsFromConfig and its
// helpers), which reads from the global config.Cfg singleton rather than Viper.
// -----------------------------------------------------------------------------

// setProviderConfig installs a Nautobot provider config block into the global
// config singleton for the duration of a test, restoring the previous value
// afterward so tests stay independent and order-insensitive.
func setProviderConfig(t *testing.T, block map[string]any) {
	t.Helper()
	prev := config.Cfg
	config.Cfg = &config.Config{
		Providers: map[string]map[string]any{providerSlug: block},
	}
	t.Cleanup(func() { config.Cfg = prev })
}

func TestLoadOptionsFromConfig_PopulatesURLTokenAndDefaults(t *testing.T) {
	setProviderConfig(t, map[string]any{
		"url":              "http://nb.example/api",
		"token":            "abc123",
		"default_location": "Site-A",
		"default_role":     "compute",
		"default_status":   "Active",
		"export": map[string]any{
			"merge":   true,
			"dry_run": true,
		},
	})

	p := New()
	if err := p.loadOptionsFromConfig(); err != nil {
		t.Fatalf("loadOptionsFromConfig() error = %v", err)
	}

	if p.Options.URL != "http://nb.example/api" {
		t.Errorf("URL = %q, want %q", p.Options.URL, "http://nb.example/api")
	}
	if p.Options.Token != "abc123" {
		t.Errorf("Token = %q, want %q", p.Options.Token, "abc123")
	}
	if p.Options.DefaultLocation != "Site-A" {
		t.Errorf("DefaultLocation = %q, want %q", p.Options.DefaultLocation, "Site-A")
	}
	if p.Options.DefaultRole != "compute" {
		t.Errorf("DefaultRole = %q, want %q", p.Options.DefaultRole, "compute")
	}
	if p.Options.DefaultStatus != "Active" {
		t.Errorf("DefaultStatus = %q, want %q", p.Options.DefaultStatus, "Active")
	}
	if !p.Options.Export.Merge {
		t.Error("Export.Merge = false, want true")
	}
	if !p.Options.Export.DryRun {
		t.Error("Export.DryRun = false, want true")
	}
}

func TestLoadOptionsFromConfig_InitializesNilSubStructs(t *testing.T) {
	setProviderConfig(t, map[string]any{})

	p := New()
	p.Options.Import = nil
	p.Options.Export = nil

	if err := p.loadOptionsFromConfig(); err != nil {
		t.Fatalf("loadOptionsFromConfig() error = %v", err)
	}
	if p.Options.Import == nil {
		t.Error("Import sub-struct should be initialized")
	}
	if p.Options.Export == nil {
		t.Error("Export sub-struct should be initialized")
	}
}

func TestLoadOptionsFromConfig_IgnoresNonStringURLAndToken(t *testing.T) {
	setProviderConfig(t, map[string]any{
		"url":   42,
		"token": true,
	})

	p := New()
	if err := p.loadOptionsFromConfig(); err != nil {
		t.Fatalf("loadOptionsFromConfig() error = %v", err)
	}

	// The default URL from New() should be retained when the config value is
	// not a string.
	if p.Options.URL != "http://localhost:8081/api" {
		t.Errorf("URL = %q, want the default to be retained", p.Options.URL)
	}
	if p.Options.Token != "" {
		t.Errorf("Token = %q, want it to remain empty", p.Options.Token)
	}
}

func TestDefaultFromConfig_PrefersTopLevelOverLegacyImport(t *testing.T) {
	setProviderConfig(t, map[string]any{
		"default_role": "top-role",
		"import": map[string]any{
			"default_role": "legacy-role",
		},
	})

	p := New()
	got, ok := p.defaultFromConfig("default_role")
	if !ok {
		t.Fatal("expected default_role to be found")
	}
	if got != "top-role" {
		t.Errorf("default_role = %q, want %q (top-level wins)", got, "top-role")
	}
}

func TestDefaultFromConfig_FallsBackToLegacyImportSection(t *testing.T) {
	setProviderConfig(t, map[string]any{
		"import": map[string]any{
			"default_location": "legacy-loc",
		},
	})

	p := New()
	got, ok := p.defaultFromConfig("default_location")
	if !ok {
		t.Fatal("expected default_location to be found via the legacy import section")
	}
	if got != "legacy-loc" {
		t.Errorf("default_location = %q, want %q", got, "legacy-loc")
	}
}

func TestDefaultFromConfig_ReturnsFalseWhenAbsent(t *testing.T) {
	setProviderConfig(t, map[string]any{})

	p := New()
	if _, ok := p.defaultFromConfig("default_status"); ok {
		t.Error("expected ok=false when the key is absent in both sections")
	}
}

func TestExportBoolFromConfig_ReadsBoolUnderExportSection(t *testing.T) {
	setProviderConfig(t, map[string]any{
		"export": map[string]any{"merge": true},
	})

	p := New()
	got, ok := p.exportBoolFromConfig("merge")
	if !ok || !got {
		t.Errorf("exportBoolFromConfig(merge) = (%v, %v), want (true, true)", got, ok)
	}
}

func TestExportBoolFromConfig_ReturnsFalseForMissingKey(t *testing.T) {
	setProviderConfig(t, map[string]any{
		"export": map[string]any{},
	})

	p := New()
	if _, ok := p.exportBoolFromConfig("dry_run"); ok {
		t.Error("expected ok=false for a key missing from the export section")
	}
}

func TestExportBoolFromConfig_ReturnsFalseForNonBoolValue(t *testing.T) {
	setProviderConfig(t, map[string]any{
		"export": map[string]any{"merge": "not-a-bool"},
	})

	p := New()
	if _, ok := p.exportBoolFromConfig("merge"); ok {
		t.Error("expected ok=false when the value is not a bool")
	}
}

func TestLoadExportOptsFromConfig_AppliesAllFlags(t *testing.T) {
	setProviderConfig(t, map[string]any{
		"export": map[string]any{
			"create_device_types":   true,
			"create_location_types": true,
			"create_module_types":   true,
			"create_locations":      true,
			"create_statuses":       true,
			"create_roles":          true,
			"merge":                 true,
			"dry_run":               true,
		},
	})

	p := New()
	p.loadExportOptsFromConfig()

	e := p.Options.Export
	if !e.CreateDeviceTypes || !e.CreateLocationTypes || !e.CreateModuleTypes ||
		!e.CreateLocations || !e.CreateStatuses || !e.CreateRoles ||
		!e.Merge || !e.DryRun {
		t.Errorf("expected all export flags true, got %+v", e)
	}
}

func TestLoadDefaultsFromConfig_LeavesOptionsUnchangedWhenAbsent(t *testing.T) {
	setProviderConfig(t, map[string]any{})

	p := New()
	p.Options.DefaultLocation = "preset"
	p.loadDefaultsFromConfig()

	if p.Options.DefaultLocation != "preset" {
		t.Errorf("DefaultLocation = %q, want it unchanged when config is empty", p.Options.DefaultLocation)
	}
}
