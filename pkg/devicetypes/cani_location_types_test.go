package devicetypes

// Test coverage for CaniLocationType methods.
//
// | Function  | Happy-path test                          | Failure test                               |
// |-----------|------------------------------------------|--------------------------------------------|
// | Validate  | TestValidateLocationHappyPath             | TestValidateLocationFailure                |
// | GetID     | TestGetIDLocationHappyPath                | TestGetIDLocationNilReceiver               |
// | GetSlug   | TestGetSlugLocationHappyPath              | TestGetSlugLocationNilReceiver             |
// | GetStatus | TestGetStatusLocationHappyPath            | TestGetStatusLocationNilReceiver           |
// | AddRack   | TestAddRackHappyPath                      | TestAddRackDuplicate                       |
// | AddChild  | TestAddChildHappyPath                     | TestAddChildDuplicate                      |

import (
	"testing"

	"github.com/google/uuid"
)

// --- Validate ---

// TestValidateLocationHappyPath verifies Validate returns nil for a location that
// has a non-empty LocationType.
//
// Why it matters: Validate gates a location before it enters the inventory
// hierarchy (building → floor → room), so a fully formed location must pass.
// Inputs: a location with ID, Name, LocationType "building", and Active status.
// Outputs: an error, nil expected.
// Data choice: a populated "building" with every identity field set represents a
// realistic top-level location, isolating the success path where LocationType is
// present.
func TestValidateLocationHappyPath(t *testing.T) {
	loc := CaniLocationType{
		ID:           uuid.New(),
		Name:         "Building-A",
		LocationType: "building",
		ObjectMeta:   ObjectMeta{Status: "Active"},
	}
	if err := loc.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestValidateLocationFailure verifies Validate returns an error when
// LocationType is empty.
//
// Why it matters: every location must declare its kind (site, building, room),
// so an unclassified location must be rejected before it joins the hierarchy.
// Inputs: a location with ID and Name but an empty LocationType. Outputs: an
// error, non-nil expected.
// Data choice: setting only LocationType to "" while leaving other fields valid
// isolates the empty-LocationType branch as the sole reason for failure.
func TestValidateLocationFailure(t *testing.T) {
	loc := CaniLocationType{
		ID:           uuid.New(),
		Name:         "NoType",
		LocationType: "",
	}
	if err := loc.Validate(); err == nil {
		t.Fatal("expected error for empty LocationType, got nil")
	}
}

// --- GetID ---

// TestGetIDLocationHappyPath verifies GetID returns the location's stored UUID.
//
// Why it matters: the ID uniquely identifies a location across the hierarchy and
// export, so the accessor must return it unchanged for parent/child linking.
// Inputs: a location with a generated ID. Outputs: the UUID, expected equal to
// the stored value.
// Data choice: a freshly generated uuid is an arbitrary distinct value, proving
// the getter returns the stored field rather than a constant.
func TestGetIDLocationHappyPath(t *testing.T) {
	id := uuid.New()
	loc := CaniLocationType{ID: id, LocationType: "site"}
	if got := loc.GetID(); got != id {
		t.Fatalf("expected %v, got %v", id, got)
	}
}

// TestGetIDLocationNilReceiver verifies GetID returns uuid.Nil for a nil
// receiver.
//
// Why it matters: callers treat uuid.Nil as the sentinel for "no location", so a
// nil receiver must yield it rather than panic.
// Inputs: a nil *CaniLocationType. Outputs: the UUID, expected uuid.Nil.
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestGetIDLocationNilReceiver(t *testing.T) {
	var loc *CaniLocationType
	if got := loc.GetID(); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %v", got)
	}
}

// --- GetSlug ---

// TestGetSlugLocationHappyPath verifies GetSlug returns the LocationType string,
// which serves as the location's slug.
//
// Why it matters: GetSlug maps a location's kind onto the generic slug accessor
// used for grouping and export, so it must surface LocationType verbatim.
// Inputs: a location with LocationType "room". Outputs: the slug string, expected
// "room".
// Data choice: "room" is a representative leaf location kind, showing the slug is
// the LocationType value rather than a separate field.
func TestGetSlugLocationHappyPath(t *testing.T) {
	loc := CaniLocationType{LocationType: "room"}
	if got := loc.GetSlug(); got != "room" {
		t.Fatalf("expected %q, got %q", "room", got)
	}
}

