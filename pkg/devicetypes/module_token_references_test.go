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
	"slices"
	"testing"

	"github.com/google/uuid"
)

func connectedTokenModule(t *testing.T) (*Inventory, *CaniModuleType, *CaniCableType) {
	t.Helper()
	inv, deviceID := tokenInventory()
	mod := customTokenModule(t, deviceID)
	if err := inv.AddModule(mod); err != nil {
		t.Fatal(err)
	}
	cable := &CaniCableType{
		ID: uuid.New(), Slug: "test-cable",
		TerminationA: mod.Interfaces[0].ID, TerminationADevice: mod.ID, TerminationAPort: mod.Interfaces[0].Name,
		TerminationB: mod.Interfaces[1].ID, TerminationBDevice: deviceID, TerminationBPort: mod.Interfaces[1].Name,
	}
	if err := inv.AddCable(cable); err != nil {
		t.Fatal(err)
	}
	return inv, mod, cable
}

func assertTokenCableReferences(t *testing.T, inv *Inventory, mod *CaniModuleType, cable *CaniCableType) {
	t.Helper()
	if cable.TerminationA != mod.Interfaces[0].ID || cable.TerminationAPort != mod.Interfaces[0].Name {
		t.Errorf("termination A = %s/%q, want %s/%q",
			cable.TerminationA, cable.TerminationAPort, mod.Interfaces[0].ID, mod.Interfaces[0].Name)
	}
	if cable.TerminationB != mod.Interfaces[1].ID || cable.TerminationBPort != mod.Interfaces[1].Name {
		t.Errorf("termination B = %s/%q, want %s/%q",
			cable.TerminationB, cable.TerminationBPort, mod.Interfaces[1].ID, mod.Interfaces[1].Name)
	}
	for _, spec := range mod.Interfaces {
		indexed := inv.Interfaces[spec.ID]
		if indexed == nil || indexed.Name != spec.Name || indexed.Label != spec.Label {
			t.Errorf("indexed interface = %+v, want name %q and label %q", indexed, spec.Name, spec.Label)
		}
	}
}

func TestResolveModuleNamesRenamesCablePorts(t *testing.T) {
	inv, mod, cable := connectedTokenModule(t)
	indexed := inv.Interfaces[mod.Interfaces[0].ID]
	addressID := uuid.New()
	indexed.IPAddresses = []uuid.UUID{addressID}

	mod.ModuleBayName = "Slot 1"
	inv.ResolveModuleNames(mod)

	assertTokenCableReferences(t, inv, mod, cable)
	assertCustomTemplatePort(t, mod, "1", indexed.ID)
	if inv.Interfaces[indexed.ID] != indexed || !slices.Equal(indexed.IPAddresses, []uuid.UUID{addressID}) {
		t.Error("rename replaced the indexed interface or discarded its derived IP assignments")
	}
	if cable.TerminationADevice != mod.ID || cable.TerminationBDevice != mod.ParentDevice {
		t.Error("rename changed endpoint ownership")
	}
	if err := inv.RebuildDerivedState().Err(); err != nil {
		t.Fatalf("rename left inconsistent cable references: %v", err)
	}
}

func TestResolveModuleNamesCablePortsFollowRepeatedMoves(t *testing.T) {
	inv, mod, cable := connectedTokenModule(t)
	cable.TerminationADevice, cable.TerminationBDevice = mod.ParentDevice, mod.ID
	firstID, secondID := cable.TerminationA, cable.TerminationB

	for _, move := range []struct{ bay, position string }{
		{"Slot 1", "1"}, {"PCIe1", "PCIe1"}, {"Slot 1", "1"},
	} {
		mod.ModuleBayName = move.bay
		inv.ResolveModuleNames(mod)
		assertTokenCableReferences(t, inv, mod, cable)
		assertCustomTemplatePort(t, mod, move.position, firstID)
		if cable.TerminationA != firstID || cable.TerminationB != secondID {
			t.Fatal("move changed cable interface UUIDs")
		}
	}
}

