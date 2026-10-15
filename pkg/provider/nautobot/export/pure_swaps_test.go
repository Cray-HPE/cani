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

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// ---------- remoteSlotKey ----------

// TestRemoteSlotKeyNil verifies remoteSlotKey returns nil when the device is
// nil, has no position, or has a position but no rack.
//
// Why it matters: position-swap detection compares slot keys; a non-nil key for
// an unplaced remote device would invent a phantom slot and trigger spurious
// moves during export.
// Inputs: nil, a device with no position, a device with position but no rack.
// Outputs: nil in all three cases.
// Data choice: each case removes one required field to prove every guard clause
// independently returns nil.
func TestRemoteSlotKeyNil(t *testing.T) {
	// nil device → nil
	if got := remoteSlotKey(nil); got != nil {
		t.Error("expected nil for nil device")
	}

	// missing position → nil
	d := &nautobotapi.Device{}
	if got := remoteSlotKey(d); got != nil {
		t.Error("expected nil for device without position")
	}

	// missing rack → nil
	pos := 5
	d = &nautobotapi.Device{Position: &pos}
	if got := remoteSlotKey(d); got != nil {
		t.Error("expected nil for device without rack")
	}
}

// TestRemoteSlotKeyValid verifies a fully placed remote device yields a slotKey
// with the right rack, position, and face, defaulting face to "front" and
// honoring an explicit "rear".
//
// Why it matters: accurate slot keys drive Nautobot rack-position reconciliation;
// a wrong face or position would move devices incorrectly.
// Inputs: a device at position 5 in a fixed rack, then with face set to rear.
// Outputs: slotKey{rack, 5, "front"} then face "rear".
// Data choice: omitting face first proves the front default, then setting rear
// proves the explicit path; a fixed rack UUID makes the assertion exact.
func TestRemoteSlotKeyValid(t *testing.T) {
	pos := 5
	rackID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	d := &nautobotapi.Device{
		Position: &pos,
	}
	mustSetRefID(t, &d.Rack, rackID)
	sk := remoteSlotKey(d)
	if sk == nil {
		t.Fatal("expected non-nil slotKey")
	}
	if sk.RackID != rackID {
		t.Errorf("rackID = %s, want %s", sk.RackID, rackID)
	}
	if sk.Position != 5 {
		t.Errorf("position = %d, want 5", sk.Position)
	}
	if sk.Face != "front" {
		t.Errorf("face = %q, want 'front'", sk.Face)
	}

	// rear face
	setNBValue(&d.Face, "rear")
	sk = remoteSlotKey(d)
	if sk == nil {
		t.Fatal("expected non-nil slotKey for rear")
	}
	if sk.Face != "rear" {
		t.Errorf("face = %q, want 'rear'", sk.Face)
	}
}

// ---------- derefSlotKey ----------

// TestDerefSlotKeyNil verifies dereferencing a nil *slotKey yields the zero
// slotKey rather than panicking.
//
// Why it matters: swap logic dereferences optional slot keys when comparing
// current vs desired placement; a nil deref there would crash the export.
// Inputs: a nil *slotKey.
// Outputs: slotKey{Position:0, Face:"", RackID:Nil}.
// Data choice: asserting all three zero fields confirms a true zero value, not a
// partially initialized struct.
func TestDerefSlotKeyNil(t *testing.T) {
	got := derefSlotKey(nil)
	if got.Position != 0 || got.Face != "" || got.RackID != uuid.Nil {
		t.Error("expected zero slotKey for nil input")
	}
}

// TestDerefSlotKeyNonNil verifies dereferencing a non-nil *slotKey copies its
// fields exactly.
//
// Why it matters: the value copy is compared against remote slot keys during
// reconciliation; a dropped field would cause incorrect move decisions.
// Inputs: a *slotKey with a fixed rack, position 42, face "rear".
// Outputs: an equal slotKey value.
// Data choice: distinctive non-zero values (42, "rear") ensure the copy is real
// and not coincidentally matching a zero value.
func TestDerefSlotKeyNonNil(t *testing.T) {
	sk := &slotKey{
		RackID:   uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		Position: 42,
		Face:     "rear",
	}
	got := derefSlotKey(sk)
	if got.RackID != sk.RackID || got.Position != 42 || got.Face != "rear" {
		t.Errorf("derefSlotKey did not dereference correctly: %+v", got)
	}
}

// ---------- pendingMove struct ----------

// TestPendingMoveFields verifies a pendingMove literal preserves its nested
// slotKey From/To fields (position and face).
//
// Why it matters: pendingMove drives rack relocation during export; mis-wired
// From/To fields would move a device to the wrong slot in Nautobot.
// Inputs: a pendingMove with From{Position 1} and To{Face "rear"}. Outputs:
// those nested fields read back unchanged.
// Data choice: differing From/To racks, positions, and faces ensure each nested
// field is checked independently.
func TestPendingMoveFields(t *testing.T) {
	devID := uuid.New()
	rackA := uuid.New()
	rackB := uuid.New()
	move := pendingMove{
		DeviceID:   devID,
		DeviceName: "compute-001",
		From:       slotKey{RackID: rackA, Position: 1, Face: "front"},
		To:         slotKey{RackID: rackB, Position: 5, Face: "rear"},
	}
	if move.DeviceName != "compute-001" {
		t.Errorf("DeviceName = %q", move.DeviceName)
	}
	if move.From.Position != 1 {
		t.Errorf("From.Position = %d, want 1", move.From.Position)
	}
	if move.To.Face != "rear" {
		t.Errorf("To.Face = %q, want 'rear'", move.To.Face)
	}
}

