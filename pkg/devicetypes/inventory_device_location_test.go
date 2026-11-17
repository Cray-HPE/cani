/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package devicetypes

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestRebuildKeepsLocationOfDeviceWithoutRack verifies that a device placed
// directly in a location (no Parent) keeps that Location across a rebuild,
// survives a JSON round trip with it, and is not reported as an orphan.
//
// Why it matters: Nautobot allows devices with a location and no rack; the
// import sets Location for them and nothing else could re-derive it, so a
// rebuild that cleared it would silently lose the placement.
// Inputs: one location and one device with Location set and Parent nil.
// Outputs: Location unchanged after RebuildDerivedState, present in the
// serialized form, no orphan or error reported.
// Data choice: a racked device alongside records a location that does not
// hold its rack's, which the rebuild still replaces with the rack's.
func TestRebuildKeepsLocationOfDeviceWithoutRack(t *testing.T) {
	inv := NewInventory()
	locID, rackID, loose, racked := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Locations[locID] = &CaniLocationType{ID: locID, Name: "test-edge"}
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "rack-01", Location: locID}
	inv.Devices[loose] = &CaniDeviceType{ID: loose, Name: "edge-001", Location: locID}
	inv.Devices[racked] = &CaniDeviceType{ID: racked, Name: "compute-001", Parent: rackID, Location: uuid.New()}

	res := inv.rebuildDeviceRelationships()

	if res.HasErrors() || len(res.Orphans) != 0 {
		t.Fatalf("a located device must be neither an error nor an orphan: errors=%v orphans=%v", res.Errors, res.Orphans)
	}
	if inv.Devices[loose].Location != locID {
		t.Fatalf("loose device lost its location: %s", inv.Devices[loose].Location)
	}
	if inv.Devices[racked].Location != locID {
		t.Fatalf("racked device must derive its location from the rack, got %s", inv.Devices[racked].Location)
	}
	raw, err := json.Marshal(inv.Devices[loose])
	if err != nil {
		t.Fatal(err)
	}
	var persisted map[string]any
	if err := json.Unmarshal(raw, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted["location"] != locID.String() {
		t.Fatalf("location not persisted: %v", persisted["location"])
	}
}

// TestRebuildReportsDanglingDeviceLocation verifies that a device without a
// rack whose Location names no known location is reported as an error, the
// way a rack with an unknown location is.
//
// Why it matters: a dangling placement would export to the provider default
// without anyone noticing.
// Inputs: a device with Parent nil and a random Location.
// Outputs: one error naming the device; no orphan.
// Data choice: no locations at all, so the lookup cannot succeed by chance.
func TestRebuildReportsDanglingDeviceLocation(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Devices[id] = &CaniDeviceType{ID: id, Name: "edge-001", Location: uuid.New()}

	res := inv.rebuildDeviceRelationships()

	if len(res.Errors) != 1 || len(res.Orphans) != 0 {
		t.Fatalf("expected one error and no orphan, got errors=%v orphans=%v", res.Errors, res.Orphans)
	}
}

// TestMergePropertiesCarriesLocation verifies that merging an incoming device
// with a Location onto one without carries the Location, while an incoming
// nil Location leaves the existing one.
//
// Why it matters: a rack-less device's placement arrives only through this
// field, so a re-import must carry it like any other scalar.
// Inputs: an existing device without Location; an incoming copy with one,
// then one without.
// Outputs: the first merge sets it and reports a change; the second keeps it
// and reports none.
// Data choice: names match so nothing but Location differs.
func TestMergePropertiesCarriesLocation(t *testing.T) {
	locID := uuid.New()
	existing := &CaniDeviceType{Name: "edge-001"}

	if !existing.MergeProperties(&CaniDeviceType{Name: "edge-001", Location: locID}) {
		t.Fatal("expected a change when the incoming device carries a location")
	}
	if existing.MergeProperties(&CaniDeviceType{Name: "edge-001"}) || existing.Location != locID {
		t.Fatalf("a nil incoming location must neither change nor clear: %s", existing.Location)
	}
}

