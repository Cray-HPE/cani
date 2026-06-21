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
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// LookupCache.List* helpers — each performs a single GET, populates its cache,
// and short-circuits on subsequent calls. These tests cover the three branches
// every List* method shares: cached short-circuit, HTTP success, HTTP error.
// -----------------------------------------------------------------------------

// jsonHandler returns a handler that always replies with the given status code
// and JSON body, recording how many times it was invoked.
func jsonHandler(calls *int, status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

// TestListLocations_ReturnsCachedItemsWithoutHTTP verifies that ListLocations
// returns the cached entries and issues no HTTP request once the cache is marked
// loaded.
//
// Why it matters: locations are referenced repeatedly while exporting devices;
// the loaded-flag short-circuit keeps the export from re-listing them on every
// call.
// Inputs: a cache pre-seeded with one location and locationsLoaded=true.
// Outputs: the single cached *CachedItem.
// Data choice: a nil client makes any HTTP attempt panic, so a clean return
// proves the cached branch was taken.
func TestListLocations_ReturnsCachedItemsWithoutHTTP(t *testing.T) {
	cache := NewLookupCache(nil) // nil client: any HTTP call would panic
	id := uuid.New()
	cache.locationsMu.Lock()
	cache.locations["Site-A"] = &CachedItem{ID: id, Name: "Site-A"}
	cache.locationsLoaded = true
	cache.locationsMu.Unlock()

	items, err := cache.ListLocations()
	if err != nil {
		t.Fatalf("ListLocations() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != id {
		t.Errorf("expected the single cached location %s, got %+v", id, items)
	}
}

// TestListLocations_FetchesAndCachesOnFirstCall verifies that the first call
// fetches locations over HTTP and a second call is served from cache, leaving
// exactly one request recorded.
//
// Why it matters: the export should pay a single bulk list cost and then reuse
// the results, avoiding repeated round-trips to Nautobot.
// Inputs: a server returning one location, "Site-A". Outputs: the parsed items
// and a call counter that must equal 1 after two invocations.
// Data choice: the call counter is the assertion vehicle; "Site-A" is an
// arbitrary location name parsed from the list payload.
func TestListLocations_FetchesAndCachesOnFirstCall(t *testing.T) {
	id := uuid.New()
	var calls int
	body := `{"count":1,"results":[{"id":"` + id.String() + `","name":"Site-A","display":"Site-A"}]}`
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	items, err := e.Cache.ListLocations()
	if err != nil {
		t.Fatalf("ListLocations() error = %v", err)
	}
	if len(items) != 1 || items[0].Name != "Site-A" {
		t.Fatalf("expected one location named Site-A, got %+v", items)
	}

	// A second call must be served from cache, issuing no further HTTP request.
	if _, err := e.Cache.ListLocations(); err != nil {
		t.Fatalf("second ListLocations() error = %v", err)
	}
	if calls != 1 {
		t.Errorf("expected exactly 1 HTTP call, got %d", calls)
	}
}

// TestListLocations_ReturnsErrorOnNon200 verifies that ListLocations returns an
// error when the API responds with a non-200 status.
//
// Why it matters: a failed location list must abort the export rather than
// proceed with an empty or partial set of placements.
// Inputs: a server that always answers 500. Outputs: a non-nil error.
// Data choice: 500 is the generic server-failure case for the list endpoint.
func TestListLocations_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{}`))
	defer cleanup()

	if _, err := e.Cache.ListLocations(); err == nil {
		t.Fatal("expected an error when the API responds with 500")
	}
}

// TestListStatuses_FetchesAndCaches verifies that ListStatuses fetches statuses
// over HTTP and returns them parsed into CachedItems.
//
// Why it matters: nearly every exported object references a status, so the
// status list must load correctly before devices and IPAM objects are written.
// Inputs: a server returning one status, "Active". Outputs: a slice containing
// that status.
// Data choice: "Active" is Nautobot's stock default status, the one most
// exports will actually resolve against.
func TestListStatuses_FetchesAndCaches(t *testing.T) {
	id := uuid.New()
	var calls int
	body := `{"count":1,"results":[{"id":"` + id.String() + `","name":"Active","display":"Active"}]}`
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	items, err := e.Cache.ListStatuses()
	if err != nil {
		t.Fatalf("ListStatuses() error = %v", err)
	}
	if len(items) != 1 || items[0].Name != "Active" {
		t.Errorf("expected one status named Active, got %+v", items)
	}
}

// TestListStatuses_ReturnsErrorOnNon200 verifies that ListStatuses returns an
// error when the API responds with a non-200 status.
//
// Why it matters: without a usable status list the export cannot assign
// lifecycle states, so the failure must propagate.
// Inputs: a server that always answers 502. Outputs: a non-nil error.
// Data choice: 502 Bad Gateway models an upstream/proxy failure, varying the
// error code across the list tests.
func TestListStatuses_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusBadGateway, `{}`))
	defer cleanup()

	if _, err := e.Cache.ListStatuses(); err == nil {
		t.Fatal("expected an error when the API responds with 502")
	}
}

// TestListStatuses_ReturnsCachedItemsWithoutHTTP verifies that ListStatuses
// returns cached statuses and makes no HTTP request once the cache is loaded.
//
// Why it matters: repeated status resolution during an export must hit memory,
// not the API, to stay efficient.
// Inputs: a cache pre-seeded with "Active" and statusesLoaded=true. Outputs:
// the single cached *CachedItem.
// Data choice: a nil client guarantees a panic on any HTTP, proving the loaded
// short-circuit fired.
func TestListStatuses_ReturnsCachedItemsWithoutHTTP(t *testing.T) {
	cache := NewLookupCache(nil)
	id := uuid.New()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: id, Name: "Active"}
	cache.statusesLoaded = true
	cache.statusesMu.Unlock()

	items, err := cache.ListStatuses()
	if err != nil {
		t.Fatalf("ListStatuses() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != id {
		t.Errorf("expected the single cached status %s, got %+v", id, items)
	}
}

// TestListRoles_FetchesAndCaches verifies that ListRoles fetches roles over HTTP
// and returns them parsed into CachedItems.
//
// Why it matters: device export assigns a role to each device, so the role list
// must load before devices are written.
// Inputs: a server returning one role, "Compute". Outputs: a slice containing
// that role.
// Data choice: "Compute" is a representative device role exercised by the
// success path of the list call.
func TestListRoles_FetchesAndCaches(t *testing.T) {
	id := uuid.New()
	var calls int
	body := `{"count":1,"results":[{"id":"` + id.String() + `","name":"Compute","display":"Compute"}]}`
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	items, err := e.Cache.ListRoles()
	if err != nil {
		t.Fatalf("ListRoles() error = %v", err)
	}
	if len(items) != 1 || items[0].Name != "Compute" {
		t.Errorf("expected one role named Compute, got %+v", items)
	}
}

// TestListRoles_ReturnsCachedItemsWithoutHTTP verifies that ListRoles returns
// cached roles and makes no HTTP request once the cache is loaded.
//
// Why it matters: role resolution recurs across devices and must be served from
// memory to avoid redundant calls to Nautobot.
// Inputs: a cache pre-seeded with "Compute" and rolesLoaded=true. Outputs: the
// single cached *CachedItem.
// Data choice: a nil client makes any HTTP attempt panic, proving the cached
// branch was used.
func TestListRoles_ReturnsCachedItemsWithoutHTTP(t *testing.T) {
	cache := NewLookupCache(nil)
	id := uuid.New()
	cache.rolesMu.Lock()
	cache.roles["Compute"] = &CachedItem{ID: id, Name: "Compute"}
	cache.rolesLoaded = true
	cache.rolesMu.Unlock()

	items, err := cache.ListRoles()
	if err != nil {
		t.Fatalf("ListRoles() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != id {
		t.Errorf("expected the single cached role %s, got %+v", id, items)
	}
}

// TestListDeviceTypes_FetchesAndCachesByModel verifies that ListDeviceTypes
// fetches device types and stores each one keyed by its model, with the model
// copied into both Name and Slug.
//
// Why it matters: device types are looked up by model/slug when exporting
// devices; the Name==Slug==model contract is what callers depend on to resolve
// them.
// Inputs: a server returning one device type whose "model" is "DL380".
// Outputs: a CachedItem with Name and Slug both set to "DL380".
// Data choice: the payload uses the "model" field (not "name") to confirm the
// mapping, and "DL380" is a realistic server model.
func TestListDeviceTypes_FetchesAndCachesByModel(t *testing.T) {
	id := uuid.New()
	var calls int
	body := `{"count":1,"results":[{"id":"` + id.String() + `","model":"DL380","display":"DL380"}]}`
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	items, err := e.Cache.ListDeviceTypes()
	if err != nil {
		t.Fatalf("ListDeviceTypes() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one device type, got %d", len(items))
	}
	// The model is used as both the cache key and the item's Name/Slug.
	if items[0].Name != "DL380" || items[0].Slug != "DL380" {
		t.Errorf("device type Name/Slug = %q/%q, want DL380/DL380", items[0].Name, items[0].Slug)
	}
}

// TestListDeviceTypes_ReturnsErrorOnNon200 verifies that ListDeviceTypes returns
// an error when the API responds with a non-200 status.
//
// Why it matters: a failed device-type list must stop the export rather than
// leave devices unable to resolve their type.
// Inputs: a server that always answers 403. Outputs: a non-nil error.
// Data choice: 403 Forbidden models a permission/auth failure on the
// device-types endpoint.
func TestListDeviceTypes_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusForbidden, `{}`))
	defer cleanup()

	if _, err := e.Cache.ListDeviceTypes(); err == nil {
		t.Fatal("expected an error when the API responds with 403")
	}
}

// TestListDeviceTypes_ReturnsCachedItemsWithoutHTTP verifies that ListDeviceTypes
// returns cached device types and makes no HTTP request once the cache is loaded.
//
// Why it matters: device-type resolution repeats for every device, so it must
// be served from memory to keep large exports efficient.
// Inputs: a cache pre-seeded with "DL380" and deviceTypesLoaded=true. Outputs:
// the single cached *CachedItem.
// Data choice: a nil client guarantees a panic on any HTTP, proving the loaded
// short-circuit fired.
func TestListDeviceTypes_ReturnsCachedItemsWithoutHTTP(t *testing.T) {
	cache := NewLookupCache(nil)
	id := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["DL380"] = &CachedItem{ID: id, Name: "DL380", Slug: "DL380"}
	cache.deviceTypesLoaded = true
	cache.deviceTypesMu.Unlock()

	items, err := cache.ListDeviceTypes()
	if err != nil {
		t.Fatalf("ListDeviceTypes() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != id {
		t.Errorf("expected the single cached device type %s, got %+v", id, items)
	}
}

// -----------------------------------------------------------------------------
// GetAllDevicesByName — unlike GetDeviceByName, returns every match uncached.
// -----------------------------------------------------------------------------

// TestGetAllDevicesByName_ReturnsEveryMatch verifies that GetAllDevicesByName
// returns every device matching a name, not just the first.
//
// Why it matters: unlike the cached GetDeviceByName, this is how the export
// detects and disambiguates duplicate same-named devices in Nautobot before
// deciding what to update.
// Inputs: name "dup". Outputs: a slice of both matching CachedItems and one
// HTTP request.
// Data choice: two results share the name "dup" but differ in display ("#1"
// vs "#2"), so ID/display assertions prove no match was dropped or rewritten.
func TestGetAllDevicesByName_ReturnsEveryMatch(t *testing.T) {
	id1, id2 := uuid.New(), uuid.New()
	var calls int
	body := `{"count":2,"results":[` +
		`{"id":"` + id1.String() + `","name":"dup","display":"dup #1"},` +
		`{"id":"` + id2.String() + `","name":"dup","display":"dup #2"}]}`
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	items, err := e.Cache.GetAllDevicesByName("dup")
	if err != nil {
		t.Fatalf("GetAllDevicesByName() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 matching devices, got %d", len(items))
	}
	if items[0].ID != id1 || items[0].Name != "dup" || items[0].Display != "dup #1" {
		t.Errorf("items[0] = %+v, want id=%s name=dup display='dup #1'", items[0], id1)
	}
	if items[1].ID != id2 || items[1].Name != "dup" || items[1].Display != "dup #2" {
		t.Errorf("items[1] = %+v, want id=%s name=dup display='dup #2'", items[1], id2)
	}
	if calls != 1 {
		t.Errorf("expected exactly one device-list request, got %d", calls)
	}
}

// TestGetAllDevicesByName_ReturnsErrorOnNon200 verifies that GetAllDevicesByName
// returns an error when the API responds with a non-200 status.
//
// Why it matters: the duplicate-detection step must fail loudly rather than
// report zero matches and risk creating a duplicate device.
// Inputs: name "anything". Outputs: a non-nil error.
// Data choice: 500 is the generic server-failure case for the device list
// query.
func TestGetAllDevicesByName_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{}`))
	defer cleanup()

	if _, err := e.Cache.GetAllDevicesByName("anything"); err == nil {
		t.Fatal("expected an error when the API responds with 500")
	}
}
