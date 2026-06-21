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
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// moduleListCreateServer routes a GET (list lookup) and a POST (create) on the
// same resource path to distinct responses. The exporter's get-or-create
// helpers list first and create on a miss, so a single handler can drive both
// branches.
func moduleListCreateServer(listBody string, createStatus int, createBody string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(createStatus)
			_, _ = w.Write([]byte(createBody))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(listBody))
	}
}

// moduleIfaceServer answers the interface-create POST with the supplied body
// and every other request (the prefetch list and best-effort role lookup) with
// an empty result set. postCalls counts only the interface POSTs.
func moduleIfaceServer(postCalls *int, createStatus int, createBody string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "dcim/interfaces") {
			*postCalls++
			w.WriteHeader(createStatus)
			_, _ = w.Write([]byte(createBody))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":0,"results":[]}`))
	}
}

// -----------------------------------------------------------------------------
// getOrCreateModuleType
// -----------------------------------------------------------------------------

// TestGetOrCreateModuleType_FindsExisting verifies that getOrCreateModuleType
// returns the existing ModuleType (mapped to a CachedItem) when the Nautobot
// list returns a match, without attempting a create.
//
// Why it matters: module-types must exist before the modules that reference
// them; reusing an existing one keeps the export idempotent and avoids
// duplicating hardware definitions in the source-of-truth.
// Inputs: a CaniModuleType{Model:"NVIDIA A100"} plus a fake server whose list
// returns one match. Outputs: *CachedItem{ID, Name:"NVIDIA A100"}, nil error.
// Data choice: a realistic GPU model and a single-result list (count:1) drive
// the find-existing branch; the random UUID confirms the server id is mapped.
func TestGetOrCreateModuleType_FindsExisting(t *testing.T) {
	mtID := uuid.New()
	body := fmt.Sprintf(`{"count":1,"results":[{"id":%q,"model":"NVIDIA A100","display":"NVIDIA A100"}]}`, mtID.String())
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	module := &devicetypes.CaniModuleType{Model: "NVIDIA A100"}
	item, err := e.getOrCreateModuleType(context.Background(), module)
	if err != nil {
		t.Fatalf("getOrCreateModuleType() error = %v", err)
	}
	if item == nil || item.ID != mtID || item.Name != "NVIDIA A100" {
		t.Errorf("expected existing module type %s/NVIDIA A100, got %+v", mtID, item)
	}
}

