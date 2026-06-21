package devicetypes

// Test coverage for CaniModuleType methods in cani_module_types.go.
//
// | Function              | Happy-path test                                  | Failure test                                      |
// |-----------------------|--------------------------------------------------|----------------------------------------------------|
// | Validate              | TestModuleValidatePassesForValid                 | TestModuleValidateReturnsErrorForNil               |
// | GetID                 | TestModuleGetIDReturnsUUID                       | TestModuleGetIDReturnsNilForNilReceiver            |
// | GetSlug               | TestModuleGetSlugReturnsSlug                     | TestModuleGetSlugReturnsEmptyForNil                |
// | GetStatus             | TestModuleGetStatusReturnsStatus                 | TestModuleGetStatusReturnsEmptyForNil              |
// | GetVendor             | TestModuleGetVendorReturnsManufacturer            | TestModuleGetVendorReturnsEmptyForNil              |
// | GetType               | TestModuleGetTypeReturnsExplicitType             | TestModuleGetTypeReturnsDefaultForEmpty            |
// | InstantiateInterfaces | TestModuleInstantiateInterfacesCreatesInstances   | TestModuleInstantiateInterfacesReturnsNilForNil    |
// | GetInterface          | TestModuleGetInterfaceFindsMatch                 | TestModuleGetInterfaceReturnsNilForMissing         |
// | GetInterfaceByName    | TestModuleGetInterfaceByNameFindsMatch            | TestModuleGetInterfaceByNameReturnsNilForNil       |

import (
	"testing"

	"github.com/google/uuid"
)

// --- Validate ---

// TestModuleValidatePassesForValid verifies Validate returns nil for a module
// whose slug is registered in the hardware library.
//
// Why it matters: Validate gates a module before it enters inventory, so a slug
// the library recognizes must pass and let the module be added.
// Inputs: a module with a registered slug. Outputs: an error, nil expected.
// Data choice: the slug is registered first (and removed via t.Cleanup) so the
// library lookup succeeds, isolating the valid-slug branch.
func TestModuleValidatePassesForValid(t *testing.T) {
	const slug = "test-valid-module-slug"
	RegisterModuleType(CaniModuleType{Slug: slug})
	t.Cleanup(func() {
		delete(allModuleTypes, slug)
	})

	m := &CaniModuleType{Name: "gpu-a100", Slug: slug}
	if err := m.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestModuleValidateReturnsErrorForUnknownSlug verifies Validate returns an error
// when the module's slug is absent from the library.
//
// Why it matters: Validate keeps unknown hardware out of inventory, so a slug the
// library does not recognize must be rejected.
// Inputs: a module with an unregistered slug. Outputs: an error, non-nil expected.
// Data choice: a deliberately bogus slug guarantees the library lookup misses,
// exercising the not-found branch.
func TestModuleValidateReturnsErrorForUnknownSlug(t *testing.T) {
	m := &CaniModuleType{Name: "gpu-a100", Slug: "definitely-not-a-real-module-slug"}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for unknown module slug")
	}
}

// TestModuleValidateAllowsBlankSlug verifies Validate accepts a module that has
// no slug set.
//
// Why it matters: custom or ad-hoc modules without a library slug are legal, so a
// blank slug must skip the library check instead of failing.
// Inputs: a name-only module with an empty Slug. Outputs: an error, nil expected.
// Data choice: a module with a name but no slug isolates the blank-slug skip path,
// the branch that bypasses the library lookup entirely.
func TestModuleValidateAllowsBlankSlug(t *testing.T) {
	m := &CaniModuleType{Name: "custom-module"}
	if err := m.Validate(); err != nil {
		t.Fatalf("expected blank slug to remain valid, got %v", err)
	}
}

// TestModuleValidateReturnsErrorForNil verifies Validate returns an error for a
// nil receiver.
//
// Why it matters: a nil module is a programming error, and Validate must fail
// loudly rather than dereference nil and panic.
// Inputs: a nil *CaniModuleType. Outputs: an error, non-nil expected.
// Data choice: a nil receiver is the only input that reaches the nil guard ahead
// of the slug check.
func TestModuleValidateReturnsErrorForNil(t *testing.T) {
	var m *CaniModuleType
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for nil receiver")
	}
}

