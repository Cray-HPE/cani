package devicetypes

// Test coverage table for CaniRackType methods:
//
// | Function             | Happy-path test                        | Failure test                            |
// |----------------------|----------------------------------------|-----------------------------------------|
// | Validate             | TestValidateRackHappyPath              | TestValidateRackNil                     |
// | GetID                | TestGetIDRackHappyPath                 | TestGetIDRackNil                        |
// | GetSlug              | TestGetSlugRackHappyPath               | TestGetSlugRackNil                      |
// | GetVendor            | TestGetVendorRackHappyPath             | TestGetVendorRackNil                    |
// | GetType              | TestGetTypeRackHappyPath               | TestGetTypeRackNil                      |
// | GetStatus            | TestGetStatusRackHappyPath             | TestGetStatusRackNil                    |
// | isSlotBlocked        | TestIsSlotBlockedHappyPath             | TestIsSlotBlockedFullDepthConflict      |
// | hasFaceOccupant      | TestHasFaceOccupantHappyPath           | TestHasFaceOccupantOccupied             |
// | CanFitDevice         | TestCanFitDeviceHappyPath              | TestCanFitDeviceOutOfBounds             |
// | PlaceDevice          | TestPlaceDeviceHappyPath               | TestPlaceDeviceConflict                 |
// | RemoveDevice         | TestRemoveDeviceHappyPath              | TestRemoveDeviceNil                     |
// | GetDeviceStartU      | TestGetDeviceStartURackHappyPath       | TestGetDeviceStartURackNil              |
// | addDevice            | TestAddDeviceHappyPath                 | TestAddDeviceDuplicate                  |
// | FindNextAvailableSlot| TestFindNextAvailableSlotHappyPath     | TestFindNextAvailableSlotFull           |
// | MigrateLegacySlots   | TestMigrateLegacySlotsHappyPath        | TestMigrateLegacySlotsEmpty             |

import (
	"testing"

	"github.com/google/uuid"
)

// --- Validate ---

func TestValidateRackHappyPath(t *testing.T) {
	r := &CaniRackType{UHeight: 42}
	if err := r.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateRackNil(t *testing.T) {
	var r *CaniRackType
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for nil receiver")
	}
}

// --- GetID ---

func TestGetIDRackHappyPath(t *testing.T) {
	id := uuid.New()
	r := &CaniRackType{ID: id}
	if got := r.GetID(); got != id {
		t.Fatalf("expected %v, got %v", id, got)
	}
}

func TestGetIDRackNil(t *testing.T) {
	var r *CaniRackType
	if got := r.GetID(); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %v", got)
	}
}

// --- GetSlug ---

func TestGetSlugRackHappyPath(t *testing.T) {
	r := &CaniRackType{Slug: "std-42u"}
	if got := r.GetSlug(); got != "std-42u" {
		t.Fatalf("expected std-42u, got %s", got)
	}
}

