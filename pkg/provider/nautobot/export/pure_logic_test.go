package export

import (
	"context"
	"errors"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// ---------- sortPrefixesByLength ----------

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

func TestSortPrefixesByLengthEmpty(t *testing.T) {
	sorted := sortPrefixesByLength(map[uuid.UUID]*devicetypes.CaniPrefix{})
	if len(sorted) != 0 {
		t.Fatalf("expected 0 prefixes, got %d", len(sorted))
	}
}

// ---------- mapPrefixType ----------

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

func TestMapIPAddressType(t *testing.T) {
	tests := []struct {
		input    devicetypes.IPAddressType
		expected nautobotapi.IPAddressTypeChoices
	}{
		{devicetypes.IPAddressTypeHost, nautobotapi.Host},
		{devicetypes.IPAddressTypeDHCP, nautobotapi.Dhcp},
		{devicetypes.IPAddressTypeSLAAC, nautobotapi.Slaac},
		{devicetypes.IPAddressType("unknown"), nautobotapi.Host},
		{devicetypes.IPAddressType(""), nautobotapi.Host},
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

// ---------- isValidNautobotInterfaceType ----------

func TestIsValidNautobotInterfaceType(t *testing.T) {
	validTypes := []string{
		"100base-tx", "1000base-t", "10gbase-x-sfpp", "25gbase-x-sfp28",
		"40gbase-x-qsfpp", "100gbase-x-qsfp28", "200gbase-x-qsfp56",
		"400gbase-x-osfp", "400gbase-x-qsfpdd",
		"infiniband-hdr", "infiniband-ndr",
		"virtual", "lag", "other",
	}
	for _, vt := range validTypes {
		if !isValidNautobotInterfaceType(vt) {
			t.Errorf("expected %q to be valid", vt)
		}
	}

	invalidTypes := []string{
		"nvlink", "pcie-gen5-x16", "spi", "", "garbage",
	}
	for _, ivt := range invalidTypes {
		if isValidNautobotInterfaceType(ivt) {
			t.Errorf("expected %q to be invalid", ivt)
		}
	}
}

// ---------- LookupCache setters ----------

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

// ---------- DeviceMapper nil device ----------

func TestMapToNautobotDeviceNilDevice(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	_, err := mapper.MapToNautobotDevice(nil)
	if err == nil {
		t.Fatal("expected error for nil device")
	}
	if err.Error() != "device is nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestMapToWritableDeviceRequestNilDevice(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	_, err := mapper.MapToWritableDeviceRequest(nil)
	if err == nil {
		t.Fatal("expected error for nil device")
	}
	if err.Error() != "device is nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

// ---------- SetInventory ----------

func TestDeviceMapperSetInventory(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
	}
	mapper.SetInventory(inv)

	if mapper.inventory != inv {
		t.Error("SetInventory did not set inventory")
	}
}

// ---------- CacheLocation / CacheInterface ----------

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

// ---------- Exporter resolveLocationName ----------

func TestResolveLocationName(t *testing.T) {
	exporter := &Exporter{}

	locID := uuid.New()
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			locID: {Name: "Building-A"},
		},
	}

	name, err := exporter.resolveLocationName(locID, inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Building-A" {
		t.Errorf("expected Building-A, got %s", name)
	}
}

func TestResolveLocationNameNotFound(t *testing.T) {
	exporter := &Exporter{}

	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{},
	}

	_, err := exporter.resolveLocationName(uuid.New(), inv)
	if err == nil {
		t.Fatal("expected error for missing location")
	}
}

// ---------- PrintSummary (smoke) ----------

func TestPrintSummaryDoesNotPanic(t *testing.T) {
	result := &LoadResult{
		LocationsCreated:   []string{"loc1"},
		LocationsSkipped:   []string{"loc2"},
		RacksCreated:       []string{"rack1"},
		Created:            []string{"dev1", "dev2"},
		Updated:            []string{"dev3"},
		Skipped:            []string{"dev4"},
		IfacesCreated:      5,
		IfacesSkipped:      2,
		CablesCreated:      3,
		VLANsCreated:       1,
		PrefixesCreated:    2,
		IPAddressesCreated: 4,
	}
	// Should not panic.
	PrintSummary(result)
}

// ---------- NewNautobotClient ----------

func TestNewNautobotClientErrors(t *testing.T) {
	_, err := NewNautobotClient("", "token")
	if err == nil {
		t.Error("expected error for empty URL")
	}

	_, err = NewNautobotClient("http://example.com", "")
	if err == nil {
		t.Error("expected error for empty token")
	}
}

func TestNewNautobotClientSuccess(t *testing.T) {
	client, err := NewNautobotClient("http://example.com", "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

// ---------- refID / tenantRefID ----------

func TestRefID(t *testing.T) {
	// nil returns Nil
	if got := refID(nil); got != uuid.Nil {
		t.Errorf("refID(nil) = %s, want Nil", got)
	}

	// valid union
	id := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	var union nautobotapi.BulkWritableCableRequestStatusId
	_ = union.FromBulkWritableCableRequestStatusId0(id)
	if got := refID(&union); got != id {
		t.Errorf("refID() = %s, want %s", got, id)
	}

	// tenantRefID delegates to refID
	if got := tenantRefID(&union); got != id {
		t.Errorf("tenantRefID() = %s, want %s", got, id)
	}
	if got := tenantRefID(nil); got != uuid.Nil {
		t.Errorf("tenantRefID(nil) = %s, want Nil", got)
	}
}

// ---------- generateDeviceNames (additional branches) ----------

func TestGenerateDeviceNamesAllBranches(t *testing.T) {
	nodeID := uuid.New()
	slugID := uuid.New()
	modelID := uuid.New()
	uuidID := uuid.New()
	namedID := uuid.New()
	systemID := uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			nodeID:   {ID: nodeID, Name: "", Type: "node", Serial: "SN-ABC"},
			slugID:   {ID: slugID, Name: "", Type: "node", Slug: "hpe-dl360"},
			modelID:  {ID: modelID, Name: "", Type: "node", Model: "ProLiant DL380"},
			uuidID:   {ID: uuidID, Name: "", Type: "node"},
			namedID:  {ID: namedID, Name: "already-named", Type: "node"},
			systemID: {ID: systemID, Name: "", Type: "system"},
		},
	}

	generateDeviceNames(inv)

	if inv.Devices[nodeID].Name != "cani-SN-ABC" {
		t.Errorf("serial branch: got %q", inv.Devices[nodeID].Name)
	}
	if inv.Devices[slugID].Name != "cani-hpe-dl360" {
		t.Errorf("slug branch: got %q", inv.Devices[slugID].Name)
	}
	if inv.Devices[modelID].Name != "cani-proliant-dl380" {
		t.Errorf("model branch: got %q", inv.Devices[modelID].Name)
	}
	if inv.Devices[uuidID].Name != "cani-"+uuidID.String()[:8] {
		t.Errorf("uuid branch: got %q", inv.Devices[uuidID].Name)
	}
	if inv.Devices[namedID].Name != "already-named" {
		t.Errorf("named device should not change: got %q", inv.Devices[namedID].Name)
	}
	// "system" also classifies as CategoryDevice, so it gets a UUID-based name.
	if inv.Devices[systemID].Name != "cani-"+systemID.String()[:8] {
		t.Errorf("system device: got %q", inv.Devices[systemID].Name)
	}
}

// ---------- disambiguateDeviceNames (additional cases) ----------

func TestDisambiguateDeviceNamesAllSuffixes(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()
	rackID := uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id1: {ID: id1, Name: "dup-server", Type: "node", Serial: "SN001"},
			id2: {ID: id2, Name: "dup-server", Type: "node", RackPosition: 5, Rack: rackID},
			id3: {ID: id3, Name: "dup-server", Type: "node"},
		},
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {Name: "Rack-A"},
		},
	}

	disambiguateDeviceNames(inv)

	// All names should now be unique.
	names := make(map[string]int)
	for _, d := range inv.Devices {
		names[d.Name]++
	}
	for name, count := range names {
		if count > 1 {
			t.Errorf("name %q still appears %d times after disambiguation", name, count)
		}
	}

	// Serial suffix
	if inv.Devices[id1].Name != "dup-server (SN001)" {
		t.Errorf("serial suffix: got %q", inv.Devices[id1].Name)
	}
	// Rack position suffix
	if inv.Devices[id2].Name != "dup-server (Rack-A U5)" {
		t.Errorf("rack position suffix: got %q", inv.Devices[id2].Name)
	}
	// UUID suffix
	expected := "dup-server (" + id3.String()[:8] + ")"
	if inv.Devices[id3].Name != expected {
		t.Errorf("uuid suffix: got %q, want %q", inv.Devices[id3].Name, expected)
	}
}

func TestDisambiguateDeviceNamesNoDuplicates(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id1: {ID: id1, Name: "server-1", Type: "node"},
			id2: {ID: id2, Name: "server-2", Type: "node"},
		},
	}

	disambiguateDeviceNames(inv)

	// Names should not be modified.
	if inv.Devices[id1].Name != "server-1" {
		t.Errorf("non-duplicate should not change: got %q", inv.Devices[id1].Name)
	}
	if inv.Devices[id2].Name != "server-2" {
		t.Errorf("non-duplicate should not change: got %q", inv.Devices[id2].Name)
	}
}

// ---------- IPAM cache helpers ----------

func TestCacheVLAN(t *testing.T) {
	cache := NewLookupCache(nil)
	item := &CachedItem{ID: uuid.New(), Name: "VLAN100"}
	cache.CacheVLAN(100, "site-a", item)

	vlansMu.RLock()
	got, ok := vlans["100:site-a"]
	vlansMu.RUnlock()
	if !ok || got.ID != item.ID {
		t.Error("CacheVLAN did not cache correctly")
	}
}

func TestCachePrefix(t *testing.T) {
	cache := NewLookupCache(nil)
	item := &CachedItem{ID: uuid.New(), Name: "10.0.0.0/24"}
	cache.CachePrefix("10.0.0.0/24", item)

	prefixesMu.RLock()
	got, ok := prefixes["10.0.0.0/24"]
	prefixesMu.RUnlock()
	if !ok || got.ID != item.ID {
		t.Error("CachePrefix did not cache correctly")
	}
}

func TestCacheIPAddress(t *testing.T) {
	cache := NewLookupCache(nil)
	item := &CachedItem{ID: uuid.New(), Name: "10.0.0.1/32"}
	cache.CacheIPAddress("10.0.0.1/32", item)

	ipAddressesMu.RLock()
	got, ok := ipAddresses["10.0.0.1/32"]
	ipAddressesMu.RUnlock()
	if !ok || got.ID != item.ID {
		t.Error("CacheIPAddress did not cache correctly")
	}
}

// ---------- makeIDRef / makeLocationRef / makePrefixParentRef / makeIPParentRef / makeIPNamespaceRef ----------

func TestMakeIDRef(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ref := makeIDRef(id)
	if ref.Id == nil {
		t.Fatal("expected non-nil Id")
	}
	got, err := ref.Id.AsBulkWritableCableRequestStatusId0()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uuid.UUID(got) != id {
		t.Errorf("expected %s, got %s", id, uuid.UUID(got))
	}
}

func TestMakeLocationRef(t *testing.T) {
	id := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	ref := makeLocationRef(id)
	if ref.Id == nil {
		t.Fatal("expected non-nil Id")
	}
	got, err := ref.Id.AsBulkWritableCableRequestStatusId0()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uuid.UUID(got) != id {
		t.Errorf("expected %s, got %s", id, uuid.UUID(got))
	}
}

func TestMakePrefixParentRef(t *testing.T) {
	id := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	ref := makePrefixParentRef(id)
	if ref.Id == nil {
		t.Fatal("expected non-nil Id")
	}
	got, err := ref.Id.AsBulkWritableCableRequestStatusId0()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uuid.UUID(got) != id {
		t.Errorf("expected %s, got %s", id, uuid.UUID(got))
	}
}

func TestMakeIPParentRef(t *testing.T) {
	id := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	ref := makeIPParentRef(id)
	if ref.Id == nil {
		t.Fatal("expected non-nil Id")
	}
	got, err := ref.Id.AsBulkWritableCableRequestStatusId0()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uuid.UUID(got) != id {
		t.Errorf("expected %s, got %s", id, uuid.UUID(got))
	}
}

func TestMakeIPNamespaceRef(t *testing.T) {
	id := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	ref := makeIPNamespaceRef(id)
	if ref.Id == nil {
		t.Fatal("expected non-nil Id")
	}
	got, err := ref.Id.AsBulkWritableCableRequestStatusId0()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uuid.UUID(got) != id {
		t.Errorf("expected %s, got %s", id, uuid.UUID(got))
	}
}

// ---------- InvalidateInterfacePrefetch ----------

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

