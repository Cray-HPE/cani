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

// ---------- compareRack ----------

// TestCompareRackBothNil verifies that compareRack reports no diff when the
// local device has no rack (Rack=Nil) and the remote Nautobot device also has a
// nil Rack.
//
// Why it matters: export should not PATCH a device's rack when neither side
// assigns one; a false diff would churn unracked devices on every run.
// Inputs: a device with Rack=Nil and a remote Device with Rack=nil.
// Outputs: an empty diff slice.
// Data choice: both-nil is the symmetric no-op case and avoids the
// GetRackByName API call, which would panic on this nil client.
func TestCompareRackBothNil(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{},
	}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{Rack: uuid.Nil}
	remote := &nautobotapi.Device{Rack: nil}
	diffs := compareRack(dev, remote, mapper)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when both nil, got %+v", diffs)
	}
}

// TestCompareRackLocalHasRack verifies that resolveLocalRackName finds the
// local rack name "Rack-A" when the device is assigned to an inventory rack.
//
// Why it matters: compareRack relies on this name to look up the remote rack
// UUID; the test documents that full compareRack coverage needs a mocked API
// because GetRackByName panics on a nil client.
// Inputs: a device whose Rack UUID maps to inventory rack "Rack-A".
// Outputs: "Rack-A".
// Data choice: a nil API client forces the test to assert only name
// resolution. NOTE: despite its name it does not actually call compareRack.
func TestCompareRackLocalHasRack(t *testing.T) {
	// When the local device has a rack but the API client is nil,
	// GetRackByName panics. This scenario requires a real or mocked API.
	// We verify here that resolveLocalRackName correctly finds the rack name.
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	rackID := uuid.New()
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {Name: "Rack-A"},
		},
	}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{Rack: rackID}
	name := resolveLocalRackName(dev, mapper)
	if name != "Rack-A" {
		t.Errorf("expected 'Rack-A', got %q", name)
	}
}

// ---------- compareRack remote has rack, local doesn't ----------

// TestCompareRackRemoteHasRackLocalDoesNot verifies compareRack emits one
// "rack" FieldDiff when the remote Nautobot device has a rack but the local
// device resolves to none.
//
// Why it matters: drift detection must flag a rack that exists remotely but not
// locally so merges can reconcile placement.
// Inputs: a parent-less CaniDeviceType (local rack "") and a remote Device
// carrying a rack id union. Outputs: a single diff whose Field == "rack".
// Data choice: a nil inventory yields an empty local rack name, so GetRackByName
// is skipped and only the remote-vs-none branch is tested.
func TestCompareRackRemoteHasRackLocalDoesNot(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	// Device with no parent → local rack is ""
	dev := &devicetypes.CaniDeviceType{Name: "no-rack-local"}

	// Remote device has a rack with an ID
	rackID := uuid.New()
	remote := &nautobotapi.Device{}
	mustSetRefID(t, &remote.Rack, rackID)
	diffs := compareRack(dev, remote, mapper)
	// Local is nil, remote is non-nil → should produce a diff
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Field != "rack" {
		t.Errorf("field = %q, want 'rack'", diffs[0].Field)
	}
}

// ---------- comparePosition ----------

// TestComparePositionNoLocalPosition verifies comparePosition returns no diffs
// when the local RackPosition is 0 ("no position").
//
// Why it matters: a missing local position must not flag or overwrite the
// remote position during export, avoiding spurious drift.
// Inputs: a device with RackPosition 0 and a remote with no Position. Outputs:
// an empty diff slice.
// Data choice: 0 is the sentinel for "unset", directly exercising the
// early-return guard.
func TestComparePositionNoLocalPosition(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Name:         "no-pos",
		RackPosition: 0, // zero means "no position"
	}
	remote := &nautobotapi.Device{}
	diffs := comparePosition(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when local position is 0, got %d", len(diffs))
	}
}

// TestComparePositionMismatch verifies comparePosition emits one "position"
// diff (LocalVal "5") when local and remote rack positions differ.
//
// Why it matters: position drift must be surfaced so a device's U-slot in
// Nautobot matches the cani inventory.
// Inputs: local RackPosition 5 vs. remote Position 3. Outputs: a single diff
// with Field "position" and LocalVal "5".
// Data choice: distinct non-zero numbers (5 vs 3) make both the mismatch and
// the string-formatted LocalVal unambiguous.
func TestComparePositionMismatch(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Name:         "pos-mismatch",
		RackPosition: 5,
	}
	remotePos := 3
	remote := &nautobotapi.Device{
		Position: &remotePos,
	}
	diffs := comparePosition(dev, remote)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Field != "position" {
		t.Errorf("field = %q, want 'position'", diffs[0].Field)
	}
	if diffs[0].LocalVal != "5" {
		t.Errorf("LocalVal = %q, want '5'", diffs[0].LocalVal)
	}
}

