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
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ---------- natural keys ----------

// TestIPAMNaturalKeysNormalizeAndScope verifies the prefix, address, and VRF
// natural keys combine the effective namespace with a canonical value.
//
// Why it matters: equivalent spellings must share one identity inside a
// namespace while identical text in another namespace must not, and the mask
// must never be part of an address's identity.
// Inputs: a host-form CIDR, a masked address, and blank versus explicit
// namespaces. Outputs: canonical network/host text and Global defaulting.
// Data choice: "10.0.0.5/24" as a prefix exercises network canonicalization,
// and the same host with two masks proves the mask is excluded.
func TestIPAMNaturalKeysNormalizeAndScope(t *testing.T) {
	prefix := &CaniPrefix{Prefix: "10.0.0.5/24"}
	if got := prefixNaturalKey(prefix); got != (prefixKey{Namespace: "Global", CIDR: "10.0.0.0/24"}) {
		t.Errorf("prefixNaturalKey(blank namespace) = %v, want Global/10.0.0.0/24", got)
	}
	prefix.Namespace = "tenant-a"
	if got := prefixNaturalKey(prefix).Namespace; got != "tenant-a" {
		t.Errorf("prefixNaturalKey namespace = %q, want tenant-a", got)
	}
	prefix.Namespace = " tenant-a "
	if got := prefixNaturalKey(prefix).Namespace; got != "tenant-a" {
		t.Errorf("prefixNaturalKey(padded namespace) = %q, want tenant-a: surrounding whitespace must not create a namespace", got)
	}

	wide := &CaniIPAddress{Host: "10.0.0.5", Address: "10.0.0.5/24"}
	narrow := &CaniIPAddress{Address: "10.0.0.5/25"}
	if ipAddressNaturalKey(wide, "") != ipAddressNaturalKey(narrow, "Global") {
		t.Error("address keys differ across masks; the mask must not be part of identity")
	}
	if ipAddressNaturalKey(wide, "tenant-a") == ipAddressNaturalKey(wide, "tenant-b") {
		t.Error("address keys collide across namespaces")
	}

	if got := vrfNaturalKey(&CaniVRF{Name: "blue"}); got != (vrfKey{Namespace: "Global", Name: "blue"}) {
		t.Errorf("vrfNaturalKey = %v, want Global/blue", got)
	}
}

// TestParsersCanonicalizeNamespace verifies ParsePrefix and ParseIPAddress
// store the namespace without surrounding whitespace.
//
// Why it matters: the stored field is what export sends as the Nautobot
// namespace name, so a padded value would create a second namespace there
// even though the key treats it as the same scope locally.
// Inputs: a prefix and an address whose namespace is " tenant-a ".
// Outputs: both store "tenant-a".
// Data choice: padding on both sides covers leading and trailing trimming.
func TestParsersCanonicalizeNamespace(t *testing.T) {
	prefix := &CaniPrefix{Prefix: "10.0.0.0/24", Namespace: " tenant-a "}
	addr := &CaniIPAddress{Address: "10.0.0.5/24", Namespace: " tenant-a "}

	if err := ParsePrefix(prefix); err != nil || prefix.Namespace != "tenant-a" {
		t.Errorf("ParsePrefix namespace = %q, err = %v; want tenant-a", prefix.Namespace, err)
	}
	if err := ParseIPAddress(addr); err != nil || addr.Namespace != "tenant-a" {
		t.Errorf("ParseIPAddress namespace = %q, err = %v; want tenant-a", addr.Namespace, err)
	}
}

