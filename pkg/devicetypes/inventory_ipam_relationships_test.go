package devicetypes

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestIPAMRelationshipsAcceptValidUUIDsAndDeferredNames verifies inventory
// validation distinguishes local UUID references from deferred interface names.
//
// Why it matters: provider-only VRF/LAG names remain valid until export.
// Inputs: a complete local IPAM graph and an interface with external-only names.
// Outputs: both validation entry points succeed without requiring named objects.
// Data choice: UUID-linked IP/VLAN/VRF objects exercise every validated collection.
func TestIPAMRelationshipsAcceptValidUUIDsAndDeferredNames(t *testing.T) {
	fixture := newIPAMReferenceFixture(t)

	if err := fixture.inventory.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
	if err := fixture.inventory.RebuildDerivedState().Err(); err != nil {
		t.Errorf("RebuildDerivedState() error = %v", err)
	}
}

// TestIPAMRelationshipsRejectDanglingUUIDs verifies local foreign keys cannot
// silently reference missing or nil objects through either validator.
//
// Why it matters: load/rebuild and transform transactions share the same graph.
// Inputs: one invalid FK at a time in an otherwise valid inventory.
// Outputs: an error naming the broken relationship from each entry point.
// Data choice: zero list entries, nil targets, and self-cycles cover non-obvious
// cases that a map-key-existence check alone would accept.
func TestIPAMRelationshipsRejectDanglingUUIDs(t *testing.T) {
	missing := uuid.New()
	cases := []struct {
		name   string
		mutate func(ipamReferenceFixture)
		want   string
	}{
		{"vlan location", func(fixture ipamReferenceFixture) { fixture.vlan.Location = missing }, "location"},
		{"prefix location", func(fixture ipamReferenceFixture) { fixture.prefix.Location = missing }, "location"},
		{"prefix vlan", func(fixture ipamReferenceFixture) { fixture.prefix.VLAN = missing }, "VLAN"},
		{"prefix parent", func(fixture ipamReferenceFixture) { fixture.prefix.Parent = missing }, "parent prefix"},
		{"prefix self-cycle", func(fixture ipamReferenceFixture) { fixture.prefix.Parent = fixture.prefix.ID }, "circular"},
		{"IP parent", func(fixture ipamReferenceFixture) { fixture.address.Parent = missing }, "parent prefix"},
		{"IP interface", func(fixture ipamReferenceFixture) { fixture.address.Interfaces = []uuid.UUID{missing} }, "interface"},
		{"IP nil interface", func(fixture ipamReferenceFixture) { fixture.address.Interfaces = []uuid.UUID{uuid.Nil} }, "interface"},
		{"IP NAT", func(fixture ipamReferenceFixture) { fixture.address.NATInside = missing }, "NAT inside"},
		{"VRF device", func(fixture ipamReferenceFixture) { fixture.vrf.Devices = []uuid.UUID{missing} }, "device"},
		{"VRF nil device", func(fixture ipamReferenceFixture) { fixture.vrf.Devices = []uuid.UUID{uuid.Nil} }, "device"},
		{"primary IPv4", func(fixture ipamReferenceFixture) { fixture.device.PrimaryIPv4 = missing }, "primary IPv4"},
		{"primary IPv6", func(fixture ipamReferenceFixture) { fixture.device.PrimaryIPv6 = missing }, "primary IPv6"},
		{"device VLAN", func(fixture ipamReferenceFixture) { fixture.device.AssignedVLANs = []uuid.UUID{missing} }, "VLAN"},
		{"nil VLAN target", func(fixture ipamReferenceFixture) { fixture.inventory.VLANs[fixture.vlan.ID] = nil }, "VLAN"},
		{"nil IP target", func(fixture ipamReferenceFixture) { fixture.inventory.IPAddresses[fixture.address.ID] = nil }, "primary IPv4"},
	}
	for _, testCase := range cases {
		fixture := newIPAMReferenceFixture(t)
		testCase.mutate(fixture)

		assertIPAMReferenceError(t, testCase.name+" Validate", fixture.inventory.Validate(), testCase.want)
		assertIPAMReferenceError(t, testCase.name+" Rebuild", fixture.inventory.RebuildDerivedState().Err(), testCase.want)
	}
}

