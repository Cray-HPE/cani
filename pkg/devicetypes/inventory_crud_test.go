package devicetypes

// Test coverage for inventory_crud.go (cascading remove & helpers)
//
// | Function               | Happy-path test                                    | Failure test                                        |
// |------------------------|----------------------------------------------------|-----------------------------------------------------|
// | unlinkDeviceFromParent | TestUnlinkDeviceFromParentRemovesChild             | TestUnlinkDeviceFromParentNilParent                 |
// | removeCablesForDevice  | TestRemoveCablesForDeviceDeletesMatching            | TestRemoveCablesForDeviceNoMatch                    |
// | removeModulesForDevice | TestRemoveModulesForDeviceDeletesMatching           | TestRemoveModulesForDeviceNoMatch                   |
// | RemoveDevice           | TestRemoveDeviceCascadesCleanup                    | TestRemoveDeviceNotFound                            |

import (
	"testing"

	"github.com/google/uuid"
)

// ---------- unlinkDeviceFromParent ----------

func TestUnlinkDeviceFromParentRemovesChild(t *testing.T) {
	inv := NewInventory()

	parentID := uuid.New()
	childID := uuid.New()

	inv.Devices[parentID] = &CaniDeviceType{
		ID:       parentID,
		Name:     "parent-device",
		Children: []uuid.UUID{childID},
	}
	inv.Devices[childID] = &CaniDeviceType{
		ID:     childID,
		Name:   "child-device",
		Parent: parentID,
	}

	inv.unlinkDeviceFromParent(inv.Devices[childID], childID)

	if containsUUID(inv.Devices[parentID].Children, childID) {
		t.Error("expected child to be removed from parent's Children list")
	}
}

func TestUnlinkDeviceFromParentNilParent(t *testing.T) {
	inv := NewInventory()

	deviceID := uuid.New()
	device := &CaniDeviceType{
		ID:     deviceID,
		Name:   "orphan-device",
		Parent: uuid.Nil,
	}
	inv.Devices[deviceID] = device

	// Should return immediately without panic when Parent is nil UUID.
	inv.unlinkDeviceFromParent(device, deviceID)

	if _, ok := inv.Devices[deviceID]; !ok {
		t.Error("device should still exist in inventory")
	}
}

// ---------- removeCablesForDevice ----------

func TestRemoveCablesForDeviceDeletesMatching(t *testing.T) {
	inv := NewInventory()

	deviceID := uuid.New()
	otherID := uuid.New()
	cableAID := uuid.New()
	cableBID := uuid.New()
	cableKeepID := uuid.New()

	inv.Cables[cableAID] = &CaniCableType{
		ID:                 cableAID,
		Label:              "cable-a",
		TerminationADevice: deviceID,
		TerminationBDevice: otherID,
	}
	inv.Cables[cableBID] = &CaniCableType{
		ID:                 cableBID,
		Label:              "cable-b",
		TerminationADevice: otherID,
		TerminationBDevice: deviceID,
	}
	inv.Cables[cableKeepID] = &CaniCableType{
		ID:                 cableKeepID,
		Label:              "cable-keep",
		TerminationADevice: otherID,
		TerminationBDevice: otherID,
	}

	inv.removeCablesForDevice(deviceID)

	if _, ok := inv.Cables[cableAID]; ok {
		t.Error("cable-a should have been deleted (terminationA matched)")
	}
	if _, ok := inv.Cables[cableBID]; ok {
		t.Error("cable-b should have been deleted (terminationB matched)")
	}
	if _, ok := inv.Cables[cableKeepID]; !ok {
		t.Error("cable-keep should remain (no termination matched)")
	}
}

func TestRemoveCablesForDeviceNoMatch(t *testing.T) {
	inv := NewInventory()

	otherID := uuid.New()
	cableID := uuid.New()

	inv.Cables[cableID] = &CaniCableType{
		ID:                 cableID,
		Label:              "unrelated",
		TerminationADevice: otherID,
		TerminationBDevice: otherID,
	}

	inv.removeCablesForDevice(uuid.New())

	if len(inv.Cables) != 1 {
		t.Errorf("expected 1 cable to remain, got %d", len(inv.Cables))
	}
}

// ---------- removeModulesForDevice ----------

func TestRemoveModulesForDeviceDeletesMatching(t *testing.T) {
	inv := NewInventory()

	deviceID := uuid.New()
	modMatchID := uuid.New()
	modKeepID := uuid.New()

	inv.Modules[modMatchID] = &CaniModuleType{
		ID:           modMatchID,
		Name:         "mod-match",
		ParentDevice: deviceID,
	}
	inv.Modules[modKeepID] = &CaniModuleType{
		ID:           modKeepID,
		Name:         "mod-keep",
		ParentDevice: uuid.New(),
	}

	inv.removeModulesForDevice(deviceID)

	if _, ok := inv.Modules[modMatchID]; ok {
		t.Error("matching module should have been deleted")
	}
	if _, ok := inv.Modules[modKeepID]; !ok {
		t.Error("unrelated module should remain")
	}
}

func TestRemoveModulesForDeviceNoMatch(t *testing.T) {
	inv := NewInventory()

	modID := uuid.New()
	inv.Modules[modID] = &CaniModuleType{
		ID:           modID,
		Name:         "some-module",
		ParentDevice: uuid.New(),
	}

	inv.removeModulesForDevice(uuid.New())

	if len(inv.Modules) != 1 {
		t.Errorf("expected 1 module to remain, got %d", len(inv.Modules))
	}
}

// ---------- RemoveDevice ----------

func TestRemoveDeviceCascadesCleanup(t *testing.T) {
	inv := NewInventory()

	rackID := uuid.New()
	deviceID := uuid.New()
	childID := uuid.New()
	cableID := uuid.New()
	modID := uuid.New()

	inv.Racks[rackID] = &CaniRackType{
		ID:      rackID,
		Name:    "rack-1",
		UHeight: 42,
		Devices: []uuid.UUID{deviceID},
	}
	inv.Devices[deviceID] = &CaniDeviceType{
		ID:       deviceID,
		Name:     "device-1",
		Parent:   rackID,
		Children: []uuid.UUID{childID},
	}
	inv.Devices[childID] = &CaniDeviceType{
		ID:     childID,
		Name:   "child-1",
		Parent: deviceID,
	}
	inv.Cables[cableID] = &CaniCableType{
		ID:                 cableID,
		Label:              "cable-1",
		TerminationADevice: deviceID,
	}
	inv.Modules[modID] = &CaniModuleType{
		ID:           modID,
		Name:         "mod-1",
		ParentDevice: deviceID,
	}

	err := inv.RemoveDevice(deviceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := inv.Devices[deviceID]; ok {
		t.Error("device should have been deleted")
	}
	if _, ok := inv.Devices[childID]; ok {
		t.Error("child device should have been recursively deleted")
	}
	if _, ok := inv.Cables[cableID]; ok {
		t.Error("cable referencing device should have been deleted")
	}
	if _, ok := inv.Modules[modID]; ok {
		t.Error("module belonging to device should have been deleted")
	}
	if containsUUID(inv.Racks[rackID].Devices, deviceID) {
		t.Error("device should have been removed from rack's device list")
	}
}

func TestRemoveDeviceNotFound(t *testing.T) {
	inv := NewInventory()

	err := inv.RemoveDevice(uuid.New())
	if err == nil {
		t.Fatal("expected error when removing non-existent device")
	}
}
