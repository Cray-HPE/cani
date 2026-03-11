package devicetypes

// Test coverage for cani_cable_types.go
//
// | Function                | Happy-path test                           | Failure test                              |
// |-------------------------|-------------------------------------------|-------------------------------------------|
// | NewCable                | TestNewCableCreatesValidCable              | TestNewCableEmptySlug                      |
// | Validate                | TestValidateHappyPath                      | TestValidateNilReceiver                    |
// | GetID                   | TestGetIDReturnsID                         | TestGetIDNilReceiver                       |
// | GetSlug                 | TestGetSlugReturnsSlug                     | TestGetSlugNilReceiver                     |
// | GetVendor               | TestGetVendorReturnsManufacturer            | TestGetVendorNilReceiver                   |
// | GetType                 | TestGetTypeReturnsCable                    | TestGetTypeAlwaysCable                     |
// | GetStatus               | TestGetStatusReturnsStatus                 | TestGetStatusNilReceiver                   |
// | SetTerminations         | TestSetTerminationsSetsUUIDs               | TestSetTerminationsNilReceiver             |
// | SetDeviceTerminations   | TestSetDeviceTerminationsSetsFields         | TestSetDeviceTerminationsNilReceiver       |
// | ValidateCable           | TestValidateCableHappyPath                 | TestValidateCableNilCable                  |
// | areInterfacesCompatible | TestAreInterfacesCompatibleSameType         | TestAreInterfacesCompatibleIncompatible     |

import (
	"testing"

	"github.com/google/uuid"
)

// --- NewCable ---

func TestNewCableCreatesValidCable(t *testing.T) {
	c := NewCable("cat6a", "Cat6a Patch")
	if c == nil {
		t.Fatal("expected non-nil cable")
	}
	if c.ID == uuid.Nil {
		t.Error("expected generated UUID, got Nil")
	}
	if c.Slug != "cat6a" {
		t.Errorf("slug = %q, want %q", c.Slug, "cat6a")
	}
	if c.Label != "Cat6a Patch" {
		t.Errorf("label = %q, want %q", c.Label, "Cat6a Patch")
	}
	if c.Status != "connected" {
		t.Errorf("status = %q, want %q", c.Status, "connected")
	}
}

func TestNewCableEmptySlug(t *testing.T) {
	c := NewCable("", "No Slug")
	if c.Slug != "" {
		t.Errorf("slug = %q, want empty string", c.Slug)
	}
	if err := c.Validate(); err == nil {
		t.Error("expected validation error for empty slug")
	}
}

// --- Validate ---