// TestMergeTransformRejectsDanglingIPAMAtomically verifies the complete-result
// transaction rejects an assigned interface UUID that is not in the inventory.
//
// Why it matters: invalid imported network state must never be committed.
// Inputs: a valid inventory and a new IP assigned to a missing interface.
// Outputs: a relationship error with both the inventory and input unchanged.
// Data choice: a valid address isolates the missing-FK validation failure.
func TestMergeTransformRejectsDanglingIPAMAtomically(t *testing.T) {
	fixture := newIPAMReferenceFixture(t)
	addressID, missingInterface := uuid.New(), uuid.New()
	result := &TransformResult{IPAddresses: map[uuid.UUID]*CaniIPAddress{
		addressID: {ID: addressID, Address: "10.0.0.2/24", Interfaces: []uuid.UUID{missingInterface}},
	}}
	before := inventoryJSONForReferenceTest(t, fixture.inventory)
	inputBefore := inventoryJSONForReferenceTest(t, result)

	_, err := fixture.inventory.MergeTransformResult(result)

	assertIPAMReferenceError(t, "MergeTransformResult", err, missingInterface.String())
	if got := inventoryJSONForReferenceTest(t, fixture.inventory); got != before {
		t.Error("rejected merge changed the live inventory")
	}
	if got := inventoryJSONForReferenceTest(t, result); got != inputBefore {
		t.Error("rejected merge mutated its transform input")
	}
}

