package import_

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
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
