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
	"strings"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// interfaceCacheKey generates a cache key for an interface
func interfaceCacheKey(deviceID uuid.UUID, ifaceName string) string {
	return deviceID.String() + ":" + ifaceName
}

// CacheInterface adds an interface to the local cache
// This is used to cache newly created interfaces so cable creation can find them
func (c *LookupCache) CacheInterface(deviceID uuid.UUID, ifaceName string, item *CachedItem) {
	c.interfacesMu.Lock()
	defer c.interfacesMu.Unlock()
	key := interfaceCacheKey(deviceID, ifaceName)
	c.interfaces[key] = item
}

// GetInterfaceByDeviceAndName looks up an interface by device ID and interface name.
// On a cache miss it fetches ALL interfaces for the device in a single API
// call rather than filtering by name.  This avoids 400 errors from the
// Nautobot API when interface names contain "/" characters (e.g. "1/1/14").
func (c *LookupCache) GetInterfaceByDeviceAndName(deviceID uuid.UUID, ifaceName string) (*CachedItem, error) {
	// Check local cache first
	c.interfacesMu.RLock()
	key := interfaceCacheKey(deviceID, ifaceName)
	if item, ok := c.interfaces[key]; ok {
		c.interfacesMu.RUnlock()
		return item, nil
	}
	c.interfacesMu.RUnlock()

	// Fetch all interfaces for this device and populate cache.
	if err := c.PrefetchInterfacesForDevice(deviceID); err != nil {
		return nil, fmt.Errorf("failed to lookup interface %s on device %s: %w", ifaceName, deviceID, err)
	}

	// Re-check cache after prefetch
	c.interfacesMu.RLock()
	item, ok := c.interfaces[key]
	c.interfacesMu.RUnlock()
	if ok {
		return item, nil
	}

	return nil, nil // Not found is not an error
}

// GetInterfacesByDevice lists the interfaces a device owns directly. Nautobot
// 3.2 also returns the ports of a module installed in the device under the
// device filter; module__isnull excludes them because the device cache is
// keyed by name, so a module port named like a device port would otherwise
// replace it. Module ports are listed by GetInterfacesByModule.
func (c *LookupCache) GetInterfacesByDevice(deviceID uuid.UUID) ([]*CachedItem, error) {
	deviceIDStr := []string{deviceID.String()}
	deviceOwned := true
	return c.listInterfaces("device", deviceID, &nautobotapi.DcimInterfacesListParams{
		Device:       &deviceIDStr,
		ModuleIsnull: &deviceOwned,
	})
}

// listInterfaces runs one interface list query for an owner and maps every
// result, including the cable attached to it, into a CachedItem.
func (c *LookupCache) listInterfaces(kind string, ownerID uuid.UUID, params *nautobotapi.DcimInterfacesListParams) ([]*CachedItem, error) {
	limit := 1000
	params.Limit = &limit
	resp, err := c.client.DcimInterfacesListWithResponse(c.ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list interfaces for %s %s: %w", kind, ownerID, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to list interfaces for %s %s: status %d: %s",
			kind, ownerID, resp.StatusCode(), strings.TrimSpace(string(resp.Body)))
	}
	var items []*CachedItem
	if resp.JSON200 != nil {
		for _, iface := range resp.JSON200.Results {
			items = append(items, cachedInterfaceItem(iface))
		}
	}
	return items, nil
}

// cachedInterfaceItem maps one Nautobot interface into a CachedItem carrying
// its remote state and the ID of any cable attached to it.
func cachedInterfaceItem(iface nautobotapi.Interface) *CachedItem {
	item := &CachedItem{
		ID:      toUUID(iface.Id),
		Name:    iface.Name,
		Display: derefString(iface.Display),
		Iface:   remoteInterfaceState(iface),
	}
	// Nautobot 3.2 returns the cable as a generic nested object map.
	if iface.Cable != nil {
		if idStr, ok := (*iface.Cable)["id"].(string); ok {
			if cableUUID, err := uuid.Parse(idStr); err == nil {
				item.CableID = cableUUID
			}
		}
	}
	return item
}

// InvalidateInterfacePrefetch marks a device's interface cache as stale so
// that the next lookup re-reads the device's ports from Nautobot, including
// any created after the first prefetch.
func (c *LookupCache) InvalidateInterfacePrefetch(deviceID uuid.UUID) {
	c.interfacesMu.Lock()
	defer c.interfacesMu.Unlock()
	delete(c.interfacesPrefetched, deviceID)
}

// PrefetchInterfacesForDevice fetches all interfaces for a device from
// Nautobot in a single API call and populates the local cache.  This
// avoids per-interface queries that can fail when interface names
// contain characters (like "/") that cause issues in query filters.
func (c *LookupCache) PrefetchInterfacesForDevice(deviceID uuid.UUID) error {
	c.interfacesMu.RLock()
	if c.interfacesPrefetched[deviceID] {
		c.interfacesMu.RUnlock()
		return nil
	}
	c.interfacesMu.RUnlock()

	items, err := c.GetInterfacesByDevice(deviceID)
	if err != nil {
		return err
	}
	for _, item := range items {
		c.CacheInterface(deviceID, item.Name, item)
	}

	c.interfacesMu.Lock()
	c.interfacesPrefetched[deviceID] = true
	c.interfacesMu.Unlock()
	return nil
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
	return matchInterfaceFuzzy(allInterfaces, ifaceName), nil
}

// matchInterfaceFuzzy returns the first interface whose name matches after
// normalization, otherwise the first whose trailing port number matches, or
// nil when neither does.
func matchInterfaceFuzzy(items []*CachedItem, ifaceName string) *CachedItem {
	normalizedSearch := normalizeInterfaceName(ifaceName)
	for _, iface := range items {
		if normalizeInterfaceName(iface.Name) == normalizedSearch {
			return iface
		}
	}

	// Try matching just the numeric portion for port-style interfaces
	searchNum := extractPortNumber(ifaceName)
	if searchNum == "" {
		return nil
	}
	for _, iface := range items {
		if existingNum := extractPortNumber(iface.Name); existingNum != "" && searchNum == existingNum {
			return iface
		}
	}
	return nil
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

// extractPortNumber extracts just the numeric port number from an interface name.
// Examples: "1" -> "1", "port1" -> "1", "1/1/9" -> "9", "GigabitEthernet1/0/24" -> "24".
// For hierarchical patterns like "1/1/9" it returns the LAST numeric segment, which
// is the port number (the earlier segments identify chassis/slot). This lets a bare
// "9" match "1/1/9".
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

	// The last segment is the port number: for hierarchical names like "1/1/9"
	// the earlier segments are chassis/slot; for simple names it is the only part.
	return numParts[len(numParts)-1]
}
