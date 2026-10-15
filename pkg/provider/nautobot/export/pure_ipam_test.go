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

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// ---------- sortPrefixesByLength ----------

// TestSortPrefixesByLength verifies that prefixes are returned sorted by
// PrefixLen ascending, so wider container prefixes precede narrower ones.
//
// Why it matters: Nautobot requires a parent prefix to exist before its
// children, so export must create /8 before /16 before /24 or the API rejects
// the child as having no parent.
// Inputs: a map of three CaniPrefix values (/24, /8, /16) keyed by UUID.
// Outputs: a slice ordered 8, 16, 24.
// Data choice: out-of-order insertion with three distinct lengths proves the
// sort actually reorders rather than preserving map/iteration order.
func TestSortPrefixesByLength(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()

	prefixes := map[uuid.UUID]*devicetypes.CaniPrefix{
		id1: {ID: id1, Prefix: "10.0.0.0/24", PrefixLen: 24},
		id2: {ID: id2, Prefix: "10.0.0.0/8", PrefixLen: 8},
		id3: {ID: id3, Prefix: "10.0.0.0/16", PrefixLen: 16},
	}

	sorted := sortPrefixesByLength(prefixes)

	if len(sorted) != 3 {
		t.Fatalf("expected 3 prefixes, got %d", len(sorted))
	}
	if sorted[0].PrefixLen != 8 {
		t.Errorf("expected first prefix len 8, got %d", sorted[0].PrefixLen)
	}
	if sorted[1].PrefixLen != 16 {
		t.Errorf("expected second prefix len 16, got %d", sorted[1].PrefixLen)
	}
	if sorted[2].PrefixLen != 24 {
		t.Errorf("expected third prefix len 24, got %d", sorted[2].PrefixLen)
	}
}

// TestSortPrefixesByLengthSkipsNil verifies that nil map entries are dropped
// and only the non-nil prefix survives in the sorted result.
//
// Why it matters: inventory maps can contain nil pointers; dereferencing them
// while building Nautobot prefix requests would panic mid-export.
// Inputs: a map with one nil entry and one valid /16 prefix.
// Outputs: a single-element slice containing only the /16.
// Data choice: pairing exactly one nil with one valid entry isolates the
// nil-skip branch while confirming the survivor is carried through.
func TestSortPrefixesByLengthSkipsNil(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()

	prefixes := map[uuid.UUID]*devicetypes.CaniPrefix{
		id1: nil,
		id2: {ID: id2, Prefix: "192.168.0.0/16", PrefixLen: 16},
	}

	sorted := sortPrefixesByLength(prefixes)
	if len(sorted) != 1 {
		t.Fatalf("expected 1 prefix (nil skipped), got %d", len(sorted))
	}
	if sorted[0].PrefixLen != 16 {
		t.Errorf("expected prefix len 16, got %d", sorted[0].PrefixLen)
	}
}

// TestSortPrefixesByLengthEmpty verifies that an empty input map yields an
// empty (non-panicking) slice.
//
// Why it matters: an inventory with no prefixes must export cleanly rather
// than erroring, since IPAM data is optional.
// Inputs: an empty CaniPrefix map. Outputs: a zero-length slice.
// Data choice: the empty map exercises the boundary where the sort loop never
// executes, guarding against off-by-one or nil-slice assumptions.
func TestSortPrefixesByLengthEmpty(t *testing.T) {
	sorted := sortPrefixesByLength(map[uuid.UUID]*devicetypes.CaniPrefix{})
	if len(sorted) != 0 {
		t.Fatalf("expected 0 prefixes, got %d", len(sorted))
	}
}

// ---------- mapPrefixType ----------

