package devicetypes

// Test coverage for inventory.go
//
// NewInventory:
//   - TestNewInventoryReturnsInitializedMaps      (core maps non-nil)
//   - TestNewInventoryMapsAreEmpty                (maps empty + independent)
//   - TestNewInventorySchemaVersionIsV1Alpha3     (default schema version)
//   - TestNewInventoryInitializesIPAMMaps         (IPAM maps non-nil + empty)
//   - TestNewInventoryInitializesMetadataAndIndex (metadata + pkIndex non-nil)
//
// EnsureUniqueDeviceNames:
//   - TestEnsureUniqueDeviceNamesSuffixes              (duplicates -> unique)
//   - TestEnsureUniqueDeviceNamesNoDuplicates          (unique names untouched)
//   - TestEnsureUniqueDeviceNamesEmptyAndNilMap        (early-return guard)
//   - TestEnsureUniqueDeviceNamesSuffixFormat          ("<name>-<n>" format)
//   - TestEnsureUniqueDeviceNamesSkipsNilAndEmptyNames (nil/empty skipped)
//   - TestEnsureUniqueDeviceNamesPartialDuplicates     (mixed unique + dup)

import (
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"
)

// ---------- NewInventory ----------

// TestNewInventoryReturnsInitializedMaps verifies the constructor returns a
// non-nil inventory with every core entity map initialized.
//
// Why it matters: inventory code inserts items into these maps by UUID key, so a
// nil map from the constructor would panic on the first assignment; the
// constructor's contract is to hand back ready-to-use maps.
// Inputs: none. Outputs: the inventory pointer and its seven core maps
// (Locations, Racks, Devices, Modules, Cables, Frus, Interfaces).
// Data choice: each core map is checked individually so a regression that leaves
// one map nil identifies exactly which map broke.
func TestNewInventoryReturnsInitializedMaps(t *testing.T) {
	inv := NewInventory()
	if inv == nil {
		t.Fatal("expected non-nil inventory")
	}
	if inv.Locations == nil {
		t.Error("expected Locations map to be initialized")
	}
	if inv.Racks == nil {
		t.Error("expected Racks map to be initialized")
	}
	if inv.Devices == nil {
		t.Error("expected Devices map to be initialized")
	}
	if inv.Modules == nil {
		t.Error("expected Modules map to be initialized")
	}
	if inv.Cables == nil {
		t.Error("expected Cables map to be initialized")
	}
	if inv.Frus == nil {
		t.Error("expected Frus map to be initialized")
	}
	if inv.Interfaces == nil {
		t.Error("expected Interfaces map to be initialized")
	}
}

// TestNewInventoryMapsAreEmpty verifies a freshly constructed inventory has zero
// entries in every core map and shares no state with another inventory.
//
// Why it matters: each NewInventory call must yield an isolated inventory; shared
// backing maps would let one inventory's writes leak into another and corrupt
// unrelated data.
// Inputs: two independent inventories, with a single Location inserted into the
// second. Outputs: the first inventory's map lengths, all expected to be zero.
// Data choice: mutating only the second inventory and asserting the first stays
// empty proves independence, not merely initial emptiness.
func TestNewInventoryMapsAreEmpty(t *testing.T) {
	inv := NewInventory()

	// A freshly created inventory must have zero entries in every map.
	// Inserting an item and then checking a *different* inventory proves
	// that inventories are independent (no shared state).
	other := NewInventory()
	other.Locations[uuid.New()] = &CaniLocationType{}

	if len(inv.Locations) != 0 {
		t.Errorf("expected 0 locations, got %d", len(inv.Locations))
	}
	if len(inv.Racks) != 0 {
		t.Errorf("expected 0 racks, got %d", len(inv.Racks))
	}
	if len(inv.Devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(inv.Devices))
	}
	if len(inv.Modules) != 0 {
		t.Errorf("expected 0 modules, got %d", len(inv.Modules))
	}
	if len(inv.Cables) != 0 {
		t.Errorf("expected 0 cables, got %d", len(inv.Cables))
	}
	if len(inv.Frus) != 0 {
		t.Errorf("expected 0 frus, got %d", len(inv.Frus))
	}
	if len(inv.Interfaces) != 0 {
		t.Errorf("expected 0 interfaces, got %d", len(inv.Interfaces))
	}
}

// TestNewInventorySchemaVersionIsV1Alpha3 verifies the constructor stamps the
// current datastore schema version on a fresh inventory.
//
// Why it matters: the schema version drives load/migration decisions, so a new
// inventory must declare the latest version (v1alpha3) rather than a zero value
// that would later be read as an unknown or older schema.
// Inputs: none. Outputs: the SchemaVersion field of the new inventory.
// Data choice: comparing against the SchemaVersionV1Alpha3 constant (not a
// string literal) keeps the test correct if the default is intentionally bumped.
func TestNewInventorySchemaVersionIsV1Alpha3(t *testing.T) {
	inv := NewInventory()
	if inv.SchemaVersion != SchemaVersionV1Alpha3 {
		t.Errorf("SchemaVersion = %q, want %q", inv.SchemaVersion, SchemaVersionV1Alpha3)
	}
}

