package export

import (
	"testing"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

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

func TestCacheCreatedInterfaces_EmptySlices(t *testing.T) {
	cache := NewLookupCache(nil)
	e := &Exporter{Cache: cache}

	// Should not panic on empty inputs
	e.cacheCreatedInterfaces(nil, nil)
	e.cacheCreatedInterfaces([]bulkInterfaceItem{}, []nautobotapi.Interface{})
}
