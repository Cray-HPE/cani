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
package visual

import (
	"fmt"
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// BuildRackTree builds tree nodes for racks → devices → modules.
func BuildRackTree(racks []*devicetypes.CaniRackType, inv *devicetypes.Inventory, tf TreeFilter) []TreeNode {
	nodes := make([]TreeNode, 0, len(racks))
	for _, rack := range racks {
		nodes = append(nodes, RackToTreeNode(rack, inv, tf))
	}
	return nodes
}

// BuildDeviceTree builds tree nodes for devices → modules.
func BuildDeviceTree(devices []*devicetypes.CaniDeviceType, inv *devicetypes.Inventory, tf TreeFilter) []TreeNode {
	nodes := make([]TreeNode, 0, len(devices))
	for _, dev := range devices {
		nodes = append(nodes, DeviceToTreeNode(dev, inv, tf))
	}
	return nodes
}

// BuildModuleTree groups modules by parent device as tree nodes.
func BuildModuleTree(modules []*devicetypes.CaniModuleType, inv *devicetypes.Inventory, tf TreeFilter) []TreeNode {
	groups := make(map[uuid.UUID][]*devicetypes.CaniModuleType)
	var order []uuid.UUID
	for _, mod := range modules {
		pid := mod.ParentDevice
		if _, seen := groups[pid]; !seen {
			order = append(order, pid)
		}
		groups[pid] = append(groups[pid], mod)
	}

	nodes := make([]TreeNode, 0, len(order))
	for _, pid := range order {
		parentLabel := ResolveDeviceName(pid, inv)
		if parentLabel == "" {
			parentLabel = "(unassigned)"
		}
		node := TreeNode{Label: parentLabel, Detail: "device"}
		for _, mod := range groups[pid] {
			node.Children = append(node.Children, ModuleToTreeNode(mod, inv, tf))
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// BuildCableTree groups cables by A-termination device as tree nodes.
func BuildCableTree(cables []*devicetypes.CaniCableType, inv *devicetypes.Inventory, tf TreeFilter) []TreeNode {
	groups := make(map[uuid.UUID][]*devicetypes.CaniCableType)
	var order []uuid.UUID
	for _, c := range cables {
		did := c.TerminationADevice
		if _, seen := groups[did]; !seen {
			order = append(order, did)
		}
		groups[did] = append(groups[did], c)
	}

	nodes := make([]TreeNode, 0, len(order))
	for _, did := range order {
		devLabel := ResolveDeviceName(did, inv)
		if devLabel == "" {
			devLabel = "(unconnected)"
		}
		node := TreeNode{Label: devLabel, Detail: "device"}
		for _, c := range groups[did] {
			node.Children = append(node.Children, CableLeafNode(c, inv, tf))
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// BuildFruTree groups FRUs by parent device as tree nodes.
func BuildFruTree(frus []*devicetypes.CaniFruType, inv *devicetypes.Inventory) []TreeNode {
	groups := make(map[uuid.UUID][]*devicetypes.CaniFruType)
	var order []uuid.UUID
	for _, f := range frus {
		did := f.Device
		if _, seen := groups[did]; !seen {
			order = append(order, did)
		}
		groups[did] = append(groups[did], f)
	}

	nodes := make([]TreeNode, 0, len(order))
	for _, did := range order {
		parentLabel := ResolveDeviceName(did, inv)
		if parentLabel == "" {
			parentLabel = "(unassigned)"
		}
		node := TreeNode{Label: parentLabel, Detail: "device"}
		for _, f := range groups[did] {
			node.Children = append(node.Children, FruToTreeNode(f, TreeFilter{}))
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// BuildInterfaceInstanceTree groups interface instances by device as tree nodes.
func BuildInterfaceInstanceTree(ifaces []*devicetypes.InterfaceInstance, inv *devicetypes.Inventory, tf TreeFilter) []TreeNode {
	groups := make(map[uuid.UUID][]*devicetypes.InterfaceInstance)
	var order []uuid.UUID
	for _, iface := range ifaces {
		did := iface.DeviceID
		if _, seen := groups[did]; !seen {
			order = append(order, did)
		}
		groups[did] = append(groups[did], iface)
	}

	nodes := make([]TreeNode, 0, len(order))
	for _, did := range order {
		parentLabel := ResolveDeviceName(did, inv)
		if parentLabel == "" {
			parentLabel = "(unassigned)"
		}
		node := TreeNode{Label: parentLabel, Detail: "device"}
		for _, iface := range groups[did] {
			detail := string(iface.InterfaceType)
			if iface.Label != "" {
				detail += " " + iface.Label
			}
			if tf.Roles && iface.Role != "" {
				detail = PipeSep(detail, "role:"+iface.Role)
			}
			node.Children = append(node.Children, TreeNode{
				Label:  iface.Name,
				Detail: detail,
			})
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// BuildFullTree builds the complete inventory tree: Locations → Racks → Devices → Modules.
func BuildFullTree(inv *devicetypes.Inventory, tf TreeFilter) []TreeNode {
	var roots []TreeNode

	locNodes := BuildLocationTree(inv, tf)
	if len(locNodes) > 0 {
		roots = append(roots, TreeNode{
			Label:    fmt.Sprintf("Locations (%d)", len(inv.Locations)),
			Children: locNodes,
		})
	}

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
		roots = append(roots, TreeNode{
			Label:    fmt.Sprintf("Racks (%d)", len(orphanRacks)),
			Children: BuildRackTree(orphanRacks, inv, tf),
		})
	}

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
		roots = append(roots, TreeNode{
			Label:    fmt.Sprintf("Devices (%d)", len(orphanDevices)),
			Children: BuildDeviceTree(orphanDevices, inv, tf),
		})
	}

	if tf.Cables && len(inv.Cables) > 0 {
		var orphanCables []*devicetypes.CaniCableType
		for _, c := range inv.Cables {
			if !CableAttachedToAnyDeviceInterface(c, inv) {
				orphanCables = append(orphanCables, c)
			}
		}
		sort.Slice(orphanCables, func(i, j int) bool {
			return orphanCables[i].Label < orphanCables[j].Label
		})
		if len(orphanCables) > 0 {
			roots = append(roots, TreeNode{
				Label:    fmt.Sprintf("Unattached Cables (%d)", len(orphanCables)),
				Children: BuildCableTree(orphanCables, inv, tf),
			})
		}
	}

	return roots
}
