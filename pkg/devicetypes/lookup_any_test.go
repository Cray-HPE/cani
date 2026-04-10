package devicetypes

// Test coverage for lookup_any.go
//
// | Function   | Happy-path test                | Failure test                 |
// |------------|--------------------------------|------------------------------|
// | LookupAny  | TestLookupAnyBySlug            | TestLookupAnyEmptyKey        |
// | LookupAny  | TestLookupAnyByPartNumber      | TestLookupAnyNoMatch         |

import (
	"strings"
	"testing"
)

// ---------- LookupAny ----------

func TestLookupAnyBySlug(t *testing.T) {
	// Register a temporary rack type so the lookup has something to find.
	RegisterRackType(CaniRackType{
		Slug:         "test-lookup-any-rack",
		Model:        "Test Rack",
		Manufacturer: "TestCo",
		PartNumber:   "TLAR-001",
	})
	defer func() { delete(allRackTypes, "test-lookup-any-rack") }()

	result, err := LookupAny("test-lookup-any-rack")
	if err != nil {
		t.Fatalf("LookupAny() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Category != CategoryRack {
		t.Errorf("Category = %q, want %q", result.Category, CategoryRack)
	}
	if result.Rack == nil {
		t.Error("expected Rack field to be non-nil")
	}
}

func TestLookupAnyByPartNumber(t *testing.T) {
	RegisterDeviceType(CaniDeviceType{
		Slug:         "test-lookup-any-device",
		Model:        "Test Device",
		Manufacturer: "TestCo",
		PartNumber:   "TLAD-002",
	})
	defer func() { delete(allDeviceTypes, "test-lookup-any-device") }()

	result, err := LookupAny("TLAD-002")
	if err != nil {
		t.Fatalf("LookupAny() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Category != CategoryDevice {
		t.Errorf("Category = %q, want %q", result.Category, CategoryDevice)
	}
	if result.Device == nil {
		t.Error("expected Device field to be non-nil")
	}
}

func TestLookupAnyEmptyKey(t *testing.T) {
	_, err := LookupAny("")
	if err == nil {
		t.Fatal("LookupAny(\"\") should return an error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("error = %q, want message containing 'required'", err.Error())
	}
}

func TestLookupAnyNoMatch(t *testing.T) {
	_, err := LookupAny("absolutely-nonexistent-slug-zzz-999")
	if err == nil {
		t.Fatal("LookupAny(nonexistent) should return an error")
	}
	if !strings.Contains(err.Error(), "no hardware type found") {
		t.Errorf("error = %q, want message containing 'no hardware type found'", err.Error())
	}
}
