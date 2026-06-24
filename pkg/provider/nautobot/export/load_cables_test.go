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

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// cableFixture bundles a cable plus the inventory, device-ID map and Nautobot
// interface IDs needed to drive createCaniCableType.
type cableFixture struct {
	cable       *devicetypes.CaniCableType
	cableID     uuid.UUID
	inv         *devicetypes.Inventory
	deviceIDMap map[string]uuid.UUID
	deviceAID   uuid.UUID
	deviceBID   uuid.UUID
	ifaceAID    uuid.UUID
	ifaceBID    uuid.UUID
}

// newCableFixture builds a two-device cable whose endpoints resolve cleanly:
// both devices are present in the inventory and in the Nautobot device-ID map.
func newCableFixture() cableFixture {
	caniDevA, caniDevB := uuid.New(), uuid.New()
	devAID, devBID := uuid.New(), uuid.New()
	devA := &devicetypes.CaniDeviceType{Name: "spine-01"}
	devB := &devicetypes.CaniDeviceType{Name: "leaf-01"}
	cable := &devicetypes.CaniCableType{
		Label:              "cable-1",
		TerminationADevice: caniDevA,
		TerminationBDevice: caniDevB,
		TerminationAPort:   "eth0",
		TerminationBPort:   "eth1",
	}
	cable.Status = "connected" // promoted from embedded ObjectMeta
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{caniDevA: devA, caniDevB: devB},
		Modules: map[uuid.UUID]*devicetypes.CaniModuleType{},
	}
	return cableFixture{
		cable:       cable,
		cableID:     uuid.New(),
		inv:         inv,
		deviceIDMap: map[string]uuid.UUID{"spine-01": devAID, "leaf-01": devBID},
		deviceAID:   devAID,
		deviceBID:   devBID,
		ifaceAID:    uuid.New(),
		ifaceBID:    uuid.New(),
	}
}

// seedInterfaces pre-populates the interface cache so the fuzzy lookups in
// createCaniCableType hit the cache instead of issuing HTTP. cableA/cableB set
// the CableID on each cached interface (uuid.Nil means "not yet cabled").
func (f cableFixture) seedInterfaces(e *Exporter, cableA, cableB uuid.UUID) {
	e.Cache.CacheInterface(f.deviceAID, f.cable.TerminationAPort,
		&CachedItem{ID: f.ifaceAID, Name: f.cable.TerminationAPort, CableID: cableA})
	e.Cache.CacheInterface(f.deviceBID, f.cable.TerminationBPort,
		&CachedItem{ID: f.ifaceBID, Name: f.cable.TerminationBPort, CableID: cableB})
}

// cableServer answers the cable existence GET (empty by default) and the cable
// create POST. existStatus lets a test force the existence check to fail so the
// warn-and-continue branch runs. postCalls counts only the create POSTs.
func cableServer(existStatus, createStatus int, postCalls *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "dcim/cables") {
			if r.Method == http.MethodPost {
				*postCalls++
				w.WriteHeader(createStatus)
				_, _ = io.WriteString(w, fmt.Sprintf(`{"id":%q,"display":"cable-1"}`, uuid.NewString()))
				return
			}
			w.WriteHeader(existStatus)
			_, _ = io.WriteString(w, emptyListJSON)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
}

// seedConnectedStatus seeds the "Connected" status so the cable create path
// resolves its status reference without HTTP.
func seedConnectedStatus(e *Exporter) uuid.UUID {
	id := uuid.New()
	e.Cache.statuses["Connected"] = &CachedItem{ID: id, Name: "Connected"}
	return id
}

// -----------------------------------------------------------------------------
// createCaniCableType
// -----------------------------------------------------------------------------

