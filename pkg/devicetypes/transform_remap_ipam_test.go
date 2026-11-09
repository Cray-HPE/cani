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
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestInterfaceRemap_FollowsMergedOwnersByPortName verifies that an incoming
// device and module resolved onto existing records map each of their ports
// to the same-named port the inventory kept, orphan a port the kept record
// lacks, and that an owner inserted as new contributes nothing.
//
// Why it matters: MergeProperties never replaces an existing record's
// interface specs, so an address assigned to an incoming port would point
// at an ID nothing owns; the remap is what lets the assignment survive a
// merge by name, and the orphan list is what lets the merge drop the
// assignments it cannot follow instead of refusing the import.
// Inputs: existing device d1 (eth0=A, then a second eth0=A2) and module m1
// (xe0=M); incoming d1 (eth0=B, bond0=O), m1 (xe0=N) and a new device d2
// (eth0=C); remaps resolving d1→d1 and m1→m1 with d2 kept as itself.
// Outputs: Remap {B→A, N→M}; Orphaned {O: port "bond0" on device "d1"};
// nothing for C.
// Data choice: one device, one module and one new owner cover the three
// branches of the walk; the duplicate kept eth0 pins first-wins, the rule
// FindInterfaceIDByPort applies.
func TestInterfaceRemap_FollowsMergedOwnersByPortName(t *testing.T) {
	a, a2, b, c, m, n, o := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	d1, m1, d2 := uuid.New(), uuid.New(), uuid.New()
	inv := NewInventory()
	inv.Devices[d1] = &CaniDeviceType{ID: d1, Name: "d1", Interfaces: []InterfaceSpec{{ID: a, Name: "eth0"}, {ID: a2, Name: "eth0"}}}
	inv.Modules[m1] = &CaniModuleType{ID: m1, Name: "m1", Interfaces: []InterfaceSpec{{ID: m, Name: "xe0"}}}

	incomingD1, incomingM1 := uuid.New(), uuid.New()
	newD2 := &CaniDeviceType{ID: d2, Name: "d2", Interfaces: []InterfaceSpec{{ID: c, Name: "eth0"}}}
	inv.Devices[d2] = newD2
	result := &TransformResult{
		Devices: map[uuid.UUID]*CaniDeviceType{
			incomingD1: {ID: incomingD1, Name: "d1", Interfaces: []InterfaceSpec{{ID: b, Name: "eth0"}, {ID: o, Name: "bond0"}}},
			d2:         newD2,
		},
		Modules: map[uuid.UUID]*CaniModuleType{
			incomingM1: {ID: incomingM1, Name: "m1", Interfaces: []InterfaceSpec{{ID: n, Name: "xe0"}}},
		},
	}
	remaps := ReferenceRemaps{
		Devices: map[uuid.UUID]uuid.UUID{incomingD1: d1, d2: d2},
		Modules: map[uuid.UUID]uuid.UUID{incomingM1: m1},
	}

	got := result.InterfaceRemap(inv, remaps)

	if want := map[uuid.UUID]uuid.UUID{b: a, n: m}; !reflect.DeepEqual(got.Remap, want) {
		t.Errorf("Remap = %v, want %v", got.Remap, want)
	}
	if want := map[uuid.UUID]string{o: `port "bond0" on device "d1"`}; !reflect.DeepEqual(got.Orphaned, want) {
		t.Errorf("Orphaned = %v, want %v", got.Orphaned, want)
	}
}

