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
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// GetDeviceType / GetStatus / GetRole / GetDeviceByName — API-fetch paths.
// These complement the existing cache-hit tests by exercising the live lookup,
// the non-200 error branch, and the not-found branch with create flags off
// (the Exporter default), which is how a strict export resolves references.
// -----------------------------------------------------------------------------

// TestGetDeviceType_FetchesAndCaches verifies GetDeviceType resolves a device
// type by model from the API on a cache miss and stores the result for reuse.
//
// Why it matters: every exported device references a device type by slug;
// resolving and caching it once keeps a large export from re-querying the same
// model for every device.
// Inputs: model slug "hpe-dl380" with the server returning one matching type.
// Outputs: a *CachedItem carrying the type UUID, with the slug present in the
// deviceTypes cache afterwards. Data choice: the first (slug) lookup hits, so the
// local-library fallback is intentionally not triggered, isolating the primary
// resolve-and-cache path.
func TestGetDeviceType_FetchesAndCaches(t *testing.T) {
	dtID := uuid.New()
	var calls int
	body := fmt.Sprintf(`{"count":1,"results":[{"id":%q,"model":"hpe-dl380","display":"HPE DL380"}]}`, dtID.String())
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	item, err := e.Cache.GetDeviceType("hpe-dl380")
	if err != nil {
		t.Fatalf("GetDeviceType() error = %v", err)
	}
	if item == nil || item.ID != dtID {
		t.Fatalf("expected device type %s, got %+v", dtID, item)
	}

	e.Cache.deviceTypesMu.RLock()
	cached, ok := e.Cache.deviceTypes["hpe-dl380"]
	e.Cache.deviceTypesMu.RUnlock()
	if !ok || cached.ID != dtID {
		t.Errorf("expected hpe-dl380 cached after lookup, got %+v (ok=%v)", cached, ok)
	}
}

// TestGetDeviceType_NonOKReturnsError verifies a non-200 device-type list
// response is surfaced as an error.
//
// Why it matters: a server fault must abort resolution rather than be mistaken
// for "type absent", which could otherwise trigger an unwanted auto-create.
// Inputs: a slug with the server returning 500. Outputs: a non-nil error.
// Data choice: a 500 with a detail body models a transient Nautobot failure.
func TestGetDeviceType_NonOKReturnsError(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{"detail":"boom"}`))
	defer cleanup()

	if _, err := e.Cache.GetDeviceType("hpe-dl380"); err == nil {
		t.Fatal("expected an error when the device-type list responds with 500")
	}
}

// TestGetDeviceType_NotFoundWithoutCreateReturnsError verifies GetDeviceType
// errors when the model is absent and auto-creation is disabled.
//
// Why it matters: by default cani must not invent device types; an unknown model
// is an operator/library error that should stop the export.
// Inputs: a slug absent from both Nautobot (empty list) and the local library,
// with createDeviceTypes=false (default). Outputs: a non-nil error. Data choice:
// the fabricated slug "ghost-device-type" guarantees the local-library fallback
// also misses, so the not-found error path is reached deterministically.
func TestGetDeviceType_NotFoundWithoutCreateReturnsError(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, emptyListJSON))
	defer cleanup()

	if _, err := e.Cache.GetDeviceType("ghost-device-type"); err == nil {
		t.Fatal("expected an error when the device type is missing and create is disabled")
	}
}

// TestGetStatus_FetchesAndCaches verifies GetStatus resolves a status by name
// from the API on a cache miss and stores the result for reuse.
//
// Why it matters: device, rack, prefix and IP records all reference a status;
// caching the resolved status avoids repeated lookups across a single export.
// Inputs: name "Active" with the server returning one match and create flags off.
// Outputs: a *CachedItem carrying the status UUID, with "Active" present in the
// statuses cache. Data choice: createStatuses is off so the content-type
// reconciliation branch is skipped, isolating the plain resolve-and-cache path.
func TestGetStatus_FetchesAndCaches(t *testing.T) {
	stID := uuid.New()
	var calls int
	body := fmt.Sprintf(`{"count":1,"results":[%s]}`, refObjectJSON(stID, "Active"))
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	item, err := e.Cache.GetStatus("Active")
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if item == nil || item.ID != stID {
		t.Fatalf("expected status %s, got %+v", stID, item)
	}

	e.Cache.statusesMu.RLock()
	cached, ok := e.Cache.statuses["Active"]
	e.Cache.statusesMu.RUnlock()
	if !ok || cached.ID != stID {
		t.Errorf("expected Active cached after lookup, got %+v (ok=%v)", cached, ok)
	}
}

// TestGetStatus_NonOKReturnsError verifies a non-200 status list response is
// surfaced as an error.
//
// Why it matters: a server fault must abort resolution rather than be treated as
// a missing status that triggers auto-create.
// Inputs: a name with the server returning 503. Outputs: a non-nil error.
// Data choice: a 503 models Nautobot being temporarily unavailable.
func TestGetStatus_NonOKReturnsError(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusServiceUnavailable, `{"detail":"down"}`))
	defer cleanup()

	if _, err := e.Cache.GetStatus("Active"); err == nil {
		t.Fatal("expected an error when the status list responds with 503")
	}
}

// TestGetStatus_NotFoundWithoutCreateReturnsError verifies GetStatus errors when
// the status is absent and auto-creation is disabled.
//
// Why it matters: by default cani must resolve against pre-defined statuses; a
// missing required status is an operator error that should stop the export.
// Inputs: a name with an empty result list and createStatuses=false (default).
// Outputs: a non-nil error. Data choice: name "Ghost" against an empty list
// models a required status the operator forgot to define.
func TestGetStatus_NotFoundWithoutCreateReturnsError(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, emptyListJSON))
	defer cleanup()

	if _, err := e.Cache.GetStatus("Ghost"); err == nil {
		t.Fatal("expected an error when the status is missing and create is disabled")
	}
}

// TestGetRole_FetchesAndCaches verifies GetRole resolves a role by name from the
// API on a cache miss and stores the result for reuse.
//
// Why it matters: every device carries a role; caching the resolved role keeps a
// large export from re-querying the same role for every device.
// Inputs: name "Compute" with the server returning one match and create flags
// off. Outputs: a *CachedItem carrying the role UUID, with "Compute" present in
// the roles cache. Data choice: createRoles is off so the content-type
// reconciliation branch is skipped, isolating the plain resolve-and-cache path.
func TestGetRole_FetchesAndCaches(t *testing.T) {
	roleID := uuid.New()
	var calls int
	body := fmt.Sprintf(`{"count":1,"results":[%s]}`, refObjectJSON(roleID, "Compute"))
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	item, err := e.Cache.GetRole("Compute")
	if err != nil {
		t.Fatalf("GetRole() error = %v", err)
	}
	if item == nil || item.ID != roleID {
		t.Fatalf("expected role %s, got %+v", roleID, item)
	}

	e.Cache.rolesMu.RLock()
	cached, ok := e.Cache.roles["Compute"]
	e.Cache.rolesMu.RUnlock()
	if !ok || cached.ID != roleID {
		t.Errorf("expected Compute cached after lookup, got %+v (ok=%v)", cached, ok)
	}
}

// TestGetRole_NonOKReturnsError verifies a non-200 role list response is surfaced
// as an error.
//
// Why it matters: a server fault must abort resolution rather than be treated as
// a missing role that triggers auto-create.
// Inputs: a name with the server returning 500. Outputs: a non-nil error.
// Data choice: a 500 with a detail body models a transient Nautobot failure.
func TestGetRole_NonOKReturnsError(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{"detail":"boom"}`))
	defer cleanup()

	if _, err := e.Cache.GetRole("Compute"); err == nil {
		t.Fatal("expected an error when the role list responds with 500")
	}
}

