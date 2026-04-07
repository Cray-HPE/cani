package devicetypes

// Test coverage for CaniFruType methods in cani_fru_types.go.
//
// | Function  | Happy-path test                        | Failure test                              |
// |-----------|----------------------------------------|-------------------------------------------|
// | Validate  | TestFruValidatePassesForValidFru       | TestFruValidateReturnsErrorForNil         |
// | GetID     | TestFruGetIDReturnsUUID                | TestFruGetIDReturnsNilForNilReceiver      |
// | GetSlug   | TestFruGetSlugReturnsSlug              | TestFruGetSlugReturnsEmptyForNil          |
// | GetStatus | TestFruGetStatusReturnsStatus          | TestFruGetStatusReturnsEmptyForNil        |
// | GetVendor | TestFruGetVendorReturnsManufacturer    | TestFruGetVendorReturnsEmptyForNil        |
// | GetType   | TestFruGetTypeReturnsExplicitType      | TestFruGetTypeReturnsEmptyForNil          |

import (
	"testing"

	"github.com/google/uuid"
)

// --- Validate ---

func TestFruValidatePassesForValidFru(t *testing.T) {
	f := &CaniFruType{Name: "fan-module", Slug: "fan-module"}
	if err := f.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFruValidateReturnsErrorForNil(t *testing.T) {
	var f *CaniFruType
	if err := f.Validate(); err == nil {
		t.Fatal("expected error for nil receiver, got nil")
	}
}

// --- GetID ---

func TestFruGetIDReturnsUUID(t *testing.T) {
	id := uuid.New()
	f := &CaniFruType{ID: id}
	if got := f.GetID(); got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

func TestFruGetIDReturnsNilForNilReceiver(t *testing.T) {
	var f *CaniFruType
	if got := f.GetID(); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}

// --- GetSlug ---

func TestFruGetSlugReturnsSlug(t *testing.T) {
	f := &CaniFruType{Slug: "psu-2000w"}
	if got := f.GetSlug(); got != "psu-2000w" {
		t.Fatalf("expected psu-2000w, got %s", got)
	}
}

func TestFruGetSlugReturnsEmptyForNil(t *testing.T) {
	var f *CaniFruType
	if got := f.GetSlug(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetStatus ---

func TestFruGetStatusReturnsStatus(t *testing.T) {
	f := &CaniFruType{ObjectMeta: ObjectMeta{Status: "Active"}}
	if got := f.GetStatus(); got != "Active" {
		t.Fatalf("expected Active, got %s", got)
	}
}

func TestFruGetStatusReturnsEmptyForNil(t *testing.T) {
	var f *CaniFruType
	if got := f.GetStatus(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetVendor ---

func TestFruGetVendorReturnsManufacturer(t *testing.T) {
	f := &CaniFruType{Manufacturer: "Cray"}
	if got := f.GetVendor(); got != "Cray" {
		t.Fatalf("expected Cray, got %s", got)
	}
}

func TestFruGetVendorReturnsEmptyForNil(t *testing.T) {
	var f *CaniFruType
	if got := f.GetVendor(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetType ---

func TestFruGetTypeReturnsExplicitType(t *testing.T) {
	f := &CaniFruType{HardwareType: string(TypeNode)}
	if got := f.GetType(); got != TypeNode {
		t.Fatalf("expected %s, got %s", TypeNode, got)
	}
}

func TestFruGetTypeReturnsEmptyForNil(t *testing.T) {
	var f *CaniFruType
	if got := f.GetType(); got != "" {
		t.Fatalf("expected empty Type, got %s", got)
	}
}
