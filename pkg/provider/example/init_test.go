package example

import (
	"testing"

	"github.com/Cray-HPE/cani/internal/cli"
)

func TestGetInstance(t *testing.T) {
	inst := GetInstance()
	if inst == nil {
		t.Fatal("GetInstance() returned nil")
	}
}

func TestNewProviderCmd(t *testing.T) {
	p := New()

	tests := []struct {
		name    string
		cmdName string
	}{
		{"import", "import"},
		{"export", "export"},
		{"show", "show"},
		{"add", "add"},
		{"remove", "remove"},
		{"update", "update"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := &cli.Command{Use: tt.cmdName}
			cmd, err := p.NewProviderCmd(base)
			if err != nil {
				t.Fatalf("NewProviderCmd(%q) error = %v", tt.cmdName, err)
			}
			if cmd == nil {
				t.Fatalf("NewProviderCmd(%q) returned nil", tt.cmdName)
			}
		})
	}
}

func TestNewProviderCmdImportFlags(t *testing.T) {
	p := New()
	base := &cli.Command{Use: "import"}

	cmd, err := p.NewProviderCmd(base)
	if err != nil {
		t.Fatalf("NewProviderCmd(import) error = %v", err)
	}
	if cmd.Flags().Lookup("file") == nil {
		t.Error("import command missing --file flag")
	}
	if cmd.Flags().Lookup("csv") == nil {
		t.Error("import command missing --csv flag")
	}
}

func TestNewProviderCmdUnknown(t *testing.T) {
	p := New()
	base := &cli.Command{Use: "unknown"}

	cmd, err := p.NewProviderCmd(base)
	if err != nil {
		t.Fatalf("NewProviderCmd(unknown) error = %v", err)
	}
	if cmd != base {
		t.Error("NewProviderCmd(unknown) should return the base command unchanged")
	}
}
