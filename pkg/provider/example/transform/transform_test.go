package transform

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/example/commands"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

func TestDeviceTypePriority(t *testing.T) {
	tests := []struct {
		hwType string
		want   int
	}{
		{"pdu", 0},
		{"cdu", 0},
		{"node", 1},
		{"blade", 1},
		{"chassis", 1},
		{"switch", 2},
		{"mgmt-switch", 2},
		{"hsn-switch", 2},
		{"", 1},
	}

	for _, tt := range tests {
		t.Run(tt.hwType, func(t *testing.T) {
			got := deviceTypePriority(tt.hwType)
			if got != tt.want {
				t.Errorf("deviceTypePriority(%q) = %d, want %d", tt.hwType, got, tt.want)
			}
		})
	}
}

func TestSortDevicesByRackPriority(t *testing.T) {
	inv := &devicetypes.Inventory{
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
	}

	switchID := uuid.New()
	nodeID := uuid.New()
	pduID := uuid.New()
	bladeID := uuid.New()
	cduID := uuid.New()

	inv.Devices[switchID] = &devicetypes.CaniDeviceType{ID: switchID, Type: devicetypes.Type("switch")}
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{ID: nodeID, Type: devicetypes.Type("node")}
	inv.Devices[pduID] = &devicetypes.CaniDeviceType{ID: pduID, Type: devicetypes.Type("pdu")}
	inv.Devices[bladeID] = &devicetypes.CaniDeviceType{ID: bladeID, Type: devicetypes.Type("blade")}
	inv.Devices[cduID] = &devicetypes.CaniDeviceType{ID: cduID, Type: devicetypes.Type("cdu")}

	ids := []uuid.UUID{switchID, nodeID, pduID, bladeID, cduID}
	sortDevicesByRackPriority(inv, ids)

	// PDUs/CDUs should come first, then nodes/blades, then switches
	for i, id := range ids {
		dev := inv.Devices[id]
		pri := deviceTypePriority(string(dev.Type))
		if i > 0 {
			prevDev := inv.Devices[ids[i-1]]
			prevPri := deviceTypePriority(string(prevDev.Type))
			if pri < prevPri {
				t.Errorf("device at index %d (%s, priority %d) sorted before index %d (%s, priority %d)",
					i, dev.Type, pri, i-1, prevDev.Type, prevPri)
			}
		}
	}

	// Verify first two are pdu/cdu (priority 0)
	for _, id := range ids[:2] {
		dev := inv.Devices[id]
		if deviceTypePriority(string(dev.Type)) != 0 {
			t.Errorf("expected pdu/cdu in first two positions, got %q", dev.Type)
		}
	}

	// Verify last is switch (priority 2)
	lastDev := inv.Devices[ids[len(ids)-1]]
	if deviceTypePriority(string(lastDev.Type)) != 2 {
		t.Errorf("expected switch in last position, got %q", lastDev.Type)
	}
}

func TestGroupDevicesByZone(t *testing.T) {
	inv := &devicetypes.Inventory{
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
	}

	pduID := uuid.New()
	cduID := uuid.New()
	nodeID := uuid.New()
	switchID := uuid.New()

	inv.Devices[pduID] = &devicetypes.CaniDeviceType{ID: pduID, Type: devicetypes.Type("pdu")}
	inv.Devices[cduID] = &devicetypes.CaniDeviceType{ID: cduID, Type: devicetypes.Type("cdu")}
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{ID: nodeID, Type: devicetypes.Type("node")}
	inv.Devices[switchID] = &devicetypes.CaniDeviceType{ID: switchID, Type: devicetypes.Type("switch")}

	ids := []uuid.UUID{pduID, cduID, nodeID, switchID}
	bottom, middle, top := groupDevicesByZone(inv, ids)

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
	// Reset global state
	resetRackPositionStates()

	inv := &devicetypes.Inventory{
		Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
	}

	rackID := uuid.New()
	inv.Racks[rackID] = &devicetypes.CaniRackType{
		ID:      rackID,
		Name:    "Rack-001",
		UHeight: 48,
		Devices: []uuid.UUID{},
	}

	switchID := uuid.New()
	nodeID := uuid.New()
	pduID := uuid.New()

	inv.Devices[switchID] = &devicetypes.CaniDeviceType{ID: switchID, Type: devicetypes.Type("switch"), Slug: "switch-1u"}
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{ID: nodeID, Type: devicetypes.Type("node"), Slug: "node-1u"}
	inv.Devices[pduID] = &devicetypes.CaniDeviceType{ID: pduID, Type: devicetypes.Type("pdu"), Slug: "pdu-1u"}

	racksByGroup := map[string][]uuid.UUID{
		"0100": {rackID},
	}
	devicesByGroup := map[string][]uuid.UUID{
		"0200": {switchID, nodeID, pduID},
	}

	assignConfigGroupParenting(inv, racksByGroup, devicesByGroup)

	pdu := inv.Devices[pduID]
	node := inv.Devices[nodeID]
	sw := inv.Devices[switchID]

	// PDU should be at the bottom (low U)
	if pdu.RackPosition != 1 {
		t.Errorf("PDU position = %d, want 1 (bottom)", pdu.RackPosition)
	}

	// Switch should be at the top (U48)
	if sw.RackPosition != 48 {
		t.Errorf("switch position = %d, want 48 (top)", sw.RackPosition)
	}

	// Node should be just below switch (U47, filling downward)
	if node.RackPosition != 47 {
		t.Errorf("node position = %d, want 47 (below switch)", node.RackPosition)
	}
}

