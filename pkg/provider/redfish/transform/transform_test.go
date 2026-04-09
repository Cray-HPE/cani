package transform

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/redfish/import"
	"github.com/google/uuid"
)

const (
	wantProduct  = "ProLiant DL325 Gen11"
	wantRedfish  = "1.20.0"
	wantProdtag  = "HPE iLO 6"
	hwTypeErrFmt = "HardwareType = %q, want %q"
	wantSlug     = "hpe-proliant-dl325-gen11"
	wantModel    = "DL325 Gen11"
	wantTestSlug = "test-slug"
)

func testRoot() import_.ServiceRoot {
	return import_.ServiceRoot{
		OdataType:      "#ServiceRoot.v1_13_0.ServiceRoot",
		ID:             "RootService",
		Name:           "HPE RESTful Root Service",
		Product:        wantProduct,
		Vendor:         "HPE",
		UUID:           "4f76fa4b-132e-59b2-bb1f-fcbec645cb17",
		RedfishVersion: wantRedfish,
		Oem: import_.OemData{
			Hpe: &import_.HpeOem{
				Manager: []import_.HpeManager{
					{
						FQDN:                   "pine-finch2-ilo.local",
						HostName:               "pine-finch2-ilo",
						ManagerFirmwareVersion: "1.61",
						ManagerType:            "iLO 6",
						Status:                 import_.HpeStatus{Health: "OK"},
					},
				},
				Moniker: import_.HpeMoniker{
					PRODTAG: wantProdtag,
					SYSFAM:  "ProLiant",
					VENDABR: "HPE",
					VENDNAM: "Hewlett-Packard Enterprise",
				},
			},
		},
	}
}

func TestBuildDeviceFromRoot(t *testing.T) {
	root := testRoot()
	dev := buildDeviceFromRoot(root, nil)
	if dev.Name != wantProduct {
		t.Errorf("Name = %q, want %q", dev.Name, wantProduct)
	}
	if dev.Manufacturer != "HPE" {
		t.Errorf("Manufacturer = %q, want %q", dev.Manufacturer, "HPE")
	}
	if dev.Type != devicetypes.TypeNode {
		t.Errorf("Type = %q, want %q", dev.Type, devicetypes.TypeNode)
	}
	if dev.ID == uuid.Nil {
		t.Error("ID is nil, expected generated UUID")
	}

	// Metadata must be nested under "redfish" key.
	redfishMeta, ok := dev.GetProviderSubMap("redfish")
	if !ok {
		t.Fatal("ProviderMetadata missing \"redfish\" sub-map")
	}

	wantKeys := []string{
		"redfish_version", "redfish_uuid", "vendor", "odata_type",
		"bmc_type", "bmc_firmware", "bmc_fqdn", "bmc_hostname",
		"product_tag", "system_family",
	}
	for _, key := range wantKeys {
		if _, ok := redfishMeta[key]; !ok {
			t.Errorf("ProviderMetadata[\"redfish\"] missing key %q", key)
		}
	}
	if redfishMeta["redfish_version"] != wantRedfish {
		t.Errorf("redfish_version = %v, want %q", redfishMeta["redfish_version"], wantRedfish)
	}
	if redfishMeta["bmc_type"] != "iLO 6" {
		t.Errorf("bmc_type = %v, want %q", redfishMeta["bmc_type"], "iLO 6")
	}

	// import_source key must exist in the redfish sub-map.
	redfishSub, _ := dev.GetProviderSubMap("redfish")
	if _, ok := redfishSub["import_source"]; !ok {
		t.Error("import_source key missing from redfish sub-map")
	}
}

func TestBuildLookupQueries(t *testing.T) {
	root := testRoot()
	queries := buildLookupQueries(root)
	if len(queries) == 0 {
		t.Fatal("expected at least one lookup query")
	}
	if queries[0] != wantProduct {
		t.Errorf("queries[0] = %q, want %q", queries[0], wantProduct)
	}
	hasTag := false
	hasCombined := false
	for _, q := range queries {
		if q == wantProdtag {
			hasTag = true
		}
		if q == "HPE "+wantProduct {
			hasCombined = true
		}
	}
	if !hasTag {
		t.Errorf("expected product tag in queries: %v", queries)
	}
	if !hasCombined {
		t.Errorf("expected combined vendor+product in queries: %v", queries)
	}
}

func TestBuildLookupQueriesNoDuplicates(t *testing.T) {
	root := testRoot()
	queries := buildLookupQueries(root)
	seen := make(map[string]bool)
	for _, q := range queries {
		if seen[q] {
			t.Errorf("duplicate query: %q", q)
		}
		seen[q] = true
	}
}

func TestTransformRoots(t *testing.T) {
	roots := []import_.ServiceRoot{testRoot()}
	result, err := transformRoots(roots, nil)
	if err != nil {
		t.Fatalf("transformRoots() error: %v", err)
	}
	if len(result.Devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(result.Devices))
	}
	if len(result.Modules) != 0 {
		t.Errorf("expected 0 modules, got %d", len(result.Modules))
	}
}

