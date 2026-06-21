package export

import (
	"context"
	"errors"
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

// TestIsValidNautobotInterfaceType verifies that supported Nautobot interface
// type slugs are accepted and unsupported ones (e.g. nvlink, pcie) are rejected.
//
// Why it matters: cani may carry hardware interface types Nautobot has no enum
// for; sending an invalid type aborts the interface create, so export must
// pre-validate and skip/translate unknowns.
// Inputs: a list of known-valid slugs and a list of known-invalid ones.
// Outputs: bool valid/invalid per slug.
// Data choice: the valid list mirrors the function's switch exactly; nvlink and
// pcie-gen5-x16 are real GPU-fabric types cani knows but Nautobot does not.
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

// ---------- DeviceMapper nil device ----------

// TestMapToNautobotDeviceNilDevice verifies that mapping a nil device returns
// the exact error "device is nil" rather than panicking.
//
// Why it matters: the bulk-create path maps many devices in a loop; a nil entry
// must fail fast with a clear message instead of crashing the whole export.
// Inputs: a nil *CaniDeviceType. Outputs: nil request and the sentinel error.
// Data choice: asserting the exact error string locks the contract callers and
// tests rely on for nil handling.
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

// TestMapToWritableDeviceRequestNilDevice verifies the single-create mapper
// also rejects a nil device with the "device is nil" error.
//
// Why it matters: the single-device create path is a separate method from the
// bulk path and must enforce the same nil guard so neither can panic.
// Inputs: a nil *CaniDeviceType. Outputs: nil request and the sentinel error.
// Data choice: mirrors TestMapToNautobotDeviceNilDevice to prove both mapper
// entry points share identical nil-safety behavior.
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

// TestDeviceMapperSetInventory verifies SetInventory stores the inventory
// reference on the mapper for later parent/rack resolution.
//
// Why it matters: the mapper needs the full inventory to resolve a device's
// rack and parent UUIDs into Nautobot references; without it those lookups
// would be empty and placement would be lost on export.
// Inputs: a mapper and an Inventory with an empty Devices map.
// Outputs: mapper.inventory points at the supplied inventory.
// Data choice: an empty-but-non-nil inventory isolates the assignment from any
// resolution side effects.
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

// ---------- Exporter resolveLocationName ----------

// TestResolveLocationName verifies a location UUID resolves to its name from
// the inventory.
//
// Why it matters: IPAM objects (VLANs, prefixes) reference locations by name in
// Nautobot, so export must turn cani location UUIDs back into names.
// Inputs: an inventory with one location "Building-A" and its UUID.
// Outputs: the name and a nil error.
// Data choice: a single named location is the minimal fixture proving the
// happy-path lookup.
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

// TestResolveLocationNameNotFound verifies that resolving an unknown location
// UUID returns an error.
//
// Why it matters: a dangling location reference must surface as an error so the
// export reports it rather than silently writing an empty location name.
// Inputs: an empty Locations map and a random UUID.
// Outputs: an error (name ignored).
// Data choice: the empty map guarantees the miss branch, complementing the
// found-case test above.
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

// TestPrintSummaryDoesNotPanic verifies PrintSummary renders a fully populated
// LoadResult without panicking.
//
// Why it matters: the summary is the operator's end-of-run report; a nil-deref
// or format panic there would mask an otherwise successful export.
// Inputs: a LoadResult with every counter and slice field populated.
// Outputs: none asserted beyond no panic (function only logs).
// Data choice: filling every field exercises all conditional print branches at
// once. Note: this is a smoke test and asserts no output content.
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

// TestNewNautobotClientErrors verifies the client constructor rejects an empty
// URL and an empty token.
//
// Why it matters: a client built without a base URL or auth token cannot reach
// Nautobot; failing at construction gives a clear error instead of opaque 401s
// or malformed requests later.
// Inputs: ("","token") then ("http://example.com","").
// Outputs: an error in both cases.
// Data choice: the two calls isolate each required-field guard independently.
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

// TestNewNautobotClientSuccess verifies a non-nil client is returned when both
// URL and token are supplied.
//
// Why it matters: this is the happy path every export run depends on; the
// constructor must succeed with valid inputs and not require a live server.
// Inputs: a valid URL and token.
// Outputs: a non-nil *NautobotClient and nil error.
// Data choice: example.com with a dummy token proves construction is offline
// and does not validate connectivity here.
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

// TestRefID verifies refID extracts a UUID from a reference union, returns Nil
// for a nil pointer, and that tenantRefID delegates identically.
//
// Why it matters: Nautobot responses wrap related-object IDs in union types;
// diffing local vs remote devices depends on pulling the UUID back out, and a
// nil-unsafe extractor would panic on absent references.
// Inputs: nil, then a union built from a fixed UUID.
// Outputs: uuid.Nil for nil, the embedded UUID otherwise.
// Data choice: a fixed all-a UUID makes the round-trip assertion exact and the
// tenantRefID delegation check reuses the same union.
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

// TestGenerateDeviceNamesAllBranches verifies name generation picks serial,
// then slug, then model, then UUID, leaves already-named devices alone, and
// names "system"-typed devices via the UUID fallback.
//
// Why it matters: Nautobot devices require a name; unnamed cani devices must get
// a deterministic, identifiable "cani-" name or they cannot be created.
// Inputs: six devices each missing one identifier tier (plus a named one).
// Outputs: in-place Name fields following the priority order.
// Data choice: one device per branch proves the precedence; "ProLiant DL380"
// checks the model lowercasing/space-to-dash transform.
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

// TestDisambiguateDeviceNamesAllSuffixes verifies duplicate names are made
// unique via serial, rack-position, or UUID suffixes (in that priority).
//
// Why it matters: Nautobot enforces device-name uniqueness per location/tenant,
// so colliding cani names must be disambiguated before export or the second
// create fails.
// Inputs: three devices all named "dup-server" with differing identifiers and a
// rack for the position suffix.
// Outputs: three now-unique names with the expected suffix formats.
// Data choice: giving each duplicate a different identifier forces all three
// suffix branches and the post-check asserts no name remains duplicated.
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

// TestDisambiguateDeviceNamesNoDuplicates verifies devices with already-unique
// names are left unchanged.
//
// Why it matters: disambiguation must not rewrite names that are already valid,
// which would churn Nautobot records and break stable identity across exports.
// Inputs: two devices with distinct names "server-1" and "server-2".
// Outputs: both names unchanged.
// Data choice: two unique names exercise the len<=1 group skip, the inverse of
// the all-suffixes test.
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

// TestCacheVLAN verifies CacheVLAN stores an item under the "vid:location" key
// in the package-level VLAN cache.
//
// Why it matters: VLANs are scoped per location in Nautobot, so caching by VID
// alone would collide across sites; the composite key keeps resolved IDs
// correct during export.
// Inputs: VID 100, location "site-a", and a CachedItem.
// Outputs: the item retrievable at key "100:site-a".
// Data choice: the test reads the global vlans map directly to confirm the
// exact composite key format. Note: this mutates package-global state.
func TestCacheVLAN(t *testing.T) {
	resetIPAMCaches()
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

// TestCachePrefix verifies CachePrefix stores an item under its CIDR key in the
// package-level prefix cache.
//
// Why it matters: prefixes are looked up by CIDR while linking IP addresses and
// child prefixes; a wrong key causes redundant API lookups or mis-parenting.
// Inputs: CIDR "10.0.0.0/24" and a CachedItem.
// Outputs: the item retrievable at key "10.0.0.0/24".
// Data choice: reading the global prefixes map confirms CIDR is used verbatim
// as the key. Note: this mutates package-global state.
func TestCachePrefix(t *testing.T) {
	resetIPAMCaches()
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

// TestCacheIPAddress verifies CacheIPAddress stores an item under its address
// key in the package-level IP cache.
//
// Why it matters: IP addresses are deduplicated by address string during
// export; an incorrect key would re-create existing addresses in Nautobot.
// Inputs: address "10.0.0.1/32" and a CachedItem.
// Outputs: the item retrievable at key "10.0.0.1/32".
// Data choice: reading the global ipAddresses map confirms the verbatim address
// key. Note: this mutates package-global state.
func TestCacheIPAddress(t *testing.T) {
	resetIPAMCaches()
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

// TestMakeIDRef verifies makeIDRef wraps a UUID into a status reference whose
// embedded Id round-trips back to the same UUID.
//
// Why it matters: device writes reference status/role/etc. by this union type;
// a malformed wrapper would send a bad or empty reference to Nautobot.
// Inputs: a fixed UUID. Outputs: a ref with non-nil Id decoding to that UUID.
// Data choice: an all-1s UUID makes the AsBulkWritableCableRequestStatusId0
// round-trip assertion unambiguous.
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

// TestMakeLocationRef verifies makeLocationRef wraps a UUID into a prefix-
// location reference whose Id round-trips back to the same UUID.
//
// Why it matters: prefix writes attach a location by this typed reference, so
// the UUID must survive wrapping to place the prefix correctly.
// Inputs: a fixed UUID. Outputs: a ref with non-nil Id decoding to that UUID.
// Data choice: an all-2s UUID distinguishes this case from the other make* tests
// while reusing the same round-trip check.
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

// TestMakePrefixParentRef verifies makePrefixParentRef wraps a UUID into a
// prefix-parent reference whose Id round-trips back to the same UUID.
//
// Why it matters: nested prefixes set their parent via this reference; a lost
// UUID would orphan child prefixes during export.
// Inputs: a fixed UUID. Outputs: a ref with non-nil Id decoding to that UUID.
// Data choice: an all-3s UUID keeps each make* test's fixture distinct for
// clearer failure diagnosis.
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

// TestMakeIPParentRef verifies makeIPParentRef wraps a UUID into an IP-address
// parent (prefix) reference whose Id round-trips back to the same UUID.
//
// Why it matters: IP addresses link to their containing prefix via this typed
// reference, which export needs to set parentage correctly.
// Inputs: a fixed UUID. Outputs: a ref with non-nil Id decoding to that UUID.
// Data choice: an all-4s UUID mirrors the sibling make* tests with a unique
// value.
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

// TestMakeIPNamespaceRef verifies makeIPNamespaceRef wraps a UUID into an IP
// namespace reference whose Id round-trips back to the same UUID.
//
// Why it matters: IP addresses and prefixes are scoped to a namespace in
// Nautobot; a broken namespace reference would file addresses in the wrong
// (or default) namespace.
// Inputs: a fixed UUID. Outputs: a ref with non-nil Id decoding to that UUID.
// Data choice: an all-5s UUID completes the distinct-per-test fixture set.
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

// ---------- locationTypeSupports ----------

// TestLocationTypeSupports verifies a registered location type reports support
// only for content types in its list, and unknown slugs return false.
//
// Why it matters: Nautobot rejects placing a rack/device under a location type
// that does not list that content type, so export must check before assigning.
// Inputs: a "test-room-export" type supporting rack+device; queried for rack,
// device, module, and an unknown slug.
// Outputs: true for rack/device, false for module and the unknown slug.
// Data choice: a type that supports two of three content types plus a bogus slug
// covers the match, no-match, and not-found branches.
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

// TestFindContentChild verifies the recursive search returns a child location
// whose type supports the requested content, and "" when none does.
//
// Why it matters: cani hierarchies often attach racks/devices to a generic
// parent (e.g. Building) that cannot hold them; export must descend to a child
// (e.g. Room) that can, or Nautobot placement fails.
// Inputs: a Building parent (no content types) with a Room child supporting
// rack+device.
// Outputs: "Server-Room" for rack; "" for an unsupported content type.
// Data choice: a two-level building->room tree is the minimal shape that forces
// a descent into the child.
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

// TestResolveContentLocation verifies content-location resolution: it descends
// to a supporting child, returns "" for nil/Nil inputs, and returns a location
// that directly supports the content type.
//
// Why it matters: this is the top-level helper that picks the correct Nautobot
// location for a rack/device, including the fallback walk; wrong results cause
// misplacement or export errors.
// Inputs: a DC parent with a Section child; queried by DC id, nil inventory,
// Nil UUID, and the Section id directly.
// Outputs: "Section-A" for the descent and direct cases, "" for nil/Nil.
// Data choice: the DC->Section tree plus the two empty inputs cover the descent,
// guard, and direct-support branches in one test.
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

// TestRemoteSlotKeyNil verifies remoteSlotKey returns nil when the device is
// nil, has no position, or has a position but no rack.
//
// Why it matters: position-swap detection compares slot keys; a non-nil key for
// an unplaced remote device would invent a phantom slot and trigger spurious
// moves during export.
// Inputs: nil, a device with no position, a device with position but no rack.
// Outputs: nil in all three cases.
// Data choice: each case removes one required field to prove every guard clause
// independently returns nil.
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

// TestRemoteSlotKeyValid verifies a fully placed remote device yields a slotKey
// with the right rack, position, and face, defaulting face to "front" and
// honoring an explicit "rear".
//
// Why it matters: accurate slot keys drive Nautobot rack-position reconciliation;
// a wrong face or position would move devices incorrectly.
// Inputs: a device at position 5 in a fixed rack, then with face set to rear.
// Outputs: slotKey{rack, 5, "front"} then face "rear".
// Data choice: omitting face first proves the front default, then setting rear
// proves the explicit path; a fixed rack UUID makes the assertion exact.
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

// TestDerefSlotKeyNil verifies dereferencing a nil *slotKey yields the zero
// slotKey rather than panicking.
//
// Why it matters: swap logic dereferences optional slot keys when comparing
// current vs desired placement; a nil deref there would crash the export.
// Inputs: a nil *slotKey.
// Outputs: slotKey{Position:0, Face:"", RackID:Nil}.
// Data choice: asserting all three zero fields confirms a true zero value, not a
// partially initialized struct.
func TestDerefSlotKeyNil(t *testing.T) {
	got := derefSlotKey(nil)
	if got.Position != 0 || got.Face != "" || got.RackID != uuid.Nil {
		t.Error("expected zero slotKey for nil input")
	}
}

// TestDerefSlotKeyNonNil verifies dereferencing a non-nil *slotKey copies its
// fields exactly.
//
// Why it matters: the value copy is compared against remote slot keys during
// reconciliation; a dropped field would cause incorrect move decisions.
// Inputs: a *slotKey with a fixed rack, position 42, face "rear".
// Outputs: an equal slotKey value.
// Data choice: distinctive non-zero values (42, "rear") ensure the copy is real
// and not coincidentally matching a zero value.
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

// TestValidateInventoryNil verifies ValidateInventory rejects a nil inventory.
//
// Why it matters: export must refuse to run against a nil inventory with a clear
// error instead of panicking on the first map access.
// Inputs: nil. Outputs: a non-nil error.
// Data choice: nil is the most basic precondition failure and the first guard
// in the function.
func TestValidateInventoryNil(t *testing.T) {
	if err := ValidateInventory(nil); err == nil {
		t.Error("expected error for nil inventory")
	}
}

// TestValidateInventoryEmpty verifies an inventory with zero devices is rejected.
//
// Why it matters: there is nothing to export from an empty inventory, so it
// should fail validation rather than make a no-op run look successful.
// Inputs: an Inventory with an empty Devices map.
// Outputs: a non-nil error.
// Data choice: an explicitly empty (non-nil) map distinguishes this from the
// nil-inventory case above.
func TestValidateInventoryEmpty(t *testing.T) {
	if err := ValidateInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
	}); err == nil {
		t.Error("expected error for empty devices")
	}
}

// TestValidateInventoryAllNilDevices verifies an inventory whose only entries
// are nil device pointers is rejected.
//
// Why it matters: a map sized >0 but holding only nils has no exportable devices;
// validation must catch this so export does not report success with nothing done.
// Inputs: a Devices map with a single nil entry.
// Outputs: a non-nil error.
// Data choice: one nil entry makes len(Devices)>0 yet validCount==0, hitting the
// "no exportable devices" branch.
func TestValidateInventoryAllNilDevices(t *testing.T) {
	if err := ValidateInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			uuid.New(): nil,
		},
	}); err == nil {
		t.Error("expected error for all nil devices")
	}
}

