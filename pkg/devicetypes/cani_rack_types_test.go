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
// | GetSlotOccupant      | TestGetSlotOccupantHappyPath           | TestGetSlotOccupantEmpty                |
// | GetDeviceFace        | TestGetDeviceFaceHappyPath             | TestGetDeviceFaceNil                    |
// | GetDeviceHeight      | TestGetDeviceHeightHappyPath           | TestGetDeviceHeightNil                  |
// | SwapDevices          | TestSwapDevicesHappyPath               | TestSwapDevicesNil                      |

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// --- Validate ---

// TestValidateRackHappyPath verifies Validate returns nil for a rack with a valid
// UHeight.
//
// Why it matters: Validate gates a rack before it enters inventory, so a rack
// with usable U-slots must pass.
// Inputs: a rack with UHeight 42. Outputs: an error, nil expected.
// Data choice: 42U is the canonical full-height rack, a representative valid
// UHeight clearing the minimum-of-1 check.
func TestValidateRackHappyPath(t *testing.T) {
	r := &CaniRackType{UHeight: 42}
	if err := r.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestValidateRackNil verifies Validate returns an error for a nil receiver.
//
// Why it matters: a nil rack is a programming error, and Validate must fail
// loudly rather than dereference nil and panic.
// Inputs: a nil *CaniRackType. Outputs: an error, non-nil expected.
// Data choice: a nil receiver is the only input that reaches the nil guard ahead
// of the UHeight check.
func TestValidateRackNil(t *testing.T) {
	var r *CaniRackType
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for nil receiver")
	}
}

// --- GetID ---

// TestGetIDRackHappyPath verifies GetID returns the rack's stored UUID.
//
// Why it matters: the ID uniquely identifies a rack across inventory and export,
// so the accessor must return it unchanged.
// Inputs: a rack with a generated ID. Outputs: the UUID, expected equal to the
// stored value.
// Data choice: a freshly generated uuid is an arbitrary distinct value proving
// the getter returns the stored field rather than a constant.
func TestGetIDRackHappyPath(t *testing.T) {
	id := uuid.New()
	r := &CaniRackType{ID: id}
	if got := r.GetID(); got != id {
		t.Fatalf("expected %v, got %v", id, got)
	}
}

// TestGetIDRackNil verifies GetID returns uuid.Nil for a nil receiver.
//
// Why it matters: callers treat uuid.Nil as the sentinel for "no rack", so a nil
// receiver must yield it rather than panic.
// Inputs: a nil *CaniRackType. Outputs: the UUID, expected uuid.Nil.
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestGetIDRackNil(t *testing.T) {
	var r *CaniRackType
	if got := r.GetID(); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %v", got)
	}
}

// --- GetSlug ---

// TestGetSlugRackHappyPath verifies GetSlug returns the rack's slug.
//
// Why it matters: the slug links a rack to its hardware-library template, so the
// accessor must surface it unchanged for lookups and export.
// Inputs: a rack with slug "std-42u". Outputs: the slug string, expected
// "std-42u".
// Data choice: a realistic rack slug is an arbitrary non-empty value showing the
// stored field is returned verbatim.
func TestGetSlugRackHappyPath(t *testing.T) {
	r := &CaniRackType{Slug: "std-42u"}
	if got := r.GetSlug(); got != "std-42u" {
		t.Fatalf("expected std-42u, got %s", got)
	}
}

