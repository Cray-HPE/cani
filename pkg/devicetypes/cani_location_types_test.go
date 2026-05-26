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

func TestGetIDLocationHappyPath(t *testing.T) {
	id := uuid.New()
	loc := CaniLocationType{ID: id, LocationType: "site"}
	if got := loc.GetID(); got != id {
		t.Fatalf("expected %v, got %v", id, got)
	}
}

func TestGetIDLocationNilReceiver(t *testing.T) {
	var loc *CaniLocationType
	if got := loc.GetID(); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %v", got)
	}
}

// --- GetSlug ---

func TestGetSlugLocationHappyPath(t *testing.T) {
	loc := CaniLocationType{LocationType: "room"}
	if got := loc.GetSlug(); got != "room" {
		t.Fatalf("expected %q, got %q", "room", got)
	}
}

func TestGetSlugLocationNilReceiver(t *testing.T) {
	var loc *CaniLocationType
	if got := loc.GetSlug(); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

// --- GetStatus ---

func TestGetStatusLocationHappyPath(t *testing.T) {
	loc := CaniLocationType{LocationType: "floor", ObjectMeta: ObjectMeta{Status: "Active"}}
	if got := loc.GetStatus(); got != "Active" {
		t.Fatalf("expected %q, got %q", "Active", got)
	}
}

func TestGetStatusLocationNilReceiver(t *testing.T) {
	var loc *CaniLocationType
	if got := loc.GetStatus(); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

// --- AddRack ---

func TestAddRackHappyPath(t *testing.T) {
	loc := CaniLocationType{LocationType: "room"}
	rackID := uuid.New()
	loc.AddRack(rackID)
	if len(loc.Racks) != 1 || loc.Racks[0] != rackID {
		t.Fatalf("expected Racks to contain %v, got %v", rackID, loc.Racks)
	}
}

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

func TestAddChildHappyPath(t *testing.T) {
	loc := CaniLocationType{LocationType: "building"}
	childID := uuid.New()
	loc.AddChild(childID)
	if len(loc.Children) != 1 || loc.Children[0] != childID {
		t.Fatalf("expected Children to contain %v, got %v", childID, loc.Children)
	}
}

func TestAddChildDuplicate(t *testing.T) {
	loc := CaniLocationType{LocationType: "building"}
	childID := uuid.New()
	loc.AddChild(childID)
	loc.AddChild(childID) // duplicate
	if len(loc.Children) != 1 {
		t.Fatalf("expected 1 child after duplicate add, got %d", len(loc.Children))
	}
}
