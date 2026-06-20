package export

import (
	"testing"

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// TestCacheCreatedInterfaces_MatchesByPosition verifies a bulk-created
// interface is cached under the device ID and name of the batch item at the
// same slice index, carrying the Nautobot ID returned in the response.
//
// Why it matters: Phase 6 cable creation resolves endpoints via this cache;
// position-based matching is how freshly POSTed interfaces (Nautobot returns
// results in request order) acquire their remote IDs without a lookup.
// Inputs: a one-item batch and one created Interface named "eth0". Outputs: a
// cache entry retrievable by GetInterfaceByDeviceAndName.
// Data choice: matching names keep the focus on the position-to-ID mapping.
func TestCacheCreatedInterfaces_MatchesByPosition(t *testing.T) {
	cache := NewLookupCache(nil)
	e := &Exporter{Cache: cache}

	devID := uuid.New()
	ifaceID := uuid.New()
	oapiID := openapi_types.UUID(ifaceID)

	batch := []bulkInterfaceItem{
		{DeviceID: devID, DeviceName: "dev1", Spec: interfaceSpec{Name: "eth0"}},
	}
	created := []nautobotapi.Interface{
		{Id: &oapiID, Name: "eth0"},
	}

	e.cacheCreatedInterfaces(batch, created)

	cached, err := cache.GetInterfaceByDeviceAndName(devID, "eth0")
	if err != nil {
		t.Fatalf("GetInterfaceByDeviceAndName: %v", err)
	}
	if cached == nil {
		t.Fatal("expected cached interface, got nil")
	}
	if cached.ID != ifaceID {
		t.Errorf("expected ID %s, got %s", ifaceID, cached.ID)
	}
	if cached.Name != "eth0" {
		t.Errorf("expected name eth0, got %s", cached.Name)
	}
}

// TestCacheCreatedInterfaces_SkipsNilID verifies an interface returned with a
// nil Id is not written to the cache.
//
// Why it matters: a created interface without a usable Nautobot ID is worthless
// for later cable/IP assignment, so caching it would plant a broken reference
// that silently misdirects downstream phases.
// Inputs: a one-item batch and one created Interface whose Id is nil. Outputs:
// the internal interfaces map is asserted to hold no entry for that key.
// Data choice: the test reads the map directly because
// GetInterfaceByDeviceAndName would fall through to a live API call on a miss.
func TestCacheCreatedInterfaces_SkipsNilID(t *testing.T) {
	cache := NewLookupCache(nil)
	e := &Exporter{Cache: cache}

	devID := uuid.New()

	batch := []bulkInterfaceItem{
		{DeviceID: devID, DeviceName: "dev1", Spec: interfaceSpec{Name: "eth0"}},
	}
	created := []nautobotapi.Interface{
		{Name: "eth0"}, // Id is nil — should be skipped
	}

	e.cacheCreatedInterfaces(batch, created)

	// Verify the cache was NOT populated by checking the internal map
	// directly. We can't call GetInterfaceByDeviceAndName because it
	// falls through to an API call when the key is missing.
	cache.interfacesMu.RLock()
	key := interfaceCacheKey(devID, "eth0")
	_, found := cache.interfaces[key]
	cache.interfacesMu.RUnlock()
	if found {
		t.Error("expected interface with nil ID to not be cached")
	}
}

// TestCacheCreatedInterfaces_MultipleBatch verifies position-based caching
// across several devices: each created interface lands under its own batch
// item's device/name with the matching Nautobot ID.
//
// Why it matters: real exports send interfaces for many devices in one batch,
// and each must cache against the correct device so cables wire the right ports.
// Inputs: a two-item batch (dev1/eth0, dev2/mgmt0) and two created Interfaces.
// Outputs: two independent, correctly-keyed cache entries.
// Data choice: distinct device IDs and names prove entries do not collide or
// cross-contaminate.
func TestCacheCreatedInterfaces_MultipleBatch(t *testing.T) {
	cache := NewLookupCache(nil)
	e := &Exporter{Cache: cache}

	dev1 := uuid.New()
	dev2 := uuid.New()
	id1 := uuid.New()
	id2 := uuid.New()
	oapi1 := openapi_types.UUID(id1)
	oapi2 := openapi_types.UUID(id2)

	batch := []bulkInterfaceItem{
		{DeviceID: dev1, DeviceName: "dev1", Spec: interfaceSpec{Name: "eth0"}},
		{DeviceID: dev2, DeviceName: "dev2", Spec: interfaceSpec{Name: "mgmt0"}},
	}
	created := []nautobotapi.Interface{
		{Id: &oapi1, Name: "eth0"},
		{Id: &oapi2, Name: "mgmt0"},
	}

	e.cacheCreatedInterfaces(batch, created)

	c1, _ := cache.GetInterfaceByDeviceAndName(dev1, "eth0")
	c2, _ := cache.GetInterfaceByDeviceAndName(dev2, "mgmt0")

	if c1 == nil || c1.ID != id1 {
		t.Errorf("dev1/eth0: expected ID %s, got %v", id1, c1)
	}
	if c2 == nil || c2.ID != id2 {
		t.Errorf("dev2/mgmt0: expected ID %s, got %v", id2, c2)
	}
}

// TestCacheCreatedInterfaces_MoreCreatedThanBatch verifies the overflow branch:
// when Nautobot returns more interfaces than were sent, the extra item is
// cached using uuid.Nil for the device and the name from the response body.
//
// Why it matters: it defends the cache against a length mismatch so a
// surprising response cannot index past the batch slice and panic mid-export.
// Inputs: a one-item batch but two created Interfaces (eth0, eth1). Outputs:
// eth0 keyed by its real device; eth1 keyed by (uuid.Nil, response name).
// Data choice: a deliberate batch/response size skew is the only way to reach
// the fallback path.
func TestCacheCreatedInterfaces_MoreCreatedThanBatch(t *testing.T) {
	cache := NewLookupCache(nil)
	e := &Exporter{Cache: cache}

	devID := uuid.New()
	id1 := uuid.New()
	id2 := uuid.New()
	oapi1 := openapi_types.UUID(id1)
	oapi2 := openapi_types.UUID(id2)

	// One item in batch, but two in created (edge case — extra item uses
	// response Name as fallback).
	batch := []bulkInterfaceItem{
		{DeviceID: devID, DeviceName: "dev1", Spec: interfaceSpec{Name: "eth0"}},
	}
	created := []nautobotapi.Interface{
		{Id: &oapi1, Name: "eth0"},
		{Id: &oapi2, Name: "eth1"},
	}

	e.cacheCreatedInterfaces(batch, created)

	c1, _ := cache.GetInterfaceByDeviceAndName(devID, "eth0")
	if c1 == nil || c1.ID != id1 {
		t.Errorf("eth0: expected ID %s, got %v", id1, c1)
	}

	// Second item exceeds batch length, so it uses zero deviceID and
	// the response's Name field. Verify via the internal cache map.
	cache.interfacesMu.RLock()
	key := interfaceCacheKey(uuid.Nil, "eth1")
	item, found := cache.interfaces[key]
	cache.interfacesMu.RUnlock()
	if !found || item == nil || item.ID != id2 {
		t.Errorf("eth1 (overflow): expected cached ID %s, found=%v item=%v", id2, found, item)
	}
}

// TestCacheCreatedInterfaces_EmptySlices verifies the method is a safe no-op
// for nil and empty inputs (no panic, nothing cached).
//
// Why it matters: an empty bulk batch (e.g. a device with no new interfaces)
// is a normal case, so the export must not crash when there is nothing to do.
// Inputs: (nil, nil) and ([]bulkInterfaceItem{}, []nautobotapi.Interface{}).
// Outputs: none — the test passes as long as neither call panics.
// Data choice: both empty-input shapes exercise the loop's zero-iteration path.
func TestCacheCreatedInterfaces_EmptySlices(t *testing.T) {
	cache := NewLookupCache(nil)
	e := &Exporter{Cache: cache}

	// Should not panic on empty inputs
	e.cacheCreatedInterfaces(nil, nil)
	e.cacheCreatedInterfaces([]bulkInterfaceItem{}, []nautobotapi.Interface{})
}
