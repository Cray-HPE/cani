package export

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// sentInterface is a minimal view of the JSON payload the exporter POSTs to
// Nautobot's /dcim/interfaces/ endpoint. Decoding into this plain struct lets
// the tests assert the on-the-wire shape independently of the generated
// client's union types.
type sentInterface struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Device struct {
		ID string `json:"id"`
	} `json:"device"`
	Status struct {
		ID string `json:"id"`
	} `json:"status"`
	MgmtOnly *bool `json:"mgmt_only"`
}

// capturedRequest records what the fake Nautobot server received.
type capturedRequest struct {
	method string
	path   string
	body   []byte
}

// newExporterWithServer wires an Exporter to an httptest server so the real
// HTTP request-building and response-parsing code paths run without a live
// Nautobot. The returned cleanup func must be deferred by the caller.
func newExporterWithServer(t *testing.T, handler http.HandlerFunc) (*Exporter, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	client, err := NewNautobotClient(srv.URL, "test-token")
	if err != nil {
		srv.Close()
		t.Fatalf("NewNautobotClient: %v", err)
	}
	e := &Exporter{
		Client:  client,
		Cache:   NewLookupCache(client),
		Options: &ExporterOpts{},
	}
	return e, srv.Close
}

// activeStatus builds an interface status reference with a fresh ID.
func activeStatus(t *testing.T) (nautobotapi.BulkWritableCableRequestStatus, uuid.UUID) {
	t.Helper()
	statusID := uuid.New()
	var union nautobotapi.BulkWritableCableRequestStatusId
	if err := union.FromBulkWritableCableRequestStatusId0(statusID); err != nil {
		t.Fatalf("build status union: %v", err)
	}
	return nautobotapi.BulkWritableCableRequestStatus{Id: &union}, statusID
}

// interfaceServer returns a handler that records the interface-create POST into
// cap and replies with the supplied created interfaces. Any other request
// (best-effort role/status lookups) receives an empty result list so it is a
// no-op.
func interfaceServer(rec *capturedRequest, created []map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "dcim/interfaces") {
			rec.method, rec.path = r.Method, r.URL.Path
			rec.body, _ = io.ReadAll(r.Body)
			body, _ := json.Marshal(created)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write(body)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":0,"results":[]}`))
	}
}

// decodeSentInterfaces unmarshals a captured request body into sentInterfaces.
func decodeSentInterfaces(t *testing.T, body []byte) []sentInterface {
	t.Helper()
	var sent []sentInterface
	if err := json.Unmarshal(body, &sent); err != nil {
		t.Fatalf("decode payload: %v\nbody: %s", err, body)
	}
	return sent
}

// indexByName keys a slice of sentInterfaces by interface name.
func indexByName(in []sentInterface) map[string]sentInterface {
	out := make(map[string]sentInterface, len(in))
	for _, s := range in {
		out[s.Name] = s
	}
	return out
}

// assertRef verifies the type, device and status references of one payload entry.
func assertRef(t *testing.T, got sentInterface, wantType string, devID, statusID uuid.UUID) {
	t.Helper()
	if got.Type != wantType {
		t.Errorf("%s type = %q, want %q", got.Name, got.Type, wantType)
	}
	if got.Device.ID != devID.String() {
		t.Errorf("%s device.id = %q, want %q", got.Name, got.Device.ID, devID)
	}
	if got.Status.ID != statusID.String() {
		t.Errorf("%s status.id = %q, want %q", got.Name, got.Status.ID, statusID)
	}
}

// assertCached verifies an interface was cached by device+name after creation.
func assertCached(t *testing.T, e *Exporter, devID uuid.UUID, name string, wantID uuid.UUID) {
	t.Helper()
	got, err := e.Cache.GetInterfaceByDeviceAndName(devID, name)
	if err != nil {
		t.Fatalf("GetInterfaceByDeviceAndName(%s): %v", name, err)
	}
	if got == nil || got.ID != wantID {
		t.Errorf("cache[%s] = %v, want ID %s", name, got, wantID)
	}
}

// assertMgmtOnly verifies the mgmt_only flag of a named payload entry.
func assertMgmtOnly(t *testing.T, byName map[string]sentInterface, name string, want bool) {
	t.Helper()
	got, ok := byName[name]
	if !ok {
		t.Fatalf("%s interface missing from payload", name)
	}
	have := got.MgmtOnly != nil && *got.MgmtOnly
	if have != want {
		t.Errorf("%s mgmt_only = %v, want %v", name, have, want)
	}
}