// --- GetID ---

// TestModuleGetIDReturnsUUID verifies GetID returns the module's stored UUID.
//
// Why it matters: the ID uniquely identifies a module across inventory and
// export, so the accessor must return it unchanged.
// Inputs: a module with a generated ID. Outputs: the UUID, expected equal to the
// stored value.
// Data choice: a freshly generated uuid is an arbitrary distinct value, proving
// the getter returns the stored field verbatim rather than a constant.
func TestModuleGetIDReturnsUUID(t *testing.T) {
	id := uuid.New()
	m := &CaniModuleType{ID: id}
	if got := m.GetID(); got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

// TestModuleGetIDReturnsNilForNilReceiver verifies GetID returns uuid.Nil for a
// nil receiver.
//
// Why it matters: callers treat uuid.Nil as the sentinel for "no module", so a
// nil receiver must yield it rather than panic.
// Inputs: a nil *CaniModuleType. Outputs: the UUID, expected uuid.Nil.
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestModuleGetIDReturnsNilForNilReceiver(t *testing.T) {
	var m *CaniModuleType
	if got := m.GetID(); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}

// --- GetSlug ---

// TestModuleGetSlugReturnsSlug verifies GetSlug returns the module's slug.
//
// Why it matters: the slug links a module to its library template, so the
// accessor must surface it unchanged for lookups and export.
// Inputs: a module with slug "nvidia-a100". Outputs: the slug string.
// Data choice: a realistic GPU slug is an arbitrary non-empty value showing the
// stored field is returned verbatim.
func TestModuleGetSlugReturnsSlug(t *testing.T) {
	m := &CaniModuleType{Slug: "nvidia-a100"}
	if got := m.GetSlug(); got != "nvidia-a100" {
		t.Fatalf("expected nvidia-a100, got %s", got)
	}
}

// TestModuleGetSlugReturnsEmptyForNil verifies GetSlug returns the empty string
// for a nil receiver.
//
// Why it matters: a nil module must degrade to "" rather than panic when callers
// read its slug.
// Inputs: a nil *CaniModuleType. Outputs: the slug string, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestModuleGetSlugReturnsEmptyForNil(t *testing.T) {
	var m *CaniModuleType
	if got := m.GetSlug(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetStatus ---

// TestModuleGetStatusReturnsStatus verifies GetStatus returns the module's status
// string from its embedded ObjectMeta.
//
// Why it matters: status drives lifecycle handling and export filtering, so the
// accessor must surface the stored value.
// Inputs: a module with Status "Active". Outputs: the status string.
// Data choice: "Active" is the common in-service status, a representative
// non-empty value proving the embedded field is returned.
func TestModuleGetStatusReturnsStatus(t *testing.T) {
	m := &CaniModuleType{ObjectMeta: ObjectMeta{Status: "Active"}}
	if got := m.GetStatus(); got != "Active" {
		t.Fatalf("expected Active, got %s", got)
	}
}

// TestModuleGetStatusReturnsEmptyForNil verifies GetStatus returns "" for a nil
// receiver.
//
// Why it matters: a nil module must report no status rather than panic when its
// lifecycle state is queried.
// Inputs: a nil *CaniModuleType. Outputs: the status string, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestModuleGetStatusReturnsEmptyForNil(t *testing.T) {
	var m *CaniModuleType
	if got := m.GetStatus(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetVendor ---

// TestModuleGetVendorReturnsManufacturer verifies GetVendor returns the module's
// Manufacturer through the generic vendor accessor.
//
// Why it matters: vendor identifies the hardware maker for classification and
// export, and GetVendor maps the Manufacturer field onto that shared accessor.
// Inputs: a module with Manufacturer "NVIDIA". Outputs: the vendor string.
// Data choice: "NVIDIA" is a recognizable maker that demonstrates the
// Manufacturer-to-vendor mapping returns the stored value.
func TestModuleGetVendorReturnsManufacturer(t *testing.T) {
	m := &CaniModuleType{Manufacturer: "NVIDIA"}
	if got := m.GetVendor(); got != "NVIDIA" {
		t.Fatalf("expected NVIDIA, got %s", got)
	}
}

// TestModuleGetVendorReturnsEmptyForNil verifies GetVendor returns "" for a nil
// receiver.
//
// Why it matters: a nil module must report no vendor rather than panic when its
// maker is queried.
// Inputs: a nil *CaniModuleType. Outputs: the vendor string, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestModuleGetVendorReturnsEmptyForNil(t *testing.T) {
	var m *CaniModuleType
	if got := m.GetVendor(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetType ---

// TestModuleGetTypeReturnsExplicitType verifies GetType returns an explicitly set
// Type rather than the default.
//
// Why it matters: an operator-set type must override the generic fallback so the
// module classifies correctly during export.
// Inputs: a module with Type TypeChassis. Outputs: the Type, expected
// TypeChassis.
// Data choice: TypeChassis differs from the TypeModule default, so a passing
// assertion proves the explicit value is preferred over the fallback.
func TestModuleGetTypeReturnsExplicitType(t *testing.T) {
	m := &CaniModuleType{Type: TypeChassis}
	if got := m.GetType(); got != TypeChassis {
		t.Fatalf("expected %s, got %s", TypeChassis, got)
	}
}

// TestModuleGetTypeReturnsDefaultForEmpty verifies GetType returns the TypeModule
// default when Type is unset on a non-nil module.
//
// Why it matters: a module without an explicit type must classify as a generic
// module rather than an empty type.
// Inputs: a module with an empty Type. Outputs: the Type, expected TypeModule.
// Data choice: an empty (non-nil) struct isolates the unset-Type fallback branch,
// distinct from the nil-receiver case that returns "".
func TestModuleGetTypeReturnsDefaultForEmpty(t *testing.T) {
	m := &CaniModuleType{}
	if got := m.GetType(); got != TypeModule {
		t.Fatalf("expected %s, got %s", TypeModule, got)
	}
}

// --- InstantiateInterfaces ---

// TestModuleInstantiateInterfacesCreatesInstances verifies InstantiateInterfaces
// builds one CaniInterface per spec, preserving names and the MgmtOnly flag.
//
// Why it matters: instantiating a module's interfaces turns its declared ports
// into concrete inventory records, so each spec must map to an instance with the
// right name and management-only flag.
// Inputs: a module with specs eth0 (MgmtOnly true) and eth1 (unset). Outputs: the
// instance slice length, names, and MgmtOnly flags.
// Data choice: one spec with MgmtOnly set and one without exercises both the
// explicit-true and default-false paths of the pointer flag in a single pass.
func TestModuleInstantiateInterfacesCreatesInstances(t *testing.T) {
	mgmt := true
	m := &CaniModuleType{
		ID: uuid.New(),
		Interfaces: []InterfaceSpec{
			{Name: "eth0", Type: InterfacesElemTypeA10GbaseT, MgmtOnly: &mgmt},
			{Name: "eth1", Type: InterfacesElemTypeA25GbaseXSfp28},
		},
	}
	instances := m.InstantiateInterfaces()
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
	if instances[0].Name != "eth0" {
		t.Fatalf("expected eth0, got %s", instances[0].Name)
	}
	if !instances[0].MgmtOnly {
		t.Fatal("expected MgmtOnly true for eth0")
	}
	if instances[1].MgmtOnly {
		t.Fatal("expected MgmtOnly false for eth1")
	}
}

// TestModuleInstantiateInterfacesReturnsNilForNil verifies InstantiateInterfaces
// returns nil for a nil receiver.
//
// Why it matters: a nil module has no interfaces to instantiate, so the method
// must return nil rather than panic.
// Inputs: a nil *CaniModuleType. Outputs: the instance slice, expected nil.
// Data choice: a nil receiver is the only input that reaches the nil guard,
// distinct from the empty-slice case a module with no specs produces.
func TestModuleInstantiateInterfacesReturnsNilForNil(t *testing.T) {
	var m *CaniModuleType
	if got := m.InstantiateInterfaces(); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

// --- GetInterface ---

// TestModuleGetInterfaceFindsMatch verifies GetInterface returns the spec whose
// name matches the query.
//
// Why it matters: interface lookup by name is how cabling and config resolve a
// module's ports, so a present name must return its spec.
// Inputs: a module with eth0/eth1 specs and the query "eth1". Outputs: the
// *InterfaceSpec, expected non-nil with Name "eth1".
// Data choice: querying the second of two specs proves the search scans past the
// first entry to find the match.
func TestModuleGetInterfaceFindsMatch(t *testing.T) {
	m := &CaniModuleType{
		Interfaces: []InterfaceSpec{
			{Name: "eth0", Type: InterfacesElemTypeA10GbaseT},
			{Name: "eth1", Type: InterfacesElemTypeA25GbaseXSfp28},
		},
	}
	iface := m.GetInterface("eth1")
	if iface == nil {
		t.Fatal("expected non-nil interface")
	}
	if iface.Name != "eth1" {
		t.Fatalf("expected eth1, got %s", iface.Name)
	}
}

// TestModuleGetInterfaceReturnsNilForMissing verifies GetInterface returns nil
// when no spec name matches the query.
//
// Why it matters: a lookup for an absent port must signal "not found" with nil
// rather than return a wrong spec.
// Inputs: a module with one eth0 spec and the query "eth99". Outputs: the
// *InterfaceSpec, expected nil.
// Data choice: a name absent from the single-spec list guarantees the loop
// exhausts every entry without a match.
func TestModuleGetInterfaceReturnsNilForMissing(t *testing.T) {
	m := &CaniModuleType{
		Interfaces: []InterfaceSpec{
			{Name: "eth0", Type: InterfacesElemTypeA10GbaseT},
		},
	}
	if got := m.GetInterface("eth99"); got != nil {
		t.Fatalf("expected nil for missing interface, got %v", got)
	}
}

// --- GetInterfaceByName ---

// TestModuleGetInterfaceByNameFindsMatch verifies GetInterfaceByName returns the
// spec matching the name, delegating to GetInterface.
//
// Why it matters: the alias gives callers a name-explicit lookup, and it must
// resolve to the same spec (including its type) as the underlying method.
// Inputs: a module with a mgmt0 spec and the query "mgmt0". Outputs: the
// *InterfaceSpec, expected non-nil with Type A1000BaseT.
// Data choice: asserting the returned Type, not just non-nil, confirms the alias
// yields the actual spec rather than a placeholder.
func TestModuleGetInterfaceByNameFindsMatch(t *testing.T) {
	m := &CaniModuleType{
		Interfaces: []InterfaceSpec{
			{Name: "mgmt0", Type: InterfacesElemTypeA1000BaseT},
		},
	}
	iface := m.GetInterfaceByName("mgmt0")
	if iface == nil {
		t.Fatal("expected non-nil interface")
	}
	if iface.Type != InterfacesElemTypeA1000BaseT {
		t.Fatalf("expected %s, got %s", InterfacesElemTypeA1000BaseT, iface.Type)
	}
}

// TestModuleGetInterfaceByNameReturnsNilForNil verifies GetInterfaceByName returns
// nil for a nil receiver.
//
// Why it matters: the alias must inherit GetInterface's nil safety so a missing
// module never panics during a port lookup.
// Inputs: a nil *CaniModuleType and the query "eth0". Outputs: the
// *InterfaceSpec, expected nil.
// Data choice: a nil receiver routes through the alias to GetInterface's nil
// guard, the only input that reaches it via this entry point.
func TestModuleGetInterfaceByNameReturnsNilForNil(t *testing.T) {
	var m *CaniModuleType
	if got := m.GetInterfaceByName("eth0"); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

// ========== additional edge-case / branch coverage ==========

// --- GetType (nil) ---

// TestModuleGetTypeReturnsEmptyForNil verifies GetType returns the empty Type for
// a nil receiver instead of falling through to the TypeModule default.
//
// Why it matters: module type drives classification and export, and a nil module
// (e.g. a missing map entry) must report "unknown" rather than be silently
// treated as a generic module, which would mask the missing data.
// Inputs: a nil *CaniModuleType. Outputs: the Type result, expected "".
// Data choice: a nil receiver is the only input that reaches the `m == nil`
// guard ahead of the TypeModule fallback, the one branch the empty-value test
// (which returns TypeModule) cannot hit.
func TestModuleGetTypeReturnsEmptyForNil(t *testing.T) {
	var m *CaniModuleType
	if got := m.GetType(); got != "" {
		t.Fatalf("expected empty Type for nil receiver, got %s", got)
	}
}

// --- InstantiateInterfaces (contents) ---

// TestModuleInstantiateInterfacesStampsIdentityAndRole verifies each instantiated
// interface carries the parent module ID, an Active status, a fresh ID, and the
// spec's explicit role.
//
// Why it matters: instantiating a module's interfaces is how its ports become
// concrete inventory records, so each instance must be linked to its module
// (DeviceID), be independently addressable (unique ID), default to Active, and
// honor an operator-declared role.
// Inputs: a module with one spec eth0 carrying Role="custom-role". Outputs: the
// instance's DeviceID, Status, ID, and Role.
// Data choice: a single spec with an explicit role isolates these identity/content
// assertions, which the existing creates-instances test (names + MgmtOnly only)
// leaves unchecked; the explicit role also drives ResolveInterfaceRole's
// short-circuit branch.
func TestModuleInstantiateInterfacesStampsIdentityAndRole(t *testing.T) {
	moduleID := uuid.New()
	m := &CaniModuleType{
		ID: moduleID,
		Interfaces: []InterfaceSpec{
			{Name: "eth0", Type: InterfacesElemTypeA10GbaseT, Role: "custom-role"},
		},
	}
	got := m.InstantiateInterfaces()
	if len(got) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(got))
	}
	if got[0].DeviceID != moduleID {
		t.Errorf("expected DeviceID %s, got %s", moduleID, got[0].DeviceID)
	}
	if got[0].Status != string(StatusActive) {
		t.Errorf("expected status %q, got %q", StatusActive, got[0].Status)
	}
	if got[0].ID == uuid.Nil {
		t.Error("expected a freshly assigned interface ID, got uuid.Nil")
	}
	if got[0].Role != "custom-role" {
		t.Errorf("expected explicit role 'custom-role', got %q", got[0].Role)
	}
}

// TestModuleInstantiateInterfacesEmptyReturnsNonNilEmpty verifies a module with no
// interface specs returns a non-nil, empty slice.
//
// Why it matters: callers append to and range over the result, so a module that
// simply has zero interfaces must yield an empty slice (distinct from the nil a
// nil receiver returns) to keep that contract unambiguous.
// Inputs: a non-nil module with an empty Interfaces slice. Outputs: the returned
// slice, expected non-nil with length 0.
// Data choice: an empty (not nil) Interfaces field separates the "no specs" case
// from the nil-receiver case the existing nil test already covers.
func TestModuleInstantiateInterfacesEmptyReturnsNonNilEmpty(t *testing.T) {
	m := &CaniModuleType{ID: uuid.New()}
	got := m.InstantiateInterfaces()
	if got == nil {
		t.Fatal("expected a non-nil empty slice, got nil")
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 instances, got %d", len(got))
	}
}

// --- GetInterface (nil) ---

// TestModuleGetInterfaceReturnsNilForNilReceiver verifies GetInterface itself
// returns nil for a nil receiver, exercising the guard directly rather than
// through the GetInterfaceByName alias.
//
// Why it matters: GetInterface is the underlying lookup used during cable wiring,
// so its own nil guard must hold even when callers reach it directly on a missing
// module.
// Inputs: a nil *CaniModuleType and any interface name. Outputs: the
// *InterfaceSpec result, expected nil.
// Data choice: calling GetInterface (not the alias) on a nil receiver targets the
// `m == nil` guard in the real method body; the existing nil test goes through
// GetInterfaceByName.
func TestModuleGetInterfaceReturnsNilForNilReceiver(t *testing.T) {
	var m *CaniModuleType
	if got := m.GetInterface("eth0"); got != nil {
		t.Fatalf("expected nil for nil receiver, got %v", got)
	}
}