// ---------- remoteSlotKey additional cases ----------

// TestRemoteSlotKeyValidFront verifies remoteSlotKey extracts RackID, Position,
// and a defaulted "front" Face from a Nautobot Device whose Face field is nil.
//
// Why it matters: slot keys identify where a device currently sits in Nautobot
// so the exporter can detect when cani inventory implies a rack-slot move; a nil
// Face must default to "front" to match Nautobot's own default.
// Inputs: a *Device with Position=10, a Rack Id union, and Face=nil. Outputs: a
// *slotKey with RackID set, Position=10, Face="front".
// Data choice: Face is left nil specifically to exercise the front-default path.
func TestRemoteSlotKeyValidFront(t *testing.T) {
	pos := 10
	rackID := uuid.New()
	d := &nautobotapi.Device{
		Position: &pos,
		Face:     nil, // nil face → defaults to "front"
	}
	mustSetRefID(t, &d.Rack, rackID)

	sk := remoteSlotKey(d)
	if sk == nil {
		t.Fatal("expected non-nil slotKey")
	}
	if sk.RackID != rackID {
		t.Errorf("RackID = %s, want %s", sk.RackID, rackID)
	}
	if sk.Position != 10 {
		t.Errorf("Position = %d, want 10", sk.Position)
	}
	if sk.Face != "front" {
		t.Errorf("Face = %q, want 'front'", sk.Face)
	}
}

// TestRemoteSlotKeyValidRear verifies remoteSlotKey reports Face="rear" when the
// Device's Face value is explicitly "rear".
//
// Why it matters: front vs. rear placement is part of a slot's identity in
// Nautobot, so a misread face would make the exporter compute spurious moves.
// Inputs: a *Device with Position=5, a Rack Id union, and Face.Value="rear".
// Outputs: a *slotKey whose Face is "rear".
// Data choice: Face.Value="rear" is the only non-default branch, so it is the
// single field this test asserts.
func TestRemoteSlotKeyValidRear(t *testing.T) {
	pos := 5
	rackID := uuid.New()
	d := &nautobotapi.Device{
		Position: &pos,
	}
	mustSetRefID(t, &d.Rack, rackID)
	setNBValue(&d.Face, "rear")

	sk := remoteSlotKey(d)
	if sk == nil {
		t.Fatal("expected non-nil slotKey")
	}
	if sk.Face != "rear" {
		t.Errorf("Face = %q, want 'rear'", sk.Face)
	}
}

// TestRemoteSlotKeyNilPosition verifies remoteSlotKey returns nil when the Device
// has no rack Position.
//
// Why it matters: an unracked device has no slot to compare against, so the
// exporter must skip it rather than fabricate a zero-position slot key.
// Inputs: a *Device with Position=nil but a valid Rack. Outputs: nil.
// Data choice: only Position is nil (Rack stays valid) to isolate the
// missing-position guard from the missing-rack guard.
func TestRemoteSlotKeyNilPosition(t *testing.T) {
	rackID := uuid.New()
	d := &nautobotapi.Device{
		Position: nil, // no position
	}
	mustSetRefID(t, &d.Rack, rackID)
	if remoteSlotKey(d) != nil {
		t.Error("expected nil when Position is nil")
	}
}

// TestRemoteSlotKeyNilRack verifies remoteSlotKey returns nil when the Device has
// a Position but no Rack.
//
// Why it matters: a position without a rack is meaningless for move detection,
// so the exporter must treat it as "no slot" instead of dereferencing nil.
// Inputs: a *Device with Position=3 and Rack=nil. Outputs: nil.
// Data choice: Rack is nil while Position is set to isolate the missing-rack
// guard from the missing-position guard.
func TestRemoteSlotKeyNilRack(t *testing.T) {
	pos := 3
	d := &nautobotapi.Device{
		Position: &pos,
		Rack:     nil,
	}
	if remoteSlotKey(d) != nil {
		t.Error("expected nil when Rack is nil")
	}
}

// ---------- derefString ----------

// TestDerefStringNil verifies derefString returns "" for a nil *string.
//
// Why it matters: many Nautobot API fields are optional pointers, so the
// exporter relies on derefString to read them without nil-pointer panics.
// Inputs: a nil *string. Outputs: "".
// Data choice: nil is the only input that triggers the guard clause.
func TestDerefStringNil(t *testing.T) {
	if got := derefString(nil); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

// TestDerefStringNonNil verifies derefString returns the pointed-to value for a
// non-nil *string.
//
// Why it matters: when an optional Nautobot field is present, the exporter must
// read its actual value, not the empty fallback.
// Inputs: a *string pointing at "hello". Outputs: "hello".
// Data choice: a simple non-empty literal confirms the value is dereferenced
// rather than dropped.
func TestDerefStringNonNil(t *testing.T) {
	s := "hello"
	if got := derefString(&s); got != "hello" {
		t.Errorf("got %q, want 'hello'", got)
	}
}
