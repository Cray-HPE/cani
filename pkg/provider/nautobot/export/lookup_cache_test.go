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
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// ---------- LookupCache setters ----------

// TestLookupCacheSetters verifies each LookupCache create-toggle setter and
// SetContext mutate the corresponding internal field.
//
// Why it matters: these flags decide whether export auto-creates missing
// device types, statuses, roles, locations, and location types in Nautobot;
// a broken setter would silently disable on-demand creation.
// Inputs: a fresh cache; each setter called with true (context set explicitly).
// Outputs: the matching private fields become true / hold the context.
// Data choice: every setter is exercised; SetCreateModuleTypes is called only
// to confirm the documented no-op does not panic.
func TestLookupCacheSetters(t *testing.T) {
	cache := NewLookupCache(nil)

	ctx := context.Background()
	cache.SetContext(ctx)
	if cache.ctx != ctx {
		t.Error("SetContext did not set context")
	}

	cache.SetCreateDeviceTypes(true)
	if !cache.createDeviceTypes {
		t.Error("SetCreateDeviceTypes(true) did not take effect")
	}

	cache.SetCreateStatuses(true)
	if !cache.createStatuses {
		t.Error("SetCreateStatuses(true) did not take effect")
	}

	cache.SetCreateRoles(true)
	if !cache.createRoles {
		t.Error("SetCreateRoles(true) did not take effect")
	}

	cache.SetCreateLocations(true)
	if !cache.createLocations {
		t.Error("SetCreateLocations(true) did not take effect")
	}

	cache.SetCreateLocationTypes(true)
	if !cache.createLocationTypes {
		t.Error("SetCreateLocationTypes(true) did not take effect")
	}

	// SetCreateModuleTypes is a no-op (gated at Exporter.Options level),
	// just verify it doesn't panic.
	cache.SetCreateModuleTypes(true)
}

// ---------- SetCreateModuleTypes ----------

// TestSetCreateModuleTypes verifies that LookupCache.SetCreateModuleTypes can be
// called with true and false without panicking.
//
// Why it matters: the setter is part of the cache's configuration surface used
// during export setup; even as a no-op it must remain safe to call.
// Inputs: boolean true then false. Outputs: none (the method is a no-op).
// Data choice: calling both boolean values is the minimal smoke check; the test
// asserts only the absence of a panic, not any observable state change.
func TestSetCreateModuleTypes(t *testing.T) {
	cache := NewLookupCache(nil)
	// Should not panic — it's a no-op.
	cache.SetCreateModuleTypes(true)
	cache.SetCreateModuleTypes(false)
}

// ---------- GetDeviceType / GetLocation / GetStatus / GetRole cache hit ----------

// TestGetDeviceTypeCacheHit verifies GetDeviceType returns the cached
// CachedItem (matching ID) when the device-type slug is already cached.
//
// Why it matters: device-type resolution runs for every exported device; a
// cache hit must skip the Nautobot API and return the correct UUID so devices
// reference the right type.
// Inputs: a LookupCache seeded with slug "cached-dt"->dtID. Outputs: the cached
// item with no error.
// Data choice: a fresh UUID under a known slug isolates the cache-hit branch
// without needing an HTTP client.
func TestGetDeviceTypeCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["cached-dt"] = &CachedItem{ID: dtID, Name: "Cached DT"}
	cache.deviceTypesMu.Unlock()

	item, err := cache.GetDeviceType("cached-dt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != dtID {
		t.Errorf("ID = %s, want %s", item.ID, dtID)
	}
}

// TestGetLocationCacheHit verifies GetLocation returns the cached CachedItem
// (matching ID) when the location name is already cached.
//
// Why it matters: locations anchor every exported device and rack; a cache hit
// must avoid an API round-trip and yield the correct location UUID.
// Inputs: a cache seeded with "cached-loc"->locID. Outputs: the cached item
// with no error.
// Data choice: a known name/UUID with a nil client proves the cache-hit branch
// is taken without HTTP.
func TestGetLocationCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	cache.locationsMu.Lock()
	cache.locations["cached-loc"] = &CachedItem{ID: locID, Name: "cached-loc"}
	cache.locationsMu.Unlock()

	item, err := cache.GetLocation("cached-loc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != locID {
		t.Errorf("ID = %s, want %s", item.ID, locID)
	}
}

