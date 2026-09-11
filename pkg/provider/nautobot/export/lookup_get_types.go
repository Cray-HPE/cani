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

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// GetDeviceType looks up a device type by model/slug and returns its ID
// If createDeviceTypes is enabled and the device type is not found, it will attempt
// to create it from the local devicetypes library.
func (c *LookupCache) GetDeviceType(slug string) (*CachedItem, error) {
	c.deviceTypesMu.RLock()
	if item, ok := c.deviceTypes[slug]; ok {
		c.deviceTypesMu.RUnlock()
		return item, nil
	}
	c.deviceTypesMu.RUnlock()

	// Fetch from API
	c.deviceTypesMu.Lock()
	defer c.deviceTypesMu.Unlock()

	// Double-check after acquiring write lock
	if item, ok := c.deviceTypes[slug]; ok {
		return item, nil
	}

	// Try searching by slug first
	model := []string{slug}
	resp, err := c.client.DcimDeviceTypesListWithResponse(c.ctx, &nautobotapi.DcimDeviceTypesListParams{
		Model: &model,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup device type %s: %w", slug, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup device type %s: status %d", slug, resp.StatusCode())
	}

	if resp.JSON200 != nil && resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		dt := (resp.JSON200.Results)[0]
		item := &CachedItem{
			ID:      toUUID(dt.Id),
			Name:    dt.Model,
			Slug:    dt.Model, // DeviceType uses model as identifier
			Display: *dt.Display,
		}
		c.deviceTypes[slug] = item
		return item, nil
	}

	// If not found by slug, try looking up by the actual model name from local library
	if localDT, found := devicetypes.GetBySlug(slug); found {
		modelName := []string{localDT.Model}
		resp2, err := c.client.DcimDeviceTypesListWithResponse(c.ctx, &nautobotapi.DcimDeviceTypesListParams{
			Model: &modelName,
		})
		if err == nil && resp2.StatusCode() == http.StatusOK {
			if resp2.JSON200 != nil && resp2.JSON200.Results != nil && len(resp2.JSON200.Results) > 0 {
				dt := (resp2.JSON200.Results)[0]
				item := &CachedItem{
					ID:      toUUID(dt.Id),
					Name:    dt.Model,
					Slug:    dt.Model,
					Display: *dt.Display,
				}
				c.deviceTypes[slug] = item
				clog.Detail("[nautobot] Found existing device type by model name: %s", dt.Model)
				return item, nil
			}
		}
	}

	// Device type not found in Nautobot - try to create from local library if enabled
	if c.createDeviceTypes {
		// Release lock before calling CreateDeviceTypeFromLocal (it acquires its own locks)
		c.deviceTypesMu.Unlock()
		item, err := c.CreateDeviceTypeFromLocal(slug)
		c.deviceTypesMu.Lock() // Re-acquire for deferred unlock
		if err != nil {
			return nil, fmt.Errorf("failed to create device type %s: %w", slug, err)
		}
		return item, nil
	}

	return nil, fmt.Errorf("device type not found: %s", slug)
}

// GetLocation looks up a location by name and returns its ID
// If createLocations is enabled and the location is not found, it will create it.
func (c *LookupCache) GetLocation(name string) (*CachedItem, error) {
	c.locationsMu.RLock()
	if item, ok := c.locations[name]; ok {
		c.locationsMu.RUnlock()
		return item, nil
	}
	c.locationsMu.RUnlock()

	// Fetch from API
	c.locationsMu.Lock()
	defer c.locationsMu.Unlock()

	// Double-check after acquiring write lock
	if item, ok := c.locations[name]; ok {
		return item, nil
	}

	nameFilter := []string{name}
	resp, err := c.client.DcimLocationsListWithResponse(c.ctx, &nautobotapi.DcimLocationsListParams{
		Name: &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup location %s: %w", name, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup location %s: status %d", name, resp.StatusCode())
	}

	if resp.JSON200 != nil && resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		loc := (resp.JSON200.Results)[0]
		item := &CachedItem{
			ID:      toUUID(loc.Id),
			Name:    loc.Name,
			Display: *loc.Display,
		}
		c.locations[name] = item
		return item, nil
	}

	// Location not found - try to create if enabled
	if c.createLocations {
		c.locationsMu.Unlock()
		item, err := c.CreateLocation(name)
		c.locationsMu.Lock() // Re-acquire for deferred unlock
		if err != nil {
			return nil, fmt.Errorf("failed to create location %s: %w", name, err)
		}
		return item, nil
	}

	return nil, fmt.Errorf("location not found: %s", name)
}

// LookupLocation checks whether a location already exists in Nautobot by name.
// Unlike GetLocation, it never auto-creates. Returns nil, nil when not found.
func (c *LookupCache) LookupLocation(name string) (*CachedItem, error) {
	c.locationsMu.RLock()
	if item, ok := c.locations[name]; ok {
		c.locationsMu.RUnlock()
		return item, nil
	}
	c.locationsMu.RUnlock()

	nameFilter := []string{name}
	resp, err := c.client.DcimLocationsListWithResponse(c.ctx, &nautobotapi.DcimLocationsListParams{
		Name: &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup location %s: %w", name, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup location %s: status %d", name, resp.StatusCode())
	}

	if resp.JSON200 != nil && resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		loc := (resp.JSON200.Results)[0]
		item := &CachedItem{
			ID:      toUUID(loc.Id),
			Name:    loc.Name,
			Display: *loc.Display,
		}
		c.locationsMu.Lock()
		c.locations[name] = item
		c.locationsMu.Unlock()
		return item, nil
	}

	return nil, nil
}