// TestMapPrefixType verifies the cani-to-Nautobot prefix-type mapping for all
// known types plus the unknown/empty fallback to Network.
//
// Why it matters: Nautobot rejects prefix writes with an invalid type enum, so
// every cani PrefixType must map to a legal PrefixTypeChoices value.
// Inputs: container/network/pool plus an unknown and empty PrefixType.
// Outputs: the matching choice, defaulting to Network.
// Data choice: the table covers each switch case and both default branches so
// a forgotten case is caught immediately.
func TestMapPrefixType(t *testing.T) {
	tests := []struct {
		input    devicetypes.PrefixType
		expected nautobotapi.PrefixTypeChoices
	}{
		{devicetypes.PrefixTypeContainer, nautobotapi.PrefixTypeChoicesContainer},
		{devicetypes.PrefixTypeNetwork, nautobotapi.PrefixTypeChoicesNetwork},
		{devicetypes.PrefixTypePool, nautobotapi.PrefixTypeChoicesPool},
		{devicetypes.PrefixType("unknown"), nautobotapi.PrefixTypeChoicesNetwork},
		{devicetypes.PrefixType(""), nautobotapi.PrefixTypeChoicesNetwork},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := mapPrefixType(tt.input)
			if got != tt.expected {
				t.Errorf("mapPrefixType(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// ---------- mapIPAddressType ----------

// TestMapIPAddressType verifies the cani-to-Nautobot IP-address-type mapping
// for host/DHCP/SLAAC plus the unknown/empty fallback to Host.
//
// Why it matters: Nautobot validates the IP address type enum on write, so an
// unmapped cani type would fail the export of that address.
// Inputs: host/DHCP/SLAAC plus an unknown and empty IPAddressType.
// Outputs: the matching choice, defaulting to Host.
// Data choice: the table hits each switch case and both fallthrough branches,
// matching mapPrefixType's structure for consistency.
func TestMapIPAddressType(t *testing.T) {
	tests := []struct {
		input    devicetypes.IPAddressType
		expected nautobotapi.IPAddressTypeChoices
	}{
		{devicetypes.IPAddressTypeHost, nautobotapi.IPAddressTypeChoicesHost},
		{devicetypes.IPAddressTypeDHCP, nautobotapi.IPAddressTypeChoicesDhcp},
		{devicetypes.IPAddressTypeSLAAC, nautobotapi.IPAddressTypeChoicesSlaac},
		{devicetypes.IPAddressType("unknown"), nautobotapi.IPAddressTypeChoicesHost},
		{devicetypes.IPAddressType(""), nautobotapi.IPAddressTypeChoicesHost},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := mapIPAddressType(tt.input)
			if got != tt.expected {
				t.Errorf("mapIPAddressType(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// ---------- IPAM cache helpers ----------

// TestCacheVLAN verifies CacheVLAN stores an item under the "vid:location" key
// in the package-level VLAN cache.
//
// Why it matters: VLANs are scoped per location in Nautobot, so caching by VID
// alone would collide across sites; the composite key keeps resolved IDs
// correct during export.
// Inputs: VID 100, location "site-a", and a CachedItem.
// Outputs: the item retrievable at key "100:site-a".
// Data choice: the test reads the cache's vlans map directly to confirm the
// exact composite key format.
func TestCacheVLAN(t *testing.T) {
	cache := NewLookupCache(nil)
	item := &CachedItem{ID: uuid.New(), Name: "VLAN100"}
	cache.CacheVLAN(100, "site-a", item)

	cache.vlansMu.RLock()
	got, ok := cache.vlans["100:site-a"]
	cache.vlansMu.RUnlock()
	if !ok || got.ID != item.ID {
		t.Error("CacheVLAN did not cache correctly")
	}
}

// TestCachePrefix verifies CachePrefix stores an item under its CIDR key in the
// package-level prefix cache.
//
// Why it matters: prefixes are looked up by namespace and CIDR while linking IP
// addresses and child prefixes; a wrong key causes redundant API lookups or
// mis-parenting across namespaces.
// Inputs: namespace "Global", CIDR "10.0.0.0/24" and a CachedItem.
// Outputs: the item retrievable at key "Global/10.0.0.0/24".
// Data choice: reading the cache's prefixes map confirms the scoped key format.
func TestCachePrefix(t *testing.T) {
	cache := NewLookupCache(nil)
	item := &CachedItem{ID: uuid.New(), Name: "10.0.0.0/24"}
	cache.CachePrefix("Global", "10.0.0.0/24", item)

	cache.prefixesMu.RLock()
	got, ok := cache.prefixes[scopedKey("Global", "10.0.0.0/24")]
	cache.prefixesMu.RUnlock()
	if !ok || got.ID != item.ID {
		t.Error("CachePrefix did not cache correctly")
	}
}

// TestCacheIPAddress verifies CacheIPAddress stores an item under its
// namespace-and-host key in the package-level IP cache.
//
// Why it matters: IP addresses are deduplicated by host within a namespace
// during export; keying on the masked text would re-create an address whose
// mask changed, and keying without the namespace would collide across tenants.
// Inputs: a blank namespace, address "10.0.0.1/32" and a CachedItem.
// Outputs: the item retrievable at key "Global/10.0.0.1".
// Data choice: the blank namespace proves the Global default and the /32 proves
// the mask is stripped.
func TestCacheIPAddress(t *testing.T) {
	cache := NewLookupCache(nil)
	item := &CachedItem{ID: uuid.New(), Name: "10.0.0.1/32"}
	cache.CacheIPAddress("", "10.0.0.1/32", item)

	cache.ipAddressesMu.RLock()
	got, ok := cache.ipAddresses[scopedKey("", "10.0.0.1")]
	cache.ipAddressesMu.RUnlock()
	if !ok || got.ID != item.ID {
		t.Error("CacheIPAddress did not cache correctly")
	}
}

// ---------- LookupVLAN cache hit ----------

// TestLookupVLANCacheHit verifies LookupVLAN returns the cached item (without an
// API call) when the VID/location key was previously cached.
//
// Why it matters: the exporter caches resolved VLANs to avoid redundant Nautobot
// lookups; a cache hit must short-circuit before touching the (here nil) client.
// Inputs: a cache seeded via CacheVLAN(100,"DC1",...); then LookupVLAN(100,"DC1").
// Outputs: the same *CachedItem (matching ID and Name) and a nil error.
// Data choice: the nil client guarantees any miss would attempt an API call, so
// a clean return proves the cache path was taken.
func TestLookupVLANCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	vlanID := uuid.New()
	cache.CacheVLAN(100, "DC1", &CachedItem{ID: vlanID, Name: "VLAN-100"})

	item, err := cache.LookupVLAN(100, "DC1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item == nil {
		t.Fatal("expected non-nil item")
	}
	if item.ID != vlanID {
		t.Errorf("ID = %s, want %s", item.ID, vlanID)
	}
	if item.Name != "VLAN-100" {
		t.Errorf("Name = %q, want 'VLAN-100'", item.Name)
	}
}

// TestLookupVLANCacheMissDifferentKey verifies a lookup for an unseeded
// VID/location key does not return a VLAN cached under another key.
//
// Why it matters: VLANs are scoped by VID and location; returning a cached VLAN
// for a different key would attach prefixes/IPs to the wrong Nautobot VLAN.
// Inputs: CacheVLAN(100,"DC1",...) then LookupVLAN(200,"DC1") against an empty
// API response. Outputs: nil item, nil error, and one HTTP lookup.
// Data choice: changing only the VID proves the composite cache key separates
// neighboring VLANs in the same location.
func TestLookupVLANCacheMissDifferentKey(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, emptyListJSON))
	defer cleanup()

	vlanID := uuid.New()
	e.Cache.CacheVLAN(100, "DC1", &CachedItem{ID: vlanID, Name: "VLAN-100"})

	item, err := e.Cache.LookupVLAN(200, "DC1")
	if err != nil {
		t.Fatalf("LookupVLAN() error = %v", err)
	}
	if item != nil {
		t.Fatalf("expected nil item for different VID, got %+v", item)
	}
	if calls != 1 {
		t.Errorf("expected exactly one HTTP request for cache miss, got %d", calls)
	}
}

// ---------- LookupPrefix cache hit ----------

// TestLookupPrefixCacheHit verifies LookupPrefix returns the cached item (without
// an API call) when the CIDR was previously cached.
//
// Why it matters: caching resolved prefixes lets the exporter reuse IPAM lookups
// instead of re-querying Nautobot for every IP it loads.
// Inputs: CachePrefix("10.0.0.0/24",...) then LookupPrefix("10.0.0.0/24").
// Outputs: the same *CachedItem (matching ID) and a nil error.
// Data choice: a /24 CIDR with a nil client ensures a miss would hit the API, so
// the clean hit confirms the cache short-circuit.
func TestLookupPrefixCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	prefixID := uuid.New()
	cache.CachePrefix("Global", "10.0.0.0/24", &CachedItem{ID: prefixID, Name: "10.0.0.0/24"})

	item, err := cache.LookupPrefix("Global", "10.0.0.0/24")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item == nil {
		t.Fatal("expected non-nil item")
	}
	if item.ID != prefixID {
		t.Errorf("ID = %s, want %s", item.ID, prefixID)
	}
}

// ---------- LookupIPAddress cache hit ----------

// TestLookupIPAddressCacheHit verifies LookupIPAddress returns the cached item
// (without an API call) when the address was previously cached.
//
// Why it matters: caching resolved IPs avoids redundant Nautobot lookups while
// the exporter assigns addresses to interfaces.
// Inputs: CacheIPAddress("192.168.1.1/32",...) then
// LookupIPAddress("192.168.1.1/32"). Outputs: the same *CachedItem (matching ID)
// and a nil error.
// Data choice: a /32 host address with a nil client makes any miss call the API,
// so the clean hit proves the cache path was used.
func TestLookupIPAddressCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	ipID := uuid.New()
	cache.CacheIPAddress("Global", "192.168.1.1/32", &CachedItem{ID: ipID, Name: "192.168.1.1/32"})

	item, err := cache.LookupIPAddress("Global", "192.168.1.1/32")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item == nil {
		t.Fatal("expected non-nil item")
	}
	if item.ID != ipID {
		t.Errorf("ID = %s, want %s", item.ID, ipID)
	}
}