// TestRemoveLocationRefusesWhileDevicesArePlacedInIt verifies that a location
// holding a rack-less device cannot be removed until the device is, and that
// a racked device elsewhere does not count against it.
//
// Why it matters: a device's Location is now persisted, so removing the
// location under it would leave a dangling reference that every later load
// reports as an error, the way removing a location under a rack would.
// Inputs: a location with one rack-less device; a second location whose
// device sits in a rack.
// Outputs: the first removal fails naming one device; after the device is
// removed it succeeds; the second location's device is not counted.
// Data choice: the racked device has a derived Location equal to its
// location, which must be attributed to the rack guard, not the device one.
func TestRemoveLocationRefusesWhileDevicesArePlacedInIt(t *testing.T) {
	inv := NewInventory()
	edge, dc, rackID, loose, racked := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Locations[edge] = &CaniLocationType{ID: edge, Name: "test-edge"}
	inv.Locations[dc] = &CaniLocationType{ID: dc, Name: "test-dc"}
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "rack-01", Location: dc}
	inv.Devices[loose] = &CaniDeviceType{ID: loose, Name: "edge-001", Location: edge}
	inv.Devices[racked] = &CaniDeviceType{ID: racked, Name: "compute-001", Parent: rackID}
	inv.RebuildDerivedState()

	err := inv.RemoveLocation(edge)

	if err == nil || !strings.Contains(err.Error(), "still has 1 device(s)") {
		t.Fatalf("expected the device guard, got %v", err)
	}
	if inv.countDevicesPlacedIn(dc) != 0 {
		t.Fatal("a racked device must be counted by the rack guard, not the device guard")
	}
	delete(inv.Devices, loose)
	if err := inv.RemoveLocation(edge); err != nil {
		t.Fatalf("removal must succeed once the device is gone, got %v", err)
	}
}

// TestRemoveRackLeavesDevicesInItsLocation verifies that removing a rack
// leaves the devices it held placed directly in the rack's location, so a
// rebuild reports neither an orphan nor an error and a save keeps the
// placement.
//
// Why it matters: before Location was persisted, an unracked device had no
// placement left to save and came back as an orphan; Nautobot keeps a
// device's location when its rack goes away, and so should the datastore.
// Inputs: a location, a rack in it listing one device, and that device with
// only its Parent set; the rack is removed.
// Outputs: the device has no Parent, its Location is the rack's, and a
// rebuild reports no orphan and no error.
// Data choice: the device's derived Location is left unset, so the removal
// itself must supply the placement rather than a rebuild having done so.
func TestRemoveRackLeavesDevicesInItsLocation(t *testing.T) {
	inv := NewInventory()
	locID, rackID, devID := uuid.New(), uuid.New(), uuid.New()
	inv.Locations[locID] = &CaniLocationType{ID: locID, Name: "test-dc"}
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "rack-01", Location: locID, Devices: []uuid.UUID{devID}}
	inv.Devices[devID] = &CaniDeviceType{ID: devID, Name: "compute-001", Parent: rackID}

	if err := inv.RemoveRack(rackID); err != nil {
		t.Fatalf("RemoveRack() error = %v", err)
	}

	dev := inv.Devices[devID]
	if dev.Parent != uuid.Nil || dev.Location != locID {
		t.Fatalf("device must stay in the rack's location without a rack: parent=%s location=%s", dev.Parent, dev.Location)
	}
	res := inv.rebuildDeviceRelationships()
	if res.HasErrors() || len(res.Orphans) != 0 {
		t.Fatalf("an unracked device in a location is not an orphan: errors=%v orphans=%v", res.Errors, res.Orphans)
	}
}

