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
package devicetypes

import (
	"testing"

	"github.com/google/uuid"
)

// TestCloneInterfaceSpecsReturnsIndependentCopy verifies CloneInterfaceSpecs
// returns a slice with its own backing array whose elements equal the source.
//
// Why it matters: device and module instances are shallow copies of a shared
// type template, so they must not alias the template's interface specs; the
// inventory relies on this clone so the interface IDs assigned lazily during a
// rebuild do not collide across instances created in one process.
// Inputs: a two-element spec slice. Outputs: a clone whose mutation is asserted
// not to affect the source.
// Data choice: two specs with distinct names is the smallest input that reveals
// both element copying and backing-array independence.
func TestCloneInterfaceSpecsReturnsIndependentCopy(t *testing.T) {
	src := []InterfaceSpec{{Name: "mgmt"}, {Name: "eth0"}}
	got := CloneInterfaceSpecs(src)
	if len(got) != len(src) {
		t.Fatalf("len = %d, want %d", len(got), len(src))
	}
	got[0].ID = uuid.New()
	if src[0].ID != uuid.Nil {
		t.Errorf("mutating clone changed source: src[0].ID = %v, want Nil", src[0].ID)
	}
}

// TestCloneInterfaceSpecsNilReturnsNil verifies a nil input yields nil.
//
// Why it matters: a device type with no interfaces must clone to no interfaces
// rather than an empty-but-non-nil slice, preserving JSON omitempty behavior.
// Inputs: a nil slice. Outputs: a nil slice.
// Data choice: nil is the boundary case distinct from an empty slice.
func TestCloneInterfaceSpecsNilReturnsNil(t *testing.T) {
	if got := CloneInterfaceSpecs(nil); got != nil {
		t.Errorf("CloneInterfaceSpecs(nil) = %v, want nil", got)
	}
}

// TestAddDevicesIndependentInterfaceSpecsNoCollision verifies that adding two
// devices that share one interface-spec backing array yields disjoint, fully
// populated interface sets per device.
//
// Why it matters: instances of one hardware type are shallow copies of a single
// registry template and therefore share an Interfaces slice; without the clone
// in AddDevices the IDs assigned during the relationship rebuild collide and one
// device's interfaces disappear from the index — the exact defect that broke
// batch runs adding many same-type devices in a single process.
// Inputs: two devices sharing a two-spec slice. Outputs: inv.Interfaces split
// 2/2 across the devices with disjoint IDs.
// Data choice: two specs and two devices are the minimum needed to force an ID
// collision when the backing array is shared.
func TestAddDevicesIndependentInterfaceSpecsNoCollision(t *testing.T) {
	shared := []InterfaceSpec{{Name: "mgmt"}, {Name: "eth0"}}
	id1, id2 := uuid.New(), uuid.New()
	d1 := &CaniDeviceType{ID: id1, Name: "dev1", Model: "M", Interfaces: shared}
	d2 := &CaniDeviceType{ID: id2, Name: "dev2", Model: "M", Interfaces: shared}

	inv := NewInventory()
	if err := inv.AddDevices(map[uuid.UUID]*CaniDeviceType{id1: d1, id2: d2}); err != nil {
		t.Fatalf("AddDevices: %v", err)
	}

	got1 := inv.GetInterfacesByDevice(id1)
	got2 := inv.GetInterfacesByDevice(id2)
	if len(got1) != 2 || len(got2) != 2 {
		t.Fatalf("interfaces per device = %d and %d, want 2 and 2", len(got1), len(got2))
	}

	seen := map[uuid.UUID]bool{}
	for _, iface := range got1 {
		seen[iface.ID] = true
	}
	for _, iface := range got2 {
		if seen[iface.ID] {
			t.Errorf("interface ID %s shared between two devices", iface.ID)
		}
	}
}

// TestAddModuleIndependentInterfaceSpecsNoCollision verifies two modules of one
// type that share an interface-spec backing array end up with disjoint specs.
//
// Why it matters: modules (e.g. NICs) are instantiated the same shallow-copy way
// as devices, and the maple inventory adds many identical ConnectX modules in a
// single batch process; AddModule must clone their interface specs so their
// lazily-assigned IDs do not collide.
// Inputs: two modules sharing a one-spec slice. Outputs: each module's spec gets
// a distinct ID after add.
// Data choice: a single shared spec is the minimum that can collide.
func TestAddModuleIndependentInterfaceSpecsNoCollision(t *testing.T) {
	shared := []InterfaceSpec{{Name: "HSN 0"}}
	inv := NewInventory()
	m1 := &CaniModuleType{ID: uuid.New(), Name: "nic1", Model: "CX", Interfaces: shared}
	m2 := &CaniModuleType{ID: uuid.New(), Name: "nic2", Model: "CX", Interfaces: shared}
	if err := inv.AddModule(m1); err != nil {
		t.Fatalf("AddModule m1: %v", err)
	}
	if err := inv.AddModule(m2); err != nil {
		t.Fatalf("AddModule m2: %v", err)
	}
	if m1.Interfaces[0].ID == uuid.Nil || m2.Interfaces[0].ID == uuid.Nil {
		t.Fatalf("interface IDs not assigned: m1=%v m2=%v", m1.Interfaces[0].ID, m2.Interfaces[0].ID)
	}
	if m1.Interfaces[0].ID == m2.Interfaces[0].ID {
		t.Errorf("modules share interface ID %s; specs were not cloned", m1.Interfaces[0].ID)
	}
}
