package export

import (
	"testing"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// --- remoteSlotKey ---

func TestRemoteSlotKey_NilDevice(t *testing.T) {
	if got := remoteSlotKey(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestRemoteSlotKey_NilPosition(t *testing.T) {
	d := &nautobotapi.Device{}
	if got := remoteSlotKey(d); got != nil {
		t.Errorf("expected nil when Position is nil, got %v", got)
	}
}

func TestRemoteSlotKey_NilRack(t *testing.T) {
	pos := 1
	d := &nautobotapi.Device{Position: &pos}
	if got := remoteSlotKey(d); got != nil {
		t.Errorf("expected nil when Rack is nil, got %v", got)
	}
}

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

func TestDerefSlotKey_Nil(t *testing.T) {
	got := derefSlotKey(nil)
	zero := slotKey{}
	if got != zero {
		t.Errorf("expected zero slotKey, got %+v", got)
	}
}

func TestDerefSlotKey_NonNil(t *testing.T) {
	id := uuid.New()
	sk := &slotKey{RackID: id, Position: 7, Face: "rear"}
	got := derefSlotKey(sk)
	if got != *sk {
		t.Errorf("expected %+v, got %+v", *sk, got)
	}
}