// TestNewInventoryInitializesIPAMMaps verifies the IPAM maps (Prefixes,
// IPAddresses, VLANs) are initialized and empty on a new inventory.
//
// Why it matters: IPAM code writes directly into these maps; if the constructor
// left them nil the first assignment would panic, so the constructor's contract
// is that every IPAM map is ready to use.
// Inputs: none. Outputs: the three IPAM maps and their combined length.
// Data choice: these maps are checked explicitly because the existing
// initialized-maps test omits them, leaving the IPAM half of the constructor
// unasserted.
func TestNewInventoryInitializesIPAMMaps(t *testing.T) {
	inv := NewInventory()

	if inv.Prefixes == nil {
		t.Error("expected Prefixes map to be initialized")
	}
	if inv.IPAddresses == nil {
		t.Error("expected IPAddresses map to be initialized")
	}
	if inv.VLANs == nil {
		t.Error("expected VLANs map to be initialized")
	}
	if n := len(inv.Prefixes) + len(inv.IPAddresses) + len(inv.VLANs); n != 0 {
		t.Errorf("expected IPAM maps to be empty, got %d total entries", n)
	}
}

// TestNewInventoryInitializesMetadataAndIndex verifies the constructor allocates
// the metadata catalog and the transient provider-key index.
//
// Why it matters: item roles/statuses/tags are resolved through Metadata and
// device dedup uses the provider-key index, so both must be non-nil before any
// item is added, otherwise the first lookup would dereference a nil value.
// Inputs: none. Outputs: the Metadata pointer and the unexported pkIndex map.
// Data choice: a white-box test in the same package lets the unexported pkIndex
// -- part of the constructor's contract -- be asserted directly.
func TestNewInventoryInitializesMetadataAndIndex(t *testing.T) {
	inv := NewInventory()

	if inv.Metadata == nil {
		t.Error("expected Metadata to be initialized, got nil")
	}
	if inv.pkIndex == nil {
		t.Error("expected pkIndex to be initialized, got nil")
	}
}

// ---------- EnsureUniqueDeviceNames ----------

// TestEnsureUniqueDeviceNamesSuffixes verifies three identically named devices
// are rewritten so that every resulting name is unique.
//
// Why it matters: device names must be unique downstream (and in Nautobot), so
// the de-duplication pass has to resolve a multi-way collision, not just a pair.
// Inputs: three devices all named "server". Outputs: the set of names after the
// call, expected to be three distinct values.
// Data choice: three identical names (rather than two) prove the suffixing scales
// past a single pair and leaves no residual duplicate.
func TestEnsureUniqueDeviceNamesSuffixes(t *testing.T) {
	tr := &TransformResult{
		Devices: map[uuid.UUID]*CaniDeviceType{
			uuid.New(): {Name: "server"},
			uuid.New(): {Name: "server"},
			uuid.New(): {Name: "server"},
		},
	}

	tr.EnsureUniqueDeviceNames()

	seen := make(map[string]bool)
	for _, d := range tr.Devices {
		if seen[d.Name] {
			t.Errorf("duplicate name after EnsureUniqueDeviceNames: %q", d.Name)
		}
		seen[d.Name] = true
	}
	if len(seen) != 3 {
		t.Errorf("expected 3 unique names, got %d", len(seen))
	}
}

// TestEnsureUniqueDeviceNamesNoDuplicates verifies already-unique device names
// are left untouched by the de-duplication pass.
//
// Why it matters: device names are stable identifiers other systems reference, so
// the pass must not append suffixes when there is no collision to resolve.
// Inputs: three distinct names (alpha, bravo, charlie). Outputs: the name set
// after the call, expected to be identical to the input.
// Data choice: three fully distinct names guarantee the collision path is never
// taken, isolating the no-op branch.
func TestEnsureUniqueDeviceNamesNoDuplicates(t *testing.T) {
	tr := &TransformResult{
		Devices: map[uuid.UUID]*CaniDeviceType{
			uuid.New(): {Name: "alpha"},
			uuid.New(): {Name: "bravo"},
			uuid.New(): {Name: "charlie"},
		},
	}

	tr.EnsureUniqueDeviceNames()

	// Names should be unchanged when all unique.
	names := make(map[string]bool)
	for _, d := range tr.Devices {
		names[d.Name] = true
	}
	for _, want := range []string{"alpha", "bravo", "charlie"} {
		if !names[want] {
			t.Errorf("expected name %q to remain unchanged", want)
		}
	}
}

