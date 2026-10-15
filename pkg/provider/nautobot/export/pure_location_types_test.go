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
	"github.com/google/uuid"
)

// ---------- Exporter resolveLocationName ----------

// TestResolveLocationName verifies a location UUID resolves to its name from
// the inventory.
//
// Why it matters: IPAM objects (VLANs, prefixes) reference locations by name in
// Nautobot, so export must turn cani location UUIDs back into names.
// Inputs: an inventory with one location "Building-A" and its UUID.
// Outputs: the name and a nil error.
// Data choice: a single named location is the minimal fixture proving the
// happy-path lookup.
func TestResolveLocationName(t *testing.T) {
	exporter := &Exporter{}

	locID := uuid.New()
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			locID: {Name: "Building-A"},
		},
	}

	name, err := exporter.resolveLocationName(locID, inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Building-A" {
		t.Errorf("expected Building-A, got %s", name)
	}
}

// TestResolveLocationNameNotFound verifies that resolving an unknown location
// UUID returns an error.
//
// Why it matters: a dangling location reference must surface as an error so the
// export reports it rather than silently writing an empty location name.
// Inputs: an empty Locations map and a random UUID.
// Outputs: an error (name ignored).
// Data choice: the empty map guarantees the miss branch, complementing the
// found-case test above.
func TestResolveLocationNameNotFound(t *testing.T) {
	exporter := &Exporter{}

	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{},
	}

	_, err := exporter.resolveLocationName(uuid.New(), inv)
	if err == nil {
		t.Fatal("expected error for missing location")
	}
}

// ---------- locationTypeSupports ----------

// TestLocationTypeSupports verifies a registered location type reports support
// only for content types in its list, and unknown slugs return false.
//
// Why it matters: Nautobot rejects placing a rack/device under a location type
// that does not list that content type, so export must check before assigning.
// Inputs: a "test-room-export" type supporting rack+device; queried for rack,
// device, module, and an unknown slug.
// Outputs: true for rack/device, false for module and the unknown slug.
// Data choice: a type that supports two of three content types plus a bogus slug
// covers the match, no-match, and not-found branches.
func TestLocationTypeSupports(t *testing.T) {
	// Register a test location type.
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Test Room",
		Slug:         "test-room-export",
		ContentTypes: []string{"rack", "device"},
	})

	if !locationTypeSupports("test-room-export", "rack") {
		t.Error("expected test-room-export to support rack")
	}
	if !locationTypeSupports("test-room-export", "device") {
		t.Error("expected test-room-export to support device")
	}
	if locationTypeSupports("test-room-export", "module") {
		t.Error("expected test-room-export to NOT support module")
	}
	if locationTypeSupports("nonexistent-slug", "rack") {
		t.Error("expected unknown slug to return false")
	}
}

// ---------- toNautobotContentTypes ----------

// TestToNautobotContentTypes verifies that toNautobotContentTypes maps the short
// content-type names device/rack/module to their dcim.* equivalents, passes
// unknown values through unchanged, and preserves order and length.
//
// Why it matters: Nautobot APIs expect fully-qualified content types (e.g.
// "dcim.device"); exporting the wrong string would make content-type assignment
// on statuses/roles fail.
// Inputs (table): single known types, an unknown "ipam.prefix", a mixed slice,
// and an empty slice. Outputs: the converted slice for each case.
// Data choice: the cases cover every known mapping, the default passthrough, a
// mixed combination, and the empty edge case in one table.
func TestToNautobotContentTypes(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		expect []string
	}{
		{"device", []string{"device"}, []string{"dcim.device"}},
		{"rack", []string{"rack"}, []string{"dcim.rack"}},
		{"module", []string{"module"}, []string{"dcim.module"}},
		{"unknown passthrough", []string{"ipam.prefix"}, []string{"ipam.prefix"}},
		{"mixed", []string{"device", "rack", "foo"}, []string{"dcim.device", "dcim.rack", "foo"}},
		{"empty", []string{}, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toNautobotContentTypes(tt.input)
			if len(got) != len(tt.expect) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.expect))
			}
			for i := range got {
				if got[i] != tt.expect[i] {
					t.Errorf("[%d] = %q, want %q", i, got[i], tt.expect[i])
				}
			}
		})
	}
}

// ---------- parentDef ----------

// TestParentDefFound verifies that parentDef returns a non-nil location-type
// definition (with matching Slug) for a registered slug.
//
// Why it matters: location-tree resolution during export looks up parent
// location-type definitions by slug; a registered slug must resolve so hierarchy
// walks work.
// Inputs: the registered slug "parent-test-type". Outputs: a non-nil
// *LocationTypeDefinition whose Slug equals the input.
// Data choice: registering a unique type first guarantees a deterministic hit
// without depending on globally pre-registered types.
func TestParentDefFound(t *testing.T) {
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Parent Test Type",
		Slug:         "parent-test-type",
		ContentTypes: []string{"rack"},
	})

	got := parentDef("parent-test-type")
	if got == nil {
		t.Fatal("expected non-nil definition")
	}
	if got.Slug != "parent-test-type" {
		t.Errorf("slug = %q", got.Slug)
	}
}

