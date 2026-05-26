package devicetypes

import (
	"testing"

	"github.com/google/uuid"
)

func TestGetProviderMeta(t *testing.T) {
	dev := &CaniDeviceType{
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{
				"import_source": "test.json",
				"redfish_uuid":  "abc-123",
				"bmc_fqdn":      "host.example.com",
			},
		}},
	}

	// Nested import_source key.
	if v, ok := dev.GetProviderMeta("import_source"); !ok || v != "test.json" {
		t.Errorf("GetProviderMeta(import_source) = %v, %v", v, ok)
	}

	// Nested key.
	if v, ok := dev.GetProviderMeta("redfish_uuid"); !ok || v != "abc-123" {
		t.Errorf("GetProviderMeta(redfish_uuid) = %v, %v", v, ok)
	}

	// Missing key.
	if _, ok := dev.GetProviderMeta("nonexistent"); ok {
		t.Error("GetProviderMeta(nonexistent) should return false")
	}

	// Nil ObjectMeta.
	var nilMeta *ObjectMeta
	if _, ok := nilMeta.GetProviderMeta("anything"); ok {
		t.Error("nil ObjectMeta should return false")
	}
}

func TestGetProviderSubMap(t *testing.T) {
	dev := &CaniDeviceType{
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"csm": map[string]any{
				"xname": "x3000c0s1b0n0",
			},
		}},
	}

	sub, ok := dev.GetProviderSubMap("csm")
	if !ok {
		t.Fatal("expected csm sub-map")
	}
	if sub["xname"] != "x3000c0s1b0n0" {
		t.Errorf("xname = %v, want x3000c0s1b0n0", sub["xname"])
	}

	_, ok = dev.GetProviderSubMap("redfish")
	if ok {
		t.Error("should not find redfish sub-map")
	}
}

func TestSetProviderMeta(t *testing.T) {
	dev := &CaniDeviceType{}
	dev.SetProviderMeta("redfish", "redfish_uuid", "abc-123")

	sub, ok := dev.GetProviderSubMap("redfish")
	if !ok {
		t.Fatal("expected redfish sub-map after SetProviderMeta")
	}
	if sub["redfish_uuid"] != "abc-123" {
		t.Errorf("redfish_uuid = %v, want abc-123", sub["redfish_uuid"])
	}

	// Overwrite existing key.
	dev.SetProviderMeta("redfish", "redfish_uuid", "xyz-789")
	if sub["redfish_uuid"] != "xyz-789" {
		t.Errorf("redfish_uuid = %v, want xyz-789", sub["redfish_uuid"])
	}
}

func TestSetImportSource(t *testing.T) {
	dev := &CaniDeviceType{}
	dev.SetImportSource("redfish", "testdata/root.json")

	src := dev.GetImportSource("redfish")
	want := "testdata/root.json"
	if src != want {
		t.Errorf("GetImportSource(redfish) = %q, want %q", src, want)
	}

	// Verify it's nested inside the provider sub-map.
	sub, ok := dev.GetProviderSubMap("redfish")
	if !ok {
		t.Fatal("expected redfish sub-map")
	}
	if sub["import_source"] != "testdata/root.json" {
		t.Errorf("sub[import_source] = %v, want testdata/root.json", sub["import_source"])
	}

	// Verify no top-level import_source key.
	if _, exists := dev.ProviderMetadata["import_source"]; exists {
		t.Error("import_source should not exist at top level")
	}
}

func TestFlattenProviderMetadata(t *testing.T) {
	dev := &CaniDeviceType{
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{
				"import_source": "test.json",
				"redfish_uuid":  "abc-123",
				"vendor":        "HPE",
			},
		}},
	}

	flat := dev.FlattenProviderMetadata()
	if flat["import_source"] != "test.json" {
		t.Errorf("import_source = %v", flat["import_source"])
	}
	if flat["redfish_uuid"] != "abc-123" {
		t.Errorf("redfish_uuid = %v", flat["redfish_uuid"])
	}
	if flat["vendor"] != "HPE" {
		t.Errorf("vendor = %v", flat["vendor"])
	}
}

func TestFlattenProviderMetadataEmpty(t *testing.T) {
	var m *ObjectMeta
	flat := m.FlattenProviderMetadata()
	if flat != nil {
		t.Errorf("expected nil for nil ObjectMeta, got %v", flat)
	}
}

