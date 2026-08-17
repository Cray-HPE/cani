/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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

// loadVLANs exports CaniVLAN records to Nautobot.
// Returns a map of cani VLAN ID → Nautobot VLAN ID for downstream FK resolution.
func (e *Exporter) loadVLANs(
	ctx context.Context,
	inventory *devicetypes.Inventory,
	locationMap map[uuid.UUID]uuid.UUID,
	result *LoadResult,
) (map[uuid.UUID]uuid.UUID, error) {
	created := make(map[uuid.UUID]uuid.UUID)

	if len(inventory.VLANs) == 0 {
		return created, nil
	}

	clog.Header("Phase 7: VLANs (%d)", len(inventory.VLANs))

	for _, vlan := range inventory.VLANs {
		if vlan == nil || vlan.Name == "" {
			continue
		}

		// Resolve location name for cache key
		locationName := e.Options.DefaultLocation
		if vlan.Location != uuid.Nil {
			if loc, err := e.resolveLocationName(vlan.Location, inventory); err == nil && loc != "" {
				locationName = loc
			}
		}

		// Check if VLAN already exists
		existing, err := e.Cache.LookupVLAN(vlan.VID, locationName)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("vlan %d (%s): lookup error: %v", vlan.VID, vlan.Name, err))
			continue
		}
		if existing != nil {
			created[vlan.ID] = existing.ID
			setExternalID(&vlan.ExternalIDs, "nautobot", existing.ID)

			if e.Options.Merge {
				if updated, mergeErr := e.mergeVLAN(ctx, vlan, existing.ID); mergeErr != nil {
					result.Errors = append(result.Errors,
						fmt.Sprintf("vlan %d (%s): merge error: %v", vlan.VID, vlan.Name, mergeErr))
				} else if updated {
					result.VLANsUpdated++
					continue
				}
			}

			result.VLANsSkipped++
			continue
		}

		// Resolve the Nautobot location for scoping (nil when unmapped).
		var nautobotLocID uuid.UUID
		if vlan.Location != uuid.Nil {
			nautobotLocID = locationMap[vlan.Location]
		}

		// Build the request
		nautobotID, err := e.createVLAN(ctx, vlan, nautobotLocID, result)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("vlan %d (%s): create error: %v", vlan.VID, vlan.Name, err))
			continue
		}

		created[vlan.ID] = nautobotID
		setExternalID(&vlan.ExternalIDs, "nautobot", nautobotID)
		e.Cache.CacheVLAN(vlan.VID, locationName, &CachedItem{
			ID:   nautobotID,
			Name: vlan.Name,
		})
		result.VLANsCreated++
	}

	clog.Info("  VLANs created: %d", result.VLANsCreated)
	return created, nil
}

// createVLAN creates a single VLAN in Nautobot, scoping it to locationID when
// that location is known (locationID may be uuid.Nil for an unscoped VLAN).
func (e *Exporter) createVLAN(
	ctx context.Context,
	vlan *devicetypes.CaniVLAN,
	locationID uuid.UUID,
	result *LoadResult,
) (uuid.UUID, error) {
	// Resolve status
	statusName := vlan.Status
	if statusName == "" {
		statusName = e.Options.DefaultStatus
	}
	if statusName == "" {
		statusName = "Active"
	}
	statusItem, err := e.Cache.GetStatus(statusName)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to resolve status %q: %w", statusName, err)
	}

	req := nautobotapi.VLANRequest{
		Vid:    vlan.VID,
		Name:   vlan.Name,
		Status: makeIDRef(statusItem.ID),
	}

	// Set description
	if vlan.Description != "" {
		req.Description = &vlan.Description
	}

	// Scope to the location when it maps to a known Nautobot location.
	if locationID != uuid.Nil {
		ref := makeLocationRef(locationID)
		req.Location = &ref
	}

	// Resolve role
	if vlan.Role != "" {
		roleItem, rerr := e.Cache.GetRole(vlan.Role)
		if rerr == nil && roleItem != nil {
			req.Role = makeObjectRef(roleItem.ID)
		}
	}

	// Merge CustomFields and ProviderMetadata into custom_fields payload
	cf := map[string]interface{}{}
	for k, v := range vlan.CustomFields {
		cf[k] = v
	}
	if flat := vlan.FlattenProviderMetadata(); len(flat) > 0 {
		for k, v := range flat {
			cf[k] = v
		}
	}
	if len(cf) > 0 {
		req.CustomFields = &cf
		clog.Info("  VLAN %d custom_fields: %d key(s)", vlan.VID, len(cf))
	}

	if e.Options.DryRun {
		clog.DryRun("Would create VLAN %d: %s", vlan.VID, vlan.Name)
		return uuid.New(), nil
	}

	nautobotID, err := e.postVLANWithFallback(ctx, vlan, &req)
	if err != nil {
		return uuid.Nil, err
	}
	clog.Created("  + VLAN %d: %s", vlan.VID, vlan.Name)
	return nautobotID, nil
}

// postVLANWithFallback POSTs the VLAN, retrying once without a location when
// Nautobot rejects the association (e.g. the location type does not permit
// VLANs). This keeps VLANs exporting even when their location type lacks the
// ipam.vlan content type.
func (e *Exporter) postVLANWithFallback(
	ctx context.Context,
	vlan *devicetypes.CaniVLAN,
	req *nautobotapi.VLANRequest,
) (uuid.UUID, error) {
	id, code, body, err := e.postVLAN(ctx, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}
	if code == http.StatusCreated {
		return id, nil
	}
	// An unsupported location type makes Nautobot reject the VLAN — cleanly with
	// 400, or with a 500 when its location validation crashes. Either way, retry
	// once without the location so the VLAN still exports (unscoped).
	if req.Location != nil && (code == http.StatusBadRequest || code == http.StatusInternalServerError) {
		clog.Warn("  VLAN %d (%s): location not accepted, retrying without location", vlan.VID, vlan.Name)
		req.Location = nil
		id, code, body, err = e.postVLAN(ctx, req)
		if err != nil {
			return uuid.Nil, fmt.Errorf("API error: %w", err)
		}
		if code == http.StatusCreated {
			return id, nil
		}
	}
	return uuid.Nil, fmt.Errorf("unexpected status %d: %s", code, body)
}

// postVLAN performs a single VLAN create, returning the new ID (on 201), the
// HTTP status code, and the response body.
func (e *Exporter) postVLAN(ctx context.Context, req *nautobotapi.VLANRequest) (uuid.UUID, int, string, error) {
	resp, err := e.Client.IpamVlansCreateWithResponse(ctx, &nautobotapi.IpamVlansCreateParams{}, *req)
	if err != nil {
		return uuid.Nil, 0, "", err
	}
	if resp.StatusCode() == http.StatusCreated && resp.JSON201 != nil {
		return toUUID(resp.JSON201.Id), resp.StatusCode(), string(resp.Body), nil
	}
	return uuid.Nil, resp.StatusCode(), string(resp.Body), nil
}

// resolveLocationName returns the name of a location from its cani UUID.
func (e *Exporter) resolveLocationName(locID uuid.UUID, inventory *devicetypes.Inventory) (string, error) {
	if loc, ok := inventory.Locations[locID]; ok && loc != nil {
		return loc.Name, nil
	}
	return "", fmt.Errorf("location %s not found in inventory", locID)
}