// TestCreateCaniCableType_SkipsIncompleteTerminations verifies that a cable
// with only one termination set (TerminationBDevice == uuid.Nil) is silently
// skipped: no error, no counter change, and no create POST.
//
// Why it matters: cani inventory can contain half-defined cables with one end
// unconnected; the cable phase must drop these rather than emit a single-ended
// cable that Nautobot would reject.
// Inputs: a fixture cable with TerminationBDevice cleared; an empty LoadResult.
// Outputs: nil error, CablesCreated==0, CablesSkipped==0, postCalls==0.
// Data choice: clearing exactly one endpoint isolates the incomplete-termination
// guard, which is the first branch of createCaniCableType.
func TestCreateCaniCableType_SkipsIncompleteTerminations(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, cableServer(http.StatusOK, http.StatusCreated, &postCalls))
	defer cleanup()

	f := newCableFixture()
	f.cable.TerminationBDevice = uuid.Nil // incomplete: only one end set

	result := &LoadResult{}
	if err := e.createCaniCableType(context.Background(), f.cableID, f.cable, f.inv, f.deviceIDMap, result); err != nil {
		t.Fatalf("createCaniCableType() error = %v", err)
	}
	if result.CablesCreated != 0 || result.CablesSkipped != 0 {
		t.Errorf("expected no counters changed for incomplete cable, got created=%d skipped=%d",
			result.CablesCreated, result.CablesSkipped)
	}
	if postCalls != 0 {
		t.Errorf("expected no create POST for incomplete cable, got %d", postCalls)
	}
}

// TestCreateCaniCableType_ErrorsOnUnknownDevices verifies that an error is
// returned when a termination references a device UUID present in neither
// inventory.Devices nor inventory.Modules.
//
// Why it matters: a dangling device reference means the inventory is
// inconsistent; surfacing an error stops the export instead of silently
// dropping a connection that should appear in Nautobot.
// Inputs: a fixture whose TerminationADevice is repointed to a fresh UUID not
// in the inventory. Outputs: a non-nil error and no cable create POST.
// Data choice: a brand-new uuid.New() guarantees device resolution misses both
// maps, exercising the "unknown devices" branch.
func TestCreateCaniCableType_ErrorsOnUnknownDevices(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, cableServer(http.StatusOK, http.StatusCreated, &postCalls))
	defer cleanup()

	f := newCableFixture()
	// Point endpoint A at a device that is neither in Devices nor Modules.
	f.cable.TerminationADevice = uuid.New()

	result := &LoadResult{}
	if err := e.createCaniCableType(context.Background(), f.cableID, f.cable, f.inv, f.deviceIDMap, result); err == nil {
		t.Fatal("expected an error when a cable endpoint references an unknown device")
	}
	if postCalls != 0 {
		t.Errorf("expected no create POST for unknown endpoint devices, got %d", postCalls)
	}
}

// TestCreateCaniCableType_ErrorsWhenDeviceNotInNautobot verifies that an error
// is returned when a device exists in the cani inventory but cannot be resolved
// to a Nautobot device ID.
//
// Why it matters: cables can only be created after their devices are exported;
// if the Nautobot device ID is unresolvable, the cable create would fail
// downstream, so failing fast here is correct.
// Inputs: a valid cable fixture but an empty deviceIDMap, forcing a
// GetDeviceByName lookup that the fake server answers with an empty list.
// Outputs: a non-nil error and no cable create POST.
// Data choice: empty map plus empty-list response together force the fallback
// name lookup to fail, which is the scenario under test.
func TestCreateCaniCableType_ErrorsWhenDeviceNotInNautobot(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, cableServer(http.StatusOK, http.StatusCreated, &postCalls))
	defer cleanup()

	f := newCableFixture()

	result := &LoadResult{}
	// Empty device-ID map forces a GetDeviceByName lookup, which the server
	// answers with an empty list, so the Nautobot device ID cannot be resolved.
	if err := e.createCaniCableType(context.Background(), f.cableID, f.cable, f.inv, map[string]uuid.UUID{}, result); err == nil {
		t.Fatal("expected an error when the device is not found in Nautobot")
	}
	if postCalls != 0 {
		t.Errorf("expected no create POST when endpoint device is missing in Nautobot, got %d", postCalls)
	}
}

