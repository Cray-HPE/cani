/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
	"net"
	"testing"

	"github.com/google/uuid"
)

// TestBroadcastInvalidIP verifies the IPv4 and IPv6 broadcast helpers return an
// empty string when the network IP cannot be coerced to the expected width.
//
// Why it matters: prefix records may carry an address family that does not match
// the helper being called; returning "" rather than indexing a nil byte slice
// keeps broadcast derivation from panicking on malformed input.
// Inputs: an IPv6 network passed to broadcastIPv4 (To4 yields nil) and a
// deliberately malformed 3-byte IP passed to broadcastIPv6 (To16 yields nil).
// Outputs: an empty string from each. Data choice: a real IPv6 CIDR and a
// 3-byte IP are the minimal cases that force each To4/To16 nil guard.
func TestBroadcastInvalidIP(t *testing.T) {
	_, n6, err := net.ParseCIDR("2001:db8::/32")
	if err != nil {
		t.Fatalf("ParseCIDR: %v", err)
	}
	if got := broadcastIPv4(n6); got != "" {
		t.Errorf("broadcastIPv4(ipv6) = %q, want empty", got)
	}

	bad := &net.IPNet{IP: net.IP{1, 2, 3}, Mask: net.CIDRMask(24, 32)}
	if got := broadcastIPv6(bad); got != "" {
		t.Errorf("broadcastIPv6(malformed) = %q, want empty", got)
	}
}

// ---------- ParsePrefix ----------

// TestParsePrefixNil verifies ParsePrefix returns an error for a nil prefix.
//
// Why it matters: the IPAM layer derives a prefix's network/broadcast fields
// before storing it, so a nil pointer must be rejected rather than panicking
// deep inside net.ParseCIDR.
// Inputs: a nil *CaniPrefix. Outputs: a non-nil error.
// Data choice: nil is the only input that exercises the explicit guard clause,
// proving the function fails closed instead of dereferencing the pointer.
func TestParsePrefixNil(t *testing.T) {
	if err := ParsePrefix(nil); err == nil {
		t.Error("ParsePrefix(nil) should return an error")
	}
}

// TestParsePrefixInvalidCIDR verifies ParsePrefix returns an error when the
// Prefix string is not valid CIDR notation.
//
// Why it matters: prefixes are loaded from user/provider data, so malformed
// CIDR must surface as an error instead of producing a zero-value network.
// Inputs: a CaniPrefix whose Prefix is "not-a-cidr". Outputs: a non-nil error.
// Data choice: a clearly non-numeric string guarantees net.ParseCIDR fails,
// isolating the error path without ambiguity about IP family.
func TestParsePrefixInvalidCIDR(t *testing.T) {
	p := &CaniPrefix{Prefix: "not-a-cidr"}
	if err := ParsePrefix(p); err == nil {
		t.Error("ParsePrefix with invalid CIDR should return an error")
	}
}

// TestParsePrefixIPv4 verifies ParsePrefix populates the derived IPv4 fields.
//
// Why it matters: downstream IPAM consumers rely on Network, Broadcast,
// PrefixLen, and IPVersion being computed from the CIDR string, so the v4
// branch (including broadcastIPv4) must be exact.
// Inputs: a CaniPrefix with Prefix "10.0.0.0/24". Outputs: Network "10.0.0.0",
// Broadcast "10.0.0.255", PrefixLen 24, IPVersion 4, nil error.
// Data choice: a /24 makes the broadcast address (last octet 255) obvious and
// distinct from the network address, proving the host-bit OR math is correct.
func TestParsePrefixIPv4(t *testing.T) {
	p := &CaniPrefix{Prefix: "10.0.0.0/24"}
	if err := ParsePrefix(p); err != nil {
		t.Fatalf("ParsePrefix() unexpected error: %v", err)
	}
	if p.IPVersion != 4 {
		t.Errorf("IPVersion = %d, want 4", p.IPVersion)
	}
	if p.PrefixLen != 24 {
		t.Errorf("PrefixLen = %d, want 24", p.PrefixLen)
	}
	if p.Network != "10.0.0.0" {
		t.Errorf("Network = %q, want %q", p.Network, "10.0.0.0")
	}
	if p.Broadcast != "10.0.0.255" {
		t.Errorf("Broadcast = %q, want %q", p.Broadcast, "10.0.0.255")
	}
}