// TestIPAddressNamespaceDerivesFromParent verifies a parented address takes its
// namespace from the parent prefix while a draft keeps its own intent.
//
// Why it matters: prefix scope is authoritative, so an address must follow its
// parent even when its own intent field is stale or blank.
// Inputs: a prefix in "tenant-a", an address parented to it with a blank
// namespace, and an orphan address with explicit intent. Outputs: "tenant-a"
// for the parented address and the intent for the orphan.
// Data choice: a blank intent on the parented address isolates the derivation
// from a trivially matching field.
func TestIPAddressNamespaceDerivesFromParent(t *testing.T) {
	inv := NewInventory()
	prefixID := uuid.New()
	inv.Prefixes[prefixID] = &CaniPrefix{ID: prefixID, Prefix: "10.0.0.0/24", Namespace: "tenant-a"}

	parented := &CaniIPAddress{Address: "10.0.0.5/24", Parent: prefixID}
	if got := inv.IPAddressNamespace(parented); got != "tenant-a" {
		t.Errorf("IPAddressNamespace(parented) = %q, want tenant-a", got)
	}
	orphan := &CaniIPAddress{Address: "10.0.0.6/24", Namespace: "tenant-b"}
	if got := inv.IPAddressNamespace(orphan); got != "tenant-b" {
		t.Errorf("IPAddressNamespace(orphan) = %q, want tenant-b", got)
	}
	if got := inv.IPAddressNamespace(&CaniIPAddress{Address: "10.0.0.7/24"}); got != DefaultNamespace {
		t.Errorf("IPAddressNamespace(blank) = %q, want %q", got, DefaultNamespace)
	}
}

// ---------- scoped parent selection ----------

// TestFindParentPrefixStaysInsideNamespace verifies parent selection ignores a
// tighter containing prefix that lives in another namespace.
//
// Why it matters: two independent networks may both hold 10.0.0.0/8; picking
// the other namespace's prefix would silently splice the hierarchies together.
// Inputs: a target /24 in "tenant-a", a /8 in "tenant-a", and a /16 in
// "tenant-b" that also contains it. Outputs: the /8 in "tenant-a".
// Data choice: the foreign candidate is more specific than the correct one so
// an unscoped longest-match would return the wrong answer.
func TestFindParentPrefixStaysInsideNamespace(t *testing.T) {
	ownEight, foreignSixteen := uuid.New(), uuid.New()
	prefixes := map[uuid.UUID]*CaniPrefix{
		ownEight:       {ID: ownEight, Prefix: "10.0.0.0/8", Namespace: "tenant-a"},
		foreignSixteen: {ID: foreignSixteen, Prefix: "10.1.0.0/16", Namespace: "tenant-b"},
	}
	target := &CaniPrefix{ID: uuid.New(), Prefix: "10.1.2.0/24", PrefixLen: 24, Namespace: "tenant-a"}
	if got := FindParentPrefix(target, prefixes); got != ownEight {
		t.Errorf("FindParentPrefix() = %v, want %v (same-namespace /8)", got, ownEight)
	}

	addr := &CaniIPAddress{Host: "10.1.2.3", Address: "10.1.2.3/24", Namespace: "tenant-a"}
	if got := FindParentPrefixForIP(addr, prefixes); got != ownEight {
		t.Errorf("FindParentPrefixForIP() = %v, want %v (same-namespace /8)", got, ownEight)
	}
	addr.Namespace = "tenant-c"
	if got := FindParentPrefixForIP(addr, prefixes); got != uuid.Nil {
		t.Errorf("FindParentPrefixForIP(unknown namespace) = %v, want uuid.Nil", got)
	}
}

// TestFindParentPrefixTieBreaksDeterministically verifies equal-length
// duplicate candidates resolve to the same parent on every call.
//
// Why it matters: legacy inventories may hold the same CIDR twice; map
// iteration must not decide which copy becomes the parent.
// Inputs: two "10.0.0.0/16" prefixes in Global and a /24 target. Outputs: the
// candidate with the lexically smaller UUID, identical across 50 calls.
// Data choice: repeating the call surfaces nondeterminism that a single
// invocation would hide.
func TestFindParentPrefixTieBreaksDeterministically(t *testing.T) {
	first, second := uuid.New(), uuid.New()
	if first.String() > second.String() {
		first, second = second, first
	}
	prefixes := map[uuid.UUID]*CaniPrefix{
		first:  {ID: first, Prefix: "10.0.0.0/16"},
		second: {ID: second, Prefix: "10.0.0.0/16"},
	}
	target := &CaniPrefix{ID: uuid.New(), Prefix: "10.0.1.0/24", PrefixLen: 24}
	for i := 0; i < 50; i++ {
		if got := FindParentPrefix(target, prefixes); got != first {
			t.Fatalf("call %d: FindParentPrefix() = %v, want %v (lexically smaller UUID)", i, got, first)
		}
	}
}

