package datastores

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// ---------- isLegacyDatastore ----------

// TestIsLegacyDatastoreV1Alpha1 verifies a v1alpha1 datastore fixture is detected as legacy.
//
// Why it matters: JSONStore.Load only backs up and migrates old CRUD inventory
// data when this detector recognizes the legacy file shape.
// Inputs: the canitestdb_v1alpha1.json fixture. Outputs: true from
// isLegacyDatastore.
// Data choice: the fixture is a representative legacy datastore with Hardware
// records and an explicit v1alpha1 schema version.
func TestIsLegacyDatastoreV1Alpha1(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/fixtures/cani/legacy/canitestdb_v1alpha1.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	if !isLegacyDatastore(raw) {
		t.Error("expected v1alpha1 fixture to be detected as legacy")
	}
}

// TestIsLegacyDatastoreV1Alpha2 verifies a current inventory is not treated as legacy.
//
// Why it matters: current datastore files must load normally so CRUD data is not
// unnecessarily backed up or rewritten through the legacy migration path.
// Inputs: a freshly marshaled v1alpha2-or-newer Inventory. Outputs: false from
// isLegacyDatastore.
// Data choice: devicetypes.NewInventory provides the canonical current datastore
// shape used by JSONStore.Save.
func TestIsLegacyDatastoreV1Alpha2(t *testing.T) {
	inv := devicetypes.NewInventory()
	data, _ := json.Marshal(inv)
	if isLegacyDatastore(data) {
		t.Error("v1alpha2 inventory should NOT be detected as legacy")
	}
}

// TestIsLegacyDatastoreInvalidJSON verifies malformed JSON is not classified as legacy.
//
// Why it matters: invalid datastore contents should surface as parse errors in
// Load instead of being routed into migration.
// Inputs: invalid JSON bytes. Outputs: false from isLegacyDatastore.
// Data choice: the malformed object is the smallest input that exercises the
// detector's unmarshal guard.
func TestIsLegacyDatastoreInvalidJSON(t *testing.T) {
	if isLegacyDatastore([]byte("{bad json}")) {
		t.Error("invalid JSON should not be detected as legacy")
	}
}

// TestIsLegacyDatastoreEmptyHardware verifies schema alone is not enough for legacy detection.
//
// Why it matters: migration requires Hardware records; classifying a file without
// them as migratable would risk producing incomplete inventory state.
// Inputs: JSON with SchemaVersion v1alpha1 and no Hardware key. Outputs: false
// from isLegacyDatastore.
// Data choice: the fixture isolates the missing Hardware condition from all
// other legacy fields.
func TestIsLegacyDatastoreEmptyHardware(t *testing.T) {
	raw := []byte(`{"SchemaVersion":"v1alpha1"}`)
	if isLegacyDatastore(raw) {
		t.Error("v1alpha1 with no Hardware key should not be detected as legacy")
	}
}

// ---------- migrateV1Alpha1 ----------

func loadFixture(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("../../testdata/fixtures/cani/legacy/canitestdb_v1alpha1.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	return raw
}

func assertMetadataInt(t *testing.T, csm map[string]any, key string, want int) {
	t.Helper()

	got, ok := metadataInt(csm[key])
	if !ok {
		t.Fatalf("%s unexpected type %T", key, csm[key])
	}
	if got != want {
		t.Fatalf("%s = %d, want %d", key, got, want)
	}
}

func metadataInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func assertMetadataStringSlice(t *testing.T, csm map[string]any, key string, want []string) {
	t.Helper()

	got, ok := csm[key].([]string)
	if !ok {
		t.Fatalf("%s unexpected type %T", key, csm[key])
	}
	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d", key, len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s = %v, want %v", key, got, want)
		}
	}
}

