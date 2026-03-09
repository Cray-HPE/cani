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
	"strings"
	"sync"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// CachedItem represents a cached Nautobot object with its ID and display name
type CachedItem struct {
	ID      uuid.UUID
	Name    string
	Slug    string
	Display string
	CableID uuid.UUID // For interfaces: the ID of the attached cable (if any)
}

// toUUID converts an openapi_types.UUID pointer to uuid.UUID
func toUUID(id *openapi_types.UUID) uuid.UUID {
	if id == nil {
		return uuid.Nil
	}
	return uuid.UUID(*id)
}

// LookupCache provides cached lookups for Nautobot reference objects
// All lookups are cached in memory to avoid repeated API calls during a session
type LookupCache struct {
	client *NautobotClient
	ctx    context.Context

	// Caches for different object types
	deviceTypes   map[string]*CachedItem // keyed by slug
	locations     map[string]*CachedItem // keyed by name
	statuses      map[string]*CachedItem // keyed by name
	roles         map[string]*CachedItem // keyed by name
	devices       map[string]*CachedItem // keyed by name
	manufacturers map[string]*CachedItem // keyed by name
	interfaces    map[string]*CachedItem // keyed by "deviceID:ifaceName"
	tags          map[string]*CachedItem // keyed by name

	// Mutexes for thread-safe access
	deviceTypesMu   sync.RWMutex
	locationsMu     sync.RWMutex
	statusesMu      sync.RWMutex
	rolesMu         sync.RWMutex
	devicesMu       sync.RWMutex
	manufacturersMu sync.RWMutex
	interfacesMu    sync.RWMutex
	tagsMu          sync.RWMutex

	// Track if full list has been fetched
	deviceTypesLoaded bool
	locationsLoaded   bool
	statusesLoaded    bool
	rolesLoaded       bool
	devicesLoaded     bool

	// Options for auto-creation
	createDeviceTypes bool
	createStatuses    bool
	createRoles       bool
	createLocations   bool
}

// NewLookupCache creates a new lookup cache for the given client
func NewLookupCache(client *NautobotClient) *LookupCache {
	return &LookupCache{
		client:        client,
		ctx:           context.Background(),
		deviceTypes:   make(map[string]*CachedItem),
		locations:     make(map[string]*CachedItem),
		statuses:      make(map[string]*CachedItem),
		roles:         make(map[string]*CachedItem),
		devices:       make(map[string]*CachedItem),
		manufacturers: make(map[string]*CachedItem),
		interfaces:    make(map[string]*CachedItem),
		tags:          make(map[string]*CachedItem),
	}
}

// SetContext sets the context for API calls
func (c *LookupCache) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// SetCreateDeviceTypes enables or disables auto-creation of device types
func (c *LookupCache) SetCreateDeviceTypes(create bool) {
	c.createDeviceTypes = create
}

// SetCreateStatuses enables or disables auto-creation of statuses
func (c *LookupCache) SetCreateStatuses(create bool) {
	c.createStatuses = create
}

// SetCreateRoles enables or disables auto-creation of roles
func (c *LookupCache) SetCreateRoles(create bool) {
	c.createRoles = create
}