// TestGetSlugRackNil verifies GetSlug returns "" for a nil receiver.
//
// Why it matters: a nil rack must degrade to "" rather than panic when its slug
// is read.
// Inputs: a nil *CaniRackType. Outputs: the slug string, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestGetSlugRackNil(t *testing.T) {
	var r *CaniRackType
	if got := r.GetSlug(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetVendor ---

// TestGetVendorRackHappyPath verifies GetVendor returns the rack's Manufacturer.
//
// Why it matters: vendor identifies the hardware maker for classification and
// export, and GetVendor maps the Manufacturer field onto that shared accessor.
// Inputs: a rack with Manufacturer "Acme". Outputs: the vendor string, expected
// "Acme".
// Data choice: "Acme" is an arbitrary maker demonstrating the
// Manufacturer-to-vendor mapping returns the stored value.
func TestGetVendorRackHappyPath(t *testing.T) {
	r := &CaniRackType{Manufacturer: "Acme"}
	if got := r.GetVendor(); got != "Acme" {
		t.Fatalf("expected Acme, got %s", got)
	}
}

// TestGetVendorRackNil verifies GetVendor returns "" for a nil receiver.
//
// Why it matters: a nil rack must report no vendor rather than panic when its
// maker is queried.
// Inputs: a nil *CaniRackType. Outputs: the vendor string, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestGetVendorRackNil(t *testing.T) {
	var r *CaniRackType
	if got := r.GetVendor(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetType ---

// TestGetTypeRackHappyPath verifies GetType returns an explicitly set Type rather
// than the default.
//
// Why it matters: an operator-set type must override the generic fallback so the
// rack classifies correctly during export.
// Inputs: a rack with Type "custom-rack". Outputs: the Type, expected
// "custom-rack".
// Data choice: a custom type distinct from the TypeRack default proves the
// explicit value is preferred over the fallback.
func TestGetTypeRackHappyPath(t *testing.T) {
	r := &CaniRackType{Type: "custom-rack"}
	if got := r.GetType(); got != Type("custom-rack") {
		t.Fatalf("expected custom-rack, got %s", got)
	}
}

// TestGetTypeRackNil verifies GetType returns the empty Type for a nil receiver.
//
// Why it matters: rack type drives classification and export, and a nil rack must
// report "unknown" rather than be silently treated as a generic rack.
// Inputs: a nil *CaniRackType. Outputs: the Type result, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard ahead
// of the TypeRack fallback.
func TestGetTypeRackNil(t *testing.T) {
	var r *CaniRackType
	if got := r.GetType(); got != "" {
		t.Fatalf("expected empty Type, got %s", got)
	}
}

// --- GetStatus ---

// TestGetStatusRackHappyPath verifies GetStatus returns the rack's status from its
// embedded ObjectMeta.
//
// Why it matters: status drives lifecycle handling and export filtering, so the
// accessor must surface the stored value.
// Inputs: a rack with Status "Active". Outputs: the status string, expected
// "Active".
// Data choice: "Active" is the common in-service status, a representative
// non-empty value proving the embedded field is returned.
func TestGetStatusRackHappyPath(t *testing.T) {
	r := &CaniRackType{ObjectMeta: ObjectMeta{Status: "Active"}}
	if got := r.GetStatus(); got != "Active" {
		t.Fatalf("expected Active, got %s", got)
	}
}

// TestGetStatusRackNil verifies GetStatus returns "" for a nil receiver.
//
// Why it matters: a nil rack must report no status rather than panic when its
// lifecycle state is queried.
// Inputs: a nil *CaniRackType. Outputs: the status string, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestGetStatusRackNil(t *testing.T) {
	var r *CaniRackType
	if got := r.GetStatus(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- isSlotBlocked ---

// TestIsSlotBlockedHappyPath verifies isSlotBlocked reports an empty slot as not
// blocked.
//
// Why it matters: placement scans slots for conflicts, so a vacant U-position
// must read as free or no device could ever be placed.
// Inputs: a rack with an empty OccupiedSlots map, queried at U1 front. Outputs:
// the blocked boolean, expected false.
// Data choice: an empty slot map drives the slot==nil early return, the baseline
// "free" path.
func TestIsSlotBlockedHappyPath(t *testing.T) {
	r := &CaniRackType{
		UHeight:       42,
		OccupiedSlots: map[int]map[string]uuid.UUID{},
	}
	if r.isSlotBlocked(1, FaceFront, false) {
		t.Fatal("expected empty slot to not be blocked")
	}
}

// TestIsSlotBlockedFullDepthConflict verifies isSlotBlocked reports a slot held by
// a full-depth device as blocked on the front face.
//
// Why it matters: a full-depth occupant consumes both faces, so any face query
// must see it as blocked to prevent overlapping placement.
// Inputs: a rack whose slot 5 holds a FaceFull occupant, queried at U5 front.
// Outputs: the blocked boolean, expected true.
// Data choice: a FaceFull entry drives the has-full-depth short-circuit branch,
// independent of the queried face.
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

// TestHasFaceOccupantHappyPath verifies hasFaceOccupant reports an empty slot as
// having no occupant on the queried face.
//
// Why it matters: face-level occupancy is the primitive isSlotBlocked builds on,
// so an empty slot must read as unoccupied.
// Inputs: an empty slot map, queried for the front face. Outputs: the occupied
// boolean, expected false.
// Data choice: an empty slot map isolates the not-present path of the map lookup.
func TestHasFaceOccupantHappyPath(t *testing.T) {
	r := &CaniRackType{}
	slot := map[string]uuid.UUID{}
	if r.hasFaceOccupant(slot, FaceFront) {
		t.Fatal("expected empty slot to have no occupant")
	}
}

// TestHasFaceOccupantOccupied verifies hasFaceOccupant reports a face as occupied
// when the slot holds a device on it.
//
// Why it matters: detecting a face occupant is what lets placement reject a
// conflicting position.
// Inputs: a slot with a rear occupant, queried for the rear face. Outputs: the
// occupied boolean, expected true.
// Data choice: occupying the rear face and querying rear drives the present path
// of the map lookup for a non-front face.
func TestHasFaceOccupantOccupied(t *testing.T) {
	r := &CaniRackType{}
	slot := map[string]uuid.UUID{FaceRear: uuid.New()}
	if !r.hasFaceOccupant(slot, FaceRear) {
		t.Fatal("expected rear face to be occupied")
	}
}

// --- CanFitDevice ---

// TestCanFitDeviceHappyPath verifies CanFitDevice reports that a device fits in an
// empty rack within bounds.
//
// Why it matters: CanFitDevice is the pre-flight check before placement, so a
// clearly free position must read as fittable.
// Inputs: an empty 42U rack and a 2U candidate at U1 front. Outputs: the fit
// boolean, expected true.
// Data choice: a 2U device at U1 in a 42U rack is comfortably in-bounds and
// unobstructed, exercising the all-clear path.
func TestCanFitDeviceHappyPath(t *testing.T) {
	r := &CaniRackType{
		UHeight:       42,
		OccupiedSlots: map[int]map[string]uuid.UUID{},
	}
	if !r.CanFitDevice(1, 2, FaceFront, false) {
		t.Fatal("expected 2U device to fit at U1 in empty rack")
	}
}

// TestCanFitDeviceOutOfBounds verifies CanFitDevice rejects a device whose span
// extends past the top of the rack.
//
// Why it matters: a device must not be placed beyond the rack's physical height,
// so the bounds check must reject an over-tall span.
// Inputs: a 10U rack and a 4U candidate at U9 (spanning U9-U12). Outputs: the fit
// boolean, expected false.
// Data choice: startU 9 + height 4 exceeds UHeight 10, driving the endU>UHeight
// rejection specifically.
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

// TestPlaceDeviceHappyPath verifies PlaceDevice places a device in an empty rack
// and records it in the Devices list.
//
// Why it matters: placement is how a device acquires a physical position and
// becomes tracked in the rack's inventory.
// Inputs: an empty 42U rack and a 2U device at U1 front. Outputs: the placement
// boolean (expected true) and the Devices slice (expected to contain the ID).
// Data choice: asserting both the boolean and the Devices entry confirms the
// device is both fitted and registered, not merely accepted.
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

// TestPlaceDeviceConflict verifies PlaceDevice rejects a device whose slots are
// already occupied.
//
// Why it matters: two devices cannot share the same U-position and face, so the
// second placement must fail to preserve physical reality.
// Inputs: a rack with device A at U1-U2 front, then device B attempted at the
// same span. Outputs: the second placement boolean, expected false.
// Data choice: placing B at the identical position and face as A drives the
// occupied-slot rejection through isSlotBlocked.
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

// TestRemoveDeviceHappyPath verifies RemoveDevice frees a placed device's slots
// and drops it from the Devices list.
//
// Why it matters: removing a device must fully release its rack position so the
// space can be reused and the inventory stays accurate.
// Inputs: a rack with a 3U device at U5, then removed. Outputs: the Devices and
// OccupiedSlots collections, both expected empty.
// Data choice: a multi-U device confirms every occupied slot is cleared, not just
// the starting U.
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

// TestRemoveDeviceNil verifies RemoveDevice is a no-op on a nil receiver and does
// not panic.
//
// Why it matters: cleanup paths may call RemoveDevice on a missing rack, so a nil
// receiver must be handled gracefully.
// Inputs: a nil *CaniRackType and a random device ID. Outputs: none; the test
// passes if the call returns without panicking.
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestRemoveDeviceNil(t *testing.T) {
	var r *CaniRackType
	// should not panic on nil receiver
	r.RemoveDevice(uuid.New())
}

// --- GetDeviceStartU ---

// TestGetDeviceStartURackHappyPath verifies GetDeviceStartU returns the lowest
// U-position a device occupies.
//
// Why it matters: the starting U anchors a device's position for display, swaps,
// and export, so it must report the true bottom of its span.
// Inputs: a rack with a 4U device placed at U10. Outputs: the start-U, expected
// 10.
// Data choice: placing at U10 with height 4 (U10-U13) confirms the minimum U is
// returned, not an arbitrary occupied slot.
func TestGetDeviceStartURackHappyPath(t *testing.T) {
	devID := uuid.New()
	r := &CaniRackType{UHeight: 42}
	r.PlaceDevice(devID, 10, 4, FaceFront, false)
	if got := r.GetDeviceStartU(devID); got != 10 {
		t.Fatalf("expected startU=10, got %d", got)
	}
}

// TestGetDeviceStartURackNil verifies GetDeviceStartU returns 0 for a nil
// receiver.
//
// Why it matters: callers treat 0 as "not placed", so a nil rack must yield it
// rather than panic.
// Inputs: a nil *CaniRackType and a random device ID. Outputs: the start-U,
// expected 0.
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestGetDeviceStartURackNil(t *testing.T) {
	var r *CaniRackType
	if got := r.GetDeviceStartU(uuid.New()); got != 0 {
		t.Fatalf("expected 0 for nil receiver, got %d", got)
	}
}

// --- addDevice ---

// TestAddDeviceHappyPath verifies addDevice appends a new device to the Devices
// list.
//
// Why it matters: the Devices list is the rack's roster of contained hardware, so
// adding a new device must record it.
// Inputs: an empty rack and one device ID. Outputs: the Devices slice length,
// expected 1.
// Data choice: a single add to an empty list isolates the append path.
func TestAddDeviceHappyPath(t *testing.T) {
	devID := uuid.New()
	r := &CaniRackType{}
	r.addDevice(devID)
	if len(r.Devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(r.Devices))
	}
}

// TestAddDeviceDuplicate verifies addDevice ignores a device already in the
// Devices list.
//
// Why it matters: the roster must not list the same device twice, so a repeat add
// must be a no-op.
// Inputs: a rack with the same device ID added twice. Outputs: the Devices slice
// length, expected 1.
// Data choice: adding the identical ID drives the already-present short-circuit;
// a distinct ID would instead exercise the append path.
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

// TestFindNextAvailableSlotHappyPath verifies FindNextAvailableSlot returns the
// next free top-down position below an occupied region.
//
// Why it matters: the scanner places devices top-to-bottom, so it must return the
// highest free slot that fits the requested height.
// Inputs: a 42U rack with a 2U device at U41-U42, searching for a 2U front slot.
// Outputs: the start-U, expected 39.
// Data choice: with U41-U42 taken, the next 2U block top-down is U39-U40, so 39
// proves the scan steps past the occupied region rather than returning the top.
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

// TestFindNextAvailableSlotFull verifies FindNextAvailableSlot returns 0 when no
// position can fit the device.
//
// Why it matters: a full rack must report "no space" so callers don't place a
// device into an occupied slot.
// Inputs: a 4U rack fully occupied by a full-depth device, searching for a 1U
// slot. Outputs: the start-U, expected 0.
// Data choice: a full-depth device spanning the entire 4U rack guarantees every
// candidate position is blocked, driving the no-slot return.
func TestFindNextAvailableSlotFull(t *testing.T) {
	r := &CaniRackType{UHeight: 4}
	r.PlaceDevice(uuid.New(), 1, 4, FaceFront, true)
	got := r.FindNextAvailableSlot(1, FaceFront, false)
	if got != 0 {
		t.Fatalf("expected 0 when rack is full, got %d", got)
	}
}

// --- MigrateLegacySlots ---

// TestMigrateLegacySlotsHappyPath verifies MigrateLegacySlots converts legacy
// face-less slots into front-face entries.
//
// Why it matters: old inventory used a face-less slot map, so migration must move
// each occupant to the front face for the current face-aware model.
// Inputs: a rack and a legacy map placing one device at U1 and U2. Outputs: the
// OccupiedSlots front-face entries at U1 and U2, expected to hold the device.
// Data choice: two legacy slots for one device confirm every entry is migrated,
// each landing under FaceFront.
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

// TestMigrateLegacySlotsEmpty verifies MigrateLegacySlots initializes the slot map
// even when the legacy input is empty.
//
// Why it matters: downstream code ranges over OccupiedSlots, so migration must
// leave a non-nil (if empty) map rather than nil.
// Inputs: a rack and an empty legacy map. Outputs: the OccupiedSlots map, expected
// non-nil with length 0.
// Data choice: an empty legacy map isolates the map-initialization path from any
// occupant migration.
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

// --- GetSlotOccupant ---

// TestGetSlotOccupantHappyPath verifies GetSlotOccupant returns the device mounted
// at a given U-position and face.
//
// Why it matters: slot lookup answers "what is mounted here", the basis for
// display and conflict checks.
// Inputs: a rack with a 1U device at U5 front, queried at U5 front. Outputs: the
// occupant UUID, expected the placed device.
// Data choice: querying the exact U and face a device was placed at drives the
// matching-face return.
func TestGetSlotOccupantHappyPath(t *testing.T) {
	devID := uuid.New()
	r := &CaniRackType{UHeight: 42}
	r.PlaceDevice(devID, 5, 1, FaceFront, false)
	if got := r.GetSlotOccupant(5, FaceFront); got != devID {
		t.Fatalf("expected %v, got %v", devID, got)
	}
}

// TestGetSlotOccupantEmpty verifies GetSlotOccupant returns uuid.Nil for an empty
// U-position.
//
// Why it matters: a vacant slot must report Nil so callers can tell it is free.
// Inputs: an empty 42U rack queried at U1 front. Outputs: the occupant UUID,
// expected uuid.Nil.
// Data choice: an unpopulated rack drives the slot==nil early return.
func TestGetSlotOccupantEmpty(t *testing.T) {
	r := &CaniRackType{UHeight: 42}
	if got := r.GetSlotOccupant(1, FaceFront); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil for empty slot, got %v", got)
	}
}

// --- GetDeviceFace ---

// TestGetDeviceFaceHappyPath verifies GetDeviceFace returns the face a device is
// mounted on.
//
// Why it matters: a device's face affects cabling and airflow, so the rack must
// report where it was placed.
// Inputs: a rack with a device placed at U3 on the rear face. Outputs: the face
// string, expected FaceRear.
// Data choice: using the rear face (not the default front) proves the stored face
// is returned rather than a default.
func TestGetDeviceFaceHappyPath(t *testing.T) {
	devID := uuid.New()
	r := &CaniRackType{UHeight: 42}
	r.PlaceDevice(devID, 3, 1, FaceRear, false)
	if got := r.GetDeviceFace(devID); got != FaceRear {
		t.Fatalf("expected %s, got %s", FaceRear, got)
	}
}

// TestGetDeviceFaceNil verifies GetDeviceFace returns "" for a nil receiver.
//
// Why it matters: a nil rack must report no face rather than panic when queried.
// Inputs: a nil *CaniRackType and a random device ID. Outputs: the face string,
// expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestGetDeviceFaceNil(t *testing.T) {
	var r *CaniRackType
	if got := r.GetDeviceFace(uuid.New()); got != "" {
		t.Fatalf("expected empty string for nil receiver, got %s", got)
	}
}