// TestValidateInventoryOnlySystem verifies an inventory containing only
// "system"-typed devices is rejected.
//
// Why it matters: system objects are organizational, not real Nautobot devices;
// an inventory with only systems has nothing to create and must fail validation.
// Inputs: a single device of type "system".
// Outputs: a non-nil error.
// Data choice: the lone system device exercises the type=="system" skip that
// drives validCount to zero.
func TestValidateInventoryOnlySystem(t *testing.T) {
	if err := ValidateInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			uuid.New(): {Name: "sys", Type: "system"},
		},
	}); err == nil {
		t.Error("expected error for only system devices")
	}
}

// TestValidateInventoryOnlyUnnamed verifies an inventory of only unnamed devices
// is rejected.
//
// Why it matters: Nautobot devices need a name; if name generation has not run,
// validation should stop the export before it attempts nameless creates.
// Inputs: a single node device with an empty Name.
// Outputs: a non-nil error.
// Data choice: an empty-name node hits the Name=="" skip, the unnamed analog of
// the only-system case.
func TestValidateInventoryOnlyUnnamed(t *testing.T) {
	if err := ValidateInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			uuid.New(): {Name: "", Type: "node"},
		},
	}); err == nil {
		t.Error("expected error for only unnamed devices")
	}
}

// TestValidateInventoryValid verifies an inventory with one named node passes
// validation.
//
// Why it matters: this is the happy path gating every export; a false rejection
// here would block valid runs entirely.
// Inputs: a single node named "server-1".
// Outputs: nil error.
// Data choice: one minimally valid device (named, type node) is the smallest
// fixture that should pass.
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

// ---------- compareDeviceFields ----------

// TestCompareDeviceFieldsSerial verifies a differing serial number produces a
// "serial" FieldDiff carrying the local and remote values.
//
// Why it matters: --merge updates rely on this diff to know which fields to
// PATCH in Nautobot; a missed serial diff would leave stale serials remotely.
// Inputs: local device serial "NEW-SN" vs remote "OLD-SN".
// Outputs: a diff list containing {serial, NEW-SN, OLD-SN}.
// Data choice: distinct new/old serials and otherwise-empty objects isolate the
// serial comparison branch.
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

// TestCompareDeviceFieldsAssetTag verifies a differing asset tag produces an
// "asset_tag" FieldDiff carrying the local and remote values.
//
// Why it matters: asset tags are tracked for inventory/audit; export's merge
// path must detect and propagate changes so Nautobot stays in sync.
// Inputs: local asset tag "NEW-TAG" vs remote "OLD-TAG".
// Outputs: a diff list containing {asset_tag, NEW-TAG, OLD-TAG}.
// Data choice: distinct new/old tags on otherwise-empty objects isolate the
// asset-tag branch, mirroring the serial test.
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

// TestCompareDeviceFieldsNoChanges verifies that two empty device objects
// produce no field diffs.
//
// Why it matters: export should skip (not update) devices that match Nautobot,
// avoiding needless PATCHes; any spurious diff would cause churn on every run.
// Inputs: an empty local CaniDeviceType and an empty remote Device.
// Outputs: an empty diff slice.
// Data choice: empty-vs-empty exercises the all-fields-skipped path, since each
// comparison branch is guarded by a non-empty/non-nil check.
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

// TestResolveLocalRackNameNilInventory verifies that resolveLocalRackName
// returns an empty string when the mapper has no inventory set, even though the
// device carries a non-nil Rack UUID.
//
// Why it matters: rack-diff comparison must not panic or fabricate a rack name
// when inventory is unavailable, so a device with an unknown rack is treated as
// unracked rather than emitting a spurious rack PATCH to Nautobot.
// Inputs: a device with a random Rack UUID and a mapper whose inventory is nil.
// Outputs: the empty string.
// Data choice: a populated Rack UUID with nil inventory isolates the nil-guard
// branch, proving the lookup short-circuits before dereferencing inventory.
func TestResolveLocalRackNameNilInventory(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{Rack: uuid.New()}
	if got := resolveLocalRackName(dev, mapper); got != "" {
		t.Errorf("expected empty string for nil inventory, got %q", got)
	}
}

// TestResolveLocalRackNameNoRack verifies that resolveLocalRackName returns an
// empty string when the device resolves to no rack (Rack and Parent are Nil)
// against an empty Racks map.
//
// Why it matters: free-standing devices with no rack assignment must export
// without a rack reference; returning "" keeps them out of rack diffs.
// Inputs: a device with Rack=Nil and Parent=Nil plus an inventory holding an
// empty Racks map. Outputs: the empty string.
// Data choice: empty Racks plus Nil IDs exercises the GetRackID==Nil branch,
// distinct from the nil-inventory case above.
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

// TestResolveLocalRackNameFound verifies that resolveLocalRackName returns the
// rack's Name when the device's Rack UUID maps to a rack in inventory.
//
// Why it matters: the rack name is the key used to look up the rack's Nautobot
// UUID during diffing, so resolving it correctly is what lets a device be
// placed in the right rack on export.
// Inputs: a device whose Rack UUID matches an inventory rack named "Rack-42".
// Outputs: "Rack-42".
// Data choice: a single named rack keyed by the exact UUID confirms the happy
// path returns the stored name verbatim.
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

// TestCompareRackBothNil verifies that compareRack reports no diff when the
// local device has no rack (Rack=Nil) and the remote Nautobot device also has a
// nil Rack.
//
// Why it matters: export should not PATCH a device's rack when neither side
// assigns one; a false diff would churn unracked devices on every run.
// Inputs: a device with Rack=Nil and a remote Device with Rack=nil.
// Outputs: an empty diff slice.
// Data choice: both-nil is the symmetric no-op case and avoids the
// GetRackByName API call, which would panic on this nil client.
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

// TestCompareRackLocalHasRack verifies that resolveLocalRackName finds the
// local rack name "Rack-A" when the device is assigned to an inventory rack.
//
// Why it matters: compareRack relies on this name to look up the remote rack
// UUID; the test documents that full compareRack coverage needs a mocked API
// because GetRackByName panics on a nil client.
// Inputs: a device whose Rack UUID maps to inventory rack "Rack-A".
// Outputs: "Rack-A".
// Data choice: a nil API client forces the test to assert only name
// resolution. NOTE: despite its name it does not actually call compareRack.
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

// TestMapToNautobotDeviceFull verifies that MapToNautobotDevice produces a
// fully-populated BulkWritableDeviceRequest (name, serial, asset_tag, comments)
// when device-type, location, status, and role caches are all pre-seeded.
//
// Why it matters: bulk device creation is the primary export path; every
// required FK must resolve from cache and every scalar field must round-trip so
// Nautobot records match the cani inventory.
// Inputs: a fully-populated CaniDeviceType plus a cache holding the slug,
// location, status, and role. Outputs: a non-nil request with all fields set.
// Data choice: realistic HPE DL360 values exercise the all-fields-present path
// with no auto-creation or error branches.
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

// TestMapToWritableDeviceRequestFull verifies that MapToWritableDeviceRequest
// builds a complete request (name, serial, asset_tag, comments) when all
// reference caches are pre-seeded.
//
// Why it matters: this is the single-device create/update path; like the bulk
// path it must resolve every FK and copy scalar fields so updates to Nautobot
// faithfully reflect cani data.
// Inputs: a populated device and a cache with the DL380 slug, location,
// "Planned" status, and "Compute" role. Outputs: a non-nil request, fields set.
// Data choice: a different model/status/role than the bulk test guards against
// hard-coded values leaking between the two mapping methods.
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

// TestMapToWritableDeviceRequestWithRack verifies that MapToWritableDeviceRequest
// succeeds and leaves Rack nil when the device has a rack position and face but
// no resolvable parent rack (rack lookup needs an API client, which is nil).
//
// Why it matters: devices may carry position/face hints before their rack is
// known; the mapper must still produce a valid request and simply omit the rack
// FK rather than fail or panic.
// Inputs: a device with RackPosition=10 and Face="rear" but no Parent.
// Outputs: a non-nil request with Rack==nil.
// Data choice: the nil API client deliberately suppresses rack resolution.
// NOTE: the name says "WithRack" yet it asserts the no-rack-resolution branch.
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

// ---------- getDeviceInterfaceSpecs ----------

// TestGetDeviceInterfaceSpecsWithInterfaces verifies that getDeviceInterfaceSpecs
// returns one spec per instantiated device interface, preserving name and type
// and deriving speed (e.g. infiniband-ndr -> 400000000 Kbps).
//
// Why it matters: when the device-type library already supplies interfaces,
// export must use them verbatim so Nautobot interface records match the real
// hardware rather than a fabricated fallback set.
// Inputs: a node with three interfaces (eth0, mgmt0, ib0). Outputs: three specs
// in order with derived speeds.
// Data choice: mixing 1000base-t and infiniband-ndr exercises both type
// pass-through and the speed-derivation lookup.
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

// TestGetDeviceInterfaceSpecsFallbackBlade verifies that getDeviceInterfaceSpecs
// synthesizes a fallback interface set (iLO + 4 Ethernet) for a blade/node with
// no instantiated interfaces, leading with the management iLO port.
//
// Why it matters: devices lacking a populated interface list still need
// interfaces in Nautobot; the type-based fallback guarantees a sensible default
// so exported servers are not left interface-less.
// Inputs: a Blade with Model "ProLiant DL360" and no Interfaces.
// Outputs: at least 5 specs, the first named "iLO".
// Data choice: a non-InfiniBand model keeps the count at the base 5 (no
// ib0/ib1 added).
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

// TestGetDeviceInterfaceSpecsFallbackHSNSwitch verifies that
// getDeviceInterfaceSpecs builds 65 specs for an HSNSwitch with no interfaces:
// one mgmt0 plus 64 InfiniBand-NDR osfp ports.
//
// Why it matters: high-speed-network switches export with a fixed dense port
// layout; the exact count and osfp type must match so fabric topology is
// represented correctly in Nautobot.
// Inputs: a device with Type HSNSwitch and no Interfaces.
// Outputs: 65 specs; specs[0]=="mgmt0", specs[1].Type=="infiniband-ndr".
// Data choice: asserting the count plus the first two specs pins down both the
// management port and the osfp block.
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

// TestGetDeviceInterfaceSpecsFallbackMgmtSwitch verifies that
// getDeviceInterfaceSpecs builds 53 specs for a MgmtSwitch with no interfaces:
// mgmt0 + 48 copper ports + 4 SFP uplinks.
//
// Why it matters: management switches have a standard Aruba-2930F-style port
// profile; the precise count ensures all switch ports are exported to Nautobot.
// Inputs: a device with Type MgmtSwitch and no Interfaces. Outputs: 53 specs.
// Data choice: the count (1+48+4) is the simplest assertion that all three
// port groups were emitted.
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

// TestGetDeviceInterfaceSpecsFallbackPDU verifies that getDeviceInterfaceSpecs
// emits exactly one mgmt0 interface for a CabinetPDU with no interfaces.
//
// Why it matters: PDUs are network-managed only; exporting a single management
// interface (and nothing else) keeps their Nautobot representation accurate.
// Inputs: a device with Type CabinetPDU and no Interfaces.
// Outputs: a single spec named "mgmt0".
// Data choice: the PDU is the minimal fallback case, asserting both length 1
// and the mgmt0 name.
func TestGetDeviceInterfaceSpecsFallbackPDU(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Type: devicetypes.Type(devicetypes.CabinetPDU),
	}

	specs := getDeviceInterfaceSpecs(dev)
	if len(specs) != 1 || specs[0].Name != "mgmt0" {
		t.Fatalf("expected 1 mgmt spec for PDU, got %d", len(specs))
	}
}

// TestGetDeviceInterfaceSpecsInfinibandNode verifies that getDeviceInterfaceSpecs
// adds two InfiniBand ports (ib0, ib1) to the node fallback when the model name
// signals InfiniBand, yielding iLO + 4 eth + 2 ib = 7 specs.
//
// Why it matters: nodes with IB adapters must export their fabric interfaces;
// the model-name heuristic (containsInfiniband) is what triggers the extra
// ports so HSN connectivity is captured.
// Inputs: a Node with Model "Quantum InfiniBand NDR Switch" and no Interfaces.
// Outputs: 7 specs; the last two named "ib0" and "ib1".
// Data choice: a model containing "infiniband"/"ndr" reliably fires the IB
// branch on top of the base node set.
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

// TestPrintSummaryAllBranches verifies that PrintSummary executes without
// panicking when given a LoadResult that exercises every counter and slice
// (created/updated/skipped/conflicts/errors across all object kinds).
//
// Why it matters: the summary is the operator-facing report of an export run;
// it must render all non-zero sections without crashing regardless of which
// counters are populated.
// Inputs: a LoadResult with every field non-empty. Outputs: none asserted
// beyond the absence of a panic (the rendered text is not captured).
// Data choice: populating all fields drives every print branch in one call.
// NOTE: this is a smoke test; it does not assert the rendered output.
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

// TestCompareDeviceFieldsPositionFace verifies that compareDeviceFields reports
// both a "position" and a "face" diff when the local rack position and face
// differ from the remote Nautobot device.
//
// Why it matters: rack position and orientation must be reconciled so a moved
// or reoriented device is PATCHed; missing either diff would leave Nautobot out
// of sync with the physical layout.
// Inputs: local RackPosition=10/Face="rear" vs remote Position=5/Face="front".
// Outputs: a diff slice containing both "position" and "face" entries.
// Data choice: deliberately mismatched values on both axes guarantee both
// comparison branches fire.
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