// TestMergeTransformResult_ImportedAssignmentFollowsTheMergedDevice drives
// the whole merge: an inventory authored locally holds device n1 with port
// eth0; an import brings the same device under a fresh UUID with its own
// port UUID and an address assigned to that port.
//
// Why it matters: this is the re-import onto a CLI-authored inventory the
// Story's seed path depends on; without the remap the address would keep
// the incoming port ID, the relationship validation would find no such
// interface, and the merge would refuse the whole import.
// Inputs: existing n1 (eth0 = kept); incoming n1 (eth0 = incoming) and
// 10.0.0.5/24 assigned to incoming, under a 10.0.0.0/24 prefix.
// Outputs: the merge succeeds, one device remains, and the address's
// Interfaces is [kept].
// Data choice: the device is matched by name alone, the weakest match the
// merge supports, so the remap cannot rely on a shared UUID.
func TestMergeTransformResult_ImportedAssignmentFollowsTheMergedDevice(t *testing.T) {
	kept, incoming := uuid.New(), uuid.New()
	inv := NewInventory()
	existingID := uuid.New()
	inv.Devices[existingID] = &CaniDeviceType{
		ID: existingID, Name: "n1", Interfaces: []InterfaceSpec{{ID: kept, Name: "eth0"}},
	}
	if res := inv.RebuildDerivedState(); res.Err() != nil {
		t.Fatalf("arrange: %v", res.Err())
	}

	prefix := &CaniPrefix{ID: uuid.New(), Prefix: "10.0.0.0/24"}
	address := &CaniIPAddress{ID: uuid.New(), Address: "10.0.0.5/24", Interfaces: []uuid.UUID{incoming}}
	requireAdded(t, ParsePrefix(prefix))
	requireAdded(t, ParseIPAddress(address))
	importedID := uuid.New()
	result := &TransformResult{
		Devices: map[uuid.UUID]*CaniDeviceType{importedID: {
			ID: importedID, Name: "n1", Interfaces: []InterfaceSpec{{ID: incoming, Name: "eth0"}},
		}},
		Prefixes:    map[uuid.UUID]*CaniPrefix{prefix.ID: prefix},
		IPAddresses: map[uuid.UUID]*CaniIPAddress{address.ID: address},
	}

	if _, err := inv.MergeTransformResult(result); err != nil {
		t.Fatalf("MergeTransformResult returned error: %v", err)
	}

	if len(inv.Devices) != 1 {
		t.Errorf("devices = %d, want 1 (merged by name)", len(inv.Devices))
	}
	merged := inv.IPAddresses[address.ID]
	if merged == nil {
		t.Fatal("address missing after merge")
	}
	if !reflect.DeepEqual(merged.Interfaces, []uuid.UUID{kept}) {
		t.Errorf("Interfaces = %v, want [%s] (the port the inventory kept)", merged.Interfaces, kept)
	}
	if iface := inv.Interfaces[kept]; iface == nil || !reflect.DeepEqual(iface.IPAddresses, []uuid.UUID{address.ID}) {
		t.Errorf("kept interface IPAddresses = %v, want [%s]", iface, address.ID)
	}
}

// TestMergeTransformResult_DropsAndReportsAnAssignmentToAPortTheKeptDeviceLacks
// verifies that an address assigned to a port the merged-onto record does not
// have loses that one assignment, keeps the rest, and is reported, while the
// import as a whole merges.
//
// Why it matters: a device merged by name keeps the existing record's ports,
// and a source often has ports the local template lacks (a bond, a loopback,
// a renamed port). Left in place, the assignment would point at an interface
// nothing owns and the relationship validation would refuse the entire
// import over one row.
// Inputs: existing n1 with eth0 (kept); incoming n1 with eth0 (incoming) and
// bond0 (orphan); 10.0.0.5/24 assigned to both incoming ports.
// Outputs: no error; the address's Interfaces is [kept]; the summary's
// Unresolved names the address, bond0 and n1 once.
// Data choice: one followed port beside one orphaned port on the same
// address shows the drop is per assignment, not per address.
func TestMergeTransformResult_DropsAndReportsAnAssignmentToAPortTheKeptDeviceLacks(t *testing.T) {
	kept, incoming, orphan := uuid.New(), uuid.New(), uuid.New()
	inv := NewInventory()
	existingID := uuid.New()
	inv.Devices[existingID] = &CaniDeviceType{
		ID: existingID, Name: "n1", Interfaces: []InterfaceSpec{{ID: kept, Name: "eth0"}},
	}
	if res := inv.RebuildDerivedState(); res.Err() != nil {
		t.Fatalf("arrange: %v", res.Err())
	}
	prefix := &CaniPrefix{ID: uuid.New(), Prefix: "10.0.0.0/24"}
	address := &CaniIPAddress{ID: uuid.New(), Address: "10.0.0.5/24", Interfaces: []uuid.UUID{incoming, orphan}}
	requireAdded(t, ParsePrefix(prefix))
	requireAdded(t, ParseIPAddress(address))
	importedID := uuid.New()
	result := &TransformResult{
		Devices: map[uuid.UUID]*CaniDeviceType{importedID: {
			ID: importedID, Name: "n1",
			Interfaces: []InterfaceSpec{{ID: incoming, Name: "eth0"}, {ID: orphan, Name: "bond0"}},
		}},
		Prefixes:    map[uuid.UUID]*CaniPrefix{prefix.ID: prefix},
		IPAddresses: map[uuid.UUID]*CaniIPAddress{address.ID: address},
	}

	summary, err := inv.MergeTransformResult(result)

	if err != nil {
		t.Fatalf("MergeTransformResult returned error: %v", err)
	}
	if merged := inv.IPAddresses[address.ID]; merged == nil || !reflect.DeepEqual(merged.Interfaces, []uuid.UUID{kept}) {
		t.Errorf("Interfaces = %v, want [%s] (the followed port only)", merged, kept)
	}
	if len(summary.Unresolved) != 1 {
		t.Fatalf("Unresolved = %q, want one entry for the dropped assignment", summary.Unresolved)
	}
	for _, want := range []string{"10.0.0.5/24", `"bond0"`, `"n1"`} {
		if !strings.Contains(summary.Unresolved[0], want) {
			t.Errorf("Unresolved[0] = %q, want it to name %s", summary.Unresolved[0], want)
		}
	}
}