func TestRackZonesFillCorrectly(t *testing.T) {
	resetRackPositionStates()

	inv := &devicetypes.Inventory{
		Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
	}

	rackID := uuid.New()
	inv.Racks[rackID] = &devicetypes.CaniRackType{
		ID:      rackID,
		Name:    "Rack-001",
		UHeight: 10,
		Devices: []uuid.UUID{},
	}

	// 2 PDUs (1U each) at bottom, 2 switches (1U each) at top, 1 node in middle
	pdu1 := uuid.New()
	pdu2 := uuid.New()
	sw1 := uuid.New()
	sw2 := uuid.New()
	node1 := uuid.New()

	inv.Devices[pdu1] = &devicetypes.CaniDeviceType{ID: pdu1, Type: devicetypes.Type("pdu")}
	inv.Devices[pdu2] = &devicetypes.CaniDeviceType{ID: pdu2, Type: devicetypes.Type("cdu")}
	inv.Devices[sw1] = &devicetypes.CaniDeviceType{ID: sw1, Type: devicetypes.Type("switch")}
	inv.Devices[sw2] = &devicetypes.CaniDeviceType{ID: sw2, Type: devicetypes.Type("mgmt-switch")}
	inv.Devices[node1] = &devicetypes.CaniDeviceType{ID: node1, Type: devicetypes.Type("node")}

	racksByGroup := map[string][]uuid.UUID{"0100": {rackID}}
	devicesByGroup := map[string][]uuid.UUID{
		"0200": {pdu1, pdu2, sw1, sw2, node1},
	}

	assignConfigGroupParenting(inv, racksByGroup, devicesByGroup)

	// Bottom zone: U1, U2
	if inv.Devices[pdu1].RackPosition != 1 {
		t.Errorf("pdu1 position = %d, want 1", inv.Devices[pdu1].RackPosition)
	}
	if inv.Devices[pdu2].RackPosition != 2 {
		t.Errorf("pdu2 position = %d, want 2", inv.Devices[pdu2].RackPosition)
	}

	// Middle zone: node fills downward from below switches → U8
	if inv.Devices[node1].RackPosition != 8 {
		t.Errorf("node position = %d, want 8", inv.Devices[node1].RackPosition)
	}

	// Top zone: U10 (first switch), U9 (second switch, fills downward)
	if inv.Devices[sw1].RackPosition != 10 {
		t.Errorf("sw1 position = %d, want 10", inv.Devices[sw1].RackPosition)
	}
	if inv.Devices[sw2].RackPosition != 9 {
		t.Errorf("sw2 position = %d, want 9", inv.Devices[sw2].RackPosition)
	}
}

