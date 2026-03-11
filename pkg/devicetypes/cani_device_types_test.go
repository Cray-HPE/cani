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

func TestIsCableReturnsTrueForCableType(t *testing.T) {
	d := &CaniDeviceType{Type: TypeCable}
	if !d.IsCable() {
		t.Fatal("expected IsCable to return true for TypeCable")
	}
}

func TestIsCableReturnsFalseForNonCable(t *testing.T) {
	d := &CaniDeviceType{Type: TypeNode}
	if d.IsCable() {
		t.Fatal("expected IsCable to return false for TypeNode")
	}
}

// --- GetVendor ---

func TestGetVendorReturnsVendor(t *testing.T) {
	d := &CaniDeviceType{Vendor: "Acme", Manufacturer: "FallbackCorp"}
	if got := d.GetVendor(); got != "Acme" {
		t.Fatalf("expected Acme, got %s", got)
	}
}

func TestGetVendorReturnsEmptyForNil(t *testing.T) {
	var d *CaniDeviceType
	if got := d.GetVendor(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetType ---

func TestGetTypeReturnsExplicitType(t *testing.T) {
	d := &CaniDeviceType{Type: TypeChassis}
	if got := d.GetType(); got != TypeChassis {
		t.Fatalf("expected %s, got %s", TypeChassis, got)
	}
}

func TestGetTypeReturnsEmptyForNil(t *testing.T) {
	var d *CaniDeviceType
	if got := d.GetType(); got != "" {
		t.Fatalf("expected empty Type, got %s", got)
	}
}

// --- MergeProperties ---

func TestMergePropertiesNilReceiverReturnsFalse(t *testing.T) {
	var d *CaniDeviceType
	other := &CaniDeviceType{Name: "test"}
	if d.MergeProperties(other) {
		t.Fatal("expected MergeProperties on nil receiver to return false")
	}
}

// --- MergeProperties ---

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

func TestMergePropertiesReturnsFalseForNils(t *testing.T) {
	var d *CaniDeviceType
	other := &CaniDeviceType{Name: "Anything"}
	if d.MergeProperties(other) {
		t.Fatal("expected MergeProperties to return false for nil receiver")
	}
}

// --- GetID ---

func TestGetIDReturnsUUID(t *testing.T) {
	id := uuid.New()
	d := &CaniDeviceType{ID: id}
	if got := d.GetID(); got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

func TestGetIDReturnsNilForNilReceiver(t *testing.T) {
	var d *CaniDeviceType
	if got := d.GetID(); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}

// --- GetSlug ---

func TestDeviceTypeGetSlugReturnsSlug(t *testing.T) {
	d := &CaniDeviceType{Slug: "my-device"}
	if got := d.GetSlug(); got != "my-device" {
		t.Fatalf("expected my-device, got %s", got)
	}
}

func TestDeviceTypeGetSlugReturnsEmptyForNil(t *testing.T) {
	var d *CaniDeviceType
	if got := d.GetSlug(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetStatus ---

func TestDeviceTypeGetStatusReturnsStatus(t *testing.T) {
	d := &CaniDeviceType{Status: "active"}
	if got := d.GetStatus(); got != "active" {
		t.Fatalf("expected active, got %s", got)
	}
}

func TestDeviceTypeGetStatusReturnsEmptyForNil(t *testing.T) {
	var d *CaniDeviceType
	if got := d.GetStatus(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- Validate ---

func TestValidatePassesForValidDevice(t *testing.T) {
	d := &CaniDeviceType{Name: "switch-1", Slug: "switch-1"}
	if err := d.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestValidateReturnsErrorForNil(t *testing.T) {
	var d *CaniDeviceType
	if err := d.Validate(); err == nil {
		t.Fatal("expected error when validating nil CaniDeviceType")
	}
}

// --- InstantiateInterfaces ---

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

func TestInstantiateInterfacesReturnsNilForNil(t *testing.T) {
	var d *CaniDeviceType
	if instances := d.InstantiateInterfaces(); instances != nil {
		t.Fatalf("expected nil, got %v", instances)
	}
}

// --- GetInterface ---

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

func TestGetRackIDReturnsExplicitRack(t *testing.T) {
	rackID := uuid.New()
	inv := NewInventory()
	d := &CaniDeviceType{Rack: rackID}
	if got := d.GetRackID(inv); got != rackID {
		t.Fatalf("expected %s, got %s", rackID, got)
	}
}

func TestGetRackIDReturnsNilForNilReceiver(t *testing.T) {
	var d *CaniDeviceType
	inv := NewInventory()
	if got := d.GetRackID(inv); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}

// --- GetUHeight ---

func TestGetUHeightReturnsSetValue(t *testing.T) {
	d := &CaniDeviceType{UHeight: 4}
	if got := d.GetUHeight(); got != 4 {
		t.Fatalf("expected 4, got %d", got)
	}
}

func TestGetUHeightReturnsDefaultForZero(t *testing.T) {
	d := &CaniDeviceType{UHeight: 0}
	if got := d.GetUHeight(); got != 1 {
		t.Fatalf("expected default of 1, got %d", got)
	}
}

// --- GetIsFullDepth ---

func TestGetIsFullDepthReturnsTrue(t *testing.T) {
	d := &CaniDeviceType{IsFullDepth: true}
	if !d.GetIsFullDepth() {
		t.Fatal("expected GetIsFullDepth to return true")
	}
}

func TestGetIsFullDepthReturnsFalseForNil(t *testing.T) {
	var d *CaniDeviceType
	if d.GetIsFullDepth() {
		t.Fatal("expected GetIsFullDepth to return false for nil receiver")
	}
}