// TestGetSlugLocationNilReceiver verifies GetSlug returns "" for a nil receiver.
//
// Why it matters: a nil location must degrade to "" rather than panic when its
// slug is read.
// Inputs: a nil *CaniLocationType. Outputs: the slug string, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestGetSlugLocationNilReceiver(t *testing.T) {
	var loc *CaniLocationType
	if got := loc.GetSlug(); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

// --- GetStatus ---

// TestGetStatusLocationHappyPath verifies GetStatus returns the location's status
// from its embedded ObjectMeta.
//
// Why it matters: status drives lifecycle handling and export filtering, so the
// accessor must surface the stored value.
// Inputs: a location with Status "Active". Outputs: the status string, expected
// "Active".
// Data choice: "Active" is the common in-service status, a representative
// non-empty value proving the embedded field is returned.
func TestGetStatusLocationHappyPath(t *testing.T) {
	loc := CaniLocationType{LocationType: "floor", ObjectMeta: ObjectMeta{Status: "Active"}}
	if got := loc.GetStatus(); got != "Active" {
		t.Fatalf("expected %q, got %q", "Active", got)
	}
}

// TestGetStatusLocationNilReceiver verifies GetStatus returns "" for a nil
// receiver.
//
// Why it matters: a nil location must report no status rather than panic when its
// lifecycle state is queried.
// Inputs: a nil *CaniLocationType. Outputs: the status string, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestGetStatusLocationNilReceiver(t *testing.T) {
	var loc *CaniLocationType
	if got := loc.GetStatus(); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

// --- AddRack ---

// TestAddRackHappyPath verifies AddRack appends a rack UUID to the location's
// rack list.
//
// Why it matters: a location's rack list anchors racks (and the devices within
// them) to a physical place, so adding a new rack must record its ID.
// Inputs: a location and one rack UUID. Outputs: the Racks slice, expected length
// 1 containing that UUID.
// Data choice: asserting both the length and the stored ID (not just the count)
// confirms the actual rack landed in the list rather than a placeholder.
func TestAddRackHappyPath(t *testing.T) {
	loc := CaniLocationType{LocationType: "room"}
	rackID := uuid.New()
	loc.AddRack(rackID)
	if len(loc.Racks) != 1 || loc.Racks[0] != rackID {
		t.Fatalf("expected Racks to contain %v, got %v", rackID, loc.Racks)
	}
}

// TestAddRackDuplicate verifies AddRack is idempotent and does not append a rack
// UUID that is already present.
//
// Why it matters: rack lists are rebuilt at load time and AddRack may be called
// repeatedly, so re-adding the same rack must not create a duplicate entry.
// Inputs: a location with the same rack UUID added twice. Outputs: the Racks
// slice, expected length 1.
// Data choice: adding the identical UUID twice isolates the already-present
// short-circuit; a distinct ID would instead exercise the append path.
func TestAddRackDuplicate(t *testing.T) {
	loc := CaniLocationType{LocationType: "room"}
	rackID := uuid.New()
	loc.AddRack(rackID)
	loc.AddRack(rackID) // duplicate
	if len(loc.Racks) != 1 {
		t.Fatalf("expected 1 rack after duplicate add, got %d", len(loc.Racks))
	}
}

// --- AddChild ---

// TestAddChildHappyPath verifies AddChild appends a child location UUID to the
// parent's child list.
//
// Why it matters: the child list builds the location hierarchy (building → floor
// → room), so adding a new child must record its ID under the parent.
// Inputs: a parent location and one child UUID. Outputs: the Children slice,
// expected length 1 containing that UUID.
// Data choice: asserting both the length and the stored ID (not just the count)
// confirms the actual child landed in the list rather than a placeholder.
func TestAddChildHappyPath(t *testing.T) {
	loc := CaniLocationType{LocationType: "building"}
	childID := uuid.New()
	loc.AddChild(childID)
	if len(loc.Children) != 1 || loc.Children[0] != childID {
		t.Fatalf("expected Children to contain %v, got %v", childID, loc.Children)
	}
}

