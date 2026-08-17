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
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

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
