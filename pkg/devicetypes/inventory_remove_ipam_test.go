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

// TestRemovePrefixMovesChildrenToItsParent verifies removing a mid-tree prefix
// hands its child prefixes and addresses to its own parent.
//
// Why it matters: Nautobot re-parents children on delete; orphaned addresses
// cannot be exported.
// Inputs: 10.0.0.0/8 holding 10.0.0.0/16 and a sibling 10.1.0.0/16; the
// first /16 holds 10.0.1.0/24 and 10.0.0.5.
// Outputs: the /24 and the address move to the /8; the sibling is untouched.
// Data choice: the removed prefix sits mid-tree, so moving up and orphaning
// leave different parents.
func TestRemovePrefixMovesChildrenToItsParent(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	root, removed, child, sibling, address := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Prefixes[root] = &CaniPrefix{ID: root, Prefix: "10.0.0.0/8"}
	inv.Prefixes[removed] = &CaniPrefix{ID: removed, Prefix: "10.0.0.0/16", Parent: root}
	inv.Prefixes[child] = &CaniPrefix{ID: child, Prefix: "10.0.1.0/24", Parent: removed}
	inv.Prefixes[sibling] = &CaniPrefix{ID: sibling, Prefix: "10.1.0.0/16", Parent: root}
	inv.IPAddresses[address] = &CaniIPAddress{ID: address, Address: "10.0.0.5/16", Parent: removed}

	// Act.
	err := inv.RemovePrefix(removed)

	// Assert.
	if err != nil {
		t.Fatalf("RemovePrefix returned error: %v", err)
	}
	if inv.Prefixes[removed] != nil {
		t.Error("removed prefix is still in the inventory")
	}
	if got := inv.Prefixes[child].Parent; got != root {
		t.Errorf("child prefix parent = %v, want the /8 %v", got, root)
	}
	if got := inv.IPAddresses[address].Parent; got != root {
		t.Errorf("address parent = %v, want the /8 %v", got, root)
	}
	if got := inv.Prefixes[sibling].Parent; got != root {
		t.Errorf("sibling parent = %v, want unchanged %v", got, root)
	}
	if err := inv.RebuildDerivedState().Err(); err != nil {
		t.Errorf("relationships invalid after removal: %v", err)
	}
}

// TestRemoveRootPrefixLetsChildPrefixesBecomeRoots verifies a root prefix
// that holds only prefixes can be removed.
//
// Why it matters: a prefix needs no parent, so nothing should block this.
// Inputs: a root 10.0.0.0/16 holding 10.0.1.0/24 and no addresses.
// Outputs: the /16 is gone and the /24 has no parent.
// Data choice: the counterpart of the address case below.
func TestRemoveRootPrefixLetsChildPrefixesBecomeRoots(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	root, child := uuid.New(), uuid.New()
	inv.Prefixes[root] = &CaniPrefix{ID: root, Prefix: "10.0.0.0/16"}
	inv.Prefixes[child] = &CaniPrefix{ID: child, Prefix: "10.0.1.0/24", Parent: root}

	// Act.
	err := inv.RemovePrefix(root)

	// Assert.
	if err != nil {
		t.Fatalf("RemovePrefix returned error: %v", err)
	}
	if inv.Prefixes[root] != nil || inv.Prefixes[child].Parent != uuid.Nil {
		t.Errorf("root kept = %v, child parent = %v; want root removed and child a root",
			inv.Prefixes[root] != nil, inv.Prefixes[child].Parent)
	}
}

// TestRemoveRootPrefixRefusesToOrphanAddresses verifies a root prefix that
// holds addresses is kept, and that the refusal changes nothing.
//
// Why it matters: Nautobot refuses the same delete, and an address with no
// parent prefix cannot be exported.
// Inputs: a root 10.0.0.0/16 holding 10.0.1.0/24 and 10.0.0.5.
// Outputs: an error naming one address; the /16 and the /24's parent remain.
// Data choice: the child prefix shows the refusal comes before any re-parenting.
func TestRemoveRootPrefixRefusesToOrphanAddresses(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	root, child, address := uuid.New(), uuid.New(), uuid.New()
	inv.Prefixes[root] = &CaniPrefix{ID: root, Prefix: "10.0.0.0/16"}
	inv.Prefixes[child] = &CaniPrefix{ID: child, Prefix: "10.0.1.0/24", Parent: root}
	inv.IPAddresses[address] = &CaniIPAddress{ID: address, Address: "10.0.0.5/16", Parent: root}

	// Act.
	err := inv.RemovePrefix(root)

	// Assert.
	if err == nil || !strings.Contains(err.Error(), "holds 1 IP address(es)") {
		t.Fatalf("RemovePrefix error = %v, want it to report the 1 held address", err)
	}
	if inv.Prefixes[root] == nil || inv.Prefixes[child].Parent != root || inv.IPAddresses[address].Parent != root {
		t.Error("a refused removal changed the inventory")
	}
}

