package imprt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/spf13/cobra"
)

// fakeProvider implements the provider interface expected by Import.
type fakeProvider struct {
	client  *nautobotapi.ClientWithResponses
	ctx     context.Context
	cleared bool
	data    RawData
}

func (f *fakeProvider) ClearRawData()                                      { f.cleared = true }
func (f *fakeProvider) SetRawData(d RawData)                               { f.data = d }
func (f *fakeProvider) GetClient() *nautobotapi.ClientWithResponses        { return f.client }
func (f *fakeProvider) GetContext() context.Context                        { return f.ctx }

// setProviderGetterForTest installs a fake providerGetter and returns a
// cleanup function that restores the original.
func setProviderGetterForTest(t *testing.T, fp *fakeProvider) {
	t.Helper()
	old := providerGetter
	t.Cleanup(func() { providerGetter = old })
	SetProviderGetter(func() interface {
		ClearRawData()
		SetRawData(RawData)
		GetClient() *nautobotapi.ClientWithResponses
		GetContext() context.Context
	} {
		return fp
	})
}

// --- SetProviderGetter / GetProvider ---

func TestSetProviderGetter(t *testing.T) {
	old := providerGetter
	defer func() { providerGetter = old }()

	called := false
	SetProviderGetter(func() interface {
		ClearRawData()
		SetRawData(RawData)
		GetClient() *nautobotapi.ClientWithResponses
		GetContext() context.Context
	} {
		called = true
		return &fakeProvider{ctx: context.Background()}
	})

	GetProvider()
	if !called {
		t.Error("expected getter to be called")
	}
}

func TestGetProvider_PanicsWhenNotSet(t *testing.T) {
	old := providerGetter
	defer func() { providerGetter = old }()
	providerGetter = nil

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when providerGetter is nil")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "providerGetter not set") {
			t.Errorf("unexpected panic value: %v", r)
		}
	}()

	GetProvider()
}

// --- Import ---

// allEndpointsServer returns an httptest.Server that responds to every
// Nautobot list endpoint used by Import with valid paginated JSON.
func allEndpointsServer(t *testing.T) *httptest.Server {
	t.Helper()

	empty := func() []byte {
		b, _ := json.Marshal(map[string]interface{}{
			"count": 0, "next": nil, "previous": nil, "results": []interface{}{},
		})
		return b
	}

	withResults := func(results interface{}) []byte {
		b, _ := json.Marshal(map[string]interface{}{
			"count": 1, "next": nil, "previous": nil, "results": results,
		})
		return b
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/dcim/locations/":
			w.Write(withResults([]map[string]interface{}{{"name": "loc1"}}))
		case "/dcim/racks/":
			w.Write(withResults([]map[string]interface{}{{"name": "rack1"}}))
		case "/dcim/devices/":
			w.Write(withResults([]map[string]interface{}{{"name": "dev1"}}))
		case "/dcim/device-types/":
			w.Write(withResults([]map[string]interface{}{{"model": "dt1"}}))
		case "/dcim/interfaces/":
			w.Write(withResults([]map[string]interface{}{{"name": "eth0"}}))
		case "/dcim/modules/":
			w.Write(withResults([]map[string]interface{}{{"serial": "SN1"}}))
		case "/dcim/module-bays/":
			w.Write(withResults([]map[string]interface{}{{"name": "bay1"}}))
		case "/dcim/cables/":
			w.Write(withResults([]map[string]interface{}{{"label": "cable1"}}))
		case "/dcim/inventory-items/":
			w.Write(withResults([]map[string]interface{}{{"name": "fru1"}}))
		case "/extras/statuses/":
			w.Write(withResults([]map[string]interface{}{
				{"name": "Active", "content_types": []string{"dcim.device"}},
			}))
		case "/extras/roles/":
			w.Write(withResults([]map[string]interface{}{
				{"name": "Server", "content_types": []string{"dcim.device"}},
			}))
		default:
			w.Write(empty())
		}
	}))
}

