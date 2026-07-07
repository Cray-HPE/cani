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

// resetRelationshipCache clears the package-level relationship-definition cache
// so each test resolves definitions from a known-empty state.
func resetRelationshipCache() {
	relationshipCacheMu.Lock()
	for k := range relationshipCache {
		delete(relationshipCache, k)
	}
	relationshipCacheMu.Unlock()
}

// relationshipServer returns a handler that find-or-creates a relationship
// definition (empty list -> POST returns relID) and records association POSTs.
// The association dedupe list is always empty so associations are created.
func relationshipServer(relKey string, relID uuid.UUID, assocPosted *bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "relationship-associations"):
			if r.Method == http.MethodPost {
				*assocPosted = true
				w.WriteHeader(http.StatusCreated)
				_, _ = io.WriteString(w, fmt.Sprintf(`{"id":%q}`, uuid.New().String()))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		case strings.Contains(r.URL.Path, "extras/relationships"):
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
				_, _ = io.WriteString(w, fmt.Sprintf(`{"id":%q,"key":%q}`, relID.String(), relKey))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		}
	}
}

// TestLoadRelationships_CreatesAssignedVLANs verifies a device's assigned VLANs
// become assigned_vlans associations, creating the relationship definition.
//
// Why it matters: the assigned_vlans relationship records which VLANs are
// configured on a switch; without it, downstream fabric tooling cannot tell
// which networks a switch carries.
// Inputs: a device with one assigned VLAN, its Nautobot ID set, and a VID->VLAN
// map; the server has no existing definition or association. Outputs:
// RelationshipsCreated=1 and an association POST.
// Data choice: a single leaf with one VLAN isolates the create-and-associate path.
func TestLoadRelationships_CreatesAssignedVLANs(t *testing.T) {
	resetRelationshipCache()
	relID := uuid.New()
	assocPosted := false
	e, cleanup := newExporterWithServer(t, relationshipServer(relKeyAssignedVLANs, relID, &assocPosted))
	defer cleanup()

	deviceID, deviceNID := uuid.New(), uuid.New()
	vlanCaniID, vlanNID := uuid.New(), uuid.New()
	device := &devicetypes.CaniDeviceType{ID: deviceID, Name: "leaf-1", AssignedVLANs: []uuid.UUID{vlanCaniID}}
	device.ExternalIDs = map[string]uuid.UUID{externalIDKeyNautobot: deviceNID}
	inv := &devicetypes.Inventory{Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{deviceID: device}}

	result := &LoadResult{}
	if err := e.loadRelationships(context.Background(), inv, map[uuid.UUID]uuid.UUID{vlanCaniID: vlanNID}, result); err != nil {
		t.Fatalf("loadRelationships: %v", err)
	}
	if result.RelationshipsCreated != 1 {
		t.Errorf("RelationshipsCreated = %d, want 1", result.RelationshipsCreated)
	}
	if !assocPosted {
		t.Error("expected an association POST")
	}
}

// TestLoadRelationships_CreatesBMCDevice verifies a BMC device's parent link
// becomes a bmc_device association.
//
// Why it matters: the bmc_device relationship ties a BMC (iLO) to the node it
// manages; losing it breaks out-of-band tooling that navigates from node to BMC.
// Inputs: a BMC device whose BMCParent points at a parent device, both with
// Nautobot IDs set, and no existing definition/association. Outputs:
// RelationshipsCreated=1 and an association POST.
// Data choice: a compute node and its iLO mirror the canonical BMC pairing.
func TestLoadRelationships_CreatesBMCDevice(t *testing.T) {
	resetRelationshipCache()
	relID := uuid.New()
	assocPosted := false
	e, cleanup := newExporterWithServer(t, relationshipServer(relKeyBMCDevice, relID, &assocPosted))
	defer cleanup()

	bmcID, bmcNID := uuid.New(), uuid.New()
	parentID, parentNID := uuid.New(), uuid.New()
	bmc := &devicetypes.CaniDeviceType{ID: bmcID, Name: "compute01-bmc", BMCParent: parentID}
	bmc.ExternalIDs = map[string]uuid.UUID{externalIDKeyNautobot: bmcNID}
	parent := &devicetypes.CaniDeviceType{ID: parentID, Name: "compute01"}
	parent.ExternalIDs = map[string]uuid.UUID{externalIDKeyNautobot: parentNID}
	inv := &devicetypes.Inventory{Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{bmcID: bmc, parentID: parent}}

	result := &LoadResult{}
	if err := e.loadRelationships(context.Background(), inv, map[uuid.UUID]uuid.UUID{}, result); err != nil {
		t.Fatalf("loadRelationships: %v", err)
	}
	if result.RelationshipsCreated != 1 {
		t.Errorf("RelationshipsCreated = %d, want 1", result.RelationshipsCreated)
	}
	if !assocPosted {
		t.Error("expected an association POST")
	}
}

// TestCreateAssociation_SkipsExisting verifies an association already present in
// Nautobot is not recreated.
//
// Why it matters: re-running an export must be idempotent; recreating an
// association would duplicate it (for many-to-many) or error (for one-to-one).
// Inputs: a server whose association list returns one match and which fails the
// test on any association POST. Outputs: RelationshipsSkipped=1, none created.
// Data choice: erroring on POST makes the no-duplicate guarantee explicit.
func TestCreateAssociation_SkipsExisting(t *testing.T) {
	resetRelationshipCache()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "relationship-associations") {
			if r.Method == http.MethodPost {
				t.Error("unexpected association POST: it already exists")
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

	result := &LoadResult{}
	e.createAssociation(context.Background(), relKeyAssignedVLANs, uuid.New(), assocEndpoints{
		srcType: contentTypeDevice, srcID: uuid.New(),
		dstType: contentTypeVLAN, dstID: uuid.New(),
	}, result)
	if result.RelationshipsSkipped != 1 {
		t.Errorf("RelationshipsSkipped = %d, want 1", result.RelationshipsSkipped)
	}
	if result.RelationshipsCreated != 0 {
		t.Errorf("RelationshipsCreated = %d, want 0", result.RelationshipsCreated)
	}
}
