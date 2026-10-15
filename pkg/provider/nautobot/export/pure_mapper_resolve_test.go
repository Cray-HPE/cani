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
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// ---------- resolveDeviceType (cache hit) ----------

// TestResolveDeviceTypeCacheHit verifies that resolveDeviceType returns the
// cached device-type item when the device's Slug matches a cache entry.
//
// Why it matters: device-type resolution is the most frequent lookup during
// export; a cache hit must return the right Nautobot ID without any API call.
// Inputs: a device with Slug "dl360" and a cache holding that slug.
// Outputs: the cached item whose ID equals the seeded UUID.
// Data choice: pre-seeding the slug isolates the cache-hit branch from the
// model-fallback and create-on-miss branches.
func TestResolveDeviceTypeCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["dl360"] = &CachedItem{ID: dtID, Name: "DL360"}
	cache.deviceTypesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{Slug: "dl360"}

	item, err := mapper.resolveDeviceType(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != dtID {
		t.Errorf("expected ID %s, got %s", dtID, item.ID)
	}
}

// TestResolveDeviceTypeNoSlug verifies that resolveDeviceType returns an error
// when both Slug and Model are empty.
//
// Why it matters: a device with no type cannot be created in Nautobot; failing
// fast lets the loader skip or report it instead of sending an invalid request.
// Inputs: a device with empty Slug and Model under non-strict opts.
// Outputs: a non-nil error.
// Data choice: empty slug AND model is the only way to reach the type-required
// guard; non-strict still errors here (it returns the ErrDeviceUnclassified
// sentinel so callers may choose to skip).
func TestResolveDeviceTypeNoSlug(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{Strict: false})
	dev := &devicetypes.CaniDeviceType{Slug: "", Model: ""}

	_, err := mapper.resolveDeviceType(dev)
	if err == nil {
		t.Error("expected error for empty slug/model")
	}
}

// TestResolveDeviceTypeFallbackModel verifies that resolveDeviceType falls back
// to the device's Model when Slug is empty and resolves it from cache.
//
// Why it matters: some cani devices carry only a Model string; honoring it as
// the type key lets those devices export instead of being rejected.
// Inputs: a device with Slug="" and Model="DL380" and a cache keyed by "DL380".
// Outputs: the cached item with the seeded UUID.
// Data choice: caching under the model string (not a slug) proves the fallback
// uses Model verbatim as the lookup key.
func TestResolveDeviceTypeFallbackModel(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["DL380"] = &CachedItem{ID: dtID, Name: "DL380"}
	cache.deviceTypesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{Slug: "", Model: "DL380"}

	item, err := mapper.resolveDeviceType(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != dtID {
		t.Errorf("expected ID %s, got %s", dtID, item.ID)
	}
}

// ---------- resolveStatus (cache hit) ----------

// TestResolveStatusCacheHit verifies that resolveStatus resolves the mapper's
// DefaultStatus from cache when the device specifies no status of its own.
//
// Why it matters: most devices inherit the export-wide default status, so the
// default must map to the correct Nautobot status ID.
// Inputs: a device with no Status and a mapper with DefaultStatus "Active",
// cache seeded with "Active". Outputs: the cached "Active" item.
// Data choice: an empty device status forces the default-status branch.
func TestResolveStatusCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	statusID := uuid.New()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultStatus: "Active"})
	dev := &devicetypes.CaniDeviceType{}

	item, err := mapper.resolveStatus(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != statusID {
		t.Errorf("expected ID %s, got %s", statusID, item.ID)
	}
}

// TestResolveStatusExplicit verifies that resolveStatus prefers the device's
// own Status field over the mapper's DefaultStatus.
//
// Why it matters: per-device status (e.g. "Planned") must win so devices export
// with their real lifecycle state rather than the global default.
// Inputs: a device with Status "Planned" while DefaultStatus is "Active"; cache
// holds "Planned". Outputs: the cached "Planned" item.
// Data choice: a default that differs from the explicit value proves the
// override path is taken, not the default.
func TestResolveStatusExplicit(t *testing.T) {
	cache := NewLookupCache(nil)
	statusID := uuid.New()
	cache.statusesMu.Lock()
	cache.statuses["Planned"] = &CachedItem{ID: statusID, Name: "Planned"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultStatus: "Active"})
	dev := &devicetypes.CaniDeviceType{ObjectMeta: devicetypes.ObjectMeta{Status: "Planned"}}

	item, err := mapper.resolveStatus(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != statusID {
		t.Errorf("expected ID %s, got %s", statusID, item.ID)
	}
}

// ---------- resolveRole (cache hit) ----------

