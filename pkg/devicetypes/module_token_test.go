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
package devicetypes

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// tokenInventory returns an inventory with one device whose module bays are
// "PCIe1" (position "PCIe1") and "Slot 1" (position "1").
func tokenInventory() (*Inventory, uuid.UUID) {
	inv := NewInventory()
	dev := &CaniDeviceType{ID: uuid.New(), Name: "node-01", ModuleBays: []ModuleBaySpec{
		{Name: "PCIe1", Position: "PCIe1"},
		{Name: "Slot 1", Position: "1"},
	}}
	inv.Devices[dev.ID] = dev
	return inv, dev.ID
}

// tokenModule returns a module whose two ports use the {module} placeholder.
func tokenModule(parent uuid.UUID, bay string) *CaniModuleType {
	return &CaniModuleType{ID: uuid.New(), ParentDevice: parent, ModuleBayName: bay, Interfaces: []InterfaceSpec{
		{Name: "HSN {module}/0", Label: "port {module}"},
		{Name: "HSN {module}/1"},
	}}
}

// TestResolveModuleNames verifies {module} becomes the bay position, falls
// back to the bay name, and is left alone when the module has no bay.
//
// Why it matters: forge-devicetype-library module types name their ports with
// Nautobot's {module} placeholder; without resolution cani would create ports
// literally named "HSN {module}/0". Inputs: the same module in a bay whose
// position equals its name, one whose position differs, an undeclared bay,
// and no bay. Outputs: the resolved first port name for each.
func TestResolveModuleNames(t *testing.T) {
	inv, devID := tokenInventory()
	cases := []struct{ bay, want string }{
		{"PCIe1", "HSN PCIe1/0"},
		{"slot 1", "HSN 1/0"},
		{"OCP", "HSN OCP/0"},
		{"", "HSN {module}/0"},
	}
	for _, c := range cases {
		mod := tokenModule(devID, c.bay)
		inv.ResolveModuleNames(mod)
		if mod.Interfaces[0].Name != c.want {
			t.Errorf("bay %q: name = %q, want %q", c.bay, mod.Interfaces[0].Name, c.want)
		}
	}
	mod := tokenModule(devID, "PCIe1")
	inv.ResolveModuleNames(mod)
	if mod.Interfaces[0].Label != "port PCIe1" || mod.Interfaces[1].Name != "HSN PCIe1/1" {
		t.Errorf("label/second port = %q, %q", mod.Interfaces[0].Label, mod.Interfaces[1].Name)
	}
}

// TestResolveModuleNamesRebaysFromLibrary verifies a module moved to another
// bay is renamed from its library type, and the library keeps its
// placeholders even when the instance shares its interface slice.
//
// Why it matters: "update module --bay" must rename ports, and resolving one
// module must never rename the ports of every later module of that type.
// Inputs: a registered library type with {module} ports; an instance built by
// struct copy, added in PCIe1 and then moved to Slot 1. Outputs: "HSN PCIe1/0",
// then "HSN 1/0", with the library still "HSN {module}/0".
func TestResolveModuleNamesRebaysFromLibrary(t *testing.T) {
	const slug = "test-token-nic"
	tmpl := *tokenModule(uuid.Nil, "")
	tmpl.Slug = slug
	RegisterModuleType(tmpl)
	t.Cleanup(func() { delete(allModuleTypes, slug) })

	inv, devID := tokenInventory()
	lib, _ := GetModuleBySlug(slug)
	mod := lib // struct copy shares the Interfaces slice with the library
	mod.ID, mod.ParentDevice, mod.ModuleBayName = uuid.New(), devID, "PCIe1"
	if err := inv.AddModule(&mod); err != nil {
		t.Fatalf("AddModule: %v", err)
	}
	if mod.Interfaces[0].Name != "HSN PCIe1/0" {
		t.Fatalf("after add: %q, want HSN PCIe1/0", mod.Interfaces[0].Name)
	}
	mod.ModuleBayName = "Slot 1"
	inv.ResolveModuleNames(&mod)
	if mod.Interfaces[0].Name != "HSN 1/0" {
		t.Errorf("after move: %q, want HSN 1/0", mod.Interfaces[0].Name)
	}
	if again, _ := GetModuleBySlug(slug); again.Interfaces[0].Name != "HSN {module}/0" ||
		again.Interfaces[0].NameTemplate != nil || again.Interfaces[0].LabelTemplate != nil {
		t.Errorf("library mutated: %+v", again.Interfaces[0])
	}
}

