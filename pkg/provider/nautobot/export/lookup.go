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
	"context"
	"sync"

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	"github.com/google/uuid"
)

// Nautobot content-type identifiers used when assigning statuses and roles.
const (
	contentTypeDevice        = "dcim.device"
	contentTypeRack          = "dcim.rack"
	contentTypeModule        = "dcim.module"
	contentTypeInterface     = "dcim.interface"
	contentTypeInventoryItem = "dcim.inventoryitem"
	contentTypePrefix        = "ipam.prefix"
	contentTypeIPAddress     = "ipam.ipaddress"
	contentTypeVLAN          = "ipam.vlan"
)

// taggableContentTypes lists the Nautobot object types cani attaches tags to:
// devices, racks, interfaces, and inventory items (FRUs). A tag must advertise
// these content types or Nautobot rejects the assignment with "Related object
// not found using the provided attributes".
func taggableContentTypes() []string {
	return []string{contentTypeDevice, contentTypeRack, contentTypeInterface, contentTypeInventoryItem}
}

// missingContentTypes returns the entries in required that are absent from
// existing, preserving the order given in required.
func missingContentTypes(existing, required []string) []string {
	have := make(map[string]bool, len(existing))
	for _, ct := range existing {
		have[ct] = true
	}
	var missing []string
	for _, ct := range required {
		if !have[ct] {
			missing = append(missing, ct)
		}
	}
	return missing
}

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
	deviceTypes          map[string]*CachedItem // keyed by slug
	locations            map[string]*CachedItem // keyed by name
	statuses             map[string]*CachedItem // keyed by name
	roles                map[string]*CachedItem // keyed by name
	devices              map[string]*CachedItem // keyed by name
	manufacturers        map[string]*CachedItem // keyed by name
	interfaces           map[string]*CachedItem // keyed by "deviceID:ifaceName"
	interfacesPrefetched map[uuid.UUID]bool     // tracks devices whose interfaces have been prefetched
	tags                 map[string]*CachedItem // keyed by name

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

	// Options for auto-creation
	createDeviceTypes   bool
	createLocationTypes bool
	createStatuses      bool
	createRoles         bool
	createLocations     bool
}

// NewLookupCache creates a new lookup cache for the given client
func NewLookupCache(client *NautobotClient) *LookupCache {
	return &LookupCache{
		client:               client,
		ctx:                  context.Background(),
		deviceTypes:          make(map[string]*CachedItem),
		locations:            make(map[string]*CachedItem),
		statuses:             make(map[string]*CachedItem),
		roles:                make(map[string]*CachedItem),
		devices:              make(map[string]*CachedItem),
		manufacturers:        make(map[string]*CachedItem),
		interfaces:           make(map[string]*CachedItem),
		interfacesPrefetched: make(map[uuid.UUID]bool),
		tags:                 make(map[string]*CachedItem),
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

// SetCreateLocationTypes enables or disables auto-creation of location types
func (c *LookupCache) SetCreateLocationTypes(create bool) {
	c.createLocationTypes = create
}

// SetCreateModuleTypes enables or disables auto-creation of module types
func (c *LookupCache) SetCreateModuleTypes(create bool) {
	// Module type creation is gated at the Exporter.Options level,
	// but this setter is provided for symmetry with the other flags.
}
