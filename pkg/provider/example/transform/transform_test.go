package transform

import (
	"testing"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/example/commands"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

// --- Test helper ---

type fakeRecordProvider struct{ records []import_.CsvRecord }

func (f *fakeRecordProvider) GetRecords() []import_.CsvRecord { return f.records }

// setupTransformTest saves and restores providerGetter and config.Cfg, then
// sets a fakeRecordProvider with the given records.
func setupTransformTest(t *testing.T, records []import_.CsvRecord) {
	t.Helper()
	oldGetter := providerGetter
	oldCfg := config.Cfg
	t.Cleanup(func() {
		providerGetter = oldGetter
		config.Cfg = oldCfg
		resetRackPositionStates()
	})
	config.Cfg = &config.Config{}
	SetProviderGetter(func() interface{ GetRecords() []import_.CsvRecord } {
		return &fakeRecordProvider{records: records}
	})
}

// --- Transform entry point tests ---

func TestSetProviderGetter(t *testing.T) {
	old := providerGetter
	t.Cleanup(func() { providerGetter = old })

	called := false
	SetProviderGetter(func() interface{ GetRecords() []import_.CsvRecord } {
		called = true
		return &fakeRecordProvider{}
	})

	if providerGetter == nil {
		t.Fatal("providerGetter should not be nil after SetProviderGetter")
	}
	providerGetter()
	if !called {
		t.Error("expected providerGetter to be called")
	}
}

func TestTransform(t *testing.T) {
	t.Run("nil provider getter", func(t *testing.T) {
		old := providerGetter
		t.Cleanup(func() { providerGetter = old })
		providerGetter = nil

		if _, err := Transform(devicetypes.Inventory{}); err == nil {
			t.Fatal("expected error when providerGetter is nil")
		}
	})

	t.Run("empty records", func(t *testing.T) {
		setupTransformTest(t, nil)
		result, err := Transform(devicetypes.Inventory{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Racks) != 0 || len(result.Devices) != 0 || len(result.Cables) != 0 {
			t.Error("expected empty result for empty records")
		}
	})

	t.Run("with records", func(t *testing.T) {
		setupTransformTest(t, []import_.CsvRecord{
			{PartNumber: "FAKE-RACK-PN", Description: "48U Rack Cabinet", Quantity: 1, ConfigGroup: "0100"},
			{PartNumber: "FAKE-SRV-PN", Description: "ProLiant DL380 Server", Quantity: 2, ConfigGroup: "0300"},
		})
		result, err := Transform(devicetypes.Inventory{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Racks) != 1 {
			t.Errorf("expected 1 rack, got %d", len(result.Racks))
		}
		if len(result.Devices) != 2 {
			t.Errorf("expected 2 devices, got %d", len(result.Devices))
		}
	})

	t.Run("classify error", func(t *testing.T) {
		setupTransformTest(t, []import_.CsvRecord{
			{PartNumber: "FAKE-UNKNOWN", Description: "", Quantity: 1},
		})
		if _, err := Transform(devicetypes.Inventory{}); err == nil {
			t.Fatal("expected error for unclassifiable record")
		}
	})
}

// --- Classification tests ---

func TestClassifyRecords(t *testing.T) {
	tests := []struct {
		name        string
		records     []import_.CsvRecord
		wantRacks   int
		wantDevices int
		wantCables  int
		wantErr     bool
	}{
		{"empty records", nil, 0, 0, 0, false},
		{"rack by description", []import_.CsvRecord{
			{PartNumber: "X", Description: "HPE 48U 800mmx1200mm G2 Enterprise Shock Rack", Quantity: 1},
		}, 1, 0, 0, false},
		{"cable by explicit endpoints", []import_.CsvRecord{
			{Description: "link", Quantity: 1, SourceDevice: "sw1", DestDevice: "srv1"},
		}, 0, 0, 1, false},
		{"cable by description pattern", []import_.CsvRecord{
			{PartNumber: "X", Description: "HPE Cat6 RJ45 M/M 2m", Quantity: 1},
		}, 0, 0, 1, false},
		{"switch by description", []import_.CsvRecord{
			{PartNumber: "X", Description: "HPE Aruba Networking 8360-48Y6C", Quantity: 1},
		}, 0, 1, 0, false},
		{"node by description", []import_.CsvRecord{
			{PartNumber: "X", Description: "HPE ProLiant DL380 Gen11", Quantity: 1},
		}, 0, 1, 0, false},
		{"unclassifiable returns error", []import_.CsvRecord{
			{PartNumber: "X", Description: "HPE 64GB DDR5 Memory Kit", Quantity: 1},
		}, 0, 0, 0, true},
		{"mixed record types", []import_.CsvRecord{
			{PartNumber: "R", Description: "HPE 48U Rack", Quantity: 1},
			{PartNumber: "S", Description: "HPE Aruba Switch", Quantity: 1},
			{PartNumber: "C", Description: "HPE Cat6 Cable 2m", Quantity: 1},
		}, 1, 1, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := classifyRecords(tt.records)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got.racks) != tt.wantRacks {
				t.Errorf("racks = %d, want %d", len(got.racks), tt.wantRacks)
			}
			if len(got.devices) != tt.wantDevices {
				t.Errorf("devices = %d, want %d", len(got.devices), tt.wantDevices)
			}
			if len(got.cables) != tt.wantCables {
				t.Errorf("cables = %d, want %d", len(got.cables), tt.wantCables)
			}
		})
	}
}

// --- Item creation tests ---

func TestCreateItemsFromRecord(t *testing.T) {
	t.Run("creates rack items", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
			Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		}
		racksByGroup := make(map[string][]uuid.UUID)
		devicesByGroup := make(map[string][]uuid.UUID)

		rec := import_.CsvRecord{PartNumber: "TEST-RACK", Description: "Test Rack", Quantity: 2, ConfigGroup: "0100"}
		result := createItemsFromRecord(inv, rec, "rack", racksByGroup, devicesByGroup)

		if len(result.Racks) != 2 {
			t.Errorf("expected 2 racks, got %d", len(result.Racks))
		}
		if len(result.Devices) != 0 {
			t.Errorf("expected 0 devices, got %d", len(result.Devices))
		}
		if len(inv.Racks) != 2 {
			t.Errorf("expected 2 racks in inventory, got %d", len(inv.Racks))
		}
		if len(racksByGroup["0100"]) != 2 {
			t.Errorf("expected 2 racks in group 0100, got %d", len(racksByGroup["0100"]))
		}
	})

	t.Run("creates device items", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
			Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		}
		racksByGroup := make(map[string][]uuid.UUID)
		devicesByGroup := make(map[string][]uuid.UUID)

		rec := import_.CsvRecord{PartNumber: "TEST-NODE", Description: "Test Server", Quantity: 3, ConfigGroup: "0200"}
		result := createItemsFromRecord(inv, rec, "node", racksByGroup, devicesByGroup)

		if len(result.Devices) != 3 {
			t.Errorf("expected 3 devices, got %d", len(result.Devices))
		}
		if len(result.Racks) != 0 {
			t.Errorf("expected 0 racks, got %d", len(result.Racks))
		}
		if len(inv.Devices) != 3 {
			t.Errorf("expected 3 devices in inventory, got %d", len(inv.Devices))
		}
		if len(devicesByGroup["0200"]) != 3 {
			t.Errorf("expected 3 devices in group 0200, got %d", len(devicesByGroup["0200"]))
		}
	})

	t.Run("no config group skips grouping", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
			Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		}
		racksByGroup := make(map[string][]uuid.UUID)
		devicesByGroup := make(map[string][]uuid.UUID)

		rec := import_.CsvRecord{PartNumber: "TEST", Description: "Standalone Device", Quantity: 1}
		createItemsFromRecord(inv, rec, "node", racksByGroup, devicesByGroup)

		if len(devicesByGroup) != 0 {
			t.Errorf("expected no config group entries, got %d", len(devicesByGroup))
		}
	})
}

