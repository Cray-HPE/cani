/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024, 2026 Hewlett Packard Enterprise Development LP
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
	"fmt"
	"net/http"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// CreateStatus creates a new status in Nautobot for use with devices, racks, modules, and IPAM objects.
func (c *LookupCache) CreateStatus(name string) (*CachedItem, error) {
	clog.Detail("[nautobot] Creating status: %s", name)

	// Status requires content_types to specify what objects it applies to.
	contentTypes := []string{contentTypeDevice, contentTypeRack, contentTypeModule, contentTypePrefix, contentTypeIPAddress, contentTypeVLAN}

	createResp, err := c.client.ExtrasStatusesCreateWithResponse(c.ctx,
		&nautobotapi.ExtrasStatusesCreateParams{},
		nautobotapi.StatusRequest{
			Name:         name,
			ContentTypes: contentTypes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create status %s: %w", name, err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		clog.Error("[nautobot] Status create failed for %s: %s", name, string(createResp.Body))
		return nil, fmt.Errorf("failed to create status %s: status %d: %s", name, createResp.StatusCode(), string(createResp.Body))
	}

	if createResp.JSON201 != nil {
		item := &CachedItem{
			ID:      toUUID(createResp.JSON201.Id),
			Name:    createResp.JSON201.Name,
			Display: *createResp.JSON201.Display,
		}
		c.statuses[name] = item
		clog.Created("[nautobot] Created status: %s (ID: %s)", name, item.ID)
		return item, nil
	}

	return nil, fmt.Errorf("failed to create status %s: no response body", name)
}

// CreateRole creates a new role in Nautobot
func (c *LookupCache) CreateRole(name string) (*CachedItem, error) {
	clog.Detail("[nautobot] Creating role: %s", name)

	// Role requires content_types and weight
	contentTypes := []string{contentTypeDevice, "dcim.interface", contentTypePrefix, contentTypeIPAddress, contentTypeVLAN}
	weight := 1000

	createResp, err := c.client.ExtrasRolesCreateWithResponse(c.ctx,
		&nautobotapi.ExtrasRolesCreateParams{},
		nautobotapi.RoleRequest{
			Name:         name,
			ContentTypes: contentTypes,
			Weight:       &weight,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create role %s: %w", name, err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		clog.Error("[nautobot] Role create failed for %s: %s", name, string(createResp.Body))
		return nil, fmt.Errorf("failed to create role %s: status %d: %s", name, createResp.StatusCode(), string(createResp.Body))
	}

	if createResp.JSON201 != nil {
		item := &CachedItem{
			ID:      toUUID(createResp.JSON201.Id),
			Name:    createResp.JSON201.Name,
			Display: *createResp.JSON201.Display,
		}
		c.roles[name] = item
		clog.Created("[nautobot] Created role: %s (ID: %s)", name, item.ID)
		return item, nil
	}

	return nil, fmt.Errorf("failed to create role %s: no response body", name)
}

// CreateLocation creates a new location in Nautobot.
// The caller must ensure the location type already exists; this method does
// not fall back to a hard-coded type.
func (c *LookupCache) CreateLocation(name string) (*CachedItem, error) {
	clog.Detail("[nautobot] Creating location: %s", name)

	// Use "Section" as the default location type for auto-created locations.
	// Section allows rack, device, and module content types.
	sectionDef := parentDef("section")
	locType, err := c.GetOrCreateLocationType("section", sectionDef)
	if err != nil {
		return nil, fmt.Errorf("cannot create location %q: failed to resolve location type: %w", name, err)
	}

	// Resolve a status for the location.
	status, err := c.GetStatus("Active")
	if err != nil {
		return nil, fmt.Errorf("cannot create location %q: failed to resolve status: %w", name, err)
	}

	req := nautobotapi.LocationRequest{
		Name: name,
	}
	if err := setRefID(&req.LocationType, locType.ID); err != nil {
		return nil, fmt.Errorf("cannot create location %q: set location type reference: %w", name, err)
	}
	if err := setRefID(&req.Status, status.ID); err != nil {
		return nil, fmt.Errorf("cannot create location %q: set status reference: %w", name, err)
	}

	resp, err := c.client.DcimLocationsCreateWithResponse(c.ctx,
		&nautobotapi.DcimLocationsCreateParams{}, req)
	if err != nil {
		return nil, fmt.Errorf("cannot create location %q: %w", name, err)
	}
	if resp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("cannot create location %q: status %d: %s", name, resp.StatusCode(), string(resp.Body))
	}

	if resp.JSON201 != nil {
		item := &CachedItem{
			ID:      toUUID(resp.JSON201.Id),
			Name:    resp.JSON201.Name,
			Display: *resp.JSON201.Display,
		}
		c.locations[name] = item
		clog.Created("[nautobot] Created location: %s (ID: %s)", name, item.ID)
		return item, nil
	}

	return nil, fmt.Errorf("cannot create location %q: no response body", name)
}
