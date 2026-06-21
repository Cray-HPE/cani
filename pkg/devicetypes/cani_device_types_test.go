package devicetypes

// Test coverage for CaniDeviceType methods in cani_device_types.go.
//
// | Function              | Happy-path test                          | Failure test                                |
// |-----------------------|------------------------------------------|---------------------------------------------|
// | IsCable               | TestIsCableReturnsTrueForCableType       | TestIsCableReturnsFalseForNonCable          |
// | GetVendor             | TestGetVendorReturnsVendor               | TestGetVendorReturnsEmptyForNil             |
// | GetType               | TestGetTypeReturnsExplicitType           | TestGetTypeReturnsEmptyForNil               |
// | MergeProperties       | TestMergePropertiesAppliesChanges        | TestMergePropertiesNilReceiverReturnsFalse  |
// | GetID                 | TestGetIDReturnsUUID                     | TestGetIDReturnsNilForNilReceiver           |
// | GetSlug               | TestDeviceTypeGetSlugReturnsSlug          | TestDeviceTypeGetSlugReturnsEmptyForNil     |
// | GetStatus             | TestDeviceTypeGetStatusReturnsStatus      | TestDeviceTypeGetStatusReturnsEmptyForNil   |
// | Validate              | TestValidatePassesForValidDevice         | TestValidateReturnsErrorForNil              |
// | InstantiateInterfaces | TestInstantiateInterfacesCreatesInstances | TestInstantiateInterfacesReturnsNilForNil   |
// | GetInterface          | TestGetInterfaceFindsMatch               | TestGetInterfaceReturnsNilForMissing        |
// | GetRackID             | TestGetRackIDReturnsExplicitRack         | TestGetRackIDReturnsNilForNilReceiver       |
// | GetUHeight            | TestGetUHeightReturnsSetValue            | TestGetUHeightReturnsDefaultForZero         |
// | GetIsFullDepth        | TestGetIsFullDepthReturnsTrue            | TestGetIsFullDepthReturnsFalseForNil        |

import (
	"testing"

	"github.com/google/uuid"
)

// --- IsCable ---

// TestIsCableReturnsTrueForCableType verifies IsCable reports true when the
// device's Type is TypeCable.
//
// Why it matters: cables are handled on a separate export path from racked
// hardware, so IsCable is the discriminator that routes a device correctly.
// Inputs: a device with Type=TypeCable. Outputs: the bool result, expected true.
// Data choice: TypeCable is the one value that must return true, so it is the
// minimal positive case for the discriminator.
func TestIsCableReturnsTrueForCableType(t *testing.T) {
	d := &CaniDeviceType{Type: TypeCable}
	if !d.IsCable() {
		t.Fatal("expected IsCable to return true for TypeCable")
	}
}

// TestIsCableReturnsFalseForNonCable verifies IsCable reports false for a
// non-cable device type.
//
// Why it matters: routing a non-cable through the cable path would mis-handle
// real hardware, so every non-cable type must be rejected by the discriminator.
// Inputs: a device with Type=TypeNode. Outputs: the bool result, expected false.
// Data choice: TypeNode is a representative ordinary device type, proving the
// check keys on TypeCable specifically rather than accepting any set type.
func TestIsCableReturnsFalseForNonCable(t *testing.T) {
	d := &CaniDeviceType{Type: TypeNode}
	if d.IsCable() {
		t.Fatal("expected IsCable to return false for TypeNode")
	}
}

// --- GetVendor ---

// TestGetVendorReturnsVendor verifies GetVendor returns the explicit Vendor when
// it is set, ignoring Manufacturer.
//
// Why it matters: the export path prefers a curated Vendor over the raw
// Manufacturer, so a present Vendor must take precedence.
// Inputs: a device with Vendor="Acme" and Manufacturer="FallbackCorp". Outputs:
// the vendor string, expected "Acme".
// Data choice: a Vendor that differs from Manufacturer proves precedence; equal
// values could pass even if the fallback were wrongly chosen.
func TestGetVendorReturnsVendor(t *testing.T) {
	d := &CaniDeviceType{Vendor: "Acme", Manufacturer: "FallbackCorp"}
	if got := d.GetVendor(); got != "Acme" {
		t.Fatalf("expected Acme, got %s", got)
	}
}

