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
// ResolveNamespace — create-on-miss (PASSING scenario)
// -----------------------------------------------------------------------------

// TestResolveNamespace_CreatesWhenMissing verifies that a namespace absent
// from Nautobot is created via POST on a real run, that the returned item and
// the cache capture its new remote ID, and that a second call is served from
// cache with no further POST.
//
// Why it matters: every prefix, address and VRF export resolves its namespace
// first, so this must create-on-miss exactly once and then memoize to avoid
// hammering the API each phase.
// Inputs: namespace name "TestNS-Create" with an empty GET lookup. Outputs:
// item.ID equals the server-assigned UUID, the item is not Planned, and
// exactly one create POST occurs.
// Data choice: the create response includes "display" because the loader
// dereferences that field when building the cached item.
func TestResolveNamespace_CreatesWhenMissing(t *testing.T) {
	// Arrange.
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
	item, err := e.Cache.ResolveNamespace(nsName)

	// Assert.
	if err != nil {
		t.Fatalf("ResolveNamespace returned error: %v", err)
	}
	if item == nil || item.Planned {
		t.Fatalf("expected an existing namespace item, got %+v", item)
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
	again, err := e.Cache.ResolveNamespace(nsName)
	if err != nil {
		t.Fatalf("second ResolveNamespace returned error: %v", err)
	}
	if again.ID != nsNID {
		t.Errorf("cached namespace ID = %s, want %s", again.ID, nsNID)
	}
	if posts != 1 {
		t.Errorf("expected no further POST on cache hit, total POSTs = %d", posts)
	}
}

// -----------------------------------------------------------------------------
// ResolveNamespace — cache hit short-circuits any API call
// -----------------------------------------------------------------------------

// TestResolveNamespace_ReturnsCachedWithoutAPICall verifies that a
// pre-populated namespace is returned straight from the cache with no HTTP
// round-trip at all.
//
// Why it matters: the namespace cache is shared across every IPAM phase of an
// export, so a cache hit must short-circuit to keep repeated resolves from
// generating redundant Nautobot calls.
// Inputs: the exporter's namespace cache seeded directly with "TestNS-Cached"
// before the call. Outputs: the cached item is returned; the fake server fails
// the test if it receives any request.
// Data choice: seeding the cache map reproduces the post-first-resolve state
// without issuing a real create.
func TestResolveNamespace_ReturnsCachedWithoutAPICall(t *testing.T) {
	// Arrange.
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
	e.Cache.namespaces[nsName] = &CachedItem{ID: cachedNID, Name: nsName}

	// Act.
	item, err := e.Cache.ResolveNamespace(nsName)

	// Assert.
	if err != nil {
		t.Fatalf("ResolveNamespace returned error: %v", err)
	}
	if item == nil || item.ID != cachedNID || item.Planned {
		t.Errorf("expected cached namespace %s, got %+v", cachedNID, item)
	}
}

// TestResolveNamespace_DryRunPlansWithoutWriting verifies that a dry run
// plans a missing namespace instead of creating it and remembers the plan.
//
// Why it matters: --dry-run must never write to Nautobot, and every prefix and
// address in a new namespace resolves it, so later calls must not repeat the
// lookup either.
// Inputs: "TestNS-Planned" absent from Nautobot, resolved twice by a cache in
// dry-run mode. Outputs: a Planned nil-ID placeholder both times, one
// namespace GET, and no write request.
// Data choice: resolving twice proves the plan is cached, not merely that the
// first call skipped the POST.
func TestResolveNamespace_DryRunPlansWithoutWriting(t *testing.T) {
	// Arrange.
	var gets, writes int
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			writes++
			w.WriteHeader(http.StatusCreated)
			return
		}
		gets++
		_, _ = w.Write([]byte(emptyListJSON))
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	e.Cache.SetDryRun(true)

	for call := 1; call <= 2; call++ {
		// Act.
		item, err := e.Cache.ResolveNamespace("TestNS-Planned")

		// Assert.
		if err != nil || item == nil || !item.Planned || item.ID != uuid.Nil {
			t.Fatalf("call %d = (%+v, %v), want a Planned nil-ID placeholder", call, item, err)
		}
	}
	if gets != 1 || writes != 0 {
		t.Errorf("requests = %d GET and %d writes, want 1 GET and no writes", gets, writes)
	}
}
