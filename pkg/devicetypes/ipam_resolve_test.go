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
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// Fixed UUIDs for the duplicate pair, so the order the error lists them in is
// known.
var (
	duplicateOne = uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	duplicateTwo = uuid.MustParse("00000000-0000-0000-0000-0000000000a2")
)

// prefixResolverFixture holds Global 10.0.0.0/24 twice (duplicateOne and
// duplicateTwo) and 10.1.0.0/16, tenant-a 10.1.0.0/16, and org/tenant-b
// 10.1.0.0/16, whose namespace contains a slash.
func prefixResolverFixture() (inv *Inventory, global, tenant, slashed uuid.UUID) {
	inv = NewInventory()
	global, tenant, slashed = uuid.New(), uuid.New(), uuid.New()
	inv.Prefixes[duplicateOne] = &CaniPrefix{ID: duplicateOne, Prefix: "10.0.0.0/24"}
	inv.Prefixes[duplicateTwo] = &CaniPrefix{ID: duplicateTwo, Prefix: "10.0.0.0/24"}
	inv.Prefixes[global] = &CaniPrefix{ID: global, Prefix: "10.1.0.0/16"}
	inv.Prefixes[tenant] = &CaniPrefix{ID: tenant, Prefix: "10.1.0.0/16", Namespace: "tenant-a"}
	inv.Prefixes[slashed] = &CaniPrefix{ID: slashed, Prefix: "10.1.0.0/16", Namespace: "org/tenant-b"}
	return inv, global, tenant, slashed
}

// TestResolvePrefixReferenceFindsTheRecord verifies a prefix resolves by
// UUID, by CIDR in the given namespace, and by a namespace-qualified CIDR,
// including a namespace that itself contains a slash.
//
// Why it matters: remove and update act on what this returns.
// Inputs: the prefix resolver fixture and the references below.
// Outputs: each reference resolves to its record.
// Data choice: 10.1.9.9/16 is another spelling of 10.1.0.0/16, the same CIDR
// in three namespaces leaves the namespace to choose, and org/tenant-b shows
// the key is the shortest suffix that parses, not whatever follows the first
// slash.
func TestResolvePrefixReferenceFindsTheRecord(t *testing.T) {
	// Arrange.
	inv, global, tenant, slashed := prefixResolverFixture()
	cases := []struct {
		namespace, ref string
		want           uuid.UUID
	}{
		{"", "10.1.0.0/16", global},
		{"", "10.1.9.9/16", global},
		{"", "tenant-a/10.1.0.0/16", tenant},
		{"tenant-a", "10.1.0.0/16", tenant},
		{"", "org/tenant-b/10.1.0.0/16", slashed},
		{"org/tenant-b", "10.1.0.0/16", slashed},
		{"tenant-a", duplicateTwo.String(), duplicateTwo},
	}
	for _, tc := range cases {
		// Act.
		prefix, err := inv.ResolvePrefixReference(tc.namespace, tc.ref)

		// Assert.
		if err != nil || prefix.ID != tc.want {
			t.Errorf("ResolvePrefixReference(%q, %q) = %v, %v; want %v", tc.namespace, tc.ref, prefix, err, tc.want)
		}
	}
}

// TestResolvePrefixReferenceExplainsAFailure verifies a duplicated CIDR is
// refused with both UUIDs, and a missing CIDR or UUID is reported.
//
// Why it matters: resolving a duplicate to either record would change the
// wrong one, and the error is how the operator learns which UUIDs to use.
// Inputs: the prefix resolver fixture and the references below.
// Outputs: each reference fails with the message noted.
// Data choice: the duplicate pair has fixed UUIDs, so the listed order is
// known.
func TestResolvePrefixReferenceExplainsAFailure(t *testing.T) {
	// Arrange.
	inv, _, _, _ := prefixResolverFixture()
	cases := map[string]string{
		"10.0.0.0/24": fmt.Sprintf(`prefix 10.0.0.0/24 is ambiguous in namespace "Global"; use one of the UUIDs [%s %s]`,
			duplicateOne, duplicateTwo),
		"10.2.0.0/16":    `prefix 10.2.0.0/16 not found in namespace "Global"`,
		uuid.NewString(): "not found",
	}
	for ref, want := range cases {
		// Act.
		_, err := inv.ResolvePrefixReference("", ref)

		// Assert.
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("ResolvePrefixReference(%q) error = %v, want %q", ref, err, want)
		}
	}
}