// TestRemoveIPAddressClearsReferencesToIt verifies removing an address clears
// every reference to it and leaves references to other addresses alone.
//
// Why it matters: a dangling primary IP or NAT inside fails validation on the
// next load, and the derived interface index must be rebuilt rather than keep
// an address that is gone.
// Inputs: a device with the removed address as primary IPv4 and another as
// primary IPv6; an address whose NAT inside is the removed one; an interface
// assigned both, indexed before the removal.
// Outputs: the IPv4, the NAT inside and the interface entry are cleared; the
// IPv6 and the other assignment stay.
// Data choice: each reference has a neighbour that must survive, so clearing
// too much fails the test, and the pre-built index fails it if the rebuild
// only appends.
func TestRemoveIPAddressClearsReferencesToIt(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	device, iface, removed, kept, outside := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Devices[device] = &CaniDeviceType{ID: device, Name: "node", PrimaryIPv4: removed, PrimaryIPv6: kept,
		Interfaces: []InterfaceSpec{{ID: iface, Name: "eth0"}}}
	inv.IPAddresses[removed] = &CaniIPAddress{ID: removed, Address: "10.0.0.5/24", Interfaces: []uuid.UUID{iface}}
	inv.IPAddresses[kept] = &CaniIPAddress{ID: kept, Address: "fd00::5/64", Interfaces: []uuid.UUID{iface}}
	inv.IPAddresses[outside] = &CaniIPAddress{ID: outside, Address: "192.0.2.5/32", NATInside: removed}
	if err := inv.RebuildDerivedState().Err(); err != nil {
		t.Fatalf("invalid fixture: %v", err)
	}

	// Act.
	err := inv.RemoveIPAddress(removed)

	// Assert.
	if err != nil {
		t.Fatalf("RemoveIPAddress returned error: %v", err)
	}
	node := inv.Devices[device]
	if inv.IPAddresses[removed] != nil || node.PrimaryIPv4 != uuid.Nil || node.PrimaryIPv6 != kept {
		t.Errorf("address kept = %v, primary IPv4 = %v, IPv6 = %v; want removed, nil, %v",
			inv.IPAddresses[removed] != nil, node.PrimaryIPv4, node.PrimaryIPv6, kept)
	}
	if got := inv.IPAddresses[outside].NATInside; got != uuid.Nil {
		t.Errorf("NAT inside = %v, want cleared", got)
	}
	if got := inv.Interfaces[iface].IPAddresses; !reflect.DeepEqual(got, []uuid.UUID{kept}) {
		t.Errorf("interface addresses = %v, want only %v", got, kept)
	}
	if err := inv.RebuildDerivedState().Err(); err != nil {
		t.Errorf("relationships invalid after removal: %v", err)
	}
}