func TestCreateRackWithoutLibraryMatch(t *testing.T) {
	inv := &devicetypes.Inventory{Racks: make(map[uuid.UUID]*devicetypes.CaniRackType)}
	racksByGroup := make(map[string][]uuid.UUID)
	id := uuid.New()

	rec := import_.CsvRecord{PartNumber: "FAKE-RACK-PN", Description: "Custom 42U Rack", ConfigGroup: "0100"}
	rack := createRack(inv, id, "Custom 42U Rack", rec, racksByGroup)

	if rack.ID != id {
		t.Errorf("rack ID = %v, want %v", rack.ID, id)
	}
	if rack.UHeight != 48 {
		t.Errorf("default UHeight = %d, want 48", rack.UHeight)
	}
	if _, ok := inv.Racks[id]; !ok {
		t.Error("rack should be in inventory")
	}
	if len(racksByGroup["0100"]) != 1 {
		t.Error("rack should be tracked in racksByGroup")
	}
}

func TestBuildDeviceFromRecord(t *testing.T) {
	id := uuid.New()
	rec := import_.CsvRecord{PartNumber: "TEST-FAKE-PN", Description: "HPE ProLiant DL380", ConfigGroup: "0200"}
	device := buildDeviceFromRecord(id, "DL380-001", rec, "node")

	if device.ID != id {
		t.Errorf("ID = %v, want %v", device.ID, id)
	}
	if device.Name != "DL380-001" {
		t.Errorf("Name = %q, want %q", device.Name, "DL380-001")
	}
	if device.PartNumber != "TEST-FAKE-PN" {
		t.Errorf("PartNumber = %q, want %q", device.PartNumber, "TEST-FAKE-PN")
	}
	if device.Type != devicetypes.Type("node") {
		t.Errorf("Type = %q, want %q", device.Type, "node")
	}
}

