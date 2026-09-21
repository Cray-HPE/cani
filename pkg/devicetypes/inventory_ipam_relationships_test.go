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
