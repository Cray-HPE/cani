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
	"strconv"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// Shared tree labels.
const (
	labelNoLabel   = "<no label>"
	labelModuleSep = " (module) "
)

// DeviceToTreeNode converts a device and its modules/children into a tree node.
func DeviceToTreeNode(dev *devicetypes.CaniDeviceType, inv *devicetypes.Inventory, tf TreeFilter) TreeNode {
	node := deviceBaseNode(dev, tf)

	childDevs := ResolveDeviceChildren(dev.Children, inv)
	for _, child := range childDevs {
		node.Children = append(node.Children, DeviceToTreeNode(child, inv, tf))
	}

	node.Children = append(node.Children, deviceModuleNodes(dev, inv, tf)...)
	node.Children = append(node.Children, deviceInterfaceNodes(dev, inv, tf)...)

	frus := FindFrusForDevice(dev.ID, inv)
	for _, fru := range frus {
		node.Children = append(node.Children, FruToTreeNode(fru, tf))
	}

	node.Children = append(node.Children, deviceCableNodes(dev, inv, tf)...)

	return node
}

// deviceBaseNode builds the label and detail for a device node (no children).
func deviceBaseNode(dev *devicetypes.CaniDeviceType, tf TreeFilter) TreeNode {
	uPos := ""
	if dev.RackPosition > 0 {
		if dev.UHeight > 1 {
			uPos = fmt.Sprintf("U%d-U%d", dev.RackPosition, dev.RackPosition-dev.UHeight+1)
		} else {
			uPos = fmt.Sprintf("U%d", dev.RackPosition)
		}
		if !tf.NoColor {
			uPos = ColorGray + uPos + ColorReset
		}
	}
	slug := Trunc(dev.Model, 30)
	if slug != "" && !tf.NoColor {
		slug = ColorGray + slug + ColorReset
	}
	name := dev.Name
	if name == "" {
		if tf.NoColor {
			name = labelNoLabel
		} else {
			name = ColorGray + labelNoLabel + ColorReset
		}
	} else if !tf.NoColor {
		name = StatusAnsi(dev.Status) + name + ColorReset
	}
	node := TreeNode{
		Label:  TreeIcon(IconDevice, ColorWhite, "device", uPos, tf.NoColor),
		Detail: PipeSep(name, slug),
	}
	if tf.Roles && dev.GetRole() != "" {
		node.Detail = PipeSep(name, slug, "role:"+dev.GetRole())
	}
	return node
}

// deviceModuleNodes builds child nodes for a device's module bays and modules.
func deviceModuleNodes(dev *devicetypes.CaniDeviceType, inv *devicetypes.Inventory, tf TreeFilter) []TreeNode {
	if !tf.Modules {
		return nil
	}
	modules := FindModulesForDevice(dev.ID, inv)
	populated := make(map[string]*devicetypes.CaniModuleType)
	for _, mod := range modules {
		if mod.ModuleBayName != "" {
			populated[mod.ModuleBayName] = mod
		}
	}

	var out []TreeNode
	for _, bay := range dev.ModuleBays {
		if mod, ok := populated[bay.Name]; ok {
			out = append(out, ModuleBayNode(bay.Name, mod, inv, tf))
			delete(populated, bay.Name)
		} else {
			out = append(out, EmptyBayNode(bay.Name, tf))
		}
	}

	for _, mod := range modules {
		if mod.ModuleBayName == "" {
			out = append(out, ModuleToTreeNode(mod, inv, tf))
		}
	}
	return out
}

// deviceInterfaceNodes builds child nodes for a device's interfaces.
func deviceInterfaceNodes(dev *devicetypes.CaniDeviceType, inv *devicetypes.Inventory, tf TreeFilter) []TreeNode {
	if !tf.Interfaces || len(dev.Interfaces) == 0 {
		return nil
	}
	ifaceNodes := make([]TreeNode, 0, len(dev.Interfaces))
	for _, iface := range dev.Interfaces {
		ifaceNodes = append(ifaceNodes, InterfaceToTreeNode(iface, dev.ID, inv, tf))
	}

	if tf.Modules && len(dev.ModuleBays) > 0 {
		lbl := fmt.Sprintf("(device) Interfaces (%d)", len(dev.Interfaces))
		if !tf.NoColor {
			lbl = ColorGray + "(device interfaces)" + ColorReset + "(" + strconv.Itoa(len(dev.Interfaces)) + ")"
		}
		return []TreeNode{{
			Label:    lbl,
			Children: ifaceNodes,
		}}
	}
	return ifaceNodes
}

// deviceCableNodes builds child nodes for a device's cables.
func deviceCableNodes(dev *devicetypes.CaniDeviceType, inv *devicetypes.Inventory, tf TreeFilter) []TreeNode {
	if !tf.Cables {
		return nil
	}
	var out []TreeNode
	cables := FindCablesForDevice(dev.ID, inv)
	for _, c := range cables {
		if tf.Interfaces && CableMatchesAnyInterface(c, dev, inv) {
			continue
		}
		out = append(out, CableLeafNode(c, inv, tf))
	}
	return out
}

