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
	"encoding/json"
	"fmt"
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/spf13/cobra"
)

func showSingleRack(cmd *cobra.Command, inv *devicetypes.Inventory, arg string) error {
	rack, err := findRackByNameOrUUID(arg, inv)
	if err != nil {
		return err
	}

	// Default to rack-view visualization unless --format was explicitly set
	formatChanged := cmd.Flags().Changed("format")

	if !formatChanged {
		noColor, _ := cmd.Flags().GetBool("no-color")
		opts := visual.CompactRenderOptions{
			NoColor:    noColor,
			RackFilter: rack.Name,
			Detail:     true,
			Inventory:  inv,
		}
		return visual.RenderMinimapDetailAll(inv, opts)
	}

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		printRackTable([]*devicetypes.CaniRackType{rack}, inv)
		return nil
	case "tree":
		nodes := buildRackTree([]*devicetypes.CaniRackType{rack}, inv)
		renderTreeOutput(nodes)
		return nil
	default:
		return marshalAndPrint(rack)
	}
}

// newCableCommand creates the "show cable" subcommand.
func newCableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "cable [label|uuid]",
		Short: "List cables in the inventory.",
		Long:  "List cables, or show a single cable by label or UUID.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  showCables,
	}
}

func showCables(cmd *cobra.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		return showSingleCable(cmd, inv, args[0])
	}

	cables := make([]*devicetypes.CaniCableType, 0, len(inv.Cables))
	for _, cable := range inv.Cables {
		cables = append(cables, cable)
	}
	sort.Slice(cables, func(i, j int) bool {
		return cables[i].Label < cables[j].Label
	})

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		printCableTableFromTypes(cables, inv)
		return nil
	case "tree":
		nodes := buildCableTree(cables, inv)
		renderTreeOutput(nodes)
		return nil
	default:
		return marshalAndPrint(cables)
	}
}

func showSingleCable(cmd *cobra.Command, inv *devicetypes.Inventory, arg string) error {
	cable, err := findCableByLabelOrUUID(arg, inv)
	if err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		printCableTableFromTypes([]*devicetypes.CaniCableType{cable}, inv)
		return nil
	case "tree":
		nodes := buildCableTree([]*devicetypes.CaniCableType{cable}, inv)
		renderTreeOutput(nodes)
		return nil
	default:
		return marshalAndPrint(cable)
	}
}

// newFruCommand creates the "show fru" subcommand.
func newFruCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "fru [name|uuid]",
		Short: "List FRUs in the inventory.",
		Long:  "List field-replaceable units, or show a single FRU by name or UUID.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  showFrus,
	}
}

func showFrus(cmd *cobra.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		return showSingleFru(cmd, inv, args[0])
	}

	frus := make([]*devicetypes.CaniFruType, 0, len(inv.Frus))
	for _, f := range inv.Frus {
		frus = append(frus, f)
	}
	sort.Slice(frus, func(i, j int) bool {
		return frus[i].Name < frus[j].Name
	})

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		printFruTable(frus, inv)
		return nil
	case "tree":
		nodes := buildFruTree(frus, inv)
		renderTreeOutput(nodes)
		return nil
	default:
		return marshalAndPrint(frus)
	}
}

func showSingleFru(cmd *cobra.Command, inv *devicetypes.Inventory, arg string) error {
	fru, err := findFruByNameOrUUID(arg, inv)
	if err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		printFruTable([]*devicetypes.CaniFruType{fru}, inv)
		return nil
	case "tree":
		nodes := []visual.TreeNode{fruToTreeNode(fru)}
		renderTreeOutput(nodes)
		return nil
	default:
		return marshalAndPrint(fru)
	}
}

// newInterfaceCommand creates the "show interface" subcommand.
func newInterfaceCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "interface [name|uuid]",
		Short: "List interfaces in the inventory.",
		Long:  "List interfaces, or show a single interface by name or UUID.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  showInterfaceInstances,
	}
}

func showInterfaceInstances(cmd *cobra.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		return showSingleInterface(cmd, inv, args[0])
	}

	ifaces := make([]*devicetypes.InterfaceInstance, 0, len(inv.Interfaces))
	for _, iface := range inv.Interfaces {
		ifaces = append(ifaces, iface)
	}
	sort.Slice(ifaces, func(i, j int) bool {
		return ifaces[i].Name < ifaces[j].Name
	})

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		printInterfaceInstanceTable(ifaces, inv)
		return nil
	case "tree":
		nodes := buildInterfaceInstanceTree(ifaces, inv)
		renderTreeOutput(nodes)
		return nil
	default:
		return marshalAndPrint(ifaces)
	}
}

func showSingleInterface(cmd *cobra.Command, inv *devicetypes.Inventory, arg string) error {
	iface, err := findInterfaceByNameOrUUID(arg, inv)
	if err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		printInterfaceInstanceTable([]*devicetypes.InterfaceInstance{iface}, inv)
		return nil
	case "tree":
		node := visual.TreeNode{
			Label:  iface.Name,
			Detail: string(iface.InterfaceType),
		}
		renderTreeOutput([]visual.TreeNode{node})
		return nil
	default:
		return marshalAndPrint(iface)
	}
}

// marshalAndPrint encodes v as indented JSON and writes it to stdout.
func marshalAndPrint(v any) error {
	output, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}
	fmt.Println(string(output))
	return nil
}