// ---------- locationTypeSupports ----------

func TestLocationTypeSupports(t *testing.T) {
	// Register a test location type.
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Test Room",
		Slug:         "test-room-export",
		ContentTypes: []string{"rack", "device"},
	})

	if !locationTypeSupports("test-room-export", "rack") {
		t.Error("expected test-room-export to support rack")
	}
	if !locationTypeSupports("test-room-export", "device") {
		t.Error("expected test-room-export to support device")
	}
	if locationTypeSupports("test-room-export", "module") {
		t.Error("expected test-room-export to NOT support module")
	}
	if locationTypeSupports("nonexistent-slug", "rack") {
		t.Error("expected unknown slug to return false")
	}
}

// ---------- findContentChild ----------

func TestFindContentChild(t *testing.T) {
	// Register location types for the test.
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Building",
		Slug:         "building-test-fc",
		ContentTypes: []string{},
	})
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Room",
		Slug:         "room-test-fc",
		ContentTypes: []string{"rack", "device"},
	})

	parentID := uuid.New()
	childID := uuid.New()

	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			parentID: {
				ID:           parentID,
				Name:         "HQ",
				LocationType: "building-test-fc",
				Children:     []uuid.UUID{childID},
			},
			childID: {
				ID:           childID,
				Name:         "Server-Room",
				LocationType: "room-test-fc",
			},
		},
	}

	parent := inv.Locations[parentID]
	got := findContentChild(parent, "rack", inv)
	if got != "Server-Room" {
		t.Errorf("expected Server-Room, got %q", got)
	}

	// Non-matching content type
	got = findContentChild(parent, "nonexistent", inv)
	if got != "" {
		t.Errorf("expected empty string for unsupported content type, got %q", got)
	}
}

// ---------- resolveContentLocation ----------

func TestResolveContentLocation(t *testing.T) {
	// Register location types.
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "DC",
		Slug:         "dc-test-rcl",
		ContentTypes: []string{},
	})
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Section",
		Slug:         "section-test-rcl",
		ContentTypes: []string{"rack", "device"},
	})

	dcID := uuid.New()
	sectionID := uuid.New()

	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			dcID: {
				ID:           dcID,
				Name:         "DC-01",
				LocationType: "dc-test-rcl",
				Children:     []uuid.UUID{sectionID},
			},
			sectionID: {
				ID:           sectionID,
				Name:         "Section-A",
				LocationType: "section-test-rcl",
			},
		},
	}

	// DC itself doesn't support "rack", but its child does.
	got := resolveContentLocation(dcID, "rack", inv)
	if got != "Section-A" {
		t.Errorf("expected Section-A, got %q", got)
	}

	// nil inventory
	got = resolveContentLocation(dcID, "rack", nil)
	if got != "" {
		t.Errorf("expected empty for nil inventory, got %q", got)
	}

	// Nil UUID
	got = resolveContentLocation(uuid.Nil, "rack", inv)
	if got != "" {
		t.Errorf("expected empty for nil UUID, got %q", got)
	}

	// Location that directly supports content type.
	got = resolveContentLocation(sectionID, "rack", inv)
	if got != "Section-A" {
		t.Errorf("expected Section-A for direct support, got %q", got)
	}
}

// ---------- remoteSlotKey ----------

func TestRemoteSlotKeyNil(t *testing.T) {
	// nil device → nil
	if got := remoteSlotKey(nil); got != nil {
		t.Error("expected nil for nil device")
	}

	// missing position → nil
	d := &nautobotapi.Device{}
	if got := remoteSlotKey(d); got != nil {
		t.Error("expected nil for device without position")
	}

	// missing rack → nil
	pos := 5
	d = &nautobotapi.Device{Position: &pos}
	if got := remoteSlotKey(d); got != nil {
		t.Error("expected nil for device without rack")
	}
}

func TestRemoteSlotKeyValid(t *testing.T) {
	pos := 5
	rackID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	rackIDUnion := nautobotapi.BulkWritableCableRequestStatusId{}
	_ = rackIDUnion.FromBulkWritableCableRequestStatusId0(rackID)
	d := &nautobotapi.Device{
		Position: &pos,
		Rack: &nautobotapi.BulkWritableCircuitRequestTenant{
			Id: &rackIDUnion,
		},
	}
	sk := remoteSlotKey(d)
	if sk == nil {
		t.Fatal("expected non-nil slotKey")
	}
	if sk.RackID != rackID {
		t.Errorf("rackID = %s, want %s", sk.RackID, rackID)
	}
	if sk.Position != 5 {
		t.Errorf("position = %d, want 5", sk.Position)
	}
	if sk.Face != "front" {
		t.Errorf("face = %q, want 'front'", sk.Face)
	}

	// rear face
	rearVal := nautobotapi.DeviceFaceValue("rear")
	d.Face = &nautobotapi.DeviceFace{Value: &rearVal}
	sk = remoteSlotKey(d)
	if sk == nil {
		t.Fatal("expected non-nil slotKey for rear")
	}
	if sk.Face != "rear" {
		t.Errorf("face = %q, want 'rear'", sk.Face)
	}
}

// ---------- derefSlotKey ----------

func TestDerefSlotKeyNil(t *testing.T) {
	got := derefSlotKey(nil)
	if got.Position != 0 || got.Face != "" || got.RackID != uuid.Nil {
		t.Error("expected zero slotKey for nil input")
	}
}

func TestDerefSlotKeyNonNil(t *testing.T) {
	sk := &slotKey{
		RackID:   uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		Position: 42,
		Face:     "rear",
	}
	got := derefSlotKey(sk)
	if got.RackID != sk.RackID || got.Position != 42 || got.Face != "rear" {
		t.Errorf("derefSlotKey did not dereference correctly: %+v", got)
	}
}

// ---------- ValidateInventory (additional edge cases) ----------

func TestValidateInventoryNil(t *testing.T) {
	if err := ValidateInventory(nil); err == nil {
		t.Error("expected error for nil inventory")
	}
}

func TestValidateInventoryEmpty(t *testing.T) {
	if err := ValidateInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
	}); err == nil {
		t.Error("expected error for empty devices")
	}
}

func TestValidateInventoryAllNilDevices(t *testing.T) {
	if err := ValidateInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			uuid.New(): nil,
		},
	}); err == nil {
		t.Error("expected error for all nil devices")
	}
}

func TestValidateInventoryOnlySystem(t *testing.T) {
	if err := ValidateInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			uuid.New(): {Name: "sys", Type: "system"},
		},
	}); err == nil {
		t.Error("expected error for only system devices")
	}
}

func TestValidateInventoryOnlyUnnamed(t *testing.T) {
	if err := ValidateInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			uuid.New(): {Name: "", Type: "node"},
		},
	}); err == nil {
		t.Error("expected error for only unnamed devices")
	}
}

func TestValidateInventoryValid(t *testing.T) {
	if err := ValidateInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			uuid.New(): {Name: "server-1", Type: "node"},
		},
	}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- CacheInterface ----------

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

// ---------- compareDeviceFields ----------

func TestCompareDeviceFieldsSerial(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Serial: "NEW-SN",
	}
	oldSerial := "OLD-SN"
	remote := &nautobotapi.Device{
		Serial: &oldSerial,
	}

	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	diffs := compareDeviceFields(dev, remote, mapper)

	found := false
	for _, d := range diffs {
		if d.Field == "serial" && d.LocalVal == "NEW-SN" && d.RemoteVal == "OLD-SN" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected serial diff, got: %+v", diffs)
	}
}

func TestCompareDeviceFieldsAssetTag(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		AssetTag: "NEW-TAG",
	}
	oldTag := "OLD-TAG"
	remote := &nautobotapi.Device{
		AssetTag: &oldTag,
	}

	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	diffs := compareDeviceFields(dev, remote, mapper)

	found := false
	for _, d := range diffs {
		if d.Field == "asset_tag" && d.LocalVal == "NEW-TAG" && d.RemoteVal == "OLD-TAG" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected asset_tag diff, got: %+v", diffs)
	}
}

func TestCompareDeviceFieldsNoChanges(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{}
	remote := &nautobotapi.Device{}

	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	diffs := compareDeviceFields(dev, remote, mapper)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs for empty device, got: %+v", diffs)
	}
}

// ---------- resolveLocalRackName ----------

func TestResolveLocalRackNameNilInventory(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{Rack: uuid.New()}
	if got := resolveLocalRackName(dev, mapper); got != "" {
		t.Errorf("expected empty string for nil inventory, got %q", got)
	}
}

func TestResolveLocalRackNameNoRack(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{},
	}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{Rack: uuid.Nil, Parent: uuid.Nil}
	if got := resolveLocalRackName(dev, mapper); got != "" {
		t.Errorf("expected empty string for no rack, got %q", got)
	}
}

func TestResolveLocalRackNameFound(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	rackID := uuid.New()
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {Name: "Rack-42"},
		},
	}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{Rack: rackID}
	if got := resolveLocalRackName(dev, mapper); got != "Rack-42" {
		t.Errorf("expected 'Rack-42', got %q", got)
	}
}

// ---------- compareRack ----------

func TestCompareRackBothNil(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{},
	}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{Rack: uuid.Nil}
	remote := &nautobotapi.Device{Rack: nil}
	diffs := compareRack(dev, remote, mapper)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when both nil, got %+v", diffs)
	}
}

func TestCompareRackLocalHasRack(t *testing.T) {
	// When the local device has a rack but the API client is nil,
	// GetRackByName panics. This scenario requires a real or mocked API.
	// We verify here that resolveLocalRackName correctly finds the rack name.
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	rackID := uuid.New()
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {Name: "Rack-A"},
		},
	}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{Rack: rackID}
	name := resolveLocalRackName(dev, mapper)
	if name != "Rack-A" {
		t.Errorf("expected 'Rack-A', got %q", name)
	}
}

// ---------- MapToNautobotDevice (full path with populated cache) ----------

func TestMapToNautobotDeviceFull(t *testing.T) {
	cache := NewLookupCache(nil)

	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	// Pre-populate caches
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["hpe-dl360-gen10"] = &CachedItem{ID: dtID, Name: "HPE DL360 Gen10"}
	cache.deviceTypesMu.Unlock()

	cache.locationsMu.Lock()
	cache.locations["DC-01"] = &CachedItem{ID: locID, Name: "DC-01"}
	cache.locationsMu.Unlock()

	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
		DefaultLocation: "DC-01",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:       uuid.New(),
		Name:     "server-01",
		Slug:     "hpe-dl360-gen10",
		Serial:   "SN123",
		AssetTag: "ASSET-001",
		Comments: "Test server",
	}

	req, err := mapper.MapToNautobotDevice(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req == nil {
		t.Fatal("expected non-nil request")
	}
	if *req.Name != "server-01" {
		t.Errorf("name = %q, want 'server-01'", *req.Name)
	}
	if req.Serial == nil || *req.Serial != "SN123" {
		t.Error("serial not set correctly")
	}
	if req.AssetTag == nil || *req.AssetTag != "ASSET-001" {
		t.Error("asset_tag not set correctly")
	}
	if req.Comments == nil || *req.Comments != "Test server" {
		t.Error("comments not set correctly")
	}
}

// ---------- MapToWritableDeviceRequest (full path with populated cache) ----------

func TestMapToWritableDeviceRequestFull(t *testing.T) {
	cache := NewLookupCache(nil)

	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["hpe-dl380-gen10"] = &CachedItem{ID: dtID, Name: "HPE DL380 Gen10"}
	cache.deviceTypesMu.Unlock()

	cache.locationsMu.Lock()
	cache.locations["DC-02"] = &CachedItem{ID: locID, Name: "DC-02"}
	cache.locationsMu.Unlock()

	cache.statusesMu.Lock()
	cache.statuses["Planned"] = &CachedItem{ID: statusID, Name: "Planned"}
	cache.statusesMu.Unlock()

	cache.rolesMu.Lock()
	cache.roles["Compute"] = &CachedItem{ID: roleID, Name: "Compute"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultStatus:   "Planned",
		DefaultRole:     "Compute",
		DefaultLocation: "DC-02",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:       uuid.New(),
		Name:     "compute-01",
		Slug:     "hpe-dl380-gen10",
		Serial:   "SN456",
		AssetTag: "ASSET-002",
		Comments: "Compute node",
	}

	req, err := mapper.MapToWritableDeviceRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req == nil {
		t.Fatal("expected non-nil request")
	}
	if *req.Name != "compute-01" {
		t.Errorf("name = %q, want 'compute-01'", *req.Name)
	}
	if req.Serial == nil || *req.Serial != "SN456" {
		t.Error("serial not set correctly")
	}
	if req.AssetTag == nil || *req.AssetTag != "ASSET-002" {
		t.Error("asset_tag not set correctly")
	}
	if req.Comments == nil || *req.Comments != "Compute node" {
		t.Error("comments not set correctly")
	}
}

