/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// GetOrCreateNamespace — create-on-miss (PASSING scenario)
// -----------------------------------------------------------------------------

// TestGetOrCreateNamespace_CreatesWhenMissing verifies that a namespace absent
// from Nautobot is created via POST, that the returned item and the package
// cache capture its new remote ID, and that a second call is served from cache
// with no further POST.
//
// Why it matters: every prefix and IP export resolves the "Global" namespace
// first, so this must create-on-miss exactly once and then memoize to avoid
// hammering the API each phase.
// Inputs: namespace name "TestNS-Create" with an empty GET lookup. Outputs:
// item.ID equals the server-assigned UUID and exactly one create POST occurs.
// Data choice: the create response includes "display" because the loader
// dereferences that field when building the cached item.
func TestGetOrCreateNamespace_CreatesWhenMissing(t *testing.T) {
	// Arrange.
	resetIPAMCaches()

	const nsName = "TestNS-Create"
	nsNID := uuid.New()

	var posts int
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !strings.Contains(r.URL.Path, "ipam/namespaces") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(emptyListJSON))
			return
		}
		if r.Method == http.MethodPost {
			posts++
			w.WriteHeader(http.StatusCreated)
			// Display must be present; the loader dereferences it.
			_, _ = w.Write([]byte(`{"id":"` + nsNID.String() + `","name":"` + nsName + `","display":"` + nsName + `"}`))
			return
		}
		// GET lookup -> not found, forcing creation.
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(emptyListJSON))
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	// Act.
	item, err := e.Cache.GetOrCreateNamespace(nsName)

	// Assert.
	if err != nil {
		t.Fatalf("GetOrCreateNamespace returned error: %v", err)
	}
	if item == nil {
		t.Fatal("expected a namespace item, got nil")
	}
	if item.ID != nsNID {
		t.Errorf("namespace ID = %s, want %s", item.ID, nsNID)
	}
	if item.Name != nsName {
		t.Errorf("namespace Name = %q, want %q", item.Name, nsName)
	}
	if posts != 1 {
		t.Errorf("expected exactly 1 create POST, got %d", posts)
	}

	// A subsequent call must hit the cache and issue no additional POST.
	again, err := e.Cache.GetOrCreateNamespace(nsName)
	if err != nil {
		t.Fatalf("second GetOrCreateNamespace returned error: %v", err)
	}
	if again.ID != nsNID {
		t.Errorf("cached namespace ID = %s, want %s", again.ID, nsNID)
	}
	if posts != 1 {
		t.Errorf("expected no further POST on cache hit, total POSTs = %d", posts)
	}
}

// -----------------------------------------------------------------------------
// GetOrCreateNamespace — cache hit short-circuits any API call
// -----------------------------------------------------------------------------

// TestGetOrCreateNamespace_ReturnsCachedWithoutAPICall verifies that a
// pre-populated namespace is returned straight from the cache with no HTTP
// round-trip at all.
//
// Why it matters: the namespace cache is a package-level global shared across
// every IPAM phase, so a cache hit must short-circuit to keep repeated resolves
// from generating redundant Nautobot calls.
// Inputs: the global namespaces map seeded directly with "TestNS-Cached" before
// the call. Outputs: the cached item is returned; the fake server fails the
// test if it receives any request.
// Data choice: seeding the global map under its lock reproduces the
// post-first-resolve state without issuing a real create.
func TestGetOrCreateNamespace_ReturnsCachedWithoutAPICall(t *testing.T) {
	// Arrange.
	resetIPAMCaches()

	const nsName = "TestNS-Cached"
	cachedNID := uuid.New()

	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected %s %s — cache hit must not call the API", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(emptyListJSON))
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	namespacesMu.Lock()
	namespaces[nsName] = &CachedItem{ID: cachedNID, Name: nsName}
	namespacesMu.Unlock()

	// Act.
	item, err := e.Cache.GetOrCreateNamespace(nsName)

	// Assert.
	if err != nil {
		t.Fatalf("GetOrCreateNamespace returned error: %v", err)
	}
	if item == nil || item.ID != cachedNID {
		t.Errorf("expected cached namespace %s, got %+v", cachedNID, item)
	}
}
