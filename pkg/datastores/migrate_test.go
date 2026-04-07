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

func TestIsLegacyDatastoreV1Alpha1(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/fixtures/cani/legacy/canitestdb_v1alpha1.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	if !isLegacyDatastore(raw) {
		t.Error("expected v1alpha1 fixture to be detected as legacy")
	}
}

func TestIsLegacyDatastoreV1Alpha2(t *testing.T) {
	inv := devicetypes.NewInventory()
	data, _ := json.Marshal(inv)
	if isLegacyDatastore(data) {
		t.Error("v1alpha2 inventory should NOT be detected as legacy")
	}
}

func TestIsLegacyDatastoreInvalidJSON(t *testing.T) {
	if isLegacyDatastore([]byte("{bad json}")) {
		t.Error("invalid JSON should not be detected as legacy")
	}
}

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

func TestMigrateSchemaVersion(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}
	if inv.SchemaVersion != devicetypes.SchemaVersionV1Alpha2 {
		t.Errorf("SchemaVersion = %q, want %q", inv.SchemaVersion, devicetypes.SchemaVersionV1Alpha2)
	}
}

func TestMigrateProvider(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}
	if inv.Provider != "csm" {
		t.Errorf("Provider = %q, want %q", inv.Provider, "csm")
	}
}

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
	if _, ok := inv.Racks[dev.Parent]; !ok {
		t.Errorf("Cabinet device parent %v does not exist in Racks", dev.Parent)
	}
}

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
	// JSON numbers unmarshal as float64 through intermediate map[string]any.
	if v, ok := vlan.(int); ok {
		if v != 1513 {
			t.Errorf("hmnVlan = %d, want 1513", v)
		}
	} else if v, ok := vlan.(float64); ok {
		if int(v) != 1513 {
			t.Errorf("hmnVlan = %v, want 1513", v)
		}
	} else {
		t.Errorf("hmnVlan unexpected type %T", vlan)
	}
}

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

func TestMigrateNodeNidAndAliases(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}
	nodeID := uuid.MustParse("00000001-0000-0000-0000-000000000005")
	csm, _ := inv.Devices[nodeID].GetProviderSubMap("csm")

	nid := csm["nid"]
	switch v := nid.(type) {
	case int:
		if v != 1001 {
			t.Errorf("nid = %d, want 1001", v)
		}
	case float64:
		if int(v) != 1001 {
			t.Errorf("nid = %v, want 1001", v)
		}
	default:
		t.Errorf("nid unexpected type %T", nid)
	}

	aliases, ok := csm["aliases"]
	if !ok {
		t.Fatal("Node missing aliases in csm metadata")
	}
	list, ok := aliases.([]string)
	if !ok {
		// Could be []interface{} after JSON round-trip
		if iface, ok := aliases.([]interface{}); ok {
			if len(iface) != 1 {
				t.Errorf("aliases length = %d, want 1", len(iface))
			}
		} else {
			t.Errorf("aliases unexpected type %T", aliases)
		}
	} else if len(list) != 1 || list[0] != "nid001001" {
		t.Errorf("aliases = %v, want [nid001001]", list)
	}
}

func TestMigrateLocationOrdinalInMetadata(t *testing.T) {
	inv, err := migrateV1Alpha1(loadFixture(t))
	if err != nil {
		t.Fatalf("migrateV1Alpha1: %v", err)
	}
	cabinetID := uuid.MustParse("00000001-0000-0000-0000-000000000002")
	csm, _ := inv.Devices[cabinetID].GetProviderSubMap("csm")
	ord := csm["locationOrdinal"]
	switch v := ord.(type) {
	case int:
		if v != 3000 {
			t.Errorf("locationOrdinal = %d, want 3000", v)
		}
	case float64:
		if int(v) != 3000 {
			t.Errorf("locationOrdinal = %v, want 3000", v)
		}
	default:
		t.Errorf("locationOrdinal unexpected type %T", ord)
	}
}

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
}

// ---------- Load integration ----------

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
}

// ---------- migrateInventoryMetadata ----------

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

func TestMigrateInventoryMetadataNoOp(t *testing.T) {
	raw := []byte(`{"schemaVersion": "v1alpha2", "devices": {}}`)
	inv := devicetypes.NewInventory()
	if migrateInventoryMetadata(raw, inv) {
		t.Error("expected no migration when providerMetadata is absent")
	}
}

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