// --- GetDeviceHeight ---

// TestGetDeviceHeightHappyPath verifies GetDeviceHeight returns the number of
// U-slots a device occupies.
//
// Why it matters: a device's height drives layout and swap math, so the rack must
// count its occupied slots accurately.
// Inputs: a rack with a 4U device at U1. Outputs: the height, expected 4.
// Data choice: a 4U device confirms every occupied U is counted, not just the
// starting slot.
func TestGetDeviceHeightHappyPath(t *testing.T) {
	devID := uuid.New()
	r := &CaniRackType{UHeight: 42}
	r.PlaceDevice(devID, 1, 4, FaceFront, false)
	if got := r.GetDeviceHeight(devID); got != 4 {
		t.Fatalf("expected height 4, got %d", got)
	}
}

// TestGetDeviceHeightNil verifies GetDeviceHeight returns 0 for a nil receiver.
//
// Why it matters: callers treat 0 as "not placed", so a nil rack must yield it
// rather than panic.
// Inputs: a nil *CaniRackType and a random device ID. Outputs: the height,
// expected 0.
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestGetDeviceHeightNil(t *testing.T) {
	var r *CaniRackType
	if got := r.GetDeviceHeight(uuid.New()); got != 0 {
		t.Fatalf("expected 0 for nil receiver, got %d", got)
	}
}