// TestResolveRoleCacheHit verifies that resolveRole resolves the mapper's
// DefaultRole from cache when the device specifies no role.
//
// Why it matters: devices without an explicit role inherit the export default,
// which must map to the right Nautobot role ID.
// Inputs: a device with no Role and DefaultRole "Server"; cache seeded with
// "Server". Outputs: the cached "Server" item.
// Data choice: an empty device role isolates the default-role branch.
func TestResolveRoleCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	roleID := uuid.New()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultRole: "Server"})
	dev := &devicetypes.CaniDeviceType{}

	item, err := mapper.resolveRole(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != roleID {
		t.Errorf("expected ID %s, got %s", roleID, item.ID)
	}
}

// TestResolveRoleExplicit verifies that resolveRole prefers the device's
// explicit Role field over the mapper's DefaultRole.
//
// Why it matters: device-specific roles must take precedence so exported
// devices get their true role, not the fallback default.
// Inputs: a device with Role "Compute" while DefaultRole is "Server"; cache
// holds "Compute". Outputs: the cached "Compute" item.
// Data choice: a differing default confirms the explicit field wins.
func TestResolveRoleExplicit(t *testing.T) {
	cache := NewLookupCache(nil)
	roleID := uuid.New()
	cache.rolesMu.Lock()
	cache.roles["Compute"] = &CachedItem{ID: roleID, Name: "Compute"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultRole: "Server"})
	dev := &devicetypes.CaniDeviceType{ObjectMeta: devicetypes.ObjectMeta{Role: "Compute"}}

	item, err := mapper.resolveRole(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != roleID {
		t.Errorf("expected ID %s, got %s", roleID, item.ID)
	}
}

// TestResolveRoleFromMetadata verifies that resolveRole falls back to
// ProviderMetadata["role"] when the explicit Role field is empty.
//
// Why it matters: legacy/provider-sourced records carry the role in metadata;
// honoring it preserves role assignments during export without a default.
// Inputs: a device with no Role but ProviderMetadata{"role":"Gateway"} and no
// DefaultRole; cache holds "Gateway". Outputs: the cached "Gateway" item.
// Data choice: omitting DefaultRole forces resolution via metadata alone.
func TestResolveRoleFromMetadata(t *testing.T) {
	cache := NewLookupCache(nil)
	roleID := uuid.New()
	cache.rolesMu.Lock()
	cache.roles["Gateway"] = &CachedItem{ID: roleID, Name: "Gateway"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{
		ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{"role": "Gateway"}},
	}

	item, err := mapper.resolveRole(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != roleID {
		t.Errorf("expected ID %s, got %s", roleID, item.ID)
	}
}

// ---------- resolveLocation (cache hit) ----------

// TestResolveLocationDefault verifies that resolveLocation resolves the mapper's
// DefaultLocation from cache when the device has no location metadata and no
// parent rack.
//
// Why it matters: devices without explicit placement fall back to the export
// default location, which must map to a valid Nautobot location ID.
// Inputs: a bare device with DefaultLocation "Site-A"; cache holds "Site-A".
// Outputs: the cached "Site-A" item.
// Data choice: an empty device isolates the default-location branch from the
// metadata and parent-rack branches.
func TestResolveLocationDefault(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	cache.locationsMu.Lock()
	cache.locations["Site-A"] = &CachedItem{ID: locID, Name: "Site-A"}
	cache.locationsMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultLocation: "Site-A"})
	dev := &devicetypes.CaniDeviceType{}

	item, err := mapper.resolveLocation(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != locID {
		t.Errorf("expected ID %s, got %s", locID, item.ID)
	}
}

// TestResolveLocationFromMetadata verifies that resolveLocation prefers
// ProviderMetadata["location"] over the mapper's DefaultLocation.
//
// Why it matters: per-device location metadata must win so devices land in the
// correct Nautobot location instead of the global default.
// Inputs: a device with ProviderMetadata{"location":"Room-B"} while
// DefaultLocation is "Site-A"; cache holds "Room-B". Outputs: the "Room-B" item.
// Data choice: a differing default proves the metadata path takes priority.
func TestResolveLocationFromMetadata(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	cache.locationsMu.Lock()
	cache.locations["Room-B"] = &CachedItem{ID: locID, Name: "Room-B"}
	cache.locationsMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultLocation: "Site-A"})
	dev := &devicetypes.CaniDeviceType{
		ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{"location": "Room-B"}},
	}

	item, err := mapper.resolveLocation(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != locID {
		t.Errorf("expected ID %s, got %s", locID, item.ID)
	}
}

// ---------- resolveLocalDeviceType / resolveLocalLocation / resolveLocalStatus / resolveLocalRole ----------

