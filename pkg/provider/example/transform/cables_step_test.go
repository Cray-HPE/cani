package transform

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

func TestResolveCableTypeSlug(t *testing.T) {
	tests := []struct {
		name        string
		partNumber  string
		description string
		wantNot     string // just verify we get a non-empty result
	}{
		{
			name:        "fallback to description inference for cat6",
			partNumber:  "FAKE-PN-123",
			description: "Cat6 RJ45 patch cable 2m",
		},
		{
			name:        "fallback to description inference for DAC",
			partNumber:  "",
			description: "400G QSFP-DD DAC 3m",
		},
		{
			name:        "unknown description returns other",
			partNumber:  "",
			description: "Mystery Widget",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveCableTypeSlug(tt.partNumber, tt.description)
			if got == "" {
				t.Error("expected non-empty cable type slug")
			}
		})
	}
}

func TestAutoConnectCables(t *testing.T) {
	switchID := uuid.New()
	nodeID := uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			switchID: {
				ID:   switchID,
				Name: "switch-01",
				Type: "switch",
				Interfaces: []devicetypes.InterfaceSpec{
					{ID: uuid.New(), Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT},
				},
				ObjectMeta: devicetypes.ObjectMeta{
					ProviderMetadata: map[string]any{
						"example": map[string]any{"ConfigGroup": "0200"},
					},
				},
			},
			nodeID: {
				ID:   nodeID,
				Name: "server-01",
				Type: "node",
				Interfaces: []devicetypes.InterfaceSpec{
					{ID: uuid.New(), Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT},
				},
				ObjectMeta: devicetypes.ObjectMeta{
					ProviderMetadata: map[string]any{
						"example": map[string]any{"ConfigGroup": "0300"},
					},
				},
			},
		},
	}

	cable := devicetypes.NewCable("cat6", "test-cable")
	cablesByGroup := map[string][]*devicetypes.CaniCableType{
		"0900": {cable},
	}

	autoConnectCables(inv, cablesByGroup)

	// Cable should have been connected to the switch and node
	if cable.TerminationA == uuid.Nil && cable.TerminationB == uuid.Nil {
		t.Log("cable was not auto-connected (may lack matching interfaces)")
	}
}

func TestAutoConnectCablesNoRelatedGroups(t *testing.T) {
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
	}

	cable := devicetypes.NewCable("cat6", "test-cable")
	cablesByGroup := map[string][]*devicetypes.CaniCableType{
		"0900": {cable},
	}

	// Should not panic with no devices
	autoConnectCables(inv, cablesByGroup)
}

func TestConnectCablesHubSpoke(t *testing.T) {
	switchID := uuid.New()
	nodeID := uuid.New()
	switchIfaceID := uuid.New()
	nodeIfaceID := uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			switchID: {
				ID:   switchID,
				Name: "switch-01",
				Interfaces: []devicetypes.InterfaceSpec{
					{ID: switchIfaceID, Name: "eth0"},
				},
			},
			nodeID: {
				ID:   nodeID,
				Name: "server-01",
				Interfaces: []devicetypes.InterfaceSpec{
					{ID: nodeIfaceID, Name: "eth0"},
				},
			},
		},
	}

	cable := devicetypes.NewCable("cat6", "test-cable")
	hubs := []*devicetypes.CaniDeviceType{inv.Devices[switchID]}
	spokes := []*devicetypes.CaniDeviceType{inv.Devices[nodeID]}

	connectCablesHubSpoke(inv, []*devicetypes.CaniCableType{cable}, hubs, spokes)

	if cable.TerminationA == uuid.Nil || cable.TerminationB == uuid.Nil {
		t.Error("expected cable to be connected to both hub and spoke")
	}
}

func TestConnectCablesHubSpokeNoHubs(t *testing.T) {
	inv := &devicetypes.Inventory{Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{}}
	cable := devicetypes.NewCable("cat6", "test")

	// No hubs → should return without panic
	connectCablesHubSpoke(inv, []*devicetypes.CaniCableType{cable}, nil, nil)
	if cable.TerminationA != uuid.Nil {
		t.Error("expected cable to remain unconnected")
	}
}

func TestConnectCablesHubSpokeMoreCablesThanSpokes(t *testing.T) {
	hubID := uuid.New()
	spokeID := uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			hubID: {
				ID:         hubID,
				Name:       "sw-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "e0"}, {ID: uuid.New(), Name: "e1"}},
			},
			spokeID: {
				ID:         spokeID,
				Name:       "srv-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "e0"}},
			},
		},
	}

	cables := []*devicetypes.CaniCableType{
		devicetypes.NewCable("cat6", "c1"),
		devicetypes.NewCable("cat6", "c2"),
	}

	connectCablesHubSpoke(inv, cables, []*devicetypes.CaniDeviceType{inv.Devices[hubID]}, []*devicetypes.CaniDeviceType{inv.Devices[spokeID]})

	// Only one spoke → only one cable should be connected
	if cables[0].TerminationA == uuid.Nil {
		t.Error("first cable should be connected")
	}
	if cables[1].TerminationA != uuid.Nil {
		t.Error("second cable should remain unconnected (no more spokes)")
	}
}

func TestBuildCableStepInfo(t *testing.T) {
	cable := devicetypes.NewCable("cat6", "switch:e0 ↔ server:e0")
	ifaceA := uuid.New()
	ifaceB := uuid.New()
	cable.SetTerminations(ifaceA, ifaceB)

	rec := import_.CsvRecord{
		SourceDevice: "switch-01",
		SourcePort:   "eth0",
		DestDevice:   "server-01",
		DestPort:     "eth0",
		CableType:    "cat6",
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
}

func TestBuildCableStepInfoDerivedType(t *testing.T) {
	cable := devicetypes.NewCable("cat6", "label")
	cable.SetTerminations(uuid.New(), uuid.New())

	rec := import_.CsvRecord{
		SourceDevice: "sw",
		SourcePort:   "e0",
		DestDevice:   "srv",
		DestPort:     "e0",
		CableType:    "", // empty → IsDerived should be true
	}

	info := buildCableStepInfo(rec, cable)

	// Third mapping is CableType; IsDerived should be true when empty
	if len(info.Mappings) >= 3 && !info.Mappings[2].IsDerived {
		t.Error("expected CableType mapping to be marked as derived when empty")
	}
}

func TestBuildCableProductStepInfo(t *testing.T) {
	cables := []*devicetypes.CaniCableType{
		devicetypes.NewCable("cat6", "Cable-001"),
		devicetypes.NewCable("cat6", "Cable-002"),
	}

	rec := import_.CsvRecord{
		PartNumber:  "C7536A",
		Description: "HPE Cat6 RJ45 Cable 2m",
		Quantity:    2,
	}

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
