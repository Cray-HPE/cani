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
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
)

// Column widths for FRU and interface tables.
const (
	colSerial    = 20
	colConnected = 12
)

// printFruTable renders FRUs as a fixed-width table.
func printFruTable(frus []*devicetypes.CaniFruType, inv *devicetypes.Inventory) {
	header := col("NAME", colName) + "  " +
		col("MODEL", colModel) + "  " +
		col("STATUS", colStatus) + "  " +
		col("DEVICE", colDevice) + "  " +
		col("SERIAL", colSerial)
	sep := col(strings.Repeat("-", colName), colName) + "  " +
		col(strings.Repeat("-", colModel), colModel) + "  " +
		col(strings.Repeat("-", colStatus), colStatus) + "  " +
		col(strings.Repeat("-", colDevice), colDevice) + "  " +
		col(strings.Repeat("-", colSerial), colSerial)

	fmt.Println(header)
	fmt.Println(sep)
	for _, f := range frus {
		devName := resolveDeviceName(f.Device, inv)
		fmt.Println(
			col(f.Name, colName) + "  " +
				col(f.Model, colModel) + "  " +
				col(f.Status, colStatus) + "  " +
				col(devName, colDevice) + "  " +
				col(f.Serial, colSerial),
		)
	}
	fmt.Printf("\nTotal: %d FRU(s)\n", len(frus))
}

// printInterfaceInstanceTable renders interface instances as a fixed-width table.
func printInterfaceInstanceTable(ifaces []*devicetypes.InterfaceInstance, inv *devicetypes.Inventory) {
	header := col("NAME", colName) + "  " +
		col("TYPE", colType) + "  " +
		col("DEVICE", colDevice) + "  " +
		col("LABEL", colLabel) + "  " +
		col("CONNECTED", colConnected)
	sep := col(strings.Repeat("-", colName), colName) + "  " +
		col(strings.Repeat("-", colType), colType) + "  " +
		col(strings.Repeat("-", colDevice), colDevice) + "  " +
		col(strings.Repeat("-", colLabel), colLabel) + "  " +
		col(strings.Repeat("-", colConnected), colConnected)

	fmt.Println(header)
	fmt.Println(sep)
	for _, iface := range ifaces {
		devName := resolveDeviceName(iface.DeviceID, inv)
		connected := "-"
		if iface.ConnectedCable != nil {
			connected = "yes"
		}
		fmt.Println(
			col(iface.Name, colName) + "  " +
				col(string(iface.InterfaceType), colType) + "  " +
				col(devName, colDevice) + "  " +
				col(iface.Label, colLabel) + "  " +
				col(connected, colConnected),
		)
	}
	fmt.Printf("\nTotal: %d interface(s)\n", len(ifaces))
}

// buildFruTree groups FRUs by parent device as tree nodes.
func buildFruTree(frus []*devicetypes.CaniFruType, inv *devicetypes.Inventory) []visual.TreeNode {
	groups := make(map[uuid.UUID][]*devicetypes.CaniFruType)
	var order []uuid.UUID
	for _, f := range frus {
		did := f.Device
		if _, seen := groups[did]; !seen {
			order = append(order, did)
		}
		groups[did] = append(groups[did], f)
	}

	nodes := make([]visual.TreeNode, 0, len(order))
	for _, did := range order {
		parentLabel := resolveDeviceName(did, inv)
		if parentLabel == "" {
			parentLabel = "(unassigned)"
		}
		node := visual.TreeNode{Label: parentLabel, Detail: "device"}
		for _, f := range groups[did] {
			node.Children = append(node.Children, fruToTreeNode(f))
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// buildInterfaceInstanceTree groups interface instances by device as tree nodes.
func buildInterfaceInstanceTree(ifaces []*devicetypes.InterfaceInstance, inv *devicetypes.Inventory) []visual.TreeNode {
	groups := make(map[uuid.UUID][]*devicetypes.InterfaceInstance)
	var order []uuid.UUID
	for _, iface := range ifaces {
		did := iface.DeviceID
		if _, seen := groups[did]; !seen {
			order = append(order, did)
		}
		groups[did] = append(groups[did], iface)
	}

	nodes := make([]visual.TreeNode, 0, len(order))
	for _, did := range order {
		parentLabel := resolveDeviceName(did, inv)
		if parentLabel == "" {
			parentLabel = "(unassigned)"
		}
		node := visual.TreeNode{Label: parentLabel, Detail: "device"}
		for _, iface := range groups[did] {
			detail := string(iface.InterfaceType)
			if iface.Label != "" {
				detail += " " + iface.Label
			}
			node.Children = append(node.Children, visual.TreeNode{
				Label:  iface.Name,
				Detail: detail,
			})
		}
		nodes = append(nodes, node)
	}
	return nodes
}
