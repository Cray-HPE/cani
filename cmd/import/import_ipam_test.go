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
package imprt

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// ipamResult builds a TransformResult populated with one of each IPAM entity.
// Passing uuid.Nil for any id allocates a fresh UUID, letting callers reuse the
// same natural keys (VID, name, CIDR, address) across two results to exercise
// the merge dedup path.
func ipamResult(vlanID, prefixID, ipID, vrfID uuid.UUID) *devicetypes.TransformResult {
	orNew := func(id uuid.UUID) uuid.UUID {
		if id == uuid.Nil {
			return uuid.New()
		}
		return id
	}
	vlanID, prefixID, ipID, vrfID = orNew(vlanID), orNew(prefixID), orNew(ipID), orNew(vrfID)
	return &devicetypes.TransformResult{
		VLANs: map[uuid.UUID]*devicetypes.CaniVLAN{
			vlanID: {ID: vlanID, VID: 100, Name: "mgmt"},
		},
		Prefixes: map[uuid.UUID]*devicetypes.CaniPrefix{
			prefixID: {ID: prefixID, Prefix: "10.0.0.0/24"},
		},
		IPAddresses: map[uuid.UUID]*devicetypes.CaniIPAddress{
			ipID: {ID: ipID, Host: "10.0.0.1", Address: "10.0.0.1/24"},
		},
		VRFs: map[uuid.UUID]*devicetypes.CaniVRF{
			vrfID: {ID: vrfID, Name: "LEGACY"},
		},
	}
}

// TestMergeTransformResult_MergesIPAM verifies that transformed IPAM entities
// (VLANs, prefixes, IP addresses, VRFs) are merged into the inventory rather
// than silently dropped.
//
// Why it matters: the transform stage produces these maps on `cani import`, but
// before this wiring mergeTransformResult only merged DCIM entities, so all
// imported IPAM was discarded. This guards the completion of that pipeline.
//
// Inputs/outputs: given a TransformResult holding one VLAN, prefix, IP address,
// and VRF, after mergeTransformResult the inventory must contain each by ID.
//
// Data choice: a single, minimally-valid instance of each IPAM type isolates
// the merge wiring from unrelated DCIM merge logic.
func TestMergeTransformResult_MergesIPAM(t *testing.T) {
	vlanID, prefixID, ipID, vrfID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	result := ipamResult(vlanID, prefixID, ipID, vrfID)

	ctx := &etlContext{inventory: devicetypes.NewInventory()}
	if err := mergeTransformResult(ctx, result); err != nil {
		t.Fatalf("mergeTransformResult() unexpected error: %v", err)
	}

	if _, ok := ctx.inventory.VLANs[vlanID]; !ok {
		t.Errorf("VLAN %s was not merged into inventory", vlanID)
	}
	if _, ok := ctx.inventory.Prefixes[prefixID]; !ok {
		t.Errorf("prefix %s was not merged into inventory", prefixID)
	}
	if _, ok := ctx.inventory.IPAddresses[ipID]; !ok {
		t.Errorf("IP address %s was not merged into inventory", ipID)
	}
	if _, ok := ctx.inventory.VRFs[vrfID]; !ok {
		t.Errorf("VRF %s was not merged into inventory", vrfID)
	}
}