// TestLocationFromParentRackNilInventory verifies that locationFromParentRack
// returns an empty string when the mapper has no inventory, despite a non-nil
// device Rack UUID.
//
// Why it matters: location inheritance from a parent rack must be skipped safely
// when inventory is unavailable, letting resolveLocation fall through to
// metadata/default instead of panicking.
// Inputs: a device with a random Rack UUID and a nil-inventory mapper.
// Outputs: the empty string.
// Data choice: a populated Rack UUID isolates the nil-inventory guard.
func TestLocationFromParentRackNilInventory(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{Rack: uuid.New()}

	if got := mapper.locationFromParentRack(dev); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// TestLocationFromParentRackNoRack verifies that locationFromParentRack returns
// an empty string when the device resolves to no rack (Rack and Parent Nil)
// against an empty Racks map.
//
// Why it matters: devices with no parent rack have no inherited location, so
// resolveLocation must continue to its other sources rather than receive a
// bogus name.
// Inputs: a device with Rack=Nil/Parent=Nil and an empty Racks map.
// Outputs: the empty string.
// Data choice: empty Racks plus Nil IDs hits the GetRackID==Nil branch.
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

// TestLocationFromParentRackResolvesChild verifies that locationFromParentRack
// returns the rack's location name when that location type supports devices,
// resolving via resolveContentLocation.
//
// Why it matters: Nautobot rejects a device whose location does not contain its
// rack, so the device must inherit the rack's device-capable location for a
// valid export.
// Inputs: a registered "server-room-lfp" location type (content types include
// "device"), a rack at that location, and a device in the rack.
// Outputs: "Server-Room-LFP".
// Data choice: a location type explicitly listing "device" content exercises
// the supported-location path without tree-walking to a child.
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

// TestMapToPatchRequestNil verifies that MapToPatchRequest returns an error when
// given a nil device.
//
// Why it matters: PATCH builds an update body for an existing Nautobot device;
// guarding nil input prevents a panic mid-export and surfaces the programming
// error instead.
// Inputs: a nil device and an arbitrary existing UUID. Outputs: a non-nil error.
// Data choice: nil is the boundary input that must be rejected before any field
// access.
func TestMapToPatchRequestNil(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	_, err := mapper.MapToPatchRequest(nil, uuid.New())
	if err == nil {
		t.Error("expected error for nil device")
	}
}

// TestMapToPatchRequestBasic verifies that MapToPatchRequest populates name,
// serial, asset_tag, comments and resolves device_type, location, status, and
// role when all caches are seeded.
//
// Why it matters: updating an existing device must carry both scalar edits and
// resolved FKs so an in-place Nautobot PATCH fully reconciles the record.
// Inputs: a populated device plus caches for slug, location, status, role.
// Outputs: a request with all scalars and all four references non-nil.
// Data choice: every reference pre-seeded drives the all-fields-set path with no
// skipped FK.
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

// TestMapToPatchRequestNoSlug verifies that MapToPatchRequest leaves DeviceType
// nil when the device has no slug, while still building the rest of the request.
//
// Why it matters: a PATCH should not overwrite a device's type with an empty
// value; omitting the device_type FK when the slug is unknown avoids corrupting
// the Nautobot record.
// Inputs: a device with only a Name and caches for location/status/role.
// Outputs: a request with DeviceType==nil.
// Data choice: an empty slug specifically targets the skip-device-type branch.
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

// TestMapToPatchRequestWithCustomFields verifies that MapToPatchRequest copies
// flattened ProviderMetadata into the request's CustomFields, e.g. a nested
// nautobot.xname becomes the "xname" custom field.
//
// Why it matters: cani stores Nautobot-specific attributes (like xname) in
// provider metadata; flattening them into custom fields is how that data reaches
// Nautobot on update.
// Inputs: a device whose ProviderMetadata nests {"nautobot":{"xname":...}}.
// Outputs: CustomFields["xname"]=="x1000c0s0b0".
// Data choice: a nested map proves FlattenProviderMetadata unwraps the
// "nautobot" namespace into a top-level custom field.
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

// TestMapToWritableRackRequestNil verifies that MapToWritableRackRequest returns
// an error when given a nil device.
//
// Why it matters: rack creation must reject nil input rather than panic, so a
// missing rack record fails cleanly during export.
// Inputs: a nil device. Outputs: a non-nil error.
// Data choice: nil is the boundary case guarded before any field access.
func TestMapToWritableRackRequestNil(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	_, err := mapper.MapToWritableRackRequest(nil)
	if err == nil {
		t.Error("expected error for nil device")
	}
}

// TestMapToWritableRackRequestBasic verifies that MapToWritableRackRequest sets
// the rack name and defaults UHeight to 48 when no u_height metadata is present.
//
// Why it matters: racks export with a sane default height so Nautobot can place
// devices by rack unit even when the cani record omits the height.
// Inputs: a rack device named "Rack-A01" with location/status cached and no
// u_height metadata. Outputs: a request with Name "Rack-A01" and UHeight 48.
// Data choice: omitting u_height drives the 48U default branch.
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

// TestMapToWritableRackRequestCustomHeight verifies that MapToWritableRackRequest
// honors ProviderMetadata["u_height"] instead of the 48U default.
//
// Why it matters: non-standard racks (e.g. 42U) must export with their real
// height so device positions remain valid in Nautobot.
// Inputs: a rack device with ProviderMetadata{"u_height":42} and cached
// location/status. Outputs: a request with UHeight 42.
// Data choice: 42 (not the 48 default) proves the metadata value overrides the
// default.
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

// TestMapToWritableRackRequestNoLocation verifies that MapToWritableRackRequest
// returns an error when no location can be resolved (no metadata, no default, no
// parent rack).
//
// Why it matters: Nautobot racks require a location; failing early stops an
// invalid rack create instead of letting the API reject it later.
// Inputs: a rack device with only Name/ID and a mapper with no DefaultLocation.
// Outputs: a non-nil error.
// Data choice: an empty MapperOpts removes every location source, forcing the
// "location is required" error.
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

// TestCompareDeviceFieldsTypeDiff verifies that when the locally resolved
// device type has a different UUID than the remote device's DeviceType ref,
// compareDeviceFields emits a "device_type" diff whose LocalVal is the cached
// local type name ("LocalType").
//
// Why it matters: the export update path relies on these diffs to detect when a
// device's hardware type drifted from Nautobot and must be patched; a wrong
// field name or value would push bad updates or mask real drift.
// Inputs: a CaniDeviceType (Slug "local-type"), a remote Device with a
// different device-type ref, and a seeded mapper. Outputs: a []FieldDiff that
// includes the device_type entry.
// Data choice: distinct local/remote device-type UUIDs force the mismatch while
// location/status/role are left matching so only device_type diverges.
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

// TestCompareDeviceFieldsStatusDiff verifies that compareDeviceFields emits a
// "status" diff (LocalVal "Planned") when the locally resolved status UUID
// differs from the remote device's Status reference.
//
// Why it matters: lifecycle status (Planned/Active/...) is exported to Nautobot,
// so the diff engine must flag status drift for the update path to reconcile it.
// Inputs: a CaniDeviceType with Status "Planned", a remote Device whose Status
// ref is a different UUID, and a mapper seeded with that status/location/role.
// Outputs: a []FieldDiff containing the status entry.
// Data choice: only the status UUID is made to differ; location and role caches
// match their defaults so the status field is the sole asserted divergence.
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

// TestCompareDeviceFieldsRoleDiff verifies that compareDeviceFields emits a
// "role" diff (LocalVal "Compute") when the locally resolved role UUID differs
// from the remote device's Role reference.
//
// Why it matters: device role drives Nautobot's device classification, so the
// exporter must detect and report role drift before patching the remote device.
// Inputs: an empty-slug CaniDeviceType, a remote Device whose Role ref is a
// fresh UUID, and a mapper whose DefaultRole "Compute" is cached. Outputs: a
// []FieldDiff containing the role entry.
// Data choice: location/status caches match their defaults while the remote role
// gets an unrelated UUID, isolating role as the only mismatch.
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

// TestCompareDeviceFieldsLocationDiff verifies that compareDeviceFields emits a
// "location" diff (LocalVal "Room-Z") when the locally resolved location UUID
// differs from the remote device's Location reference.
//
// Why it matters: placing devices at the correct Nautobot location is core to
// the export, so drift in location assignment must be surfaced for reconciliation.
// Inputs: an empty CaniDeviceType, a remote Device whose Location ref is a fresh
// UUID, and a mapper with DefaultLocation "Room-Z" cached. Outputs: a []FieldDiff
// containing the location entry.
// Data choice: status/role caches match their defaults while the remote location
// UUID is unrelated, so location is the only field that diverges.
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

// TestCompareDeviceFieldsNoTypeDiffWhenMatching verifies that compareDeviceFields
// returns an empty diff slice when every locally resolved reference UUID equals
// the corresponding remote reference.
//
// Why it matters: the export update path must NOT report changes when local and
// remote already agree, otherwise it would issue needless PATCH churn to Nautobot.
// Inputs: a CaniDeviceType (Slug "shared"), a remote Device whose four refs all
// use one shared UUID, and a mapper whose caches all resolve to that same UUID.
// Outputs: an empty []FieldDiff.
// Data choice: reusing a single sharedID for every cache entry and remote ref
// makes all four ID comparisons equal, exercising the no-diff path precisely.
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

// TestResolveContentLocationNilInventory verifies that resolveContentLocation
// returns "" when the inventory pointer is nil.
//
// Why it matters: the exporter calls this helper while resolving where to place
// a device/rack; a nil-inventory guard prevents a panic and lets the caller fall
// back to its default location instead of crashing the export.
// Inputs: a random location UUID, content type "device", and a nil *Inventory.
// Outputs: an empty string.
// Data choice: a nil inventory is the minimal fixture that triggers the early
// guard clause without needing any location data.
func TestResolveContentLocationNilInventory(t *testing.T) {
	got := resolveContentLocation(uuid.New(), "device", nil)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// TestResolveContentLocationNilID verifies that resolveContentLocation returns
// "" when the supplied location ID is uuid.Nil.
//
// Why it matters: devices/racks without a resolved location UUID must not match
// an arbitrary entry; returning empty lets the exporter fall back to a default.
// Inputs: uuid.Nil, content type "device", and an Inventory with an empty
// Locations map. Outputs: an empty string.
// Data choice: uuid.Nil plus an empty map isolates the nil-ID guard from the
// "not found in map" path tested separately.
func TestResolveContentLocationNilID(t *testing.T) {
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{},
	}
	got := resolveContentLocation(uuid.Nil, "device", inv)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// TestResolveContentLocationNotFound verifies that resolveContentLocation
// returns "" when the location ID is not present in the inventory map.
//
// Why it matters: a stale or dangling location reference must not silently
// resolve to the wrong place; empty output signals the caller to use a default.
// Inputs: a random UUID absent from the map, content type "device", and an
// Inventory with an empty Locations map. Outputs: an empty string.
// Data choice: a fresh UUID guaranteed not in the empty map exercises the
// map-miss branch distinctly from the nil-ID branch.
func TestResolveContentLocationNotFound(t *testing.T) {
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{},
	}
	got := resolveContentLocation(uuid.New(), "device", inv)
	if got != "" {
		t.Errorf("expected empty for missing location, got %q", got)
	}
}

// TestResolveContentLocationDirect verifies that resolveContentLocation returns
// the location's own name when that location's type directly supports the
// requested content type.
//
// Why it matters: devices must be placed at a Nautobot location whose type
// allows "device"; the direct-support case is the common happy path for export.
// Inputs: a location of type "server-room-rcl" (ContentTypes include "device"),
// its UUID, and content type "device". Outputs: the location name "Room-A-RCL".
// Data choice: a registered location type that lists "device" lets the helper
// short-circuit on direct support without walking any children.
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

// TestResolveContentLocationChildResolution verifies that resolveContentLocation
// walks the location tree and returns a descendant's name when the parent's own
// type does not support the content type but a child's type does.
//
// Why it matters: cani hierarchies often place racks/devices at a child location
// (e.g. a floor) under a parent (e.g. a building) that cannot hold them; the
// exporter must resolve down to the valid child.
// Inputs: a parent "building-rcl" (no content types) linking a child "floor-rcl"
// (supports rack/device), the parent UUID, and content type "rack". Outputs: the
// child name "Floor-2-RCL".
// Data choice: an empty parent plus a supporting child models the real
// building->floor nesting and forces the depth-first child search.
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

// TestResolveContentLocationNoMatch verifies that resolveContentLocation returns
// "" when neither the location nor any descendant supports the content type.
//
// Why it matters: if no part of the hierarchy can legally hold a device, the
// exporter must fall back to a default rather than force an invalid placement.
// Inputs: a single location of type "campus-rcl" (empty ContentTypes, no
// children), its UUID, and content type "device". Outputs: an empty string.
// Data choice: a childless location whose type supports nothing exercises the
// exhausted-search branch that yields the empty fallback signal.
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

// TestResolveFaceRear verifies that resolveFace("rear") returns a RackFace
// whose generated Nautobot union decodes to FaceEnumRear.
//
// Why it matters: Nautobot requires a face value whenever a rack position is
// set, so the exporter must always produce a usable RackFace for rack-mounted
// devices and must preserve an explicit rear orientation.
// Inputs: the face string "rear". Outputs: a non-nil *nautobotapi.RackFace
// whose enum value is FaceEnumRear.
// Data choice: "rear" exercises the explicit non-default switch case.
func TestResolveFaceRear(t *testing.T) {
	rf := resolveFace("rear")
	if rf == nil {
		t.Fatal("expected non-nil RackFace")
	}
	face, err := rf.AsFaceEnum()
	if err != nil {
		t.Fatalf("decode RackFace: %v", err)
	}
	if face != nautobotapi.FaceEnumRear {
		t.Errorf("face = %v, want %v", face, nautobotapi.FaceEnumRear)
	}
}

// TestResolveFaceFront verifies that resolveFace("front") returns a RackFace
// whose generated Nautobot union decodes to FaceEnumFront.
//
// Why it matters: rack-mounted devices exported to Nautobot need a valid face;
// "front" is the most common orientation and must map to a usable RackFace.
// Inputs: the face string "front". Outputs: a non-nil *nautobotapi.RackFace
// whose enum value is FaceEnumFront.
// Data choice: "front" hits the default switch branch with an explicit value.
func TestResolveFaceFront(t *testing.T) {
	rf := resolveFace("front")
	if rf == nil {
		t.Fatal("expected non-nil RackFace")
	}
	face, err := rf.AsFaceEnum()
	if err != nil {
		t.Fatalf("decode RackFace: %v", err)
	}
	if face != nautobotapi.FaceEnumFront {
		t.Errorf("face = %v, want %v", face, nautobotapi.FaceEnumFront)
	}
}

// TestResolveFaceEmpty verifies that resolveFace("") defaults to a RackFace
// whose generated Nautobot union decodes to FaceEnumFront.
//
// Why it matters: cani devices may omit a face; since Nautobot demands one when
// a rack position exists, the exporter must default rather than emit nil/null.
// Inputs: an empty face string. Outputs: a non-nil *nautobotapi.RackFace
// whose enum value is FaceEnumFront.
// Data choice: the empty string drives the default branch without an explicit
// input face.
func TestResolveFaceEmpty(t *testing.T) {
	rf := resolveFace("")
	if rf == nil {
		t.Fatal("expected non-nil RackFace (default to front)")
	}
	face, err := rf.AsFaceEnum()
	if err != nil {
		t.Fatalf("decode RackFace: %v", err)
	}
	if face != nautobotapi.FaceEnumFront {
		t.Errorf("face = %v, want %v", face, nautobotapi.FaceEnumFront)
	}
}

// ---------- MapToNautobotDevice with ProviderMetadata (custom fields) ----------

// TestMapToNautobotDeviceWithCustomFields verifies that MapToNautobotDevice
// flattens a device's nested ProviderMetadata ("nautobot" map) into the request's
// CustomFields, surfacing keys like "xname" and "import_date" at top level.
//
// Why it matters: cani provider metadata (e.g. HPC xnames, import dates) must be
// carried into Nautobot custom fields so exported devices stay traceable to source.
// Inputs: a fully classified CaniDeviceType whose ProviderMetadata nests a
// "nautobot" map. Outputs: a BulkWritableDeviceRequest whose *CustomFields holds
// the flattened keys.
// Data choice: nesting under "nautobot" exercises FlattenProviderMetadata's
// lifting of nested keys; all reference caches are seeded so mapping reaches the
// custom-field step without resolution errors.
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

// ---------- MapToWritableDeviceRequest with custom fields ----------

// TestMapToWritableDeviceRequestCustomFields verifies that
// MapToWritableDeviceRequest populates the request's CustomFields from the
// device's flattened ProviderMetadata (e.g. "custom_key").
//
// Why it matters: the single-create export path (not just bulk) must also carry
// cani provider metadata into Nautobot custom fields for parity and traceability.
// Inputs: a classified CaniDeviceType with ProviderMetadata {"custom_key":
// "custom_val"} and a fully seeded mapper. Outputs: a WritableDeviceRequest whose
// *CustomFields contains custom_key.
// Data choice: a flat top-level metadata key (no "nautobot" nesting) confirms the
// mapper passes through simple custom fields on the single-create request.
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

// TestToNautobotContentTypes verifies that toNautobotContentTypes maps the short
// content-type names device/rack/module to their dcim.* equivalents, passes
// unknown values through unchanged, and preserves order and length.
//
// Why it matters: Nautobot APIs expect fully-qualified content types (e.g.
// "dcim.device"); exporting the wrong string would make content-type assignment
// on statuses/roles fail.
// Inputs (table): single known types, an unknown "ipam.prefix", a mixed slice,
// and an empty slice. Outputs: the converted slice for each case.
// Data choice: the cases cover every known mapping, the default passthrough, a
// mixed combination, and the empty edge case in one table.
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

// TestParentDefFound verifies that parentDef returns a non-nil location-type
// definition (with matching Slug) for a registered slug.
//
// Why it matters: location-tree resolution during export looks up parent
// location-type definitions by slug; a registered slug must resolve so hierarchy
// walks work.
// Inputs: the registered slug "parent-test-type". Outputs: a non-nil
// *LocationTypeDefinition whose Slug equals the input.
// Data choice: registering a unique type first guarantees a deterministic hit
// without depending on globally pre-registered types.
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

// TestParentDefNotFound verifies that parentDef returns nil for a slug that is
// not registered.
//
// Why it matters: an unknown location-type slug must not resolve to a bogus
// definition during export; nil lets callers handle the missing-type case.
// Inputs: the unregistered slug "nonexistent-slug-xyz". Outputs: a nil
// *LocationTypeDefinition.
// Data choice: a deliberately implausible slug ensures it is absent from the
// global registry regardless of other registered types.
func TestParentDefNotFound(t *testing.T) {
	got := parentDef("nonexistent-slug-xyz")
	if got != nil {
		t.Errorf("expected nil for unknown slug, got %+v", got)
	}
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

// ---------- resolveDeviceType error paths ----------

// TestResolveDeviceTypeStrictNoSlug verifies that resolveDeviceType returns the
// error "device type slug is required" when Strict is true and the device has no
// slug or model.
//
// Why it matters: in strict mode the exporter must refuse to guess a device type,
// failing loudly so unclassified hardware is never silently mis-exported.
// Inputs: a CaniDeviceType with empty Slug and Model and a Strict mapper.
// Outputs: a non-nil error with the exact required-slug message.
// Data choice: empty slug+model with Strict=true is the precise condition that
// triggers the hard required-slug failure, and the message is asserted verbatim.
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

// TestResolveDeviceTypeNonStrictNoSlug verifies that resolveDeviceType returns a
// value satisfying errors.Is(err, ErrDeviceUnclassified) when Strict is false and
// the device has no slug or model.
//
// Why it matters: non-strict export must tag unclassifiable devices with a known
// sentinel so callers can skip them gracefully instead of erroring hard.
// Inputs: a CaniDeviceType with empty Slug and Model and a non-strict mapper.
// Outputs: an error wrapping ErrDeviceUnclassified.
// Data choice: empty slug+model with Strict=false isolates the sentinel-error
// branch; using errors.Is tolerates any wrapping around the sentinel.
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

// TestResolveStatusNoDefaultNoCreate verifies that resolveStatus returns the
// error "status is required (use --default-status)" when the device has no status
// and the mapper has no DefaultStatus.
//
// Why it matters: Nautobot devices require a status; the exporter must fail with
// actionable guidance rather than push a device with no status.
// Inputs: an empty CaniDeviceType and a mapper with DefaultStatus "". Outputs: a
// non-nil error with the exact guidance message.
// Data choice: empty device status plus empty default removes every status source,
// forcing the required-status error, which is asserted verbatim.
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

// TestResolveRoleNoDefaultNoCreate verifies that resolveRole returns the error
// "role is required (use --default-role)" when the device has no role and the
// mapper has no DefaultRole.
//
// Why it matters: Nautobot devices require a role; the exporter must fail with a
// clear hint instead of exporting a role-less device.
// Inputs: an empty CaniDeviceType and a mapper with DefaultRole "". Outputs: a
// non-nil error with the exact guidance message.
// Data choice: empty role plus empty default eliminates every role source so the
// required-role error path is the only outcome, asserted verbatim.
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

// TestResolveLocationNoDefaultNoCreate verifies that resolveLocation returns the
// error "location is required (use --default-location)" when the device has no
// location and the mapper has no DefaultLocation.
//
// Why it matters: every exported device must land at a Nautobot location; absent
// any source the exporter must fail with actionable guidance, not a blank location.
// Inputs: an empty CaniDeviceType and a mapper with DefaultLocation "". Outputs: a
// non-nil error with the exact guidance message.
// Data choice: empty location plus empty default removes all location sources,
// forcing the required-location error, which is asserted verbatim.
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

// TestMapToNautobotDeviceNoDeviceType verifies that MapToNautobotDevice returns
// an error wrapping ErrDeviceUnclassified when the device type cannot be
// resolved (empty slug/model).
//
// Why it matters: the bulk export must abort a device whose type is unknown
// rather than emit a request with a missing/zero device-type reference.
// Inputs: a CaniDeviceType with empty Slug and a non-strict mapper whose
// location/status/role are seeded. Outputs: an error matching
// ErrDeviceUnclassified.
// Data choice: only the device-type source is removed (others seeded) so the
// failure is attributable to type resolution.
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
		t.Fatalf("expected ErrDeviceUnclassified, got %v", err)
	}
}

