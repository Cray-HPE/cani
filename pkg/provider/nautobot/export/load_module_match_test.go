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

// remotePlacement is Nautobot's view of where a module sits.
type remotePlacement struct {
	bay, location uuid.UUID
}

// moduleServer is a stateful fake of the module endpoints: modules are
// retrieved by ID with their placement, listed by the bay they occupy, moved
// by PATCH (which updates the state) and created by POST. The module bay
// "slot1", module type "NIC" and location "site" resolve to fixed IDs.
type moduleServer struct {
	bayID, otherBayID, moduleTypeID, locationID, stockID uuid.UUID
	modules                                              map[uuid.UUID]remotePlacement
	patches                                              []string // "<module id> <body>" in order
	posts                                                int
}

func newModuleServer() *moduleServer {
	return &moduleServer{
		bayID: uuid.New(), otherBayID: uuid.New(), moduleTypeID: uuid.New(), locationID: uuid.New(), stockID: uuid.New(),
		modules: map[uuid.UUID]remotePlacement{},
	}
}

func (s *moduleServer) placementJSON(p remotePlacement) string {
	if p.bay != uuid.Nil {
		return fmt.Sprintf(`"parent_module_bay":{"id":%q},"location":null`, p.bay)
	}
	return fmt.Sprintf(`"parent_module_bay":null,"location":{"id":%q}`, p.location)
}

func (s *moduleServer) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch {
	case strings.HasPrefix(r.URL.Path, "/dcim/modules/"):
		s.serveModules(w, r)
	case strings.Contains(r.URL.Path, "dcim/module-types"):
		fmt.Fprintf(w, `{"count":1,"results":[{"id":%q,"model":"NIC","display":"NIC"}]}`, s.moduleTypeID)
	case strings.Contains(r.URL.Path, "dcim/module-bays"):
		fmt.Fprintf(w, `{"count":1,"results":[{"id":%q,"name":"slot1","display":"slot1"}]}`, s.bayID)
	case strings.Contains(r.URL.Path, "dcim/locations"):
		fmt.Fprintf(w, `{"count":1,"results":[{"id":%q,"name":"site","display":"site"}]}`, s.locationID)
	default:
		_, _ = io.WriteString(w, emptyListJSON)
	}
}

