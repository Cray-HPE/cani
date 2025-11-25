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
	"encoding/json"
	"fmt"
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

// newLocationCommand creates the "show location" subcommand.
func newLocationCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "location",
		Short: "List locations in the inventory.",
		Long:  "List locations in the inventory.",
		RunE:  showLocations,
	}
}

func showLocations(cmd *cobra.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	locations := make([]*devicetypes.CaniLocationType, 0, len(inv.Locations))
	for _, loc := range inv.Locations {
		locations = append(locations, loc)
	}
	sort.Slice(locations, func(i, j int) bool {
		return locations[i].Name < locations[j].Name
	})
	return marshalAndPrint(locations)
}

// newRackShowCommand creates the "show rack" subcommand.
func newRackShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rack",
		Short: "List racks in the inventory.",
		Long:  "List racks in the inventory.",
		RunE:  showRacks,
	}
}

func showRacks(cmd *cobra.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	racks := make([]*devicetypes.CaniRackType, 0, len(inv.Racks))
	for _, rack := range inv.Racks {
		racks = append(racks, rack)
	}
	sort.Slice(racks, func(i, j int) bool {
		return racks[i].Name < racks[j].Name
	})
	return marshalAndPrint(racks)
}

// newDeviceShowCommand creates the "show device" subcommand.
func newDeviceShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "device",
		Short: "List devices in the inventory.",
		Long:  "List devices in the inventory.",
		RunE:  showDevices,
	}
}

func showDevices(cmd *cobra.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
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
	return marshalAndPrint(devices)
}

// newModuleCommand creates the "show module" subcommand.
func newModuleCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "module",
		Short: "List modules in the inventory.",
		Long:  "List modules in the inventory.",
		RunE:  showModules,
	}
}

func showModules(cmd *cobra.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	modules := make([]*devicetypes.CaniModuleType, 0, len(inv.Modules))
	for _, mod := range inv.Modules {
		modules = append(modules, mod)
	}
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})
	return marshalAndPrint(modules)
}

// newCableCommand creates the "show cable" subcommand (replaces show cables).
func newCableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "cable",
		Short: "List cables in the inventory.",
		Long:  "List cables in the inventory.",
		RunE:  showCables,
	}
}

func showCables(cmd *cobra.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	cables := make([]*devicetypes.CaniCableType, 0, len(inv.Cables))
	for _, cable := range inv.Cables {
		cables = append(cables, cable)
	}
	sort.Slice(cables, func(i, j int) bool {
		return cables[i].Label < cables[j].Label
	})
	return marshalAndPrint(cables)
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
