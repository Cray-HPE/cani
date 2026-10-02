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

// loadVRFs exports CaniVRF records to Nautobot, creating each VRF that does not
// already exist in its namespace and caching the result by namespace and name
// so interface enrichment and prefix assignments can resolve it. It runs
// before the VLAN and interface-enrichment phases.
func (e *Exporter) loadVRFs(ctx context.Context, inventory *devicetypes.Inventory, result *LoadResult) error {
	if len(inventory.VRFs) == 0 {
		return nil
	}

	clog.Header("Phase 6c: VRFs (%d)", len(inventory.VRFs))

	for _, vrf := range inventory.VRFs {
		if vrf == nil || vrf.Name == "" {
			continue
		}
		namespace := vrf.EffectiveNamespace()

		existingID, err := e.findVRF(ctx, namespace, vrf.Name)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("vrf %s/%s: lookup error: %v", namespace, vrf.Name, err))
			continue
		}
		item := &CachedItem{ID: existingID, Name: vrf.Name}
		if existingID != uuid.Nil {
			result.VRFsSkipped++
		} else {
			nautobotID, err := e.createVRF(ctx, vrf)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("vrf %s/%s: create error: %v", namespace, vrf.Name, err))
				continue
			}
			// A VRF a dry run would create is cached as Planned so assignment
			// checks know it cannot hold anything remotely yet.
			item.ID, item.Planned = nautobotID, e.Options.DryRun
			result.VRFsCreated++
		}
		e.Cache.CacheVRF(namespace, vrf.Name, item)
		setExternalID(&vrf.ExternalIDs, "nautobot", item.ID)

		// Create VRF-device assignments for explicitly listed devices.
		e.loadVRFDeviceAssignments(ctx, vrf, item, inventory, result)
	}

	clog.Info("  VRFs created: %d", result.VRFsCreated)
	return nil
}

// loadVRFDeviceAssignments links a VRF to each device the operator listed.
func (e *Exporter) loadVRFDeviceAssignments(
	ctx context.Context, vrf *devicetypes.CaniVRF, remote *CachedItem,
	inventory *devicetypes.Inventory, result *LoadResult,
) {
	for _, devID := range vrf.Devices {
		dev, ok := inventory.Devices[devID]
		if !ok || dev == nil {
			clog.Skipped("  VRF %s: device %s not found in inventory", vrf.Name, devID)
			result.VRFDeviceAssignmentsSkipped++
			continue
		}
		nautobotDevID, ok := dev.ExternalIDs[externalIDKeyNautobot]
		if !ok || nautobotDevID == uuid.Nil {
			clog.Skipped("  VRF %s: device %q not in Nautobot; assignment skipped", vrf.Name, dev.Name)
			result.VRFDeviceAssignmentsSkipped++
			continue
		}
		if err := e.assignVRFToDevice(ctx, nautobotDevID, remote); err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("vrf %s: device assignment %s: %v", vrf.Name, dev.Name, err))
			continue
		}
		result.VRFDeviceAssignmentsCreated++
	}
}

