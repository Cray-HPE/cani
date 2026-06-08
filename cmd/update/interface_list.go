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
package update

import (
	"fmt"

	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

// interfaceRowFormat is the column layout shared by the interface listing
// header, separator, and data rows.
const interfaceRowFormat = "  %-20s %-24s %-20s %s\n"

// listDeviceInterfaces prints all interfaces for a device (and its modules).
func listDeviceInterfaces(cmd *cobra.Command, inv *devicetypes.Inventory) error {
	deviceRef, _ := cmd.Flags().GetString("device")
	if deviceRef == "" {
		return fmt.Errorf("--device is required with -L/--list")
	}

	deviceID, err := resolve.Device(inv, deviceRef)
	if err != nil {
		return fmt.Errorf("resolving --device: %w", err)
	}
	device := inv.Devices[deviceID]

	fmt.Printf("Interfaces for %s (%s):\n", device.Name, deviceID)
	fmt.Printf(interfaceRowFormat, "NAME", "TYPE", "ROLE", "SOURCE")
	fmt.Printf(interfaceRowFormat, "----", "----", "----", "------")

	// Device's own interfaces
	for _, iface := range device.Interfaces {
		printInterfaceRow(iface, "device")
	}

	// Module interfaces
	for _, mod := range inv.Modules {
		if mod == nil || mod.ParentDevice != deviceID {
			continue
		}
		for _, iface := range mod.Interfaces {
			printInterfaceRow(iface, "module:"+mod.Name)
		}
	}

	return nil
}

// printInterfaceRow prints a single interface row with its resolved role.
func printInterfaceRow(iface devicetypes.InterfaceSpec, source string) {
	mgmtOnly := iface.MgmtOnly != nil && *iface.MgmtOnly
	role := devicetypes.ResolveInterfaceRole(iface.Role, iface.Name, iface.Type, mgmtOnly)
	if role == "" {
		role = "-"
	}
	fmt.Printf(interfaceRowFormat, iface.Name, iface.Type, role, source)
}
