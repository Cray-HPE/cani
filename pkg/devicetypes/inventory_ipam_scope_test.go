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

func scopeErrors(result *RelationshipResult) string {
	if err := result.Err(); err != nil {
		return err.Error()
	}
	return ""
}

// TestIPAMScopeRejectsCrossNamespaceLinks verifies a prefix parent, a prefix
// VRF membership, and an address parent in another namespace are errors.
//
// Why it matters: a link that crosses namespaces means two independent
// networks have been spliced together; it must block the mutation rather than
// be persisted.
// Inputs: a Global /16 parenting a tenant-a /24, a tenant-a prefix listing a
// Global VRF, and a tenant-b address parented to a tenant-a prefix. Outputs:
// three errors naming the offending namespaces.
// Data choice: each link is otherwise valid (containing, existing) so only the
// namespace mismatch can explain the error.
func TestIPAMScopeRejectsCrossNamespaceLinks(t *testing.T) {
	inv := NewInventory()
	globalSixteen, tenantTwentyFour, globalVRF, addrID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Prefixes[globalSixteen] = &CaniPrefix{ID: globalSixteen, Prefix: "10.0.0.0/16"}
	inv.VRFs[globalVRF] = &CaniVRF{ID: globalVRF, Name: "blue"}
	inv.Prefixes[tenantTwentyFour] = &CaniPrefix{
		ID: tenantTwentyFour, Prefix: "10.0.1.0/24", Namespace: "tenant-a",
		Parent: globalSixteen, VRFs: []uuid.UUID{globalVRF},
	}
	inv.IPAddresses[addrID] = &CaniIPAddress{
		ID: addrID, Host: "10.0.1.5", Address: "10.0.1.5/24", Namespace: "tenant-b", Parent: tenantTwentyFour,
	}

	result := &RelationshipResult{}
	inv.validateIPAMScope(result)

	if len(result.Errors) != 3 {
		t.Fatalf("errors = %d (%v), want 3", len(result.Errors), result.Errors)
	}
	msg := scopeErrors(result)
	for _, want := range []string{
		`parent prefix "10.0.0.0/16" is in namespace "Global", not "tenant-a"`,
		`VRF "blue" is in namespace "Global", not "tenant-a"`,
		`parent prefix "10.0.1.0/24" is in namespace "tenant-a", not "tenant-b"`,
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("errors missing %q in:\n%s", want, msg)
		}
	}
}

// TestIPAMScopeRejectsNonContainingParent verifies an address whose parent
// prefix does not contain it is an error.
//
// Why it matters: export sends the parent as the Nautobot parent and namespace
// source; a non-containing parent would be rejected remotely or, worse,
// accepted into the wrong network.
// Inputs: a 10.0.0.0/24 prefix and an address 192.168.1.1/24 parented to it.
// Outputs: exactly one error mentioning containment.
// Data choice: the address is in a disjoint network so Contains() is
// unambiguously false.
func TestIPAMScopeRejectsNonContainingParent(t *testing.T) {
	inv := NewInventory()
	prefixID, addrID := uuid.New(), uuid.New()
	inv.Prefixes[prefixID] = &CaniPrefix{ID: prefixID, Prefix: "10.0.0.0/24"}
	inv.IPAddresses[addrID] = &CaniIPAddress{ID: addrID, Host: "192.168.1.1", Address: "192.168.1.1/24", Parent: prefixID}

	result := &RelationshipResult{}
	inv.validateIPAMScope(result)

	if len(result.Errors) != 1 || !strings.Contains(scopeErrors(result), "does not contain the address") {
		t.Fatalf("errors = %v, want one containment error", result.Errors)
	}
}

