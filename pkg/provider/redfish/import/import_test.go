package import_

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	fixtureFile = "redfish-root.json"
	parseErrFmt = "ParseServiceRoots() error: %v"
	wantProduct = "ProLiant DL325 Gen11"
	wantServerA = "Server A"
)

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

func TestParseServiceRootsArray(t *testing.T) {
	single := loadFixture(t, fixtureFile)
	arrayJSON := []byte("[" + string(single) + "," + string(single) + "]")
	roots, err := ParseServiceRoots(arrayJSON)
	if err != nil {
		t.Fatalf(parseErrFmt, err)
	}
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
	for i, r := range roots {
		if r.Product != wantProduct {
			t.Errorf("root[%d].Product = %q, want %q", i, r.Product, wantProduct)
		}
	}
}

func TestParseServiceRootsInvalid(t *testing.T) {
	_, err := ParseServiceRoots([]byte(`{"not": "a service root"}`))
	if err == nil {
		t.Error("expected error for non-ServiceRoot JSON, got nil")
	}
}

func TestParseServiceRootsBadJSON(t *testing.T) {
	_, err := ParseServiceRoots([]byte(`{bad json`))
	if err == nil {
		t.Error("expected error for bad JSON, got nil")
	}
}

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
	if root.ManagerFQDN() == "" {
		t.Error("ManagerFQDN() is empty, expected non-empty")
	}
	if root.ManagerHostName() == "" {
		t.Error("ManagerHostName() is empty, expected non-empty")
	}
}

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
}

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
