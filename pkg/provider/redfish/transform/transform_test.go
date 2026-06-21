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

// TestBuildDeviceFromRoot verifies a CaniDeviceType built from a ServiceRoot maps
// the core fields, nests all metadata under the "redfish" key, and stamps an
// import source.
//
// Why it matters: buildDeviceFromRoot is the first step of the Redfish import
// that turns a raw BMC response into a CANI device, so its field and metadata
// mapping is what the rest of the inventory depends on.
// Inputs: a full testRoot() and a nil existing inventory. Outputs: a device whose
// Name, Manufacturer, Type, generated ID, and nested redfish metadata are
// asserted.
// Data choice: the HPE iLO root populates every optional metadata key (bmc_*,
// product_tag, system_family) so the full expected key set can be checked at once.
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

// TestBuildLookupQueries verifies lookup-query building orders the Product field
// first and also includes the HPE product tag and the vendor+product
// combination.
//
// Why it matters: these queries drive the device-type library match, so the most
// specific string must be tried first to maximize classification accuracy.
// Inputs: a full testRoot(). Outputs: the ordered query slice, checked for its
// first element and for membership of the tag and combined queries.
// Data choice: the root carries both a PRODTAG and distinct Vendor and Product
// values so the tag and combined-query branches each produce a checkable string.
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

// TestBuildLookupQueriesNoDuplicates verifies lookup-query building never emits
// the same query string twice.
//
// Why it matters: duplicate queries waste library lookups and could skew match
// scoring, so deduplication keeps enrichment efficient and deterministic.
// Inputs: a full testRoot(). Outputs: the query slice, scanned for repeats with a
// seen-set.
// Data choice: an HPE root whose vendor and product strings overlap is the case
// most likely to generate collisions, so it exercises the dedup guard.
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

// TestTransformRoots verifies transforming one ServiceRoot yields exactly one
// device with the expected identity and no non-device core inventory objects.
//
// Why it matters: transformRoots is the core conversion loop of the Redfish
// import, so each root must map to the correct CANI device entry without
// fabricating Nautobot-equivalent records that are not present in ServiceRoot.
// Inputs: a one-element root slice and a nil existing inventory. Outputs: a
// TransformResult whose single device identity, map key, and empty non-device
// buckets are asserted.
// Data choice: a single root isolates the one-root-to-one-device contract, and
// the full HPE root makes the resulting device fields concrete.
func TestTransformRoots(t *testing.T) {
	roots := []import_.ServiceRoot{testRoot()}
	result, err := transformRoots(roots, nil)
	if err != nil {
		t.Fatalf("transformRoots() error: %v", err)
	}
	deviceID, dev := singleResultDevice(t, result)
	if dev.ID != deviceID {
		t.Errorf("device map key = %s, but device.ID = %s", deviceID, dev.ID)
	}
	if dev.Name != wantProduct {
		t.Errorf("device Name = %q, want %q", dev.Name, wantProduct)
	}
	if dev.Manufacturer != "HPE" {
		t.Errorf("device Manufacturer = %q, want %q", dev.Manufacturer, "HPE")
	}
	if dev.Type == "" {
		t.Error("device Type is empty, want a CANI hardware type")
	}
	assertUnsupportedCoreTypesEmpty(t, result)
}

// TestTransformRootsEmpty verifies transforming an empty root slice returns a
// result with no devices and no error.
//
// Why it matters: importing from a source with nothing to offer must be a
// graceful no-op rather than an error that aborts the import pipeline.
// Inputs: an empty root slice and a nil existing inventory. Outputs: a
// TransformResult with zero devices and a nil error.
// Data choice: an empty slice is the canonical "nothing to import" signal the
// conversion loop must tolerate.
func TestTransformRootsEmpty(t *testing.T) {
	result, err := transformRoots([]import_.ServiceRoot{}, nil)
	if err != nil {
		t.Fatalf("transformRoots() error: %v", err)
	}
	if len(result.Devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(result.Devices))
	}
}

