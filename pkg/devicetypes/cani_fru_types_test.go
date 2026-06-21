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

// TestFruValidatePassesForValidFru verifies Validate returns nil for a populated
// FRU.
//
// Why it matters: Validate gates a FRU (a field-replaceable spare) before it
// enters inventory, so a well-formed FRU must pass.
// Inputs: a FRU with a name and slug. Outputs: an error, nil expected.
// Data choice: a named "fan-module" FRU is a minimal valid instance; FRU
// validation rejects only a nil receiver, so any non-nil value exercises the
// success path.
func TestFruValidatePassesForValidFru(t *testing.T) {
	f := &CaniFruType{Name: "fan-module", Slug: "fan-module"}
	if err := f.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestFruValidateReturnsErrorForNil verifies Validate returns an error for a nil
// receiver.
//
// Why it matters: a nil FRU is a programming error, and Validate must fail
// loudly rather than dereference nil and panic.
// Inputs: a nil *CaniFruType. Outputs: an error, non-nil expected.
// Data choice: a nil receiver is the only input that reaches the nil guard, the
// sole branch FRU validation can fail on.
func TestFruValidateReturnsErrorForNil(t *testing.T) {
	var f *CaniFruType
	if err := f.Validate(); err == nil {
		t.Fatal("expected error for nil receiver, got nil")
	}
}

// --- GetID ---

// TestFruGetIDReturnsUUID verifies GetID returns the FRU's stored UUID.
//
// Why it matters: the ID uniquely identifies a FRU across inventory and export,
// so the accessor must return it unchanged.
// Inputs: a FRU with a generated ID. Outputs: the UUID, expected equal to the
// stored value.
// Data choice: a freshly generated uuid is an arbitrary distinct value, proving
// the getter returns the stored field rather than a constant.
func TestFruGetIDReturnsUUID(t *testing.T) {
	id := uuid.New()
	f := &CaniFruType{ID: id}
	if got := f.GetID(); got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

// TestFruGetIDReturnsNilForNilReceiver verifies GetID returns uuid.Nil for a nil
// receiver.
//
// Why it matters: callers treat uuid.Nil as the sentinel for "no FRU", so a nil
// receiver must yield it rather than panic.
// Inputs: a nil *CaniFruType. Outputs: the UUID, expected uuid.Nil.
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestFruGetIDReturnsNilForNilReceiver(t *testing.T) {
	var f *CaniFruType
	if got := f.GetID(); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}

// --- GetSlug ---

// TestFruGetSlugReturnsSlug verifies GetSlug returns the FRU's slug.
//
// Why it matters: the slug links a FRU to its hardware-library template, so the
// accessor must surface it unchanged for lookups and export.
// Inputs: a FRU with slug "psu-2000w". Outputs: the slug string, expected
// "psu-2000w".
// Data choice: a realistic PSU slug is an arbitrary non-empty value showing the
// stored field is returned verbatim.
func TestFruGetSlugReturnsSlug(t *testing.T) {
	f := &CaniFruType{Slug: "psu-2000w"}
	if got := f.GetSlug(); got != "psu-2000w" {
		t.Fatalf("expected psu-2000w, got %s", got)
	}
}

// TestFruGetSlugReturnsEmptyForNil verifies GetSlug returns "" for a nil
// receiver.
//
// Why it matters: a nil FRU must degrade to "" rather than panic when its slug
// is read.
// Inputs: a nil *CaniFruType. Outputs: the slug string, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestFruGetSlugReturnsEmptyForNil(t *testing.T) {
	var f *CaniFruType
	if got := f.GetSlug(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetStatus ---

// TestFruGetStatusReturnsStatus verifies GetStatus returns the FRU's status from
// its embedded ObjectMeta.
//
// Why it matters: status drives lifecycle handling and export filtering, so the
// accessor must surface the stored value.
// Inputs: a FRU with Status "Active". Outputs: the status string, expected
// "Active".
// Data choice: "Active" is the common in-service status, a representative
// non-empty value proving the embedded field is returned.
func TestFruGetStatusReturnsStatus(t *testing.T) {
	f := &CaniFruType{ObjectMeta: ObjectMeta{Status: "Active"}}
	if got := f.GetStatus(); got != "Active" {
		t.Fatalf("expected Active, got %s", got)
	}
}

// TestFruGetStatusReturnsEmptyForNil verifies GetStatus returns "" for a nil
// receiver.
//
// Why it matters: a nil FRU must report no status rather than panic when its
// lifecycle state is queried.
// Inputs: a nil *CaniFruType. Outputs: the status string, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestFruGetStatusReturnsEmptyForNil(t *testing.T) {
	var f *CaniFruType
	if got := f.GetStatus(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetVendor ---

// TestFruGetVendorReturnsManufacturer verifies GetVendor returns the FRU's
// Manufacturer through the generic vendor accessor.
//
// Why it matters: vendor identifies the hardware maker for classification and
// export, and GetVendor maps the Manufacturer field onto that shared accessor.
// Inputs: a FRU with Manufacturer "Cray". Outputs: the vendor string, expected
// "Cray".
// Data choice: "Cray" is a recognizable maker that demonstrates the
// Manufacturer-to-vendor mapping returns the stored value.
func TestFruGetVendorReturnsManufacturer(t *testing.T) {
	f := &CaniFruType{Manufacturer: "Cray"}
	if got := f.GetVendor(); got != "Cray" {
		t.Fatalf("expected Cray, got %s", got)
	}
}

// TestFruGetVendorReturnsEmptyForNil verifies GetVendor returns "" for a nil
// receiver.
//
// Why it matters: a nil FRU must report no vendor rather than panic when its
// maker is queried.
// Inputs: a nil *CaniFruType. Outputs: the vendor string, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestFruGetVendorReturnsEmptyForNil(t *testing.T) {
	var f *CaniFruType
	if got := f.GetVendor(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetType ---

// TestFruGetTypeReturnsExplicitType verifies GetType returns an explicitly set
// Type rather than the default.
//
// Why it matters: an operator-set type must override the generic fallback so the
// FRU classifies correctly during export.
// Inputs: a FRU with Type TypeNode. Outputs: the Type, expected TypeNode.
// Data choice: TypeNode differs from the TypeFru default, so a passing assertion
// proves the explicit value is preferred over the fallback.
func TestFruGetTypeReturnsExplicitType(t *testing.T) {
	f := &CaniFruType{Type: TypeNode}
	if got := f.GetType(); got != TypeNode {
		t.Fatalf("expected %s, got %s", TypeNode, got)
	}
}

// TestFruGetTypeReturnsDefaultForEmpty verifies GetType returns the TypeFru
// default when Type is unset on a non-nil FRU.
//
// Why it matters: a FRU without an explicit type must classify as a generic FRU
// rather than an empty type.
// Inputs: a FRU with an empty Type. Outputs: the Type, expected TypeFru.
// Data choice: an empty (non-nil) struct isolates the unset-Type fallback
// branch, distinct from the nil-receiver case that returns "".
func TestFruGetTypeReturnsDefaultForEmpty(t *testing.T) {
	f := &CaniFruType{}
	if got := f.GetType(); got != TypeFru {
		t.Fatalf("expected %s, got %s", TypeFru, got)
	}
}

// TestFruGetTypeReturnsEmptyForNil verifies GetType returns the empty Type for a
// nil receiver instead of falling through to the TypeFru default.
//
// Why it matters: FRU type drives classification and export, and a nil FRU must
// report "unknown" rather than be silently treated as a generic FRU.
// Inputs: a nil *CaniFruType. Outputs: the Type result, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard ahead
// of the TypeFru fallback, the one branch the empty-value test cannot hit.
func TestFruGetTypeReturnsEmptyForNil(t *testing.T) {
	var f *CaniFruType
	if got := f.GetType(); got != "" {
		t.Fatalf("expected empty Type, got %s", got)
	}
}