// TestAddChildDuplicate verifies AddChild is idempotent and does not append a
// child UUID that is already present.
//
// Why it matters: the child hierarchy is rebuilt at load time and AddChild may be
// called repeatedly, so re-adding the same child must not create a duplicate.
// Inputs: a parent location with the same child UUID added twice. Outputs: the
// Children slice, expected length 1.
// Data choice: adding the identical UUID twice isolates the already-present
// short-circuit; a distinct ID would instead exercise the append path.
func TestAddChildDuplicate(t *testing.T) {
	loc := CaniLocationType{LocationType: "building"}
	childID := uuid.New()
	loc.AddChild(childID)
	loc.AddChild(childID) // duplicate
	if len(loc.Children) != 1 {
		t.Fatalf("expected 1 child after duplicate add, got %d", len(loc.Children))
	}
}

// ========== additional edge-case / branch coverage ==========

// --- Validate (nil) ---

// TestValidateLocationNilReceiver verifies Validate returns an error for a nil
// receiver instead of dereferencing it.
//
// Why it matters: a nil location is a programming error, and Validate gates a
// location before it enters the inventory hierarchy, so it must fail loudly
// rather than panic.
// Inputs: a nil *CaniLocationType. Outputs: an error, non-nil expected.
// Data choice: a nil receiver is the only input that reaches the nil guard ahead
// of the empty-LocationType check, the one branch the value-receiver tests cannot
// hit.
func TestValidateLocationNilReceiver(t *testing.T) {
	var loc *CaniLocationType
	if err := loc.Validate(); err == nil {
		t.Fatal("expected error for nil receiver, got nil")
	}
}

// --- ValidateContentType ---

// TestValidateContentType verifies ValidateContentType permits a content type
// when the receiver is nil or its allow-list is empty, accepts a type present in
// the list, and rejects one that is absent.
//
// Why it matters: locations gate which kinds of objects may be placed in them, so
// the check must enforce a populated allow-list while staying permissive (for
// backwards compatibility) when no list is configured.
// Inputs: a *CaniLocationType (nil, empty-list, or two-entry allow-list) and a
// content-type string. Outputs: an error, nil when permitted and non-nil when
// rejected.
// Data choice: a two-entry allow-list forces the loop to scan past the first
// entry for the match case, and the reject case queries a third distinct type so
// the failure is unambiguous rather than a coincidental near-miss.
func TestValidateContentType(t *testing.T) {
	allowed := &CaniLocationType{
		Name:         "Room-1",
		LocationType: "room",
		ContentTypes: []string{"dcim.device", "dcim.rack"},
	}
	cases := []struct {
		name    string
		loc     *CaniLocationType
		ct      string
		wantErr bool
	}{
		{"nil receiver permits anything", nil, "dcim.device", false},
		{"empty list permits anything", &CaniLocationType{LocationType: "room"}, "dcim.device", false},
		{"content type in allow list", allowed, "dcim.rack", false},
		{"content type not in allow list", allowed, "dcim.cable", true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loc.ValidateContentType(tt.ct)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for content type %q, got nil", tt.ct)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error for content type %q, got %v", tt.ct, err)
			}
		})
	}
}

// --- AddRack (nil) ---

// TestAddRackNilReceiver verifies AddRack is a no-op on a nil receiver and does
// not panic.
//
// Why it matters: rack lists are rebuilt at load time by calling AddRack across
// locations, so a nil location entry must be skipped silently rather than crash
// the load.
// Inputs: a nil *CaniLocationType and a rack UUID. Outputs: none; the test passes
// if the call returns without panicking.
// Data choice: a nil receiver is the only input that reaches the nil guard ahead
// of the append, the branch the value-receiver tests cannot exercise.
func TestAddRackNilReceiver(t *testing.T) {
	var loc *CaniLocationType
	loc.AddRack(uuid.New()) // must not panic
}

// --- AddChild (nil) ---

// TestAddChildNilReceiver verifies AddChild is a no-op on a nil receiver and does
// not panic.
//
// Why it matters: the child hierarchy is rebuilt at load time by calling AddChild
// from parent locations, so a nil parent must be skipped silently rather than
// crash the load.
// Inputs: a nil *CaniLocationType and a child UUID. Outputs: none; the test passes
// if the call returns without panicking.
// Data choice: a nil receiver is the only input that reaches the nil guard ahead
// of the append, the branch the value-receiver tests cannot exercise.
func TestAddChildNilReceiver(t *testing.T) {
	var loc *CaniLocationType
	loc.AddChild(uuid.New()) // must not panic
}
