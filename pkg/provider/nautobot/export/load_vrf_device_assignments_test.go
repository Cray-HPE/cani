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

// vrfDeviceAssignmentServer answers assignment listings with nothing, accepts
// assignment creates except for the refused device, and records each body.
func vrfDeviceAssignmentServer(t *testing.T, refusedNID uuid.UUID, posted *[]map[string]any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !strings.Contains(r.URL.Path, "vrf-device-assignments") || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding assignment body: %v", err)
		}
		*posted = append(*posted, body)
		if strings.Contains(stringify(body["device"]), refusedNID.String()) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, `{"detail":"refused"}`)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":"`+uuid.New().String()+`"}`)
	}
}

// TestLoadVRFDeviceAssignments_AssignsExportedDevicesAndReportsTheRest
// verifies the devices a VRF lists are each assigned, skipped or reported
// according to what the inventory and Nautobot know about them.
//
// Why it matters: a VRF's device list is operator-authored; a device that was
// never exported, or that Nautobot refuses, must be counted and named rather
// than fail the VRF phase or vanish from the summary.
// Inputs: VRF blue listing four devices: one absent from the inventory, one
// without a Nautobot ID, one exported, one exported but refused by the server
// with a 500 on create; the server lists no existing assignments.
// Outputs: one POST carrying the exported device's and the VRF's Nautobot
// IDs; created=1, skipped=2, one error naming the refused device.
// Data choice: each outcome has its own device so the counters distinguish
// the four branches.
func TestLoadVRFDeviceAssignments_AssignsExportedDevicesAndReportsTheRest(t *testing.T) {
	// Arrange.
	vrfNID, exportedNID, refusedNID := uuid.New(), uuid.New(), uuid.New()
	var posted []map[string]any
	e, cleanup := newExporterWithServer(t, vrfDeviceAssignmentServer(t, refusedNID, &posted))
	defer cleanup()

	missing, unexported, exported, refused := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv := devicetypes.NewInventory()
	inv.Devices[unexported] = &devicetypes.CaniDeviceType{ID: unexported, Name: "staged"}
	inv.Devices[exported] = &devicetypes.CaniDeviceType{ID: exported, Name: "node01",
		ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{externalIDKeyNautobot: exportedNID}}}
	inv.Devices[refused] = &devicetypes.CaniDeviceType{ID: refused, Name: "node02",
		ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{externalIDKeyNautobot: refusedNID}}}
	vrf := &devicetypes.CaniVRF{ID: uuid.New(), Name: "blue", Devices: []uuid.UUID{missing, unexported, exported, refused}}
	result := &LoadResult{}

	// Act.
	e.loadVRFDeviceAssignments(context.Background(), vrf, &CachedItem{ID: vrfNID, Name: "blue"}, inv, result)

	// Assert.
	counts := [3]int{result.VRFDeviceAssignmentsCreated, result.VRFDeviceAssignmentsSkipped, len(result.Errors)}
	if counts != [3]int{1, 2, 1} {
		t.Fatalf("created, skipped, errors = %v; want [1 2 1] (%v)", counts, result.Errors)
	}
	assertContainsAll(t, "error", result.Errors[0], "node02", "500")
	if len(posted) != 2 {
		t.Fatalf("assignment POSTs = %d, want 2 (the exported and the refused device)", len(posted))
	}
	assertContainsAll(t, "first POST device", stringify(posted[0]["device"]), exportedNID.String())
	assertContainsAll(t, "first POST vrf", stringify(posted[0]["vrf"]), vrfNID.String())
}

// TestLoadVRFDeviceAssignments_DryRunMakesNoRequest verifies a dry run counts
// the assignment it would create without touching Nautobot.
//
// Why it matters: the dry-run summary must show the assignments a real run
// would add, and a dry run must issue no request for them.
// Inputs: DryRun set; a VRF listing one exported device; a server that fails
// the test on any request.
// Outputs: created=1, skipped=0, no errors, no request.
// Data choice: an exported device is the only one that would reach Nautobot.
func TestLoadVRFDeviceAssignments_DryRunMakesNoRequest(t *testing.T) {
	// Arrange.
	e, cleanup := newExporterWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("dry run sent %s %s", r.Method, r.URL.Path)
	})
	defer cleanup()
	e.Options.DryRun = true
	exported := uuid.New()
	inv := devicetypes.NewInventory()
	inv.Devices[exported] = &devicetypes.CaniDeviceType{ID: exported, Name: "node01",
		ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{externalIDKeyNautobot: uuid.New()}}}
	vrf := &devicetypes.CaniVRF{ID: uuid.New(), Name: "blue", Devices: []uuid.UUID{exported}}
	result := &LoadResult{}

	// Act.
	e.loadVRFDeviceAssignments(context.Background(), vrf, &CachedItem{ID: uuid.New(), Name: "blue"}, inv, result)

	// Assert.
	if result.VRFDeviceAssignmentsCreated != 1 || result.VRFDeviceAssignmentsSkipped != 0 || len(result.Errors) != 0 {
		t.Errorf("created = %d, skipped = %d, errors = %v; want 1, 0, none",
			result.VRFDeviceAssignmentsCreated, result.VRFDeviceAssignmentsSkipped, result.Errors)
	}
}

// stringify renders a decoded JSON value so a test can look for an ID inside
// whichever reference shape the client sent.
func stringify(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// assertContainsAll fails the test for each want absent from text.
func assertContainsAll(t *testing.T, label, text string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(text, want) {
			t.Errorf("%s = %q, want it to contain %q", label, text, want)
		}
	}
}