// TestCreateCaniCableType_SkipsWhenInterfaceAlreadyCabled verifies that the
// cable is skipped (CablesSkipped incremented, no POST) when one endpoint
// interface already has a CableID attached.
//
// Why it matters: Nautobot rejects a second cable on an already-cabled
// interface ("must make a unique set"); skipping keeps the export idempotent
// across re-runs. This is an input-side conflict (an endpoint port reused by
// two cables), so it is tallied as CablesConflicted rather than CablesSkipped,
// which is reserved for cables that already exist in Nautobot.
// Inputs: a fixture whose seedInterfaces sets endpoint A's CableID to a non-nil
// UUID. Outputs: nil error, CablesConflicted==1, CablesSkipped==0, postCalls==0.
// Data choice: cabling only endpoint A (B left uuid.Nil) proves a single
// already-cabled side is enough to trigger the skip.
func TestCreateCaniCableType_SkipsWhenInterfaceAlreadyCabled(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, cableServer(http.StatusOK, http.StatusCreated, &postCalls))
	defer cleanup()

	f := newCableFixture()
	// Endpoint A already has a cable attached; the create must be skipped.
	f.seedInterfaces(e, uuid.New(), uuid.Nil)

	result := &LoadResult{}
	if err := e.createCaniCableType(context.Background(), f.cableID, f.cable, f.inv, f.deviceIDMap, result); err != nil {
		t.Fatalf("createCaniCableType() error = %v", err)
	}
	if result.CablesConflicted != 1 {
		t.Errorf("CablesConflicted = %d, want 1", result.CablesConflicted)
	}
	if result.CablesSkipped != 0 {
		t.Errorf("CablesSkipped = %d, want 0 (interface conflict is not an already-exists skip)", result.CablesSkipped)
	}
	if postCalls != 0 {
		t.Errorf("expected no create POST when an interface is already cabled, got %d", postCalls)
	}
}

// TestCreateCaniCableType_DryRunRecordsWithoutPost verifies that in dry-run
// mode the cable is counted as created (CablesCreated==1) but no create POST is
// issued.
//
// Why it matters: operators preview an export with --dry-run; the planned-cable
// count must be accurate while nothing is written to Nautobot.
// Inputs: e.Options.DryRun=true with both interfaces seeded uncabled.
// Outputs: nil error, CablesCreated==1, postCalls==0.
// Data choice: both interfaces seeded with a uuid.Nil CableID so the dry-run
// branch (not the already-cabled skip branch) is the one exercised.
func TestCreateCaniCableType_DryRunRecordsWithoutPost(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, cableServer(http.StatusOK, http.StatusCreated, &postCalls))
	defer cleanup()
	e.Options.DryRun = true

	f := newCableFixture()
	f.seedInterfaces(e, uuid.Nil, uuid.Nil)

	result := &LoadResult{}
	if err := e.createCaniCableType(context.Background(), f.cableID, f.cable, f.inv, f.deviceIDMap, result); err != nil {
		t.Fatalf("createCaniCableType() error = %v", err)
	}
	if result.CablesCreated != 1 {
		t.Errorf("CablesCreated = %d, want 1", result.CablesCreated)
	}
	if postCalls != 0 {
		t.Errorf("expected no create POST in dry-run, got %d", postCalls)
	}
}

