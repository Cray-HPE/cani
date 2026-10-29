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
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// placementServer records every POST body by path and answers lookups so the
// module-type and module-bay resolve to existing objects, modules and
// interfaces are absent, and locations resolve by name. missingBay and
// missingModuleType make those lookups miss; interfaceListStatus, when set,
// is returned for interface list requests so a lookup failure can be staged.
type placementServer struct {
	moduleTypeID, moduleBayID, locationID uuid.UUID
	posts                                 map[string][][]byte
	requests                              []string
	missingBay, missingModuleType         bool
	interfaceListStatus                   int
}

func newPlacementServer() *placementServer {
	return &placementServer{moduleTypeID: uuid.New(), moduleBayID: uuid.New(), locationID: uuid.New(), posts: map[string][][]byte{}}
}

func (s *placementServer) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.requests = append(s.requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
	if r.Method == http.MethodPost {
		body, _ := io.ReadAll(r.Body)
		s.posts[r.URL.Path] = append(s.posts[r.URL.Path], body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":"`+uuid.New().String()+`","name":"created","display":"created"}`)
		return
	}
	switch {
	case strings.Contains(r.URL.Path, "dcim/module-types") && !s.missingModuleType:
		_, _ = io.WriteString(w, `{"count":1,"results":[{"id":"`+s.moduleTypeID.String()+`","model":"NIC","display":"NIC"}]}`)
	case strings.Contains(r.URL.Path, "dcim/module-bays") && !s.missingBay:
		_, _ = io.WriteString(w, `{"count":1,"results":[{"id":"`+s.moduleBayID.String()+`","name":"slot1","display":"slot1"}]}`)
	case strings.Contains(r.URL.Path, "dcim/interfaces") && s.interfaceListStatus != 0:
		w.WriteHeader(s.interfaceListStatus)
		_, _ = io.WriteString(w, `{"detail":"staged failure"}`)
	case strings.Contains(r.URL.Path, "dcim/locations"):
		_, _ = io.WriteString(w, `{"count":1,"results":[{"id":"`+s.locationID.String()+`","name":"site","display":"site"}]}`)
	default:
		_, _ = io.WriteString(w, emptyListJSON)
	}
}

// sentModule is the placement-relevant view of a ModuleRequest body.
type sentModule struct {
	ParentModuleBay *wireIDRef `json:"parent_module_bay"`
	Location        *wireIDRef `json:"location"`
}

// sentOwnedInterface is the owner-relevant view of a WritableInterfaceRequest body.
type sentOwnedInterface struct {
	Name   string     `json:"name"`
	Device *wireIDRef `json:"device"`
	Module *wireIDRef `json:"module"`
}

func decodeJSON[T any](t *testing.T, body []byte) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decoding %s: %v", body, err)
	}
	return out
}

// placementInventory returns a racked chassis "chassis" at "site" with module
// bay slot1 and the Nautobot ID map for it.
func placementInventory() (*devicetypes.Inventory, *devicetypes.CaniDeviceType, map[string]uuid.UUID) {
	inv := devicetypes.NewInventory()
	siteID, chassisID := uuid.New(), uuid.New()
	inv.Locations[siteID] = &devicetypes.CaniLocationType{ID: siteID, Name: "site"}
	chassis := &devicetypes.CaniDeviceType{ID: chassisID, Name: "chassis", ModuleBays: []devicetypes.ModuleBaySpec{{Name: "slot1"}}}
	inv.Devices[chassisID] = chassis
	return inv, chassis, map[string]uuid.UUID{"chassis": uuid.New()}
}