// TestMapToNautobotDeviceNoLocation verifies that MapToNautobotDevice returns a
// non-nil error when the location cannot be resolved.
//
// Why it matters: a device with a resolvable type but no location must not be
// exported, since Nautobot requires a location reference.
// Inputs: a CaniDeviceType (Slug "dt-noloc") with its device type cached but a
// mapper that has no DefaultLocation and no cached location. Outputs: a non-nil
// error.
// Data choice: caching the device type while leaving location unsatisfiable
// pinpoints location resolution as the failing step.
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

// TestMapToNautobotDeviceNoStatus verifies that MapToNautobotDevice returns a
// non-nil error when the status cannot be resolved.
//
// Why it matters: status is mandatory on Nautobot devices, so the exporter must
// stop before building a request that lacks one.
// Inputs: a CaniDeviceType (Slug "dt-nostat") with device type and location
// cached but a mapper with no DefaultStatus. Outputs: a non-nil error.
// Data choice: seeding type and location but not status isolates the status step
// as the cause of the failure.
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

// TestMapToNautobotDeviceNoRole verifies that MapToNautobotDevice returns a
// non-nil error when the role cannot be resolved.
//
// Why it matters: role is required on Nautobot devices, so a device missing it
// must fail mapping rather than export an incomplete request.
// Inputs: a CaniDeviceType (Slug "dt-norole") with device type, location, and
// status cached but a mapper with no DefaultRole. Outputs: a non-nil error.
// Data choice: satisfying every reference except role pins the failure to the
// role resolution step.
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

// TestMapToWritableDeviceRequestNoDeviceType verifies that
// MapToWritableDeviceRequest returns a non-nil error when the device type cannot
// be resolved.
//
// Why it matters: the single-create export path must enforce the same device-type
// requirement as the bulk path so no under-specified device reaches Nautobot.
// Inputs: a CaniDeviceType with no slug and a non-strict mapper carrying default
// location/status/role names. Outputs: a non-nil error.
// Data choice: an unclassifiable device with defaults set elsewhere isolates the
// device-type resolution failure on the single-create code path.
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

// TestMapToWritableDeviceRequestNoLocation verifies that
// MapToWritableDeviceRequest returns a non-nil error when the location cannot be
// resolved.
//
// Why it matters: the single-create path must reject a device with no resolvable
// location, matching Nautobot's location requirement.
// Inputs: a CaniDeviceType (Slug "wr-noloc") with its device type cached but a
// mapper with empty DefaultLocation. Outputs: a non-nil error.
// Data choice: caching only the device type leaves location unsatisfiable,
// pinpointing the location step on the single-create path.
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

// TestMapToWritableDeviceRequestNoStatus verifies that
// MapToWritableDeviceRequest returns a non-nil error when the status cannot be
// resolved.
//
// Why it matters: the single-create path must enforce the mandatory status field
// just like the bulk path before issuing a create.
// Inputs: a CaniDeviceType (Slug "wr-nostat") with device type and location
// cached but a mapper with empty DefaultStatus. Outputs: a non-nil error.
// Data choice: seeding type and location but not status isolates status as the
// failing step on the single-create path.
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

// TestMapToWritableDeviceRequestNoRole verifies that MapToWritableDeviceRequest
// returns a non-nil error when the role cannot be resolved.
//
// Why it matters: the single-create path must enforce the required role field so
// no role-less device is created in Nautobot.
// Inputs: a CaniDeviceType (Slug "wr-norole") with device type, location, and
// status cached but a mapper with empty DefaultRole. Outputs: a non-nil error.
// Data choice: satisfying every reference except role pins the failure to the
// role step on the single-create path.
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

// TestMapToWritableRackRequestNoStatus verifies that MapToWritableRackRequest
// returns a non-nil error when the rack's status cannot be resolved.
//
// Why it matters: racks exported to Nautobot require a status; mapping must fail
// rather than create a rack with a missing status reference.
// Inputs: a CaniDeviceType (Name "Rack-Fail") with its location cached but a
// mapper with empty DefaultStatus. Outputs: a non-nil error.
// Data choice: caching the location while leaving status unsatisfiable isolates
// status resolution as the failure point for the rack request.
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

// ---------- NewExporter ----------

// TestNewExporterOptions verifies NewExporter builds a non-nil Exporter that
// carries the supplied ExporterOpts (DryRun and Merge) unchanged.
//
// Why it matters: export behavior (dry-run vs. write, merge-on-conflict) is
// governed by these options, so the constructor must wire them through
// faithfully.
// Inputs: a nil client, a cache, and opts with DryRun/Merge true. Outputs: an
// Exporter whose Options reflect those flags.
// Data choice: DryRun/Merge are asserted as representative toggles; the nil
// client is fine because no API call is made.
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

// TestMapToPatchRequestSerial verifies MapToPatchRequest copies a device's
// Serial into the patch request as a non-nil pointer.
//
// Why it matters: PATCH updates an existing Nautobot device; serial numbers
// must survive the mapping so inventory stays accurate on re-sync.
// Inputs: a fully cached mapper and a CaniDeviceType with Serial "SN123456".
// Outputs: a patch request whose *Serial equals "SN123456".
// Data choice: device-type/location/status/role are all pre-cached so mapping
// succeeds and the test isolates the Serial field.
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

// TestMapToPatchRequestAssetTag verifies MapToPatchRequest copies a device's
// AssetTag into the patch request as a non-nil pointer.
//
// Why it matters: PATCH updates an existing Nautobot device; asset tags must
// survive the mapping so inventory stays accurate on re-sync.
// Inputs: a fully cached mapper and a device with AssetTag "ASSET-001".
// Outputs: a patch request whose *AssetTag equals "ASSET-001".
// Data choice: all FK caches are pre-seeded so mapping succeeds and the test
// isolates the AssetTag field.
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

// TestMapToPatchRequestComments verifies MapToPatchRequest copies a device's
// Comments into the patch request as a non-nil pointer.
//
// Why it matters: free-form comments carry operator context that must be
// preserved when patching an existing Nautobot device.
// Inputs: a mapper with empty opts and a device carrying only Comments.
// Outputs: a patch request whose *Comments matches the input.
// Data choice: omitting a slug/defaults keeps the test on the comment-copy
// branch with no FK resolution needed.
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

// TestMapToPatchRequestNoSlugSkipsDeviceType verifies that an empty Slug makes
// MapToPatchRequest leave DeviceType nil instead of resolving one.
//
// Why it matters: a PATCH should only touch the device type when the caller
// supplies one; otherwise the existing Nautobot device type must be left
// untouched.
// Inputs: a device with Name set but Slug "". Outputs: a patch request whose
// DeviceType pointer is nil.
// Data choice: an empty slug directly exercises the skip branch without needing
// a populated device-type cache.
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

// ---------- compareRack remote has rack, local doesn't ----------

// TestCompareRackRemoteHasRackLocalDoesNot verifies compareRack emits one
// "rack" FieldDiff when the remote Nautobot device has a rack but the local
// device resolves to none.
//
// Why it matters: drift detection must flag a rack that exists remotely but not
// locally so merges can reconcile placement.
// Inputs: a parent-less CaniDeviceType (local rack "") and a remote Device
// carrying a rack id union. Outputs: a single diff whose Field == "rack".
// Data choice: a nil inventory yields an empty local rack name, so GetRackByName
// is skipped and only the remote-vs-none branch is tested.
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

// TestComparePositionNoLocalPosition verifies comparePosition returns no diffs
// when the local RackPosition is 0 ("no position").
//
// Why it matters: a missing local position must not flag or overwrite the
// remote position during export, avoiding spurious drift.
// Inputs: a device with RackPosition 0 and a remote with no Position. Outputs:
// an empty diff slice.
// Data choice: 0 is the sentinel for "unset", directly exercising the
// early-return guard.
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