// TestCreateCaniCableType_CreatesOn201 verifies the happy path: exactly one
// create POST is issued and CablesCreated is incremented when Nautobot returns
// 201.
//
// Why it matters: this is the core of the cable phase — each fully-resolved
// cani cable must produce exactly one Nautobot cable, no more.
// Inputs: both interfaces seeded uncabled, the "Connected" status seeded, and a
// server returning 201. Outputs: nil error, CablesCreated==1, postCalls==1.
// Data choice: seedConnectedStatus avoids an HTTP status lookup so the test
// isolates the create POST; uncabled interfaces let execution reach it.
func TestCreateCaniCableType_CreatesOn201(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, cableServer(http.StatusOK, http.StatusCreated, &postCalls))
	defer cleanup()
	seedConnectedStatus(e)

	f := newCableFixture()
	f.seedInterfaces(e, uuid.Nil, uuid.Nil)

	result := &LoadResult{}
	if err := e.createCaniCableType(context.Background(), f.cableID, f.cable, f.inv, f.deviceIDMap, result); err != nil {
		t.Fatalf("createCaniCableType() error = %v", err)
	}
	if result.CablesCreated != 1 {
		t.Errorf("CablesCreated = %d, want 1", result.CablesCreated)
	}
	if postCalls != 1 {
		t.Errorf("expected exactly one create POST, got %d", postCalls)
	}
}

// TestCreateCaniCableType_ContinuesWhenExistenceCheckErrors verifies that when
// the duplicate-cable check (GetCableByTerminations) returns 500, the function
// warns and still proceeds to create the cable.
//
// Why it matters: a transient failure of the duplicate check should not block
// exporting a valid cable; the code deliberately treats that error as
// non-fatal.
// Inputs: the existence GET configured to return 500 while the create POST
// returns 201. Outputs: nil error, CablesCreated==1, postCalls==1.
// Data choice: existStatus=500 with createStatus=201 isolates the
// warn-and-continue branch from the create path.
func TestCreateCaniCableType_ContinuesWhenExistenceCheckErrors(t *testing.T) {
	var postCalls int
	// The existence GET returns 500: createCaniCableType warns and proceeds.
	e, cleanup := newExporterWithServer(t, cableServer(http.StatusInternalServerError, http.StatusCreated, &postCalls))
	defer cleanup()
	seedConnectedStatus(e)

	f := newCableFixture()
	f.seedInterfaces(e, uuid.Nil, uuid.Nil)

	result := &LoadResult{}
	if err := e.createCaniCableType(context.Background(), f.cableID, f.cable, f.inv, f.deviceIDMap, result); err != nil {
		t.Fatalf("createCaniCableType() error = %v", err)
	}
	if result.CablesCreated != 1 {
		t.Errorf("CablesCreated = %d, want 1", result.CablesCreated)
	}
	if postCalls != 1 {
		t.Errorf("expected exactly one create POST after the existence check warns, got %d", postCalls)
	}
}

// TestCreateCaniCableType_ReturnsErrorOnNon201 verifies that an error is
// returned when the cable create POST responds with a non-success status (400).
//
// Why it matters: a failed write must surface so the operator knows the cable
// was not recorded in Nautobot rather than assuming success.
// Inputs: the create POST configured to return 400 with both interfaces seeded
// uncabled. Outputs: a non-nil error.
// Data choice: 400 (instead of 201) on the POST drives the unexpected-status
// error branch without tripping any of the earlier guards; postCalls==1 proves
// the error comes from the create request itself.
func TestCreateCaniCableType_ReturnsErrorOnNon201(t *testing.T) {
	var postCalls int
	e, cleanup := newExporterWithServer(t, cableServer(http.StatusOK, http.StatusBadRequest, &postCalls))
	defer cleanup()
	seedConnectedStatus(e)

	f := newCableFixture()
	f.seedInterfaces(e, uuid.Nil, uuid.Nil)

	result := &LoadResult{}
	if err := e.createCaniCableType(context.Background(), f.cableID, f.cable, f.inv, f.deviceIDMap, result); err == nil {
		t.Fatal("expected an error when the cable create responds with 400")
	}
	if postCalls != 1 {
		t.Errorf("expected exactly one failed cable create POST, got %d", postCalls)
	}
}
