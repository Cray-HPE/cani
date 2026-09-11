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

// GetStatus looks up a status by name and returns its ID
// If createStatuses is enabled and the status is not found, it will create it.
func (c *LookupCache) GetStatus(name string) (*CachedItem, error) {
	c.statusesMu.RLock()
	if item, ok := c.statuses[name]; ok {
		c.statusesMu.RUnlock()
		return item, nil
	}
	c.statusesMu.RUnlock()

	// Fetch from API
	c.statusesMu.Lock()
	defer c.statusesMu.Unlock()

	// Double-check after acquiring write lock
	if item, ok := c.statuses[name]; ok {
		return item, nil
	}

	nameFilter := []string{name}
	resp, err := c.client.ExtrasStatusesListWithResponse(c.ctx, &nautobotapi.ExtrasStatusesListParams{
		Name: &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup status %s: %w", name, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup status %s: status %d", name, resp.StatusCode())
	}

	if resp.JSON200 != nil && resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		st := (resp.JSON200.Results)[0]

		// Ensure the status covers all content types the exporter may
		// need (DCIM + IPAM); patch if any are missing.
		if c.createStatuses {
			required := []string{contentTypeDevice, contentTypeModule, contentTypeRack, contentTypePrefix, contentTypeIPAddress, contentTypeVLAN}
			var missing []string
			for _, req := range required {
				found := false
				for _, ct := range st.ContentTypes {
					if ct == req {
						found = true
						break
					}
				}
				if !found {
					missing = append(missing, req)
				}
			}
			if len(missing) > 0 {
				clog.Detail("[nautobot] Status '%s' missing content types %v, updating...", name, missing)
				updated := append(st.ContentTypes, missing...)
				c.statusesMu.Unlock()
				updatedItem, err := c.UpdateStatusContentTypes(toUUID(st.Id), name, updated)
				c.statusesMu.Lock()
				if err != nil {
					clog.Warn("[nautobot] WARNING: Failed to update status '%s' content types: %v", name, err)
				} else {
					c.statuses[name] = updatedItem
					return updatedItem, nil
				}
			}
		}

		item := &CachedItem{
			ID:      toUUID(st.Id),
			Name:    st.Name,
			Display: *st.Display,
		}
		c.statuses[name] = item
		return item, nil
	}

	// Status not found - try to create if enabled
	if c.createStatuses {
		c.statusesMu.Unlock()
		item, err := c.CreateStatus(name)
		c.statusesMu.Lock() // Re-acquire for deferred unlock
		if err != nil {
			return nil, fmt.Errorf("failed to create status %s: %w", name, err)
		}
		return item, nil
	}

	return nil, fmt.Errorf("status not found: %s", name)
}

// GetRole looks up a role by name and returns its ID
func (c *LookupCache) GetRole(name string) (*CachedItem, error) {
	c.rolesMu.RLock()
	if item, ok := c.roles[name]; ok {
		c.rolesMu.RUnlock()
		return item, nil
	}
	c.rolesMu.RUnlock()

	// Fetch from API
	c.rolesMu.Lock()
	defer c.rolesMu.Unlock()

	// Double-check after acquiring write lock
	if item, ok := c.roles[name]; ok {
		return item, nil
	}

	nameFilter := []string{name}
	clog.Detail("[nautobot] Looking up role: '%s'", name)
	resp, err := c.client.ExtrasRolesListWithResponse(c.ctx, &nautobotapi.ExtrasRolesListParams{
		Name: &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup role %s: %w", name, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup role %s: status %d", name, resp.StatusCode())
	}

	if resp.JSON200 != nil && resp.JSON200.Results != nil {
		resultCount := len(resp.JSON200.Results)
		clog.Detail("[nautobot] Role lookup for '%s' returned %d results", name, resultCount)
		if resultCount > 0 {
			r := (resp.JSON200.Results)[0]
			clog.Detail("[nautobot] Found role: name='%s', id=%s, content_types=%v", r.Name, r.Id, r.ContentTypes)

			// Check if role has all required content types
			requiredTypes := []string{contentTypeDevice, contentTypePrefix, contentTypeIPAddress, contentTypeVLAN}
			existingTypes := make(map[string]bool)
			for _, ct := range r.ContentTypes {
				existingTypes[ct] = true
			}
			var missingTypes []string
			for _, rt := range requiredTypes {
				if !existingTypes[rt] {
					missingTypes = append(missingTypes, rt)
				}
			}
			if len(missingTypes) > 0 && c.createRoles {
				clog.Detail("[nautobot] Role '%s' missing content types %v, updating...", name, missingTypes)
				updatedContentTypes := append(r.ContentTypes, missingTypes...)
				c.rolesMu.Unlock()
				updatedItem, err := c.UpdateRoleContentTypes(toUUID(r.Id), name, updatedContentTypes)
				c.rolesMu.Lock()
				if err != nil {
					clog.Warn("[nautobot] WARNING: Failed to update role '%s' content types: %v", name, err)
					// Fall through to return the existing role anyway
				} else {
					c.roles[name] = updatedItem
					return updatedItem, nil
				}
			}

			item := &CachedItem{
				ID:      toUUID(r.Id),
				Name:    r.Name,
				Display: *r.Display,
			}
			c.roles[name] = item
			return item, nil
		}
	} else {
		clog.Detail("[nautobot] Role lookup for '%s': JSON200=%v, Results=%v", name, resp.JSON200 != nil, resp.JSON200 != nil && resp.JSON200.Results != nil)
	}

	// Role not found - try to create if enabled
	if c.createRoles {
		c.rolesMu.Unlock()
		item, err := c.CreateRole(name)
		c.rolesMu.Lock() // Re-acquire for deferred unlock
		if err != nil {
			return nil, fmt.Errorf("failed to create role %s: %w", name, err)
		}
		return item, nil
	}

	return nil, fmt.Errorf("role not found: %s", name)
}
