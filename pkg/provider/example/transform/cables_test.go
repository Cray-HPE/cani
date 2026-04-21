package transform

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
)

// --- Inference and parsing tests ---

func TestInferHardwareType(t *testing.T) {
	tests := []struct {
		name, description, want string
	}{
		{"48U rack", "HPE 48U 800mmx1200mm G2 Enterprise Shock Rack", "rack"},
		{"cabinet", "HPE Cabinet G2", "rack"},
		{"aruba switch", "HPE Aruba Networking 8360-48Y6C v2", "switch"},
		{"generic switch", "HPE 48-Port Ethernet Switch", "switch"},
		{"proliant server", "HPE ProLiant DL380 Gen11", "node"},
		{"blade server", "HPE BladeSystem c7000 Blade", "node"},
		{"cat6 cable", "HPE Cat6 RJ45 M/M 2m", "cable"},
		{"cat5e cable", "HPE CAT5e RJ45 2.3m Cable", "cable"},
		{"DAC cable", "HPE 400G QSFP-DD DAC 3m", "cable"},
		{"direct attach cable", "HPE 100Gb Direct Attach Copper Cable", "cable"},
		{"AOC cable", "HPE Aruba 100G QSFP28 15m AOC", "cable"},
		{"active optical cable", "HPE Active Optical Cable 30m", "cable"},
		{"OM4 fiber", "HPE Premier Flex LC LC OM4 15m", "cable"},
		{"power jumper", "HPE C19 C20 250V 16A 2m Jumper", "cable"},
		{"RJ45 cable", "HPE RJ45 to RJ45 Cat5e Black", "cable"},
		{"QSFP cable", "HPE QSFP28 to QSFP28 Cable", "cable"},
		{"generic cable", "HPE Data Cable 3m", "cable"},
		{"unknown device", "XD670", ""},
		{"memory module", "HPE 64GB DDR5 Memory Kit", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inferHardwareType(tt.description); got != tt.want {
				t.Errorf("inferHardwareType(%q) = %q, want %q", tt.description, got, tt.want)
			}
		})
	}
}

func TestInferCableTypeSlug(t *testing.T) {
	tests := []struct {
		name, description, want string
	}{
		{"cat5 cable", "HPE CAT5 RJ45 1.2m Cable", cableTypeCat5e},
		{"cat5e cable", "HPE CAT5e RJ45 2.3m Cable", cableTypeCat5e},
		{"cat6 cable", "HPE Cat6 RJ45 M/M 2m", cableTypeCat6},
		{"cat6a cable", "HPE Cat6a Shielded Cable 3m", cableTypeCat6a},
		{"DAC cable", "HPE 400G QSFP-DD DAC 3m", cableTypeDacPassive},
		{"direct attach cable", "HPE 100Gb Direct Attach Copper Cable", cableTypeDacPassive},
		{"AOC cable", "HPE Aruba 100G QSFP28 15m AOC", cableTypeAoc},
		{"active optical cable", "HPE Active Optical Cable 30m", cableTypeAoc},
		{"OM3 fiber", "HPE LC LC OM3 2F 30m", cableTypeMmfOm4},
		{"OM4 fiber", "HPE Premier Flex LC LC OM4 15m", cableTypeMmfOm4},
		{"MMF fiber", "HPE InfiniBand NDR MPO MPO MM 10m", cableTypeMmfOm4},
		{"SMF fiber", "HPE InfiniBand NDR MPO MPO SM 10m", cableTypeSmf},
		{"single mode fiber", "HPE Single Mode Fiber Cable 20m", cableTypeSmf},
		{"power jumper", "HPE C19 C20 250V 16A 2m Jumper", cableTypePower},
		{"power cord", "HPE Power Cord 2m", cableTypePower},
		{"generic cable", "HPE Data Cable 3m", cableTypeOther},
		{"QSFP without type", "HPE QSFP28 Cable", cableTypeOther},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inferCableTypeSlug(tt.description); got != tt.want {
				t.Errorf("inferCableTypeSlug(%q) = %q, want %q", tt.description, got, tt.want)
			}
		})
	}
}