// TestMergeTransformRejectsConflictingIPAMIdentityAtomically verifies the
// complete-result transaction rolls back when an incoming prefix shares a
// natural key with an existing prefix but carries a different source UUID.
//
// Why it matters: the conflict surfaces in the IPAM merge stage, after
// locations, racks, devices and VLANs have already merged into the working
// copy; the receiver must still be untouched.
// Inputs: the reference fixture (Global 10.0.0.0/24 with a nautobot external
// ID) and a transform result with the same key under another nautobot UUID.
// Outputs: an error naming the source-identity conflict and byte-identical
// inventory and input.
// Data choice: giving the existing prefix an external ID is what turns a plain
// natural-key match into a conflict.
func TestMergeTransformRejectsConflictingIPAMIdentityAtomically(t *testing.T) {
	fixture := newIPAMReferenceFixture(t)
	fixture.prefix.ExternalIDs = map[string]uuid.UUID{"nautobot": uuid.New()}
	incomingID := uuid.New()
	result := &TransformResult{Prefixes: map[uuid.UUID]*CaniPrefix{
		incomingID: {ID: incomingID, Prefix: "10.0.0.0/24", ObjectMeta: ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": uuid.New()}}},
	}}
	before := inventoryJSONForReferenceTest(t, fixture.inventory)
	inputBefore := inventoryJSONForReferenceTest(t, result)

	_, err := fixture.inventory.MergeTransformResult(result)

	assertIPAMReferenceError(t, "MergeTransformResult", err, "different source identity")
	if got := inventoryJSONForReferenceTest(t, fixture.inventory); got != before {
		t.Error("rejected merge changed the live inventory")
	}
	if got := inventoryJSONForReferenceTest(t, result); got != inputBefore {
		t.Error("rejected merge mutated its transform input")
	}
}

// TestMergeTransformKeepsNamespacesSeparate verifies a transform result that
// repeats the fixture's prefix and host text in another namespace is merged
// as new objects whose relationships stay inside that namespace.
//
// Why it matters: this is the acceptance case for combining inventories from
// independent networks — equal object text must not merge, and the new
// address must parent to the new prefix, not the Global one.
// Inputs: the reference fixture plus a tenant-a 10.0.0.0/24 with a tenant-a
// VRF membership and a 10.0.0.1/24 address intended for tenant-a. Outputs:
// two prefixes, two addresses, two VRFs; the new address parented to the
// tenant-a prefix; the membership remapped to the merged tenant-a VRF.
// Data choice: the VRF carries the same name as the fixture's Global VRF so a
// name-only merge would wrongly collapse it.
func TestMergeTransformKeepsNamespacesSeparate(t *testing.T) {
	fixture := newIPAMReferenceFixture(t)
	vrfID, prefixID, addressID := uuid.New(), uuid.New(), uuid.New()
	result := &TransformResult{
		VRFs:     map[uuid.UUID]*CaniVRF{vrfID: {ID: vrfID, Name: "local", Namespace: "tenant-a"}},
		Prefixes: map[uuid.UUID]*CaniPrefix{prefixID: {ID: prefixID, Prefix: "10.0.0.0/24", Namespace: "tenant-a", VRFs: []uuid.UUID{vrfID}}},
		IPAddresses: map[uuid.UUID]*CaniIPAddress{
			addressID: {ID: addressID, Host: "10.0.0.1", Address: "10.0.0.1/24", Namespace: "tenant-a"},
		},
	}

	summary, err := fixture.inventory.MergeTransformResult(result)
	if err != nil {
		t.Fatalf("MergeTransformResult returned error: %v", err)
	}

	inv := fixture.inventory
	if len(inv.Prefixes) != 2 || len(inv.IPAddresses) != 2 || len(inv.VRFs) != 2 {
		t.Fatalf("counts = prefixes %d, addresses %d, vrfs %d; want 2 each", len(inv.Prefixes), len(inv.IPAddresses), len(inv.VRFs))
	}
	mergedPrefix := inv.Prefixes[summary.Remaps.Prefixes[prefixID]]
	mergedAddress := inv.IPAddresses[summary.Remaps.IPAddresses[addressID]]
	mergedVRF := summary.Remaps.VRFs[vrfID]
	if mergedPrefix == nil || mergedAddress == nil || inv.VRFs[mergedVRF] == nil {
		t.Fatal("merged tenant-a objects are missing from the inventory")
	}
	if mergedAddress.Parent != mergedPrefix.ID {
		t.Errorf("tenant-a address parent = %v, want the tenant-a prefix %v, not the Global one", mergedAddress.Parent, mergedPrefix.ID)
	}
	if len(mergedPrefix.VRFs) != 1 || mergedPrefix.VRFs[0] != mergedVRF {
		t.Errorf("tenant-a prefix VRFs = %v, want [%v]", mergedPrefix.VRFs, mergedVRF)
	}
	if fixture.address.Parent != fixture.prefix.ID {
		t.Errorf("Global address parent changed to %v", fixture.address.Parent)
	}
}

type ipamReferenceFixture struct {
	inventory *Inventory
	device    *CaniDeviceType
	prefix    *CaniPrefix
	address   *CaniIPAddress
	vlan      *CaniVLAN
	vrf       *CaniVRF
}

func newIPAMReferenceFixture(t *testing.T) ipamReferenceFixture {
	t.Helper()
	inventory := NewInventory()
	locationID, deviceID, interfaceID := uuid.New(), uuid.New(), uuid.New()
	vlanID, prefixID, addressID, vrfID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inventory.Locations[locationID] = &CaniLocationType{ID: locationID, Name: "site"}
	device := &CaniDeviceType{
		ID: deviceID, Name: "switch", PrimaryIPv4: addressID, AssignedVLANs: []uuid.UUID{vlanID},
		Interfaces: []InterfaceSpec{{ID: interfaceID, Name: "port1", VRF: "external-only", Lag: "external-bond", TaggedVLANs: []int{999}}},
	}
	prefix := &CaniPrefix{ID: prefixID, Prefix: "10.0.0.0/24", VLAN: vlanID, Location: locationID}
	address := &CaniIPAddress{ID: addressID, Address: "10.0.0.1/24", Parent: prefixID, Interfaces: []uuid.UUID{interfaceID}}
	vlan := &CaniVLAN{ID: vlanID, Name: "management", VID: 10, Location: locationID}
	vrf := &CaniVRF{ID: vrfID, Name: "local", Devices: []uuid.UUID{deviceID}}
	inventory.Devices[deviceID] = device
	inventory.Prefixes[prefixID] = prefix
	inventory.IPAddresses[addressID] = address
	inventory.VLANs[vlanID] = vlan
	inventory.VRFs[vrfID] = vrf
	if err := inventory.RebuildDerivedState().Err(); err != nil {
		t.Fatalf("invalid fixture: %v", err)
	}
	return ipamReferenceFixture{inventory, device, prefix, address, vlan, vrf}
}

func assertIPAMReferenceError(t *testing.T, name string, err error, want string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("%s error = %v, want %q", name, err, want)
	}
}

func inventoryJSONForReferenceTest(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal inventory: %v", err)
	}
	return string(data)
}