// TestMigrateSchemaVersion verifies v1alpha1 migration writes the current schema version.
//
// Why it matters: migrated CRUD inventory must be saved in the new datastore
// schema so future loads skip the legacy migration path.
// Inputs: the legacy canitest fixture. Outputs: an Inventory whose SchemaVersion
// is v1alpha2.
// Data choice: the fixture contains enough legacy hardware records to exercise a
// real migration rather than a schema-only probe.
func TestMigrateSchemaVersion(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}
	if inv.SchemaVersion != devicetypes.SchemaVersionV1Alpha2 {
		t.Errorf("SchemaVersion = %q, want %q", inv.SchemaVersion, devicetypes.SchemaVersionV1Alpha2)
	}
}

// TestMigrateProvider verifies the legacy provider field is preserved.
//
// Why it matters: provider-specific CRUD/import data relies on the inventory's
// provider identity after migration.
// Inputs: the legacy canitest fixture. Outputs: an Inventory with Provider csm.
// Data choice: the fixture's Provider value is csm, matching the metadata stored
// on the migrated hardware records.
func TestMigrateProvider(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}
	if inv.Provider != "csm" {
		t.Errorf("Provider = %q, want %q", inv.Provider, "csm")
	}
}

// TestMigrateSystemBecomesLocation verifies the legacy System record becomes a location.
//
// Why it matters: migrated devices and racks need a valid location hierarchy so
// later CRUD operations can preserve physical placement.
// Inputs: the legacy canitest fixture. Outputs: a system location with active
// status at the legacy System UUID.
// Data choice: the fixed System UUID identifies the root hardware record in the
// fixture and makes the location mapping explicit.
func TestMigrateSystemBecomesLocation(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}
	systemID := uuid.MustParse("00000001-0000-0000-0000-000000000001")
	loc, ok := inv.Locations[systemID]
	if !ok {
		t.Fatal("System was not migrated to a Location")
	}
	if loc.LocationType != "system" {
		t.Errorf("Location type = %q, want %q", loc.LocationType, "system")
	}
	if loc.Status != "Active" {
		t.Errorf("Location status = %q, want %q", loc.Status, "Active")
	}
}

// TestMigrateCabinetCreatesDeviceAndRack verifies a cabinet produces both device and rack records.
//
// Why it matters: legacy cabinet data has to remain addressable as hardware while
// also creating the rack parent used by child device placement.
// Inputs: the legacy canitest fixture. Outputs: a cabinet device whose parent is
// a migrated rack with matching cabinet content.
// Data choice: the fixture cabinet has a stable UUID, name, vendor, and model so
// identity and descriptive fields can be asserted directly.
func TestMigrateCabinetCreatesDeviceAndRack(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}
	cabinetID := uuid.MustParse("00000001-0000-0000-0000-000000000002")
	dev, ok := inv.Devices[cabinetID]
	if !ok {
		t.Fatal("Cabinet was not migrated to a Device")
	}
	if dev.Type != devicetypes.TypeCabinet {
		t.Errorf("Device Type = %q, want %q", dev.Type, devicetypes.TypeCabinet)
	}

	// A rack must have been created for the cabinet.
	if len(inv.Racks) == 0 {
		t.Fatal("No racks created for cabinet")
	}
	// The cabinet device should point at the rack.
	rack, ok := inv.Racks[dev.Parent]
	if !ok {
		t.Errorf("Cabinet device parent %v does not exist in Racks", dev.Parent)
	}
	if rack != nil && rack.Name != dev.Name {
		t.Errorf("Rack Name = %q, want cabinet device name %q", rack.Name, dev.Name)
	}
	if rack != nil && rack.UHeight != 42 {
		t.Errorf("Rack UHeight = %d, want 42", rack.UHeight)
	}
}

