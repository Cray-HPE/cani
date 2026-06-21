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

// moduleCreateConfig configures the fake Nautobot server used by
// createModuleFromCani tests.
type moduleCreateConfig struct {
	moduleTypeID    uuid.UUID
	moduleBayID     uuid.UUID
	moduleID        uuid.UUID
	idempotencyBody string // body for the modules-list (idempotency) GET
	createStatus    int    // status for the modules POST
}

// moduleCreateServer routes the four request kinds createModuleFromCani issues:
// the module-type lookup, the module-bay lookup (both returning an existing
// object so no sub-create is needed), the modules-list idempotency GET, and the
// modules create POST. modPostCalls counts only the module create POSTs.
func moduleCreateServer(cfg moduleCreateConfig, modPostCalls *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "dcim/module-types"):
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"count":1,"results":[{"id":%q,"model":"NVIDIA A100","display":"NVIDIA A100"}]}`, cfg.moduleTypeID.String())
		case strings.Contains(r.URL.Path, "dcim/module-bays"):
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"count":1,"results":[{"id":%q,"name":"GPU Bay 0","display":"GPU Bay 0"}]}`, cfg.moduleBayID.String())
		case strings.Contains(r.URL.Path, "dcim/modules"):
			if r.Method == http.MethodPost {
				*modPostCalls++
				w.WriteHeader(cfg.createStatus)
				fmt.Fprintf(w, `{"id":%q,"display":"gpu-0"}`, cfg.moduleID.String())
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(cfg.idempotencyBody))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":0,"results":[]}`))
		}
	}
}

// newModuleFixture builds a module plus the inventory and device-ID map needed
// to resolve its parent device.
func newModuleFixture(deviceName string) (*devicetypes.CaniModuleType, *devicetypes.Inventory, map[string]uuid.UUID) {
	parentDeviceID := uuid.New()
	parentNautobotID := uuid.New()
	dev := &devicetypes.CaniDeviceType{Name: deviceName}
	module := &devicetypes.CaniModuleType{
		Name:          "gpu-0",
		Model:         "NVIDIA A100",
		ParentDevice:  parentDeviceID,
		ModuleBayName: "GPU Bay 0",
		Serial:        "SN-GPU-001",
	}
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{parentDeviceID: dev},
	}
	createdDeviceIDs := map[string]uuid.UUID{deviceName: parentNautobotID}
	return module, inv, createdDeviceIDs
}

// defaultModuleCfg returns a config whose module type and bay already exist and
// whose bay is unoccupied, so the create POST proceeds.
func defaultModuleCfg(createStatus int) (moduleCreateConfig, *int) {
	var modPostCalls int
	return moduleCreateConfig{
		moduleTypeID:    uuid.New(),
		moduleBayID:     uuid.New(),
		moduleID:        uuid.New(),
		idempotencyBody: `{"count":0,"results":[]}`,
		createStatus:    createStatus,
	}, &modPostCalls
}

// -----------------------------------------------------------------------------
// createModuleFromCani
// -----------------------------------------------------------------------------

// TestCreateModuleFromCani_CreatesOn201 verifies a module is created in Nautobot
// when its type and bay already exist and the bay is unoccupied, incrementing
// ModulesCreated and issuing exactly one module POST.
//
// Why it matters: modules (e.g. GPUs, line cards) are exported into a specific
// module bay on their parent device; this is the primary create path and depends
// on resolving the module-type and module-bay foreign keys first.
// Inputs: a context, a CaniModuleType, the inventory, the device-name->ID map,
// and a LoadResult. Outputs: an error; side effects are the counters and POST.
// Data choice: a config whose module-type ("NVIDIA A100") and bay ("GPU Bay 0")
// resolve to existing objects with an empty idempotency list isolates the create
// branch from the type/bay sub-creation paths.
func TestCreateModuleFromCani_CreatesOn201(t *testing.T) {
	cfg, modPostCalls := defaultModuleCfg(http.StatusCreated)
	e, cleanup := newExporterWithServer(t, moduleCreateServer(cfg, modPostCalls))
	defer cleanup()
	seedActiveStatus(t, e)

	module, inv, createdDeviceIDs := newModuleFixture("compute-001")

	result := &LoadResult{}
	if err := e.createModuleFromCani(context.Background(), module, inv, createdDeviceIDs, result); err != nil {
		t.Fatalf("createModuleFromCani() error = %v", err)
	}
	if result.ModulesCreated != 1 {
		t.Errorf("ModulesCreated = %d, want 1", result.ModulesCreated)
	}
	if *modPostCalls != 1 {
		t.Errorf("expected exactly one module create POST, got %d", *modPostCalls)
	}
}

