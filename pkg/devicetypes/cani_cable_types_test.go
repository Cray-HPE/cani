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
	"strings"
	"testing"

	"github.com/google/uuid"
)

// --- NewCable ---

// TestNewCableCreatesValidCable verifies NewCable returns a populated cable with a
// generated ID, the given slug and label, and a default "Connected" status.
//
// Why it matters: NewCable is the constructor every cable flows through, so it
// must stamp identity and the default lifecycle state callers rely on.
// Inputs: slug "cat6a" and label "Cat6a Patch". Outputs: a *CaniCableType with a
// non-nil ID, matching slug/label, and status "Connected".
// Data choice: a realistic copper-patch slug and label make the field mapping
// obvious while exercising the default-status assignment.
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
	if c.Status != "Connected" {
		t.Errorf("status = %q, want %q", c.Status, "Connected")
	}
}

// TestNewCableEmptySlug verifies NewCable stores an empty slug verbatim and that
// the resulting cable fails Validate.
//
// Why it matters: the constructor must not invent a slug, and a slug-less cable
// must be caught by validation rather than entering inventory.
// Inputs: an empty slug and label "No Slug". Outputs: the stored slug (expected
// "") and a Validate error.
// Data choice: an empty slug is the minimal input that isolates the
// required-field check downstream of construction.
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

// TestValidateHappyPath verifies Validate returns nil for a cable with a slug.
//
// Why it matters: Validate gates a cable before use, so a well-formed cable must
// pass.
// Inputs: a cable built with slug "smf". Outputs: an error, nil expected.
// Data choice: a single-mode-fiber slug is a representative valid cable clearing
// the required-slug check.
func TestValidateHappyPath(t *testing.T) {
	c := NewCable("smf", "Single-Mode Fiber")
	if err := c.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestValidateNilReceiver verifies Validate returns a descriptive error for a nil
// receiver.
//
// Why it matters: a nil cable is a programming error, and Validate must fail with
// a clear message rather than panic.
// Inputs: a nil *CaniCableType. Outputs: an error equal to "cannot validate nil
// CaniCableType".
// Data choice: asserting the exact message confirms the nil guard, not some later
// field check, produced the error.
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

// TestGetIDReturnsID verifies GetID returns the cable's stored UUID.
//
// Why it matters: the ID keys a cable across inventory and export, so the
// accessor must return it unchanged.
// Inputs: a cable built by NewCable. Outputs: the UUID, expected equal to the
// cable's ID field.
// Data choice: comparing GetID() to the struct field proves the getter reflects
// the stored value rather than a constant.
func TestGetIDReturnsID(t *testing.T) {
	c := NewCable("cat5e", "Cat5e")
	if c.GetID() != c.ID {
		t.Errorf("GetID() = %v, want %v", c.GetID(), c.ID)
	}
}

// TestGetIDNilReceiver verifies GetID returns uuid.Nil for a nil receiver.
//
// Why it matters: callers treat uuid.Nil as "no cable", so a nil receiver must
// yield it rather than panic.
// Inputs: a nil *CaniCableType. Outputs: the UUID, expected uuid.Nil.
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestGetIDNilReceiver(t *testing.T) {
	var c *CaniCableType
	if c.GetID() != uuid.Nil {
		t.Errorf("GetID() on nil = %v, want %v", c.GetID(), uuid.Nil)
	}
}

// --- GetSlug ---

// TestGetSlugReturnsSlug verifies GetSlug returns the cable's slug.
//
// Why it matters: the slug links a cable to its library template, so the accessor
// must surface it unchanged.
// Inputs: a cable with slug "dac-passive". Outputs: the slug string, expected
// "dac-passive".
// Data choice: a passive-DAC slug is an arbitrary non-empty value showing the
// stored field is returned verbatim.
func TestGetSlugReturnsSlug(t *testing.T) {
	c := NewCable("dac-passive", "DAC Passive")
	if c.GetSlug() != "dac-passive" {
		t.Errorf("GetSlug() = %q, want %q", c.GetSlug(), "dac-passive")
	}
}

// TestGetSlugNilReceiver verifies GetSlug returns "" for a nil receiver.
//
// Why it matters: a nil cable must degrade to "" rather than panic when its slug
// is read.
// Inputs: a nil *CaniCableType. Outputs: the slug string, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestGetSlugNilReceiver(t *testing.T) {
	var c *CaniCableType
	if c.GetSlug() != "" {
		t.Errorf("GetSlug() on nil = %q, want empty", c.GetSlug())
	}
}

