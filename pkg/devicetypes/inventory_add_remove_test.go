package devicetypes

// Test coverage for inventory_add_remove.go
//
// | Function       | Happy-path test                              | Failure test                                    |
// |----------------|----------------------------------------------|-------------------------------------------------|
// | AddLocation    | TestAddLocationValid                         | TestAddLocationNil                              |
// | AddRack        | TestAddRackValid                             | TestAddRackDuplicateUUID                        |
// | AddModule      | TestAddModuleValid                           | TestAddModuleNil                                |
// | AddCable       | TestAddCableValid                            | TestAddCableDuplicateUUID                       |
// | RemoveLocation | TestRemoveLocationEmpty                      | TestRemoveLocationHasRacks                      |
// | RemoveRack     | TestRemoveRackOrphansDevices                 | TestRemoveRackNotFound                          |
// | RemoveModule   | TestRemoveModuleValid                        | TestRemoveModuleNotFound                        |
// | RemoveCable    | TestRemoveCableValid                         | TestRemoveCableNotFound                         |

import (
	"testing"

	"github.com/google/uuid"
)

// ---------- AddLocation ----------

func TestAddLocationValid(t *testing.T) {
	inv := NewInventory()
	loc := &CaniLocationType{
		ID:           uuid.New(),
		Name:         "site-alpha",
		LocationType: "site",
		ObjectMeta:   ObjectMeta{Status: "Active"},
	}

	if err := inv.AddLocation(loc); err != nil {
		t.Fatalf("AddLocation() unexpected error: %v", err)
	}
	if _, ok := inv.Locations[loc.ID]; !ok {
		t.Error("expected location to be present in inventory after AddLocation")
	}
}

func TestAddLocationNil(t *testing.T) {
	inv := NewInventory()
	if err := inv.AddLocation(nil); err == nil {
		t.Error("AddLocation(nil) should return an error")
	}
}

// ---------- AddRack ----------

func TestAddRackValid(t *testing.T) {
	inv := NewInventory()
	rack := &CaniRackType{
		ID:      uuid.New(),
		Name:    "rack-01",
		UHeight: 42,
	}

	if err := inv.AddRack(rack); err != nil {
		t.Fatalf("AddRack() unexpected error: %v", err)
	}
	if _, ok := inv.Racks[rack.ID]; !ok {
		t.Error("expected rack to be present in inventory after AddRack")
	}
}

func TestAddRackDuplicateUUID(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Racks[id] = &CaniRackType{ID: id, Name: "existing-rack"}

	dup := &CaniRackType{ID: id, Name: "duplicate-rack"}
	if err := inv.AddRack(dup); err == nil {
		t.Error("AddRack(duplicate UUID) should return an error")
	}
}

// ---------- AddModule ----------

func TestAddModuleValid(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	inv.Devices[devID] = &CaniDeviceType{ID: devID, Name: "parent-dev"}

	mod := &CaniModuleType{
		ID:           uuid.New(),
		Name:         "nic-0",
		ParentDevice: devID,
	}

	if err := inv.AddModule(mod); err != nil {
		t.Fatalf("AddModule() unexpected error: %v", err)
	}
	if _, ok := inv.Modules[mod.ID]; !ok {
		t.Error("expected module to be present in inventory after AddModule")
	}
}

func TestAddModuleNil(t *testing.T) {
	inv := NewInventory()
	if err := inv.AddModule(nil); err == nil {
		t.Error("AddModule(nil) should return an error")
	}
}

// ---------- AddCable ----------

func TestAddCableValid(t *testing.T) {
	inv := NewInventory()
	cable := &CaniCableType{
		ID:    uuid.New(),
		Label: "cable-01",
	}

	if err := inv.AddCable(cable); err != nil {
		t.Fatalf("AddCable() unexpected error: %v", err)
	}
	if _, ok := inv.Cables[cable.ID]; !ok {
		t.Error("expected cable to be present in inventory after AddCable")
	}
}

