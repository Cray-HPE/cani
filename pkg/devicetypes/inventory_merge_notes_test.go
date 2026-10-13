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

// TestMergePropertiesCarriesCommentsAndAssetTag verifies that merging an
// incoming device onto an existing one copies Comments and AssetTag and
// reports the change, while an empty incoming value leaves the existing one.
//
// Why it matters: a re-import must carry an operator's Nautobot-side comment
// or asset tag edit onto the record cani already holds; NAUTOBOT_MAPPING.md
// marks both fields Mapped in both directions.
// Inputs: an existing device with old notes; one incoming copy with new notes,
// one with empty notes.
// Outputs: the first merge changes both fields and returns true; the second
// keeps them and returns false.
// Data choice: the two incoming copies differ only in the merged fields so a
// true from the other branches cannot mask a missed one.
func TestMergePropertiesCarriesCommentsAndAssetTag(t *testing.T) {
	existing := &CaniDeviceType{Name: "compute-002", Comments: "compute node 2", AssetTag: "AT-OLD"}

	if !existing.MergeProperties(&CaniDeviceType{Name: "compute-002", Comments: "edited again", AssetTag: "AT-NEW"}) {
		t.Fatal("expected a changed merge when comments and asset tag differ")
	}
	if existing.Comments != "edited again" || existing.AssetTag != "AT-NEW" {
		t.Fatalf("notes not carried: comments=%q assetTag=%q", existing.Comments, existing.AssetTag)
	}
	if existing.MergeProperties(&CaniDeviceType{Name: "compute-002"}) {
		t.Fatal("an incoming device without notes must not report a change")
	}
	if existing.Comments != "edited again" || existing.AssetTag != "AT-NEW" {
		t.Fatalf("empty incoming notes cleared the existing ones: comments=%q assetTag=%q", existing.Comments, existing.AssetTag)
	}
}

// TestMergeRacksCarriesCommentsAndAssetTag verifies that a rack merged by name
// takes the incoming Comments and AssetTag.
//
// Why it matters: racks follow the same import path as devices, and the
// import mapper fills both fields from Nautobot.
// Inputs: an existing rack named rack-01 with old notes; an incoming rack of
// the same name under a new UUID with new notes.
// Outputs: the incoming UUID is remapped onto the existing one and the
// existing rack carries the new notes.
// Data choice: no external IDs on either side so the match is by name alone.
func TestMergeRacksCarriesCommentsAndAssetTag(t *testing.T) {
	inventory := NewInventory()
	existingID, incomingID := uuid.New(), uuid.New()
	inventory.Racks[existingID] = &CaniRackType{ID: existingID, Name: "rack-01", Comments: "primary test rack", AssetTag: "R-OLD"}

	remap := inventory.MergeRacks(map[uuid.UUID]*CaniRackType{
		incomingID: {ID: incomingID, Name: "rack-01", Comments: "moved to row 2", AssetTag: "R-NEW"},
	})

	if remap[incomingID] != existingID {
		t.Fatalf("incoming rack not remapped onto the existing one: %v", remap)
	}
	rack := inventory.Racks[existingID]
	if rack.Comments != "moved to row 2" || rack.AssetTag != "R-NEW" {
		t.Fatalf("rack notes not carried: comments=%q assetTag=%q", rack.Comments, rack.AssetTag)
	}
}

// TestMergeLocationsCarriesComments verifies that a location merged by name
// takes the incoming Comments.
//
// Why it matters: Location.Comments is Mapped in both directions and was the
// one scalar the location merge left behind.
// Inputs: an existing location test-dc with a comment; an incoming copy under
// a new UUID with a different comment.
// Outputs: the incoming UUID is remapped and the existing location carries
// the new comment.
// Data choice: Description is left empty on the incoming copy to show the
// merge does not need another field to change.
func TestMergeLocationsCarriesComments(t *testing.T) {
	inventory := NewInventory()
	existingID, incomingID := uuid.New(), uuid.New()
	inventory.Locations[existingID] = &CaniLocationType{ID: existingID, Name: "test-dc", Comments: "old"}

	remap := inventory.MergeLocations(map[uuid.UUID]*CaniLocationType{
		incomingID: {ID: incomingID, Name: "test-dc", Comments: "edited in Nautobot"},
	})

	if remap[incomingID] != existingID {
		t.Fatalf("incoming location not remapped onto the existing one: %v", remap)
	}
	if got := inventory.Locations[existingID].Comments; got != "edited in Nautobot" {
		t.Fatalf("location comment not carried: %q", got)
	}
}