// TestMergeTransformResult_IPAMIdempotent verifies that re-importing the same
// IPAM (fresh UUIDs but identical natural keys) updates in place instead of
// creating duplicates.
//
// Why it matters: transform assigns new UUIDs on every run, so a UUID-only
// merge would multiply IPAM on each import. The natural-key merge (VID+Location,
// CIDR+VRF, address, VRF name) keeps re-imports stable.
//
// Inputs/outputs: two TransformResults with different UUIDs but the same VLAN
// VID, prefix CIDR, IP address, and VRF name; after merging both, each map must
// hold exactly one entry.
//
// Data choice: distinct UUIDs with shared natural keys directly exercise the
// dedup branch that a UUID-only merge would miss.
func TestMergeTransformResult_IPAMIdempotent(t *testing.T) {
	ctx := &etlContext{inventory: devicetypes.NewInventory()}

	first := ipamResult(uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil)
	var firstVRFID uuid.UUID
	for firstVRFID = range first.VRFs {
	}
	if err := mergeTransformResult(ctx, first); err != nil {
		t.Fatalf("first mergeTransformResult() unexpected error: %v", err)
	}
	if err := mergeTransformResult(ctx, ipamResult(uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil)); err != nil {
		t.Fatalf("second mergeTransformResult() unexpected error: %v", err)
	}

	if got := len(ctx.inventory.VLANs); got != 1 {
		t.Errorf("VLANs: got %d, want 1 (re-import duplicated)", got)
	}
	if got := len(ctx.inventory.Prefixes); got != 1 {
		t.Errorf("Prefixes: got %d, want 1 (re-import duplicated)", got)
	}
	if got := len(ctx.inventory.IPAddresses); got != 1 {
		t.Errorf("IPAddresses: got %d, want 1 (re-import duplicated)", got)
	}
	if got := len(ctx.inventory.VRFs); got != 1 {
		t.Errorf("VRFs: got %d, want 1 (re-import duplicated)", got)
	}
	if _, ok := ctx.inventory.VRFs[firstVRFID]; !ok {
		t.Errorf("VRF canonical UUID %s was not retained", firstVRFID)
	}
}

func TestMergeTransformResultDerivesIPAMParents(t *testing.T) {
	parentID, childID, addressID := uuid.New(), uuid.New(), uuid.New()
	result := &devicetypes.TransformResult{
		Prefixes: map[uuid.UUID]*devicetypes.CaniPrefix{
			parentID: {ID: parentID, Prefix: "10.20.0.0/16", PrefixLen: 16},
			childID:  {ID: childID, Prefix: "10.20.1.0/24", PrefixLen: 24},
		},
		IPAddresses: map[uuid.UUID]*devicetypes.CaniIPAddress{
			addressID: {ID: addressID, Address: "10.20.1.10/24", Host: "10.20.1.10", MaskLength: 24},
		},
	}
	inventory := devicetypes.NewInventory()

	if err := mergeTransformResult(&etlContext{inventory: inventory}, result); err != nil {
		t.Fatalf("mergeTransformResult() error = %v", err)
	}
	if inventory.Prefixes[childID].Parent != parentID {
		t.Errorf("child prefix parent = %s, want %s", inventory.Prefixes[childID].Parent, parentID)
	}
	if inventory.IPAddresses[addressID].Parent != childID {
		t.Errorf("address parent = %s, want %s", inventory.IPAddresses[addressID].Parent, childID)
	}
}

// TestMergeTransformResultRejectsInvalidCable verifies transformed inventory
// with an unresolved cable port fails before the import Load phase.
//
// Why it matters: import merges in memory before persistence, so propagating
// this error keeps the existing datastore atomic and prevents partial imports.
// Inputs: firewall-01 with port1 and an imported cable naming missing port3.
// Outputs: a contextual relationship error from mergeTransformResult.
// Data choice: a valid device reference isolates the exact port-resolution
// defect that previously survived import and failed during provider export.
func TestMergeTransformResultRejectsInvalidCable(t *testing.T) {
	originalConfig := config.Cfg
	config.Cfg = &config.Config{}
	t.Cleanup(func() { config.Cfg = originalConfig })

	deviceID := uuid.New()
	cableID := uuid.New()
	ctx := &etlContext{inventory: devicetypes.NewInventory()}
	result := &devicetypes.TransformResult{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			deviceID: {
				ID: deviceID, Name: "firewall-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "port1"}},
			},
		},
		Cables: map[uuid.UUID]*devicetypes.CaniCableType{
			cableID: {
				ID:                 cableID,
				Label:              "invalid-uplink",
				TerminationBDevice: deviceID,
				TerminationBPort:   "port3",
			},
		},
	}

	err := mergeTransformResult(ctx, result)

	if err == nil {
		t.Fatal("expected invalid cable relationship to fail transform merge")
	}
	want := `termination B port "port3" not found on device "firewall-01"`
	if !strings.Contains(err.Error(), want) {
		t.Errorf("mergeTransformResult() error = %q, want it to contain %q", err, want)
	}
	if len(ctx.inventory.Devices) != 0 || len(ctx.inventory.Cables) != 0 {
		t.Fatalf("failed merge mutated inventory: devices=%d cables=%d", len(ctx.inventory.Devices), len(ctx.inventory.Cables))
	}
}