// --- SwapDevices ---

// TestSwapDevicesHappyPath verifies SwapDevices exchanges the positions of two
// placed devices.
//
// Why it matters: operators reorder hardware, so a successful swap must leave each
// device at the other's former start-U while preserving height and face.
// Inputs: a rack with device A at U1 and device B at U2, then swapped. Outputs:
// the post-swap start-U of each, expected A at U2 and B at U1.
// Data choice: two adjacent 1U front devices make the exchange unambiguous and
// the expected positions easy to verify.
func TestSwapDevicesHappyPath(t *testing.T) {
	devA := uuid.New()
	devB := uuid.New()
	r := &CaniRackType{UHeight: 42}
	r.PlaceDevice(devA, 1, 1, FaceFront, false)
	r.PlaceDevice(devB, 2, 1, FaceFront, false)
	if err := r.SwapDevices(devA, devB); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := r.GetDeviceStartU(devA); got != 2 {
		t.Fatalf("expected devA at U2 after swap, got U%d", got)
	}
	if got := r.GetDeviceStartU(devB); got != 1 {
		t.Fatalf("expected devB at U1 after swap, got U%d", got)
	}
}

// TestSwapDevicesNil verifies SwapDevices returns an error for a nil receiver.
//
// Why it matters: a nil rack cannot host a swap, so the method must fail loudly
// rather than dereference nil and panic.
// Inputs: a nil *CaniRackType and two random device IDs. Outputs: an error,
// non-nil expected.
// Data choice: a nil receiver is the only input that reaches the nil guard ahead
// of the placement lookups.
func TestSwapDevicesNil(t *testing.T) {
	var r *CaniRackType
	if err := r.SwapDevices(uuid.New(), uuid.New()); err == nil {
		t.Fatal("expected error for nil receiver")
	}
}

