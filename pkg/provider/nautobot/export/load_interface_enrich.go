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
			clog.Warn("failed to prefetch interfaces for %s: %v", deviceName, err)
		}
		for _, spec := range specs {
			if interfaceNeedsEnrichment(spec) {
				e.enrichOneInterface(ctx, deviceID, spec, vidToVLAN, result)
			}
		}
	}
	if result.IfacesUnresolvedRefs > 0 {
		clog.Warn("%d interface reference(s) could not be resolved and were skipped during enrichment", result.IfacesUnresolvedRefs)
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
	return spec.Lag != "" || spec.Mode != "" || spec.UntaggedVLAN != 0 ||
		len(spec.TaggedVLANs) > 0 || spec.VRF != ""
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

	// A VRF referenced by an interface must first be assigned to the interface's
	// device, or Nautobot rejects the PATCH with "VRF must be assigned to same
	// Device." On failure, drop the VRF so mode/VLAN settings still apply.
	if spec.VRF != "" {
		if aerr := e.ensureVRFDeviceAssignment(ctx, deviceID, spec.VRF); aerr != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("enrich %s: vrf-device assignment: %v", spec.Name, aerr))
			spec.VRF = ""
		}
	}

	req, changed, unresolved := e.buildInterfaceEnrichment(deviceID, spec, vidToVLAN)
	result.IfacesUnresolvedRefs += unresolved
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
// settings, returning the request, whether any field was set, and the number of
// references that could not be resolved (and were warned about + skipped).
func (e *Exporter) buildInterfaceEnrichment(
	deviceID uuid.UUID,
	spec interfaceSpec,
	vidToVLAN map[int]uuid.UUID,
) (nautobotapi.PatchedWritableInterfaceRequest, bool, int) {
	req := nautobotapi.PatchedWritableInterfaceRequest{}
	changed := false
	unresolved := 0

	if ref := e.lagRef(deviceID, spec.Lag); ref != nil {
		req.Lag = ref
		changed = true
	} else if spec.Lag != "" {
		unresolved++
	}
	if ref := modeRef(spec.Mode); ref != nil {
		req.Mode = ref
		changed = true
	}
	if ref := untaggedVLANRef(spec.UntaggedVLAN, vidToVLAN); ref != nil {
		req.UntaggedVlan = ref
		changed = true
	} else if spec.UntaggedVLAN != 0 {
		unresolved++
	}
	tagged := resolveTaggedVLANRefs(spec.TaggedVLANs, vidToVLAN)
	if len(tagged) > 0 {
		req.TaggedVlans = &tagged
		changed = true
	}
	unresolved += len(spec.TaggedVLANs) - len(tagged)
	if ref := e.vrfRef(spec.VRF); ref != nil {
		req.Vrf = ref
		changed = true
	} else if spec.VRF != "" {
		unresolved++
	}

	if spec.Description != "" {
		desc := spec.Description
		req.Description = &desc
	}

	// Preserve the interface role during enrichment so that adding LAG/VRF
	// settings does not clear the role previously set during creation.
	if ref := e.roleRef(spec.Role); ref != nil {
		req.Role = ref
	}

	// Nautobot's interface serializer validates that a device (or module) is
	// set even on a partial update; omitting it fails with "Either device or
	// module must be set". Include the device reference whenever we PATCH.
	if changed {
		req.Device = makeTenantRef(deviceID)
	}
	return req, changed, unresolved
}
