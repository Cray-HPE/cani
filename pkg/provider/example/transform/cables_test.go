package transform

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
)

func TestInferHardwareType(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
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
			got := inferHardwareType(tt.description)
			if got != tt.want {
				t.Errorf("inferHardwareType(%q) = %q, want %q", tt.description, got, tt.want)
			}
		})
	}
}

func TestInferCableTypeSlug(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
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
			got := inferCableTypeSlug(tt.description)
			if got != tt.want {
				t.Errorf("inferCableTypeSlug(%q) = %q, want %q", tt.description, got, tt.want)
			}
		})
	}
}

func TestParseLengthFromDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		wantLength  float64
		wantUnit    string
	}{
		{"meters", "HPE Cat6 RJ45 M/M 2m", 2, "m"},
		{"feet", "HPE Cable 10ft", 10, "ft"},
		{"decimal meters", "HPE 1.5m DAC Cable", 1.5, "m"},
		{"centimeters", "HPE 50cm Patch Cable", 50, "cm"},
		{"no length", "HPE Generic Cable", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLength, gotUnit := parseLengthFromDescription(tt.description)
			if gotLength != tt.wantLength {
				t.Errorf("parseLengthFromDescription(%q) length = %v, want %v", tt.description, gotLength, tt.wantLength)
			}
			if gotUnit != tt.wantUnit {
				t.Errorf("parseLengthFromDescription(%q) unit = %q, want %q", tt.description, gotUnit, tt.wantUnit)
			}
		})
	}
}

func TestGenerateCableLabel(t *testing.T) {
	tests := []struct {
		name        string
		description string
		index       int
		total       int
		want        string
	}{
		{"single cable", "HPE Cat6 Cable", 0, 1, "HPE Cat6 Cable"},
		{"first of multiple", "HPE Cat6 Cable", 0, 3, "HPE Cat6 Cable-001"},
		{"second of multiple", "HPE Cat6 Cable", 1, 3, "HPE Cat6 Cable-002"},
		{"tenth cable", "HPE DAC Cable", 9, 20, "HPE DAC Cable-010"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateCableLabel(tt.description, tt.index, tt.total)
			if got != tt.want {
				t.Errorf("generateCableLabel(%q, %d, %d) = %q, want %q",
					tt.description, tt.index, tt.total, got, tt.want)
			}
		})
	}
}

func TestParseCableLength(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantLength float64
		wantUnit   string
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
				t.Errorf("parseCableLength(%q) length = %v, want %v", tt.input, length, tt.wantLength)
			}
			if unit != tt.wantUnit {
				t.Errorf("parseCableLength(%q) unit = %q, want %q", tt.input, unit, tt.wantUnit)
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
			got := inferCableType(tt.ifaceType)
			if got != tt.want {
				t.Errorf("inferCableType(%q) = %q, want %q", tt.ifaceType, got, tt.want)
			}
		})
	}
}

