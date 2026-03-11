/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
package devicetypes

import (
	"fmt"

	"github.com/google/uuid"
)

// providerKeyIndex maps [provider][key][value] → device UUID.
// It is a transient in-memory cache, never serialized. Rebuilt on
// load and kept current during merge operations.
type providerKeyIndex map[string]map[string]map[string]uuid.UUID

// RebuildProviderKeyIndex rebuilds the transient provider-key lookup
// index from all devices' ProviderMetadata. Call this once after
// loading the inventory from the datastore.
func (inv *Inventory) RebuildProviderKeyIndex() {
	idx := make(providerKeyIndex)
	for id, dev := range inv.Devices {
		if dev == nil || dev.ProviderMetadata == nil {
			continue
		}
		indexDeviceMetadata(idx, id, dev.ProviderMetadata)
	}
	inv.pkIndex = idx
}

// lookupProviderKey returns the device UUID for provider/key/value,
// or uuid.Nil if the index has no match.
func (inv *Inventory) lookupProviderKey(provider, key string, value any) uuid.UUID {
	if inv.pkIndex == nil {
		return uuid.Nil
	}
	valStr := toIndexValue(value)
	if valStr == "" {
		return uuid.Nil
	}
	byKey, ok := inv.pkIndex[provider]
	if !ok {
		return uuid.Nil
	}
	byVal, ok := byKey[key]
	if !ok {
		return uuid.Nil
	}
	id, ok := byVal[valStr]
	if !ok {
		return uuid.Nil
	}
	return id
}

// indexDevice adds all provider-key entries for a single device.
func (inv *Inventory) indexDevice(id uuid.UUID, dev *CaniDeviceType) {
	if inv.pkIndex == nil {
		inv.pkIndex = make(providerKeyIndex)
	}
	if dev == nil || dev.ProviderMetadata == nil {
		return
	}
	indexDeviceMetadata(inv.pkIndex, id, dev.ProviderMetadata)
}

// unindexDevice removes all provider-key entries for a single device.
func (inv *Inventory) unindexDevice(id uuid.UUID, dev *CaniDeviceType) {
	if inv.pkIndex == nil || dev == nil || dev.ProviderMetadata == nil {
		return
	}
	for provider, v := range dev.ProviderMetadata {
		sub, ok := v.(map[string]any)
		if !ok {
			continue
		}
		byKey, ok := inv.pkIndex[provider]
		if !ok {
			continue
		}
		for key, val := range sub {
			valStr := toIndexValue(val)
			if valStr == "" {
				continue
			}
			byVal, ok := byKey[key]
			if !ok {
				continue
			}
			if byVal[valStr] == id {
				delete(byVal, valStr)
			}
		}
	}
}

// indexDeviceMetadata populates the index for one device's metadata.
func indexDeviceMetadata(idx providerKeyIndex, id uuid.UUID, meta map[string]any) {
	for provider, v := range meta {
		sub, ok := v.(map[string]any)
		if !ok {
			continue
		}
		byKey, ok := idx[provider]
		if !ok {
			byKey = make(map[string]map[string]uuid.UUID)
			idx[provider] = byKey
		}
		for key, val := range sub {
			valStr := toIndexValue(val)
			if valStr == "" {
				continue
			}
			byVal, ok := byKey[key]
			if !ok {
				byVal = make(map[string]uuid.UUID)
				byKey[key] = byVal
			}
			byVal[valStr] = id
		}
	}
}

// toIndexValue converts an arbitrary metadata value to a string key
// suitable for index lookups. Returns "" for nil, empty strings, and
// types that cannot be meaningfully indexed (slices, maps).
func toIndexValue(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}