func TestPopulateFromDeviceType(t *testing.T) {
	t.Run("copies all fields", func(t *testing.T) {
		device := &devicetypes.CaniDeviceType{Name: "original", Slug: "old-slug"}
		dt := &devicetypes.CaniDeviceType{
			Slug: "new-slug", Manufacturer: "HPE", Model: "DL380 Gen11",
			Description: "Rack server",
			Type:        devicetypes.Type("node"),
			UHeight:     2,
			IsFullDepth: true,
			Weight:      15.5,
			WeightUnit:  "kg",
			Comments:    "from library",
			Interfaces:  []devicetypes.InterfaceSpec{{Name: "eth0"}},
			ConsolePorts: []devicetypes.ConsolePortSpec{{
				Name: "console0",
			}},
			PowerPorts: []devicetypes.PowerPortSpec{{
				Name: "PSU1",
			}},
			ModuleBays: []devicetypes.ModuleBaySpec{{
				Name: "PCIe1",
			}},
			DeviceBays: []devicetypes.DeviceBaySpec{{
				Name: "bay1",
			}},
			Identifications: []devicetypes.Identification{{
				Manufacturer: "HPE",
				Model:        "DL380 Gen11",
			}},
		}
		populateFromDeviceType(device, dt)

		if device.Slug != "new-slug" {
			t.Errorf("Slug = %q, want %q", device.Slug, "new-slug")
		}
		if device.Manufacturer != "HPE" {
			t.Errorf("Manufacturer = %q, want %q", device.Manufacturer, "HPE")
		}
		if device.Model != "DL380 Gen11" {
			t.Errorf("Model = %q, want %q", device.Model, "DL380 Gen11")
		}
		if device.Type != devicetypes.Type("node") {
			t.Errorf("Type = %q, want %q", device.Type, "node")
		}
		if device.Description != "Rack server" {
			t.Errorf("Description = %q, want %q", device.Description, "Rack server")
		}
		if device.UHeight != 2 {
			t.Errorf("UHeight = %d, want 2", device.UHeight)
		}
		if !device.IsFullDepth {
			t.Fatal("expected IsFullDepth to be true")
		}
		if device.Weight != 15.5 {
			t.Errorf("Weight = %v, want 15.5", device.Weight)
		}
		if device.WeightUnit != "kg" {
			t.Errorf("WeightUnit = %q, want %q", device.WeightUnit, "kg")
		}
		if device.Comments != "from library" {
			t.Errorf("Comments = %q, want %q", device.Comments, "from library")
		}
		if len(device.Interfaces) != 1 {
			t.Errorf("Interfaces len = %d, want 1", len(device.Interfaces))
		}
		if len(device.ConsolePorts) != 1 {
			t.Errorf("ConsolePorts len = %d, want 1", len(device.ConsolePorts))
		}
		if len(device.PowerPorts) != 1 {
			t.Errorf("PowerPorts len = %d, want 1", len(device.PowerPorts))
		}
		if len(device.ModuleBays) != 1 {
			t.Errorf("ModuleBays len = %d, want 1", len(device.ModuleBays))
		}
		if len(device.DeviceBays) != 1 {
			t.Errorf("DeviceBays len = %d, want 1", len(device.DeviceBays))
		}
		if len(device.Identifications) != 1 {
			t.Errorf("Identifications len = %d, want 1", len(device.Identifications))
		}
	})

	t.Run("does not overwrite type when empty", func(t *testing.T) {
		device := &devicetypes.CaniDeviceType{Type: devicetypes.Type("switch")}
		dt := &devicetypes.CaniDeviceType{Slug: "new-slug", Type: ""}
		populateFromDeviceType(device, dt)

		if device.Type != devicetypes.Type("switch") {
			t.Errorf("Type = %q, want %q (should not overwrite with empty)", device.Type, "switch")
		}
	})
}