// TestSendInterfaceBatch_PayloadAndCache is a round-trip test for the bulk
// interface create path — the mechanism every export uses to populate device
// interfaces in Nautobot. It verifies three things that the export depends on
// but that no prior test exercised:
//
//  1. the JSON payload POSTed to Nautobot carries the correct device, name,
//     type and status for every interface in the batch;
//  2. the array response is parsed back into Interface objects; and
//  3. the created interfaces are cached by device+name so cable creation
//     (Phase 6) can resolve them.
func TestSendInterfaceBatch_PayloadAndCache(t *testing.T) {
	id0, id1 := uuid.New(), uuid.New()
	rec := &capturedRequest{}
	created := []map[string]string{
		{"id": id0.String(), "name": "iLO"},
		{"id": id1.String(), "name": "eth0"},
	}

	e, cleanup := newExporterWithServer(t, interfaceServer(rec, created))
	defer cleanup()

	devID := uuid.New()
	status, statusID := activeStatus(t)
	batch := []bulkInterfaceItem{
		{DeviceID: devID, DeviceName: "node1", Spec: interfaceSpec{Name: "iLO", Type: "1000base-t"}},
		{DeviceID: devID, DeviceName: "node1", Spec: interfaceSpec{Name: "eth0", Type: "1000base-t"}},
	}

	result, err := e.sendInterfaceBatch(context.Background(), batch, status)
	if err != nil {
		t.Fatalf("sendInterfaceBatch: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("parsed %d created interfaces, want 2", len(result))
	}

	// Request shape.
	if rec.method != http.MethodPost {
		t.Errorf("method = %s, want POST", rec.method)
	}
	if !strings.Contains(rec.path, "dcim/interfaces") {
		t.Errorf("path = %q, want it to contain dcim/interfaces", rec.path)
	}

	// Payload.
	byName := indexByName(decodeSentInterfaces(t, rec.body))
	if len(byName) != 2 {
		t.Fatalf("payload had %d interfaces, want 2", len(byName))
	}
	assertRef(t, byName["iLO"], "1000base-t", devID, statusID)
	assertRef(t, byName["eth0"], "1000base-t", devID, statusID)

	// Cache wiring used by cable creation.
	e.cacheCreatedInterfaces(batch, result)
	assertCached(t, e, devID, "iLO", id0)
	assertCached(t, e, devID, "eth0", id1)
}

// TestInterfaceExport_PreservesMgmtOnly guards a data-fidelity guarantee: an
// interface flagged management-only in the cani inventory (e.g. iLO/BMC) must
// be exported to Nautobot with mgmt_only=true. Nautobot models mgmt_only as a
// first-class field that is distinct from interface role, so dropping it
// silently changes the meaning of the device in the source of truth.
//
// The test drives the real export path: getDeviceInterfaceSpecs builds the
// specs from the device, and sendInterfaceBatch serializes them to the wire
// format which is captured and inspected.
func TestInterfaceExport_PreservesMgmtOnly(t *testing.T) {
	mgmt := true
	dev := &devicetypes.CaniDeviceType{
		Name: "node1",
		Type: "node",
		Interfaces: []devicetypes.InterfaceSpec{
			{Name: "iLO", Type: "1000base-t", MgmtOnly: &mgmt},
			{Name: "eth0", Type: "1000base-t"},
		},
	}

	specs := getDeviceInterfaceSpecs(dev)
	if len(specs) != 2 {
		t.Fatalf("getDeviceInterfaceSpecs returned %d specs, want 2", len(specs))
	}

	devID := uuid.New()
	created := make([]map[string]string, 0, len(specs))
	batch := make([]bulkInterfaceItem, 0, len(specs))
	for _, s := range specs {
		created = append(created, map[string]string{"id": uuid.New().String(), "name": s.Name})
		batch = append(batch, bulkInterfaceItem{DeviceID: devID, DeviceName: dev.Name, Spec: s})
	}

	rec := &capturedRequest{}
	e, cleanup := newExporterWithServer(t, interfaceServer(rec, created))
	defer cleanup()

	status, _ := activeStatus(t)
	if _, err := e.sendInterfaceBatch(context.Background(), batch, status); err != nil {
		t.Fatalf("sendInterfaceBatch: %v", err)
	}

	byName := indexByName(decodeSentInterfaces(t, rec.body))
	assertMgmtOnly(t, byName, "iLO", true)   // management-only flag must survive export
	assertMgmtOnly(t, byName, "eth0", false) // data port must not be management-only
}