// TestRemoveVRFDetachesPrefixMemberships verifies removing a VRF drops it
// from every prefix and leaves the other memberships as they were.
//
// Why it matters: a membership naming a missing VRF fails validation and
// export.
// Inputs: tenant-a VRFs red and blue; one prefix in both, one in neither.
// Outputs: the first prefix keeps only blue; the second still has none.
// Data choice: the prefix without memberships shows untouched prefixes are not
// rewritten.
func TestRemoveVRFDetachesPrefixMemberships(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	red, blue, member, other := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.VRFs[red] = &CaniVRF{ID: red, Name: "red", Namespace: "tenant-a"}
	inv.VRFs[blue] = &CaniVRF{ID: blue, Name: "blue", Namespace: "tenant-a"}
	inv.Prefixes[member] = &CaniPrefix{ID: member, Prefix: "10.0.0.0/24", Namespace: "tenant-a", VRFs: []uuid.UUID{red, blue}}
	inv.Prefixes[other] = &CaniPrefix{ID: other, Prefix: "10.0.1.0/24", Namespace: "tenant-a"}

	// Act.
	err := inv.RemoveVRF(red)

	// Assert.
	if err != nil {
		t.Fatalf("RemoveVRF returned error: %v", err)
	}
	if inv.VRFs[red] != nil {
		t.Error("removed VRF is still in the inventory")
	}
	if got := inv.Prefixes[member].VRFs; !reflect.DeepEqual(got, []uuid.UUID{blue}) {
		t.Errorf("member VRFs = %v, want only blue %v", got, blue)
	}
	if got := inv.Prefixes[other].VRFs; got != nil {
		t.Errorf("untouched prefix VRFs = %#v, want nil", got)
	}
	if err := inv.RebuildDerivedState().Err(); err != nil {
		t.Errorf("relationships invalid after removal: %v", err)
	}
}

// TestRemoveIPAMRejectsUnknownIDs verifies each IPAM removal reports an
// unknown UUID instead of succeeding silently.
//
// Why it matters: the remove commands rely on the error to tell the operator
// that nothing was removed.
// Inputs: an empty inventory and a random UUID.
// Outputs: a "not found" error naming the kind of object.
// Data choice: an empty inventory holds nothing a lookup could match by mistake.
func TestRemoveIPAMRejectsUnknownIDs(t *testing.T) {
	inv := NewInventory()
	removals := map[string]func(uuid.UUID) error{
		"prefix": inv.RemovePrefix, "ip address": inv.RemoveIPAddress, "vrf": inv.RemoveVRF,
	}
	for kind, remove := range removals {
		t.Run(kind, func(t *testing.T) {
			// Act.
			err := remove(uuid.New())

			// Assert.
			if err == nil || !strings.HasPrefix(err.Error(), kind+" ") || !strings.HasSuffix(err.Error(), "not found") {
				t.Errorf("error = %v, want %q ... not found", err, kind)
			}
		})
	}
}

// TestRemovePrefixTreatsAMissingParentAsNone verifies a prefix whose parent
// record is gone is handled as a root: holding addresses, it is refused; holding
// only prefixes, they become roots rather than inherit the dangling UUID.
//
// Why it matters: the orphan refusal exists for damaged data, so a dangling
// parent must not be the way around it.
// Inputs: two prefixes whose Parent is a UUID not in the inventory; one holds
// an address, the other a child prefix.
// Outputs: the first removal is refused as holding 1 address; the second
// succeeds and leaves the child with no parent.
// Data choice: the same dangling UUID on both shows the rule, not the data,
// decides the outcome.
func TestRemovePrefixTreatsAMissingParentAsNone(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	dangling, withAddress, withChild, address, child := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Prefixes[withAddress] = &CaniPrefix{ID: withAddress, Prefix: "10.0.0.0/24", Parent: dangling}
	inv.IPAddresses[address] = &CaniIPAddress{ID: address, Host: "10.0.0.5", Address: "10.0.0.5/24", Parent: withAddress}
	inv.Prefixes[withChild] = &CaniPrefix{ID: withChild, Prefix: "10.1.0.0/16", Parent: dangling}
	inv.Prefixes[child] = &CaniPrefix{ID: child, Prefix: "10.1.1.0/24", Parent: withChild}

	// Act.
	refused := inv.RemovePrefix(withAddress)
	err := inv.RemovePrefix(withChild)

	// Assert.
	if refused == nil || !strings.Contains(refused.Error(), "holds 1 IP address(es)") || inv.Prefixes[withAddress] == nil {
		t.Errorf("removing the prefix holding an address = %v, kept = %v; want a refusal naming the address and the prefix kept",
			refused, inv.Prefixes[withAddress] != nil)
	}
	if err != nil || inv.Prefixes[withChild] != nil || inv.Prefixes[child].Parent != uuid.Nil {
		t.Errorf("removing the prefix holding a child = %v, kept = %v, child parent = %v; want removed and a root child",
			err, inv.Prefixes[withChild] != nil, inv.Prefixes[child].Parent)
	}
}
