/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// GetInterfaceByModuleAndName looks up a module-owned interface by the
// module's Nautobot ID and the interface name. The interface cache is keyed
// by owner ID, and device lookups list only device-owned ports, so a module's
// ports never collide with same-named ports on a sibling module or on the
// parent device. Returns nil, nil when not found.
func (c *LookupCache) GetInterfaceByModuleAndName(moduleID uuid.UUID, ifaceName string) (*CachedItem, error) {
	c.interfacesMu.RLock()
	key := interfaceCacheKey(moduleID, ifaceName)
	item, ok := c.interfaces[key]
	c.interfacesMu.RUnlock()
	if ok {
		return item, nil
	}
	if err := c.prefetchInterfacesForModule(moduleID); err != nil {
		return nil, fmt.Errorf("failed to lookup interface %s on module %s: %w", ifaceName, moduleID, err)
	}
	c.interfacesMu.RLock()
	item, ok = c.interfaces[key]
	c.interfacesMu.RUnlock()
	if ok {
		return item, nil
	}
	return nil, nil
}

// GetInterfaceByModuleAndNameFuzzy looks up a module-owned interface by exact
// name and then by the fuzzy name, over a fresh list of the module's own
// ports. A naming difference between the inventory and Nautobot is absorbed
// without ever matching a port of the parent device or of a sibling module.
// Returns nil, nil when not found.
func (c *LookupCache) GetInterfaceByModuleAndNameFuzzy(moduleID uuid.UUID, ifaceName string) (*CachedItem, error) {
	item, err := c.GetInterfaceByModuleAndName(moduleID, ifaceName)
	if err != nil || item != nil {
		return item, err
	}
	ports, err := c.GetInterfacesByModule(moduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup interface %s on module %s: %w", ifaceName, moduleID, err)
	}
	return matchInterfaceFuzzy(ports, ifaceName), nil
}

// GetInterfacesByModule lists the interfaces a module owns, each with the
// cable attached to it, as GetInterfacesByDevice does for a device.
func (c *LookupCache) GetInterfacesByModule(moduleID uuid.UUID) ([]*CachedItem, error) {
	moduleFilter := []string{moduleID.String()}
	return c.listInterfaces("module", moduleID, &nautobotapi.DcimInterfacesListParams{Module: &moduleFilter})
}

// prefetchInterfacesForModule fetches a module's interfaces once and caches
// them under the module's ID.
func (c *LookupCache) prefetchInterfacesForModule(moduleID uuid.UUID) error {
	c.interfacesMu.RLock()
	done := c.interfacesPrefetched[moduleID]
	c.interfacesMu.RUnlock()
	if done {
		return nil
	}
	items, err := c.GetInterfacesByModule(moduleID)
	if err != nil {
		return err
	}
	for _, item := range items {
		c.CacheInterface(moduleID, item.Name, item)
	}
	c.interfacesMu.Lock()
	c.interfacesPrefetched[moduleID] = true
	c.interfacesMu.Unlock()
	return nil
}
