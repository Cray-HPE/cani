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
package devicetypes

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// identityIndex tracks the natural key and provider external IDs of every
// record during an IPAM merge, so each incoming object resolves in constant
// time and a record taken over by an incoming object answers to its new key
// and external IDs, not its old ones.
type identityIndex[K comparable] struct {
	byKey    map[K][]uuid.UUID
	bySource map[string]map[uuid.UUID]uuid.UUID
	keys     map[uuid.UUID]K
	sources  map[uuid.UUID]map[string]uuid.UUID
}

func newIdentityIndex[T any, K comparable](
	items map[uuid.UUID]*T, keyOf func(*T) K, externalIDsOf func(*T) map[string]uuid.UUID,
) *identityIndex[K] {
	index := &identityIndex[K]{
		byKey:    make(map[K][]uuid.UUID, len(items)),
		bySource: make(map[string]map[uuid.UUID]uuid.UUID),
		keys:     make(map[uuid.UUID]K, len(items)),
		sources:  make(map[uuid.UUID]map[string]uuid.UUID, len(items)),
	}
	for _, id := range sortedIDs(items) {
		if item := items[id]; item != nil {
			index.add(id, keyOf(item), externalIDsOf(item))
		}
	}
	return index
}

// add indexes a record; an external ID already held by a lower UUID keeps
// pointing there.
func (x *identityIndex[K]) add(id uuid.UUID, key K, externalIDs map[string]uuid.UUID) {
	x.keys[id] = key
	x.byKey[key] = append(x.byKey[key], id)
	x.sources[id] = externalIDs
	for source, externalID := range externalIDs {
		if externalID == uuid.Nil {
			continue
		}
		if x.bySource[source] == nil {
			x.bySource[source] = make(map[uuid.UUID]uuid.UUID)
		}
		if _, held := x.bySource[source][externalID]; !held {
			x.bySource[source][externalID] = id
		}
	}
}

// remove drops a record from the index.
func (x *identityIndex[K]) remove(id uuid.UUID) {
	key, ok := x.keys[id]
	if !ok {
		return
	}
	x.byKey[key] = slices.DeleteFunc(x.byKey[key], func(other uuid.UUID) bool { return other == id })
	for source, externalID := range x.sources[id] {
		if x.bySource[source][externalID] == id {
			delete(x.bySource[source], externalID)
		}
	}
	delete(x.keys, id)
	delete(x.sources, id)
}

// replace re-indexes a record under the incoming object that takes it over.
func (x *identityIndex[K]) replace(id uuid.UUID, key K, externalIDs map[string]uuid.UUID) {
	x.remove(id)
	x.add(id, key, externalIDs)
}

// identityMatch returns the record that is the incoming object itself or the
// one record sharing its provider external IDs. External IDs that name
// several records are an error: those records are one object the sources
// never merged locally, and picking one of them would be a guess.
func (x *identityIndex[K]) identityMatch(id uuid.UUID, externalIDs map[string]uuid.UUID) (uuid.UUID, bool, error) {
	if _, ok := x.keys[id]; ok {
		return id, true, nil
	}
	records := x.recordsBySource(externalIDs)
	switch len(records) {
	case 0:
		return uuid.Nil, false, nil
	case 1:
		return records[0], true, nil
	default:
		return uuid.Nil, false, fmt.Errorf("source IDs name %d existing records (%v); resolve the duplicates explicitly",
			len(records), records)
	}
}

// recordsBySource returns the distinct records holding any of the external
// IDs, lowest UUID first.
func (x *identityIndex[K]) recordsBySource(externalIDs map[string]uuid.UUID) []uuid.UUID {
	var records []uuid.UUID
	for source, externalID := range externalIDs {
		record, ok := x.bySource[source][externalID]
		if ok && externalID != uuid.Nil && !slices.Contains(records, record) {
			records = append(records, record)
		}
	}
	slices.SortFunc(records, func(a, b uuid.UUID) int { return bytes.Compare(a[:], b[:]) })
	return records
}

// keyMatch returns the single record holding key, or the incoming ID itself
// when none does. A match whose provider external IDs conflict, and a key
// held by several records, are errors rather than a second record as module,
// FRU and cable merges would insert: Nautobot holds one object per natural key
// in a namespace.
func (x *identityIndex[K]) keyMatch(id uuid.UUID, key K, externalIDs map[string]uuid.UUID) (uuid.UUID, error) {
	matches := x.byKey[key]
	switch len(matches) {
	case 0:
		return id, nil
	case 1:
		if conflictingExternalID(x.sources[matches[0]], externalIDs) {
			return uuid.Nil, fmt.Errorf("existing %s has a different source identity; "+
				"resolve the conflict explicitly", matches[0])
		}
		return matches[0], nil
	default:
		return uuid.Nil, fmt.Errorf("natural key is ambiguous among %d existing records (%v); "+
			"resolve the duplicates explicitly", len(matches), matches)
	}
}

// keyTakenFrom returns another record holding the same natural key as id.
// Checked only after the whole source pass, so two records that swapped keys
// in one batch each hold their key alone.
func (x *identityIndex[K]) keyTakenFrom(id uuid.UUID) (uuid.UUID, bool) {
	for _, holder := range x.byKey[x.keys[id]] {
		if holder != id {
			return holder, true
		}
	}
	return uuid.Nil, false
}
