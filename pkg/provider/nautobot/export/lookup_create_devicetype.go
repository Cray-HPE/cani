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

// GetOrCreateManufacturer looks up a manufacturer by name, creating it if not found
func (c *LookupCache) GetOrCreateManufacturer(name string) (*CachedItem, error) {
	c.manufacturersMu.RLock()
	if item, ok := c.manufacturers[name]; ok {
		c.manufacturersMu.RUnlock()
		return item, nil
	}
	c.manufacturersMu.RUnlock()

	c.manufacturersMu.Lock()
	defer c.manufacturersMu.Unlock()

	// Double-check after acquiring write lock
	if item, ok := c.manufacturers[name]; ok {
		return item, nil
	}

	// Try to find existing manufacturer
	nameFilter := []string{name}
	resp, err := c.client.DcimManufacturersListWithResponse(c.ctx, &nautobotapi.DcimManufacturersListParams{
		Name: &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup manufacturer %s: %w", name, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup manufacturer %s: status %d", name, resp.StatusCode())
	}

	if resp.JSON200 != nil && resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		mfr := (resp.JSON200.Results)[0]
		item := &CachedItem{
			ID:      toUUID(mfr.Id),
			Name:    mfr.Name,
			Display: *mfr.Display,
		}
		c.manufacturers[name] = item
		return item, nil
	}

	// Manufacturer not found, create it
	clog.Detail("[nautobot] Creating manufacturer: %s", name)
	createResp, err := c.client.DcimManufacturersCreateWithResponse(c.ctx,
		&nautobotapi.DcimManufacturersCreateParams{},
		nautobotapi.ManufacturerRequest{
			Name: name,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create manufacturer %s: %w", name, err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("failed to create manufacturer %s: status %d", name, createResp.StatusCode())
	}

	if createResp.JSON201 != nil {
		item := &CachedItem{
			ID:      toUUID(createResp.JSON201.Id),
			Name:    createResp.JSON201.Name,
			Display: *createResp.JSON201.Display,
		}
		c.manufacturers[name] = item
		clog.Created("[nautobot] Created manufacturer: %s (ID: %s)", name, item.ID)
		return item, nil
	}

	return nil, fmt.Errorf("failed to create manufacturer %s: no response body", name)
}

// CreateDeviceTypeFromLocal creates a device type in Nautobot from the local devicetypes library
func (c *LookupCache) CreateDeviceTypeFromLocal(slug string) (*CachedItem, error) {
	// Look up device type in local library
	localDT, found := devicetypes.GetBySlug(slug)
	if !found {
		return nil, fmt.Errorf("device type not found in local library: %s", slug)
	}

	// Get or create the manufacturer
	manufacturer, err := c.GetOrCreateManufacturer(localDT.Manufacturer)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create manufacturer for device type %s: %w", slug, err)
	}

	// Build the device type request
	req := nautobotapi.WritableDeviceTypeRequest{
		Model: localDT.Model,
	}
	if err := setRefID(&req.Manufacturer, manufacturer.ID); err != nil {
		return nil, fmt.Errorf("set manufacturer reference for device type %s: %w", slug, err)
	}

	// Set optional fields if available
	if localDT.PartNumber != "" {
		req.PartNumber = &localDT.PartNumber
	}
	if localDT.IsFullDepth {
		isFullDepth := localDT.IsFullDepth
		req.IsFullDepth = &isFullDepth
	}

	// Map SubdeviceRole for chassis/blade relationships
	if localDT.SubdeviceRole != "" {
		sr := &nautobotapi.WritableDeviceTypeRequest_SubdeviceRole{}
		switch localDT.SubdeviceRole {
		case "parent":
			if err := sr.FromSubdeviceRoleEnum(nautobotapi.SubdeviceRoleEnumParent); err == nil {
				req.SubdeviceRole = sr
			}
		case "child":
			if err := sr.FromSubdeviceRoleEnum(nautobotapi.SubdeviceRoleEnumChild); err == nil {
				req.SubdeviceRole = sr
			}
			// Nautobot requires child device types to have u_height=0
			zeroHeight := 0
			req.UHeight = &zeroHeight
		default:
			clog.Warn("[nautobot] Unknown SubdeviceRole %q for %s, skipping", localDT.SubdeviceRole, slug)
		}
	}

	// Set u_height for non-child device types (child types are forced to 0 above)
	if req.UHeight == nil && localDT.UHeight > 0 {
		uHeight := localDT.UHeight
		req.UHeight = &uHeight
	}

	clog.Detail("[nautobot] Creating device type: %s (manufacturer: %s)", localDT.Model, localDT.Manufacturer)

	createResp, err := c.client.DcimDeviceTypesCreateWithResponse(c.ctx,
		&nautobotapi.DcimDeviceTypesCreateParams{},
		req,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create device type %s: %w", slug, err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		// Log the response body for debugging
		clog.Error("[nautobot] Device type create failed for %s: %s", slug, string(createResp.Body))
		return nil, fmt.Errorf("failed to create device type %s: status %d: %s", slug, createResp.StatusCode(), string(createResp.Body))
	}

	if createResp.JSON201 != nil {
		item := &CachedItem{
			ID:      toUUID(createResp.JSON201.Id),
			Name:    createResp.JSON201.Model,
			Slug:    createResp.JSON201.Model,
			Display: *createResp.JSON201.Display,
		}
		c.deviceTypes[slug] = item
		clog.Created("[nautobot] Created device type: %s (ID: %s)", slug, item.ID)
		return item, nil
	}

	return nil, fmt.Errorf("failed to create device type %s: no response body", slug)
}

