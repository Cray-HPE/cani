package transform

import (
	"sort"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/redfish/import"
)

// firstLibraryDeviceWithModel returns a deterministic embedded-library device
// that has both a slug and a model, so an exact-slug lookup yields a meaningful
// enrichment. It fails the test if the library has no such entry.
func firstLibraryDeviceWithModel(t *testing.T) devicetypes.CaniDeviceType {
	t.Helper()
	all := devicetypes.All()
	slugs := make([]string, 0, len(all))
	for s := range all {
		slugs = append(slugs, s)
	}
	sort.Strings(slugs)
	for _, s := range slugs {
		if dt := all[s]; dt.Slug != "" && dt.Model != "" {
			return dt
		}
	}
	t.Fatal("embedded device-type library has no entry with both slug and model")
	return devicetypes.CaniDeviceType{}
}

// containsString reports whether want appears in items.
func containsString(items []string, want string) bool {
	for _, s := range items {
		if s == want {
			return true
		}
	}
	return false
}

// TestEnrichDeviceFromLibrary_MatchAppliesDefaults verifies an exact library hit
// returns score 100 and copies the library defaults onto the device.
//
// Why it matters: enrichment is what upgrades a bare Redfish device into a fully
// classified inventory entry (slug, model, part number), so a confident match
// must populate those fields.
// Inputs: a real library device's slug as the root Product and an otherwise bare
// device. Outputs: the matched slug, winning query, score 100, and a device with
// the library slug and model applied.
// Data choice: querying by a known slug guarantees the exact-match branch (score
// 100) instead of a fuzzy score that could drift with library contents.
func TestEnrichDeviceFromLibrary_MatchAppliesDefaults(t *testing.T) {
	lib := firstLibraryDeviceWithModel(t)
	root := import_.ServiceRoot{Product: lib.Slug}
	dev := devicetypes.CaniDeviceType{Name: "raw", Type: devicetypes.TypeNode}

	slug, query, score := enrichDeviceFromLibrary(&dev, root)

	if slug != lib.Slug {
		t.Errorf("slug = %q, want %q", slug, lib.Slug)
	}
	if query != lib.Slug {
		t.Errorf("query = %q, want %q", query, lib.Slug)
	}
	if score != 100 {
		t.Errorf("score = %d, want 100", score)
	}
	if dev.Slug != lib.Slug {
		t.Errorf("dev.Slug = %q, want %q (defaults not applied)", dev.Slug, lib.Slug)
	}
	if dev.Model != lib.Model {
		t.Errorf("dev.Model = %q, want %q", dev.Model, lib.Model)
	}
}

// TestEnrichDeviceFromLibrary_NoMatch verifies an unrecognized product leaves the
// device untouched and reports an empty match.
//
// Why it matters: when no library entry fits, the import must keep the raw device
// as-is rather than stamping it with an unrelated slug.
// Inputs: a root whose Product is gibberish and a device with preset fields.
// Outputs: empty slug and query, score 0, and unchanged device fields.
// Data choice: a long unique nonsense string scores below the fuzzy-accept
// threshold, guaranteeing the no-match branch.
func TestEnrichDeviceFromLibrary_NoMatch(t *testing.T) {
	root := import_.ServiceRoot{Product: "zzz-nonexistent-hardware-9q8w7e6r5t"}
	dev := devicetypes.CaniDeviceType{Name: "raw", Manufacturer: "ACME", Type: devicetypes.TypeNode}

	slug, query, score := enrichDeviceFromLibrary(&dev, root)

	if slug != "" || query != "" || score != 0 {
		t.Errorf("got (%q, %q, %d), want an empty match", slug, query, score)
	}
	if dev.Slug != "" {
		t.Errorf("dev.Slug = %q, want empty (device must be untouched)", dev.Slug)
	}
	if dev.Manufacturer != "ACME" {
		t.Errorf("dev.Manufacturer = %q, want %q", dev.Manufacturer, "ACME")
	}
}

// TestBuildLookupQueries_IncludesOEMMonikerNames verifies the OEM PRODNAM and
// PRODGEN monikers are appended as lookup queries.
//
// Why it matters: HPE BMCs sometimes only spell out the model in OEM monikers, so
// those strings must be tried to maximize library hits.
// Inputs: a testRoot() with PRODNAM and PRODGEN set. Outputs: the ordered query
// list.
// Data choice: distinctive moniker values that do not collide with the other
// derived queries prove they are added rather than deduped away.
func TestBuildLookupQueries_IncludesOEMMonikerNames(t *testing.T) {
	root := testRoot()
	root.Oem.Hpe.Moniker.PRODNAM = "Apollo 6500 Gen11"
	root.Oem.Hpe.Moniker.PRODGEN = "Gen11 XL"

	queries := buildLookupQueries(root)

	if !containsString(queries, "Apollo 6500 Gen11") {
		t.Errorf("expected PRODNAM in queries: %v", queries)
	}
	if !containsString(queries, "Gen11 XL") {
		t.Errorf("expected PRODGEN in queries: %v", queries)
	}
}

