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
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// seedDeviceRefs pre-populates the cache with the device type, location and
// role the mapper resolves (plus the "Active" status). Keys match the device
// built by newMappableDevice, so the mapper resolves every reference from the
// cache and issues no HTTP of its own — each CRUD helper then makes exactly one
// request.
func seedDeviceRefs(t *testing.T, e *Exporter) {
	t.Helper()
	e.Cache.deviceTypesMu.Lock()
	e.Cache.deviceTypes["hpe-dl380"] = &CachedItem{ID: uuid.New(), Name: "ProLiant DL380"}
	e.Cache.deviceTypesMu.Unlock()
	e.Cache.locationsMu.Lock()
	e.Cache.locations["DC1"] = &CachedItem{ID: uuid.New(), Name: "DC1"}
	e.Cache.locationsMu.Unlock()
	e.Cache.rolesMu.Lock()
	e.Cache.roles["Compute"] = &CachedItem{ID: uuid.New(), Name: "Compute"}
	e.Cache.rolesMu.Unlock()
	seedActiveStatus(t, e)
}

// newMappableDevice builds a device whose references match seedDeviceRefs.
func newMappableDevice(name string) *devicetypes.CaniDeviceType {
	return &devicetypes.CaniDeviceType{
		Name:       name,
		Slug:       "hpe-dl380",
		ObjectMeta: devicetypes.ObjectMeta{Status: "Active", Role: "Compute"},
	}
}

// newCrudMapper wires a mapper to the exporter cache with defaults that match
// the seeded references.
func newCrudMapper(e *Exporter) *DeviceMapper {
	return NewDeviceMapper(e.Cache, &MapperOpts{
		DefaultLocation: "DC1",
		DefaultStatus:   "Active",
		DefaultRole:     "Compute",
	})
}

// containsName reports whether names contains want.
func containsName(names []string, want string) bool {
	for _, n := range names {
		if n == want {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// createDevice
// -----------------------------------------------------------------------------

// TestCreateDevice_DryRunRecordsWithoutHTTP verifies that with DryRun enabled
// createDevice records the device name (with the dry-run suffix) in
// result.Created and issues no HTTP request.
//
// Why it matters: a dry run must preview the export of a cani device without
// mutating Nautobot, so operators can audit changes before committing them.
// Inputs: a mappable device, DryRun=true. Outputs: result.Created gains
// "compute-001"+suffixDryRun and the fake server's call counter stays at 0.
// Data choice: seedDeviceRefs pre-caches every reference so the mapper resolves
// offline, isolating the dry-run short-circuit from any cache-fill HTTP.
func TestCreateDevice_DryRunRecordsWithoutHTTP(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusCreated, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)
	e.Options.DryRun = true

	result := &LoadResult{}
	if err := e.createDevice(context.Background(), newMappableDevice("compute-001"), newCrudMapper(e), result); err != nil {
		t.Fatalf("createDevice() error = %v", err)
	}
	if !containsName(result.Created, "compute-001"+suffixDryRun) {
		t.Errorf("expected dry-run entry in Created, got %v", result.Created)
	}
	if calls != 0 {
		t.Errorf("expected no HTTP calls in dry-run, got %d", calls)
	}
}

// TestCreateDevice_CreatesOn201 verifies that a 201 response causes createDevice
// to record the device name in result.Created and issue exactly one POST.
//
// Why it matters: creating a device is the core write path for exporting cani
// inventory to Nautobot, so the success accounting and single round-trip must
// be exact. Inputs: a mappable device, DryRun off. Outputs: "compute-001" in
// result.Created and calls==1.
// Data choice: the handler returns 201 with an empty body because createDevice
// checks only the status code, and seedDeviceRefs keeps the mapper HTTP-free so
// the one counted call is the device POST itself.
func TestCreateDevice_CreatesOn201(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusCreated, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)

	result := &LoadResult{}
	if err := e.createDevice(context.Background(), newMappableDevice("compute-001"), newCrudMapper(e), result); err != nil {
		t.Fatalf("createDevice() error = %v", err)
	}
	if !containsName(result.Created, "compute-001") {
		t.Errorf("expected compute-001 in Created, got %v", result.Created)
	}
	if calls != 1 {
		t.Errorf("expected exactly one device POST, got %d", calls)
	}
}

// TestCreateDevice_ReturnsErrorOnNon201 verifies that a 400 response makes
// createDevice return an error instead of recording a success.
//
// Why it matters: a rejected device write must surface as an error so the
// export never reports inventory as present in Nautobot when it was not stored.
// Inputs: a mappable device, server replying 400. Outputs: a non-nil error.
// Data choice: a plain {"detail":"bad request"} body (without the status and
// content-type markers) exercises the generic unexpected-status path rather
// than the specialized friendly-error branch.
func TestCreateDevice_ReturnsErrorOnNon201(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusBadRequest, `{"detail":"bad request"}`))
	defer cleanup()
	seedDeviceRefs(t, e)

	result := &LoadResult{}
	if err := e.createDevice(context.Background(), newMappableDevice("compute-001"), newCrudMapper(e), result); err == nil {
		t.Fatal("expected an error when device create responds with 400")
	}
	if calls != 1 {
		t.Errorf("expected exactly one failed device POST, got %d", calls)
	}
}