func TestResolveModuleNamesPreservesContradictoryCablePorts(t *testing.T) {
	inv, mod, cable := connectedTokenModule(t)
	differentID, differentOwner := *cable, *cable
	differentName, unresolved := *cable, *cable
	differentID.TerminationA, differentID.TerminationB = uuid.New(), uuid.New()
	differentOwner.TerminationADevice, differentOwner.TerminationBDevice = uuid.New(), uuid.New()
	differentName.TerminationAPort, differentName.TerminationBPort = "another-port", "another-port"
	unresolved.TerminationA, unresolved.TerminationB = uuid.Nil, uuid.Nil
	unchanged := []CaniCableType{differentID, differentOwner, differentName, unresolved}
	ids := make([]uuid.UUID, len(unchanged))
	for i := range unchanged {
		candidate := unchanged[i]
		candidate.ID = uuid.New()
		ids[i] = candidate.ID
		unchanged[i] = candidate
		inv.Cables[candidate.ID] = &candidate
	}
	inv.Cables[uuid.New()] = nil

	mod.ModuleBayName = "Slot 1"
	inv.ResolveModuleNames(mod)

	assertTokenCableReferences(t, inv, mod, cable)
	for i, id := range ids {
		if !reflect.DeepEqual(*inv.Cables[id], unchanged[i]) {
			t.Errorf("rename modified unrelated or contradictory cable: %+v", inv.Cables[id])
		}
	}
	if err := inv.RebuildDerivedState().Err(); err == nil {
		t.Error("contradictory cable references were no longer reported")
	}
}

func TestResolveModuleNamesLeavesSameNamedPeerPorts(t *testing.T) {
	inv, mod, cable := connectedTokenModule(t)
	peerID := uuid.New()
	peer := &CaniDeviceType{ID: peerID, Name: "peer", Interfaces: []InterfaceSpec{
		{ID: uuid.New(), Name: cable.TerminationAPort}, {ID: uuid.New(), Name: cable.TerminationBPort},
	}}
	inv.Devices[peerID] = peer
	peerCable := &CaniCableType{ID: uuid.New(), Slug: "test-cable",
		TerminationA: peer.Interfaces[0].ID, TerminationADevice: peerID, TerminationAPort: peer.Interfaces[0].Name,
		TerminationB: peer.Interfaces[1].ID, TerminationBDevice: peerID, TerminationBPort: peer.Interfaces[1].Name}
	if err := inv.AddCable(peerCable); err != nil {
		t.Fatal(err)
	}
	before := *peerCable

	mod.ModuleBayName = "Slot 1"
	inv.ResolveModuleNames(mod)

	assertTokenCableReferences(t, inv, mod, cable)
	if !reflect.DeepEqual(*peerCable, before) {
		t.Fatalf("rename modified same-named peer endpoints: %+v", peerCable)
	}
	if err := inv.RebuildDerivedState().Err(); err != nil {
		t.Fatalf("rename left inconsistent cable references: %v", err)
	}
}

func TestResolveModuleNamesDoesNotRenameReferencesForUnregisteredCopy(t *testing.T) {
	inv, mod, cable := connectedTokenModule(t)
	copy := *mod
	copy.ModuleBayName = "Slot 1"

	inv.ResolveModuleNames(&copy)

	if copy.Interfaces[0].Name != "custom-1/0" || mod.Interfaces[0].Name != "custom-PCIe1/0" {
		t.Fatal("resolution failed to isolate the copied module's interface specs")
	}
	assertTokenCableReferences(t, inv, mod, cable)
}

func TestResolveModuleNamesDoesNotMatchUnsetInterfaceIDs(t *testing.T) {
	inv, deviceID := tokenInventory()
	mod := tokenModule(deviceID, "PCIe1")
	inv.Modules[mod.ID] = mod
	cable := &CaniCableType{ID: uuid.New(),
		TerminationADevice: mod.ID, TerminationAPort: mod.Interfaces[0].Name,
		TerminationBDevice: deviceID, TerminationBPort: mod.Interfaces[1].Name}
	before := *cable
	inv.Cables[cable.ID] = cable

	inv.ResolveModuleNames(mod)

	if !reflect.DeepEqual(*cable, before) {
		t.Fatalf("unset interface IDs matched unresolved cable ends: %+v", cable)
	}
}

func TestResolveModuleNamesRefreshesLabelWithoutRenamingPort(t *testing.T) {
	inv, deviceID := tokenInventory()
	mod := customTokenModule(t, deviceID)
	mod.Interfaces[0].Name = "literal"
	if err := inv.AddModule(mod); err != nil {
		t.Fatal(err)
	}
	indexed := inv.Interfaces[mod.Interfaces[0].ID]

	mod.ModuleBayName = "Slot 1"
	inv.ResolveModuleNames(mod)

	if indexed.Name != "literal" || indexed.Label != "custom 1" {
		t.Fatalf("indexed interface not refreshed for label-only resolution: %+v", indexed)
	}
}
