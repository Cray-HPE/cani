package import_

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	fixtureFile      = "redfish-root.json"
	arrayFixtureFile = "redfish-root-array.json"
	parseErrFmt      = "ParseServiceRoots() error: %v"
	wantProduct      = "ProLiant DL325 Gen11"
	wantServerA      = "Server A"
)

// TestParseServiceRootsSingleObject verifies a single Redfish ServiceRoot JSON
// object is parsed into one ServiceRoot with the expected core fields.
//
// Why it matters: Redfish import accepts a single BMC response as its primary
// input shape, so the parser must preserve the identifiers consumed later by the
// transform phase.
// Inputs: the redfish-root.json fixture. Outputs: one ServiceRoot whose product,
// vendor, UUID, Redfish version, and name match the fixture.
// Data choice: the fixture is a realistic HPE iLO ServiceRoot with populated
// product, vendor, UUID, and version fields.
func TestParseServiceRootsSingleObject(t *testing.T) {
	data := loadFixture(t, fixtureFile)
	roots, err := ParseServiceRoots(data)
	if err != nil {
		t.Fatalf("ParseServiceRoots() error: %v", err)
	}
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	root := roots[0]
	assertField(t, "Product", root.Product, wantProduct)
	assertField(t, "Vendor", root.Vendor, "HPE")
	assertField(t, "UUID", root.UUID, "946a7d44-9967-4940-9490-f2d581950512")
	assertField(t, "RedfishVersion", root.RedfishVersion, "1.20.0")
	assertField(t, "Name", root.Name, "HPE RESTful Root Service")
}

// TestParseServiceRootsArray verifies an array of Redfish ServiceRoot objects is
// parsed without collapsing or reordering the records.
//
// Why it matters: multi-node imports can feed an array into the import stage,
// and deduplication is a later step, not part of JSON parsing.
// Inputs: the redfish-root-array.json fixture. Outputs: two ServiceRoot records
// with distinct UUIDs and BMC FQDNs in fixture order.
// Data choice: the fixture contains two realistic HPE iLO roots with different
// Redfish UUIDs and endpoint names, which makes identity/order assertions useful.
func TestParseServiceRootsArray(t *testing.T) {
	roots, err := ParseServiceRoots(loadFixture(t, arrayFixtureFile))
	if err != nil {
		t.Fatalf(parseErrFmt, err)
	}
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
	checks := []struct {
		index int
		uuid  string
		fqdn  string
	}{
		{index: 0, uuid: "946a7d44-9967-4940-9490-f2d581950512", fqdn: "bin.example.com"},
		{index: 1, uuid: "4f76fa4b-132e-59b2-bb1f-fcbec645cb17", fqdn: "baz.example.com"},
	}
	for _, check := range checks {
		root := roots[check.index]
		assertField(t, "Product", root.Product, wantProduct)
		assertField(t, "UUID", root.UUID, check.uuid)
		assertField(t, "ManagerFQDN", root.ManagerFQDN(), check.fqdn)
	}
}

// TestParseServiceRootsInvalid verifies object-shaped JSON that lacks any
// ServiceRoot identity fields is rejected.
//
// Why it matters: accepting arbitrary JSON as a root would store unusable records
// and defer a clear input error until later import phases.
// Inputs: a syntactically valid non-ServiceRoot JSON object. Outputs: a non-nil
// parser error.
// Data choice: the object has an unrelated key, exercising semantic validation
// after JSON decoding succeeds.
func TestParseServiceRootsInvalid(t *testing.T) {
	_, err := ParseServiceRoots([]byte(`{"not": "a service root"}`))
	if err == nil {
		t.Error("expected error for non-ServiceRoot JSON, got nil")
	}
}

// TestParseServiceRootsBadJSON verifies malformed JSON is returned as a parse
// error.
//
// Why it matters: the import command should fail loudly on unreadable Redfish
// input rather than attempting to store a partial record.
// Inputs: an invalid JSON byte slice. Outputs: a non-nil parser error.
// Data choice: a truncated object is the smallest malformed input that exercises
// the JSON decoder error path.
func TestParseServiceRootsBadJSON(t *testing.T) {
	_, err := ParseServiceRoots([]byte(`{bad json`))
	if err == nil {
		t.Error("expected error for bad JSON, got nil")
	}
}

// TestServiceRootOemAccessors verifies HPE OEM manager and moniker accessors
// return the first manager's concrete values.
//
// Why it matters: transform and dedup logic rely on these accessors for BMC
// identity, firmware metadata, and library lookup hints.
// Inputs: the redfish-root.json fixture parsed through ParseServiceRoots.
// Outputs: manager type, firmware, health, FQDN, hostname, product tag, and
// system family values.
// Data choice: the fixture has a fully populated HPE OEM block, so every accessor
// should return a real value instead of the empty fallback.
func TestServiceRootOemAccessors(t *testing.T) {
	data := loadFixture(t, fixtureFile)
	roots, err := ParseServiceRoots(data)
	if err != nil {
		t.Fatalf(parseErrFmt, err)
	}
	root := roots[0]
	assertField(t, "ManagerType", root.ManagerType(), "iLO 6")
	assertField(t, "ManagerFirmwareVersion", root.ManagerFirmwareVersion(), "1.61")
	assertField(t, "ManagerHealth", root.ManagerHealth(), "OK")
	assertField(t, "ProductTag", root.ProductTag(), "HPE iLO 6")
	assertField(t, "SystemFamily", root.SystemFamily(), "ProLiant")
	assertField(t, "ManagerFQDN", root.ManagerFQDN(), "foo.example.com")
	assertField(t, "ManagerHostName", root.ManagerHostName(), "foo")
}