// TestResolveLocalDeviceTypeHit verifies that resolveLocalDeviceType returns the
// cached CachedItem (matching ID) when the device's slug resolves in the cache.
//
// Why it matters: the diff engine resolves the local device type to compare it
// against Nautobot; a cache hit must yield the right item so comparisons use the
// correct UUID.
// Inputs: a CaniDeviceType with Slug "dt-x" and a mapper whose cache holds that
// slug. Outputs: a non-nil *CachedItem whose ID equals the cached UUID.
// Data choice: pre-seeding the device-type cache under the matching slug isolates
// the happy-path resolution from any API/error handling.
func TestResolveLocalDeviceTypeHit(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["dt-x"] = &CachedItem{ID: dtID, Name: "DT-X"}
	cache.deviceTypesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{Slug: "dt-x"}
	item := resolveLocalDeviceType(dev, mapper)
	if item == nil || item.ID != dtID {
		t.Errorf("expected cached item with ID %s", dtID)
	}
}

// TestResolveLocalDeviceTypeMiss verifies that resolveLocalDeviceType returns nil
// when the device has no slug/model and Strict is false (the underlying
// ErrDeviceUnclassified is swallowed into a nil result).
//
// Why it matters: in the diff path an unclassifiable device must not produce a
// device_type diff; returning nil makes compareDeviceFields skip that field.
// Inputs: a CaniDeviceType with empty Slug and Model and a non-strict mapper.
// Outputs: a nil *CachedItem.
// Data choice: empty slug+model with Strict=false is the exact condition that
// yields ErrDeviceUnclassified, which the local resolver converts to nil.
func TestResolveLocalDeviceTypeMiss(t *testing.T) {
	// When the slug is empty and Strict is false, resolveDeviceType returns
	// ErrDeviceUnclassified without hitting the API.
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{Strict: false})
	dev := &devicetypes.CaniDeviceType{Slug: "", Model: ""}
	item := resolveLocalDeviceType(dev, mapper)
	if item != nil {
		t.Error("expected nil for unclassified device (no slug/model)")
	}
}

// TestResolveLocalLocationHit verifies that resolveLocalLocation returns the
// cached CachedItem (matching ID) when the mapper's DefaultLocation resolves in
// the cache.
//
// Why it matters: the diff engine needs the local location's UUID to compare
// against Nautobot; a cache hit must return the correct item.
// Inputs: an empty CaniDeviceType and a mapper with DefaultLocation "Room-RL"
// pre-cached. Outputs: a non-nil *CachedItem whose ID equals the cached UUID.
// Data choice: seeding the location cache under the default name exercises the
// straightforward hit path used during diffing.
func TestResolveLocalLocationHit(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	cache.locationsMu.Lock()
	cache.locations["Room-RL"] = &CachedItem{ID: locID, Name: "Room-RL"}
	cache.locationsMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultLocation: "Room-RL"})
	dev := &devicetypes.CaniDeviceType{}
	item := resolveLocalLocation(dev, mapper)
	if item == nil || item.ID != locID {
		t.Errorf("expected cached item with ID %s", locID)
	}
}

// TestResolveLocalStatusHit verifies that resolveLocalStatus returns the cached
// CachedItem (matching ID) when the mapper's DefaultStatus resolves in the cache.
//
// Why it matters: status comparison during export diffing depends on resolving
// the local status to its Nautobot UUID; a hit must return the right item.
// Inputs: an empty CaniDeviceType and a mapper with DefaultStatus "Planned"
// pre-cached. Outputs: a non-nil *CachedItem whose ID equals the cached UUID.
// Data choice: pre-seeding the status cache under the default name isolates the
// successful resolution from list/create fallbacks.
func TestResolveLocalStatusHit(t *testing.T) {
	cache := NewLookupCache(nil)
	statusID := uuid.New()
	cache.statusesMu.Lock()
	cache.statuses["Planned"] = &CachedItem{ID: statusID, Name: "Planned"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultStatus: "Planned"})
	dev := &devicetypes.CaniDeviceType{}
	item := resolveLocalStatus(dev, mapper)
	if item == nil || item.ID != statusID {
		t.Errorf("expected cached item with ID %s", statusID)
	}
}

// TestResolveLocalRoleHit verifies that resolveLocalRole returns the cached
// CachedItem (matching ID) when the mapper's DefaultRole resolves in the cache.
//
// Why it matters: role drift detection during export relies on resolving the
// local role to its Nautobot UUID; a cache hit must return the correct item.
// Inputs: an empty CaniDeviceType and a mapper with DefaultRole "Gateway"
// pre-cached. Outputs: a non-nil *CachedItem whose ID equals the cached UUID.
// Data choice: seeding the role cache under the default name exercises the hit
// path without invoking any API lookup.
func TestResolveLocalRoleHit(t *testing.T) {
	cache := NewLookupCache(nil)
	roleID := uuid.New()
	cache.rolesMu.Lock()
	cache.roles["Gateway"] = &CachedItem{ID: roleID, Name: "Gateway"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultRole: "Gateway"})
	dev := &devicetypes.CaniDeviceType{}
	item := resolveLocalRole(dev, mapper)
	if item == nil || item.ID != roleID {
		t.Errorf("expected cached item with ID %s", roleID)
	}
}