// TestComparePositionMismatch verifies comparePosition emits one "position"
// diff (LocalVal "5") when local and remote rack positions differ.
//
// Why it matters: position drift must be surfaced so a device's U-slot in
// Nautobot matches the cani inventory.
// Inputs: local RackPosition 5 vs. remote Position 3. Outputs: a single diff
// with Field "position" and LocalVal "5".
// Data choice: distinct non-zero numbers (5 vs 3) make both the mismatch and
// the string-formatted LocalVal unambiguous.
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

// TestComparePositionMatch verifies comparePosition returns no diffs when the
// local and remote positions are equal.
//
// Why it matters: matching positions must not be reported as drift, keeping
// export summaries free of false positives.
// Inputs: local RackPosition 7 and remote Position 7. Outputs: an empty diff
// slice.
// Data choice: identical non-zero values exercise the equal branch past the
// zero-position guard.
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

// TestCompareFaceNoLocalFace verifies compareFace returns no diffs when the
// local Face is empty ("no face").
//
// Why it matters: an unset face must not flag drift or clobber the remote rack
// face during export.
// Inputs: a device with Face "" and a remote with no Face. Outputs: an empty
// diff slice.
// Data choice: the empty string is the sentinel for "unset", exercising the
// early-return guard.
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

// TestCompareFaceMismatch verifies compareFace emits one "face" diff (LocalVal
// "rear") when local and remote rack faces differ.
//
// Why it matters: face drift (front vs rear) must be surfaced so the exported
// device orientation matches the cani inventory.
// Inputs: local Face "rear" vs. a remote DeviceFace value "front". Outputs: a
// single diff with Field "face" and LocalVal "rear".
// Data choice: opposite faces (rear vs front) make the mismatch and the
// reported LocalVal explicit.
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

// TestCompareFaceMatch verifies compareFace returns no diffs when local and
// remote rack faces are equal.
//
// Why it matters: matching faces must not register as drift, avoiding noise in
// export conflict reporting.
// Inputs: local Face "front" and a remote DeviceFace value "front". Outputs: an
// empty diff slice.
// Data choice: identical non-empty faces exercise the equal branch past the
// empty-face guard.
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

// ---------- SetInventory ----------

// TestSetInventory verifies SetInventory attaches an Inventory to the mapper so
// later rack-name resolution can read it.
//
// Why it matters: resolveLocalRackName (used by rack diffing) depends on the
// mapper holding the local inventory; without it, rack comparisons cannot run.
// Inputs: a mapper and an Inventory with empty Devices/Racks maps. Outputs: a
// mapper whose inventory field is non-nil.
// Data choice: empty maps are enough to assert the wiring without exercising
// rack resolution itself.
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

// TestMapToWritableDeviceRequestSerial verifies MapToWritableDeviceRequest
// copies Serial, AssetTag, and Comments into the create request.
//
// Why it matters: creating a new Nautobot device must carry over all
// identifying metadata so the exported record matches cani inventory.
// Inputs: a fully cached mapper and a device with Serial/AssetTag/Comments set.
// Outputs: a request whose *Serial, *AssetTag, and *Comments match.
// Data choice: all FK caches are pre-seeded so the create path succeeds and the
// test focuses on the optional string fields.
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

// TestRefIDNilInput verifies refID returns uuid.Nil for a nil reference union.
//
// Why it matters: remote Nautobot references are frequently nil; refID must
// degrade to uuid.Nil so diffing/mapping treats them as "absent" rather than
// panicking.
// Inputs: a nil *BulkWritableCableRequestStatusId. Outputs: uuid.Nil.
// Data choice: nil is the boundary input that the guard clause protects
// against.
func TestRefIDNilInput(t *testing.T) {
	result := refID(nil)
	if result != uuid.Nil {
		t.Errorf("expected uuid.Nil for nil input, got %s", result)
	}
}

// TestTenantRefIDNilInput verifies tenantRefID returns uuid.Nil for a nil
// reference union (it delegates to refID).
//
// Why it matters: tenant/rack references on remote devices may be nil; the
// helper must yield uuid.Nil so comparisons treat them as unset.
// Inputs: a nil *BulkWritableCableRequestStatusId. Outputs: uuid.Nil.
// Data choice: nil is the boundary input exercising the delegated guard.
func TestTenantRefIDNilInput(t *testing.T) {
	result := tenantRefID(nil)
	if result != uuid.Nil {
		t.Errorf("expected uuid.Nil for nil input, got %s", result)
	}
}

// ---------- LoadResult / ConflictInfo structs ----------

// TestLoadResultZeroValue verifies a freshly constructed LoadResult has zero
// counters (CablesCreated, IfacesCreated).
//
// Why it matters: summary tallies start empty and are incremented during load;
// a non-zero zero-value would corrupt every export summary.
// Inputs: an empty &LoadResult{}. Outputs: zero-valued counter fields.
// Data choice: checking two representative counters confirms the struct's zero
// value without enumerating every field.
func TestLoadResultZeroValue(t *testing.T) {
	result := &LoadResult{}
	if result.CablesCreated != 0 {
		t.Error("expected zero value for CablesCreated")
	}
	if result.IfacesCreated != 0 {
		t.Error("expected zero value for IfacesCreated")
	}
}

// TestConflictInfoDiffs verifies a ConflictInfo carries its FieldDiff slice
// intact (length and Field value).
//
// Why it matters: ConflictInfo is what the exporter reports for devices that
// differ from Nautobot; its diffs must round-trip so operators see what would
// change under --merge.
// Inputs: a ConflictInfo literal with one "status" FieldDiff. Outputs: a slice
// of length 1 whose Field is "status".
// Data choice: a single status diff is the minimal case proving the slice field
// is populated and readable.
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

// TestPendingMoveFields verifies a pendingMove literal preserves its nested
// slotKey From/To fields (position and face).
//
// Why it matters: pendingMove drives rack relocation during export; mis-wired
// From/To fields would move a device to the wrong slot in Nautobot.
// Inputs: a pendingMove with From{Position 1} and To{Face "rear"}. Outputs:
// those nested fields read back unchanged.
// Data choice: differing From/To racks, positions, and faces ensure each nested
// field is checked independently.
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

// TestExporterOptsAllFlags verifies an ExporterOpts literal stores its
// representative flags (DefaultLocation, CreateDeviceTypes, Strict).
//
// Why it matters: these flags gate whether the exporter creates missing
// Nautobot prerequisites and how strictly it treats conflicts.
// Inputs: an ExporterOpts with all create/strict flags set. Outputs: those
// fields read back with the assigned values.
// Data choice: asserting a string, a create-flag, and Strict samples the
// distinct flag categories without restating every field.
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

// ---------- printLoadSummary ----------

// TestPrintLoadSummaryAllBranches verifies printLoadSummary runs without
// panicking when every result category is populated.
//
// Why it matters: the end-of-export summary must render every created/skipped
// tally (locations, racks, devices, interfaces, IPAM, ...) so operators get a
// complete report.
// Inputs: a LoadResult with all slices/counters non-empty. Outputs: console
// output only (no return value); the test asserts no panic.
// Data choice: populating every field forces each conditional print branch to
// execute at least once.
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

// TestPrintLoadSummaryEmpty verifies printLoadSummary runs without panicking
// when the result is empty (most branches skipped).
//
// Why it matters: an export that changed nothing must still print a clean
// summary rather than crash on empty slices/zero counters.
// Inputs: an empty &LoadResult{}. Outputs: console output only; the test
// asserts no panic.
// Data choice: the zero-value result exercises the all-branches-skipped path,
// complementing the all-populated test.
func TestPrintLoadSummaryEmpty(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}
	result := &LoadResult{}

	// Empty result — most branches are skipped
	e.printLoadSummary(result)
}

// ---------- printConflictDiffs ----------

// TestPrintConflictDiffsNoDiffs verifies printConflictDiffs returns early
// (prints nothing) when no conflict carries any FieldDiff.
//
// Why it matters: the "pending changes" banner should only appear when there is
// actually something to merge, keeping export output uncluttered.
// Inputs: a LoadResult with one conflict whose Diffs is nil. Outputs: console
// output only; the test asserts no panic on the early-return path.
// Data choice: a conflict with nil Diffs is the exact condition that triggers
// the early return.
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

// TestPrintConflictDiffsWithDiffs verifies printConflictDiffs prints the diff
// banner and per-device diffs when a conflict has FieldDiffs.
//
// Why it matters: operators rely on this output to see exactly which fields
// would change under --merge before mutating Nautobot.
// Inputs: a LoadResult with one conflict carrying two FieldDiffs. Outputs:
// console output only; the test asserts no panic.
// Data choice: two diffs (status, role) exercise the per-diff loop more than
// once.
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

// TestPrintConflictDiffsEmptyConflicts verifies printConflictDiffs handles an
// empty Conflicts slice without panicking (nothing to print).
//
// Why it matters: a clean export with zero conflicts must not crash the summary
// stage on an empty slice.
// Inputs: a LoadResult with Conflicts set to an empty slice. Outputs: console
// output only; the test asserts no panic.
// Data choice: an explicitly empty (non-nil) slice exercises the no-iteration
// path of the conflict loop.
func TestPrintConflictDiffsEmptyConflicts(t *testing.T) {
	e := &Exporter{Options: &ExporterOpts{}}
	result := &LoadResult{
		Conflicts: []ConflictInfo{},
	}
	e.printConflictDiffs(result)
}

// ---------- remoteSlotKey additional cases ----------

// TestRemoteSlotKeyValidFront verifies remoteSlotKey extracts RackID, Position,
// and a defaulted "front" Face from a Nautobot Device whose Face field is nil.
//
// Why it matters: slot keys identify where a device currently sits in Nautobot
// so the exporter can detect when cani inventory implies a rack-slot move; a nil
// Face must default to "front" to match Nautobot's own default.
// Inputs: a *Device with Position=10, a Rack Id union, and Face=nil. Outputs: a
// *slotKey with RackID set, Position=10, Face="front".
// Data choice: Face is left nil specifically to exercise the front-default path.
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

// TestRemoteSlotKeyValidRear verifies remoteSlotKey reports Face="rear" when the
// Device's Face value is explicitly "rear".
//
// Why it matters: front vs. rear placement is part of a slot's identity in
// Nautobot, so a misread face would make the exporter compute spurious moves.
// Inputs: a *Device with Position=5, a Rack Id union, and Face.Value="rear".
// Outputs: a *slotKey whose Face is "rear".
// Data choice: Face.Value="rear" is the only non-default branch, so it is the
// single field this test asserts.
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

// TestRemoteSlotKeyNilPosition verifies remoteSlotKey returns nil when the Device
// has no rack Position.
//
// Why it matters: an unracked device has no slot to compare against, so the
// exporter must skip it rather than fabricate a zero-position slot key.
// Inputs: a *Device with Position=nil but a valid Rack. Outputs: nil.
// Data choice: only Position is nil (Rack stays valid) to isolate the
// missing-position guard from the missing-rack guard.
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

// TestRemoteSlotKeyNilRack verifies remoteSlotKey returns nil when the Device has
// a Position but no Rack.
//
// Why it matters: a position without a rack is meaningless for move detection,
// so the exporter must treat it as "no slot" instead of dereferencing nil.
// Inputs: a *Device with Position=3 and Rack=nil. Outputs: nil.
// Data choice: Rack is nil while Position is set to isolate the missing-rack
// guard from the missing-position guard.
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

// TestMapInterfaceType25g verifies mapInterfaceType passes "25gbase-x-sfp28"
// through unchanged.
//
// Why it matters: interface types must land on exact Nautobot
// InterfaceTypeChoices slugs or interface creation is rejected; 25G already
// matches Nautobot's slug and must not be rewritten.
// Inputs: the devicetypes string "25gbase-x-sfp28". Outputs: the identical
// Nautobot slug "25gbase-x-sfp28".
// Data choice: 25G is a common SFP28 fabric link whose cani and Nautobot names
// coincide, covering the identity branch of the mapping.
func TestMapInterfaceType25g(t *testing.T) {
	got := mapInterfaceType("25gbase-x-sfp28")
	if got != "25gbase-x-sfp28" {
		t.Errorf("got %q, want '25gbase-x-sfp28'", got)
	}
}

// TestMapInterfaceType40g verifies mapInterfaceType passes "40gbase-x-qsfpp"
// through unchanged.
//
// Why it matters: 40G QSFP+ uplinks must export with Nautobot's exact slug or
// the interface is rejected on create.
// Inputs: "40gbase-x-qsfpp". Outputs: the identical "40gbase-x-qsfpp".
// Data choice: 40G QSFP+ is a standard switch uplink whose cani and Nautobot
// slugs already match, exercising the pass-through branch.
func TestMapInterfaceType40g(t *testing.T) {
	got := mapInterfaceType("40gbase-x-qsfpp")
	if got != "40gbase-x-qsfpp" {
		t.Errorf("got %q, want '40gbase-x-qsfpp'", got)
	}
}

// TestMapInterfaceType100g verifies mapInterfaceType passes "100gbase-x-qsfp28"
// through unchanged.
//
// Why it matters: 100G fabric links must export with the exact Nautobot slug to
// be accepted on create.
// Inputs: "100gbase-x-qsfp28". Outputs: the identical "100gbase-x-qsfp28".
// Data choice: 100G QSFP28 is the dominant HPC fabric speed whose names match
// across cani and Nautobot, covering the identity branch.
func TestMapInterfaceType100g(t *testing.T) {
	got := mapInterfaceType("100gbase-x-qsfp28")
	if got != "100gbase-x-qsfp28" {
		t.Errorf("got %q, want '100gbase-x-qsfp28'", got)
	}
}

// TestMapInterfaceType200g verifies mapInterfaceType passes "200gbase-x-qsfp56"
// through unchanged.
//
// Why it matters: 200G QSFP56 links must export with Nautobot's exact slug to
// avoid rejection.
// Inputs: "200gbase-x-qsfp56". Outputs: the identical "200gbase-x-qsfp56".
// Data choice: 200G QSFP56 is a high-speed fabric link whose cani and Nautobot
// slugs coincide, exercising the pass-through branch.
func TestMapInterfaceType200g(t *testing.T) {
	got := mapInterfaceType("200gbase-x-qsfp56")
	if got != "200gbase-x-qsfp56" {
		t.Errorf("got %q, want '200gbase-x-qsfp56'", got)
	}
}

// TestMapInterfaceType100baseTx verifies mapInterfaceType passes "100base-tx"
// through unchanged.
//
// Why it matters: low-speed management/copper ports must still export with the
// exact Nautobot slug.
// Inputs: "100base-tx". Outputs: the identical "100base-tx".
// Data choice: 100BASE-TX is a typical management-port type whose names already
// match, covering the identity branch for a non-fabric speed.
func TestMapInterfaceType100baseTx(t *testing.T) {
	got := mapInterfaceType("100base-tx")
	if got != "100base-tx" {
		t.Errorf("got %q, want '100base-tx'", got)
	}
}

