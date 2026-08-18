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
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	providertransform "github.com/Cray-HPE/cani/pkg/provider/nautobot/transform"
	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// createInterface
// -----------------------------------------------------------------------------

// TestCreateInterface_DryRunIncrementsWithoutHTTP verifies that in dry-run mode
// createInterface increments IfacesCreated and issues no HTTP request.
//
// Why it matters: interfaces are a prerequisite phase for cables; a dry-run
// must report how many interfaces would be created without writing to Nautobot.
// Inputs: e.Options.DryRun=true with interfaceSpec{Name:"eth0",Type:"1000base-t"}.
// Outputs: nil error, IfacesCreated==1, calls==0.
// Data choice: a minimal interfaceSpec suffices because the dry-run branch
// returns before any request is built.
func TestCreateInterface_DryRunIncrementsWithoutHTTP(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()
	e.Options.DryRun = true

	result := &LoadResult{}
	iface := interfaceSpec{Name: "eth0", Type: "1000base-t"}
	if err := e.createInterface(context.Background(), uuid.New(), iface, result); err != nil {
		t.Fatalf("createInterface() error = %v", err)
	}
	if result.IfacesCreated != 1 {
		t.Errorf("IfacesCreated = %d, want 1", result.IfacesCreated)
	}
	if calls != 0 {
		t.Errorf("expected no HTTP calls in dry-run, got %d", calls)
	}
}

// TestCreateInterface_CreatesAndCaches verifies that createInterface POSTs the
// interface, increments IfacesCreated, and caches the returned ID under
// (deviceID, name).
//
// Why it matters: the cable phase resolves interface IDs from this cache;
// caching on create avoids a redundant lookup and lets cables find
// freshly-created interfaces.
// Inputs: a deviceID and interfaceSpec with MgmtOnly=true; server returns 201
// with the interface id; "Active" status seeded. Outputs: IfacesCreated==1 and
// a cache entry for interfaceCacheKey(deviceID,"eth0") with the matching ID.
// Data choice: seedActiveStatus avoids a status round-trip; MgmtOnly=true
// exercises the optional flag while keeping the assert on create+cache.
func TestCreateInterface_CreatesAndCaches(t *testing.T) {
	ifaceID := uuid.New()
	deviceID := uuid.New()
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusCreated,
		fmt.Sprintf(`{"id":%q,"name":"eth0","display":"eth0"}`, ifaceID.String())))
	defer cleanup()
	seedActiveStatus(t, e)

	result := &LoadResult{}
	iface := interfaceSpec{Name: "eth0", Type: "1000base-t", MgmtOnly: true}
	if err := e.createInterface(context.Background(), deviceID, iface, result); err != nil {
		t.Fatalf("createInterface() error = %v", err)
	}
	if result.IfacesCreated != 1 {
		t.Errorf("IfacesCreated = %d, want 1", result.IfacesCreated)
	}

	e.Cache.interfacesMu.RLock()
	cached, ok := e.Cache.interfaces[interfaceCacheKey(deviceID, "eth0")]
	e.Cache.interfacesMu.RUnlock()
	if !ok || cached.ID != ifaceID {
		t.Errorf("expected created interface %s cached, got %+v (ok=%v)", ifaceID, cached, ok)
	}
}

// TestCreateInterface_ReturnsErrorOnNon201 verifies that an error is returned
// when the interface create POST responds with 400.
//
// Why it matters: if an interface fails to export, dependent cables would later
// fail to resolve it; surfacing the error stops the pipeline at the real cause.
// Inputs: a server returning 400 with "Active" status seeded. Outputs: a
// non-nil error.
// Data choice: seeding the status isolates the failure to the create POST's
// status code rather than status resolution.
func TestCreateInterface_ReturnsErrorOnNon201(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusBadRequest, `{"detail":"bad"}`))
	defer cleanup()
	seedActiveStatus(t, e)

	result := &LoadResult{}
	iface := interfaceSpec{Name: "eth0", Type: "1000base-t"}
	if err := e.createInterface(context.Background(), uuid.New(), iface, result); err == nil {
		t.Fatal("expected an error when interface create responds with 400")
	}
}

