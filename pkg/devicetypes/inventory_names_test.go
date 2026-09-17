package devicetypes

import (
	"testing"

	"github.com/google/uuid"
)

// TestEnsureUniqueDeviceNamesReservesExistingSuffixes verifies generated names
// never replace names already present in the transform result.
//
// Why it matters: the merge matches by name and would collapse distinct devices.
// Inputs: two "node" devices plus existing "node-1" and "node-2" devices.
// Outputs: every UUID keeps a unique name, including after a second call.
// Data choice: consecutive occupied suffixes require skipping more than one name.
func TestEnsureUniqueDeviceNamesReservesExistingSuffixes(t *testing.T) {
	firstID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	secondID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	thirdID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	fourthID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	result := &TransformResult{Devices: map[uuid.UUID]*CaniDeviceType{
		firstID:  {ID: firstID, Name: "node"},
		secondID: {ID: secondID, Name: "node"},
		thirdID:  {ID: thirdID, Name: "node-1"},
		fourthID: {ID: fourthID, Name: "node-2"},
	}}
	want := map[uuid.UUID]string{firstID: "node-3", secondID: "node-4", thirdID: "node-1", fourthID: "node-2"}

	result.EnsureUniqueDeviceNames()
	result.EnsureUniqueDeviceNames()

	for deviceID, name := range want {
		if got := result.Devices[deviceID].Name; got != name {
			t.Errorf("device %s name = %q, want %q", deviceID, got, name)
		}
	}
}

// TestMergeTransformResultPreservesSuffixCollisionDevices verifies name
// allocation cannot merge separate incoming UUIDs onto the same device.
//
// Why it matters: a successful transform must not silently discard hardware.
// Inputs: three devices named "node", "node", and "node-1" with distinct serials.
// Outputs: all original UUIDs and serials survive, with identity remaps unchanged.
// Data choice: the pre-suffixed name reproduces the smallest data-loss case.
func TestMergeTransformResultPreservesSuffixCollisionDevices(t *testing.T) {
	inventory := NewInventory()
	result := &TransformResult{Devices: make(map[uuid.UUID]*CaniDeviceType)}
	for _, device := range []CaniDeviceType{
		{Name: "node", Serial: "first"},
		{Name: "node", Serial: "second"},
		{Name: "node-1", Serial: "third"},
	} {
		device.ID = uuid.New()
		result.Devices[device.ID] = &device
	}

	summary, err := inventory.MergeTransformResult(result)
	if err != nil {
		t.Fatalf("MergeTransformResult() error = %v", err)
	}
	if len(inventory.Devices) != len(result.Devices) {
		t.Fatalf("retained %d devices, want %d", len(inventory.Devices), len(result.Devices))
	}
	for deviceID, original := range result.Devices {
		retained := inventory.Devices[deviceID]
		if retained == nil || retained.Serial != original.Serial || summary.Remaps.Devices[deviceID] != deviceID {
			t.Errorf("device %s lost its identity or serial: retained=%+v remap=%s", deviceID, retained, summary.Remaps.Devices[deviceID])
		}
	}
}