// TestComparePositionMatch verifies comparePosition returns no diffs when the
// local and remote positions are equal.
//
// Why it matters: matching positions must not be reported as drift, keeping
// export summaries free of false positives.
// Inputs: local RackPosition 7 and remote Position 7. Outputs: an empty diff
// slice.
// Data choice: identical non-zero values exercise the equal branch past the
// zero-position guard.
func TestComparePositionMatch(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Name:         "pos-match",
		RackPosition: 7,
	}
	remotePos := 7
	remote := &nautobotapi.Device{
		Position: &remotePos,
	}
	diffs := comparePosition(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when positions match, got %d", len(diffs))
	}
}

// ---------- compareFace ----------

// TestCompareFaceNoLocalFace verifies compareFace returns no diffs when the
// local Face is empty ("no face").
//
// Why it matters: an unset face must not flag drift or clobber the remote rack
// face during export.
// Inputs: a device with Face "" and a remote with no Face. Outputs: an empty
// diff slice.
// Data choice: the empty string is the sentinel for "unset", exercising the
// early-return guard.
func TestCompareFaceNoLocalFace(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Name: "no-face",
		Face: "", // empty means "no face"
	}
	remote := &nautobotapi.Device{}
	diffs := compareFace(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when local face is empty, got %d", len(diffs))
	}
}

// TestCompareFaceMismatch verifies compareFace emits one "face" diff (LocalVal
// "rear") when local and remote rack faces differ.
//
// Why it matters: face drift (front vs rear) must be surfaced so the exported
// device orientation matches the cani inventory.
// Inputs: local Face "rear" vs. a remote DeviceFace value "front". Outputs: a
// single diff with Field "face" and LocalVal "rear".
// Data choice: opposite faces (rear vs front) make the mismatch and the
// reported LocalVal explicit.
func TestCompareFaceMismatch(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Name: "face-mismatch",
		Face: "rear",
	}
	remote := &nautobotapi.Device{}
	setNBValue(&remote.Face, "front")
	diffs := compareFace(dev, remote)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Field != "face" {
		t.Errorf("field = %q, want 'face'", diffs[0].Field)
	}
	if diffs[0].LocalVal != "rear" {
		t.Errorf("LocalVal = %q, want 'rear'", diffs[0].LocalVal)
	}
}

// TestCompareFaceMatch verifies compareFace returns no diffs when local and
// remote rack faces are equal.
//
// Why it matters: matching faces must not register as drift, avoiding noise in
// export conflict reporting.
// Inputs: local Face "front" and a remote DeviceFace value "front". Outputs: an
// empty diff slice.
// Data choice: identical non-empty faces exercise the equal branch past the
// empty-face guard.
func TestCompareFaceMatch(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Name: "face-match",
		Face: "front",
	}
	remote := &nautobotapi.Device{}
	setNBValue(&remote.Face, "front")
	diffs := compareFace(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when faces match, got %d", len(diffs))
	}
}

// ---------- FieldDiff struct ----------

// TestFieldDiffStruct verifies a FieldDiff literal exposes the Field, LocalVal,
// and RemoteVal values assigned to it.
//
// Why it matters: FieldDiff is the unit carried through every device comparison
// and printed in the sync summary, so its fields must hold what callers set.
// Inputs: a FieldDiff literal for a device_type drift (DL380 vs DL360).
// Outputs: the struct's own fields (no package logic is invoked).
// Data choice: a realistic device-type mismatch keeps the example meaningful.
// NOTE: this only checks Go struct assignment, not any export behavior.
func TestFieldDiffStruct(t *testing.T) {
	diff := FieldDiff{
		Field:     "device_type",
		LocalVal:  "DL380",
		RemoteVal: "DL360",
	}
	if diff.Field != "device_type" {
		t.Errorf("Field = %q", diff.Field)
	}
	if diff.LocalVal != "DL380" {
		t.Errorf("LocalVal = %q", diff.LocalVal)
	}
}

// ---------- comparePosition matching (no diff) ----------

