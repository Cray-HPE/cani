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
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// deviceExistsByID maps Nautobot's retrieve status onto a tri-state result:
// 200 -> exists, 404 -> definitively gone (no error), anything else -> error so
// a transient failure is never mistaken for a deletion.
// -----------------------------------------------------------------------------

// TestDeviceExistsByID_TrueWhenFound verifies that a 200 retrieve response makes
// deviceExistsByID return (true, nil).
//
// Why it matters: confirming a stored Nautobot ID still resolves lets the export
// reuse it instead of re-creating an already-present device. Inputs: a random
// device ID, server replying 200. Outputs: exists=true, err=nil.
// Data choice: an empty {} body suffices because deviceExistsByID branches only
// on the HTTP status code, not the payload.
func TestDeviceExistsByID_TrueWhenFound(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	exists, err := e.deviceExistsByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("deviceExistsByID() error = %v", err)
	}
	if !exists {
		t.Error("expected exists=true for a 200 response")
	}
}

// TestDeviceExistsByID_FalseWithoutErrorOn404 verifies that a 404 retrieve
// response makes deviceExistsByID return (false, nil).
//
// Why it matters: a definitive "not found" must be reported as absence without
// an error so the export can safely re-create or re-resolve the cani device.
// Inputs: a random device ID, server replying 404. Outputs: exists=false,
// err=nil.
// Data choice: 404 is the one non-200 status the method treats as a clean
// negative; an empty {} body is enough since only the code is examined.
func TestDeviceExistsByID_FalseWithoutErrorOn404(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusNotFound, `{}`))
	defer cleanup()

	exists, err := e.deviceExistsByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("deviceExistsByID() error = %v, want nil for a 404", err)
	}
	if exists {
		t.Error("expected exists=false for a 404 response")
	}
}

// TestDeviceExistsByID_ErrorOnUnexpectedStatus verifies that a 500 retrieve
// response makes deviceExistsByID return an error.
//
// Why it matters: a transient failure must not be read as a deletion, otherwise
// the export could wrongly discard a valid Nautobot ID and duplicate the device.
// Inputs: a random device ID, server replying 500. Outputs: a non-nil error.
// Data choice: 500 stands in for any status outside {200,404}, exercising the
// default error branch that protects the prune logic from false negatives.
func TestDeviceExistsByID_ErrorOnUnexpectedStatus(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{}`))
	defer cleanup()

	if _, err := e.deviceExistsByID(context.Background(), uuid.New()); err == nil {
		t.Fatal("expected an error for a 500 response so a transient failure is not read as a deletion")
	}
}

// -----------------------------------------------------------------------------
// pruneStaleDeviceID drops a stored Nautobot ID only when the remote device is
// confirmed gone (404). It must leave the ID in place otherwise.
// -----------------------------------------------------------------------------

// TestPruneStaleDeviceID_NoStoredIDIsNoOp verifies that pruneStaleDeviceID does
// nothing when the device carries no stored Nautobot ID.
//
// Why it matters: pruning only makes sense for a recorded ID, so a device with
// none must skip the existence check entirely and leave its ExternalIDs
// untouched. Inputs: a device with nil ExternalIDs. Outputs: ExternalIDs stays
// nil and no HTTP occurs.
// Data choice: a bare &Exporter{} with no client (and no fake server) proves the
// method returns before touching the network when there is nothing to prune.
func TestPruneStaleDeviceID_NoStoredIDIsNoOp(t *testing.T) {
	// No client is needed: with no stored ID the method returns before any HTTP.
	e := &Exporter{}
	device := &devicetypes.CaniDeviceType{Name: "node-1"}

	e.pruneStaleDeviceID(context.Background(), device)

	if device.ExternalIDs != nil {
		t.Errorf("ExternalIDs should remain nil, got %+v", device.ExternalIDs)
	}
}

// TestPruneStaleDeviceID_RemovesIDWhenRemoteGone verifies that a 404 existence
// check makes pruneStaleDeviceID delete the stored Nautobot ID.
//
// Why it matters: clearing a dead ID lets the export fall back to name-based
// resolution instead of trusting a UUID that no longer exists (e.g. after a
// Nautobot rebuild). Inputs: a device whose ExternalIDs holds a stale ID, server
// replying 404. Outputs: the externalIDKeyNautobot entry is removed.
// Data choice: 404 is the only signal treated as "definitively gone", so it is
// the precise condition that should trigger deletion.
func TestPruneStaleDeviceID_RemovesIDWhenRemoteGone(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusNotFound, `{}`))
	defer cleanup()

	stale := uuid.New()
	device := &devicetypes.CaniDeviceType{
		Name:       "node-1",
		ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{externalIDKeyNautobot: stale}},
	}

	e.pruneStaleDeviceID(context.Background(), device)

	if _, ok := device.ExternalIDs[externalIDKeyNautobot]; ok {
		t.Error("expected the stale Nautobot ID to be removed after a 404")
	}
}

// TestPruneStaleDeviceID_KeepsIDWhenRemoteExists verifies that a 200 existence
// check makes pruneStaleDeviceID retain the stored Nautobot ID.
//
// Why it matters: a live device must keep its ID so the export updates the
// existing Nautobot record rather than creating a duplicate. Inputs: a device
// whose ExternalIDs holds a live ID, server replying 200. Outputs: the ID is
// unchanged.
// Data choice: 200 represents a present device, the case where pruning must be a
// no-op even though an ID is stored.
func TestPruneStaleDeviceID_KeepsIDWhenRemoteExists(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	live := uuid.New()
	device := &devicetypes.CaniDeviceType{
		Name:       "node-1",
		ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{externalIDKeyNautobot: live}},
	}

	e.pruneStaleDeviceID(context.Background(), device)

	if device.ExternalIDs[externalIDKeyNautobot] != live {
		t.Error("expected the Nautobot ID to be retained when the remote device still exists")
	}
}

// TestPruneStaleDeviceID_KeepsIDOnTransientError verifies that a 500 existence
// check leaves the stored Nautobot ID in place.
//
// Why it matters: a momentary server error must never be mistaken for a
// deletion, or the export would drop a valid ID and risk duplicating the device
// on the next run. Inputs: a device whose ExternalIDs holds an ID, server
// replying 500. Outputs: the ID is retained.
// Data choice: 500 makes deviceExistsByID return an error, so prune's
// (err==nil && !exists) guard is false and the ID survives.
func TestPruneStaleDeviceID_KeepsIDOnTransientError(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{}`))
	defer cleanup()

	id := uuid.New()
	device := &devicetypes.CaniDeviceType{
		Name:       "node-1",
		ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{externalIDKeyNautobot: id}},
	}

	e.pruneStaleDeviceID(context.Background(), device)

	if device.ExternalIDs[externalIDKeyNautobot] != id {
		t.Error("a transient 500 must not delete the stored ID")
	}
}

