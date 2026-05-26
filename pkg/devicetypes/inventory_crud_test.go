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

// ---------- EnsureLocation ----------

func TestEnsureLocationCreatesDefault(t *testing.T) {
	inv := NewInventory()

	locID := inv.EnsureLocation()
	if locID == uuid.Nil {
		t.Fatal("expected non-nil UUID from EnsureLocation")
	}
	if len(inv.Locations) != 1 {
		t.Errorf("expected 1 location, got %d", len(inv.Locations))
	}
}

func TestEnsureLocationIdempotent(t *testing.T) {
	inv := NewInventory()

	first := inv.EnsureLocation()
	second := inv.EnsureLocation()

	if first != second {
		t.Errorf("EnsureLocation not idempotent: first=%s, second=%s", first, second)
	}
	if len(inv.Locations) != 1 {
		t.Errorf("expected 1 location after two calls, got %d", len(inv.Locations))
	}
}

// ---------- AssignRacksToLocation ----------

func TestAssignRacksToLocationLinksOrphans(t *testing.T) {
	inv := NewInventory()

	locID := uuid.New()
	inv.Locations[locID] = &CaniLocationType{ID: locID, Name: "site-1"}

	orphanID := uuid.New()
	inv.Racks[orphanID] = &CaniRackType{ID: orphanID, Name: "orphan-rack"}

	inv.AssignRacksToLocation(locID)

	if inv.Racks[orphanID].Location != locID {
		t.Errorf("expected rack Location = %s, got %s", locID, inv.Racks[orphanID].Location)
	}
}

func TestAssignRacksToLocationSkipsAssigned(t *testing.T) {
	inv := NewInventory()

	loc1 := uuid.New()
	loc2 := uuid.New()
	inv.Locations[loc1] = &CaniLocationType{ID: loc1, Name: "site-1"}
	inv.Locations[loc2] = &CaniLocationType{ID: loc2, Name: "site-2"}

	rackID := uuid.New()
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "assigned-rack", Location: loc1}

	inv.AssignRacksToLocation(loc2)

	if inv.Racks[rackID].Location != loc1 {
		t.Error("assigned rack should not have been reassigned")
	}
}

// ---------- AddDevices ----------

func TestAddDevicesValid(t *testing.T) {
	inv := NewInventory()

	id := uuid.New()
	batch := map[uuid.UUID]*CaniDeviceType{
		id: {ID: id, Name: "server-new", Slug: "test-slug"},
	}

	err := inv.AddDevices(batch)
	if err != nil {
		t.Fatalf("AddDevices() unexpected error: %v", err)
	}
	if _, ok := inv.Devices[id]; !ok {
		t.Error("expected device to be present after AddDevices")
	}
}

func TestAddDevicesDuplicateUUID(t *testing.T) {
	inv := NewInventory()

	id := uuid.New()
	inv.Devices[id] = &CaniDeviceType{ID: id, Name: "existing"}

	batch := map[uuid.UUID]*CaniDeviceType{
		id: {ID: id, Name: "duplicate"},
	}

	err := inv.AddDevices(batch)
	if err == nil {
		t.Error("AddDevices with duplicate UUID should return an error")
	}
}

// ---------- MergeDevices ----------

func TestMergeDevicesByUUID(t *testing.T) {
	inv := NewInventory()

	id := uuid.New()
	inv.Devices[id] = &CaniDeviceType{
		ID:   id,
		Name: "server-1",
		Slug: "old-slug",
	}

	incoming := map[uuid.UUID]*CaniDeviceType{
		id: {ID: id, Name: "server-1", Slug: "new-slug", Serial: "SN-123"},
	}

	inv.MergeDevices(incoming)

	if inv.Devices[id].Serial != "SN-123" {
		t.Errorf("expected Serial to be merged, got %q", inv.Devices[id].Serial)
	}
}

func TestMergeDevicesByName(t *testing.T) {
	inv := NewInventory()

	existingID := uuid.New()
	inv.Devices[existingID] = &CaniDeviceType{
		ID:   existingID,
		Name: "server-named",
		Slug: "existing-slug",
	}

	incomingID := uuid.New()
	incoming := map[uuid.UUID]*CaniDeviceType{
		incomingID: {ID: incomingID, Name: "server-named", Serial: "SN-NEW"},
	}

	inv.MergeDevices(incoming)

	// The existing device should have the merged serial.
	if inv.Devices[existingID].Serial != "SN-NEW" {
		t.Errorf("expected Serial merged by name, got %q", inv.Devices[existingID].Serial)
	}
}

// ---------- MergeRacks ----------

