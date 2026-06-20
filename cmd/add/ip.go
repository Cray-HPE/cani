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

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func newIPCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "ip <cidr>",
		Short: "Add an IP address to the inventory.",
		Long: `Add an IP address to the inventory in CIDR notation (host/mask).

Examples:
  cani alpha add ip 10.0.1.1/24 --interface "switch1:vlan100" --status active
  cani alpha add ip 10.0.1.10/24 --dns-name "node1-ilo.example.com"
  cani alpha add ip 10.0.1.254/24 --interface "router1:vlan100" --role vip`,
		Args: cli.ExactArgs(1),
		RunE: addIP,
	}

	cmd.Flags().StringArray(flagInterface, nil, "Interface reference (repeatable)")
	cmd.Flags().String("type", "", "Address type: host, dhcp, or slaac")
	cmd.Flags().String("role", "", "Address role: loopback, secondary, anycast, vip, vrrp, hsrp, glbp")
	cmd.Flags().String(flagDNSName, "", "Forward DNS name for this address")
	cmd.Flags().String(flagDescription, "", "Address description")

	return cmd
}

func addIP(cmd *cli.Command, args []string) error {
	cidr := args[0]

	if err := store.Setup(cmd); err != nil {
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

	applyIPFlags(cmd, addr)

	if err := inventory.AddIPAddress(addr); err != nil {
		return fmt.Errorf("failed to add ip address: %w", err)
	}

	// Resolve --interface flags to interface UUIDs after AddIPAddress
	// (which computes parent prefix), so we can use the full inventory.
	if err := resolveIPInterfaces(cmd, inventory, addr); err != nil {
		return err
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Added IP address %s (%s)", addr.Address, addr.ID)
	return nil
}

// applyIPFlags copies optional IP-address fields from CLI flags when set.
func applyIPFlags(cmd *cli.Command, addr *devicetypes.CaniIPAddress) {
	if cmd.Flags().Changed("type") {
		t, _ := cmd.Flags().GetString("type")
		addr.Type = devicetypes.IPAddressType(t)
	}
	if cmd.Flags().Changed("role") {
		r, _ := cmd.Flags().GetString("role")
		addr.IPRole = devicetypes.IPAddressRole(r)
	}
	if cmd.Flags().Changed(flagDNSName) {
		addr.DNSName, _ = cmd.Flags().GetString(flagDNSName)
	}
	if cmd.Flags().Changed(flagDescription) {
		addr.Description, _ = cmd.Flags().GetString(flagDescription)
	}
	if cmd.Flags().Changed("status") {
		addr.Status, _ = cmd.Flags().GetString("status")
	}
	if tags, _ := cmd.Flags().GetStringArray("tag"); len(tags) > 0 {
		addr.Tags = tags
	}
}

// resolveIPInterfaces resolves each --interface reference to an interface UUID,
// auto-creating a virtual interface when the port does not already exist.
func resolveIPInterfaces(cmd *cli.Command, inventory *devicetypes.Inventory, addr *devicetypes.CaniIPAddress) error {
	if !cmd.Flags().Changed(flagInterface) {
		return nil
	}
	refs, _ := cmd.Flags().GetStringArray(flagInterface)
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
			ifaceID = createVirtualInterface(inventory, device, portName)
		}
		addr.Interfaces = append(addr.Interfaces, ifaceID)
	}
	return nil
}

// createVirtualInterface registers a virtual interface (loopback/vlan/SVI) on a
// device and the inventory, returning its new UUID.
func createVirtualInterface(inventory *devicetypes.Inventory, device *devicetypes.CaniDeviceType, portName string) uuid.UUID {
	ifaceID := uuid.New()
	device.Interfaces = append(device.Interfaces, devicetypes.InterfaceSpec{
		ID:   ifaceID,
		Name: portName,
		Type: devicetypes.InterfacesElemTypeVirtual,
	})
	inventory.Interfaces[ifaceID] = &devicetypes.CaniInterface{
		ID:            ifaceID,
		Name:          portName,
		InterfaceType: devicetypes.InterfacesElemTypeVirtual,
		DeviceID:      device.ID,
	}
	return ifaceID
}