// --- GetVendor ---

// TestGetVendorReturnsManufacturer verifies GetVendor returns the cable's
// Manufacturer.
//
// Why it matters: vendor identifies the cable maker for export, and GetVendor
// maps the Manufacturer field onto that shared accessor.
// Inputs: a cable with Manufacturer "Acme Corp". Outputs: the vendor string,
// expected "Acme Corp".
// Data choice: a set Manufacturer distinct from the default proves the mapping
// returns the stored value.
func TestGetVendorReturnsManufacturer(t *testing.T) {
	c := NewCable("aoc", "AOC Cable")
	c.Manufacturer = "Acme Corp"
	if c.GetVendor() != "Acme Corp" {
		t.Errorf("GetVendor() = %q, want %q", c.GetVendor(), "Acme Corp")
	}
}

// TestGetVendorNilReceiver verifies GetVendor returns "" for a nil receiver.
//
// Why it matters: a nil cable must report no vendor rather than panic when its
// maker is queried.
// Inputs: a nil *CaniCableType. Outputs: the vendor string, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestGetVendorNilReceiver(t *testing.T) {
	var c *CaniCableType
	if c.GetVendor() != "" {
		t.Errorf("GetVendor() on nil = %q, want empty", c.GetVendor())
	}
}

// --- GetType ---

// TestGetTypeReturnsCable verifies GetType reports TypeCable for a constructed
// cable.
//
// Why it matters: type drives classification and export, and every cable must
// classify as a cable.
// Inputs: a cable built by NewCable. Outputs: the Type, expected TypeCable.
// Data choice: a normally constructed cable exercises the common path that must
// always yield TypeCable.
func TestGetTypeReturnsCable(t *testing.T) {
	c := NewCable("cat6", "Cat6")
	if c.GetType() != TypeCable {
		t.Errorf("GetType() = %v, want %v", c.GetType(), TypeCable)
	}
}

// TestGetTypeAlwaysCable verifies GetType reports TypeCable even for a bare struct
// with no constructor.
//
// Why it matters: GetType is a constant for cables, so it must not depend on any
// field being set.
// Inputs: a bare &CaniCableType{Slug: "custom"}. Outputs: the Type, expected
// TypeCable.
// Data choice: a hand-built struct bypassing NewCable confirms the type is fixed,
// not assigned during construction.
func TestGetTypeAlwaysCable(t *testing.T) {
	c := &CaniCableType{Slug: "custom"}
	if c.GetType() != TypeCable {
		t.Errorf("GetType() = %v, want %v even for bare struct", c.GetType(), TypeCable)
	}
}

// --- GetStatus ---

// TestGetStatusReturnsStatus verifies GetStatus returns the cable's status.
//
// Why it matters: status drives lifecycle handling and export filtering, so the
// accessor must surface the stored value.
// Inputs: a cable built by NewCable (default status "Connected"). Outputs: the
// status string, expected "Connected".
// Data choice: the constructor's default status is the representative value the
// getter must echo.
func TestGetStatusReturnsStatus(t *testing.T) {
	c := NewCable("cat6a", "Cat6a")
	if c.GetStatus() != "Connected" {
		t.Errorf("GetStatus() = %q, want %q", c.GetStatus(), "Connected")
	}
}

