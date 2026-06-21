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

// locationCreateOnMissServer answers the requests GetLocation makes when a
// location is missing and auto-creation is enabled: the locations list GET
// (returns empty so the lookup misses), the location-type lookup (returns an
// existing "Section" type), and the locations create POST. postCalls counts
// only the create POSTs.
func locationCreateOnMissServer(locTypeID, createdLocID uuid.UUID, postCalls *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "dcim/location-types"):
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, fmt.Sprintf(`{"count":1,"results":[%s]}`,
				refObjectJSON(locTypeID, "Section")))
		case strings.Contains(r.URL.Path, "dcim/locations"):
			if r.Method == http.MethodPost {
				*postCalls++
				w.WriteHeader(http.StatusCreated)
				_, _ = io.WriteString(w, refObjectJSON(createdLocID, "DC1"))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		}
	}
}

// -----------------------------------------------------------------------------
// GetLocation — caches on hit, auto-creates when enabled, errors otherwise.
// -----------------------------------------------------------------------------

// TestGetLocation_FetchesAndCachesFromAPI verifies GetLocation fetches a location
// by name from the API and caches it for subsequent resolution.
//
// Why it matters: devices, racks and IPAM all resolve their location by name;
// caching the hit avoids repeated lookups during a single export.
// Inputs: a location name. Outputs: a *CachedItem and an error; the cache map is
// asserted to hold the result.
// Data choice: name "DC1" with one matching result is the canonical hit, and the
// test inspects the locations cache directly to confirm the write.
func TestGetLocation_FetchesAndCachesFromAPI(t *testing.T) {
	locID := uuid.New()
	var calls int
	body := fmt.Sprintf(`{"count":1,"results":[%s]}`, refObjectJSON(locID, "DC1"))
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	item, err := e.Cache.GetLocation("DC1")
	if err != nil {
		t.Fatalf("GetLocation() error = %v", err)
	}
	if item == nil || item.ID != locID || item.Name != "DC1" {
		t.Fatalf("expected location %s, got %+v", locID, item)
	}

	e.Cache.locationsMu.RLock()
	cached, ok := e.Cache.locations["DC1"]
	e.Cache.locationsMu.RUnlock()
	if !ok || cached.ID != locID {
		t.Errorf("expected DC1 cached after lookup, got %+v (ok=%v)", cached, ok)
	}
}

