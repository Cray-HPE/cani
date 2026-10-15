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

// ---------- generateDeviceNames (additional branches) ----------

// TestGenerateDeviceNamesAllBranches verifies name generation picks serial,
// then slug, then model, then UUID, leaves already-named devices alone, and
// names "system"-typed devices via the UUID fallback.
//
// Why it matters: Nautobot devices require a name; unnamed cani devices must get
// a deterministic, identifiable "cani-" name or they cannot be created.
// Inputs: six devices each missing one identifier tier (plus a named one).
// Outputs: in-place Name fields following the priority order.
// Data choice: one device per branch proves the precedence; "ProLiant DL380"
// checks the model lowercasing/space-to-dash transform.
func TestGenerateDeviceNamesAllBranches(t *testing.T) {
	nodeID := uuid.New()
	slugID := uuid.New()
	modelID := uuid.New()
	uuidID := uuid.New()
	namedID := uuid.New()
	systemID := uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			nodeID:   {ID: nodeID, Name: "", Type: "node", Serial: "SN-ABC"},
			slugID:   {ID: slugID, Name: "", Type: "node", Slug: "hpe-dl360"},
			modelID:  {ID: modelID, Name: "", Type: "node", Model: "ProLiant DL380"},
			uuidID:   {ID: uuidID, Name: "", Type: "node"},
			namedID:  {ID: namedID, Name: "already-named", Type: "node"},
			systemID: {ID: systemID, Name: "", Type: "system"},
		},
	}

	generateDeviceNames(inv)

	if inv.Devices[nodeID].Name != "cani-SN-ABC" {
		t.Errorf("serial branch: got %q", inv.Devices[nodeID].Name)
	}
	if inv.Devices[slugID].Name != "cani-hpe-dl360" {
		t.Errorf("slug branch: got %q", inv.Devices[slugID].Name)
	}
	if inv.Devices[modelID].Name != "cani-proliant-dl380" {
		t.Errorf("model branch: got %q", inv.Devices[modelID].Name)
	}
	if inv.Devices[uuidID].Name != "cani-"+uuidID.String()[:8] {
		t.Errorf("uuid branch: got %q", inv.Devices[uuidID].Name)
	}
	if inv.Devices[namedID].Name != "already-named" {
		t.Errorf("named device should not change: got %q", inv.Devices[namedID].Name)
	}
	// "system" also classifies as CategoryDevice, so it gets a UUID-based name.
	if inv.Devices[systemID].Name != "cani-"+systemID.String()[:8] {
		t.Errorf("system device: got %q", inv.Devices[systemID].Name)
	}
}

// ---------- disambiguateDeviceNames (additional cases) ----------

// TestDisambiguateDeviceNamesAllSuffixes verifies duplicate names are made
// unique via serial, rack-position, or UUID suffixes (in that priority).
//
// Why it matters: Nautobot enforces device-name uniqueness per location/tenant,
// so colliding cani names must be disambiguated before export or the second
// create fails.
// Inputs: three devices all named "dup-server" with differing identifiers and a
// rack for the position suffix.
// Outputs: three now-unique names with the expected suffix formats.
// Data choice: giving each duplicate a different identifier forces all three
// suffix branches and the post-check asserts no name remains duplicated.
func TestDisambiguateDeviceNamesAllSuffixes(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()
	rackID := uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id1: {ID: id1, Name: "dup-server", Type: "node", Serial: "SN001"},
			id2: {ID: id2, Name: "dup-server", Type: "node", RackPosition: 5, Rack: rackID},
			id3: {ID: id3, Name: "dup-server", Type: "node"},
		},
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {Name: "Rack-A"},
		},
	}

	disambiguateDeviceNames(inv)

	// All names should now be unique.
	names := make(map[string]int)
	for _, d := range inv.Devices {
		names[d.Name]++
	}
	for name, count := range names {
		if count > 1 {
			t.Errorf("name %q still appears %d times after disambiguation", name, count)
		}
	}

	// Serial suffix
	if inv.Devices[id1].Name != "dup-server (SN001)" {
		t.Errorf("serial suffix: got %q", inv.Devices[id1].Name)
	}
	// Rack position suffix
	if inv.Devices[id2].Name != "dup-server (Rack-A U5)" {
		t.Errorf("rack position suffix: got %q", inv.Devices[id2].Name)
	}
	// UUID suffix
	expected := "dup-server (" + id3.String()[:8] + ")"
	if inv.Devices[id3].Name != expected {
		t.Errorf("uuid suffix: got %q, want %q", inv.Devices[id3].Name, expected)
	}
}