// TestGetRole_NotFoundWithoutCreateReturnsError verifies GetRole errors when the
// role is absent and auto-creation is disabled.
//
// Why it matters: by default cani must resolve against pre-defined roles; a
// missing required role is an operator error that should stop the export.
// Inputs: a name with an empty result list and createRoles=false (default).
// Outputs: a non-nil error. Data choice: name "Ghost" against an empty list
// models a required role the operator forgot to define.
func TestGetRole_NotFoundWithoutCreateReturnsError(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, emptyListJSON))
	defer cleanup()

	if _, err := e.Cache.GetRole("Ghost"); err == nil {
		t.Fatal("expected an error when the role is missing and create is disabled")
	}
}

// TestGetDeviceByName_FetchesAndCaches verifies GetDeviceByName resolves a device
// by name from the API on a cache miss and stores the result for reuse.
//
// Why it matters: cable and relationship resolution look devices up by name;
// caching the hit avoids repeated lookups while wiring an inventory.
// Inputs: name "compute-001" with the server returning one match. Outputs: a
// *CachedItem carrying the device UUID, with the name present in the devices
// cache. Data choice: a single matching result is the canonical hit and the test
// inspects the cache directly to confirm the write.
func TestGetDeviceByName_FetchesAndCaches(t *testing.T) {
	devID := uuid.New()
	var calls int
	body := fmt.Sprintf(`{"count":1,"results":[%s]}`, refObjectJSON(devID, "compute-001"))
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	item, err := e.Cache.GetDeviceByName("compute-001")
	if err != nil {
		t.Fatalf("GetDeviceByName() error = %v", err)
	}
	if item == nil || item.ID != devID {
		t.Fatalf("expected device %s, got %+v", devID, item)
	}

	e.Cache.devicesMu.RLock()
	cached, ok := e.Cache.devices["compute-001"]
	e.Cache.devicesMu.RUnlock()
	if !ok || cached.ID != devID {
		t.Errorf("expected compute-001 cached after lookup, got %+v (ok=%v)", cached, ok)
	}
}

