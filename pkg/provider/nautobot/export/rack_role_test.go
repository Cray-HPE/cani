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
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// roleTraffic records what a fake Nautobot saw for one role.
type roleTraffic struct {
	reads   int
	patched []string
}

// roleServer answers a role retrieve with the given content types and a
// role PATCH with the content types it was sent, recording both.
func roleServer(id uuid.UUID, contentTypes []string, seen *roleTraffic) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		role := map[string]any{"id": id.String(), "name": "Storage", "display": "Storage", "content_types": contentTypes}
		if r.Method == http.MethodPatch {
			var body struct {
				ContentTypes []string `json:"content_types"`
			}
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &body)
			seen.patched = body.ContentTypes
			role["content_types"] = body.ContentTypes
		} else {
			seen.reads++
		}
		_ = json.NewEncoder(w).Encode(role)
	}
}

// roleExporter wires an exporter to a fake Nautobot that holds one role,
// Storage, with the given content types, and seeds the cache with the role
// as an earlier lookup would have left it.
func roleExporter(t *testing.T, contentTypes []string, seen *roleTraffic) (*Exporter, uuid.UUID, func()) {
	t.Helper()
	id := uuid.New()
	e, cleanup := newExporterWithServer(t, roleServer(id, contentTypes, seen))
	e.Cache.rolesMu.Lock()
	e.Cache.roles["Storage"] = &CachedItem{ID: id, Name: "Storage"}
	e.Cache.rolesMu.Unlock()
	return e, id, cleanup
}

// rackRequestRole returns the role reference a rack request carries as JSON,
// "" when the request has none.
func rackRequestRole(t *testing.T, req *nautobotapi.WritableRackRequest) string {
	t.Helper()
	blob, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal rack request: %v", err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(blob, &payload); err != nil {
		t.Fatalf("unmarshal rack request: %v", err)
	}
	return string(payload["role"])
}

// TestGetRoleFor_AddsContentTypeOnceWhenRolesMayChange verifies that a role
// missing the rack content type gains it, keeping the types it had, and that
// the role is not read again for the next rack.
//
// Why it matters: a role cani finds or creates lists the content types of
// devices and IPAM objects only; Nautobot refuses a rack that names a role
// without dcim.rack, and the rack's devices fail with it.
// Inputs: a cached role whose Nautobot record lists dcim.device, with
// create_roles on, asked for twice.
// Outputs: one PATCH carrying dcim.device and dcim.rack, and one read.
// Data choice: an existing type in the role proves the update appends
// instead of replacing; two calls prove the per-run memo.
func TestGetRoleFor_AddsContentTypeOnceWhenRolesMayChange(t *testing.T) {
	// Arrange.
	seen := &roleTraffic{}
	e, id, cleanup := roleExporter(t, []string{contentTypeDevice}, seen)
	defer cleanup()
	e.Cache.SetCreateRoles(true)

	// Act.
	first, err := e.Cache.GetRoleFor("Storage", contentTypeRack)
	if err != nil {
		t.Fatalf("GetRoleFor() error = %v", err)
	}
	second, err := e.Cache.GetRoleFor("Storage", contentTypeRack)
	if err != nil {
		t.Fatalf("second GetRoleFor() error = %v", err)
	}

	// Assert.
	if first.ID != id || second.ID != id {
		t.Errorf("role IDs = %s, %s, want %s", first.ID, second.ID, id)
	}
	if want := []string{contentTypeDevice, contentTypeRack}; strings.Join(seen.patched, ",") != strings.Join(want, ",") {
		t.Errorf("PATCH content_types = %v, want %v", seen.patched, want)
	}
	if seen.reads != 1 {
		t.Errorf("role was read %d time(s), want 1", seen.reads)
	}
}

