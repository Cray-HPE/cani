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
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// seedTag inserts a tag directly into the cache so resolveTagRefs resolves it
// without any HTTP round-trip.
func seedTag(e *Exporter, name string, id uuid.UUID) {
	e.Cache.tagsMu.Lock()
	e.Cache.tags[name] = &CachedItem{ID: id, Name: name}
	e.Cache.tagsMu.Unlock()
}

// TestResolveTagRefs_EmptyReturnsNil verifies resolveTagRefs returns a nil
// pointer for nil and empty inputs and issues no HTTP calls.
//
// Why it matters: callers assign the result straight onto an optional request
// field; returning nil (not an empty slice) leaves the field unset so exports
// don't send an empty tags array or contact Nautobot when there is nothing to do.
// Inputs: nil and an empty []string. Outputs: nil both times, with the fake
// server's call counter at 0.
// Data choice: the two empty forms cover both guard branches; the counter proves
// the short-circuit happens before any lookup.
func TestResolveTagRefs_EmptyReturnsNil(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	if got := e.Cache.resolveTagRefs(nil); got != nil {
		t.Errorf("resolveTagRefs(nil) = %v, want nil", got)
	}
	if got := e.Cache.resolveTagRefs([]string{}); got != nil {
		t.Errorf("resolveTagRefs([]) = %v, want nil", got)
	}
	if calls != 0 {
		t.Errorf("expected no HTTP calls, got %d", calls)
	}
}

// TestResolveTagRefs_ResolvesCachedTagsWithoutHTTP verifies that each supplied
// tag name is resolved to a reference carrying the cached tag's UUID.
//
// Why it matters: exporting an object's tags depends on turning tag names into
// Nautobot references; a mismatch here would silently drop or mis-associate
// annotations like multi-chassis or hosts-edge on the exported objects.
// Inputs: two names pre-seeded into the cache. Outputs: a two-element ref slice
// whose JSON encoding contains both UUIDs, with zero HTTP calls.
// Data choice: pre-seeding isolates the resolve logic from lookup HTTP, and
// asserting the marshalled UUIDs proves the reference identity end to end.
func TestResolveTagRefs_ResolvesCachedTagsWithoutHTTP(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	id1, id2 := uuid.New(), uuid.New()
	seedTag(e, "multi-chassis", id1)
	seedTag(e, "hosts-edge", id2)

	refs := e.Cache.resolveTagRefs([]string{"multi-chassis", "hosts-edge"})
	if refs == nil || len(refs) != 2 {
		t.Fatalf("expected 2 tag refs, got %v", refs)
	}
	blob, _ := json.Marshal(refs)
	for _, id := range []uuid.UUID{id1, id2} {
		if !strings.Contains(string(blob), id.String()) {
			t.Errorf("tag ref payload missing id %s: %s", id, blob)
		}
	}
	if calls != 0 {
		t.Errorf("expected cache hits with no HTTP, got %d calls", calls)
	}
}

// TestResolveTagRefs_SkipsEmptyNames verifies that empty tag names are ignored
// while valid names still resolve.
//
// Why it matters: inventory data can carry blank tag entries; sending an empty
// tag name to Nautobot would error, so the resolver must drop them and still
// emit references for the real tags.
// Inputs: a slice of one empty string and one seeded name "real". Outputs: a
// single-element ref slice.
// Data choice: mixing a blank with a real name exercises the skip branch without
// hiding the happy path.
func TestResolveTagRefs_SkipsEmptyNames(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	id := uuid.New()
	seedTag(e, "real", id)

	refs := e.Cache.resolveTagRefs([]string{"", "real"})
	if refs == nil || len(refs) != 1 {
		t.Fatalf("expected 1 tag ref (empty name skipped), got %v", refs)
	}
}
