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
package export

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// ---------- SetInventory ----------

// TestDeviceMapperSetInventory verifies SetInventory stores the inventory
// reference on the mapper for later parent/rack resolution.
//
// Why it matters: the mapper needs the full inventory to resolve a device's
// rack and parent UUIDs into Nautobot references; without it those lookups
// would be empty and placement would be lost on export.
// Inputs: a mapper and an Inventory with an empty Devices map.
// Outputs: mapper.inventory points at the supplied inventory.
// Data choice: an empty-but-non-nil inventory isolates the assignment from any
// resolution side effects.
func TestDeviceMapperSetInventory(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
	}
	mapper.SetInventory(inv)

	if mapper.inventory != inv {
		t.Error("SetInventory did not set inventory")
	}
}

// TestSetInventory verifies SetInventory attaches an Inventory to the mapper so
// later rack-name resolution can read it.
//
// Why it matters: resolveLocalRackName (used by rack diffing) depends on the
// mapper holding the local inventory; without it, rack comparisons cannot run.
// Inputs: a mapper and an Inventory with empty Devices/Racks maps. Outputs: a
// mapper whose inventory field is non-nil.
// Data choice: empty maps are enough to assert the wiring without exercising
// rack resolution itself.
func TestSetInventory(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
		Racks:   map[uuid.UUID]*devicetypes.CaniRackType{},
	}
	mapper.SetInventory(inv)

	// Verify mapper now has the inventory (used by resolveLocalRackName)
	if mapper.inventory == nil {
		t.Fatal("expected non-nil inventory after SetInventory")
	}
}

// ---------- resolveLocalRackName ----------

// TestResolveLocalRackNameNilInventory verifies that resolveLocalRackName
// returns an empty string when the mapper has no inventory set, even though the
// device carries a non-nil Rack UUID.
//
// Why it matters: rack-diff comparison must not panic or fabricate a rack name
// when inventory is unavailable, so a device with an unknown rack is treated as
// unracked rather than emitting a spurious rack PATCH to Nautobot.
// Inputs: a device with a random Rack UUID and a mapper whose inventory is nil.
// Outputs: the empty string.
// Data choice: a populated Rack UUID with nil inventory isolates the nil-guard
// branch, proving the lookup short-circuits before dereferencing inventory.
func TestResolveLocalRackNameNilInventory(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{Rack: uuid.New()}
	if got := resolveLocalRackName(dev, mapper); got != "" {
		t.Errorf("expected empty string for nil inventory, got %q", got)
	}
}

// TestResolveLocalRackNameNoRack verifies that resolveLocalRackName returns an
// empty string when the device resolves to no rack (Rack and Parent are Nil)
// against an empty Racks map.
//
// Why it matters: free-standing devices with no rack assignment must export
// without a rack reference; returning "" keeps them out of rack diffs.
// Inputs: a device with Rack=Nil and Parent=Nil plus an inventory holding an
// empty Racks map. Outputs: the empty string.
// Data choice: empty Racks plus Nil IDs exercises the GetRackID==Nil branch,
// distinct from the nil-inventory case above.
func TestResolveLocalRackNameNoRack(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{},
	}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{Rack: uuid.Nil, Parent: uuid.Nil}
	if got := resolveLocalRackName(dev, mapper); got != "" {
		t.Errorf("expected empty string for no rack, got %q", got)
	}
}

// TestResolveLocalRackNameFound verifies that resolveLocalRackName returns the
// rack's Name when the device's Rack UUID maps to a rack in inventory.
//
// Why it matters: the rack name is the key used to look up the rack's Nautobot
// UUID during diffing, so resolving it correctly is what lets a device be
// placed in the right rack on export.
// Inputs: a device whose Rack UUID matches an inventory rack named "Rack-42".
// Outputs: "Rack-42".
// Data choice: a single named rack keyed by the exact UUID confirms the happy
// path returns the stored name verbatim.
func TestResolveLocalRackNameFound(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	rackID := uuid.New()
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {Name: "Rack-42"},
		},
	}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{Rack: rackID}
	if got := resolveLocalRackName(dev, mapper); got != "Rack-42" {
		t.Errorf("expected 'Rack-42', got %q", got)
	}
}

// ---------- locationFromParentRack ----------