// TestParsePrefixIPv6 verifies ParsePrefix populates the derived IPv6 fields.
//
// Why it matters: IPv6 prefixes take a separate code path (broadcastIPv6), so
// the v6 last-address computation must be validated independently of v4.
// Inputs: a CaniPrefix with Prefix "2001:db8::/64". Outputs: IPVersion 6,
// PrefixLen 64, Broadcast "2001:db8::ffff:ffff:ffff:ffff", nil error.
// Data choice: a documentation-range /64 yields a predictable last address with
// all host bits set, proving the 16-byte OR math differs from the v4 path.
func TestParsePrefixIPv6(t *testing.T) {
	p := &CaniPrefix{Prefix: "2001:db8::/64"}
	if err := ParsePrefix(p); err != nil {
		t.Fatalf("ParsePrefix() unexpected error: %v", err)
	}
	if p.IPVersion != 6 {
		t.Errorf("IPVersion = %d, want 6", p.IPVersion)
	}
	if p.PrefixLen != 64 {
		t.Errorf("PrefixLen = %d, want 64", p.PrefixLen)
	}
	if p.Broadcast != "2001:db8::ffff:ffff:ffff:ffff" {
		t.Errorf("Broadcast = %q, want %q", p.Broadcast, "2001:db8::ffff:ffff:ffff:ffff")
	}
}

// ---------- ParseIPAddress ----------

// TestParseIPAddressNil verifies ParseIPAddress returns an error for a nil
// address.
//
// Why it matters: the IPAM layer normalizes addresses before storing them, so
// a nil pointer must be rejected rather than panicking.
// Inputs: a nil *CaniIPAddress. Outputs: a non-nil error.
// Data choice: nil is the only input that exercises the guard clause.
func TestParseIPAddressNil(t *testing.T) {
	if err := ParseIPAddress(nil); err == nil {
		t.Error("ParseIPAddress(nil) should return an error")
	}
}

// TestParseIPAddressCIDR verifies ParseIPAddress normalizes a CIDR-form
// address into its Host, MaskLength, IPVersion, and canonical Address fields.
//
// Why it matters: addresses arrive in mixed notation, so the "/"-bearing branch
// must split the host from the mask and re-emit a canonical "host/len" string.
// Inputs: a CaniIPAddress with Address "10.0.0.5/24". Outputs: Host "10.0.0.5",
// MaskLength 24, IPVersion 4, Address "10.0.0.5/24", nil error.
// Data choice: a host bit (the .5) inside a /24 proves the mask length is read
// from the CIDR suffix and not defaulted to /32.
func TestParseIPAddressCIDR(t *testing.T) {
	addr := &CaniIPAddress{Address: "10.0.0.5/24"}
	if err := ParseIPAddress(addr); err != nil {
		t.Fatalf("ParseIPAddress() unexpected error: %v", err)
	}
	if addr.Host != "10.0.0.5" {
		t.Errorf("Host = %q, want %q", addr.Host, "10.0.0.5")
	}
	if addr.MaskLength != 24 {
		t.Errorf("MaskLength = %d, want 24", addr.MaskLength)
	}
	if addr.IPVersion != 4 {
		t.Errorf("IPVersion = %d, want 4", addr.IPVersion)
	}
	if addr.Address != "10.0.0.5/24" {
		t.Errorf("Address = %q, want %q", addr.Address, "10.0.0.5/24")
	}
}

// TestParseIPAddressBareIPv4 verifies ParseIPAddress defaults a maskless IPv4
// address to /32.
//
// Why it matters: a bare host address has no subnet context, so the function
// must assign the host route length (/32) to keep the stored Address canonical.
// Inputs: a CaniIPAddress with Address "10.0.0.7". Outputs: MaskLength 32,
// Address "10.0.0.7/32", IPVersion 4, nil error.
// Data choice: omitting the "/" forces the bare-IP branch, and /32 is the only
// correct default for a single IPv4 host.
func TestParseIPAddressBareIPv4(t *testing.T) {
	addr := &CaniIPAddress{Address: "10.0.0.7"}
	if err := ParseIPAddress(addr); err != nil {
		t.Fatalf("ParseIPAddress() unexpected error: %v", err)
	}
	if addr.MaskLength != 32 {
		t.Errorf("MaskLength = %d, want 32", addr.MaskLength)
	}
	if addr.Address != "10.0.0.7/32" {
		t.Errorf("Address = %q, want %q", addr.Address, "10.0.0.7/32")
	}
}

