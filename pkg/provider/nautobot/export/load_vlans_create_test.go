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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// sentVLANRole is a minimal view of the VLAN create payload that exposes the
// optional role reference so tests can assert it was resolved and attached.
type sentVLANRole struct {
	Role *wireIDRef `json:"role"`
}

// vlanCreateServer captures the body of the ipam/vlans create POST and replies
// with the supplied status. Any other request receives an empty result list.
func vlanCreateServer(createStatus int, vlanNID uuid.UUID, captured *[]byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "ipam/vlans") && r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			*captured = body
			w.WriteHeader(createStatus)
			_, _ = io.WriteString(w,
				`{"id":"`+vlanNID.String()+`","vid":100,"name":"vlan100","display":"vlan100"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
}

// -----------------------------------------------------------------------------
// createVLAN — direct unit tests for the inner VLAN creator
// -----------------------------------------------------------------------------

// TestCreateVLAN_CreatesWithStatusFallback verifies a VLAN with no explicit
// status is created using the default "Active" status, returning the new ID and
// capturing a create POST body.
//
// Why it matters: cani VLANs may omit a status; Nautobot requires one, so the
// export must substitute a sensible default rather than fail or send an empty
// status.
// Inputs: a context, a CaniVLAN (no Status), and a LoadResult. Outputs: the new
// VLAN UUID and an error; the POST body is captured for inspection.
// Data choice: the VLAN sets only VID/Name (no Status, no option default) and
// "Active" is seeded so the test pins the fallback to the seeded status.
func TestCreateVLAN_CreatesWithStatusFallback(t *testing.T) {
	resetIPAMCaches()
	vlanNID := uuid.New()
	var captured []byte
	e, cleanup := newExporterWithServer(t, vlanCreateServer(http.StatusCreated, vlanNID, &captured))
	defer cleanup()
	seedActiveStatus(t, e) // empty vlan.Status must fall back to "Active"

	// No Status set on the VLAN or in the exporter options: the default applies.
	vlan := &devicetypes.CaniVLAN{VID: 100, Name: "vlan100"}

	result := &LoadResult{}
	got, err := e.createVLAN(context.Background(), vlan, result)
	if err != nil {
		t.Fatalf("createVLAN() error = %v", err)
	}
	if got != vlanNID {
		t.Errorf("returned ID = %s, want %s", got, vlanNID)
	}
	if len(captured) == 0 {
		t.Fatal("expected a create POST body to be captured")
	}
}

// TestCreateVLAN_ResolvesRoleIntoPayload verifies that a VLAN's role name is
// resolved to a Nautobot role ID and attached to the create payload.
//
// Why it matters: VLAN roles (e.g. compute, storage) classify networks in
// Nautobot; the export must translate the cani role name into the role's remote
// ID or the classification is lost.
// Inputs: a CaniVLAN with Role "compute" plus a seeded role cache entry.
// Outputs: an error; the captured payload's role.id is asserted.
// Data choice: the role "compute" is pre-seeded with a known ID so the test can
// assert the payload carries exactly that ID, proving the lookup-and-attach.
func TestCreateVLAN_ResolvesRoleIntoPayload(t *testing.T) {
	resetIPAMCaches()
	vlanNID := uuid.New()
	var captured []byte
	e, cleanup := newExporterWithServer(t, vlanCreateServer(http.StatusCreated, vlanNID, &captured))
	defer cleanup()
	seedActiveStatus(t, e)

	roleID := uuid.New()
	e.Cache.roles["compute"] = &CachedItem{ID: roleID, Name: "compute"}

	vlan := &devicetypes.CaniVLAN{VID: 100, Name: "vlan100"}
	vlan.Status = "Active"
	vlan.Role = "compute"

	result := &LoadResult{}
	if _, err := e.createVLAN(context.Background(), vlan, result); err != nil {
		t.Fatalf("createVLAN() error = %v", err)
	}

	var payload sentVLANRole
	if err := json.Unmarshal(captured, &payload); err != nil {
		t.Fatalf("decode VLAN payload: %v\nbody: %s", err, captured)
	}
	if payload.Role == nil || payload.Role.ID != roleID.String() {
		t.Errorf("payload role = %v, want id %s", payload.Role, roleID)
	}
}

// TestCreateVLAN_DryRunSkipsCreate verifies dry-run returns a non-nil
// placeholder ID and issues no create POST.
//
// Why it matters: previewing must not create VLANs in Nautobot, yet downstream
// mapping still needs a non-nil ID to keep references coherent in the preview.
// Inputs: the create path with Options.DryRun=true. Outputs: a non-nil UUID and
// an error; the captured body must stay empty.
// Data choice: a status of "Active" is set so the code reaches the create
// decision, isolating dry-run as the sole reason no POST is sent.
func TestCreateVLAN_DryRunSkipsCreate(t *testing.T) {
	resetIPAMCaches()
	var captured []byte
	e, cleanup := newExporterWithServer(t, vlanCreateServer(http.StatusCreated, uuid.New(), &captured))
	defer cleanup()
	seedActiveStatus(t, e)
	e.Options.DryRun = true

	vlan := &devicetypes.CaniVLAN{VID: 100, Name: "vlan100"}
	vlan.Status = "Active"

	result := &LoadResult{}
	got, err := e.createVLAN(context.Background(), vlan, result)
	if err != nil {
		t.Fatalf("createVLAN() error = %v", err)
	}
	if got == uuid.Nil {
		t.Error("dry-run should still return a non-nil placeholder ID")
	}
	if len(captured) != 0 {
		t.Errorf("expected no create POST in dry-run, captured %d bytes", len(captured))
	}
}

// TestCreateVLAN_ReturnsErrorWhenStatusLookupFails verifies the create aborts
// when the status cannot be resolved.
//
// Why it matters: a VLAN cannot be created without a valid status FK; if the
// status lookup fails the export must error rather than POST an invalid VLAN.
// Inputs: a server returning 500 for extras/statuses, with "Active" not seeded.
// Outputs: a non-nil error; no VLAN POST occurs.
// Data choice: the handler fails only the statuses endpoint (500) so the failure
// is attributable specifically to status resolution.
func TestCreateVLAN_ReturnsErrorWhenStatusLookupFails(t *testing.T) {
	resetIPAMCaches()
	// The status lookup returns 500 and "Active" is not seeded, so GetStatus —
	// and therefore createVLAN — fails before any VLAN POST.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "extras/statuses") {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, `{"detail":"boom"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	vlan := &devicetypes.CaniVLAN{VID: 100, Name: "vlan100"}

	result := &LoadResult{}
	if _, err := e.createVLAN(context.Background(), vlan, result); err == nil {
		t.Fatal("expected an error when the status cannot be resolved")
	}
}

// TestCreateVLAN_ReturnsErrorOnNon201 verifies a non-201 VLAN create response is
// surfaced as an error.
//
// Why it matters: a rejected VLAN create must abort rather than be treated as
// success, since prefixes and IPs depend on the VLAN existing.
// Inputs: the create path with the ipam/vlans POST returning 400. Outputs: a
// non-nil error.
// Data choice: only the create status is flipped to 400 (status still seeded) so
// the failure is isolated to the VLAN POST.
func TestCreateVLAN_ReturnsErrorOnNon201(t *testing.T) {
	resetIPAMCaches()
	var captured []byte
	e, cleanup := newExporterWithServer(t, vlanCreateServer(http.StatusBadRequest, uuid.New(), &captured))
	defer cleanup()
	seedActiveStatus(t, e)

	vlan := &devicetypes.CaniVLAN{VID: 100, Name: "vlan100"}
	vlan.Status = "Active"

	result := &LoadResult{}
	if _, err := e.createVLAN(context.Background(), vlan, result); err == nil {
		t.Fatal("expected an error when the VLAN create responds with 400")
	}
}