// ========== additional edge-case / branch coverage ==========

// --- Validate (UHeight) ---

// TestValidateRackRejectsZeroUHeight verifies Validate returns an error when the
// rack's UHeight is below 1.
//
// Why it matters: a rack with no usable U-slots cannot hold devices, so Validate
// must reject it before it enters inventory and corrupts placement math.
// Inputs: a non-nil rack with UHeight 0. Outputs: an error, non-nil expected.
// Data choice: UHeight 0 is the boundary just below the minimum of 1, isolating
// the UHeight check from the nil-receiver guard the existing nil test covers.
func TestValidateRackRejectsZeroUHeight(t *testing.T) {
	r := &CaniRackType{UHeight: 0}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for UHeight < 1")
	}
}

// --- GetType (default) ---

// TestGetTypeRackReturnsDefault verifies GetType returns the TypeRack default when
// Type is unset on a non-nil rack.
//
// Why it matters: a rack without an explicit type must classify as a generic rack
// rather than an empty type during export.
// Inputs: a rack with an empty Type. Outputs: the Type, expected TypeRack.
// Data choice: an empty (non-nil) struct isolates the unset-Type fallback branch,
// distinct from the explicit-type and nil-receiver cases the existing tests
// cover.
func TestGetTypeRackReturnsDefault(t *testing.T) {
	r := &CaniRackType{}
	if got := r.GetType(); got != TypeRack {
		t.Fatalf("expected %s, got %s", TypeRack, got)
	}
}

