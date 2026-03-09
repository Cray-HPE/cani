package transform

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func TestAssignRackPositions_SwitchesAtTop(t *testing.T) {
	rackID := uuid.New()
	sw1ID := uuid.New()
	sw2ID := uuid.New()
	sw3ID := uuid.New()

	result := &devicetypes.TransformResult{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			sw1ID: {
				ID: sw1ID, Name: "sw-leaf-bmc-001", Type: devicetypes.TypeMgmtSwitch,
				UHeight:          1,
				ProviderMetadata: map[string]any{"csm": map[string]any{"xname": "x3000c0w1"}},
			},
			sw2ID: {
				ID: sw2ID, Name: "sw-leaf-bmc-002", Type: devicetypes.TypeMgmtSwitch,
				UHeight:          1,
				ProviderMetadata: map[string]any{"csm": map[string]any{"xname": "x3000c0w2"}},
			},
			sw3ID: {
				ID: sw3ID, Name: "sw-spine-001", Type: devicetypes.TypeSwitch,
				UHeight:          1,
				ProviderMetadata: map[string]any{"csm": map[string]any{"xname": "x3000c0h1s1"}},
			},
		},
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {ID: rackID, Name: "x3000", UHeight: 42},
		},
	}

	assignRackPositions(result)

	// Sorted by xname: h1s1 < w1 < w2 → positions 42, 41, 40.
	cases := []struct {
		name string
		id   uuid.UUID
		want int
	}{
		{"sw-spine-001 (h1s1)", sw3ID, 42},
		{"sw-leaf-bmc-001 (w1)", sw1ID, 41},
		{"sw-leaf-bmc-002 (w2)", sw2ID, 40},
	}
	for _, tc := range cases {
		got := result.Devices[tc.id].RackPosition
		if got != tc.want {
			t.Errorf("%s: RackPosition = %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestAssignRackPositions_BladesAtBottom(t *testing.T) {
	rackID := uuid.New()
	b1 := uuid.New()
	b2 := uuid.New()
	b3 := uuid.New()

	result := &devicetypes.TransformResult{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			b1: {
				ID: b1, Name: "x3000c0s1", Type: devicetypes.TypeBlade,
				UHeight:          2,
				ProviderMetadata: map[string]any{"csm": map[string]any{"xname": "x3000c0s1"}},
			},
			b2: {
				ID: b2, Name: "x3000c0s2", Type: devicetypes.TypeBlade,
				UHeight:          2,
				ProviderMetadata: map[string]any{"csm": map[string]any{"xname": "x3000c0s2"}},
			},
			b3: {
				ID: b3, Name: "x3000c0s3", Type: devicetypes.TypeBlade,
				UHeight:          2,
				ProviderMetadata: map[string]any{"csm": map[string]any{"xname": "x3000c0s3"}},
			},
		},
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {ID: rackID, Name: "x3000", UHeight: 42},
		},
	}

	assignRackPositions(result)

	// Blades should start from U1 going up: 1, 3, 5.
	cases := []struct {
		name string
		id   uuid.UUID
		want int
	}{
		{"blade s1", b1, 1},
		{"blade s2", b2, 3},
		{"blade s3", b3, 5},
	}
	for _, tc := range cases {
		got := result.Devices[tc.id].RackPosition
		if got != tc.want {
			t.Errorf("%s: RackPosition = %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestAssignRackPositions_NoPositionForInternal(t *testing.T) {
	rackID := uuid.New()
	nodeID := uuid.New()
	ncID := uuid.New()

	result := &devicetypes.TransformResult{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			nodeID: {
				ID: nodeID, Name: "x3000c0s1b0n0", Type: devicetypes.TypeNode,
				ProviderMetadata: map[string]any{"csm": map[string]any{"xname": "x3000c0s1b0n0"}},
			},
			ncID: {
				ID: ncID, Name: "x3000c0s1b0", Type: devicetypes.TypeNodeCard,
				ProviderMetadata: map[string]any{"csm": map[string]any{"xname": "x3000c0s1b0"}},
			},
		},
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {ID: rackID, Name: "x3000", UHeight: 42},
		},
	}

	assignRackPositions(result)

	if result.Devices[nodeID].RackPosition != 0 {
		t.Errorf("node should not have RackPosition, got %d", result.Devices[nodeID].RackPosition)
	}
	if result.Devices[ncID].RackPosition != 0 {
		t.Errorf("nodecard should not have RackPosition, got %d", result.Devices[ncID].RackPosition)
	}
}

func TestCabinetForDevice_ExtractsCabinet(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		ProviderMetadata: map[string]any{
			"csm": map[string]any{"xname": "x3000c0s5b0"},
		},
	}
	got := cabinetForDevice(dev)
	if got != "x3000" {
		t.Errorf("cabinetForDevice = %q, want %q", got, "x3000")
	}
}

func TestCabinetForDevice_NoMetadata(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{}
	got := cabinetForDevice(dev)
	if got != "" {
		t.Errorf("cabinetForDevice = %q, want empty", got)
	}
}