func TestFindDeviceByName(t *testing.T) {
	deviceID := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			deviceID: {ID: deviceID, Name: "server-01"},
		},
	}

	t.Run("found", func(t *testing.T) {
		got := findDeviceByName(inv, "server-01")
		if got == nil || got.ID != deviceID {
			t.Errorf("findDeviceByName() = %v, want device with ID %s", got, deviceID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		got := findDeviceByName(inv, "nonexistent")
		if got != nil {
			t.Errorf("findDeviceByName() = %v, want nil", got)
		}
	})
}

func TestFindAvailableInterface(t *testing.T) {
	cableID := uuid.New()

	t.Run("returns unconnected interface", func(t *testing.T) {
		device := &devicetypes.CaniDeviceType{
			Interfaces: []devicetypes.InterfaceSpec{
				{Name: "eth0", ConnectedCable: &cableID},
				{Name: "eth1"},
			},
		}
		inv := &devicetypes.Inventory{}

		got := findAvailableInterface(inv, device)
		if got == nil || got.Name != "eth1" {
			t.Errorf("findAvailableInterface() = %v, want eth1", got)
		}
	})

	t.Run("all connected returns nil", func(t *testing.T) {
		device := &devicetypes.CaniDeviceType{
			Interfaces: []devicetypes.InterfaceSpec{
				{Name: "eth0", ConnectedCable: &cableID},
			},
		}
		inv := &devicetypes.Inventory{}

		got := findAvailableInterface(inv, device)
		if got != nil {
			t.Errorf("findAvailableInterface() = %v, want nil", got)
		}
	})

	t.Run("no interfaces returns nil", func(t *testing.T) {
		device := &devicetypes.CaniDeviceType{}
		inv := &devicetypes.Inventory{}

		got := findAvailableInterface(inv, device)
		if got != nil {
			t.Errorf("findAvailableInterface() = %v, want nil", got)
		}
	})
}

func TestLinkInterfacesToCable(t *testing.T) {
	ifaceAID := uuid.New()
	ifaceBID := uuid.New()
	deviceAID := uuid.New()
	deviceBID := uuid.New()
	cableID := uuid.New()

	t.Run("links both interfaces", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				deviceAID: {
					ID:         deviceAID,
					Interfaces: []devicetypes.InterfaceSpec{{ID: ifaceAID, Name: "eth0"}},
				},
				deviceBID: {
					ID:         deviceBID,
					Interfaces: []devicetypes.InterfaceSpec{{ID: ifaceBID, Name: "eth0"}},
				},
			},
			Interfaces: map[uuid.UUID]*devicetypes.InterfaceInstance{
				ifaceAID: {ID: ifaceAID, DeviceID: deviceAID},
				ifaceBID: {ID: ifaceBID, DeviceID: deviceBID},
			},
		}

		cable := &devicetypes.CaniCableType{
			ID:           cableID,
			TerminationA: ifaceAID,
			TerminationB: ifaceBID,
		}

		err := linkInterfacesToCable(inv, cable)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing interface returns error", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices:    map[uuid.UUID]*devicetypes.CaniDeviceType{},
			Interfaces: map[uuid.UUID]*devicetypes.InterfaceInstance{},
		}

		cable := &devicetypes.CaniCableType{
			ID:           cableID,
			TerminationA: uuid.New(),
			TerminationB: uuid.New(),
		}

		err := linkInterfacesToCable(inv, cable)
		if err == nil {
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
							"example": map[string]any{
								"ConfigGroup": "0200",
							},
						},
					},
				},
			},
		}

		result := groupDevicesByConfigGroup(inv)
		if len(result["0200"]) != 1 {
			t.Errorf("expected 1 device in group 0200, got %d", len(result["0200"]))
		}
	})

	t.Run("skips devices without metadata", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				deviceID: {ID: deviceID},
			},
		}

		result := groupDevicesByConfigGroup(inv)
		if len(result) != 0 {
			t.Errorf("expected empty result, got %v", result)
		}
	})

	t.Run("skips nil devices", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				deviceID: nil,
			},
		}

		result := groupDevicesByConfigGroup(inv)
		if len(result) != 0 {
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
		{
			name:       "finds non-rack non-cable groups",
			cableGroup: "0900",
			devicesByGroup: map[string][]*devicetypes.CaniDeviceType{
				"0200": {{ID: deviceID}},
				"0300": {{ID: deviceID}},
			},
			wantLen: 2,
		},
		{
			name:       "excludes rack group 01XX",
			cableGroup: "0900",
			devicesByGroup: map[string][]*devicetypes.CaniDeviceType{
				"0100": {{ID: deviceID}},
				"0200": {{ID: deviceID}},
			},
			wantLen: 1,
		},
		{
			name:       "excludes own cable group",
			cableGroup: "0900",
			devicesByGroup: map[string][]*devicetypes.CaniDeviceType{
				"0900": {{ID: deviceID}},
			},
			wantLen: 0,
		},
		{
			name:           "short cable group returns empty",
			cableGroup:     "1",
			devicesByGroup: map[string][]*devicetypes.CaniDeviceType{},
			wantLen:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findRelatedDeviceGroups(tt.cableGroup, tt.devicesByGroup)
			if len(got) != tt.wantLen {
				t.Errorf("findRelatedDeviceGroups() returned %d groups, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestTransformCablesEmpty(t *testing.T) {
	inv := &devicetypes.Inventory{
		Cables: make(map[uuid.UUID]*devicetypes.CaniCableType),
	}
	tally := &visual.StepTally{}
	recordNum := 0

	cables, err := transformCables(inv, nil, false, visual.ETLOptions{}, tally, &recordNum, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cables) != 0 {
		t.Errorf("expected 0 cables, got %d", len(cables))
	}
}

func TestTransformCablesProductRecords(t *testing.T) {
	inv := &devicetypes.Inventory{
		Cables: make(map[uuid.UUID]*devicetypes.CaniCableType),
	}
	tally := &visual.StepTally{}
	recordNum := 0

	records := []import_.CsvRecord{
		{
			PartNumber:  "C7536A",
			Description: "HPE Cat6 RJ45 M/M 2m",
			Quantity:    3,
		},
	}

	cables, err := transformCables(inv, records, false, visual.ETLOptions{}, tally, &recordNum, len(records))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cables) != 3 {
		t.Errorf("expected 3 cables, got %d", len(cables))
	}
}

func TestCreateCableFromExplicitRecord(t *testing.T) {
	srcDeviceID := uuid.New()
	dstDeviceID := uuid.New()
	srcIfaceID := uuid.New()
	dstIfaceID := uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			srcDeviceID: {
				ID:   srcDeviceID,
				Name: "switch-01",
				Interfaces: []devicetypes.InterfaceSpec{
					{ID: srcIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT},
				},
			},
			dstDeviceID: {
				ID:   dstDeviceID,
				Name: "server-01",
				Interfaces: []devicetypes.InterfaceSpec{
					{ID: dstIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT},
				},
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		rec := import_.CsvRecord{
			SourceDevice: "switch-01",
			SourcePort:   "eth0",
			DestDevice:   "server-01",
			DestPort:     "eth0",
		}

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

	t.Run("source device not found", func(t *testing.T) {
		rec := import_.CsvRecord{
			SourceDevice: "nonexistent",
			SourcePort:   "eth0",
			DestDevice:   "server-01",
			DestPort:     "eth0",
		}

		_, err := createCableFromExplicitRecord(inv, rec)
		if err == nil {
			t.Error("expected error but got none")
		}
	})

	t.Run("source port not found", func(t *testing.T) {
		rec := import_.CsvRecord{
			SourceDevice: "switch-01",
			SourcePort:   "nonexistent",
			DestDevice:   "server-01",
			DestPort:     "eth0",
		}

		_, err := createCableFromExplicitRecord(inv, rec)
		if err == nil {
			t.Error("expected error but got none")
		}
	})

	t.Run("dest device not found", func(t *testing.T) {
		rec := import_.CsvRecord{
			SourceDevice: "switch-01",
			SourcePort:   "eth0",
			DestDevice:   "nonexistent",
			DestPort:     "eth0",
		}

		_, err := createCableFromExplicitRecord(inv, rec)
		if err == nil {
			t.Error("expected error but got none")
		}
	})

	t.Run("dest port not found", func(t *testing.T) {
		rec := import_.CsvRecord{
			SourceDevice: "switch-01",
			SourcePort:   "eth0",
			DestDevice:   "server-01",
			DestPort:     "nonexistent",
		}

		_, err := createCableFromExplicitRecord(inv, rec)
		if err == nil {
			t.Error("expected error but got none")
		}
	})
}
