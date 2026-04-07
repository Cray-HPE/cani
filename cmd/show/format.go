/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
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
package show

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
)

// buildLocationTree builds tree nodes for the location hierarchy.
func buildLocationTree(inv *devicetypes.Inventory) []visual.TreeNode {
	if inv == nil {
		return nil
	}

	// Find root locations (no parent)
	var roots []*devicetypes.CaniLocationType
	for _, loc := range inv.Locations {
		if loc.Parent == uuid.Nil {
			roots = append(roots, loc)
		}
	}
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].Name < roots[j].Name
	})

	nodes := make([]visual.TreeNode, 0, len(roots))
	for _, root := range roots {
		nodes = append(nodes, locationToTreeNode(root, inv))
	}
	return nodes
}

// locationToTreeNode recursively converts a location and its children into a tree node.
func locationToTreeNode(loc *devicetypes.CaniLocationType, inv *devicetypes.Inventory) visual.TreeNode {
	node := visual.TreeNode{
		Label:  loc.Name,
		Detail: loc.LocationType,
	}

	// Add child locations
	childLocs := resolveLocationChildren(loc.Children, inv)
	for _, child := range childLocs {
		node.Children = append(node.Children, locationToTreeNode(child, inv))
	}

	// Add racks under this location
	rackList := resolveRacks(loc.Racks, inv)
	for _, rack := range rackList {
		node.Children = append(node.Children, rackToTreeNode(rack, inv))
	}

	return node
}

// rackToTreeNode converts a rack and its child devices into a tree node.
func rackToTreeNode(rack *devicetypes.CaniRackType, inv *devicetypes.Inventory) visual.TreeNode {
	detail := "rack"
	if rack.Status != "" {
		detail += " " + rack.Status
	}
	if rack.UHeight > 0 {
		detail += fmt.Sprintf(" %dU", rack.UHeight)
	}
	node := visual.TreeNode{
		Label:  rack.Name,
		Detail: detail,
	}

	devices := resolveDevices(rack.Devices, inv)
	for _, dev := range devices {
		node.Children = append(node.Children, deviceToTreeNode(dev, inv))
	}
	return node
}

// deviceToTreeNode converts a device and its modules/children into a tree node.
func deviceToTreeNode(dev *devicetypes.CaniDeviceType, inv *devicetypes.Inventory) visual.TreeNode {
	var details []string
	if dev.HardwareType != "" {
		details = append(details, dev.HardwareType)
	}
	if dev.RackPosition > 0 {
		details = append(details, fmt.Sprintf("U%d", dev.RackPosition))
	}
	if dev.Status != "" {
		details = append(details, dev.Status)
	}
	if dev.Model != "" {
		details = append(details, dev.Model)
	}
	node := visual.TreeNode{
		Label:  dev.Name,
		Detail: joinNonEmpty(details, " "),
	}

	// Add child devices (e.g. blades in a chassis)
	childDevs := resolveDeviceChildren(dev.Children, inv)
	for _, child := range childDevs {
		node.Children = append(node.Children, deviceToTreeNode(child, inv))
	}

	// Add modules
	modules := findModulesForDevice(dev.ID, inv)
	for _, mod := range modules {
		node.Children = append(node.Children, moduleToTreeNode(mod, inv))
	}

	// Add interfaces directly on the device
	for _, iface := range dev.Interfaces {
		node.Children = append(node.Children, interfaceToTreeNode(iface))
	}

	// Add FRUs
	frus := findFrusForDevice(dev.ID, inv)
	for _, fru := range frus {
		node.Children = append(node.Children, fruToTreeNode(fru))
	}

	// Add cables connected to this device
	cables := findCablesForDevice(dev.ID, inv)
	for _, c := range cables {
		node.Children = append(node.Children, cableLeafNode(c, inv))
	}

	return node
}

// moduleToTreeNode converts a module into a tree node with its interfaces and FRUs.
func moduleToTreeNode(mod *devicetypes.CaniModuleType, inv *devicetypes.Inventory) visual.TreeNode {
	var details []string
	if mod.HardwareType != "" {
		details = append(details, mod.HardwareType)
	}
	if mod.ModuleBayName != "" {
		details = append(details, "bay:"+mod.ModuleBayName)
	}
	if mod.Status != "" {
		details = append(details, mod.Status)
	}
	node := visual.TreeNode{
		Label:  mod.Name,
		Detail: joinNonEmpty(details, " "),
	}

	// Add interfaces on the module
	for _, iface := range mod.Interfaces {
		node.Children = append(node.Children, interfaceToTreeNode(iface))
	}

	// Add FRUs belonging to this module
	for _, fruID := range mod.Frus {
		if inv != nil && inv.Frus != nil {
			if fru, ok := inv.Frus[fruID]; ok {
				node.Children = append(node.Children, fruToTreeNode(fru))
			}
		}
	}

	return node
}

