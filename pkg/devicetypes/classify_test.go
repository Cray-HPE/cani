package devicetypes

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
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

// ---------- FindUnclassifiedDevices ----------

// TestFindUnclassifiedDevices verifies FindUnclassifiedDevices returns only
// devices lacking both a Slug and a Model, sorted by name, and skips nil
// entries.
//
// Why it matters: the interactive classifier operates on devices that cannot be
// matched to a CaniType, so the scan must select exactly those (no slug and no
// model) and present them deterministically.
// Inputs: an inventory with one unclassified device ("zeta"), one classified by
// slug, one classified by model, and a nil map entry. Outputs: a single summary
// for "zeta".
// Data choice: covering both classification escape hatches (slug-set and
// model-set) plus a nil entry proves each skip branch; a lone qualifying device
// keeps the expected result unambiguous.
func TestFindUnclassifiedDevices(t *testing.T) {
	inv := NewInventory()
	uID := uuid.New()
	inv.Devices[uID] = &CaniDeviceType{ID: uID, Name: "zeta"}
	inv.Devices[uuid.New()] = &CaniDeviceType{Name: "has-slug", Slug: "x"}
	inv.Devices[uuid.New()] = &CaniDeviceType{Name: "has-model", Model: "m"}
	inv.Devices[uuid.New()] = nil

	got := FindUnclassifiedDevices(inv)
	if len(got) != 1 {
		t.Fatalf("expected 1 unclassified device, got %d: %+v", len(got), got)
	}
	if got[0].Name != "zeta" || got[0].ID != uID {
		t.Errorf("unclassified device = %+v, want name zeta id %v", got[0], uID)
	}
}

// ---------- ApplyDeviceType ----------

// TestApplyDeviceTypeFillsEmptyFields verifies ApplyDeviceType copies template
// fields onto a device only where the device's fields are empty.
//
// Why it matters: classification enriches a discovered device from the library
// without clobbering data the operator already provided, so non-empty fields
// must be preserved while empty ones are filled.
// Inputs: a registered "srv-x" template (model, manufacturer, U-height 2) and a
// device with a pre-set Model "Custom". Outputs: nil error; Slug set from the
// template, Model preserved as "Custom", Manufacturer and UHeight filled.
// Data choice: pre-setting only Model isolates the "keep existing" branch while
// the untouched Manufacturer/UHeight exercise the "fill empty" branches.
func TestApplyDeviceTypeFillsEmptyFields(t *testing.T) {
	resetRegistries()
	RegisterDeviceType(CaniDeviceType{
		Slug:         "srv-x",
		Model:        "ServerX",
		Manufacturer: "Acme",
		UHeight:      2,
	})

	dev := &CaniDeviceType{Name: "node-1", Model: "Custom"}
	if err := ApplyDeviceType(dev, "srv-x"); err != nil {
		t.Fatalf("ApplyDeviceType() unexpected error: %v", err)
	}
	if dev.Slug != "srv-x" {
		t.Errorf("Slug = %q, want %q", dev.Slug, "srv-x")
	}
	if dev.Model != "Custom" {
		t.Errorf("Model = %q, want preserved %q", dev.Model, "Custom")
	}
	if dev.Manufacturer != "Acme" {
		t.Errorf("Manufacturer = %q, want filled %q", dev.Manufacturer, "Acme")
	}
	if dev.UHeight != 2 {
		t.Errorf("UHeight = %d, want filled 2", dev.UHeight)
	}
}

// TestApplyDeviceTypeErrors verifies ApplyDeviceType rejects a nil device and
// an unknown slug.
//
// Why it matters: classification must fail loudly on bad inputs rather than
// silently producing a half-populated device.
// Inputs: a nil device pointer; a valid device with slug "ghost" against an
// empty registry. Outputs: ErrNilDevice for the first; a non-nil error
// mentioning the slug for the second.
// Data choice: errors.Is pins the nil case to the sentinel ErrNilDevice, and an
// empty registry guarantees the slug miss.
func TestApplyDeviceTypeErrors(t *testing.T) {
	resetRegistries()
	if err := ApplyDeviceType(nil, "any"); !errors.Is(err, ErrNilDevice) {
		t.Errorf("ApplyDeviceType(nil) error = %v, want ErrNilDevice", err)
	}
	err := ApplyDeviceType(&CaniDeviceType{Name: "d"}, "ghost")
	if err == nil {
		t.Fatal("ApplyDeviceType(unknown slug) should return an error")
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Errorf("error should mention the slug, got: %v", err)
	}
}