func TestGetDeviceUHeight(t *testing.T) {
	tests := []struct {
		name   string
		device *devicetypes.CaniDeviceType
	}{
		{"unknown part number and slug", &devicetypes.CaniDeviceType{PartNumber: "FAKE-UNKNOWN-PN", Slug: "fake-unknown-slug"}},
		{"empty part number and slug", &devicetypes.CaniDeviceType{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDeviceUHeight(tt.device); got != 0 {
				t.Errorf("expected 0 for unknown device, got %d", got)
			}
		})
	}
}

// --- Rack position and zone tests ---

func TestDeviceTypePriority(t *testing.T) {
	tests := []struct {
		hwType string
		want   int
	}{
		{"pdu", 0}, {"cdu", 0},
		{"node", 1}, {"blade", 1}, {"chassis", 1},
		{"switch", 2}, {"mgmt-switch", 2}, {"hsn-switch", 2},
		{"", 1},
	}
	for _, tt := range tests {
		t.Run(tt.hwType, func(t *testing.T) {
			if got := deviceTypePriority(tt.hwType); got != tt.want {
				t.Errorf("deviceTypePriority(%q) = %d, want %d", tt.hwType, got, tt.want)
			}
		})
	}
}

func TestSortDevicesByRackPriority(t *testing.T) {
	inv := &devicetypes.Inventory{Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType)}

	switchID, nodeID, pduID, bladeID, cduID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Devices[switchID] = &devicetypes.CaniDeviceType{ID: switchID, Type: devicetypes.Type("switch")}
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{ID: nodeID, Type: devicetypes.Type("node")}
	inv.Devices[pduID] = &devicetypes.CaniDeviceType{ID: pduID, Type: devicetypes.Type("pdu")}
	inv.Devices[bladeID] = &devicetypes.CaniDeviceType{ID: bladeID, Type: devicetypes.Type("blade")}
	inv.Devices[cduID] = &devicetypes.CaniDeviceType{ID: cduID, Type: devicetypes.Type("cdu")}

	ids := []uuid.UUID{switchID, nodeID, pduID, bladeID, cduID}
	sortDevicesByRackPriority(inv, ids)

	// Verify ordering: PDUs/CDUs first, then nodes/blades, then switches
	for i := 1; i < len(ids); i++ {
		pri := deviceTypePriority(string(inv.Devices[ids[i]].Type))
		prevPri := deviceTypePriority(string(inv.Devices[ids[i-1]].Type))
		if pri < prevPri {
			t.Errorf("device at index %d (%s, priority %d) sorted before index %d (%s, priority %d)",
				i, inv.Devices[ids[i]].Type, pri, i-1, inv.Devices[ids[i-1]].Type, prevPri)
		}
	}

	for _, id := range ids[:2] {
		if deviceTypePriority(string(inv.Devices[id].Type)) != 0 {
			t.Errorf("expected pdu/cdu in first two positions, got %q", inv.Devices[id].Type)
		}
	}
	if deviceTypePriority(string(inv.Devices[ids[len(ids)-1]].Type)) != 2 {
		t.Errorf("expected switch in last position, got %q", inv.Devices[ids[len(ids)-1]].Type)
	}
}

