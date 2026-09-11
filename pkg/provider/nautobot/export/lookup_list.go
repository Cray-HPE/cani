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

// ListLocations returns all available locations
func (c *LookupCache) ListLocations() ([]*CachedItem, error) {
	c.locationsMu.Lock()
	defer c.locationsMu.Unlock()

	if c.locationsLoaded {
		items := make([]*CachedItem, 0, len(c.locations))
		for _, item := range c.locations {
			items = append(items, item)
		}
		return items, nil
	}

	resp, err := c.client.DcimLocationsListWithResponse(c.ctx, &nautobotapi.DcimLocationsListParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to list locations: status %d", resp.StatusCode())
	}

	var items []*CachedItem
	if resp.JSON200 != nil && resp.JSON200.Results != nil {
		for _, loc := range resp.JSON200.Results {
			item := &CachedItem{
				ID:      toUUID(loc.Id),
				Name:    loc.Name,
				Display: *loc.Display,
			}
			c.locations[loc.Name] = item
			items = append(items, item)
		}
	}
	c.locationsLoaded = true
	return items, nil
}

// ListStatuses returns all available statuses
func (c *LookupCache) ListStatuses() ([]*CachedItem, error) {
	c.statusesMu.Lock()
	defer c.statusesMu.Unlock()

	if c.statusesLoaded {
		items := make([]*CachedItem, 0, len(c.statuses))
		for _, item := range c.statuses {
			items = append(items, item)
		}
		return items, nil
	}

	resp, err := c.client.ExtrasStatusesListWithResponse(c.ctx, &nautobotapi.ExtrasStatusesListParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list statuses: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to list statuses: status %d", resp.StatusCode())
	}

	var items []*CachedItem
	if resp.JSON200 != nil && resp.JSON200.Results != nil {
		for _, st := range resp.JSON200.Results {
			item := &CachedItem{
				ID:      toUUID(st.Id),
				Name:    st.Name,
				Display: *st.Display,
			}
			c.statuses[st.Name] = item
			items = append(items, item)
		}
	}
	c.statusesLoaded = true
	return items, nil
}

// ListRoles returns all available roles
func (c *LookupCache) ListRoles() ([]*CachedItem, error) {
	c.rolesMu.Lock()
	defer c.rolesMu.Unlock()

	if c.rolesLoaded {
		items := make([]*CachedItem, 0, len(c.roles))
		for _, item := range c.roles {
			items = append(items, item)
		}
		return items, nil
	}

	resp, err := c.client.ExtrasRolesListWithResponse(c.ctx, &nautobotapi.ExtrasRolesListParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to list roles: status %d", resp.StatusCode())
	}

	var items []*CachedItem
	if resp.JSON200 != nil && resp.JSON200.Results != nil {
		for _, r := range resp.JSON200.Results {
			item := &CachedItem{
				ID:      toUUID(r.Id),
				Name:    r.Name,
				Display: *r.Display,
			}
			c.roles[r.Name] = item
			items = append(items, item)
		}
	}
	c.rolesLoaded = true
	return items, nil
}

// ListDeviceTypes returns all available device types
func (c *LookupCache) ListDeviceTypes() ([]*CachedItem, error) {
	c.deviceTypesMu.Lock()
	defer c.deviceTypesMu.Unlock()

	if c.deviceTypesLoaded {
		items := make([]*CachedItem, 0, len(c.deviceTypes))
		for _, item := range c.deviceTypes {
			items = append(items, item)
		}
		return items, nil
	}

	resp, err := c.client.DcimDeviceTypesListWithResponse(c.ctx, &nautobotapi.DcimDeviceTypesListParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list device types: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to list device types: status %d", resp.StatusCode())
	}

	var items []*CachedItem
	if resp.JSON200 != nil && resp.JSON200.Results != nil {
		for _, dt := range resp.JSON200.Results {
			item := &CachedItem{
				ID:      toUUID(dt.Id),
				Name:    dt.Model,
				Slug:    dt.Model,
				Display: *dt.Display,
			}
			c.deviceTypes[dt.Model] = item
			items = append(items, item)
		}
	}
	c.deviceTypesLoaded = true
	return items, nil
}
