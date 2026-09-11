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
	"sync"

	"github.com/google/uuid"
)

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