// TestCreateModuleFromCani_InstalledSendsBayOnlyAndOwnsPorts verifies an
// installed module is sent with its parent module bay and no location even
// when a legacy location is authored, and that its ports are created on the
// module rather than on the parent device.
//
// Why it matters: sending both placement fields is rejected by Nautobot, and
// ports created on the parent collapse same-named ports across modules.
// Inputs: a module in slot1 of the chassis with Location set to the chassis's
// site and one port "p1". Outputs: one module POST carrying parent_module_bay
// and no location, one interface POST carrying module and no device, and the
// module's Nautobot ID recorded in its external IDs.
// Data choice: authoring the matching location is the legacy inherited shape
// the ticket requires to stay compatible without producing two remote fields.
func TestCreateModuleFromCani_InstalledSendsBayOnlyAndOwnsPorts(t *testing.T) {
	server := newPlacementServer()
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	seedActiveStatus(t, e)
	inv, chassis, created := placementInventory()
	module := &devicetypes.CaniModuleType{
		ID: uuid.New(), Name: "nic1", Model: "NIC", ParentDevice: chassis.ID, ModuleBayName: "slot1",
		Location:   chassis.Parent,
		Interfaces: []devicetypes.InterfaceSpec{{Name: "p1", Type: devicetypes.InterfacesElemTypeA25GbaseXSfp28}},
	}
	result := &LoadResult{}

	if err := e.createModuleFromCani(context.Background(), module, inv, created, result); err != nil {
		t.Fatalf("createModuleFromCani returned error: %v", err)
	}

	modulePosts := server.posts["/dcim/modules/"]
	if len(modulePosts) != 1 || result.ModulesCreated != 1 {
		t.Fatalf("module POSTs = %d, created = %d; want 1 and 1", len(modulePosts), result.ModulesCreated)
	}
	sent := decodeJSON[sentModule](t, modulePosts[0])
	if sent.ParentModuleBay == nil || sent.ParentModuleBay.ID != server.moduleBayID.String() || sent.Location != nil {
		t.Errorf("module body = bay %v, location %v; want bay %s and no location", sent.ParentModuleBay, sent.Location, server.moduleBayID)
	}
	ifacePosts := server.posts["/dcim/interfaces/"]
	if len(ifacePosts) != 1 {
		t.Fatalf("interface POSTs = %d, want 1", len(ifacePosts))
	}
	port := decodeJSON[sentOwnedInterface](t, ifacePosts[0])
	if port.Module == nil || port.Device != nil || port.Module.ID != module.ExternalIDs[externalIDKeyNautobot].String() {
		t.Errorf("interface body = device %v, module %v; want module %s and no device", port.Device, port.Module, module.ExternalIDs[externalIDKeyNautobot])
	}
}

// TestCreateModuleFromCani_SpareSendsLocationOnly verifies a module with a
// location and no parent is exported as a spare with only its location, and
// that a module with neither placement is rejected.
//
// Why it matters: spares were previously an error, forcing operators to invent
// a parent; the no-placement case must still fail instead of defaulting.
// Inputs: a spare module at "site", and a module with neither parent nor
// location. Outputs: one module POST with location and no parent_module_bay;
// an error naming the missing placement for the second.
// Data choice: the two modules isolate the spare branch from the guard.
func TestCreateModuleFromCani_SpareSendsLocationOnly(t *testing.T) {
	server := newPlacementServer()
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	seedActiveStatus(t, e)
	inv, _, created := placementInventory()
	var siteID uuid.UUID
	for id := range inv.Locations {
		siteID = id
	}
	spare := &devicetypes.CaniModuleType{ID: uuid.New(), Name: "spare-nic", Model: "NIC", Location: siteID}
	unplaced := &devicetypes.CaniModuleType{ID: uuid.New(), Name: "loose-nic", Model: "NIC"}
	result := &LoadResult{}

	if err := e.createModuleFromCani(context.Background(), spare, inv, created, result); err != nil {
		t.Fatalf("spare createModuleFromCani returned error: %v", err)
	}
	err := e.createModuleFromCani(context.Background(), unplaced, inv, created, result)

	posts := server.posts["/dcim/modules/"]
	if len(posts) != 1 {
		t.Fatalf("module POSTs = %d, want 1 (spare only)", len(posts))
	}
	sent := decodeJSON[sentModule](t, posts[0])
	if sent.Location == nil || sent.Location.ID != server.locationID.String() || sent.ParentModuleBay != nil {
		t.Errorf("spare body = bay %v, location %v; want location %s and no bay", sent.ParentModuleBay, sent.Location, server.locationID)
	}
	if err == nil || !strings.Contains(err.Error(), "neither a parent device nor a location") {
		t.Errorf("unplaced module error = %v, want missing-placement error", err)
	}
}

