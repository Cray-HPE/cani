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
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestRebuildReidentifiesSharedInterfaceIDs verifies that when two owners
// persist the same interface ID, the rebuild keeps the ID on the first owner
// in byte order, gives the other a fresh ID, repoints that owner's cable end,
// and reports the repair.
//
// Why it matters: datastores written before per-instance copies exist carry
// shared IDs; the index must stop collapsing them without breaking cables.
// Inputs: two devices sharing one spec ID, a cable from the second device's
// port to a third device. Outputs: three indexed interfaces, the cable end on
// the second device moved to its new ID, and a "re-identified" fix message.
// Data choice: device IDs are chosen so the keeper is unambiguous, and the
// cable sits on the loser to prove rebinding rather than the no-op case.
func TestRebuildReidentifiesSharedInterfaceIDs(t *testing.T) {
	keeperID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	loserID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	peerID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	sharedID := uuid.New()
	peerIfaceID := uuid.New()

	inv := NewInventory()
	inv.Devices[keeperID] = &CaniDeviceType{ID: keeperID, Name: "keeper",
		Interfaces: []InterfaceSpec{{ID: sharedID, Name: "eth0", Type: InterfacesElemTypeA1000BaseT}}}
	inv.Devices[loserID] = &CaniDeviceType{ID: loserID, Name: "loser",
		Interfaces: []InterfaceSpec{{ID: sharedID, Name: "eth0", Type: InterfacesElemTypeA1000BaseT}}}
	inv.Devices[peerID] = &CaniDeviceType{ID: peerID, Name: "peer",
		Interfaces: []InterfaceSpec{{ID: peerIfaceID, Name: "eth0", Type: InterfacesElemTypeA1000BaseT}}}
	cable := &CaniCableType{ID: uuid.New(), Label: "c1",
		TerminationADevice: loserID, TerminationAPort: "eth0", TerminationA: sharedID,
		TerminationBDevice: peerID, TerminationBPort: "eth0", TerminationB: peerIfaceID}
	inv.Cables[cable.ID] = cable

	result := inv.RebuildDerivedState()

	if err := result.Err(); err != nil {
		t.Fatalf("unexpected relationship errors: %v", err)
	}
	if len(inv.Interfaces) != 3 {
		t.Fatalf("indexed %d interfaces, want 3", len(inv.Interfaces))
	}
	if inv.Devices[keeperID].Interfaces[0].ID != sharedID {
		t.Fatalf("keeper lost its ID: %s", inv.Devices[keeperID].Interfaces[0].ID)
	}
	loserIfaceID := inv.Devices[loserID].Interfaces[0].ID
	if loserIfaceID == sharedID || loserIfaceID == uuid.Nil {
		t.Fatalf("loser kept the shared ID: %s", loserIfaceID)
	}
	if cable.TerminationA != loserIfaceID {
		t.Fatalf("cable termination A = %s, want re-identified %s", cable.TerminationA, loserIfaceID)
	}
	if !slices.ContainsFunc(result.Fixed, func(fix string) bool { return strings.Contains(fix, "re-identified") }) {
		t.Fatalf("expected a re-identified fix message, got %v", result.Fixed)
	}
}

// TestRebuildReidentifiesModulePortSharedWithSibling verifies that two modules
// on one device whose specs share an ID end up with distinct indexed
// interfaces: a cable addressed to the second module follows its new ID, while
// a cable addressed through the parent device keeps resolving to the module
// that kept the ID.
//
// Why it matters: `cani add module --qty N` was the main producer of shared
// IDs, and cables to module ports may be recorded against the module or the
// parent device; neither may end up dangling or ambiguous after the repair.
// Inputs: one device, two modules sharing a spec ID, one cable per addressing
// style. Outputs: two indexed interfaces, no relationship errors, the module
// cable on the second module's new ID and the device cable on the shared ID.
// Data choice: the two cables cover both branches of the rebind condition.
func TestRebuildReidentifiesModulePortSharedWithSibling(t *testing.T) {
	deviceID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	firstModID := uuid.MustParse("00000000-0000-0000-0000-000000000011")
	secondModID := uuid.MustParse("00000000-0000-0000-0000-000000000012")
	sharedID := uuid.New()

	inv := NewInventory()
	inv.Devices[deviceID] = &CaniDeviceType{ID: deviceID, Name: "node"}
	inv.Modules[firstModID] = &CaniModuleType{ID: firstModID, Name: "nic-1", ParentDevice: deviceID,
		Interfaces: []InterfaceSpec{{ID: sharedID, Name: "Port 1", Type: InterfacesElemTypeA25GbaseXSfp28}}}
	inv.Modules[secondModID] = &CaniModuleType{ID: secondModID, Name: "nic-2", ParentDevice: deviceID,
		Interfaces: []InterfaceSpec{{ID: sharedID, Name: "Port 1", Type: InterfacesElemTypeA25GbaseXSfp28}}}
	moduleCable := &CaniCableType{ID: uuid.New(), Label: "via-module",
		TerminationADevice: secondModID, TerminationAPort: "Port 1", TerminationA: sharedID}
	deviceCable := &CaniCableType{ID: uuid.New(), Label: "via-device",
		TerminationADevice: deviceID, TerminationAPort: "Port 1", TerminationA: sharedID}
	inv.Cables[moduleCable.ID] = moduleCable
	inv.Cables[deviceCable.ID] = deviceCable

	result := inv.RebuildDerivedState()

	if err := result.Err(); err != nil {
		t.Fatalf("unexpected relationship errors: %v", err)
	}
	secondID := inv.Modules[secondModID].Interfaces[0].ID
	if inv.Modules[firstModID].Interfaces[0].ID != sharedID || secondID == sharedID {
		t.Fatalf("module port IDs not separated: first=%s second=%s",
			inv.Modules[firstModID].Interfaces[0].ID, secondID)
	}
	if len(inv.Interfaces) != 2 {
		t.Fatalf("indexed %d interfaces, want 2", len(inv.Interfaces))
	}
	if moduleCable.TerminationA != secondID {
		t.Fatalf("module cable termination A = %s, want %s", moduleCable.TerminationA, secondID)
	}
	if deviceCable.TerminationA != sharedID {
		t.Fatalf("device cable termination A = %s, want the kept ID %s", deviceCable.TerminationA, sharedID)
	}
}