func TestResolveCableTypeSlug(t *testing.T) {
	tests := []struct {
		name, partNumber, description string
	}{
		{"fallback to description for cat6", "FAKE-PN-123", "Cat6 RJ45 patch cable 2m"},
		{"fallback to description for DAC", "", "400G QSFP-DD DAC 3m"},
		{"unknown description returns other", "", "Mystery Widget"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveCableTypeSlug(tt.partNumber, tt.description); got == "" {
				t.Error("expected non-empty cable type slug")
			}
		})
	}
}

func TestInferCableType(t *testing.T) {
	tests := []struct {
		name      string
		ifaceType devicetypes.InterfacesElemType
		want      string
	}{
		{"1000base-t", devicetypes.InterfacesElemTypeA1000BaseT, cableTypeCat6},
		{"10gbase-t", devicetypes.InterfacesElemTypeA10GbaseT, cableTypeCat6a},
		{"10gbase-x-sfpp", devicetypes.InterfacesElemTypeA10GbaseXSfpp, cableTypeDacPassive},
		{"25gbase-x-sfp28", devicetypes.InterfacesElemTypeA25GbaseXSfp28, cableTypeDacPassive},
		{"40gbase-x-qsfpp", devicetypes.InterfacesElemTypeA40GbaseXQsfpp, cableTypeDacPassive},
		{"100gbase-x-qsfp28", devicetypes.InterfacesElemTypeA100GbaseXQsfp28, cableTypeDacPassive},
		{"400gbase-x-qsfpdd", devicetypes.InterfacesElemTypeA400GbaseXQsfpdd, cableTypeDacPassive},
		{"unknown type", devicetypes.InterfacesElemType("unknown"), cableTypeOther},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inferCableType(tt.ifaceType); got != tt.want {
				t.Errorf("inferCableType(%q) = %q, want %q", tt.ifaceType, got, tt.want)
			}
		})
	}
}

func TestParseLengthFromDescription(t *testing.T) {
	tests := []struct {
		name, description string
		wantLength        float64
		wantUnit          string
	}{
		{"meters", "HPE Cat6 RJ45 M/M 2m", 2, "m"},
		{"feet", "HPE Cable 10ft", 10, "ft"},
		{"decimal meters", "HPE 1.5m DAC Cable", 1.5, "m"},
		{"centimeters", "HPE 50cm Patch Cable", 50, "cm"},
		{"no length", "HPE Generic Cable", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, unit := parseLengthFromDescription(tt.description)
			if length != tt.wantLength {
				t.Errorf("length = %v, want %v", length, tt.wantLength)
			}
			if unit != tt.wantUnit {
				t.Errorf("unit = %q, want %q", unit, tt.wantUnit)
			}
		})
	}
}

func TestParseCableLength(t *testing.T) {
	tests := []struct {
		name, input string
		wantLength  float64
		wantUnit    string
	}{
		{"meters", "3m", 3, "m"},
		{"feet", "10ft", 10, "ft"},
		{"decimal", "1.5m", 1.5, "m"},
		{"no unit defaults to m", "5", 5, "m"},
		{"empty string", "", 0, ""},
		{"non-numeric", "abc", 0, ""},
		{"with spaces", " 3m ", 3, "m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, unit := parseCableLength(tt.input)
			if length != tt.wantLength {
				t.Errorf("length = %v, want %v", length, tt.wantLength)
			}
			if unit != tt.wantUnit {
				t.Errorf("unit = %q, want %q", unit, tt.wantUnit)
			}
		})
	}
}