// TestGetStatusNilReceiver verifies GetStatus returns "" for a nil receiver.
//
// Why it matters: a nil cable must report no status rather than panic when its
// lifecycle state is queried.
// Inputs: a nil *CaniCableType. Outputs: the status string, expected "".
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestGetStatusNilReceiver(t *testing.T) {
	var c *CaniCableType
	if c.GetStatus() != "" {
		t.Errorf("GetStatus() on nil = %q, want empty", c.GetStatus())
	}
}

// --- SetTerminations ---

// TestSetTerminationsSetsUUIDs verifies SetTerminations records both endpoint
// interface IDs on the cable.
//
// Why it matters: terminations define which interfaces a cable connects, the core
// of the connection model.
// Inputs: two generated interface UUIDs passed as A and B. Outputs: the cable's
// TerminationA and TerminationB fields, expected to equal the inputs.
// Data choice: two distinct UUIDs make it clear each endpoint is stored in its
// own field without transposition.
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

// TestSetTerminationsNilReceiver verifies SetTerminations is a no-op on a nil
// receiver and does not panic.
//
// Why it matters: callers may set terminations on a missing cable, so a nil
// receiver must be handled gracefully.
// Inputs: a nil *CaniCableType and two generated UUIDs. Outputs: none; the test
// passes if the call returns without panicking.
// Data choice: a nil receiver is the only input that reaches the nil guard.
func TestSetTerminationsNilReceiver(t *testing.T) {
	var c *CaniCableType
	// Should not panic
	c.SetTerminations(uuid.New(), uuid.New())
}

// --- SetDeviceTerminations ---

// TestSetDeviceTerminationsSetsFields verifies SetDeviceTerminations records both
// device IDs and both port names on the cable.
//
// Why it matters: device-and-port terminations capture exactly where a cable
// plugs in, which export and topology rendering depend on.
// Inputs: two device UUIDs and ports "eth0"/"eth1". Outputs: the four termination
// fields, each expected to equal its input.
// Data choice: distinct devices and named ports confirm all four fields are set
// independently and in the right slots.
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

// TestSetDeviceTerminationsNilReceiver verifies SetDeviceTerminations is a no-op
// on a nil receiver and does not panic.
//
// Why it matters: callers may set device terminations on a missing cable, so a
// nil receiver must be handled gracefully.
// Inputs: a nil *CaniCableType, two UUIDs, and ports "p0"/"p1". Outputs: none; the
// test passes if the call returns without panicking.
// Data choice: a nil receiver is the only input that reaches the nil guard.
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
	inv.Interfaces[ifaceAID] = &CaniInterface{
		ID:       ifaceAID,
		DeviceID: devAID,
	}
	inv.Interfaces[ifaceBID] = &CaniInterface{
		ID:       ifaceBID,
		DeviceID: devBID,
	}

	cable := NewCable("cat6a", "Cat6a Patch")
	cable.SetTerminations(ifaceAID, ifaceBID)
	inv.Cables[cable.ID] = cable

	return cable, inv
}