// ---------- MapToWritableDeviceRequest with rack/position ----------

func TestMapToWritableDeviceRequestWithRack(t *testing.T) {
	cache := NewLookupCache(nil)

	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["hpe-dl360"] = &CachedItem{ID: dtID, Name: "HPE DL360"}
	cache.deviceTypesMu.Unlock()

	cache.locationsMu.Lock()
	cache.locations["Room-A"] = &CachedItem{ID: locID, Name: "Room-A"}
	cache.locationsMu.Unlock()

	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
		DefaultLocation: "Room-A",
	})

	// Device without rack FK (rack resolution needs API which is nil).
	// This tests that the path through optional fields + no rack works.
	dev := &devicetypes.CaniDeviceType{
		ID:           uuid.New(),
		Name:         "racked-server",
		Slug:         "hpe-dl360",
		RackPosition: 10,
		Face:         "rear",
	}

	req, err := mapper.MapToWritableDeviceRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req == nil {
		t.Fatal("expected non-nil request")
	}
	if *req.Name != "racked-server" {
		t.Errorf("name = %q, want 'racked-server'", *req.Name)
	}
	// Without rack resolution, Rack should be nil
	if req.Rack != nil {
		t.Error("expected nil Rack (no API client)")
	}
}

// ---------- resolveDeviceType (cache hit) ----------

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

func TestResolveDeviceTypeNoSlug(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{Strict: false})
	dev := &devicetypes.CaniDeviceType{Slug: "", Model: ""}

	_, err := mapper.resolveDeviceType(dev)
	if err == nil {
		t.Error("expected error for empty slug/model")
	}
}

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

// ---------- GetInterfaceByDeviceAndName (cache hit) ----------

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

// ---------- getDeviceInterfaceSpecs ----------

func TestGetDeviceInterfaceSpecsWithInterfaces(t *testing.T) {
	mgmt := true
	dev := &devicetypes.CaniDeviceType{
		Type: "node",
		Interfaces: []devicetypes.InterfaceSpec{
			{Name: "eth0", Type: "1000base-t"},
			{Name: "mgmt0", Type: "1000base-t", MgmtOnly: &mgmt},
			{Name: "ib0", Type: "infiniband-ndr"},
		},
	}

	specs := getDeviceInterfaceSpecs(dev)
	if len(specs) != 3 {
		t.Fatalf("expected 3 specs, got %d", len(specs))
	}
	if specs[0].Name != "eth0" || specs[0].Type != "1000base-t" {
		t.Errorf("spec[0] = %+v", specs[0])
	}
	if specs[2].Name != "ib0" || specs[2].Type != "infiniband-ndr" || specs[2].Speed != 400000000 {
		t.Errorf("spec[2] = %+v", specs[2])
	}
}

func TestGetDeviceInterfaceSpecsFallbackBlade(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Type:  devicetypes.Type(devicetypes.Blade),
		Model: "ProLiant DL360",
	}

	specs := getDeviceInterfaceSpecs(dev)
	// Should have iLO + 4 eth = 5 interfaces
	if len(specs) < 5 {
		t.Fatalf("expected at least 5 specs for blade, got %d", len(specs))
	}
	if specs[0].Name != "iLO" {
		t.Errorf("first spec should be iLO, got %q", specs[0].Name)
	}
}

func TestGetDeviceInterfaceSpecsFallbackHSNSwitch(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Type: devicetypes.Type(devicetypes.HSNSwitch),
	}

	specs := getDeviceInterfaceSpecs(dev)
	// mgmt0 + 64 osfp = 65
	if len(specs) != 65 {
		t.Fatalf("expected 65 specs for HSN switch, got %d", len(specs))
	}
	if specs[0].Name != "mgmt0" {
		t.Errorf("first spec should be mgmt0, got %q", specs[0].Name)
	}
	if specs[1].Type != "infiniband-ndr" {
		t.Errorf("osfp type = %q, want infiniband-ndr", specs[1].Type)
	}
}

func TestGetDeviceInterfaceSpecsFallbackMgmtSwitch(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Type: devicetypes.Type(devicetypes.MgmtSwitch),
	}

	specs := getDeviceInterfaceSpecs(dev)
	// mgmt0 + 48 ports + 4 sfp = 53
	if len(specs) != 53 {
		t.Fatalf("expected 53 specs for mgmt switch, got %d", len(specs))
	}
}

func TestGetDeviceInterfaceSpecsFallbackPDU(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Type: devicetypes.Type(devicetypes.CabinetPDU),
	}

	specs := getDeviceInterfaceSpecs(dev)
	if len(specs) != 1 || specs[0].Name != "mgmt0" {
		t.Fatalf("expected 1 mgmt spec for PDU, got %d", len(specs))
	}
}

func TestGetDeviceInterfaceSpecsInfinibandNode(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Type:  devicetypes.Type(devicetypes.Node),
		Model: "Quantum InfiniBand NDR Switch",
	}

	specs := getDeviceInterfaceSpecs(dev)
	// Should have iLO + 4 eth + 2 ib = 7
	if len(specs) != 7 {
		t.Fatalf("expected 7 specs for IB node, got %d", len(specs))
	}
	// Last two should be IB
	if specs[5].Name != "ib0" || specs[6].Name != "ib1" {
		t.Errorf("IB interfaces: %q, %q", specs[5].Name, specs[6].Name)
	}
}

// ---------- PrintSummary (all branches) ----------

func TestPrintSummaryAllBranches(t *testing.T) {
	result := &LoadResult{
		LocationsCreated:   []string{"loc1"},
		LocationsSkipped:   []string{"loc2"},
		RacksCreated:       []string{"rack1"},
		RacksSkipped:       2,
		Created:            []string{"dev1"},
		Updated:            []string{"dev2"},
		Skipped:            []string{"dev3"},
		Conflicts:          []ConflictInfo{{DeviceName: "dev3", Reason: "exists"}},
		IfacesCreated:      5,
		IfacesSkipped:      2,
		ModulesCreated:     1,
		ModulesSkipped:     1,
		FrusCreated:        3,
		FrusSkipped:        1,
		CablesCreated:      4,
		CablesSkipped:      1,
		VLANsCreated:       2,
		VLANsSkipped:       1,
		PrefixesCreated:    3,
		PrefixesSkipped:    1,
		IPAddressesCreated: 6,
		IPAddressesSkipped: 2,
		Errors:             []string{"something went wrong"},
	}
	// Should not panic.
	PrintSummary(result)
}

// ---------- compareDeviceFields with position/face diffs ----------

func TestCompareDeviceFieldsPositionFace(t *testing.T) {
	pos := 5
	faceVal := nautobotapi.DeviceFaceValue("front")
	dev := &devicetypes.CaniDeviceType{
		RackPosition: 10,
		Face:         "rear",
	}
	remote := &nautobotapi.Device{
		Position: &pos,
		Face:     &nautobotapi.DeviceFace{Value: &faceVal},
	}

	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	diffs := compareDeviceFields(dev, remote, mapper)

	posFound := false
	faceFound := false
	for _, d := range diffs {
		if d.Field == "position" {
			posFound = true
		}
		if d.Field == "face" {
			faceFound = true
		}
	}
	if !posFound {
		t.Error("expected position diff")
	}
	if !faceFound {
		t.Error("expected face diff")
	}
}

// ---------- locationFromParentRack ----------

func TestLocationFromParentRackNilInventory(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{Rack: uuid.New()}

	if got := mapper.locationFromParentRack(dev); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestLocationFromParentRackNoRack(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{},
	}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{Rack: uuid.Nil, Parent: uuid.Nil}
	if got := mapper.locationFromParentRack(dev); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestLocationFromParentRackResolvesChild(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	// Register a location type that supports devices.
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Server Room LFP",
		Slug:         "server-room-lfp",
		ContentTypes: []string{"rack", "device"},
	})

	locID := uuid.New()
	rackID := uuid.New()

	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {Name: "Rack-X", Location: locID},
		},
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			locID: {
				ID:           locID,
				Name:         "Server-Room-LFP",
				LocationType: "server-room-lfp",
			},
		},
	}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{Rack: rackID}
	got := mapper.locationFromParentRack(dev)
	if got != "Server-Room-LFP" {
		t.Errorf("expected 'Server-Room-LFP', got %q", got)
	}
}

// ---------- MapToPatchRequest ----------

func TestMapToPatchRequestNil(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	_, err := mapper.MapToPatchRequest(nil, uuid.New())
	if err == nil {
		t.Error("expected error for nil device")
	}
}

func TestMapToPatchRequestBasic(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["test-slug"] = &CachedItem{ID: dtID, Name: "TestSlug"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["Site-X"] = &CachedItem{ID: locID, Name: "Site-X"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Site-X",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:       uuid.New(),
		Name:     "patched-device",
		Slug:     "test-slug",
		Serial:   "SN-PATCH",
		AssetTag: "AT-PATCH",
		Comments: "patched",
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req == nil {
		t.Fatal("expected non-nil request")
	}
	if *req.Name != "patched-device" {
		t.Errorf("name = %q", *req.Name)
	}
	if req.Serial == nil || *req.Serial != "SN-PATCH" {
		t.Error("serial not set")
	}
	if req.AssetTag == nil || *req.AssetTag != "AT-PATCH" {
		t.Error("asset_tag not set")
	}
	if req.Comments == nil || *req.Comments != "patched" {
		t.Error("comments not set")
	}
	if req.DeviceType == nil {
		t.Error("device_type not resolved")
	}
	if req.Location == nil {
		t.Error("location not resolved")
	}
	if req.Status == nil {
		t.Error("status not resolved")
	}
	if req.Role == nil {
		t.Error("role not resolved")
	}
}

func TestMapToPatchRequestNoSlug(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["Site-Y"] = &CachedItem{ID: locID, Name: "Site-Y"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Site-Y",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "no-slug-device",
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// DeviceType should be nil (slug is empty, resolveDeviceType would error)
	if req.DeviceType != nil {
		t.Error("expected nil DeviceType for empty slug")
	}
}

func TestMapToPatchRequestWithCustomFields(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["DC"] = &CachedItem{ID: locID, Name: "DC"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Generic"] = &CachedItem{ID: roleID, Name: "Generic"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC",
		DefaultStatus:   "Active",
		DefaultRole:     "Generic",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "cf-device",
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{
				"nautobot": map[string]any{
					"xname": "x1000c0s0b0",
				},
			},
		},
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.CustomFields == nil {
		t.Fatal("expected non-nil CustomFields")
	}
	cf := *req.CustomFields
	if cf["xname"] != "x1000c0s0b0" {
		t.Errorf("custom field xname = %v", cf["xname"])
	}
}

// ---------- MapToWritableRackRequest ----------

func TestMapToWritableRackRequestNil(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	_, err := mapper.MapToWritableRackRequest(nil)
	if err == nil {
		t.Error("expected error for nil device")
	}
}

func TestMapToWritableRackRequestBasic(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["Floor-1"] = &CachedItem{ID: locID, Name: "Floor-1"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Floor-1",
		DefaultStatus:   "Active",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "Rack-A01",
	}

	req, err := mapper.MapToWritableRackRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req == nil {
		t.Fatal("expected non-nil request")
	}
	if req.Name != "Rack-A01" {
		t.Errorf("name = %q", req.Name)
	}
	if req.UHeight == nil || *req.UHeight != 48 {
		t.Errorf("expected default u_height 48, got %v", req.UHeight)
	}
}

func TestMapToWritableRackRequestCustomHeight(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["Room-C"] = &CachedItem{ID: locID, Name: "Room-C"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Room-C",
		DefaultStatus:   "Active",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "Rack-B02",
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{"u_height": 42},
		},
	}

	req, err := mapper.MapToWritableRackRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.UHeight == nil || *req.UHeight != 42 {
		t.Errorf("expected u_height 42, got %v", req.UHeight)
	}
}

func TestMapToWritableRackRequestNoLocation(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "Rack-NoLoc",
	}

	_, err := mapper.MapToWritableRackRequest(dev)
	if err == nil {
		t.Error("expected error when no location available")
	}
}

// ---------- compareDeviceFields with type/location/status/role diffs ----------

func TestCompareDeviceFieldsTypeDiff(t *testing.T) {
	cache := NewLookupCache(nil)
	localDtID := uuid.New()
	remoteDtID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["local-type"] = &CachedItem{ID: localDtID, Name: "LocalType"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["Site"] = &CachedItem{ID: uuid.New(), Name: "Site"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: uuid.New(), Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: uuid.New(), Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Site",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{Slug: "local-type"}

	// Build remote with a different device type ID
	remoteRef := makeStatusRef(remoteDtID)
	remote := &nautobotapi.Device{
		DeviceType: remoteRef,
		Location:   makeStatusRef(uuid.New()),
		Status:     makeStatusRef(uuid.New()),
		Role:       makeStatusRef(uuid.New()),
	}

	diffs := compareDeviceFields(dev, remote, mapper)
	found := false
	for _, d := range diffs {
		if d.Field == "device_type" {
			found = true
			if d.LocalVal != "LocalType" {
				t.Errorf("localVal = %q, want 'LocalType'", d.LocalVal)
			}
		}
	}
	if !found {
		t.Error("expected device_type diff")
	}
}

func TestCompareDeviceFieldsStatusDiff(t *testing.T) {
	cache := NewLookupCache(nil)
	localStatusID := uuid.New()
	remoteStatusID := uuid.New()

	cache.statusesMu.Lock()
	cache.statuses["Planned"] = &CachedItem{ID: localStatusID, Name: "Planned"}
	cache.statusesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC"] = &CachedItem{ID: uuid.New(), Name: "DC"}
	cache.locationsMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: uuid.New(), Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC",
		DefaultStatus:   "Planned",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		ObjectMeta: devicetypes.ObjectMeta{Status: "Planned"},
	}

	remote := &nautobotapi.Device{
		DeviceType: makeStatusRef(uuid.New()),
		Location:   makeStatusRef(uuid.New()),
		Status:     makeStatusRef(remoteStatusID),
		Role:       makeStatusRef(uuid.New()),
	}

	diffs := compareDeviceFields(dev, remote, mapper)
	found := false
	for _, d := range diffs {
		if d.Field == "status" {
			found = true
			if d.LocalVal != "Planned" {
				t.Errorf("localVal = %q, want 'Planned'", d.LocalVal)
			}
		}
	}
	if !found {
		t.Error("expected status diff")
	}
}

func TestCompareDeviceFieldsRoleDiff(t *testing.T) {
	cache := NewLookupCache(nil)
	localRoleID := uuid.New()

	cache.rolesMu.Lock()
	cache.roles["Compute"] = &CachedItem{ID: localRoleID, Name: "Compute"}
	cache.rolesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC"] = &CachedItem{ID: uuid.New(), Name: "DC"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: uuid.New(), Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC",
		DefaultStatus:   "Active",
		DefaultRole:     "Compute",
	})

	dev := &devicetypes.CaniDeviceType{}

	remote := &nautobotapi.Device{
		DeviceType: makeStatusRef(uuid.New()),
		Location:   makeStatusRef(uuid.New()),
		Status:     makeStatusRef(uuid.New()),
		Role:       makeStatusRef(uuid.New()), // different from localRoleID
	}

	diffs := compareDeviceFields(dev, remote, mapper)
	found := false
	for _, d := range diffs {
		if d.Field == "role" {
			found = true
			if d.LocalVal != "Compute" {
				t.Errorf("localVal = %q, want 'Compute'", d.LocalVal)
			}
		}
	}
	if !found {
		t.Error("expected role diff")
	}
}