// TestDisambiguateDeviceNamesNoDuplicates verifies devices with already-unique
// names are left unchanged.
//
// Why it matters: disambiguation must not rewrite names that are already valid,
// which would churn Nautobot records and break stable identity across exports.
// Inputs: two devices with distinct names "server-1" and "server-2".
// Outputs: both names unchanged.
// Data choice: two unique names exercise the len<=1 group skip, the inverse of
// the all-suffixes test.
func TestDisambiguateDeviceNamesNoDuplicates(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id1: {ID: id1, Name: "server-1", Type: "node"},
			id2: {ID: id2, Name: "server-2", Type: "node"},
		},
	}

	disambiguateDeviceNames(inv)

	// Names should not be modified.
	if inv.Devices[id1].Name != "server-1" {
		t.Errorf("non-duplicate should not change: got %q", inv.Devices[id1].Name)
	}
	if inv.Devices[id2].Name != "server-2" {
		t.Errorf("non-duplicate should not change: got %q", inv.Devices[id2].Name)
	}
}

// ---------- generateDeviceNames with Model path ----------

// TestGenerateDeviceNamesModelPath verifies generateDeviceNames derives a name
// from the device Model when Name is empty, slugified under a "cani-" prefix.
//
// Why it matters: Nautobot requires a device name, so unnamed cani devices need
// a deterministic, human-readable fallback before export.
// Inputs: a node device with empty Name and Model="HPE DL380 Gen10".
// Outputs: the device's Name set to "cani-hpe-dl380-gen10" in place.
// Data choice: a spaced, mixed-case model proves the lowercase/space-to-dash
// slugging.
func TestGenerateDeviceNamesModelPath(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id: {ID: id, Name: "", Type: "node", Model: "HPE DL380 Gen10"},
		},
	}
	generateDeviceNames(inv)
	expected := "cani-hpe-dl380-gen10"
	if inv.Devices[id].Name != expected {
		t.Errorf("expected %q, got %q", expected, inv.Devices[id].Name)
	}
}

// TestGenerateDeviceNamesSkipsNonDevice verifies generateDeviceNames leaves the
// Name untouched for inventory entries that are not the device category.
//
// Why it matters: only Nautobot "device" objects need generated names; racks and
// other categories must not be renamed by this pass.
// Inputs: an entry with Type="rack" and empty Name. Outputs: the Name stays empty
// (the entry is skipped).
// Data choice: Type="rack" exercises the ClassifyForNautobot non-device branch.
func TestGenerateDeviceNamesSkipsNonDevice(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id: {ID: id, Name: "", Type: "rack", Model: "should-not-change"},
		},
	}
	generateDeviceNames(inv)
	if inv.Devices[id].Name != "" {
		t.Errorf("expected empty name for non-device, got %q", inv.Devices[id].Name)
	}
}

// TestGenerateDeviceNamesUUIDFallback verifies generateDeviceNames falls back to
// the UUID prefix when a device has no serial, slug, or model.
//
// Why it matters: even a device with no descriptive fields must still receive a
// unique, stable name so the export can create it in Nautobot.
// Inputs: a node device with empty Name and no serial/slug/model.
// Outputs: Name set to "cani-" plus the first 8 chars of the device UUID.
// Data choice: omitting every name source forces the last-resort UUID branch.
func TestGenerateDeviceNamesUUIDFallback(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id: {ID: id, Name: "", Type: "node"},
		},
	}
	generateDeviceNames(inv)
	expected := "cani-" + id.String()[:8]
	if inv.Devices[id].Name != expected {
		t.Errorf("expected %q, got %q", expected, inv.Devices[id].Name)
	}
}

// ---------- disambiguateDeviceNames — RackPosition without rack name ----------

// TestDisambiguateDeviceNamesRackPositionNoName verifies disambiguateDeviceNames
// makes duplicate names unique using the rack position when the rack is unnamed.
//
// Why it matters: Nautobot enforces name uniqueness per location, so colliding
// device names must be disambiguated before export.
// Inputs: two "dup-node" devices at positions 3 and 7 sharing an unnamed rack.
// Outputs: the two devices end with different Names (position suffixes).
// Data choice: no serials plus an empty rack name forces the rack-position
// branch rather than the serial or index suffix.
func TestDisambiguateDeviceNamesRackPositionNoName(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	rackID := uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id1: {ID: id1, Name: "dup-node", Type: "node", RackPosition: 3, Rack: rackID},
			id2: {ID: id2, Name: "dup-node", Type: "node", RackPosition: 7, Rack: rackID},
		},
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {Name: ""},
		},
	}

	disambiguateDeviceNames(inv)

	// Both should have unique names with position suffix
	if inv.Devices[id1].Name == inv.Devices[id2].Name {
		t.Errorf("names should be unique, both are %q", inv.Devices[id1].Name)
	}
}
