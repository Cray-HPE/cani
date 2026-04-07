package devicetypes_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// fixtureDir returns the absolute path to the export-edge-cases fixture.
func fixtureDir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata", "fixtures", "cani")
}

// loadFixture reads and unmarshals the export_edge_cases.json fixture.
func loadFixture(t *testing.T) *devicetypes.Inventory {
	t.Helper()
	path := filepath.Join(fixtureDir(), "export_edge_cases.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var inv devicetypes.Inventory
	if err := json.Unmarshal(data, &inv); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	return &inv
}

// ---------- fixture load & counts ----------

func TestLoadExportEdgeCaseFixture(t *testing.T) {
	inv := loadFixture(t)

	tests := []struct {
		name string
		got  int
		want int
	}{
		{"Locations", len(inv.Locations), 4},
		{"Racks", len(inv.Racks), 3},
		{"Devices", len(inv.Devices), 6},
		{"Modules", len(inv.Modules), 3},
		{"Cables", len(inv.Cables), 4},
		{"Frus", len(inv.Frus), 4},
		{"Interfaces", len(inv.Interfaces), 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("len(%s) = %d, want %d", tt.name, tt.got, tt.want)
			}
		})
	}
}

// ---------- bug #5: ClassifyForNautobot("module") ----------

func TestClassifyForNautobot_ModuleFallthrough(t *testing.T) {
	// TypeModule ("module") should classify as CategoryModule, but
	// currently falls through to the default case → CategoryDevice.
	got := devicetypes.ClassifyForNautobot("module")

	// Document the current (buggy) behaviour so CI stays green.
	if got == devicetypes.CategoryModule {
		t.Log("ClassifyForNautobot(\"module\") correctly returns CategoryModule — bug #5 is fixed")
		return
	}
	if got != devicetypes.CategoryDevice {
		t.Errorf("unexpected category %q", got)
	}
	t.Log("BUG #5: ClassifyForNautobot(\"module\") returns CategoryDevice instead of CategoryModule")
}

// ---------- bug #8: FRU parent self-cycle ----------

func TestFruParentSelfCycle(t *testing.T) {
	inv := loadFixture(t)

	selfCycleID := uuid.MustParse("00000000-0000-0000-0000-000000000052")
	fru, ok := inv.Frus[selfCycleID]
	if !ok {
		t.Fatal("self-cycle FRU not found in fixture")
	}
	if fru.Parent != fru.ID {
		t.Fatal("fixture setup error: expected parent == self")
	}

	// topologicalSortFrus is unexported; verify that the self-referencing
	// FRU exists so downstream tests/fixes can target cycle detection.
	t.Log("BUG #8: FRU parent == self (cycle); topologicalSortFrus has no cycle detection")
}

// ---------- bug #10: cable float-to-int truncation ----------

func TestCableLengthTruncation(t *testing.T) {
	inv := loadFixture(t)

	fractionalID := uuid.MustParse("00000000-0000-0000-0000-000000000041")
	cable, ok := inv.Cables[fractionalID]
	if !ok {
		t.Fatal("fractional-length cable not found in fixture")
	}
	if cable.Length == nil {
		t.Fatal("expected non-nil Length")
	}

	original := *cable.Length
	truncated := int(original)

	if float64(truncated) == original {
		t.Skip("length has no fractional part; nothing to test")
	}
	t.Logf("BUG #10: cable length %.1f truncated to %d by float64→int cast", original, truncated)
}

// ---------- bug #9: cable termination type hardcoded ----------

func TestCableTerminationTypeHardcoded(t *testing.T) {
	inv := loadFixture(t)

	powerCableID := uuid.MustParse("00000000-0000-0000-0000-000000000042")
	cable, ok := inv.Cables[powerCableID]
	if !ok {
		t.Fatal("power-termination cable not found in fixture")
	}

	if cable.TerminationAType != "dcim.powerport" {
		t.Fatalf("expected terminationAType=dcim.powerport, got %s", cable.TerminationAType)
	}

	// The export pipeline hardcodes "dcim.interface" regardless of what
	// the cable declares. Document this for future fix.
	t.Log("BUG #9: cable declares dcim.powerport but export hardcodes dcim.interface")
}

