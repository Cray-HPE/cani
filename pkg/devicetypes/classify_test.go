package devicetypes

import (
	"testing"
)

func TestSuggestTypesCompoundName(t *testing.T) {
	// Register a device type that should match "dl360gen11" via sub-token
	// decomposition: "dl360" matches a token in the slug.
	RegisterDeviceType(CaniDeviceType{
		Slug:         "test-dl360-gen11-8sff",
		Model:        "ProLiant DL360 Gen11 8SFF",
		Manufacturer: "HPE",
		Type:         "blade",
	})
	defer func() {
		delete(allDeviceTypes, "test-dl360-gen11-8sff")
	}()

	device := UnclassifiedDevice{
		Name:       "dl360gen11",
		DeviceType: "compute",
	}
	results := SuggestTypes(device, 8)
	if len(results) == 0 {
		t.Fatal("SuggestTypes returned 0 results for dl360gen11, want >= 1")
	}

	found := false
	for _, r := range results {
		if r.Slug == "test-dl360-gen11-8sff" {
			found = true
			if r.Score < 30 {
				t.Errorf("test-dl360-gen11-8sff score = %d, want >= 30", r.Score)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected test-dl360-gen11-8sff in results, got %v", results)
	}
}

func TestSuggestTypesMgmtSwitchFallback(t *testing.T) {
	// Register a management switch device type.
	RegisterDeviceType(CaniDeviceType{
		Slug:         "test-aruba-2930f",
		Model:        "Aruba 2930F 48G 4SFP+",
		Manufacturer: "HPE",
		Type:         "mgmt-switch",
	})
	defer func() {
		delete(allDeviceTypes, "test-aruba-2930f")
	}()

	// "mgmtsw0" has no direct text match but HardwareType "mgmt_switch"
	// should trigger the hardware-type fallback.
	device := UnclassifiedDevice{
		Name:       "mgmtsw0",
		DeviceType: "mgmt_switch",
	}
	results := SuggestTypes(device, 8)
	if len(results) == 0 {
		t.Fatal("SuggestTypes returned 0 results for mgmtsw0, want >= 1")
	}

	found := false
	for _, r := range results {
		if r.Slug == "test-aruba-2930f" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected test-aruba-2930f in fallback results, got %v", results)
	}
}

func TestSuggestTypesFMN(t *testing.T) {
	// Register a device type with FMN in the slug and model.
	RegisterDeviceType(CaniDeviceType{
		Slug:         "test-proliant-dl325-gen11-fmn",
		Model:        "ProLiant DL325 Gen11 FMN",
		Manufacturer: "HPE",
		Type:         "blade",
	})
	defer func() {
		delete(allDeviceTypes, "test-proliant-dl325-gen11-fmn")
	}()

	device := UnclassifiedDevice{
		Name:       "fmn",
		DeviceType: "compute",
	}
	results := SuggestTypes(device, 8)

	found := false
	for _, r := range results {
		if r.Slug == "test-proliant-dl325-gen11-fmn" {
			found = true
			if r.Score < 30 {
				t.Errorf("fmn slug score = %d, want >= 30", r.Score)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected test-proliant-dl325-gen11-fmn in results, got %v", results)
	}
}

func TestNormalizeHardwareType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"mgmt_switch", "mgmt-switch"},
		{"hsn_switch", "hsn-switch"},
		{"compute", "compute"},
		{"cabinet-pdu", "cabinet-pdu"},
		{"", ""},
	}
	for _, tt := range tests {
		got := normalizeHardwareType(tt.input)
		if got != tt.want {
			t.Errorf("normalizeHardwareType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCollectQueriesDecomposition(t *testing.T) {
	device := UnclassifiedDevice{
		Name:       "dl360gen11",
		DeviceType: "compute",
	}
	queries := collectQueries(device)

	// Should contain the original name plus decomposed sub-tokens.
	wantPresent := map[string]bool{
		"dl360gen11": true,
		"dl360":      true,
		"gen11":      true,
		"compute":    true,
	}

	got := make(map[string]bool)
	for _, q := range queries {
		got[q] = true
	}

	for want := range wantPresent {
		if !got[want] {
			t.Errorf("collectQueries missing expected query %q, got %v", want, queries)
		}
	}
}

func TestRelatedHardwareTypes(t *testing.T) {
	tests := []struct {
		input string
		want  int // minimum number of related types
	}{
		{"compute", 2},
		{"blade", 1},
		{"node", 1},
		{"mgmt-switch", 1},
		{"switch", 1},
		{"chassis", 0},
	}
	for _, tt := range tests {
		got := relatedHardwareTypes(tt.input)
		if len(got) < tt.want {
			t.Errorf("relatedHardwareTypes(%q) returned %d types, want >= %d", tt.input, len(got), tt.want)
		}
	}
}

// ---------- hardwareTypeFallback ----------

func TestHardwareTypeFallbackReturns(t *testing.T) {
	// Register a device with matching HardwareType so the fallback has data.
	RegisterDeviceType(CaniDeviceType{
		Slug:         "test-hwfb-blade",
		Model:        "Test Blade",
		Manufacturer: "TestCo",
		Type:         TypeBlade,
	})
	defer func() {
		delete(allDeviceTypes, "test-hwfb-blade")
		delete(deviceTypesByPartNum, "")
	}()

	slugs := hardwareTypeFallback("blade", 10)
	if len(slugs) == 0 {
		t.Error("hardwareTypeFallback(\"blade\", 10) returned 0 slugs, want > 0")
	}
}

func TestHardwareTypeFallbackZeroMax(t *testing.T) {
	slugs := hardwareTypeFallback("blade", 0)
	if len(slugs) != 0 {
		t.Errorf("expected 0 slugs with max=0, got %d", len(slugs))
	}
}
