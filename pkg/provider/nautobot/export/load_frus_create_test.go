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

// fruServer answers the two requests createFruFromCani makes: a GET list for
// the idempotency check and a POST to create the inventory item. postCalls
// counts only the create POSTs.
func fruServer(idempotencyBody string, createStatus int, createBody string, postCalls *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "dcim/inventory-items") {
			*postCalls++
			w.WriteHeader(createStatus)
			_, _ = w.Write([]byte(createBody))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(idempotencyBody))
	}
}

// newFruFixture builds a FRU plus the inventory and device-ID map needed to
// resolve its parent device. The parent device is named deviceName and is
// already "created" in Nautobot under a fresh UUID.
func newFruFixture(deviceName string) (*devicetypes.CaniFruType, *devicetypes.Inventory, map[string]uuid.UUID) {
	parentDeviceID := uuid.New()
	parentNautobotID := uuid.New()
	dev := &devicetypes.CaniDeviceType{Name: deviceName}
	fru := &devicetypes.CaniFruType{
		Name:       "psu-0",
		Device:     parentDeviceID,
		PartNumber: "P12345",
		Serial:     "SN-001",
	}
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{parentDeviceID: dev},
	}
	createdDeviceIDs := map[string]uuid.UUID{deviceName: parentNautobotID}
	return fru, inv, createdDeviceIDs
}

// -----------------------------------------------------------------------------
// createFruFromCani
// -----------------------------------------------------------------------------

// TestCreateFruFromCani_CreatesOn201 verifies a FRU is created as a Nautobot
// inventory-item when the idempotency check finds nothing, returning the new ID
// and incrementing FrusCreated.
//
// Why it matters: FRUs (PSUs, fans, GPUs) are exported as inventory-items
// attached to their parent device; this is the core create path and it must also
// wire the manufacturer and parent-FRU foreign keys.
// Inputs: a context, a CaniFruType, the inventory, the device-name->Nautobot-ID
// and FRU->Nautobot-ID maps, and a LoadResult. Outputs: the new FRU UUID and an
// error; side effects increment result counters and issue one POST.
// Data choice: Manufacturer "HPE" (pre-seeded) and a Parent FRU present in the
// created-FRU map exercise both FK branches without extra HTTP round-trips.
func TestCreateFruFromCani_CreatesOn201(t *testing.T) {
	fruID := uuid.New()
	var postCalls int
	e, cleanup := newExporterWithServer(t, fruServer(emptyListJSON, http.StatusCreated,
		fmt.Sprintf(`{"id":%q,"display":"psu-0"}`, fruID.String()), &postCalls))
	defer cleanup()

	fru, inv, createdDeviceIDs := newFruFixture("compute-001")
	// Exercise the manufacturer and parent-FRU FK branches without extra HTTP.
	fru.Manufacturer = "HPE"
	seedManufacturer(e, "HPE")
	parentFru := uuid.New()
	fru.Parent = parentFru
	createdFruIDs := map[uuid.UUID]uuid.UUID{parentFru: uuid.New()}

	result := &LoadResult{}
	got, err := e.createFruFromCani(context.Background(), fru, inv, createdDeviceIDs, createdFruIDs, result)
	if err != nil {
		t.Fatalf("createFruFromCani() error = %v", err)
	}
	if got != fruID {
		t.Errorf("returned ID = %s, want %s", got, fruID)
	}
	if result.FrusCreated != 1 {
		t.Errorf("FrusCreated = %d, want 1", result.FrusCreated)
	}
	if postCalls != 1 {
		t.Errorf("expected exactly one create POST, got %d", postCalls)
	}
}

// TestCreateFruFromCani_SkipsWhenAlreadyExists verifies that when the idempotency
// GET already returns a matching inventory-item, no POST is issued, the existing
// ID is returned, and FrusSkipped is incremented.
//
// Why it matters: re-running an export must not duplicate FRUs in Nautobot.
// Inputs: the create path, but the server's idempotency body reports one
// existing item. Outputs: the existing UUID and an error; no POST occurs.
// Data choice: a count:1 list with a known ID simulates the FRU already being
// present from a prior run so the skip branch and returned ID can be asserted.
func TestCreateFruFromCani_SkipsWhenAlreadyExists(t *testing.T) {
	existingID := uuid.New()
	idempotency := fmt.Sprintf(`{"count":1,"results":[{"id":%q,"display":"psu-0"}]}`, existingID.String())
	var postCalls int
	e, cleanup := newExporterWithServer(t, fruServer(idempotency, http.StatusCreated, `{}`, &postCalls))
	defer cleanup()

	fru, inv, createdDeviceIDs := newFruFixture("compute-001")

	result := &LoadResult{}
	got, err := e.createFruFromCani(context.Background(), fru, inv, createdDeviceIDs, map[uuid.UUID]uuid.UUID{}, result)
	if err != nil {
		t.Fatalf("createFruFromCani() error = %v", err)
	}
	if got != existingID {
		t.Errorf("returned ID = %s, want existing %s", got, existingID)
	}
	if result.FrusSkipped != 1 {
		t.Errorf("FrusSkipped = %d, want 1", result.FrusSkipped)
	}
	if postCalls != 0 {
		t.Errorf("expected no create POST when the item already exists, got %d", postCalls)
	}
}