func TestMergeRacksByName(t *testing.T) {
	inv := NewInventory()

	existingID := uuid.New()
	inv.Racks[existingID] = &CaniRackType{
		ID:   existingID,
		Name: "rack-shared",
	}

	incomingID := uuid.New()
	incoming := map[uuid.UUID]*CaniRackType{
		incomingID: {ID: incomingID, Name: "rack-shared", UHeight: 48},
	}

	remap := inv.MergeRacks(incoming)

	if remap[incomingID] != existingID {
		t.Errorf("expected remap[%s] = %s, got %s", incomingID, existingID, remap[incomingID])
	}
	if inv.Racks[existingID].UHeight != 48 {
		t.Errorf("expected UHeight merged to 48, got %d", inv.Racks[existingID].UHeight)
	}
}

func TestMergeRacksNewInsert(t *testing.T) {
	inv := NewInventory()

	id := uuid.New()
	incoming := map[uuid.UUID]*CaniRackType{
		id: {ID: id, Name: "brand-new-rack", UHeight: 42},
	}

	remap := inv.MergeRacks(incoming)

	if len(remap) != 0 {
		t.Errorf("expected empty remap for new insert, got %v", remap)
	}
	if _, ok := inv.Racks[id]; !ok {
		t.Error("expected new rack in inventory after MergeRacks")
	}
}

// ---------- MergeLocations ----------

func TestMergeLocationsRemap(t *testing.T) {
	inv := NewInventory()

	existingID := uuid.New()
	inv.Locations[existingID] = &CaniLocationType{
		ID:   existingID,
		Name: "site-shared",
	}

	incomingID := uuid.New()
	incoming := map[uuid.UUID]*CaniLocationType{
		incomingID: {ID: incomingID, Name: "site-shared", LocationType: "building"},
	}

	remap := inv.MergeLocations(incoming)

	if remap[incomingID] != existingID {
		t.Errorf("expected remap[%s] = %s, got %s", incomingID, existingID, remap[incomingID])
	}
	if inv.Locations[existingID].LocationType != "building" {
		t.Errorf("expected LocationType merged, got %q", inv.Locations[existingID].LocationType)
	}
}

func TestMergeLocationsNewInsert(t *testing.T) {
	inv := NewInventory()

	id := uuid.New()
	incoming := map[uuid.UUID]*CaniLocationType{
		id: {ID: id, Name: "new-site", LocationType: "site"},
	}

	remap := inv.MergeLocations(incoming)

	if len(remap) != 0 {
		t.Errorf("expected empty remap for new insert, got %v", remap)
	}
	if _, ok := inv.Locations[id]; !ok {
		t.Error("expected new location in inventory after MergeLocations")
	}
}

// ---------- providerIdentityCompatible ----------

func TestProviderIdentityCompatibleMatch(t *testing.T) {
	a := &CaniDeviceType{
		Name: "server-1",
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{"bmc_fqdn": "bmc1.example.com"},
		}},
	}
	b := &CaniDeviceType{
		Name: "server-1",
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{"bmc_fqdn": "bmc1.example.com"},
		}},
	}

	if !providerIdentityCompatible(a, b) {
		t.Error("expected compatible when bmc_fqdn values match")
	}
}

func TestProviderIdentityCompatibleConflict(t *testing.T) {
	a := &CaniDeviceType{
		Name: "server-1",
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{"bmc_fqdn": "bmc1.example.com"},
		}},
	}
	b := &CaniDeviceType{
		Name: "server-1",
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{"bmc_fqdn": "bmc2.example.com"},
		}},
	}

	if providerIdentityCompatible(a, b) {
		t.Error("expected incompatible when bmc_fqdn values differ")
	}
}

// ---------- mergeRackProperties ----------

func TestMergeRackPropertiesCopiesNonEmpty(t *testing.T) {
	existing := &CaniRackType{
		ID:   uuid.New(),
		Name: "rack-existing",
	}
	incoming := &CaniRackType{
		UHeight:      48,
		Manufacturer: "HPE",
		Model:        "G2",
		Serial:       "SN-RACK",
	}

	mergeRackProperties(existing, incoming)

	if existing.UHeight != 48 {
		t.Errorf("UHeight = %d, want 48", existing.UHeight)
	}
	if existing.Manufacturer != "HPE" {
		t.Errorf("Manufacturer = %q, want HPE", existing.Manufacturer)
	}
	if existing.Model != "G2" {
		t.Errorf("Model = %q, want G2", existing.Model)
	}
	if existing.Serial != "SN-RACK" {
		t.Errorf("Serial = %q, want SN-RACK", existing.Serial)
	}
}
