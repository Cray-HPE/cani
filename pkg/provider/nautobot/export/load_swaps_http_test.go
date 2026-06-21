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

// swapInventoryServer routes the two requests collectPendingMoves makes: the
// rack-name lookup and the per-device retrieve.
func swapInventoryServer(rackNautobotID, deviceID uuid.UUID) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		switch {
		case strings.Contains(r.URL.Path, "dcim/racks"):
			_, _ = io.WriteString(w, fmt.Sprintf(
				`{"count":1,"results":[{"id":%q,"name":"rack-1","display":"rack-1"}]}`,
				rackNautobotID.String()))
		case strings.Contains(r.URL.Path, "dcim/devices"):
			_, _ = io.WriteString(w, deviceObjectJSON(deviceID, "compute-001"))
		default:
			_, _ = io.WriteString(w, `{"count":0,"results":[]}`)
		}
	}
}

// -----------------------------------------------------------------------------
// clearDevicePosition — PATCH /dcim/devices/{id}/ clearing rack/position/face.
// -----------------------------------------------------------------------------

// TestClearDevicePosition_SucceedsOn200 verifies clearDevicePosition issues
// exactly one PATCH and returns nil when Nautobot answers 200.
//
// Why it matters: clearing a device's rack/position/face is how the exporter
// frees a slot to break a swap cycle; it must fire once and report success so
// the freed device can be re-placed by the normal merge pass.
// Inputs: a device UUID; a fake server returning 200 "{}". Outputs: a nil error
// and a call count of 1.
// Data choice: counting calls guards against accidental retries or duplicate
// PATCHes against the single slot-clearing endpoint.
func TestClearDevicePosition_SucceedsOn200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	if err := e.clearDevicePosition(context.Background(), uuid.New()); err != nil {
		t.Fatalf("clearDevicePosition() error = %v, want nil", err)
	}
	if calls != 1 {
		t.Errorf("expected exactly one PATCH call, got %d", calls)
	}
}

// TestClearDevicePosition_ReturnsErrorOnNon200 verifies clearDevicePosition
// returns an error when the PATCH responds with a non-200 status.
//
// Why it matters: if freeing a slot silently failed, the subsequent
// re-placement PATCH would hit Nautobot's unique (rack, position, face)
// constraint and the swap would be left half-applied.
// Inputs: a device UUID; a fake server returning 400 with a detail body.
// Outputs: a non-nil error.
// Data choice: 400 represents Nautobot rejecting the clear via validation, the
// realistic failure mode this guard must surface.
func TestClearDevicePosition_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusBadRequest, `{"detail":"bad"}`))
	defer cleanup()

	if err := e.clearDevicePosition(context.Background(), uuid.New()); err == nil {
		t.Fatal("expected an error when the position clear responds with 400")
	}
}

// -----------------------------------------------------------------------------
// collectPendingMoves — detect devices whose rack slot is changing.
// -----------------------------------------------------------------------------

// TestCollectPendingMoves_RecordsChangedPosition verifies collectPendingMoves
// records one move when a device's desired rack slot differs from its remote
// placement.
//
// Why it matters: swap resolution only acts on devices whose slot is actually
// changing, so the move list must capture the device name and target position
// the merge intends to apply.
// Inputs: an Inventory placing compute-001 at rack-1 U2/front with a nautobot
// external ID; the fake server resolves the rack and returns the remote device.
// Outputs: one pendingMove with DeviceName "compute-001" and To.Position 2.
// Data choice: deviceObjectJSON returns a device with no position/rack, so the
// remote slot is nil and a move is unambiguously detected.
func TestCollectPendingMoves_RecordsChangedPosition(t *testing.T) {
	deviceID := uuid.New()
	rackLocalID := uuid.New()
	rackNautobotID := uuid.New()

	e, cleanup := newExporterWithServer(t, swapInventoryServer(rackNautobotID, deviceID))
	defer cleanup()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			deviceID: {
				Name:         "compute-001",
				Type:         devicetypes.TypeNode,
				RackPosition: 2,
				Face:         "front",
				Parent:       rackLocalID,
				ObjectMeta:   devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": deviceID}},
			},
		},
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackLocalID: {Name: "rack-1"},
		},
	}

	moves, _, err := e.collectPendingMoves(context.Background(), inv)
	if err != nil {
		t.Fatalf("collectPendingMoves() error = %v", err)
	}
	if len(moves) != 1 {
		t.Fatalf("expected 1 pending move, got %d", len(moves))
	}
	if moves[0].DeviceName != "compute-001" || moves[0].To.Position != 2 {
		t.Errorf("unexpected move: %+v", moves[0])
	}
}

