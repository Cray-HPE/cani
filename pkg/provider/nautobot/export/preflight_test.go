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
	"reflect"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// preflightInventory is a location of the default type and one device whose
// status exists in the fake Nautobot while its device type and role do not,
// carrying one module whose type does not either.
func preflightInventory() *devicetypes.Inventory {
	inv := devicetypes.NewInventory()
	locID, devID, modID := uuid.New(), uuid.New(), uuid.New()
	inv.Locations[locID] = &devicetypes.CaniLocationType{ID: locID, Name: "test-dc"}
	inv.Devices[devID] = &devicetypes.CaniDeviceType{
		ID: devID, Name: "compute-001", Slug: "probe-slug", Type: devicetypes.Node,
		ObjectMeta: devicetypes.ObjectMeta{Status: "Active", Role: "Compute"},
	}
	inv.Modules[modID] = &devicetypes.CaniModuleType{ID: modID, Name: "nic-1", Model: "probe-module", ParentDevice: devID}
	return inv
}

// TestLoadRefusesMissingReferencesBeforeWriting verifies that with every
// create_* option off, Load stops before its first phase when the inventory
// names a location type, device type, module type or role Nautobot does not
// have, lists each gap with the message the phase would have given, and
// writes nothing.
//
// Why it matters: the phases used to run on after a refusal, so a missing
// status left prefixes exported and devices not; the pre-flight makes the
// export all-or-nothing for catalog references.
// Inputs: an empty fake Nautobot holding only the Active status; the
// pre-flight inventory; all create options off.
// Outputs: an error naming four missing objects; zero non-GET requests.
// Data choice: one gap per kind the location, device and module phases
// resolve, plus a status that is found, so the pass branch is exercised
// alongside the refusals.
func TestLoadRefusesMissingReferencesBeforeWriting(t *testing.T) {
	ensureTypesLoaded(t)
	server := &emptyNautobot{}
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()

	err := e.Load(preflightInventory())

	if err == nil || !strings.Contains(err.Error(), "4 missing reference object(s)") {
		t.Fatalf("expected a pre-flight refusal, got %v", err)
	}
	if writes := server.writes(); len(writes) != 0 {
		t.Fatalf("pre-flight refusal must not write, saw %v", writes)
	}
	missing := e.preflight(preflightInventory())
	want := []string{
		"device type not found: probe-slug",
		`location type "Site" not found in Nautobot (enable create_location_types)`,
		`module type "probe-module" not in Nautobot (enable create_module_types)`,
		"role not found: Compute",
	}
	if !reflect.DeepEqual(missing, want) {
		t.Fatalf("missing = %q, want %q", missing, want)
	}
}

// TestPreflightSkipsKindsThePhasesWillCreate verifies that a kind whose
// create_* option is on is not reported, that the lookups run with creation
// off so nothing is created by the check itself, and that the options are
// restored afterwards.
//
// Why it matters: a missing role is only a refusal when nothing will create
// it; and the check must stay read-only or the atomicity it provides is lost.
// Inputs: the pre-flight inventory with every create option on.
// Outputs: no gaps, no non-GET request, every create option still on.
// Data choice: the same inventory that produced four refusals above.
func TestPreflightSkipsKindsThePhasesWillCreate(t *testing.T) {
	ensureTypesLoaded(t)
	server := &emptyNautobot{}
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	for _, enable := range []func(bool){
		e.Cache.SetCreateDeviceTypes, e.Cache.SetCreateLocationTypes, e.Cache.SetCreateStatuses, e.Cache.SetCreateRoles,
	} {
		enable(true)
	}
	e.Options.CreateModuleTypes = true

	missing := e.preflight(preflightInventory())

	if len(missing) != 0 {
		t.Fatalf("kinds the phases create must not be reported, got %q", missing)
	}
	if writes := server.writes(); len(writes) != 0 {
		t.Fatalf("pre-flight must not write, saw %v", writes)
	}
	if !e.Cache.createDeviceTypes || !e.Cache.createLocationTypes || !e.Cache.createStatuses || !e.Cache.createRoles {
		t.Fatal("create options were not restored after the pre-flight")
	}
}