// TestGetVendorReturnsEmptyForNil verifies GetVendor returns an empty string for
// a nil receiver instead of panicking.
//
// Why it matters: vendor lookups can run over map entries that may be nil, so the
// guard keeps export resilient to missing devices.
// Inputs: a nil *CaniDeviceType. Outputs: the vendor string, expected "".
// Data choice: a nil receiver is the only input that exercises the `c == nil`
// guard ahead of the field access.
func TestGetVendorReturnsEmptyForNil(t *testing.T) {
	var d *CaniDeviceType
	if got := d.GetVendor(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetType ---

// TestGetTypeReturnsExplicitType verifies GetType returns the device's set Type
// constant.
//
// Why it matters: downstream classification and export branch on the hardware
// type, so the accessor must return the stored value faithfully.
// Inputs: a device with Type=TypeChassis. Outputs: the Type result, expected
// TypeChassis.
// Data choice: TypeChassis is a concrete non-empty type, distinguishing a real
// return from the empty-string zero value.
func TestGetTypeReturnsExplicitType(t *testing.T) {
	d := &CaniDeviceType{Type: TypeChassis}
	if got := d.GetType(); got != TypeChassis {
		t.Fatalf("expected %s, got %s", TypeChassis, got)
	}
}

// TestGetTypeReturnsEmptyForNil verifies GetType returns the empty Type for a nil
// receiver instead of panicking.
//
// Why it matters: type lookups may run over nil map entries, so the guard keeps
// classification crash-free.
// Inputs: a nil *CaniDeviceType. Outputs: the Type result, expected "".
// Data choice: a nil receiver is the only input that reaches the `c == nil`
// guard before the field read.
func TestGetTypeReturnsEmptyForNil(t *testing.T) {
	var d *CaniDeviceType
	if got := d.GetType(); got != "" {
		t.Fatalf("expected empty Type, got %s", got)
	}
}

// --- MergeProperties ---

// TestMergePropertiesNilReceiverReturnsFalse verifies merging into a nil receiver
// reports no change.
//
// Why it matters: MergeProperties folds provider updates into an existing record,
// and a nil receiver must be rejected without a panic so callers can merge
// optional records safely.
// Inputs: a nil receiver and a non-empty other named "test". Outputs: the bool
// result, expected false.
// Data choice: a non-empty other ensures the false result comes from the nil
// guard, not from an empty source that would also report no change.
func TestMergePropertiesNilReceiverReturnsFalse(t *testing.T) {
	var d *CaniDeviceType
	other := &CaniDeviceType{Name: "test"}
	if d.MergeProperties(other) {
		t.Fatal("expected MergeProperties on nil receiver to return false")
	}
}

// --- MergeProperties ---

// TestMergePropertiesAppliesChanges verifies a differing non-empty field is
// copied from the source and reported as a change.
//
// Why it matters: MergeProperties is the one place provider updates are applied,
// so a changed field must both copy across and signal that a write occurred.
// Inputs: a receiver named "Old" and an other named "New". Outputs: the bool
// result (true) and the receiver's updated Name ("New").
// Data choice: distinct old/new names make the overwrite observable and the
// change flag unambiguous.
func TestMergePropertiesAppliesChanges(t *testing.T) {
	d := &CaniDeviceType{Name: "Old"}
	other := &CaniDeviceType{Name: "New"}
	if changed := d.MergeProperties(other); !changed {
		t.Fatal("expected MergeProperties to report a change")
	}
	if d.Name != "New" {
		t.Fatalf("expected Name to be New, got %s", d.Name)
	}
}

// TestMergePropertiesReturnsFalseForNils verifies a nil receiver reports no
// change even when the source carries data.
//
// Why it matters: the merge guard must protect against a nil receiver so callers
// merging optional records never crash.
// Inputs: a nil receiver and a non-empty other named "Anything". Outputs: the
// bool result, expected false.
// Data choice: a populated other proves the false stems from the nil-receiver
// guard rather than an empty payload.
func TestMergePropertiesReturnsFalseForNils(t *testing.T) {
	var d *CaniDeviceType
	other := &CaniDeviceType{Name: "Anything"}
	if d.MergeProperties(other) {
		t.Fatal("expected MergeProperties to return false for nil receiver")
	}
}

// --- GetID ---

// TestGetIDReturnsUUID verifies GetID returns the device's stored identifier.
//
// Why it matters: the ID keys the device in inventory maps and FK references, so
// the accessor must echo the exact stored UUID.
// Inputs: a device whose ID is a freshly generated UUID. Outputs: the UUID
// result, expected to equal that ID.
// Data choice: a random UUID avoids coincidental matches with any default.
func TestGetIDReturnsUUID(t *testing.T) {
	id := uuid.New()
	d := &CaniDeviceType{ID: id}
	if got := d.GetID(); got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

// TestGetIDReturnsNilForNilReceiver verifies GetID returns uuid.Nil for a nil
// receiver instead of panicking.
//
// Why it matters: ID lookups may run over nil map entries, so the guard keeps
// inventory traversal safe.
// Inputs: a nil *CaniDeviceType. Outputs: the UUID result, expected uuid.Nil.
// Data choice: a nil receiver is the only input reaching the `c == nil` guard.
func TestGetIDReturnsNilForNilReceiver(t *testing.T) {
	var d *CaniDeviceType
	if got := d.GetID(); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}

// --- GetSlug ---

// TestDeviceTypeGetSlugReturnsSlug verifies GetSlug returns the device's slug.
//
// Why it matters: the slug links an inventory device back to its hardware-library
// template, so the accessor must return it verbatim for that lookup.
// Inputs: a device with Slug="my-device". Outputs: the slug string, expected
// "my-device".
// Data choice: a non-empty slug distinguishes a real return from the empty zero
// value.
func TestDeviceTypeGetSlugReturnsSlug(t *testing.T) {
	d := &CaniDeviceType{Slug: "my-device"}
	if got := d.GetSlug(); got != "my-device" {
		t.Fatalf("expected my-device, got %s", got)
	}
}

// TestDeviceTypeGetSlugReturnsEmptyForNil verifies GetSlug returns an empty string
// for a nil receiver instead of panicking.
//
// Why it matters: slug lookups may run over nil map entries, so the guard keeps
// template resolution crash-free.
// Inputs: a nil *CaniDeviceType. Outputs: the slug string, expected "".
// Data choice: a nil receiver is the only input that reaches the `c == nil` guard.
func TestDeviceTypeGetSlugReturnsEmptyForNil(t *testing.T) {
	var d *CaniDeviceType
	if got := d.GetSlug(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetStatus ---

// TestDeviceTypeGetStatusReturnsStatus verifies GetStatus returns the status held
// in the embedded ObjectMeta.
//
// Why it matters: status drives lifecycle handling and export, so the accessor
// must surface the embedded metadata value.
// Inputs: a device with ObjectMeta.Status="Active". Outputs: the status string,
// expected "Active".
// Data choice: "Active" is a real status string, proving the accessor reads
// through to the embedded field rather than returning empty.
func TestDeviceTypeGetStatusReturnsStatus(t *testing.T) {
	d := &CaniDeviceType{ObjectMeta: ObjectMeta{Status: "Active"}}
	if got := d.GetStatus(); got != "Active" {
		t.Fatalf("expected Active, got %s", got)
	}
}

// TestDeviceTypeGetStatusReturnsEmptyForNil verifies GetStatus returns an empty
// string for a nil receiver instead of panicking.
//
// Why it matters: status lookups may run over nil map entries, so the guard keeps
// traversal safe.
// Inputs: a nil *CaniDeviceType. Outputs: the status string, expected "".
// Data choice: a nil receiver is the only input that reaches the `c == nil` guard.
func TestDeviceTypeGetStatusReturnsEmptyForNil(t *testing.T) {
	var d *CaniDeviceType
	if got := d.GetStatus(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- Validate ---

// TestValidatePassesForValidDevice verifies Validate succeeds when the device's
// slug exists in the hardware library.
//
// Why it matters: Validate gates devices before they enter inventory, so a device
// whose slug is registered must be accepted.
// Inputs: a slug registered in the library (removed in cleanup) and a device
// referencing it. Outputs: the error result, expected nil.
// Data choice: registering then deleting the slug via t.Cleanup isolates the test
// from global registry state while guaranteeing the lookup hits.
func TestValidatePassesForValidDevice(t *testing.T) {
	const slug = "test-valid-device-slug"
	RegisterDeviceType(CaniDeviceType{Slug: slug})
	t.Cleanup(func() {
		delete(allDeviceTypes, slug)
	})

	d := &CaniDeviceType{Name: "switch-1", Slug: slug}
	if err := d.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

// TestValidateReturnsErrorForUnknownSlug verifies Validate fails when the device's
// slug is not in the hardware library.
//
// Why it matters: Validate must reject devices that reference an unknown template
// so bad data never reaches inventory.
// Inputs: a device whose Slug is absent from the library. Outputs: the error
// result, expected non-nil.
// Data choice: an obviously fake slug guarantees a miss without depending on which
// real templates happen to be registered.
func TestValidateReturnsErrorForUnknownSlug(t *testing.T) {
	d := &CaniDeviceType{Name: "switch-1", Slug: "definitely-not-a-real-device-slug"}
	if err := d.Validate(); err == nil {
		t.Fatal("expected error for unknown device slug")
	}
}

// TestValidateAllowsBlankSlug verifies Validate succeeds when the slug is blank.
//
// Why it matters: custom devices without a library template are permitted, so a
// blank slug must skip the library lookup rather than fail.
// Inputs: a device with an empty Slug. Outputs: the error result, expected nil.
// Data choice: leaving Slug empty exercises the `c.Slug != ""` guard's false
// branch, the path that skips validation against the library.
func TestValidateAllowsBlankSlug(t *testing.T) {
	d := &CaniDeviceType{Name: "custom-device"}
	if err := d.Validate(); err != nil {
		t.Fatalf("unexpected validation error for blank slug: %v", err)
	}
}

// TestValidateReturnsErrorForNil verifies Validate returns an error for a nil
// receiver instead of panicking.
//
// Why it matters: Validate may be called on a device fetched from a map that
// turned out nil, so it must fail cleanly rather than crash.
// Inputs: a nil *CaniDeviceType. Outputs: the error result, expected non-nil.
// Data choice: a nil receiver is the only input that reaches the explicit nil
// check at the top of Validate.
func TestValidateReturnsErrorForNil(t *testing.T) {
	var d *CaniDeviceType
	if err := d.Validate(); err == nil {
		t.Fatal("expected error when validating nil CaniDeviceType")
	}
}

// --- InstantiateInterfaces ---

// TestInstantiateInterfacesCreatesInstances verifies each interface spec becomes a
// CaniInterface carrying the spec's name, mgmt flag, and the parent device ID.
//
// Why it matters: instantiating interfaces is how a device template's ports become
// concrete inventory records, so each spec must map to a correctly populated
// instance.
// Inputs: a device with two specs, eth0 (MgmtOnly=true) and eth1 (unset). Outputs:
// two instances; the eth0 name, its MgmtOnly=true, its DeviceID, and eth1's
// defaulted MgmtOnly=false.
// Data choice: one spec with an explicit MgmtOnly pointer and one without proves
// both the dereference and the nil-defaults-to-false branches in a single pass.
func TestInstantiateInterfacesCreatesInstances(t *testing.T) {
	mgmt := true
	deviceID := uuid.New()
	d := &CaniDeviceType{
		ID: deviceID,
		Interfaces: []InterfaceSpec{
			{Name: "eth0", Type: InterfacesElemTypeA1000BaseT, MgmtOnly: &mgmt},
			{Name: "eth1", Type: InterfacesElemTypeA10GbaseT},
		},
	}
	instances := d.InstantiateInterfaces()
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
	if instances[0].Name != "eth0" {
		t.Fatalf("expected eth0, got %s", instances[0].Name)
	}
	if !instances[0].MgmtOnly {
		t.Fatal("expected MgmtOnly to be true for eth0")
	}
	if instances[0].DeviceID != deviceID {
		t.Fatal("expected DeviceID to match the parent device ID")
	}
	if instances[1].MgmtOnly {
		t.Fatal("expected MgmtOnly to default to false for eth1")
	}
}

// TestInstantiateInterfacesReturnsNilForNil verifies a nil receiver yields a nil
// slice instead of panicking.
//
// Why it matters: instantiation may be attempted on a device lookup that returned
// nil, so the guard keeps the build resilient.
// Inputs: a nil *CaniDeviceType. Outputs: the slice result, expected nil.
// Data choice: a nil receiver is the only input reaching the `c == nil` guard
// ahead of the spec loop.
func TestInstantiateInterfacesReturnsNilForNil(t *testing.T) {
	var d *CaniDeviceType
	if instances := d.InstantiateInterfaces(); instances != nil {
		t.Fatalf("expected nil, got %v", instances)
	}
}

// --- GetInterface ---

// TestGetInterfaceFindsMatch verifies GetInterface returns the spec whose name
// matches the query.
//
// Why it matters: cable wiring resolves endpoints by interface name, so a present
// interface must be found and returned with its details intact.
// Inputs: a device with specs mgmt0 and eth0; the query "eth0". Outputs: the
// matched *InterfaceSpec, and its Type checked to confirm identity.
// Data choice: two distinct interfaces ensure the lookup selects by name rather
// than returning the only element, and the Type assertion proves the right one
// came back.
func TestGetInterfaceFindsMatch(t *testing.T) {
	d := &CaniDeviceType{
		Interfaces: []InterfaceSpec{
			{Name: "mgmt0", Type: InterfacesElemTypeA1000BaseT},
			{Name: "eth0", Type: InterfacesElemTypeA10GbaseT},
		},
	}
	iface := d.GetInterface("eth0")
	if iface == nil {
		t.Fatal("expected to find interface eth0")
	}
	if iface.Type != InterfacesElemTypeA10GbaseT {
		t.Fatalf("expected type %s, got %s", InterfacesElemTypeA10GbaseT, iface.Type)
	}
}

// TestGetInterfaceReturnsNilForMissing verifies GetInterface returns nil when no
// spec matches the queried name.
//
// Why it matters: wiring must detect an absent endpoint cleanly rather than
// mis-resolve to the wrong interface.
// Inputs: a device with a single spec eth0; the query "nonexistent". Outputs: the
// result, expected nil.
// Data choice: a query that matches none of the present names exercises the
// loop's fall-through return.
func TestGetInterfaceReturnsNilForMissing(t *testing.T) {
	d := &CaniDeviceType{
		Interfaces: []InterfaceSpec{
			{Name: "eth0"},
		},
	}
	if iface := d.GetInterface("nonexistent"); iface != nil {
		t.Fatal("expected nil for a non-existent interface name")
	}
}

// --- GetRackID ---

// TestGetRackIDReturnsExplicitRack verifies GetRackID returns the explicit Rack FK
// when one is set.
//
// Why it matters: a device's curated rack assignment must win for correct
// placement and export, ahead of any parent-based inference.
// Inputs: a device with Rack set and an empty inventory. Outputs: the UUID result,
// expected to equal the explicit Rack.
// Data choice: an empty inventory ensures the returned value can only come from the
// explicit Rack branch, not from a parent lookup.
func TestGetRackIDReturnsExplicitRack(t *testing.T) {
	rackID := uuid.New()
	inv := NewInventory()
	d := &CaniDeviceType{Rack: rackID}
	if got := d.GetRackID(inv); got != rackID {
		t.Fatalf("expected %s, got %s", rackID, got)
	}
}

// TestGetRackIDReturnsNilForNilReceiver verifies GetRackID returns uuid.Nil for a
// nil receiver instead of panicking.
//
// Why it matters: placement queries may run over nil map entries, so the guard
// keeps rack resolution safe.
// Inputs: a nil *CaniDeviceType and a valid inventory. Outputs: the UUID result,
// expected uuid.Nil.
// Data choice: a nil receiver isolates the `c == nil` half of the guard while a
// real inventory rules out a nil-inventory short-circuit.
func TestGetRackIDReturnsNilForNilReceiver(t *testing.T) {
	var d *CaniDeviceType
	inv := NewInventory()
	if got := d.GetRackID(inv); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}

// --- GetUHeight ---

// TestGetUHeightReturnsSetValue verifies GetUHeight returns the stored height when
// it is a positive value.
//
// Why it matters: rack-layout math depends on the real occupied height, so a set
// value must pass through unmodified.
// Inputs: a device with UHeight=4. Outputs: the int result, expected 4.
// Data choice: 4 is a typical multi-U height greater than the default of 1, so the
// assertion fails if the default were wrongly returned.
func TestGetUHeightReturnsSetValue(t *testing.T) {
	d := &CaniDeviceType{UHeight: 4}
	if got := d.GetUHeight(); got != 4 {
		t.Fatalf("expected 4, got %d", got)
	}
}

// TestGetUHeightReturnsDefaultForZero verifies GetUHeight returns the default of 1
// when the height is unset (zero).
//
// Why it matters: an unspecified height must still occupy one rack unit so layout
// math never places a zero-height device.
// Inputs: a device with UHeight=0. Outputs: the int result, expected 1.
// Data choice: zero is the unset zero value that triggers the `< 1` default branch.
func TestGetUHeightReturnsDefaultForZero(t *testing.T) {
	d := &CaniDeviceType{UHeight: 0}
	if got := d.GetUHeight(); got != 1 {
		t.Fatalf("expected default of 1, got %d", got)
	}
}

// --- GetIsFullDepth ---

// TestGetIsFullDepthReturnsTrue verifies GetIsFullDepth returns true when the
// device is marked full-depth.
//
// Why it matters: full-depth occupancy affects rack rendering and slot conflicts,
// so a true flag must be reported faithfully.
// Inputs: a device with IsFullDepth=true. Outputs: the bool result, expected true.
// Data choice: the explicit true is the minimal positive case for the accessor.
func TestGetIsFullDepthReturnsTrue(t *testing.T) {
	d := &CaniDeviceType{IsFullDepth: true}
	if !d.GetIsFullDepth() {
		t.Fatal("expected GetIsFullDepth to return true")
	}
}

// TestGetIsFullDepthReturnsFalseForNil verifies GetIsFullDepth returns false for a
// nil receiver instead of panicking.
//
// Why it matters: depth lookups may run over nil map entries, so the guard keeps
// rack rendering safe.
// Inputs: a nil *CaniDeviceType. Outputs: the bool result, expected false.
// Data choice: a nil receiver is the only input that reaches the `c == nil` guard.
func TestGetIsFullDepthReturnsFalseForNil(t *testing.T) {
	var d *CaniDeviceType
	if d.GetIsFullDepth() {
		t.Fatal("expected GetIsFullDepth to return false for nil receiver")
	}
}

// ========== additional edge-case / branch coverage ==========

// --- IsCable (nil) ---

// TestIsCableReturnsFalseForNil verifies a nil receiver reports not-a-cable
// instead of panicking.
//
// Why it matters: IsCable is called while classifying mixed device collections
// where a map lookup may yield a nil pointer, so the nil guard must hold to keep
// classification crash-free.
// Inputs: a nil *CaniDeviceType. Outputs: the bool result, expected false.
// Data choice: an untyped nil pointer is the only input that exercises the
// `c == nil` guard, the one branch the type-based tests cannot reach.
func TestIsCableReturnsFalseForNil(t *testing.T) {
	var d *CaniDeviceType
	if d.IsCable() {
		t.Fatal("expected IsCable to return false for nil receiver")
	}
}

// --- GetVendor (fallback) ---

// TestGetVendorFallsBackToManufacturer verifies GetVendor returns Manufacturer
// when the dedicated Vendor field is empty.
//
// Why it matters: many device-type YAMLs set only Manufacturer, so the export
// path relies on this fallback to populate a vendor rather than emitting a blank.
// Inputs: a device with Vendor="" and Manufacturer="FallbackCorp". Outputs: the
// vendor string, expected "FallbackCorp".
// Data choice: an empty Vendor with a distinct Manufacturer value isolates the
// fallback branch; the existing vendor test sets Vendor and never reaches it.
func TestGetVendorFallsBackToManufacturer(t *testing.T) {
	d := &CaniDeviceType{Vendor: "", Manufacturer: "FallbackCorp"}
	if got := d.GetVendor(); got != "FallbackCorp" {
		t.Fatalf("expected fallback to Manufacturer 'FallbackCorp', got %q", got)
	}
}

// --- MergeProperties (branches) ---

// TestMergePropertiesNilOtherReturnsFalse verifies merging a nil source into a
// real receiver is a no-op that reports no change.
//
// Why it matters: callers merge an optional "other" record that may be nil, so
// the guard must protect the receiver's existing data and avoid a nil deref.
// Inputs: a receiver named "keep" and a nil other. Outputs: the bool result
// (false) and the receiver's unchanged Name.
// Data choice: a nil other (as opposed to the nil receiver the existing tests
// use) targets the second half of the `c == nil || other == nil` guard.
func TestMergePropertiesNilOtherReturnsFalse(t *testing.T) {
	d := &CaniDeviceType{Name: "keep"}
	if d.MergeProperties(nil) {
		t.Fatal("expected MergeProperties(nil) to return false")
	}
	if d.Name != "keep" {
		t.Fatalf("expected Name to stay 'keep', got %q", d.Name)
	}
}

// TestMergePropertiesNoChangeWhenOtherEmpty verifies an all-zero source produces
// no change and leaves the receiver untouched.
//
// Why it matters: idempotent re-merges must not report spurious changes, since a
// false "changed" result would trigger needless downstream writes/updates.
// Inputs: a receiver with Name/Slug set and an empty other. Outputs: the bool
// result (false) and the receiver's preserved fields.
// Data choice: RackPosition is left at 0 so the face-defaulting branch cannot
// fire, ensuring the only possible outcome is the all-fields-skipped no-op path.
func TestMergePropertiesNoChangeWhenOtherEmpty(t *testing.T) {
	d := &CaniDeviceType{Name: "node", Slug: "node-slug"}
	if d.MergeProperties(&CaniDeviceType{}) {
		t.Fatal("expected no change when other is empty")
	}
	if d.Name != "node" || d.Slug != "node-slug" {
		t.Fatalf("expected receiver unchanged, got Name=%q Slug=%q", d.Name, d.Slug)
	}
}

// TestMergePropertiesMergesIndividualFields verifies each simple non-empty source
// field overwrites the receiver and reports a change.
//
// Why it matters: MergeProperties is the single place provider updates are folded
// into an existing record, so every covered field must actually copy across for
// updates to take effect.
// Inputs: per case, an empty receiver and an other carrying one set field.
// Outputs: the bool result (true) and the merged field value, checked per case.
// Data choice: only side-effect-free fields are tabled here (RackPosition/Face
// are tested separately because they also trigger the face default), so each row
// asserts exactly one branch in isolation.
func TestMergePropertiesMergesIndividualFields(t *testing.T) {
	parent := uuid.New()
	cases := []struct {
		name  string
		other *CaniDeviceType
		check func(*CaniDeviceType) bool
	}{
		{"slug", &CaniDeviceType{Slug: "new-slug"}, func(d *CaniDeviceType) bool { return d.Slug == "new-slug" }},
		{"manufacturer", &CaniDeviceType{Manufacturer: "HPE"}, func(d *CaniDeviceType) bool { return d.Manufacturer == "HPE" }},
		{"model", &CaniDeviceType{Model: "X1"}, func(d *CaniDeviceType) bool { return d.Model == "X1" }},
		{"status", &CaniDeviceType{ObjectMeta: ObjectMeta{Status: "Active"}}, func(d *CaniDeviceType) bool { return d.Status == "Active" }},
		{"type", &CaniDeviceType{Type: TypeChassis}, func(d *CaniDeviceType) bool { return d.Type == TypeChassis }},
		{"uheight", &CaniDeviceType{UHeight: 4}, func(d *CaniDeviceType) bool { return d.UHeight == 4 }},
		{"parent", &CaniDeviceType{Parent: parent}, func(d *CaniDeviceType) bool { return d.Parent == parent }},
		{"role", &CaniDeviceType{ObjectMeta: ObjectMeta{Role: "leaf"}}, func(d *CaniDeviceType) bool { return d.Role == "leaf" }},
		{"serial", &CaniDeviceType{Serial: "SN1"}, func(d *CaniDeviceType) bool { return d.Serial == "SN1" }},
		{"description", &CaniDeviceType{Description: "desc"}, func(d *CaniDeviceType) bool { return d.Description == "desc" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &CaniDeviceType{}
			if !d.MergeProperties(tc.other) {
				t.Errorf("expected a change when merging %s", tc.name)
			}
			if !tc.check(d) {
				t.Errorf("expected %s to be merged into receiver", tc.name)
			}
		})
	}
}

// TestMergePropertiesDefaultsFaceWhenRackPositionSet verifies merging a rack
// position with no face defaults the face to "front".
//
// Why it matters: Nautobot requires rack, position, and face together, so the
// merge must supply a default face whenever a position arrives without one, or
// the export would be rejected.
// Inputs: an empty receiver and an other with RackPosition=5, Face="". Outputs:
// the bool result (true), the merged RackPosition, and the defaulted Face.
// Data choice: a positive position paired with an empty face is the exact
// condition (`c.RackPosition > 0 && c.Face == ""`) that triggers the default.
func TestMergePropertiesDefaultsFaceWhenRackPositionSet(t *testing.T) {
	d := &CaniDeviceType{}
	if !d.MergeProperties(&CaniDeviceType{RackPosition: 5}) {
		t.Fatal("expected a change when merging a rack position")
	}
	if d.RackPosition != 5 {
		t.Fatalf("expected RackPosition 5, got %d", d.RackPosition)
	}
	if d.Face != "front" {
		t.Fatalf("expected Face to default to 'front', got %q", d.Face)
	}
}

// TestMergePropertiesPreservesExplicitFace verifies an explicit source face is
// merged verbatim and not overridden by the "front" default.
//
// Why it matters: rear-mounted gear must keep its declared face, so an explicit
// value has to win over the default-fill behavior.
// Inputs: an empty receiver and an other with RackPosition=5, Face="rear".
// Outputs: the bool result (true) and the merged Face, expected "rear".
// Data choice: "rear" differs from the "front" default, so the assertion fails
// loudly if the default ever clobbers an explicitly provided face.
func TestMergePropertiesPreservesExplicitFace(t *testing.T) {
	d := &CaniDeviceType{}
	if !d.MergeProperties(&CaniDeviceType{RackPosition: 5, Face: "rear"}) {
		t.Fatal("expected a change when merging position and face")
	}
	if d.Face != "rear" {
		t.Fatalf("expected Face 'rear' to be preserved, got %q", d.Face)
	}
}

// TestMergePropertiesMergesProviderMetadataScalar verifies a top-level scalar
// provider-metadata key is copied into a receiver that had none.
//
// Why it matters: provider metadata (xnames, aliases, external IDs) must survive
// a merge so provider-specific identifiers are not lost when records combine.
// Inputs: an empty receiver and an other whose ProviderMetadata holds {"xname":
// "x3000c0s1b0"}. Outputs: the bool result (true) and the copied map value.
// Data choice: a string value (not a map) under a top-level key drives the
// scalar branch, and the nil receiver map forces the lazy `make` allocation path.
func TestMergePropertiesMergesProviderMetadataScalar(t *testing.T) {
	d := &CaniDeviceType{}
	other := &CaniDeviceType{
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{"xname": "x3000c0s1b0"}},
	}
	if !d.MergeProperties(other) {
		t.Fatal("expected a change when merging provider metadata")
	}
	if d.ProviderMetadata["xname"] != "x3000c0s1b0" {
		t.Fatalf("expected xname to be merged, got %v", d.ProviderMetadata["xname"])
	}
}

// TestMergePropertiesMergesProviderMetadataSubMap verifies a provider sub-map is
// merged key-by-key, preserving the receiver's existing sub-keys while also
// allocating sub-maps for provider keys the receiver lacks.
//
// Why it matters: providers store nested metadata (e.g. a "nautobot" sub-map),
// and a merge must union the inner keys rather than replace the whole sub-map and
// drop previously captured data, while still accepting entirely new providers.
// Inputs: a receiver with ProviderMetadata{"nautobot":{"a":"1"}} and an other
// with {"nautobot":{"b":"2"},"csm":{"xname":"x1"}}. Outputs: the bool result
// (true) and the merged metadata.
// Data choice: disjoint inner keys (a vs. b) prove a key-by-key union of an
// existing sub-map, while the brand-new "csm" key drives the allocate-missing
// sub-map branch in the same pass.
func TestMergePropertiesMergesProviderMetadataSubMap(t *testing.T) {
	d := &CaniDeviceType{
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{"nautobot": map[string]any{"a": "1"}}},
	}
	other := &CaniDeviceType{
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"nautobot": map[string]any{"b": "2"},
			"csm":      map[string]any{"xname": "x1"},
		}},
	}
	if !d.MergeProperties(other) {
		t.Fatal("expected a change when merging a provider sub-map")
	}
	sub, ok := d.ProviderMetadata["nautobot"].(map[string]any)
	if !ok {
		t.Fatalf("expected nautobot sub-map, got %T", d.ProviderMetadata["nautobot"])
	}
	if sub["a"] != "1" || sub["b"] != "2" {
		t.Fatalf("expected sub-map to contain a=1 and b=2, got %v", sub)
	}
	csm, ok := d.ProviderMetadata["csm"].(map[string]any)
	if !ok || csm["xname"] != "x1" {
		t.Fatalf("expected newly allocated csm sub-map with xname=x1, got %v", d.ProviderMetadata["csm"])
	}
}

// --- InstantiateInterfaces (edge) ---

// TestInstantiateInterfacesEmptyReturnsNonNilEmpty verifies a device with no
// interface specs returns a non-nil, empty slice.
//
// Why it matters: callers append to and range over the result, so a device that
// simply has zero interfaces must yield an empty slice (distinct from the nil a
// nil receiver returns) to keep that contract unambiguous.
// Inputs: a non-nil device with an empty Interfaces slice. Outputs: the returned
// slice, expected non-nil with length 0.
// Data choice: an empty (not nil) Interfaces field separates the "no specs" case
// from the nil-receiver case, which the existing nil test already covers.
func TestInstantiateInterfacesEmptyReturnsNonNilEmpty(t *testing.T) {
	d := &CaniDeviceType{ID: uuid.New()}
	got := d.InstantiateInterfaces()
	if got == nil {
		t.Fatal("expected a non-nil empty slice, got nil")
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 instances, got %d", len(got))
	}
}

// TestInstantiateInterfacesUsesExplicitRole verifies an explicit spec role is
// passed through and each instance is stamped Active with a fresh ID.
//
// Why it matters: instantiated interfaces feed inventory records, so an
// operator-declared role must be honored and every instance needs a unique ID
// and an Active status to be valid downstream.
// Inputs: a device with one interface spec carrying Role="custom-role". Outputs:
// the instance's Role, Status, and ID.
// Data choice: a non-empty explicit role exercises the ResolveInterfaceRole
// short-circuit (explicit wins over inference), which name-based inference tests
// cannot prove.
func TestInstantiateInterfacesUsesExplicitRole(t *testing.T) {
	d := &CaniDeviceType{
		ID: uuid.New(),
		Interfaces: []InterfaceSpec{
			{Name: "eth0", Type: InterfacesElemTypeA10GbaseT, Role: "custom-role"},
		},
	}
	got := d.InstantiateInterfaces()
	if len(got) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(got))
	}
	if got[0].Role != "custom-role" {
		t.Errorf("expected explicit role 'custom-role', got %q", got[0].Role)
	}
	if got[0].Status != string(StatusActive) {
		t.Errorf("expected status %q, got %q", StatusActive, got[0].Status)
	}
	if got[0].ID == uuid.Nil {
		t.Error("expected a freshly assigned interface ID, got uuid.Nil")
	}
}

// --- GetInterface (nil) ---

// TestGetInterfaceReturnsNilForNilReceiver verifies a nil receiver yields nil
// instead of panicking.
//
// Why it matters: GetInterface is used during cable wiring where a device lookup
// may return nil, so the guard keeps wiring resilient to missing devices.
// Inputs: a nil *CaniDeviceType and any interface name. Outputs: the *InterfaceSpec
// result, expected nil.
// Data choice: a nil receiver is the only input reaching the `c == nil` guard,
// the branch the found/missing tests on a real device cannot hit.
func TestGetInterfaceReturnsNilForNilReceiver(t *testing.T) {
	var d *CaniDeviceType
	if iface := d.GetInterface("eth0"); iface != nil {
		t.Fatalf("expected nil for nil receiver, got %v", iface)
	}
}

// --- GetRackID (branches) ---

// TestGetRackIDReturnsNilForNilInventory verifies a nil inventory yields uuid.Nil.
//
// Why it matters: GetRackID consults the inventory's Racks map, so a nil
// inventory must be handled without a deref to keep placement queries safe.
// Inputs: a device with an explicit Rack set and a nil inventory. Outputs: the
// uuid result, expected uuid.Nil.
// Data choice: an explicit Rack is set deliberately to prove the nil-inventory
// guard short-circuits before the Rack value would otherwise be returned.
func TestGetRackIDReturnsNilForNilInventory(t *testing.T) {
	d := &CaniDeviceType{Rack: uuid.New()}
	if got := d.GetRackID(nil); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil for nil inventory, got %s", got)
	}
}

// TestGetRackIDFallsBackToParentRack verifies that with no explicit Rack the
// device resolves its rack via a Parent present in the inventory's Racks.
//
// Why it matters: devices placed directly in a rack carry only a Parent link, so
// this fallback is how their rack is discovered for placement and export.
// Inputs: a device with Rack unset and Parent=R, and an inventory whose Racks
// map contains R. Outputs: the uuid result, expected to equal Parent.
// Data choice: leaving Rack at uuid.Nil forces the code past the explicit-Rack
// branch into the parent-lookup branch this test is targeting.
func TestGetRackIDFallsBackToParentRack(t *testing.T) {
	parent := uuid.New()
	inv := NewInventory()
	inv.Racks[parent] = &CaniRackType{}

	d := &CaniDeviceType{Parent: parent}
	if got := d.GetRackID(inv); got != parent {
		t.Fatalf("expected fallback to parent rack %s, got %s", parent, got)
	}
}

// TestGetRackIDReturnsNilWhenNotInRack verifies a device with neither an explicit
// Rack nor a rack-backed Parent resolves to uuid.Nil.
//
// Why it matters: free-standing devices are not in any rack, and the resolver
// must report that cleanly rather than inventing a placement.
// Inputs: a device with Rack unset and a Parent that is absent from Racks, plus
// an empty inventory. Outputs: the uuid result, expected uuid.Nil.
// Data choice: a Parent deliberately not inserted into Racks exercises the final
// "not found" return after both prior branches fail.
func TestGetRackIDReturnsNilWhenNotInRack(t *testing.T) {
	d := &CaniDeviceType{Parent: uuid.New()}
	inv := NewInventory()
	if got := d.GetRackID(inv); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil when not in a rack, got %s", got)
	}
}