// TestClearDeviceReverseListsResetsFields verifies that the reset before a
// rebuild clears every derived field and the rack's device list, clears the
// location of a device installed in another device, and leaves a
// rack-mounted device the location it records.
//
// Why it matters: a stale derived value would survive into the rebuild and
// mask a changed Parent, while a racked device's recorded location is input
// the rebuild still has to judge against its rack.
// Inputs: a rack listing one device with Parent, Children, Rack, Location and
// ParentDevice all set, and a second device installed in the first with a
// Location of its own.
// Outputs: the rack lists no devices; both devices lose Children, Rack and
// ParentDevice; the racked device keeps its Location and the installed one
// loses it.
// Data choice: both devices carry a Location, so the kind of parent is the
// only thing that decides which survives.
func TestClearDeviceReverseListsResetsFields(t *testing.T) {
	rackID := uuid.New()
	deviceID := uuid.New()
	childID := uuid.New()
	locID := uuid.New()
	parentDevID := uuid.New()

	inv := NewInventory()
	inv.Racks[rackID] = &CaniRackType{
		Name:    "r1",
		Devices: []uuid.UUID{deviceID},
	}
	inv.Devices[deviceID] = &CaniDeviceType{
		Name:         "d1",
		Parent:       rackID,
		Children:     []uuid.UUID{childID},
		Rack:         rackID,
		Location:     locID,
		ParentDevice: parentDevID,
	}
	inv.Devices[childID] = &CaniDeviceType{Name: "blade", Parent: deviceID, Rack: rackID, Location: locID, ParentDevice: deviceID}

	inv.clearDeviceReverseLists()

	if len(inv.Racks[rackID].Devices) != 0 {
		t.Fatal("expected rack devices to be cleared")
	}
	for _, dev := range []*CaniDeviceType{inv.Devices[deviceID], inv.Devices[childID]} {
		if len(dev.Children) != 0 || dev.Rack != uuid.Nil || dev.ParentDevice != uuid.Nil {
			t.Fatalf("expected the derived fields of %s to be cleared", dev.Name)
		}
	}
	if got := inv.Devices[deviceID].Location; got != locID {
		t.Errorf("racked device location = %s, want the recorded one kept for the rebuild to judge", got)
	}
	if got := inv.Devices[childID].Location; got != uuid.Nil {
		t.Errorf("installed device location = %s, want it cleared", got)
	}
}

// TestClearDeviceReverseListsEmptyInventory verifies the reset tolerates an
// inventory with no racks or devices.
//
// Why it matters: a fresh datastore is rebuilt like any other.
// Inputs: an empty inventory. Outputs: no panic.
// Data choice: the empty maps are the only input that exercises the loops'
// zero-iteration path.
func TestClearDeviceReverseListsEmptyInventory(t *testing.T) {
	inv := NewInventory()
	inv.clearDeviceReverseLists()
}

// locationTree is an inventory with a site that holds two rooms and one rack
// in the first room.
type locationTree struct {
	inv                         *Inventory
	site, room, otherRoom, rack uuid.UUID
}

func newLocationTree() locationTree {
	tree := locationTree{inv: NewInventory(), site: uuid.New(), room: uuid.New(), otherRoom: uuid.New(), rack: uuid.New()}
	tree.inv.Locations[tree.site] = &CaniLocationType{ID: tree.site, Name: "site"}
	tree.inv.Locations[tree.room] = &CaniLocationType{ID: tree.room, Name: "room-1", Parent: tree.site}
	tree.inv.Locations[tree.otherRoom] = &CaniLocationType{ID: tree.otherRoom, Name: "room-2", Parent: tree.site}
	tree.inv.Racks[tree.rack] = &CaniRackType{ID: tree.rack, Name: "rack-01", Location: tree.room}
	return tree
}

// addDevice puts a device with the given parent and recorded location into
// the tree's inventory and returns it.
func (tree locationTree) addDevice(name string, parent, location uuid.UUID) *CaniDeviceType {
	device := &CaniDeviceType{ID: uuid.New(), Name: name, Parent: parent, Location: location}
	tree.inv.Devices[device.ID] = device
	return device
}