// TestGetLocation_ReturnsErrorOnNon200 verifies a non-200 locations list response
// is surfaced as an error.
//
// Why it matters: a server error must abort resolution rather than be treated as
// "location absent", which could trigger an unwanted auto-create.
// Inputs: a name with the server returning 500. Outputs: a non-nil error.
// Data choice: a 500 with a detail body models a transient Nautobot failure.
func TestGetLocation_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{"detail":"boom"}`))
	defer cleanup()

	if _, err := e.Cache.GetLocation("DC1"); err == nil {
		t.Fatal("expected an error when the locations list responds with 500")
	}
}

// TestGetLocation_NotFoundWithoutCreateReturnsError verifies GetLocation errors
// when the location is missing and auto-creation is disabled.
//
// Why it matters: by default cani must not invent locations; a missing required
// location is an operator error that should stop the export.
// Inputs: a name with an empty result list and createLocations=false (default).
// Outputs: a non-nil error.
// Data choice: name "ghost" against an empty list models a required location the
// operator forgot to define.
func TestGetLocation_NotFoundWithoutCreateReturnsError(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, emptyListJSON))
	defer cleanup()
	// createLocations defaults to false: a missing location is an error.

	if _, err := e.Cache.GetLocation("ghost"); err == nil {
		t.Fatal("expected an error when the location is missing and creation is disabled")
	}
}

// TestGetLocation_CreatesWhenNotFoundAndCreateEnabled verifies GetLocation
// auto-creates the location when it is missing and creation is enabled.
//
// Why it matters: some workflows opt into letting the export create locations on
// demand; this exercises that resolve-or-create path end to end.
// Inputs: a name, with SetCreateLocations(true) and a server that misses then
// accepts a create POST. Outputs: the created *CachedItem and an error; exactly
// one create POST is expected.
// Data choice: the server returns an existing "Section" location-type so the
// create can resolve its type FK, isolating the location create itself.
func TestGetLocation_CreatesWhenNotFoundAndCreateEnabled(t *testing.T) {
	locTypeID := uuid.New()
	createdLocID := uuid.New()
	var postCalls int
	e, cleanup := newExporterWithServer(t, locationCreateOnMissServer(locTypeID, createdLocID, &postCalls))
	defer cleanup()
	seedActiveStatus(t, e)
	e.Cache.SetCreateLocations(true)

	item, err := e.Cache.GetLocation("DC1")
	if err != nil {
		t.Fatalf("GetLocation() error = %v", err)
	}
	if item == nil || item.ID != createdLocID {
		t.Errorf("expected created location %s, got %+v", createdLocID, item)
	}
	if postCalls != 1 {
		t.Errorf("expected exactly one location create POST, got %d", postCalls)
	}
}

// -----------------------------------------------------------------------------
// LookupLocation — never creates; returns nil,nil when not found.
// -----------------------------------------------------------------------------

// TestLookupLocation_FetchesAndCachesFromAPI verifies LookupLocation fetches a
// location by name and caches it, without ever creating one.
//
// Why it matters: LookupLocation is the non-creating sibling of GetLocation used
// where a miss should be tolerated; it must still cache hits for reuse.
// Inputs: a location name. Outputs: a *CachedItem and an error; the cache is
// asserted to hold the result.
// Data choice: name "Room-A" with one match exercises the hit-and-cache path.
func TestLookupLocation_FetchesAndCachesFromAPI(t *testing.T) {
	locID := uuid.New()
	var calls int
	body := fmt.Sprintf(`{"count":1,"results":[%s]}`, refObjectJSON(locID, "Room-A"))
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	item, err := e.Cache.LookupLocation("Room-A")
	if err != nil {
		t.Fatalf("LookupLocation() error = %v", err)
	}
	if item == nil || item.ID != locID {
		t.Fatalf("expected location %s, got %+v", locID, item)
	}

	e.Cache.locationsMu.RLock()
	_, ok := e.Cache.locations["Room-A"]
	e.Cache.locationsMu.RUnlock()
	if !ok {
		t.Error("expected Room-A to be cached after lookup")
	}
}

// TestLookupLocation_ReturnsNilWhenNotFound verifies LookupLocation returns
// (nil, nil) when no location matches.
//
// Why it matters: callers rely on a nil result (not an error) to decide whether
// to create the location during the locations phase.
// Inputs: an unmatched name. Outputs: nil item and nil error.
// Data choice: name "ghost" against an empty list models a clean miss.
func TestLookupLocation_ReturnsNilWhenNotFound(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, emptyListJSON))
	defer cleanup()

	item, err := e.Cache.LookupLocation("ghost")
	if err != nil {
		t.Fatalf("LookupLocation() error = %v", err)
	}
	if item != nil {
		t.Errorf("expected nil item when the location is not found, got %+v", item)
	}
}

// TestLookupLocation_ReturnsErrorOnNon200 verifies a non-200 locations list
// response is surfaced as an error.
//
// Why it matters: a server error must not be mistaken for "location absent",
// which would cause a spurious create.
// Inputs: a name with the server returning 500. Outputs: a non-nil error.
// Data choice: a 500 with a detail body models a transient Nautobot failure.
func TestLookupLocation_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{"detail":"boom"}`))
	defer cleanup()

	if _, err := e.Cache.LookupLocation("Room-A"); err == nil {
		t.Fatal("expected an error when the locations list responds with 500")
	}
}
