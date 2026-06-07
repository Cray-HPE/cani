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
package add

import (
	"fmt"
	"log"
	"strings"

	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func newIPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ip <cidr>",
		Short: "Add an IP address to the inventory.",
		Long: `Add an IP address to the inventory in CIDR notation (host/mask).

Examples:
  cani alpha add ip 10.0.1.1/24 --interface "switch1:vlan100" --status active
  cani alpha add ip 10.0.1.10/24 --dns-name "node1-ilo.example.com"
  cani alpha add ip 10.0.1.254/24 --interface "router1:vlan100" --role vip`,
		Args: cobra.ExactArgs(1),
		RunE: addIP,
	}

	cmd.Flags().StringArray("interface", nil, "Interface reference (repeatable)")
	cmd.Flags().String("type", "", "Address type: host, dhcp, or slaac")
	cmd.Flags().String("role", "", "Address role: loopback, secondary, anycast, vip, vrrp, hsrp, glbp")
	cmd.Flags().String("dns-name", "", "Forward DNS name for this address")
	cmd.Flags().String("description", "", "Address description")

	return cmd
}

func addIP(cmd *cobra.Command, args []string) error {
	cidr := args[0]

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	addr := &devicetypes.CaniIPAddress{
		ID:      uuid.New(),
		Address: cidr,
	}

	if cmd.Flags().Changed("type") {
		t, _ := cmd.Flags().GetString("type")
		addr.Type = devicetypes.IPAddressType(t)
	}
	if cmd.Flags().Changed("role") {
		r, _ := cmd.Flags().GetString("role")
		addr.IPRole = devicetypes.IPAddressRole(r)
	}
	if cmd.Flags().Changed("dns-name") {
		addr.DNSName, _ = cmd.Flags().GetString("dns-name")
	}
	if cmd.Flags().Changed("description") {
		addr.Description, _ = cmd.Flags().GetString("description")
	}
	if cmd.Flags().Changed("status") {
		addr.Status, _ = cmd.Flags().GetString("status")
	}
	tags, _ := cmd.Flags().GetStringArray("tag")
	if len(tags) > 0 {
		addr.Tags = tags
	}

	if err := inventory.AddIPAddress(addr); err != nil {
		return fmt.Errorf("failed to add ip address: %w", err)
	}

	// Resolve --interface flags to interface UUIDs after AddIPAddress
	// (which computes parent prefix), so we can use the full inventory.
	if cmd.Flags().Changed("interface") {
		refs, _ := cmd.Flags().GetStringArray("interface")
		for _, ref := range refs {
			parts := strings.SplitN(ref, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid interface reference %q: expected device:port", ref)
			}
			deviceName, portName := parts[0], parts[1]
			device := inventory.FindDeviceByNameOrID(deviceName)
			if device == nil {
				return fmt.Errorf("interface reference %q: device %q not found", ref, deviceName)
			}
			ifaceID := inventory.FindInterfaceIDByPort(device.ID, portName)
			if ifaceID == uuid.Nil {
				// Auto-create virtual interface for loopback/vlan/SVI names
				ifaceID = uuid.New()
				iface := devicetypes.InterfaceSpec{
					ID:   ifaceID,
					Name: portName,
					Type: devicetypes.InterfacesElemTypeVirtual,
				}
				device.Interfaces = append(device.Interfaces, iface)
				inventory.Interfaces[ifaceID] = &devicetypes.CaniInterface{
					ID:            ifaceID,
					Name:          portName,
					InterfaceType: devicetypes.InterfacesElemTypeVirtual,
					DeviceID:      device.ID,
				}
			}
			addr.Interfaces = append(addr.Interfaces, ifaceID)
		}
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Added IP address %s (%s)", addr.Address, addr.ID)
	return nil
}