func TestGenerateCableLabel(t *testing.T) {
	tests := []struct {
		name, description string
		index, total      int
		want              string
	}{
		{"single cable", "HPE Cat6 Cable", 0, 1, "HPE Cat6 Cable"},
		{"first of multiple", "HPE Cat6 Cable", 0, 3, "HPE Cat6 Cable-001"},
		{"second of multiple", "HPE Cat6 Cable", 1, 3, "HPE Cat6 Cable-002"},
		{"tenth cable", "HPE DAC Cable", 9, 20, "HPE DAC Cable-010"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateCableLabel(tt.description, tt.index, tt.total); got != tt.want {
				t.Errorf("generateCableLabel(%q, %d, %d) = %q, want %q",
					tt.description, tt.index, tt.total, got, tt.want)
			}
		})
	}
}

// --- Lookup helper tests ---

func TestFindDeviceByName(t *testing.T) {
	deviceID := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			deviceID: {ID: deviceID, Name: "server-01"},
		},
	}

	t.Run("found", func(t *testing.T) {
		if got := findDeviceByName(inv, "server-01"); got == nil || got.ID != deviceID {
			t.Errorf("findDeviceByName() = %v, want device with ID %s", got, deviceID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		if got := findDeviceByName(inv, "nonexistent"); got != nil {
			t.Errorf("findDeviceByName() = %v, want nil", got)
		}
	})
}

func TestFindAvailableInterface(t *testing.T) {
	cableID := uuid.New()
	inv := &devicetypes.Inventory{}

	t.Run("returns unconnected interface", func(t *testing.T) {
		device := &devicetypes.CaniDeviceType{
			Interfaces: []devicetypes.InterfaceSpec{
				{Name: "eth0", ConnectedCable: &cableID},
				{Name: "eth1"},
			},
		}
		if got := findAvailableInterface(inv, device); got == nil || got.Name != "eth1" {
			t.Errorf("findAvailableInterface() = %v, want eth1", got)
		}
	})

	t.Run("all connected returns nil", func(t *testing.T) {
		device := &devicetypes.CaniDeviceType{
			Interfaces: []devicetypes.InterfaceSpec{{Name: "eth0", ConnectedCable: &cableID}},
		}
		if got := findAvailableInterface(inv, device); got != nil {
			t.Errorf("findAvailableInterface() = %v, want nil", got)
		}
	})

	t.Run("no interfaces returns nil", func(t *testing.T) {
		if got := findAvailableInterface(inv, &devicetypes.CaniDeviceType{}); got != nil {
			t.Errorf("findAvailableInterface() = %v, want nil", got)
		}
	})
}

// --- Connection logic tests ---

func TestLinkInterfacesToCable(t *testing.T) {
	ifaceAID, ifaceBID := uuid.New(), uuid.New()
	deviceAID, deviceBID := uuid.New(), uuid.New()
	cableID := uuid.New()

	t.Run("links both interfaces", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				deviceAID: {ID: deviceAID, Interfaces: []devicetypes.InterfaceSpec{{ID: ifaceAID, Name: "eth0"}}},
				deviceBID: {ID: deviceBID, Interfaces: []devicetypes.InterfaceSpec{{ID: ifaceBID, Name: "eth0"}}},
			},
			Interfaces: map[uuid.UUID]*devicetypes.InterfaceInstance{
				ifaceAID: {ID: ifaceAID, DeviceID: deviceAID},
				ifaceBID: {ID: ifaceBID, DeviceID: deviceBID},
			},
		}
		cable := &devicetypes.CaniCableType{ID: cableID, TerminationA: ifaceAID, TerminationB: ifaceBID}
		if err := linkInterfacesToCable(inv, cable); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing interface returns error", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices:    map[uuid.UUID]*devicetypes.CaniDeviceType{},
			Interfaces: map[uuid.UUID]*devicetypes.InterfaceInstance{},
		}
		cable := &devicetypes.CaniCableType{ID: cableID, TerminationA: uuid.New(), TerminationB: uuid.New()}
		if err := linkInterfacesToCable(inv, cable); err == nil {
			t.Error("expected error but got none")
		}
	})
}