// TestCreateDevice_ReturnsFriendlyStatusContentTypeError verifies that a 400
// whose body mentions both "status" and "Related object not found" is
// translated into guidance about an unsupported device status.
//
// Why it matters: a status not enabled for the dcim.device content type is a
// common Nautobot misconfiguration; a raw 400 dump is opaque, so the export
// rewrites it into actionable advice. Inputs: a mappable device, server
// replying 400 with that specific body. Outputs: an error containing "does not
// support dcim.device content type".
// Data choice: the body reproduces Nautobot's exact wording so the
// strings.Contains guards in createDevice fire and the friendly branch is hit.
func TestCreateDevice_ReturnsFriendlyStatusContentTypeError(t *testing.T) {
	var calls int
	body := `{"status":["Related object not found using the provided value"]}`
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusBadRequest, body))
	defer cleanup()
	seedDeviceRefs(t, e)

	result := &LoadResult{}
	err := e.createDevice(context.Background(), newMappableDevice("compute-001"), newCrudMapper(e), result)
	if err == nil || !strings.Contains(err.Error(), "does not support dcim.device content type") {
		t.Fatalf("expected a friendly content-type error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected exactly one failed device POST, got %d", calls)
	}
}

// TestCreateDevice_ReturnsMapperErrorForNilDevice verifies that a nil device
// makes createDevice fail during mapping, before any HTTP is attempted.
//
// Why it matters: mapping turns a cani device into a Nautobot request, so a nil
// input must short-circuit with an error rather than POST a malformed body.
// Inputs: device=nil. Outputs: a non-nil error and zero HTTP calls.
// Data choice: the cache is deliberately left unseeded — mapping fails on the
// nil device first, so no reference resolution or request is ever reached.
func TestCreateDevice_ReturnsMapperErrorForNilDevice(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusCreated, `{}`))
	defer cleanup()

	if err := e.createDevice(context.Background(), nil, newCrudMapper(e), &LoadResult{}); err == nil {
		t.Fatal("expected an error when the device is nil")
	}
	if calls != 0 {
		t.Errorf("expected no HTTP when mapping fails, got %d", calls)
	}
}

// -----------------------------------------------------------------------------
// updateDevice
// -----------------------------------------------------------------------------