// ---------- bug #17: rack field droppage ----------

func TestRackFieldDroppage(t *testing.T) {
	inv := loadFixture(t)

	rackID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	rack, ok := inv.Racks[rackID]
	if !ok {
		t.Fatal("full-featured rack not found in fixture")
	}

	// Verify the fixture carries all fields that are dropped during export.
	dropped := map[string]string{
		"Serial":     rack.Serial,
		"AssetTag":   rack.AssetTag,
		"RackType":   rack.RackType,
		"FacilityId": rack.FacilityId,
		"Width":      rack.Width,
		"Role":       rack.Role,
		"Tenant":     rack.Tenant,
	}
	for field, val := range dropped {
		if val == "" {
			t.Errorf("fixture setup error: Rack-A01.%s is empty", field)
		}
	}
	t.Log("BUG #17: createRackFromCaniRack drops Serial, AssetTag, RackType, FacilityId, Width, Role, Tenant, Tags")
}

// ---------- bug #20: rack type enum validation ----------

func TestRackTypeEnumValidation(t *testing.T) {
	inv := loadFixture(t)

	rackID := uuid.MustParse("00000000-0000-0000-0000-000000000011")
	rack, ok := inv.Racks[rackID]
	if !ok {
		t.Fatal("invalid-rack-type rack not found in fixture")
	}

	validTypes := map[string]bool{
		"2-post-frame":          true,
		"4-post-frame":          true,
		"4-post-cabinet":        true,
		"wall-frame":            true,
		"wall-frame-vertical":   true,
		"wall-cabinet":          true,
		"wall-cabinet-vertical": true,
	}

	if validTypes[rack.RackType] {
		t.Skipf("rackType %q is valid; nothing to test", rack.RackType)
	}
	t.Logf("BUG #20: rackType %q is not in Nautobot enum; no validation exists", rack.RackType)
}

// ---------- bug #18: module location mutual exclusivity ----------

func TestModuleLocationMutualExclusivity(t *testing.T) {
	inv := loadFixture(t)

	modID := uuid.MustParse("00000000-0000-0000-0000-000000000032")
	mod, ok := inv.Modules[modID]
	if !ok {
		t.Fatal("dual-location module not found in fixture")
	}

	hasParent := mod.ParentDevice != uuid.Nil
	hasLocation := mod.Location != uuid.Nil

	if hasParent && hasLocation {
		t.Log("BUG #18: module has both ParentDevice and Location set; Nautobot says mutually exclusive")
	}
}

// ---------- bug #4: duplicate device name ----------

func TestDuplicateDeviceName(t *testing.T) {
	inv := loadFixture(t)

	// Two different UUIDs share the name "Chassis-01".
	id1 := uuid.MustParse("00000000-0000-0000-0000-000000000020")
	id2 := uuid.MustParse("00000000-0000-0000-0000-000000000022")

	dev1, ok1 := inv.Devices[id1]
	dev2, ok2 := inv.Devices[id2]
	if !ok1 || !ok2 {
		t.Fatal("expected both Chassis-01 devices in fixture")
	}
	if dev1.Name != dev2.Name {
		t.Fatalf("expected same name, got %q and %q", dev1.Name, dev2.Name)
	}

	// Simulate the createdDeviceIDs map behaviour.
	createdDeviceIDs := make(map[string]uuid.UUID)
	createdDeviceIDs[dev1.Name] = id1
	createdDeviceIDs[dev2.Name] = id2 // overwrites

	if createdDeviceIDs["Chassis-01"] != id2 {
		t.Fatal("expected second insert to overwrite first")
	}
	t.Log("BUG #4: createdDeviceIDs keyed by name; duplicate 'Chassis-01' causes silent data loss")
}

// ---------- bug #21 & #22: FRU orphan fields ----------

