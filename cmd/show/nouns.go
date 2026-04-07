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

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/spf13/cobra"
)

// newLocationCommand creates the "show location" subcommand.
func newLocationCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "location [name|uuid]",
		Short: "List locations in the inventory.",
		Long:  "List locations, or show a single location by name or UUID.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  showLocations,
	}
}

func showLocations(cmd *cobra.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		return showSingleLocation(cmd, inv, args[0])
	}

	locations := make([]*devicetypes.CaniLocationType, 0, len(inv.Locations))
	for _, loc := range inv.Locations {
		locations = append(locations, loc)
	}
	sort.Slice(locations, func(i, j int) bool {
		return locations[i].Name < locations[j].Name
	})

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		printLocationTable(locations, inv)
		return nil
	case "tree":
		nodes := buildLocationTree(inv)
		renderTreeOutput(nodes)
		return nil
	default:
		return marshalAndPrint(locations)
	}
}

func showSingleLocation(cmd *cobra.Command, inv *devicetypes.Inventory, arg string) error {
	loc, err := findLocationByNameOrUUID(arg, inv)
	if err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		printLocationTable([]*devicetypes.CaniLocationType{loc}, inv)
		return nil
	case "tree":
		nodes := []visual.TreeNode{locationToTreeNode(loc, inv)}
		renderTreeOutput(nodes)
		return nil
	default:
		return marshalAndPrint(loc)
	}
}

// newRackShowCommand creates the "show rack" subcommand.
func newRackShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rack [name|uuid]",
		Short: "List racks in the inventory.",
		Long:  "List racks, or show a single rack by name or UUID.\nWhen a rack is specified, the default output is a visual rack view.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  showRacks,
	}
}

func showRacks(cmd *cobra.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		return showSingleRack(cmd, inv, args[0])
	}

	visualMode, _ := cmd.Flags().GetBool("visual")
	if visualMode {
		noColor, _ := cmd.Flags().GetBool("no-color")
		opts := visual.CompactRenderOptions{
			NoColor:   noColor,
			Detail:    true,
			Inventory: inv,
		}
		return visual.RenderMinimapDetailAll(inv, opts)
	}

	racks := make([]*devicetypes.CaniRackType, 0, len(inv.Racks))
	for _, rack := range inv.Racks {
		racks = append(racks, rack)
	}
	sort.Slice(racks, func(i, j int) bool {
		return racks[i].Name < racks[j].Name
	})

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		printRackTable(racks, inv)
		return nil
	case "tree":
		nodes := buildRackTree(racks, inv)
		renderTreeOutput(nodes)
		return nil
	default:
		return marshalAndPrint(racks)
	}
}

// newDeviceShowCommand creates the "show device" subcommand.
func newDeviceShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "device [name|uuid]",
		Short: "List devices in the inventory.",
		Long:  "List devices, or show a single device by name or UUID.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  showDevices,
	}
}

func showDevices(cmd *cobra.Command, args []string) error {
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
		printDeviceTable(devices, inv)
		return nil
	case "tree":
		nodes := buildDeviceTree(devices, inv)
		renderTreeOutput(nodes)
		return nil
	default:
		return marshalAndPrint(devices)
	}
}

func showSingleDevice(cmd *cobra.Command, inv *devicetypes.Inventory, arg string) error {
	dev, err := findDeviceByNameOrUUID(arg, inv)
	if err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		printDeviceTable([]*devicetypes.CaniDeviceType{dev}, inv)
		return nil
	case "tree":
		nodes := []visual.TreeNode{deviceToTreeNode(dev, inv)}
		renderTreeOutput(nodes)
		return nil
	default:
		return marshalAndPrint(dev)
	}
}

// newModuleCommand creates the "show module" subcommand.
func newModuleCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "module [name|uuid]",
		Short: "List modules in the inventory.",
		Long:  "List modules, or show a single module by name or UUID.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  showModules,
	}
}

func showModules(cmd *cobra.Command, args []string) error {
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
		printModuleTable(modules, inv)
		return nil
	case "tree":
		nodes := buildModuleTree(modules, inv)
		renderTreeOutput(nodes)
		return nil
	default:
		return marshalAndPrint(modules)
	}
}

func showSingleModule(cmd *cobra.Command, inv *devicetypes.Inventory, arg string) error {
	mod, err := findModuleByNameOrUUID(arg, inv)
	if err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		printModuleTable([]*devicetypes.CaniModuleType{mod}, inv)
		return nil
	case "tree":
		nodes := []visual.TreeNode{moduleToTreeNode(mod, inv)}
		renderTreeOutput(nodes)
		return nil
	default:
		return marshalAndPrint(mod)
	}
}