// interfaceToTreeNode converts an interface spec into a tree leaf.
func interfaceToTreeNode(iface devicetypes.InterfaceSpec) visual.TreeNode {
	detail := string(iface.Type)
	if iface.Label != "" {
		detail += " " + iface.Label
	}
	return visual.TreeNode{
		Label:  iface.Name,
		Detail: detail,
	}
}

// fruToTreeNode converts a FRU into a tree leaf.
func fruToTreeNode(fru *devicetypes.CaniFruType) visual.TreeNode {
	var details []string
	if fru.HardwareType != "" {
		details = append(details, fru.HardwareType)
	}
	if fru.PartNumber != "" {
		details = append(details, fru.PartNumber)
	}
	return visual.TreeNode{
		Label:  fru.Name,
		Detail: joinNonEmpty(details, " "),
	}
}

// cableLeafNode converts a cable into a tree leaf showing the connection.
func cableLeafNode(c *devicetypes.CaniCableType, inv *devicetypes.Inventory) visual.TreeNode {
	label := c.Label
	if label == "" {
		label = c.CableType
	}
	var details []string
	if c.CableType != "" {
		details = append(details, c.CableType)
	}
	bTerm := formatTermination(c.TerminationBDevice, c.TerminationBPort, inv)
	if bTerm != "-" {
		details = append(details, "→ "+bTerm)
	}
	return visual.TreeNode{
		Label:  label,
		Detail: joinNonEmpty(details, " "),
	}
}

// buildRackTree builds tree nodes for racks → devices → modules.
func buildRackTree(racks []*devicetypes.CaniRackType, inv *devicetypes.Inventory) []visual.TreeNode {
	nodes := make([]visual.TreeNode, 0, len(racks))
	for _, rack := range racks {
		nodes = append(nodes, rackToTreeNode(rack, inv))
	}
	return nodes
}

// buildDeviceTree builds tree nodes for devices → modules.
func buildDeviceTree(devices []*devicetypes.CaniDeviceType, inv *devicetypes.Inventory) []visual.TreeNode {
	nodes := make([]visual.TreeNode, 0, len(devices))
	for _, dev := range devices {
		nodes = append(nodes, deviceToTreeNode(dev, inv))
	}
	return nodes
}

