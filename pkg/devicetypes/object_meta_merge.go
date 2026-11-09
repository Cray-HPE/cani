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
	"reflect"
	"slices"
)

// Merge carries every set field of incoming into m and reports whether
// anything changed. An empty incoming value never clears an existing one;
// custom fields, external IDs and provider metadata merge key by key, and a
// non-nil tag list replaces the existing one.
func (m *ObjectMeta) Merge(incoming ObjectMeta) bool {
	if m == nil {
		return false
	}
	// Every merge must run, so the results are collected rather than
	// short-circuited.
	changes := []bool{
		mergeSet(&m.Status, incoming.Status),
		mergeSet(&m.Role, incoming.Role),
		mergeSet(&m.Tenant, incoming.Tenant),
		m.mergeTags(incoming.Tags),
		mergeMapKeys(&m.CustomFields, incoming.CustomFields),
		mergeMapKeys(&m.ExternalIDs, incoming.ExternalIDs),
		m.mergeProviderMetadata(incoming.ProviderMetadata),
	}
	return slices.Contains(changes, true)
}

func (m *ObjectMeta) mergeTags(incoming []string) bool {
	if incoming == nil || (m.Tags != nil && slices.Equal(m.Tags, incoming)) {
		return false
	}
	m.Tags = slices.Clone(incoming)
	return true
}

// mergeMapKeys copies every incoming key over *dst, allocating it on first
// use, and reports whether any value differed.
func mergeMapKeys[V any](dst *map[string]V, incoming map[string]V) bool {
	if len(incoming) == 0 {
		return false
	}
	if *dst == nil {
		*dst = make(map[string]V, len(incoming))
	}
	changed := false
	for k, v := range incoming {
		if old, ok := (*dst)[k]; ok && reflect.DeepEqual(old, v) {
			continue
		}
		(*dst)[k] = v
		changed = true
	}
	return changed
}

// mergeProviderMetadata merges provider sub-maps key by key and replaces
// top-level scalar keys.
func (m *ObjectMeta) mergeProviderMetadata(incoming map[string]any) bool {
	if len(incoming) == 0 {
		return false
	}
	if m.ProviderMetadata == nil {
		m.ProviderMetadata = make(map[string]any, len(incoming))
	}
	changed := false
	for k, v := range incoming {
		changed = m.mergeProviderKey(k, v) || changed
	}
	return changed
}

func (m *ObjectMeta) mergeProviderKey(k string, v any) bool {
	sub, isMap := v.(map[string]any)
	if !isMap {
		if reflect.DeepEqual(m.ProviderMetadata[k], v) {
			return false
		}
		m.ProviderMetadata[k] = v
		return true
	}
	existing, _ := m.ProviderMetadata[k].(map[string]any)
	if existing == nil {
		existing = make(map[string]any, len(sub))
		m.ProviderMetadata[k] = existing
	}
	return mergeMapKeys(&existing, sub)
}