// TestEnsureUniqueDeviceNamesEmptyAndNilMap verifies the de-duplication step is
// a no-op when the transform produced no devices, for both an empty and a nil
// map.
//
// Why it matters: providers that emit only non-device types (e.g. cables or
// VLANs) hand an empty or nil Devices map to EnsureUniqueDeviceNames before the
// merge, so the early-return guard must neither panic nor allocate names.
// Inputs: a TransformResult with an empty Devices map, and one with a nil map.
// Outputs: both maps are left untouched (empty stays empty, nil stays nil).
// Data choice: the two zero-device shapes are the only inputs that exercise the
// `len(tr.Devices) == 0` guard, and the nil map additionally proves no implicit
// allocation occurs.
func TestEnsureUniqueDeviceNamesEmptyAndNilMap(t *testing.T) {
	empty := &TransformResult{Devices: map[uuid.UUID]*CaniDeviceType{}}
	empty.EnsureUniqueDeviceNames()
	if len(empty.Devices) != 0 {
		t.Errorf("expected empty Devices to stay empty, got %d", len(empty.Devices))
	}

	nilMap := &TransformResult{}
	nilMap.EnsureUniqueDeviceNames()
	if nilMap.Devices != nil {
		t.Errorf("expected nil Devices to stay nil, got %v", nilMap.Devices)
	}
}

// TestEnsureUniqueDeviceNamesSuffixFormat verifies duplicate names are rewritten
// to the exact "<name>-<n>" form with a 1-based, incrementing sequence.
//
// Why it matters: downstream code and Nautobot require unique device names, and
// the suffix format is the contract other layers display and match against, so
// the literal output -- not just uniqueness -- must be pinned.
// Inputs: two devices both named "server". Outputs: the resulting name set,
// collected and sorted for a deterministic comparison.
// Data choice: exactly two identical names is the minimal collision that yields
// the predictable, fully-enumerable result {"server-1","server-2"}, letting the
// test assert the format without depending on Go's random map order.
func TestEnsureUniqueDeviceNamesSuffixFormat(t *testing.T) {
	tr := &TransformResult{
		Devices: map[uuid.UUID]*CaniDeviceType{
			uuid.New(): {Name: "server"},
			uuid.New(): {Name: "server"},
		},
	}

	tr.EnsureUniqueDeviceNames()

	got := make([]string, 0, len(tr.Devices))
	for _, d := range tr.Devices {
		got = append(got, d.Name)
	}
	sort.Strings(got)

	want := []string{"server-1", "server-2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("EnsureUniqueDeviceNames() names = %v, want %v", got, want)
	}
}

// TestEnsureUniqueDeviceNamesSkipsNilAndEmptyNames verifies nil device pointers
// and devices with empty names are ignored by the de-duplication pass.
//
// Why it matters: a transform may leave placeholder/nil entries or not-yet-named
// devices in the map, and treating empty names as a "duplicate" group would
// wrongly rewrite them to "-1"/"-2", corrupting later naming.
// Inputs: a map holding a nil pointer, two empty-named devices, and one named
// "server". Outputs: the nil stays nil, both empty names stay empty, and the
// lone real name is left unsuffixed because it is unique.
// Data choice: two empty-named entries specifically prove the empty string is
// not counted as a collision -- two of them would trigger suffixing if it were.
func TestEnsureUniqueDeviceNamesSkipsNilAndEmptyNames(t *testing.T) {
	nilID := uuid.New()
	emptyA := uuid.New()
	emptyB := uuid.New()
	named := uuid.New()

	tr := &TransformResult{
		Devices: map[uuid.UUID]*CaniDeviceType{
			nilID:  nil,
			emptyA: {Name: ""},
			emptyB: {Name: ""},
			named:  {Name: "server"},
		},
	}

	tr.EnsureUniqueDeviceNames()

	if tr.Devices[nilID] != nil {
		t.Errorf("expected nil device to remain nil, got %v", tr.Devices[nilID])
	}
	if tr.Devices[emptyA].Name != "" || tr.Devices[emptyB].Name != "" {
		t.Errorf("expected empty names to remain empty, got %q and %q",
			tr.Devices[emptyA].Name, tr.Devices[emptyB].Name)
	}
	if tr.Devices[named].Name != "server" {
		t.Errorf("expected unique name to remain 'server', got %q", tr.Devices[named].Name)
	}
}

// TestEnsureUniqueDeviceNamesPartialDuplicates verifies only colliding names are
// suffixed while already-unique names are left untouched in the same pass.
//
// Why it matters: real transforms mix unique and duplicate names, and over-eager
// rewriting of a unique name would change a stable identifier other systems
// already reference.
// Inputs: two devices named "node" plus one named "head". Outputs: "head" is
// unchanged, and the two "node" entries become {"node-1","node-2"}.
// Data choice: combining one unique name with one duplicated pair exercises both
// branches of the per-name count check (<=1 skip vs. >1 suffix) in a single run.
func TestEnsureUniqueDeviceNamesPartialDuplicates(t *testing.T) {
	dupA := uuid.New()
	dupB := uuid.New()
	unique := uuid.New()

	tr := &TransformResult{
		Devices: map[uuid.UUID]*CaniDeviceType{
			dupA:   {Name: "node"},
			dupB:   {Name: "node"},
			unique: {Name: "head"},
		},
	}

	tr.EnsureUniqueDeviceNames()

	if tr.Devices[unique].Name != "head" {
		t.Errorf("expected unique name 'head' to be unchanged, got %q", tr.Devices[unique].Name)
	}

	got := []string{tr.Devices[dupA].Name, tr.Devices[dupB].Name}
	sort.Strings(got)
	want := []string{"node-1", "node-2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("duplicate names = %v, want %v", got, want)
	}
}