// TestMapInterfaceType1gBaseT verifies mapInterfaceType rewrites the devicetypes
// alias "1gbase-t" to Nautobot's canonical "1000base-t".
//
// Why it matters: the cani devicetype library and Nautobot disagree on the 1G
// copper slug, so the exporter must translate or interface creation fails.
// Inputs: "1gbase-t". Outputs: "1000base-t".
// Data choice: 1G copper is the common speed where the two vocabularies differ,
// making it the key non-identity mapping branch.
func TestMapInterfaceType1gBaseT(t *testing.T) {
	got := mapInterfaceType("1gbase-t")
	if got != "1000base-t" {
		t.Errorf("got %q, want '1000base-t'", got)
	}
}

// TestMapInterfaceTypeInfinibandHDR verifies mapInterfaceType passes
// "infiniband-hdr" through unchanged.
//
// Why it matters: InfiniBand HDR fabric ports must export with Nautobot's exact
// slug; HDR must not be swallowed by the NDR case that is matched first.
// Inputs: "infiniband-hdr". Outputs: the identical "infiniband-hdr".
// Data choice: HDR is checked after NDR in the switch, so this input (lacking
// "ndr") confirms the HDR branch is actually reached.
func TestMapInterfaceTypeInfinibandHDR(t *testing.T) {
	got := mapInterfaceType("infiniband-hdr")
	if got != "infiniband-hdr" {
		t.Errorf("got %q, want 'infiniband-hdr'", got)
	}
}

// TestMapInterfaceType400gQSFPDD verifies mapInterfaceType folds the QSFP-DD form
// factor "400gbase-x-qsfpdd" into Nautobot's "400gbase-x-osfp".
//
// Why it matters: cani may emit either 400G form factor, but the exporter
// normalizes both onto the single 400G slug it sends to Nautobot on create.
// Inputs: "400gbase-x-qsfpdd". Outputs: "400gbase-x-osfp".
// Data choice: QSFP-DD is the alternate 400G connector, chosen to prove the
// normalization branch rather than a pass-through.
func TestMapInterfaceType400gQSFPDD(t *testing.T) {
	got := mapInterfaceType("400gbase-x-qsfpdd")
	if got != "400gbase-x-osfp" {
		t.Errorf("got %q, want '400gbase-x-osfp'", got)
	}
}

// ---------- resolveCableType slug-based fallback ----------

// TestResolveCableTypeSlugCat verifies resolveCableType's slug fallback maps any
// slug containing "cat" to a non-nil Cat5e cable-type choice.
//
// Why it matters: legacy cani cables carry only a slug, so the exporter must
// still resolve a Nautobot CableTypeChoices value to create the cable.
// Inputs: a CaniCableType with Slug "some-cat5e-cable" and no explicit type,
// category, or connector. Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: the slug embeds "cat" with the higher-priority fields empty,
// forcing resolution down to the slug branch.
func TestResolveCableTypeSlugCat(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "some-cat5e-cable"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'cat'")
	}
}

// TestResolveCableTypeSlugDAC verifies resolveCableType's slug fallback maps a
// slug containing "dac" to a non-nil passive-DAC cable-type choice.
//
// Why it matters: direct-attach copper cables are common in racks and must still
// resolve from a bare slug for export.
// Inputs: a CaniCableType with Slug "hpe-dac-1m" and the other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: "hpe-dac-1m" embeds "dac" while leaving the higher-priority fields
// empty, isolating the DAC slug branch.
func TestResolveCableTypeSlugDAC(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "hpe-dac-1m"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'dac'")
	}
}

// TestResolveCableTypeSlugAOC verifies resolveCableType's slug fallback maps a
// slug containing "aoc" to a non-nil active-optical cable-type choice.
//
// Why it matters: active optical cables must resolve from a bare slug so the
// exporter can still classify them in Nautobot.
// Inputs: a CaniCableType with Slug "mellanox-aoc-10m" and other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: "mellanox-aoc-10m" embeds "aoc" with no higher-priority fields,
// exercising only the AOC slug branch.
func TestResolveCableTypeSlugAOC(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "mellanox-aoc-10m"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'aoc'")
	}
}

// TestResolveCableTypeSlugFiber verifies resolveCableType's slug fallback maps a
// slug containing "fiber" to a non-nil multimode (MMF-OM4) cable-type choice.
//
// Why it matters: generically named fiber cables must still resolve to a
// concrete Nautobot type for export.
// Inputs: a CaniCableType with Slug "fiber-lc-om4-5m" and other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: the slug contains "fiber" (which shares the MMF branch) while
// leaving higher-priority fields empty.
func TestResolveCableTypeSlugFiber(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "fiber-lc-om4-5m"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'fiber'")
	}
}

// TestResolveCableTypeSlugMMF verifies resolveCableType's slug fallback maps a
// slug containing "mmf" to a non-nil multimode (MMF-OM4) cable-type choice.
//
// Why it matters: multimode fiber cables identified only by slug must still be
// classified so the exporter can create them.
// Inputs: a CaniCableType with Slug "mmf-om3-cable" and other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: "mmf" shares the fiber branch; because the value is in Slug (not
// CableCategory), resolution falls through to the slug heuristic.
func TestResolveCableTypeSlugMMF(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "mmf-om3-cable"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'mmf'")
	}
}

// TestResolveCableTypeConnectorLC verifies resolveCableType uses the connector
// heuristic to map an "lc" connector to a non-nil (SMF-OS2) cable type.
//
// Why it matters: when a cani cable lacks an explicit type and category, the
// connector is the best signal for picking a Nautobot cable type on export.
// Inputs: a CaniCableType with ConnectorType "lc" and other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: only ConnectorType is set so resolution skips the type/category
// tiers and exercises the connector-heuristic branch.
func TestResolveCableTypeConnectorLC(t *testing.T) {
	cable := &devicetypes.CaniCableType{ConnectorType: "lc"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for LC connector")
	}
}

// TestResolveCableTypeConnectorMPO verifies resolveCableType uses the connector
// heuristic to map an "mpo" connector to a non-nil (MMF-OM4) cable type.
//
// Why it matters: MPO trunks are common in fabrics, so the connector heuristic
// must classify them when no explicit type/category is present.
// Inputs: a CaniCableType with ConnectorType "mpo" and other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: ConnectorType alone is set, isolating the connector branch for a
// multimode connector distinct from the LC (single-mode) case.
func TestResolveCableTypeConnectorMPO(t *testing.T) {
	cable := &devicetypes.CaniCableType{ConnectorType: "mpo"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for MPO connector")
	}
}

// TestResolveCableTypeCategoryMMF verifies resolveCableType resolves the
// CableCategory "mmf-om3" via the category map to a non-nil cable type.
//
// Why it matters: when a cani cable names its category, the exporter should use
// that exact grade rather than guessing from a connector or slug.
// Inputs: a CaniCableType with CableCategory "mmf-om3" and other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: "mmf-om3" is a real key in cableTypeMap, confirming the
// category-lookup tier resolves a specific multimode grade.
func TestResolveCableTypeCategoryMMF(t *testing.T) {
	cable := &devicetypes.CaniCableType{CableCategory: "mmf-om3"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for mmf-om3 category")
	}
}

// TestResolveCableTypeCategorySMF verifies resolveCableType resolves the
// CableCategory "smf-os2" via the category map to a non-nil cable type.
//
// Why it matters: single-mode fiber grades must resolve from the named category
// so long-haul links export with the correct Nautobot type.
// Inputs: a CaniCableType with CableCategory "smf-os2" and other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: "smf-os2" is a real cableTypeMap key, pairing with the MMF case
// to cover both fiber families through the category tier.
func TestResolveCableTypeCategorySMF(t *testing.T) {
	cable := &devicetypes.CaniCableType{CableCategory: "smf-os2"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for smf-os2 category")
	}
}

// ---------- getSpeedForType additional ----------

// TestGetSpeedForType200g verifies getSpeedForType returns 200000000 Kbps for
// "200gbase-x-qsfp56".
//
// Why it matters: Nautobot interfaces carry an explicit speed, so the exporter
// must derive the right Kbps from the type or it records wrong link speeds.
// Inputs: "200gbase-x-qsfp56". Outputs: 200000000 (200 Gbps in Kbps).
// Data choice: 200G shares its speed case with InfiniBand HDR, so this asserts
// the Ethernet half of that combined branch.
func TestGetSpeedForType200g(t *testing.T) {
	if got := getSpeedForType("200gbase-x-qsfp56"); got != 200000000 {
		t.Errorf("got %d, want 200000000", got)
	}
}

// TestGetSpeedForTypeInfinibandHDR verifies getSpeedForType returns 200000000
// Kbps for "infiniband-hdr".
//
// Why it matters: InfiniBand HDR ports must report 200 Gbps in Nautobot just
// like 200G Ethernet, so fabric link speeds stay accurate.
// Inputs: "infiniband-hdr". Outputs: 200000000 (200 Gbps in Kbps).
// Data choice: HDR is the InfiniBand member of the shared 200G speed case,
// confirming both members resolve identically.
func TestGetSpeedForTypeInfinibandHDR(t *testing.T) {
	if got := getSpeedForType("infiniband-hdr"); got != 200000000 {
		t.Errorf("got %d, want 200000000", got)
	}
}

// TestGetSpeedForTypeInfinibandNDR verifies getSpeedForType returns 400000000
// Kbps for "infiniband-ndr".
//
// Why it matters: InfiniBand NDR ports must report 400 Gbps so the exporter
// records the correct fabric speed in Nautobot.
// Inputs: "infiniband-ndr". Outputs: 400000000 (400 Gbps in Kbps).
// Data choice: NDR shares the 400G speed case with 400G Ethernet, so this
// asserts the InfiniBand member of that branch.
func TestGetSpeedForTypeInfinibandNDR(t *testing.T) {
	if got := getSpeedForType("infiniband-ndr"); got != 400000000 {
		t.Errorf("got %d, want 400000000", got)
	}
}

// TestGetSpeedForType400gOSFP verifies getSpeedForType returns 400000000 Kbps for
// "400gbase-x-osfp".
//
// Why it matters: 400G OSFP links are the fastest exported ports and must carry
// the correct 400 Gbps speed in Nautobot.
// Inputs: "400gbase-x-osfp". Outputs: 400000000 (400 Gbps in Kbps).
// Data choice: OSFP is the Ethernet member of the shared 400G case (with QSFP-DD
// and InfiniBand NDR), asserting that branch from the Ethernet side.
func TestGetSpeedForType400gOSFP(t *testing.T) {
	if got := getSpeedForType("400gbase-x-osfp"); got != 400000000 {
		t.Errorf("got %d, want 400000000", got)
	}
}

// TestGetSpeedForType25g verifies getSpeedForType returns 25000000 Kbps for
// "25gbase-x-sfp28".
//
// Why it matters: 25G server links must report the correct speed so Nautobot
// reflects real edge connectivity.
// Inputs: "25gbase-x-sfp28". Outputs: 25000000 (25 Gbps in Kbps).
// Data choice: 25G has its own standalone speed case, chosen to cover a
// mid-range edge speed distinct from the shared high-speed branches.
func TestGetSpeedForType25g(t *testing.T) {
	if got := getSpeedForType("25gbase-x-sfp28"); got != 25000000 {
		t.Errorf("got %d, want 25000000", got)
	}
}

// TestGetSpeedForTypeUnknown verifies getSpeedForType falls back to 1000000 Kbps
// (1 Gbps) for an unrecognized interface type.
//
// Why it matters: unknown types must still yield a sane default speed so the
// exporter can create the interface rather than failing or sending zero.
// Inputs: "some-unknown-type". Outputs: 1000000 (default 1 Gbps in Kbps).
// Data choice: a deliberately bogus type string forces the switch's default
// branch.
func TestGetSpeedForTypeUnknown(t *testing.T) {
	if got := getSpeedForType("some-unknown-type"); got != 1000000 {
		t.Errorf("got %d, want 1000000 (default)", got)
	}
}

// ---------- containsInfiniband additional ----------

// TestContainsInfinibandModel verifies containsInfiniband returns true when a
// device model name contains "infiniband".
//
// Why it matters: the exporter uses this signal to decide whether to synthesize
// InfiniBand interfaces, so IB hardware must be recognized by model name.
// Inputs: a CaniDeviceType with Model "Mellanox InfiniBand Switch". Outputs:
// true.
// Data choice: a realistic Mellanox IB switch name exercises the literal
// "infiniband" substring match (case-insensitive).
func TestContainsInfinibandModel(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "Mellanox InfiniBand Switch"}
	if !containsInfiniband(dev) {
		t.Error("expected true for model containing 'infiniband'")
	}
}

// TestContainsInfinibandNDR verifies containsInfiniband returns true when a model
// name contains the generation token "ndr".
//
// Why it matters: IB hardware is often named by generation rather than the word
// "infiniband", so the exporter must also key off "ndr".
// Inputs: a CaniDeviceType with Model "NDR-400G Switch". Outputs: true.
// Data choice: an NDR-generation switch name lacking "infiniband" proves the
// "ndr" substring branch independently.
func TestContainsInfinibandNDR(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "NDR-400G Switch"}
	if !containsInfiniband(dev) {
		t.Error("expected true for model containing 'ndr'")
	}
}

// TestContainsInfinibandHDR verifies containsInfiniband returns true when a model
// name contains the generation token "hdr".
//
// Why it matters: HDR-generation IB switches must be detected so the exporter
// adds their fabric interfaces.
// Inputs: a CaniDeviceType with Model "QM8700 HDR Switch". Outputs: true.
// Data choice: the QM8700 is a real HDR switch; its name carries "hdr" but not
// "infiniband", isolating the "hdr" branch.
func TestContainsInfinibandHDR(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "QM8700 HDR Switch"}
	if !containsInfiniband(dev) {
		t.Error("expected true for model containing 'hdr'")
	}
}

// TestContainsInfinibandMCX verifies containsInfiniband returns true when a model
// name contains the adapter token "mcx".
//
// Why it matters: Mellanox ConnectX adapters use "MCX" part numbers, so the
// exporter must recognize them as InfiniBand-capable.
// Inputs: a CaniDeviceType with Model "MCX653105A-HDAT". Outputs: true.
// Data choice: a genuine ConnectX-6 part number exercises the "mcx" branch via a
// raw SKU rather than a descriptive name.
func TestContainsInfinibandMCX(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "MCX653105A-HDAT"}
	if !containsInfiniband(dev) {
		t.Error("expected true for model containing 'mcx'")
	}
}

// TestContainsInfinibandFalse verifies containsInfiniband returns false for a
// standard Ethernet server model.
//
// Why it matters: false positives would make the exporter add phantom InfiniBand
// interfaces to non-IB hardware.
// Inputs: a CaniDeviceType with Model "HPE ProLiant DL380". Outputs: false.
// Data choice: a ubiquitous non-IB server whose name contains none of the
// infiniband/ndr/hdr/mcx tokens, covering the negative path.
func TestContainsInfinibandFalse(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "HPE ProLiant DL380"}
	if containsInfiniband(dev) {
		t.Error("expected false for standard server model")
	}
}

// ---------- derefString ----------