// buildModuleTree groups modules by parent device as tree nodes.
func buildModuleTree(modules []*devicetypes.CaniModuleType, inv *devicetypes.Inventory) []visual.TreeNode {
	groups := make(map[uuid.UUID][]*devicetypes.CaniModuleType)
	var order []uuid.UUID
	for _, mod := range modules {
		pid := mod.ParentDevice
		if _, seen := groups[pid]; !seen {
			order = append(order, pid)
		}
		groups[pid] = append(groups[pid], mod)
	}

	nodes := make([]visual.TreeNode, 0, len(order))
	for _, pid := range order {
		parentLabel := resolveDeviceName(pid, inv)
		if parentLabel == "" {
			parentLabel = "(unassigned)"
		}
		node := visual.TreeNode{Label: parentLabel, Detail: "device"}
		for _, mod := range groups[pid] {
			node.Children = append(node.Children, moduleToTreeNode(mod, inv))
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// buildCableTree groups cables by A-termination device as tree nodes.
func buildCableTree(cables []*devicetypes.CaniCableType, inv *devicetypes.Inventory) []visual.TreeNode {
	groups := make(map[uuid.UUID][]*devicetypes.CaniCableType)
	var order []uuid.UUID
	for _, c := range cables {
		did := c.TerminationADevice
		if _, seen := groups[did]; !seen {
			order = append(order, did)
		}
		groups[did] = append(groups[did], c)
	}

	nodes := make([]visual.TreeNode, 0, len(order))
	for _, did := range order {
		devLabel := resolveDeviceName(did, inv)
		if devLabel == "" {
			devLabel = "(unconnected)"
		}
		node := visual.TreeNode{Label: devLabel, Detail: "device"}
		for _, c := range groups[did] {
			node.Children = append(node.Children, cableLeafNode(c, inv))
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// buildFullTree builds the complete inventory tree: Locations → Racks → Devices → Modules.
func buildFullTree(inv *devicetypes.Inventory) []visual.TreeNode {
	var roots []visual.TreeNode

	// Locations (includes nested racks/devices/modules)
	locNodes := buildLocationTree(inv)
	if len(locNodes) > 0 {
		roots = append(roots, visual.TreeNode{
			Label:    fmt.Sprintf("Locations (%d)", len(inv.Locations)),
			Children: locNodes,
		})
	}

	// Orphan racks (not under any location)
	var orphanRacks []*devicetypes.CaniRackType
	for _, rack := range inv.Racks {
		if rack.Location == uuid.Nil {
			orphanRacks = append(orphanRacks, rack)
		}
	}
	sort.Slice(orphanRacks, func(i, j int) bool {
		return orphanRacks[i].Name < orphanRacks[j].Name
	})
	if len(orphanRacks) > 0 {
		roots = append(roots, visual.TreeNode{
			Label:    fmt.Sprintf("Racks (%d)", len(orphanRacks)),
			Children: buildRackTree(orphanRacks, inv),
		})
	}

	// Orphan devices (not in any rack)
	var orphanDevices []*devicetypes.CaniDeviceType
	for _, dev := range inv.Devices {
		if dev.Rack == uuid.Nil && dev.ParentDevice == uuid.Nil {
			orphanDevices = append(orphanDevices, dev)
		}
	}
	sort.Slice(orphanDevices, func(i, j int) bool {
		return orphanDevices[i].Name < orphanDevices[j].Name
	})
	if len(orphanDevices) > 0 {
		roots = append(roots, visual.TreeNode{
			Label:    fmt.Sprintf("Devices (%d)", len(orphanDevices)),
			Children: buildDeviceTree(orphanDevices, inv),
		})
	}

	// Cables
	if len(inv.Cables) > 0 {
		var cables []*devicetypes.CaniCableType
		for _, c := range inv.Cables {
			cables = append(cables, c)
		}
		sort.Slice(cables, func(i, j int) bool {
			return cables[i].Label < cables[j].Label
		})
		roots = append(roots, visual.TreeNode{
			Label:    fmt.Sprintf("Cables (%d)", len(inv.Cables)),
			Children: buildCableTree(cables, inv),
		})
	}

	return roots
}

// renderTreeOutput is a helper that renders tree nodes to stdout.
func renderTreeOutput(nodes []visual.TreeNode) {
	opts := visual.TreeOptions{Writer: os.Stdout}
	visual.RenderTreeToStdout(nodes, opts)
}

// resolveLocationChildren resolves a slice of location UUIDs into sorted location pointers.
func resolveLocationChildren(ids []uuid.UUID, inv *devicetypes.Inventory) []*devicetypes.CaniLocationType {
	if inv == nil || inv.Locations == nil {
		return nil
	}
	result := make([]*devicetypes.CaniLocationType, 0, len(ids))
	for _, id := range ids {
		if loc, ok := inv.Locations[id]; ok {
			result = append(result, loc)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// resolveRacks resolves a slice of rack UUIDs into sorted rack pointers.
func resolveRacks(ids []uuid.UUID, inv *devicetypes.Inventory) []*devicetypes.CaniRackType {
	if inv == nil || inv.Racks == nil {
		return nil
	}
	result := make([]*devicetypes.CaniRackType, 0, len(ids))
	for _, id := range ids {
		if r, ok := inv.Racks[id]; ok {
			result = append(result, r)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// resolveDevices resolves a slice of device UUIDs into sorted device pointers.
func resolveDevices(ids []uuid.UUID, inv *devicetypes.Inventory) []*devicetypes.CaniDeviceType {
	if inv == nil || inv.Devices == nil {
		return nil
	}
	result := make([]*devicetypes.CaniDeviceType, 0, len(ids))
	for _, id := range ids {
		if d, ok := inv.Devices[id]; ok {
			result = append(result, d)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// resolveDeviceChildren resolves child device UUIDs into sorted device pointers.
func resolveDeviceChildren(ids []uuid.UUID, inv *devicetypes.Inventory) []*devicetypes.CaniDeviceType {
	return resolveDevices(ids, inv)
}

// findModulesForDevice returns all modules whose ParentDevice matches deviceID.
func findModulesForDevice(deviceID uuid.UUID, inv *devicetypes.Inventory) []*devicetypes.CaniModuleType {
	if inv == nil || inv.Modules == nil {
		return nil
	}
	var result []*devicetypes.CaniModuleType
	for _, mod := range inv.Modules {
		if mod.ParentDevice == deviceID {
			result = append(result, mod)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// printAllTables renders all inventory sections as tables with section headers.
func printAllTables(inv *devicetypes.Inventory) {
	// Locations
	if len(inv.Locations) > 0 {
		locations := sortedLocations(inv)
		fmt.Printf("\nLocations (%d):\n", len(locations))
		printLocationTable(locations, inv)
	}

	// Racks
	if len(inv.Racks) > 0 {
		racks := sortedRacks(inv)
		fmt.Printf("\nRacks (%d):\n", len(racks))
		printRackTable(racks, inv)
	}

	// Devices
	if len(inv.Devices) > 0 {
		devices := sortedDevices(inv)
		fmt.Printf("\nDevices (%d):\n", len(devices))
		printDeviceTable(devices, inv)
	}

	// Modules
	if len(inv.Modules) > 0 {
		modules := sortedModules(inv)
		fmt.Printf("\nModules (%d):\n", len(modules))
		printModuleTable(modules, inv)
	}

	// Cables
	if len(inv.Cables) > 0 {
		cables := sortedCables(inv)
		fmt.Printf("\nCables (%d):\n", len(cables))
		printCableTableFromTypes(cables, inv)
	}

	fmt.Println()
}

func sortedLocations(inv *devicetypes.Inventory) []*devicetypes.CaniLocationType {
	locs := make([]*devicetypes.CaniLocationType, 0, len(inv.Locations))
	for _, loc := range inv.Locations {
		locs = append(locs, loc)
	}
	sort.Slice(locs, func(i, j int) bool { return locs[i].Name < locs[j].Name })
	return locs
}

func sortedRacks(inv *devicetypes.Inventory) []*devicetypes.CaniRackType {
	racks := make([]*devicetypes.CaniRackType, 0, len(inv.Racks))
	for _, r := range inv.Racks {
		racks = append(racks, r)
	}
	sort.Slice(racks, func(i, j int) bool { return racks[i].Name < racks[j].Name })
	return racks
}

func sortedDevices(inv *devicetypes.Inventory) []*devicetypes.CaniDeviceType {
	devs := make([]*devicetypes.CaniDeviceType, 0, len(inv.Devices))
	for _, d := range inv.Devices {
		devs = append(devs, d)
	}
	sort.Slice(devs, func(i, j int) bool { return devs[i].Name < devs[j].Name })
	return devs
}

func sortedModules(inv *devicetypes.Inventory) []*devicetypes.CaniModuleType {
	mods := make([]*devicetypes.CaniModuleType, 0, len(inv.Modules))
	for _, m := range inv.Modules {
		mods = append(mods, m)
	}
	sort.Slice(mods, func(i, j int) bool { return mods[i].Name < mods[j].Name })
	return mods
}

func sortedCables(inv *devicetypes.Inventory) []*devicetypes.CaniCableType {
	cables := make([]*devicetypes.CaniCableType, 0, len(inv.Cables))
	for _, c := range inv.Cables {
		cables = append(cables, c)
	}
	sort.Slice(cables, func(i, j int) bool { return cables[i].Label < cables[j].Label })
	return cables
}

// intOrDash returns the integer as a string, or "-" if zero.
func intOrDash(n int) string {
	if n == 0 {
		return "-"
	}
	return strconv.Itoa(n)
}

// joinNonEmpty joins non-empty strings with sep.
func joinNonEmpty(parts []string, sep string) string {
	var filtered []string
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) == 0 {
		return ""
	}
	result := filtered[0]
	for _, p := range filtered[1:] {
		result += sep + p
	}
	return result
}

// findFrusForDevice returns all FRUs whose Device matches deviceID.
func findFrusForDevice(deviceID uuid.UUID, inv *devicetypes.Inventory) []*devicetypes.CaniFruType {
	if inv == nil || inv.Frus == nil {
		return nil
	}
	var result []*devicetypes.CaniFruType
	for _, fru := range inv.Frus {
		if fru.Device == deviceID {
			result = append(result, fru)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// findCablesForDevice returns all cables where TerminationADevice matches deviceID.
func findCablesForDevice(deviceID uuid.UUID, inv *devicetypes.Inventory) []*devicetypes.CaniCableType {
	if inv == nil || inv.Cables == nil {
		return nil
	}
	var result []*devicetypes.CaniCableType
	for _, c := range inv.Cables {
		if c.TerminationADevice == deviceID {
			result = append(result, c)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Label < result[j].Label
	})
	return result
}