func TestValidateHappyPath(t *testing.T) {
	c := NewCable("smf", "Single-Mode Fiber")
	if err := c.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateNilReceiver(t *testing.T) {
	var c *CaniCableType
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error for nil receiver")
	}
	want := "cannot validate nil CaniCableType"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

// --- GetID ---

func TestGetIDReturnsID(t *testing.T) {
	c := NewCable("cat5e", "Cat5e")
	if c.GetID() != c.ID {
		t.Errorf("GetID() = %v, want %v", c.GetID(), c.ID)
	}
}

func TestGetIDNilReceiver(t *testing.T) {
	var c *CaniCableType
	if c.GetID() != uuid.Nil {
		t.Errorf("GetID() on nil = %v, want %v", c.GetID(), uuid.Nil)
	}
}

// --- GetSlug ---

func TestGetSlugReturnsSlug(t *testing.T) {
	c := NewCable("dac-passive", "DAC Passive")
	if c.GetSlug() != "dac-passive" {
		t.Errorf("GetSlug() = %q, want %q", c.GetSlug(), "dac-passive")
	}
}

func TestGetSlugNilReceiver(t *testing.T) {
	var c *CaniCableType
	if c.GetSlug() != "" {
		t.Errorf("GetSlug() on nil = %q, want empty", c.GetSlug())
	}
}

// --- GetVendor ---

func TestGetVendorReturnsManufacturer(t *testing.T) {
	c := NewCable("aoc", "AOC Cable")
	c.Manufacturer = "Acme Corp"
	if c.GetVendor() != "Acme Corp" {
		t.Errorf("GetVendor() = %q, want %q", c.GetVendor(), "Acme Corp")
	}
}

func TestGetVendorNilReceiver(t *testing.T) {
	var c *CaniCableType
	if c.GetVendor() != "" {
		t.Errorf("GetVendor() on nil = %q, want empty", c.GetVendor())
	}
}

// --- GetType ---

func TestGetTypeReturnsCable(t *testing.T) {
	c := NewCable("cat6", "Cat6")
	if c.GetType() != TypeCable {
		t.Errorf("GetType() = %v, want %v", c.GetType(), TypeCable)
	}
}

func TestGetTypeAlwaysCable(t *testing.T) {
	c := &CaniCableType{Slug: "custom"}
	if c.GetType() != TypeCable {
		t.Errorf("GetType() = %v, want %v even for bare struct", c.GetType(), TypeCable)
	}
}

// --- GetStatus ---

func TestGetStatusReturnsStatus(t *testing.T) {
	c := NewCable("cat6a", "Cat6a")
	if c.GetStatus() != "connected" {
		t.Errorf("GetStatus() = %q, want %q", c.GetStatus(), "connected")
	}
}

func TestGetStatusNilReceiver(t *testing.T) {
	var c *CaniCableType
	if c.GetStatus() != "" {
		t.Errorf("GetStatus() on nil = %q, want empty", c.GetStatus())
	}
}

// --- SetTerminations ---

func TestSetTerminationsSetsUUIDs(t *testing.T) {
	c := NewCable("cat6a", "Cat6a")
	a := uuid.New()
	b := uuid.New()
	c.SetTerminations(a, b)
	if c.TerminationA != a {
		t.Errorf("TerminationA = %v, want %v", c.TerminationA, a)
	}
	if c.TerminationB != b {
		t.Errorf("TerminationB = %v, want %v", c.TerminationB, b)
	}
}

func TestSetTerminationsNilReceiver(t *testing.T) {
	var c *CaniCableType
	// Should not panic
	c.SetTerminations(uuid.New(), uuid.New())
}

// --- SetDeviceTerminations ---

func TestSetDeviceTerminationsSetsFields(t *testing.T) {
	c := NewCable("smf", "SMF")
	dA := uuid.New()
	dB := uuid.New()
	c.SetDeviceTerminations(dA, dB, "eth0", "eth1")
	if c.TerminationADevice != dA {
		t.Errorf("TerminationADevice = %v, want %v", c.TerminationADevice, dA)
	}
	if c.TerminationBDevice != dB {
		t.Errorf("TerminationBDevice = %v, want %v", c.TerminationBDevice, dB)
	}
	if c.TerminationAPort != "eth0" {
		t.Errorf("TerminationAPort = %q, want %q", c.TerminationAPort, "eth0")
	}
	if c.TerminationBPort != "eth1" {
		t.Errorf("TerminationBPort = %q, want %q", c.TerminationBPort, "eth1")
	}
}

func TestSetDeviceTerminationsNilReceiver(t *testing.T) {
	var c *CaniCableType
	// Should not panic
	c.SetDeviceTerminations(uuid.New(), uuid.New(), "p0", "p1")
}

// --- ValidateCable ---

// newTestInventoryWithCable builds a minimal inventory with two devices, two
// interfaces, and a cable connecting them.  Caller gets the cable back for
// assertions.
func newTestInventoryWithCable(ifaceType InterfacesElemType) (*CaniCableType, *Inventory) {
	inv := NewInventory()

	devAID := uuid.New()
	devBID := uuid.New()
	ifaceAID := uuid.New()
	ifaceBID := uuid.New()

	inv.Devices[devAID] = &CaniDeviceType{
		ID:   devAID,
		Name: "switch-a",
		Interfaces: []InterfaceSpec{
			{ID: ifaceAID, Name: "eth0", Type: ifaceType},
		},
	}
	inv.Devices[devBID] = &CaniDeviceType{
		ID:   devBID,
		Name: "switch-b",
		Interfaces: []InterfaceSpec{
			{ID: ifaceBID, Name: "eth0", Type: ifaceType},
		},
	}
	inv.Interfaces[ifaceAID] = &InterfaceInstance{
		ID:       ifaceAID,
		DeviceID: devAID,
	}
	inv.Interfaces[ifaceBID] = &InterfaceInstance{
		ID:       ifaceBID,
		DeviceID: devBID,
	}

	cable := NewCable("cat6a", "Cat6a Patch")
	cable.SetTerminations(ifaceAID, ifaceBID)
	inv.Cables[cable.ID] = cable

	return cable, inv
}

func TestValidateCableHappyPath(t *testing.T) {
	cable, inv := newTestInventoryWithCable(InterfacesElemTypeA1000BaseT)
	if err := ValidateCable(cable, inv); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateCableNilCable(t *testing.T) {
	inv := NewInventory()
	err := ValidateCable(nil, inv)
	if err == nil {
		t.Fatal("expected error for nil cable")
	}
	want := "cable is nil"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

// --- areInterfacesCompatible ---

func TestAreInterfacesCompatibleSameType(t *testing.T) {
	if !areInterfacesCompatible(InterfacesElemTypeA100GbaseXQsfp28, InterfacesElemTypeA100GbaseXQsfp28) {
		t.Error("same interface type should be compatible")
	}
}

func TestAreInterfacesCompatibleIncompatible(t *testing.T) {
	if areInterfacesCompatible(InterfacesElemTypeA1000BaseT, InterfacesElemTypeA100GbaseXQsfp28) {
		t.Error("1000base-t and 100gbase-x-qsfp28 should not be compatible")
	}
}