// TestMigrateCabinetHMNVlan verifies cabinet CSM metadata keeps the HMN VLAN.
//
// Why it matters: data integrity for migrated cabinets includes provider-specific
// metadata used by later import/export workflows.
// Inputs: the legacy canitest fixture. Outputs: cabinet ProviderMetadata.csm with
// hmnVlan equal to 1513.
// Data choice: VLAN 1513 is the fixture value and is distinctive enough to catch
// missing or mis-typed metadata.
func TestMigrateCabinetHMNVlan(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}
	cabinetID := uuid.MustParse("00000001-0000-0000-0000-000000000002")
	dev := inv.Devices[cabinetID]
	csm, ok := dev.GetProviderSubMap("csm")
	if !ok {
		t.Fatal("Cabinet missing csm provider metadata")
	}
	vlan, ok := csm["hmnVlan"]
	if !ok {
		t.Fatal("Cabinet csm metadata missing hmnVlan")
	}
	assertMetadataInt(t, map[string]any{"hmnVlan": vlan}, "hmnVlan", 1513)
}

// TestMigrateNodeBladeMapped verifies legacy NodeBlade records map to blade devices.
//
// Why it matters: preserving device type during migration keeps downstream CRUD,
// classification, and export behavior from treating blades as generic devices.
// Inputs: the legacy canitest fixture. Outputs: a migrated device with TypeBlade
// at the legacy NodeBlade UUID.
// Data choice: the fixture's NodeBlade UUID isolates the type-mapping case from
// cabinet and node migration cases.
func TestMigrateNodeBladeMapped(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}
	bladeID := uuid.MustParse("00000001-0000-0000-0000-000000000004")
	dev, ok := inv.Devices[bladeID]
	if !ok {
		t.Fatal("NodeBlade was not migrated")
	}
	if dev.Type != devicetypes.TypeBlade {
		t.Errorf("NodeBlade Type = %q, want %q", dev.Type, devicetypes.TypeBlade)
	}
}

// TestMigrateNodeMetadata verifies node role and subrole metadata are preserved.
//
// Why it matters: migrated node provider metadata informs later selection,
// classification, and export decisions after datastore load.
// Inputs: the legacy canitest fixture. Outputs: the node's csm metadata contains
// role Compute and subRole Worker.
// Data choice: the fixture node carries both fields, proving multiple metadata
// keys survive migration.
func TestMigrateNodeMetadata(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}
	nodeID := uuid.MustParse("00000001-0000-0000-0000-000000000005")
	dev, ok := inv.Devices[nodeID]
	if !ok {
		t.Fatal("Node was not migrated")
	}
	csm, ok := dev.GetProviderSubMap("csm")
	if !ok {
		t.Fatal("Node missing csm provider metadata")
	}
	if role, _ := csm["role"].(string); role != "Compute" {
		t.Errorf("role = %q, want %q", role, "Compute")
	}
	if subRole, _ := csm["subRole"].(string); subRole != "Worker" {
		t.Errorf("subRole = %q, want %q", subRole, "Worker")
	}
}

// TestMigrateNodeNidAndAliases verifies node NID and alias metadata are preserved.
//
// Why it matters: NID and aliases are provider identity data that must survive
// migration for idempotent imports and user-facing lookups.
// Inputs: the legacy canitest fixture. Outputs: csm metadata with nid 1001 and
// alias nid001001.
// Data choice: the fixture includes numeric and list-shaped metadata so the test
// covers both decoded value forms.
func TestMigrateNodeNidAndAliases(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}
	nodeID := uuid.MustParse("00000001-0000-0000-0000-000000000005")
	csm, _ := inv.Devices[nodeID].GetProviderSubMap("csm")

	assertMetadataInt(t, csm, "nid", 1001)
	assertMetadataStringSlice(t, csm, "aliases", []string{"nid001001"})
}

// TestMigrateLocationOrdinalInMetadata verifies legacy location ordinals move into metadata.
//
// Why it matters: location ordinal is part of CSM placement identity and must not
// be dropped when legacy hardware becomes new inventory records.
// Inputs: the legacy canitest fixture. Outputs: cabinet csm metadata with
// locationOrdinal 3000.
// Data choice: cabinet ordinal 3000 is the fixture value and identifies the
// cabinet placement branch.
func TestMigrateLocationOrdinalInMetadata(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}
	cabinetID := uuid.MustParse("00000001-0000-0000-0000-000000000002")
	csm, _ := inv.Devices[cabinetID].GetProviderSubMap("csm")
	assertMetadataInt(t, csm, "locationOrdinal", 3000)
}