func TestTransformRootsEmpty(t *testing.T) {
	result, err := transformRoots([]import_.ServiceRoot{}, nil)
	if err != nil {
		t.Fatalf("transformRoots() error: %v", err)
	}
	if len(result.Devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(result.Devices))
	}
}

func TestApplyDeviceDefaults(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Name:         "test",
		Manufacturer: "HPE",
		Type:         devicetypes.Type("server"),
	}
	lib := devicetypes.CaniDeviceType{
		Slug:         wantSlug,
		PartNumber:   "P12345",
		Model:        wantModel,
		Description:  "HPE " + wantProduct,
		UHeight:      1,
		Manufacturer: "Hewlett Packard Enterprise",
		Type:         devicetypes.Type("node"),
	}
	applyDeviceDefaults(dev, lib)
	if dev.Slug != wantSlug {
		t.Errorf("Slug = %q, want %q", dev.Slug, wantSlug)
	}
	if dev.PartNumber != "P12345" {
		t.Errorf("PartNumber = %q, want %q", dev.PartNumber, "P12345")
	}
	if dev.Model != wantModel {
		t.Errorf("Model = %q, want %q", dev.Model, wantModel)
	}
	if dev.Manufacturer != "HPE" {
		t.Errorf("Manufacturer = %q, want %q (should keep original)", dev.Manufacturer, "HPE")
	}
	if dev.UHeight != 1 {
		t.Errorf("UHeight = %d, want %d", dev.UHeight, 1)
	}
	if dev.Type != "node" {
		t.Errorf(hwTypeErrFmt, dev.Type, "node")
	}
}

func TestBuildProviderMetadata(t *testing.T) {
	root := testRoot()
	meta := buildProviderMetadata(root)

	// Meta should be nested under "redfish" key.
	redfishMeta, ok := meta["redfish"].(map[string]any)
	if !ok {
		t.Fatal("buildProviderMetadata must nest under \"redfish\" key")
	}

	checks := map[string]string{
		"redfish_version": wantRedfish,
		"redfish_uuid":    "4f76fa4b-132e-59b2-bb1f-fcbec645cb17",
		"vendor":          "HPE",
		"bmc_type":        "iLO 6",
		"bmc_firmware":    "1.61",
		"product_tag":     wantProdtag,
		"system_family":   "ProLiant",
	}
	for key, want := range checks {
		got, ok := redfishMeta[key]
		if !ok {
			t.Errorf("missing key %q", key)
			continue
		}
		if got != want {
			t.Errorf("redfishMeta[%q] = %v, want %q", key, got, want)
		}
	}
}

func TestBuildRootStepInfo(t *testing.T) {
	root := testRoot()
	dev := buildDeviceFromRoot(root, nil)
	info := buildRootStepInfo(stepInput{
		Num:        1,
		Total:      1,
		Root:       root,
		Dev:        &dev,
		LibSlug:    wantTestSlug,
		MatchQuery: wantProduct,
		MatchScore: 85,
	})
	if info.NodeNum != 1 {
		t.Errorf("NodeNum = %d, want 1", info.NodeNum)
	}
	if info.RawName != wantProduct {
		t.Errorf("RawName = %q, want %q", info.RawName, wantProduct)
	}
	if info.RawType != "server" {
		t.Errorf("RawType = %q, want %q", info.RawType, "server")
	}
	if info.MatchScore != 85 {
		t.Errorf("MatchScore = %d, want 85", info.MatchScore)
	}
	if info.LibMatch != wantTestSlug {
		t.Errorf("LibMatch = %q, want %q", info.LibMatch, wantTestSlug)
	}
	if len(info.Mappings) == 0 {
		t.Error("expected non-empty Mappings")
	}
}

func TestIdempotentReimport(t *testing.T) {
	root := testRoot()

	// First import: no existing inventory.
	dev1 := buildDeviceFromRoot(root, nil)
	if dev1.ID == uuid.Nil {
		t.Fatal("first import produced nil ID")
	}

	// Build an existing inventory containing the first device.
	existing := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			dev1.ID: &dev1,
		},
	}

	// Second import: supply the existing inventory.
	dev2 := buildDeviceFromRoot(root, existing)

	// The second import must reuse the UUID from the first.
	if dev2.ID != dev1.ID {
		t.Errorf("re-import generated new UUID %s, want existing %s", dev2.ID, dev1.ID)
	}
}

func TestIdempotentNoMatchGeneratesNewID(t *testing.T) {
	root := testRoot()

	// Existing inventory with a different device (different redfish_uuid).
	other := devicetypes.CaniDeviceType{
		ID: uuid.New(),
		ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{
			"redfish": map[string]any{
				"redfish_uuid": "00000000-0000-0000-0000-000000000000",
			},
		}},
	}
	existing := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			other.ID: &other,
		},
	}

	dev := buildDeviceFromRoot(root, existing)
	if dev.ID == other.ID {
		t.Error("should not match a different device's UUID")
	}
}
