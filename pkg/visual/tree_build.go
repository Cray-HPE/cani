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

// BuildLocationTree builds tree nodes for the location hierarchy.
func BuildLocationTree(inv *devicetypes.Inventory, tf TreeFilter) []TreeNode {
	if inv == nil {
		return nil
	}

	var roots []*devicetypes.CaniLocationType
	for _, loc := range inv.Locations {
		if loc.Parent == uuid.Nil {
			roots = append(roots, loc)
		}
	}
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].Name < roots[j].Name
	})

	nodes := make([]TreeNode, 0, len(roots))
	for _, root := range roots {
		nodes = append(nodes, LocationToTreeNode(root, inv, tf))
	}
	return nodes
}

// LocationToTreeNode recursively converts a location and its children into a tree node.
func LocationToTreeNode(loc *devicetypes.CaniLocationType, inv *devicetypes.Inventory, tf TreeFilter) TreeNode {
	nRacks, nDevices := CountLocationDescendants(loc, inv)
	counts := ""
	if nRacks > 0 || nDevices > 0 {
		var parts []string
		if nRacks > 0 {
			parts = append(parts, fmt.Sprintf("%d racks", nRacks))
		}
		if nDevices > 0 {
			parts = append(parts, fmt.Sprintf("%d devices", nDevices))
		}
		counts = "(" + JoinNonEmpty(parts, ", ") + ")"
	}
	node := TreeNode{
		Label:  TreeIcon(IconLocation, ColorCyan, "location", loc.Name, tf.NoColor),
		Detail: PipeSep(loc.LocationType, counts),
	}

	childLocs := ResolveLocationChildren(loc.Children, inv)
	for _, child := range childLocs {
		node.Children = append(node.Children, LocationToTreeNode(child, inv, tf))
	}

	rackList := ResolveRacks(loc.Racks, inv)
	for _, rack := range rackList {
		node.Children = append(node.Children, RackToTreeNode(rack, inv, tf))
	}

	return node
}

// RackToTreeNode converts a rack and its child devices into a tree node.
func RackToTreeNode(rack *devicetypes.CaniRackType, inv *devicetypes.Inventory, tf TreeFilter) TreeNode {
	rackModel := Trunc(rack.Model, 30)

	devices := ResolveDevices(rack.Devices, inv)
	nDevices := len(devices)
	uUsed := 0
	for _, dev := range devices {
		if dev.UHeight > 0 {
			uUsed += dev.UHeight
		} else {
			uUsed++
		}
	}
	devCount := fmt.Sprintf("%d device(s)", nDevices)
	uUsage := ""
	if rack.UHeight > 0 {
		uUsage = fmt.Sprintf("%dU/%dU", uUsed, rack.UHeight)
	}

	node := TreeNode{
		Label:  TreeIconColored(IconRack, ColorCyan, "rack", rack.Name, StatusAnsi(rack.Status), tf.NoColor),
		Detail: PipeSep(rackModel, devCount, uUsage),
	}

	if tf.EmptyUs && rack.UHeight > 0 {
		occupied := make(map[int]*devicetypes.CaniDeviceType)
		for _, dev := range devices {
			if dev.RackPosition > 0 {
				h := dev.UHeight
				if h < 1 {
					h = 1
				}
				for u := dev.RackPosition; u > dev.RackPosition-h; u-- {
					occupied[u] = dev
				}
			}
		}
		emitted := make(map[uuid.UUID]bool)
		for u := rack.UHeight; u >= 1; u-- {
			if dev, ok := occupied[u]; ok {
				if !emitted[dev.ID] {
					node.Children = append(node.Children, DeviceToTreeNode(dev, inv, tf))
					emitted[dev.ID] = true
				}
			} else {
				node.Children = append(node.Children, EmptyUNode(u, tf))
			}
		}
		for _, dev := range devices {
			if dev.RackPosition == 0 {
				node.Children = append(node.Children, DeviceToTreeNode(dev, inv, tf))
			}
		}
	} else {
		for _, dev := range devices {
			node.Children = append(node.Children, DeviceToTreeNode(dev, inv, tf))
		}
	}
	return node
}

// EmptyUNode renders a dim placeholder for an unoccupied rack unit.
func EmptyUNode(u int, tf TreeFilter) TreeNode {
	lbl := fmt.Sprintf("U%d", u)
	if tf.NoColor {
		return TreeNode{Label: IconDevice + " (device) " + lbl + " (empty)"}
	}
	return TreeNode{
		Label: ColorGray + IconDevice + " (device) " + lbl + " (empty)" + ColorReset,
	}
}