// TestParseIPAddressBareIPv6 verifies ParseIPAddress defaults a maskless IPv6
// address to /128 and records IPVersion 6.
//
// Why it matters: the bare-IP branch must distinguish address families so an
// IPv6 host route gets /128 rather than the IPv4 /32.
// Inputs: a CaniIPAddress with Address "2001:db8::1". Outputs: MaskLength 128,
// IPVersion 6, Address "2001:db8::1/128", nil error.
// Data choice: a v6 literal with no "/" is the smallest input that proves the
// family check selects /128 instead of /32.
func TestParseIPAddressBareIPv6(t *testing.T) {
	addr := &CaniIPAddress{Address: "2001:db8::1"}
	if err := ParseIPAddress(addr); err != nil {
		t.Fatalf("ParseIPAddress() unexpected error: %v", err)
	}
	if addr.MaskLength != 128 {
		t.Errorf("MaskLength = %d, want 128", addr.MaskLength)
	}
	if addr.IPVersion != 6 {
		t.Errorf("IPVersion = %d, want 6", addr.IPVersion)
	}
}

// TestParseIPAddressInvalid verifies ParseIPAddress reports errors for both
// malformed CIDR and malformed bare addresses.
//
// Why it matters: invalid host data must not be silently stored, so both
// parsing branches need an error return.
// Inputs: addresses "bad/24" (CIDR branch) and "999.1.1.1" (bare branch).
// Outputs: a non-nil error for each case.
// Data choice: one input carries a "/" to hit net.ParseCIDR and the other omits
// it to hit net.ParseIP, covering both failure branches in one table.
func TestParseIPAddressInvalid(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"invalid cidr", "bad/24"},
		{"invalid bare ip", "999.1.1.1"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			addr := &CaniIPAddress{Address: tt.in}
			if err := ParseIPAddress(addr); err == nil {
				t.Errorf("ParseIPAddress(%q) should return an error", tt.in)
			}
		})
	}
}

// ---------- FindParentPrefix ----------

// TestFindParentPrefixMostSpecific verifies FindParentPrefix returns the
// most-specific containing prefix among several candidates.
//
// Why it matters: prefix hierarchy is built by attaching each prefix to its
// tightest enclosing parent, so the longest matching mask must win.
// Inputs: a target "10.1.2.0/24" and a pool containing "10.0.0.0/8" and
// "10.1.0.0/16". Outputs: the UUID of the /16 prefix.
// Data choice: two nested supernets that both contain the target force the
// function to choose by mask length rather than by map iteration order.
func TestFindParentPrefixMostSpecific(t *testing.T) {
	eightID := uuid.New()
	sixteenID := uuid.New()
	prefixes := map[uuid.UUID]*CaniPrefix{
		eightID:   {ID: eightID, Prefix: "10.0.0.0/8"},
		sixteenID: {ID: sixteenID, Prefix: "10.1.0.0/16"},
	}
	target := &CaniPrefix{ID: uuid.New(), Prefix: "10.1.2.0/24", PrefixLen: 24}

	got := FindParentPrefix(target, prefixes)
	if got != sixteenID {
		t.Errorf("FindParentPrefix() = %v, want %v (the /16)", got, sixteenID)
	}
}

// TestFindParentPrefixNoParent verifies FindParentPrefix returns uuid.Nil when
// no candidate contains the target.
//
// Why it matters: a top-level prefix has no parent, so the function must report
// "none" rather than attaching it to an unrelated network.
// Inputs: a target "10.0.0.0/24" and a pool containing only "192.168.0.0/16".
// Outputs: uuid.Nil.
// Data choice: a disjoint supernet guarantees Contains() is false, isolating
// the no-match path.
func TestFindParentPrefixNoParent(t *testing.T) {
	otherID := uuid.New()
	prefixes := map[uuid.UUID]*CaniPrefix{
		otherID: {ID: otherID, Prefix: "192.168.0.0/16"},
	}
	target := &CaniPrefix{ID: uuid.New(), Prefix: "10.0.0.0/24", PrefixLen: 24}

	if got := FindParentPrefix(target, prefixes); got != uuid.Nil {
		t.Errorf("FindParentPrefix() = %v, want uuid.Nil", got)
	}
}