// TestFindModuleInBay_ConflictsOnDifferentSerial verifies a bay occupied by a
// module with a different known serial is a conflict, while an occupant with
// the same serial is reused.
//
// Why it matters: idempotent re-export must recognise its own module but must
// not treat someone else's hardware in the bay as success.
// Inputs: a bay whose occupant has serial "OTHER", checked against modules
// with serial "MINE" and "OTHER". Outputs: an error naming both serials, then
// the occupant's ID.
// Data choice: serial is the only identity available before an external ID
// has been recorded.
func TestFindModuleInBay_ConflictsOnDifferentSerial(t *testing.T) {
	occupantID := uuid.New()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"count":1,"results":[{"id":"`+occupantID.String()+`","serial":"OTHER","display":"x"}]}`)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	placement := modulePlacement{parentName: "chassis", bayName: "slot1", bay: &CachedItem{ID: uuid.New()}}

	_, err := e.findModuleInBay(context.Background(), &devicetypes.CaniModuleType{Serial: "MINE"}, placement)
	if err == nil || !strings.Contains(err.Error(), `occupied by module serial "OTHER", not "MINE"`) {
		t.Errorf("different serial error = %v, want conflict", err)
	}
	occupant, err := e.findModuleInBay(context.Background(), &devicetypes.CaniModuleType{Serial: "OTHER"}, placement)
	if err != nil || occupant == nil || toUUID(occupant.Id) != occupantID {
		t.Errorf("same serial = %v, %v; want occupant %s", occupant, err, occupantID)
	}
}

// TestCreateModuleFromCani_DryRunWritesNothing verifies a dry run of an
// installed module whose bay and module type Nautobot lacks performs no
// write: the bay and type are reported as planned, the module is counted as
// planned, and its ports are planned.
//
// Why it matters: a preview that creates a module bay or module type leaves
// state behind that the operator did not approve, and the bay was created on
// every dry run before this check existed.
// Inputs: DryRun and CreateModuleTypes on, a server whose module-bay and
// module-type lists are empty, and a module in slot1 with one port. Outputs:
// zero POSTs to any path, ModulesCreated == 1, no error.
// Data choice: both lookups miss at once so a single test covers the two
// create paths a dry run must avoid.
func TestCreateModuleFromCani_DryRunWritesNothing(t *testing.T) {
	server := newPlacementServer()
	server.missingBay, server.missingModuleType = true, true
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	e.Options.DryRun, e.Options.CreateModuleTypes = true, true
	seedActiveStatus(t, e)
	inv, chassis, created := placementInventory()
	module := &devicetypes.CaniModuleType{
		ID: uuid.New(), Name: "nic1", Model: "NIC", Manufacturer: "Acme", ParentDevice: chassis.ID, ModuleBayName: "slot1",
		Interfaces: []devicetypes.InterfaceSpec{{Name: "p1", Type: devicetypes.InterfacesElemTypeA25GbaseXSfp28}},
	}
	result := &LoadResult{}

	if err := e.createModuleFromCani(context.Background(), module, inv, created, result); err != nil {
		t.Fatalf("createModuleFromCani returned error: %v", err)
	}

	if len(server.posts) != 0 {
		t.Errorf("dry run wrote to %v, want no POST", slices.Sorted(maps.Keys(server.posts)))
	}
	if result.ModulesCreated != 1 || len(result.Errors) != 0 {
		t.Errorf("created = %d, errors = %v; want 1 planned module and no errors", result.ModulesCreated, result.Errors)
	}
	for _, req := range server.requests {
		if !strings.HasPrefix(req, http.MethodGet+" ") {
			t.Errorf("dry run issued %s, want only GET requests", req)
		}
	}
}

// TestExportModulePortsReportsLookupFailure verifies that a failed existence
// check for a module port is recorded as an export error and creates nothing,
// rather than being read as absence.
//
// Why it matters: treating an API failure as "not found" would attempt a
// duplicate create on the next line, and the earlier code dropped the error.
// Inputs: an exported module with one port and a server that answers the
// module's interface list with 500. Outputs: one result error naming the
// module and the existence check, zero interface POSTs.
// Data choice: the 500 is staged only on interface lists so the module
// itself resolves normally and the failure is isolated to the lookup.
func TestExportModulePortsReportsLookupFailure(t *testing.T) {
	server := newPlacementServer()
	server.interfaceListStatus = http.StatusInternalServerError
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	module := &devicetypes.CaniModuleType{
		ID: uuid.New(), Name: "nic1",
		Interfaces: []devicetypes.InterfaceSpec{{Name: "p1", Type: devicetypes.InterfacesElemTypeA25GbaseXSfp28}},
	}
	result := &LoadResult{}

	e.exportModulePorts(context.Background(), module, uuid.New(), result)

	if len(result.Errors) != 1 || !strings.Contains(result.Errors[0], "module nic1") || !strings.Contains(result.Errors[0], "existence check") {
		t.Errorf("errors = %v, want one naming module nic1 and the existence check", result.Errors)
	}
	if n := len(server.posts["/dcim/interfaces/"]); n != 0 {
		t.Errorf("interface POSTs = %d, want 0 after a failed lookup", n)
	}
}

// TestResolveModulePlacementDescendsSpareLocation verifies a spare authored
// at a location whose type holds no modules is placed in that location's
// descendant that does, falls back to the default location when there is no
// such descendant, and is an error when neither is available.
//
// Why it matters: devices and racks already descend to a content-capable
// location; a spare that sent the container location verbatim would be
// rejected by Nautobot or, with auto-create on, spawn a same-named section.
// Inputs: a "dc" location with a "section" child, then a lone "dc" with and
// without --default-location.
// Outputs: the Nautobot location lookup carries the section's name; then the
// default's name; then an error naming the dc location.
// Data choice: "dc" and "section" are the embedded location types whose
// content types differ in exactly this way.
func TestResolveModulePlacementDescendsSpareLocation(t *testing.T) {
	server := newPlacementServer()
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	inv := devicetypes.NewInventory()
	dcID, sectionID, loneID := uuid.New(), uuid.New(), uuid.New()
	inv.Locations[dcID] = &devicetypes.CaniLocationType{ID: dcID, Name: "dc-01", LocationType: "dc"}
	inv.Locations[sectionID] = &devicetypes.CaniLocationType{ID: sectionID, Name: "row-a", LocationType: "section", Parent: dcID}
	inv.Locations[loneID] = &devicetypes.CaniLocationType{ID: loneID, Name: "dc-02", LocationType: "dc"}
	inv.RebuildDerivedState()
	spareAt := func(loc uuid.UUID) *devicetypes.CaniModuleType {
		return &devicetypes.CaniModuleType{ID: uuid.New(), Name: "spare", Location: loc}
	}
	lastLocationLookup := func() string {
		for i := len(server.requests) - 1; i >= 0; i-- {
			if strings.Contains(server.requests[i], "dcim/locations") {
				return server.requests[i]
			}
		}
		return ""
	}

	if _, err := e.resolveModulePlacement(context.Background(), spareAt(dcID), inv, nil); err != nil {
		t.Fatalf("spare under dc with a section: %v", err)
	}
	if got := lastLocationLookup(); !strings.Contains(got, "row-a") {
		t.Errorf("location lookup = %q, want the section row-a", got)
	}

	_, err := e.resolveModulePlacement(context.Background(), spareAt(loneID), inv, nil)
	if err == nil || !strings.Contains(err.Error(), `"dc-02" cannot hold modules`) {
		t.Errorf("spare at a lone dc error = %v, want a cannot-hold-modules error", err)
	}

	e.Options.DefaultLocation = "fallback-section"
	if _, err := e.resolveModulePlacement(context.Background(), spareAt(loneID), inv, nil); err != nil {
		t.Fatalf("spare at a lone dc with a default: %v", err)
	}
	if got := lastLocationLookup(); !strings.Contains(got, "fallback-section") {
		t.Errorf("location lookup = %q, want the default location", got)
	}
}

// TestResolveModulePlacementRefusesUndeclaredFallbackBay verifies a module
// with no named bay is refused on a parent that declares module bays, while
// on a parent that declares none it still falls back to its own name.
//
// Why it matters: the fallback would create a module bay the parent's device
// type does not declare, which is drift Nautobot's own module bays never
// have; on ad hoc hardware there is no declared bay to drift from.
// Inputs: a server whose module-bay lookups miss, a bay-less module in the
// chassis that declares slot1, and a bay-less module in a device that declares
// no bays. Outputs: an error naming the chassis and `update module --bay` with
// no module-bay request, and a created bay named after the second module.
// Data choice: the missing bay makes the fallback reach the create path, so
// only the refusal can explain the absence of a module-bay POST.
func TestResolveModulePlacementRefusesUndeclaredFallbackBay(t *testing.T) {
	server := newPlacementServer()
	server.missingBay = true
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	inv, chassis, created := placementInventory()
	bare := &devicetypes.CaniDeviceType{ID: uuid.New(), Name: "bare"}
	inv.Devices[bare.ID] = bare
	created[bare.Name] = uuid.New()

	refused := &devicetypes.CaniModuleType{ID: uuid.New(), Name: "nic-legacy", ParentDevice: chassis.ID}
	_, err := e.resolveModulePlacement(context.Background(), refused, inv, created)
	if err == nil || !strings.Contains(err.Error(), "chassis") || !strings.Contains(err.Error(), "update module --bay") {
		t.Fatalf("placement error = %v, want a refusal naming chassis and the remedy", err)
	}
	for _, req := range server.requests {
		if strings.Contains(req, "dcim/module-bays") {
			t.Fatalf("refused module issued %s, want no module-bay request", req)
		}
	}

	fallback := &devicetypes.CaniModuleType{ID: uuid.New(), Name: "nic-adhoc", ParentDevice: bare.ID}
	placement, err := e.resolveModulePlacement(context.Background(), fallback, inv, created)
	if err != nil || placement.bayName != "nic-adhoc" {
		t.Fatalf("placement = %+v, %v; want bay nic-adhoc", placement, err)
	}
	posts := server.posts["/dcim/module-bays/"]
	if len(posts) != 1 || decodeJSON[struct{ Name string }](t, posts[0]).Name != "nic-adhoc" {
		t.Errorf("module-bay POSTs = %d, want one named nic-adhoc", len(posts))
	}
}

// TestCreateModuleFromCani_DryRunPlansModuleOnPlannedParent verifies a dry
// run against a Nautobot that lacks the parent device plans the installed
// module, its bay and its ports instead of reporting the parent as missing,
// and records the module as planned.
//
// Why it matters: the first export of an inventory is the dry run operators
// rely on most, and before this the module phase errored for every installed
// module because the device phase hands out no ID for a planned device.
// Inputs: DryRun on, the chassis planned by the device phase and absent from
// the created-ID map, a module in slot1 with one port. Outputs: no error, one
// planned module, no module-bay request of any kind, ModulePlanned true.
// Data choice: an empty created-ID map is exactly the state a dry run leaves
// behind for a device it only planned.
func TestCreateModuleFromCani_DryRunPlansModuleOnPlannedParent(t *testing.T) {
	server := newPlacementServer()
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	e.Options.DryRun = true
	seedActiveStatus(t, e)
	inv, chassis, _ := placementInventory()
	e.Cache.PlanDevice(chassis.Name)
	module := &devicetypes.CaniModuleType{
		ID: uuid.New(), Name: "nic1", Model: "NIC", Manufacturer: "Acme", ParentDevice: chassis.ID, ModuleBayName: "slot1",
		Interfaces: []devicetypes.InterfaceSpec{{Name: "p1", Type: devicetypes.InterfacesElemTypeA25GbaseXSfp28}},
	}
	result := &LoadResult{}

	if err := e.createModuleFromCani(context.Background(), module, inv, map[string]uuid.UUID{}, result); err != nil {
		t.Fatalf("createModuleFromCani returned error: %v", err)
	}

	if result.ModulesCreated != 1 || result.IfacesCreated != 1 || len(result.Errors) != 0 {
		t.Errorf("modules = %d, interfaces = %d, errors = %v; want 1, 1 and none", result.ModulesCreated, result.IfacesCreated, result.Errors)
	}
	if !e.Cache.ModulePlanned(module.ID) {
		t.Error("module must be recorded as planned")
	}
	for _, req := range server.requests {
		if strings.Contains(req, "dcim/module-bays") || !strings.HasPrefix(req, http.MethodGet+" ") {
			t.Errorf("dry run on a planned parent issued %s", req)
		}
	}
}