// TestLocationFromParentRackNilInventory verifies that locationFromParentRack
// returns an empty string when the mapper has no inventory, despite a non-nil
// device Rack UUID.
//
// Why it matters: location inheritance from a parent rack must be skipped safely
// when inventory is unavailable, letting resolveLocation fall through to
// metadata/default instead of panicking.
// Inputs: a device with a random Rack UUID and a nil-inventory mapper.
// Outputs: the empty string.
// Data choice: a populated Rack UUID isolates the nil-inventory guard.
func TestLocationFromParentRackNilInventory(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{Rack: uuid.New()}

	if got := mapper.locationFromParentRack(dev); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// TestLocationFromParentRackNoRack verifies that locationFromParentRack returns
// an empty string when the device resolves to no rack (Rack and Parent Nil)
// against an empty Racks map.
//
// Why it matters: devices with no parent rack have no inherited location, so
// resolveLocation must continue to its other sources rather than receive a
// bogus name.
// Inputs: a device with Rack=Nil/Parent=Nil and an empty Racks map.
// Outputs: the empty string.
// Data choice: empty Racks plus Nil IDs hits the GetRackID==Nil branch.
func TestLocationFromParentRackNoRack(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{},
	}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{Rack: uuid.Nil, Parent: uuid.Nil}
	if got := mapper.locationFromParentRack(dev); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// TestLocationFromParentRackResolvesChild verifies that locationFromParentRack
// returns the rack's location name when that location type supports devices,
// resolving via resolveContentLocation.
//
// Why it matters: Nautobot rejects a device whose location does not contain its
// rack, so the device must inherit the rack's device-capable location for a
// valid export.
// Inputs: a registered "server-room-lfp" location type (content types include
// "device"), a rack at that location, and a device in the rack.
// Outputs: "Server-Room-LFP".
// Data choice: a location type explicitly listing "device" content exercises
// the supported-location path without tree-walking to a child.
func TestLocationFromParentRackResolvesChild(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	// Register a location type that supports devices.
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Server Room LFP",
		Slug:         "server-room-lfp",
		ContentTypes: []string{"rack", "device"},
	})

	locID := uuid.New()
	rackID := uuid.New()

	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {Name: "Rack-X", Location: locID},
		},
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			locID: {
				ID:           locID,
				Name:         "Server-Room-LFP",
				LocationType: "server-room-lfp",
			},
		},
	}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{Rack: rackID}
	got := mapper.locationFromParentRack(dev)
	if got != "Server-Room-LFP" {
		t.Errorf("expected 'Server-Room-LFP', got %q", got)
	}
}

// ---------- setDeviceFace ----------

// TestSetDeviceFaceRear verifies that setDeviceFace(&req.Face, "rear") sets a
// RackFace whose generated Nautobot union decodes to FaceEnumRear.
//
// Why it matters: Nautobot requires a face value whenever a rack position is
// set, so the exporter must always produce a usable RackFace for rack-mounted
// devices and must preserve an explicit rear orientation.
// Inputs: the face string "rear". Outputs: a non-nil *nautobotapi.RackFace
// whose enum value is FaceEnumRear.
// Data choice: "rear" exercises the explicit non-default switch case.
func TestSetDeviceFaceRear(t *testing.T) {
	var req nautobotapi.WritableDeviceRequest
	setDeviceFace(&req.Face, "rear")
	if req.Face == nil {
		t.Fatal("expected non-nil RackFace")
	}
	face, err := req.Face.AsFaceEnum()
	if err != nil {
		t.Fatalf("decode RackFace: %v", err)
	}
	if face != nautobotapi.FaceEnumRear {
		t.Errorf("face = %v, want %v", face, nautobotapi.FaceEnumRear)
	}
}

// TestSetDeviceFaceFront verifies that setDeviceFace(&req.Face, "front")
// sets a RackFace whose generated Nautobot union decodes to FaceEnumFront.
//
// Why it matters: rack-mounted devices exported to Nautobot need a valid face;
// "front" is the most common orientation and must map to a usable RackFace.
// Inputs: the face string "front". Outputs: a non-nil *nautobotapi.RackFace
// whose enum value is FaceEnumFront.
// Data choice: "front" hits the default switch branch with an explicit value.
func TestSetDeviceFaceFront(t *testing.T) {
	var req nautobotapi.WritableDeviceRequest
	setDeviceFace(&req.Face, "front")
	if req.Face == nil {
		t.Fatal("expected non-nil RackFace")
	}
	face, err := req.Face.AsFaceEnum()
	if err != nil {
		t.Fatalf("decode RackFace: %v", err)
	}
	if face != nautobotapi.FaceEnumFront {
		t.Errorf("face = %v, want %v", face, nautobotapi.FaceEnumFront)
	}
}

// TestSetDeviceFaceEmpty verifies that setDeviceFace(&req.Face, "") defaults
// to a RackFace whose generated Nautobot union decodes to FaceEnumFront.
//
// Why it matters: cani devices may omit a face; since Nautobot demands one when
// a rack position exists, the exporter must default rather than emit nil/null.
// Inputs: an empty face string. Outputs: a non-nil *nautobotapi.RackFace
// whose enum value is FaceEnumFront.
// Data choice: the empty string drives the default branch without an explicit
// input face.
func TestSetDeviceFaceEmpty(t *testing.T) {
	var req nautobotapi.WritableDeviceRequest
	setDeviceFace(&req.Face, "")
	if req.Face == nil {
		t.Fatal("expected non-nil RackFace (default to front)")
	}
	face, err := req.Face.AsFaceEnum()
	if err != nil {
		t.Fatalf("decode RackFace: %v", err)
	}
	if face != nautobotapi.FaceEnumFront {
		t.Errorf("face = %v, want %v", face, nautobotapi.FaceEnumFront)
	}
}
