/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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

// +---------------------------------------+--------------------------------------------------+--------------------------------------------------+
// | Function                              | Happy-path test                                  | Failure test                                     |
// +---------------------------------------+--------------------------------------------------+--------------------------------------------------+
// | FindLocationByName                    | TestFindLocationByNameFound                      | TestFindLocationByNameNotFound                   |
// | LocationExists                        | TestLocationExistsTrue                            | TestLocationExistsFalse                           |
// | FindRackByName                        | TestFindRackByNameFound                          | TestFindRackByNameNotFound                       |
// | RackExists                            | TestRackExistsTrue                                | TestRackExistsFalse                               |
// | RacksByLocation                       | TestRacksByLocation                              | TestRacksByLocationEmpty                         |
// | FindLocationByNameOrID                | TestFindLocationByNameOrIDByUUID                 | TestFindLocationByNameOrIDNotFound               |
// | FindDeviceByNameOrID                  | TestFindDeviceByNameOrIDByUUID                   | TestFindDeviceByNameOrIDNotFound                 |
// | DevicesBySlug                         | TestDevicesBySlugFound                           | TestDevicesBySlugEmpty                           |
// | OccupiedModuleBays                    | TestOccupiedModuleBaysFound                      | TestOccupiedModuleBaysEmpty                      |
// | AvailableModuleBays                   | TestAvailableModuleBaysFound                     | TestAvailableModuleBaysAllOccupied               |
// | FindModuleByName                      | TestFindModuleByNameFound                        | TestFindModuleByNameNotFound                     |
// | ModuleExists                          | TestModuleExistsTrue                              | TestModuleExistsFalse                             |
// | FindFruByName                         | TestFindFruByNameFound                           | TestFindFruByNameNotFound                        |
// | FruExists                             | TestFruExistsTrue                                 | TestFruExistsFalse                                |
// | FindCableByLabel                      | TestFindCableByLabelFound                        | TestFindCableByLabelNotFound                     |
// | GetDevicesInRack                      | TestGetDevicesInRackFound                        | TestGetDevicesInRackEmpty                        |
// | GetCablesForDevice                    | TestGetCablesForDeviceFound                      | TestGetCablesForDeviceEmpty                      |
// | GetModulesForDevice                   | TestGetModulesForDeviceFound                     | TestGetModulesForDeviceEmpty                     |
// | findInterfaceOnDevice                 | TestFindInterfaceOnDeviceFound                   | TestFindInterfaceOnDeviceNotFound                |
// | findInterfaceInModules                | TestFindInterfaceInModulesFound                  | TestFindInterfaceInModulesNotFound               |
// | GetInterfaceByID                      | TestGetInterfaceByIDFound                        | TestGetInterfaceByIDNotFound                     |
// | GetInterfacesByDevice                 | TestGetInterfacesByDeviceFound                   | TestGetInterfacesByDeviceEmpty                   |
// | validateDeviceRefs                    | TestValidateDeviceRefsValid                      | TestValidateDeviceRefsBroken                     |
// | validateLocationRefs                  | TestValidateLocationRefsValid                    | TestValidateLocationRefsBroken                   |
// | validateRackRefs                      | TestValidateRackRefsValid                        | TestValidateRackRefsBroken                       |
// | validateModuleRefs                    | TestValidateModuleRefsValid                      | TestValidateModuleRefsBroken                     |
// | validateCableRefs                     | TestValidateCableRefsValid                       | TestValidateCableRefsBroken                      |
// | validateFruRefs                       | TestValidateFruRefsValid                         | TestValidateFruRefsBroken                        |
// | Validate                              | TestValidateValid                                 | TestValidateNilInventory                          |
// | parentExists                          | TestParentExistsTrue                              | TestParentExistsFalse                             |
// +---------------------------------------+--------------------------------------------------+--------------------------------------------------+

package devicetypes

import (
	"testing"

	"github.com/google/uuid"
)

// --- FindLocationByName ---

func TestFindLocationByNameFound(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Locations[id] = &CaniLocationType{ID: id, Name: "site-alpha"}

	got := inv.FindLocationByName("site-alpha")
	if got == nil {
		t.Fatal("expected location, got nil")
	}
	if got.ID != id {
		t.Errorf("expected ID %s, got %s", id, got.ID)
	}
}