// TestCollectPendingMoves_SkipsUnplacedDevices verifies collectPendingMoves
// produces no moves for a device with a non-positive rack position.
//
// Why it matters: unplaced devices have no slot to swap into, so including them
// would generate spurious moves and needless PATCHes during a merge export.
// Inputs: an Inventory with one device at RackPosition 0 carrying a nautobot
// external ID. Outputs: an empty move slice and a nil error.
// Data choice: RackPosition 0 is the sentinel for "unplaced" that the
// RackPosition<=0 guard is meant to skip.
func TestCollectPendingMoves_SkipsUnplacedDevices(t *testing.T) {
	deviceID := uuid.New()
	e, cleanup := newExporterWithServer(t, swapInventoryServer(uuid.New(), deviceID))
	defer cleanup()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			deviceID: {
				Name:         "compute-001",
				Type:         devicetypes.TypeNode,
				RackPosition: 0, // unplaced: must be skipped
				ObjectMeta:   devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": deviceID}},
			},
		},
	}

	moves, _, err := e.collectPendingMoves(context.Background(), inv)
	if err != nil {
		t.Fatalf("collectPendingMoves() error = %v", err)
	}
	if len(moves) != 0 {
		t.Errorf("expected no moves for an unplaced device, got %d", len(moves))
	}
}

// -----------------------------------------------------------------------------
// resolvePositionSwaps — top-level guard behavior.
// -----------------------------------------------------------------------------

// TestResolvePositionSwaps_NoopWhenNotMerging verifies resolvePositionSwaps does
// nothing and makes no HTTP calls when Merge is disabled.
//
// Why it matters: swap reconciliation only applies to merge exports; on a fresh
// create-only run it must not touch Nautobot or mutate existing placements.
// Inputs: an empty Inventory with Options.Merge=false. Outputs: a nil error and
// a call count of 0.
// Data choice: asserting zero calls proves the merge guard short-circuits before
// any device fetch or PATCH is attempted.
func TestResolvePositionSwaps_NoopWhenNotMerging(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()
	e.Options.Merge = false

	if err := e.resolvePositionSwaps(context.Background(), &devicetypes.Inventory{}); err != nil {
		t.Fatalf("resolvePositionSwaps() error = %v, want nil", err)
	}
	if calls != 0 {
		t.Errorf("expected no HTTP calls when merge is disabled, got %d", calls)
	}
}

// TestResolvePositionSwaps_NoopWhenNoMoves verifies resolvePositionSwaps returns
// nil without error when merging an inventory that yields no pending moves.
//
// Why it matters: with no slots changing there is no swap cycle to break, so the
// function must exit cleanly rather than issue any clear-position PATCHes.
// Inputs: an empty Inventory with Options.Merge=true. Outputs: a nil error and
// zero HTTP calls.
// Data choice: an empty inventory has no placed devices, the simplest way to
// drive collectPendingMoves to an empty result on the merge path.
func TestResolvePositionSwaps_NoopWhenNoMoves(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()
	e.Options.Merge = true

	// An inventory with no placed devices yields no pending moves.
	if err := e.resolvePositionSwaps(context.Background(), &devicetypes.Inventory{}); err != nil {
		t.Fatalf("resolvePositionSwaps() error = %v, want nil", err)
	}
	if calls != 0 {
		t.Errorf("expected no HTTP calls when there are no pending moves, got %d", calls)
	}
}