// TestGetDeviceByName_NonOKReturnsError verifies a non-200 device list response
// is surfaced as an error.
//
// Why it matters: a server fault must abort resolution rather than be silently
// treated as "device absent", which could corrupt cable wiring decisions.
// Inputs: a name with the server returning 500. Outputs: a non-nil error.
// Data choice: a 500 with a detail body models a transient Nautobot failure.
func TestGetDeviceByName_NonOKReturnsError(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{"detail":"boom"}`))
	defer cleanup()

	if _, err := e.Cache.GetDeviceByName("compute-001"); err == nil {
		t.Fatal("expected an error when the device list responds with 500")
	}
}

// TestGetDeviceByName_NotFoundReturnsNilNil verifies GetDeviceByName returns a
// nil item and nil error when no device matches.
//
// Why it matters: unlike required references, a missing device is a normal
// condition (the device may not exist remotely yet), so the lookup signals
// absence without raising an error that would abort the export.
// Inputs: a name with an empty result list. Outputs: a nil *CachedItem and nil
// error. Data choice: an empty list for an unknown name models the not-yet-
// created device case the contract explicitly tolerates.
func TestGetDeviceByName_NotFoundReturnsNilNil(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, emptyListJSON))
	defer cleanup()

	item, err := e.Cache.GetDeviceByName("nope")
	if err != nil {
		t.Fatalf("GetDeviceByName() unexpected error = %v", err)
	}
	if item != nil {
		t.Errorf("expected nil item for a missing device, got %+v", item)
	}
}

// reconcileServer answers a resolver lookup when auto-create reconciliation is
// enabled: the GET list returns one matching item carrying no content_types (so
// every required type is "missing" and a PATCH is forced), and the PATCH returns
// the updated item. patchCalls counts only the PATCH requests.
func reconcileServer(id uuid.UUID, name string, patchCalls *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch {
			*patchCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, patchedRefJSON(id, name))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{"count":1,"results":[%s]}`, refObjectJSON(id, name))
	}
}

// TestGetStatus_ReconcilesMissingContentTypes verifies that, with createStatuses
// enabled, GetStatus patches an existing status to add the content types it is
// missing and returns the reconciled item.
//
// Why it matters: a shared status like "Active" must cover every object type the
// export touches (devices, modules, racks, prefixes, IPs, VLANs); if GetStatus
// skipped the reconciliation it would resolve a status that some objects could
// not legally reference, failing the export downstream.
// Inputs: name "Active" with createStatuses=true, a server returning the status
// with no content types (forcing all required types to be missing) and accepting
// the PATCH. Outputs: a reconciled *CachedItem and exactly one PATCH. Data
// choice: an empty content_types list is the strongest trigger because it makes
// every required type missing, guaranteeing the update branch runs.
func TestGetStatus_ReconcilesMissingContentTypes(t *testing.T) {
	stID := uuid.New()
	var patchCalls int
	e, cleanup := newExporterWithServer(t, reconcileServer(stID, "Active", &patchCalls))
	defer cleanup()
	e.Cache.createStatuses = true

	item, err := e.Cache.GetStatus("Active")
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if item == nil || item.ID != stID {
		t.Fatalf("expected reconciled status %s, got %+v", stID, item)
	}
	if patchCalls != 1 {
		t.Errorf("expected exactly one PATCH to add missing content types, got %d", patchCalls)
	}
}

// TestGetRole_ReconcilesMissingContentTypes verifies that, with createRoles
// enabled, GetRole patches an existing role to add the content types it is
// missing and returns the reconciled item.
//
// Why it matters: a device role must be associated with the dcim.device (and
// IPAM) content types before objects can reference it; skipping reconciliation
// would resolve a role that cannot be assigned, breaking the export.
// Inputs: name "Compute" with createRoles=true, a server returning the role with
// no content types (forcing all required types to be missing) and accepting the
// PATCH. Outputs: a reconciled *CachedItem and exactly one PATCH. Data choice: an
// empty content_types list makes every required role type missing, guaranteeing
// the update branch executes.
func TestGetRole_ReconcilesMissingContentTypes(t *testing.T) {
	roleID := uuid.New()
	var patchCalls int
	e, cleanup := newExporterWithServer(t, reconcileServer(roleID, "Compute", &patchCalls))
	defer cleanup()
	e.Cache.createRoles = true

	item, err := e.Cache.GetRole("Compute")
	if err != nil {
		t.Fatalf("GetRole() error = %v", err)
	}
	if item == nil || item.ID != roleID {
		t.Fatalf("expected reconciled role %s, got %+v", roleID, item)
	}
	if patchCalls != 1 {
		t.Errorf("expected exactly one PATCH to add missing content types, got %d", patchCalls)
	}
}