// TestFindParentPrefixGuards verifies FindParentPrefix returns uuid.Nil for a
// nil target and for a target whose Prefix is unparseable.
//
// Why it matters: malformed or missing targets must not crash hierarchy
// computation; they should resolve to "no parent".
// Inputs: a nil target, then a target with Prefix "garbage". Outputs: uuid.Nil
// in both cases.
// Data choice: nil exercises the first guard and an unparseable string exercises
// the ParseCIDR error guard, covering both early returns.
func TestFindParentPrefixGuards(t *testing.T) {
	if got := FindParentPrefix(nil, nil); got != uuid.Nil {
		t.Errorf("FindParentPrefix(nil) = %v, want uuid.Nil", got)
	}
	target := &CaniPrefix{ID: uuid.New(), Prefix: "garbage", PrefixLen: 24}
	if got := FindParentPrefix(target, map[uuid.UUID]*CaniPrefix{}); got != uuid.Nil {
		t.Errorf("FindParentPrefix(invalid) = %v, want uuid.Nil", got)
	}
}

// TestFindParentPrefixSkipsSelfAndEqualMask verifies FindParentPrefix ignores
// the target's own entry and any candidate that is not strictly less specific.
//
// Why it matters: a prefix must never be its own parent, and an equally specific
// peer is not a parent, so both must be excluded from the search.
// Inputs: a pool containing the target itself (same ID) and a same-mask peer
// "10.1.2.0/24"; target "10.1.2.0/24". Outputs: uuid.Nil.
// Data choice: reusing the target's ID exercises the self-skip branch while the
// equal-length peer exercises the "ones >= target.PrefixLen" continue.
func TestFindParentPrefixSkipsSelfAndEqualMask(t *testing.T) {
	targetID := uuid.New()
	peerID := uuid.New()
	prefixes := map[uuid.UUID]*CaniPrefix{
		targetID: {ID: targetID, Prefix: "10.1.2.0/24"},
		peerID:   {ID: peerID, Prefix: "10.1.2.0/24"},
	}
	target := &CaniPrefix{ID: targetID, Prefix: "10.1.2.0/24", PrefixLen: 24}

	if got := FindParentPrefix(target, prefixes); got != uuid.Nil {
		t.Errorf("FindParentPrefix() = %v, want uuid.Nil", got)
	}
}

// ---------- FindParentPrefixForIP ----------

// TestFindParentPrefixForIPMostSpecific verifies FindParentPrefixForIP returns
// the tightest prefix that contains the host address.
//
// Why it matters: an IP address is auto-attached to its most-specific enclosing
// prefix, so the longest matching mask must win.
// Inputs: an address with Host "10.1.2.3" and a pool of "10.0.0.0/8" and
// "10.1.2.0/24". Outputs: the UUID of the /24 prefix.
// Data choice: nested supernets that both contain the host force selection by
// mask length rather than iteration order.
func TestFindParentPrefixForIPMostSpecific(t *testing.T) {
	eightID := uuid.New()
	twentyFourID := uuid.New()
	prefixes := map[uuid.UUID]*CaniPrefix{
		eightID:      {ID: eightID, Prefix: "10.0.0.0/8"},
		twentyFourID: {ID: twentyFourID, Prefix: "10.1.2.0/24"},
	}
	addr := &CaniIPAddress{ID: uuid.New(), Host: "10.1.2.3"}

	if got := FindParentPrefixForIP(addr, prefixes); got != twentyFourID {
		t.Errorf("FindParentPrefixForIP() = %v, want %v (the /24)", got, twentyFourID)
	}
}

// TestFindParentPrefixForIPGuards verifies FindParentPrefixForIP returns
// uuid.Nil for a nil address, an unparseable host, and a host with no
// containing prefix.
//
// Why it matters: addresses with missing or unmatched parents must resolve to
// "no parent" instead of crashing or attaching arbitrarily.
// Inputs: nil; an address with Host "nope"; an address with Host "10.0.0.1"
// against a pool of only "192.168.0.0/16". Outputs: uuid.Nil for each.
// Data choice: each input targets one early-return branch (nil guard, ParseIP
// failure, and no-Contains match) so all three are covered.
func TestFindParentPrefixForIPGuards(t *testing.T) {
	if got := FindParentPrefixForIP(nil, nil); got != uuid.Nil {
		t.Errorf("FindParentPrefixForIP(nil) = %v, want uuid.Nil", got)
	}
	bad := &CaniIPAddress{ID: uuid.New(), Host: "nope"}
	if got := FindParentPrefixForIP(bad, map[uuid.UUID]*CaniPrefix{}); got != uuid.Nil {
		t.Errorf("FindParentPrefixForIP(bad host) = %v, want uuid.Nil", got)
	}
	otherID := uuid.New()
	prefixes := map[uuid.UUID]*CaniPrefix{
		otherID: {ID: otherID, Prefix: "192.168.0.0/16"},
	}
	addr := &CaniIPAddress{ID: uuid.New(), Host: "10.0.0.1"}
	if got := FindParentPrefixForIP(addr, prefixes); got != uuid.Nil {
		t.Errorf("FindParentPrefixForIP(no match) = %v, want uuid.Nil", got)
	}
}

