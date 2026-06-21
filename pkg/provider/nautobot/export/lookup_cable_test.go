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

// cableListJSON renders a /dcim/cables/ list response containing a single cable
// that terminates at terminationB.
func cableListJSON(cableID, terminationB uuid.UUID, label string) string {
	return fmt.Sprintf(`{"count":1,"results":[{"id":%q,"termination_b_id":%q,"label":%q}]}`,
		cableID.String(), terminationB.String(), label)
}

const emptyCableList = `{"count":0,"results":[]}`

// TestGetCableByTerminations_FindsCableInAToBDirection verifies that the cable
// (ID and label) is returned when the first A->B query yields a cable whose
// termination_b matches interfaceB.
//
// Why it matters: before creating a cable the exporter checks whether one
// already exists; the forward-direction match prevents duplicate cables on
// re-export.
// Inputs: interfaceA/interfaceB UUIDs; the server returns a one-cable list with
// termination_b == ifaceB. Outputs: a CachedItem with ID==cableID, Name=="cab-1".
// Data choice: a single result whose termination_b equals ifaceB exercises the
// forward match and label extraction in one pass.
func TestGetCableByTerminations_FindsCableInAToBDirection(t *testing.T) {
	ifaceA, ifaceB := uuid.New(), uuid.New()
	cableID := uuid.New()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, cableListJSON(cableID, ifaceB, "cab-1"))
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	item, err := e.Cache.GetCableByTerminations(ifaceA, ifaceB)
	if err != nil {
		t.Fatalf("GetCableByTerminations() error = %v", err)
	}
	if item == nil || item.ID != cableID || item.Name != "cab-1" {
		t.Errorf("expected cable %s labelled cab-1, got %+v", cableID, item)
	}
}

// TestGetCableByTerminations_FindsCableInReverseDirection verifies that a cable
// stored B->A is still found: the forward query is empty and the reverse query
// (termination_a_id == ifaceB) returns a cable terminating at interfaceA.
//
// Why it matters: Nautobot cables are bidirectional, so the duplicate check
// must match regardless of which endpoint was stored as A; missing this would
// create duplicate reverse cables.
// Inputs: a server that returns empty for the A query and a cable
// (termination_b == ifaceA) for the B query. Outputs: a CachedItem with
// ID==cableID.
// Data choice: keying the handler on termination_a_id == ifaceB.String() forces
// the forward search to miss and the reverse branch to hit.
func TestGetCableByTerminations_FindsCableInReverseDirection(t *testing.T) {
	ifaceA, ifaceB := uuid.New(), uuid.New()
	cableID := uuid.New()
	// The cable is stored as B->A, so the first (A) query is empty and the
	// reverse (B) query returns a cable terminating at interface A.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.URL.Query().Get("termination_a_id") == ifaceB.String() {
			_, _ = io.WriteString(w, cableListJSON(cableID, ifaceA, "rev"))
			return
		}
		_, _ = io.WriteString(w, emptyCableList)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	item, err := e.Cache.GetCableByTerminations(ifaceA, ifaceB)
	if err != nil {
		t.Fatalf("GetCableByTerminations() error = %v", err)
	}
	if item == nil || item.ID != cableID {
		t.Errorf("expected reverse-direction cable %s, got %+v", cableID, item)
	}
}

// TestGetCableByTerminations_ReturnsNilWhenNoCableExists verifies that
// (nil, nil) is returned when neither direction yields a matching cable.
//
// Why it matters: a clean "no existing cable" signal is what lets
// createCaniCableType proceed to create a new cable; a false positive would
// wrongly skip it.
// Inputs: two random interface UUIDs; the server always returns an empty cable
// list. Outputs: nil item and nil error.
// Data choice: emptyCableList for every request models the common first-time
// export case where no cable exists yet.
func TestGetCableByTerminations_ReturnsNilWhenNoCableExists(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyCableList)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	item, err := e.Cache.GetCableByTerminations(uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("GetCableByTerminations() error = %v", err)
	}
	if item != nil {
		t.Errorf("expected nil when no cable connects the interfaces, got %+v", item)
	}
}

// TestGetCableByTerminations_ReturnsErrorOnNon200 verifies that an error is
// returned when the cable list query responds with 500.
//
// Why it matters: the caller (createCaniCableType) treats this error as a
// warn-and-continue signal, so the lookup must reliably surface non-200
// responses.
// Inputs: random UUIDs; the server returns 500 for any dcim/cables path.
// Outputs: a non-nil error.
// Data choice: gating the 500 on the dcim/cables path (other paths 200) ensures
// the failure comes from the cable query itself, not unrelated setup traffic.
func TestGetCableByTerminations_ReturnsErrorOnNon200(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "dcim/cables") {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, `{}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	if _, err := e.Cache.GetCableByTerminations(uuid.New(), uuid.New()); err == nil {
		t.Fatal("expected an error when the cable query responds with 500")
	}
}

// TestGetCableByTerminations_ReturnsErrorWhenContextMissing verifies that an
// error is returned and zero HTTP calls are made when the cache context
// (c.ctx) is nil.
//
// Why it matters: the lookup cache requires SetContext before use; failing fast
// without a request guards against a nil-context panic and wasted API traffic
// during export.
// Inputs: a cache with ctx set to nil and a call-counting server. Outputs: a
// non-nil error with calls==0.
// Data choice: asserting calls==0 proves the guard short-circuits before any
// DcimCablesList request is issued.
func TestGetCableByTerminations_ReturnsErrorWhenContextMissing(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, emptyCableList))
	defer cleanup()

	e.Cache.ctx = nil
	if _, err := e.Cache.GetCableByTerminations(uuid.New(), uuid.New()); err == nil {
		t.Fatal("expected an error when the lookup cache context is not set")
	}
	if calls != 0 {
		t.Errorf("expected no HTTP calls when context is missing, got %d", calls)
	}
}