// TestComparePositionMatching verifies comparePosition reports no diff when the
// local and remote rack positions are equal.
//
// Why it matters: identical positions must not be flagged as drift, or every
// sync would needlessly re-write unchanged devices.
// Inputs: a device at RackPosition 12 and a remote Device with Position=12.
// Outputs: an empty diff slice.
// Data choice: equal non-zero positions isolate the matching branch.
func TestComparePositionMatching(t *testing.T) {
	pos := 12
	dev := &devicetypes.CaniDeviceType{RackPosition: 12}
	remote := &nautobotapi.Device{Position: &pos}
	diffs := comparePosition(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when positions match, got %+v", diffs)
	}
}

// TestComparePositionLocalZero verifies comparePosition reports no diff when the
// local position is zero (unset), even if the remote has a position.
//
// Why it matters: a device with no local position must not clobber the position
// already recorded in Nautobot.
// Inputs: a device with RackPosition=0 and a remote Device with Position=5.
// Outputs: an empty diff slice (RackPosition<=0 short-circuits).
// Data choice: local 0 vs remote 5 proves the unset-local guard wins over a
// differing remote.
func TestComparePositionLocalZero(t *testing.T) {
	pos := 5
	dev := &devicetypes.CaniDeviceType{RackPosition: 0}
	remote := &nautobotapi.Device{Position: &pos}
	diffs := comparePosition(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when local position is zero, got %+v", diffs)
	}
}

// ---------- compareFace matching (no diff) ----------

// TestCompareFaceMatching verifies compareFace reports no diff when the local and
// remote rack faces are equal.
//
// Why it matters: a matching face must not register as drift during the device
// comparison that decides whether to update Nautobot.
// Inputs: a device with Face="front" and a remote whose Face value is "front".
// Outputs: an empty diff slice.
// Data choice: equal "front" faces isolate the no-diff branch.
func TestCompareFaceMatching(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Face: "front"}
	remote := &nautobotapi.Device{}
	setNBValue(&remote.Face, "front")
	diffs := compareFace(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when faces match, got %+v", diffs)
	}
}

// TestCompareFaceLocalEmpty verifies compareFace reports no diff when the local
// face is empty, regardless of the remote value.
//
// Why it matters: an unspecified local face must not overwrite the face already
// stored in Nautobot.
// Inputs: a device with Face="" and a remote whose Face value is "rear".
// Outputs: an empty diff slice (empty local face short-circuits).
// Data choice: empty local vs "rear" remote proves the guard ignores a populated
// remote.
func TestCompareFaceLocalEmpty(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Face: ""}
	remote := &nautobotapi.Device{}
	setNBValue(&remote.Face, "rear")
	diffs := compareFace(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when local face is empty, got %+v", diffs)
	}
}

// ---------- compareRack matching UUIDs ----------

// TestCompareRackBothMatchingSameRack verifies compareRack returns nil when
// neither the local device nor the remote has a rack.
//
// Why it matters: two "no rack" sides must agree, avoiding spurious rack drift
// for unracked devices during sync.
// Inputs: a device with Rack=uuid.Nil, a remote with Rack=nil, and a mapper with
// an empty inventory. Outputs: nil diffs.
// Data choice: empty rack on both sides isolates the both-Nil match branch.
func TestCompareRackBothMatchingSameRack(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{},
	}
	mapper.SetInventory(inv)

	// Both have no rack → match (nil)
	dev := &devicetypes.CaniDeviceType{Rack: uuid.Nil}
	remote := &nautobotapi.Device{Rack: nil}
	diffs := compareRack(dev, remote, mapper)
	if diffs != nil {
		t.Errorf("expected nil diffs, got %+v", diffs)
	}
}

// TestCompareRackRemoteHasRackLocalNil verifies compareRack emits one "rack" diff
// when the remote has a rack but the local device does not.
//
// Why it matters: a rack present in Nautobot but absent locally is real drift the
// operator must see before merging.
// Inputs: a device with Rack=uuid.Nil and a remote whose Rack ID union holds a
// fresh UUID. Outputs: a single FieldDiff with Field=="rack".
// Data choice: a populated remote union against a Nil local forces exactly the
// one-sided mismatch.
func TestCompareRackRemoteHasRackLocalNil(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{},
	}
	mapper.SetInventory(inv)

	rackUUID := uuid.New()

	dev := &devicetypes.CaniDeviceType{Rack: uuid.Nil}
	remote := &nautobotapi.Device{}
	mustSetRefID(t, &remote.Rack, rackUUID)
	diffs := compareRack(dev, remote, mapper)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Field != "rack" {
		t.Errorf("expected field 'rack', got %q", diffs[0].Field)
	}
}
