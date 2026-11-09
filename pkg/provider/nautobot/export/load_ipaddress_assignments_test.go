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

// assignmentFixture is an address that already exists in Nautobot and is
// assigned, locally, to two ports of one device; Nautobot already holds the
// assignment of the first.
type assignmentFixture struct {
	ipNID, deviceNID, ifaceANID, ifaceBNID uuid.UUID
	inv                                    *devicetypes.Inventory
	deviceMap                              map[string]uuid.UUID
	assignPosts                            *int
	assignStatus                           int
}

// existingAddressServer answers the address lookup with an existing address
// and the assignment list with one row (interface A); it counts and answers
// assignment POSTs with assignStatus.
func existingAddressServer(f *assignmentFixture) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "ip-address-to-interface"):
			if r.Method == http.MethodPost {
				*f.assignPosts++
				w.WriteHeader(f.assignStatus)
				_, _ = io.WriteString(w, `{"interface":["Must associate to either an Interface or a VMInterface."]}`)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"count":1,"results":[{"id":"%s","ip_address":{"id":"%s"},"interface":{"id":"%s"}}]}`,
				uuid.New(), f.ipNID, f.ifaceANID)
		case strings.Contains(r.URL.Path, "ip-addresses"):
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"count":1,"results":[{"id":"%s","address":"10.0.1.5/24"}]}`, f.ipNID)
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		}
	}
}

// newAssignmentFixture wires an exporter to existingAddressServer with the
// device and both interfaces pre-cached, so only the assignment calls hit
// the fake.
func newAssignmentFixture(t *testing.T, assignStatus int) (*Exporter, *assignmentFixture, func()) {
	t.Helper()
	f := &assignmentFixture{
		ipNID: uuid.New(), deviceNID: uuid.New(), ifaceANID: uuid.New(), ifaceBNID: uuid.New(),
		assignPosts: new(int), assignStatus: assignStatus,
	}
	e, cleanup := newExporterWithServer(t, existingAddressServer(f))
	seedActiveStatus(t, e)
	seedGlobalNamespace(e, uuid.New())
	e.Cache.CacheInterface(f.deviceNID, "eth0", &CachedItem{ID: f.ifaceANID, Name: "eth0"})
	e.Cache.CacheInterface(f.deviceNID, "eth1", &CachedItem{ID: f.ifaceBNID, Name: "eth1"})

	deviceID, ifaceA, ifaceB, prefixID, ipID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	f.inv = &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{deviceID: {ID: deviceID, Name: "server-1"}},
		Interfaces: map[uuid.UUID]*devicetypes.CaniInterface{
			ifaceA: {ID: ifaceA, Name: "eth0", DeviceID: deviceID},
			ifaceB: {ID: ifaceB, Name: "eth1", DeviceID: deviceID},
		},
		Prefixes: map[uuid.UUID]*devicetypes.CaniPrefix{prefixID: {ID: prefixID, Prefix: "10.0.1.0/24"}},
		IPAddresses: map[uuid.UUID]*devicetypes.CaniIPAddress{ipID: {
			ID: ipID, Address: "10.0.1.5/24", Parent: prefixID, Interfaces: []uuid.UUID{ifaceA, ifaceB},
			ObjectMeta: devicetypes.ObjectMeta{Status: "Active"},
		}},
	}
	f.deviceMap = map[string]uuid.UUID{"server-1": f.deviceNID}
	return e, f, cleanup
}

