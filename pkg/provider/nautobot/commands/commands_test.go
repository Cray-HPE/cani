package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewImportCommand(t *testing.T) {
	base := &cobra.Command{Use: "import"}

	cmd, err := NewImportCommand(base)
	if err != nil {
		t.Fatalf("NewImportCommand: %v", err)
	}
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	expectedFlags := []struct {
		name     string
		defValue string
	}{
		{"default-location", ""},
		{"default-role", ""},
		{"default-status", "Active"},
	}
	for _, ef := range expectedFlags {
		f := cmd.Flags().Lookup(ef.name)
		if f == nil {
			t.Errorf("expected flag %q to be registered", ef.name)
			continue
		}
		if f.DefValue != ef.defValue {
			t.Errorf("flag %q: expected default %q, got %q", ef.name, ef.defValue, f.DefValue)
		}
	}
}

func TestNewImportCommand_FlagTypes(t *testing.T) {
	cmd, _ := NewImportCommand(&cobra.Command{})

	tests := []struct {
		flag     string
		flagType string
	}{
		{"default-location", "string"},
		{"default-role", "string"},
		{"default-status", "string"},
	}
	for _, tt := range tests {
		f := cmd.Flags().Lookup(tt.flag)
		if f == nil {
			t.Errorf("flag %q not found", tt.flag)
			continue
		}
		if f.Value.Type() != tt.flagType {
			t.Errorf("flag %q: expected type %s, got %s", tt.flag, tt.flagType, f.Value.Type())
		}
	}
}

func TestNewExportCommand(t *testing.T) {
	base := &cobra.Command{Use: "export"}

	cmd, err := NewExportCommand(base)
	if err != nil {
		t.Fatalf("NewExportCommand: %v", err)
	}
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	expectedFlags := []struct {
		name     string
		defValue string
	}{
		{"create-device-types", "true"},
		{"create-location-types", "true"},
		{"create-module-types", "true"},
		{"create-locations", "true"},
		{"create-statuses", "true"},
		{"create-roles", "true"},
		{"merge", "false"},
		{"dry-run", "false"},
	}
	for _, ef := range expectedFlags {
		f := cmd.Flags().Lookup(ef.name)
		if f == nil {
			t.Errorf("expected flag %q to be registered", ef.name)
			continue
		}
		if f.DefValue != ef.defValue {
			t.Errorf("flag %q: expected default %q, got %q", ef.name, ef.defValue, f.DefValue)
		}
	}
}

func TestNewExportCommand_FlagTypes(t *testing.T) {
	cmd, _ := NewExportCommand(&cobra.Command{})

	boolFlags := []string{
		"create-device-types", "create-location-types", "create-module-types",
		"create-locations", "create-statuses", "create-roles",
		"merge", "dry-run",
	}
	for _, name := range boolFlags {
		f := cmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("flag %q not found", name)
			continue
		}
		if f.Value.Type() != "bool" {
			t.Errorf("flag %q: expected type bool, got %s", name, f.Value.Type())
		}
	}
}

func TestNewExportCommand_UsageStrings(t *testing.T) {
	cmd, _ := NewExportCommand(&cobra.Command{})

	tests := []struct {
		flag  string
		usage string
	}{
		{"merge", "Merge with existing devices instead of skipping conflicts"},
		{"dry-run", "Log planned actions without making API calls"},
	}
	for _, tt := range tests {
		f := cmd.Flags().Lookup(tt.flag)
		if f == nil {
			t.Errorf("flag %q not found", tt.flag)
			continue
		}
		if f.Usage != tt.usage {
			t.Errorf("flag %q: expected usage %q, got %q", tt.flag, tt.usage, f.Usage)
		}
	}
}
