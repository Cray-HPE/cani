package transform

import (
	"testing"

	import_ "github.com/Cray-HPE/cani/pkg/provider/redfish/import"
	"github.com/Cray-HPE/cani/pkg/visual"
)

// findMapping returns the first field mapping whose TargetField matches, and
// whether one was found.
func findMapping(maps []visual.FieldMapping, targetField string) (visual.FieldMapping, bool) {
	for _, m := range maps {
		if m.TargetField == targetField {
			return m, true
		}
	}
	return visual.FieldMapping{}, false
}

// assertMappingValue checks that a mapping for field exists with the wanted
// TargetValue.
func assertMappingValue(t *testing.T, maps []visual.FieldMapping, field, want string) {
	t.Helper()
	m, ok := findMapping(maps, field)
	if !ok {
		t.Errorf("missing mapping for %q", field)
		return
	}
	if m.TargetValue != want {
		t.Errorf("mapping[%q] TargetValue = %q, want %q", field, m.TargetValue, want)
	}
}

// TestBuildFieldMappings_BaseAndOEM verifies the four base mappings plus the OEM
// BMC mappings are emitted with the device's resolved values.
//
// Why it matters: the step-through display must faithfully show how each raw
// Redfish field lands in the CANI device so operators can audit the transform.
// Inputs: a full testRoot(), its built device, and an empty library slug.
// Outputs: a mapping slice asserted by target field and value.
// Data choice: testRoot() carries FQDN, hostname, manager type, and firmware, so
// every OEM mapping branch fires; the empty libSlug proves no library rows leak
// in.
func TestBuildFieldMappings_BaseAndOEM(t *testing.T) {
	root := testRoot()
	dev := buildDeviceFromRoot(root, nil)

	mappings := buildFieldMappings(root, &dev, "")

	wantValues := map[string]string{
		"Name":         wantProduct,
		"Manufacturer": "HPE",
		"ProviderMetadata[redfish][redfish_uuid]":    root.UUID,
		"ProviderMetadata[redfish][redfish_version]": wantRedfish,
		"ProviderMetadata[redfish][bmc_fqdn]":        "pine-finch2-ilo.local",
		"ProviderMetadata[redfish][bmc_hostname]":    "pine-finch2-ilo",
		"ProviderMetadata[redfish][bmc_type]":        "iLO 6",
		"ProviderMetadata[redfish][bmc_firmware]":    "1.61",
	}
	for field, want := range wantValues {
		assertMappingValue(t, mappings, field, want)
	}
	if _, ok := findMapping(mappings, "Slug"); ok {
		t.Error("did not expect a library Slug mapping when libSlug is empty")
	}
}

// TestBuildFieldMappings_NoOEM verifies only the four base mappings appear when
// the root has no OEM manager data.
//
// Why it matters: the display must not invent BMC rows for fields the BMC never
// reported.
// Inputs: a minimal ServiceRoot and its device, with an empty library slug.
// Outputs: a four-element mapping slice.
// Data choice: omitting OEM data drives the false branch of every non-empty
// manager-field guard.
func TestBuildFieldMappings_NoOEM(t *testing.T) {
	root := import_.ServiceRoot{
		Product: "Generic", Vendor: "ACME", UUID: "u-1", RedfishVersion: "1.0",
	}
	dev := buildDeviceFromRoot(root, nil)

	mappings := buildFieldMappings(root, &dev, "")

	if len(mappings) != 4 {
		t.Fatalf("len(mappings) = %d, want 4", len(mappings))
	}
	if _, ok := findMapping(mappings, "ProviderMetadata[redfish][bmc_fqdn]"); ok {
		t.Error("did not expect a bmc_fqdn mapping for a root without OEM data")
	}
}

// TestBuildFieldMappings_LibraryEnrichment verifies derived library mappings are
// added for slug, model, part number, and U-height when present.
//
// Why it matters: when a library match enriches the device, the display must mark
// those values as derived so operators see they came from the library, not the
// BMC.
// Inputs: a device populated with library-sourced fields and a non-empty library
// slug. Outputs: derived mappings flagged IsDerived.
// Data choice: a UHeight of 2 verifies the integer is formatted as the string
// "2".
func TestBuildFieldMappings_LibraryEnrichment(t *testing.T) {
	root := testRoot()
	dev := buildDeviceFromRoot(root, nil)
	dev.Slug = "lib-slug"
	dev.Model = "DL325 Gen11"
	dev.PartNumber = "P12345"
	dev.UHeight = 2

	mappings := buildFieldMappings(root, &dev, "lib-slug")

	slugMap, ok := findMapping(mappings, "Slug")
	if !ok || !slugMap.IsDerived {
		t.Errorf("expected a derived Slug mapping, got %+v (found=%v)", slugMap, ok)
	}
	assertMappingValue(t, mappings, "Model", "DL325 Gen11")
	assertMappingValue(t, mappings, "PartNumber", "P12345")
	assertMappingValue(t, mappings, "UHeight", "2")
}

// TestBuildFieldMappings_LibrarySlugOnly verifies that with a library slug but no
// enriched model, part number, or U-height, only the Slug derived mapping is
// added.
//
// Why it matters: the display must not show derived rows for library fields that
// did not actually populate the device.
// Inputs: a device with only a Slug set and a non-empty library slug. Outputs: a
// Slug mapping present while Model, PartNumber, and UHeight are absent.
// Data choice: leaving Model/PartNumber empty and UHeight zero drives the false
// branch of each optional library mapping.
func TestBuildFieldMappings_LibrarySlugOnly(t *testing.T) {
	root := import_.ServiceRoot{Product: "Generic", UUID: "u-1"}
	dev := buildDeviceFromRoot(root, nil)
	dev.Slug = "lib-slug"

	mappings := buildFieldMappings(root, &dev, "lib-slug")

	if _, ok := findMapping(mappings, "Slug"); !ok {
		t.Error("expected a Slug mapping")
	}
	for _, field := range []string{"Model", "PartNumber", "UHeight"} {
		if _, ok := findMapping(mappings, field); ok {
			t.Errorf("did not expect a %q mapping when the field is empty", field)
		}
	}
}