func TestClassifyRecords(t *testing.T) {
	tests := []struct {
		name        string
		records     []import_.CsvRecord
		wantRacks   int
		wantDevices int
		wantCables  int
		wantErr     bool
	}{
		{
			name:    "empty records",
			records: nil,
		},
		{
			name: "rack by description",
			records: []import_.CsvRecord{
				{PartNumber: "X", Description: "HPE 48U 800mmx1200mm G2 Enterprise Shock Rack", Quantity: 1},
			},
			wantRacks: 1,
		},
		{
			name: "cable by explicit endpoints",
			records: []import_.CsvRecord{
				{Description: "link", Quantity: 1, SourceDevice: "sw1", DestDevice: "srv1"},
			},
			wantCables: 1,
		},
		{
			name: "cable by description pattern",
			records: []import_.CsvRecord{
				{PartNumber: "X", Description: "HPE Cat6 RJ45 M/M 2m", Quantity: 1},
			},
			wantCables: 1,
		},
		{
			name: "switch by description",
			records: []import_.CsvRecord{
				{PartNumber: "X", Description: "HPE Aruba Networking 8360-48Y6C", Quantity: 1},
			},
			wantDevices: 1,
		},
		{
			name: "node by description",
			records: []import_.CsvRecord{
				{PartNumber: "X", Description: "HPE ProLiant DL380 Gen11", Quantity: 1},
			},
			wantDevices: 1,
		},
		{
			name: "unclassifiable returns error",
			records: []import_.CsvRecord{
				{PartNumber: "X", Description: "HPE 64GB DDR5 Memory Kit", Quantity: 1},
			},
			wantErr: true,
		},
		{
			name: "mixed record types",
			records: []import_.CsvRecord{
				{PartNumber: "R", Description: "HPE 48U Rack", Quantity: 1},
				{PartNumber: "S", Description: "HPE Aruba Switch", Quantity: 1},
				{PartNumber: "C", Description: "HPE Cat6 Cable 2m", Quantity: 1},
			},
			wantRacks:   1,
			wantDevices: 1,
			wantCables:  1,
		},
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

func TestGenerateName(t *testing.T) {
	tests := []struct {
		name        string
		description string
		index       int
		total       int
		want        string
	}{
		{"single item", "HPE Server", 0, 1, "HPE Server"},
		{"first of three", "HPE Server", 0, 3, "HPE Server-001"},
		{"second of three", "HPE Server", 1, 3, "HPE Server-002"},
		{"truncates long name single", "HPE ProLiant DL380 Gen11 High Performance Server", 0, 1, "HPE ProLiant DL380 Gen11 High "},
		{"truncates long name multi", "HPE ProLiant DL380 Gen11 High Performance Server", 0, 2, "HPE ProLiant DL380 Gen11 High -001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateName(tt.description, tt.index, tt.total)
			if got != tt.want {
				t.Errorf("generateName(%q, %d, %d) = %q, want %q",
					tt.description, tt.index, tt.total, got, tt.want)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
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
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTruncateName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"short string", "Hello", "Hello"},
		{"exactly 30 chars", "123456789012345678901234567890", "123456789012345678901234567890"},
		{"over 30 chars", "1234567890123456789012345678901", "123456789012345678901234567890"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateName(tt.input)
			if got != tt.want {
				t.Errorf("truncateName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
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
			Racks: map[uuid.UUID]*devicetypes.CaniRackType{
				rackID: {Name: "existing"},
			},
		}
		initInventoryMaps(&inv)
		if _, ok := inv.Racks[rackID]; !ok {
			t.Error("initInventoryMaps overwrote existing Racks map")
		}
	})
}

func TestBuildProviderMetadata(t *testing.T) {
	commands.CsvFlag = "test.csv"
	t.Cleanup(func() { commands.CsvFlag = "" })

	rec := import_.CsvRecord{
		PartNumber:  "P12345",
		ConfigGroup: "0200",
	}
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

func TestFindParentRackIDs(t *testing.T) {
	rack1 := uuid.New()
	rack2 := uuid.New()
	device1 := uuid.New()

	tests := []struct {
		name         string
		racksByGroup map[string][]uuid.UUID
		wantLen      int
	}{
		{
			name:         "empty map",
			racksByGroup: map[string][]uuid.UUID{},
			wantLen:      0,
		},
		{
			name:         "racks in 01XX group",
			racksByGroup: map[string][]uuid.UUID{"0100": {rack1, rack2}},
			wantLen:      2,
		},
		{
			name:         "racks in non-01XX group only",
			racksByGroup: map[string][]uuid.UUID{"0200": {device1}},
			wantLen:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findParentRackIDs(tt.racksByGroup)
			if len(got) != tt.wantLen {
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
		{"0100", false},
		{"0200", true},
		{"0300", true},
		{"0900", true},
		{"1", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.group, func(t *testing.T) {
			got := shouldLinkToRacks(tt.group)
			if got != tt.want {
				t.Errorf("shouldLinkToRacks(%q) = %t, want %t", tt.group, got, tt.want)
			}
		})
	}
}

func TestGetConfigGroupPrefixTransform(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0100", "01"},
		{"0200", "02"},
		{"0315", "03"},
		{"1", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := getConfigGroupPrefix(tt.input)
			if got != tt.want {
				t.Errorf("getConfigGroupPrefix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCreateItemsFromRecord(t *testing.T) {
	t.Run("creates rack items", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
			Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		}
		racksByGroup := make(map[string][]uuid.UUID)
		devicesByGroup := make(map[string][]uuid.UUID)

		rec := import_.CsvRecord{
			PartNumber:  "TEST-RACK",
			Description: "Test Rack",
			Quantity:    2,
			ConfigGroup: "0100",
		}

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

		rec := import_.CsvRecord{
			PartNumber:  "TEST-NODE",
			Description: "Test Server",
			Quantity:    3,
			ConfigGroup: "0200",
		}

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

		rec := import_.CsvRecord{
			PartNumber:  "TEST",
			Description: "Standalone Device",
			Quantity:    1,
		}

		createItemsFromRecord(inv, rec, "node", racksByGroup, devicesByGroup)

		if len(devicesByGroup) != 0 {
			t.Errorf("expected no config group entries, got %d", len(devicesByGroup))
		}
	})
}

func TestBuildDeviceFromRecord(t *testing.T) {
	id := uuid.New()
	rec := import_.CsvRecord{
		PartNumber:  "TEST-FAKE-PN",
		Description: "HPE ProLiant DL380",
		ConfigGroup: "0200",
	}

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
		device := &devicetypes.CaniDeviceType{
			Name: "original",
			Slug: "old-slug",
		}
		dt := &devicetypes.CaniDeviceType{
			Slug:         "new-slug",
			Manufacturer: "HPE",
			Model:        "DL380 Gen11",
			Type:         devicetypes.Type("node"),
			Interfaces:   []devicetypes.InterfaceSpec{{Name: "eth0"}},
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
		if len(device.Interfaces) != 1 {
			t.Errorf("Interfaces len = %d, want 1", len(device.Interfaces))
		}
	})

	t.Run("does not overwrite type when empty", func(t *testing.T) {
		device := &devicetypes.CaniDeviceType{
			Type: devicetypes.Type("switch"),
		}
		dt := &devicetypes.CaniDeviceType{
			Slug: "new-slug",
			Type: "",
		}

		populateFromDeviceType(device, dt)

		if device.Type != devicetypes.Type("switch") {
			t.Errorf("Type = %q, want %q (should not overwrite with empty)", device.Type, "switch")
		}
	})
}

func TestBuildTransformSummary(t *testing.T) {
	rackID := uuid.New()
	deviceID := uuid.New()

	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {ID: rackID, Name: "Rack-001"},
		},
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			deviceID: {ID: deviceID, Name: "Server-001", Parent: rackID},
		},
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
	rackID := uuid.New()

	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {ID: rackID, Name: "Rack-001"},
		},
	}

	created := CreatedItems{
		Racks: []*devicetypes.CaniRackType{
			{ID: rackID, Name: "Rack-001"},
		},
		Devices: []*devicetypes.CaniDeviceType{
			{Name: "Server-001", Parent: rackID},
		},
	}

	summary := buildNewItemsSummary(created, inv)

	if len(summary.RackNames) != 1 || summary.RackNames[0] != "Rack-001" {
		t.Errorf("RackNames = %v, want [Rack-001]", summary.RackNames)
	}
	if devices, ok := summary.DevicesByRack["Rack-001"]; !ok || len(devices) != 1 {
		t.Errorf("expected 1 device in Rack-001, got %v", summary.DevicesByRack)
	}
}

func TestBuildNewItemsSummaryOrphanDevice(t *testing.T) {
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{},
	}

	created := CreatedItems{
		Devices: []*devicetypes.CaniDeviceType{
			{Name: "Orphan-001"},
		},
	}

	summary := buildNewItemsSummary(created, inv)

	if devices, ok := summary.DevicesByRack[""]; !ok || len(devices) != 1 {
		t.Errorf("expected 1 orphan device under empty key, got %v", summary.DevicesByRack)
	}
}