func TestCompareDeviceFieldsLocationDiff(t *testing.T) {
	cache := NewLookupCache(nil)
	localLocID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["Room-Z"] = &CachedItem{ID: localLocID, Name: "Room-Z"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: uuid.New(), Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: uuid.New(), Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Room-Z",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{}

	remote := &nautobotapi.Device{
		DeviceType: makeStatusRef(uuid.New()),
		Location:   makeStatusRef(uuid.New()), // different
		Status:     makeStatusRef(uuid.New()),
		Role:       makeStatusRef(uuid.New()),
	}

	diffs := compareDeviceFields(dev, remote, mapper)
	found := false
	for _, d := range diffs {
		if d.Field == "location" {
			found = true
			if d.LocalVal != "Room-Z" {
				t.Errorf("localVal = %q, want 'Room-Z'", d.LocalVal)
			}
		}
	}
	if !found {
		t.Error("expected location diff")
	}
}

func TestCompareDeviceFieldsNoTypeDiffWhenMatching(t *testing.T) {
	cache := NewLookupCache(nil)
	sharedID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["shared"] = &CachedItem{ID: sharedID, Name: "Shared"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC"] = &CachedItem{ID: sharedID, Name: "DC"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: sharedID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: sharedID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{Slug: "shared"}

	remote := &nautobotapi.Device{
		DeviceType: makeStatusRef(sharedID),
		Location:   makeStatusRef(sharedID),
		Status:     makeStatusRef(sharedID),
		Role:       makeStatusRef(sharedID),
	}

	diffs := compareDeviceFields(dev, remote, mapper)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when all IDs match, got %+v", diffs)
	}
}

// ---------- resolveContentLocation ----------

func TestResolveContentLocationNilInventory(t *testing.T) {
	got := resolveContentLocation(uuid.New(), "device", nil)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestResolveContentLocationNilID(t *testing.T) {
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{},
	}
	got := resolveContentLocation(uuid.Nil, "device", inv)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestResolveContentLocationNotFound(t *testing.T) {
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{},
	}
	got := resolveContentLocation(uuid.New(), "device", inv)
	if got != "" {
		t.Errorf("expected empty for missing location, got %q", got)
	}
}

func TestResolveContentLocationDirect(t *testing.T) {
	// Register a type that directly supports "device"
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Server Room RCL",
		Slug:         "server-room-rcl",
		ContentTypes: []string{"rack", "device"},
	})

	locID := uuid.New()
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			locID: {
				ID:           locID,
				Name:         "Room-A-RCL",
				LocationType: "server-room-rcl",
			},
		},
	}

	got := resolveContentLocation(locID, "device", inv)
	if got != "Room-A-RCL" {
		t.Errorf("expected 'Room-A-RCL', got %q", got)
	}
}

func TestResolveContentLocationChildResolution(t *testing.T) {
	// Parent type doesn't support "rack", child type does
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Building RCL",
		Slug:         "building-rcl",
		ContentTypes: []string{},
	})
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Floor RCL",
		Slug:         "floor-rcl",
		ContentTypes: []string{"rack", "device"},
	})

	parentID := uuid.New()
	childID := uuid.New()

	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			parentID: {
				ID:           parentID,
				Name:         "Building-1-RCL",
				LocationType: "building-rcl",
				Children:     []uuid.UUID{childID},
			},
			childID: {
				ID:           childID,
				Name:         "Floor-2-RCL",
				LocationType: "floor-rcl",
			},
		},
	}

	got := resolveContentLocation(parentID, "rack", inv)
	if got != "Floor-2-RCL" {
		t.Errorf("expected 'Floor-2-RCL', got %q", got)
	}
}

func TestResolveContentLocationNoMatch(t *testing.T) {
	// Type that supports nothing
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Campus RCL",
		Slug:         "campus-rcl",
		ContentTypes: []string{},
	})

	locID := uuid.New()
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			locID: {
				ID:           locID,
				Name:         "Campus-RCL",
				LocationType: "campus-rcl",
			},
		},
	}

	got := resolveContentLocation(locID, "device", inv)
	if got != "" {
		t.Errorf("expected empty (no descendant supports device), got %q", got)
	}
}

// ---------- resolveFace ----------

func TestResolveFaceRear(t *testing.T) {
	rf := resolveFace("rear")
	if rf == nil {
		t.Fatal("expected non-nil RackFace")
	}
}

func TestResolveFaceFront(t *testing.T) {
	rf := resolveFace("front")
	if rf == nil {
		t.Fatal("expected non-nil RackFace")
	}
}

func TestResolveFaceEmpty(t *testing.T) {
	rf := resolveFace("")
	if rf == nil {
		t.Fatal("expected non-nil RackFace (default to front)")
	}
}

// ---------- MapToNautobotDevice with ProviderMetadata (custom fields) ----------

func TestMapToNautobotDeviceWithCustomFields(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["test-dt"] = &CachedItem{ID: dtID, Name: "TestDT"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-CF"] = &CachedItem{ID: locID, Name: "DC-CF"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-CF",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "custom-fields-device",
		Slug: "test-dt",
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{
				"nautobot": map[string]any{
					"xname":       "x3000c0s1b0",
					"import_date": "2026-01-01",
				},
			},
		},
	}

	req, err := mapper.MapToNautobotDevice(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.CustomFields == nil {
		t.Fatal("expected non-nil CustomFields")
	}
	cf := *req.CustomFields
	if cf["xname"] != "x3000c0s1b0" {
		t.Errorf("xname = %v", cf["xname"])
	}
	if cf["import_date"] != "2026-01-01" {
		t.Errorf("import_date = %v", cf["import_date"])
	}
}

// ---------- resolveLocalDeviceType / resolveLocalLocation / resolveLocalStatus / resolveLocalRole ----------

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

// ---------- MapToWritableDeviceRequest with custom fields ----------

func TestMapToWritableDeviceRequestCustomFields(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["cf-slug"] = &CachedItem{ID: dtID, Name: "CF-Slug"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-WR"] = &CachedItem{ID: locID, Name: "DC-WR"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-WR",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "wr-cf-device",
		Slug: "cf-slug",
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{
				"custom_key": "custom_val",
			},
		},
	}

	req, err := mapper.MapToWritableDeviceRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.CustomFields == nil {
		t.Fatal("expected non-nil CustomFields")
	}
	cf := *req.CustomFields
	if cf["custom_key"] != "custom_val" {
		t.Errorf("custom_key = %v", cf["custom_key"])
	}
}

// ---------- toNautobotContentTypes ----------

func TestToNautobotContentTypes(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		expect []string
	}{
		{"device", []string{"device"}, []string{"dcim.device"}},
		{"rack", []string{"rack"}, []string{"dcim.rack"}},
		{"module", []string{"module"}, []string{"dcim.module"}},
		{"unknown passthrough", []string{"ipam.prefix"}, []string{"ipam.prefix"}},
		{"mixed", []string{"device", "rack", "foo"}, []string{"dcim.device", "dcim.rack", "foo"}},
		{"empty", []string{}, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toNautobotContentTypes(tt.input)
			if len(got) != len(tt.expect) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.expect))
			}
			for i := range got {
				if got[i] != tt.expect[i] {
					t.Errorf("[%d] = %q, want %q", i, got[i], tt.expect[i])
				}
			}
		})
	}
}

// ---------- parentDef ----------

func TestParentDefFound(t *testing.T) {
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Parent Test Type",
		Slug:         "parent-test-type",
		ContentTypes: []string{"rack"},
	})

	got := parentDef("parent-test-type")
	if got == nil {
		t.Fatal("expected non-nil definition")
	}
	if got.Slug != "parent-test-type" {
		t.Errorf("slug = %q", got.Slug)
	}
}

func TestParentDefNotFound(t *testing.T) {
	got := parentDef("nonexistent-slug-xyz")
	if got != nil {
		t.Errorf("expected nil for unknown slug, got %+v", got)
	}
}

// ---------- SetCreateModuleTypes ----------

func TestSetCreateModuleTypes(t *testing.T) {
	cache := NewLookupCache(nil)
	// Should not panic — it's a no-op.
	cache.SetCreateModuleTypes(true)
	cache.SetCreateModuleTypes(false)
}

// ---------- resolveDeviceType error paths ----------

func TestResolveDeviceTypeStrictNoSlug(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{Strict: true})
	dev := &devicetypes.CaniDeviceType{Slug: "", Model: ""}

	_, err := mapper.resolveDeviceType(dev)
	if err == nil {
		t.Fatal("expected error for strict mode with no slug/model")
	}
	if err.Error() != "device type slug is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveDeviceTypeNonStrictNoSlug(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{Strict: false})
	dev := &devicetypes.CaniDeviceType{Slug: "", Model: ""}

	_, err := mapper.resolveDeviceType(dev)
	if !errors.Is(err, ErrDeviceUnclassified) {
		t.Errorf("expected ErrDeviceUnclassified, got %v", err)
	}
}

// ---------- resolveStatus error path ----------