// TestDerefStringNil verifies derefString returns "" for a nil *string.
//
// Why it matters: many Nautobot API fields are optional pointers, so the
// exporter relies on derefString to read them without nil-pointer panics.
// Inputs: a nil *string. Outputs: "".
// Data choice: nil is the only input that triggers the guard clause.
func TestDerefStringNil(t *testing.T) {
	if got := derefString(nil); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

// TestDerefStringNonNil verifies derefString returns the pointed-to value for a
// non-nil *string.
//
// Why it matters: when an optional Nautobot field is present, the exporter must
// read its actual value, not the empty fallback.
// Inputs: a *string pointing at "hello". Outputs: "hello".
// Data choice: a simple non-empty literal confirms the value is dereferenced
// rather than dropped.
func TestDerefStringNonNil(t *testing.T) {
	s := "hello"
	if got := derefString(&s); got != "hello" {
		t.Errorf("got %q, want 'hello'", got)
	}
}

// ---------- isValidNautobotInterfaceType ----------

// TestIsValidNautobotInterfaceTypeValid verifies isValidNautobotInterfaceType
// returns true for every slug in Nautobot's accepted InterfaceTypeChoices set.
//
// Why it matters: the exporter gates interface creation on this whitelist, so
// every legitimate type must pass or valid interfaces would be skipped.
// Inputs: a slice of accepted slugs (100base-tx … infiniband-ndr, virtual, lag,
// other). Outputs: true for each.
// Data choice: the list mirrors every speed plus virtual/lag/other choices the
// exporter emits, asserting full positive coverage of the whitelist.
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

// TestIsValidNautobotInterfaceTypeInvalid verifies isValidNautobotInterfaceType
// returns false for strings outside the accepted choice set.
//
// Why it matters: rejecting unknown types prevents the exporter from sending
// values Nautobot would refuse on create.
// Inputs: "not-a-type", "ethernet", "10gbase-t", and "". Outputs: false for each.
// Data choice: the cases include a bogus string, a too-generic name, a
// plausible-but-unsupported copper slug, and the empty string to cover near
// misses and the zero value.
func TestIsValidNautobotInterfaceTypeInvalid(t *testing.T) {
	invalidTypes := []string{"not-a-type", "ethernet", "10gbase-t", ""}
	for _, vt := range invalidTypes {
		if isValidNautobotInterfaceType(vt) {
			t.Errorf("expected false for %q", vt)
		}
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
	resetIPAMCaches()
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
	resetIPAMCaches()
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
	resetIPAMCaches()
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
	resetIPAMCaches()
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

// TestPrintDeviceDiffsNoDiffs verifies printDeviceDiffs returns early without
// printing when the diff slice is nil or empty.
//
// Why it matters: when a local device already matches its Nautobot counterpart
// the export must stay silent, so the sync summary never lists phantom changes.
// Inputs: a device name plus nil and an empty []FieldDiff. Outputs: none; the
// function returns before touching the colored logger.
// Data choice: both nil and an empty slice are exercised to cover the two
// "no diffs" shapes callers produce. NOTE: a smoke test — it only confirms no
// panic and asserts nothing.
func TestPrintDeviceDiffsNoDiffs(t *testing.T) {
	// Should not panic or output when diffs is empty
	printDeviceDiffs("test-device", nil)
	printDeviceDiffs("test-device", []FieldDiff{})
}

// TestPrintDeviceDiffsWithDiffs verifies printDeviceDiffs runs without panicking
// when given a non-empty list of field diffs.
//
// Why it matters: diff output is how a dry-run export tells operators which
// fields would change on a Nautobot device before --merge is applied.
// Inputs: a device name and three FieldDiffs (status, role, location).
// Outputs: colored log lines only — there is no return value to assert.
// Data choice: three commonly drifted fields exercise the print loop body.
// NOTE: a smoke test — it asserts nothing beyond not panicking.
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

// TestGetSpeedForType40g verifies getSpeedForType maps the 40G QSFP+ interface
// type string to 40000000 kbps.
//
// Why it matters: Nautobot stores interface speed in kbps, so cani must convert
// each cani interface type to the correct numeric speed when exporting NICs.
// Inputs: the type slug "40gbase-x-qsfpp". Outputs: the integer speed in kbps.
// Data choice: the canonical 40G QSFP+ slug pins the high-speed mapping that
// fabric uplinks rely on.
func TestGetSpeedForType40g(t *testing.T) {
	if got := getSpeedForType("40gbase-x-qsfpp"); got != 40000000 {
		t.Errorf("got %d, want 40000000", got)
	}
}

// TestGetSpeedForType100g verifies getSpeedForType maps the 100G QSFP28
// interface type string to 100000000 kbps.
//
// Why it matters: exported interface speeds must match Nautobot's kbps field so
// 100G fabric links report their real bandwidth.
// Inputs: the type slug "100gbase-x-qsfp28". Outputs: the integer kbps speed.
// Data choice: the 100G QSFP28 slug is the most common high-speed switch uplink,
// making it the key mapping to guard.
func TestGetSpeedForType100g(t *testing.T) {
	if got := getSpeedForType("100gbase-x-qsfp28"); got != 100000000 {
		t.Errorf("got %d, want 100000000", got)
	}
}

// TestGetSpeedForType10g verifies getSpeedForType maps the 10G SFP+ interface
// type string to 10000000 kbps.
//
// Why it matters: server and ToR links are frequently 10G, so this conversion
// keeps exported Nautobot interface speeds accurate.
// Inputs: the type slug "10gbase-x-sfpp". Outputs: the integer kbps speed.
// Data choice: the 10G SFP+ slug represents the common mid-speed case between
// the 1G default and the 40G/100G fabric tiers.
func TestGetSpeedForType10g(t *testing.T) {
	if got := getSpeedForType("10gbase-x-sfpp"); got != 10000000 {
		t.Errorf("got %d, want 10000000", got)
	}
}

// TestGetSpeedForType100baseTx verifies getSpeedForType maps the 100BASE-TX
// copper interface type string to 100000 kbps.
//
// Why it matters: management/BMC ports are often 100M copper, and exporting the
// wrong speed would misrepresent them in Nautobot.
// Inputs: the type slug "100base-tx". Outputs: the integer kbps speed.
// Data choice: 100BASE-TX is the slowest mapped tier, guarding the low end away
// from the 1Gbps default fallback.
func TestGetSpeedForType100baseTx(t *testing.T) {
	if got := getSpeedForType("100base-tx"); got != 100000 {
		t.Errorf("got %d, want 100000", got)
	}
}

// ---------- MapToWritableRackRequest custom fields filtering ----------

// TestMapToWritableRackRequestCustomFieldsFiltered verifies
// MapToWritableRackRequest drops u_height and rack_position from CustomFields
// while preserving genuine custom keys.
//
// Why it matters: u_height/rack_position map to first-class Nautobot rack
// fields, so leaking them into CustomFields would duplicate data and risk a
// schema-invalid rack create.
// Inputs: a CaniDeviceType whose ProviderMetadata mixes the reserved keys with
// custom_key/site_code. Outputs: a WritableRackRequest with filtered
// CustomFields.
// Data choice: the metadata deliberately blends reserved and arbitrary keys to
// prove only the reserved ones are stripped.
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

// TestMapToWritableRackRequestDefaultHeight verifies MapToWritableRackRequest
// defaults UHeight to 48 when the device carries no u_height metadata.
//
// Why it matters: a rack created in Nautobot needs a sane height, so cani must
// supply a default rather than emit a zero-height rack.
// Inputs: a CaniDeviceType with no ProviderMetadata. Outputs: a
// WritableRackRequest whose UHeight is 48.
// Data choice: omitting all metadata isolates the default-height branch with no
// other fields interfering.
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

// TestMapToWritableDeviceRequestNoParentNoRack verifies
// MapToWritableDeviceRequest leaves Rack and Position unset when the device has
// no parent (Parent == uuid.Nil).
//
// Why it matters: only rack-mounted devices should carry rack/position in
// Nautobot; an unparented device must export without a phantom rack assignment.
// Inputs: a device with Parent=uuid.Nil but RackPosition/Face set, plus a
// seeded cache and inventory. Outputs: a request with nil Rack and Position.
// Data choice: RackPosition/Face are populated to prove they are ignored when
// no parent resolves to a rack, avoiding the nil-client GetRackByName path.
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

// TestResolveLocationNameFromExporter verifies Exporter.resolveLocationName
// returns the location's name for a UUID present in the inventory.
//
// Why it matters: VLAN/prefix export scopes objects to a location by name, so
// resolving the cani location UUID to its Nautobot name must succeed.
// Inputs: a location UUID and an inventory containing that location ("Floor-42").
// Outputs: the resolved name string and a nil error.
// Data choice: a single named location keeps the happy path unambiguous.
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

// TestResolveLocationNameFromExporterNotFound verifies
// Exporter.resolveLocationName returns an error when the UUID is absent from the
// inventory.
//
// Why it matters: a dangling location reference must fail loudly during export
// instead of silently scoping data to the wrong (empty) location.
// Inputs: a random UUID and an inventory with an empty Locations map.
// Outputs: a non-nil error.
// Data choice: the empty map guarantees the lookup misses, isolating the error
// branch.
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

// TestTopologicalSortLocationsEmpty verifies topologicalSortLocations returns an
// empty slice for an empty location map.
//
// Why it matters: locations are created parent-first in Nautobot, and the empty
// case must not panic or fabricate entries when there is nothing to export.
// Inputs: an empty map[uuid.UUID]*CaniLocationType. Outputs: a length-0 slice.
// Data choice: the empty map is the boundary condition for the sort.
func TestTopologicalSortLocationsEmpty(t *testing.T) {
	locs := map[uuid.UUID]*devicetypes.CaniLocationType{}
	sorted := topologicalSortLocations(locs)
	if len(sorted) != 0 {
		t.Errorf("expected 0, got %d", len(sorted))
	}
}

// TestTopologicalSortLocationsSingleRoot verifies topologicalSortLocations
// returns a single root location unchanged.
//
// Why it matters: a lone top-level location (Parent=uuid.Nil) must export as
// itself with no ordering surprises.
// Inputs: a one-entry map with a root location named "Root". Outputs: a
// one-element slice containing that location.
// Data choice: a single Nil-parent node is the simplest non-empty graph.
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

// TestTopologicalSortLocationsParentBeforeChild verifies
// topologicalSortLocations orders a parent ahead of its child regardless of map
// iteration order.
//
// Why it matters: Nautobot rejects a child location whose parent does not yet
// exist, so the export order must always place parents first.
// Inputs: a two-entry map (child inserted first) where Child.Parent == Parent.ID.
// Outputs: a slice with Parent at index 0 and Child at index 1.
// Data choice: inserting the child first defeats any accidental reliance on map
// ordering, proving the topo-sort actually reorders.
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

// TestMapToPatchRequestWithRackPositionNoParent verifies MapToPatchRequest skips
// the rack block (Rack stays nil) when the device has no parent.
//
// Why it matters: updating an existing Nautobot device must not attach a rack
// when the local device is unparented, even if a RackPosition is present.
// Inputs: a device with Parent=uuid.Nil, RackPosition=15, Face="front", and a
// seeded cache/inventory. Outputs: a patch request with nil Rack.
// Data choice: RackPosition/Face are set to confirm they are ignored without a
// parent, sidestepping the nil-client rack API call.
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

// TestRefIDValidUnion verifies refID extracts the embedded UUID from a populated
// BulkWritableCableRequestStatusId union.
//
// Why it matters: Nautobot foreign keys arrive as oapi-codegen union types, and
// diffing/exporting relies on pulling the raw UUID back out correctly.
// Inputs: a union built via FromBulkWritableCableRequestStatusId0(expected).
// Outputs: the same UUID that was stored.
// Data choice: a freshly generated UUID round-trips the union to prove identity
// is preserved.
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

// TestFieldDiffStruct verifies a FieldDiff literal exposes the Field, LocalVal,
// and RemoteVal values assigned to it.
//
// Why it matters: FieldDiff is the unit carried through every device comparison
// and printed in the sync summary, so its fields must hold what callers set.
// Inputs: a FieldDiff literal for a device_type drift (DL380 vs DL360).
// Outputs: the struct's own fields (no package logic is invoked).
// Data choice: a realistic device-type mismatch keeps the example meaningful.
// NOTE: this only checks Go struct assignment, not any export behavior.
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

// TestMapToNautobotDeviceComments verifies MapToNautobotDevice copies the cani
// Comments field into the Nautobot device request.
//
// Why it matters: operator notes on a device must survive the export so they
// remain visible in Nautobot.
// Inputs: a fully resolvable device (seeded type/location/status/role) with
// Comments="Important device". Outputs: a request whose Comments points to that
// string.
// Data choice: all FK caches are pre-seeded so the test isolates comment mapping
// from FK-resolution failures.
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

// TestMapToNautobotDeviceSerialAndAssetTag verifies MapToNautobotDevice copies
// Serial and AssetTag into the Nautobot device request.
//
// Why it matters: serial and asset tag are key asset-tracking fields that must
// transfer accurately from cani into Nautobot.
// Inputs: a resolvable device with Serial="SN-99887766" and AssetTag="ASSET-XYZ".
// Outputs: a request whose Serial and AssetTag pointers hold those values.
// Data choice: distinctive sentinel strings make an accidental swap or drop
// obvious.
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

// TestResolveStatusAutoCreateActive verifies resolveStatus falls back to the
// "Active" status when auto-create is enabled and the device sets no status.
//
// Why it matters: every Nautobot device requires a status, so cani must supply a
// sensible default rather than fail the export.
// Inputs: a device with no status, createStatuses=true, and a cache pre-seeded
// with "Active". Outputs: the cached "Active" CachedItem.
// Data choice: seeding the cache lets the default resolve without hitting the
// nil HTTP client that a real create would use.
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

// TestResolveStatusNoAutoCreateFails verifies resolveStatus errors when the
// device has no status and auto-create is disabled.
//
// Why it matters: without a status and without permission to create one, the
// export must stop rather than push an invalid device to Nautobot.
// Inputs: a device with no status and the default createStatuses=false cache.
// Outputs: a non-nil error.
// Data choice: leaving both the device status and default empty isolates the
// failure path.
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

// TestResolveRoleAutoCreateGeneric verifies resolveRole falls back to the
// "Generic" role when auto-create is enabled and the device sets no role.
//
// Why it matters: Nautobot devices need a role, so cani defaults unclassified
// devices to "Generic" instead of aborting the export.
// Inputs: a device with no role, createRoles=true, and a cache seeded with
// "Generic". Outputs: the cached "Generic" CachedItem.
// Data choice: the seeded cache resolves the default without a live API call.
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

// TestResolveRoleNoAutoCreateFails verifies resolveRole errors when the device
// has no role and auto-create is disabled.
//
// Why it matters: a device with no resolvable role must fail fast rather than be
// exported into Nautobot without one.
// Inputs: a bare device and a cache with createRoles=false.
// Outputs: a non-nil error.
// Data choice: the empty device and default cache isolate the no-role error
// branch.
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

// TestResolveLocationAutoCreateDefault verifies resolveLocation falls back to the
// "Default" location when auto-create is enabled and no location is set.
//
// Why it matters: every Nautobot device must live somewhere, so cani supplies a
// default location instead of failing the export.
// Inputs: a device with no location, createLocations=true, and a cache seeded
// with "Default". Outputs: the cached "Default" CachedItem.
// Data choice: seeding the cache resolves the default without the nil-client
// create path.
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

// TestResolveLocationNoAutoCreateFails verifies resolveLocation errors when no
// location is set and auto-create is disabled.
//
// Why it matters: exporting a device with no resolvable location would orphan it
// in Nautobot, so the mapper must refuse.
// Inputs: a bare device and a cache with createLocations=false.
// Outputs: a non-nil error.
// Data choice: the empty device and default cache isolate the failure path.
func TestResolveLocationNoAutoCreateFails(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{}
	_, err := mapper.resolveLocation(dev)
	if err == nil {
		t.Fatal("expected error when no location and no auto-create")
	}
}

// TestResolveLocationFromMetadataExplicit verifies resolveLocation honors an
// explicit ProviderMetadata["location"] over any default.
//
// Why it matters: when cani data names a specific location, the export must place
// the device there rather than at a fallback.
// Inputs: a device whose ProviderMetadata sets location="Lab-42", with "Lab-42"
// seeded in the cache. Outputs: the CachedItem whose ID matches the seeded one.
// Data choice: asserting on the seeded UUID proves the named location (not a
// default) was selected.
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

// TestMapToPatchRequestSerialAssetComments verifies MapToPatchRequest carries
// Serial, AssetTag, Comments, and custom fields into the patch body.
//
// Why it matters: an in-place update of an existing Nautobot device must
// propagate edits to these tracking fields, not just structural ones.
// Inputs: a resolvable device with serial/asset/comments and a custom_key in
// ProviderMetadata. Outputs: a patch request mirroring all four values.
// Data choice: distinct "PATCH" sentinel strings make a dropped field obvious.
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

// TestMapToPatchRequestNilDevice verifies MapToPatchRequest returns an error when
// handed a nil device.
//
// Why it matters: a nil device must be rejected defensively so the exporter
// never dereferences it while building an update.
// Inputs: nil device plus a random target UUID. Outputs: a non-nil error.
// Data choice: nil is the degenerate input that guards the mapper's entry check.
func TestMapToPatchRequestNilDevice(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	_, err := mapper.MapToPatchRequest(nil, uuid.New())
	if err == nil {
		t.Fatal("expected error for nil device")
	}
}

// ---------- MapToPatchRequest rack via Racks collection ----------

// TestMapToPatchRequestRackViaRacksCollection verifies MapToPatchRequest
// resolves a parent rack from the inventory's Racks collection through the real
// Nautobot rack lookup path.
//
// Why it matters: merge exports PATCH existing Nautobot devices; if rack
// placement is not resolved into a rack UUID, moved devices lose their slot.
// Inputs: a device whose Parent is an inventory rack named "rack-1" and a fake
// Nautobot server returning that rack UUID. Outputs: a patch request with Rack,
// Position=10, and Face=rear set.
// Data choice: Parent (not Rack) is the field the patch mapper uses, so this
// fixture exercises the update-specific rack branch directly.
func TestMapToPatchRequestRackViaRacksCollection(t *testing.T) {
	nautobotRackID := uuid.New()
	rackID := uuid.New()
	e, cleanup := newExporterWithServer(t, rackLookupServer(nautobotRackID))
	defer cleanup()
	seedDeviceRefs(t, e)

	mapper := newCrudMapper(e)
	inv := devicetypes.NewInventory()
	inv.Racks[rackID] = &devicetypes.CaniRackType{ID: rackID, Name: "rack-1"}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{
		Name:         "rack-dev",
		Parent:       rackID,
		RackPosition: 10,
		Face:         "rear",
		ObjectMeta:   devicetypes.ObjectMeta{Status: "Active", Role: "Compute"},
	}

	req, err := mapper.MapToPatchRequest(dev, uuid.New())
	if err != nil {
		t.Fatalf("MapToPatchRequest() error = %v", err)
	}
	if req.Rack == nil || req.Rack.Id == nil {
		t.Fatal("expected rack reference to be set")
	}
	if got := extractRackID(t, req.Rack.Id); got != nautobotRackID {
		t.Errorf("rack id = %s, want %s", got, nautobotRackID)
	}
	if req.Position == nil || *req.Position != 10 {
		t.Errorf("position = %v, want 10", req.Position)
	}
	if req.Face == nil {
		t.Fatal("expected face to be set")
	}
	face, err := req.Face.AsFaceEnum()
	if err != nil {
		t.Fatalf("decode face: %v", err)
	}
	if face != nautobotapi.FaceEnumRear {
		t.Errorf("face = %v, want %v", face, nautobotapi.FaceEnumRear)
	}
}

// ---------- MapToWritableDeviceRequest with rack in Racks collection ----------

// TestMapToWritableDeviceRequestWithRacksCollection verifies
// MapToWritableDeviceRequest resolves a rack from the inventory's Racks
// collection through the real Nautobot rack lookup path.
//
// Why it matters: new device creates must include rack UUID, position, and face
// so Nautobot stores the device in the same physical slot as the cani inventory.
// Inputs: a classified device whose Rack field points at inventory rack
// "rack-1" and a fake Nautobot server returning that rack UUID. Outputs: a
// writable request with Rack, Position=22, and Face=rear set.
// Data choice: Rack (not Parent) is the field the create mapper resolves via
// GetRackID, so this fixture exercises the create-specific rack branch.
func TestMapToWritableDeviceRequestWithRacksCollection(t *testing.T) {
	nautobotRackID := uuid.New()
	rackID := uuid.New()
	e, cleanup := newExporterWithServer(t, rackLookupServer(nautobotRackID))
	defer cleanup()
	seedDeviceRefs(t, e)

	mapper := newCrudMapper(e)
	inv := devicetypes.NewInventory()
	inv.Racks[rackID] = &devicetypes.CaniRackType{ID: rackID, Name: "rack-1"}
	mapper.SetInventory(inv)

	dev := &devicetypes.CaniDeviceType{
		Name:         "rack-srv",
		Slug:         "hpe-dl380",
		Rack:         rackID,
		RackPosition: 22,
		Face:         "rear",
		ObjectMeta:   devicetypes.ObjectMeta{Status: "Active", Role: "Compute"},
	}

	req, err := mapper.MapToWritableDeviceRequest(dev)
	if err != nil {
		t.Fatalf("MapToWritableDeviceRequest() error = %v", err)
	}
	if req.Rack == nil || req.Rack.Id == nil {
		t.Fatal("expected rack reference to be set")
	}
	if got := extractRackID(t, req.Rack.Id); got != nautobotRackID {
		t.Errorf("rack id = %s, want %s", got, nautobotRackID)
	}
	if req.Position == nil || *req.Position != 22 {
		t.Errorf("position = %v, want 22", req.Position)
	}
	if req.Face == nil {
		t.Fatal("expected face to be set")
	}
	face, err := req.Face.AsFaceEnum()
	if err != nil {
		t.Fatalf("decode face: %v", err)
	}
	if face != nautobotapi.FaceEnumRear {
		t.Errorf("face = %v, want %v", face, nautobotapi.FaceEnumRear)
	}
}

// ---------- comparePosition matching (no diff) ----------

// TestComparePositionMatching verifies comparePosition reports no diff when the
// local and remote rack positions are equal.
//
// Why it matters: identical positions must not be flagged as drift, or every
// sync would needlessly re-write unchanged devices.
// Inputs: a device at RackPosition 12 and a remote Device with Position=12.
// Outputs: an empty diff slice.
// Data choice: equal non-zero positions isolate the matching branch.
func TestComparePositionMatching(t *testing.T) {
	pos := 12
	dev := &devicetypes.CaniDeviceType{RackPosition: 12}
	remote := &nautobotapi.Device{Position: &pos}
	diffs := comparePosition(dev, remote)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when positions match, got %+v", diffs)
	}
}

// TestComparePositionLocalZero verifies comparePosition reports no diff when the
// local position is zero (unset), even if the remote has a position.
//
// Why it matters: a device with no local position must not clobber the position
// already recorded in Nautobot.
// Inputs: a device with RackPosition=0 and a remote Device with Position=5.
// Outputs: an empty diff slice (RackPosition<=0 short-circuits).
// Data choice: local 0 vs remote 5 proves the unset-local guard wins over a
// differing remote.
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

// TestCompareFaceMatching verifies compareFace reports no diff when the local and
// remote rack faces are equal.
//
// Why it matters: a matching face must not register as drift during the device
// comparison that decides whether to update Nautobot.
// Inputs: a device with Face="front" and a remote whose Face value is "front".
// Outputs: an empty diff slice.
// Data choice: equal "front" faces isolate the no-diff branch.
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

// TestCompareFaceLocalEmpty verifies compareFace reports no diff when the local
// face is empty, regardless of the remote value.
//
// Why it matters: an unspecified local face must not overwrite the face already
// stored in Nautobot.
// Inputs: a device with Face="" and a remote whose Face value is "rear".
// Outputs: an empty diff slice (empty local face short-circuits).
// Data choice: empty local vs "rear" remote proves the guard ignores a populated
// remote.
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

// TestCompareRackBothMatchingSameRack verifies compareRack returns nil when
// neither the local device nor the remote has a rack.
//
// Why it matters: two "no rack" sides must agree, avoiding spurious rack drift
// for unracked devices during sync.
// Inputs: a device with Rack=uuid.Nil, a remote with Rack=nil, and a mapper with
// an empty inventory. Outputs: nil diffs.
// Data choice: empty rack on both sides isolates the both-Nil match branch.
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

// TestCompareRackRemoteHasRackLocalNil verifies compareRack emits one "rack" diff
// when the remote has a rack but the local device does not.
//
// Why it matters: a rack present in Nautobot but absent locally is real drift the
// operator must see before merging.
// Inputs: a device with Rack=uuid.Nil and a remote whose Rack ID union holds a
// fresh UUID. Outputs: a single FieldDiff with Field=="rack".
// Data choice: a populated remote union against a Nil local forces exactly the
// one-sided mismatch.
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

// TestGenerateDeviceNamesModelPath verifies generateDeviceNames derives a name
// from the device Model when Name is empty, slugified under a "cani-" prefix.
//
// Why it matters: Nautobot requires a device name, so unnamed cani devices need
// a deterministic, human-readable fallback before export.
// Inputs: a node device with empty Name and Model="HPE DL380 Gen10".
// Outputs: the device's Name set to "cani-hpe-dl380-gen10" in place.
// Data choice: a spaced, mixed-case model proves the lowercase/space-to-dash
// slugging.
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

// TestGenerateDeviceNamesSkipsNonDevice verifies generateDeviceNames leaves the
// Name untouched for inventory entries that are not the device category.
//
// Why it matters: only Nautobot "device" objects need generated names; racks and
// other categories must not be renamed by this pass.
// Inputs: an entry with Type="rack" and empty Name. Outputs: the Name stays empty
// (the entry is skipped).
// Data choice: Type="rack" exercises the ClassifyForNautobot non-device branch.
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

// TestGenerateDeviceNamesUUIDFallback verifies generateDeviceNames falls back to
// the UUID prefix when a device has no serial, slug, or model.
//
// Why it matters: even a device with no descriptive fields must still receive a
// unique, stable name so the export can create it in Nautobot.
// Inputs: a node device with empty Name and no serial/slug/model.
// Outputs: Name set to "cani-" plus the first 8 chars of the device UUID.
// Data choice: omitting every name source forces the last-resort UUID branch.
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

// TestResolveContentLocationSiblingResolution verifies resolveContentLocation
// returns a same-name sibling's device-capable child when the primary location
// lacks a supporting child.
//
// Why it matters: when a location can't host a content type, export tries a
// same-named sibling's subtree before falling back to a default.
// Inputs: two "DC-Main" locations (one childless, one with a "row" child) and a
// target contentType="device". Outputs: "Row-1".
// Data choice: duplicate names with differing hierarchies trigger the sibling
// walk while unique test-only location-type slugs keep the fixture independent
// of embedded YAML contents.
func TestResolveContentLocationSiblingResolution(t *testing.T) {
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Sibling DC RCL",
		Slug:         "sibling-dc-rcl",
		ContentTypes: []string{},
	})
	devicetypes.RegisterLocationType(devicetypes.LocationTypeDefinition{
		Name:         "Sibling Row RCL",
		Slug:         "sibling-row-rcl",
		ContentTypes: []string{"device"},
	})

	locID := uuid.New()
	siblingID := uuid.New()
	childID := uuid.New()

	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			locID: {
				ID:           locID,
				Name:         "DC-Main",
				LocationType: "sibling-dc-rcl",
				Children:     []uuid.UUID{},
			},
			siblingID: {
				ID:           siblingID,
				Name:         "DC-Main",
				LocationType: "sibling-dc-rcl",
				Children:     []uuid.UUID{childID},
			},
			childID: {
				ID:           childID,
				Name:         "Row-1",
				LocationType: "sibling-row-rcl",
			},
		},
	}

	result := resolveContentLocation(locID, "device", inv)
	if result != "Row-1" {
		t.Errorf("resolveContentLocation() = %q, want Row-1", result)
	}
}