// TestUpdateDevice_DryRunRecordsWithoutHTTP verifies that with DryRun enabled
// updateDevice records the device name (with the dry-run suffix) in
// result.Updated and issues no HTTP request.
//
// Why it matters: previewing an update to an already-exported cani device must
// not PATCH Nautobot, mirroring the create dry-run contract. Inputs: a mappable
// device, an existing ID, DryRun=true. Outputs: "compute-001"+suffixDryRun in
// result.Updated and zero calls.
// Data choice: seedDeviceRefs pre-caches references so the patch mapper resolves
// offline, isolating the dry-run guard from any cache-fill HTTP.
func TestUpdateDevice_DryRunRecordsWithoutHTTP(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)
	e.Options.DryRun = true

	result := &LoadResult{}
	if err := e.updateDevice(context.Background(), newMappableDevice("compute-001"), uuid.New(), newCrudMapper(e), result); err != nil {
		t.Fatalf("updateDevice() error = %v", err)
	}
	if !containsName(result.Updated, "compute-001"+suffixDryRun) {
		t.Errorf("expected dry-run entry in Updated, got %v", result.Updated)
	}
	if calls != 0 {
		t.Errorf("expected no HTTP calls in dry-run, got %d", calls)
	}
}

// TestUpdateDevice_UpdatesOn200 verifies that a 200 response causes updateDevice
// to record the device name in result.Updated and issue exactly one PATCH.
//
// Why it matters: updating reconciles drift between cani inventory and an
// existing Nautobot device, so the success tally and single round-trip must be
// exact. Inputs: a mappable device, an existing ID, DryRun off. Outputs:
// "compute-001" in result.Updated and calls==1.
// Data choice: the handler returns 200 (the only code updateDevice accepts for a
// PATCH) with an empty body, since only the status code is inspected.
func TestUpdateDevice_UpdatesOn200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)

	result := &LoadResult{}
	if err := e.updateDevice(context.Background(), newMappableDevice("compute-001"), uuid.New(), newCrudMapper(e), result); err != nil {
		t.Fatalf("updateDevice() error = %v", err)
	}
	if !containsName(result.Updated, "compute-001") {
		t.Errorf("expected compute-001 in Updated, got %v", result.Updated)
	}
	if calls != 1 {
		t.Errorf("expected exactly one device PATCH, got %d", calls)
	}
}

// TestUpdateDevice_ReturnsErrorOnNon200 verifies that a 500 response makes
// updateDevice return an error instead of recording a success.
//
// Why it matters: a failed reconciliation must surface so the export does not
// claim a device was updated in Nautobot when the PATCH was rejected. Inputs: a
// mappable device, an existing ID, server replying 500. Outputs: a non-nil
// error.
// Data choice: unlike createDevice, updateDevice treats any non-200 (including
// 201) as an error; 500 represents a server-side failure exercising that strict
// check.
func TestUpdateDevice_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)

	result := &LoadResult{}
	if err := e.updateDevice(context.Background(), newMappableDevice("compute-001"), uuid.New(), newCrudMapper(e), result); err == nil {
		t.Fatal("expected an error when device update responds with 500")
	}
	if calls != 1 {
		t.Errorf("expected exactly one failed device PATCH, got %d", calls)
	}
}

// TestUpdateDevice_ReturnsMapperErrorForNilDevice verifies that a nil device
// makes updateDevice fail during patch mapping, before any HTTP is attempted.
//
// Why it matters: building a PATCH from a nil cani device must short-circuit
// with an error rather than send a malformed update to Nautobot. Inputs:
// device=nil with a non-nil existing ID. Outputs: a non-nil error and zero HTTP
// calls.
// Data choice: the cache is left unseeded because mapping fails on the nil
// device first, so reference resolution and the PATCH are never reached.
func TestUpdateDevice_ReturnsMapperErrorForNilDevice(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	if err := e.updateDevice(context.Background(), nil, uuid.New(), newCrudMapper(e), &LoadResult{}); err == nil {
		t.Fatal("expected an error when the device is nil")
	}
	if calls != 0 {
		t.Errorf("expected no HTTP when mapping fails, got %d", calls)
	}
}

// -----------------------------------------------------------------------------
// createRack
// -----------------------------------------------------------------------------