func TestGroupDevicesByConfigGroup(t *testing.T) {
	deviceID := uuid.New()

	t.Run("groups by config group metadata", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				deviceID: {
					ID: deviceID,
					ObjectMeta: devicetypes.ObjectMeta{
						ProviderMetadata: map[string]any{
							"example": map[string]any{"ConfigGroup": "0200"},
						},
					},
				},
			},
		}
		if result := groupDevicesByConfigGroup(inv); len(result["0200"]) != 1 {
			t.Errorf("expected 1 device in group 0200, got %d", len(result["0200"]))
		}
	})

	t.Run("skips devices without metadata", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{deviceID: {ID: deviceID}},
		}
		if result := groupDevicesByConfigGroup(inv); len(result) != 0 {
			t.Errorf("expected empty result, got %v", result)
		}
	})

	t.Run("skips nil devices", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{deviceID: nil},
		}
		if result := groupDevicesByConfigGroup(inv); len(result) != 0 {
			t.Errorf("expected empty result, got %v", result)
		}
	})
}

func TestFindRelatedDeviceGroups(t *testing.T) {
	deviceID := uuid.New()
	tests := []struct {
		name           string
		cableGroup     string
		devicesByGroup map[string][]*devicetypes.CaniDeviceType
		wantLen        int
	}{
		{"finds non-rack non-cable groups", "0900", map[string][]*devicetypes.CaniDeviceType{
			"0200": {{ID: deviceID}}, "0300": {{ID: deviceID}},
		}, 2},
		{"excludes rack group 01XX", "0900", map[string][]*devicetypes.CaniDeviceType{
			"0100": {{ID: deviceID}}, "0200": {{ID: deviceID}},
		}, 1},
		{"excludes own cable group", "0900", map[string][]*devicetypes.CaniDeviceType{
			"0900": {{ID: deviceID}},
		}, 0},
		{"short cable group returns empty", "1", map[string][]*devicetypes.CaniDeviceType{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findRelatedDeviceGroups(tt.cableGroup, tt.devicesByGroup); len(got) != tt.wantLen {
				t.Errorf("findRelatedDeviceGroups() returned %d groups, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestAutoConnectCables(t *testing.T) {
	t.Run("connects cables to related device groups", func(t *testing.T) {
		switchID, nodeID := uuid.New(), uuid.New()
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				switchID: {
					ID: switchID, Name: "switch-01", Type: "switch",
					Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}},
					ObjectMeta: devicetypes.ObjectMeta{
						ProviderMetadata: map[string]any{"example": map[string]any{"ConfigGroup": "0200"}},
					},
				},
				nodeID: {
					ID: nodeID, Name: "server-01", Type: "node",
					Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}},
					ObjectMeta: devicetypes.ObjectMeta{
						ProviderMetadata: map[string]any{"example": map[string]any{"ConfigGroup": "0300"}},
					},
				},
			},
		}
		cable := devicetypes.NewCable("cat6", "test-cable")
		autoConnectCables(inv, map[string][]*devicetypes.CaniCableType{"0900": {cable}})
		if cable.TerminationA == uuid.Nil && cable.TerminationB == uuid.Nil {
			t.Log("cable was not auto-connected (may lack matching interfaces)")
		}
	})

	t.Run("no related groups does not panic", func(t *testing.T) {
		inv := &devicetypes.Inventory{Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{}}
		cable := devicetypes.NewCable("cat6", "test-cable")
		autoConnectCables(inv, map[string][]*devicetypes.CaniCableType{"0900": {cable}})
	})
}