// SetCreateLocations enables or disables auto-creation of locations
func (c *LookupCache) SetCreateLocations(create bool) {
	c.createLocations = create
}

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

		// Ensure the status covers dcim.module; patch if missing.
		if c.createStatuses {
			hasModule := false
			for _, ct := range st.ContentTypes {
				if ct == "dcim.module" {
					hasModule = true
					break
				}
			}
			if !hasModule {
				clog.Detail("[nautobot] Status '%s' exists but lacks dcim.module content type, updating...", name)
				updated := append(st.ContentTypes, "dcim.module")
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

			// Check if role has dcim.device content type
			hasDeviceType := false
			for _, ct := range r.ContentTypes {
				if ct == "dcim.device" {
					hasDeviceType = true
					break
				}
			}
			if !hasDeviceType && c.createRoles {
				clog.Detail("[nautobot] Role '%s' exists but doesn't have dcim.device content type, updating...", name)
				// Update the role to add dcim.device content type
				updatedContentTypes := append(r.ContentTypes, "dcim.device")
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

	// Build manufacturer reference using the proper type
	mfrID := nautobotapi.BulkWritableCableRequestStatusId{}
	mfrID.FromBulkWritableCableRequestStatusId0(manufacturer.ID)

	// Build the device type request
	req := nautobotapi.WritableDeviceTypeRequest{
		Model: localDT.Model,
		Manufacturer: nautobotapi.BulkWritableCableRequestStatus{
			Id: &mfrID,
		},
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
		sr := &nautobotapi.ParentChildStatus{}
		switch localDT.SubdeviceRole {
		case "parent":
			if err := sr.FromSubdeviceRoleEnum(nautobotapi.Parent); err == nil {
				req.SubdeviceRole = sr
			}
		case "child":
			if err := sr.FromSubdeviceRoleEnum(nautobotapi.Child); err == nil {
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

	mfrID := nautobotapi.BulkWritableCableRequestStatusId{}
	mfrID.FromBulkWritableCableRequestStatusId0(manufacturer.ID)

	model := device.Model
	if model == "" {
		model = device.Slug
	}

	req := nautobotapi.WritableDeviceTypeRequest{
		Model: model,
		Manufacturer: nautobotapi.BulkWritableCableRequestStatus{
			Id: &mfrID,
		},
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

// CreateStatus creates a new status in Nautobot for use with devices, racks, and modules.
func (c *LookupCache) CreateStatus(name string) (*CachedItem, error) {
	clog.Detail("[nautobot] Creating status: %s", name)

	// Status requires content_types to specify what objects it applies to.
	contentTypes := []string{"dcim.device", "dcim.rack", "dcim.module"}

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
	contentTypes := []string{"dcim.device"}
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

// CreateLocation creates a new location in Nautobot
// It first looks up or creates a default location type called "Site"
func (c *LookupCache) CreateLocation(name string) (*CachedItem, error) {
	clog.Detail("[nautobot] Creating location: %s", name)

	// First, we need a location type - look for "Site" or create one
	locationType, err := c.GetOrCreateLocationType("Site")
	if err != nil {
		return nil, fmt.Errorf("failed to get location type for %s: %w", name, err)
	}

	// Get a valid status for the location
	status, err := c.GetStatus("Active")
	if err != nil {
		// Try "active" lowercase
		status, err = c.GetStatus("active")
		if err != nil {
			return nil, fmt.Errorf("failed to get status for location %s: no 'Active' status found", name)
		}
	}

	// Create the location
	statusRef := nautobotapi.BulkWritableCableRequestStatus{}
	statusID := nautobotapi.BulkWritableCableRequestStatusId{}
	statusID.FromBulkWritableCableRequestStatusId0(status.ID)
	statusRef.Id = &statusID

	locTypeRef := nautobotapi.BulkWritableCableRequestStatus{}
	locTypeID := nautobotapi.BulkWritableCableRequestStatusId{}
	locTypeID.FromBulkWritableCableRequestStatusId0(locationType.ID)
	locTypeRef.Id = &locTypeID

	createResp, err := c.client.DcimLocationsCreateWithResponse(c.ctx,
		&nautobotapi.DcimLocationsCreateParams{},
		nautobotapi.LocationRequest{
			Name:         name,
			Status:       statusRef,
			LocationType: locTypeRef,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create location %s: %w", name, err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		clog.Error("[nautobot] Location create failed for %s: %s", name, string(createResp.Body))
		return nil, fmt.Errorf("failed to create location %s: status %d: %s", name, createResp.StatusCode(), string(createResp.Body))
	}

	if createResp.JSON201 != nil {
		item := &CachedItem{
			ID:      toUUID(createResp.JSON201.Id),
			Name:    createResp.JSON201.Name,
			Display: *createResp.JSON201.Display,
		}
		c.locations[name] = item
		clog.Created("[nautobot] Created location: %s (ID: %s)", name, item.ID)
		return item, nil
	}

	return nil, fmt.Errorf("failed to create location %s: no response body", name)
}

// GetOrCreateLocationType gets or creates a location type by name
func (c *LookupCache) GetOrCreateLocationType(name string) (*CachedItem, error) {
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
		return &CachedItem{
			ID:      toUUID(lt.Id),
			Name:    lt.Name,
			Display: *lt.Display,
		}, nil
	}

	// Create the location type
	clog.Detail("[nautobot] Creating location type: %s", name)
	contentTypes := []string{"dcim.device", "dcim.rack"}

	createResp, err := c.client.DcimLocationTypesCreateWithResponse(c.ctx,
		&nautobotapi.DcimLocationTypesCreateParams{},
		nautobotapi.LocationTypeRequest{
			Name:         name,
			ContentTypes: &contentTypes,
		},
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

// FindNameByID searches the cache for an item with the given UUID and returns
// its display name. The cacheType parameter narrows the search to a specific
// cache: "deviceType", "location", "status", "role", "rack", or "device".
// Returns the UUID string when no cached name is found.
func (c *LookupCache) FindNameByID(cacheType string, id uuid.UUID) string {
	if id == uuid.Nil {
		return "(none)"
	}

	var caches []struct {
		mu    *sync.RWMutex
		items map[string]*CachedItem
	}

	switch cacheType {
	case "deviceType":
		caches = append(caches, struct {
			mu    *sync.RWMutex
			items map[string]*CachedItem
		}{&c.deviceTypesMu, c.deviceTypes})
	case "location":
		caches = append(caches, struct {
			mu    *sync.RWMutex
			items map[string]*CachedItem
		}{&c.locationsMu, c.locations})
	case "status":
		caches = append(caches, struct {
			mu    *sync.RWMutex
			items map[string]*CachedItem
		}{&c.statusesMu, c.statuses})
	case "role":
		caches = append(caches, struct {
			mu    *sync.RWMutex
			items map[string]*CachedItem
		}{&c.rolesMu, c.roles})
	case "device":
		caches = append(caches, struct {
			mu    *sync.RWMutex
			items map[string]*CachedItem
		}{&c.devicesMu, c.devices})
	default:
		// Search all caches
		caches = append(caches,
			struct {
				mu    *sync.RWMutex
				items map[string]*CachedItem
			}{&c.deviceTypesMu, c.deviceTypes},
			struct {
				mu    *sync.RWMutex
				items map[string]*CachedItem
			}{&c.locationsMu, c.locations},
			struct {
				mu    *sync.RWMutex
				items map[string]*CachedItem
			}{&c.statusesMu, c.statuses},
			struct {
				mu    *sync.RWMutex
				items map[string]*CachedItem
			}{&c.rolesMu, c.roles},
			struct {
				mu    *sync.RWMutex
				items map[string]*CachedItem
			}{&c.devicesMu, c.devices},
		)
	}

	for _, cache := range caches {
		cache.mu.RLock()
		for _, item := range cache.items {
			if item.ID == id {
				name := item.Name
				if name == "" {
					name = item.Display
				}
				cache.mu.RUnlock()
				return name
			}
		}
		cache.mu.RUnlock()
	}

	return id.String()
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

// interfaceCacheKey generates a cache key for an interface
func interfaceCacheKey(deviceID uuid.UUID, ifaceName string) string {
	return deviceID.String() + ":" + ifaceName
}

// CacheLocation adds a location to the local cache.
func (c *LookupCache) CacheLocation(name string, item *CachedItem) {
	c.locationsMu.Lock()
	defer c.locationsMu.Unlock()
	c.locations[name] = item
}

// CacheInterface adds an interface to the local cache
// This is used to cache newly created interfaces so cable creation can find them
func (c *LookupCache) CacheInterface(deviceID uuid.UUID, ifaceName string, item *CachedItem) {
	c.interfacesMu.Lock()
	defer c.interfacesMu.Unlock()
	key := interfaceCacheKey(deviceID, ifaceName)
	c.interfaces[key] = item
}

// GetInterfaceByDeviceAndName looks up an interface by device ID and interface name
func (c *LookupCache) GetInterfaceByDeviceAndName(deviceID uuid.UUID, ifaceName string) (*CachedItem, error) {
	// Check local cache first
	c.interfacesMu.RLock()
	key := interfaceCacheKey(deviceID, ifaceName)
	if item, ok := c.interfaces[key]; ok {
		c.interfacesMu.RUnlock()
		return item, nil
	}
	c.interfacesMu.RUnlock()

	// Query from Nautobot API
	deviceIDStr := []string{deviceID.String()}
	nameFilter := []string{ifaceName}

	resp, err := c.client.DcimInterfacesListWithResponse(c.ctx, &nautobotapi.DcimInterfacesListParams{
		Device: &deviceIDStr,
		Name:   &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup interface %s on device %s: %w", ifaceName, deviceID, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup interface %s on device %s: status %d", ifaceName, deviceID, resp.StatusCode())
	}

	if resp.JSON200 != nil && resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		iface := (resp.JSON200.Results)[0]
		item := &CachedItem{
			ID:      toUUID(iface.Id),
			Name:    iface.Name,
			Display: *iface.Display,
		}
		// Extract cable ID if interface has a cable attached
		if iface.Cable != nil && iface.Cable.Id != nil {
			if cableUUID, err := iface.Cable.Id.AsBulkWritableCableRequestStatusId0(); err == nil {
				item.CableID = uuid.UUID(cableUUID)
			}
		}
		// Cache the result
		c.CacheInterface(deviceID, ifaceName, item)
		return item, nil
	}

	return nil, nil // Not found is not an error
}

// GetInterfacesByDevice lists all interfaces for a device
func (c *LookupCache) GetInterfacesByDevice(deviceID uuid.UUID) ([]*CachedItem, error) {
	deviceIDStr := []string{deviceID.String()}

	resp, err := c.client.DcimInterfacesListWithResponse(c.ctx, &nautobotapi.DcimInterfacesListParams{
		Device: &deviceIDStr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list interfaces for device %s: %w", deviceID, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to list interfaces for device %s: status %d", deviceID, resp.StatusCode())
	}

	var items []*CachedItem
	if resp.JSON200 != nil && resp.JSON200.Results != nil {
		for _, iface := range resp.JSON200.Results {
			item := &CachedItem{
				ID:      toUUID(iface.Id),
				Name:    iface.Name,
				Display: *iface.Display,
			}
			// Extract cable ID if interface has a cable attached
			if iface.Cable != nil && iface.Cable.Id != nil {
				if cableUUID, err := iface.Cable.Id.AsBulkWritableCableRequestStatusId0(); err == nil {
					item.CableID = uuid.UUID(cableUUID)
				}
			}
			items = append(items, item)
		}
	}

	return items, nil
}

// GetInterfaceByDeviceAndNameFuzzy looks up an interface by device ID and interface name,
// with fuzzy matching to handle naming variations between cani and Nautobot.
// It tries: (1) exact match, (2) normalized numeric match (e.g., "1" matches "port1", "1/1"),
// (3) prefix-stripped match (e.g., "eth0" matches "Gig-E 0").
func (c *LookupCache) GetInterfaceByDeviceAndNameFuzzy(deviceID uuid.UUID, ifaceName string) (*CachedItem, error) {
	// First, try exact match
	item, err := c.GetInterfaceByDeviceAndName(deviceID, ifaceName)
	if err != nil {
		return nil, err
	}
	if item != nil {
		return item, nil
	}

	// Exact match failed, fetch all interfaces for this device and try fuzzy matching
	allInterfaces, err := c.GetInterfacesByDevice(deviceID)
	if err != nil {
		return nil, err
	}

	// Try to find a match using normalization
	normalizedSearch := normalizeInterfaceName(ifaceName)
	for _, iface := range allInterfaces {
		normalizedExisting := normalizeInterfaceName(iface.Name)
		if normalizedSearch == normalizedExisting {
			return iface, nil
		}
	}

	// Try matching just the numeric portion for port-style interfaces
	searchNum := extractPortNumber(ifaceName)
	if searchNum != "" {
		for _, iface := range allInterfaces {
			existingNum := extractPortNumber(iface.Name)
			if existingNum != "" && searchNum == existingNum {
				return iface, nil
			}
		}
	}

	return nil, nil // Not found
}

// normalizeInterfaceName normalizes an interface name by lowercasing and removing common prefixes
func normalizeInterfaceName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	// Remove common interface prefixes
	prefixes := []string{"port", "eth", "ethernet", "gigabitethernet", "gig-e ", "gig-e", "osfp", "sfp", "mgmt", "ib"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			name = strings.TrimPrefix(name, prefix)
			break
		}
	}
	// Remove leading separators
	name = strings.TrimLeft(name, " -_/")
	return name
}

// extractPortNumber extracts just the numeric port number from an interface name
// Examples: "1" -> "1", "port1" -> "1", "1/1" -> "1", "GigabitEthernet1/0/1" -> "1"
// For complex patterns like "1/0/1", returns the last number segment
func extractPortNumber(name string) string {
	// First try to find any numbers in the name
	var numParts []string
	current := ""
	for _, ch := range name {
		if ch >= '0' && ch <= '9' {
			current += string(ch)
		} else if current != "" {
			numParts = append(numParts, current)
			current = ""
		}
	}
	if current != "" {
		numParts = append(numParts, current)
	}

	if len(numParts) == 0 {
		return ""
	}

	// For simple ports like "1", "port1", return the first number
	// For hierarchical like "1/0/1", we use the first number as the primary identifier
	return numParts[0]
}

// GetCableByTerminations checks if a cable exists between two interfaces
// It checks both directions (A-B and B-A) since cables are bidirectional
// Returns the cable CachedItem if found, nil if not found
func (c *LookupCache) GetCableByTerminations(interfaceAID, interfaceBID uuid.UUID) (*CachedItem, error) {
	if c.ctx == nil {
		return nil, fmt.Errorf("lookup cache context not set, call SetContext first")
	}

	// Search for cables with termination_a matching interfaceA
	aID := openapi_types.UUID(interfaceAID)
	bID := openapi_types.UUID(interfaceBID)

	// Try A->B direction
	params := &nautobotapi.DcimCablesListParams{
		TerminationAId: &[]openapi_types.UUID{aID},
	}
	resp, err := c.client.DcimCablesListWithResponse(c.ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query cables: %w", err)
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected status %d querying cables", resp.StatusCode())
	}

	// Check if any cable has termination_b matching interfaceB
	for _, cable := range resp.JSON200.Results {
		if openapi_types.UUID(cable.TerminationBId) == bID {
			cableID := uuid.UUID(*cable.Id)
			label := ""
			if cable.Label != nil {
				label = *cable.Label
			}
			return &CachedItem{ID: cableID, Name: label}, nil
		}
	}

	// Try B->A direction (cable may be stored in reverse)
	params2 := &nautobotapi.DcimCablesListParams{
		TerminationAId: &[]openapi_types.UUID{bID},
	}
	resp2, err := c.client.DcimCablesListWithResponse(c.ctx, params2)
	if err != nil {
		return nil, fmt.Errorf("failed to query cables: %w", err)
	}
	if resp2.StatusCode() != http.StatusOK || resp2.JSON200 == nil {
		return nil, fmt.Errorf("unexpected status %d querying cables", resp2.StatusCode())
	}

	// Check if any cable has termination_b matching interfaceA
	for _, cable := range resp2.JSON200.Results {
		if openapi_types.UUID(cable.TerminationBId) == aID {
			cableID := uuid.UUID(*cable.Id)
			label := ""
			if cable.Label != nil {
				label = *cable.Label
			}
			return &CachedItem{ID: cableID, Name: label}, nil
		}
	}

	return nil, nil
}

// GetOrCreateTag looks up a tag by name, creating it if not found.
func (c *LookupCache) GetOrCreateTag(name string) (*CachedItem, error) {
	c.tagsMu.RLock()
	if item, ok := c.tags[name]; ok {
		c.tagsMu.RUnlock()
		return item, nil
	}
	c.tagsMu.RUnlock()

	c.tagsMu.Lock()
	defer c.tagsMu.Unlock()

	// Double-check after acquiring write lock.
	if item, ok := c.tags[name]; ok {
		return item, nil
	}

	// Try to find existing tag.
	nameFilter := []string{name}
	resp, err := c.client.ExtrasTagsListWithResponse(c.ctx, &nautobotapi.ExtrasTagsListParams{
		Name: &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup tag %s: %w", name, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup tag %s: status %d", name, resp.StatusCode())
	}

	if resp.JSON200 != nil && len(resp.JSON200.Results) > 0 {
		t := resp.JSON200.Results[0]
		item := &CachedItem{
			ID:   toUUID(t.Id),
			Name: t.Name,
		}
		c.tags[name] = item
		return item, nil
	}

	// Tag not found — create it.
	clog.Detail("[nautobot] Creating tag: %s", name)
	createResp, err := c.client.ExtrasTagsCreateWithResponse(c.ctx,
		&nautobotapi.ExtrasTagsCreateParams{},
		nautobotapi.TagRequest{
			Name:         name,
			ContentTypes: []string{},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag %s: %w", name, err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("failed to create tag %s: status %d: %s",
			name, createResp.StatusCode(), string(createResp.Body))
	}

	if createResp.JSON201 != nil {
		item := &CachedItem{
			ID:   toUUID(createResp.JSON201.Id),
			Name: createResp.JSON201.Name,
		}
		c.tags[name] = item
		clog.Created("[nautobot] Created tag: %s (ID: %s)", name, item.ID)
		return item, nil
	}

	return nil, fmt.Errorf("failed to create tag %s: no response body", name)
}