// TestMergeTransformResult_CableFollowsTheMergedDevicesPorts verifies that a
// cable between two devices merged by name ends on the ports the inventory
// kept.
//
// Why it matters: the merge remaps a cable's device ends but its interface
// ends carry the source's port IDs; on a device merged onto an existing
// record those IDs belong to no interface, so the relationship validation
// refused the whole import of any source with a cable onto a locally
// authored inventory.
// Inputs: existing n1 and n2, each with eth0 (keptA, keptB); incoming n1 and
// n2 under fresh UUIDs with their own eth0 IDs and a cable between them.
// Outputs: no error, one cable, terminations keptA and keptB.
// Data choice: both ends on merged devices, so neither termination can
// survive without the interface remap.
func TestMergeTransformResult_CableFollowsTheMergedDevicesPorts(t *testing.T) {
	keptA, keptB, inA, inB := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv := NewInventory()
	n1, n2 := uuid.New(), uuid.New()
	inv.Devices[n1] = &CaniDeviceType{ID: n1, Name: "n1", Interfaces: []InterfaceSpec{{ID: keptA, Name: "eth0"}}}
	inv.Devices[n2] = &CaniDeviceType{ID: n2, Name: "n2", Interfaces: []InterfaceSpec{{ID: keptB, Name: "eth0"}}}
	if res := inv.RebuildDerivedState(); res.Err() != nil {
		t.Fatalf("arrange: %v", res.Err())
	}
	impN1, impN2, cableID := uuid.New(), uuid.New(), uuid.New()
	cable := &CaniCableType{ID: cableID}
	cable.SetTerminations(inA, inB)
	cable.SetDeviceTerminations(impN1, impN2, "eth0", "eth0")
	result := &TransformResult{
		Devices: map[uuid.UUID]*CaniDeviceType{
			impN1: {ID: impN1, Name: "n1", Interfaces: []InterfaceSpec{{ID: inA, Name: "eth0"}}},
			impN2: {ID: impN2, Name: "n2", Interfaces: []InterfaceSpec{{ID: inB, Name: "eth0"}}},
		},
		Cables: map[uuid.UUID]*CaniCableType{cableID: cable},
	}

	if _, err := inv.MergeTransformResult(result); err != nil {
		t.Fatalf("MergeTransformResult returned error: %v", err)
	}

	if len(inv.Cables) != 1 {
		t.Fatalf("cables = %d, want 1", len(inv.Cables))
	}
	for _, merged := range inv.Cables {
		if merged.TerminationA != keptA || merged.TerminationB != keptB {
			t.Errorf("terminations = %s/%s, want %s/%s (the kept ports)", merged.TerminationA, merged.TerminationB, keptA, keptB)
		}
	}
}