// TestApplyDeviceDefaults verifies applying library defaults fills empty device
// fields while preserving values already set, and replaces the placeholder
// "server" type with the library's "node".
//
// Why it matters: enrichment must complete a bare device from the matched library
// entry without clobbering data already discovered from the BMC.
// Inputs: a device with Name, Manufacturer, and a "server" type set, plus a fully
// populated library entry. Outputs: a device whose Slug, PartNumber, Model, and
// UHeight come from the library, whose Manufacturer stays "HPE", and whose Type
// becomes "node".
// Data choice: a differing library Manufacturer proves the keep-existing rule,
// while the server-to-node pairing exercises the placeholder-type replacement.
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

// TestBuildProviderMetadata verifies provider-metadata building nests all Redfish
// fields under the "redfish" key with the expected values.
//
// Why it matters: this metadata is how later re-imports identify the device and
// how operators trace its origin, so the keys and values must be exact.
// Inputs: a full testRoot(). Outputs: the nested redfish sub-map, checked for
// seven key and value pairs.
// Data choice: the HPE iLO root supplies real version, UUID, BMC, tag, and family
// values so every asserted key has a concrete expected value.
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

// TestBuildRootStepInfo verifies step-info building copies the position, raw
// name and type, match score, library slug, and concrete field mappings into the
// display struct.
//
// Why it matters: StepMode shows operators how each raw field maps to CANI before
// committing, so the display struct must carry the right summary values and the
// right source-to-target mapping details.
// Inputs: a stepInput with num and total of 1, the built device, a test slug, and
// a score of 85. Outputs: a NodeStepInfo whose header fields and selected
// mapping contents are asserted.
// Data choice: a score of 85 and a distinct test slug are recognizable markers
// that prove the values are passed through rather than defaulted; setting the
// device slug mirrors the real enrichment path when LibSlug is non-empty.
func TestBuildRootStepInfo(t *testing.T) {
	root := testRoot()
	dev := buildDeviceFromRoot(root, nil)
	dev.Slug = wantTestSlug
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
	assertMappingValue(t, info.Mappings, "Name", wantProduct)
	assertMappingValue(t, info.Mappings, "Manufacturer", "HPE")
	assertMappingValue(t, info.Mappings, "ProviderMetadata[redfish][redfish_uuid]", root.UUID)
	slugMapping, ok := findMapping(info.Mappings, "Slug")
	if !ok {
		t.Fatal("expected Slug mapping")
	}
	if !slugMapping.IsDerived {
		t.Error("Slug mapping IsDerived = false, want true")
	}
	if slugMapping.TargetValue != wantTestSlug {
		t.Errorf("Slug mapping TargetValue = %q, want %q", slugMapping.TargetValue, wantTestSlug)
	}
}

// TestIdempotentReimport verifies re-importing the same root against an inventory
// that already contains it reuses the existing device UUID.
//
// Why it matters: idempotent imports are essential so repeated syncs update the
// device in place instead of creating duplicates.
// Inputs: testRoot() imported once with a nil inventory, then again with an
// inventory holding that first device. Outputs: the second device's ID, asserted
// equal to the first.
// Data choice: reusing the same root for both imports isolates the dedup behavior
// to the inventory argument alone.
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

// TestIdempotentNoMatchGeneratesNewID verifies importing a root whose
// redfish_uuid matches no existing device produces a new UUID rather than reusing
// an unrelated device.
//
// Why it matters: the dedup logic must distinguish distinct hardware, or separate
// servers would be conflated under a single identity.
// Inputs: testRoot() and an inventory whose only device has a different
// redfish_uuid. Outputs: the new device's ID, asserted not equal to the unrelated
// device's ID.
// Data choice: an all-zeros redfish_uuid is an unmistakably different value,
// guaranteeing the no-match branch.
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