// TestCollectCatalogRefsAppliesLoaderDefaults verifies that the names the
// pre-flight collects follow the loaders' fallbacks: locations, racks,
// modules and IPAM objects fall back to the provider default status and then
// Active, a device falls back to the default status and role only, its
// interfaces are Active with their declared or inferred role, a module is
// looked up by model, a cable is Connected unless planned, and empty names
// are dropped.
//
// Why it matters: a pre-flight that looked up different names than the phases
// would refuse exports the phases could complete, or pass ones they could not.
// Inputs: a location and VLAN without status, a device without status or role
// whose interfaces are a declared "uplink" port and a management-only port, a
// module without status, a connected and a planned cable, a prefix with a
// role, and a default status of "Planned" with no default role.
// Outputs: statuses {Planned, Active, Connected}, roles {network, uplink,
// management}, device type {probe-slug}, module type {probe-module}, location
// types {Site}.
// Data choice: the device's empty role must not surface as "", the module
// shares the location's fallback, and the management port shows an inferred
// role is collected like a declared one.
func TestCollectCatalogRefsAppliesLoaderDefaults(t *testing.T) {
	inv := preflightInventory()
	mgmtOnly := true
	for _, device := range inv.Devices {
		device.Status, device.Role = "", ""
		device.Interfaces = []devicetypes.InterfaceSpec{
			{Name: "eth0", Type: ifaceType1000BaseT, Role: "uplink"},
			{Name: "bmc", Type: ifaceType1000BaseT, MgmtOnly: &mgmtOnly},
		}
	}
	vlanID, prefixID, moduleID, cableA, cableB := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.VLANs[vlanID] = &devicetypes.CaniVLAN{ID: vlanID, VID: 100}
	inv.Prefixes[prefixID] = &devicetypes.CaniPrefix{ID: prefixID, Prefix: "10.0.0.0/24", ObjectMeta: devicetypes.ObjectMeta{Role: "network"}}
	inv.Modules[moduleID] = &devicetypes.CaniModuleType{ID: moduleID, Name: "nic-1", Model: "probe-module"}
	inv.Cables[cableA] = &devicetypes.CaniCableType{ID: cableA, TerminationADevice: uuid.New(), TerminationBDevice: uuid.New()}
	inv.Cables[cableB] = &devicetypes.CaniCableType{ID: cableB, TerminationADevice: uuid.New(), TerminationBDevice: uuid.New(),
		ObjectMeta: devicetypes.ObjectMeta{Status: "planned"}}

	refs := collectCatalogRefs(inv, "Planned", "")

	want := catalogRefs{
		locationTypes: map[string]bool{"Site": true}, deviceTypes: map[string]bool{"probe-slug": true},
		moduleTypes: map[string]bool{"probe-module": true},
		statuses:    map[string]bool{"Planned": true, "Active": true, "Connected": true},
		roles:       map[string]bool{"network": true, "uplink": true, "management": true},
	}
	if !reflect.DeepEqual(refs, want) {
		t.Fatalf("refs = %+v, want %+v", refs, want)
	}
}

// TestCollectCatalogRefsSkipsWhatThePhasesSkip verifies that a cable missing
// an end, an interface of a type Nautobot lacks, and a module's own role
// contribute nothing, because the phases skip the first two and treat the
// third as optional.
//
// Why it matters: collecting them would refuse exports the phases complete.
// Inputs: an inventory of one such cable, one device with one unsupported
// interface, and one module with a role; no defaults.
// Outputs: no statuses from the cable or interface, no module role, and
// Active, from the device and the module's fallback, as the only status.
// Data choice: each object is the smallest that reaches its phase's skip.
func TestCollectCatalogRefsSkipsWhatThePhasesSkip(t *testing.T) {
	inv := devicetypes.NewInventory()
	devID, moduleID, cableID := uuid.New(), uuid.New(), uuid.New()
	inv.Devices[devID] = &devicetypes.CaniDeviceType{ID: devID, Name: "compute-001", Slug: "probe-slug", Type: devicetypes.Node,
		ObjectMeta: devicetypes.ObjectMeta{Status: "Active", Role: "Compute"},
		Interfaces: []devicetypes.InterfaceSpec{{Name: "odd0", Type: "not-a-nautobot-type", Role: "uplink"}}}
	inv.Modules[moduleID] = &devicetypes.CaniModuleType{ID: moduleID, Name: "nic-1", ObjectMeta: devicetypes.ObjectMeta{Role: "NIC"}}
	inv.Cables[cableID] = &devicetypes.CaniCableType{ID: cableID, TerminationADevice: uuid.New()}

	refs := collectCatalogRefs(inv, "", "")

	want := catalogRefs{
		locationTypes: map[string]bool{}, deviceTypes: map[string]bool{"probe-slug": true},
		moduleTypes: map[string]bool{"nic-1": true},
		statuses:    map[string]bool{"Active": true}, roles: map[string]bool{"Compute": true},
	}
	if !reflect.DeepEqual(refs, want) {
		t.Fatalf("refs = %+v, want %+v", refs, want)
	}
}
