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
package transform

import (
	"slices"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// scopedRaw is the discriminating fixture: two namespaces that both hold
// 10.0.0.0/8 > 10.0.0.0/24 with a 10.0.0.5 host, plus a VRF "blue" in each,
// where only tenant-a's /24 is a member of its VRF.
type scopedRaw struct {
	namespaces  []nautobotapi.Namespace
	vrfs        []nautobotapi.VRF
	prefixes    []nautobotapi.Prefix
	addresses   []nautobotapi.IPAddress
	assignments []nautobotapi.VRFPrefixAssignment
	ids         map[string]uuid.UUID
}

func newScopedRaw() scopedRaw {
	raw := scopedRaw{ids: map[string]uuid.UUID{}}
	for _, namespace := range []string{"tenant-a", "tenant-b"} {
		nsID := uuid.New()
		raw.ids[namespace] = nsID
		raw.namespaces = append(raw.namespaces, nautobotapi.Namespace{Id: oaPtr(nsID), Name: namespace})

		vrfID, eightID, twentyFourID, hostID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
		raw.ids[namespace+"/vrf"], raw.ids[namespace+"/8"] = vrfID, eightID
		raw.ids[namespace+"/24"], raw.ids[namespace+"/host"] = twentyFourID, hostID

		vrf := nautobotapi.VRF{Id: oaPtr(vrfID), Name: "blue"}
		setNBRef(&vrf.Namespace, nsID)
		raw.vrfs = append(raw.vrfs, vrf)

		eight := nautobotapi.Prefix{Id: oaPtr(eightID), Prefix: "10.0.0.0/8"}
		setNBRef(&eight.Namespace, nsID)
		twentyFour := nautobotapi.Prefix{Id: oaPtr(twentyFourID), Prefix: "10.0.0.0/24"}
		setNBRef(&twentyFour.Namespace, nsID)
		setNBRef(&twentyFour.Parent, eightID)
		raw.prefixes = append(raw.prefixes, eight, twentyFour)

		host := nautobotapi.IPAddress{Id: oaPtr(hostID), Address: "10.0.0.5/24"}
		setNBRef(&host.Parent, twentyFourID)
		raw.addresses = append(raw.addresses, host)
	}
	assignment := nautobotapi.VRFPrefixAssignment{}
	setNBRef(&assignment.Vrf, raw.ids["tenant-a/vrf"])
	setNBRef(&assignment.Prefix, raw.ids["tenant-a/24"])
	raw.assignments = append(raw.assignments, assignment)
	return raw
}

// byNautobotID returns the mapped object carrying the given Nautobot ID.
func byNautobotID[T any](t *testing.T, items map[uuid.UUID]*T, externalIDs func(*T) map[string]uuid.UUID, nbID uuid.UUID) *T {
	t.Helper()
	for _, item := range items {
		if externalIDs(item)[externalIDKeyNautobot] == nbID {
			return item
		}
	}
	t.Fatalf("no mapped object carries Nautobot ID %s", nbID)
	return nil
}

// scopeAll runs the three scoping passes in production order and returns
// their combined warnings.
func scopeAll(raw scopedRaw, vrfs map[uuid.UUID]*devicetypes.CaniVRF, prefixes map[uuid.UUID]*devicetypes.CaniPrefix, addresses map[uuid.UUID]*devicetypes.CaniIPAddress) []string {
	namespaceNameMap := BuildNamespaceNameMap(raw.namespaces)
	warnings := ScopeVRFs(vrfs, raw.vrfs, namespaceNameMap)
	warnings = append(warnings, ScopePrefixes(prefixes, raw.prefixes, namespaceNameMap, vrfs, BuildVRFPrefixMembershipMap(raw.assignments))...)
	return append(warnings, ScopeIPAddresses(addresses, raw.addresses, prefixes)...)
}

// TestScopeIPAMPreservesNamespaceParentAndMemberships verifies the scoping
// pass stamps namespaces, keeps the Nautobot parent inside the namespace,
// derives the address namespace from its parent, and attaches VRF
// memberships only where an assignment record exists.
//
// Why it matters: the pre-namespace import dropped namespaces and recomputed
// parents across the whole inventory, so two tenants with identical address
// text collapsed into one hierarchy.
// Inputs: the two-tenant fixture with input order reversed. Outputs: no
// warnings; each /24 parented to its own tenant's /8, each host parented to
// its own /24 with the tenant namespace, blue VRFs in both namespaces, and a
// membership on tenant-a's /24 only.
// Data choice: identical CIDR, host and VRF text across tenants means only the
// namespace references can explain the separation.
func TestScopeIPAMPreservesNamespaceParentAndMemberships(t *testing.T) {
	raw := newScopedRaw()
	slices.Reverse(raw.prefixes)
	vrfs := MapVRFs(raw.vrfs, nil, nil, nil)
	prefixes := MapPrefixes(raw.prefixes, nil, nil, nil, nil, nil)
	addresses := MapIPAddresses(raw.addresses, nil, nil)

	warnings := scopeAll(raw, vrfs, prefixes, addresses)

	if len(warnings) != 0 {
		t.Fatalf("warnings = %q, want none", warnings)
	}
	for _, namespace := range []string{"tenant-a", "tenant-b"} {
		assertTenantHierarchy(t, prefixes, addresses, raw, namespace)
		if vrf := byNautobotID(t, vrfs, vrfExternalIDs, raw.ids[namespace+"/vrf"]); vrf.Namespace != namespace {
			t.Errorf("%s: VRF namespace = %q, want %q", namespace, vrf.Namespace, namespace)
		}
	}
	assertMembershipOnTenantAOnly(t, prefixes, vrfs, raw)
}

// TestScopeIPAMLeavesOutUnresolvedScopeButKeepsOmitted verifies an object
// whose namespace reference is unknown is left out and reported, while an
// object with no namespace reference at all keeps the Global default.
//
// Why it matters: an unresolved imported namespace is not evidence of Global;
// filing it there would merge a tenant network into the default one. A
// genuinely omitted reference, as in older fixtures, is the only case that
// may default.
// Inputs: a VRF and a prefix referencing an unknown namespace, a prefix with
// no namespace reference, and an address whose parent prefix was not
// imported. Outputs: three warnings naming the VRF, the prefix and the
// address; those three removed; the plain prefix retained with an empty
// (Global) namespace.
// Data choice: the four objects cover the VRF, prefix, omitted and
// missing-parent branches independently.
func TestScopeIPAMLeavesOutUnresolvedScopeButKeepsOmitted(t *testing.T) {
	unknownNS, plainID := uuid.New(), uuid.New()
	vrf := nautobotapi.VRF{Id: oaPtr(uuid.New()), Name: "blue"}
	setNBRef(&vrf.Namespace, unknownNS)
	unknown := nautobotapi.Prefix{Id: oaPtr(uuid.New()), Prefix: "10.1.0.0/24"}
	setNBRef(&unknown.Namespace, unknownNS)
	plain := nautobotapi.Prefix{Id: oaPtr(plainID), Prefix: "10.2.0.0/24"}
	orphan := nautobotapi.IPAddress{Id: oaPtr(uuid.New()), Address: "10.3.0.1/24"}
	setNBRef(&orphan.Parent, uuid.New())
	raw := scopedRaw{vrfs: []nautobotapi.VRF{vrf}, prefixes: []nautobotapi.Prefix{unknown, plain}, addresses: []nautobotapi.IPAddress{orphan}}
	vrfs := MapVRFs(raw.vrfs, nil, nil, nil)
	prefixes := MapPrefixes(raw.prefixes, nil, nil, nil, nil, nil)
	addresses := MapIPAddresses(raw.addresses, nil, nil)

	warnings := scopeAll(raw, vrfs, prefixes, addresses)

	want := []string{
		"VRF blue: namespace " + unknownNS.String() + " not found",
		"prefix 10.1.0.0/24: namespace " + unknownNS.String() + " not found",
		"IP address 10.3.0.1/24: parent prefix " + refID(orphan.Parent).String() + " not imported",
	}
	if !slices.Equal(warnings, want) {
		t.Errorf("warnings = %q, want %q", warnings, want)
	}
	if len(vrfs) != 0 || len(prefixes) != 1 || len(addresses) != 0 {
		t.Fatalf("vrfs/prefixes/addresses = %d/%d/%d; want 0/1/0", len(vrfs), len(prefixes), len(addresses))
	}
	if kept := byNautobotID(t, prefixes, prefixExternalIDs, plainID); kept.Namespace != "" {
		t.Errorf("omitted-namespace prefix namespace = %q, want empty (Global default)", kept.Namespace)
	}
}

// TestScopePrefixesDetachesChildOfLeftOutParent verifies a prefix whose
// Nautobot parent was left out of the result keeps no parent, so the merge
// derives one instead of rejecting the import.
//
// Why it matters: linking from an index built before the namespace pass left
// the child pointing at a removed record, and relationship validation then
// failed the whole import for one unresolvable parent.
// Inputs: a parent prefix in an unknown namespace and a child in a known one
// that references it. Outputs: one warning for the parent, the child retained
// with a nil parent, and a merge into an empty inventory that succeeds.
// Data choice: parent and child in different namespaces is the only shape
// that leaves one out and keeps the other.
func TestScopePrefixesDetachesChildOfLeftOutParent(t *testing.T) {
	knownNS, parentID, childID := uuid.New(), uuid.New(), uuid.New()
	parent := nautobotapi.Prefix{Id: oaPtr(parentID), Prefix: "10.0.0.0/8"}
	setNBRef(&parent.Namespace, uuid.New())
	child := nautobotapi.Prefix{Id: oaPtr(childID), Prefix: "10.0.0.0/24"}
	setNBRef(&child.Namespace, knownNS)
	setNBRef(&child.Parent, parentID)
	rawPrefixes := []nautobotapi.Prefix{parent, child}
	prefixes := MapPrefixes(rawPrefixes, nil, nil, nil, nil, nil)

	warnings := ScopePrefixes(prefixes, rawPrefixes, map[uuid.UUID]string{knownNS: "known"}, nil, nil)

	if len(warnings) != 1 {
		t.Fatalf("warnings = %q, want one for the parent", warnings)
	}
	if kept := byNautobotID(t, prefixes, prefixExternalIDs, childID); kept.Parent != uuid.Nil {
		t.Errorf("child parent = %v, want nil after its parent was left out", kept.Parent)
	}
	if _, err := devicetypes.NewInventory().MergeTransformResult(&devicetypes.TransformResult{Prefixes: prefixes}); err != nil {
		t.Errorf("MergeTransformResult() error = %v, want the detached child to merge", err)
	}
}

// TestScopeIPAMSurvivesMergeAndReimport verifies the scoped transform merges
// into an inventory as two separate hierarchies and that importing the same
// data again changes nothing.
//
// Why it matters: the integration spec covers this only against a live
// Nautobot; here the full Transform → MergeTransformResult path is exercised
// offline, including the second import a round-trip performs.
// Inputs: the two-tenant fixture as RawData, transformed and merged twice into
// one inventory. Outputs: no warnings; four prefixes, two addresses and two
// VRFs after each round, each /24 under its own /8 and each host under its
// own /24 with the tenant namespace, and the membership on tenant-a's /24
// only.
// Data choice: identical CIDR, host and VRF text across tenants means a merge
// that ignored the namespace would collide on every record.
func TestScopeIPAMSurvivesMergeAndReimport(t *testing.T) {
	raw := newScopedRaw()
	rawData := &RawData{
		Namespaces:           raw.namespaces,
		VRFs:                 raw.vrfs,
		Prefixes:             raw.prefixes,
		IPAddresses:          raw.addresses,
		VRFPrefixAssignments: raw.assignments,
		Statuses:             []nautobotapi.Status{{Id: oaPtr(uuid.New()), Name: "Active"}},
	}
	inv := devicetypes.NewInventory()
	for round := 1; round <= 2; round++ {
		result, err := Transform(*inv, rawData)
		if err != nil || len(result.Warnings) != 0 {
			t.Fatalf("round %d: Transform error = %v, warnings = %q", round, err, result.Warnings)
		}
		if _, err := inv.MergeTransformResult(result); err != nil {
			t.Fatalf("round %d: MergeTransformResult: %v", round, err)
		}
		assertTwoTenantInventory(t, inv, raw)
	}
}

// assertTwoTenantInventory checks the merged inventory holds exactly the two
// tenant hierarchies of the fixture with the membership on tenant-a only.
func assertTwoTenantInventory(t *testing.T, inv *devicetypes.Inventory, raw scopedRaw) {
	t.Helper()
	if len(inv.Prefixes) != 4 || len(inv.IPAddresses) != 2 || len(inv.VRFs) != 2 {
		t.Fatalf("prefixes/addresses/VRFs = %d/%d/%d, want 4/2/2", len(inv.Prefixes), len(inv.IPAddresses), len(inv.VRFs))
	}
	for _, namespace := range []string{"tenant-a", "tenant-b"} {
		assertTenantHierarchy(t, inv.Prefixes, inv.IPAddresses, raw, namespace)
	}
	assertMembershipOnTenantAOnly(t, inv.Prefixes, inv.VRFs, raw)
}

// assertTenantHierarchy checks one tenant's /8 > /24 > host chain, each record
// carrying the tenant namespace.
func assertTenantHierarchy(t *testing.T, prefixes map[uuid.UUID]*devicetypes.CaniPrefix, addresses map[uuid.UUID]*devicetypes.CaniIPAddress, raw scopedRaw, namespace string) {
	t.Helper()
	eight := byNautobotID(t, prefixes, prefixExternalIDs, raw.ids[namespace+"/8"])
	twentyFour := byNautobotID(t, prefixes, prefixExternalIDs, raw.ids[namespace+"/24"])
	host := byNautobotID(t, addresses, addressExternalIDs, raw.ids[namespace+"/host"])
	if twentyFour.Parent != eight.ID || host.Parent != twentyFour.ID {
		t.Errorf("%s: /24 parent = %v (want %v), host parent = %v (want %v)",
			namespace, twentyFour.Parent, eight.ID, host.Parent, twentyFour.ID)
	}
	if eight.Namespace != namespace || twentyFour.Namespace != namespace || host.Namespace != namespace {
		t.Errorf("%s: namespaces = %q/%q/%q, want all %q",
			namespace, eight.Namespace, twentyFour.Namespace, host.Namespace, namespace)
	}
}

// assertMembershipOnTenantAOnly checks tenant-a's /24 belongs to its blue VRF
// and tenant-b's /24, which has no assignment record, belongs to none.
func assertMembershipOnTenantAOnly(t *testing.T, prefixes map[uuid.UUID]*devicetypes.CaniPrefix, vrfs map[uuid.UUID]*devicetypes.CaniVRF, raw scopedRaw) {
	t.Helper()
	memberA := byNautobotID(t, prefixes, prefixExternalIDs, raw.ids["tenant-a/24"])
	vrfA := byNautobotID(t, vrfs, vrfExternalIDs, raw.ids["tenant-a/vrf"])
	if len(memberA.VRFs) != 1 || memberA.VRFs[0] != vrfA.ID {
		t.Errorf("tenant-a /24 VRFs = %v, want [%v]", memberA.VRFs, vrfA.ID)
	}
	if memberB := byNautobotID(t, prefixes, prefixExternalIDs, raw.ids["tenant-b/24"]); len(memberB.VRFs) != 0 {
		t.Errorf("tenant-b /24 VRFs = %v, want none (no assignment record)", memberB.VRFs)
	}
}