func (s *moduleServer) serveModules(w http.ResponseWriter, r *http.Request) {
	id, isOne := uuid.Parse(strings.Trim(strings.TrimPrefix(r.URL.Path, "/dcim/modules/"), "/"))
	switch {
	case isOne == nil && r.Method == http.MethodGet:
		placement, ok := s.modules[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, `{"detail":"Not found."}`)
			return
		}
		fmt.Fprintf(w, `{"id":%q,"display":"m",%s}`, id, s.placementJSON(placement))
	case isOne == nil && r.Method == http.MethodPatch:
		body, _ := io.ReadAll(r.Body)
		s.patches = append(s.patches, id.String()+" "+string(body))
		var patch struct {
			Bay      *wireIDRef `json:"parent_module_bay"`
			Location *wireIDRef `json:"location"`
		}
		_ = json.Unmarshal(body, &patch)
		placement := remotePlacement{}
		if patch.Bay != nil {
			placement.bay = uuid.MustParse(patch.Bay.ID)
		}
		if patch.Location != nil {
			placement.location = uuid.MustParse(patch.Location.ID)
		}
		s.modules[id] = placement
		fmt.Fprintf(w, `{"id":%q,"display":"m",%s}`, id, s.placementJSON(placement))
	case r.Method == http.MethodPost:
		s.posts++
		created := uuid.New()
		s.modules[created] = remotePlacement{bay: s.bayID}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"id":%q,"display":"created"}`, created)
	default:
		s.listByBay(w, r)
	}
}

func (s *moduleServer) listByBay(w http.ResponseWriter, r *http.Request) {
	bay, err := uuid.Parse(r.URL.Query().Get("parent_module_bay"))
	if err != nil {
		_, _ = io.WriteString(w, emptyListJSON)
		return
	}
	for id, placement := range s.modules {
		if placement.bay == bay {
			fmt.Fprintf(w, `{"count":1,"results":[{"id":%q,"display":"m",%s}]}`, id, s.placementJSON(placement))
			return
		}
	}
	_, _ = io.WriteString(w, emptyListJSON)
}

// moveFixture returns an exporter against the fake, the fixture inventory with
// its chassis and the created-device map, and a module whose recorded external
// ID Nautobot knows at the given placement.
func moveFixture(t *testing.T, server *moduleServer, remote remotePlacement, merge, dryRun bool) (*Exporter, func(), *devicetypes.Inventory, *devicetypes.CaniDeviceType, map[string]uuid.UUID, *devicetypes.CaniModuleType) {
	t.Helper()
	e, cleanup := newExporterWithServer(t, server.handler)
	e.Options.Merge, e.Options.DryRun = merge, dryRun
	seedActiveStatus(t, e)
	inv, chassis, created := placementInventory()
	nautobotID := uuid.New()
	server.modules[nautobotID] = remote
	module := &devicetypes.CaniModuleType{ID: uuid.New(), Name: "nic", Model: "NIC"}
	module.ExternalIDs = map[string]uuid.UUID{externalIDKeyNautobot: nautobotID}
	return e, cleanup, inv, chassis, created, module
}

// TestCreateModuleFromCani_MovesModuleFoundByExternalID verifies a module
// Nautobot holds under the recorded external ID is moved to the placement
// the inventory records, in both directions, clearing the field it leaves.
//
// Why it matters: this is the export half of `swap module` and of
// `update module --bay`; before identity-first matching an installed module
// was matched to whatever occupied its new bay and never moved.
// Inputs: with --merge, a module installed in slot1 whose remote sits in
// another bay, and a spare at "site" whose remote sits in slot1. Outputs: one
// PATCH each carrying the new reference and null for the other, ModulesMoved
// incremented, no POST, the external ID kept and recorded for the run.
// Data choice: the two cases cover both PATCH shapes placementPatch emits.
func TestCreateModuleFromCani_MovesModuleFoundByExternalID(t *testing.T) {
	cases := []struct {
		name     string
		remote   func(s *moduleServer) remotePlacement
		place    func(inv *devicetypes.Inventory, chassis *devicetypes.CaniDeviceType, m *devicetypes.CaniModuleType)
		wantBody func(s *moduleServer) string
	}{
		{
			"installed module moved between bays",
			func(s *moduleServer) remotePlacement { return remotePlacement{bay: s.otherBayID} },
			func(_ *devicetypes.Inventory, chassis *devicetypes.CaniDeviceType, m *devicetypes.CaniModuleType) {
				m.ParentDevice, m.ModuleBayName = chassis.ID, "slot1"
			},
			func(s *moduleServer) string {
				return fmt.Sprintf(`{"location":null,"parent_module_bay":{"id":%q}}`, s.bayID)
			},
		},
		{
			"installed module stored as a spare",
			func(s *moduleServer) remotePlacement { return remotePlacement{bay: s.bayID} },
			func(inv *devicetypes.Inventory, _ *devicetypes.CaniDeviceType, m *devicetypes.CaniModuleType) {
				for id := range inv.Locations {
					m.Location = id
				}
			},
			func(s *moduleServer) string {
				return fmt.Sprintf(`{"location":{"id":%q},"parent_module_bay":null}`, s.locationID)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := newModuleServer()
			e, cleanup, inv, chassis, created, module := moveFixture(t, server, tc.remote(server), true, false)
			defer cleanup()
			tc.place(inv, chassis, module)
			nautobotID := module.ExternalIDs[externalIDKeyNautobot]
			result := &LoadResult{}

			if err := e.createModuleFromCani(context.Background(), module, inv, created, result); err != nil {
				t.Fatalf("createModuleFromCani: %v", err)
			}

			want := nautobotID.String() + " " + tc.wantBody(server)
			if len(server.patches) != 1 || server.patches[0] != want {
				t.Errorf("patches = %q, want [%q]", server.patches, want)
			}
			if result.ModulesMoved != 1 || result.ModulesSkipped != 0 || server.posts != 0 {
				t.Errorf("moved = %d, skipped = %d, posts = %d; want 1, 0, 0", result.ModulesMoved, result.ModulesSkipped, server.posts)
			}
			if module.ExternalIDs[externalIDKeyNautobot] != nautobotID || e.Cache.ExportedModuleID(module.ID) != nautobotID {
				t.Errorf("external ID = %s, recorded = %s; want %s for both", module.ExternalIDs[externalIDKeyNautobot], e.Cache.ExportedModuleID(module.ID), nautobotID)
			}
		})
	}
}

// TestCreateModuleFromCani_PlacementGates verifies the three outcomes short
// of a write for a module found by external ID: it is skipped when it
// already sits where the inventory places it, reported and counted as not
// exported as authored without --merge, and previewed without a PATCH in a
// dry run.
//
// Why it matters: an idempotent re-export must stay silent, a move must not
// happen behind an export that was not asked to update, and a dry run must
// still show the move it would make.
// Inputs: a module in slot1 whose remote sits in slot1 (merge on); the same
// module with its remote in another bay, without --merge; then with --merge
// and --dry-run. Outputs: ModulesSkipped 1; ModulesMisplaced 1 and a lossy
// count of 1; ModulesMoved 1; and no PATCH or POST in any case.
// Data choice: one module and one server per case keep the state of the fake
// independent of case order.
func TestCreateModuleFromCani_PlacementGates(t *testing.T) {
	cases := []struct {
		name          string
		remoteBay     func(s *moduleServer) uuid.UUID
		merge, dryRun bool
		check         func(t *testing.T, result *LoadResult)
	}{
		{"current placement is skipped", func(s *moduleServer) uuid.UUID { return s.bayID }, true, false,
			func(t *testing.T, r *LoadResult) {
				if r.ModulesSkipped != 1 || r.ModulesMoved != 0 || r.ModulesMisplaced != 0 {
					t.Errorf("skipped = %d, moved = %d, misplaced = %d; want 1, 0, 0", r.ModulesSkipped, r.ModulesMoved, r.ModulesMisplaced)
				}
			}},
		{"misplaced without --merge is reported", func(s *moduleServer) uuid.UUID { return s.otherBayID }, false, false,
			func(t *testing.T, r *LoadResult) {
				if r.ModulesMisplaced != 1 || r.ModulesMoved != 0 || lossyExportCount(r) != 1 {
					t.Errorf("misplaced = %d, moved = %d, lossy = %d; want 1, 0, 1", r.ModulesMisplaced, r.ModulesMoved, lossyExportCount(r))
				}
			}},
		{"dry run previews the move", func(s *moduleServer) uuid.UUID { return s.otherBayID }, true, true,
			func(t *testing.T, r *LoadResult) {
				if r.ModulesMoved != 1 || r.ModulesMisplaced != 0 {
					t.Errorf("moved = %d, misplaced = %d; want 1, 0", r.ModulesMoved, r.ModulesMisplaced)
				}
			}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := newModuleServer()
			e, cleanup, inv, chassis, created, module := moveFixture(t, server, remotePlacement{bay: tc.remoteBay(server)}, tc.merge, tc.dryRun)
			defer cleanup()
			module.ParentDevice, module.ModuleBayName = chassis.ID, "slot1"
			result := &LoadResult{}

			if err := e.createModuleFromCani(context.Background(), module, inv, created, result); err != nil {
				t.Fatalf("createModuleFromCani: %v", err)
			}

			tc.check(t, result)
			if len(server.patches) != 0 || server.posts != 0 {
				t.Errorf("patches = %d, posts = %d; want no writes", len(server.patches), server.posts)
			}
			if e.Cache.ExportedModuleID(module.ID) != module.ExternalIDs[externalIDKeyNautobot] {
				t.Errorf("recorded ID = %s, want the external ID %s", e.Cache.ExportedModuleID(module.ID), module.ExternalIDs[externalIDKeyNautobot])
			}
		})
	}
}

// TestCreateModuleFromCani_StaleExternalIDIsReplaced verifies a recorded
// external ID Nautobot no longer has is dropped, after which the module is
// resolved by its placement and created.
//
// Why it matters: a module deleted in Nautobot must not block its re-export,
// and the stale ID must not be sent in a filter Nautobot answers with 400.
// Inputs: a module in slot1 whose external ID the fake does not know, with an
// empty bay. Outputs: one POST, ModulesCreated 1, and an external ID equal to
// the created module's, not the stale one.
// Data choice: the empty bay isolates the stale-ID path from bay matching.
func TestCreateModuleFromCani_StaleExternalIDIsReplaced(t *testing.T) {
	server := newModuleServer()
	e, cleanup, inv, chassis, created, module := moveFixture(t, server, remotePlacement{}, true, false)
	defer cleanup()
	stale := uuid.New()
	module.ExternalIDs[externalIDKeyNautobot] = stale
	module.ParentDevice, module.ModuleBayName = chassis.ID, "slot1"
	result := &LoadResult{}

	if err := e.createModuleFromCani(context.Background(), module, inv, created, result); err != nil {
		t.Fatalf("createModuleFromCani: %v", err)
	}

	got := module.ExternalIDs[externalIDKeyNautobot]
	if server.posts != 1 || result.ModulesCreated != 1 || got == stale || got == uuid.Nil || server.modules[got].bay != server.bayID {
		t.Errorf("posts = %d, created = %d, external ID = %s (stale %s); want one create recorded under the new ID", server.posts, result.ModulesCreated, got, stale)
	}
}

// TestCreateModuleFromCani_RefusesOccupiedTargetBay verifies a move into a
// bay another module still holds is refused before any PATCH.
//
// Why it matters: Nautobot holds one module per bay; a PATCH into an occupied
// bay fails with a constraint error, and the operator needs the remedy.
// Inputs: with --merge, a module in slot1 whose remote sits in another bay,
// while a different module occupies slot1 in the fake. Outputs: an error
// naming the occupant and the remedy, and no PATCH.
// Data choice: the occupant is a module the inventory does not know, the case
// a swap ordering cannot resolve.
func TestCreateModuleFromCani_RefusesOccupiedTargetBay(t *testing.T) {
	server := newModuleServer()
	e, cleanup, inv, chassis, created, module := moveFixture(t, server, remotePlacement{bay: server.otherBayID}, true, false)
	defer cleanup()
	occupant := uuid.New()
	server.modules[occupant] = remotePlacement{bay: server.bayID}
	module.ParentDevice, module.ModuleBayName = chassis.ID, "slot1"

	err := e.createModuleFromCani(context.Background(), module, inv, created, &LoadResult{})

	want := fmt.Sprintf(`bay "slot1" on chassis is occupied by module %s; store or move that module first`, occupant)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("error = %v, want %q", err, want)
	}
	if len(server.patches) != 0 {
		t.Errorf("patches = %q, want none", server.patches)
	}
}

// TestLoadModulesExportsSwapInOrder verifies a swapped pair is exported so
// the module leaving the bay vacates it before the module entering it is
// moved, with both PATCHes landing and the final state matching the inventory.
//
// Why it matters: this is the whole-run contract behind `swap module` then
// `export nautobot --merge`; a bay-first order would refuse the incoming
// module because the outgoing one still holds the bay.
// Inputs: an inventory where "a-spare" is now in slot1 and "z-bad" is stored
// at the site, against a fake where "z-bad" holds slot1 and "a-spare" sits at
// the site. Outputs: two PATCHes, "z-bad" first, and the fake showing
// "a-spare" in slot1 and "z-bad" at the site.
// Data choice: names sort "a-spare" before "z-bad", so a name order would
// move the incoming module first and be refused; the observed order can only
// come from the spares-first rule.
func TestLoadModulesExportsSwapInOrder(t *testing.T) {
	server := newModuleServer()
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	e.Options.Merge = true
	seedActiveStatus(t, e)
	inv, chassis, created := placementInventory()
	var site uuid.UUID
	for id := range inv.Locations {
		site = id
	}
	badID, spareID := uuid.New(), uuid.New()
	server.modules[badID] = remotePlacement{bay: server.bayID}
	server.modules[spareID] = remotePlacement{location: server.locationID}
	bad := &devicetypes.CaniModuleType{ID: uuid.New(), Name: "z-bad", Model: "NIC", Location: site}
	bad.ExternalIDs = map[string]uuid.UUID{externalIDKeyNautobot: badID}
	spare := &devicetypes.CaniModuleType{ID: uuid.New(), Name: "a-spare", Model: "NIC", ParentDevice: chassis.ID, ModuleBayName: "slot1"}
	spare.ExternalIDs = map[string]uuid.UUID{externalIDKeyNautobot: spareID}
	inv.Modules[bad.ID], inv.Modules[spare.ID] = bad, spare
	result := &LoadResult{}

	if err := e.loadModules(context.Background(), inv, created, result); err != nil {
		t.Fatalf("loadModules: %v", err)
	}

	if len(result.Errors) != 0 || result.ModulesMoved != 2 {
		t.Fatalf("errors = %v, moved = %d; want none and 2", result.Errors, result.ModulesMoved)
	}
	if len(server.patches) != 2 || !strings.HasPrefix(server.patches[0], badID.String()) || !strings.HasPrefix(server.patches[1], spareID.String()) {
		t.Errorf("patch order = %q, want z-bad then a-spare", server.patches)
	}
	if server.modules[spareID] != (remotePlacement{bay: server.bayID}) || server.modules[badID] != (remotePlacement{location: server.locationID}) {
		t.Errorf("remote state = spare %+v, bad %+v; want spare in slot1 and bad at site", server.modules[spareID], server.modules[badID])
	}
}

// TestModuleExportOrderPutsSparesFirst verifies the export order lists spares
// before installed modules, sorts by name within each group, and drops nil
// and unnamed modules.
//
// Why it matters: a stable, spares-first order is what lets a swap export in
// one run and keeps the log comparable between runs.
// Inputs: installed "b" and "a", spares "z" and "y", a nil entry and an
// unnamed module. Outputs: y, z, a, b.
// Data choice: reverse-alphabetical insertion proves the sort rather than
// map order.
func TestModuleExportOrderPutsSparesFirst(t *testing.T) {
	inv := devicetypes.NewInventory()
	site, parent := uuid.New(), uuid.New()
	for _, name := range []string{"b", "a"} {
		id := uuid.New()
		inv.Modules[id] = &devicetypes.CaniModuleType{ID: id, Name: name, ParentDevice: parent}
	}
	for _, name := range []string{"z", "y"} {
		id := uuid.New()
		inv.Modules[id] = &devicetypes.CaniModuleType{ID: id, Name: name, Location: site}
	}
	inv.Modules[uuid.New()] = nil
	inv.Modules[uuid.New()] = &devicetypes.CaniModuleType{ID: uuid.New(), Location: site}

	var names []string
	for _, module := range moduleExportOrder(inv) {
		names = append(names, module.Name)
	}
	if got := strings.Join(names, ","); got != "y,z,a,b" {
		t.Errorf("order = %s, want y,z,a,b", got)
	}
}
