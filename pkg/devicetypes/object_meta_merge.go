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
// non-nil tag list replaces the existing one. A custom field the source
// system reports also retires the copy of that key held in the source's own
// provider-metadata bucket, so the two cannot disagree.
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
		m.dropSupersededProviderKeys(incoming),
		m.mergeProviderMetadata(incoming.ProviderMetadata),
	}
	return slices.Contains(changes, true)
}

// dropSupersededProviderKeys removes, from the provider-metadata bucket of
// each source system that identifies incoming (the keys of its ExternalIDs),
// every key incoming reports as a custom field. Such a bucket entry is an
// older local write of the same field for that system; left in place it
// would override the value just read back on the next export. Buckets of
// other providers are untouched, and an emptied bucket is removed.
func (m *ObjectMeta) dropSupersededProviderKeys(incoming ObjectMeta) bool {
	changed := false
	for source := range incoming.ExternalIDs {
		bucket, ok := m.ProviderMetadata[source].(map[string]any)
		if !ok || !dropKeys(bucket, incoming.CustomFields) {
			continue
		}
		changed = true
		if len(bucket) == 0 {
			delete(m.ProviderMetadata, source)
		}
	}
	return changed
}

// dropKeys deletes from bucket every key present in keys and reports
// whether any was removed.
func dropKeys(bucket, keys map[string]any) bool {
	removed := false
	for key := range keys {
		if _, held := bucket[key]; held {
			delete(bucket, key)
			removed = true
		}
	}
	return removed
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
