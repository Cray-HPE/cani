/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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

// newDeviceShowCommand creates the "show device" subcommand.
func newDeviceShowCommand() *cli.Command {
	return &cli.Command{
		Use:   "device [name|uuid]",
		Short: "List devices in the inventory.",
		Long:  "List devices, or show a single device by name or UUID.",
		Args:  cli.MaximumNArgs(1),
		RunE:  showDevices,
	}
}

func showDevices(cmd *cli.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		return showSingleDevice(cmd, inv, args[0])
	}

	devices := make([]*devicetypes.CaniDeviceType, 0, len(inv.Devices))
	for _, device := range inv.Devices {
		devices = append(devices, device)
	}

	sortKey, _ := cmd.Flags().GetString("sort")
	sort.Slice(devices, func(i, j int) bool {
		switch sortKey {
		case "type":
			return devices[i].Type < devices[j].Type
		case "id":
			return devices[i].ID.String() < devices[j].ID.String()
		case "status":
			return devices[i].Status < devices[j].Status
		case "vendor":
			return devices[i].GetVendor() < devices[j].GetVendor()
		case "model":
			return devices[i].Model < devices[j].Model
		default:
			return devices[i].Name < devices[j].Name
		}
	})

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		visual.PrintDeviceTable(devices, inv, visual.TreeFilter{})
		return nil
	case "tree":
		tf := treeFilterFromCmd(cmd)
		nodes := visual.BuildDeviceTree(devices, inv, tf)
		visual.RenderTreeOutput(nodes, tf.NoColor)
		return nil
	default:
		return marshalAndPrint(devices)
	}
}

func showSingleDevice(cmd *cli.Command, inv *devicetypes.Inventory, arg string) error {
	dev, err := findDeviceByNameOrUUID(arg, inv)
	if err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		visual.PrintDeviceTable([]*devicetypes.CaniDeviceType{dev}, inv, visual.TreeFilter{})
		return nil
	case "tree":
		tf := treeFilterFromCmd(cmd)
		nodes := []visual.TreeNode{visual.DeviceToTreeNode(dev, inv, tf)}
		visual.RenderTreeOutput(nodes, tf.NoColor)
		return nil
	default:
		return marshalAndPrint(dev)
	}
}

// newModuleCommand creates the "show module" subcommand.
func newModuleCommand() *cli.Command {
	return &cli.Command{
		Use:   "module [name|uuid]",
		Short: "List modules in the inventory.",
		Long:  "List modules, or show a single module by name or UUID.",
		Args:  cli.MaximumNArgs(1),
		RunE:  showModules,
	}
}

func showModules(cmd *cli.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		return showSingleModule(cmd, inv, args[0])
	}

	modules := make([]*devicetypes.CaniModuleType, 0, len(inv.Modules))
	for _, mod := range inv.Modules {
		modules = append(modules, mod)
	}
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		visual.PrintModuleTable(modules, inv)
		return nil
	case "tree":
		tf := treeFilterFromCmd(cmd)
		nodes := visual.BuildModuleTree(modules, inv, tf)
		visual.RenderTreeOutput(nodes, tf.NoColor)
		return nil
	default:
		return marshalAndPrint(modules)
	}
}

func showSingleModule(cmd *cli.Command, inv *devicetypes.Inventory, arg string) error {
	mod, err := findModuleByNameOrUUID(arg, inv)
	if err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		visual.PrintModuleTable([]*devicetypes.CaniModuleType{mod}, inv)
		return nil
	case "tree":
		tf := treeFilterFromCmd(cmd)
		nodes := []visual.TreeNode{visual.ModuleToTreeNode(mod, inv, tf)}
		visual.RenderTreeOutput(nodes, tf.NoColor)
		return nil
	default:
		return marshalAndPrint(mod)
	}
}