// TestGetRoleFor_RefusesRoleItMayNotChange verifies that a role without the
// content type is an error, and is left alone, when create_roles is off.
//
// Why it matters: with create_roles off the export must not edit catalog
// objects, and sending the role anyway would make Nautobot refuse the rack.
// Inputs: a cached role whose Nautobot record lists dcim.device only, with
// create_roles off.
// Outputs: an error naming the role and the content type, and no PATCH.
// Data choice: the same role as the passing case, so the option is the only
// difference between the two outcomes.
func TestGetRoleFor_RefusesRoleItMayNotChange(t *testing.T) {
	// Arrange.
	seen := &roleTraffic{}
	e, _, cleanup := roleExporter(t, []string{contentTypeDevice}, seen)
	defer cleanup()

	// Act.
	_, err := e.Cache.GetRoleFor("Storage", contentTypeRack)

	// Assert.
	if err == nil || !strings.Contains(err.Error(), "Storage") || !strings.Contains(err.Error(), contentTypeRack) {
		t.Errorf("error = %v, want one naming the role and %s", err, contentTypeRack)
	}
	if seen.patched != nil {
		t.Errorf("role was patched to %v, want no PATCH", seen.patched)
	}
}

// TestSetRackRole_SendsRoleRacksMayUse verifies that a rack's role reaches
// the create request when Nautobot's role already lists racks.
//
// Why it matters: `update rack --role` and a role imported from Nautobot set
// a field the export never sent, so a round trip lost every rack role.
// Inputs: a rack with role Storage and a role record listing dcim.rack.
// Outputs: a request whose role reference is the role's ID, nothing counted
// as skipped and no PATCH.
// Data choice: the role already lists racks, so the test covers the plain
// path without the content-type update.
func TestSetRackRole_SendsRoleRacksMayUse(t *testing.T) {
	// Arrange.
	seen := &roleTraffic{}
	e, id, cleanup := roleExporter(t, []string{contentTypeRack}, seen)
	defer cleanup()
	rack := &devicetypes.CaniRackType{Name: "x3000", ObjectMeta: devicetypes.ObjectMeta{Role: "Storage"}}
	req := &nautobotapi.WritableRackRequest{Name: rack.Name}
	result := &LoadResult{}

	// Act.
	e.setRackRole(req, rack, result)

	// Assert.
	if role := rackRequestRole(t, req); !strings.Contains(role, id.String()) {
		t.Errorf("request role = %s, want a reference to %s", role, id)
	}
	if result.RackRolesSkipped != 0 || seen.patched != nil {
		t.Errorf("RackRolesSkipped = %d, PATCH = %v, want 0 and none", result.RackRolesSkipped, seen.patched)
	}
}

// TestSetRackRole_LeavesUnusableRoleOffAndCountsIt verifies that a role
// racks may not use is left off the request and counted, and that a rack
// without a role asks Nautobot nothing.
//
// Why it matters: a refused role must cost the rack its role, not the rack
// and every device in it, and the loss must show in the summary.
// Inputs: a rack with role Storage against a role record without dcim.rack
// and create_roles off; then a rack with no role.
// Outputs: no role in either request, RackRolesSkipped == 1 and lossy count
// 1 after the first, and no further read for the second.
// Data choice: the roleless rack runs second so the read counter isolates
// its request traffic.
func TestSetRackRole_LeavesUnusableRoleOffAndCountsIt(t *testing.T) {
	// Arrange.
	seen := &roleTraffic{}
	e, _, cleanup := roleExporter(t, []string{contentTypeDevice}, seen)
	defer cleanup()
	result := &LoadResult{}
	withRole := &nautobotapi.WritableRackRequest{Name: "x3000"}
	withoutRole := &nautobotapi.WritableRackRequest{Name: "x3001"}

	// Act.
	e.setRackRole(withRole, &devicetypes.CaniRackType{Name: "x3000", ObjectMeta: devicetypes.ObjectMeta{Role: "Storage"}}, result)
	readsAfterFirst := seen.reads
	e.setRackRole(withoutRole, &devicetypes.CaniRackType{Name: "x3001"}, result)

	// Assert.
	if role := rackRequestRole(t, withRole); role != "" && role != "null" {
		t.Errorf("request role = %s, want none", role)
	}
	if role := rackRequestRole(t, withoutRole); role != "" && role != "null" {
		t.Errorf("roleless rack request role = %s, want none", role)
	}
	if result.RackRolesSkipped != 1 || lossyExportCount(result) != 1 {
		t.Errorf("RackRolesSkipped = %d, lossy = %d, want 1 and 1", result.RackRolesSkipped, lossyExportCount(result))
	}
	if seen.reads != readsAfterFirst {
		t.Errorf("roleless rack caused %d read(s), want none", seen.reads-readsAfterFirst)
	}
}