// TestParentDefNotFound verifies that parentDef returns nil for a slug that is
// not registered.
//
// Why it matters: an unknown location-type slug must not resolve to a bogus
// definition during export; nil lets callers handle the missing-type case.
// Inputs: the unregistered slug "nonexistent-slug-xyz". Outputs: a nil
// *LocationTypeDefinition.
// Data choice: a deliberately implausible slug ensures it is absent from the
// global registry regardless of other registered types.
func TestParentDefNotFound(t *testing.T) {
	got := parentDef("nonexistent-slug-xyz")
	if got != nil {
		t.Errorf("expected nil for unknown slug, got %+v", got)
	}
}

// ---------- resolveLocationName (Exporter method in load_vlans.go) ----------

// TestResolveLocationNameFromExporter verifies Exporter.resolveLocationName
// returns the location's name for a UUID present in the inventory.
//
// Why it matters: VLAN/prefix export scopes objects to a location by name, so
// resolving the cani location UUID to its Nautobot name must succeed.
// Inputs: a location UUID and an inventory containing that location ("Floor-42").
// Outputs: the resolved name string and a nil error.
// Data choice: a single named location keeps the happy path unambiguous.
func TestResolveLocationNameFromExporter(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}

	locID := uuid.New()
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			locID: {ID: locID, Name: "Floor-42"},
		},
	}

	name, err := e.resolveLocationName(locID, inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Floor-42" {
		t.Errorf("got %q, want 'Floor-42'", name)
	}
}

// TestResolveLocationNameFromExporterNotFound verifies
// Exporter.resolveLocationName returns an error when the UUID is absent from the
// inventory.
//
// Why it matters: a dangling location reference must fail loudly during export
// instead of silently scoping data to the wrong (empty) location.
// Inputs: a random UUID and an inventory with an empty Locations map.
// Outputs: a non-nil error.
// Data choice: the empty map guarantees the lookup misses, isolating the error
// branch.
func TestResolveLocationNameFromExporterNotFound(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{},
	}

	_, err := e.resolveLocationName(uuid.New(), inv)
	if err == nil {
		t.Fatal("expected error for unknown location ID")
	}
}

// ---------- topologicalSortLocations additional cases ----------

// TestTopologicalSortLocationsEmpty verifies topologicalSortLocations returns an
// empty slice for an empty location map.
//
// Why it matters: locations are created parent-first in Nautobot, and the empty
// case must not panic or fabricate entries when there is nothing to export.
// Inputs: an empty map[uuid.UUID]*CaniLocationType. Outputs: a length-0 slice.
// Data choice: the empty map is the boundary condition for the sort.
func TestTopologicalSortLocationsEmpty(t *testing.T) {
	locs := map[uuid.UUID]*devicetypes.CaniLocationType{}
	sorted := topologicalSortLocations(locs)
	if len(sorted) != 0 {
		t.Errorf("expected 0, got %d", len(sorted))
	}
}

// TestTopologicalSortLocationsSingleRoot verifies topologicalSortLocations
// returns a single root location unchanged.
//
// Why it matters: a lone top-level location (Parent=uuid.Nil) must export as
// itself with no ordering surprises.
// Inputs: a one-entry map with a root location named "Root". Outputs: a
// one-element slice containing that location.
// Data choice: a single Nil-parent node is the simplest non-empty graph.
func TestTopologicalSortLocationsSingleRoot(t *testing.T) {
	id := uuid.New()
	locs := map[uuid.UUID]*devicetypes.CaniLocationType{
		id: {ID: id, Name: "Root", Parent: uuid.Nil},
	}
	sorted := topologicalSortLocations(locs)
	if len(sorted) != 1 {
		t.Fatalf("expected 1, got %d", len(sorted))
	}
	if sorted[0].Name != "Root" {
		t.Errorf("Name = %q, want 'Root'", sorted[0].Name)
	}
}

// TestTopologicalSortLocationsParentBeforeChild verifies
// topologicalSortLocations orders a parent ahead of its child regardless of map
// iteration order.
//
// Why it matters: Nautobot rejects a child location whose parent does not yet
// exist, so the export order must always place parents first.
// Inputs: a two-entry map (child inserted first) where Child.Parent == Parent.ID.
// Outputs: a slice with Parent at index 0 and Child at index 1.
// Data choice: inserting the child first defeats any accidental reliance on map
// ordering, proving the topo-sort actually reorders.
func TestTopologicalSortLocationsParentBeforeChild(t *testing.T) {
	parentID := uuid.New()
	childID := uuid.New()
	locs := map[uuid.UUID]*devicetypes.CaniLocationType{
		childID:  {ID: childID, Name: "Child", Parent: parentID},
		parentID: {ID: parentID, Name: "Parent", Parent: uuid.Nil},
	}
	sorted := topologicalSortLocations(locs)
	if len(sorted) != 2 {
		t.Fatalf("expected 2, got %d", len(sorted))
	}
	if sorted[0].Name != "Parent" {
		t.Errorf("expected Parent first, got %q", sorted[0].Name)
	}
	if sorted[1].Name != "Child" {
		t.Errorf("expected Child second, got %q", sorted[1].Name)
	}
}