// TestResolveIPAddressReference verifies an address resolves by UUID, and by
// host with or without a mask, qualified by namespace or not, and that an
// unknown host is an error.
//
// Why it matters: operators write addresses as they added them, mask
// included, while the mask is not part of an address's identity.
// Inputs: Global 10.0.0.5 under a Global /24 and tenant-a 10.0.0.5 under a
// tenant-a /24.
// Outputs: 10.0.0.5 and 10.0.0.5/32 resolve to the Global address,
// tenant-a/10.0.0.5/24 and the tenant address's UUID to the tenant one, and
// 10.0.0.6 is not found.
// Data choice: the same host in two namespaces leaves only the namespace, or
// the UUID, to tell them apart.
func TestResolveIPAddressReference(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	addresses := map[string]uuid.UUID{}
	for _, namespace := range []string{"Global", "tenant-a"} {
		prefix, address := uuid.New(), uuid.New()
		inv.Prefixes[prefix] = &CaniPrefix{ID: prefix, Prefix: "10.0.0.0/24", Namespace: namespace}
		inv.IPAddresses[address] = &CaniIPAddress{ID: address, Host: "10.0.0.5", Address: "10.0.0.5/24", Parent: prefix}
		addresses[namespace] = address
	}
	cases := map[string]uuid.UUID{
		"10.0.0.5": addresses["Global"], "10.0.0.5/32": addresses["Global"], "tenant-a/10.0.0.5/24": addresses["tenant-a"],
		addresses["tenant-a"].String(): addresses["tenant-a"],
	}

	// Act and assert.
	for ref, want := range cases {
		if address, err := inv.ResolveIPAddressReference("", ref); err != nil || address.ID != want {
			t.Errorf("ResolveIPAddressReference(%q) = %v, %v; want %v", ref, address, err, want)
		}
	}
	if _, err := inv.ResolveIPAddressReference("", "10.0.0.6"); err == nil ||
		!strings.Contains(err.Error(), `IP address 10.0.0.6 not found in namespace "Global"`) {
		t.Errorf("ResolveIPAddressReference(10.0.0.6) error = %v, want not found", err)
	}
}

// TestRemovePrefixWithAddressesRemovesWhatARootHolds verifies the cascade
// removes the addresses a root prefix holds, where RemovePrefix refuses with
// a recognisable error.
//
// Why it matters: the remove command offers the cascade when it sees that
// error, and the removed addresses must not leave dangling references.
// Inputs: a root 10.0.0.0/24 holding two addresses, one a device's primary
// IP.
// Outputs: RemovePrefix fails with ErrPrefixHoldsAddresses; the cascade
// reports 2 and leaves no prefix, no address and no primary IP.
// Data choice: the primary IP shows the addresses go through RemoveIPAddress.
func TestRemovePrefixWithAddressesRemovesWhatARootHolds(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	prefix, first, second, device := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Prefixes[prefix] = &CaniPrefix{ID: prefix, Prefix: "10.0.0.0/24"}
	inv.IPAddresses[first] = &CaniIPAddress{ID: first, Host: "10.0.0.5", Address: "10.0.0.5/24", Parent: prefix}
	inv.IPAddresses[second] = &CaniIPAddress{ID: second, Host: "10.0.0.6", Address: "10.0.0.6/24", Parent: prefix}
	inv.Devices[device] = &CaniDeviceType{ID: device, Name: "node", PrimaryIPv4: first}

	// Act.
	refusal := inv.RemovePrefix(prefix)
	removed, err := inv.RemovePrefixWithAddresses(prefix)

	// Assert.
	if !errors.Is(refusal, ErrPrefixHoldsAddresses) {
		t.Errorf("RemovePrefix error = %v, want ErrPrefixHoldsAddresses", refusal)
	}
	if err != nil || removed != 2 {
		t.Fatalf("RemovePrefixWithAddresses = %d, %v; want 2, nil", removed, err)
	}
	if len(inv.Prefixes) != 0 || len(inv.IPAddresses) != 0 || inv.Devices[device].PrimaryIPv4 != uuid.Nil {
		t.Errorf("left %d prefixes, %d addresses, primary IP %v; want none",
			len(inv.Prefixes), len(inv.IPAddresses), inv.Devices[device].PrimaryIPv4)
	}
}

