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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// tagJSON renders a single tag object as returned by Nautobot.
func tagJSON(id uuid.UUID, name string) string {
	return fmt.Sprintf(`{"id":%q,"name":%q,"display":%q}`, id.String(), name, name)
}

// TestGetOrCreateTag_ReturnsCachedWithoutHTTP verifies that a tag already present
// in the in-memory cache is returned directly, with zero HTTP calls.
//
// Why it matters: tags annotate many exported objects; serving repeat requests
// from cache avoids re-querying Nautobot for the same tag and keeps the export
// fast and idempotent.
// Inputs: name "cached-tag", pre-seeded into the tags map. Outputs: the cached
// *CachedItem and a call count of 0.
// Data choice: seeding the tags map directly isolates the cache-hit path; the
// jsonHandler call counter proves the fake server is never contacted.
func TestGetOrCreateTag_ReturnsCachedWithoutHTTP(t *testing.T) {
	id := uuid.New()
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	e.Cache.tagsMu.Lock()
	e.Cache.tags["cached-tag"] = &CachedItem{ID: id, Name: "cached-tag"}
	e.Cache.tagsMu.Unlock()

	item, err := e.Cache.GetOrCreateTag("cached-tag")
	if err != nil {
		t.Fatalf("GetOrCreateTag() error = %v", err)
	}
	if item == nil || item.ID != id {
		t.Errorf("expected cached tag %s, got %+v", id, item)
	}
	if calls != 0 {
		t.Errorf("expected no HTTP calls for a cache hit, got %d", calls)
	}
}

// TestGetOrCreateTag_ReturnsExistingFromList verifies that when the tag already
// exists in Nautobot (returned by the list GET), it is reused and never
// re-created (no POST is issued).
//
// Why it matters: find-or-create idempotency means re-running an export must
// adopt the existing tag instead of duplicating it in the source-of-truth.
// Inputs: name "net"; the list GET returns one match. Outputs: the existing
// *CachedItem; the handler fails the test if any POST occurs.
// Data choice: a count:1 list result drives the "found" branch, and erroring on
// POST makes the no-create guarantee explicit.
func TestGetOrCreateTag_ReturnsExistingFromList(t *testing.T) {
	id := uuid.New()
	posted := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			posted = true
			t.Error("unexpected POST: tag already exists and must not be re-created")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, fmt.Sprintf(`{"count":1,"results":[%s]}`, tagJSON(id, "net")))
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	item, err := e.Cache.GetOrCreateTag("net")
	if err != nil {
		t.Fatalf("GetOrCreateTag() error = %v", err)
	}
	if item == nil || item.ID != id {
		t.Errorf("expected existing tag %s, got %+v", id, item)
	}
	if posted {
		t.Error("GetOrCreateTag must not POST when the tag already exists")
	}
}

// TestGetOrCreateTag_CreatesOnMiss verifies that when the list GET returns no
// match, a POST creates the tag and the created item is returned.
//
// Why it matters: first-time exports into a fresh Nautobot must auto-create any
// referenced tags so object annotation succeeds without manual pre-seeding.
// Inputs: name "fresh"; the list GET returns count:0 and the POST to extras/tags
// returns 201. Outputs: the created *CachedItem with the new ID/name.
// Data choice: an empty list forces the create branch, and matching the
// "extras/tags" path confirms the POST targets the tag endpoint specifically.
func TestGetOrCreateTag_CreatesOnMiss(t *testing.T) {
	created := uuid.New()
	posted := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "extras/tags") {
			posted = true
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, tagJSON(created, "fresh"))
			return
		}
		// GET list: report no existing tag so the create branch is taken.
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"count":0,"results":[]}`)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	item, err := e.Cache.GetOrCreateTag("fresh")
	if err != nil {
		t.Fatalf("GetOrCreateTag() error = %v", err)
	}
	if !posted {
		t.Error("expected a POST to create the missing tag")
	}
	if item == nil || item.ID != created || item.Name != "fresh" {
		t.Errorf("expected created tag %s, got %+v", created, item)
	}
}

// TestGetOrCreateTag_ReturnsErrorOnListFailure verifies that a 500 from the tag
// list GET is surfaced as an error.
//
// Why it matters: if the lookup itself fails, the export must abort rather than
// silently skip tagging and emit incompletely-annotated objects.
// Inputs: name "net"; server replies 500 to the list. Outputs: a non-nil error.
// Data choice: 500 simulates Nautobot/list unavailability; the empty `{}` body is
// irrelevant because only the status code is inspected.
func TestGetOrCreateTag_ReturnsErrorOnListFailure(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{}`))
	defer cleanup()

	if _, err := e.Cache.GetOrCreateTag("net"); err == nil {
		t.Fatal("expected an error when the tag list responds with 500")
	}
}

// TestGetOrCreateTag_ReturnsErrorWhenCreateFails verifies that, after an empty
// list, a 400 from the create POST is surfaced as an error.
//
// Why it matters: a failed tag creation must surface so the operator fixes it
// rather than the export continuing with objects that lack required tags.
// Inputs: name "fresh"; the list GET returns empty and the POST replies 400.
// Outputs: a non-nil error.
// Data choice: empty list then 400 isolates the create-failure branch, and the
// `{"detail":"bad"}` body mimics a Nautobot validation rejection.
func TestGetOrCreateTag_ReturnsErrorWhenCreateFails(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"detail":"bad"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"count":0,"results":[]}`)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	if _, err := e.Cache.GetOrCreateTag("fresh"); err == nil {
		t.Fatal("expected an error when tag creation responds with 400")
	}
}