// findVRF returns the Nautobot ID of the VRF with the given name in the
// namespace, or uuid.Nil when none exists. Several matches are an error
// rather than a first-result guess.
func (e *Exporter) findVRF(ctx context.Context, namespace, name string) (uuid.UUID, error) {
	nameFilter := []string{name}
	namespaceFilter := []string{namespace}
	resp, err := e.Client.IpamVrfsListWithResponse(ctx, &nautobotapi.IpamVrfsListParams{
		Name: &nameFilter, Namespace: &namespaceFilter,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return uuid.Nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
		return uuid.Nil, nil
	}
	if resp.JSON200.Count > 1 || len(resp.JSON200.Results) > 1 {
		return uuid.Nil, fmt.Errorf("%d VRFs named %q in namespace %q", resp.JSON200.Count, name, namespace)
	}
	return toUUID(resp.JSON200.Results[0].Id), nil
}

// createVRF creates a single VRF in Nautobot from a CaniVRF, resolving its
// namespace (defaulting to Global), status, and tags.
func (e *Exporter) createVRF(ctx context.Context, vrf *devicetypes.CaniVRF) (uuid.UUID, error) {
	req := nautobotapi.VRFRequest{Name: vrf.Name}
	if vrf.RD != "" {
		req.Rd = &vrf.RD
	}
	if vrf.Description != "" {
		req.Description = &vrf.Description
	}

	ns, err := e.Cache.GetOrCreateNamespace(vrf.EffectiveNamespace())
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve namespace %q: %w", vrf.EffectiveNamespace(), err)
	}
	if err := setRefID(&req.Namespace, ns.ID); err != nil {
		return uuid.Nil, fmt.Errorf("set VRF namespace reference: %w", err)
	}

	if vrf.Status != "" {
		if statusItem, err := e.Cache.GetStatus(vrf.Status); err == nil && statusItem != nil {
			if err := setRefID(&req.Status, statusItem.ID); err != nil {
				return uuid.Nil, fmt.Errorf("set VRF status reference: %w", err)
			}
		}
	}

	if err := setRefSlice(&req.Tags, e.Cache.resolveTagRefs(vrf.Tags)); err != nil {
		return uuid.Nil, fmt.Errorf("set VRF tag references: %w", err)
	}

	if e.Options.DryRun {
		clog.DryRun("Would create VRF: %s", vrf.Name)
		return uuid.New(), nil
	}

	resp, err := e.Client.IpamVrfsCreateWithResponse(ctx, &nautobotapi.IpamVrfsCreateParams{}, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}
	if resp.StatusCode() != http.StatusCreated || resp.JSON201 == nil {
		return uuid.Nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}

	clog.Created("  + VRF: %s", vrf.Name)
	return toUUID(resp.JSON201.Id), nil
}

// ensureVRFDeviceAssignment links the VRF an interface references to the
// interface's device. Nautobot rejects an interface VRF assignment with
// "VRF must be assigned to same Device" unless the VRF-device assignment already
// exists. vrfRef is "name" or "namespace/name"; an unknown VRF is a no-op so
// the caller can fall back to omitting the VRF, while an ambiguous bare name
// is an error.
func (e *Exporter) ensureVRFDeviceAssignment(ctx context.Context, deviceID uuid.UUID, vrfRef string) error {
	vrf, err := e.Cache.LookupCachedVRF(vrfRef)
	if err != nil || vrf == nil {
		return err
	}
	return e.assignVRFToDevice(ctx, deviceID, vrf)
}

// assignVRFToDevice is a find-or-create keyed on (vrf, device); an existing
// assignment is left untouched and a dry run makes no lookup.
func (e *Exporter) assignVRFToDevice(ctx context.Context, deviceID uuid.UUID, vrf *CachedItem) error {
	if e.Options.DryRun {
		return nil
	}

	vrfFilter := []string{vrf.ID.String()}
	deviceFilter := []string{deviceID.String()}
	listResp, err := e.Client.IpamVrfDeviceAssignmentsListWithResponse(ctx,
		&nautobotapi.IpamVrfDeviceAssignmentsListParams{Vrf: &vrfFilter, Device: &deviceFilter})
	if err != nil {
		return fmt.Errorf("API error: %w", err)
	}
	if listResp.JSON200 != nil && len(listResp.JSON200.Results) > 0 {
		return nil
	}

	req := nautobotapi.VRFDeviceAssignmentRequest{}
	if err := setRefID(&req.Device, deviceID); err != nil {
		return fmt.Errorf("set VRF assignment device reference: %w", err)
	}
	if err := setRefID(&req.Vrf, vrf.ID); err != nil {
		return fmt.Errorf("set VRF assignment VRF reference: %w", err)
	}
	createResp, err := e.Client.IpamVrfDeviceAssignmentsCreateWithResponse(ctx,
		&nautobotapi.IpamVrfDeviceAssignmentsCreateParams{}, req)
	if err != nil {
		return fmt.Errorf("API error: %w", err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("unexpected status %d: %s", createResp.StatusCode(), string(createResp.Body))
	}
	return nil
}
