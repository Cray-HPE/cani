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
package connections

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// applyTestInventory builds an inventory with two devices, each carrying a
// single named interface, and returns the device IDs so callers can craft
// ResolvedConnections that reference real interfaces for MAC assignment.
func applyTestInventory(t *testing.T) (*devicetypes.Inventory, uuid.UUID, uuid.UUID) {
	t.Helper()
	inv := devicetypes.NewInventory()
	aID := uuid.New()
	bID := uuid.New()
	inv.Devices[aID] = &devicetypes.CaniDeviceType{
		ID:   aID,
		Name: "node-a",
		Interfaces: []devicetypes.InterfaceSpec{
			{ID: uuid.New(), Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT},
		},
	}
	inv.Devices[bID] = &devicetypes.CaniDeviceType{
		ID:   bID,
		Name: "node-b",
		Interfaces: []devicetypes.InterfaceSpec{
			{ID: uuid.New(), Name: "eth1", Type: devicetypes.InterfacesElemTypeA1000BaseT},
		},
	}
	return inv, aID, bID
}

// singleCable asserts the inventory holds exactly one cable and returns it.
func singleCable(t *testing.T, inv *devicetypes.Inventory) *devicetypes.CaniCableType {
	t.Helper()
	if len(inv.Cables) != 1 {
		t.Fatalf("cables = %d, want 1", len(inv.Cables))
	}
	for _, c := range inv.Cables {
		return c
	}
	return nil
}

// TestApplyConnections_CreatesCableWithAllProps verifies ApplyConnections builds
// a cable from a resolved connection, copies every optional property, and writes
// both endpoint MACs onto the referenced interfaces.
//
// Why it matters: ApplyConnections is the final step that turns resolved
// connection intent into persisted cables and interface MACs, so each property
// and MAC must land on the right object.
// Inputs: one ResolvedConnection with Color/Length/LengthUnit/Status and valid A
// and B MACs against an inventory with matching interfaces. Outputs: a created
// count of 1, no errors, a stored cable carrying the props, and updated
// interface MAC addresses.
// Data choice: a fractional length and a non-default "Active" status prove the
// optional-property branches actually override NewCable's defaults rather than
// coinciding with them.
func TestApplyConnections_CreatesCableWithAllProps(t *testing.T) {
	inv, aID, bID := applyTestInventory(t)
	length := 2.5
	resolved := []ResolvedConnection{
		{
			ADevice: aID,
			APort:   "eth0",
			AMac:    "aa:bb:cc:dd:ee:01",
			BDevice: bID,
			BPort:   "eth1",
			BMac:    "aa:bb:cc:dd:ee:02",
			Cable: CableProps{
				Type:       "cat6a",
				Label:      "uplink",
				Color:      "blue",
				Length:     &length,
				LengthUnit: "m",
				Status:     "Active",
			},
		},
	}

	created, errs := ApplyConnections(resolved, inv)
	if created != 1 {
		t.Fatalf("created = %d, want 1", created)
	}
	if len(errs) != 0 {
		t.Fatalf("errors = %v, want none", errs)
	}

	cable := singleCable(t, inv)
	if cable.Slug != "cat6a" {
		t.Errorf("cable slug = %q, want cat6a", cable.Slug)
	}
	if cable.Label != "uplink" {
		t.Errorf("cable label = %q, want uplink", cable.Label)
	}
	if cable.Color != "blue" {
		t.Errorf("cable color = %q, want blue", cable.Color)
	}
	if cable.Length == nil || *cable.Length != 2.5 {
		t.Errorf("cable length = %v, want 2.5", cable.Length)
	}
	if cable.LengthUnit != "m" {
		t.Errorf("cable length_unit = %q, want m", cable.LengthUnit)
	}
	if cable.Status != "Active" {
		t.Errorf("cable status = %q, want Active", cable.Status)
	}
	if got := inv.Devices[aID].Interfaces[0].MacAddress; got != "aa:bb:cc:dd:ee:01" {
		t.Errorf("A interface mac = %q, want aa:bb:cc:dd:ee:01", got)
	}
	if got := inv.Devices[bID].Interfaces[0].MacAddress; got != "aa:bb:cc:dd:ee:02" {
		t.Errorf("B interface mac = %q, want aa:bb:cc:dd:ee:02", got)
	}
}

// TestApplyConnections_NoOptionalProps verifies ApplyConnections creates a cable
// when the resolved connection sets no optional properties and no MACs, leaving
// NewCable's defaults intact and touching no interface.
//
// Why it matters: a bare connection (just endpoints) is the common case, and the
// optional-property and MAC-assignment steps must be skipped cleanly rather than
// stamping empty values or erroring.
// Inputs: one ResolvedConnection with only Type/Label and empty Color, Length,
// LengthUnit, Status, and MACs. Outputs: a created count of 1, no errors, a
// stored cable keeping the default "Connected" status, and an unchanged
// interface MAC.
// Data choice: leaving every optional field zero drives the false side of each
// optional-property guard and the empty-MAC skip, the exact branches the
// all-props test does not reach.
func TestApplyConnections_NoOptionalProps(t *testing.T) {
	inv, aID, bID := applyTestInventory(t)
	resolved := []ResolvedConnection{
		{
			ADevice: aID,
			APort:   "eth0",
			BDevice: bID,
			BPort:   "eth1",
			Cable:   CableProps{Type: "cat6a", Label: "uplink"},
		},
	}

	created, errs := ApplyConnections(resolved, inv)
	if created != 1 {
		t.Fatalf("created = %d, want 1", created)
	}
	if len(errs) != 0 {
		t.Fatalf("errors = %v, want none", errs)
	}

	cable := singleCable(t, inv)
	if cable.Status != "Connected" {
		t.Errorf("cable status = %q, want default Connected", cable.Status)
	}
	if cable.Color != "" || cable.Length != nil || cable.LengthUnit != "" {
		t.Errorf("expected no optional props, got color=%q length=%v unit=%q", cable.Color, cable.Length, cable.LengthUnit)
	}
	if got := inv.Devices[aID].Interfaces[0].MacAddress; got != "" {
		t.Errorf("A interface mac = %q, want empty (no mac supplied)", got)
	}
}

// TestApplyConnections_MacErrorsCollected verifies ApplyConnections still creates
// the cable but collects one error per side when a MAC fails to apply.
//
// Why it matters: a malformed MAC must not abort the whole apply; the cable is
// recorded and the per-endpoint failures are surfaced so the caller can report
// every problem at once.
// Inputs: one ResolvedConnection with invalid A and B MAC strings. Outputs: a
// created count of 1 and a two-element error slice naming both ports.
// Data choice: "not-a-mac" fails net.ParseMAC for both endpoints, exercising the
// A-side and B-side MAC error branches together while the cable itself remains
// valid.
func TestApplyConnections_MacErrorsCollected(t *testing.T) {
	inv, aID, bID := applyTestInventory(t)
	resolved := []ResolvedConnection{
		{
			ADevice: aID,
			APort:   "eth0",
			AMac:    "not-a-mac",
			BDevice: bID,
			BPort:   "eth1",
			BMac:    "also-bad",
			Cable:   CableProps{Type: "cat6a"},
		},
	}

	created, errs := ApplyConnections(resolved, inv)
	if created != 1 {
		t.Fatalf("created = %d, want 1 (cable still added)", created)
	}
	if len(errs) != 2 {
		t.Fatalf("errors = %d, want 2 (one per bad mac)", len(errs))
	}
}