// TestRebuildKeepsRackedDeviceLocationThatHoldsItsRack verifies which
// recorded location a rack-mounted device keeps across a rebuild: the rack's
// own or an ancestor of it, and nothing else; a device installed in another
// device takes that device's location whatever it records.
//
// Why it matters: Nautobot accepts a racked device in any location that
// holds its rack's, so replacing a recorded parent location with the rack's
// made every re-export move such a device.
// Inputs: a rack in room-1 under site, with devices recording the site, the
// room, the sibling room and nothing, and a blade installed in the first
// that records the sibling room.
// Outputs: site, room-1, room-1, room-1, and site for the blade.
// Data choice: the sibling room is a real location that does not hold the
// rack's, so existence alone cannot be what keeps a location.
func TestRebuildKeepsRackedDeviceLocationThatHoldsItsRack(t *testing.T) {
	// Arrange.
	tree := newLocationTree()
	atSite := tree.addDevice("at-site", tree.rack, tree.site)
	atRoom := tree.addDevice("at-room", tree.rack, tree.room)
	elsewhere := tree.addDevice("elsewhere", tree.rack, tree.otherRoom)
	unrecorded := tree.addDevice("unrecorded", tree.rack, uuid.Nil)
	blade := tree.addDevice("blade", atSite.ID, tree.otherRoom)

	// Act.
	res := tree.inv.rebuildDeviceRelationships()

	// Assert.
	if res.HasErrors() {
		t.Fatalf("rebuild errors: %v", res.Errors)
	}
	want := map[*CaniDeviceType]uuid.UUID{
		atSite: tree.site, atRoom: tree.room, elsewhere: tree.room, unrecorded: tree.room, blade: tree.site,
	}
	for device, location := range want {
		if device.Location != location {
			t.Errorf("%s location = %s, want %s (site %s, room-1 %s, room-2 %s)",
				device.Name, device.Location, location, tree.site, tree.room, tree.otherRoom)
		}
	}
}

// TestSetDeviceLocationFollowsThePlacementRules verifies each outcome of
// placing a device: any existing location for a device with no parent, the
// rack's location or one that holds it for a racked device, and a refusal
// for every other case.
//
// Why it matters: a location the rebuild would silently replace, or one that
// does not exist, must be refused when it is typed, not discovered at the
// next export.
// Inputs: a device with no parent, a racked device and a blade installed in
// it, each given the site, the sibling room or an unknown location.
// Outputs: the loose device takes the site; the racked device takes the site
// and keeps it when the sibling room and the unknown location are refused;
// the blade is refused.
// Data choice: the refused calls follow an accepted one, so an unchanged
// location proves a refusal writes nothing.
func TestSetDeviceLocationFollowsThePlacementRules(t *testing.T) {
	// Arrange.
	tree := newLocationTree()
	loose := tree.addDevice("loose", uuid.Nil, uuid.Nil)
	racked := tree.addDevice("racked", tree.rack, uuid.Nil)
	blade := tree.addDevice("blade", racked.ID, uuid.Nil)

	// Act.
	looseErr := tree.inv.SetDeviceLocation(loose, tree.site)
	rackedErr := tree.inv.SetDeviceLocation(racked, tree.site)
	siblingErr := tree.inv.SetDeviceLocation(racked, tree.otherRoom)
	unknownErr := tree.inv.SetDeviceLocation(racked, uuid.New())
	bladeErr := tree.inv.SetDeviceLocation(blade, tree.site)

	// Assert.
	if looseErr != nil || loose.Location != tree.site {
		t.Errorf("loose device: error = %v, location = %s; want the site", looseErr, loose.Location)
	}
	if rackedErr != nil || racked.Location != tree.site {
		t.Errorf("racked device: error = %v, location = %s; want the site kept", rackedErr, racked.Location)
	}
	if siblingErr == nil || !strings.Contains(siblingErr.Error(), "room-2") {
		t.Errorf("sibling room error = %v, want a refusal naming room-2", siblingErr)
	}
	if unknownErr == nil || bladeErr == nil || blade.Location != uuid.Nil {
		t.Errorf("unknown location error = %v, blade error = %v, blade location = %s; want two refusals and no location",
			unknownErr, bladeErr, blade.Location)
	}
}

