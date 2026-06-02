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
			result.VLANsSkipped++
			continue
		}

		// Build the request
		nautobotID, err := e.createVLAN(ctx, vlan, result)
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

// createVLAN creates a single VLAN in Nautobot.
func (e *Exporter) createVLAN(
	ctx context.Context,
	vlan *devicetypes.CaniVLAN,
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

	// Note: Location is intentionally omitted for VLANs because Nautobot
	// restricts which location types may associate with VLANs. The location
	// type "Section" (used by cani) does not support VLAN associations.

	// Resolve role
	if vlan.Role != "" {
		roleItem, err := e.Cache.GetRole(vlan.Role)
		if err == nil && roleItem != nil {
			ref := makeTenantRef(roleItem.ID)
			req.Role = ref
		}
	}

	if e.Options.DryRun {
		clog.DryRun("Would create VLAN %d: %s", vlan.VID, vlan.Name)
		return uuid.New(), nil
	}

	resp, err := e.Client.IpamVlansCreateWithResponse(ctx, &nautobotapi.IpamVlansCreateParams{}, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}
	if resp.StatusCode() != http.StatusCreated {
		return uuid.Nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	if resp.JSON201 == nil {
		return uuid.Nil, fmt.Errorf("empty response body")
	}

	nautobotID := toUUID(resp.JSON201.Id)
	clog.Created("  + VLAN %d: %s", vlan.VID, vlan.Name)
	return nautobotID, nil
}

// resolveLocationName returns the name of a location from its cani UUID.
func (e *Exporter) resolveLocationName(locID uuid.UUID, inventory *devicetypes.Inventory) (string, error) {
	if loc, ok := inventory.Locations[locID]; ok && loc != nil {
		return loc.Name, nil
	}
	return "", fmt.Errorf("location %s not found in inventory", locID)
}
