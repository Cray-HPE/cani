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
package export

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// enrichInterfaces performs a second pass over device interfaces to apply
// settings that depend on other objects existing first: LAG membership (needs
// the parent LAG interface) and switchport mode plus untagged/tagged VLAN
// assignment (needs the VLANs created in the VLAN phase). Each interface that
// carries any of these settings is PATCHed.
func (e *Exporter) enrichInterfaces(
	ctx context.Context,
	inventory *devicetypes.Inventory,
	createdDeviceIDs map[string]uuid.UUID,
	createdVLANIDs map[uuid.UUID]uuid.UUID,
	result *LoadResult,
) error {
	vidToVLAN := buildVIDMap(inventory, createdVLANIDs)

	for deviceName, deviceID := range createdDeviceIDs {
		device := findDeviceByName(inventory, deviceName)
		if device == nil {
			continue
		}
		specs := getDeviceInterfaceSpecs(device)
		if !anyInterfaceNeedsEnrichment(specs) {
			continue
		}
		if err := e.Cache.PrefetchInterfacesForDevice(deviceID); err != nil {
			clog.Warn("Warning: failed to prefetch interfaces for %s: %v", deviceName, err)
		}
		for _, spec := range specs {
			if interfaceNeedsEnrichment(spec) {
				e.enrichOneInterface(ctx, deviceID, spec, vidToVLAN, result)
			}
		}
	}
	return nil
}

// buildVIDMap maps each VLAN ID to its Nautobot UUID using the cani-to-Nautobot
// map returned by the VLAN phase.
func buildVIDMap(inventory *devicetypes.Inventory, createdVLANIDs map[uuid.UUID]uuid.UUID) map[int]uuid.UUID {
	vidToVLAN := make(map[int]uuid.UUID, len(inventory.VLANs))
	for _, vlan := range inventory.VLANs {
		if vlan == nil {
			continue
		}
		if nid, ok := createdVLANIDs[vlan.ID]; ok {
			vidToVLAN[vlan.VID] = nid
		}
	}
	return vidToVLAN
}

// findDeviceByName returns the device with the given name, or nil.
func findDeviceByName(inventory *devicetypes.Inventory, name string) *devicetypes.CaniDeviceType {
	for _, d := range inventory.Devices {
		if d != nil && d.Name == name {
			return d
		}
	}
	return nil
}

// interfaceNeedsEnrichment reports whether a spec carries any setting that must
// be applied in the enrichment pass.
func interfaceNeedsEnrichment(spec interfaceSpec) bool {
	return spec.Lag != "" || spec.Mode != "" || spec.UntaggedVLAN != 0 || len(spec.TaggedVLANs) > 0
}

// anyInterfaceNeedsEnrichment reports whether any spec in the slice needs enrichment.
func anyInterfaceNeedsEnrichment(specs []interfaceSpec) bool {
	for _, s := range specs {
		if interfaceNeedsEnrichment(s) {
			return true
		}
	}
	return false
}

// enrichOneInterface resolves the interface and PATCHes it with the LAG, mode,
// and VLAN settings from its spec.
func (e *Exporter) enrichOneInterface(
	ctx context.Context,
	deviceID uuid.UUID,
	spec interfaceSpec,
	vidToVLAN map[int]uuid.UUID,
	result *LoadResult,
) {
	iface, err := e.Cache.GetInterfaceByDeviceAndName(deviceID, spec.Name)
	if err != nil || iface == nil {
		result.Errors = append(result.Errors, fmt.Sprintf("enrich %s: interface not found: %v", spec.Name, err))
		return
	}

	req, changed := e.buildInterfaceEnrichment(deviceID, spec, vidToVLAN)
	if !changed {
		return
	}

	if e.Options.DryRun {
		clog.DryRun("Would enrich interface: %s", spec.Name)
		return
	}

	resp, err := e.Client.DcimInterfacesPartialUpdateWithResponse(
		ctx, iface.ID, &nautobotapi.DcimInterfacesPartialUpdateParams{}, req)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("enrich %s: API error: %v", spec.Name, err))
		return
	}
	if resp.StatusCode() != http.StatusOK {
		result.Errors = append(result.Errors,
			fmt.Sprintf("enrich %s: unexpected status %d: %s", spec.Name, resp.StatusCode(), string(resp.Body)))
		return
	}
	clog.Changed("  ~ enriched interface: %s", spec.Name)
}