// ---------- VRF resolution ----------

// TestResolveVRFReferenceScopesAndReportsAmbiguity verifies VRF lookup by name
// is namespace-scoped, accepts a "namespace/name" qualifier, and refuses to
// pick a winner among same-named VRFs.
//
// Why it matters: namespace-qualified names are selectors, not identities;
// taking the first match would attach a prefix to the wrong routing table.
// Inputs: "blue" in Global, "blue" in tenant-a, and two "red" VRFs in Global.
// Outputs: the Global blue for an unqualified lookup, the tenant copy for the
// qualified one, an ambiguity error for red, and a UUID lookup that bypasses
// names entirely.
// Data choice: the duplicate pair isolates the ambiguity branch from the
// not-found branch.
func TestResolveVRFReferenceScopesAndReportsAmbiguity(t *testing.T) {
	inv := NewInventory()
	globalBlue, tenantBlue, redOne, redTwo := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.VRFs[globalBlue] = &CaniVRF{ID: globalBlue, Name: "blue"}
	inv.VRFs[tenantBlue] = &CaniVRF{ID: tenantBlue, Name: "blue", Namespace: "tenant-a"}
	inv.VRFs[redOne] = &CaniVRF{ID: redOne, Name: "red"}
	inv.VRFs[redTwo] = &CaniVRF{ID: redTwo, Name: "red"}
	cases := []struct {
		name, namespace, ref string
		want                 uuid.UUID
		wantErr              string
	}{
		{name: "unqualified name in Global", ref: "blue", want: globalBlue},
		{name: "qualified name crosses namespaces", namespace: "Global", ref: "tenant-a/blue", want: tenantBlue},
		{name: "uuid bypasses names", ref: redOne.String(), want: redOne},
		{name: "duplicate name is ambiguous", ref: "red", wantErr: "ambiguous"},
		{name: "unknown namespace", namespace: "tenant-b", ref: "blue", wantErr: "not found"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vrf, err := inv.ResolveVRFReference(tc.namespace, tc.ref)
			assertVRFResolution(t, vrf, err, tc.want, tc.wantErr)
		})
	}
}

// assertVRFResolution checks a lookup against the wanted VRF, or against the
// wanted error text when wantErr is set.
func assertVRFResolution(t *testing.T, vrf *CaniVRF, err error, want uuid.UUID, wantErr string) {
	t.Helper()
	if wantErr != "" {
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Errorf("error = %v, want containing %q", err, wantErr)
		}
		return
	}
	if err != nil || vrf.ID != want {
		t.Errorf("= %v, %v; want %v", vrf, err, want)
	}
}

// TestResolveVRFReferencePrefersTheExactName verifies a name lookup returns
// the VRF named exactly so, falls back to a single case-insensitive match,
// and reports several case-insensitive matches as ambiguous.
//
// Why it matters: Nautobot treats "red" and "Red" as different VRFs, and so
// does the merge key; folding case first would leave both unreachable by name
// and disagree with the identity the duplicate report uses.
// Inputs: Global VRFs red, Red and blue.
// Outputs: red and Red each resolve to themselves, BLUE resolves to blue, and
// RED is ambiguous.
// Data choice: red and Red differ only in case, the pair a case-folding lookup
// cannot tell apart.
func TestResolveVRFReferencePrefersTheExactName(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	red, upper, blue := uuid.New(), uuid.New(), uuid.New()
	inv.VRFs[red] = &CaniVRF{ID: red, Name: "red"}
	inv.VRFs[upper] = &CaniVRF{ID: upper, Name: "Red"}
	inv.VRFs[blue] = &CaniVRF{ID: blue, Name: "blue"}

	for ref, want := range map[string]uuid.UUID{"red": red, "Red": upper, "BLUE": blue} {
		// Act.
		vrf, err := inv.ResolveVRFReference("", ref)

		// Assert.
		if err != nil || vrf.ID != want {
			t.Errorf("ResolveVRFReference(%q) = %v, %v; want %v", ref, vrf, err, want)
		}
	}
	if _, err := inv.ResolveVRFReference("", "RED"); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("ResolveVRFReference(RED) error = %v, want ambiguity error", err)
	}
}
