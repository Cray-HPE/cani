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
package update

import (
	"encoding/json"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func moduleTemplateHarness(t *testing.T) (*cmdtest.Harness, *devicetypes.CaniModuleType) {
	t.Helper()
	inventory := devicetypes.NewInventory()
	deviceID, moduleID := uuid.New(), uuid.New()
	inventory.Devices[deviceID] = &devicetypes.CaniDeviceType{
		ID: deviceID, Name: "node", ModuleBays: []devicetypes.ModuleBaySpec{
			{Name: "PCIe1", Position: "PCIe1"}, {Name: "Slot 1", Position: "1"},
		},
	}
	module := &devicetypes.CaniModuleType{
		ID: moduleID, Name: "custom-nic", ParentDevice: deviceID, ModuleBayName: "PCIe1",
		Interfaces: []devicetypes.InterfaceSpec{
			{Name: "custom-{module}/0", Label: "port {module}"},
			{Name: "custom-{module}/1"},
		},
	}
	if err := inventory.AddModule(module); err != nil {
		t.Fatal(err)
	}
	cable := &devicetypes.CaniCableType{
		ID: uuid.New(), Slug: "test-cable",
		TerminationA: module.Interfaces[0].ID, TerminationADevice: deviceID, TerminationAPort: module.Interfaces[0].Name,
		TerminationB: module.Interfaces[1].ID, TerminationBDevice: moduleID, TerminationBPort: module.Interfaces[1].Name,
	}
	if err := inventory.AddCable(cable); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(inventory)
	if err != nil {
		t.Fatal(err)
	}
	var loaded devicetypes.Inventory
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}
	if loaded.Modules[moduleID] == nil {
		t.Fatal("module missing after JSON round trip")
	}
	return cmdtest.New(t, NewCommand(), newModuleCommand(), &loaded), loaded.Modules[moduleID]
}

func TestUpdateModuleBayPreservesCustomTemplates(t *testing.T) {
	harness, module := moduleTemplateHarness(t)
	if err := harness.Execute(t, "update", "module", module.Name, "--bay", "Slot 1"); err != nil {
		t.Fatal(err)
	}
	assertUpdatedModuleTemplates(t, harness, module)
}

func TestUpdateModuleSetBayPreservesCustomTemplates(t *testing.T) {
	harness, module := moduleTemplateHarness(t)
	if err := harness.Execute(t, "update", "module", module.Name, "--set", "module_bay_name=Slot 1"); err != nil {
		t.Fatal(err)
	}
	assertUpdatedModuleTemplates(t, harness, module)
}

func assertUpdatedModuleTemplates(t *testing.T, harness *cmdtest.Harness, module *devicetypes.CaniModuleType) {
	t.Helper()
	module = harness.Inventory().Modules[module.ID]
	if module == nil {
		t.Fatal("saved module missing")
	}
	port := module.Interfaces[0]
	if port.Name != "custom-1/0" || port.Label != "port 1" {
		t.Fatalf("updated port=%+v", port)
	}
	if port.NameTemplate == nil || *port.NameTemplate != "custom-{module}/0" {
		t.Fatalf("name template lost: %+v", port)
	}
	if port.LabelTemplate == nil || *port.LabelTemplate != "port {module}" {
		t.Fatalf("label template lost: %+v", port)
	}
	assertUpdatedModuleCable(t, harness.Inventory(), module)
	if harness.Store.Saves != 1 {
		t.Fatalf("saves=%d, want 1", harness.Store.Saves)
	}
}

func assertUpdatedModuleCable(t *testing.T, inventory *devicetypes.Inventory, module *devicetypes.CaniModuleType) {
	t.Helper()
	if len(inventory.Cables) != 1 {
		t.Fatalf("saved cables=%d, want 1", len(inventory.Cables))
	}
	port := module.Interfaces[0]
	for _, cable := range inventory.Cables {
		if cable.TerminationA != port.ID || cable.TerminationAPort != port.Name ||
			cable.TerminationB != module.Interfaces[1].ID || cable.TerminationBPort != module.Interfaces[1].Name ||
			cable.TerminationADevice != module.ParentDevice || cable.TerminationBDevice != module.ID {
			t.Fatalf("saved cable no longer matches module ports: %+v", cable)
		}
	}
	indexed := inventory.Interfaces[port.ID]
	if indexed == nil || indexed.Name != port.Name || indexed.Label != port.Label {
		t.Fatalf("saved interface index differs from module port: %+v", indexed)
	}
}
