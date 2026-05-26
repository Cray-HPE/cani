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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestGetDefaultOptions(t *testing.T) {
	p := New()
	opts := p.GetDefaultOptions()

	if opts == nil {
		t.Fatal("expected non-nil map")
	}
	if opts["url"] != "http://localhost:8081/api" {
		t.Errorf("url = %v, want %q", opts["url"], "http://localhost:8081/api")
	}
}

func TestGetOptionsStruct(t *testing.T) {
	p := New()
	s := p.GetOptionsStruct()

	if s == nil {
		t.Fatal("expected non-nil struct")
	}
	opts, ok := s.(*NautobotOpts)
	if !ok {
		t.Fatalf("expected *NautobotOpts, got %T", s)
	}
	if opts.Import == nil {
		t.Error("expected non-nil Import sub-struct")
	}
	if opts.Export == nil {
		t.Error("expected non-nil Export sub-struct")
	}
}

func TestGetImportOptionsStruct(t *testing.T) {
	p := New()
	s := p.GetImportOptionsStruct()

	if s == nil {
		t.Fatal("expected non-nil struct")
	}
	if _, ok := s.(*NautobotImportOpts); !ok {
		t.Fatalf("expected *NautobotImportOpts, got %T", s)
	}
}

func TestGetImportDefaults(t *testing.T) {
	p := New()
	d := p.GetImportDefaults()

	if d == nil {
		t.Fatal("expected non-nil map")
	}
}

func TestGetExportOptionsStruct(t *testing.T) {
	p := New()
	s := p.GetExportOptionsStruct()

	if s == nil {
		t.Fatal("expected non-nil struct")
	}
	if _, ok := s.(*NautobotExportOpts); !ok {
		t.Fatalf("expected *NautobotExportOpts, got %T", s)
	}
}

func TestGetExportDefaults(t *testing.T) {
	p := New()
	d := p.GetExportDefaults()

	if d == nil {
		t.Fatal("expected non-nil map")
	}

	// Verify the create flags default to true.
	boolFields := []string{
		"create_device_types", "create_location_types", "create_module_types",
		"create_locations", "create_statuses", "create_roles",
	}
	for _, key := range boolFields {
		val, ok := d[key]
		if !ok {
			t.Errorf("missing key %q in export defaults", key)
			continue
		}
		if val != true {
			t.Errorf("export default %q = %v, want true", key, val)
		}
	}

	// Verify merge and dry_run default to false.
	for _, key := range []string{"merge", "dry_run"} {
		val, ok := d[key]
		if !ok {
			t.Errorf("missing key %q in export defaults", key)
			continue
		}
		if val != false {
			t.Errorf("export default %q = %v, want false", key, val)
		}
	}
}

func TestBindImportFlags(t *testing.T) {
	p := New()
	cmd := &cobra.Command{}
	cmd.Flags().String("default-location", "", "")
	cmd.Flags().String("default-role", "", "")
	cmd.Flags().String("default-status", "", "")

	err := p.BindImportFlags(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBindExportFlags(t *testing.T) {
	p := New()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("create-device-types", true, "")
	cmd.Flags().Bool("create-location-types", true, "")
	cmd.Flags().Bool("create-module-types", true, "")
	cmd.Flags().Bool("create-locations", true, "")
	cmd.Flags().Bool("create-statuses", true, "")
	cmd.Flags().Bool("create-roles", true, "")
	cmd.Flags().Bool("merge", false, "")
	cmd.Flags().Bool("dry-run", false, "")

	err := p.BindExportFlags(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadOptionsFromViper(t *testing.T) {
	t.Run("loads URL and token from viper", func(t *testing.T) {
		v := viper.New()
		viper.Reset()
		defer viper.Reset()

		viper.Set("nautobot.url", "http://nb.example.com/api")
		viper.Set("nautobot.token", "secret-token")

		p := New()
		p.LoadOptionsFromViper()

		if p.Options.URL != "http://nb.example.com/api" {
			t.Errorf("URL = %q, want %q", p.Options.URL, "http://nb.example.com/api")
		}
		if p.Options.Token != "secret-token" {
			t.Errorf("Token = %q, want %q", p.Options.Token, "secret-token")
		}
		_ = v
	})

	t.Run("loads default location from primary key", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		viper.Set("nautobot.default_location", "Site-A")

		p := New()
		p.LoadOptionsFromViper()

		if p.Options.DefaultLocation != "Site-A" {
			t.Errorf("DefaultLocation = %q, want %q", p.Options.DefaultLocation, "Site-A")
		}
	})

	t.Run("falls back to legacy import key", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		viper.Set("nautobot.import.default_location", "Legacy-Site")

		p := New()
		p.LoadOptionsFromViper()

		if p.Options.DefaultLocation != "Legacy-Site" {
			t.Errorf("DefaultLocation = %q, want %q", p.Options.DefaultLocation, "Legacy-Site")
		}
	})

	t.Run("loads export options from viper", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		viper.Set("nautobot.export.merge", true)
		viper.Set("nautobot.export.dry_run", true)
		viper.Set("nautobot.export.create_device_types", true)

		p := New()
		p.LoadOptionsFromViper()

		if !p.Options.Export.Merge {
			t.Error("Export.Merge should be true")
		}
		if !p.Options.Export.DryRun {
			t.Error("Export.DryRun should be true")
		}
		if !p.Options.Export.CreateDeviceTypes {
			t.Error("Export.CreateDeviceTypes should be true")
		}
	})

	t.Run("initializes nil Import/Export sub-structs", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		p := New()
		p.Options.Import = nil
		p.Options.Export = nil

		p.LoadOptionsFromViper()

		if p.Options.Import == nil {
			t.Error("Import should be initialized")
		}
		if p.Options.Export == nil {
			t.Error("Export should be initialized")
		}
	})
}