// TestCreateFruFromCani_DryRunSkipsCreate verifies dry-run mode reports the FRU
// as created (for the summary) but issues no POST and returns a Nil ID.
//
// Why it matters: operators preview an export with --dry-run; it must never
// mutate Nautobot yet still surface what would happen.
// Inputs: the create path with Options.DryRun=true. Outputs: uuid.Nil and an
// error; FrusCreated is incremented but postCalls stays 0.
// Data choice: an empty idempotency list ensures the code reaches the create
// decision, isolating the dry-run guard as the reason no POST is sent.
func TestCreateFruFromCani_DryRunSkipsCreate(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, fruServer(emptyListJSON, http.StatusCreated, `{}`, &postCalls))
	defer cleanup()
	e.Options.DryRun = true

	fru, inv, createdDeviceIDs := newFruFixture("compute-001")

	result := &LoadResult{}
	got, err := e.createFruFromCani(context.Background(), fru, inv, createdDeviceIDs, map[uuid.UUID]uuid.UUID{}, result)
	if err != nil {
		t.Fatalf("createFruFromCani() error = %v", err)
	}
	if got != uuid.Nil {
		t.Errorf("dry-run returned ID = %s, want Nil", got)
	}
	if result.FrusCreated != 1 {
		t.Errorf("FrusCreated = %d, want 1", result.FrusCreated)
	}
	if postCalls != 0 {
		t.Errorf("expected no create POST in dry-run, got %d", postCalls)
	}
}

// TestCreateFruFromCani_ErrorsWhenParentDeviceNotInInventory verifies the create
// fails when the FRU's parent device UUID is absent from the inventory.
//
// Why it matters: a FRU cannot be attached without a known parent device;
// failing fast prevents creating an orphaned inventory-item in Nautobot.
// Inputs: a FRU whose Device points at a UUID not present in inv.Devices.
// Outputs: a non-nil error.
// Data choice: fru.Device is reassigned to a fresh random UUID so the inventory
// lookup misses, isolating this specific precondition.
func TestCreateFruFromCani_ErrorsWhenParentDeviceNotInInventory(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, fruServer(emptyListJSON, http.StatusCreated, `{}`, &postCalls))
	defer cleanup()

	fru, inv, createdDeviceIDs := newFruFixture("compute-001")
	fru.Device = uuid.New() // points at a device absent from the inventory

	result := &LoadResult{}
	if _, err := e.createFruFromCani(context.Background(), fru, inv, createdDeviceIDs, map[uuid.UUID]uuid.UUID{}, result); err == nil {
		t.Fatal("expected an error when the parent device is not in the inventory")
	}
}

// TestCreateFruFromCani_ErrorsWhenParentDeviceNotInNautobot verifies the create
// fails when the parent device exists in the inventory but was never created in
// Nautobot (absent from the device-name->ID map).
//
// Why it matters: the inventory-item's device FK must reference a real Nautobot
// device ID; without it the export would post an unresolvable reference.
// Inputs: a valid inventory but an empty createdDeviceIDs map. Outputs: a
// non-nil error.
// Data choice: passing an empty createdDeviceIDs map models the case where the
// device-creation phase skipped or failed for this device.
func TestCreateFruFromCani_ErrorsWhenParentDeviceNotInNautobot(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, fruServer(emptyListJSON, http.StatusCreated, `{}`, &postCalls))
	defer cleanup()

	fru, inv, _ := newFruFixture("compute-001")

	result := &LoadResult{}
	// Empty createdDeviceIDs: the parent device was never created in Nautobot.
	if _, err := e.createFruFromCani(context.Background(), fru, inv, map[string]uuid.UUID{}, map[uuid.UUID]uuid.UUID{}, result); err == nil {
		t.Fatal("expected an error when the parent device is not in Nautobot")
	}
}

// TestCreateFruFromCani_ReturnsErrorOnNon201 verifies a non-201 create response
// is surfaced as an error.
//
// Why it matters: Nautobot rejections (validation, permissions) must abort the
// FRU create rather than be silently treated as success.
// Inputs: the create path with the server returning 400. Outputs: a non-nil
// error.
// Data choice: 400 Bad Request with a {"detail":"bad"} body mirrors a typical
// Nautobot validation rejection.
func TestCreateFruFromCani_ReturnsErrorOnNon201(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, fruServer(emptyListJSON, http.StatusBadRequest, `{"detail":"bad"}`, &postCalls))
	defer cleanup()

	fru, inv, createdDeviceIDs := newFruFixture("compute-001")

	result := &LoadResult{}
	if _, err := e.createFruFromCani(context.Background(), fru, inv, createdDeviceIDs, map[uuid.UUID]uuid.UUID{}, result); err == nil {
		t.Fatal("expected an error when inventory-item create responds with 400")
	}
}
