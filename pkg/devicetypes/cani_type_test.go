package devicetypes

// Test coverage for CaniType interface methods in cani_type.go.
//
// | Function  | Happy-path test                              | Failure test                                    |
// |-----------|----------------------------------------------|-------------------------------------------------|
// | Validate  | TestCaniTypeValidatePassesForValidInstance    | TestCaniTypeValidateReturnsErrorForNilReceiver  |
// | GetID     | TestCaniTypeGetIDReturnsExpectedUUID          | TestCaniTypeGetIDReturnsNilForNilReceiver       |
// | GetSlug   | TestCaniTypeGetSlugReturnsExpectedSlug        | TestCaniTypeGetSlugReturnsEmptyForNilReceiver   |
// | GetStatus | TestCaniTypeGetStatusReturnsExpectedStatus    | TestCaniTypeGetStatusReturnsEmptyForNilReceiver |

import (
	"testing"

	"github.com/google/uuid"
)

// --- Validate ---

func TestCaniTypeValidatePassesForValidInstance(t *testing.T) {
	id := uuid.New()
	var ct CaniType = &CaniDeviceType{ID: id, ObjectMeta: ObjectMeta{Status: "Staged"}}
	if err := ct.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCaniTypeValidateReturnsErrorForNilReceiver(t *testing.T) {
	var ct CaniType = (*CaniDeviceType)(nil)
	if err := ct.Validate(); err == nil {
		t.Fatal("expected error for nil receiver, got nil")
	}
}

// --- GetID ---

func TestCaniTypeGetIDReturnsExpectedUUID(t *testing.T) {
	id := uuid.New()
	var ct CaniType = &CaniDeviceType{ID: id}
	if got := ct.GetID(); got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

func TestCaniTypeGetIDReturnsNilForNilReceiver(t *testing.T) {
	var ct CaniType = (*CaniDeviceType)(nil)
	if got := ct.GetID(); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}

// --- GetSlug ---

func TestCaniTypeGetSlugReturnsExpectedSlug(t *testing.T) {
	var ct CaniType = &CaniDeviceType{Slug: "hpe-cray-ex235n"}
	if got := ct.GetSlug(); got != "hpe-cray-ex235n" {
		t.Fatalf("expected hpe-cray-ex235n, got %s", got)
	}
}

func TestCaniTypeGetSlugReturnsEmptyForNilReceiver(t *testing.T) {
	var ct CaniType = (*CaniDeviceType)(nil)
	if got := ct.GetSlug(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetStatus ---

func TestCaniTypeGetStatusReturnsExpectedStatus(t *testing.T) {
	var ct CaniType = &CaniDeviceType{ObjectMeta: ObjectMeta{Status: "Active"}}
	if got := ct.GetStatus(); got != "Active" {
		t.Fatalf("expected Active, got %s", got)
	}
}

func TestCaniTypeGetStatusReturnsEmptyForNilReceiver(t *testing.T) {
	var ct CaniType = (*CaniDeviceType)(nil)
	if got := ct.GetStatus(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}