// buildInterfaceEnrichment assembles the PATCH request for LAG, mode, and VLAN
// settings, returning the request and whether any field was set.
func (e *Exporter) buildInterfaceEnrichment(
	deviceID uuid.UUID,
	spec interfaceSpec,
	vidToVLAN map[int]uuid.UUID,
) (nautobotapi.PatchedWritableInterfaceRequest, bool) {
	req := nautobotapi.PatchedWritableInterfaceRequest{}
	changed := false

	if ref := e.lagRef(deviceID, spec.Lag); ref != nil {
		req.Lag = ref
		changed = true
	}
	if ref := modeRef(spec.Mode); ref != nil {
		req.Mode = ref
		changed = true
	}
	if ref := untaggedVLANRef(spec.UntaggedVLAN, vidToVLAN); ref != nil {
		req.UntaggedVlan = ref
		changed = true
	}
	if tagged := resolveTaggedVLANRefs(spec.TaggedVLANs, vidToVLAN); len(tagged) > 0 {
		req.TaggedVlans = &tagged
		changed = true
	}
	return req, changed
}

// lagRef resolves the parent LAG interface on the same device to a ParentLAG
// reference, or nil when the LAG name is empty or cannot be resolved.
func (e *Exporter) lagRef(deviceID uuid.UUID, lagName string) *nautobotapi.ParentLAG {
	if lagName == "" {
		return nil
	}
	lagIface, err := e.Cache.GetInterfaceByDeviceAndName(deviceID, lagName)
	if err != nil || lagIface == nil {
		return nil
	}
	return makeParentLagRef(lagIface.ID)
}

// modeRef builds a switchport-mode reference, or nil when mode is empty or invalid.
func modeRef(mode string) *nautobotapi.PatchedWritableInterfaceRequestMode {
	if mode == "" {
		return nil
	}
	var m nautobotapi.PatchedWritableInterfaceRequestMode
	if err := m.FromModeEnum(nautobotapi.ModeEnum(mode)); err != nil {
		return nil
	}
	return &m
}

// untaggedVLANRef resolves a VLAN ID to an untagged-VLAN reference, or nil.
func untaggedVLANRef(vid int, vidToVLAN map[int]uuid.UUID) *nautobotapi.BulkWritableCircuitRequestTenant {
	if vid == 0 {
		return nil
	}
	if vlanID, ok := vidToVLAN[vid]; ok {
		return makeTenantRef(vlanID)
	}
	return nil
}

// resolveTaggedVLANRefs maps tagged VLAN IDs to Nautobot tagged-VLAN references,
// skipping any VLAN that was not created.
func resolveTaggedVLANRefs(vids []int, vidToVLAN map[int]uuid.UUID) []nautobotapi.TaggedVLANs {
	refs := make([]nautobotapi.TaggedVLANs, 0, len(vids))
	for _, vid := range vids {
		if vlanID, ok := vidToVLAN[vid]; ok {
			refs = append(refs, makeTaggedVLANRef(vlanID))
		}
	}
	return refs
}

// makeParentLagRef builds a ParentLAG reference from a Nautobot interface UUID.
func makeParentLagRef(id uuid.UUID) *nautobotapi.ParentLAG {
	var union nautobotapi.BulkWritableCableRequestStatusId
	_ = union.FromBulkWritableCableRequestStatusId0(id)
	return &nautobotapi.ParentLAG{Id: &union}
}

// makeTaggedVLANRef builds a TaggedVLANs reference from a Nautobot VLAN UUID.
func makeTaggedVLANRef(id uuid.UUID) nautobotapi.TaggedVLANs {
	var union nautobotapi.BulkWritableCableRequestStatusId
	_ = union.FromBulkWritableCableRequestStatusId0(id)
	return nautobotapi.TaggedVLANs{Id: &union}
}