// TestBuildLookupQueries_NoOEM verifies a root without OEM data yields only the
// product and vendor+product queries without dereferencing the nil OEM block.
//
// Why it matters: non-HPE or minimal BMCs have no Oem.Hpe, and query building
// must degrade gracefully instead of panicking on nil.
// Inputs: a ServiceRoot with Product and Vendor but nil Oem.Hpe. Outputs: exactly
// the product and "vendor product" queries, in order.
// Data choice: setting both Vendor and Product exercises the combined-query path
// while the empty ProductTag/SystemFamily/OEM paths are all skipped.
func TestBuildLookupQueries_NoOEM(t *testing.T) {
	root := import_.ServiceRoot{Product: "Generic Server", Vendor: "ACME"}

	queries := buildLookupQueries(root)

	want := []string{"Generic Server", "ACME Generic Server"}
	if len(queries) != len(want) {
		t.Fatalf("queries = %v, want %v", queries, want)
	}
	for i, q := range want {
		if queries[i] != q {
			t.Errorf("queries[%d] = %q, want %q", i, queries[i], q)
		}
	}
}

// TestApplyDeviceDefaults_KeepsExistingValues verifies populated device fields are
// never overwritten by library values.
//
// Why it matters: data already discovered from the live BMC is more trustworthy
// than library defaults, so enrichment must only fill gaps.
// Inputs: a fully populated device and a library entry with different values in
// every field. Outputs: a device whose fields all retain their originals.
// Data choice: making every library field differ ensures any accidental
// overwrite would be caught.
func TestApplyDeviceDefaults_KeepsExistingValues(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Slug: "keep-slug", PartNumber: "KEEP-PN", Manufacturer: "KeepMfr",
		Model: "KeepModel", Description: "KeepDesc", UHeight: 2,
		Type: devicetypes.Type("node"),
	}
	lib := devicetypes.CaniDeviceType{
		Slug: "lib-slug", PartNumber: "LIB-PN", Manufacturer: "LibMfr",
		Model: "LibModel", Description: "LibDesc", UHeight: 4,
		Type: devicetypes.Type("server"),
	}

	applyDeviceDefaults(dev, lib)

	if dev.Slug != "keep-slug" || dev.PartNumber != "KEEP-PN" ||
		dev.Manufacturer != "KeepMfr" || dev.Model != "KeepModel" ||
		dev.Description != "KeepDesc" || dev.UHeight != 2 || dev.Type != "node" {
		t.Errorf("library values overwrote existing device: %+v", dev)
	}
}

// TestApplyDeviceDefaults_FillsEmptyFields verifies every empty device field is
// populated from the library, including Description and an empty Type.
//
// Why it matters: a freshly built device starts mostly empty, and enrichment is
// responsible for completing it from the matched library entry.
// Inputs: a zero-value device and a fully populated library entry. Outputs: a
// device with all fields copied from the library.
// Data choice: an empty Type plus a node-typed library entry exercises the
// type-fill branch alongside the other field copies.
func TestApplyDeviceDefaults_FillsEmptyFields(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{}
	lib := devicetypes.CaniDeviceType{
		Slug: "lib-slug", PartNumber: "LIB-PN", Manufacturer: "LibMfr",
		Model: "LibModel", Description: "LibDesc", UHeight: 3,
		Type: devicetypes.Type("node"),
	}

	applyDeviceDefaults(dev, lib)

	if dev.Slug != "lib-slug" || dev.PartNumber != "LIB-PN" ||
		dev.Manufacturer != "LibMfr" || dev.Model != "LibModel" ||
		dev.Description != "LibDesc" || dev.UHeight != 3 || dev.Type != "node" {
		t.Errorf("empty fields not filled from library: %+v", dev)
	}
}

// TestApplyDeviceDefaults_ServerTypeKeptWhenLibTypeEmpty verifies a "server" type
// is left intact when the library entry has no type to offer.
//
// Why it matters: the placeholder "server" type should only be replaced by a real
// library type, never blanked out by an empty one.
// Inputs: a device typed "server" and a library entry with an empty Type.
// Outputs: a device still typed "server".
// Data choice: an empty library Type isolates the false branch of the inner
// non-empty-Type guard.
func TestApplyDeviceDefaults_ServerTypeKeptWhenLibTypeEmpty(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Type: devicetypes.Type("server")}
	lib := devicetypes.CaniDeviceType{Slug: "lib-slug"}

	applyDeviceDefaults(dev, lib)

	if dev.Type != "server" {
		t.Errorf("Type = %q, want %q", dev.Type, "server")
	}
}
