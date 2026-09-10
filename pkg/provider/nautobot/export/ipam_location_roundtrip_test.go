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
	"net/http/httptest"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	imprt "github.com/Cray-HPE/cani/pkg/provider/nautobot/import"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/transform"
	"github.com/google/uuid"
)

type scopedIPAMRequest struct {
	Location *wireIDRef `json:"location"`
	VLAN     *wireIDRef `json:"vlan"`
}

func TestScopedIPAMImportExportRoundTrip(t *testing.T) {
	resetIPAMCaches()
	t.Cleanup(resetIPAMCaches)

	sourceLocationID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sourceVLANID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	sourcePrefixID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	sourceStatusID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var results any
		switch r.URL.Path {
		case "/dcim/locations/":
			results = []map[string]any{{
				"id": sourceLocationID, "name": "site-a",
				"location_type": map[string]any{"id": uuid.New()},
				"status":        map[string]any{"id": sourceStatusID},
			}}
		case "/extras/statuses/":
			results = []map[string]any{{
				"id": sourceStatusID, "name": "Active", "content_types": []string{"ipam.vlan", "ipam.prefix"},
			}}
		case "/ipam/vlans/":
			results = []map[string]any{{
				"id": sourceVLANID, "vid": 100, "name": "production",
				"status": map[string]any{"id": sourceStatusID},
			}}
		case "/ipam/vlan-location-assignments/":
			results = []map[string]any{{
				"location": map[string]any{"id": sourceLocationID},
				"vlan":     map[string]any{"id": sourceVLANID},
			}}
		case "/ipam/prefixes/":
			results = []map[string]any{{
				"id": sourcePrefixID, "prefix": "10.0.0.0/24", "prefix_length": 24, "ip_version": 4,
				"status": map[string]any{"id": sourceStatusID},
				"vlan":   map[string]any{"id": sourceVLANID},
			}}
		case "/ipam/prefix-location-assignments/":
			results = []map[string]any{{
				"location": map[string]any{"id": sourceLocationID},
				"prefix":   map[string]any{"id": sourcePrefixID},
			}}
		default:
			t.Errorf("unexpected import request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"count": len(results.([]map[string]any)), "next": nil, "previous": nil, "results": results,
		})
	}))
	defer source.Close()

	sourceClient, err := nautobotapi.NewClientWithResponses(source.URL)
	if err != nil {
		t.Fatalf("create import client: %v", err)
	}
	ctx := context.Background()
	raw := &transform.RawData{}
	if raw.Locations, err = imprt.FetchLocations(ctx, sourceClient); err != nil {
		t.Fatalf("fetch locations: %v", err)
	}
	if raw.Statuses, err = imprt.FetchStatuses(ctx, sourceClient); err != nil {
		t.Fatalf("fetch statuses: %v", err)
	}
	if raw.VLANs, err = imprt.FetchVLANs(ctx, sourceClient); err != nil {
		t.Fatalf("fetch VLANs: %v", err)
	}
	if raw.VLANLocationAssignments, err = imprt.FetchVLANLocationAssignments(ctx, sourceClient); err != nil {
		t.Fatalf("fetch VLAN location assignments: %v", err)
	}
	if raw.Prefixes, err = imprt.FetchPrefixes(ctx, sourceClient); err != nil {
		t.Fatalf("fetch prefixes: %v", err)
	}
	if raw.PrefixLocationAssignments, err = imprt.FetchPrefixLocationAssignments(ctx, sourceClient); err != nil {
		t.Fatalf("fetch prefix location assignments: %v", err)
	}

	transformed, err := transform.Transform(*devicetypes.NewInventory(), raw)
	if err != nil {
		t.Fatalf("transform imported IPAM data: %v", err)
	}
	if len(transformed.Locations) != 1 || len(transformed.VLANs) != 1 || len(transformed.Prefixes) != 1 {
		t.Fatalf("transformed counts = locations:%d VLANs:%d prefixes:%d, want 1 each",
			len(transformed.Locations), len(transformed.VLANs), len(transformed.Prefixes))
	}

	inventory := devicetypes.NewInventory()
	inventory.Locations = transformed.Locations
	inventory.VLANs = transformed.VLANs
	inventory.Prefixes = transformed.Prefixes
	var caniLocationID uuid.UUID
	for id := range inventory.Locations {
		caniLocationID = id
	}
	for _, vlan := range inventory.VLANs {
		if vlan.Location != caniLocationID {
			t.Fatalf("imported VLAN location = %s, want %s", vlan.Location, caniLocationID)
		}
	}
	for _, prefix := range inventory.Prefixes {
		if prefix.Location != caniLocationID {
			t.Fatalf("imported prefix location = %s, want %s", prefix.Location, caniLocationID)
		}
	}

	destinationLocationID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	destinationVLANID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	destinationPrefixID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	var vlanBody, prefixBody []byte
	destinationHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/ipam/vlans/":
			vlanBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusCreated)
			_, _ = fmt.Fprintf(w, `{"id":%q,"vid":100,"name":"production"}`, destinationVLANID)
		case r.Method == http.MethodPost && r.URL.Path == "/ipam/prefixes/":
			prefixBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusCreated)
			_, _ = fmt.Fprintf(w, `{"id":%q}`, destinationPrefixID)
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		}
	}
	exporter, cleanup := newExporterWithServer(t, destinationHandler)
	defer cleanup()
	seedActiveStatus(t, exporter)
	seedGlobalNamespace(uuid.New())

	locationMap := map[uuid.UUID]uuid.UUID{caniLocationID: destinationLocationID}
	loadResult := &LoadResult{}
	vlanMap, err := exporter.loadVLANs(ctx, inventory, locationMap, loadResult)
	if err != nil {
		t.Fatalf("export VLANs: %v", err)
	}
	if _, err := exporter.loadPrefixes(ctx, inventory, locationMap, vlanMap, loadResult); err != nil {
		t.Fatalf("export prefixes: %v", err)
	}
	if len(loadResult.Errors) != 0 {
		t.Fatalf("export errors: %v", loadResult.Errors)
	}

	assertScopedIPAMRequest(t, "VLAN", vlanBody, destinationLocationID, uuid.Nil)
	assertScopedIPAMRequest(t, "prefix", prefixBody, destinationLocationID, destinationVLANID)
}

func assertScopedIPAMRequest(t *testing.T, object string, body []byte, locationID, vlanID uuid.UUID) {
	t.Helper()
	var request scopedIPAMRequest
	if err := json.Unmarshal(body, &request); err != nil {
		t.Fatalf("decode %s request: %v\n%s", object, err, body)
	}
	if request.Location == nil || request.Location.ID != locationID.String() {
		t.Fatalf("%s location = %+v, want %s", object, request.Location, locationID)
	}
	if vlanID != uuid.Nil && (request.VLAN == nil || request.VLAN.ID != vlanID.String()) {
		t.Fatalf("%s VLAN = %+v, want %s", object, request.VLAN, vlanID)
	}
}
