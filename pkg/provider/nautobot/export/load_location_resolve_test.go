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

// ---------- findContentChild ----------

// TestFindContentChild verifies the recursive search returns a child location
// whose type supports the requested content, and "" when none does.
//
// Why it matters: cani hierarchies often attach racks/devices to a generic
// parent (e.g. Building) that cannot hold them; export must descend to a child
// (e.g. Room) that can, or Nautobot placement fails.
// Inputs: a Building parent (no content types) with a Room child supporting
// rack+device.
// Outputs: "Server-Room" for rack; "" for an unsupported content type.
// Data choice: a two-level building->room tree is the minimal shape that forces
// a descent into the child.
func TestFindContentChild(t *testing.T) {
	// Register location types for the test.
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Building",
		Slug:         "building-test-fc",
		ContentTypes: []string{},
	})
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Room",
		Slug:         "room-test-fc",
		ContentTypes: []string{"rack", "device"},
	})

	parentID := uuid.New()
	childID := uuid.New()

	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			parentID: {
				ID:           parentID,
				Name:         "HQ",
				LocationType: "building-test-fc",
				Children:     []uuid.UUID{childID},
			},
			childID: {
				ID:           childID,
				Name:         "Server-Room",
				LocationType: "room-test-fc",
			},
		},
	}

	parent := inv.Locations[parentID]
	got := findContentChild(parent, "rack", inv)
	if got != "Server-Room" {
		t.Errorf("expected Server-Room, got %q", got)
	}

	// Non-matching content type
	got = findContentChild(parent, "nonexistent", inv)
	if got != "" {
		t.Errorf("expected empty string for unsupported content type, got %q", got)
	}
}

// ---------- resolveContentLocation ----------

// TestResolveContentLocation verifies content-location resolution: it descends
// to a supporting child, returns "" for nil/Nil inputs, and returns a location
// that directly supports the content type.
//
// Why it matters: this is the top-level helper that picks the correct Nautobot
// location for a rack/device, including the fallback walk; wrong results cause
// misplacement or export errors.
// Inputs: a DC parent with a Section child; queried by DC id, nil inventory,
// Nil UUID, and the Section id directly.
// Outputs: "Section-A" for the descent and direct cases, "" for nil/Nil.
// Data choice: the DC->Section tree plus the two empty inputs cover the descent,
// guard, and direct-support branches in one test.
func TestResolveContentLocation(t *testing.T) {
	// Register location types.
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "DC",
		Slug:         "dc-test-rcl",
		ContentTypes: []string{},
	})
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Section",
		Slug:         "section-test-rcl",
		ContentTypes: []string{"rack", "device"},
	})

	dcID := uuid.New()
	sectionID := uuid.New()

	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			dcID: {
				ID:           dcID,
				Name:         "DC-01",
				LocationType: "dc-test-rcl",
				Children:     []uuid.UUID{sectionID},
			},
			sectionID: {
				ID:           sectionID,
				Name:         "Section-A",
				LocationType: "section-test-rcl",
			},
		},
	}

	// DC itself doesn't support "rack", but its child does.
	got := resolveContentLocation(dcID, "rack", inv)
	if got != "Section-A" {
		t.Errorf("expected Section-A, got %q", got)
	}

	// nil inventory
	got = resolveContentLocation(dcID, "rack", nil)
	if got != "" {
		t.Errorf("expected empty for nil inventory, got %q", got)
	}

	// Nil UUID
	got = resolveContentLocation(uuid.Nil, "rack", inv)
	if got != "" {
		t.Errorf("expected empty for nil UUID, got %q", got)
	}

	// Location that directly supports content type.
	got = resolveContentLocation(sectionID, "rack", inv)
	if got != "Section-A" {
		t.Errorf("expected Section-A for direct support, got %q", got)
	}
}

// TestResolveContentLocationNilInventory verifies that resolveContentLocation
// returns "" when the inventory pointer is nil.
//
// Why it matters: the exporter calls this helper while resolving where to place
// a device/rack; a nil-inventory guard prevents a panic and lets the caller fall
// back to its default location instead of crashing the export.
// Inputs: a random location UUID, content type "device", and a nil *Inventory.
// Outputs: an empty string.
// Data choice: a nil inventory is the minimal fixture that triggers the early
// guard clause without needing any location data.
func TestResolveContentLocationNilInventory(t *testing.T) {
	got := resolveContentLocation(uuid.New(), "device", nil)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// TestResolveContentLocationNilID verifies that resolveContentLocation returns
// "" when the supplied location ID is uuid.Nil.
//
// Why it matters: devices/racks without a resolved location UUID must not match
// an arbitrary entry; returning empty lets the exporter fall back to a default.
// Inputs: uuid.Nil, content type "device", and an Inventory with an empty
// Locations map. Outputs: an empty string.
// Data choice: uuid.Nil plus an empty map isolates the nil-ID guard from the
// "not found in map" path tested separately.
func TestResolveContentLocationNilID(t *testing.T) {
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{},
	}
	got := resolveContentLocation(uuid.Nil, "device", inv)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// TestResolveContentLocationNotFound verifies that resolveContentLocation
// returns "" when the location ID is not present in the inventory map.
//
// Why it matters: a stale or dangling location reference must not silently
// resolve to the wrong place; empty output signals the caller to use a default.
// Inputs: a random UUID absent from the map, content type "device", and an
// Inventory with an empty Locations map. Outputs: an empty string.
// Data choice: a fresh UUID guaranteed not in the empty map exercises the
// map-miss branch distinctly from the nil-ID branch.
func TestResolveContentLocationNotFound(t *testing.T) {
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{},
	}
	got := resolveContentLocation(uuid.New(), "device", inv)
	if got != "" {
		t.Errorf("expected empty for missing location, got %q", got)
	}
}

