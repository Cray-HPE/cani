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

// GetOrCreateLocationType gets or creates a location type by name.
// When def is non-nil, its Nestable, ContentTypes, and Parent fields are used
// when creating the LocationType in Nautobot. When nil, defaults are applied.
func (c *LookupCache) GetOrCreateLocationType(name string, def *devicetypes.LocationTypeDefinition) (*CachedItem, error) {
	// Prefer the display name from the YAML definition over the raw slug.
	if def != nil && def.Name != "" {
		name = def.Name
	}

	// Try to find existing location type
	nameFilter := []string{name}
	resp, err := c.client.DcimLocationTypesListWithResponse(c.ctx, &nautobotapi.DcimLocationTypesListParams{
		Name: &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup location type %s: %w", name, err)
	}
	if resp.StatusCode() == http.StatusOK && resp.JSON200 != nil && resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		lt := (resp.JSON200.Results)[0]
		c.reconcileLocationTypeIPAM(lt)
		return &CachedItem{
			ID:      toUUID(lt.Id),
			Name:    lt.Name,
			Display: *lt.Display,
		}, nil
	}

	if !c.createLocationTypes {
		return nil, fmt.Errorf("location type %q not found in Nautobot (enable create_location_types)", name)
	}

	return c.createLocationType(name, def)
}

// locationTypeIPAMContentTypes lists the IPAM object types that every
// cani-created location type must advertise so VLANs and prefixes can be
// scoped to any location Nautobot builds. Without these, Nautobot rejects
// location-scoped VLANs/prefixes. IP addresses are not location-scoped in
// Nautobot, so ipam.ipaddress is intentionally excluded (it is rejected).
func locationTypeIPAMContentTypes() []string {
	return []string{contentTypeVLAN, contentTypePrefix}
}

// locationTypeContentTypes returns the Nautobot content types a cani-created
// location type advertises: its declared DCIM types (device/rack by default,
// or the YAML definition's) plus the IPAM types.
func locationTypeContentTypes(def *devicetypes.LocationTypeDefinition) []string {
	base := []string{contentTypeDevice, contentTypeRack}
	if def != nil && len(def.ContentTypes) > 0 {
		base = toNautobotContentTypes(def.ContentTypes)
	}
	return append(base, missingContentTypes(base, locationTypeIPAMContentTypes())...)
}

// reconcileLocationTypeIPAM PATCHes an existing location type to add any
// missing IPAM content types so re-exports keep scoping VLANs/prefixes to it.
func (c *LookupCache) reconcileLocationTypeIPAM(lt nautobotapi.LocationType) {
	existing := []string{}
	if lt.ContentTypes != nil {
		existing = *lt.ContentTypes
	}
	missing := missingContentTypes(existing, locationTypeIPAMContentTypes())
	if len(missing) == 0 {
		return
	}
	updated := append(existing, missing...)
	if _, err := c.UpdateLocationTypeContentTypes(toUUID(lt.Id), lt.Name, updated); err != nil {
		clog.Warn("[nautobot] could not update location type %s content types: %v", lt.Name, err)
	}
}

// createLocationType creates a new LocationType in Nautobot.
func (c *LookupCache) createLocationType(name string, def *devicetypes.LocationTypeDefinition) (*CachedItem, error) {
	clog.Detail("[nautobot] Creating location type: %s", name)

	req := nautobotapi.LocationTypeRequest{
		Name: name,
	}

	if def != nil {
		req.Nestable = &def.Nestable
		if def.Description != "" {
			req.Description = &def.Description
		}
		if def.Parent != "" {
			parentItem, perr := c.GetOrCreateLocationType(def.Parent, parentDef(def.Parent))
			if perr == nil && parentItem != nil {
				if err := setRefID(&req.Parent, parentItem.ID); err != nil {
					return nil, fmt.Errorf("set parent reference for location type %s: %w", name, err)
				}
			}
		}
	}
	ct := locationTypeContentTypes(def)
	req.ContentTypes = &ct

	createResp, err := c.client.DcimLocationTypesCreateWithResponse(c.ctx,
		&nautobotapi.DcimLocationTypesCreateParams{},
		req,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create location type %s: %w", name, err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		clog.Error("[nautobot] Location type create failed for %s: %s", name, string(createResp.Body))
		return nil, fmt.Errorf("failed to create location type %s: status %d: %s", name, createResp.StatusCode(), string(createResp.Body))
	}

	if createResp.JSON201 != nil {
		item := &CachedItem{
			ID:      toUUID(createResp.JSON201.Id),
			Name:    createResp.JSON201.Name,
			Display: *createResp.JSON201.Display,
		}
		clog.Created("[nautobot] Created location type: %s (ID: %s)", name, item.ID)
		return item, nil
	}

	return nil, fmt.Errorf("failed to create location type %s: no response body", name)
}

// parentDef looks up a LocationTypeDefinition by slug for use as a parent.
func parentDef(slug string) *devicetypes.LocationTypeDefinition {
	lt, ok := devicetypes.GetLocationTypeBySlug(slug)
	if !ok {
		return nil
	}
	return &lt
}

// toNautobotContentTypes converts short content type names ("device", "rack",
// "module") to their fully-qualified Nautobot equivalents ("dcim.device", etc.).
func toNautobotContentTypes(types []string) []string {
	out := make([]string, 0, len(types))
	for _, t := range types {
		switch t {
		case "device":
			out = append(out, "dcim.device")
		case "rack":
			out = append(out, "dcim.rack")
		case "module":
			out = append(out, "dcim.module")
		case "vlan":
			out = append(out, "ipam.vlan")
		case "prefix":
			out = append(out, "ipam.prefix")
		case "ipaddress":
			out = append(out, "ipam.ipaddress")
		default:
			out = append(out, t)
		}
	}
	return out
}

// CacheLocation adds a location to the local cache.
func (c *LookupCache) CacheLocation(name string, item *CachedItem) {
	c.locationsMu.Lock()
	defer c.locationsMu.Unlock()
	c.locations[name] = item
}
