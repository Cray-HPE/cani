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
	"testing"

	"github.com/google/uuid"
)

// cloneFixtureIDs are the identifiers the clone fixtures are built around.
type cloneFixtureIDs struct {
	location uuid.UUID
	rack     uuid.UUID
	device   uuid.UUID
	module   uuid.UUID
	cable    uuid.UUID
	fru      uuid.UUID
}

// newCloneFixtureIDs mints one identifier per entity the restore table covers.
func newCloneFixtureIDs() cloneFixtureIDs {
	return cloneFixtureIDs{
		location: uuid.New(),
		rack:     uuid.New(),
		device:   uuid.New(),
		module:   uuid.New(),
		cable:    uuid.New(),
		fru:      uuid.New(),
	}
}

// fixtureObjectMeta returns metadata whose maps hold values that a JSON
// round-trip would re-type: an int would come back as float64, so comparing it
// afterwards proves the value was restored rather than decoded.
func fixtureObjectMeta(tag string) ObjectMeta {
	return ObjectMeta{
		Status:       "active",
		Tags:         []string{tag},
		CustomFields: map[string]any{"count": 7, "nested": map[string]any{"depth": 2}},
		ExternalIDs:  map[string]uuid.UUID{"nautobot": uuid.New()},
		ProviderMetadata: map[string]any{
			"csm": map[string]any{"xname": "x3000c0s1b0n0", "slot": 1},
		},
	}
}

// fixtureDeviceBays returns bays exercising every dynamic field restoreDeviceBays
// touches, including both DeviceBaySlugRef pointers.
func fixtureDeviceBays() []DeviceBaySpec {
	return []DeviceBaySpec{{
		Name:     "bay-1",
		Position: "1",
		Allowed:  &DeviceBaySlugRef{Slug: []any{"blade-a", "blade-b"}, Types: "node"},
		Default:  &DeviceBaySlugRef{Slug: "blade-a"},
		Extra:    map[string]any{"depth": 3},
	}}
}

// fixtureInterfaces returns interface specs carrying provider metadata, the one
// InterfaceSpec field the restore table has to put back by hand.
func fixtureInterfaces(name string) []InterfaceSpec {
	return []InterfaceSpec{{
		ID:               uuid.New(),
		Name:             name,
		ProviderMetadata: map[string]any{"speed": 400},
	}}
}

// newCloneFixture builds an inventory that populates every field
// restoreInventoryDynamicFields is responsible for putting back.
func newCloneFixture(ids cloneFixtureIDs) *Inventory {
	inventory := NewInventory()

	inventory.Locations[ids.location] = &CaniLocationType{
		ID: ids.location, Name: "hall-1", ObjectMeta: fixtureObjectMeta("loc"),
	}
	inventory.Racks[ids.rack] = &CaniRackType{
		ID: ids.rack, Name: "rack-01", Location: ids.location,
		Source:           "fixture-rack",
		ProviderDefaults: map[string]any{"height": 42},
		DeviceBays:       fixtureDeviceBays(),
		ObjectMeta:       fixtureObjectMeta("rack"),
	}
	inventory.Devices[ids.device] = &CaniDeviceType{
		ID: ids.device, Name: "cn-01", Parent: ids.rack,
		Source:     "fixture-device",
		Interfaces: fixtureInterfaces("eth0"),
		DeviceBays: fixtureDeviceBays(),
		ObjectMeta: fixtureObjectMeta("device"),
	}
	inventory.Modules[ids.module] = &CaniModuleType{
		ID: ids.module, Name: "nic-01", ParentDevice: ids.device,
		Source:     "fixture-module",
		Interfaces: fixtureInterfaces("port0"),
		ObjectMeta: fixtureObjectMeta("module"),
	}
	inventory.Cables[ids.cable] = &CaniCableType{
		ID: ids.cable, Label: "cable-01",
		Source: "fixture-cable", ObjectMeta: fixtureObjectMeta("cable"),
	}
	inventory.Frus[ids.fru] = &CaniFruType{
		ID: ids.fru, Name: "fru-01", Device: ids.device,
		Source: "fixture-fru", ObjectMeta: fixtureObjectMeta("fru"),
	}
	inventory.Metadata = &InventoryMetadata{
		CustomFields: []CustomFieldDefinition{{Key: "rack-class", Default: 3}},
	}
	return inventory
}