// TestCreateModuleFromCani_SkipsWhenBayOccupied verifies that when the
// idempotency GET shows the bay already holds a module, no POST is issued and
// ModulesSkipped is incremented.
//
// Why it matters: a module bay holds one module; re-exporting must not attempt
// to create a duplicate in an occupied bay.
// Inputs: the create path, but the idempotency body reports count:1. Outputs: an
// error; ModulesSkipped increments and no POST occurs.
// Data choice: a count:1 results list models the bay already populated from a
// prior export run.
func TestCreateModuleFromCani_SkipsWhenBayOccupied(t *testing.T) {
	cfg, modPostCalls := defaultModuleCfg(http.StatusCreated)
	cfg.idempotencyBody = fmt.Sprintf(`{"count":1,"results":[{"id":%q,"display":"gpu-0"}]}`, uuid.NewString())
	e, cleanup := newExporterWithServer(t, moduleCreateServer(cfg, modPostCalls))
	defer cleanup()
	seedActiveStatus(t, e)

	module, inv, createdDeviceIDs := newModuleFixture("compute-001")

	result := &LoadResult{}
	if err := e.createModuleFromCani(context.Background(), module, inv, createdDeviceIDs, result); err != nil {
		t.Fatalf("createModuleFromCani() error = %v", err)
	}
	if result.ModulesSkipped != 1 {
		t.Errorf("ModulesSkipped = %d, want 1", result.ModulesSkipped)
	}
	if *modPostCalls != 0 {
		t.Errorf("expected no module create POST when the bay is occupied, got %d", *modPostCalls)
	}
}

// TestCreateModuleFromCani_DryRunSkipsCreate verifies dry-run mode counts the
// module as created but issues no POST.
//
// Why it matters: previewing an export must not mutate Nautobot while still
// reporting the intended change.
// Inputs: the create path with Options.DryRun=true. Outputs: an error;
// ModulesCreated increments but modPostCalls stays 0.
// Data choice: an unoccupied bay (empty idempotency list) ensures the code
// reaches the create decision so the dry-run guard is the only reason no POST is
// sent.
func TestCreateModuleFromCani_DryRunSkipsCreate(t *testing.T) {
	cfg, modPostCalls := defaultModuleCfg(http.StatusCreated)
	e, cleanup := newExporterWithServer(t, moduleCreateServer(cfg, modPostCalls))
	defer cleanup()
	seedActiveStatus(t, e)
	e.Options.DryRun = true

	module, inv, createdDeviceIDs := newModuleFixture("compute-001")

	result := &LoadResult{}
	if err := e.createModuleFromCani(context.Background(), module, inv, createdDeviceIDs, result); err != nil {
		t.Fatalf("createModuleFromCani() error = %v", err)
	}
	if result.ModulesCreated != 1 {
		t.Errorf("ModulesCreated = %d, want 1", result.ModulesCreated)
	}
	if *modPostCalls != 0 {
		t.Errorf("expected no module create POST in dry-run, got %d", *modPostCalls)
	}
}

