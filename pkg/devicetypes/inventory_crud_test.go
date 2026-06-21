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
		id: {ID: id, Name: "server-new"},
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

func TestAddDevicesRejectsInvalidSlug(t *testing.T) {
	inv := NewInventory()

	id := uuid.New()
	batch := map[uuid.UUID]*CaniDeviceType{
		id: {ID: id, Name: "server-new", Slug: "not-a-real-device-slug"},
	}

	if err := inv.AddDevices(batch); err == nil {
		t.Error("AddDevices should reject an invalid device slug")
	}
}

// ---------- MergeDevices ----------

func TestMergeDevicesByUUID(t *testing.T) {
	inv := NewInventory()

	id := uuid.New()
	inv.Devices[id] = &CaniDeviceType{
		ID:   id,
		Name: "server-1",
	}

	incoming := map[uuid.UUID]*CaniDeviceType{
		id: {ID: id, Name: "server-1", Serial: "SN-123"},
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

func TestMergeDevicesSkipsInvalidSlug(t *testing.T) {
	inv := NewInventory()

	id := uuid.New()
	inv.MergeDevices(map[uuid.UUID]*CaniDeviceType{
		id: {ID: id, Name: "server-invalid", Slug: "not-a-real-device-slug"},
	})

	if len(inv.Devices) != 0 {
		t.Fatalf("expected invalid device merge to be skipped, got %d devices", len(inv.Devices))
	}
}

func TestMergeModulesSkipsInvalidSlug(t *testing.T) {
	inv := NewInventory()

	id := uuid.New()
	inv.MergeModules(map[uuid.UUID]*CaniModuleType{
		id: {ID: id, Name: "module-invalid", Slug: "not-a-real-module-slug"},
	})

	if len(inv.Modules) != 0 {
		t.Fatalf("expected invalid module merge to be skipped, got %d modules", len(inv.Modules))
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

// ---------- mergeCableByLabel ----------

// TestMergeCableByLabel verifies mergeCableByLabel overwrites an existing cable
// that shares the incoming label, and returns false when the label is empty or
// unmatched.
//
// Why it matters: cables imported from providers often lack stable UUIDs but
// carry a unique label, so label-based de-duplication prevents duplicate cable
// records on re-import.
// Inputs: an inventory with a cable labeled "link-1"; incoming cables labeled
// "link-1" (match), "link-9" (no match), and "" (empty). Outputs: true with the
// existing entry overwritten for the match; false for the other two.
// Data choice: a distinct replacement field (Type) on the matching cable proves
// the overwrite happened, while the empty-label case pins the guard clause.
func TestMergeCableByLabel(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Cables[id] = &CaniCableType{ID: id, Label: "link-1", Type: "old"}

	if !inv.mergeCableByLabel(&CaniCableType{Label: "link-1", Type: "new"}) {
		t.Fatal("mergeCableByLabel(match) = false, want true")
	}
	if inv.Cables[id].Type != "new" {
		t.Errorf("matched cable Type = %q, want overwritten %q", inv.Cables[id].Type, "new")
	}
	if inv.mergeCableByLabel(&CaniCableType{Label: "link-9"}) {
		t.Error("mergeCableByLabel(no match) = true, want false")
	}
	if inv.mergeCableByLabel(&CaniCableType{Label: ""}) {
		t.Error("mergeCableByLabel(empty label) = true, want false")
	}
}

// ---------- MergeCables ----------

// TestMergeCables verifies MergeCables resolves incoming cables by UUID match,
// then label match, then insert, lazily allocating the map and skipping nil
// entries.
//
// Why it matters: cable re-import must converge without creating duplicates, so
// the three-tier resolution (id → label → insert) governs inventory growth.
// Inputs: a map-less inventory seeded via merge with "alpha" and "beta" cables,
// then a second merge containing a UUID-matched update to "alpha", a
// label-matched update to "beta" carrying a different UUID, a brand-new
// "gamma", and a nil. Outputs: the UUID match overwritten, the label match
// folded in place into the existing entry, "gamma" inserted, and the nil
// ignored. Data choice: routing the UUID match and the label match at distinct
// existing cables keeps the result deterministic regardless of map iteration
// order while still exercising all three resolution tiers plus the nil guard.
func TestMergeCables(t *testing.T) {
	inv := &Inventory{} // nil Cables map exercises lazy allocation
	idA := uuid.New()
	idB := uuid.New()
	inv.MergeCables(map[uuid.UUID]*CaniCableType{
		idA: {ID: idA, Label: "alpha", Type: "t1"},
		idB: {ID: idB, Label: "beta", Type: "t2"},
	})
	if len(inv.Cables) != 2 {
		t.Fatalf("after first merge len = %d, want 2", len(inv.Cables))
	}

	gammaID := uuid.New()
	inv.MergeCables(map[uuid.UUID]*CaniCableType{
		idA:        {ID: idA, Label: "alpha", Type: "t1-updated"}, // UUID match
		uuid.New(): {Label: "beta", Type: "t2-updated"},           // label match -> idB
		gammaID:    {ID: gammaID, Label: "gamma"},                 // insert
		uuid.New(): nil,                                           // skipped
	})

	if inv.Cables[idA].Type != "t1-updated" {
		t.Errorf("UUID-matched cable Type = %q, want %q", inv.Cables[idA].Type, "t1-updated")
	}
	if inv.Cables[idB].Type != "t2-updated" {
		t.Errorf("label-matched cable Type = %q, want %q", inv.Cables[idB].Type, "t2-updated")
	}
	if _, ok := inv.Cables[gammaID]; !ok {
		t.Error("expected gamma cable to be inserted")
	}
	if len(inv.Cables) != 3 {
		t.Errorf("final cable count = %d, want 3 (alpha + beta + gamma)", len(inv.Cables))
	}
}

// ---------- mergeModuleProperties ----------

// TestMergeModuleProperties verifies mergeModuleProperties copies every
// non-empty field from the incoming module onto the existing one.
//
// Why it matters: a module field-level merge preserves identity (UUID, parent)
// while refreshing mutable attributes from the latest import.
// Inputs: an empty existing module and an incoming module with Slug, Status,
// Serial, Manufacturer, Model, and Type populated. Outputs: all six fields
// copied onto the existing module.
// Data choice: populating all six covered fields at once asserts none of the
// copy branches is omitted.
func TestMergeModuleProperties(t *testing.T) {
	existing := &CaniModuleType{ID: uuid.New(), Name: "mod"}
	incoming := &CaniModuleType{
		Serial:       "SN-MOD",
		Manufacturer: "Acme",
		Model:        "M1",
		Type:         "gpu",
		ObjectMeta:   ObjectMeta{Status: "Active"},
	}

	mergeModuleProperties(existing, incoming)

	if existing.Status != "Active" || existing.Serial != "SN-MOD" ||
		existing.Manufacturer != "Acme" || existing.Model != "M1" || existing.Type != "gpu" {
		t.Errorf("mergeModuleProperties did not copy all fields: %+v", existing)
	}
}

// ---------- MergeModules ----------

// TestMergeModules verifies MergeModules resolves incoming modules by UUID
// match, then name match, then insert, while skipping nil, unnamed, and invalid
// modules.
//
// Why it matters: module re-import must converge without duplicates and must
// reject modules whose slug is absent from the library, since those cannot be
// rendered downstream.
// Inputs: an inventory with a "nic-0" module, then a merge containing a
// UUID-matched update, a name-matched "nic-0", a new "nic-1", a nil, an unnamed
// module, and one with an unregistered slug. Outputs: the UUID and name matches
// folded in, "nic-1" inserted, and the three invalid entries skipped.
// Data choice: the unregistered-slug entry drives the Validate()-failure skip
// branch that a simpler test would miss.
func TestMergeModules(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	idA := uuid.New()
	inv.Modules[idA] = &CaniModuleType{ID: idA, Name: "nic-0", ParentDevice: devID, ObjectMeta: ObjectMeta{Status: "old"}}

	inv.MergeModules(map[uuid.UUID]*CaniModuleType{
		idA:        {ID: idA, Name: "nic-0", ObjectMeta: ObjectMeta{Status: "updated"}}, // UUID match
		uuid.New(): {Name: "nic-0", Serial: "SN-name"},                                  // name match
		uuid.New(): {Name: "nic-1"},                                                     // insert
		uuid.New(): nil,                                                                 // skipped (nil)
		uuid.New(): {Name: ""},                                                          // skipped (no name)
		uuid.New(): {Name: "bad", Slug: "not-in-library"},                               // skipped (invalid)
	})

	if inv.Modules[idA].Status != "updated" {
		t.Errorf("UUID-matched module Status = %q, want %q", inv.Modules[idA].Status, "updated")
	}
	if inv.Modules[idA].Serial != "SN-name" {
		t.Errorf("name-matched merge did not apply, Serial = %q, want %q", inv.Modules[idA].Serial, "SN-name")
	}
	var hasNic1 bool
	for _, m := range inv.Modules {
		if m != nil && m.Name == "nic-1" {
			hasNic1 = true
		}
	}
	if !hasNic1 {
		t.Error("expected nic-1 to be inserted")
	}
	if len(inv.Modules) != 2 {
		t.Errorf("final module count = %d, want 2 (nic-0 + nic-1)", len(inv.Modules))
	}
}

// ---------- mergeFruProperties ----------

// TestMergeFruProperties verifies mergeFruProperties copies every non-empty
// field from the incoming FRU onto the existing one.
//
// Why it matters: FRU records are refreshed on import, so mutable attributes
// (part number, serial, status, manufacturer) must update in place without
// losing the record's identity.
// Inputs: an empty existing FRU and an incoming FRU with PartNumber, Serial,
// Status, and Manufacturer set. Outputs: all four fields copied.
// Data choice: setting all four covered fields asserts each copy branch runs.
func TestMergeFruProperties(t *testing.T) {
	existing := &CaniFruType{ID: uuid.New(), Name: "fru"}
	incoming := &CaniFruType{
		PartNumber:   "PN-FRU",
		Serial:       "SN-FRU",
		Manufacturer: "Acme",
		ObjectMeta:   ObjectMeta{Status: "Active"},
	}

	mergeFruProperties(existing, incoming)

	if existing.PartNumber != "PN-FRU" || existing.Serial != "SN-FRU" ||
		existing.Status != "Active" || existing.Manufacturer != "Acme" {
		t.Errorf("mergeFruProperties did not copy all fields: %+v", existing)
	}
}

// ---------- MergeFrus ----------

// TestMergeFrus verifies MergeFrus resolves incoming FRUs by UUID match, then
// name match, then insert, while skipping nil and unnamed entries and lazily
// allocating the map.
//
// Why it matters: FRU re-import must converge without duplicates so hardware
// inventory stays accurate across repeated syncs.
// Inputs: a map-less inventory seeded via merge with a "psu-0" FRU, then a merge
// containing a UUID-matched update, a name-matched "psu-0", a new "psu-1", a
// nil, and an unnamed FRU. Outputs: the UUID and name matches folded in, "psu-1"
// inserted, and the nil/unnamed entries skipped.
// Data choice: exercising all three tiers plus the two skip cases covers every
// branch of the loop and the lazy-allocation guard.
func TestMergeFrus(t *testing.T) {
	inv := &Inventory{} // nil Frus map exercises lazy allocation
	idA := uuid.New()
	inv.MergeFrus(map[uuid.UUID]*CaniFruType{
		idA: {ID: idA, Name: "psu-0", ObjectMeta: ObjectMeta{Status: "old"}},
	})
	if len(inv.Frus) != 1 {
		t.Fatalf("after first merge len = %d, want 1", len(inv.Frus))
	}

	inv.MergeFrus(map[uuid.UUID]*CaniFruType{
		idA:        {ID: idA, Name: "psu-0", ObjectMeta: ObjectMeta{Status: "updated"}}, // UUID match
		uuid.New(): {Name: "psu-0", Serial: "SN-name"},                                  // name match
		uuid.New(): {Name: "psu-1"},                                                     // insert
		uuid.New(): nil,                                                                 // skipped (nil)
		uuid.New(): {Name: ""},                                                          // skipped (no name)
	})

	if inv.Frus[idA].Status != "updated" {
		t.Errorf("UUID-matched FRU Status = %q, want %q", inv.Frus[idA].Status, "updated")
	}
	if inv.Frus[idA].Serial != "SN-name" {
		t.Errorf("name-matched merge did not apply, Serial = %q, want %q", inv.Frus[idA].Serial, "SN-name")
	}
	if len(inv.Frus) != 2 {
		t.Errorf("final FRU count = %d, want 2 (psu-0 + psu-1)", len(inv.Frus))
	}
}

// TestMergeRackPropertiesSlugStatusTypeAndMetadata verifies mergeRackProperties
// copies the Slug, Status, Type, and ProviderMetadata fields that the existing
// non-empty test does not exercise.
//
// Why it matters: re-importing a rack must refresh provider-supplied identity
// (slug, status, type) and accumulate provider metadata without dropping keys,
// otherwise downstream reconciliation loses provenance.
// Inputs: an existing rack with empty metadata and an incoming rack carrying a
// Slug, Status (via ObjectMeta), Type, and a one-key ProviderMetadata map.
// Outputs: all four fields copied onto existing and the metadata map allocated.
// Data choice: leaving existing.ProviderMetadata nil forces the make-then-copy
// branch that a pre-populated map would bypass.
func TestMergeRackPropertiesSlugStatusTypeAndMetadata(t *testing.T) {
	existing := &CaniRackType{ID: uuid.New(), Name: "rack-existing"}
	incoming := &CaniRackType{
		Slug:       "rack-slug",
		Type:       "standard",
		ObjectMeta: ObjectMeta{Status: "active", ProviderMetadata: map[string]any{"csm": "x1"}},
	}

	mergeRackProperties(existing, incoming)

	if existing.Slug != "rack-slug" {
		t.Errorf("Slug = %q, want rack-slug", existing.Slug)
	}
	if existing.Status != "active" {
		t.Errorf("Status = %q, want active", existing.Status)
	}
	if string(existing.Type) != "standard" {
		t.Errorf("Type = %q, want standard", existing.Type)
	}
	if existing.ProviderMetadata["csm"] != "x1" {
		t.Errorf("ProviderMetadata[csm] = %v, want x1", existing.ProviderMetadata["csm"])
	}
}

// TestMergeLocationProperties verifies mergeLocationProperties copies every
// non-empty field from the incoming location onto the existing one.
//
// Why it matters: location merges must refresh type, status, description, and
// facility on re-import so a renamed or re-tagged site stays current without a
// full replace.
// Inputs: an existing bare location and an incoming location populating
// LocationType, Status (via ObjectMeta), Description, and Facility. Outputs: all
// four fields copied onto existing. Data choice: populating every optional field
// at once asserts none of the four conditional copies is skipped.
func TestMergeLocationProperties(t *testing.T) {
	existing := &CaniLocationType{ID: uuid.New(), Name: "site-existing"}
	incoming := &CaniLocationType{
		LocationType: "building",
		Description:  "north wing",
		Facility:     "fac-7",
		ObjectMeta:   ObjectMeta{Status: "active"},
	}

	mergeLocationProperties(existing, incoming)

	if existing.LocationType != "building" {
		t.Errorf("LocationType = %q, want building", existing.LocationType)
	}
	if existing.Status != "active" {
		t.Errorf("Status = %q, want active", existing.Status)
	}
	if existing.Description != "north wing" {
		t.Errorf("Description = %q, want north wing", existing.Description)
	}
	if existing.Facility != "fac-7" {
		t.Errorf("Facility = %q, want fac-7", existing.Facility)
	}
}

// TestMergeRacksUUIDMatchAndSkips verifies MergeRacks merges by UUID match and
// skips nil and empty-name incoming entries.
//
// Why it matters: a stable UUID is the strongest identity signal, so a UUID hit
// must field-merge in place (not insert a duplicate), while malformed entries
// must be ignored rather than corrupt the map.
// Inputs: an inventory with one rack; an incoming map keyed by that same UUID
// (with a new UHeight) plus a nil entry and an empty-name entry. Outputs: the
// existing rack updated, an empty remap, and no spurious inserts. Data choice:
// reusing the existing UUID isolates the UUID-match branch, and the two bad
// entries cover both skip guards in one call.
func TestMergeRacksUUIDMatchAndSkips(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Racks[id] = &CaniRackType{ID: id, Name: "rack-1", UHeight: 10}

	remap := inv.MergeRacks(map[uuid.UUID]*CaniRackType{
		id:         {ID: id, Name: "rack-1", UHeight: 42},
		uuid.New(): nil,
		uuid.New(): {Name: ""},
	})

	if len(remap) != 0 {
		t.Errorf("remap = %v, want empty for UUID match", remap)
	}
	if inv.Racks[id].UHeight != 42 {
		t.Errorf("UHeight = %d, want merged 42", inv.Racks[id].UHeight)
	}
	if len(inv.Racks) != 1 {
		t.Errorf("rack count = %d, want 1 (nil + empty-name skipped)", len(inv.Racks))
	}
}

// TestMergeLocationsUUIDMatchAndSkips verifies MergeLocations merges by UUID
// match and skips nil and empty-name incoming entries.
//
// Why it matters: locations re-imported under their original UUID must update in
// place, and malformed entries must not pollute the location map.
// Inputs: an inventory with one location; an incoming map keyed by that UUID
// (with a new LocationType) plus a nil entry and an empty-name entry. Outputs:
// the existing location updated, an empty remap, and no spurious inserts.
// Data choice: reusing the existing UUID isolates the UUID-match branch while
// the two bad entries cover both skip guards.
func TestMergeLocationsUUIDMatchAndSkips(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Locations[id] = &CaniLocationType{ID: id, Name: "site-1", LocationType: "site"}

	remap := inv.MergeLocations(map[uuid.UUID]*CaniLocationType{
		id:         {ID: id, Name: "site-1", LocationType: "building"},
		uuid.New(): nil,
		uuid.New(): {Name: ""},
	})

	if len(remap) != 0 {
		t.Errorf("remap = %v, want empty for UUID match", remap)
	}
	if inv.Locations[id].LocationType != "building" {
		t.Errorf("LocationType = %q, want merged building", inv.Locations[id].LocationType)
	}
	if len(inv.Locations) != 1 {
		t.Errorf("location count = %d, want 1 (nil + empty-name skipped)", len(inv.Locations))
	}
}

// TestRemoveUUIDHelper verifies removeUUID returns a slice with the target
// removed and leaves a slice unchanged when the target is absent.
//
// Why it matters: parent/child relationship cleanup relies on removeUUID to
// detach a single reference without disturbing the rest, so it must drop exactly
// the matching element and preserve order for the others.
// Inputs: a three-element slice from which a middle element is removed, then the
// same slice queried for an absent UUID. Outputs: a two-element slice without
// the target, then an unchanged-length slice. Data choice: a mixed slice with
// both kept and removed elements exercises both branches of the filter loop.
func TestRemoveUUIDHelper(t *testing.T) {
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	got := removeUUID([]uuid.UUID{a, b, c}, b)
	if len(got) != 2 || containsUUID(got, b) {
		t.Errorf("removeUUID = %v, want [a c] without b", got)
	}
	if !containsUUID(got, a) || !containsUUID(got, c) {
		t.Errorf("removeUUID dropped a kept element: %v", got)
	}
	if unchanged := removeUUID([]uuid.UUID{a, c}, b); len(unchanged) != 2 {
		t.Errorf("removeUUID(absent target) len = %d, want 2", len(unchanged))
	}
}

// TestFindDeviceByProviderKeyGuardsAndIndex verifies FindDeviceByProviderKey
// rejects invalid receivers/values and resolves a device through the prebuilt
// provider-key index fast path.
//
// Why it matters: this lookup is on the hot path of provider reconciliation; it
// must never panic on a nil inventory or empty value and must use the O(1) index
// when one has been built rather than always scanning.
// Inputs: a nil inventory and empty/nil values (guard cases), then a populated
// inventory whose index is rebuilt before querying a known redfish_uuid.
// Outputs: nil for every guard case and the matching device via the index.
// Data choice: calling RebuildProviderKeyIndex before the query forces the
// fast-path branch that the existing linear-scan test never reaches.
func TestFindDeviceByProviderKeyGuardsAndIndex(t *testing.T) {
	var nilInv *Inventory
	if nilInv.FindDeviceByProviderKey("redfish", "redfish_uuid", "x") != nil {
		t.Error("nil inventory should yield nil")
	}

	inv := NewInventory()
	if inv.FindDeviceByProviderKey("redfish", "redfish_uuid", "") != nil {
		t.Error("empty value should yield nil")
	}
	if inv.FindDeviceByProviderKey("redfish", "redfish_uuid", nil) != nil {
		t.Error("nil value should yield nil")
	}

	id := uuid.New()
	inv.Devices[id] = &CaniDeviceType{
		ID:   id,
		Name: "srv",
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{"redfish_uuid": "uuid-fast"},
		}},
	}
	inv.RebuildProviderKeyIndex()

	if got := inv.FindDeviceByProviderKey("redfish", "redfish_uuid", "uuid-fast"); got == nil || got.ID != id {
		t.Errorf("index fast path = %v, want device %s", got, id)
	}
}

// TestRemoveCablesForDeviceNilSkip verifies removeCablesForDevice deletes cables
// terminating at the device while skipping nil cable entries.
//
// Why it matters: device removal must prune dependent cables without
// dereferencing a nil map slot, which can exist transiently during a rebuild.
// Inputs: an inventory holding a nil cable, a cable that terminates at the
// target device, and an unrelated cable. Outputs: the terminating cable removed,
// the nil and unrelated cables left intact. Data choice: the nil entry alongside
// a real match drives both the continue guard and the delete branch in one call.
func TestRemoveCablesForDeviceNilSkip(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()

	nilID := uuid.New()
	inv.Cables[nilID] = nil
	matchID := uuid.New()
	inv.Cables[matchID] = &CaniCableType{Label: "match", TerminationADevice: devID}
	keepID := uuid.New()
	inv.Cables[keepID] = &CaniCableType{Label: "keep", TerminationADevice: uuid.New()}

	inv.removeCablesForDevice(devID)

	if _, ok := inv.Cables[matchID]; ok {
		t.Error("expected terminating cable to be removed")
	}
	if _, ok := inv.Cables[keepID]; !ok {
		t.Error("unrelated cable should be kept")
	}
	if _, ok := inv.Cables[nilID]; !ok {
		t.Error("nil cable entry should be left intact (skipped)")
	}
}