// TestCreateInterface_ReturnsErrorWhenStatusUnresolvable verifies that
// createInterface errors before issuing the interface POST when the required
// "Active" status cannot be resolved.
//
// Why it matters: every Nautobot interface needs a status reference; failing
// early avoids sending an invalid request and gives a clear cause.
// Inputs: no "Active" status seeded and a status lookup that returns an empty
// list. Outputs: a non-nil error.
// Data choice: an empty status list (count 0) is the precise condition that
// makes statusRef("Active") fail, exercising the pre-POST guard.
func TestCreateInterface_ReturnsErrorWhenStatusUnresolvable(t *testing.T) {
	// No "Active" status is seeded and the status lookup returns an empty
	// list, so statusRef cannot resolve a status and createInterface fails
	// before issuing the interface POST.
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{"count":0,"results":[]}`))
	defer cleanup()

	result := &LoadResult{}
	iface := interfaceSpec{Name: "eth0", Type: "1000base-t"}
	if err := e.createInterface(context.Background(), uuid.New(), iface, result); err == nil {
		t.Fatal("expected an error when the Active status cannot be resolved")
	}
}

// TestCreateInterface_RejectsUnresolvedRole verifies fallback creation fails
// before POSTing an interface when its requested role cannot be resolved.
func TestCreateInterface_RejectsUnresolvedRole(t *testing.T) {
	interfacePosts := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "dcim/interfaces") {
			interfacePosts++
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(emptyListJSON))
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)

	result := &LoadResult{}
	iface := interfaceSpec{Name: "eth0", Type: "1000base-t", Role: "GhostRole"}
	if err := e.createInterface(context.Background(), uuid.New(), iface, result); err == nil {
		t.Fatal("createInterface() error = nil, want unresolved role error")
	}
	if interfacePosts != 0 {
		t.Errorf("interface POSTs = %d, want 0", interfacePosts)
	}
}

// -----------------------------------------------------------------------------
// updateInterface
// -----------------------------------------------------------------------------

// TestUpdateInterface_DryRunReturnsNilWithoutHTTP verifies that in dry-run mode
// updateInterface returns nil and makes no HTTP call.
//
// Why it matters: re-exports may update existing interfaces; a dry-run must
// preview without mutating Nautobot.
// Inputs: e.Options.DryRun=true with a minimal interfaceSpec. Outputs: nil
// error and calls==0 (updateInterface intentionally bumps no counter).
// Data choice: a minimal spec suffices since the dry-run branch returns before
// the PATCH request is built.
func TestUpdateInterface_DryRunReturnsNilWithoutHTTP(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()
	e.Options.DryRun = true

	result := &LoadResult{}
	iface := interfaceSpec{Name: "eth0", Type: "1000base-t"}
	if err := e.updateInterface(context.Background(), uuid.New(), uuid.New(), iface, result); err != nil {
		t.Fatalf("updateInterface() error = %v", err)
	}
	if calls != 0 {
		t.Errorf("expected no HTTP calls in dry-run, got %d", calls)
	}
}

// TestUpdateInterface_UpdatesOn200 verifies that updateInterface issues exactly
// one PATCH and succeeds when Nautobot returns 200.
//
// Why it matters: keeping interface attributes (type, status, MAC) in sync on
// re-export depends on the PATCH happening exactly once.
// Inputs: an interfaceSpec with a MAC address; server returns 200. Outputs: nil
// error and calls==1.
// Data choice: supplying Mac="00:11:22:33:44:55" exercises the optional MAC
// field path while asserting a single PATCH.
func TestUpdateInterface_UpdatesOn200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{"id":"`+uuid.NewString()+`"}`))
	defer cleanup()
	seedActiveStatus(t, e)

	result := &LoadResult{}
	iface := interfaceSpec{Name: "eth0", Type: "1000base-t", Mac: "00:11:22:33:44:55"}
	if err := e.updateInterface(context.Background(), uuid.New(), uuid.New(), iface, result); err != nil {
		t.Fatalf("updateInterface() error = %v", err)
	}
	if calls != 1 {
		t.Errorf("expected exactly one PATCH call, got %d", calls)
	}
}

