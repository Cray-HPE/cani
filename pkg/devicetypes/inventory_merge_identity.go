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
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// IPAMKind names the IPAM object a merge error is about.
type IPAMKind string

const (
	IPAMKindPrefix    IPAMKind = "prefix"
	IPAMKindIPAddress IPAMKind = "IP address"
	IPAMKindVRF       IPAMKind = "VRF"
)

// ErrSourceIdentityConflict marks a natural-key match whose provider external
// ID differs from the incoming object's: the source deleted and recreated the
// object. The record is left as it is rather than following the new object by
// guesswork, so the operator removes it and imports again.
var ErrSourceIdentityConflict = errors.New("remove the record and import again")

// SourceIdentityConflictError carries what acting on ErrSourceIdentityConflict
// needs: which record to remove and the two IDs that disagree.
type SourceIdentityConflictError struct {
	Kind     IPAMKind
	Record   uuid.UUID // the record holding the natural key
	Source   string    // the provider whose IDs differ
	Existing uuid.UUID // the record's ID at the source
	Incoming uuid.UUID // the ID the source holds now
}

func (e *SourceIdentityConflictError) Error() string {
	return fmt.Sprintf("record %s carries %s ID %s but the source now holds %s; %v",
		e.Record, e.Source, e.Existing, e.Incoming, ErrSourceIdentityConflict)
}

func (e *SourceIdentityConflictError) Unwrap() error { return ErrSourceIdentityConflict }

// resolveIdentities maps each incoming object to the inventory UUID it takes:
// its own when present, else the record sharing a provider external ID, else
// the single record with its natural key, else its own (insert). The identity
// pass runs over the whole batch before any key matching, so a record its
// source moved to another key no longer answers to the old one, and a record
// the source moved onto a key another record holds is an error rather than a
// second record under one key. Errors name the object with kind and describe.
func resolveIdentities[T any, K comparable](
	kind IPAMKind,
	existing, incoming map[uuid.UUID]*T,
	keyOf func(*T) K,
	externalIDsOf func(*T) map[string]uuid.UUID,
	describe func(*T) string,
) (map[uuid.UUID]uuid.UUID, error) {
	index := newIdentityIndex(kind, existing, keyOf, externalIDsOf)
	remap := make(map[uuid.UUID]uuid.UUID, len(incoming))
	unmatched, moved, err := matchBySource(index, incoming, keyOf, externalIDsOf, describe, remap)
	if err != nil {
		return nil, err
	}
	for _, id := range moved {
		if holder, taken := index.keyTakenFrom(remap[id]); taken {
			return nil, fmt.Errorf("%s %s: record %s now holds the natural key of record %s; resolve the duplicate explicitly",
				kind, describe(incoming[id]), remap[id], holder)
		}
	}
	for _, id := range unmatched {
		obj := incoming[id]
		target, err := index.keyMatch(id, keyOf(obj), externalIDsOf(obj))
		if err != nil {
			return nil, fmt.Errorf("%s %s: %w", kind, describe(obj), err)
		}
		index.replace(target, keyOf(obj), externalIDsOf(obj))
		remap[id] = target
	}
	return remap, nil
}

// matchBySource resolves, in incoming-UUID order, every object that is a
// record itself or shares a provider external ID with one, re-indexing the
// record under the object's key and external IDs. It returns the IDs left
// for key matching and the IDs whose record changed key; external IDs that
// name several records are an error.
func matchBySource[T any, K comparable](
	index *identityIndex[K], incoming map[uuid.UUID]*T,
	keyOf func(*T) K, externalIDsOf func(*T) map[string]uuid.UUID, describe func(*T) string,
	remap map[uuid.UUID]uuid.UUID,
) (unmatched, moved []uuid.UUID, err error) {
	for _, id := range sortedIDs(incoming) {
		obj := incoming[id]
		if obj == nil {
			continue
		}
		target, ok, err := index.identityMatch(id, externalIDsOf(obj))
		if err != nil {
			return nil, nil, fmt.Errorf("%s %s: %w", index.kind, describe(obj), err)
		}
		if !ok {
			unmatched = append(unmatched, id)
			continue
		}
		if index.keys[target] != keyOf(obj) {
			moved = append(moved, id)
		}
		index.replace(target, keyOf(obj), externalIDsOf(obj))
		remap[id] = target
	}
	return unmatched, moved, nil
}

// applyIdentities stores each incoming object under its resolved UUID, in
// incoming-UUID order so a record several objects resolve to ends the same
// way on every run.
func applyIdentities[T any](
	existing, incoming map[uuid.UUID]*T, remap map[uuid.UUID]uuid.UUID, setID func(*T, uuid.UUID),
) {
	for _, id := range sortedIDs(incoming) {
		if target, ok := remap[id]; ok {
			setID(incoming[id], target)
			existing[target] = incoming[id]
		}
	}
}