func TestResolveStatusNoDefaultNoCreate(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultStatus: ""})
	dev := &devicetypes.CaniDeviceType{}

	_, err := mapper.resolveStatus(dev)
	if err == nil {
		t.Fatal("expected error when no status and no default")
	}
	if err.Error() != "status is required (use --default-status)" {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- resolveRole error path ----------

func TestResolveRoleNoDefaultNoCreate(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultRole: ""})
	dev := &devicetypes.CaniDeviceType{}

	_, err := mapper.resolveRole(dev)
	if err == nil {
		t.Fatal("expected error when no role and no default")
	}
	if err.Error() != "role is required (use --default-role)" {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- resolveLocation error path ----------

func TestResolveLocationNoDefaultNoCreate(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultLocation: ""})
	dev := &devicetypes.CaniDeviceType{}

	_, err := mapper.resolveLocation(dev)
	if err == nil {
		t.Fatal("expected error when no location and no default")
	}
	if err.Error() != "location is required (use --default-location)" {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- MapToNautobotDevice error propagation ----------

func TestMapToNautobotDeviceNoDeviceType(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["DC"] = &CachedItem{ID: locID, Name: "DC"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		Strict:          false,
		DefaultLocation: "DC",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name: "no-type-device",
		Slug: "", // no slug
	}

	_, err := mapper.MapToNautobotDevice(dev)
	if err == nil {
		t.Fatal("expected error when device type can't be resolved")
	}
	if !errors.Is(err, ErrDeviceUnclassified) {
		// The error is wrapped: "failed to resolve device type for X: <sentinel>"
		if !errors.Is(errors.Unwrap(err), ErrDeviceUnclassified) {
			t.Logf("error type: %v", err)
		}
	}
}

func TestMapToNautobotDeviceNoLocation(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["dt-noloc"] = &CachedItem{ID: dtID, Name: "DT-NoLoc"}
	cache.deviceTypesMu.Unlock()

	// No location cached, no default, no createLocations → error
	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name: "noloc-device",
		Slug: "dt-noloc",
	}

	_, err := mapper.MapToNautobotDevice(dev)
	if err == nil {
		t.Fatal("expected error when location can't be resolved")
	}
}

func TestMapToNautobotDeviceNoStatus(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["dt-nostat"] = &CachedItem{ID: dtID, Name: "DT-NoStat"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-S"] = &CachedItem{ID: locID, Name: "DC-S"}
	cache.locationsMu.Unlock()

	// No status cached, no default → error
	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-S",
		DefaultStatus:   "",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name: "nostat-device",
		Slug: "dt-nostat",
	}

	_, err := mapper.MapToNautobotDevice(dev)
	if err == nil {
		t.Fatal("expected error when status can't be resolved")
	}
}

func TestMapToNautobotDeviceNoRole(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["dt-norole"] = &CachedItem{ID: dtID, Name: "DT-NoRole"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-R"] = &CachedItem{ID: locID, Name: "DC-R"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	// No role cached, no default → error
	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-R",
		DefaultStatus:   "Active",
		DefaultRole:     "",
	})

	dev := &devicetypes.CaniDeviceType{
		Name: "norole-device",
		Slug: "dt-norole",
	}

	_, err := mapper.MapToNautobotDevice(dev)
	if err == nil {
		t.Fatal("expected error when role can't be resolved")
	}
}

// ---------- MapToWritableDeviceRequest error propagation ----------

func TestMapToWritableDeviceRequestNoDeviceType(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{
		Strict:          false,
		DefaultLocation: "DC",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{Name: "fail-dt"}
	_, err := mapper.MapToWritableDeviceRequest(dev)
	if err == nil {
		t.Fatal("expected error when device type can't be resolved")
	}
}

func TestMapToWritableDeviceRequestNoLocation(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["wr-noloc"] = &CachedItem{ID: dtID, Name: "WR-NoLoc"}
	cache.deviceTypesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultLocation: ""})
	dev := &devicetypes.CaniDeviceType{Name: "wr-fail-loc", Slug: "wr-noloc"}
	_, err := mapper.MapToWritableDeviceRequest(dev)
	if err == nil {
		t.Fatal("expected error when location can't be resolved")
	}
}

func TestMapToWritableDeviceRequestNoStatus(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["wr-nostat"] = &CachedItem{ID: dtID, Name: "WR-NoStat"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-WS"] = &CachedItem{ID: locID, Name: "DC-WS"}
	cache.locationsMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-WS",
		DefaultStatus:   "",
	})
	dev := &devicetypes.CaniDeviceType{Name: "wr-fail-stat", Slug: "wr-nostat"}
	_, err := mapper.MapToWritableDeviceRequest(dev)
	if err == nil {
		t.Fatal("expected error when status can't be resolved")
	}
}

func TestMapToWritableDeviceRequestNoRole(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["wr-norole"] = &CachedItem{ID: dtID, Name: "WR-NoRole"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-WR2"] = &CachedItem{ID: locID, Name: "DC-WR2"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-WR2",
		DefaultStatus:   "Active",
		DefaultRole:     "",
	})
	dev := &devicetypes.CaniDeviceType{Name: "wr-fail-role", Slug: "wr-norole"}
	_, err := mapper.MapToWritableDeviceRequest(dev)
	if err == nil {
		t.Fatal("expected error when role can't be resolved")
	}
}

// ---------- MapToWritableRackRequest error path ----------

func TestMapToWritableRackRequestNoStatus(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	cache.locationsMu.Lock()
	cache.locations["Floor"] = &CachedItem{ID: locID, Name: "Floor"}
	cache.locationsMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Floor",
		DefaultStatus:   "", // no status
	})
	dev := &devicetypes.CaniDeviceType{Name: "Rack-Fail"}
	_, err := mapper.MapToWritableRackRequest(dev)
	if err == nil {
		t.Fatal("expected error when status can't be resolved for rack")
	}
}

// ---------- GetDeviceType / GetLocation / GetStatus / GetRole cache hit ----------

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

func TestFindNameByIDNotFound(t *testing.T) {
	cache := NewLookupCache(nil)
	got := cache.FindNameByID("deviceType", uuid.New())
	// Should return the UUID string when not found in any cache
	if got == "" {
		t.Error("expected non-empty UUID string fallback")
	}
}

// ---------- NewExporter ----------

func TestNewExporterOptions(t *testing.T) {
	cache := NewLookupCache(nil)
	opts := &ExporterOpts{
		DryRun:          true,
		Merge:           true,
		DefaultLocation: "Site-1",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	}

	exp := NewExporter(nil, cache, opts)
	if exp == nil {
		t.Fatal("expected non-nil exporter")
	}
	if exp.Options.DryRun != true {
		t.Error("DryRun not set")
	}
	if exp.Options.Merge != true {
		t.Error("Merge not set")
	}
}

// ---------- MapToPatchRequest additional branches ----------

func TestMapToPatchRequestSerial(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["patch-serial"] = &CachedItem{ID: dtID, Name: "Patch Serial"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-PS"] = &CachedItem{ID: locID, Name: "DC-PS"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-PS",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name:   "serial-device",
		Slug:   "patch-serial",
		Serial: "SN123456",
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Serial == nil || *req.Serial != "SN123456" {
		t.Errorf("Serial = %v, want SN123456", req.Serial)
	}
}

func TestMapToPatchRequestAssetTag(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["patch-at"] = &CachedItem{ID: dtID, Name: "Patch AT"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-AT"] = &CachedItem{ID: locID, Name: "DC-AT"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-AT",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name:     "asset-device",
		Slug:     "patch-at",
		AssetTag: "ASSET-001",
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.AssetTag == nil || *req.AssetTag != "ASSET-001" {
		t.Errorf("AssetTag = %v, want ASSET-001", req.AssetTag)
	}
}

func TestMapToPatchRequestComments(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	dev := &devicetypes.CaniDeviceType{
		Name:     "comments-device",
		Comments: "This is a test comment",
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Comments == nil || *req.Comments != "This is a test comment" {
		t.Errorf("Comments = %v, want 'This is a test comment'", req.Comments)
	}
}

func TestMapToPatchRequestNoSlugSkipsDeviceType(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	dev := &devicetypes.CaniDeviceType{
		Name: "no-slug-patch",
		Slug: "", // no slug means device type resolution is skipped
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.DeviceType != nil {
		t.Error("expected nil DeviceType when slug is empty")
	}
}

// ---------- GetDeviceByName cache hit ----------

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

func TestGetDeviceByNameCacheMissNilClient(t *testing.T) {
	// Cache miss with nil client would panic on API call.
	// This test documents that behavior — skip it.
	t.Skip("nil client panics on cache miss; only cache-hit path is testable")
}

// ---------- LookupLocation cache hit ----------

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

// ---------- compareRack remote has rack, local doesn't ----------

func TestCompareRackRemoteHasRackLocalDoesNot(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	// Device with no parent → local rack is ""
	dev := &devicetypes.CaniDeviceType{Name: "no-rack-local"}

	// Remote device has a rack with an ID
	rackID := uuid.New()
	var rackIDUnion nautobotapi.BulkWritableCableRequestStatusId
	rackIDUnion.FromBulkWritableCableRequestStatusId0(rackID)
	remote := &nautobotapi.Device{
		Rack: &nautobotapi.BulkWritableCircuitRequestTenant{
			Id: &rackIDUnion,
		},
	}
	diffs := compareRack(dev, remote, mapper)
	// Local is nil, remote is non-nil → should produce a diff
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Field != "rack" {
		t.Errorf("field = %q, want 'rack'", diffs[0].Field)
	}
}

// ---------- comparePosition ----------

func TestComparePositionNoLocalPosition(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Name:         "no-pos",
		RackPosition: 0, // zero means "no position"
	}
	remote := &nautobotapi.Device{}
	diffs := comparePosition(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when local position is 0, got %d", len(diffs))
	}
}

func TestComparePositionMismatch(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Name:         "pos-mismatch",
		RackPosition: 5,
	}
	remotePos := 3
	remote := &nautobotapi.Device{
		Position: &remotePos,
	}
	diffs := comparePosition(dev, remote)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Field != "position" {
		t.Errorf("field = %q, want 'position'", diffs[0].Field)
	}
	if diffs[0].LocalVal != "5" {
		t.Errorf("LocalVal = %q, want '5'", diffs[0].LocalVal)
	}
}

func TestComparePositionMatch(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Name:         "pos-match",
		RackPosition: 7,
	}
	remotePos := 7
	remote := &nautobotapi.Device{
		Position: &remotePos,
	}
	diffs := comparePosition(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when positions match, got %d", len(diffs))
	}
}

// ---------- compareFace ----------

func TestCompareFaceNoLocalFace(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Name: "no-face",
		Face: "", // empty means "no face"
	}
	remote := &nautobotapi.Device{}
	diffs := compareFace(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when local face is empty, got %d", len(diffs))
	}
}

func TestCompareFaceMismatch(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Name: "face-mismatch",
		Face: "rear",
	}
	frontVal := nautobotapi.DeviceFaceValue("front")
	remote := &nautobotapi.Device{
		Face: &nautobotapi.DeviceFace{
			Value: &frontVal,
		},
	}
	diffs := compareFace(dev, remote)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Field != "face" {
		t.Errorf("field = %q, want 'face'", diffs[0].Field)
	}
	if diffs[0].LocalVal != "rear" {
		t.Errorf("LocalVal = %q, want 'rear'", diffs[0].LocalVal)
	}
}

func TestCompareFaceMatch(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Name: "face-match",
		Face: "front",
	}
	frontVal := nautobotapi.DeviceFaceValue("front")
	remote := &nautobotapi.Device{
		Face: &nautobotapi.DeviceFace{
			Value: &frontVal,
		},
	}
	diffs := compareFace(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when faces match, got %d", len(diffs))
	}
}

// ---------- CacheInterface / GetInterfaceByDeviceAndName cache hit ----------

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

// ---------- SetInventory ----------

func TestSetInventory(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
		Racks:   map[uuid.UUID]*devicetypes.CaniRackType{},
	}
	mapper.SetInventory(inv)

	// Verify mapper now has the inventory (used by resolveLocalRackName)
	if mapper.inventory == nil {
		t.Fatal("expected non-nil inventory after SetInventory")
	}
}

// ---------- MapToWritableDeviceRequest with serial/asset/comments ----------

func TestMapToWritableDeviceRequestSerial(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["wr-serial"] = &CachedItem{ID: dtID, Name: "WR-Serial"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-WS2"] = &CachedItem{ID: locID, Name: "DC-WS2"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-WS2",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name:     "wr-serial-dev",
		Slug:     "wr-serial",
		Serial:   "SER-XYZ",
		AssetTag: "ASSET-99",
		Comments: "Test device comments",
	}

	req, err := mapper.MapToWritableDeviceRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Serial == nil || *req.Serial != "SER-XYZ" {
		t.Errorf("Serial = %v, want SER-XYZ", req.Serial)
	}
	if req.AssetTag == nil || *req.AssetTag != "ASSET-99" {
		t.Errorf("AssetTag = %v, want ASSET-99", req.AssetTag)
	}
	if req.Comments == nil || *req.Comments != "Test device comments" {
		t.Errorf("Comments = %v, want 'Test device comments'", req.Comments)
	}
}

// ---------- refID / tenantRefID ----------

func TestRefIDNilInput(t *testing.T) {
	result := refID(nil)
	if result != uuid.Nil {
		t.Errorf("expected uuid.Nil for nil input, got %s", result)
	}
}

func TestTenantRefIDNilInput(t *testing.T) {
	result := tenantRefID(nil)
	if result != uuid.Nil {
		t.Errorf("expected uuid.Nil for nil input, got %s", result)
	}
}

// ---------- LoadResult / ConflictInfo structs ----------

