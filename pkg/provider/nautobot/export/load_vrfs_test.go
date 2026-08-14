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
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestLoadVRFs_CreatesWhenMissing verifies a VRF absent from Nautobot is created
// and cached by name.
//
// Why it matters: interface VRF assignment resolves VRFs by name from the cache;
// a VRF that fails to create (or cache) would silently drop the isolation an
// interface depends on (e.g. keepalive, LEGACY).
// Inputs: an inventory with one VRF "LEGACY" and a server whose VRF list is empty
// and whose create returns 201. Outputs: VRFsCreated=1, a POST issued, and the
// VRF cached under its name with the new ID.
// Data choice: "LEGACY" with an RD mirrors the real legacy CSM routing VRF.
func TestLoadVRFs_CreatesWhenMissing(t *testing.T) {
	resetIPAMCaches()
	vrfNID := uuid.New()
	posted := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "ipam/vrfs") {
			if r.Method == http.MethodPost {
				posted = true
				w.WriteHeader(http.StatusCreated)
				_, _ = io.WriteString(w, fmt.Sprintf(`{"id":%q,"name":"LEGACY"}`, vrfNID.String()))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedGlobalNamespace(uuid.New())

	vrfID := uuid.New()
	inv := &devicetypes.Inventory{
		VRFs: map[uuid.UUID]*devicetypes.CaniVRF{
			vrfID: {ID: vrfID, Name: "LEGACY", RD: "65000:2000"},
		},
	}
	result := &LoadResult{}
	if err := e.loadVRFs(context.Background(), inv, result); err != nil {
		t.Fatalf("loadVRFs: %v", err)
	}
	if result.VRFsCreated != 1 {
		t.Errorf("VRFsCreated = %d, want 1", result.VRFsCreated)
	}
	if !posted {
		t.Error("expected a VRF create POST")
	}
	item, ok := e.Cache.LookupCachedVRF("LEGACY")
	if !ok || item.ID != vrfNID {
		t.Errorf("cached VRF = %v, want id %s", item, vrfNID)
	}
}

// TestLoadVRFs_SkipsExisting verifies an already-present VRF is reused, not
// recreated.
//
// Why it matters: re-running an export must be idempotent; recreating a VRF would
// either duplicate it or error, breaking repeat syncs.
// Inputs: a server whose VRF list returns one match and which fails the test on
// any POST. Outputs: VRFsSkipped=1 and the existing VRF cached by name.
// Data choice: erroring on POST makes the no-recreate guarantee explicit.
func TestLoadVRFs_SkipsExisting(t *testing.T) {
	resetIPAMCaches()
	vrfNID := uuid.New()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "ipam/vrfs") {
			if r.Method == http.MethodPost {
				t.Error("unexpected POST: VRF already exists and must not be recreated")
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, fmt.Sprintf(`{"count":1,"results":[{"id":%q,"name":"LEGACY"}]}`, vrfNID.String()))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	vrfID := uuid.New()
	inv := &devicetypes.Inventory{
		VRFs: map[uuid.UUID]*devicetypes.CaniVRF{
			vrfID: {ID: vrfID, Name: "LEGACY"},
		},
	}
	result := &LoadResult{}
	if err := e.loadVRFs(context.Background(), inv, result); err != nil {
		t.Fatalf("loadVRFs: %v", err)
	}
	if result.VRFsSkipped != 1 {
		t.Errorf("VRFsSkipped = %d, want 1", result.VRFsSkipped)
	}
	item, ok := e.Cache.LookupCachedVRF("LEGACY")
	if !ok || item.ID != vrfNID {
		t.Errorf("cached VRF = %v, want id %s", item, vrfNID)
	}
}

// TestBuildInterfaceEnrichment_AssignsVRF verifies a cached VRF is attached to
// the interface enrichment PATCH.
//
// Why it matters: placing an interface in a VRF (e.g. the LEGACY upstream or the
// keepalive link) is how per-tenant routing isolation reaches Nautobot; dropping
// it would merge the interface into the default routing table.
// Inputs: an interface spec naming VRF "LEGACY" with that VRF pre-cached.
// Outputs: changed=true and a payload containing the VRF UUID.
// Data choice: a pre-cached VRF isolates the enrichment wiring from VRF creation.
func TestBuildInterfaceEnrichment_AssignsVRF(t *testing.T) {
	resetIPAMCaches()
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, `{}`))
	defer cleanup()

	vrfNID := uuid.New()
	e.Cache.CacheVRF("LEGACY", &CachedItem{ID: vrfNID, Name: "LEGACY"})

	deviceID := uuid.New()
	spec := interfaceSpec{Name: "1/1/25", VRF: "LEGACY"}

	req, changed, unresolved := e.buildInterfaceEnrichment(deviceID, spec, map[int]uuid.UUID{})
	if !changed {
		t.Fatal("buildInterfaceEnrichment: changed = false, want true")
	}
	if unresolved != 0 {
		t.Errorf("buildInterfaceEnrichment: unresolved = %d, want 0", unresolved)
	}
	blob, _ := json.Marshal(req)
	if !strings.Contains(string(blob), vrfNID.String()) {
		t.Errorf("enrichment payload missing vrf id %s:\n%s", vrfNID, blob)
	}
}