// TestRemovePrefixWithAddressesKeepsWhatAParentCanTake verifies the cascade
// moves a mid-tree prefix's addresses up instead of removing them.
//
// Why it matters: the cascade overrides the root refusal only; addresses a
// parent can hold are kept, as RemovePrefix keeps them.
// Inputs: 10.0.0.0/16 holding 10.0.0.0/24, which holds 10.0.0.5.
// Outputs: 0 removed, and the address now under the /16.
// Data choice: the counterpart of the root case above.
func TestRemovePrefixWithAddressesKeepsWhatAParentCanTake(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	parent, prefix, address := uuid.New(), uuid.New(), uuid.New()
	inv.Prefixes[parent] = &CaniPrefix{ID: parent, Prefix: "10.0.0.0/16"}
	inv.Prefixes[prefix] = &CaniPrefix{ID: prefix, Prefix: "10.0.0.0/24", Parent: parent}
	inv.IPAddresses[address] = &CaniIPAddress{ID: address, Host: "10.0.0.5", Address: "10.0.0.5/24", Parent: prefix}

	// Act.
	removed, err := inv.RemovePrefixWithAddresses(prefix)

	// Assert.
	if err != nil || removed != 0 {
		t.Fatalf("RemovePrefixWithAddresses = %d, %v; want 0, nil", removed, err)
	}
	if got := inv.IPAddresses[address].Parent; got != parent {
		t.Errorf("address parent = %v, want the /16 %v", got, parent)
	}
}

// TestRemovePrefixWithAddressesCascadesUnderAMissingParent verifies a prefix
// whose parent record is gone is cascaded like a root, not re-parented onto
// the dangling UUID.
//
// Why it matters: the cascade is offered when RemovePrefix refuses, and that
// refusal treats a missing parent as none; both must agree.
// Inputs: 10.0.0.0/24 whose Parent is a UUID not in the inventory, holding
// one address.
// Outputs: 1 removed; no prefix and no address remain.
// Data choice: a single address makes the count exact.
func TestRemovePrefixWithAddressesCascadesUnderAMissingParent(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	prefix, address := uuid.New(), uuid.New()
	inv.Prefixes[prefix] = &CaniPrefix{ID: prefix, Prefix: "10.0.0.0/24", Parent: uuid.New()}
	inv.IPAddresses[address] = &CaniIPAddress{ID: address, Host: "10.0.0.5", Address: "10.0.0.5/24", Parent: prefix}

	// Act.
	removed, err := inv.RemovePrefixWithAddresses(prefix)

	// Assert.
	if err != nil || removed != 1 || len(inv.Prefixes) != 0 || len(inv.IPAddresses) != 0 {
		t.Errorf("RemovePrefixWithAddresses = %d, %v, leaving %d prefixes and %d addresses; want 1, nil and none",
			removed, err, len(inv.Prefixes), len(inv.IPAddresses))
	}
}
