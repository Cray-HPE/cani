/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package resolve

import (
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// Device — exercises both resolution branches plus every findByName outcome
// -----------------------------------------------------------------------------

// TestDevice_ResolvesByUUID covers the happy path where the arg parses as a
// UUID that exists in the inventory.
func TestDevice_ResolvesByUUID(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id: {ID: id, Name: "server-1"},
		},
	}

	got, err := Device(inv, id.String())
	if err != nil {
		t.Fatalf("Device by UUID returned error: %v", err)
	}
	if got != id {
		t.Errorf("Device by UUID = %s, want %s", got, id)
	}
}

// TestDevice_ResolvesByNameCaseInsensitively covers name resolution and proves
// the match ignores case.
func TestDevice_ResolvesByNameCaseInsensitively(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id: {ID: id, Name: "Server-1"},
		},
	}

	got, err := Device(inv, "server-1")
	if err != nil {
		t.Fatalf("Device by name returned error: %v", err)
	}
	if got != id {
		t.Errorf("Device by name = %s, want %s", got, id)
	}
}

// TestDevice_UUIDNotInInventoryReturnsError covers a well-formed UUID that is
// absent from the inventory.
func TestDevice_UUIDNotInInventoryReturnsError(t *testing.T) {
	missing := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
	}

	got, err := Device(inv, missing.String())
	if err == nil {
		t.Fatal("expected an error for a UUID not in inventory")
	}
	if got != uuid.Nil {
		t.Errorf("expected uuid.Nil on error, got %s", got)
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error %q should mention 'not found'", err)
	}
}

// TestDevice_NameNotFoundReturnsError covers the zero-match name branch.
func TestDevice_NameNotFoundReturnsError(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id: {ID: id, Name: "server-1"},
		},
	}

	_, err := Device(inv, "does-not-exist")
	if err == nil {
		t.Fatal("expected an error when no name matches")
	}
	if !strings.Contains(err.Error(), "no item found matching") {
		t.Errorf("error %q should mention 'no item found matching'", err)
	}
}

// TestDevice_AmbiguousNameReturnsError covers the multiple-match branch, which
// must refuse to guess and instead instruct the caller to use a UUID.
func TestDevice_AmbiguousNameReturnsError(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id1: {ID: id1, Name: "dup"},
			id2: {ID: id2, Name: "dup"},
		},
	}

	_, err := Device(inv, "dup")
	if err == nil {
		t.Fatal("expected an error for an ambiguous name")
	}
	if !strings.Contains(err.Error(), "multiple items match") {
		t.Errorf("error %q should mention 'multiple items match'", err)
	}
}

// -----------------------------------------------------------------------------
// Remaining resolvers — one representative path each so all are covered
// -----------------------------------------------------------------------------

// TestLocation_ResolvesByUUIDAndName checks both branches of the location
// resolver.
func TestLocation_ResolvesByUUIDAndName(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			id: {ID: id, Name: "hpc-dc"},
		},
	}

	byID, err := Location(inv, id.String())
	if err != nil || byID != id {
		t.Fatalf("Location by UUID = (%s, %v), want (%s, nil)", byID, err, id)
	}

	byName, err := Location(inv, "HPC-DC")
	if err != nil || byName != id {
		t.Fatalf("Location by name = (%s, %v), want (%s, nil)", byName, err, id)
	}
}

// TestRack_ResolvesByName checks the rack resolver's name branch.
func TestRack_ResolvesByName(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			id: {ID: id, Name: "rack-a1"},
		},
	}

	got, err := Rack(inv, "rack-a1")
	if err != nil || got != id {
		t.Fatalf("Rack by name = (%s, %v), want (%s, nil)", got, err, id)
	}
}

// TestModule_ResolvesByUUID checks the module resolver's UUID branch.
func TestModule_ResolvesByUUID(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Modules: map[uuid.UUID]*devicetypes.CaniModuleType{
			id: {ID: id, Name: "nic-0"},
		},
	}

	got, err := Module(inv, id.String())
	if err != nil || got != id {
		t.Fatalf("Module by UUID = (%s, %v), want (%s, nil)", got, err, id)
	}
}

// TestCable_ResolvesByLabel checks the cable resolver, which matches on Label
// rather than Name.
func TestCable_ResolvesByLabel(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Cables: map[uuid.UUID]*devicetypes.CaniCableType{
			id: {ID: id, Label: "dac-12"},
		},
	}

	got, err := Cable(inv, "dac-12")
	if err != nil || got != id {
		t.Fatalf("Cable by label = (%s, %v), want (%s, nil)", got, err, id)
	}
}

// TestLocation_UUIDNotInInventoryReturnsError covers the location resolver's
// UUID-miss branch.
func TestLocation_UUIDNotInInventoryReturnsError(t *testing.T) {
	inv := &devicetypes.Inventory{Locations: map[uuid.UUID]*devicetypes.CaniLocationType{}}
	if _, err := Location(inv, uuid.New().String()); err == nil {
		t.Fatal("expected an error for a location UUID not in inventory")
	}
}

