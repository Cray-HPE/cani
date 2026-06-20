package export

import (
	"testing"

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// --- remoteSlotKey ---

// TestRemoteSlotKey_NilDevice verifies remoteSlotKey returns nil when handed a
// nil *Device instead of dereferencing it.
//
// Why it matters: swap reconciliation calls remoteSlotKey on whatever
// fetchFullDeviceByID returns, and a device that could not be fetched must not
// panic the export or fabricate a phantom rack slot.
// Inputs: a nil *nautobotapi.Device. Outputs: a nil *slotKey.
// Data choice: nil is the boundary input that exercises the first guard clause
// before any field is accessed.
func TestRemoteSlotKey_NilDevice(t *testing.T) {
	if got := remoteSlotKey(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

// TestRemoteSlotKey_NilPosition verifies remoteSlotKey returns nil when the
// device reports no rack U-position.
//
// Why it matters: a device with no position is unplaced in Nautobot, so it
// occupies no slot and must never be counted as a move or as blocking a swap.
// Inputs: a Device with Position nil (Rack also unset). Outputs: a nil *slotKey.
// Data choice: an otherwise-empty Device isolates the Position guard from the
// later rack checks.
func TestRemoteSlotKey_NilPosition(t *testing.T) {
	d := &nautobotapi.Device{}
	if got := remoteSlotKey(d); got != nil {
		t.Errorf("expected nil when Position is nil, got %v", got)
	}
}

// TestRemoteSlotKey_NilRack verifies remoteSlotKey returns nil when a position
// is set but the device is assigned to no rack.
//
// Why it matters: a slot key is meaningless without a rack UUID; treating a
// rackless device as occupying a slot would corrupt swap-cycle detection.
// Inputs: a Device with Position 1 but Rack nil. Outputs: a nil *slotKey.
// Data choice: position-without-rack is the exact inconsistent state the Rack
// guard exists to reject.
func TestRemoteSlotKey_NilRack(t *testing.T) {
	pos := 1
	d := &nautobotapi.Device{Position: &pos}
	if got := remoteSlotKey(d); got != nil {
		t.Errorf("expected nil when Rack is nil, got %v", got)
	}
}

// TestRemoteSlotKey_NilRackID verifies remoteSlotKey returns nil when the rack
// reference is present but carries no ID.
//
// Why it matters: the slot key is keyed by rack UUID, so an empty rack reference
// from Nautobot must not yield a zero-UUID slot that could collide with other
// devices during swap detection.
// Inputs: a Device with Position set and a Rack ref whose Id is nil. Outputs: a
// nil *slotKey.
// Data choice: an empty BulkWritableCircuitRequestTenant reproduces a rack ref
// lacking the Id union, hitting the final nil guard.
func TestRemoteSlotKey_NilRackID(t *testing.T) {
	pos := 1
	d := &nautobotapi.Device{
		Position: &pos,
		Rack:     &nautobotapi.BulkWritableCircuitRequestTenant{},
	}
	if got := remoteSlotKey(d); got != nil {
		t.Errorf("expected nil when Rack.Id is nil, got %v", got)
	}
}

func makeRackRef(id uuid.UUID) *nautobotapi.BulkWritableCircuitRequestTenant {
	var union nautobotapi.BulkWritableCableRequestStatusId
	_ = union.FromBulkWritableCableRequestStatusId0(openapi_types.UUID(id))
	return &nautobotapi.BulkWritableCircuitRequestTenant{Id: &union}
}

// TestRemoteSlotKey_FrontDefault verifies remoteSlotKey builds a slotKey from a
// fully-placed device and defaults Face to "front" when none is reported.
//
// Why it matters: Nautobot's unique slot constraint is (rack, position, face),
// so the exporter must derive a stable, defaulted face to compare remote and
// desired placement during a merge.
// Inputs: a Device with a rack UUID, Position 5, and no Face. Outputs: a slotKey
// carrying that RackID, Position 5, and Face "front".
// Data choice: omitting Face exercises the default branch; position 5 is an
// arbitrary occupied U-slot.
func TestRemoteSlotKey_FrontDefault(t *testing.T) {
	rackID := uuid.New()
	pos := 5
	d := &nautobotapi.Device{
		Position: &pos,
		Rack:     makeRackRef(rackID),
	}
	got := remoteSlotKey(d)
	if got == nil {
		t.Fatal("expected non-nil slotKey")
	}
	if got.RackID != rackID {
		t.Errorf("RackID: expected %s, got %s", rackID, got.RackID)
	}
	if got.Position != 5 {
		t.Errorf("Position: expected 5, got %d", got.Position)
	}
	if got.Face != "front" {
		t.Errorf("Face: expected front, got %s", got.Face)
	}
}

// TestRemoteSlotKey_RearFace verifies remoteSlotKey preserves an explicit
// "rear" face reported by the Nautobot device.
//
// Why it matters: front and rear are distinct slots in Nautobot, so a
// rear-mounted device must map to Face "rear" or the merge would compare it
// against the wrong slot and mis-detect a move.
// Inputs: a placed Device whose Face value is "rear". Outputs: a slotKey with
// Face "rear".
// Data choice: "rear" is the only non-default face value the function honors, so
// it is the case worth asserting.
func TestRemoteSlotKey_RearFace(t *testing.T) {
	rackID := uuid.New()
	pos := 10
	rear := nautobotapi.DeviceFaceValue("rear")
	d := &nautobotapi.Device{
		Position: &pos,
		Rack:     makeRackRef(rackID),
		Face:     &nautobotapi.DeviceFace{Value: &rear},
	}
	got := remoteSlotKey(d)
	if got == nil {
		t.Fatal("expected non-nil slotKey")
	}
	if got.Face != "rear" {
		t.Errorf("Face: expected rear, got %s", got.Face)
	}
}

// TestRemoteSlotKey_NonRearFaceDefaultsToFront verifies remoteSlotKey maps any
// face other than "rear" (here an explicit "front") to "front".
//
// Why it matters: the function only special-cases "rear", so confirming the
// front/default collapse prevents drift if Nautobot ever returns an unexpected
// face string.
// Inputs: a placed Device whose Face value is "front". Outputs: a slotKey with
// Face "front".
// Data choice: passing "front" explicitly distinguishes this from the
// face-absent default test and pins the non-rear branch.
func TestRemoteSlotKey_NonRearFaceDefaultsToFront(t *testing.T) {
	rackID := uuid.New()
	pos := 3
	front := nautobotapi.DeviceFaceValue("front")
	d := &nautobotapi.Device{
		Position: &pos,
		Rack:     makeRackRef(rackID),
		Face:     &nautobotapi.DeviceFace{Value: &front},
	}
	got := remoteSlotKey(d)
	if got == nil {
		t.Fatal("expected non-nil slotKey")
	}
	if got.Face != "front" {
		t.Errorf("Face: expected front, got %s", got.Face)
	}
}

// --- derefSlotKey ---

// TestDerefSlotKey_Nil verifies derefSlotKey returns the zero slotKey when its
// pointer argument is nil.
//
// Why it matters: pendingMove.From is filled from derefSlotKey, so a device with
// no known current slot must record an empty origin rather than panic.
// Inputs: a nil *slotKey. Outputs: the zero-value slotKey.
// Data choice: nil is the boundary condition the guard clause protects against.
func TestDerefSlotKey_Nil(t *testing.T) {
	got := derefSlotKey(nil)
	zero := slotKey{}
	if got != zero {
		t.Errorf("expected zero slotKey, got %+v", got)
	}
}

// TestDerefSlotKey_NonNil verifies derefSlotKey returns a copy of the
// pointed-to slotKey when the pointer is non-nil.
//
// Why it matters: recording an accurate "from" slot lets swap reconciliation
// reason about which slots are being vacated during a merge export.
// Inputs: a *slotKey with a RackID, Position 7, and Face "rear". Outputs: the
// same slotKey returned by value.
// Data choice: a non-front face and non-zero position ensure every field is
// carried through, not just defaulted ones.
func TestDerefSlotKey_NonNil(t *testing.T) {
	id := uuid.New()
	sk := &slotKey{RackID: id, Position: 7, Face: "rear"}
	got := derefSlotKey(sk)
	if got != *sk {
		t.Errorf("expected %+v, got %+v", *sk, got)
	}
}
