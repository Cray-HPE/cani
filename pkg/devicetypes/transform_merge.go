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
	"encoding/json"
	"fmt"
)

// TransformMergeSummary describes the canonical UUIDs retained while merging a
// provider transform result into an inventory.
type TransformMergeSummary struct {
	Remaps ReferenceRemaps
}

// MergeTransformResult atomically merges a complete provider transform result.
// Both inputs are cloned before mutation; the receiver is replaced only after
// every entity has merged and all relationships validate.
func (inv *Inventory) MergeTransformResult(result *TransformResult) (*TransformMergeSummary, error) {
	if inv == nil {
		return nil, fmt.Errorf("inventory must not be nil")
	}
	if result == nil {
		return nil, fmt.Errorf("transform result must not be nil")
	}
	working, err := cloneInventory(inv)
	if err != nil {
		return nil, fmt.Errorf("clone inventory: %w", err)
	}
	incoming, err := cloneTransformResult(result)
	if err != nil {
		return nil, fmt.Errorf("clone transform result: %w", err)
	}

	incoming.EnsureUniqueDeviceNames()
	mergeTransformMetadata(working, incoming.Metadata)
	remaps := ReferenceRemaps{Locations: working.MergeLocations(incoming.Locations)}
	incoming.RemapReferences(working, remaps)
	remaps.Racks = working.MergeRacks(incoming.Racks)
	incoming.RemapReferences(working, remaps)
	remaps.Devices, _ = working.MergeDevicesStrict(incoming.Devices, false)
	incoming.RemapReferences(working, remaps)
	remaps.Modules = working.MergeModules(incoming.Modules)
	incoming.RemapReferences(working, remaps)
	remaps.Frus = working.MergeFrus(incoming.Frus)
	incoming.RemapReferences(working, remaps)
	remaps.Cables = working.MergeCables(incoming.Cables)
	incoming.RemapReferences(working, remaps)
	remaps.VLANs = working.MergeVLANs(incoming.VLANs)
	incoming.RemapReferences(working, remaps)
	incoming.DerivePrefixParents(working.Prefixes)
	remaps.Prefixes = working.MergePrefixes(incoming.Prefixes)
	incoming.RemapReferences(working, remaps)
	incoming.DeriveIPAddressParents(working.Prefixes)
	remaps.IPAddresses = working.MergeIPAddresses(incoming.IPAddresses)
	incoming.RemapReferences(working, remaps)
	remaps.VRFs = working.MergeVRFs(incoming.VRFs)
	incoming.RemapReferences(working, remaps)

	if relationshipResult := working.RebuildDerivedState(); relationshipResult.Err() != nil {
		return nil, fmt.Errorf("relationship validation failed after merge: %w", relationshipResult.Err())
	}
	working.RebuildProviderKeyIndex()
	*inv = *working
	return &TransformMergeSummary{Remaps: remaps}, nil
}

func cloneInventory(inventory *Inventory) (*Inventory, error) {
	data, err := json.Marshal(inventory)
	if err != nil {
		return nil, err
	}
	clone := NewInventory()
	if err := json.Unmarshal(data, clone); err != nil {
		return nil, err
	}
	restoreInventoryDynamicFields(inventory, clone)
	clone.RebuildProviderKeyIndex()
	return clone, nil
}

func cloneTransformResult(result *TransformResult) (*TransformResult, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	clone := &TransformResult{}
	if err := json.Unmarshal(data, clone); err != nil {
		return nil, err
	}
	restoreTransformDynamicFields(result, clone)
	return clone, nil
}

func mergeTransformMetadata(inventory *Inventory, metadata *InventoryMetadata) {
	if metadata == nil {
		return
	}
	for _, role := range metadata.Roles {
		_ = inventory.AddMetadata("roles", role)
	}
	for _, status := range metadata.Statuses {
		_ = inventory.AddMetadata("statuses", status)
	}
	for _, tag := range metadata.Tags {
		_ = inventory.AddMetadata("tags", tag)
	}
}
