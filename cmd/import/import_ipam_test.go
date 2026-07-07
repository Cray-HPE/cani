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
	"testing"

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
	mergeTransformResult(ctx, result)

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

	mergeTransformResult(ctx, ipamResult(uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil))
	mergeTransformResult(ctx, ipamResult(uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil))

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
}