func TestGroupDevicesByZone(t *testing.T) {
	inv := &devicetypes.Inventory{Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType)}

	pduID, cduID, nodeID, switchID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Devices[pduID] = &devicetypes.CaniDeviceType{ID: pduID, Type: devicetypes.Type("pdu")}
	inv.Devices[cduID] = &devicetypes.CaniDeviceType{ID: cduID, Type: devicetypes.Type("cdu")}
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{ID: nodeID, Type: devicetypes.Type("node")}
	inv.Devices[switchID] = &devicetypes.CaniDeviceType{ID: switchID, Type: devicetypes.Type("switch")}

	bottom, middle, top := groupDevicesByZone(inv, []uuid.UUID{pduID, cduID, nodeID, switchID})

	if len(bottom) != 2 {
		t.Errorf("expected 2 bottom devices, got %d", len(bottom))
	}
	if len(middle) != 1 {
		t.Errorf("expected 1 middle device, got %d", len(middle))
	}
	if len(top) != 1 {
		t.Errorf("expected 1 top device, got %d", len(top))
	}
}

func TestRackPositionOrdering(t *testing.T) {
	resetRackPositionStates()
	inv := &devicetypes.Inventory{
		Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
	}

	rackID := uuid.New()
	inv.Racks[rackID] = &devicetypes.CaniRackType{ID: rackID, Name: "Rack-001", UHeight: 48, Devices: []uuid.UUID{}}

	switchID, nodeID, pduID := uuid.New(), uuid.New(), uuid.New()
	inv.Devices[switchID] = &devicetypes.CaniDeviceType{ID: switchID, Type: devicetypes.Type("switch"), Slug: "switch-1u"}
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{ID: nodeID, Type: devicetypes.Type("node"), Slug: "node-1u"}
	inv.Devices[pduID] = &devicetypes.CaniDeviceType{ID: pduID, Type: devicetypes.Type("pdu"), Slug: "pdu-1u"}

	assignConfigGroupParenting(inv,
		map[string][]uuid.UUID{"0100": {rackID}},
		map[string][]uuid.UUID{"0200": {switchID, nodeID, pduID}})

	if pos := inv.Devices[pduID].RackPosition; pos != 1 {
		t.Errorf("PDU position = %d, want 1 (bottom)", pos)
	}
	if pos := inv.Devices[switchID].RackPosition; pos != 48 {
		t.Errorf("switch position = %d, want 48 (top)", pos)
	}
	if pos := inv.Devices[nodeID].RackPosition; pos != 47 {
		t.Errorf("node position = %d, want 47 (below switch)", pos)
	}
}