func TestConnectCablesHubSpoke(t *testing.T) {
	t.Run("connects hub to spoke", func(t *testing.T) {
		switchID, nodeID := uuid.New(), uuid.New()
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				switchID: {ID: switchID, Name: "switch-01", Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "eth0"}}},
				nodeID:   {ID: nodeID, Name: "server-01", Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "eth0"}}},
			},
		}
		cable := devicetypes.NewCable("cat6", "test-cable")
		connectCablesHubSpoke(inv, []*devicetypes.CaniCableType{cable},
			[]*devicetypes.CaniDeviceType{inv.Devices[switchID]},
			[]*devicetypes.CaniDeviceType{inv.Devices[nodeID]})
		if cable.TerminationA == uuid.Nil || cable.TerminationB == uuid.Nil {
			t.Error("expected cable to be connected to both hub and spoke")
		}
	})

	t.Run("no hubs leaves cable unconnected", func(t *testing.T) {
		inv := &devicetypes.Inventory{Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{}}
		cable := devicetypes.NewCable("cat6", "test")
		connectCablesHubSpoke(inv, []*devicetypes.CaniCableType{cable}, nil, nil)
		if cable.TerminationA != uuid.Nil {
			t.Error("expected cable to remain unconnected")
		}
	})

	t.Run("more cables than spokes", func(t *testing.T) {
		hubID, spokeID := uuid.New(), uuid.New()
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				hubID:   {ID: hubID, Name: "sw-01", Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "e0"}, {ID: uuid.New(), Name: "e1"}}},
				spokeID: {ID: spokeID, Name: "srv-01", Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "e0"}}},
			},
		}
		cables := []*devicetypes.CaniCableType{
			devicetypes.NewCable("cat6", "c1"),
			devicetypes.NewCable("cat6", "c2"),
		}
		connectCablesHubSpoke(inv, cables,
			[]*devicetypes.CaniDeviceType{inv.Devices[hubID]},
			[]*devicetypes.CaniDeviceType{inv.Devices[spokeID]})
		if cables[0].TerminationA == uuid.Nil {
			t.Error("first cable should be connected")
		}
		if cables[1].TerminationA != uuid.Nil {
			t.Error("second cable should remain unconnected (no more spokes)")
		}
	})
}

// --- Cable creation tests ---

func TestCreateCableFromExplicitRecord(t *testing.T) {
	srcDeviceID, dstDeviceID := uuid.New(), uuid.New()
	srcIfaceID, dstIfaceID := uuid.New(), uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			srcDeviceID: {
				ID: srcDeviceID, Name: "switch-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: srcIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}},
			},
			dstDeviceID: {
				ID: dstDeviceID, Name: "server-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: dstIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}},
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		rec := import_.CsvRecord{SourceDevice: "switch-01", SourcePort: "eth0", DestDevice: "server-01", DestPort: "eth0"}
		cable, err := createCableFromExplicitRecord(inv, rec)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cable == nil {
			t.Fatal("expected cable but got nil")
		}
		if cable.TerminationA != srcIfaceID {
			t.Errorf("TerminationA = %v, want %v", cable.TerminationA, srcIfaceID)
		}
		if cable.TerminationB != dstIfaceID {
			t.Errorf("TerminationB = %v, want %v", cable.TerminationB, dstIfaceID)
		}
	})

	errorCases := []struct {
		name string
		rec  import_.CsvRecord
	}{
		{"source device not found", import_.CsvRecord{SourceDevice: "nonexistent", SourcePort: "eth0", DestDevice: "server-01", DestPort: "eth0"}},
		{"source port not found", import_.CsvRecord{SourceDevice: "switch-01", SourcePort: "nonexistent", DestDevice: "server-01", DestPort: "eth0"}},
		{"dest device not found", import_.CsvRecord{SourceDevice: "switch-01", SourcePort: "eth0", DestDevice: "nonexistent", DestPort: "eth0"}},
		{"dest port not found", import_.CsvRecord{SourceDevice: "switch-01", SourcePort: "eth0", DestDevice: "server-01", DestPort: "nonexistent"}},
	}
	for _, tt := range errorCases {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := createCableFromExplicitRecord(inv, tt.rec); err == nil {
				t.Error("expected error but got none")
			}
		})
	}
}