// CreateDeviceTypeFromCaniDevice creates a device type in Nautobot from
// inventory data when the local YAML library does not contain the slug.
func (c *LookupCache) CreateDeviceTypeFromCaniDevice(device *devicetypes.CaniDeviceType) (*CachedItem, error) {
	if device == nil || device.Slug == "" {
		return nil, fmt.Errorf("device or slug is empty")
	}

	mfr := device.Manufacturer
	if mfr == "" {
		mfr = "Unknown"
	}
	manufacturer, err := c.GetOrCreateManufacturer(mfr)
	if err != nil {
		return nil, fmt.Errorf("manufacturer for %s: %w", device.Slug, err)
	}

	model := device.Model
	if model == "" {
		model = device.Slug
	}

	req := nautobotapi.WritableDeviceTypeRequest{
		Model: model,
	}
	if err := setRefID(&req.Manufacturer, manufacturer.ID); err != nil {
		return nil, fmt.Errorf("set manufacturer reference for device type %s: %w", device.Slug, err)
	}

	if device.PartNumber != "" {
		req.PartNumber = &device.PartNumber
	}
	if device.IsFullDepth {
		v := device.IsFullDepth
		req.IsFullDepth = &v
	}
	if device.UHeight > 0 {
		h := device.UHeight
		req.UHeight = &h
	}

	clog.Detail("[nautobot] Creating device type from inventory: %s (manufacturer: %s)", model, mfr)

	createResp, err := c.client.DcimDeviceTypesCreateWithResponse(c.ctx,
		&nautobotapi.DcimDeviceTypesCreateParams{},
		req,
	)
	if err != nil {
		return nil, fmt.Errorf("API error creating device type %s: %w", device.Slug, err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("device type %s: status %d: %s", device.Slug, createResp.StatusCode(), string(createResp.Body))
	}

	if createResp.JSON201 != nil {
		item := &CachedItem{
			ID:      toUUID(createResp.JSON201.Id),
			Name:    createResp.JSON201.Model,
			Slug:    createResp.JSON201.Model,
			Display: *createResp.JSON201.Display,
		}
		c.deviceTypesMu.Lock()
		c.deviceTypes[device.Slug] = item
		c.deviceTypesMu.Unlock()
		clog.Created("[nautobot] Created device type from inventory: %s (ID: %s)", device.Slug, item.ID)
		return item, nil
	}

	return nil, fmt.Errorf("device type %s: no response body", device.Slug)
}
