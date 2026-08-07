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

func vlanMergeHandler(vlanID uuid.UUID, remoteJSON string, patchCalls *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "ipam/vlans/"+vlanID.String()) {
			if r.Method == http.MethodPatch {
				*patchCalls++
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, `{}`)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, remoteJSON)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
}

func TestMergeVLAN_PatchesWhenDrifted(t *testing.T) {
	vlanID := uuid.New()
	remoteJSON := fmt.Sprintf(`{"id":%q,"name":"old-name","vid":100,"status":{"id":"%s"}}`,
		vlanID, uuid.New())
	var patchCalls int
	e, cleanup := newExporterWithServer(t, vlanMergeHandler(vlanID, remoteJSON, &patchCalls))
	defer cleanup()
	e.Options.Merge = true

	vlan := &devicetypes.CaniVLAN{ID: uuid.New(), VID: 100, Name: "new-name"}

	updated, err := e.mergeVLAN(context.Background(), vlan, vlanID)
	if err != nil {
		t.Fatalf("mergeVLAN() error = %v", err)
	}
	if !updated {
		t.Error("expected updated=true when name drifted")
	}
	if patchCalls != 1 {
		t.Errorf("expected 1 PATCH, got %d", patchCalls)
	}
}

func TestMergeVLAN_SkipsWhenNoDrift(t *testing.T) {
	vlanID := uuid.New()
	remoteJSON := fmt.Sprintf(`{"id":%q,"name":"mgmt","vid":100,"status":{"id":"%s"}}`,
		vlanID, uuid.New())
	var patchCalls int
	e, cleanup := newExporterWithServer(t, vlanMergeHandler(vlanID, remoteJSON, &patchCalls))
	defer cleanup()
	e.Options.Merge = true

	vlan := &devicetypes.CaniVLAN{ID: uuid.New(), VID: 100, Name: "mgmt"}

	updated, err := e.mergeVLAN(context.Background(), vlan, vlanID)
	if err != nil {
		t.Fatalf("mergeVLAN() error = %v", err)
	}
	if updated {
		t.Error("expected updated=false when no drift")
	}
	if patchCalls != 0 {
		t.Errorf("expected 0 PATCH calls, got %d", patchCalls)
	}
}

func TestMergeVLAN_DryRunDoesNotPatch(t *testing.T) {
	vlanID := uuid.New()
	remoteJSON := fmt.Sprintf(`{"id":%q,"name":"old","vid":200,"status":{"id":"%s"}}`,
		vlanID, uuid.New())
	var patchCalls int
	e, cleanup := newExporterWithServer(t, vlanMergeHandler(vlanID, remoteJSON, &patchCalls))
	defer cleanup()
	e.Options.Merge = true
	e.Options.DryRun = true

	vlan := &devicetypes.CaniVLAN{ID: uuid.New(), VID: 200, Name: "new"}

	updated, err := e.mergeVLAN(context.Background(), vlan, vlanID)
	if err != nil {
		t.Fatalf("mergeVLAN() error = %v", err)
	}
	if !updated {
		t.Error("dry-run should still report updated=true")
	}
	if patchCalls != 0 {
		t.Errorf("dry-run should not PATCH, got %d", patchCalls)
	}
}

func loadVLANsHandler(vlanID uuid.UUID, existingJSON, retrieveJSON string, patchCalls *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "ipam/vlans/"+vlanID.String()):
			if r.Method == http.MethodPatch {
				*patchCalls++
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, `{}`)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, retrieveJSON)
		case strings.Contains(r.URL.Path, "ipam/vlans"):
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, existingJSON)
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		}
	}
}

