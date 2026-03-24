package import_

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func fixtureDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "testdata", "fixtures")
}

func TestParseSlsDumpstate(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(fixtureDir(), "csm/simulator/sls.json"))
	if err != nil {
		t.Fatalf("read sls.json: %v", err)
	}
	sls, err := ParseSlsDumpstate(data)
	if err != nil {
		t.Fatalf("ParseSlsDumpstate: %v", err)
	}
	if len(sls.Hardware) == 0 {
		t.Fatal("expected at least one hardware entry")
	}
	hw, ok := sls.Hardware["x3000"]
	if !ok {
		t.Fatal("expected hardware entry for x3000")
	}
	if hw.TypeString != "Cabinet" {
		t.Errorf("x3000 TypeString = %q, want Cabinet", hw.TypeString)
	}
	if hw.Class != "River" {
		t.Errorf("x3000 Class = %q, want River", hw.Class)
	}
}

func TestParseSlsDumpstate_Mountain(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(fixtureDir(), "sls/small_mountain.json"))
	if err != nil {
		t.Fatalf("read small_mountain.json: %v", err)
	}
	sls, err := ParseSlsDumpstate(data)
	if err != nil {
		t.Fatalf("ParseSlsDumpstate: %v", err)
	}
	if len(sls.Hardware) == 0 {
		t.Fatal("expected at least one hardware entry")
	}
}

func TestParseSmdComponents(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(fixtureDir(), "csm/simulator/smd.json"))
	if err != nil {
		t.Fatalf("read smd.json: %v", err)
	}
	smd, err := ParseSmdComponents(data)
	if err != nil {
		t.Fatalf("ParseSmdComponents: %v", err)
	}
	if len(smd.Components) == 0 {
		t.Fatal("expected at least one component")
	}
	found := false
	for _, c := range smd.Components {
		if c.ID == "x3000c0s9b0n0" {
			found = true
			if c.Role != "Management" {
				t.Errorf("x3000c0s9b0n0 Role = %q, want Management", c.Role)
			}
			if c.NID != 100005 {
				t.Errorf("x3000c0s9b0n0 NID = %d, want 100005", c.NID)
			}
			break
		}
	}
	if !found {
		t.Error("expected SMD component x3000c0s9b0n0")
	}
}

func TestDecodeExtraProperties(t *testing.T) {
	props := map[string]any{
		"NID":     float64(42),
		"Role":    "Compute",
		"SubRole": "UAN",
		"Aliases": []any{"nid000042"},
	}
	ep, err := DecodeExtraProperties[SlsNodeExtraProperties](props)
	if err != nil {
		t.Fatalf("DecodeExtraProperties: %v", err)
	}
	if ep.NID != 42 {
		t.Errorf("NID = %d, want 42", ep.NID)
	}
	if ep.Role != "Compute" {
		t.Errorf("Role = %q, want Compute", ep.Role)
	}
	if len(ep.Aliases) != 1 || ep.Aliases[0] != "nid000042" {
		t.Errorf("Aliases = %v, want [nid000042]", ep.Aliases)
	}
}

func TestDecodeExtraProperties_Empty(t *testing.T) {
	ep, err := DecodeExtraProperties[SlsNodeExtraProperties](nil)
	if err != nil {
		t.Fatalf("DecodeExtraProperties(nil): %v", err)
	}
	if ep.NID != 0 {
		t.Errorf("NID = %d, want 0", ep.NID)
	}
}

// newTestCmd builds a cobra.Command with the flags that import.go reads.
func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("use-simulator", false, "")
	cmd.Flags().String("csm-api-host", "", "")
	cmd.Flags().String("sls-file", "", "")
	cmd.Flags().String("smd-file", "", "")
	cmd.Flags().Bool("insecure", false, "")
	cmd.Flags().String("csm-keycloak-username", "", "")
	cmd.Flags().String("csm-keycloak-password", "", "")
	cmd.Flags().String("csm-url-sls", "", "")
	cmd.Flags().String("csm-url-hsm", "", "")
	cmd.Flags().String("csm-ca-cert", "", "")
	cmd.Flags().String("csm-k8s-pods-cidr", "", "")
	cmd.Flags().String("csm-k8s-services-cidr", "", "")
	cmd.Flags().String("csm-kube-config", "", "")
	cmd.Flags().String("csm-secret-name", "", "")
	cmd.Flags().String("csm-client-id", "", "")
	cmd.Flags().String("csm-client-secret", "", "")
	return cmd
}

func TestShouldUseAPI(t *testing.T) {
	tests := []struct {
		name string
		flag string
		val  string
		want bool
	}{
		{
			name: "simulator flag returns true",
			flag: "use-simulator",
			val:  "true",
			want: true,
		},
		{
			name: "no flags returns false",
			flag: "",
			val:  "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTestCmd()
			if tt.flag != "" {
				cmd.Flags().Set(tt.flag, tt.val)
			}
			got, err := ShouldUseAPI(cmd)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ShouldUseAPI() = %t, want %t", got, tt.want)
			}
		})
	}
}

// fakeImportProvider satisfies the interface expected by importFromFiles.
type fakeImportProvider struct {
	sls *SlsDumpstate
	smd *SmdComponentList
}

func (f *fakeImportProvider) ClearRawData()                { f.sls = nil; f.smd = nil }
func (f *fakeImportProvider) SetSls(sls *SlsDumpstate)     { f.sls = sls }
func (f *fakeImportProvider) SetSmd(smd *SmdComponentList) { f.smd = smd }

func TestImportFromFiles(t *testing.T) {
	slsFixture := filepath.Join(fixtureDir(), "csm/simulator/sls.json")

	tests := []struct {
		name      string
		slsFile   string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid SLS file imports successfully",
			slsFile:   slsFixture,
			expectErr: false,
		},
		{
			name:      "missing sls-file flag returns error",
			slsFile:   "",
			expectErr: true,
			errMsg:    "--sls-file is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTestCmd()
			if tt.slsFile != "" {
				cmd.Flags().Set("sls-file", tt.slsFile)
			}
			fake := &fakeImportProvider{}
			err := importFromFiles(cmd, fake)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want substring %q", err.Error(), tt.errMsg)
				}
			}
			if !tt.expectErr && fake.sls == nil {
				t.Error("expected SLS data to be set on provider")
			}
		})
	}
}

func TestReadAndParseSls(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		expectErr bool
	}{
		{
			name:      "valid fixture parses successfully",
			path:      filepath.Join(fixtureDir(), "csm/simulator/sls.json"),
			expectErr: false,
		},
		{
			name:      "nonexistent file returns error",
			path:      "nonexistent.json",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sls, err := readAndParseSls(tt.path)
			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectErr && len(sls.Hardware) == 0 {
				t.Error("expected at least one hardware entry")
			}
		})
	}
}

func TestReadAndParseSmd(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		expectErr bool
	}{
		{
			name:      "valid fixture parses successfully",
			path:      filepath.Join(fixtureDir(), "csm/simulator/smd.json"),
			expectErr: false,
		},
		{
			name:      "nonexistent file returns error",
			path:      "nonexistent.json",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smd, err := readAndParseSmd(tt.path)
			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectErr && len(smd.Components) == 0 {
				t.Error("expected at least one component")
			}
		})
	}
}