func TestRackZonesFillCorrectly(t *testing.T) {
	resetRackPositionStates()
	inv := &devicetypes.Inventory{
		Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
	}

	rackID := uuid.New()
	inv.Racks[rackID] = &devicetypes.CaniRackType{ID: rackID, Name: "Rack-001", UHeight: 10, Devices: []uuid.UUID{}}

	pdu1, pdu2, sw1, sw2, node1 := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Devices[pdu1] = &devicetypes.CaniDeviceType{ID: pdu1, Type: devicetypes.Type("pdu")}
	inv.Devices[pdu2] = &devicetypes.CaniDeviceType{ID: pdu2, Type: devicetypes.Type("cdu")}
	inv.Devices[sw1] = &devicetypes.CaniDeviceType{ID: sw1, Type: devicetypes.Type("switch")}
	inv.Devices[sw2] = &devicetypes.CaniDeviceType{ID: sw2, Type: devicetypes.Type("mgmt-switch")}
	inv.Devices[node1] = &devicetypes.CaniDeviceType{ID: node1, Type: devicetypes.Type("node")}

	assignConfigGroupParenting(inv,
		map[string][]uuid.UUID{"0100": {rackID}},
		map[string][]uuid.UUID{"0200": {pdu1, pdu2, sw1, sw2, node1}})

	checks := []struct {
		name string
		id   uuid.UUID
		want int
	}{
		{"pdu1", pdu1, 1}, {"pdu2", pdu2, 2},
		{"node", node1, 8},
		{"sw1", sw1, 10}, {"sw2", sw2, 9},
	}
	for _, c := range checks {
		if pos := inv.Devices[c.id].RackPosition; pos != c.want {
			t.Errorf("%s position = %d, want %d", c.name, pos, c.want)
		}
	}
}

func TestLinkDeviceToRack(t *testing.T) {
	t.Run("missing device does not panic", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
			Racks:   map[uuid.UUID]*devicetypes.CaniRackType{},
		}
		linkDeviceToRack(inv, uuid.New(), uuid.New(), zoneMiddle)
	})

	t.Run("missing rack sets parent without panic", func(t *testing.T) {
		devID := uuid.New()
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{devID: {ID: devID, Name: "srv"}},
			Racks:   map[uuid.UUID]*devicetypes.CaniRackType{},
		}
		linkDeviceToRack(inv, devID, uuid.New(), zoneMiddle)
	})
}

// --- Config group and parenting helpers ---