func TestImport_Success(t *testing.T) {
	srv := allEndpointsServer(t)
	defer srv.Close()

	client, err := nautobotapi.NewClientWithResponses(srv.URL)
	if err != nil {
		t.Fatalf("NewClientWithResponses: %v", err)
	}

	fp := &fakeProvider{
		client: client,
		ctx:    context.Background(),
	}
	setProviderGetterForTest(t, fp)

	cmd := &cobra.Command{}
	if err := Import(cmd, nil, nil); err != nil {
		t.Fatalf("Import: %v", err)
	}

	if !fp.cleared {
		t.Error("expected ClearRawData to be called")
	}
	if len(fp.data.Locations) != 1 {
		t.Errorf("expected 1 location, got %d", len(fp.data.Locations))
	}
	if len(fp.data.Racks) != 1 {
		t.Errorf("expected 1 rack, got %d", len(fp.data.Racks))
	}
	if len(fp.data.Devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(fp.data.Devices))
	}
	if len(fp.data.DeviceTypes) != 1 {
		t.Errorf("expected 1 device type, got %d", len(fp.data.DeviceTypes))
	}
	if len(fp.data.Interfaces) != 1 {
		t.Errorf("expected 1 interface, got %d", len(fp.data.Interfaces))
	}
	if len(fp.data.Modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(fp.data.Modules))
	}
	if len(fp.data.ModuleBays) != 1 {
		t.Errorf("expected 1 module bay, got %d", len(fp.data.ModuleBays))
	}
	if len(fp.data.Cables) != 1 {
		t.Errorf("expected 1 cable, got %d", len(fp.data.Cables))
	}
	if len(fp.data.InventoryItems) != 1 {
		t.Errorf("expected 1 inventory item, got %d", len(fp.data.InventoryItems))
	}
	if len(fp.data.Statuses) != 1 {
		t.Errorf("expected 1 status, got %d", len(fp.data.Statuses))
	}
	if len(fp.data.Roles) != 1 {
		t.Errorf("expected 1 role, got %d", len(fp.data.Roles))
	}
}

func TestImport_PropagatesFirstFetchError(t *testing.T) {
	// Server that fails on the very first endpoint (locations).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client, err := nautobotapi.NewClientWithResponses(srv.URL)
	if err != nil {
		t.Fatalf("NewClientWithResponses: %v", err)
	}

	fp := &fakeProvider{
		client: client,
		ctx:    context.Background(),
	}
	setProviderGetterForTest(t, fp)

	cmd := &cobra.Command{}
	err = Import(cmd, nil, nil)
	if err == nil {
		t.Fatal("expected error when server returns 500")
	}
	if !strings.Contains(err.Error(), "fetching locations") {
		t.Errorf("expected 'fetching locations' in error, got: %s", err.Error())
	}
}

func TestImport_PropagatesLaterFetchError(t *testing.T) {
	// Server that succeeds for locations/racks/devices but fails on device-types.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ok := func(results interface{}) {
			b, _ := json.Marshal(map[string]interface{}{
				"count": 1, "next": nil, "previous": nil, "results": results,
			})
			w.Write(b)
		}
		switch r.URL.Path {
		case "/dcim/locations/":
			ok([]map[string]interface{}{{"name": "loc1"}})
		case "/dcim/racks/":
			ok([]map[string]interface{}{{"name": "rack1"}})
		case "/dcim/devices/":
			ok([]map[string]interface{}{{"name": "dev1"}})
		case "/dcim/device-types/":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	client, err := nautobotapi.NewClientWithResponses(srv.URL)
	if err != nil {
		t.Fatalf("NewClientWithResponses: %v", err)
	}

	fp := &fakeProvider{
		client: client,
		ctx:    context.Background(),
	}
	setProviderGetterForTest(t, fp)

	cmd := &cobra.Command{}
	err = Import(cmd, nil, nil)
	if err == nil {
		t.Fatal("expected error when device-types fetch fails")
	}
	if !strings.Contains(err.Error(), "fetching device types") {
		t.Errorf("expected 'fetching device types' in error, got: %s", err.Error())
	}
}

func TestImport_CancelledContext(t *testing.T) {
	srv := allEndpointsServer(t)
	defer srv.Close()

	client, err := nautobotapi.NewClientWithResponses(srv.URL)
	if err != nil {
		t.Fatalf("NewClientWithResponses: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fp := &fakeProvider{
		client: client,
		ctx:    ctx,
	}
	setProviderGetterForTest(t, fp)

	cmd := &cobra.Command{}
	err = Import(cmd, nil, nil)
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
}