// TestRack_ResolvesByUUIDAndReportsMissing covers both of the rack resolver's
// UUID branches: a hit returns the ID, a miss returns an error.
func TestRack_ResolvesByUUIDAndReportsMissing(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			id: {ID: id, Name: "rack-a1"},
		},
	}

	got, err := Rack(inv, id.String())
	if err != nil || got != id {
		t.Fatalf("Rack by UUID = (%s, %v), want (%s, nil)", got, err, id)
	}
	if _, err := Rack(inv, uuid.New().String()); err == nil {
		t.Error("expected an error for a rack UUID not in inventory")
	}
}

// TestModule_ResolvesByNameAndReportsMissing covers the module resolver's name
// branch (and the moduleNames helper) for both a match and a miss.
func TestModule_ResolvesByNameAndReportsMissing(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Modules: map[uuid.UUID]*devicetypes.CaniModuleType{
			id: {ID: id, Name: "nic-0"},
		},
	}

	got, err := Module(inv, "nic-0")
	if err != nil || got != id {
		t.Fatalf("Module by name = (%s, %v), want (%s, nil)", got, err, id)
	}
	if _, err := Module(inv, "absent"); err == nil {
		t.Error("expected an error when no module name matches")
	}
}

// TestModule_ResolvesBySerialAndAssetTag verifies a module is found by its
// serial number or asset tag, ignoring case, so a stock of identical spares
// that share the model as their name can each be addressed by the identity
// printed on the card; and that an empty serial or asset tag never matches.
//
// Why it matters: `add module --spare` without --name gives every spare of a
// type the same name, and Nautobot identifies a module by type, serial and
// asset tag rather than a name.
// Inputs: two spares named for their model, one with serial CX4000002 and
// asset tag AT-7, one with no identity; lookups by "cx4000002", "at-7",
// "" and the shared name.
// Outputs: the identified spare for both identity values; an error for the
// empty value; the duplicate-name ambiguity error for the shared name.
func TestModule_ResolvesBySerialAndAssetTag(t *testing.T) {
	tagged, blank := uuid.New(), uuid.New()
	inv := &devicetypes.Inventory{
		Modules: map[uuid.UUID]*devicetypes.CaniModuleType{
			tagged: {ID: tagged, Name: "ConnectX-4", Serial: "CX4000002", AssetTag: "AT-7"},
			blank:  {ID: blank, Name: "ConnectX-4"},
		},
	}
	for _, arg := range []string{"cx4000002", "AT-7"} {
		if got, err := Module(inv, arg); err != nil || got != tagged {
			t.Errorf("Module(%q) = (%s, %v), want the tagged spare %s", arg, got, err, tagged)
		}
	}
	if _, err := Module(inv, ""); err == nil {
		t.Error("expected an error for an empty value; blank identity fields must not match")
	}
	if _, err := Module(inv, "connectx-4"); err == nil || !strings.Contains(err.Error(), "multiple items match") {
		t.Errorf("shared name error = %v, want the ambiguity report", err)
	}
}

// TestModule_NameEqualToAnotherSerialIsAmbiguous verifies a value that is
// one module's name and another module's serial is reported as ambiguous,
// listing both with their identity values, rather than resolved to either.
//
// Why it matters: a first-field-wins match would let a name silently shadow
// a serial, and the operator would act on the wrong card.
// Inputs: module "A1" and a module whose serial is "a1".
// Outputs: an error naming both UUIDs and the serial.
func TestModule_NameEqualToAnotherSerialIsAmbiguous(t *testing.T) {
	named, serialled := uuid.New(), uuid.New()
	inv := &devicetypes.Inventory{
		Modules: map[uuid.UUID]*devicetypes.CaniModuleType{
			named:     {ID: named, Name: "A1"},
			serialled: {ID: serialled, Name: "nic", Serial: "a1"},
		},
	}
	_, err := Module(inv, "A1")
	if err == nil {
		t.Fatal("expected an ambiguity error")
	}
	for _, want := range []string{named.String(), serialled.String(), "a1", "use a UUID"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

// TestDevice_DoesNotResolveBySerial verifies the identity match is scoped to
// modules: a device's serial is not a device identifier.
//
// Why it matters: device names are the identity cani and Nautobot share, and
// widening devices would add an ambiguity surface without a workflow.
// Inputs: a device with serial "SN-1" looked up by "SN-1".
// Outputs: a not-found error.
func TestDevice_DoesNotResolveBySerial(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{id: {ID: id, Name: "cn-01", Serial: "SN-1"}},
	}
	if _, err := Device(inv, "SN-1"); err == nil {
		t.Error("expected a device serial not to resolve")
	}
}

// TestCable_ResolvesByUUIDAndReportsMissing covers both of the cable resolver's
// UUID branches.
func TestCable_ResolvesByUUIDAndReportsMissing(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Cables: map[uuid.UUID]*devicetypes.CaniCableType{
			id: {ID: id, Label: "dac-12"},
		},
	}

	got, err := Cable(inv, id.String())
	if err != nil || got != id {
		t.Fatalf("Cable by UUID = (%s, %v), want (%s, nil)", got, err, id)
	}
	if _, err := Cable(inv, uuid.New().String()); err == nil {
		t.Error("expected an error for a cable UUID not in inventory")
	}
}
