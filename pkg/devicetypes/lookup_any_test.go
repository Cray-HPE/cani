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

// TestLookupAnyModuleAndCable verifies LookupAny resolves module and cable
// registry entries, the two categories its existing slug/part-number tests omit.
//
// Why it matters: LookupAny is the single entry point used by callers that do
// not know a key's category, so every registry (rack, device, module, cable)
// must be reachable or some hardware becomes invisible to lookups.
// Inputs: a freshly registered module slug and cable slug, each looked up by
// slug. Outputs: results categorized as Module and Cable with the matching
// pointer populated. Data choice: distinct fake slugs avoid colliding with the
// embedded library, and cleanup deletes exactly those two entries.
func TestLookupAnyModuleAndCable(t *testing.T) {
	RegisterModuleType(CaniModuleType{
		Slug:         "test-lookup-any-module",
		Model:        "Test Module",
		Manufacturer: "TestCo",
		PartNumber:   "TLAM-003",
	})
	RegisterCableType(CaniCableType{
		Slug:         "test-lookup-any-cable",
		Model:        "Test Cable",
		Manufacturer: "TestCo",
		PartNumber:   "TLAC-004",
	})
	t.Cleanup(func() {
		delete(allModuleTypes, "test-lookup-any-module")
		delete(allCableTypes, "test-lookup-any-cable")
	})

	modResult, err := LookupAny("test-lookup-any-module")
	if err != nil {
		t.Fatalf("LookupAny(module) unexpected error: %v", err)
	}
	if modResult.Category != CategoryModule || modResult.Module == nil {
		t.Errorf("module lookup = (%q, %v), want Module category with non-nil Module", modResult.Category, modResult.Module)
	}

	cableResult, err := LookupAny("test-lookup-any-cable")
	if err != nil {
		t.Fatalf("LookupAny(cable) unexpected error: %v", err)
	}
	if cableResult.Category != CategoryCable || cableResult.Cable == nil {
		t.Errorf("cable lookup = (%q, %v), want Cable category with non-nil Cable", cableResult.Category, cableResult.Cable)
	}
}

// TestLookupAnyByPartNumberAllKinds verifies LookupAny resolves rack, module,
// and cable registry entries by part number, not just by slug.
//
// Why it matters: hardware is frequently referenced by manufacturer part number,
// so LookupAny must consult each registry's part-number index or those lookups
// silently fail and callers cannot find known hardware.
// Inputs: a rack, module, and cable each registered with a distinct part number,
// then looked up by that number. Outputs: results categorized as Rack, Module,
// and Cable respectively. Data choice: fake part numbers avoid colliding with
// the embedded library, and cleanup clears both the slug and part-number maps
// that Register*Type populates.
func TestLookupAnyByPartNumberAllKinds(t *testing.T) {
	RegisterRackType(CaniRackType{Slug: "tla-pn-rack", PartNumber: "PN-RACK-1", Model: "R", Manufacturer: "TestCo"})
	RegisterModuleType(CaniModuleType{Slug: "tla-pn-mod", PartNumber: "PN-MOD-1", Model: "M", Manufacturer: "TestCo"})
	RegisterCableType(CaniCableType{Slug: "tla-pn-cable", PartNumber: "PN-CABLE-1", Model: "C", Manufacturer: "TestCo"})
	t.Cleanup(func() {
		delete(allRackTypes, "tla-pn-rack")
		delete(rackTypesByPartNum, "PN-RACK-1")
		delete(allModuleTypes, "tla-pn-mod")
		delete(moduleTypesByPartNum, "PN-MOD-1")
		delete(allCableTypes, "tla-pn-cable")
		delete(cableTypesByPartNum, "PN-CABLE-1")
	})

	rackRes, err := LookupAny("PN-RACK-1")
	if err != nil || rackRes.Category != CategoryRack {
		t.Errorf("LookupAny(rack PN) = (%v, %v), want Rack category", rackRes, err)
	}
	modRes, err := LookupAny("PN-MOD-1")
	if err != nil || modRes.Category != CategoryModule {
		t.Errorf("LookupAny(module PN) = (%v, %v), want Module category", modRes, err)
	}
	cableRes, err := LookupAny("PN-CABLE-1")
	if err != nil || cableRes.Category != CategoryCable {
		t.Errorf("LookupAny(cable PN) = (%v, %v), want Cable category", cableRes, err)
	}
}
