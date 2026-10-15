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

	"github.com/google/uuid"
)

// ---------- CacheLocation / CacheInterface ----------

// TestCacheLocationAndInterface verifies CacheLocation and CacheInterface store
// items under the expected keys (interface keyed by device+name).
//
// Why it matters: export caches resolved Nautobot IDs to avoid re-querying the
// API per device; a wrong key means cache misses, duplicate lookups, or wrong
// IDs attached to interfaces.
// Inputs: a location cached by name and an interface cached by (deviceID,"eth0").
// Outputs: both items retrievable from the private maps with matching IDs.
// Data choice: reading the private maps under their mutexes confirms the real
// storage key (via interfaceCacheKey) rather than just the public API.
func TestCacheLocationAndInterface(t *testing.T) {
	cache := NewLookupCache(nil)

	locID := uuid.New()
	cache.CacheLocation("test-loc", &CachedItem{ID: locID, Name: "test-loc"})

	// Verify it was cached.
	cache.locationsMu.RLock()
	item, ok := cache.locations["test-loc"]
	cache.locationsMu.RUnlock()
	if !ok || item.ID != locID {
		t.Error("CacheLocation did not cache correctly")
	}

	// Cache an interface.
	ifaceID := uuid.New()
	deviceID := uuid.New()
	cache.CacheInterface(deviceID, "eth0", &CachedItem{ID: ifaceID, Name: "eth0"})

	cache.interfacesMu.RLock()
	key := interfaceCacheKey(deviceID, "eth0")
	iitem, ok := cache.interfaces[key]
	cache.interfacesMu.RUnlock()
	if !ok || iitem.ID != ifaceID {
		t.Error("CacheInterface did not cache correctly")
	}
}

// ---------- InvalidateInterfacePrefetch ----------

// TestInvalidateInterfacePrefetch verifies the per-device prefetch flag is
// removed so the next interface lookup re-fetches from Nautobot.
//
// Why it matters: after creating/changing a device's interfaces, the cached
// "already prefetched" marker is stale; failing to clear it would hide the new
// interfaces from subsequent export steps.
// Inputs: a cache with one device marked prefetched.
// Outputs: that device's prefetch entry is deleted.
// Data choice: seeding exactly one prefetched device isolates the delete and
// confirms the map no longer contains the key.
func TestInvalidateInterfacePrefetch(t *testing.T) {
	cache := NewLookupCache(nil)
	deviceID := uuid.New()

	// Mark as prefetched.
	cache.interfacesMu.Lock()
	cache.interfacesPrefetched[deviceID] = true
	cache.interfacesMu.Unlock()

	cache.InvalidateInterfacePrefetch(deviceID)

	cache.interfacesMu.RLock()
	_, ok := cache.interfacesPrefetched[deviceID]
	cache.interfacesMu.RUnlock()
	if ok {
		t.Error("expected prefetch entry to be deleted")
	}
}

// TestInvalidateInterfacePrefetchNoEntry verifies InvalidateInterfacePrefetch
// is a safe no-op when the device has no prefetch entry to remove.
//
// Why it matters: the exporter invalidates a device's interface prefetch after
// module creation; doing so for a device that was never prefetched must not
// panic.
// Inputs: a random, unseeded device UUID. Outputs: no panic (delete of a
// missing map key is a no-op).
// Data choice: a fresh UUID guarantees the "no entry" branch is exercised.
func TestInvalidateInterfacePrefetchNoEntry(t *testing.T) {
	cache := NewLookupCache(nil)
	// Should not panic even when no entry exists
	cache.InvalidateInterfacePrefetch(uuid.New())
}

// ---------- CacheInterface ----------

// TestCacheInterfaceStore verifies CacheInterface stores an item under the
// "deviceID:ifaceName" composite key.
//
// Why it matters: interfaces are unique per device, so caching by name alone
// would collide across devices; the composite key keeps each device's interface
// IDs separate during export.
// Inputs: a device UUID, interface "eth0", and a CachedItem.
// Outputs: the item retrievable at key "<deviceID>:eth0".
// Data choice: the test reconstructs the key string directly to pin the exact
// format. Note: this overlaps TestCacheLocationAndInterface's interface check.
func TestCacheInterfaceStore(t *testing.T) {
	cache := NewLookupCache(nil)
	deviceID := uuid.New()
	ifaceID := uuid.New()
	cache.CacheInterface(deviceID, "eth0", &CachedItem{ID: ifaceID, Name: "eth0"})

	key := deviceID.String() + ":eth0"
	cache.interfacesMu.RLock()
	got, ok := cache.interfaces[key]
	cache.interfacesMu.RUnlock()
	if !ok || got.ID != ifaceID {
		t.Error("CacheInterface did not store correctly")
	}
}

// ---------- GetInterfaceByDeviceAndName (cache hit) ----------

// TestGetInterfaceByDeviceAndNameCacheHit verifies that
// GetInterfaceByDeviceAndName returns an interface previously stored via
// CacheInterface for the same device UUID and interface name.
//
// Why it matters: interface export caches device+name -> Nautobot interface ID
// so IP and cable assignment can reference interfaces without re-querying the
// API.
// Inputs: a CacheInterface(deviceID,"eth0",item) seed, then a lookup by the
// same key. Outputs: the cached item with matching ID, no error.
// Data choice: an identical device UUID and "eth0" name confirm the composite
// cache key round-trips.
func TestGetInterfaceByDeviceAndNameCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	deviceID := uuid.New()
	ifaceID := uuid.New()

	cache.CacheInterface(deviceID, "eth0", &CachedItem{ID: ifaceID, Name: "eth0"})

	item, err := cache.GetInterfaceByDeviceAndName(deviceID, "eth0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item == nil || item.ID != ifaceID {
		t.Error("expected to find cached interface")
	}
}

// ---------- CacheInterface / GetInterfaceByDeviceAndName cache hit ----------

// TestCacheInterfaceAndLookup verifies CacheInterface stores an interface and
// GetInterfaceByDeviceAndName retrieves it by (deviceID, name).
//
// Why it matters: interface caching lets the exporter attach IPs/cables to the
// right interface UUID without repeated Nautobot queries that can fail on names
// containing "/".
// Inputs: a (devID, "eth0")->ifaceID cache write, then a read of the same key.
// Outputs: the cached item whose ID equals ifaceID.
// Data choice: a single device/interface pair isolates the round-trip through
// the interface cache key.
func TestCacheInterfaceAndLookup(t *testing.T) {
	cache := NewLookupCache(nil)
	devID := uuid.New()
	ifaceID := uuid.New()

	cache.CacheInterface(devID, "eth0", &CachedItem{ID: ifaceID, Name: "eth0"})

	item, err := cache.GetInterfaceByDeviceAndName(devID, "eth0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item == nil {
		t.Fatal("expected non-nil item")
	}
	if item.ID != ifaceID {
		t.Errorf("ID = %s, want %s", item.ID, ifaceID)
	}
}