func customTokenModule(t *testing.T, parent uuid.UUID) *CaniModuleType {
	t.Helper()
	const slug = "test-custom-token-nic"
	tmpl := *tokenModule(uuid.Nil, "")
	tmpl.Slug = slug
	RegisterModuleType(tmpl)
	t.Cleanup(func() { delete(allModuleTypes, slug) })
	mod := tmpl.NewInstance()
	mod.Name, mod.ParentDevice, mod.ModuleBayName = "custom-nic", parent, "PCIe1"
	mod.Interfaces = append([]InterfaceSpec{{
		Name: "custom-{module}/0", Label: "custom {module}", Type: InterfacesElemTypeA25GbaseXSfp28,
		Description: "custom port", MacAddress: "02:00:00:00:00:01",
	}}, mod.Interfaces...)
	return mod
}

func TestResolveModuleNamesCustomPortsSurviveMove(t *testing.T) {
	inv, deviceID := tokenInventory()
	mod := customTokenModule(t, deviceID)
	if err := inv.AddModule(mod); err != nil {
		t.Fatal(err)
	}
	if mod.Interfaces[0].Name != "custom-PCIe1/0" || mod.Interfaces[0].Label != "custom PCIe1" {
		t.Fatalf("initial resolution=%+v", mod.Interfaces[0])
	}
	mod.ModuleBayName = "Slot 1"
	inv.ResolveModuleNames(mod)
	if mod.Interfaces[0].Name != "custom-1/0" || mod.Interfaces[0].Label != "custom 1" {
		t.Errorf("after move=%+v; want custom-1/0 and custom 1", mod.Interfaces[0])
	}
}

func assertCustomTemplatePort(t *testing.T, mod *CaniModuleType, position string, id uuid.UUID) {
	t.Helper()
	port := mod.Interfaces[0]
	if port.Name != "custom-"+position+"/0" || port.Label != "custom "+position {
		t.Fatalf("resolved name/label=%q/%q, position=%q", port.Name, port.Label, position)
	}
	if port.NameTemplate == nil || *port.NameTemplate != "custom-{module}/0" {
		t.Fatalf("missing original name template: %+v", port)
	}
	if port.LabelTemplate == nil || *port.LabelTemplate != "custom {module}" {
		t.Fatalf("missing original label template: %+v", port)
	}
	if port.ID != id || port.Description != "custom port" || port.MacAddress != "02:00:00:00:00:01" {
		t.Fatalf("port identity or metadata changed: %+v", port)
	}
}

func TestResolveModuleNamesCustomPortsJSONRoundTrip(t *testing.T) {
	inv, deviceID := tokenInventory()
	mod := customTokenModule(t, deviceID)
	if err := inv.AddModule(mod); err != nil {
		t.Fatal(err)
	}
	id := mod.Interfaces[0].ID
	data, err := json.Marshal(inv)
	if err != nil {
		t.Fatal(err)
	}
	var loaded Inventory
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}
	mod = loaded.Modules[mod.ID]
	if mod == nil {
		t.Fatal("module missing after JSON round trip")
	}
	mod.ModuleBayName = "Slot 1"
	loaded.ResolveModuleNames(mod)
	assertCustomTemplatePort(t, mod, "1", id)
}

func TestResolveModuleNamesCustomPortsYAMLRoundTrip(t *testing.T) {
	inv, deviceID := tokenInventory()
	mod := customTokenModule(t, deviceID)
	if err := inv.AddModule(mod); err != nil {
		t.Fatal(err)
	}
	id := mod.Interfaces[0].ID
	data, err := yaml.Marshal(inv)
	if err != nil {
		t.Fatal(err)
	}
	var loaded Inventory
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}
	mod = loaded.Modules[mod.ID]
	if mod == nil {
		t.Fatal("module missing after YAML round trip")
	}
	mod.ModuleBayName = "Slot 1"
	loaded.ResolveModuleNames(mod)
	assertCustomTemplatePort(t, mod, "1", id)
}

func TestResolveModuleNamesCustomPortsRepeatedMoves(t *testing.T) {
	inv, deviceID := tokenInventory()
	mod := customTokenModule(t, deviceID)
	if err := inv.AddModule(mod); err != nil {
		t.Fatal(err)
	}
	id := mod.Interfaces[0].ID
	for _, move := range []struct{ bay, position string }{
		{"Slot 1", "1"}, {"PCIe1", "PCIe1"}, {"Slot 1", "1"},
	} {
		mod.ModuleBayName = move.bay
		inv.ResolveModuleNames(mod)
		assertCustomTemplatePort(t, mod, move.position, id)
	}
}