// TestServiceRootOemAccessorsNoOEM verifies HPE OEM accessors return empty
// strings when OEM data or manager records are absent.
//
// Why it matters: Redfish can come from non-HPE or minimal BMCs, and callers use
// empty strings to decide whether optional metadata is present.
// Inputs: ServiceRoot values with no OEM block and with an empty HPE manager
// list. Outputs: empty accessor strings for manager, tag, and family data.
// Data choice: both nil OEM and empty-manager cases exercise the defensive guards
// around nested optional fields.
func TestServiceRootOemAccessorsNoOEM(t *testing.T) {
	cases := []struct {
		name string
		root ServiceRoot
	}{
		{name: "nil HPE OEM", root: ServiceRoot{}},
		{name: "empty manager list", root: ServiceRoot{Oem: OemData{Hpe: &HpeOem{}}}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			root := tt.root
			for name, got := range map[string]string{
				"ManagerType":            root.ManagerType(),
				"ManagerFirmwareVersion": root.ManagerFirmwareVersion(),
				"ManagerFQDN":            root.ManagerFQDN(),
				"ManagerHostName":        root.ManagerHostName(),
				"ManagerHealth":          root.ManagerHealth(),
				"ProductTag":             root.ProductTag(),
				"SystemFamily":           root.SystemFamily(),
			} {
				if got != "" {
					t.Errorf("%s = %q, want empty", name, got)
				}
			}
		})
	}
}

// TestDeduplicateRoots verifies duplicate roots with the same deduplication key
// keep the first record while preserving distinct later records.
//
// Why it matters: import must be idempotent for repeated BMC records but must not
// reorder or overwrite the first instance that will be transformed later.
// Inputs: three roots where the first and third share UUID "aaa" and the second
// has UUID "bbb". Outputs: two roots, preserving the first two products in order.
// Data choice: a duplicate placed after a distinct record proves both duplicate
// rejection and stable order.
func TestDeduplicateRoots(t *testing.T) {
	roots := []ServiceRoot{
		{UUID: "aaa", Product: wantServerA},
		{UUID: "bbb", Product: "Server B"},
		{UUID: "aaa", Product: "Server A duplicate"},
	}
	deduped := deduplicateRoots(roots)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 roots after dedup, got %d", len(deduped))
	}
	if deduped[0].Product != wantServerA {
		t.Errorf("expected first occurrence kept, got %q", deduped[0].Product)
	}
	if deduped[1].Product != "Server B" {
		t.Errorf("deduped[1].Product = %q, want %q", deduped[1].Product, "Server B")
	}
}

// TestDeduplicateRootsEmptyUUID verifies roots without UUIDs fall back to product
// plus BMC identity for deduplication.
//
// Why it matters: some BMC responses may omit UUID, and import still needs to
// collapse exact endpoint duplicates without losing distinct endpoints.
// Inputs: two Server A roots with the same FQDN and one Server B root with a
// different FQDN. Outputs: two roots with Server A and Server B preserved.
// Data choice: matching product plus FQDN drives duplicate removal, while the
// distinct FQDN proves another endpoint survives.
func TestDeduplicateRootsEmptyUUID(t *testing.T) {
	a := HpeManager{FQDN: "a.local"}
	b := HpeManager{FQDN: "b.local"}
	roots := []ServiceRoot{
		{Product: wantServerA, Oem: OemData{Hpe: &HpeOem{Manager: []HpeManager{a}}}},
		{Product: wantServerA, Oem: OemData{Hpe: &HpeOem{Manager: []HpeManager{a}}}},
		{Product: "Server B", Oem: OemData{Hpe: &HpeOem{Manager: []HpeManager{b}}}},
	}
	deduped := deduplicateRoots(roots)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 roots after dedup, got %d", len(deduped))
	}
	assertField(t, "deduped[0].Product", deduped[0].Product, wantServerA)
	assertField(t, "deduped[0].ManagerFQDN", deduped[0].ManagerFQDN(), "a.local")
	assertField(t, "deduped[1].Product", deduped[1].Product, "Server B")
	assertField(t, "deduped[1].ManagerFQDN", deduped[1].ManagerFQDN(), "b.local")
}

// TestDeduplicateRootsPreservesSharedUUIDDifferentBMC verifies a shared UUID is
// not enough to collapse roots when BMC endpoints differ.
//
// Why it matters: the dedup key intentionally includes BMC FQDN or hostname so
// separate physical endpoints that report the same UUID remain importable.
// Inputs: two roots with UUID "aaa" but different FQDN values. Outputs: both
// roots preserved in order.
// Data choice: identical UUIDs with different FQDNs isolate the anti-collapse
// behavior from product or ordering concerns.
func TestDeduplicateRootsPreservesSharedUUIDDifferentBMC(t *testing.T) {
	roots := []ServiceRoot{
		{UUID: "aaa", Product: "Server A", Oem: OemData{Hpe: &HpeOem{Manager: []HpeManager{{FQDN: "a.local"}}}}},
		{UUID: "aaa", Product: "Server B", Oem: OemData{Hpe: &HpeOem{Manager: []HpeManager{{FQDN: "b.local"}}}}},
	}

	deduped := deduplicateRoots(roots)

	if len(deduped) != 2 {
		t.Fatalf("deduplicated roots = %d, want 2", len(deduped))
	}
	assertField(t, "deduped[0].ManagerFQDN", deduped[0].ManagerFQDN(), "a.local")
	assertField(t, "deduped[1].ManagerFQDN", deduped[1].ManagerFQDN(), "b.local")
}

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "..", "..", "testdata", "fixtures", "redfish", "v1", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", name, err)
	}
	return data
}

func assertField(t *testing.T, name, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %q, want %q", name, got, want)
	}
}