// -----------------------------------------------------------------------------
// statusRef resolves a status name into a Nautobot reference union.
// -----------------------------------------------------------------------------

// TestStatusRef_BuildsReferenceFromCachedStatus verifies that statusRef turns a
// cached status name into a reference whose ID round-trips back to the cached
// UUID.
//
// Why it matters: nearly every exported object (devices, interfaces, cables)
// carries a status reference, so the name->UUID->union encoding must be exact.
// Inputs: a cache holding "Active"->statusID. Outputs: a reference whose Id
// decodes via AsBulkWritableCableRequestStatusId0 back to statusID.
// Data choice: seeding the cache directly avoids HTTP and lets the test assert
// the union encoding in isolation from status lookup or creation.
func TestStatusRef_BuildsReferenceFromCachedStatus(t *testing.T) {
	cache := NewLookupCache(nil)
	statusID := uuid.New()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	e := &Exporter{Cache: cache}
	ref, err := e.statusRef("Active")
	if err != nil {
		t.Fatalf("statusRef() error = %v", err)
	}
	if ref.Id == nil {
		t.Fatal("expected a populated status reference ID")
	}
	got, err := ref.Id.AsBulkWritableCableRequestStatusId0()
	if err != nil {
		t.Fatalf("decoding reference ID: %v", err)
	}
	if got != statusID {
		t.Errorf("status reference ID = %s, want %s", got, statusID)
	}
}

// TestStatusRef_ReturnsErrorWhenStatusUnresolvable verifies that statusRef
// returns an error when the status cannot be resolved.
//
// Why it matters: building a reference to a non-existent status would produce an
// invalid export request, so the failure must propagate from GetStatus rather
// than yield a half-formed reference. Inputs: a status name absent from cache,
// server returning an empty list with auto-create off (the default). Outputs: a
// non-nil error.
// Data choice: an empty {"count":0,"results":[]} list with createStatuses=false
// forces GetStatus to report "not found" instead of creating the status.
func TestStatusRef_ReturnsErrorWhenStatusUnresolvable(t *testing.T) {
	// Server reports no matching status and auto-creation is off (the default),
	// so GetStatus — and therefore statusRef — fails.
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{"count":0,"results":[]}`))
	defer cleanup()

	if _, err := e.statusRef("Nonexistent"); err == nil {
		t.Fatal("expected an error when the status cannot be resolved")
	}
}
