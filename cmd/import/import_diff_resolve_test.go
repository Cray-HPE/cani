/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024, 2026 Hewlett Packard Enterprise Development LP
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
package imprt

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// diffInventory builds an inventory holding rack-a under its local ID and
// node-1 in it, plus a transform result reporting rack-a under the source
// system's ID and rack-b under another.
func diffInventory() (*devicetypes.CaniDeviceType, parentNamer, uuid.UUID, uuid.UUID) {
	inv := devicetypes.NewInventory()
	localRack, srcRack, otherRack := uuid.New(), uuid.New(), uuid.New()
	inv.Racks[localRack] = &devicetypes.CaniRackType{ID: localRack, Name: "rack-a"}
	existing := &devicetypes.CaniDeviceType{Name: "node-1", Parent: localRack}
	result := &devicetypes.TransformResult{Racks: map[uuid.UUID]*devicetypes.CaniRackType{
		srcRack:   {ID: srcRack, Name: "rack-a"},
		otherRack: {ID: otherRack, Name: "rack-b"},
	}}
	return existing, parentNamer{inv: inv, result: result}, srcRack, otherRack
}

// TestComputeDeviceDiffsParentByName verifies the parent is compared by name,
// so the same rack under the source system's ID is not a change while a
// different rack is.
//
// Why it matters: before the merge remaps IDs every device showed a spurious
// "parent: <local id> --> <source id>" line (PR #281 review).
// Inputs: node-1 in rack-a locally; incoming parent rack-a (source ID), then
// rack-b. Expected: no diff, then "parent: rack-a --> rack-b".
func TestComputeDeviceDiffsParentByName(t *testing.T) {
	existing, names, sameRack, otherRack := diffInventory()

	same := computeDeviceDiffs(existing, &devicetypes.CaniDeviceType{Parent: sameRack}, names)
	moved := computeDeviceDiffs(existing, &devicetypes.CaniDeviceType{Parent: otherRack}, names)

	if len(same) != 0 {
		t.Errorf("same rack under source ID reported: %+v", same)
	}
	want := fieldDiff{"parent", "rack-a", "rack-b"}
	if len(moved) != 1 || moved[0] != want {
		t.Errorf("moved = %+v, want [%+v]", moved, want)
	}
}

// TestComputeDeviceDiffsCustomFields verifies new and changed custom field
// values are listed per key and unchanged ones are not.
//
// Why it matters: the import preview must show custom field edits made in the
// source system before they are merged.
// Inputs: existing keep=1, x=old; incoming keep=1, x=new, y=z.
// Expected: customFields.x old->new, customFields.y <unset>->z, in key order.
func TestComputeDeviceDiffsCustomFields(t *testing.T) {
	existing := &devicetypes.CaniDeviceType{}
	existing.CustomFields = map[string]any{"keep": 1, "x": "old"}
	dev := &devicetypes.CaniDeviceType{}
	dev.CustomFields = map[string]any{"keep": 1, "x": "new", "y": "z"}

	diffs := computeDeviceDiffs(existing, dev, parentNamer{})

	want := []fieldDiff{{"customFields.x", "old", "new"}, {"customFields.y", "<unset>", "z"}}
	if len(diffs) != len(want) || diffs[0] != want[0] || diffs[1] != want[1] {
		t.Errorf("diffs = %+v, want %+v", diffs, want)
	}
}
