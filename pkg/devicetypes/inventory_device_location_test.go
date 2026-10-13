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
// Data choice: a racked device alongside shows the derived case still clears
// and re-derives Location from its rack.
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
// rebuild clears every derived field of a racked device and its rack's
// device list.
//
// Why it matters: a stale derived value would survive into the rebuild and
// mask a changed Parent.
// Inputs: a rack listing one device; that device with Parent, Children, Rack,
// Location and ParentDevice all set.
// Outputs: the rack lists no devices and every derived device field is nil.
// Data choice: Parent is set so Location counts as derived here; the
// rack-less case is covered by TestRebuildKeepsLocationOfDeviceWithoutRack.
func TestClearDeviceReverseListsResetsFields(t *testing.T) {
	rackID := uuid.New()
	deviceID := uuid.New()
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
		Children:     []uuid.UUID{uuid.New()},
		Rack:         rackID,
		Location:     locID,
		ParentDevice: parentDevID,
	}

	inv.clearDeviceReverseLists()

	if len(inv.Racks[rackID].Devices) != 0 {
		t.Fatal("expected rack devices to be cleared")
	}
	dev := inv.Devices[deviceID]
	if len(dev.Children) != 0 || dev.Rack != uuid.Nil ||
		dev.Location != uuid.Nil || dev.ParentDevice != uuid.Nil {
		t.Fatal("expected all device reverse fields to be cleared")
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