// TestGetStatusCacheHit verifies GetStatus returns the cached CachedItem
// (matching ID) when the status name is already cached.
//
// Why it matters: every exported device needs a status reference; a cache hit
// must avoid an API call and return the correct status UUID.
// Inputs: a cache seeded with "cached-status"->statusID. Outputs: the cached
// item with no error.
// Data choice: a known name/UUID with a nil client isolates the cache-hit path.
func TestGetStatusCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	statusID := uuid.New()
	cache.statusesMu.Lock()
	cache.statuses["cached-status"] = &CachedItem{ID: statusID, Name: "cached-status"}
	cache.statusesMu.Unlock()

	item, err := cache.GetStatus("cached-status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != statusID {
		t.Errorf("ID = %s, want %s", item.ID, statusID)
	}
}

// TestGetRoleCacheHit verifies GetRole returns the cached CachedItem (matching
// ID) when the role name is already cached.
//
// Why it matters: device roles are resolved during export; a cache hit must
// return the right role UUID without a Nautobot round-trip.
// Inputs: a cache seeded with "cached-role"->roleID. Outputs: the cached item
// with no error.
// Data choice: a known name/UUID with a nil client isolates the cache-hit path.
func TestGetRoleCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	roleID := uuid.New()
	cache.rolesMu.Lock()
	cache.roles["cached-role"] = &CachedItem{ID: roleID, Name: "cached-role"}
	cache.rolesMu.Unlock()

	item, err := cache.GetRole("cached-role")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != roleID {
		t.Errorf("ID = %s, want %s", item.ID, roleID)
	}
}

// ---------- FindNameByID cache searches ----------

// TestFindNameByIDDeviceType verifies FindNameByID resolves a deviceType UUID
// back to its human-readable Name via the device-type cache.
//
// Why it matters: reverse ID->name lookups produce readable diff/summary output
// when reporting what an export changed.
// Inputs: cacheType "deviceType" and a cached id->"My Device Type". Outputs:
// the matching Name string.
// Data choice: a distinct Name proves the lookup returns Name, not the slug key
// or the UUID fallback.
func TestFindNameByIDDeviceType(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["my-dt"] = &CachedItem{ID: dtID, Name: "My Device Type"}
	cache.deviceTypesMu.Unlock()

	got := cache.FindNameByID("deviceType", dtID)
	if got != "My Device Type" {
		t.Errorf("got %q, want 'My Device Type'", got)
	}
}

// TestFindNameByIDLocation verifies FindNameByID resolves a location UUID to
// its Name (preferred over Display) via the location cache.
//
// Why it matters: location names appear in export summaries/diffs, so the
// reverse lookup must prefer the canonical Name field.
// Inputs: cacheType "location" and a cached id with Name "Room-X" and a
// different Display. Outputs: "Room-X".
// Data choice: setting Display differently from Name guards that Name wins when
// both are populated.
func TestFindNameByIDLocation(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	cache.locationsMu.Lock()
	cache.locations["room"] = &CachedItem{ID: locID, Name: "Room-X", Display: "Room-X Display"}
	cache.locationsMu.Unlock()

	got := cache.FindNameByID("location", locID)
	if got != "Room-X" {
		t.Errorf("got %q, want 'Room-X'", got)
	}
}