// TestCloneInventoryRestoresEveryNonSerializedField verifies cloneInventory puts
// back every field a JSON round-trip destroys.
//
// Why it matters: cloneInventory is the rollback boundary for
// MergeTransformResult, and it clones by marshalling to JSON and back. That
// discards every field tagged json:"-" and re-types every map[string]any value,
// so a hand-written restore table in transform_clone.go has to repair the
// damage. A field added to a Cani*Type without a matching line there is silent,
// untyped data loss that no compiler catches — this test is the thing that
// catches it.
// Inputs: an inventory whose Source strings, ProviderDefaults, DeviceBays,
// interface ProviderMetadata, ObjectMeta maps and metadata defaults are all set.
// Outputs: a clone whose corresponding fields equal the originals.
// Data choice: the map values are ints, not strings, because JSON decodes a
// number into float64; a value that survives as int proves restoration ran
// rather than the decoder happening to agree.
func TestCloneInventoryRestoresEveryNonSerializedField(t *testing.T) {
	ids := newCloneFixtureIDs()
	source := newCloneFixture(ids)

	clone, err := cloneInventory(source)
	if err != nil {
		t.Fatalf("cloneInventory: %v", err)
	}

	for _, check := range []struct {
		name string
		got  string
		want string
	}{
		{"rack source", clone.Racks[ids.rack].Source, "fixture-rack"},
		{"device source", clone.Devices[ids.device].Source, "fixture-device"},
		{"module source", clone.Modules[ids.module].Source, "fixture-module"},
		{"cable source", clone.Cables[ids.cable].Source, "fixture-cable"},
		{"fru source", clone.Frus[ids.fru].Source, "fixture-fru"},
	} {
		if check.got != check.want {
			t.Errorf("%s = %q, want %q (json:\"-\" field lost by the clone)", check.name, check.got, check.want)
		}
	}

	if got := clone.Racks[ids.rack].ProviderDefaults["height"]; got != 42 {
		t.Errorf("rack ProviderDefaults[height] = %#v, want int 42", got)
	}
	if got := clone.Devices[ids.device].Interfaces[0].ProviderMetadata["speed"]; got != 400 {
		t.Errorf("device interface ProviderMetadata[speed] = %#v, want int 400", got)
	}
	if got := clone.Modules[ids.module].Interfaces[0].ProviderMetadata["speed"]; got != 400 {
		t.Errorf("module interface ProviderMetadata[speed] = %#v, want int 400", got)
	}
	if got := clone.Metadata.CustomFields[0].Default; got != 3 {
		t.Errorf("metadata custom field default = %#v, want int 3", got)
	}

	bay := clone.Devices[ids.device].DeviceBays[0]
	if got := bay.Extra["depth"]; got != 3 {
		t.Errorf("device bay Extra[depth] = %#v, want int 3", got)
	}
	if bay.Allowed == nil || bay.Default == nil {
		t.Fatalf("device bay slug refs lost: allowed=%v default=%v", bay.Allowed, bay.Default)
	}
	if got := bay.Default.Slug; got != "blade-a" {
		t.Errorf("device bay default slug = %#v, want \"blade-a\"", got)
	}

	meta := clone.Devices[ids.device].ObjectMeta
	if got := meta.CustomFields["count"]; got != 7 {
		t.Errorf("device CustomFields[count] = %#v, want int 7", got)
	}
	if got := meta.ProviderMetadata["csm"].(map[string]any)["slot"]; got != 1 {
		t.Errorf("device ProviderMetadata[csm][slot] = %#v, want int 1", got)
	}
	if got, want := meta.ExternalIDs["nautobot"], source.Devices[ids.device].ExternalIDs["nautobot"]; got != want {
		t.Errorf("device ExternalIDs[nautobot] = %v, want %v", got, want)
	}
}

// TestCloneInventoryDoesNotAliasTheSource verifies mutating a clone leaves the
// original untouched.
//
// Why it matters: MergeTransformResult only assigns the clone over the receiver
// once every step has succeeded, so a failed merge must leave the caller's
// inventory exactly as it was. That guarantee is worthless if the restore table
// copied a map or slice header instead of its contents, because the discarded
// clone would have written through to the original anyway.
// Inputs: a fully populated inventory, cloned, then mutated only via the clone.
// Outputs: the source's corresponding values are unchanged.
// Data choice: nested containers are mutated — a map inside a map, and a slice
// element — because a shallow copy passes a top-level check and fails here.
func TestCloneInventoryDoesNotAliasTheSource(t *testing.T) {
	ids := newCloneFixtureIDs()
	source := newCloneFixture(ids)

	clone, err := cloneInventory(source)
	if err != nil {
		t.Fatalf("cloneInventory: %v", err)
	}

	clone.Devices[ids.device].ProviderMetadata["csm"].(map[string]any)["xname"] = "mutated"
	clone.Devices[ids.device].CustomFields["count"] = 99
	clone.Devices[ids.device].Tags[0] = "mutated"
	clone.Devices[ids.device].Interfaces[0].ProviderMetadata["speed"] = 0
	clone.Racks[ids.rack].ProviderDefaults["height"] = 0
	clone.Devices[ids.device].DeviceBays[0].Extra["depth"] = 0

	original := source.Devices[ids.device]
	if got := original.ProviderMetadata["csm"].(map[string]any)["xname"]; got != "x3000c0s1b0n0" {
		t.Errorf("source ProviderMetadata was aliased: xname = %v", got)
	}
	if got := original.CustomFields["count"]; got != 7 {
		t.Errorf("source CustomFields was aliased: count = %v", got)
	}
	if got := original.Tags[0]; got != "device" {
		t.Errorf("source Tags was aliased: tag = %v", got)
	}
	if got := original.Interfaces[0].ProviderMetadata["speed"]; got != 400 {
		t.Errorf("source interface ProviderMetadata was aliased: speed = %v", got)
	}
	if got := source.Racks[ids.rack].ProviderDefaults["height"]; got != 42 {
		t.Errorf("source ProviderDefaults was aliased: height = %v", got)
	}
	if got := original.DeviceBays[0].Extra["depth"]; got != 3 {
		t.Errorf("source DeviceBays Extra was aliased: depth = %v", got)
	}
}