// TestGetOrCreateModuleType_ErrorsWhenNotFoundAndCreateDisabled verifies that
// getOrCreateModuleType returns an error (and attempts no create) when the list
// is empty and CreateModuleTypes is false, its default.
//
// Why it matters: by default the export must not invent module-types; a missing
// type should fail loudly so operators explicitly opt in to creation rather
// than silently diverging from Nautobot.
// Inputs: a CaniModuleType{Model:"Unknown NIC"} plus an empty list; the create
// flag left at its default false. Outputs: a non-nil error.
// Data choice: empty results (count:0) force the not-found path, and leaving
// the option unset exercises the guard without extra setup.
func TestGetOrCreateModuleType_ErrorsWhenNotFoundAndCreateDisabled(t *testing.T) {
	// CreateModuleTypes defaults to false, so an absent module type is fatal.
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{"count":0,"results":[]}`))
	defer cleanup()

	module := &devicetypes.CaniModuleType{Model: "Unknown NIC"}
	if _, err := e.getOrCreateModuleType(context.Background(), module); err == nil {
		t.Fatal("expected an error when the module type is absent and creation is disabled")
	}
}

// TestGetOrCreateModuleType_CreatesWhenEnabled verifies that, with
// CreateModuleTypes enabled and an empty list, getOrCreateModuleType POSTs a
// new ModuleType and returns the created CachedItem.
//
// Why it matters: opt-in creation lets the export seed Nautobot from cani
// inventory so modules can attach to a type that did not previously exist.
// Inputs: a CaniModuleType{ConnectX-6, NVIDIA, PartNumber}, a seeded
// manufacturer, and a server that lists empty then answers 201. Outputs: a
// *CachedItem carrying the created id and model.
// Data choice: seeding "NVIDIA" avoids an extra manufacturer round-trip;
// PartNumber exercises the optional-field path; moduleListCreateServer drives
// the empty-list and 201-create on a single resource path.
func TestGetOrCreateModuleType_CreatesWhenEnabled(t *testing.T) {
	mtID := uuid.New()
	createBody := fmt.Sprintf(`{"id":%q,"model":"ConnectX-6","display":"ConnectX-6"}`, mtID.String())
	e, cleanup := newExporterWithServer(t, moduleListCreateServer(`{"count":0,"results":[]}`, http.StatusCreated, createBody))
	defer cleanup()
	e.Options.CreateModuleTypes = true
	seedManufacturer(e, "NVIDIA")

	module := &devicetypes.CaniModuleType{Model: "ConnectX-6", Manufacturer: "NVIDIA", PartNumber: "MCX653"}
	item, err := e.getOrCreateModuleType(context.Background(), module)
	if err != nil {
		t.Fatalf("getOrCreateModuleType() error = %v", err)
	}
	if item == nil || item.ID != mtID || item.Name != "ConnectX-6" {
		t.Errorf("expected created module type %s/ConnectX-6, got %+v", mtID, item)
	}
}

// TestGetOrCreateModuleType_ErrorsOnCreateNon201 verifies that
// getOrCreateModuleType surfaces an error when the create POST returns a
// non-201 (here 400) status.
//
// Why it matters: a failed type creation must abort rather than leave modules
// referencing a type that was never persisted in Nautobot.
// Inputs: an empty list plus a 400 create body, with creation enabled and a
// seeded manufacturer. Outputs: a non-nil error.
// Data choice: a 400 with a {"detail"} body mimics a Nautobot validation
// rejection; the seeded manufacturer isolates the failure to the create step.
func TestGetOrCreateModuleType_ErrorsOnCreateNon201(t *testing.T) {
	e, cleanup := newExporterWithServer(t, moduleListCreateServer(`{"count":0,"results":[]}`, http.StatusBadRequest, `{"detail":"bad"}`))
	defer cleanup()
	e.Options.CreateModuleTypes = true
	seedManufacturer(e, "NVIDIA")

	module := &devicetypes.CaniModuleType{Model: "ConnectX-6", Manufacturer: "NVIDIA"}
	if _, err := e.getOrCreateModuleType(context.Background(), module); err == nil {
		t.Fatal("expected an error when module type creation responds with 400")
	}
}

// -----------------------------------------------------------------------------
// getOrCreateModuleBay
// -----------------------------------------------------------------------------

// TestGetOrCreateModuleBay_FindsExisting verifies that getOrCreateModuleBay
// returns the existing ModuleBay when the name+parent-device list returns a
// match, without issuing a create.
//
// Why it matters: module bays are the slots modules plug into; reusing an
// existing bay keeps re-exports idempotent and prevents duplicate bays on a
// device.
// Inputs: a parent device UUID and bay name "GPU Bay 0" plus a single-result
// list. Outputs: the existing *CachedItem{ID, Name}.
// Data choice: a count:1 result drives the find branch, and a random device
// UUID matches the lookup's parent-device filter shape.
func TestGetOrCreateModuleBay_FindsExisting(t *testing.T) {
	bayID := uuid.New()
	body := fmt.Sprintf(`{"count":1,"results":[{"id":%q,"name":"GPU Bay 0","display":"GPU Bay 0"}]}`, bayID.String())
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, body))
	defer cleanup()

	item, err := e.getOrCreateModuleBay(context.Background(), uuid.New(), "GPU Bay 0")
	if err != nil {
		t.Fatalf("getOrCreateModuleBay() error = %v", err)
	}
	if item == nil || item.ID != bayID || item.Name != "GPU Bay 0" {
		t.Errorf("expected existing module bay %s/GPU Bay 0, got %+v", bayID, item)
	}
}

// TestGetOrCreateModuleBay_CreatesWhenNotFound verifies that
// getOrCreateModuleBay creates the bay (POST -> 201) when the list is empty and
// returns the created CachedItem.
//
// Why it matters: a bay must exist before a module can be installed into one;
// auto-creating it lets the export place modules accurately. Unlike
// module-types, bay creation has no opt-in flag.
// Inputs: an empty list then a 201 create, a device UUID, and bay name
// "GPU Bay 1". Outputs: the created *CachedItem.
// Data choice: moduleListCreateServer supplies empty-list-then-201 on one path,
// and a fresh UUID stands in for the parent device.
func TestGetOrCreateModuleBay_CreatesWhenNotFound(t *testing.T) {
	bayID := uuid.New()
	createBody := fmt.Sprintf(`{"id":%q,"name":"GPU Bay 1","display":"GPU Bay 1"}`, bayID.String())
	e, cleanup := newExporterWithServer(t, moduleListCreateServer(`{"count":0,"results":[]}`, http.StatusCreated, createBody))
	defer cleanup()

	item, err := e.getOrCreateModuleBay(context.Background(), uuid.New(), "GPU Bay 1")
	if err != nil {
		t.Fatalf("getOrCreateModuleBay() error = %v", err)
	}
	if item == nil || item.ID != bayID || item.Name != "GPU Bay 1" {
		t.Errorf("expected created module bay %s/GPU Bay 1, got %+v", bayID, item)
	}
}

// TestGetOrCreateModuleBay_ErrorsOnCreateNon201 verifies that
// getOrCreateModuleBay returns an error when the create responds with 500.
//
// Why it matters: if a bay cannot be persisted the export must stop, otherwise
// module placement would reference a bay that does not exist in Nautobot.
// Inputs: an empty list then a 500 create. Outputs: a non-nil error.
// Data choice: a 500 with an empty body models a server-side failure, distinct
// from the 400 validation case used for module-type creation.
func TestGetOrCreateModuleBay_ErrorsOnCreateNon201(t *testing.T) {
	e, cleanup := newExporterWithServer(t, moduleListCreateServer(`{"count":0,"results":[]}`, http.StatusInternalServerError, `{}`))
	defer cleanup()

	if _, err := e.getOrCreateModuleBay(context.Background(), uuid.New(), "GPU Bay 2"); err == nil {
		t.Fatal("expected an error when module bay creation responds with 500")
	}
}

// -----------------------------------------------------------------------------
// createModuleInterfaces
// -----------------------------------------------------------------------------

// TestCreateModuleInterfaces_CreatesValidSkipsInvalid verifies that
// createModuleInterfaces creates only interfaces whose type Nautobot supports,
// skipping internal interconnects, and bumps IfacesCreated and the POST count
// by exactly one.
//
// Why it matters: a module's interfaces (e.g. HSN ports from a NIC) become
// device interfaces in Nautobot; filtering unsupported types (nvlink) avoids
// create failures while still recording real network ports.
// Inputs: a module with one 100gbase-x-qsfp28 and one nvlink interface, plus a
// seeded Active status. Outputs: IfacesCreated == 1 and one interface POST.
// Data choice: pairing a Nautobot-valid type with nvlink, an internal GPU
// interconnect, directly exercises the isValidNautobotInterfaceType filter.
func TestCreateModuleInterfaces_CreatesValidSkipsInvalid(t *testing.T) {
	var postCalls int
	created := fmt.Sprintf(`{"id":%q,"name":"hsn0","display":"hsn0"}`, uuid.NewString())
	e, cleanup := newExporterWithServer(t, moduleIfaceServer(&postCalls, http.StatusCreated, created))
	defer cleanup()
	seedActiveStatus(t, e)

	module := &devicetypes.CaniModuleType{
		Name: "ConnectX-6",
		Interfaces: []devicetypes.InterfaceSpec{
			{Name: "hsn0", Type: devicetypes.InterfacesElemType("100gbase-x-qsfp28")},
			{Name: "nvlink0", Type: devicetypes.InterfacesElemType("nvlink")},
		},
	}

	result := &LoadResult{}
	if err := e.createModuleInterfaces(context.Background(), module, uuid.New(), result); err != nil {
		t.Fatalf("createModuleInterfaces() error = %v", err)
	}
	if result.IfacesCreated != 1 {
		t.Errorf("IfacesCreated = %d, want 1 (only the Nautobot-supported interface)", result.IfacesCreated)
	}
	if postCalls != 1 {
		t.Errorf("expected exactly one interface POST, got %d", postCalls)
	}
}

// TestCreateModuleInterfaces_SkipsExistingInterface verifies that
// createModuleInterfaces skips an interface already present in the cache,
// issuing no POST and leaving IfacesCreated at zero.
//
// Why it matters: several modules can declare same-named interfaces on one
// device; deduping via the cache keeps the export idempotent and avoids
// duplicate-interface errors.
// Inputs: a module declaring "hsn0" with that interface pre-seeded via
// CacheInterface on the device. Outputs: IfacesCreated == 0 and zero POSTs.
// Data choice: pre-seeding the cache under the same device+name key models the
// "another module already created it" case the code guards against.
func TestCreateModuleInterfaces_SkipsExistingInterface(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, moduleIfaceServer(&postCalls, http.StatusCreated, `{}`))
	defer cleanup()
	deviceID := uuid.New()
	// Pre-seed the cache so the interface is treated as already present.
	e.Cache.CacheInterface(deviceID, "hsn0", &CachedItem{ID: uuid.New(), Name: "hsn0"})

	module := &devicetypes.CaniModuleType{
		Name: "ConnectX-6",
		Interfaces: []devicetypes.InterfaceSpec{
			{Name: "hsn0", Type: devicetypes.InterfacesElemType("100gbase-x-qsfp28")},
		},
	}

	result := &LoadResult{}
	if err := e.createModuleInterfaces(context.Background(), module, deviceID, result); err != nil {
		t.Fatalf("createModuleInterfaces() error = %v", err)
	}
	if result.IfacesCreated != 0 {
		t.Errorf("IfacesCreated = %d, want 0 (interface already exists)", result.IfacesCreated)
	}
	if postCalls != 0 {
		t.Errorf("expected no interface POST for an existing interface, got %d", postCalls)
	}
}

// TestCreateModuleInterfaces_ReturnsErrorWhenCreateFails verifies that
// createModuleInterfaces returns an error when the underlying interface create
// fails with 400.
//
// Why it matters: a failed interface create must propagate so the export does
// not silently drop a module's connectivity from Nautobot.
// Inputs: a module with one valid interface and a server that answers the
// interface POST with 400. Outputs: a non-nil error.
// Data choice: a Nautobot-supported type (100gbase-x-qsfp28) ensures the
// failure comes from the HTTP create, not the interface-type filter.
func TestCreateModuleInterfaces_ReturnsErrorWhenCreateFails(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, moduleIfaceServer(&postCalls, http.StatusBadRequest, `{"detail":"bad"}`))
	defer cleanup()
	seedActiveStatus(t, e)

	module := &devicetypes.CaniModuleType{
		Name: "ConnectX-6",
		Interfaces: []devicetypes.InterfaceSpec{
			{Name: "hsn0", Type: devicetypes.InterfacesElemType("100gbase-x-qsfp28")},
		},
	}

	result := &LoadResult{}
	if err := e.createModuleInterfaces(context.Background(), module, uuid.New(), result); err == nil {
		t.Fatal("expected an error when the underlying interface create fails")
	}
}

// TestCreateModuleInterfaces_EmptyInterfacesNoOp verifies that
// createModuleInterfaces does nothing (no error, no creates, no POSTs) for a
// module that declares no interfaces.
//
// Why it matters: many modules (e.g. plain GPUs) expose no network interfaces;
// the export must handle that cleanly without spurious API calls.
// Inputs: a module with an empty Interfaces slice. Outputs: IfacesCreated == 0,
// zero POSTs, and a nil error.
// Data choice: an interface-less module is the minimal fixture that proves the
// loop short-circuits.
func TestCreateModuleInterfaces_EmptyInterfacesNoOp(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, moduleIfaceServer(&postCalls, http.StatusCreated, `{}`))
	defer cleanup()

	module := &devicetypes.CaniModuleType{Name: "Empty"}
	result := &LoadResult{}
	if err := e.createModuleInterfaces(context.Background(), module, uuid.New(), result); err != nil {
		t.Fatalf("createModuleInterfaces() error = %v", err)
	}
	if result.IfacesCreated != 0 || postCalls != 0 {
		t.Errorf("expected no work for a module with no interfaces, got created=%d posts=%d",
			result.IfacesCreated, postCalls)
	}
}
