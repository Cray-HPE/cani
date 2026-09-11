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
package show

import (
	"sort"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/visual"
)

// newInterfaceCommand creates the "show interface" subcommand.
func newInterfaceCommand() *cli.Command {
	return &cli.Command{
		Use:     "interface [name|uuid]",
		Aliases: []string{"interfaces"},
		Short:   "List interfaces in the inventory.",
		Long:    "List interfaces, or show a single interface by name or UUID.",
		Args:    cli.MaximumNArgs(1),
		RunE:    showInterfaces,
	}
}

func showInterfaces(cmd *cli.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		interfaces, err := findInterfacesByNameOrUUID(args[0], inv)
		if err != nil {
			return err
		}
		if len(interfaces) == 1 {
			return showSingleInterface(cmd, inv, interfaces[0])
		}
		return showInterfaceList(cmd, inv, interfaces)
	}

	interfaces := make([]*devicetypes.CaniInterface, 0, len(inv.Interfaces))
	for _, iface := range inv.Interfaces {
		interfaces = append(interfaces, iface)
	}
	return showInterfaceList(cmd, inv, interfaces)
}

func showInterfaceList(cmd *cli.Command, inv *devicetypes.Inventory, interfaces []*devicetypes.CaniInterface) error {
	sort.Slice(interfaces, func(i, j int) bool {
		if interfaces[i].Name != interfaces[j].Name {
			return interfaces[i].Name < interfaces[j].Name
		}
		leftDevice := visual.ResolveDeviceName(interfaces[i].DeviceID, inv)
		rightDevice := visual.ResolveDeviceName(interfaces[j].DeviceID, inv)
		if leftDevice != rightDevice {
			return leftDevice < rightDevice
		}
		return interfaces[i].ID.String() < interfaces[j].ID.String()
	})

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		visual.PrintInterfaceTable(interfaces, inv)
		return nil
	case "tree":
		treeFilter := treeFilterFromCmd(cmd)
		nodes := visual.BuildInterfaceTree(interfaces, inv, treeFilter)
		visual.RenderTreeOutput(nodes, treeFilter.NoColor)
		return nil
	default:
		return marshalAndPrint(interfaces)
	}
}

func showSingleInterface(cmd *cli.Command, inv *devicetypes.Inventory, iface *devicetypes.CaniInterface) error {
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		visual.PrintInterfaceTable([]*devicetypes.CaniInterface{iface}, inv)
		return nil
	case "tree":
		treeFilter := treeFilterFromCmd(cmd)
		detail := string(iface.InterfaceType)
		if iface.MacAddress != "" {
			detail = visual.PipeSep(detail, "mac:"+iface.MacAddress)
		}
		node := visual.TreeNode{Label: iface.Name, Detail: detail}
		visual.RenderTreeOutput([]visual.TreeNode{node}, treeFilter.NoColor)
		return nil
	default:
		return marshalAndPrint(iface)
	}
}
