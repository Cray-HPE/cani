package imprt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// newTestClient creates a ClientWithResponses pointed at the given httptest server.
func newTestClient(t *testing.T, serverURL string) *nautobotapi.ClientWithResponses {
	t.Helper()
	client, err := nautobotapi.NewClientWithResponses(serverURL)
	if err != nil {
		t.Fatalf("NewClientWithResponses: %v", err)
	}
	return client
}

// paginatedResponse builds a JSON paginated response with an optional next URL.
func paginatedResponse(count int, results interface{}, nextURL string) []byte {
	m := map[string]interface{}{
		"count":    count,
		"results":  results,
		"previous": nil,
	}
	if nextURL != "" {
		m["next"] = nextURL
	} else {
		m["next"] = nil
	}
	b, _ := json.Marshal(m)
	return b
}

// --- FetchLocations ---

func TestFetchLocations_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dcim/locations/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedResponse(1, []map[string]interface{}{
			{"name": "loc1"},
		}, ""))
	}))
	defer srv.Close()

	got, err := FetchLocations(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchLocations: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 location, got %d", len(got))
	}
}

func TestFetchLocations_Pagination(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		if page == 1 {
			next := fmt.Sprintf("http://%s/dcim/locations/?limit=100&offset=100", r.Host)
			w.Write(paginatedResponse(2, []map[string]interface{}{
				{"name": "loc1"},
			}, next))
		} else {
			w.Write(paginatedResponse(2, []map[string]interface{}{
				{"name": "loc2"},
			}, ""))
		}
	}))
	defer srv.Close()

	got, err := FetchLocations(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchLocations: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 locations, got %d", len(got))
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 API calls, got %d", calls)
	}
}

func TestFetchLocations_EmptyResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedResponse(0, []map[string]interface{}{}, ""))
	}))
	defer srv.Close()

	got, err := FetchLocations(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchLocations: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 locations, got %d", len(got))
	}
}

func TestFetchLocations_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := FetchLocations(context.Background(), newTestClient(t, srv.URL))
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// --- FetchRacks ---

func TestFetchRacks_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dcim/racks/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedResponse(1, []map[string]interface{}{
			{"name": "rack1"},
		}, ""))
	}))
	defer srv.Close()

	got, err := FetchRacks(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchRacks: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 rack, got %d", len(got))
	}
}

func TestFetchRacks_Pagination(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		if page == 1 {
			next := fmt.Sprintf("http://%s/dcim/racks/?limit=100&offset=100", r.Host)
			w.Write(paginatedResponse(2, []map[string]interface{}{
				{"name": "r1"},
			}, next))
		} else {
			w.Write(paginatedResponse(2, []map[string]interface{}{
				{"name": "r2"},
			}, ""))
		}
	}))
	defer srv.Close()

	got, err := FetchRacks(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchRacks: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 racks, got %d", len(got))
	}
}

func TestFetchRacks_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := FetchRacks(context.Background(), newTestClient(t, srv.URL))
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// --- FetchDevices ---

func TestFetchDevices_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dcim/devices/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedResponse(1, []map[string]interface{}{
			{"name": "dev1"},
		}, ""))
	}))
	defer srv.Close()

	got, err := FetchDevices(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchDevices: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 device, got %d", len(got))
	}
}

func TestFetchDevices_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	_, err := FetchDevices(context.Background(), newTestClient(t, srv.URL))
	if err == nil {
		t.Fatal("expected error for 502 response")
	}
}

// --- FetchDeviceTypes ---

func TestFetchDeviceTypes_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dcim/device-types/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedResponse(1, []map[string]interface{}{
			{"model": "dt1"},
		}, ""))
	}))
	defer srv.Close()

	got, err := FetchDeviceTypes(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchDeviceTypes: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 device type, got %d", len(got))
	}
}

func TestFetchDeviceTypes_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	_, err := FetchDeviceTypes(context.Background(), newTestClient(t, srv.URL))
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
}

// --- FetchInterfaces ---

func TestFetchInterfaces_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dcim/interfaces/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedResponse(1, []map[string]interface{}{
			{"name": "eth0"},
		}, ""))
	}))
	defer srv.Close()

	got, err := FetchInterfaces(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchInterfaces: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(got))
	}
}

func TestFetchInterfaces_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := FetchInterfaces(context.Background(), newTestClient(t, srv.URL))
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// --- FetchModules ---

func TestFetchModules_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dcim/modules/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedResponse(1, []map[string]interface{}{
			{"serial": "SN1"},
		}, ""))
	}))
	defer srv.Close()

	got, err := FetchModules(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchModules: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 module, got %d", len(got))
	}
}