// TestFindNameByIDStatus verifies FindNameByID resolves a status UUID back to
// its Name via the status cache.
//
// Why it matters: readable status labels in export diffs/summaries depend on
// this reverse lookup.
// Inputs: cacheType "status" and a cached id->"Active". Outputs: "Active".
// Data choice: a canonical status name confirms the status branch of the
// cache-type switch.
func TestFindNameByIDStatus(t *testing.T) {
	cache := NewLookupCache(nil)
	statusID := uuid.New()
	cache.statusesMu.Lock()
	cache.statuses["active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	got := cache.FindNameByID("status", statusID)
	if got != "Active" {
		t.Errorf("got %q, want 'Active'", got)
	}
}

// TestFindNameByIDRole verifies FindNameByID resolves a role UUID back to its
// Name via the role cache.
//
// Why it matters: readable role labels in export reporting depend on this
// reverse lookup.
// Inputs: cacheType "role" and a cached id->"Server". Outputs: "Server".
// Data choice: a canonical role name confirms the role branch of the cache-type
// switch.
func TestFindNameByIDRole(t *testing.T) {
	cache := NewLookupCache(nil)
	roleID := uuid.New()
	cache.rolesMu.Lock()
	cache.roles["svr"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	got := cache.FindNameByID("role", roleID)
	if got != "Server" {
		t.Errorf("got %q, want 'Server'", got)
	}
}

// TestFindNameByIDNotFound verifies FindNameByID falls back to the UUID string
// when the id is absent from the searched cache.
//
// Why it matters: export reporting must never emit an empty label; an
// unresolved reference should still show a stable identifier.
// Inputs: cacheType "deviceType" and a random, uncached UUID. Outputs: a
// non-empty string (the UUID itself).
// Data choice: a fresh UUID guarantees a miss so the UUID-string fallback
// branch is exercised.
func TestFindNameByIDNotFound(t *testing.T) {
	cache := NewLookupCache(nil)
	got := cache.FindNameByID("deviceType", uuid.New())
	// Should return the UUID string when not found in any cache
	if got == "" {
		t.Error("expected non-empty UUID string fallback")
	}
}

// ---------- GetDeviceByName cache hit ----------

// TestGetDeviceByNameCacheHit verifies GetDeviceByName returns the cached
// device CachedItem (matching ID) on a cache hit.
//
// Why it matters: device-by-name resolution decides create-vs-update during
// export; a cache hit must return the correct existing device without an API
// call.
// Inputs: a cache seeded with "my-server"->devID. Outputs: a non-nil item whose
// ID equals devID.
// Data choice: a known name/UUID pair with a nil client proves the cache path
// is taken (an API call would panic).
func TestGetDeviceByNameCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	devID := uuid.New()
	cache.devicesMu.Lock()
	cache.devices["my-server"] = &CachedItem{ID: devID, Name: "my-server", Display: "My Server"}
	cache.devicesMu.Unlock()

	item, err := cache.GetDeviceByName("my-server")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item == nil {
		t.Fatal("expected non-nil item")
	}
	if item.ID != devID {
		t.Errorf("ID = %s, want %s", item.ID, devID)
	}
}

// TestGetDeviceByNameCacheMissReturnsNilFromAPI verifies a cache miss calls the
// Nautobot device list endpoint and returns (nil, nil) when no device matches.
//
// Why it matters: device-by-name lookup decides whether export creates a new
// Nautobot device or updates an existing one; a clean miss must be distinguishable
// from an API error.
// Inputs: name "missing-server" against an empty device-list response. Outputs:
// nil item, nil error, and exactly one HTTP request.
// Data choice: an empty Nautobot list response is the smallest valid not-found
// payload and keeps the test FAST without relying on a live server.
func TestGetDeviceByNameCacheMissReturnsNilFromAPI(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, emptyListJSON))
	defer cleanup()

	item, err := e.Cache.GetDeviceByName("missing-server")
	if err != nil {
		t.Fatalf("GetDeviceByName() error = %v", err)
	}
	if item != nil {
		t.Fatalf("expected nil item for missing device, got %+v", item)
	}
	if calls != 1 {
		t.Errorf("expected exactly one HTTP request for cache miss, got %d", calls)
	}
}

// ---------- LookupLocation cache hit ----------

// TestLookupLocationCacheHit verifies LookupLocation returns the cached
// location CachedItem (matching ID) on a cache hit.
//
// Why it matters: locations anchor every exported device/rack; resolving a
// cached location must avoid an API round-trip and yield the right UUID.
// Inputs: a cache seeded with "cached-room"->locID. Outputs: a non-nil item
// whose ID equals locID.
// Data choice: a known name/UUID with a nil client confirms the cache branch (a
// miss would require the client).
func TestLookupLocationCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	cache.locationsMu.Lock()
	cache.locations["cached-room"] = &CachedItem{ID: locID, Name: "cached-room", Display: "Cached Room"}
	cache.locationsMu.Unlock()

	item, err := cache.LookupLocation("cached-room")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item == nil {
		t.Fatal("expected non-nil item")
	}
	if item.ID != locID {
		t.Errorf("ID = %s, want %s", item.ID, locID)
	}
}