// ModuleToTreeNode converts a module into a tree node with its interfaces and FRUs.
func ModuleToTreeNode(mod *devicetypes.CaniModuleType, inv *devicetypes.Inventory, tf TreeFilter) TreeNode {
	return ModuleBayNode(mod.ModuleBayName, mod, inv, tf)
}

// ModuleBayNode renders a module bay. When mod is non-nil, the bay is populated.
func ModuleBayNode(bayName string, mod *devicetypes.CaniModuleType, inv *devicetypes.Inventory, tf TreeFilter) TreeNode {
	bayTag := "bay:" + bayName
	modName := mod.Name

	var label string
	if tf.NoColor {
		label = IconModule + labelModuleSep + bayTag + " | " + modName
	} else {
		label = ColorYellow + IconModule + ColorReset + " " +
			ColorGray + "(module) " + bayTag + ColorReset + " | " +
			StatusAnsi(mod.Status) + modName + ColorReset
	}

	node := TreeNode{
		Label:  label,
		Detail: string(mod.Type),
	}

	if tf.Interfaces {
		for _, iface := range mod.Interfaces {
			node.Children = append(node.Children, InterfaceToTreeNode(iface, mod.ParentDevice, inv, tf))
		}
	}

	for _, fruID := range mod.Frus {
		if inv != nil && inv.Frus != nil {
			if fru, ok := inv.Frus[fruID]; ok {
				node.Children = append(node.Children, FruToTreeNode(fru, tf))
			}
		}
	}

	return node
}

// EmptyBayNode renders an empty module bay in all-dim text.
func EmptyBayNode(bayName string, tf TreeFilter) TreeNode {
	bayTag := "bay:" + bayName
	var label string
	if tf.NoColor {
		label = IconModule + labelModuleSep + bayTag + " | empty"
	} else {
		label = ColorGray + IconModule + labelModuleSep + bayTag + " | empty" + ColorReset
	}
	return TreeNode{Label: label}
}

// InterfaceToTreeNode converts an interface spec into a tree node.
func InterfaceToTreeNode(iface devicetypes.InterfaceSpec, deviceID uuid.UUID, inv *devicetypes.Inventory, tf TreeFilter) TreeNode {
	cable := FindCableForInterface(iface, deviceID, inv)
	connected := cable != nil

	role := iface.Role
	if role == "" {
		mgmt := iface.MgmtOnly != nil && *iface.MgmtOnly
		role = devicetypes.InferInterfaceRole(iface.Name, iface.Type, mgmt)
	}

	var label, detail string
	if connected || tf.NoColor {
		label = TreeIcon(IconInterface, ColorGray, "interface", iface.Name, tf.NoColor)
		detail = string(iface.Type)
	} else {
		label = ColorGray + IconInterface + " (interface) " + iface.Name + ColorReset
		detail = ColorGray + string(iface.Type) + ColorReset
	}
	if tf.Roles && role != "" {
		detail = PipeSep(detail, "role:"+role)
	}

	node := TreeNode{Label: label, Detail: detail}
	if connected {
		node.Children = append(node.Children, CableLeafNode(cable, inv, tf))
	} else {
		node.Children = append(node.Children, DisconnectedCableNode(tf))
	}

	return node
}

// FruToTreeNode converts a FRU into a tree leaf.
func FruToTreeNode(fru *devicetypes.CaniFruType, tf TreeFilter) TreeNode {
	return TreeNode{
		Label:  TreeIcon(IconFru, ColorGray, "fru", fru.Name, tf.NoColor),
		Detail: PipeSep(string(fru.Type), fru.PartNumber),
	}
}

// CableLeafNode converts a cable into a tree leaf showing the connection.
func CableLeafNode(c *devicetypes.CaniCableType, inv *devicetypes.Inventory, tf TreeFilter) TreeNode {
	label := c.Label
	if label == "" {
		if tf.NoColor {
			label = labelNoLabel
		} else {
			label = ColorGray + labelNoLabel + ColorReset
		}
	}
	aTerm := FormatTermination(c.TerminationADevice, c.TerminationAPort, inv)
	bTerm := FormatTermination(c.TerminationBDevice, c.TerminationBPort, inv)

	fullyConnected := aTerm != "-" && bTerm != "-"
	icon := IconCableDisconnected
	iconColor := ColorRed
	if fullyConnected {
		icon = IconCable
		iconColor = ColorGreen
	}

	conn := ""
	if aTerm != "-" || bTerm != "-" {
		a := ColorInGray("A:", ColorCyan, tf.NoColor) + aTerm
		b := ColorInGray("B:", ColorGreen, tf.NoColor) + bTerm
		conn = a + " → " + b
	}
	return TreeNode{
		Label:  TreeIcon(icon, iconColor, "cable", label, tf.NoColor),
		Detail: conn,
	}
}

// DisconnectedCableNode returns a placeholder cable node for an unconnected interface.
func DisconnectedCableNode(tf TreeFilter) TreeNode {
	if tf.NoColor {
		return TreeNode{Label: IconCableDisconnected + " (cable) ✗ not connected"}
	}
	return TreeNode{
		Label: ColorRed + IconCableDisconnected + ColorReset + " " +
			ColorGray + "(cable)" + ColorReset + " " +
			ColorRed + "✗ not connected" + ColorReset,
	}
}
