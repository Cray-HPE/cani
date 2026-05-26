package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewImportCommand(t *testing.T) {
	base := &cobra.Command{Use: "import"}
	cmd, err := NewImportCommand(base)
	if err != nil {
		t.Fatalf("NewImportCommand() error = %v", err)
	}
	if cmd == nil {
		t.Fatal("NewImportCommand() returned nil")
	}

	fileFlag := cmd.Flags().Lookup("file")
	if fileFlag == nil {
		t.Error("missing --file flag")
	} else {
		if fileFlag.Shorthand != "f" {
			t.Errorf("--file shorthand = %q, want %q", fileFlag.Shorthand, "f")
		}
	}

	csvFlag := cmd.Flags().Lookup("csv")
	if csvFlag == nil {
		t.Error("missing --csv flag")
	} else {
		if csvFlag.Shorthand != "c" {
			t.Errorf("--csv shorthand = %q, want %q", csvFlag.Shorthand, "c")
		}
	}
}

func TestNewExportCommand(t *testing.T) {
	base := &cobra.Command{Use: "export"}
	cmd, err := NewExportCommand(base)
	if err != nil {
		t.Fatalf("NewExportCommand() error = %v", err)
	}
	if cmd == nil {
		t.Fatal("NewExportCommand() returned nil")
	}
}

func TestNewShowCommand(t *testing.T) {
	base := &cobra.Command{Use: "show"}
	cmd, err := NewShowCommand(base)
	if err != nil {
		t.Fatalf("NewShowCommand() error = %v", err)
	}
	if cmd == nil {
		t.Fatal("NewShowCommand() returned nil")
	}
}

func TestNewAddCommand(t *testing.T) {
	base := &cobra.Command{Use: "add"}
	cmd, err := NewAddCommand(base)
	if err != nil {
		t.Fatalf("NewAddCommand() error = %v", err)
	}
	if cmd == nil {
		t.Fatal("NewAddCommand() returned nil")
	}
}

func TestNewRemoveCommand(t *testing.T) {
	base := &cobra.Command{Use: "remove"}
	cmd, err := NewRemoveCommand(base)
	if err != nil {
		t.Fatalf("NewRemoveCommand() error = %v", err)
	}
	if cmd == nil {
		t.Fatal("NewRemoveCommand() returned nil")
	}
}

func TestNewUpdateCommand(t *testing.T) {
	base := &cobra.Command{Use: "update"}
	cmd, err := NewUpdateCommand(base)
	if err != nil {
		t.Fatalf("NewUpdateCommand() error = %v", err)
	}
	if cmd == nil {
		t.Fatal("NewUpdateCommand() returned nil")
	}
}