func TestResolveModuleNamesTemplatesFollowReorderedPorts(t *testing.T) {
	inv, deviceID := tokenInventory()
	mod := customTokenModule(t, deviceID)
	if err := inv.AddModule(mod); err != nil {
		t.Fatal(err)
	}
	id := mod.Interfaces[0].ID
	mod.Interfaces[0], mod.Interfaces[2] = mod.Interfaces[2], mod.Interfaces[0]
	mod.ModuleBayName = "Slot 1"
	inv.ResolveModuleNames(mod)
	port := mod.GetInterface("custom-1/0")
	if port == nil || port.ID != id || port.Label != "custom 1" {
		t.Fatalf("template did not follow its port: %+v", port)
	}
}

func TestResolveModuleNamesSavedTemplatesOverrideChangedLibrary(t *testing.T) {
	inv, deviceID := tokenInventory()
	mod := customTokenModule(t, deviceID)
	if err := inv.AddModule(mod); err != nil {
		t.Fatal(err)
	}
	id := mod.Interfaces[0].ID
	tmpl, _ := GetModuleBySlug(mod.Slug)
	tmpl.Interfaces = append(tmpl.Interfaces, InterfaceSpec{Name: "new-{module}"})
	RegisterModuleType(tmpl)
	mod.ModuleBayName = "Slot 1"
	inv.ResolveModuleNames(mod)
	assertCustomTemplatePort(t, mod, "1", id)
}

func TestResolveModuleNamesLegacyCustomLiteralsAreNotGuessed(t *testing.T) {
	inv, deviceID := tokenInventory()
	mod := customTokenModule(t, deviceID)
	mod.Interfaces[0].Name, mod.Interfaces[0].Label = "custom-PCIe1/0", "custom PCIe1"
	mod.ModuleBayName = "Slot 1"
	inv.ResolveModuleNames(mod)
	port := mod.Interfaces[0]
	if port.Name != "custom-PCIe1/0" || port.Label != "custom PCIe1" {
		t.Fatalf("guessed lost template: %+v", port)
	}
	if port.NameTemplate != nil || port.LabelTemplate != nil {
		t.Fatalf("invented template for literal fields: %+v", port)
	}
}

func TestResolveModuleNamesRecoversLegacyLibraryTemplates(t *testing.T) {
	inv, deviceID := tokenInventory()
	mod := customTokenModule(t, deviceID)
	tmpl, _ := GetModuleBySlug(mod.Slug)
	mod.Interfaces = newInstanceInterfaces(tmpl.Interfaces)
	mod.Interfaces[0].Name, mod.Interfaces[0].Label = "HSN PCIe1/0", "port PCIe1"
	mod.Interfaces[1].Name = "HSN PCIe1/1"
	mod.ModuleBayName = "Slot 1"
	inv.ResolveModuleNames(mod)
	port := mod.Interfaces[0]
	if port.Name != "HSN 1/0" || port.Label != "port 1" {
		t.Fatalf("legacy library template not recovered: %+v", port)
	}
	if port.NameTemplate == nil || *port.NameTemplate != "HSN {module}/0" {
		t.Fatalf("recovered name template not retained: %+v", port)
	}
	if port.LabelTemplate == nil || *port.LabelTemplate != "port {module}" {
		t.Fatalf("recovered label template not retained: %+v", port)
	}
}

func TestModuleInstanceCopiesInterfaceTemplates(t *testing.T) {
	name, label := "HSN {module}/0", "port {module}"
	tmpl := &CaniModuleType{Interfaces: []InterfaceSpec{{NameTemplate: &name, LabelTemplate: &label}}}
	first, second := tmpl.NewInstance(), tmpl.NewInstance()
	if first.Interfaces[0].NameTemplate == nil || first.Interfaces[0].LabelTemplate == nil {
		t.Fatal("instance dropped template fields")
	}
	*first.Interfaces[0].NameTemplate = "custom-{module}"
	*first.Interfaces[0].LabelTemplate = ""
	if name != "HSN {module}/0" || label != "port {module}" {
		t.Fatal("instance template mutation reached source")
	}
	if *second.Interfaces[0].NameTemplate != name || *second.Interfaces[0].LabelTemplate != label {
		t.Fatal("instance template mutation reached sibling")
	}
}

// TestModuleBaysOfLibraryFallback verifies bay positions come from the
// device's library type when the instance has no bay specs.
func TestModuleBaysOfLibraryFallback(t *testing.T) {
	for slug, dt := range allDeviceTypes {
		if len(dt.ModuleBays) == 0 {
			continue
		}
		got := moduleBaysOf(&CaniDeviceType{Slug: slug})
		if len(got) != len(dt.ModuleBays) {
			t.Errorf("moduleBaysOf(%s) = %d bays, want %d", slug, len(got), len(dt.ModuleBays))
		}
		return
	}
	t.Skip("no library device type with module bays")
}
