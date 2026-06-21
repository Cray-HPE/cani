/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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

// sentVLAN is the minimal view of the VLANRequest body POSTed to /ipam/vlans/.
type sentVLAN struct {
	Vid         int       `json:"vid"`
	Name        string    `json:"name"`
	Status      wireIDRef `json:"status"`
	Description string    `json:"description"`
}

// -----------------------------------------------------------------------------
// loadVLANs — happy path (PASSING scenario)
// -----------------------------------------------------------------------------

// TestLoadVLANs_SendsCorrectPayloadAndReturnsMapping verifies a new VLAN is
// POSTed with vid/name/status/description mapped correctly and that the returned
// mapping plus the inventory's ExternalIDs capture Nautobot's assigned ID.
//
// Why it matters: VLANs are a prerequisite for IPAM export; downstream prefixes
// and IP addresses reference the VLAN's Nautobot ID, so the create payload and
// the recorded remote ID must both be exact.
// Inputs: a context, an Inventory holding one CaniVLAN, an empty ID map, and a
// LoadResult. Outputs: a cani->Nautobot VLAN ID map and an error; side effects
// are the POST body, result counters, and ExternalIDs.
// Data choice: VID 100 / "vlan100" with status "Active" is a minimal realistic
// VLAN; the handler returns a fresh remote ID so the test can assert the mapping
// and ExternalIDs were populated from the response.
func TestLoadVLANs_SendsCorrectPayloadAndReturnsMapping(t *testing.T) {
	// Arrange.
	resetIPAMCaches()

	vlanNID := uuid.New() // remote ID assigned to the created VLAN

	var posts [][]byte
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !strings.Contains(r.URL.Path, "ipam/vlans") {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
			return
		}
		if r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			posts = append(posts, body)
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w,
				`{"id":"`+vlanNID.String()+`","vid":100,"name":"vlan100","display":"vlan100"}`)
			return
		}
		// GET lookup -> not found.
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	statusID := seedActiveStatus(t, e)

	vlanID := uuid.New()
	inv := &devicetypes.Inventory{
		VLANs: map[uuid.UUID]*devicetypes.CaniVLAN{
			vlanID: {
				ID:          vlanID,
				VID:         100,
				Name:        "vlan100",
				Description: "prod",
				ObjectMeta:  devicetypes.ObjectMeta{Status: "Active"},
			},
		},
	}
	result := &LoadResult{}

	// Act.
	mapping, err := e.loadVLANs(context.Background(), inv, map[uuid.UUID]uuid.UUID{}, result)

	// Assert.
	if err != nil {
		t.Fatalf("loadVLANs returned error: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 VLAN create POST, got %d", len(posts))
	}

	vlan := decodeSentVLAN(t, posts[0])
	if vlan.Vid != 100 {
		t.Errorf("vid = %d, want 100", vlan.Vid)
	}
	if vlan.Name != "vlan100" {
		t.Errorf("name = %q, want vlan100", vlan.Name)
	}
	if vlan.Status.ID != statusID.String() {
		t.Errorf("status.id = %q, want %q", vlan.Status.ID, statusID)
	}
	if vlan.Description != "prod" {
		t.Errorf("description = %q, want prod", vlan.Description)
	}

	if result.VLANsCreated != 1 {
		t.Errorf("VLANsCreated = %d, want 1", result.VLANsCreated)
	}
	if mapping[vlanID] != vlanNID {
		t.Errorf("mapping[vlanID] = %s, want %s", mapping[vlanID], vlanNID)
	}
	if got := inv.VLANs[vlanID].ExternalIDs[externalIDKeyNautobot]; got != vlanNID {
		t.Errorf("ExternalIDs[nautobot] = %s, want %s", got, vlanNID)
	}
}

// -----------------------------------------------------------------------------
// loadVLANs — idempotency: an existing VLAN is skipped (negative scenario)
// -----------------------------------------------------------------------------

// TestLoadVLANs_SkipsExistingVLAN verifies that when the cache already resolves
// a VLAN, no POST is issued, it is counted as skipped, and its remote ID still
// flows into the returned mapping.
//
// Why it matters: exports are re-run repeatedly; idempotency prevents duplicate
// VLANs in Nautobot while still letting dependent IPAM records resolve the VLAN.
// Inputs: a context, an Inventory with one already-cached CaniVLAN, an empty ID
// map, and a LoadResult. Outputs: the VLAN ID mapping and an error; the handler
// fails the test if any POST occurs.
// Data choice: the cache is pre-seeded with VID 100 keyed by an empty location
// (matching the empty DefaultLocation) so the lookup short-circuits exactly as
// it would on a second export run.
func TestLoadVLANs_SkipsExistingVLAN(t *testing.T) {
	// Arrange.
	resetIPAMCaches()

	existingNID := uuid.New()
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			t.Errorf("unexpected POST to %s for an existing VLAN", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	// Seed the cache so the lookup resolves without an API round-trip.
	// DefaultLocation is empty, so the cache key location segment is "".
	e.Cache.CacheVLAN(100, "", &CachedItem{ID: existingNID, Name: "vlan100"})

	vlanID := uuid.New()
	inv := &devicetypes.Inventory{
		VLANs: map[uuid.UUID]*devicetypes.CaniVLAN{
			vlanID: {
				ID:         vlanID,
				VID:        100,
				Name:       "vlan100",
				ObjectMeta: devicetypes.ObjectMeta{Status: "Active"},
			},
		},
	}
	result := &LoadResult{}

	// Act.
	mapping, err := e.loadVLANs(context.Background(), inv, map[uuid.UUID]uuid.UUID{}, result)

	// Assert.
	if err != nil {
		t.Fatalf("loadVLANs returned error: %v", err)
	}
	if result.VLANsSkipped != 1 {
		t.Errorf("VLANsSkipped = %d, want 1", result.VLANsSkipped)
	}
	if result.VLANsCreated != 0 {
		t.Errorf("VLANsCreated = %d, want 0", result.VLANsCreated)
	}
	if mapping[vlanID] != existingNID {
		t.Errorf("mapping[vlanID] = %s, want existing remote ID %s", mapping[vlanID], existingNID)
	}
}

func decodeSentVLAN(t *testing.T, body []byte) sentVLAN {
	t.Helper()
	var v sentVLAN
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("decode VLAN payload: %v\nbody: %s", err, body)
	}
	return v
}