// ---------- disambiguateDeviceNames — RackPosition without rack name ----------

// TestDisambiguateDeviceNamesRackPositionNoName verifies disambiguateDeviceNames
// makes duplicate names unique using the rack position when the rack is unnamed.
//
// Why it matters: Nautobot enforces name uniqueness per location, so colliding
// device names must be disambiguated before export.
// Inputs: two "dup-node" devices at positions 3 and 7 sharing an unnamed rack.
// Outputs: the two devices end with different Names (position suffixes).
// Data choice: no serials plus an empty rack name forces the rack-position
// branch rather than the serial or index suffix.
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

// TestErrDeviceUnclassifiedSentinel verifies resolveDeviceType returns the
// ErrDeviceUnclassified sentinel (matchable via errors.Is) for a device with no
// slug or model.
//
// Why it matters: callers branch on this sentinel to skip or warn about
// unclassifiable devices instead of pushing junk device types to Nautobot.
// Inputs: a non-strict mapper and an empty device. Outputs: an error that
// errors.Is matches against ErrDeviceUnclassified.
// Data choice: an empty device (no slug/model) is the only input that yields the
// sentinel.
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

// TestPrintLoadSummaryConflictNoDiffs verifies printLoadSummary handles a skipped
// device whose conflict carries no field diffs (the "up to date" branch).
//
// Why it matters: the end-of-sync summary must distinguish a conflict with real
// pending changes from one that is already current.
// Inputs: a LoadResult with Skipped=["dev-5"] and a ConflictInfo with nil Diffs.
// Outputs: colored summary log lines only.
// Data choice: a conflict with nil Diffs targets the empty-diffs "up to date"
// path. NOTE: a smoke test — it asserts nothing beyond not panicking.
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