// TestCreateRack_DryRunRecordsWithoutHTTP verifies that with DryRun enabled
// createRack records the rack name (with the dry-run suffix) in
// result.RacksCreated and issues no HTTP request.
//
// Why it matters: racks are the location containers cani devices export into, so
// a dry run must preview rack creation without mutating Nautobot. Inputs: a
// mappable device used as a rack, DryRun=true. Outputs: "rack-1"+suffixDryRun in
// result.RacksCreated and zero calls.
// Data choice: seedDeviceRefs supplies the location/status the rack mapper needs
// so it resolves offline, isolating the dry-run guard from cache-fill HTTP.
func TestCreateRack_DryRunRecordsWithoutHTTP(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusCreated, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)
	e.Options.DryRun = true

	result := &LoadResult{}
	if err := e.createRack(context.Background(), newMappableDevice("rack-1"), newCrudMapper(e), result); err != nil {
		t.Fatalf("createRack() error = %v", err)
	}
	if !containsName(result.RacksCreated, "rack-1"+suffixDryRun) {
		t.Errorf("expected dry-run entry in RacksCreated, got %v", result.RacksCreated)
	}
	if calls != 0 {
		t.Errorf("expected no HTTP calls in dry-run, got %d", calls)
	}
}

// TestCreateRack_CreatesOn201 verifies that a 201 response causes createRack to
// record the rack name in result.RacksCreated and issue exactly one POST.
//
// Why it matters: racks must exist before devices can reference them, so the
// success accounting and single round-trip must be exact. Inputs: a mappable
// device used as a rack, DryRun off. Outputs: "rack-1" in result.RacksCreated
// and calls==1.
// Data choice: the handler returns 201 with an empty body because createRack
// inspects only the status code, and seeded refs keep the mapper HTTP-free.
func TestCreateRack_CreatesOn201(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusCreated, `{}`))
	defer cleanup()
	seedDeviceRefs(t, e)

	result := &LoadResult{}
	if err := e.createRack(context.Background(), newMappableDevice("rack-1"), newCrudMapper(e), result); err != nil {
		t.Fatalf("createRack() error = %v", err)
	}
	if !containsName(result.RacksCreated, "rack-1") {
		t.Errorf("expected rack-1 in RacksCreated, got %v", result.RacksCreated)
	}
	if calls != 1 {
		t.Errorf("expected exactly one rack POST, got %d", calls)
	}
}

// TestCreateRack_ReturnsErrorOnNon201 verifies that a 400 response makes
// createRack return an error instead of recording a success.
//
// Why it matters: a rejected rack write must surface as an error so downstream
// devices are not exported against a rack that never persisted in Nautobot.
// Inputs: a mappable device used as a rack, server replying 400. Outputs: a
// non-nil error.
// Data choice: a generic {"detail":"bad"} body drives the unexpected-status
// path; createRack has no specialized friendly-error branch to bypass.
func TestCreateRack_ReturnsErrorOnNon201(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusBadRequest, `{"detail":"bad"}`))
	defer cleanup()
	seedDeviceRefs(t, e)

	result := &LoadResult{}
	if err := e.createRack(context.Background(), newMappableDevice("rack-1"), newCrudMapper(e), result); err == nil {
		t.Fatal("expected an error when rack create responds with 400")
	}
	if calls != 1 {
		t.Errorf("expected exactly one failed rack POST, got %d", calls)
	}
}

// TestCreateRack_ReturnsMapperErrorForNilDevice verifies that a nil device makes
// createRack fail during mapping, before any HTTP is attempted.
//
// Why it matters: mapping a nil cani device to a rack request must short-circuit
// with an error rather than POST a malformed rack to Nautobot. Inputs:
// device=nil. Outputs: a non-nil error and zero HTTP calls.
// Data choice: the cache is left unseeded because mapping fails on the nil
// device first, so reference resolution and the rack POST are never reached.
func TestCreateRack_ReturnsMapperErrorForNilDevice(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusCreated, `{}`))
	defer cleanup()

	if err := e.createRack(context.Background(), nil, newCrudMapper(e), &LoadResult{}); err == nil {
		t.Fatal("expected an error when the rack device is nil")
	}
	if calls != 0 {
		t.Errorf("expected no HTTP when mapping fails, got %d", calls)
	}
}