// TestUpdateInterface_ReturnsErrorOnNon200 verifies that updateInterface returns
// an error when the PATCH responds with 500.
//
// Why it matters: a failed update must surface so stale interface data in
// Nautobot is not silently accepted as correct.
// Inputs: a server returning 500 with "Active" status seeded. Outputs: a
// non-nil error.
// Data choice: 500 (instead of 200) drives the unexpected-status branch;
// seeding the status keeps the failure attributable to the PATCH.
func TestUpdateInterface_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{}`))
	defer cleanup()
	seedActiveStatus(t, e)

	result := &LoadResult{}
	iface := interfaceSpec{Name: "eth0", Type: "1000base-t"}
	if err := e.updateInterface(context.Background(), uuid.New(), uuid.New(), iface, result); err == nil {
		t.Fatal("expected an error when interface update responds with 500")
	}
}

// TestUpdateInterface_ClearsDescriptionWhenEmpty verifies that an empty local
// description is sent as "description":"" on the PATCH so the inventory value
// (authoritative on reconcile) clears any stale text in Nautobot.
//
// Why it matters: `description` uses `json:",omitempty"`, so only a non-nil
// pointer to "" reaches Nautobot; sending it unconditionally is what makes an
// emptied description round-trip instead of silently diverging.
// Inputs: an interfaceSpec with no Description; server returns 200. Outputs:
// a PATCH body containing "description":"".
func TestUpdateInterface_ClearsDescriptionWhenEmpty(t *testing.T) {
	var body string
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			b, _ := io.ReadAll(r.Body)
			body = string(b)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(emptyListJSON))
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)

	result := &LoadResult{}
	iface := interfaceSpec{Name: "eth0", Type: "1000base-t"} // Description intentionally empty
	if err := e.updateInterface(context.Background(), uuid.New(), uuid.New(), iface, result); err != nil {
		t.Fatalf("updateInterface() error = %v", err)
	}
	if !strings.Contains(body, `"description":""`) {
		t.Errorf("PATCH body missing empty description clear:\n%s", body)
	}
}

// TestInterfaceDescriptionImportExportRoundTrip verifies a description fetched
// from Nautobot survives transformation into the portable model and is sent
// back by interface reconciliation.
func TestInterfaceDescriptionImportExportRoundTrip(t *testing.T) {
	deviceID := uuid.New()
	interfaceID := uuid.New()
	deviceOpenAPIID := openapi_types.UUID(deviceID)
	interfaceOpenAPIID := openapi_types.UUID(interfaceID)
	deviceName := "switch-01"
	description := "ISL uplink to spine"
	interfaceType := nautobotapi.InterfaceTypeValue("100gbase-x-qsfp28")
	deviceRef := makeObjectRef(deviceID)

	devices, idMap := providertransform.MapDevices(
		[]nautobotapi.Device{{Id: &deviceOpenAPIID, Name: &deviceName}},
		nil, nil, nil,
		map[uuid.UUID][]nautobotapi.Interface{
			deviceID: {{
				Id:          &interfaceOpenAPIID,
				Name:        "1/1/49",
				Device:      deviceRef,
				Type:        nautobotapi.InterfaceType{Value: &interfaceType},
				Description: &description,
			}},
		},
		nil, nil,
	)
	device := devices[idMap[deviceID]]
	if device == nil {
		t.Fatal("MapDevices() did not return the imported device")
	}
	specs := getDeviceInterfaceSpecs(device)
	if len(specs) != 1 {
		t.Fatalf("getDeviceInterfaceSpecs() returned %d specs, want 1", len(specs))
	}

	var body string
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			payload, _ := io.ReadAll(r.Body)
			body = string(payload)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}
	exporter, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, exporter)

	if err := exporter.updateInterface(context.Background(), interfaceID, deviceID, specs[0], &LoadResult{}); err != nil {
		t.Fatalf("updateInterface() error = %v", err)
	}
	if !strings.Contains(body, `"description":"ISL uplink to spine"`) {
		t.Errorf("PATCH body missing imported description:\n%s", body)
	}
}

// TestUpdateInterface_ClearsRoleWhenEmpty verifies that an empty local role is
// sent as "role":null on the PATCH so the inventory value (authoritative on
// reconcile) clears any stale role FK in Nautobot.
//
// Why it matters: the `role` FK has no `omitempty`, so a nil pointer serializes
// as null; this pins the FK-clear contract that mirrors the description clear.
// Inputs: an interfaceSpec with no Role; server returns 200. Outputs: a PATCH
// body containing "role":null.
func TestUpdateInterface_ClearsRoleWhenEmpty(t *testing.T) {
	var body string
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			b, _ := io.ReadAll(r.Body)
			body = string(b)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(emptyListJSON))
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)

	result := &LoadResult{}
	iface := interfaceSpec{Name: "eth0", Type: "1000base-t"} // Role intentionally empty
	if err := e.updateInterface(context.Background(), uuid.New(), uuid.New(), iface, result); err != nil {
		t.Fatalf("updateInterface() error = %v", err)
	}
	if !strings.Contains(body, `"role":null`) {
		t.Errorf("PATCH body missing role null-clear:\n%s", body)
	}
}

// TestUpdateInterface_SerializesResolvedRoleAndDevice verifies both generated
// foreign-key fields are non-null references in the actual PATCH body.
func TestUpdateInterface_SerializesResolvedRoleAndDevice(t *testing.T) {
	deviceID := uuid.New()
	roleID := uuid.New()
	var body string
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			payload, _ := io.ReadAll(r.Body)
			body = string(payload)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)
	e.Cache.roles["UplinkInterface"] = &CachedItem{ID: roleID, Name: "UplinkInterface"}

	iface := interfaceSpec{Name: "eth0", Type: "1000base-t", Role: "UplinkInterface"}
	if err := e.updateInterface(context.Background(), uuid.New(), deviceID, iface, &LoadResult{}); err != nil {
		t.Fatalf("updateInterface() error = %v", err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("unmarshal PATCH body: %v", err)
	}
	if role := string(payload["role"]); role == "" || role == "null" || !strings.Contains(role, roleID.String()) {
		t.Errorf("PATCH role = %s, want reference to %s", role, roleID)
	}
	if device := string(payload["device"]); device == "" || device == "null" || !strings.Contains(device, deviceID.String()) {
		t.Errorf("PATCH device = %s, want reference to %s", device, deviceID)
	}
}

// TestUpdateInterface_RejectsUnresolvedRole verifies a non-empty role lookup
// failure aborts the update instead of serializing the role as null.
func TestUpdateInterface_RejectsUnresolvedRole(t *testing.T) {
	patchCalls := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			patchCalls++
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(emptyListJSON))
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)

	result := &LoadResult{}
	iface := interfaceSpec{Name: "eth0", Type: "1000base-t", Role: "GhostRole"}
	if err := e.updateInterface(context.Background(), uuid.New(), uuid.New(), iface, result); err == nil {
		t.Fatal("updateInterface() error = nil, want unresolved role error")
	}
	if patchCalls != 0 {
		t.Errorf("PATCH calls = %d, want 0", patchCalls)
	}
}