func TestFindLocationByNameNotFound(t *testing.T) {
	inv := NewInventory()
	inv.Locations[uuid.New()] = &CaniLocationType{Name: "site-alpha"}

	got := inv.FindLocationByName("nonexistent")
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

// --- LocationExists ---

func TestLocationExistsTrue(t *testing.T) {
	inv := NewInventory()
	inv.Locations[uuid.New()] = &CaniLocationType{Name: "lab-1"}

	if !inv.LocationExists("lab-1") {
		t.Error("expected LocationExists to return true")
	}
}

func TestLocationExistsFalse(t *testing.T) {
	inv := NewInventory()

	if inv.LocationExists("missing") {
		t.Error("expected LocationExists to return false for empty inventory")
	}
}

// --- FindRackByName ---

func TestFindRackByNameFound(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Racks[id] = &CaniRackType{ID: id, Name: "rack-01"}

	got := inv.FindRackByName("rack-01")
	if got == nil {
		t.Fatal("expected rack, got nil")
	}
	if got.ID != id {
		t.Errorf("expected ID %s, got %s", id, got.ID)
	}
}

func TestFindRackByNameNotFound(t *testing.T) {
	inv := NewInventory()
	inv.Racks[uuid.New()] = &CaniRackType{Name: "rack-01"}

	got := inv.FindRackByName("rack-99")
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

// --- RackExists ---

func TestRackExistsTrue(t *testing.T) {
	inv := NewInventory()
	inv.Racks[uuid.New()] = &CaniRackType{Name: "rack-A"}

	if !inv.RackExists("rack-A") {
		t.Error("expected RackExists to return true")
	}
}

func TestRackExistsFalse(t *testing.T) {
	inv := NewInventory()

	if inv.RackExists("no-such-rack") {
		t.Error("expected RackExists to return false for empty inventory")
	}
}

// --- FindModuleByName ---

func TestFindModuleByNameFound(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Modules[id] = &CaniModuleType{ID: id, Name: "nic-module"}

	got := inv.FindModuleByName("nic-module")
	if got == nil {
		t.Fatal("expected module, got nil")
	}
	if got.ID != id {
		t.Errorf("expected ID %s, got %s", id, got.ID)
	}
}

func TestFindModuleByNameNotFound(t *testing.T) {
	inv := NewInventory()
	inv.Modules[uuid.New()] = &CaniModuleType{Name: "nic-module"}

	got := inv.FindModuleByName("gpu-module")
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

// --- ModuleExists ---

func TestModuleExistsTrue(t *testing.T) {
	inv := NewInventory()
	inv.Modules[uuid.New()] = &CaniModuleType{Name: "psu-module"}

	if !inv.ModuleExists("psu-module") {
		t.Error("expected ModuleExists to return true")
	}
}

func TestModuleExistsFalse(t *testing.T) {
	inv := NewInventory()

	if inv.ModuleExists("missing-module") {
		t.Error("expected ModuleExists to return false for empty inventory")
	}
}

// --- FindFruByName ---

func TestFindFruByNameFound(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Frus[id] = &CaniFruType{ID: id, Name: "fan-tray"}

	got := inv.FindFruByName("fan-tray")
	if got == nil {
		t.Fatal("expected FRU, got nil")
	}
	if got.ID != id {
		t.Errorf("expected ID %s, got %s", id, got.ID)
	}
}

func TestFindFruByNameNotFound(t *testing.T) {
	inv := NewInventory()
	inv.Frus[uuid.New()] = &CaniFruType{Name: "fan-tray"}

	got := inv.FindFruByName("power-supply")
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

// --- FruExists ---

func TestFruExistsTrue(t *testing.T) {
	inv := NewInventory()
	inv.Frus[uuid.New()] = &CaniFruType{Name: "dimm-slot"}

	if !inv.FruExists("dimm-slot") {
		t.Error("expected FruExists to return true")
	}
}

func TestFruExistsFalse(t *testing.T) {
	inv := NewInventory()

	if inv.FruExists("no-fru") {
		t.Error("expected FruExists to return false for empty inventory")
	}
}

// --- FindCableByLabel ---

func TestFindCableByLabelFound(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Cables[id] = &CaniCableType{ID: id, Label: "cable-001"}

	got := inv.FindCableByLabel("cable-001")
	if got == nil {
		t.Fatal("expected cable, got nil")
	}
	if got.ID != id {
		t.Errorf("expected ID %s, got %s", id, got.ID)
	}
}

func TestFindCableByLabelNotFound(t *testing.T) {
	inv := NewInventory()
	inv.Cables[uuid.New()] = &CaniCableType{Label: "cable-001"}

	got := inv.FindCableByLabel("cable-999")
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

// --- GetDevicesInRack ---

func TestGetDevicesInRackFound(t *testing.T) {
	inv := NewInventory()
	rackID := uuid.New()
	devID := uuid.New()
	inv.Devices[devID] = &CaniDeviceType{ID: devID, Name: "server-1", Parent: rackID}

	got := inv.GetDevicesInRack(rackID)
	if len(got) != 1 {
		t.Fatalf("expected 1 device, got %d", len(got))
	}
	if got[0].ID != devID {
		t.Errorf("expected device ID %s, got %s", devID, got[0].ID)
	}
}

func TestGetDevicesInRackEmpty(t *testing.T) {
	inv := NewInventory()
	rackID := uuid.New()

	got := inv.GetDevicesInRack(rackID)
	if len(got) != 0 {
		t.Errorf("expected 0 devices, got %d", len(got))
	}
}

// --- GetCablesForDevice ---

func TestGetCablesForDeviceFound(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	cableID := uuid.New()
	inv.Cables[cableID] = &CaniCableType{
		ID:                 cableID,
		Label:              "uplink",
		TerminationADevice: devID,
	}

	got := inv.GetCablesForDevice(devID)
	if len(got) != 1 {
		t.Fatalf("expected 1 cable, got %d", len(got))
	}
	if got[0].ID != cableID {
		t.Errorf("expected cable ID %s, got %s", cableID, got[0].ID)
	}
}

func TestGetCablesForDeviceEmpty(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()

	got := inv.GetCablesForDevice(devID)
	if len(got) != 0 {
		t.Errorf("expected 0 cables, got %d", len(got))
	}
}

// --- GetModulesForDevice ---

func TestGetModulesForDeviceFound(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	modID := uuid.New()
	inv.Modules[modID] = &CaniModuleType{ID: modID, Name: "nic-1", ParentDevice: devID}

	got := inv.GetModulesForDevice(devID)
	if len(got) != 1 {
		t.Fatalf("expected 1 module, got %d", len(got))
	}
	if got[0].ID != modID {
		t.Errorf("expected module ID %s, got %s", modID, got[0].ID)
	}
}

func TestGetModulesForDeviceEmpty(t *testing.T) {
	inv := NewInventory()

	got := inv.GetModulesForDevice(uuid.New())
	if len(got) != 0 {
		t.Errorf("expected 0 modules, got %d", len(got))
	}
}

// --- findInterfaceOnDevice ---

func TestFindInterfaceOnDeviceFound(t *testing.T) {
	ifaceID := uuid.New()
	device := &CaniDeviceType{
		Interfaces: []InterfaceSpec{
			{ID: ifaceID, Name: "eth0", Type: "1000base-t"},
		},
	}

	got := findInterfaceOnDevice(device, ifaceID)
	if got == nil {
		t.Fatal("expected interface spec, got nil")
	}
	if got.Name != "eth0" {
		t.Errorf("expected name eth0, got %s", got.Name)
	}
}

func TestFindInterfaceOnDeviceNotFound(t *testing.T) {
	device := &CaniDeviceType{
		Interfaces: []InterfaceSpec{
			{ID: uuid.New(), Name: "eth0"},
		},
	}

	got := findInterfaceOnDevice(device, uuid.New())
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

// --- findInterfaceInModules ---

func TestFindInterfaceInModulesFound(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	modID := uuid.New()
	ifaceID := uuid.New()

	inv.Devices[devID] = &CaniDeviceType{ID: devID, Name: "switch-1"}
	inv.Modules[modID] = &CaniModuleType{
		ID:           modID,
		ParentDevice: devID,
		Interfaces: []InterfaceSpec{
			{ID: ifaceID, Name: "ge-0/0/1", Type: "1000base-t"},
		},
	}

	spec, parentDev := inv.findInterfaceInModules(ifaceID)
	if spec == nil {
		t.Fatal("expected interface spec, got nil")
	}
	if spec.Name != "ge-0/0/1" {
		t.Errorf("expected name ge-0/0/1, got %s", spec.Name)
	}
	if parentDev == nil || parentDev.ID != devID {
		t.Errorf("expected parent device %s", devID)
	}
}

func TestFindInterfaceInModulesNotFound(t *testing.T) {
	inv := NewInventory()
	inv.Modules[uuid.New()] = &CaniModuleType{
		Interfaces: []InterfaceSpec{
			{ID: uuid.New(), Name: "ge-0/0/0"},
		},
	}

	spec, dev := inv.findInterfaceInModules(uuid.New())
	if spec != nil {
		t.Errorf("expected nil spec, got %+v", spec)
	}
	if dev != nil {
		t.Errorf("expected nil device, got %+v", dev)
	}
}

// --- GetInterfaceByID ---

func TestGetInterfaceByIDFound(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	ifaceID := uuid.New()

	inv.Devices[devID] = &CaniDeviceType{
		ID:   devID,
		Name: "router-1",
		Interfaces: []InterfaceSpec{
			{ID: ifaceID, Name: "eth0", Type: "1000base-t"},
		},
	}
	inv.Interfaces[ifaceID] = &CaniInterface{
		ID:       ifaceID,
		DeviceID: devID,
	}

	spec, dev := inv.GetInterfaceByID(ifaceID)
	if spec == nil {
		t.Fatal("expected interface spec, got nil")
	}
	if spec.Name != "eth0" {
		t.Errorf("expected name eth0, got %s", spec.Name)
	}
	if dev == nil || dev.ID != devID {
		t.Errorf("expected device %s", devID)
	}
}

func TestGetInterfaceByIDNotFound(t *testing.T) {
	inv := NewInventory()

	spec, dev := inv.GetInterfaceByID(uuid.New())
	if spec != nil {
		t.Errorf("expected nil spec, got %+v", spec)
	}
	if dev != nil {
		t.Errorf("expected nil device, got %+v", dev)
	}
}

// --- GetInterfacesByDevice ---

func TestGetInterfacesByDeviceFound(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	ifaceID := uuid.New()

	inv.Interfaces[ifaceID] = &CaniInterface{
		ID:       ifaceID,
		Name:     "eth0",
		DeviceID: devID,
	}

	got := inv.GetInterfacesByDevice(devID)
	if len(got) != 1 {
		t.Fatalf("expected 1 interface instance, got %d", len(got))
	}
	if got[0].ID != ifaceID {
		t.Errorf("expected interface ID %s, got %s", ifaceID, got[0].ID)
	}
}

func TestGetInterfacesByDeviceEmpty(t *testing.T) {
	inv := NewInventory()

	got := inv.GetInterfacesByDevice(uuid.New())
	if len(got) != 0 {
		t.Errorf("expected 0 interface instances, got %d", len(got))
	}
}

// --- validateDeviceRefs ---

func TestValidateDeviceRefsValid(t *testing.T) {
	inv := NewInventory()
	parentID := uuid.New()
	childID := uuid.New()

	inv.Devices[parentID] = &CaniDeviceType{
		ID:       parentID,
		Name:     "chassis",
		Children: []uuid.UUID{childID},
	}
	inv.Devices[childID] = &CaniDeviceType{
		ID:     childID,
		Name:   "blade",
		Parent: parentID,
	}

	errs := inv.validateDeviceRefs()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateDeviceRefsBroken(t *testing.T) {
	inv := NewInventory()
	missingChild := uuid.New()

	inv.Devices[uuid.New()] = &CaniDeviceType{
		Name:     "chassis",
		Children: []uuid.UUID{missingChild},
	}

	errs := inv.validateDeviceRefs()
	if len(errs) == 0 {
		t.Error("expected validation errors for missing child")
	}
}

// --- validateLocationRefs ---

func TestValidateLocationRefsValid(t *testing.T) {
	inv := NewInventory()
	parentID := uuid.New()
	childID := uuid.New()
	rackID := uuid.New()

	inv.Locations[parentID] = &CaniLocationType{
		ID:   parentID,
		Name: "datacenter",
	}
	inv.Locations[childID] = &CaniLocationType{
		ID:     childID,
		Name:   "row-1",
		Parent: parentID,
		Racks:  []uuid.UUID{rackID},
	}
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "rack-01"}

	errs := inv.validateLocationRefs()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateLocationRefsBroken(t *testing.T) {
	inv := NewInventory()
	missingRack := uuid.New()

	inv.Locations[uuid.New()] = &CaniLocationType{
		Name:  "row-1",
		Racks: []uuid.UUID{missingRack},
	}

	errs := inv.validateLocationRefs()
	if len(errs) == 0 {
		t.Error("expected validation errors for missing rack reference")
	}
}

// --- validateRackRefs ---

func TestValidateRackRefsValid(t *testing.T) {
	inv := NewInventory()
	locID := uuid.New()
	rackID := uuid.New()

	inv.Locations[locID] = &CaniLocationType{ID: locID, Name: "room-1"}
	inv.Racks[rackID] = &CaniRackType{
		ID:       rackID,
		Name:     "rack-01",
		Location: locID,
	}

	errs := inv.validateRackRefs()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateRackRefsBroken(t *testing.T) {
	inv := NewInventory()
	missingLoc := uuid.New()

	inv.Racks[uuid.New()] = &CaniRackType{
		Name:     "rack-01",
		Location: missingLoc,
	}

	errs := inv.validateRackRefs()
	if len(errs) == 0 {
		t.Error("expected validation errors for missing location reference")
	}
}

// --- validateModuleRefs ---

func TestValidateModuleRefsValid(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()

	inv.Devices[devID] = &CaniDeviceType{ID: devID, Name: "switch-1"}
	inv.Modules[uuid.New()] = &CaniModuleType{
		Name:         "line-card",
		ParentDevice: devID,
	}

	errs := inv.validateModuleRefs()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateModuleRefsBroken(t *testing.T) {
	inv := NewInventory()
	missingDev := uuid.New()

	inv.Modules[uuid.New()] = &CaniModuleType{
		Name:         "line-card",
		ParentDevice: missingDev,
	}

	errs := inv.validateModuleRefs()
	if len(errs) == 0 {
		t.Error("expected validation errors for missing parent device")
	}
}

// --- validateCableRefs ---

func TestValidateCableRefsValid(t *testing.T) {
	inv := NewInventory()
	devA := uuid.New()
	devB := uuid.New()

	inv.Devices[devA] = &CaniDeviceType{ID: devA, Name: "switch-a"}
	inv.Devices[devB] = &CaniDeviceType{ID: devB, Name: "switch-b"}
	inv.Cables[uuid.New()] = &CaniCableType{
		Label:              "link-1",
		TerminationADevice: devA,
		TerminationBDevice: devB,
	}

	errs := inv.validateCableRefs()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateCableRefsBroken(t *testing.T) {
	inv := NewInventory()
	missingDev := uuid.New()

	inv.Cables[uuid.New()] = &CaniCableType{
		Label:              "link-1",
		TerminationADevice: missingDev,
	}

	errs := inv.validateCableRefs()
	if len(errs) == 0 {
		t.Error("expected validation errors for missing termination device")
	}
}

// --- validateFruRefs ---

func TestValidateFruRefsValid(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()

	inv.Devices[devID] = &CaniDeviceType{ID: devID, Name: "server-1"}
	inv.Frus[uuid.New()] = &CaniFruType{
		Name:   "psu-1",
		Device: devID,
	}

	errs := inv.validateFruRefs()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateFruRefsBroken(t *testing.T) {
	inv := NewInventory()
	missingDev := uuid.New()

	inv.Frus[uuid.New()] = &CaniFruType{
		Name:   "psu-1",
		Device: missingDev,
	}

	errs := inv.validateFruRefs()
	if len(errs) == 0 {
		t.Error("expected validation errors for missing FRU device")
	}
}

// --- Validate ---

func TestValidateValid(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	inv.Devices[devID] = &CaniDeviceType{ID: devID, Name: "node-1"}

	err := inv.Validate()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestValidateNilInventory(t *testing.T) {
	var inv *Inventory

	err := inv.Validate()
	if err == nil {
		t.Error("expected error for nil inventory")
	}
}

// --- parentExists ---

func TestParentExistsTrue(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	inv.Devices[devID] = &CaniDeviceType{ID: devID, Name: "parent-device"}

	if !inv.parentExists(devID) {
		t.Error("expected parentExists to return true for existing device")
	}
}

func TestParentExistsFalse(t *testing.T) {
	inv := NewInventory()

	if inv.parentExists(uuid.New()) {
		t.Error("expected parentExists to return false for unknown UUID")
	}
}

// --- RacksByLocation ---

func TestRacksByLocation(t *testing.T) {
	inv := NewInventory()
	locID := uuid.New()
	otherLocID := uuid.New()

	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()
	inv.Racks[id1] = &CaniRackType{ID: id1, Name: "z-rack", Location: locID}
	inv.Racks[id2] = &CaniRackType{ID: id2, Name: "a-rack", Location: locID}
	inv.Racks[id3] = &CaniRackType{ID: id3, Name: "other-rack", Location: otherLocID}

	result := inv.RacksByLocation(locID)
	if len(result) != 2 {
		t.Fatalf("expected 2 racks, got %d", len(result))
	}
	if result[0].Name != "a-rack" || result[1].Name != "z-rack" {
		t.Errorf("expected sorted [a-rack, z-rack], got [%s, %s]", result[0].Name, result[1].Name)
	}
}

func TestRacksByLocationEmpty(t *testing.T) {
	inv := NewInventory()
	result := inv.RacksByLocation(uuid.New())
	if len(result) != 0 {
		t.Fatalf("expected 0 racks, got %d", len(result))
	}
}

// --- FindLocationByNameOrID ---

func TestFindLocationByNameOrIDByUUID(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Locations[id] = &CaniLocationType{ID: id, Name: "my-site"}

	got := inv.FindLocationByNameOrID(id.String())
	if got == nil || got.ID != id {
		t.Fatal("expected to find location by UUID")
	}
}

func TestFindLocationByNameOrIDByName(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Locations[id] = &CaniLocationType{ID: id, Name: "NON-MSFT"}

	got := inv.FindLocationByNameOrID("NON-MSFT")
	if got == nil || got.ID != id {
		t.Fatal("expected to find location by name")
	}
}

func TestFindLocationByNameOrIDNotFound(t *testing.T) {
	inv := NewInventory()
	got := inv.FindLocationByNameOrID("ghost")
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

// --- FindDeviceByNameOrID ---

func TestFindDeviceByNameOrIDByUUID(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Devices[id] = &CaniDeviceType{ID: id, Name: "dev-1"}

	got := inv.FindDeviceByNameOrID(id.String())
	if got == nil || got.ID != id {
		t.Fatal("expected to find device by UUID")
	}
}

func TestFindDeviceByNameOrIDByName(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Devices[id] = &CaniDeviceType{ID: id, Name: "gh-x3701u1"}

	got := inv.FindDeviceByNameOrID("gh-x3701u1")
	if got == nil || got.ID != id {
		t.Fatal("expected to find device by name")
	}
}

func TestFindDeviceByNameOrIDNotFound(t *testing.T) {
	inv := NewInventory()
	got := inv.FindDeviceByNameOrID("ghost")
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

// --- DevicesBySlug ---

func TestDevicesBySlugFound(t *testing.T) {
	inv := NewInventory()
	id1, id2 := uuid.New(), uuid.New()
	inv.Devices[id1] = &CaniDeviceType{ID: id1, Name: "b-dev", Slug: "hpe-xd670"}
	inv.Devices[id2] = &CaniDeviceType{ID: id2, Name: "a-dev", Slug: "hpe-xd670"}

	got := inv.DevicesBySlug("hpe-xd670")
	if len(got) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(got))
	}
	if got[0].Name != "a-dev" {
		t.Errorf("expected sorted by name, first=%q", got[0].Name)
	}
}

func TestDevicesBySlugCaseInsensitive(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Devices[id] = &CaniDeviceType{ID: id, Slug: "hpe-xd670"}

	got := inv.DevicesBySlug("HPE-XD670")
	if len(got) != 1 {
		t.Fatalf("expected 1 device, got %d", len(got))
	}
}

func TestDevicesBySlugEmpty(t *testing.T) {
	inv := NewInventory()
	got := inv.DevicesBySlug("nonexistent")
	if len(got) != 0 {
		t.Errorf("expected 0 devices, got %d", len(got))
	}
}

// --- OccupiedModuleBays ---

func TestOccupiedModuleBaysFound(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	modID := uuid.New()
	inv.Modules[modID] = &CaniModuleType{
		ID: modID, ParentDevice: devID, ModuleBayName: "GPU 0",
	}

	got := inv.OccupiedModuleBays(devID)
	if len(got) != 1 {
		t.Fatalf("expected 1 occupied bay, got %d", len(got))
	}
	if got["GPU 0"] != modID {
		t.Errorf("expected bay 'GPU 0' occupied by %s", modID)
	}
}

func TestOccupiedModuleBaysEmpty(t *testing.T) {
	inv := NewInventory()
	got := inv.OccupiedModuleBays(uuid.New())
	if len(got) != 0 {
		t.Errorf("expected 0 occupied bays, got %d", len(got))
	}
}

// --- AvailableModuleBays ---

func TestAvailableModuleBaysFound(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	inv.Devices[devID] = &CaniDeviceType{
		ID: devID,
		ModuleBays: []ModuleBaySpec{
			{Name: "GPU 0", Position: "GPU0"},
			{Name: "GPU 1", Position: "GPU1"},
			{Name: "PSU1", Position: "PSU1"},
		},
	}
	modID := uuid.New()
	inv.Modules[modID] = &CaniModuleType{
		ID: modID, ParentDevice: devID, ModuleBayName: "GPU 0",
	}

	got := inv.AvailableModuleBays(devID, "gpu")
	if len(got) != 1 {
		t.Fatalf("expected 1 available GPU bay, got %d", len(got))
	}
	if got[0].Name != "GPU 1" {
		t.Errorf("expected GPU 1, got %s", got[0].Name)
	}
}

func TestAvailableModuleBaysNoFilter(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	inv.Devices[devID] = &CaniDeviceType{
		ID: devID,
		ModuleBays: []ModuleBaySpec{
			{Name: "GPU 0", Position: "GPU0"},
			{Name: "PSU1", Position: "PSU1"},
		},
	}

	got := inv.AvailableModuleBays(devID, "")
	if len(got) != 2 {
		t.Fatalf("expected 2 available bays, got %d", len(got))
	}
}

func TestAvailableModuleBaysAllOccupied(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	inv.Devices[devID] = &CaniDeviceType{
		ID: devID,
		ModuleBays: []ModuleBaySpec{
			{Name: "GPU 0", Position: "GPU0"},
		},
	}
	inv.Modules[uuid.New()] = &CaniModuleType{
		ID: uuid.New(), ParentDevice: devID, ModuleBayName: "GPU 0",
	}

	got := inv.AvailableModuleBays(devID, "")
	if len(got) != 0 {
		t.Errorf("expected 0 available bays, got %d", len(got))
	}
}

// TestAllRacksSortedSkipsNil verifies AllRacks returns every non-nil rack in
// name order and that a nil inventory yields nil.
//
// Why it matters: AllRacks backs rack listings and views, so it must present a
// stable, alphabetical ordering and tolerate nil maps/receivers without
// panicking.
// Inputs: an inventory with racks "bravo" and "alpha" plus a nil map entry;
// separately, a nil *Inventory. Outputs: ["alpha","bravo"] for the populated
// case; nil for the nil receiver.
// Data choice: inserting "bravo" before "alpha" proves the sort runs rather
// than echoing map order, and the nil entry exercises the skip branch.
func TestAllRacksSortedSkipsNil(t *testing.T) {
	inv := NewInventory()
	inv.Racks[uuid.New()] = &CaniRackType{Name: "bravo"}
	inv.Racks[uuid.New()] = &CaniRackType{Name: "alpha"}
	inv.Racks[uuid.New()] = nil

	got := inv.AllRacks()
	if len(got) != 2 {
		t.Fatalf("expected 2 racks, got %d", len(got))
	}
	if got[0].Name != "alpha" || got[1].Name != "bravo" {
		t.Errorf("AllRacks() order = [%q %q], want [alpha bravo]", got[0].Name, got[1].Name)
	}

	var nilInv *Inventory
	if got := nilInv.AllRacks(); got != nil {
		t.Errorf("nil Inventory AllRacks() = %v, want nil", got)
	}
}

// TestGetCablesForDeviceNilAndBSide verifies GetCablesForDevice returns nil for
// a nil receiver, skips nil cable entries, and matches on the B-side termination.
//
// Why it matters: a device can sit on either end of a cable, so a query that
// only checked the A-side would miss half the topology; nil guards also prevent
// panics on partially built inventories.
// Inputs: a nil inventory; then a real inventory holding a nil cable and a cable
// whose TerminationBDevice is the queried device. Outputs: nil for the nil
// receiver and exactly the B-side cable for the populated case. Data choice:
// placing the device only on the B-side forces the second half of the OR
// condition that the existing A-side test never reaches.
func TestGetCablesForDeviceNilAndBSide(t *testing.T) {
	var nilInv *Inventory
	if got := nilInv.GetCablesForDevice(uuid.New()); got != nil {
		t.Errorf("nil inventory = %v, want nil", got)
	}

	inv := NewInventory()
	devID := uuid.New()
	inv.Cables[uuid.New()] = nil
	bID := uuid.New()
	inv.Cables[bID] = &CaniCableType{ID: bID, Label: "b-link", TerminationBDevice: devID}

	got := inv.GetCablesForDevice(devID)
	if len(got) != 1 || got[0].ID != bID {
		t.Fatalf("GetCablesForDevice = %v, want the B-side cable %s", got, bID)
	}
}

// TestGetInterfaceByIDNilAndModuleOwned verifies GetInterfaceByID returns nil for
// a nil receiver and resolves a module-owned interface by falling through to the
// module search.
//
// Why it matters: interfaces can live on a device directly or on one of its
// modules; the lookup must transparently find both so cabling and queries work
// regardless of where the port is mounted.
// Inputs: a nil inventory; then an inventory whose index maps the interface to a
// device that lacks it directly, while a child module actually owns it. Outputs:
// nil/nil for the nil receiver and the module's interface spec plus parent
// device for the populated case. Data choice: keeping the interface off the
// device but on its module drives the findInterfaceInModules fallback that the
// device-owned test bypasses.
func TestGetInterfaceByIDNilAndModuleOwned(t *testing.T) {
	var nilInv *Inventory
	if spec, dev := nilInv.GetInterfaceByID(uuid.New()); spec != nil || dev != nil {
		t.Errorf("nil inventory = (%v, %v), want (nil, nil)", spec, dev)
	}

	inv := NewInventory()
	devID := uuid.New()
	modID := uuid.New()
	ifaceID := uuid.New()
	inv.Devices[devID] = &CaniDeviceType{ID: devID, Name: "switch-1"} // no direct interface
	inv.Modules[modID] = &CaniModuleType{
		ID:           modID,
		ParentDevice: devID,
		Interfaces:   []InterfaceSpec{{ID: ifaceID, Name: "ge-0/0/1", Type: "1000base-t"}},
	}
	inv.Interfaces[ifaceID] = &CaniInterface{ID: ifaceID, DeviceID: devID}

	spec, dev := inv.GetInterfaceByID(ifaceID)
	if spec == nil || spec.Name != "ge-0/0/1" {
		t.Fatalf("spec = %v, want the module interface ge-0/0/1", spec)
	}
	if dev == nil || dev.ID != devID {
		t.Errorf("parent device = %v, want %s", dev, devID)
	}
}

// TestParentExistsAllKinds verifies parentExists reports membership across all
// four entity maps (device, module, rack, location) and rejects an unknown UUID.
//
// Why it matters: referential-integrity checks accept any of these as a valid
// parent, so each map must be consulted; a regression that dropped one would let
// dangling references slip through validation.
// Inputs: an inventory seeded with one device, module, rack, and location, each
// probed by its UUID, plus an unrelated UUID. Outputs: true for all four seeded
// IDs and false for the stranger. Data choice: one entry per map is the minimum
// that independently exercises every return-true branch.
func TestParentExistsAllKinds(t *testing.T) {
	inv := NewInventory()
	devID, modID, rackID, locID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Devices[devID] = &CaniDeviceType{ID: devID}
	inv.Modules[modID] = &CaniModuleType{ID: modID}
	inv.Racks[rackID] = &CaniRackType{ID: rackID}
	inv.Locations[locID] = &CaniLocationType{ID: locID}

	for name, id := range map[string]uuid.UUID{"device": devID, "module": modID, "rack": rackID, "location": locID} {
		if !inv.parentExists(id) {
			t.Errorf("parentExists(%s) = false, want true", name)
		}
	}
	if inv.parentExists(uuid.New()) {
		t.Error("parentExists(unknown) = true, want false")
	}
}

// TestFindersNilReceiver verifies the name/label finders and cross-reference
// queries return nil when called on a nil inventory.
//
// Why it matters: these helpers are invoked on optional inventories that may be
// nil before load; the nil-receiver guard must prevent a panic and yield a clean
// empty result so callers can treat "no inventory" like "no match".
// Inputs: a nil *Inventory receiver against every finder. Outputs: nil from each.
// Data choice: covering the whole finder family in one test exercises each
// function's identical nil-guard branch that the populated tests never hit.
func TestFindersNilReceiver(t *testing.T) {
	var inv *Inventory
	id := uuid.New()
	if inv.FindLocationByName("x") != nil {
		t.Error("FindLocationByName(nil) should be nil")
	}
	if inv.FindLocationByNameOrID("x") != nil {
		t.Error("FindLocationByNameOrID(nil) should be nil")
	}
	if inv.FindRackByName("x") != nil {
		t.Error("FindRackByName(nil) should be nil")
	}
	if inv.FindModuleByName("x") != nil {
		t.Error("FindModuleByName(nil) should be nil")
	}
	if inv.FindFruByName("x") != nil {
		t.Error("FindFruByName(nil) should be nil")
	}
	if inv.FindCableByLabel("x") != nil {
		t.Error("FindCableByLabel(nil) should be nil")
	}
	if inv.FindDeviceByNameOrID("x") != nil {
		t.Error("FindDeviceByNameOrID(nil) should be nil")
	}
	if inv.GetDevicesInRack(id) != nil {
		t.Error("GetDevicesInRack(nil) should be nil")
	}
	if inv.GetModulesForDevice(id) != nil {
		t.Error("GetModulesForDevice(nil) should be nil")
	}
	if inv.RacksByLocation(id) != nil {
		t.Error("RacksByLocation(nil) should be nil")
	}
	if inv.DevicesBySlug("x") != nil {
		t.Error("DevicesBySlug(nil) should be nil")
	}
	if inv.Exists("x") {
		t.Error("Exists(nil) should be false")
	}
	if _, ok := inv.FindName("x"); ok {
		t.Error("FindName(nil) should be false")
	}
}

// TestFindByNameOrIDBranches verifies the NameOrID finders reject an empty ref,
// resolve a valid UUID string, and that FindConnectableByNameOrID falls back to
// a module match.
//
// Why it matters: callers pass either a UUID or a human name interchangeably, so
// both resolution paths must work, and an empty ref must fail fast rather than
// scan; connectable resolution must reach modules so cables can terminate on
// them.
// Inputs: a populated inventory queried with "" (empty), valid UUID strings for
// a location and device, a module name, and an unknown name. Outputs: nil for
// empty, the matching entities for UUIDs, the module's ID, and uuid.Nil for the
// unknown. Data choice: distinct seeded entities isolate the UUID-parse-success
// branch and the device-miss→module-hit fallback that name-only tests skip.
func TestFindByNameOrIDBranches(t *testing.T) {
	inv := NewInventory()
	locID := uuid.New()
	inv.Locations[locID] = &CaniLocationType{ID: locID, Name: "site-a"}
	devID := uuid.New()
	inv.Devices[devID] = &CaniDeviceType{ID: devID, Name: "dev-a"}
	modID := uuid.New()
	inv.Modules[modID] = &CaniModuleType{ID: modID, Name: "mod-a"}

	if inv.FindLocationByNameOrID("") != nil {
		t.Error("empty ref should yield nil location")
	}
	if inv.FindDeviceByNameOrID("") != nil {
		t.Error("empty ref should yield nil device")
	}
	if got := inv.FindLocationByNameOrID(locID.String()); got == nil || got.ID != locID {
		t.Error("FindLocationByNameOrID should resolve a valid UUID")
	}
	if got := inv.FindDeviceByNameOrID(devID.String()); got == nil || got.ID != devID {
		t.Error("FindDeviceByNameOrID should resolve a valid UUID")
	}
	if got := inv.FindConnectableByNameOrID("mod-a"); got != modID {
		t.Errorf("FindConnectableByNameOrID(module) = %s, want %s", got, modID)
	}
	if got := inv.FindConnectableByNameOrID("nope"); got != uuid.Nil {
		t.Errorf("FindConnectableByNameOrID(unknown) = %s, want uuid.Nil", got)
	}
}

// TestValidateSkipsNilEntries verifies Validate tolerates nil entries in every
// entity map and reports no broken references.
//
// Why it matters: maps can briefly hold nil slots during incremental builds, and
// validation must skip them rather than dereference a nil pointer and crash.
// Inputs: an inventory with a single nil entry seeded into Devices, Locations,
// Racks, Modules, Cables, and Frus. Outputs: a nil error from Validate. Data
// choice: one nil per map drives the nil-continue guard inside all six
// validate*Refs helpers in a single pass.
func TestValidateSkipsNilEntries(t *testing.T) {
	inv := NewInventory()
	inv.Devices[uuid.New()] = nil
	inv.Locations[uuid.New()] = nil
	inv.Racks[uuid.New()] = nil
	inv.Modules[uuid.New()] = nil
	inv.Cables[uuid.New()] = nil
	inv.Frus[uuid.New()] = nil

	if err := inv.Validate(); err != nil {
		t.Errorf("Validate with only nil entries should pass, got: %v", err)
	}
}