// --- GetUHeight (defaults) ---

// TestGetUHeightReturnsDefaultForNil verifies a nil receiver returns the default
// height of 1.
//
// Why it matters: U-height feeds rack-layout math, so a missing device must still
// occupy a sane single unit rather than panic or report zero.
// Inputs: a nil *CaniDeviceType. Outputs: the int result, expected 1.
// Data choice: a nil receiver hits the `c == nil` half of the guard that the
// zero-value test on a real device cannot reach.
func TestGetUHeightReturnsDefaultForNil(t *testing.T) {
	var d *CaniDeviceType
	if got := d.GetUHeight(); got != 1 {
		t.Fatalf("expected default of 1 for nil receiver, got %d", got)
	}
}

// TestGetUHeightReturnsDefaultForNegative verifies a negative height is floored
// to the default of 1.
//
// Why it matters: corrupt or under-specified data must not yield a non-positive
// height that would break rack-layout arithmetic.
// Inputs: a device with UHeight=-3. Outputs: the int result, expected 1.
// Data choice: a negative value (rather than zero) proves the guard uses `< 1`,
// covering the sub-zero edge the existing zero test leaves open.
func TestGetUHeightReturnsDefaultForNegative(t *testing.T) {
	d := &CaniDeviceType{UHeight: -3}
	if got := d.GetUHeight(); got != 1 {
		t.Fatalf("expected default of 1 for negative height, got %d", got)
	}
}

// --- GetIsFullDepth (explicit false) ---

// TestGetIsFullDepthReturnsFalseWhenUnset verifies a real device with
// IsFullDepth=false reports false.
//
// Why it matters: half-depth gear must be reported accurately for rack rendering,
// so the non-nil false path has to return the stored value unchanged.
// Inputs: a device with IsFullDepth explicitly false. Outputs: the bool result,
// expected false.
// Data choice: a non-nil receiver with the field left false distinguishes the
// real-value path from the nil-guard path the existing nil test covers.
func TestGetIsFullDepthReturnsFalseWhenUnset(t *testing.T) {
	d := &CaniDeviceType{IsFullDepth: false}
	if d.GetIsFullDepth() {
		t.Fatal("expected GetIsFullDepth to return false when unset")
	}
}