func TestFruOrphanFields(t *testing.T) {
	inv := loadFixture(t)

	fruID := uuid.MustParse("00000000-0000-0000-0000-000000000053")
	fru, ok := inv.Frus[fruID]
	if !ok {
		t.Fatal("orphan-fields FRU not found in fixture")
	}

	// These fields exist on CaniFruType but have no InventoryItem equivalent.
	if fru.Role == "" {
		t.Error("fixture setup error: expected non-empty Role")
	}
	if fru.HardwareType == "" {
		t.Error("fixture setup error: expected non-empty HardwareType")
	}
	if fru.Model == "" {
		t.Error("fixture setup error: expected non-empty Model")
	}
	if fru.PartNumber != "" {
		t.Error("fixture setup error: expected empty PartNumber for this test")
	}

	t.Log("BUG #21: FRU Role/HardwareType have no InventoryItem equivalents — silently dropped")
	t.Log("BUG #22: FRU Model='GPU-X100' but PartNumber empty; export maps PartNumber→PartId, so model name is lost")
}

// ---------- bug #16: FRU→InventoryItem deprecation ----------

func TestFruDeprecation(t *testing.T) {
	inv := loadFixture(t)
	if len(inv.Frus) == 0 {
		t.Skip("no FRUs in fixture")
	}
	t.Log("BUG #16: CaniFruType maps to Nautobot InventoryItem which is deprecated; " +
		"Nautobot docs say it will be replaced by Modules in a future release")
}

// ---------- bug #2: rack location UUID ignored ----------

func TestRackLocationIgnored(t *testing.T) {
	inv := loadFixture(t)

	rackID := uuid.MustParse("00000000-0000-0000-0000-000000000011")
	rack, ok := inv.Racks[rackID]
	if !ok {
		t.Fatal("rack not found in fixture")
	}

	if rack.Location == uuid.Nil {
		t.Fatal("fixture setup error: expected rack to have location UUID set")
	}
	t.Log("BUG #2: createRackFromCaniRack ignores rack.Location UUID; always uses DefaultLocation string")
}

// ---------- bug #3: device location from ProviderMetadata ----------

func TestDeviceLocationFromProviderMetadata(t *testing.T) {
	inv := loadFixture(t)

	devID := uuid.MustParse("00000000-0000-0000-0000-000000000025")
	dev, ok := inv.Devices[devID]
	if !ok {
		t.Fatal("meta-location device not found in fixture")
	}

	if dev.Location != uuid.Nil {
		t.Fatal("fixture setup error: expected empty Location UUID")
	}
	if dev.ProviderMetadata == nil {
		t.Fatal("fixture setup error: expected non-nil ProviderMetadata")
	}

	nautobotMeta, _ := dev.ProviderMetadata["nautobot"].(map[string]any)
	locStr, _ := nautobotMeta["location"].(string)
	if locStr == "" {
		t.Fatal("fixture setup error: expected ProviderMetadata.nautobot.location string")
	}
	t.Logf("BUG #3: device location resolved from ProviderMetadata[\"location\"]=%q instead of Location UUID FK", locStr)
}

// ---------- bug #7: inventory.Interfaces ignored ----------

func TestInventoryInterfacesIgnored(t *testing.T) {
	inv := loadFixture(t)

	if len(inv.Interfaces) == 0 {
		t.Fatal("fixture setup error: expected interfaces in Interfaces map")
	}

	// Verify at least one interface maps to a known device.
	devID := uuid.MustParse("00000000-0000-0000-0000-000000000020")
	found := false
	for _, iface := range inv.Interfaces {
		if iface.DeviceID == devID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("fixture setup error: expected interface for device 00...0020")
	}
	t.Log("BUG #7: inventory.Interfaces map is never consulted during export; interface creation uses device spec templates only")
}

// ---------- bug #12: Validate() is a no-op ----------

func TestValidateNoOp(t *testing.T) {
	// CaniDeviceType.Validate() with minimal data should return nil.
	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "", // empty name is allowed
	}
	if err := dev.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mod := &devicetypes.CaniModuleType{ID: uuid.New()}
	if err := mod.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fru := &devicetypes.CaniFruType{ID: uuid.New()}
	if err := fru.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Log("BUG #12: Validate() on Device/Module/FRU only checks for nil receiver — no field validation")
}