// TestResolveContentLocationDirect verifies that resolveContentLocation returns
// the location's own name when that location's type directly supports the
// requested content type.
//
// Why it matters: devices must be placed at a Nautobot location whose type
// allows "device"; the direct-support case is the common happy path for export.
// Inputs: a location of type "server-room-rcl" (ContentTypes include "device"),
// its UUID, and content type "device". Outputs: the location name "Room-A-RCL".
// Data choice: a registered location type that lists "device" lets the helper
// short-circuit on direct support without walking any children.
func TestResolveContentLocationDirect(t *testing.T) {
	// Register a type that directly supports "device"
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Server Room RCL",
		Slug:         "server-room-rcl",
		ContentTypes: []string{"rack", "device"},
	})

	locID := uuid.New()
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			locID: {
				ID:           locID,
				Name:         "Room-A-RCL",
				LocationType: "server-room-rcl",
			},
		},
	}

	got := resolveContentLocation(locID, "device", inv)
	if got != "Room-A-RCL" {
		t.Errorf("expected 'Room-A-RCL', got %q", got)
	}
}

// TestResolveContentLocationChildResolution verifies that resolveContentLocation
// walks the location tree and returns a descendant's name when the parent's own
// type does not support the content type but a child's type does.
//
// Why it matters: cani hierarchies often place racks/devices at a child location
// (e.g. a floor) under a parent (e.g. a building) that cannot hold them; the
// exporter must resolve down to the valid child.
// Inputs: a parent "building-rcl" (no content types) linking a child "floor-rcl"
// (supports rack/device), the parent UUID, and content type "rack". Outputs: the
// child name "Floor-2-RCL".
// Data choice: an empty parent plus a supporting child models the real
// building->floor nesting and forces the depth-first child search.
func TestResolveContentLocationChildResolution(t *testing.T) {
	// Parent type doesn't support "rack", child type does
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Building RCL",
		Slug:         "building-rcl",
		ContentTypes: []string{},
	})
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Floor RCL",
		Slug:         "floor-rcl",
		ContentTypes: []string{"rack", "device"},
	})

	parentID := uuid.New()
	childID := uuid.New()

	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			parentID: {
				ID:           parentID,
				Name:         "Building-1-RCL",
				LocationType: "building-rcl",
				Children:     []uuid.UUID{childID},
			},
			childID: {
				ID:           childID,
				Name:         "Floor-2-RCL",
				LocationType: "floor-rcl",
			},
		},
	}

	got := resolveContentLocation(parentID, "rack", inv)
	if got != "Floor-2-RCL" {
		t.Errorf("expected 'Floor-2-RCL', got %q", got)
	}
}

// TestResolveContentLocationNoMatch verifies that resolveContentLocation returns
// "" when neither the location nor any descendant supports the content type.
//
// Why it matters: if no part of the hierarchy can legally hold a device, the
// exporter must fall back to a default rather than force an invalid placement.
// Inputs: a single location of type "campus-rcl" (empty ContentTypes, no
// children), its UUID, and content type "device". Outputs: an empty string.
// Data choice: a childless location whose type supports nothing exercises the
// exhausted-search branch that yields the empty fallback signal.
func TestResolveContentLocationNoMatch(t *testing.T) {
	// Type that supports nothing
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Campus RCL",
		Slug:         "campus-rcl",
		ContentTypes: []string{},
	})

	locID := uuid.New()
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			locID: {
				ID:           locID,
				Name:         "Campus-RCL",
				LocationType: "campus-rcl",
			},
		},
	}

	got := resolveContentLocation(locID, "device", inv)
	if got != "" {
		t.Errorf("expected empty (no descendant supports device), got %q", got)
	}
}

// ---------- resolveContentLocation sibling path ----------

// TestResolveContentLocationSiblingResolution verifies resolveContentLocation
// returns a same-name sibling's device-capable child when the primary location
// lacks a supporting child.
//
// Why it matters: when a location can't host a content type, export tries a
// same-named sibling's subtree before falling back to a default.
// Inputs: two "DC-Main" locations (one childless, one with a "row" child) and a
// target contentType="device". Outputs: "Row-1".
// Data choice: duplicate names with differing hierarchies trigger the sibling
// walk while unique test-only location-type slugs keep the fixture independent
// of embedded YAML contents.
func TestResolveContentLocationSiblingResolution(t *testing.T) {
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Sibling DC RCL",
		Slug:         "sibling-dc-rcl",
		ContentTypes: []string{},
	})
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Sibling Row RCL",
		Slug:         "sibling-row-rcl",
		ContentTypes: []string{"device"},
	})

	locID := uuid.New()
	siblingID := uuid.New()
	childID := uuid.New()

	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			locID: {
				ID:           locID,
				Name:         "DC-Main",
				LocationType: "sibling-dc-rcl",
				Children:     []uuid.UUID{},
			},
			siblingID: {
				ID:           siblingID,
				Name:         "DC-Main",
				LocationType: "sibling-dc-rcl",
				Children:     []uuid.UUID{childID},
			},
			childID: {
				ID:           childID,
				Name:         "Row-1",
				LocationType: "sibling-row-rcl",
			},
		},
	}

	result := resolveContentLocation(locID, "device", inv)
	if result != "Row-1" {
		t.Errorf("resolveContentLocation() = %q, want Row-1", result)
	}
}
