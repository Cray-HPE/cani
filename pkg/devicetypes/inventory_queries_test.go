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
	inv.Interfaces[ifaceID] = &InterfaceInstance{
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

	inv.Interfaces[ifaceID] = &InterfaceInstance{
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