// TestEnsureVRFDeviceAssignment_CreatesWhenMissing verifies a VRF not yet
// assigned to a device is linked via a vrf-device-assignment create.
//
// Why it matters: Nautobot rejects an interface VRF assignment ("VRF must be
// assigned to same Device") unless the VRF↔device link exists first; creating it
// on demand is what lets interface VRF enrichment succeed.
// Inputs: a cached VRF "LEGACY", a device UUID, and a server whose assignment
// list is empty and whose create returns 201. Outputs: no error and exactly one
// assignment POST.
// Data choice: an empty list drives the create branch; a POST flag proves the
// link is created.
func TestEnsureVRFDeviceAssignment_CreatesWhenMissing(t *testing.T) {
	resetIPAMCaches()
	posted := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "vrf-device-assignments") {
			if r.Method == http.MethodPost {
				posted = true
				w.WriteHeader(http.StatusCreated)
				_, _ = io.WriteString(w, fmt.Sprintf(`{"id":%q}`, uuid.New().String()))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	e.Cache.CacheVRF("LEGACY", &CachedItem{ID: uuid.New(), Name: "LEGACY"})

	if err := e.ensureVRFDeviceAssignment(context.Background(), uuid.New(), "LEGACY"); err != nil {
		t.Fatalf("ensureVRFDeviceAssignment() error = %v", err)
	}
	if !posted {
		t.Error("expected a vrf-device-assignment create POST")
	}
}

// TestEnsureVRFDeviceAssignment_SkipsWhenPresent verifies an existing VRF↔device
// link is reused rather than duplicated.
//
// Why it matters: re-running an export must be idempotent; re-posting an existing
// assignment would error or duplicate the link.
// Inputs: a cached VRF and a server whose assignment list returns one match and
// which fails the test on any POST. Outputs: no error and no POST.
// Data choice: erroring on POST makes the no-duplicate guarantee explicit.
func TestEnsureVRFDeviceAssignment_SkipsWhenPresent(t *testing.T) {
	resetIPAMCaches()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "vrf-device-assignments") {
			if r.Method == http.MethodPost {
				t.Error("unexpected POST: assignment already exists and must not be recreated")
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, fmt.Sprintf(`{"count":1,"results":[{"id":%q}]}`, uuid.New().String()))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	e.Cache.CacheVRF("LEGACY", &CachedItem{ID: uuid.New(), Name: "LEGACY"})

	if err := e.ensureVRFDeviceAssignment(context.Background(), uuid.New(), "LEGACY"); err != nil {
		t.Fatalf("ensureVRFDeviceAssignment() error = %v", err)
	}
}
