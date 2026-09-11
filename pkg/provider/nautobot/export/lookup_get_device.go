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

// GetDeviceByName looks up a device by name
func (c *LookupCache) GetDeviceByName(name string) (*CachedItem, error) {
	c.devicesMu.RLock()
	if item, ok := c.devices[name]; ok {
		c.devicesMu.RUnlock()
		return item, nil
	}
	c.devicesMu.RUnlock()

	// Fetch from API
	c.devicesMu.Lock()
	defer c.devicesMu.Unlock()

	// Double-check after acquiring write lock
	if item, ok := c.devices[name]; ok {
		return item, nil
	}

	nameFilter := []string{name}
	resp, err := c.client.DcimDevicesListWithResponse(c.ctx, &nautobotapi.DcimDevicesListParams{
		Name: &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup device %s: %w", name, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup device %s: status %d", name, resp.StatusCode())
	}

	if resp.JSON200 != nil && resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		d := (resp.JSON200.Results)[0]
		item := &CachedItem{
			ID:      toUUID(d.Id),
			Name:    *d.Name,
			Display: *d.Display,
		}
		c.devices[name] = item
		return item, nil
	}

	return nil, nil // Not found is not an error for device lookup
}

// GetAllDevicesByName returns all Nautobot devices matching the given name.
// Unlike GetDeviceByName (which caches and returns only the first), this
// queries the API each time and returns every result so callers can
// disambiguate same-name devices.
func (c *LookupCache) GetAllDevicesByName(name string) ([]*CachedItem, error) {
	nameFilter := []string{name}
	resp, err := c.client.DcimDevicesListWithResponse(c.ctx, &nautobotapi.DcimDevicesListParams{
		Name: &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup devices %s: %w", name, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup devices %s: status %d", name, resp.StatusCode())
	}

	if resp.JSON200 == nil || resp.JSON200.Results == nil {
		return nil, nil
	}

	items := make([]*CachedItem, 0, len(resp.JSON200.Results))
	for _, d := range resp.JSON200.Results {
		items = append(items, &CachedItem{
			ID:      toUUID(d.Id),
			Name:    *d.Name,
			Display: *d.Display,
		})
	}
	return items, nil
}

// GetRackByName looks up a rack by name
func (c *LookupCache) GetRackByName(name string) (*CachedItem, error) {
	// Racks don't have a dedicated cache, so query directly
	nameFilter := []string{name}
	resp, err := c.client.DcimRacksListWithResponse(c.ctx, &nautobotapi.DcimRacksListParams{
		Name: &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup rack %s: %w", name, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup rack %s: status %d", name, resp.StatusCode())
	}

	if resp.JSON200 != nil && resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		r := (resp.JSON200.Results)[0]
		item := &CachedItem{
			ID:      toUUID(r.Id),
			Name:    r.Name,
			Display: *r.Display,
		}
		return item, nil
	}

	return nil, nil // Not found is not an error for rack lookup
}