// conflictingIPAM returns an inventory holding one record of kind under one
// source ID and a transform result carrying the same natural key under
// another, the shape of an object the source deleted and recreated. It also
// returns the record's UUID.
func conflictingIPAM(kind devicetypes.IPAMKind) (*devicetypes.Inventory, *devicetypes.TransformResult, uuid.UUID) {
	existing, incoming := uuid.New(), uuid.New()
	sourced := func() devicetypes.ObjectMeta {
		return devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"source": uuid.New()}}
	}
	inventory, result := devicetypes.NewInventory(), &devicetypes.TransformResult{}
	switch kind {
	case devicetypes.IPAMKindPrefix:
		inventory.Prefixes[existing] = &devicetypes.CaniPrefix{ID: existing, Prefix: "10.0.0.0/24", ObjectMeta: sourced()}
		result.Prefixes = map[uuid.UUID]*devicetypes.CaniPrefix{incoming: {ID: incoming, Prefix: "10.0.0.0/24", ObjectMeta: sourced()}}
	case devicetypes.IPAMKindIPAddress:
		inventory.IPAddresses[existing] = &devicetypes.CaniIPAddress{ID: existing, Host: "10.0.0.1", Address: "10.0.0.1/24", ObjectMeta: sourced()}
		result.IPAddresses = map[uuid.UUID]*devicetypes.CaniIPAddress{incoming: {ID: incoming, Host: "10.0.0.1", Address: "10.0.0.1/24", ObjectMeta: sourced()}}
	case devicetypes.IPAMKindVRF:
		inventory.VRFs[existing] = &devicetypes.CaniVRF{ID: existing, Name: "blue", ObjectMeta: sourced()}
		result.VRFs = map[uuid.UUID]*devicetypes.CaniVRF{incoming: {ID: incoming, Name: "blue", ObjectMeta: sourced()}}
	}
	return inventory, result, existing
}

// TestMergeTransformResultNamesTheRemoveCommandForAConflict verifies an import
// whose merge meets a source-identity conflict fails with the exact remove
// command for the record, for each IPAM kind, while the error still unwraps
// to the model's sentinel and the record is left as it was.
//
// Why it matters: the model can only say "remove the record and import again"
// and name a UUID; which subcommand removes it is the command layer's to say,
// and a wrong noun would send the operator to a command that rejects the UUID.
// Inputs: one prefix, address or VRF recorded under one source ID and a
// transform result with the same natural key under another.
// Outputs: an error containing "cani alpha remove <prefix|ip|vrf> <record>"
// that errors.Is ErrSourceIdentityConflict; the inventory still holds one
// record of the kind.
// Data choice: the three kinds map to three different nouns, so one case per
// kind proves the mapping and not a default.
func TestMergeTransformResultNamesTheRemoveCommandForAConflict(t *testing.T) {
	cases := []struct {
		kind devicetypes.IPAMKind
		noun string
	}{
		{devicetypes.IPAMKindPrefix, "prefix"},
		{devicetypes.IPAMKindIPAddress, "ip"},
		{devicetypes.IPAMKindVRF, "vrf"},
	}
	for _, tc := range cases {
		t.Run(string(tc.kind), func(t *testing.T) {
			// Arrange.
			inventory, result, record := conflictingIPAM(tc.kind)
			before := len(inventory.Prefixes) + len(inventory.IPAddresses) + len(inventory.VRFs)

			// Act.
			err := mergeTransformResult(&etlContext{inventory: inventory}, result)

			// Assert.
			want := fmt.Sprintf("cani alpha remove %s %s", tc.noun, record)
			if !errors.Is(err, devicetypes.ErrSourceIdentityConflict) || !strings.Contains(err.Error(), want) {
				t.Fatalf("mergeTransformResult() error = %v; want ErrSourceIdentityConflict naming %q", err, want)
			}
			if strings.Count(err.Error(), "cani alpha remove") != 1 {
				t.Errorf("error names the remove command more than once: %v", err)
			}
			if after := len(inventory.Prefixes) + len(inventory.IPAddresses) + len(inventory.VRFs); after != before {
				t.Errorf("IPAM records = %d after the refused merge, want %d", after, before)
			}
		})
	}
}