// TestMoveRackTakesItsDevicesAlong verifies that moving a rack moves every
// device mounted in it, including one that recorded a location of its own,
// and that an unknown rack or location is refused without moving anything.
//
// Why it matters: the site still holds the rack's new room, so without the
// reset a device recorded at the site would stay there; Nautobot moves the
// devices of a rack to the rack's location whenever the rack is saved.
// Inputs: a rack in room-1 with one device recorded at the site and one that
// records nothing; the rack is moved to room-2, then to an unknown location.
// Outputs: both devices in room-2 after a rebuild; an error and the rack
// still in room-2 for the unknown location; an error for an unknown rack.
// Data choice: both rooms share the site, which is the case where a recorded
// ancestor would otherwise survive the move.
func TestMoveRackTakesItsDevicesAlong(t *testing.T) {
	// Arrange.
	tree := newLocationTree()
	atSite := tree.addDevice("at-site", tree.rack, tree.site)
	unrecorded := tree.addDevice("unrecorded", tree.rack, uuid.Nil)
	tree.inv.rebuildDeviceRelationships()

	// Act.
	moveErr := tree.inv.MoveRack(tree.rack, tree.otherRoom)
	tree.inv.rebuildDeviceRelationships()
	unknownLocationErr := tree.inv.MoveRack(tree.rack, uuid.New())
	unknownRackErr := tree.inv.MoveRack(uuid.New(), tree.room)

	// Assert.
	if moveErr != nil || atSite.Location != tree.otherRoom || unrecorded.Location != tree.otherRoom {
		t.Errorf("after the move: error = %v, locations = %s, %s; want both in room-2",
			moveErr, atSite.Location, unrecorded.Location)
	}
	if unknownLocationErr == nil || unknownRackErr == nil || tree.inv.Racks[tree.rack].Location != tree.otherRoom {
		t.Errorf("unknown location error = %v, unknown rack error = %v, rack location = %s; want two refusals and room-2",
			unknownLocationErr, unknownRackErr, tree.inv.Racks[tree.rack].Location)
	}
}

// TestMergeReleasesRecordedLocationWhenRackOrParentChanges verifies that a
// merge which moves a rack, or gives a device another parent, clears the
// location the device recorded unless the incoming device names one.
//
// Why it matters: an import that moves a rack must move its devices, as a
// rack move at the command line does; a location recorded under the old
// placement would otherwise outlive it whenever it still held the new one.
// Inputs: a device recorded at the site in a rack in room-1. The rack is
// merged with a copy in room-2; then the device is merged with a copy in a
// second rack, without and with a location.
// Outputs: no recorded location after the rack merge and after the first
// device merge; the incoming location after the second.
// Data choice: the incoming device names the sibling room, which is not
// where it was, so the merge is seen to take it rather than keep the old.
func TestMergeReleasesRecordedLocationWhenRackOrParentChanges(t *testing.T) {
	// Arrange.
	tree := newLocationTree()
	device := tree.addDevice("at-site", tree.rack, tree.site)
	rack2, rack3 := uuid.New(), uuid.New()

	// Act.
	tree.inv.MergeRacks(map[uuid.UUID]*CaniRackType{tree.rack: {ID: tree.rack, Name: "rack-01", Location: tree.otherRoom}})
	afterRackMerge := device.Location
	device.Location = tree.site
	device.MergeProperties(&CaniDeviceType{Name: "at-site", Parent: rack2})
	afterReparent := device.Location
	device.Location = tree.site
	device.MergeProperties(&CaniDeviceType{Name: "at-site", Parent: rack3, Location: tree.otherRoom})

	// Assert.
	if afterRackMerge != uuid.Nil || afterReparent != uuid.Nil {
		t.Errorf("recorded location after rack merge = %s, after re-parenting = %s; want none for both", afterRackMerge, afterReparent)
	}
	if device.Parent != rack3 || device.Location != tree.otherRoom {
		t.Errorf("device = {parent %s, location %s}, want the incoming rack and location", device.Parent, device.Location)
	}
}

// TestRemoveRackKeepsRecordedLocationThatHeldIt verifies that a device which
// recorded a location holding its rack's stays there when the rack is
// removed.
//
// Why it matters: the device was in that location, not the rack's, before
// the rack went away; replacing it with the rack's would move the device.
// Inputs: a rack in room-1 holding a device recorded at the site.
// Outputs: the device has no parent and is still in the site.
// Data choice: the complementary case, a device that records nothing, is
// TestRemoveRackLeavesDevicesInItsLocation.
func TestRemoveRackKeepsRecordedLocationThatHeldIt(t *testing.T) {
	// Arrange.
	tree := newLocationTree()
	device := tree.addDevice("at-site", tree.rack, tree.site)
	tree.inv.rebuildDeviceRelationships()

	// Act.
	err := tree.inv.RemoveRack(tree.rack)

	// Assert.
	if err != nil || device.Parent != uuid.Nil || device.Location != tree.site {
		t.Errorf("error = %v, parent = %s, location = %s; want no parent and the site", err, device.Parent, device.Location)
	}
}