// TestCloneTransformResultRestoresEveryNonSerializedField verifies the
// TransformResult clone repairs the same damage as the inventory clone.
//
// Why it matters: MergeTransformResult clones both sides before touching
// anything, so provider output gets the same JSON round-trip as the inventory.
// The two restore paths are separate functions over separate structs, so one
// can be extended while the other is forgotten; only exercising both catches it.
// Inputs: a TransformResult carrying the same entities as the inventory fixture.
// Outputs: sources, provider defaults and interface metadata all survive.
// Data choice: TransformResult has no Interfaces map of its own, so interface
// metadata is asserted through the device that embeds it.
func TestCloneTransformResultRestoresEveryNonSerializedField(t *testing.T) {
	ids := newCloneFixtureIDs()
	fixture := newCloneFixture(ids)
	source := &TransformResult{
		Locations: fixture.Locations,
		Racks:     fixture.Racks,
		Devices:   fixture.Devices,
		Modules:   fixture.Modules,
		Cables:    fixture.Cables,
		Frus:      fixture.Frus,
		Metadata:  fixture.Metadata,
	}

	clone, err := cloneTransformResult(source)
	if err != nil {
		t.Fatalf("cloneTransformResult: %v", err)
	}

	if got := clone.Racks[ids.rack].Source; got != "fixture-rack" {
		t.Errorf("rack source = %q, want %q", got, "fixture-rack")
	}
	if got := clone.Devices[ids.device].Source; got != "fixture-device" {
		t.Errorf("device source = %q, want %q", got, "fixture-device")
	}
	if got := clone.Cables[ids.cable].Source; got != "fixture-cable" {
		t.Errorf("cable source = %q, want %q", got, "fixture-cable")
	}
	if got := clone.Frus[ids.fru].Source; got != "fixture-fru" {
		t.Errorf("fru source = %q, want %q", got, "fixture-fru")
	}
	if got := clone.Racks[ids.rack].ProviderDefaults["height"]; got != 42 {
		t.Errorf("rack ProviderDefaults[height] = %#v, want int 42", got)
	}
	if got := clone.Modules[ids.module].Interfaces[0].ProviderMetadata["speed"]; got != 400 {
		t.Errorf("module interface ProviderMetadata[speed] = %#v, want int 400", got)
	}
	if got := clone.Metadata.CustomFields[0].Default; got != 3 {
		t.Errorf("metadata custom field default = %#v, want int 3", got)
	}

	clone.Racks[ids.rack].ProviderDefaults["height"] = 0
	if got := source.Racks[ids.rack].ProviderDefaults["height"]; got != 42 {
		t.Errorf("source ProviderDefaults was aliased: height = %v", got)
	}
}

// TestCloneInventoryDropsDerivedStateAndRebuildsTheIndex verifies reverse
// indices are not carried across the clone, but the provider-key index is.
//
// Why it matters: reverse pointers are derived data that must be rebuilt from
// the forward FKs, never trusted from a previous state — that is the whole
// reason they carry json:"-". The provider-key index is different: it is a
// lookup cache the merge relies on immediately, so cloneInventory rebuilds it
// explicitly rather than leaving it empty.
// Inputs: an inventory whose reverse indices have been populated by hand, as a
// stale state would be.
// Outputs: a clone with the hand-set reverse index gone and a working
// provider-key lookup.
// Data choice: the reverse index is deliberately set to a wrong value so a
// clone that copied it would be distinguishable from one that dropped it.
func TestCloneInventoryDropsDerivedStateAndRebuildsTheIndex(t *testing.T) {
	ids := newCloneFixtureIDs()
	source := newCloneFixture(ids)
	source.Racks[ids.rack].Devices = []uuid.UUID{uuid.New()}
	source.Locations[ids.location].Racks = []uuid.UUID{uuid.New()}

	clone, err := cloneInventory(source)
	if err != nil {
		t.Fatalf("cloneInventory: %v", err)
	}

	if got := clone.Racks[ids.rack].Devices; len(got) != 0 {
		t.Errorf("rack reverse index survived the clone: %v", got)
	}
	if got := clone.Locations[ids.location].Racks; len(got) != 0 {
		t.Errorf("location reverse index survived the clone: %v", got)
	}

	if result := clone.RebuildDerivedState(); result.Err() != nil {
		t.Fatalf("rebuilding derived state on the clone: %v", result.Err())
	}
	if got := clone.Racks[ids.rack].Devices; len(got) != 1 || got[0] != ids.device {
		t.Errorf("rack devices after rebuild = %v, want [%v]", got, ids.device)
	}
}