// ---------- GetID ----------

// TestIPAMGetID verifies the GetID accessors on CaniPrefix, CaniIPAddress, and
// CaniVLAN return the stored UUID and uuid.Nil for a nil receiver.
//
// Why it matters: generic IPAM code keys objects by GetID(), so a nil-safe
// accessor prevents panics when iterating partially populated maps.
// Inputs: populated and nil pointers of each type. Outputs: the stored ID for
// populated receivers; uuid.Nil for nil receivers.
// Data choice: a fresh uuid.New() per type proves the accessor returns that
// object's own ID, and the nil receiver proves the guard clause.
func TestIPAMGetID(t *testing.T) {
	pid := uuid.New()
	if got := (&CaniPrefix{ID: pid}).GetID(); got != pid {
		t.Errorf("CaniPrefix.GetID() = %v, want %v", got, pid)
	}
	var nilPrefix *CaniPrefix
	if got := nilPrefix.GetID(); got != uuid.Nil {
		t.Errorf("nil CaniPrefix.GetID() = %v, want uuid.Nil", got)
	}

	aid := uuid.New()
	if got := (&CaniIPAddress{ID: aid}).GetID(); got != aid {
		t.Errorf("CaniIPAddress.GetID() = %v, want %v", got, aid)
	}
	var nilAddr *CaniIPAddress
	if got := nilAddr.GetID(); got != uuid.Nil {
		t.Errorf("nil CaniIPAddress.GetID() = %v, want uuid.Nil", got)
	}

	vid := uuid.New()
	if got := (&CaniVLAN{ID: vid}).GetID(); got != vid {
		t.Errorf("CaniVLAN.GetID() = %v, want %v", got, vid)
	}
	var nilVLAN *CaniVLAN
	if got := nilVLAN.GetID(); got != uuid.Nil {
		t.Errorf("nil CaniVLAN.GetID() = %v, want uuid.Nil", got)
	}
}

// ---------- AddVLAN ----------

// TestAddVLAN verifies AddVLAN inserts a VLAN, rejects nil, and rejects a
// duplicate ID.
//
// Why it matters: the VLAN catalog is keyed by UUID, so inserts must be unique
// and nil-safe to keep the inventory consistent.
// Inputs: a valid VLAN, a nil VLAN, and a second VLAN reusing the first ID.
// Outputs: nil error and map membership for the valid case; non-nil errors for
// nil and duplicate.
// Data choice: reusing the same UUID for the duplicate case isolates the
// uniqueness check, independent of any other field.
func TestAddVLAN(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	vlan := &CaniVLAN{ID: id, VID: 100, Name: "vlan100"}

	if err := inv.AddVLAN(vlan); err != nil {
		t.Fatalf("AddVLAN() unexpected error: %v", err)
	}
	if _, ok := inv.VLANs[id]; !ok {
		t.Error("expected VLAN to be present after AddVLAN")
	}
	if err := inv.AddVLAN(nil); err == nil {
		t.Error("AddVLAN(nil) should return an error")
	}
	if err := inv.AddVLAN(&CaniVLAN{ID: id, VID: 200, Name: "dup"}); err == nil {
		t.Error("AddVLAN(duplicate ID) should return an error")
	}
}

// ---------- AddPrefix ----------

// TestAddPrefixComputesParent verifies AddPrefix parses the CIDR and
// auto-attaches the new prefix to an existing enclosing prefix.
//
// Why it matters: AddPrefix is the entry point that builds the prefix
// hierarchy, so it must both derive fields (via ParsePrefix) and set Parent
// (via FindParentPrefix) in one call.
// Inputs: a /16 prefix added first, then a /24 nested inside it. Outputs: nil
// errors and the /24's Parent equal to the /16's ID.
// Data choice: a /24 strictly inside a /16 guarantees a single unambiguous
// parent, proving the auto-attach wiring without iteration-order risk.
func TestAddPrefixComputesParent(t *testing.T) {
	inv := NewInventory()
	parentID := uuid.New()
	parent := &CaniPrefix{ID: parentID, Prefix: "10.1.0.0/16"}
	if err := inv.AddPrefix(parent); err != nil {
		t.Fatalf("AddPrefix(parent) error: %v", err)
	}

	childID := uuid.New()
	child := &CaniPrefix{ID: childID, Prefix: "10.1.2.0/24"}
	if err := inv.AddPrefix(child); err != nil {
		t.Fatalf("AddPrefix(child) error: %v", err)
	}
	if child.Parent != parentID {
		t.Errorf("child.Parent = %v, want %v", child.Parent, parentID)
	}
}