func TestGetSlugRackNil(t *testing.T) {
	var r *CaniRackType
	if got := r.GetSlug(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetVendor ---

func TestGetVendorRackHappyPath(t *testing.T) {
	r := &CaniRackType{Manufacturer: "Acme"}
	if got := r.GetVendor(); got != "Acme" {
		t.Fatalf("expected Acme, got %s", got)
	}
}

func TestGetVendorRackNil(t *testing.T) {
	var r *CaniRackType
	if got := r.GetVendor(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetType ---

func TestGetTypeRackHappyPath(t *testing.T) {
	r := &CaniRackType{HardwareType: "custom-rack"}
	if got := r.GetType(); got != Type("custom-rack") {
		t.Fatalf("expected custom-rack, got %s", got)
	}
}

func TestGetTypeRackNil(t *testing.T) {
	var r *CaniRackType
	if got := r.GetType(); got != "" {
		t.Fatalf("expected empty Type, got %s", got)
	}
}

// --- GetStatus ---

func TestGetStatusRackHappyPath(t *testing.T) {
	r := &CaniRackType{Status: "active"}
	if got := r.GetStatus(); got != "active" {
		t.Fatalf("expected active, got %s", got)
	}
}

func TestGetStatusRackNil(t *testing.T) {
	var r *CaniRackType
	if got := r.GetStatus(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- isSlotBlocked ---

func TestIsSlotBlockedHappyPath(t *testing.T) {
	r := &CaniRackType{
		UHeight:       42,
		OccupiedSlots: map[int]map[string]uuid.UUID{},
	}
	if r.isSlotBlocked(1, FaceFront, false) {
		t.Fatal("expected empty slot to not be blocked")
	}
}

func TestIsSlotBlockedFullDepthConflict(t *testing.T) {
	devID := uuid.New()
	r := &CaniRackType{
		UHeight: 42,
		OccupiedSlots: map[int]map[string]uuid.UUID{
			5: {FaceFull: devID},
		},
	}
	if !r.isSlotBlocked(5, FaceFront, false) {
		t.Fatal("expected full-depth occupant to block front face")
	}
}

// --- hasFaceOccupant ---

func TestHasFaceOccupantHappyPath(t *testing.T) {
	r := &CaniRackType{}
	slot := map[string]uuid.UUID{}
	if r.hasFaceOccupant(slot, FaceFront) {
		t.Fatal("expected empty slot to have no occupant")
	}
}

func TestHasFaceOccupantOccupied(t *testing.T) {
	r := &CaniRackType{}
	slot := map[string]uuid.UUID{FaceRear: uuid.New()}
	if !r.hasFaceOccupant(slot, FaceRear) {
		t.Fatal("expected rear face to be occupied")
	}
}

// --- CanFitDevice ---

func TestCanFitDeviceHappyPath(t *testing.T) {
	r := &CaniRackType{
		UHeight:       42,
		OccupiedSlots: map[int]map[string]uuid.UUID{},
	}
	if !r.CanFitDevice(1, 2, FaceFront, false) {
		t.Fatal("expected 2U device to fit at U1 in empty rack")
	}
}

func TestCanFitDeviceOutOfBounds(t *testing.T) {
	r := &CaniRackType{
		UHeight:       10,
		OccupiedSlots: map[int]map[string]uuid.UUID{},
	}
	if r.CanFitDevice(9, 4, FaceFront, false) {
		t.Fatal("expected device exceeding rack height to not fit")
	}
}

// --- PlaceDevice ---

func TestPlaceDeviceHappyPath(t *testing.T) {
	devID := uuid.New()
	r := &CaniRackType{UHeight: 42}
	if !r.PlaceDevice(devID, 1, 2, FaceFront, false) {
		t.Fatal("expected placement to succeed in empty rack")
	}
	if len(r.Devices) != 1 || r.Devices[0] != devID {
		t.Fatal("expected device to appear in Devices list")
	}
}

func TestPlaceDeviceConflict(t *testing.T) {
	devA := uuid.New()
	devB := uuid.New()
	r := &CaniRackType{UHeight: 42}
	r.PlaceDevice(devA, 1, 2, FaceFront, false)
	if r.PlaceDevice(devB, 1, 2, FaceFront, false) {
		t.Fatal("expected placement to fail on occupied slot")
	}
}

// --- RemoveDevice ---

func TestRemoveDeviceHappyPath(t *testing.T) {
	devID := uuid.New()
	r := &CaniRackType{UHeight: 42}
	r.PlaceDevice(devID, 5, 3, FaceFront, false)
	r.RemoveDevice(devID)
	if len(r.Devices) != 0 {
		t.Fatal("expected Devices list to be empty after removal")
	}
	if len(r.OccupiedSlots) != 0 {
		t.Fatal("expected OccupiedSlots to be empty after removal")
	}
}

func TestRemoveDeviceNil(t *testing.T) {
	var r *CaniRackType
	// should not panic on nil receiver
	r.RemoveDevice(uuid.New())
}

// --- GetDeviceStartU ---

func TestGetDeviceStartURackHappyPath(t *testing.T) {
	devID := uuid.New()
	r := &CaniRackType{UHeight: 42}
	r.PlaceDevice(devID, 10, 4, FaceFront, false)
	if got := r.GetDeviceStartU(devID); got != 10 {
		t.Fatalf("expected startU=10, got %d", got)
	}
}

func TestGetDeviceStartURackNil(t *testing.T) {
	var r *CaniRackType
	if got := r.GetDeviceStartU(uuid.New()); got != 0 {
		t.Fatalf("expected 0 for nil receiver, got %d", got)
	}
}

// --- addDevice ---

func TestAddDeviceHappyPath(t *testing.T) {
	devID := uuid.New()
	r := &CaniRackType{}
	r.addDevice(devID)
	if len(r.Devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(r.Devices))
	}
}

func TestAddDeviceDuplicate(t *testing.T) {
	devID := uuid.New()
	r := &CaniRackType{}
	r.addDevice(devID)
	r.addDevice(devID)
	if len(r.Devices) != 1 {
		t.Fatalf("expected duplicate to be ignored, got %d devices", len(r.Devices))
	}
}

// --- FindNextAvailableSlot ---

func TestFindNextAvailableSlotHappyPath(t *testing.T) {
	r := &CaniRackType{UHeight: 42}
	devID := uuid.New()
	// Place a 2U device at U41-U42 (top of rack).
	r.PlaceDevice(devID, 41, 2, FaceFront, false)
	// Top-down: next available 2U slot should be U39 (occupies U39-U40).
	got := r.FindNextAvailableSlot(2, FaceFront, false)
	if got != 39 {
		t.Fatalf("expected next slot at U39, got %d", got)
	}
}

func TestFindNextAvailableSlotFull(t *testing.T) {
	r := &CaniRackType{UHeight: 4}
	r.PlaceDevice(uuid.New(), 1, 4, FaceFront, true)
	got := r.FindNextAvailableSlot(1, FaceFront, false)
	if got != 0 {
		t.Fatalf("expected 0 when rack is full, got %d", got)
	}
}

// --- MigrateLegacySlots ---

func TestMigrateLegacySlotsHappyPath(t *testing.T) {
	devID := uuid.New()
	r := &CaniRackType{UHeight: 42}
	legacy := map[int]uuid.UUID{1: devID, 2: devID}
	r.MigrateLegacySlots(legacy)
	if r.OccupiedSlots[1][FaceFront] != devID {
		t.Fatal("expected legacy slot 1 migrated to front face")
	}
	if r.OccupiedSlots[2][FaceFront] != devID {
		t.Fatal("expected legacy slot 2 migrated to front face")
	}
}

func TestMigrateLegacySlotsEmpty(t *testing.T) {
	r := &CaniRackType{UHeight: 42}
	r.MigrateLegacySlots(map[int]uuid.UUID{})
	if r.OccupiedSlots == nil {
		t.Fatal("expected OccupiedSlots to be initialized even with empty legacy map")
	}
	if len(r.OccupiedSlots) != 0 {
		t.Fatalf("expected 0 occupied slots, got %d", len(r.OccupiedSlots))
	}
}