func TestFindParentRackIDs(t *testing.T) {
	rack1, rack2, device1 := uuid.New(), uuid.New(), uuid.New()
	tests := []struct {
		name         string
		racksByGroup map[string][]uuid.UUID
		wantLen      int
	}{
		{"empty map", map[string][]uuid.UUID{}, 0},
		{"racks in 01XX group", map[string][]uuid.UUID{"0100": {rack1, rack2}}, 2},
		{"racks in non-01XX group only", map[string][]uuid.UUID{"0200": {device1}}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findParentRackIDs(tt.racksByGroup); len(got) != tt.wantLen {
				t.Errorf("findParentRackIDs() returned %d IDs, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestShouldLinkToRacks(t *testing.T) {
	tests := []struct {
		group string
		want  bool
	}{
		{"0100", false}, {"0200", true}, {"0300", true},
		{"0900", true}, {"1", false}, {"", false},
	}
	for _, tt := range tests {
		t.Run(tt.group, func(t *testing.T) {
			if got := shouldLinkToRacks(tt.group); got != tt.want {
				t.Errorf("shouldLinkToRacks(%q) = %t, want %t", tt.group, got, tt.want)
			}
		})
	}
}

func TestGetConfigGroupPrefixTransform(t *testing.T) {
	tests := []struct{ input, want string }{
		{"0100", "01"}, {"0200", "02"}, {"0315", "03"},
		{"1", ""}, {"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := getConfigGroupPrefix(tt.input); got != tt.want {
				t.Errorf("getConfigGroupPrefix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- Step info and field mapping tests ---

func TestBuildTransformStepInfo(t *testing.T) {
	rec := import_.CsvRecord{PartNumber: "FAKE-PN", Description: "Test Device", Quantity: 2, ConfigGroup: "0300"}

	t.Run("rack items", func(t *testing.T) {
		created := CreatedItems{Racks: []*devicetypes.CaniRackType{{ID: uuid.New(), Name: "Rack-01"}}}
		info := buildTransformStepInfo(rec, "rack", created)

		if info.HwType != "rack" {
			t.Errorf("HwType = %q, want %q", info.HwType, "rack")
		}
		if info.Quantity != 2 {
			t.Errorf("Quantity = %d, want 2", info.Quantity)
		}
		if len(info.CreatedItems) != 1 {
			t.Errorf("CreatedItems = %d, want 1", len(info.CreatedItems))
		}
	})

	t.Run("device items", func(t *testing.T) {
		created := CreatedItems{Devices: []*devicetypes.CaniDeviceType{{ID: uuid.New(), Name: "Server-01"}}}
		info := buildTransformStepInfo(rec, "node", created)

		if info.HwType != "node" {
			t.Errorf("HwType = %q, want %q", info.HwType, "node")
		}
		if len(info.CreatedItems) != 1 {
			t.Errorf("CreatedItems = %d, want 1", len(info.CreatedItems))
		}
	})
}

func TestBuildFieldMappings(t *testing.T) {
	t.Run("with config group", func(t *testing.T) {
		rec := import_.CsvRecord{PartNumber: "FAKE-PN", Description: "Some Device Description", ConfigGroup: "0200"}
		mappings := buildFieldMappings(rec, "switch", CreatedItems{})

		if len(mappings) != 4 {
			t.Fatalf("expected 4 mappings, got %d", len(mappings))
		}
		if mappings[0].SourceField != "PartNumber" {
			t.Errorf("first mapping source = %q, want PartNumber", mappings[0].SourceField)
		}
		if mappings[1].TargetField != "Name" {
			t.Errorf("second mapping target = %q, want Name", mappings[1].TargetField)
		}
		if mappings[2].SourceField != "ConfigGroup" {
			t.Errorf("third mapping source = %q, want ConfigGroup", mappings[2].SourceField)
		}
		if mappings[3].TargetField != "HardwareType" {
			t.Errorf("fourth mapping target = %q, want HardwareType", mappings[3].TargetField)
		}
	})

	t.Run("without config group", func(t *testing.T) {
		rec := import_.CsvRecord{PartNumber: "FAKE-PN", Description: "Device"}
		mappings := buildFieldMappings(rec, "node", CreatedItems{})

		if len(mappings) != 3 {
			t.Fatalf("expected 3 mappings without ConfigGroup, got %d", len(mappings))
		}
	})
}

// --- Metadata tests ---

func TestBuildProviderMetadata(t *testing.T) {
	commands.CsvFlag = "test.csv"
	t.Cleanup(func() { commands.CsvFlag = "" })

	rec := import_.CsvRecord{PartNumber: "P12345", ConfigGroup: "0200"}
	meta := buildProviderMetadata(rec)

	example, ok := meta["example"].(map[string]any)
	if !ok {
		t.Fatal("expected 'example' key in metadata")
	}
	if example["Source"] != "test.csv" {
		t.Errorf("Source = %v, want %q", example["Source"], "test.csv")
	}
	if example["PartNumber"] != "P12345" {
		t.Errorf("PartNumber = %v, want %q", example["PartNumber"], "P12345")
	}
	if example["ConfigGroup"] != "0200" {
		t.Errorf("ConfigGroup = %v, want %q", example["ConfigGroup"], "0200")
	}
}

func TestInitInventoryMaps(t *testing.T) {
	t.Run("initializes nil maps", func(t *testing.T) {
		inv := devicetypes.Inventory{}
		initInventoryMaps(&inv)
		if inv.Racks == nil {
			t.Error("expected Racks map to be initialized")
		}
		if inv.Devices == nil {
			t.Error("expected Devices map to be initialized")
		}
		if inv.Cables == nil {
			t.Error("expected Cables map to be initialized")
		}
	})

	t.Run("preserves existing maps", func(t *testing.T) {
		rackID := uuid.New()
		inv := devicetypes.Inventory{
			Racks: map[uuid.UUID]*devicetypes.CaniRackType{rackID: {Name: "existing"}},
		}
		initInventoryMaps(&inv)
		if _, ok := inv.Racks[rackID]; !ok {
			t.Error("initInventoryMaps overwrote existing Racks map")
		}
	})
}

// --- Summary tests ---

func TestBuildTransformSummary(t *testing.T) {
	rackID, deviceID := uuid.New(), uuid.New()
	inv := &devicetypes.Inventory{
		Racks:   map[uuid.UUID]*devicetypes.CaniRackType{rackID: {ID: rackID, Name: "Rack-001"}},
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{deviceID: {ID: deviceID, Name: "Server-001", Parent: rackID}},
	}

	summary := BuildTransformSummary(inv)
	if len(summary.RackNames) != 1 {
		t.Errorf("RackNames len = %d, want 1", len(summary.RackNames))
	}
	if devices, ok := summary.DevicesByRack["Rack-001"]; !ok || len(devices) != 1 {
		t.Errorf("expected 1 device in Rack-001, got %v", summary.DevicesByRack)
	}
}

func TestBuildNewItemsSummary(t *testing.T) {
	t.Run("with parent rack", func(t *testing.T) {
		rackID := uuid.New()
		inv := &devicetypes.Inventory{
			Racks: map[uuid.UUID]*devicetypes.CaniRackType{rackID: {ID: rackID, Name: "Rack-001"}},
		}
		created := CreatedItems{
			Racks:   []*devicetypes.CaniRackType{{ID: rackID, Name: "Rack-001"}},
			Devices: []*devicetypes.CaniDeviceType{{Name: "Server-001", Parent: rackID}},
		}
		summary := buildNewItemsSummary(created, inv)

		if len(summary.RackNames) != 1 || summary.RackNames[0] != "Rack-001" {
			t.Errorf("RackNames = %v, want [Rack-001]", summary.RackNames)
		}
		if devices, ok := summary.DevicesByRack["Rack-001"]; !ok || len(devices) != 1 {
			t.Errorf("expected 1 device in Rack-001, got %v", summary.DevicesByRack)
		}
	})

	t.Run("orphan device", func(t *testing.T) {
		inv := &devicetypes.Inventory{Racks: map[uuid.UUID]*devicetypes.CaniRackType{}}
		created := CreatedItems{Devices: []*devicetypes.CaniDeviceType{{Name: "Orphan-001"}}}
		summary := buildNewItemsSummary(created, inv)

		if devices, ok := summary.DevicesByRack[""]; !ok || len(devices) != 1 {
			t.Errorf("expected 1 orphan device under empty key, got %v", summary.DevicesByRack)
		}
	})
}

// --- Utility function tests ---

func TestGenerateName(t *testing.T) {
	tests := []struct {
		name, description string
		index, total      int
		want              string
	}{
		{"single item", "HPE Server", 0, 1, "HPE Server"},
		{"first of three", "HPE Server", 0, 3, "HPE Server-001"},
		{"second of three", "HPE Server", 1, 3, "HPE Server-002"},
		{"truncates long name single", "HPE ProLiant DL380 Gen11 High Performance Server", 0, 1, "HPE ProLiant DL380 Gen11 High "},
		{"truncates long name multi", "HPE ProLiant DL380 Gen11 High Performance Server", 0, 2, "HPE ProLiant DL380 Gen11 High -001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateName(tt.description, tt.index, tt.total); got != tt.want {
				t.Errorf("generateName(%q, %d, %d) = %q, want %q",
					tt.description, tt.index, tt.total, got, tt.want)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct{ input, want string }{
		{"HPE Cat6 Cable", "hpe-cat6-cable"},
		{"Simple", "simple"},
		{"ALLCAPS", "allcaps"},
		{"with-dashes", "with-dashes"},
		{"Special!@#$Characters", "specialcharacters"},
		{"", ""},
		{"spaces  and   tabs", "spaces--and---tabs"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := slugify(tt.input); got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTruncateName(t *testing.T) {
	tests := []struct{ name, input, want string }{
		{"short string", "Hello", "Hello"},
		{"exactly 30 chars", "123456789012345678901234567890", "123456789012345678901234567890"},
		{"over 30 chars", "1234567890123456789012345678901", "123456789012345678901234567890"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateName(tt.input); got != tt.want {
				t.Errorf("truncateName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