func TestFetchModules_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := FetchModules(context.Background(), newTestClient(t, srv.URL))
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// --- FetchModuleBays ---

func TestFetchModuleBays_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dcim/module-bays/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedResponse(1, []map[string]interface{}{
			{"name": "bay1"},
		}, ""))
	}))
	defer srv.Close()

	got, err := FetchModuleBays(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchModuleBays: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 module bay, got %d", len(got))
	}
}

func TestFetchModuleBays_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := FetchModuleBays(context.Background(), newTestClient(t, srv.URL))
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// --- FetchCables ---

func TestFetchCables_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dcim/cables/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedResponse(1, []map[string]interface{}{
			{"label": "cable1"},
		}, ""))
	}))
	defer srv.Close()

	got, err := FetchCables(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchCables: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 cable, got %d", len(got))
	}
}

func TestFetchCables_Pagination(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		if page == 1 {
			next := fmt.Sprintf("http://%s/dcim/cables/?limit=100&offset=100", r.Host)
			w.Write(paginatedResponse(3, []map[string]interface{}{
				{"label": "c1"}, {"label": "c2"},
			}, next))
		} else {
			w.Write(paginatedResponse(3, []map[string]interface{}{
				{"label": "c3"},
			}, ""))
		}
	}))
	defer srv.Close()

	got, err := FetchCables(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchCables: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 cables, got %d", len(got))
	}
}

func TestFetchCables_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := FetchCables(context.Background(), newTestClient(t, srv.URL))
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// --- FetchInventoryItems ---

func TestFetchInventoryItems_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dcim/inventory-items/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedResponse(1, []map[string]interface{}{
			{"name": "fru1"},
		}, ""))
	}))
	defer srv.Close()

	got, err := FetchInventoryItems(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchInventoryItems: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 inventory item, got %d", len(got))
	}
}

func TestFetchInventoryItems_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := FetchInventoryItems(context.Background(), newTestClient(t, srv.URL))
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// --- FetchStatuses ---

func TestFetchStatuses_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/extras/statuses/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedResponse(1, []map[string]interface{}{
			{"name": "Active", "content_types": []string{"dcim.device"}},
		}, ""))
	}))
	defer srv.Close()

	got, err := FetchStatuses(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchStatuses: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 status, got %d", len(got))
	}
	if got[0].Name != "Active" {
		t.Errorf("expected name Active, got %s", got[0].Name)
	}
}

func TestFetchStatuses_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := FetchStatuses(context.Background(), newTestClient(t, srv.URL))
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

// --- FetchRoles ---

func TestFetchRoles_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/extras/roles/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedResponse(1, []map[string]interface{}{
			{"name": "Server", "content_types": []string{"dcim.device"}},
		}, ""))
	}))
	defer srv.Close()

	got, err := FetchRoles(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchRoles: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 role, got %d", len(got))
	}
	if got[0].Name != "Server" {
		t.Errorf("expected name Server, got %s", got[0].Name)
	}
}

func TestFetchRoles_Pagination(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		if page == 1 {
			next := fmt.Sprintf("http://%s/extras/roles/?limit=100&offset=100", r.Host)
			w.Write(paginatedResponse(2, []map[string]interface{}{
				{"name": "Server", "content_types": []string{"dcim.device"}},
			}, next))
		} else {
			w.Write(paginatedResponse(2, []map[string]interface{}{
				{"name": "Switch", "content_types": []string{"dcim.device"}},
			}, ""))
		}
	}))
	defer srv.Close()

	got, err := FetchRoles(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchRoles: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(got))
	}
}

func TestFetchRoles_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := FetchRoles(context.Background(), newTestClient(t, srv.URL))
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// --- Cross-cutting: context cancellation ---

func TestFetch_CancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedResponse(0, []map[string]interface{}{}, ""))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client := newTestClient(t, srv.URL)

	_, err := FetchLocations(ctx, client)
	if err == nil {
		t.Error("FetchLocations: expected error with cancelled context")
	}

	_, err = FetchDevices(ctx, client)
	if err == nil {
		t.Error("FetchDevices: expected error with cancelled context")
	}
}

// --- Query parameter validation ---

func TestFetchLocations_SendsLimitAndOffset(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("limit") != "100" {
			t.Errorf("expected limit=100, got %s", q.Get("limit"))
		}
		if q.Get("offset") != "0" {
			t.Errorf("expected offset=0, got %s", q.Get("offset"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedResponse(0, []map[string]interface{}{}, ""))
	}))
	defer srv.Close()

	FetchLocations(context.Background(), newTestClient(t, srv.URL))
}