func TestFindDeviceByProviderKey(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	inv := &Inventory{
		Devices: map[uuid.UUID]*CaniDeviceType{
			id1: {
				ID:   id1,
				Name: "server1",
				ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
					"redfish": map[string]any{
						"redfish_uuid": "uuid-aaa",
						"bmc_fqdn":     "server1.example.com",
					},
				}},
			},
			id2: {
				ID:   id2,
				Name: "server2",
				ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
					"redfish": map[string]any{
						"redfish_uuid": "uuid-bbb",
						"bmc_fqdn":     "server2.example.com",
					},
				}},
			},
		},
	}

	// Match by redfish_uuid.
	found := inv.FindDeviceByProviderKey("redfish", "redfish_uuid", "uuid-bbb")
	if found == nil || found.ID != id2 {
		t.Errorf("expected server2, got %v", found)
	}

	// Match by bmc_fqdn.
	found = inv.FindDeviceByProviderKey("redfish", "bmc_fqdn", "server1.example.com")
	if found == nil || found.ID != id1 {
		t.Errorf("expected server1, got %v", found)
	}

	// No match.
	found = inv.FindDeviceByProviderKey("redfish", "redfish_uuid", "nonexistent")
	if found != nil {
		t.Errorf("expected nil, got %v", found)
	}

	// Wrong provider.
	found = inv.FindDeviceByProviderKey("csm", "redfish_uuid", "uuid-aaa")
	if found != nil {
		t.Errorf("expected nil for wrong provider, got %v", found)
	}
}

func TestFindDeviceByProviderKeys(t *testing.T) {
	id := uuid.New()
	inv := &Inventory{
		Devices: map[uuid.UUID]*CaniDeviceType{
			id: {
				ID:   id,
				Name: "server1",
				ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
					"redfish": map[string]any{
						"redfish_uuid": "uuid-aaa",
						"bmc_fqdn":     "server1.example.com",
					},
				}},
			},
		},
	}

	// First check fails, second succeeds.
	checks := []ProviderKeyCheck{
		{Key: "redfish_uuid", Value: "no-match"},
		{Key: "bmc_fqdn", Value: "server1.example.com"},
	}
	found := inv.FindDeviceByProviderKeys("redfish", checks)
	if found == nil || found.ID != id {
		t.Errorf("expected server1 via fallback, got %v", found)
	}

	// All checks fail.
	checks = []ProviderKeyCheck{
		{Key: "redfish_uuid", Value: "no-match"},
		{Key: "bmc_fqdn", Value: "no-match"},
	}
	found = inv.FindDeviceByProviderKeys("redfish", checks)
	if found != nil {
		t.Errorf("expected nil, got %v", found)
	}
}

func TestMergePropertiesWithMetadata(t *testing.T) {
	existing := &CaniDeviceType{
		ID:   uuid.New(),
		Name: "old-name",
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{
				"import_source": "old.json",
				"redfish_uuid":  "uuid-aaa",
				"bmc_firmware":  "1.50",
			},
		}},
	}
	incoming := &CaniDeviceType{
		Name: "new-name",
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{
				"import_source": "new.json",
				"redfish_uuid":  "uuid-aaa",
				"bmc_firmware":  "1.61",
				"bmc_fqdn":      "host.example.com",
			},
		}},
	}

	changed := existing.MergeProperties(incoming)
	if !changed {
		t.Error("expected changes from merge")
	}
	if existing.Name != "new-name" {
		t.Errorf("Name = %q, want new-name", existing.Name)
	}
	if existing.GetImportSource("redfish") != "new.json" {
		t.Errorf("import_source = %q, want new.json", existing.GetImportSource("redfish"))
	}
	redfishMeta, _ := existing.GetProviderSubMap("redfish")
	if redfishMeta["bmc_firmware"] != "1.61" {
		t.Errorf("bmc_firmware = %v, want 1.61", redfishMeta["bmc_firmware"])
	}
	if redfishMeta["bmc_fqdn"] != "host.example.com" {
		t.Errorf("bmc_fqdn = %v, want host.example.com", redfishMeta["bmc_fqdn"])
	}
}
