/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package export

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// bayServer fakes Nautobot's device-bay endpoints. existing is the bay the
// list call returns (nil means none); every POST/PATCH body is recorded.
type bayServer struct {
	existing map[string]any
	posts    [][]byte
	patches  [][]byte
}

func (s *bayServer) handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !strings.Contains(r.URL.Path, "dcim/device-bays") {
		_, _ = w.Write([]byte(`{"count":0,"results":[]}`))
		return
	}
	body, _ := io.ReadAll(r.Body)
	switch r.Method {
	case http.MethodGet:
		results := []any{}
		if s.existing != nil {
			results = append(results, s.existing)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"count": len(results), "results": results})
	case http.MethodPost:
		s.posts = append(s.posts, body)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(deviceBayJSON(uuid.New(), nil))
	case http.MethodPatch:
		s.patches = append(s.patches, body)
		_ = json.NewEncoder(w).Encode(deviceBayJSON(uuid.New(), nil))
	}
}

// deviceBayJSON builds a device-bay API object, optionally with an installed device.
func deviceBayJSON(id uuid.UUID, installed *uuid.UUID) map[string]any {
	bay := map[string]any{"id": id.String(), "name": "Bay 1", "device": map[string]any{"id": uuid.New().String()}}
	if installed != nil {
		bay["installed_device"] = map[string]any{"id": installed.String()}
	}
	return bay
}

// bayExportInventory builds a chassis and a child in "Bay 1" and returns the
// Nautobot device-ID map for both.
func bayExportInventory() (*devicetypes.Inventory, map[string]uuid.UUID, uuid.UUID) {
	inv := devicetypes.NewInventory()
	chassis := &devicetypes.CaniDeviceType{ID: uuid.New(), Name: "chassis-15"}
	child := &devicetypes.CaniDeviceType{ID: uuid.New(), Name: "leaf-001", Parent: chassis.ID, ParentBay: "Bay 1"}
	inv.Devices[chassis.ID], inv.Devices[child.ID] = chassis, child
	childNID := uuid.New()
	return inv, map[string]uuid.UUID{"chassis-15": uuid.New(), "leaf-001": childNID}, childNID
}

// TestLoadDeviceBaysInstallsChild verifies an existing empty bay is patched
// with the child, and a missing bay is created first.
//
// Why it matters: this is how half-width switches end up inside their chassis
// in Nautobot instead of floating in the rack. Inputs: a chassis/child pair,
// with the bay present and then absent. Outputs: one PATCH naming the child,
// plus one POST only when the bay was absent.
func TestLoadDeviceBaysInstallsChild(t *testing.T) {
	for _, bayExists := range []bool{true, false} {
		srv := &bayServer{}
		if bayExists {
			srv.existing = deviceBayJSON(uuid.New(), nil)
		}
		e, done := newExporterWithServer(t, srv.handle)
		inv, ids, childNID := bayExportInventory()
		result := &LoadResult{}

		e.loadDeviceBays(t.Context(), inv, ids, result)
		done()

		wantPosts := map[bool]int{true: 0, false: 1}[bayExists]
		if len(srv.posts) != wantPosts || len(srv.patches) != 1 {
			t.Fatalf("bayExists=%v: posts=%d patches=%d errors=%v", bayExists, len(srv.posts), len(srv.patches), result.Errors)
		}
		if !strings.Contains(string(srv.patches[0]), childNID.String()) {
			t.Errorf("PATCH body %s does not install child %s", srv.patches[0], childNID)
		}
		if result.DeviceBaysInstalled != 1 {
			t.Errorf("DeviceBaysInstalled = %d, want 1", result.DeviceBaysInstalled)
		}
	}
}

// TestLoadDeviceBaysOccupied verifies a bay already holding the child is
// skipped and a bay holding another device is reported.
//
// Why it matters: re-exports must be idempotent, and silently replacing a
// different installed device would corrupt Nautobot.
func TestLoadDeviceBaysOccupied(t *testing.T) {
	inv, ids, childNID := bayExportInventory()
	other := uuid.New()
	for _, tc := range []struct {
		installed  uuid.UUID
		wantSkip   int
		wantErrors int
	}{{childNID, 1, 0}, {other, 0, 1}} {
		srv := &bayServer{existing: deviceBayJSON(uuid.New(), &tc.installed)}
		e, done := newExporterWithServer(t, srv.handle)
		result := &LoadResult{}
		e.loadDeviceBays(t.Context(), inv, ids, result)
		done()
		if len(srv.patches) != 0 || result.DeviceBaysSkipped != tc.wantSkip || len(result.Errors) != tc.wantErrors {
			t.Errorf("installed=%s: patches=%d skipped=%d errors=%v", tc.installed, len(srv.patches), result.DeviceBaysSkipped, result.Errors)
		}
	}
}

// TestLoadDeviceBaysSkipsUnexported verifies children whose parent or self
// has no Nautobot ID make no API calls.
//
// Why it matters: a failed device create is already reported; a second error
// from the bay phase would only add noise.
func TestLoadDeviceBaysSkipsUnexported(t *testing.T) {
	srv := &bayServer{}
	e, done := newExporterWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "dcim/device-bays") {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		srv.handle(w, r)
	})
	defer done()
	inv, ids, _ := bayExportInventory()
	delete(ids, "chassis-15")
	result := &LoadResult{}
	e.loadDeviceBays(t.Context(), inv, ids, result)
	if len(result.Errors) != 0 {
		t.Errorf("errors = %v, want none", result.Errors)
	}
}