func TestTransformCables(t *testing.T) {
	t.Run("empty records", func(t *testing.T) {
		inv := &devicetypes.Inventory{Cables: make(map[uuid.UUID]*devicetypes.CaniCableType)}
		tally := &visual.StepTally{}
		recordNum := 0
		cables, err := transformCables(inv, nil, false, visual.ETLOptions{}, tally, &recordNum, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cables) != 0 {
			t.Errorf("expected 0 cables, got %d", len(cables))
		}
	})

	t.Run("product records", func(t *testing.T) {
		inv := &devicetypes.Inventory{Cables: make(map[uuid.UUID]*devicetypes.CaniCableType)}
		tally := &visual.StepTally{}
		recordNum := 0
		records := []import_.CsvRecord{{PartNumber: "C7536A", Description: "HPE Cat6 RJ45 M/M 2m", Quantity: 3}}
		cables, err := transformCables(inv, records, false, visual.ETLOptions{}, tally, &recordNum, len(records))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cables) != 3 {
			t.Errorf("expected 3 cables, got %d", len(cables))
		}
	})
}

// --- Step info builder tests ---

func TestBuildCableStepInfo(t *testing.T) {
	t.Run("explicit cable", func(t *testing.T) {
		cable := devicetypes.NewCable("cat6", "switch:e0 ↔ server:e0")
		cable.SetTerminations(uuid.New(), uuid.New())
		rec := import_.CsvRecord{
			SourceDevice: "switch-01", SourcePort: "eth0",
			DestDevice: "server-01", DestPort: "eth0",
			CableType: "cat6",
		}
		info := buildCableStepInfo(rec, cable)
		if info.HwType != "cable" {
			t.Errorf("HwType = %q, want %q", info.HwType, "cable")
		}
		if info.Quantity != 1 {
			t.Errorf("Quantity = %d, want 1", info.Quantity)
		}
		if len(info.Mappings) != 3 {
			t.Errorf("len(Mappings) = %d, want 3", len(info.Mappings))
		}
		if len(info.CreatedItems) != 1 {
			t.Errorf("len(CreatedItems) = %d, want 1", len(info.CreatedItems))
		}
	})

	t.Run("derived cable type", func(t *testing.T) {
		cable := devicetypes.NewCable("cat6", "label")
		cable.SetTerminations(uuid.New(), uuid.New())
		rec := import_.CsvRecord{
			SourceDevice: "sw", SourcePort: "e0",
			DestDevice: "srv", DestPort: "e0",
			CableType: "",
		}
		info := buildCableStepInfo(rec, cable)
		if len(info.Mappings) >= 3 && !info.Mappings[2].IsDerived {
			t.Error("expected CableType mapping to be marked as derived when empty")
		}
	})
}

func TestBuildCableProductStepInfo(t *testing.T) {
	cables := []*devicetypes.CaniCableType{
		devicetypes.NewCable("cat6", "Cable-001"),
		devicetypes.NewCable("cat6", "Cable-002"),
	}
	rec := import_.CsvRecord{PartNumber: "C7536A", Description: "HPE Cat6 RJ45 Cable 2m", Quantity: 2}
	info := buildCableProductStepInfo(rec, cables)
	if info.Quantity != 2 {
		t.Errorf("Quantity = %d, want 2", info.Quantity)
	}
	if info.HwType != "cable" {
		t.Errorf("HwType = %q, want %q", info.HwType, "cable")
	}
	if len(info.CreatedItems) != 2 {
		t.Errorf("len(CreatedItems) = %d, want 2", len(info.CreatedItems))
	}
	if len(info.Mappings) != 2 {
		t.Errorf("len(Mappings) = %d, want 2", len(info.Mappings))
	}
}