// --- isSlotBlocked (full-depth) ---

// TestIsSlotBlockedFullDepthChecksBothFaces verifies isSlotBlocked reports a
// conflict for a full-depth candidate when the slot has a rear occupant.
//
// Why it matters: a full-depth device spans both faces, so placement must detect
// an occupant on either face to prevent physically overlapping hardware.
// Inputs: a rack whose slot 5 holds a rear occupant, queried with isFullDepth
// true on the front face. Outputs: the blocked boolean, expected true.
// Data choice: occupying only the rear face (not FaceFull) forces the
// full-depth branch to OR both faces; the existing full-depth test uses a
// FaceFull occupant and returns earlier, so this path was unreached.
func TestIsSlotBlockedFullDepthChecksBothFaces(t *testing.T) {
	devID := uuid.New()
	r := &CaniRackType{
		UHeight: 42,
		OccupiedSlots: map[int]map[string]uuid.UUID{
			5: {FaceRear: devID},
		},
	}
	if !r.isSlotBlocked(5, FaceFront, true) {
		t.Fatal("expected full-depth check to detect the rear occupant")
	}
}

// --- CanFitDevice (guards) ---

// TestCanFitDeviceRejectsInvalidArgs verifies CanFitDevice returns false for a nil
// receiver, a startU below 1, or a height below 1.
//
// Why it matters: placement math depends on positive, in-range coordinates, so
// invalid inputs must be rejected up front rather than producing a bogus fit.
// Inputs: per case, a receiver (nil or valid) plus startU/height. Outputs: the
// fit boolean, expected false in every case.
// Data choice: each case violates exactly one guard clause (nil, startU 0,
// height 0) so the false result is attributable to that specific condition,
// branches the happy-path and out-of-bounds tests never reach.
func TestCanFitDeviceRejectsInvalidArgs(t *testing.T) {
	valid := &CaniRackType{UHeight: 42, OccupiedSlots: map[int]map[string]uuid.UUID{}}
	cases := []struct {
		name   string
		r      *CaniRackType
		startU int
		height int
	}{
		{"nil receiver", nil, 1, 1},
		{"startU below 1", valid, 0, 1},
		{"height below 1", valid, 1, 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.r.CanFitDevice(tt.startU, tt.height, FaceFront, false) {
				t.Errorf("expected CanFitDevice to reject %s", tt.name)
			}
		})
	}
}