func TestGetOrCreateDeviceBayRejectsHTTPFailures(t *testing.T) {
	for _, status := range []int{
		http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden,
		http.StatusNotFound, http.StatusTooManyRequests, http.StatusInternalServerError,
		http.StatusServiceUnavailable,
	} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			assertBayLookupDoesNotCreate(t, status, "application/json",
				`{"detail":"lookup failed"}`, fmt.Sprintf("status %d", status))
		})
	}
}

func TestGetOrCreateDeviceBayRejectsMalformedSuccess(t *testing.T) {
	for _, tc := range []struct {
		name, contentType, body, want string
	}{
		{"missing results", "application/json", `{}`, "missing results"},
		{"null results", "application/json", `{"count":0,"results":null}`, "missing results"},
		{"invalid JSON", "application/json", `{`, "device bay lookup"},
		{"HTML", "text/html", `<html>upstream failure</html>`, "status 200"},
		{"positive count", "application/json", `{"count":1,"results":[]}`, "empty results"},
		{"empty next page", "application/json", `{"count":1,"results":[],"next":"?offset=1"}`, "empty page"},
		{"invalid bay", "application/json", `{"count":1,"results":[{"name":"Bay 1"}]}`, "invalid bay"},
		{"zero bay ID", "application/json",
			fmt.Sprintf(`{"count":1,"results":[{"id":%q,"name":"Bay 1","device":{"id":%q}}]}`, uuid.Nil, uuid.New()), "invalid bay"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertBayLookupDoesNotCreate(t, http.StatusOK, tc.contentType, tc.body, tc.want)
		})
	}
}

func assertBayLookupDoesNotCreate(t *testing.T, status int, contentType, body, want string) {
	t.Helper()
	var posts atomic.Int32
	e, cleanup := newExporterWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			posts.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(deviceBayJSON(uuid.New(), nil))
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	})
	defer cleanup()
	_, err := e.getOrCreateDeviceBay(t.Context(), uuid.New(), "Bay 1")
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("lookup error=%v, want %q", err, want)
	}
	if got := posts.Load(); got != 0 {
		t.Fatalf("failed lookup caused %d create request(s)", got)
	}
}

func TestGetOrCreateDeviceBayPreservesCancellation(t *testing.T) {
	var requests atomic.Int32
	e, cleanup := newExporterWithServer(t, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	})
	defer cleanup()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := e.getOrCreateDeviceBay(ctx, uuid.New(), "Bay 1")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v, want context cancellation", err)
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("canceled lookup made %d request(s)", got)
	}
}

func TestGetOrCreateDeviceBayPreservesTransportError(t *testing.T) {
	e, cleanup := newExporterWithServer(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	})
	cleanup()
	_, err := e.getOrCreateDeviceBay(t.Context(), uuid.New(), "Bay 1")
	if err == nil || !strings.Contains(err.Error(), "device bay lookup") {
		t.Fatalf("transport error was not returned: %v", err)
	}
}

func TestLoadDeviceBaysReportsCreateFailures(t *testing.T) {
	bayID, parentID := uuid.New(), uuid.New()
	for _, tc := range []struct {
		name, contentType, body, want string
		status                        int
	}{
		{"HTTP failure", "application/json", `{"detail":"create failed"}`, "status 500", http.StatusInternalServerError},
		{"invalid JSON", "application/json", `{`, "device bay create", http.StatusCreated},
		{"HTML", "text/html", `<html>upstream failure</html>`, "status 201", http.StatusCreated},
		{"missing ID", "application/json",
			fmt.Sprintf(`{"name":"Bay 1","device":{"id":%q}}`, parentID), "invalid bay", http.StatusCreated},
		{"zero ID", "application/json",
			fmt.Sprintf(`{"id":%q,"name":"Bay 1","device":{"id":%q}}`, uuid.Nil, parentID), "invalid bay", http.StatusCreated},
		{"missing parent", "application/json",
			fmt.Sprintf(`{"id":%q,"name":"Bay 1"}`, bayID), "invalid bay", http.StatusCreated},
		{"missing name", "application/json",
			fmt.Sprintf(`{"id":%q,"device":{"id":%q}}`, bayID, parentID), "invalid bay", http.StatusCreated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertBayCreateFailure(t, tc.status, tc.contentType, tc.body, tc.want)
		})
	}
}

func assertBayCreateFailure(t *testing.T, status int, contentType, body, want string) {
	t.Helper()
	srv := &bayServer{}
	var posts atomic.Int32
	e, cleanup := newExporterWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			srv.handle(w, r)
			return
		}
		posts.Add(1)
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	})
	inventory, ids, _ := bayExportInventory()
	result := &LoadResult{}

	e.loadDeviceBays(t.Context(), inventory, ids, result)
	cleanup()

	assertDeviceBayErrorContext(t, result, "leaf-001", `"Bay 1"`, "chassis-15", "device bay create", want)
	if posts.Load() != 1 || len(srv.patches) != 0 {
		t.Fatalf("failed creation: posts=%d patches=%d, want one POST and no PATCH", posts.Load(), len(srv.patches))
	}
	if result.DeviceBayCounts != (DeviceBayCounts{}) {
		t.Fatalf("failed creation counted as success: %+v", result.DeviceBayCounts)
	}
}

func assertDeviceBayErrorContext(t *testing.T, result *LoadResult, want ...string) {
	t.Helper()
	if len(result.Errors) != 1 {
		t.Fatalf("errors=%v, want one contextual error", result.Errors)
	}
	for _, text := range want {
		if !strings.Contains(result.Errors[0], text) {
			t.Errorf("error=%q, want containing %q", result.Errors[0], text)
		}
	}
}