// TestCreateModuleFromCani_ErrorsWhenParentDeviceNotInInventory verifies the
// create fails when the module's parent device UUID is absent from the
// inventory.
//
// Why it matters: a module must attach to a known parent device; failing fast
// avoids creating an orphaned module in Nautobot.
// Inputs: a module whose ParentDevice points at a UUID not in inv.Devices.
// Outputs: a non-nil error.
// Data choice: ParentDevice is reassigned to a fresh random UUID so the
// inventory lookup misses.
func TestCreateModuleFromCani_ErrorsWhenParentDeviceNotInInventory(t *testing.T) {
	cfg, modPostCalls := defaultModuleCfg(http.StatusCreated)
	e, cleanup := newExporterWithServer(t, moduleCreateServer(cfg, modPostCalls))
	defer cleanup()

	module, inv, createdDeviceIDs := newModuleFixture("compute-001")
	module.ParentDevice = uuid.New() // absent from the inventory

	result := &LoadResult{}
	if err := e.createModuleFromCani(context.Background(), module, inv, createdDeviceIDs, result); err == nil {
		t.Fatal("expected an error when the parent device is not in the inventory")
	}
}

// TestCreateModuleFromCani_ErrorsWhenParentDeviceNotInNautobot verifies the
// create fails when the parent device exists in the inventory but was never
// created in Nautobot (absent from the device-name->ID map).
//
// Why it matters: the module's device FK must reference a real Nautobot device
// ID; without it the export would post an unresolvable reference.
// Inputs: a valid inventory but an empty createdDeviceIDs map. Outputs: a
// non-nil error.
// Data choice: an empty createdDeviceIDs map models the device-creation phase
// having skipped or failed for this device.
func TestCreateModuleFromCani_ErrorsWhenParentDeviceNotInNautobot(t *testing.T) {
	cfg, modPostCalls := defaultModuleCfg(http.StatusCreated)
	e, cleanup := newExporterWithServer(t, moduleCreateServer(cfg, modPostCalls))
	defer cleanup()

	module, inv, _ := newModuleFixture("compute-001")

	result := &LoadResult{}
	// Empty createdDeviceIDs: the parent device was never created in Nautobot.
	if err := e.createModuleFromCani(context.Background(), module, inv, map[string]uuid.UUID{}, result); err == nil {
		t.Fatal("expected an error when the parent device is not in Nautobot")
	}
}

// TestCreateModuleFromCani_ErrorsWhenModuleTypeUnresolvable verifies the create
// fails when the module type cannot be found and auto-creation is disabled.
//
// Why it matters: a module references a module-type FK; without an existing or
// creatable type the export cannot proceed and must error rather than guess.
// Inputs: a server returning empty lists for every lookup, with CreateModuleTypes
// left at its default (off). Outputs: a non-nil error.
// Data choice: an all-empty handler makes the module-type lookup miss, which
// together with the disabled create flag forces the unresolvable-type branch.
func TestCreateModuleFromCani_ErrorsWhenModuleTypeUnresolvable(t *testing.T) {
	// The module-type lookup returns an empty list and CreateModuleTypes is off
	// (default), so getOrCreateModuleType — and createModuleFromCani — fail.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":0,"results":[]}`))
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)

	module, inv, createdDeviceIDs := newModuleFixture("compute-001")

	result := &LoadResult{}
	if err := e.createModuleFromCani(context.Background(), module, inv, createdDeviceIDs, result); err == nil {
		t.Fatal("expected an error when the module type cannot be resolved")
	}
}

// TestCreateModuleFromCani_ReturnsErrorOnNon201 verifies a non-201 module create
// response is surfaced as an error.
//
// Why it matters: Nautobot rejections must abort the module create rather than
// be treated as success.
// Inputs: the create path with the modules POST returning 400. Outputs: a
// non-nil error.
// Data choice: defaultModuleCfg(http.StatusBadRequest) reuses the happy-path
// fixture but flips only the create status, isolating the failure to the POST.
func TestCreateModuleFromCani_ReturnsErrorOnNon201(t *testing.T) {
	cfg, modPostCalls := defaultModuleCfg(http.StatusBadRequest)
	e, cleanup := newExporterWithServer(t, moduleCreateServer(cfg, modPostCalls))
	defer cleanup()
	seedActiveStatus(t, e)

	module, inv, createdDeviceIDs := newModuleFixture("compute-001")

	result := &LoadResult{}
	if err := e.createModuleFromCani(context.Background(), module, inv, createdDeviceIDs, result); err == nil {
		t.Fatal("expected an error when module create responds with 400")
	}
}