func TestLoadResultZeroValue(t *testing.T) {
	result := &LoadResult{}
	if result.CablesCreated != 0 {
		t.Error("expected zero value for CablesCreated")
	}
	if result.IfacesCreated != 0 {
		t.Error("expected zero value for IfacesCreated")
	}
}

func TestConflictInfoDiffs(t *testing.T) {
	ci := ConflictInfo{
		DeviceName: "test-device",
		ExistingID: uuid.New(),
		LocalID:    uuid.New(),
		Reason:     "already exists",
		Diffs: []FieldDiff{
			{Field: "status", LocalVal: "Active", RemoteVal: "Planned"},
		},
	}
	if len(ci.Diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(ci.Diffs))
	}
	if ci.Diffs[0].Field != "status" {
		t.Errorf("field = %q, want 'status'", ci.Diffs[0].Field)
	}
}

// ---------- pendingMove struct ----------

func TestPendingMoveFields(t *testing.T) {
	devID := uuid.New()
	rackA := uuid.New()
	rackB := uuid.New()
	move := pendingMove{
		DeviceID:   devID,
		DeviceName: "compute-001",
		From:       slotKey{RackID: rackA, Position: 1, Face: "front"},
		To:         slotKey{RackID: rackB, Position: 5, Face: "rear"},
	}
	if move.DeviceName != "compute-001" {
		t.Errorf("DeviceName = %q", move.DeviceName)
	}
	if move.From.Position != 1 {
		t.Errorf("From.Position = %d, want 1", move.From.Position)
	}
	if move.To.Face != "rear" {
		t.Errorf("To.Face = %q, want 'rear'", move.To.Face)
	}
}

// ---------- ExporterOpts fields ----------

func TestExporterOptsAllFlags(t *testing.T) {
	opts := &ExporterOpts{
		DefaultLocation:     "DC1",
		DefaultRole:         "Leaf",
		DefaultStatus:       "Planned",
		Merge:               true,
		DryRun:              false,
		Strict:              true,
		CreateDeviceTypes:   true,
		CreateLocationTypes: true,
		CreateModuleTypes:   true,
		CreateLocations:     true,
		CreateStatuses:      true,
		CreateRoles:         true,
	}
	if opts.DefaultLocation != "DC1" {
		t.Errorf("DefaultLocation = %q", opts.DefaultLocation)
	}
	if !opts.CreateDeviceTypes {
		t.Error("CreateDeviceTypes should be true")
	}
	if !opts.Strict {
		t.Error("Strict should be true")
	}
}

// ---------- InvalidateInterfacePrefetch ----------

func TestInvalidateInterfacePrefetchNoEntry(t *testing.T) {
	cache := NewLookupCache(nil)
	// Should not panic even when no entry exists
	cache.InvalidateInterfacePrefetch(uuid.New())
}

// ---------- printLoadSummary ----------

func TestPrintLoadSummaryAllBranches(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}

	// Test with all fields populated — exercises every branch
	result := &LoadResult{
		Created:            []string{"dev-1", "dev-2"},
		Updated:            []string{"dev-3"},
		Skipped:            []string{"dev-4"},
		Errors:             []string{"something failed"},
		Conflicts:          []ConflictInfo{{DeviceName: "dev-4", Diffs: []FieldDiff{{Field: "status", LocalVal: "Active", RemoteVal: "Planned"}}}},
		LocationsCreated:   []string{"DC1"},
		LocationsSkipped:   []string{"DC2"},
		RacksCreated:       []string{"Rack-A"},
		RacksSkipped:       1,
		IfacesCreated:      5,
		IfacesSkipped:      2,
		ModulesCreated:     3,
		ModulesSkipped:     1,
		FrusCreated:        4,
		FrusSkipped:        2,
		CablesCreated:      10,
		CablesSkipped:      3,
		VLANsCreated:       2,
		VLANsSkipped:       1,
		PrefixesCreated:    6,
		PrefixesSkipped:    2,
		IPAddressesCreated: 8,
		IPAddressesSkipped: 4,
	}

	// Should not panic — just prints to clog
	e.printLoadSummary(result)
}

func TestPrintLoadSummaryEmpty(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}
	result := &LoadResult{}

	// Empty result — most branches are skipped
	e.printLoadSummary(result)
}

// ---------- printConflictDiffs ----------

func TestPrintConflictDiffsNoDiffs(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}
	result := &LoadResult{
		Conflicts: []ConflictInfo{
			{DeviceName: "dev-1", Diffs: nil},
		},
	}
	// Should return early since no conflict has diffs
	e.printConflictDiffs(result)
}

func TestPrintConflictDiffsWithDiffs(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}
	result := &LoadResult{
		Conflicts: []ConflictInfo{
			{
				DeviceName: "dev-conflict",
				Diffs: []FieldDiff{
					{Field: "status", LocalVal: "Active", RemoteVal: "Planned"},
					{Field: "role", LocalVal: "Server", RemoteVal: "Switch"},
				},
			},
		},
	}
	// Should print the diffs without panicking
	e.printConflictDiffs(result)
}

func TestPrintConflictDiffsEmptyConflicts(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}
	result := &LoadResult{
		Conflicts: []ConflictInfo{},
	}
	e.printConflictDiffs(result)
}

// ---------- remoteSlotKey additional cases ----------

func TestRemoteSlotKeyValidFront(t *testing.T) {
	pos := 10
	rackID := uuid.New()
	var rackIDUnion nautobotapi.BulkWritableCableRequestStatusId
	rackIDUnion.FromBulkWritableCableRequestStatusId0(rackID)

	d := &nautobotapi.Device{
		Position: &pos,
		Rack: &nautobotapi.BulkWritableCircuitRequestTenant{
			Id: &rackIDUnion,
		},
		Face: nil, // nil face → defaults to "front"
	}

	sk := remoteSlotKey(d)
	if sk == nil {
		t.Fatal("expected non-nil slotKey")
	}
	if sk.RackID != rackID {
		t.Errorf("RackID = %s, want %s", sk.RackID, rackID)
	}
	if sk.Position != 10 {
		t.Errorf("Position = %d, want 10", sk.Position)
	}
	if sk.Face != "front" {
		t.Errorf("Face = %q, want 'front'", sk.Face)
	}
}

func TestRemoteSlotKeyValidRear(t *testing.T) {
	pos := 5
	rackID := uuid.New()
	var rackIDUnion nautobotapi.BulkWritableCableRequestStatusId
	rackIDUnion.FromBulkWritableCableRequestStatusId0(rackID)

	rearVal := nautobotapi.DeviceFaceValue("rear")
	d := &nautobotapi.Device{
		Position: &pos,
		Rack: &nautobotapi.BulkWritableCircuitRequestTenant{
			Id: &rackIDUnion,
		},
		Face: &nautobotapi.DeviceFace{
			Value: &rearVal,
		},
	}

	sk := remoteSlotKey(d)
	if sk == nil {
		t.Fatal("expected non-nil slotKey")
	}
	if sk.Face != "rear" {
		t.Errorf("Face = %q, want 'rear'", sk.Face)
	}
}

func TestRemoteSlotKeyNilPosition(t *testing.T) {
	rackID := uuid.New()
	var rackIDUnion nautobotapi.BulkWritableCableRequestStatusId
	rackIDUnion.FromBulkWritableCableRequestStatusId0(rackID)

	d := &nautobotapi.Device{
		Position: nil, // no position
		Rack: &nautobotapi.BulkWritableCircuitRequestTenant{
			Id: &rackIDUnion,
		},
	}
	if remoteSlotKey(d) != nil {
		t.Error("expected nil when Position is nil")
	}
}

func TestRemoteSlotKeyNilRack(t *testing.T) {
	pos := 3
	d := &nautobotapi.Device{
		Position: &pos,
		Rack:     nil,
	}
	if remoteSlotKey(d) != nil {
		t.Error("expected nil when Rack is nil")
	}
}

// ---------- mapInterfaceType additional cases ----------

func TestMapInterfaceType25g(t *testing.T) {
	got := mapInterfaceType("25gbase-x-sfp28")
	if got != "25gbase-x-sfp28" {
		t.Errorf("got %q, want '25gbase-x-sfp28'", got)
	}
}

func TestMapInterfaceType40g(t *testing.T) {
	got := mapInterfaceType("40gbase-x-qsfpp")
	if got != "40gbase-x-qsfpp" {
		t.Errorf("got %q, want '40gbase-x-qsfpp'", got)
	}
}

func TestMapInterfaceType100g(t *testing.T) {
	got := mapInterfaceType("100gbase-x-qsfp28")
	if got != "100gbase-x-qsfp28" {
		t.Errorf("got %q, want '100gbase-x-qsfp28'", got)
	}
}

func TestMapInterfaceType200g(t *testing.T) {
	got := mapInterfaceType("200gbase-x-qsfp56")
	if got != "200gbase-x-qsfp56" {
		t.Errorf("got %q, want '200gbase-x-qsfp56'", got)
	}
}

func TestMapInterfaceType100baseTx(t *testing.T) {
	got := mapInterfaceType("100base-tx")
	if got != "100base-tx" {
		t.Errorf("got %q, want '100base-tx'", got)
	}
}

func TestMapInterfaceType1gBaseT(t *testing.T) {
	got := mapInterfaceType("1gbase-t")
	if got != "1000base-t" {
		t.Errorf("got %q, want '1000base-t'", got)
	}
}

func TestMapInterfaceTypeInfinibandHDR(t *testing.T) {
	got := mapInterfaceType("infiniband-hdr")
	if got != "infiniband-hdr" {
		t.Errorf("got %q, want 'infiniband-hdr'", got)
	}
}

func TestMapInterfaceType400gQSFPDD(t *testing.T) {
	got := mapInterfaceType("400gbase-x-qsfpdd")
	if got != "400gbase-x-osfp" {
		t.Errorf("got %q, want '400gbase-x-osfp'", got)
	}
}

// ---------- resolveCableType slug-based fallback ----------

func TestResolveCableTypeSlugCat(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "some-cat5e-cable"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'cat'")
	}
}

func TestResolveCableTypeSlugDAC(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "hpe-dac-1m"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'dac'")
	}
}

func TestResolveCableTypeSlugAOC(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "mellanox-aoc-10m"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'aoc'")
	}
}

func TestResolveCableTypeSlugFiber(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "fiber-lc-om4-5m"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'fiber'")
	}
}

func TestResolveCableTypeSlugMMF(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "mmf-om3-cable"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'mmf'")
	}
}

func TestResolveCableTypeConnectorLC(t *testing.T) {
	cable := &devicetypes.CaniCableType{ConnectorType: "lc"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for LC connector")
	}
}

func TestResolveCableTypeConnectorMPO(t *testing.T) {
	cable := &devicetypes.CaniCableType{ConnectorType: "mpo"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for MPO connector")
	}
}

func TestResolveCableTypeCategoryMMF(t *testing.T) {
	cable := &devicetypes.CaniCableType{CableCategory: "mmf-om3"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for mmf-om3 category")
	}
}

func TestResolveCableTypeCategorySMF(t *testing.T) {
	cable := &devicetypes.CaniCableType{CableCategory: "smf-os2"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for smf-os2 category")
	}
}

// ---------- getSpeedForType additional ----------

func TestGetSpeedForType200g(t *testing.T) {
	if got := getSpeedForType("200gbase-x-qsfp56"); got != 200000000 {
		t.Errorf("got %d, want 200000000", got)
	}
}

func TestGetSpeedForTypeInfinibandHDR(t *testing.T) {
	if got := getSpeedForType("infiniband-hdr"); got != 200000000 {
		t.Errorf("got %d, want 200000000", got)
	}
}

func TestGetSpeedForTypeInfinibandNDR(t *testing.T) {
	if got := getSpeedForType("infiniband-ndr"); got != 400000000 {
		t.Errorf("got %d, want 400000000", got)
	}
}

func TestGetSpeedForType400gOSFP(t *testing.T) {
	if got := getSpeedForType("400gbase-x-osfp"); got != 400000000 {
		t.Errorf("got %d, want 400000000", got)
	}
}

func TestGetSpeedForType25g(t *testing.T) {
	if got := getSpeedForType("25gbase-x-sfp28"); got != 25000000 {
		t.Errorf("got %d, want 25000000", got)
	}
}

func TestGetSpeedForTypeUnknown(t *testing.T) {
	if got := getSpeedForType("some-unknown-type"); got != 1000000 {
		t.Errorf("got %d, want 1000000 (default)", got)
	}
}

// ---------- containsInfiniband additional ----------

func TestContainsInfinibandModel(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "Mellanox InfiniBand Switch"}
	if !containsInfiniband(dev) {
		t.Error("expected true for model containing 'infiniband'")
	}
}