// TestLoadIPAddresses_ReExportCreatesOnlyTheMissingAssignment verifies that
// an address Nautobot already holds is read for its assignments first: the
// one already present is skipped, the missing one is created.
//
// Why it matters: the assignment table is not change-logged, so a duplicate
// POST on every re-export would never show in the round-trip ledger; reading
// first is what makes a repeated export duplicate nothing.
// Inputs: an existing address assigned remotely to eth0 and locally to eth0
// and eth1.
// Outputs: one assignment POST, IPAssignmentsSkipped == 1,
// IPAssignmentsCreated == 1, no errors.
// Data choice: one present and one missing port on the same address make
// "skip all" and "create all" both fail the assertions.
func TestLoadIPAddresses_ReExportCreatesOnlyTheMissingAssignment(t *testing.T) {
	e, f, cleanup := newAssignmentFixture(t, http.StatusCreated)
	defer cleanup()
	result := &LoadResult{}

	if _, err := e.loadIPAddresses(context.Background(), f.inv, map[uuid.UUID]uuid.UUID{}, f.deviceMap, result); err != nil {
		t.Fatalf("loadIPAddresses() error = %v", err)
	}

	if *f.assignPosts != 1 {
		t.Errorf("assignment POSTs = %d, want 1 (eth1 only)", *f.assignPosts)
	}
	if result.IPAssignmentsSkipped != 1 || result.IPAssignmentsCreated != 1 {
		t.Errorf("skipped/created = %d/%d, want 1/1", result.IPAssignmentsSkipped, result.IPAssignmentsCreated)
	}
	if len(result.Errors) != 0 {
		t.Errorf("Errors = %v, want none", result.Errors)
	}
}

// TestLoadIPAddresses_DryRunReportsOnlyTheMissingAssignment verifies that a
// dry run against an existing address consults the same list and plans
// only the assignment Nautobot lacks, without a POST.
//
// Why it matters: the preview must say what the real run would do; an
// existing assignment is not "would assign", and nothing may be written.
// Inputs: the fixture with DryRun set.
// Outputs: zero POSTs, IPAssignmentsSkipped == 1, IPAssignmentsCreated == 1
// (the planned one).
// Data choice: same fixture as the real run so the two counts can be
// compared directly.
func TestLoadIPAddresses_DryRunReportsOnlyTheMissingAssignment(t *testing.T) {
	e, f, cleanup := newAssignmentFixture(t, http.StatusCreated)
	defer cleanup()
	e.Options.DryRun = true
	result := &LoadResult{}

	if _, err := e.loadIPAddresses(context.Background(), f.inv, map[uuid.UUID]uuid.UUID{}, f.deviceMap, result); err != nil {
		t.Fatalf("loadIPAddresses() error = %v", err)
	}

	if *f.assignPosts != 0 {
		t.Errorf("assignment POSTs = %d in a dry run, want 0", *f.assignPosts)
	}
	if result.IPAssignmentsSkipped != 1 || result.IPAssignmentsCreated != 1 {
		t.Errorf("skipped/planned = %d/%d, want 1/1", result.IPAssignmentsSkipped, result.IPAssignmentsCreated)
	}
}

// TestLoadIPAddresses_RejectedAssignmentIsAnError verifies that a 400 from
// the assignment POST is recorded as an error rather than read as "already
// assigned".
//
// Why it matters: before the read-first check every 400 was taken for an
// existing assignment, which hid malformed references and real refusals.
// Inputs: the fixture with the POST answering 400.
// Outputs: one error naming the address and port; IPAssignmentsCreated == 0.
// Data choice: the body is Nautobot's wording for a join row without a valid
// interface, the case the old shortcut misread.
func TestLoadIPAddresses_RejectedAssignmentIsAnError(t *testing.T) {
	e, f, cleanup := newAssignmentFixture(t, http.StatusBadRequest)
	defer cleanup()
	result := &LoadResult{}

	if _, err := e.loadIPAddresses(context.Background(), f.inv, map[uuid.UUID]uuid.UUID{}, f.deviceMap, result); err != nil {
		t.Fatalf("loadIPAddresses() error = %v", err)
	}

	if result.IPAssignmentsCreated != 0 {
		t.Errorf("IPAssignmentsCreated = %d, want 0", result.IPAssignmentsCreated)
	}
	if len(result.Errors) != 1 || !strings.Contains(result.Errors[0], "server-1:eth1") || !strings.Contains(result.Errors[0], "400") {
		t.Errorf("Errors = %v, want one naming server-1:eth1 with status 400", result.Errors)
	}
}
