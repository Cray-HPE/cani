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
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// PrintFruTable renders FRUs as a fixed-width table.
func PrintFruTable(frus []*devicetypes.CaniFruType, inv *devicetypes.Inventory) {
	header := Col("NAME", ColName) + "  " +
		Col("MODEL", ColModel) + "  " +
		Col("STATUS", ColStatus) + "  " +
		Col("DEVICE", ColDevice) + "  " +
		Col("SERIAL", ColSerial)
	sep := Col(strings.Repeat("-", ColName), ColName) + "  " +
		Col(strings.Repeat("-", ColModel), ColModel) + "  " +
		Col(strings.Repeat("-", ColStatus), ColStatus) + "  " +
		Col(strings.Repeat("-", ColDevice), ColDevice) + "  " +
		Col(strings.Repeat("-", ColSerial), ColSerial)

	fmt.Println(header)
	fmt.Println(sep)
	for _, f := range frus {
		devName := ResolveDeviceName(f.Device, inv)
		fmt.Println(
			Col(f.Name, ColName) + "  " +
				Col(f.Model, ColModel) + "  " +
				Col(f.Status, ColStatus) + "  " +
				Col(devName, ColDevice) + "  " +
				Col(f.Serial, ColSerial),
		)
	}
	fmt.Printf("\nTotal: %d FRU(s)\n", len(frus))
}

// PrintInterfaceTable renders interface instances as a fixed-width table.
func PrintInterfaceTable(ifaces []*devicetypes.CaniInterface, inv *devicetypes.Inventory) {
	header := Col("NAME", ColName) + "  " +
		Col("TYPE", ColType) + "  " +
		Col("ROLE", ColRole) + "  " +
		Col("DEVICE", ColDevice) + "  " +
		Col("LABEL", ColLabel) + "  " +
		Col("MAC", ColMac) + "  " +
		Col("CONNECTED", ColConnected)
	sep := Col(strings.Repeat("-", ColName), ColName) + "  " +
		Col(strings.Repeat("-", ColType), ColType) + "  " +
		Col(strings.Repeat("-", ColRole), ColRole) + "  " +
		Col(strings.Repeat("-", ColDevice), ColDevice) + "  " +
		Col(strings.Repeat("-", ColLabel), ColLabel) + "  " +
		Col(strings.Repeat("-", ColMac), ColMac) + "  " +
		Col(strings.Repeat("-", ColConnected), ColConnected)

	fmt.Println(header)
	fmt.Println(sep)
	for _, iface := range ifaces {
		devName := ResolveDeviceName(iface.DeviceID, inv)
		connected := "-"
		if iface.ConnectedCable != nil {
			connected = "yes"
		}
		mac := iface.MacAddress
		if mac == "" {
			mac = "-"
		}
		fmt.Println(
			Col(iface.Name, ColName) + "  " +
				Col(string(iface.InterfaceType), ColType) + "  " +
				Col(iface.Role, ColRole) + "  " +
				Col(devName, ColDevice) + "  " +
				Col(iface.Label, ColLabel) + "  " +
				Col(mac, ColMac) + "  " +
				Col(connected, ColConnected),
		)
	}
	fmt.Printf("\nTotal: %d interface(s)\n", len(ifaces))
}

// CollectInterfacesForDevices returns all CaniInterface instances belonging to the given devices.
func CollectInterfacesForDevices(devices []*devicetypes.CaniDeviceType, inv *devicetypes.Inventory) []*devicetypes.CaniInterface {
	devIDs := make(map[string]bool, len(devices))
	for _, d := range devices {
		devIDs[d.ID.String()] = true
	}
	var result []*devicetypes.CaniInterface
	for _, iface := range inv.Interfaces {
		if devIDs[iface.DeviceID.String()] {
			result = append(result, iface)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// PrintAllTables renders all inventory sections as tables with section headers.
func PrintAllTables(inv *devicetypes.Inventory) {
	if len(inv.Locations) > 0 {
		locations := sortedLocations(inv)
		fmt.Printf("\nLocations (%d):\n", len(locations))
		PrintLocationTable(locations, inv)
	}

	if len(inv.Racks) > 0 {
		racks := sortedRacks(inv)
		fmt.Printf("\nRacks (%d):\n", len(racks))
		PrintRackTable(racks, inv)
	}

	if len(inv.Devices) > 0 {
		devices := sortedDevices(inv)
		fmt.Printf("\nDevices (%d):\n", len(devices))
		PrintDeviceTable(devices, inv, TreeFilter{})
	}

	if len(inv.Modules) > 0 {
		modules := sortedModules(inv)
		fmt.Printf("\nModules (%d):\n", len(modules))
		PrintModuleTable(modules, inv)
	}

	if len(inv.Cables) > 0 {
		cables := sortedCables(inv)
		fmt.Printf("\nCables (%d):\n", len(cables))
		PrintCableTable(cables, inv)
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