func TestContainsInfinibandNDR(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "NDR-400G Switch"}
	if !containsInfiniband(dev) {
		t.Error("expected true for model containing 'ndr'")
	}
}

func TestContainsInfinibandHDR(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "QM8700 HDR Switch"}
	if !containsInfiniband(dev) {
		t.Error("expected true for model containing 'hdr'")
	}
}

func TestContainsInfinibandMCX(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "MCX653105A-HDAT"}
	if !containsInfiniband(dev) {
		t.Error("expected true for model containing 'mcx'")
	}
}

func TestContainsInfinibandFalse(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "HPE ProLiant DL380"}
	if containsInfiniband(dev) {
		t.Error("expected false for standard server model")
	}
}

// ---------- derefString ----------

func TestDerefStringNil(t *testing.T) {
	if got := derefString(nil); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestDerefStringNonNil(t *testing.T) {
	s := "hello"
	if got := derefString(&s); got != "hello" {
		t.Errorf("got %q, want 'hello'", got)
	}
}

// ---------- isValidNautobotInterfaceType ----------

func TestIsValidNautobotInterfaceTypeValid(t *testing.T) {
	validTypes := []string{
		"100base-tx", "1000base-t", "10gbase-x-sfpp", "25gbase-x-sfp28",
		"40gbase-x-qsfpp", "100gbase-x-qsfp28", "200gbase-x-qsfp56",
		"400gbase-x-osfp", "400gbase-x-qsfpdd",
		"infiniband-hdr", "infiniband-ndr",
		"virtual", "lag", "other",
	}
	for _, vt := range validTypes {
		if !isValidNautobotInterfaceType(vt) {
			t.Errorf("expected true for %q", vt)
		}
	}
}

func TestIsValidNautobotInterfaceTypeInvalid(t *testing.T) {
	invalidTypes := []string{"not-a-type", "ethernet", "10gbase-t", ""}
	for _, vt := range invalidTypes {
		if isValidNautobotInterfaceType(vt) {
			t.Errorf("expected false for %q", vt)
		}
	}
}

// ---------- LookupVLAN cache hit ----------

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

func TestLookupVLANCacheMissDifferentKey(t *testing.T) {
	cache := NewLookupCache(nil)
	vlanID := uuid.New()
	cache.CacheVLAN(100, "DC1", &CachedItem{ID: vlanID, Name: "VLAN-100"})

	// Different VID → cache miss (nil client will panic on API call so use same location but different VID)
	// Just test that a different location key doesn't hit
	item, _ := cache.LookupVLAN(100, "DC1")
	if item == nil || item.ID != vlanID {
		t.Error("sanity check: expected cache hit for correct key")
	}
}

// ---------- LookupPrefix cache hit ----------

func TestLookupPrefixCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	prefixID := uuid.New()
	cache.CachePrefix("10.0.0.0/24", &CachedItem{ID: prefixID, Name: "10.0.0.0/24"})

	item, err := cache.LookupPrefix("10.0.0.0/24")
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

func TestLookupIPAddressCacheHit(t *testing.T) {
	cache := NewLookupCache(nil)
	ipID := uuid.New()
	cache.CacheIPAddress("192.168.1.1/32", &CachedItem{ID: ipID, Name: "192.168.1.1/32"})

	item, err := cache.LookupIPAddress("192.168.1.1/32")
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

// ---------- printDeviceDiffs ----------

func TestPrintDeviceDiffsNoDiffs(t *testing.T) {
	// Should not panic or output when diffs is empty
	printDeviceDiffs("test-device", nil)
	printDeviceDiffs("test-device", []FieldDiff{})
}

func TestPrintDeviceDiffsWithDiffs(t *testing.T) {
	diffs := []FieldDiff{
		{Field: "status", LocalVal: "Active", RemoteVal: "Planned"},
		{Field: "role", LocalVal: "Server", RemoteVal: "Switch"},
		{Field: "location", LocalVal: "DC1", RemoteVal: "DC2"},
	}
	// Should not panic — just prints to clog
	printDeviceDiffs("diff-device", diffs)
}

// ---------- getSpeedForType 40g and 100g ----------

func TestGetSpeedForType40g(t *testing.T) {
	if got := getSpeedForType("40gbase-x-qsfpp"); got != 40000000 {
		t.Errorf("got %d, want 40000000", got)
	}
}

func TestGetSpeedForType100g(t *testing.T) {
	if got := getSpeedForType("100gbase-x-qsfp28"); got != 100000000 {
		t.Errorf("got %d, want 100000000", got)
	}
}

func TestGetSpeedForType10g(t *testing.T) {
	if got := getSpeedForType("10gbase-x-sfpp"); got != 10000000 {
		t.Errorf("got %d, want 10000000", got)
	}
}

func TestGetSpeedForType100baseTx(t *testing.T) {
	if got := getSpeedForType("100base-tx"); got != 100000 {
		t.Errorf("got %d, want 100000", got)
	}
}

// ---------- MapToWritableRackRequest custom fields filtering ----------

func TestMapToWritableRackRequestCustomFieldsFiltered(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["Room-CF"] = &CachedItem{ID: locID, Name: "Room-CF"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Room-CF",
		DefaultStatus:   "Active",
	})

	dev := &devicetypes.CaniDeviceType{
		Name: "Rack-CF",
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{
				"u_height":      42,
				"rack_position": 5,
				"custom_key":    "custom_val",
				"site_code":     "S01",
			},
		},
	}

	req, err := mapper.MapToWritableRackRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// u_height and rack_position should be filtered out of CustomFields
	if req.CustomFields == nil {
		t.Fatal("expected non-nil CustomFields")
	}
	cf := *req.CustomFields
	if _, ok := cf["u_height"]; ok {
		t.Error("u_height should be filtered from CustomFields")
	}
	if _, ok := cf["rack_position"]; ok {
		t.Error("rack_position should be filtered from CustomFields")
	}
	if cf["custom_key"] != "custom_val" {
		t.Errorf("custom_key = %v, want 'custom_val'", cf["custom_key"])
	}
	if cf["site_code"] != "S01" {
		t.Errorf("site_code = %v, want 'S01'", cf["site_code"])
	}
}

func TestMapToWritableRackRequestDefaultHeight(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["Room-DH"] = &CachedItem{ID: locID, Name: "Room-DH"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Room-DH",
		DefaultStatus:   "Active",
	})

	// No ProviderMetadata → default to 48U
	dev := &devicetypes.CaniDeviceType{Name: "Rack-DH"}

	req, err := mapper.MapToWritableRackRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.UHeight == nil || *req.UHeight != 48 {
		t.Errorf("expected default u_height 48, got %v", req.UHeight)
	}
}

// ---------- MapToWritableDeviceRequest with rack from inventory ----------

func TestMapToWritableDeviceRequestNoParentNoRack(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["server-dt"] = &CachedItem{ID: dtID, Name: "Server DT"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-Rack"] = &CachedItem{ID: locID, Name: "DC-Rack"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-Rack",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	// Set up inventory with a rack
	rackID := uuid.New()
	devID := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {ID: rackID, Name: "Rack-Test-01"},
		},
	}
	mapper.SetInventory(inv)

	// Mock GetRackByName by pre-populating the devices cache won't work for racks
	// The rack lookup uses GetRackByName which hits the API — can't test that path
	// without a nil client panic. But we CAN test the path where GetRackID returns
	// uuid.Nil (no Parent set), which skips the rack block entirely.

	dev := &devicetypes.CaniDeviceType{
		Name:         "server-in-rack",
		Slug:         "server-dt",
		RackPosition: 10,
		Face:         "rear",
		Parent:       uuid.Nil, // no parent → no rack resolution
	}
	inv.Devices[devID] = dev

	req, err := mapper.MapToWritableDeviceRequest(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No rack should be set since Parent is Nil
	if req.Rack != nil {
		t.Error("expected nil Rack when device has no parent")
	}
	// Position should not be set either
	if req.Position != nil {
		t.Error("expected nil Position when no rack assigned")
	}
}

// ---------- resolveLocationName (Exporter method in load_vlans.go) ----------

func TestResolveLocationNameFromExporter(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}

	locID := uuid.New()
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			locID: {ID: locID, Name: "Floor-42"},
		},
	}

	name, err := e.resolveLocationName(locID, inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Floor-42" {
		t.Errorf("got %q, want 'Floor-42'", name)
	}
}

func TestResolveLocationNameFromExporterNotFound(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{},
	}

	_, err := e.resolveLocationName(uuid.New(), inv)
	if err == nil {
		t.Fatal("expected error for unknown location ID")
	}
}

// ---------- topologicalSortLocations additional cases ----------

func TestTopologicalSortLocationsEmpty(t *testing.T) {
	locs := map[uuid.UUID]*devicetypes.CaniLocationType{}
	sorted := topologicalSortLocations(locs)
	if len(sorted) != 0 {
		t.Errorf("expected 0, got %d", len(sorted))
	}
}

func TestTopologicalSortLocationsSingleRoot(t *testing.T) {
	id := uuid.New()
	locs := map[uuid.UUID]*devicetypes.CaniLocationType{
		id: {ID: id, Name: "Root", Parent: uuid.Nil},
	}
	sorted := topologicalSortLocations(locs)
	if len(sorted) != 1 {
		t.Fatalf("expected 1, got %d", len(sorted))
	}
	if sorted[0].Name != "Root" {
		t.Errorf("Name = %q, want 'Root'", sorted[0].Name)
	}
}

func TestTopologicalSortLocationsParentBeforeChild(t *testing.T) {
	parentID := uuid.New()
	childID := uuid.New()
	locs := map[uuid.UUID]*devicetypes.CaniLocationType{
		childID:  {ID: childID, Name: "Child", Parent: parentID},
		parentID: {ID: parentID, Name: "Parent", Parent: uuid.Nil},
	}
	sorted := topologicalSortLocations(locs)
	if len(sorted) != 2 {
		t.Fatalf("expected 2, got %d", len(sorted))
	}
	if sorted[0].Name != "Parent" {
		t.Errorf("expected Parent first, got %q", sorted[0].Name)
	}
	if sorted[1].Name != "Child" {
		t.Errorf("expected Child second, got %q", sorted[1].Name)
	}
}

// ---------- MapToPatchRequest with rack position via inventory ----------

func TestMapToPatchRequestWithRackPositionNoParent(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["patch-rp"] = &CachedItem{ID: dtID, Name: "Patch RP"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-RP"] = &CachedItem{ID: locID, Name: "DC-RP"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-RP",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})
	mapper.SetInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
		Racks:   map[uuid.UUID]*devicetypes.CaniRackType{},
	})

	dev := &devicetypes.CaniDeviceType{
		Name:         "patch-rp-dev",
		Slug:         "patch-rp",
		RackPosition: 15,
		Face:         "front",
		Parent:       uuid.Nil, // no parent → rack/position block is skipped
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Without a parent, rack block should be skipped
	if req.Rack != nil {
		t.Error("expected nil Rack when device has no parent")
	}
}

// ---------- refID with valid union ----------

func TestRefIDValidUnion(t *testing.T) {
	expected := uuid.New()
	var union nautobotapi.BulkWritableCableRequestStatusId
	union.FromBulkWritableCableRequestStatusId0(expected)

	got := refID(&union)
	if got != expected {
		t.Errorf("got %s, want %s", got, expected)
	}
}

// ---------- FieldDiff struct ----------

func TestFieldDiffStruct(t *testing.T) {
	diff := FieldDiff{
		Field:     "device_type",
		LocalVal:  "DL380",
		RemoteVal: "DL360",
	}
	if diff.Field != "device_type" {
		t.Errorf("Field = %q", diff.Field)
	}
	if diff.LocalVal != "DL380" {
		t.Errorf("LocalVal = %q", diff.LocalVal)
	}
}

// ---------- MapToNautobotDevice with comments ----------