// TestCanFitDeviceDefaultsEmptyFace verifies CanFitDevice treats an empty face as
// the front face and reports a fit in an empty rack.
//
// Why it matters: callers may omit the face, so the default must resolve to front
// rather than mismatch the slot map and wrongly reject a placement.
// Inputs: an empty 42U rack and a 2U candidate at U1 with face "". Outputs: the
// fit boolean, expected true.
// Data choice: passing an empty face string drives the face-default branch; an
// empty rack guarantees the only variable under test is the face defaulting.
func TestCanFitDeviceDefaultsEmptyFace(t *testing.T) {
	r := &CaniRackType{UHeight: 42, OccupiedSlots: map[int]map[string]uuid.UUID{}}
	if !r.CanFitDevice(1, 2, "", false) {
		t.Fatal("expected empty face to default to front and fit")
	}
}

// --- PlaceDevice (default face) ---

// TestPlaceDeviceDefaultsEmptyFace verifies PlaceDevice stores a device under the
// front face when called with an empty face string.
//
// Why it matters: placement is how a device acquires its physical rack position,
// so an omitted face must default to front consistently with CanFitDevice rather
// than create an unreadable "" face entry.
// Inputs: an empty 42U rack and a 1U device placed at U1 with face "". Outputs:
// the placement boolean (expected true) and the front-face occupant of U1.
// Data choice: asserting the occupant via GetSlotOccupant(1, FaceFront) confirms
// the empty face resolved to front, the branch the explicit-face tests skip.
func TestPlaceDeviceDefaultsEmptyFace(t *testing.T) {
	devID := uuid.New()
	r := &CaniRackType{UHeight: 42}
	if !r.PlaceDevice(devID, 1, 1, "", false) {
		t.Fatal("expected placement with empty face to succeed")
	}
	if got := r.GetSlotOccupant(1, FaceFront); got != devID {
		t.Fatalf("expected device at front face, got %v", got)
	}
}

// --- SwapDevices (error + rollback) ---

// TestSwapDevicesErrorPaths verifies SwapDevices rejects a swap when a device is
// unplaced and rolls both devices back to their original positions when either
// post-swap placement fails.
//
// Why it matters: SwapDevices promises an atomic exchange, so any failure must
// leave the rack exactly as it was rather than half-applied or with a device
// lost.
// Inputs: per case, a rack plus two device IDs in a state that triggers one
// failure mode. Outputs: an error containing the case substring, and the devices'
// start-U positions, expected unchanged from before the call.
// Data choice: the cases are built so each hits a distinct branch — an unplaced
// devB (startB 0), a tall devA that cannot fit at devB's high slot (first
// rollback), and a tall devB that collides with the just-moved devA (second
// rollback) — and each asserts restoration to prove atomicity.
func TestSwapDevicesErrorPaths(t *testing.T) {
	cases := []struct {
		name    string
		setup   func() (*CaniRackType, uuid.UUID, uuid.UUID)
		wantErr string
	}{
		{
			name: "one device not placed",
			setup: func() (*CaniRackType, uuid.UUID, uuid.UUID) {
				devA, devB := uuid.New(), uuid.New()
				r := &CaniRackType{UHeight: 42}
				r.PlaceDevice(devA, 1, 1, FaceFront, false)
				return r, devA, devB
			},
			wantErr: "both devices must be placed",
		},
		{
			name: "rollback when first placement fails",
			setup: func() (*CaniRackType, uuid.UUID, uuid.UUID) {
				devA, devB := uuid.New(), uuid.New()
				r := &CaniRackType{UHeight: 5}
				r.PlaceDevice(devA, 1, 4, FaceFront, false)
				r.PlaceDevice(devB, 5, 1, FaceFront, false)
				return r, devA, devB
			},
			wantErr: "cannot place device A",
		},
		{
			name: "rollback when second placement fails",
			setup: func() (*CaniRackType, uuid.UUID, uuid.UUID) {
				devA, devB := uuid.New(), uuid.New()
				r := &CaniRackType{UHeight: 42}
				r.PlaceDevice(devA, 1, 1, FaceFront, false)
				r.PlaceDevice(devB, 2, 4, FaceFront, false)
				return r, devA, devB
			},
			wantErr: "cannot place device B",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r, idA, idB := tt.setup()
			startA := r.GetDeviceStartU(idA)
			startB := r.GetDeviceStartU(idB)
			err := r.SwapDevices(idA, idB)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
			if got := r.GetDeviceStartU(idA); got != startA {
				t.Errorf("device A moved to U%d, want original U%d", got, startA)
			}
			if startB != 0 {
				if got := r.GetDeviceStartU(idB); got != startB {
					t.Errorf("device B moved to U%d, want original U%d", got, startB)
				}
			}
		})
	}
}