func TestAddCableDuplicateUUID(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Cables[id] = &CaniCableType{ID: id, Label: "existing-cable"}

	dup := &CaniCableType{ID: id, Label: "dup-cable"}
	if err := inv.AddCable(dup); err == nil {
		t.Error("AddCable(duplicate UUID) should return an error")
	}
}

// ---------- RemoveLocation ----------

func TestRemoveLocationEmpty(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Locations[id] = &CaniLocationType{
		ID:       id,
		Name:     "leaf-site",
		Children: nil,
		Racks:    nil,
	}

	if err := inv.RemoveLocation(id); err != nil {
		t.Fatalf("RemoveLocation() unexpected error: %v", err)
	}
	if _, ok := inv.Locations[id]; ok {
		t.Error("expected location to be removed from inventory")
	}
}

func TestRemoveLocationHasRacks(t *testing.T) {
	inv := NewInventory()
	locID := uuid.New()
	rackID := uuid.New()

	inv.Locations[locID] = &CaniLocationType{
		ID:    locID,
		Name:  "busy-site",
		Racks: []uuid.UUID{rackID},
	}
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "rack-in-site", Location: locID}

	if err := inv.RemoveLocation(locID); err == nil {
		t.Error("RemoveLocation should fail when location still has racks")
	}
}

// ---------- RemoveRack ----------

func TestRemoveRackOrphansDevices(t *testing.T) {
	inv := NewInventory()
	rackID := uuid.New()
	devID := uuid.New()

	inv.Racks[rackID] = &CaniRackType{
		ID:      rackID,
		Name:    "rack-with-devs",
		UHeight: 42,
		Devices: []uuid.UUID{devID},
	}
	inv.Devices[devID] = &CaniDeviceType{
		ID:     devID,
		Name:   "child-server",
		Parent: rackID,
	}

	if err := inv.RemoveRack(rackID); err != nil {
		t.Fatalf("RemoveRack() unexpected error: %v", err)
	}
	if _, ok := inv.Racks[rackID]; ok {
		t.Error("expected rack to be removed from inventory")
	}
	if inv.Devices[devID].Parent != uuid.Nil {
		t.Error("expected device to be orphaned (Parent = uuid.Nil) after RemoveRack")
	}
}

func TestRemoveRackNotFound(t *testing.T) {
	inv := NewInventory()
	if err := inv.RemoveRack(uuid.New()); err == nil {
		t.Error("RemoveRack(non-existent) should return an error")
	}
}

// ---------- RemoveModule ----------

func TestRemoveModuleValid(t *testing.T) {
	inv := NewInventory()
	modID := uuid.New()
	inv.Modules[modID] = &CaniModuleType{ID: modID, Name: "nic-to-remove"}

	if err := inv.RemoveModule(modID); err != nil {
		t.Fatalf("RemoveModule() unexpected error: %v", err)
	}
	if _, ok := inv.Modules[modID]; ok {
		t.Error("expected module to be removed from inventory")
	}
}

func TestRemoveModuleNotFound(t *testing.T) {
	inv := NewInventory()
	if err := inv.RemoveModule(uuid.New()); err == nil {
		t.Error("RemoveModule(non-existent) should return an error")
	}
}

// ---------- RemoveCable ----------

func TestRemoveCableValid(t *testing.T) {
	inv := NewInventory()
	cableID := uuid.New()
	inv.Cables[cableID] = &CaniCableType{ID: cableID, Label: "cable-to-remove"}

	if err := inv.RemoveCable(cableID); err != nil {
		t.Fatalf("RemoveCable() unexpected error: %v", err)
	}
	if _, ok := inv.Cables[cableID]; ok {
		t.Error("expected cable to be removed from inventory")
	}
}

func TestRemoveCableNotFound(t *testing.T) {
	inv := NewInventory()
	if err := inv.RemoveCable(uuid.New()); err == nil {
		t.Error("RemoveCable(non-existent) should return an error")
	}
}
