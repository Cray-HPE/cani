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
	"os"
	"sort"
	"text/tabwriter"

	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

// ListInterfacesCmd represents the interfaces list command
var ListInterfacesCmd = &cobra.Command{
	Use:   "interfaces",
	Short: "List interfaces in the inventory.",
	Long: `List interfaces for devices in the inventory.

Examples:
  cani show interfaces                     # Show all interfaces
  cani show interfaces --format table      # Show as table
  cani show interfaces --format json       # Show as JSON
  cani show interfaces --device "server1"  # Show interfaces for specific device
  cani show interfaces --type mgmt         # Show only management interfaces`,
	RunE: listInterfaces,
}

func init() {
	ListInterfacesCmd.Flags().String("device", "", "Filter interfaces by device name")
	ListInterfacesCmd.Flags().String("type", "", "Filter interfaces by type (mgmt, ethernet, infiniband, sfp, osfp, outlet)")
}

// InterfaceDisplay represents an interface for display
type InterfaceDisplay struct {
	DeviceName    string `json:"deviceName"`
	InterfaceName string `json:"interfaceName"`
	InterfaceType string `json:"interfaceType"`
	Label         string `json:"label,omitempty"`
	Connected     bool   `json:"connected"`
	ConnectedTo   string `json:"connectedTo,omitempty"`
}

// listInterfaces lists interfaces in the inventory
func listInterfaces(cmd *cobra.Command, args []string) error {
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inv, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load device store: %w", err)
	}

	deviceFilter, _ := cmd.Flags().GetString("device")
	typeFilter, _ := cmd.Flags().GetString("type")

	// Collect interfaces from devices
	var interfaces []InterfaceDisplay
	for _, device := range inv.Devices {
		if device == nil {
			continue
		}

		// Skip if device filter doesn't match
		if deviceFilter != "" && device.Name != deviceFilter {
			continue
		}

		// Skip cables and modules - only show interfaces for devices
		category := devicetypes.ClassifyForNautobot(string(device.Type))
		if category == devicetypes.CategoryModule {
			continue
		}

		// Generate expected interfaces for this device type
		deviceInterfaces := getExpectedInterfaces(string(device.Type), device.Name)

		for _, iface := range deviceInterfaces {
			// Apply type filter
			if typeFilter != "" && iface.InterfaceType != typeFilter {
				continue
			}

			interfaces = append(interfaces, iface)
		}
	}

	// Sort by device name, then interface name
	sort.Slice(interfaces, func(i, j int) bool {
		if interfaces[i].DeviceName != interfaces[j].DeviceName {
			return interfaces[i].DeviceName < interfaces[j].DeviceName
		}
		return interfaces[i].InterfaceName < interfaces[j].InterfaceName
	})

	// Output based on format
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(interfaces)

	case "table":
		fallthrough
	default:
		return printInterfacesTable(interfaces)
	}
}

// printInterfacesTable prints interfaces in a formatted table
func printInterfacesTable(interfaces []InterfaceDisplay) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "DEVICE\tINTERFACE\tTYPE\tLABEL\tCONNECTED\tCONNECTED TO")
	fmt.Fprintln(w, "------\t---------\t----\t-----\t---------\t------------")

	for _, iface := range interfaces {
		label := iface.Label
		if label == "" {
			label = "-"
		}
		connected := "No"
		if iface.Connected {
			connected = "Yes"
		}
		connectedTo := iface.ConnectedTo
		if connectedTo == "" {
			connectedTo = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			truncate(iface.DeviceName, 30),
			iface.InterfaceName,
			iface.InterfaceType,
			label,
			connected,
			connectedTo,
		)
	}

	fmt.Printf("\nTotal: %d interface(s)\n", len(interfaces))
	return nil
}

// getExpectedInterfaces returns expected interfaces for a device type
func getExpectedInterfaces(hardwareType, deviceName string) []InterfaceDisplay {
	var interfaces []InterfaceDisplay

	hwType := devicetypes.Type(hardwareType)

	switch hwType {
	case devicetypes.Blade:
		// ProLiant servers: iLO + 2x Ethernet + NDR adapter ports
		interfaces = append(interfaces, InterfaceDisplay{
			DeviceName:    deviceName,
			InterfaceName: "iLO",
			InterfaceType: "mgmt",
			Label:         "iLO Management",
		})
		for i := 1; i <= 2; i++ {
			interfaces = append(interfaces, InterfaceDisplay{
				DeviceName:    deviceName,
				InterfaceName: fmt.Sprintf("NIC.Slot.%d-1", i),
				InterfaceType: "ethernet",
				Label:         fmt.Sprintf("Ethernet Port %d", i),
			})
		}
		// NDR ports
		for i := 1; i <= 2; i++ {
			interfaces = append(interfaces, InterfaceDisplay{
				DeviceName:    deviceName,
				InterfaceName: fmt.Sprintf("NDR.Slot.1-%d", i),
				InterfaceType: "osfp",
				Label:         fmt.Sprintf("NDR Port %d", i),
			})
		}

	case devicetypes.MgmtSwitch:
		// Aruba 2930F: 48x 1G Ethernet + 4x SFP+
		for i := 1; i <= 48; i++ {
			interfaces = append(interfaces, InterfaceDisplay{
				DeviceName:    deviceName,
				InterfaceName: fmt.Sprintf("1/%d", i),
				InterfaceType: "ethernet",
				Label:         fmt.Sprintf("Port %d", i),
			})
		}
		for i := 49; i <= 52; i++ {
			interfaces = append(interfaces, InterfaceDisplay{
				DeviceName:    deviceName,
				InterfaceName: fmt.Sprintf("1/%d", i),
				InterfaceType: "sfp",
				Label:         fmt.Sprintf("SFP+ Port %d", i-48),
			})
		}

	case devicetypes.HSNSwitch:
		// NDR Switch: 64x OSFP ports
		for i := 1; i <= 64; i++ {
			interfaces = append(interfaces, InterfaceDisplay{
				DeviceName:    deviceName,
				InterfaceName: fmt.Sprintf("Port%d", i),
				InterfaceType: "osfp",
				Label:         fmt.Sprintf("OSFP Port %d", i),
			})
		}

	case devicetypes.CabinetPDU:
		// PDU: 24 outlets (typical)
		for i := 1; i <= 24; i++ {
			interfaces = append(interfaces, InterfaceDisplay{
				DeviceName:    deviceName,
				InterfaceName: fmt.Sprintf("Outlet-%d", i),
				InterfaceType: "outlet",
				Label:         fmt.Sprintf("Outlet %d", i),
			})
		}

	case devicetypes.Rack:
		// Racks don't have interfaces
		return nil

	default:
		// Generic device: at least one management interface
		interfaces = append(interfaces, InterfaceDisplay{
			DeviceName:    deviceName,
			InterfaceName: "mgmt0",
			InterfaceType: "mgmt",
			Label:         "Management",
		})
	}

	return interfaces
}