// TestValidateCableHappyPath verifies ValidateCable accepts a cable whose
// terminations resolve to compatible, unused interfaces in the inventory.
//
// Why it matters: ValidateCable is the inventory-aware gate, so a correctly wired
// cable must pass before it is persisted.
// Inputs: a cable and inventory from newTestInventoryWithCable using matching
// 1000base-t interfaces. Outputs: an error, nil expected.
// Data choice: both endpoints share one interface type, the simplest setup that
// satisfies the existence and compatibility checks.
func TestValidateCableHappyPath(t *testing.T) {
	cable, inv := newTestInventoryWithCable(InterfacesElemTypeA1000BaseT)
	if err := ValidateCable(cable, inv); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestValidateCableNilCable verifies ValidateCable returns a descriptive error
// when the cable is nil.
//
// Why it matters: validating a nil cable is a programming error that must fail
// with a clear message rather than panic.
// Inputs: a nil cable and an empty inventory. Outputs: an error equal to "cable is
// nil".
// Data choice: asserting the exact message confirms the nil-cable guard fired
// ahead of any inventory lookup.
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

// TestAreInterfacesCompatibleSameType verifies areInterfacesCompatible accepts two
// identical interface types.
//
// Why it matters: a cable's endpoints must be electrically compatible, and the
// same type on both ends is the clearest compatible case.
// Inputs: 100gbase-x-qsfp28 on both sides. Outputs: the compatibility boolean,
// expected true.
// Data choice: an identical high-speed QSFP28 type exercises the same-type match
// without relying on cross-type grouping.
func TestAreInterfacesCompatibleSameType(t *testing.T) {
	if !areInterfacesCompatible(InterfacesElemTypeA100GbaseXQsfp28, InterfacesElemTypeA100GbaseXQsfp28) {
		t.Error("same interface type should be compatible")
	}
}

// TestAreInterfacesCompatibleIncompatible verifies areInterfacesCompatible rejects
// two interface types from different speed/media groups.
//
// Why it matters: connecting incompatible endpoints would model an impossible
// link, so the check must reject mismatched types.
// Inputs: 1000base-t against 100gbase-x-qsfp28. Outputs: the compatibility
// boolean, expected false.
// Data choice: a copper 1GbE type versus an optical 100GbE type are in distinct
// groups, driving the negative path.
func TestAreInterfacesCompatibleIncompatible(t *testing.T) {
	if areInterfacesCompatible(InterfacesElemTypeA1000BaseT, InterfacesElemTypeA100GbaseXQsfp28) {
		t.Error("1000base-t and 100gbase-x-qsfp28 should not be compatible")
	}
}

// ========== additional edge-case / branch coverage ==========

// --- ValidateCable (error paths) ---

// TestValidateCableErrorPaths verifies ValidateCable rejects a cable for each
// distinct failure: a nil inventory, a missing termination A or B interface,
// incompatible endpoint types, and an endpoint already wired to a different
// cable.
//
// Why it matters: ValidateCable is the integrity gate before a cable joins the
// inventory wiring graph, so every precondition it enforces must produce a clear,
// distinguishable error rather than a silent accept or a panic.
// Inputs: per case, a *CaniCableType and *Inventory derived from a fully valid
// fixture with exactly one precondition broken. Outputs: an error whose message
// contains the case's expected substring.
// Data choice: each case starts from the shared two-device/two-interface helper
// (a known-valid wiring) and breaks a single guard — deleting an interface from
// the index, retyping one endpoint to an incompatible 100G type, or pointing an
// endpoint at a different random cable UUID — so the asserted message is
// unambiguously attributable to that one guard.
func TestValidateCableErrorPaths(t *testing.T) {
	cases := []struct {
		name    string
		setup   func() (*CaniCableType, *Inventory)
		wantErr string
	}{
		{
			name: "nil inventory",
			setup: func() (*CaniCableType, *Inventory) {
				return NewCable("cat6a", "Cat6a"), nil
			},
			wantErr: "inventory is nil",
		},
		{
			name: "termination A interface missing",
			setup: func() (*CaniCableType, *Inventory) {
				cable, inv := newTestInventoryWithCable(InterfacesElemTypeA1000BaseT)
				delete(inv.Interfaces, cable.TerminationA)
				return cable, inv
			},
			wantErr: "termination A interface",
		},
		{
			name: "termination B interface missing",
			setup: func() (*CaniCableType, *Inventory) {
				cable, inv := newTestInventoryWithCable(InterfacesElemTypeA1000BaseT)
				delete(inv.Interfaces, cable.TerminationB)
				return cable, inv
			},
			wantErr: "termination B interface",
		},
		{
			name: "interface type mismatch",
			setup: func() (*CaniCableType, *Inventory) {
				cable, inv := newTestInventoryWithCable(InterfacesElemTypeA1000BaseT)
				specB, _ := inv.GetInterfaceByID(cable.TerminationB)
				specB.Type = InterfacesElemTypeA100GbaseXQsfp28
				return cable, inv
			},
			wantErr: "interface type mismatch",
		},
		{
			name: "termination A already connected",
			setup: func() (*CaniCableType, *Inventory) {
				cable, inv := newTestInventoryWithCable(InterfacesElemTypeA1000BaseT)
				other := uuid.New()
				specA, _ := inv.GetInterfaceByID(cable.TerminationA)
				specA.ConnectedCable = &other
				return cable, inv
			},
			wantErr: "already connected to another cable",
		},
		{
			name: "termination B already connected",
			setup: func() (*CaniCableType, *Inventory) {
				cable, inv := newTestInventoryWithCable(InterfacesElemTypeA1000BaseT)
				other := uuid.New()
				specB, _ := inv.GetInterfaceByID(cable.TerminationB)
				specB.ConnectedCable = &other
				return cable, inv
			},
			wantErr: "already connected to another cable",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cable, inv := tt.setup()
			err := ValidateCable(cable, inv)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// --- ValidateCable (already wired to this cable) ---

// TestValidateCableAllowsSameCableTerminations verifies ValidateCable accepts a
// cable whose endpoints are already connected to that very cable.
//
// Why it matters: re-validating an already-wired cable must be idempotent — an
// endpoint pointing back at this same cable is the wired state, not a conflict,
// so it must pass rather than report a double-connection error.
// Inputs: a valid cable and inventory whose two endpoint specs have ConnectedCable
// set to the cable's own ID. Outputs: an error, nil expected.
// Data choice: setting ConnectedCable to cable.ID (not a random UUID) drives the
// second operand of the already-connected guard and makes it false, the path the
// error-path cases (which use a different UUID) cannot reach.
func TestValidateCableAllowsSameCableTerminations(t *testing.T) {
	cable, inv := newTestInventoryWithCable(InterfacesElemTypeA1000BaseT)
	id := cable.ID
	specA, _ := inv.GetInterfaceByID(cable.TerminationA)
	specB, _ := inv.GetInterfaceByID(cable.TerminationB)
	specA.ConnectedCable = &id
	specB.ConnectedCable = &id
	if err := ValidateCable(cable, inv); err != nil {
		t.Errorf("expected nil for endpoints wired to this same cable, got %v", err)
	}
}

// --- areInterfacesCompatible (compatible groups) ---

// TestAreInterfacesCompatibleWithinGroup verifies areInterfacesCompatible returns
// true for two different interface types that belong to the same compatibility
// group.
//
// Why it matters: cabling allows endpoints of distinct-but-compatible media (e.g.
// 1GbE copper variants), so the compatibility check must accept cross-type pairs
// within a group, not only identical types.
// Inputs: per case, two distinct InterfacesElemType values from one group.
// Outputs: the boolean result, expected true.
// Data choice: 1000BaseT/1000BaseKx and 10GbaseXSfpp/10GbaseT are the two
// configured groups; using distinct members of each forces the aInGroup &&
// bInGroup branch that the same-type and fully-incompatible tests skip.
func TestAreInterfacesCompatibleWithinGroup(t *testing.T) {
	cases := []struct {
		name string
		a, b InterfacesElemType
	}{
		{"1GbE copper group", InterfacesElemTypeA1000BaseT, InterfacesElemTypeA1000BaseKx},
		{"10GbE group", InterfacesElemTypeA10GbaseXSfpp, InterfacesElemTypeA10GbaseT},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if !areInterfacesCompatible(tt.a, tt.b) {
				t.Errorf("areInterfacesCompatible(%s, %s) = false, want true", tt.a, tt.b)
			}
		})
	}
}