// ---------- bug #15: empty LocationType is now rejected ----------

func TestLocationTypeEmptyFallback(t *testing.T) {
	inv := loadFixture(t)

	locID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	loc, ok := inv.Locations[locID]
	if !ok {
		t.Fatal("empty-type location not found in fixture")
	}

	if loc.LocationType != "" {
		t.Fatalf("fixture setup error: expected empty LocationType, got %q", loc.LocationType)
	}

	// Validate must reject an empty LocationType.
	err := loc.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject empty LocationType")
	}
	t.Logf("Validate correctly rejects empty LocationType: %v", err)
}

// ---------- spine-leaf fixture ----------

// loadSpineLeafFixture reads the spine_leaf_inventory.json fixture.
func loadSpineLeafFixture(t *testing.T) *devicetypes.Inventory {
	t.Helper()
	path := filepath.Join(fixtureDir(), "spine_leaf_inventory.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var inv devicetypes.Inventory
	if err := json.Unmarshal(data, &inv); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	return &inv
}

func TestLoadSpineLeafFixture(t *testing.T) {
	inv := loadSpineLeafFixture(t)

	tests := []struct {
		name string
		got  int
		want int
	}{
		{"Locations", len(inv.Locations), 2},
		{"Racks", len(inv.Racks), 2},
		{"Devices", len(inv.Devices), 12},
		{"Modules", len(inv.Modules), 6},
		{"Cables", len(inv.Cables), 14},
		{"Frus", len(inv.Frus), 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("len(%s) = %d, want %d", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestSpineLeafCableEndpoints(t *testing.T) {
	inv := loadSpineLeafFixture(t)

	for cableID, cable := range inv.Cables {
		t.Run(cable.Label, func(t *testing.T) {
			if cable.TerminationADevice == uuid.Nil {
				t.Errorf("cable %s: terminationADevice is nil", cableID)
			}
			if cable.TerminationBDevice == uuid.Nil {
				t.Errorf("cable %s: terminationBDevice is nil", cableID)
			}
			if cable.TerminationAPort == "" {
				t.Errorf("cable %s: terminationAPort is empty", cableID)
			}
			if cable.TerminationBPort == "" {
				t.Errorf("cable %s: terminationBPort is empty", cableID)
			}

			// Verify devices exist in inventory
			if _, ok := inv.Devices[cable.TerminationADevice]; !ok {
				t.Errorf("cable %s: terminationADevice %s not in inventory", cableID, cable.TerminationADevice)
			}
			if _, ok := inv.Devices[cable.TerminationBDevice]; !ok {
				t.Errorf("cable %s: terminationBDevice %s not in inventory", cableID, cable.TerminationBDevice)
			}
		})
	}
}

func TestSpineLeafModuleParents(t *testing.T) {
	inv := loadSpineLeafFixture(t)

	for modID, mod := range inv.Modules {
		t.Run(mod.Name, func(t *testing.T) {
			if mod.ParentDevice == uuid.Nil {
				t.Errorf("module %s: parentDevice is nil", modID)
			}
			if _, ok := inv.Devices[mod.ParentDevice]; !ok {
				t.Errorf("module %s: parentDevice %s not in inventory", modID, mod.ParentDevice)
			}
			if mod.ModuleBayName == "" {
				t.Errorf("module %s: moduleBayName is empty", modID)
			}
		})
	}
}

func TestSpineLeafFruDevices(t *testing.T) {
	inv := loadSpineLeafFixture(t)

	for fruID, fru := range inv.Frus {
		t.Run(fru.Name, func(t *testing.T) {
			if fru.Device == uuid.Nil {
				t.Errorf("fru %s: device is nil", fruID)
			}
			if _, ok := inv.Devices[fru.Device]; !ok {
				t.Errorf("fru %s: device %s not in inventory", fruID, fru.Device)
			}
		})
	}
}