func TestLoadVLANs_MergeUpdatesWhenDrifted(t *testing.T) {
	resetIPAMCaches()
	vlanID := uuid.New()
	existingJSON := fmt.Sprintf(`{"count":1,"results":[{"id":%q,"name":"old-mgmt","vid":100,"display":"old-mgmt"}]}`, vlanID)
	retrieveJSON := fmt.Sprintf(`{"id":%q,"name":"old-mgmt","vid":100,"status":{"id":"%s"}}`, vlanID, uuid.New())
	var patchCalls int
	e, cleanup := newExporterWithServer(t, loadVLANsHandler(vlanID, existingJSON, retrieveJSON, &patchCalls))
	defer cleanup()
	e.Options.Merge = true
	e.Options.DefaultLocation = "DC1"

	caniID := uuid.New()
	inv := &devicetypes.Inventory{
		VLANs: map[uuid.UUID]*devicetypes.CaniVLAN{
			caniID: {
				ID:   caniID,
				VID:  100,
				Name: "new-mgmt",
			},
		},
	}
	result := &LoadResult{}

	_, err := e.loadVLANs(context.Background(), inv, map[uuid.UUID]uuid.UUID{}, result)
	if err != nil {
		t.Fatalf("loadVLANs() error = %v", err)
	}
	if result.VLANsUpdated != 1 {
		t.Errorf("VLANsUpdated = %d, want 1", result.VLANsUpdated)
	}
	if result.VLANsSkipped != 0 {
		t.Errorf("VLANsSkipped = %d, want 0", result.VLANsSkipped)
	}
	if patchCalls != 1 {
		t.Errorf("expected 1 PATCH, got %d", patchCalls)
	}
}

func TestLoadVLANs_MergeSkipsWhenIdentical(t *testing.T) {
	resetIPAMCaches()
	vlanID := uuid.New()
	existingJSON := fmt.Sprintf(`{"count":1,"results":[{"id":%q,"name":"mgmt","vid":100,"display":"mgmt"}]}`, vlanID)
	retrieveJSON := fmt.Sprintf(`{"id":%q,"name":"mgmt","vid":100,"status":{"id":"%s"}}`, vlanID, uuid.New())
	var patchCalls int
	e, cleanup := newExporterWithServer(t, loadVLANsHandler(vlanID, existingJSON, retrieveJSON, &patchCalls))
	defer cleanup()
	e.Options.Merge = true
	e.Options.DefaultLocation = "DC1"

	caniID := uuid.New()
	inv := &devicetypes.Inventory{
		VLANs: map[uuid.UUID]*devicetypes.CaniVLAN{
			caniID: {
				ID:   caniID,
				VID:  100,
				Name: "mgmt",
			},
		},
	}
	result := &LoadResult{}

	_, err := e.loadVLANs(context.Background(), inv, map[uuid.UUID]uuid.UUID{}, result)
	if err != nil {
		t.Fatalf("loadVLANs() error = %v", err)
	}
	if result.VLANsUpdated != 0 {
		t.Errorf("VLANsUpdated = %d, want 0", result.VLANsUpdated)
	}
	if result.VLANsSkipped != 1 {
		t.Errorf("VLANsSkipped = %d, want 1", result.VLANsSkipped)
	}
	if patchCalls != 0 {
		t.Errorf("expected 0 PATCH calls, got %d", patchCalls)
	}
}

func TestLoadVLANs_MergeDryRunIssuesNoMutations(t *testing.T) {
	resetIPAMCaches()
	vlanID := uuid.New()
	existingJSON := fmt.Sprintf(`{"count":1,"results":[{"id":%q,"name":"old-mgmt","vid":100,"display":"old-mgmt"}]}`, vlanID)
	retrieveJSON := fmt.Sprintf(`{"id":%q,"name":"old-mgmt","vid":100,"status":{"id":"%s"}}`, vlanID, uuid.New())
	var patchCalls int
	e, cleanup := newExporterWithServer(t, loadVLANsHandler(vlanID, existingJSON, retrieveJSON, &patchCalls))
	defer cleanup()
	e.Options.Merge = true
	e.Options.DryRun = true
	e.Options.DefaultLocation = "DC1"

	caniID := uuid.New()
	inv := &devicetypes.Inventory{
		VLANs: map[uuid.UUID]*devicetypes.CaniVLAN{
			caniID: {
				ID:   caniID,
				VID:  100,
				Name: "new-mgmt",
			},
		},
	}
	result := &LoadResult{}

	_, err := e.loadVLANs(context.Background(), inv, map[uuid.UUID]uuid.UUID{}, result)
	if err != nil {
		t.Fatalf("loadVLANs() error = %v", err)
	}
	if patchCalls != 0 {
		t.Errorf("dry-run should issue 0 PATCH calls, got %d", patchCalls)
	}
	if result.VLANsUpdated != 1 {
		t.Errorf("VLANsUpdated = %d, want 1 (reported even in dry-run)", result.VLANsUpdated)
	}
}