// TestMigrateDeviceCount verifies migration creates the expected inventory object counts.
//
// Why it matters: object counts catch broad data loss across the legacy System,
// Cabinet, Chassis, NodeBlade, and Node records.
// Inputs: the five-record legacy canitest fixture. Outputs: four devices, one
// location, and one rack.
// Data choice: the fixture mixes records that migrate to different inventory maps
// so the counts prove the high-level split.
func TestMigrateDeviceCount(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}
	// 5 legacy entries: System + Cabinet + Chassis + NodeBlade + Node.
	// System → Location (not a device). Cabinet + Chassis + NodeBlade + Node → 4 devices.
	if got := len(inv.Devices); got != 4 {
		t.Errorf("device count = %d, want 4", got)
	}
	if got := len(inv.Locations); got != 1 {
		t.Errorf("location count = %d, want 1", got)
	}
	if got := len(inv.Racks); got != 1 {
		t.Errorf("rack count = %d, want 1", got)
	}
}

// ---------- backupDatastore ----------

// TestBackupDatastore verifies backupDatastore copies a datastore to .canisave.
//
// Why it matters: legacy migration must preserve the original datastore before
// rewriting CRUD inventory data on disk.
// Inputs: a temporary canidb.json file containing known JSON bytes. Outputs: a
// .canisave file with identical bytes.
// Data choice: the small JSON payload makes byte-for-byte backup integrity easy
// to assert.
func TestBackupDatastore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "canidb.json")
	content := []byte(`{"test": true}`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	if err := backupDatastore(path); err != nil {
		t.Fatalf("backupDatastore: %v", err)
	}

	bak, err := os.ReadFile(path + ".canisave")
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}
	if string(bak) != string(content) {
		t.Errorf("backup content mismatch: got %q, want %q", bak, content)
	}
}

// ---------- mapLegacyType ----------