// TestIPAMScopeReportsLegacyDuplicatesAsUnresolved verifies same-key
// duplicates and an unresolved legacy VRF name are reported for the operator
// to resolve without blocking the load.
//
// Why it matters: legacy inventories may hold a CIDR twice (separated only by
// a VRF string) or a host twice (separated only by mask); migration must keep
// them and tell the operator rather than deleting or merging by guesswork.
// Inputs: two Global 10.0.0.0/24 prefixes (one carrying an unresolved legacy
// VRF name), 10.0.0.5/24 and 10.0.0.5/25, and two Global VRFs named "blue".
// Outputs: four unresolved conflicts, no debug-only warnings and no errors.
// Data choice: the address pair differs only by mask, proving the key ignores
// it; the VRF pair exercises the ambiguity report.
func TestIPAMScopeReportsLegacyDuplicatesAsUnresolved(t *testing.T) {
	inv := NewInventory()
	p1, p2, a1, a2, v1, v2 := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Prefixes[p1] = &CaniPrefix{ID: p1, Prefix: "10.0.0.0/24", VRF: "stale"}
	inv.Prefixes[p2] = &CaniPrefix{ID: p2, Prefix: "10.0.0.0/24"}
	inv.IPAddresses[a1] = &CaniIPAddress{ID: a1, Host: "10.0.0.5", Address: "10.0.0.5/24"}
	inv.IPAddresses[a2] = &CaniIPAddress{ID: a2, Host: "10.0.0.5", Address: "10.0.0.5/25"}
	inv.VRFs[v1] = &CaniVRF{ID: v1, Name: "blue"}
	inv.VRFs[v2] = &CaniVRF{ID: v2, Name: "blue"}

	result := &RelationshipResult{}
	inv.validateIPAMScope(result)

	if len(result.Errors) != 0 || len(result.Warnings) != 0 {
		t.Fatalf("errors = %v, warnings = %v; want none for legacy duplicates", result.Errors, result.Warnings)
	}
	if len(result.Unresolved) != 4 {
		t.Fatalf("unresolved = %d (%v), want 4", len(result.Unresolved), result.Unresolved)
	}
	joined := strings.Join(result.Unresolved, "\n")
	for _, want := range []string{
		`prefix 10.0.0.0/24 exists twice in namespace "Global"`,
		`IP address 10.0.0.5 exists twice in namespace "Global"`,
		`VRF "blue" exists twice in namespace "Global"`,
		`legacy VRF name "stale" does not resolve`,
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("unresolved missing %q in:\n%s", want, joined)
		}
	}
}

// TestIPAMScopeSeparatesNamespaces verifies identical prefix and host text in
// two namespaces produces neither errors nor duplicate warnings.
//
// Why it matters: this is the discriminating case for the whole feature — the
// same 10.0.0.0/24 and 10.0.0.5 must coexist once per namespace.
// Inputs: 10.0.0.0/24 and a parented 10.0.0.5/24 in each of "tenant-a" and
// "tenant-b". Outputs: an empty RelationshipResult.
// Data choice: the addresses are parented so their namespace is derived from
// the prefix, exercising the authoritative-prefix rule.
func TestIPAMScopeSeparatesNamespaces(t *testing.T) {
	inv := NewInventory()
	for _, namespace := range []string{"tenant-a", "tenant-b"} {
		prefixID, addrID := uuid.New(), uuid.New()
		inv.Prefixes[prefixID] = &CaniPrefix{ID: prefixID, Prefix: "10.0.0.0/24", Namespace: namespace}
		inv.IPAddresses[addrID] = &CaniIPAddress{ID: addrID, Host: "10.0.0.5", Address: "10.0.0.5/24", Parent: prefixID}
	}

	result := &RelationshipResult{}
	inv.validateIPAMScope(result)

	if len(result.Errors) != 0 || len(result.Warnings) != 0 || len(result.Unresolved) != 0 {
		t.Fatalf("errors = %v, warnings = %v, unresolved = %v; want none", result.Errors, result.Warnings, result.Unresolved)
	}
}

// TestAddPrefixRejectsMembershipOutsideItsNamespace verifies AddPrefix refuses
// a VRF membership that is missing or belongs to another namespace.
//
// Why it matters: the model owns the namespace boundary, so every writer (CLI,
// provider, future API) gets the same refusal before the prefix is stored.
// Inputs: a tenant-a VRF; a Global prefix listing it, then a prefix listing an
// unknown VRF ID. Outputs: a namespace-mismatch error, a not-found error, and
// no stored prefix.
// Data choice: both prefixes are otherwise valid, so only the membership check
// can reject them.
func TestAddPrefixRejectsMembershipOutsideItsNamespace(t *testing.T) {
	inv := NewInventory()
	vrfID := uuid.New()
	inv.VRFs[vrfID] = &CaniVRF{ID: vrfID, Name: "blue", Namespace: "tenant-a"}

	foreign := inv.AddPrefix(&CaniPrefix{ID: uuid.New(), Prefix: "10.0.0.0/24", VRFs: []uuid.UUID{vrfID}})
	missing := inv.AddPrefix(&CaniPrefix{ID: uuid.New(), Prefix: "10.0.1.0/24", VRFs: []uuid.UUID{uuid.New()}})

	if foreign == nil || !strings.Contains(foreign.Error(), `VRF "blue" is in namespace "tenant-a", not "Global"`) {
		t.Errorf("foreign membership error = %v, want a namespace mismatch", foreign)
	}
	if missing == nil || !strings.Contains(missing.Error(), "not found") {
		t.Errorf("missing membership error = %v, want VRF not found", missing)
	}
	if len(inv.Prefixes) != 0 {
		t.Errorf("prefixes = %d, want none stored", len(inv.Prefixes))
	}
}