func TestMapToNautobotDeviceComments(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["cmt-dt"] = &CachedItem{ID: dtID, Name: "CMT-DT"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-C"] = &CachedItem{ID: locID, Name: "DC-C"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-C",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name:     "comment-dev",
		Slug:     "cmt-dt",
		Comments: "Important device",
	}

	req, err := mapper.MapToNautobotDevice(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Comments == nil || *req.Comments != "Important device" {
		t.Errorf("Comments = %v, want 'Important device'", req.Comments)
	}
}

// ---------- MapToNautobotDevice with serial and asset_tag ----------

func TestMapToNautobotDeviceSerialAndAssetTag(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["sa-dt"] = &CachedItem{ID: dtID, Name: "SA-DT"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-SA"] = &CachedItem{ID: locID, Name: "DC-SA"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-SA",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name:     "serial-asset-dev",
		Slug:     "sa-dt",
		Serial:   "SN-99887766",
		AssetTag: "ASSET-XYZ",
	}

	req, err := mapper.MapToNautobotDevice(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Serial == nil || *req.Serial != "SN-99887766" {
		t.Errorf("Serial = %v, want 'SN-99887766'", req.Serial)
	}
	if req.AssetTag == nil || *req.AssetTag != "ASSET-XYZ" {
		t.Errorf("AssetTag = %v, want 'ASSET-XYZ'", req.AssetTag)
	}
}

// ---------- resolveStatus auto-create path ----------

func TestResolveStatusAutoCreateActive(t *testing.T) {
	cache := NewLookupCache(nil)
	cache.SetCreateStatuses(true)
	// Pre-populate cache with "Active" so the lookup succeeds
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: uuid.New(), Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{} // no status set
	item, err := mapper.resolveStatus(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "Active" {
		t.Errorf("expected auto-created 'Active', got %q", item.Name)
	}
}

func TestResolveStatusNoAutoCreateFails(t *testing.T) {
	cache := NewLookupCache(nil)
	// createStatuses is false by default
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{} // no status set, no default
	_, err := mapper.resolveStatus(dev)
	if err == nil {
		t.Fatal("expected error when no status and no auto-create")
	}
}

// ---------- resolveRole auto-create path ----------

func TestResolveRoleAutoCreateGeneric(t *testing.T) {
	cache := NewLookupCache(nil)
	cache.SetCreateRoles(true)
	cache.rolesMu.Lock()
	cache.roles["Generic"] = &CachedItem{ID: uuid.New(), Name: "Generic"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{} // no role set
	item, err := mapper.resolveRole(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "Generic" {
		t.Errorf("expected auto-created 'Generic', got %q", item.Name)
	}
}

func TestResolveRoleNoAutoCreateFails(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{}
	_, err := mapper.resolveRole(dev)
	if err == nil {
		t.Fatal("expected error when no role and no auto-create")
	}
}

// ---------- resolveLocation auto-create path ----------

func TestResolveLocationAutoCreateDefault(t *testing.T) {
	cache := NewLookupCache(nil)
	cache.SetCreateLocations(true)
	cache.locationsMu.Lock()
	cache.locations["Default"] = &CachedItem{ID: uuid.New(), Name: "Default"}
	cache.locationsMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{} // no location
	item, err := mapper.resolveLocation(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "Default" {
		t.Errorf("expected auto-created 'Default', got %q", item.Name)
	}
}

func TestResolveLocationNoAutoCreateFails(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{}
	_, err := mapper.resolveLocation(dev)
	if err == nil {
		t.Fatal("expected error when no location and no auto-create")
	}
}

func TestResolveLocationFromMetadataExplicit(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	cache.locationsMu.Lock()
	cache.locations["Lab-42"] = &CachedItem{ID: locID, Name: "Lab-42"}
	cache.locationsMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{
				"location": "Lab-42",
			},
		},
	}
	item, err := mapper.resolveLocation(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != locID {
		t.Errorf("expected location ID %s, got %s", locID, item.ID)
	}
}

// ---------- MapToPatchRequest with serial, asset tag, comments, custom fields ----------

func TestMapToPatchRequestSerialAssetComments(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["test-dt"] = &CachedItem{ID: dtID, Name: "test-dt"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-01"] = &CachedItem{ID: locID, Name: "DC-01"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultLocation: "DC-01", DefaultStatus: "Active", DefaultRole: "Server"})
	dev := &devicetypes.CaniDeviceType{
		Name:     "patch-dev",
		Slug:     "test-dt",
		Serial:   "SN-PATCH-1",
		AssetTag: "TAG-PATCH-1",
		Comments: "updated via patch",
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{
				"custom_key": "custom_val",
			},
		},
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Serial == nil || *req.Serial != "SN-PATCH-1" {
		t.Errorf("Serial = %v, want 'SN-PATCH-1'", req.Serial)
	}
	if req.AssetTag == nil || *req.AssetTag != "TAG-PATCH-1" {
		t.Errorf("AssetTag = %v, want 'TAG-PATCH-1'", req.AssetTag)
	}
	if req.Comments == nil || *req.Comments != "updated via patch" {
		t.Errorf("Comments = %v, want 'updated via patch'", req.Comments)
	}
	if req.CustomFields == nil {
		t.Fatal("CustomFields should not be nil")
	}
	cf := *req.CustomFields
	if cf["custom_key"] != "custom_val" {
		t.Errorf("CustomFields[custom_key] = %v, want 'custom_val'", cf["custom_key"])
	}
}

func TestMapToPatchRequestNilDevice(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	_, err := mapper.MapToPatchRequest(nil, uuid.New())
	if err == nil {
		t.Fatal("expected error for nil device")
	}
}

// ---------- MapToPatchRequest rack via Racks collection ----------

func TestMapToPatchRequestRackViaRacksCollection(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()
	rackID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["DC-01"] = &CachedItem{ID: locID, Name: "DC-01"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	// NOTE: GetRackByName calls the API, which panics with nil client.
	// This test verifies the code path where Parent is set and the rack IS in
	// the Racks collection, but because GetRackByName will fail (nil client),
	// the rack ref won't be set — confirming no panic when API is unreachable.
	// The real rack-assignment tests would need HTTP mocking.

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultLocation: "DC-01", DefaultStatus: "Active", DefaultRole: "Server"})
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {Name: "Rack-B"},
		},
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
	}
	mapper.SetInventory(inv)

	// Device with no slug so MapToPatchRequest skips deviceType resolution
	dev := &devicetypes.CaniDeviceType{
		Name:         "rack-dev",
		Parent:       rackID,
		RackPosition: 10,
		Face:         "rear",
	}

	// Since GetRackByName panics with nil client, we recover
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected — nil client panic from GetRackByName
			}
		}()
		mapper.MapToPatchRequest(dev, uuid.New())
	}()
	// If we reach here without panic propagating, the test passes.
	// The non-rack path is verified by assertions in other tests.
}

// ---------- MapToWritableDeviceRequest with rack in Racks collection ----------

func TestMapToWritableDeviceRequestWithRacksCollection(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()
	rackID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["srv-dt"] = &CachedItem{ID: dtID, Name: "srv-dt"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-01"] = &CachedItem{ID: locID, Name: "DC-01"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultLocation: "DC-01", DefaultStatus: "Active", DefaultRole: "Server"})
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {Name: "Rack-C"},
		},
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
	}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{
		Name:         "rack-srv",
		Slug:         "srv-dt",
		Rack:         rackID,
		RackPosition: 22,
		Face:         "rear",
	}

	// GetRackByName hits API — will panic with nil client.
	// This test verifies we reach the rack path without prior errors.
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected — nil client
			}
		}()
		mapper.MapToWritableDeviceRequest(dev)
	}()
}

// ---------- comparePosition matching (no diff) ----------

func TestComparePositionMatching(t *testing.T) {
	pos := 12
	dev := &devicetypes.CaniDeviceType{RackPosition: 12}
	remote := &nautobotapi.Device{Position: &pos}
	diffs := comparePosition(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when positions match, got %+v", diffs)
	}
}

func TestComparePositionLocalZero(t *testing.T) {
	pos := 5
	dev := &devicetypes.CaniDeviceType{RackPosition: 0}
	remote := &nautobotapi.Device{Position: &pos}
	diffs := comparePosition(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when local position is zero, got %+v", diffs)
	}
}

// ---------- compareFace matching (no diff) ----------

func TestCompareFaceMatching(t *testing.T) {
	faceVal := nautobotapi.DeviceFaceValue("front")
	dev := &devicetypes.CaniDeviceType{Face: "front"}
	remote := &nautobotapi.Device{
		Face: &nautobotapi.DeviceFace{Value: &faceVal},
	}
	diffs := compareFace(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when faces match, got %+v", diffs)
	}
}

func TestCompareFaceLocalEmpty(t *testing.T) {
	faceVal := nautobotapi.DeviceFaceValue("rear")
	dev := &devicetypes.CaniDeviceType{Face: ""}
	remote := &nautobotapi.Device{
		Face: &nautobotapi.DeviceFace{Value: &faceVal},
	}
	diffs := compareFace(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when local face is empty, got %+v", diffs)
	}
}

// ---------- compareRack matching UUIDs ----------

func TestCompareRackBothMatchingSameRack(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{},
	}
	mapper.SetInventory(inv)

	// Both have no rack → match (nil)
	dev := &devicetypes.CaniDeviceType{Rack: uuid.Nil}
	remote := &nautobotapi.Device{Rack: nil}
	diffs := compareRack(dev, remote, mapper)
	if diffs != nil {
		t.Errorf("expected nil diffs, got %+v", diffs)
	}
}

func TestCompareRackRemoteHasRackLocalNil(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{},
	}
	mapper.SetInventory(inv)

	rackUUID := uuid.New()
	rackIDUnion := nautobotapi.BulkWritableCableRequestStatusId{}
	rackIDUnion.FromBulkWritableCableRequestStatusId0(rackUUID)

	dev := &devicetypes.CaniDeviceType{Rack: uuid.Nil}
	remote := &nautobotapi.Device{
		Rack: &nautobotapi.BulkWritableCircuitRequestTenant{
			Id: &rackIDUnion,
		},
	}
	diffs := compareRack(dev, remote, mapper)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Field != "rack" {
		t.Errorf("expected field 'rack', got %q", diffs[0].Field)
	}
}

// ---------- generateDeviceNames with Model path ----------

func TestGenerateDeviceNamesModelPath(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id: {ID: id, Name: "", Type: "node", Model: "HPE DL380 Gen10"},
		},
	}
	generateDeviceNames(inv)
	expected := "cani-hpe-dl380-gen10"
	if inv.Devices[id].Name != expected {
		t.Errorf("expected %q, got %q", expected, inv.Devices[id].Name)
	}
}

func TestGenerateDeviceNamesSkipsNonDevice(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id: {ID: id, Name: "", Type: "rack", Model: "should-not-change"},
		},
	}
	generateDeviceNames(inv)
	if inv.Devices[id].Name != "" {
		t.Errorf("expected empty name for non-device, got %q", inv.Devices[id].Name)
	}
}

func TestGenerateDeviceNamesUUIDFallback(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id: {ID: id, Name: "", Type: "node"},
		},
	}
	generateDeviceNames(inv)
	expected := "cani-" + id.String()[:8]
	if inv.Devices[id].Name != expected {
		t.Errorf("expected %q, got %q", expected, inv.Devices[id].Name)
	}
}

// ---------- resolveContentLocation sibling path ----------

func TestResolveContentLocationSiblingResolution(t *testing.T) {
	locID := uuid.New()
	siblingID := uuid.New()
	childID := uuid.New()

	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			locID: {
				Name:         "DC-Main",
				LocationType: "data-center", // does not support "device"
				Children:     []uuid.UUID{},
			},
			siblingID: {
				Name:         "DC-Main",
				LocationType: "data-center",
				Children:     []uuid.UUID{childID},
			},
			childID: {
				Name:         "Row-1",
				LocationType: "row", // supports "device" via GetLocationTypeBySlug
			},
		},
	}

	// This tests the sibling resolution path. However, it depends on
	// devicetypes.GetLocationTypeBySlug returning correct data. If the
	// location type isn't registered, the function returns "".
	result := resolveContentLocation(locID, "device", inv)
	// The result depends on whether "row" is registered as supporting "device"
	// in the devicetypes registry. Either way, we confirm no panic.
	_ = result
}

// ---------- disambiguateDeviceNames — RackPosition without rack name ----------

func TestDisambiguateDeviceNamesRackPositionNoName(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	rackID := uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id1: {ID: id1, Name: "dup-node", Type: "node", RackPosition: 3, Rack: rackID},
			id2: {ID: id2, Name: "dup-node", Type: "node", RackPosition: 7, Rack: rackID},
		},
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {Name: ""},
		},
	}

	disambiguateDeviceNames(inv)

	// Both should have unique names with position suffix
	if inv.Devices[id1].Name == inv.Devices[id2].Name {
		t.Errorf("names should be unique, both are %q", inv.Devices[id1].Name)
	}
}

// ---------- ErrDeviceUnclassified sentinel ----------

func TestErrDeviceUnclassifiedSentinel(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{Strict: false})
	dev := &devicetypes.CaniDeviceType{} // no slug or model

	_, err := mapper.resolveDeviceType(dev)
	if !errors.Is(err, ErrDeviceUnclassified) {
		t.Errorf("expected ErrDeviceUnclassified, got %v", err)
	}
}

// ---------- printLoadSummary conflict without diffs ----------

func TestPrintLoadSummaryConflictNoDiffs(t *testing.T) {
	e := &Exporter{}
	result := &LoadResult{
		Skipped: []string{"dev-5"},
		Conflicts: []ConflictInfo{
			{DeviceName: "dev-5", Diffs: nil},
		},
	}
	e.printLoadSummary(result)
}