// TestMapLegacyType verifies known legacy hardware types map to current types.
//
// Why it matters: migrated devices need the right current Type values for CRUD,
// classification, and export behavior after datastore migration.
// Inputs: representative legacy hardware type strings. Outputs: the expected
// devicetypes.Type values.
// Data choice: the table covers cabinet, chassis, node, switch, PDU, CDU, and
// module mappings that have explicit switch cases.
func TestMapLegacyType(t *testing.T) {
	tests := []struct {
		input string
		want  devicetypes.Type
	}{
		{"Cabinet", devicetypes.TypeCabinet},
		{"Chassis", devicetypes.TypeChassis},
		{"NodeBlade", devicetypes.TypeBlade},
		{"NodeCard", devicetypes.TypeNodeCard},
		{"Node", devicetypes.TypeNode},
		{"ManagementSwitch", devicetypes.TypeMgmtSwitch},
		{"HighSpeedSwitch", devicetypes.TypeHsnSwitch},
		{"CabinetPDU", devicetypes.TypeCabinetPDU},
		{"CoolingDistributionUnit", devicetypes.TypeCDU},
		{"ChassisManagementModule", devicetypes.TypeModule},
		{"CabinetEnvironmentalController", devicetypes.TypeModule},
		{"NodeController", devicetypes.TypeModule},
	}
	for _, tc := range tests {
		got := mapLegacyType(tc.input)
		if got != tc.want {
			t.Errorf("mapLegacyType(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ---------- round-trip ----------

// TestMigrateRoundTrip verifies migrated inventory content survives JSON encoding and decoding.
//
// Why it matters: after migration, JSONStore.Save writes this structure back to
// disk; identity, relationships, and metadata must remain intact on reload.
// Inputs: the migrated legacy canitest fixture encoded to JSON and decoded into a
// new Inventory. Outputs: matching schema, provider, counts, IDs, and metadata.
// Data choice: the fixture contains locations, racks, devices, and CSM metadata,
// giving the round trip meaningful data-integrity coverage.
func TestMigrateRoundTrip(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}

	// Marshal and unmarshal to simulate save/load cycle.
	data, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	loaded := devicetypes.NewInventory()
	if err := json.Unmarshal(data, loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if loaded.SchemaVersion != devicetypes.SchemaVersionV1Alpha2 {
		t.Errorf("round-trip SchemaVersion = %q, want %q", loaded.SchemaVersion, devicetypes.SchemaVersionV1Alpha2)
	}
	if loaded.Provider != "csm" {
		t.Errorf("round-trip Provider = %q, want %q", loaded.Provider, "csm")
	}
	if len(loaded.Devices) != len(inv.Devices) {
		t.Errorf("round-trip device count = %d, want %d", len(loaded.Devices), len(inv.Devices))
	}
	if len(loaded.Racks) != len(inv.Racks) {
		t.Errorf("round-trip rack count = %d, want %d", len(loaded.Racks), len(inv.Racks))
	}
	if len(loaded.Locations) != len(inv.Locations) {
		t.Errorf("round-trip location count = %d, want %d", len(loaded.Locations), len(inv.Locations))
	}
	nodeID := uuid.MustParse("00000001-0000-0000-0000-000000000005")
	node, ok := loaded.Devices[nodeID]
	if !ok {
		t.Fatalf("round-trip node %s missing", nodeID)
	}
	wantNode := inv.Devices[nodeID]
	if node.Name != wantNode.Name || node.Type != wantNode.Type {
		t.Errorf("round-trip node = %q/%q, want %q/%q", node.Name, node.Type, wantNode.Name, wantNode.Type)
	}
	csm, ok := node.GetProviderSubMap("csm")
	if !ok {
		t.Fatal("round-trip node missing csm provider metadata")
	}
	if role, _ := csm["role"].(string); role != "Compute" {
		t.Errorf("round-trip node role = %q, want Compute", role)
	}
}

// ---------- Load integration ----------

// TestLoadMigratesLegacyDatastore verifies Load backs up and rewrites legacy datastores.
//
// Why it matters: users opening old datastore files should keep their original
// data and receive a current inventory file with migrated CRUD records.
// Inputs: a temporary copy of the legacy canitest datastore. Outputs: a migrated
// inventory, .canisave backup, and non-legacy on-disk JSON.
// Data choice: copying the fixture into t.TempDir isolates the rewrite while
// exercising the real JSONStore.Load migration path.
func TestLoadMigratesLegacyDatastore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "canidb.json")

	raw, err := os.ReadFile("../../testdata/fixtures/cani/legacy/canitestdb_v1alpha1.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	if err := os.WriteFile(path, raw, 0644); err != nil {
		t.Fatalf("writing test datastore: %v", err)
	}

	store := &JSONStore{Path: path}
	inv, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if inv.SchemaVersion != devicetypes.SchemaVersionV1Alpha2 {
		t.Errorf("SchemaVersion = %q, want %q", inv.SchemaVersion, devicetypes.SchemaVersionV1Alpha2)
	}

	// Backup must exist.
	if _, err := os.Stat(path + ".canisave"); os.IsNotExist(err) {
		t.Error("expected .canisave backup file to exist")
	}

	// The on-disk file should now be v1alpha2.
	reread, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("re-reading migrated file: %v", err)
	}
	if isLegacyDatastore(reread) {
		t.Error("on-disk file should no longer be detected as legacy")
	}

	var saved devicetypes.Inventory
	if err := json.Unmarshal(reread, &saved); err != nil {
		t.Fatalf("unmarshaling migrated file: %v", err)
	}
	nodeID := uuid.MustParse("00000001-0000-0000-0000-000000000005")
	if saved.Devices[nodeID] == nil {
		t.Fatalf("migrated on-disk file missing node %s", nodeID)
	}
	if saved.Devices[nodeID].Name != inv.Devices[nodeID].Name {
		t.Fatalf("migrated on-disk node name = %q, want %q", saved.Devices[nodeID].Name, inv.Devices[nodeID].Name)
	}
}

// ---------- migrateInventoryMetadata ----------

// TestMigrateInventoryMetadataFromOldFormat verifies old providerMetadata catalogs move into Metadata.
//
// Why it matters: inventory-level roles, statuses, and tags must remain
// available after loading datastores written by intermediate builds.
// Inputs: v1alpha2 JSON with providerMetadata.nautobot roles, statuses, and tags.
// Outputs: true from migration and populated Inventory.Metadata slices.
// Data choice: one entry of each metadata kind proves all three catalog branches
// are merged.
func TestMigrateInventoryMetadataFromOldFormat(t *testing.T) {
	raw := []byte(`{
		"schemaVersion": "v1alpha2",
		"providerMetadata": {
			"nautobot": {
				"roles": [{"name": "ComputeNode", "contentTypes": ["dcim.device"]}],
				"statuses": [{"name": "Active", "color": "green"}],
				"tags": [{"name": "gpu-node"}]
			}
		},
		"devices": {}
	}`)

	inv := devicetypes.NewInventory()
	if err := json.Unmarshal(raw, inv); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !migrateInventoryMetadata(raw, inv) {
		t.Fatal("expected migration to occur")
	}

	if len(inv.Metadata.Roles) != 1 || inv.Metadata.Roles[0].Name != "ComputeNode" {
		t.Errorf("roles = %+v, want [ComputeNode]", inv.Metadata.Roles)
	}
	if len(inv.Metadata.Statuses) != 1 || inv.Metadata.Statuses[0].Name != "Active" {
		t.Errorf("statuses = %+v, want [Active]", inv.Metadata.Statuses)
	}
	if len(inv.Metadata.Tags) != 1 || inv.Metadata.Tags[0].Name != "gpu-node" {
		t.Errorf("tags = %+v, want [gpu-node]", inv.Metadata.Tags)
	}
}

// TestMigrateInventoryMetadataNoOp verifies absent old metadata leaves inventory unchanged.
//
// Why it matters: normal current datastores should not be rewritten when the old
// providerMetadata catalog key is missing.
// Inputs: v1alpha2 JSON with devices and no providerMetadata. Outputs: false from
// migrateInventoryMetadata.
// Data choice: the minimal current-shape JSON isolates the no-op branch.
func TestMigrateInventoryMetadataNoOp(t *testing.T) {
	raw := []byte(`{"schemaVersion": "v1alpha2", "devices": {}}`)
	inv := devicetypes.NewInventory()
	if migrateInventoryMetadata(raw, inv) {
		t.Error("expected no migration when providerMetadata is absent")
	}
}

// TestMigrateInventoryMetadataSkipsWhenTypedExists verifies typed Metadata wins over old catalogs.
//
// Why it matters: loading a datastore that already has current metadata should
// not overwrite user-visible roles, statuses, or tags with stale providerMetadata.
// Inputs: JSON containing both old providerMetadata and current metadata roles.
// Outputs: false from migration and preservation of the existing role.
// Data choice: conflicting role names make overwrite bugs visible.
func TestMigrateInventoryMetadataSkipsWhenTypedExists(t *testing.T) {
	raw := []byte(`{
		"schemaVersion": "v1alpha2",
		"providerMetadata": {"nautobot": {"roles": [{"name": "Old"}]}},
		"metadata": {"roles": [{"name": "Existing"}]},
		"devices": {}
	}`)

	inv := devicetypes.NewInventory()
	if err := json.Unmarshal(raw, inv); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if migrateInventoryMetadata(raw, inv) {
		t.Error("expected no migration when typed Metadata already has entries")
	}
	if inv.Metadata.Roles[0].Name != "Existing" {
		t.Errorf("existing role should be preserved, got %q", inv.Metadata.Roles[0].Name)
	}
}