// --- GetSlotOccupant (edge cases) ---

// TestGetSlotOccupantEdgeCases verifies GetSlotOccupant returns uuid.Nil for a nil
// receiver, defaults an empty face to front, matches a full-depth occupant on any
// face, and returns uuid.Nil when the queried face is unoccupied.
//
// Why it matters: slot lookup answers "what is mounted here", so it must be
// nil-safe, honor the front default, surface full-depth devices regardless of the
// queried face, and report an empty face as vacant.
// Inputs: per case, a receiver (nil or populated), a U-position, and a face.
// Outputs: the occupant UUID, expected Nil or the matching device per case.
// Data choice: separate front-mounted and full-depth racks let each case drive a
// distinct branch — nil guard, empty-face default, the FaceFull short-circuit, and
// the final unoccupied-face return — none of which the existing happy/empty tests
// reach.
func TestGetSlotOccupantEdgeCases(t *testing.T) {
	frontDev := uuid.New()
	frontRack := &CaniRackType{UHeight: 42}
	frontRack.PlaceDevice(frontDev, 5, 1, FaceFront, false)

	fullDev := uuid.New()
	fullRack := &CaniRackType{UHeight: 42}
	fullRack.PlaceDevice(fullDev, 7, 1, FaceFront, true)

	cases := []struct {
		name string
		r    *CaniRackType
		u    int
		face string
		want uuid.UUID
	}{
		{"nil receiver", nil, 5, FaceFront, uuid.Nil},
		{"empty face defaults to front", frontRack, 5, "", frontDev},
		{"full-depth matches any face", fullRack, 7, FaceRear, fullDev},
		{"unoccupied face returns nil", frontRack, 5, FaceRear, uuid.Nil},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.GetSlotOccupant(tt.u, tt.face); got != tt.want {
				t.Errorf("GetSlotOccupant(%d, %q) = %v, want %v", tt.u, tt.face, got, tt.want)
			}
		})
	}
}

// --- GetDeviceFace (not found) ---

// TestGetDeviceFaceReturnsEmptyWhenNotFound verifies GetDeviceFace returns "" for a
// device absent from a populated rack.
//
// Why it matters: callers use the empty string to mean "this device is not in the
// rack", so the scan must report it rather than a stale or default face.
// Inputs: a rack holding one placed device, queried with a different random ID.
// Outputs: the face string, expected "".
// Data choice: placing one device makes the slot map non-empty so the search loop
// actually runs, then querying an unrelated UUID exercises the not-found return
// the nil-receiver test cannot reach.
func TestGetDeviceFaceReturnsEmptyWhenNotFound(t *testing.T) {
	r := &CaniRackType{UHeight: 42}
	r.PlaceDevice(uuid.New(), 1, 1, FaceFront, false)
	if got := r.GetDeviceFace(uuid.New()); got != "" {
		t.Fatalf("expected empty face for absent device, got %q", got)
	}
}

// --- FindNextAvailableSlot (guard) ---

// TestFindNextAvailableSlotRejectsInvalidHeight verifies FindNextAvailableSlot
// returns 0 for a non-positive height.
//
// Why it matters: a zero or negative height has no valid placement, so the search
// must short-circuit to 0 ("no slot") rather than scan or underflow the loop
// bounds.
// Inputs: a 42U rack and height 0. Outputs: the start-U result, expected 0.
// Data choice: height 0 is the boundary just below the minimum of 1, isolating
// the guard from the rack-full case the existing test covers.
func TestFindNextAvailableSlotRejectsInvalidHeight(t *testing.T) {
	r := &CaniRackType{UHeight: 42}
	if got := r.FindNextAvailableSlot(0, FaceFront, false); got != 0 {
		t.Fatalf("expected 0 for non-positive height, got %d", got)
	}
}