// TestAddPrefixErrors verifies AddPrefix rejects nil, duplicate IDs, and
// invalid CIDR strings.
//
// Why it matters: invalid prefixes must never enter the inventory, or the
// hierarchy and IPAM math built on them would be corrupt.
// Inputs: a nil prefix; a duplicate of a pre-inserted ID; a prefix with an
// unparseable CIDR. Outputs: a non-nil error for each.
// Data choice: each case targets one guard (nil, existence, ParsePrefix error)
// so the three rejection paths are covered distinctly.
func TestAddPrefixErrors(t *testing.T) {
	inv := NewInventory()
	if err := inv.AddPrefix(nil); err == nil {
		t.Error("AddPrefix(nil) should return an error")
	}
	id := uuid.New()
	inv.Prefixes[id] = &CaniPrefix{ID: id, Prefix: "10.0.0.0/24"}
	if err := inv.AddPrefix(&CaniPrefix{ID: id, Prefix: "10.0.0.0/24"}); err == nil {
		t.Error("AddPrefix(duplicate) should return an error")
	}
	if err := inv.AddPrefix(&CaniPrefix{ID: uuid.New(), Prefix: "not-cidr"}); err == nil {
		t.Error("AddPrefix(invalid CIDR) should return an error")
	}
}

// ---------- AddIPAddress ----------

// TestAddIPAddressComputesParent verifies AddIPAddress parses the address and
// auto-attaches it to the enclosing prefix.
//
// Why it matters: AddIPAddress is the entry point for host records, so it must
// derive Host/MaskLength (via ParseIPAddress) and resolve Parent (via
// FindParentPrefixForIP) in one call.
// Inputs: a /24 prefix added first, then a host "10.1.2.3/24". Outputs: nil
// errors and the address's Parent equal to the prefix's ID.
// Data choice: a host that falls inside exactly one prefix proves the auto
// attach without ambiguity.
func TestAddIPAddressComputesParent(t *testing.T) {
	inv := NewInventory()
	prefixID := uuid.New()
	if err := inv.AddPrefix(&CaniPrefix{ID: prefixID, Prefix: "10.1.2.0/24"}); err != nil {
		t.Fatalf("AddPrefix() error: %v", err)
	}

	addrID := uuid.New()
	addr := &CaniIPAddress{ID: addrID, Address: "10.1.2.3/24"}
	if err := inv.AddIPAddress(addr); err != nil {
		t.Fatalf("AddIPAddress() error: %v", err)
	}
	if addr.Parent != prefixID {
		t.Errorf("addr.Parent = %v, want %v", addr.Parent, prefixID)
	}
}

// TestAddIPAddressErrors verifies AddIPAddress rejects nil, duplicate IDs, and
// invalid address strings.
//
// Why it matters: malformed host records must never enter the inventory.
// Inputs: a nil address; a duplicate of a pre-inserted ID; an address with an
// unparseable value. Outputs: a non-nil error for each.
// Data choice: each case targets one guard (nil, existence, ParseIPAddress
// error), covering the three rejection paths distinctly.
func TestAddIPAddressErrors(t *testing.T) {
	inv := NewInventory()
	if err := inv.AddIPAddress(nil); err == nil {
		t.Error("AddIPAddress(nil) should return an error")
	}
	id := uuid.New()
	inv.IPAddresses[id] = &CaniIPAddress{ID: id, Host: "10.0.0.1"}
	if err := inv.AddIPAddress(&CaniIPAddress{ID: id, Address: "10.0.0.1/24"}); err == nil {
		t.Error("AddIPAddress(duplicate) should return an error")
	}
	if err := inv.AddIPAddress(&CaniIPAddress{ID: uuid.New(), Address: "bad/24"}); err == nil {
		t.Error("AddIPAddress(invalid) should return an error")
	}
}
