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
	"github.com/google/uuid"
)

// UpdateStatusContentTypes updates an existing status to include additional content types.
func (c *LookupCache) UpdateStatusContentTypes(id uuid.UUID, name string, contentTypes []string) (*CachedItem, error) {
	clog.Detail("[nautobot] Updating status '%s' content types to: %v", name, contentTypes)

	patchResp, err := c.client.ExtrasStatusesPartialUpdateWithResponse(c.ctx,
		id,
		&nautobotapi.ExtrasStatusesPartialUpdateParams{},
		nautobotapi.PatchedStatusRequest{
			ContentTypes: &contentTypes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update status %s: %w", name, err)
	}
	if patchResp.StatusCode() != http.StatusOK {
		clog.Error("[nautobot] Status update failed for %s: %s", name, string(patchResp.Body))
		return nil, fmt.Errorf("failed to update status %s: status %d: %s", name, patchResp.StatusCode(), string(patchResp.Body))
	}

	if patchResp.JSON200 != nil {
		item := &CachedItem{
			ID:      toUUID(patchResp.JSON200.Id),
			Name:    patchResp.JSON200.Name,
			Display: *patchResp.JSON200.Display,
		}
		clog.Created("[nautobot] Updated status: %s (ID: %s) with content_types: %v", name, item.ID, contentTypes)
		return item, nil
	}

	return nil, fmt.Errorf("failed to update status %s: no response body", name)
}

// UpdateLocationTypeContentTypes patches an existing location type to include
// additional content types (e.g. ipam.vlan/ipam.prefix so VLANs and prefixes
// can be scoped to locations of this type).
func (c *LookupCache) UpdateLocationTypeContentTypes(id uuid.UUID, name string, contentTypes []string) (*CachedItem, error) {
	clog.Detail("[nautobot] Updating location type '%s' content types to: %v", name, contentTypes)

	patchResp, err := c.client.DcimLocationTypesPartialUpdateWithResponse(c.ctx,
		id,
		&nautobotapi.DcimLocationTypesPartialUpdateParams{},
		nautobotapi.PatchedLocationTypeRequest{
			ContentTypes: &contentTypes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update location type %s: %w", name, err)
	}
	if patchResp.StatusCode() != http.StatusOK {
		clog.Error("[nautobot] Location type update failed for %s: %s", name, string(patchResp.Body))
		return nil, fmt.Errorf("failed to update location type %s: status %d: %s", name, patchResp.StatusCode(), string(patchResp.Body))
	}

	if patchResp.JSON200 != nil {
		item := &CachedItem{
			ID:   toUUID(patchResp.JSON200.Id),
			Name: patchResp.JSON200.Name,
		}
		if patchResp.JSON200.Display != nil {
			item.Display = *patchResp.JSON200.Display
		}
		clog.Created("[nautobot] Updated location type: %s (ID: %s) with content_types: %v", name, item.ID, contentTypes)
		return item, nil
	}

	return nil, fmt.Errorf("failed to update location type %s: no response body", name)
}

// UpdateRoleContentTypes updates an existing role to add content types
func (c *LookupCache) UpdateRoleContentTypes(id uuid.UUID, name string, contentTypes []string) (*CachedItem, error) {
	clog.Detail("[nautobot] Updating role '%s' content types to: %v", name, contentTypes)

	weight := 1000
	patchResp, err := c.client.ExtrasRolesPartialUpdateWithResponse(c.ctx,
		id,
		&nautobotapi.ExtrasRolesPartialUpdateParams{},
		nautobotapi.PatchedRoleRequest{
			ContentTypes: &contentTypes,
			Weight:       &weight,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update role %s: %w", name, err)
	}
	if patchResp.StatusCode() != http.StatusOK {
		clog.Error("[nautobot] Role update failed for %s: %s", name, string(patchResp.Body))
		return nil, fmt.Errorf("failed to update role %s: status %d: %s", name, patchResp.StatusCode(), string(patchResp.Body))
	}

	if patchResp.JSON200 != nil {
		item := &CachedItem{
			ID:      toUUID(patchResp.JSON200.Id),
			Name:    patchResp.JSON200.Name,
			Display: *patchResp.JSON200.Display,
		}
		clog.Created("[nautobot] Updated role: %s (ID: %s) with content_types: %v", name, item.ID, contentTypes)
		return item, nil
	}

	return nil, fmt.Errorf("failed to update role %s: no response body", name)
}
