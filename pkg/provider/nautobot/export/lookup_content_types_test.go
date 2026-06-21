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

// patchedRefJSON renders the object Nautobot returns from a PATCH that updates
// content types on a status or role.
func patchedRefJSON(id uuid.UUID, name string) string {
	return fmt.Sprintf(`{"id":%q,"name":%q,"display":%q}`, id.String(), name, name)
}

// TestUpdateStatusContentTypes_ReturnsUpdatedItem verifies that a successful
// (200) PATCH of a status' content types returns a CachedItem with the matching
// ID/name and issues exactly one HTTP call.
//
// Why it matters: the exporter patches an existing status (e.g. "Active") to add
// any DCIM/IPAM content types it is missing so devices, racks, modules, prefixes
// and IPs can all reference that status; it must reuse the status, not duplicate
// requests.
// Inputs: status UUID, name "Active", content types ["dcim.device"]. Outputs:
// updated *CachedItem, nil error, and one PATCH.
// Data choice: "Active" is the canonical status the exporter resolves/seeds; the
// call counter asserts the find-or-update path stays a single request.
func TestUpdateStatusContentTypes_ReturnsUpdatedItem(t *testing.T) {
	id := uuid.New()
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, patchedRefJSON(id, "Active")))
	defer cleanup()

	item, err := e.Cache.UpdateStatusContentTypes(id, "Active", []string{"dcim.device"})
	if err != nil {
		t.Fatalf("UpdateStatusContentTypes() error = %v", err)
	}
	if item == nil || item.ID != id || item.Name != "Active" {
		t.Errorf("expected updated status %s, got %+v", id, item)
	}
	if calls != 1 {
		t.Errorf("expected exactly one PATCH call, got %d", calls)
	}
}

// TestUpdateStatusContentTypes_ReturnsErrorOnNon200 verifies that a non-200
// (500) response from the status PATCH is surfaced as an error.
//
// Why it matters: if a status cannot be updated to cover the content types the
// export needs, continuing would create objects referencing an unusable status;
// the failure must abort loudly instead of silently proceeding.
// Inputs: random UUID, name "Active", content types ["dcim.device"]; server
// replies 500. Outputs: a non-nil error.
// Data choice: 500 models a server-side failure, and an empty `{}` body suffices
// because only the status code drives the error branch.
func TestUpdateStatusContentTypes_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusInternalServerError, `{}`))
	defer cleanup()

	if _, err := e.Cache.UpdateStatusContentTypes(uuid.New(), "Active", []string{"dcim.device"}); err == nil {
		t.Fatal("expected an error when the status PATCH responds with 500")
	}
}

// TestUpdateRoleContentTypes_ReturnsUpdatedItem verifies that a successful (200)
// PATCH of a role's content types returns a CachedItem with the matching ID/name
// and issues exactly one HTTP call.
//
// Why it matters: device roles (e.g. leaf/spine) must be associated with the
// dcim.device content type before devices can be assigned that role during
// export; the exporter patches the role once and reuses it.
// Inputs: role UUID, name "leaf", content types ["dcim.device"]. Outputs: updated
// *CachedItem, nil error, and one PATCH.
// Data choice: "leaf" is a representative switch role in cani inventory; the call
// counter guards against redundant PATCHes.
func TestUpdateRoleContentTypes_ReturnsUpdatedItem(t *testing.T) {
	id := uuid.New()
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, patchedRefJSON(id, "leaf")))
	defer cleanup()

	item, err := e.Cache.UpdateRoleContentTypes(id, "leaf", []string{"dcim.device"})
	if err != nil {
		t.Fatalf("UpdateRoleContentTypes() error = %v", err)
	}
	if item == nil || item.ID != id || item.Name != "leaf" {
		t.Errorf("expected updated role %s, got %+v", id, item)
	}
	if calls != 1 {
		t.Errorf("expected exactly one PATCH call, got %d", calls)
	}
}

// TestUpdateRoleContentTypes_ReturnsErrorOnNon200 verifies that a non-200 (400)
// response from the role PATCH is surfaced as an error.
//
// Why it matters: a role that cannot be updated to cover dcim.device would leave
// devices unassignable to that role, so the export must fail rather than create
// objects against a broken role.
// Inputs: random UUID, name "leaf", content types ["dcim.device"]; server replies
// 400. Outputs: a non-nil error.
// Data choice: 400 (bad request) is the realistic Nautobot response to an invalid
// PATCH payload; the empty body is irrelevant to the status-code-driven branch.
func TestUpdateRoleContentTypes_ReturnsErrorOnNon200(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusBadRequest, `{}`))
	defer cleanup()

	if _, err := e.Cache.UpdateRoleContentTypes(uuid.New(), "leaf", []string{"dcim.device"}); err == nil {
		t.Fatal("expected an error when the role PATCH responds with 400")
	}
}
