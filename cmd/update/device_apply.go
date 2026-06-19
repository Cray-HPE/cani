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
	"strconv"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// applyScalarFields applies simple string field updates from changed flags.
func applyScalarFields(cmd *cli.Command, device *devicetypes.CaniDeviceType) {
	if cmd.Flags().Changed("name") {
		device.Name, _ = cmd.Flags().GetString("name")
	}
	if cmd.Flags().Changed("status") {
		device.Status, _ = cmd.Flags().GetString("status")
	}
	if cmd.Flags().Changed("role") {
		device.Role, _ = cmd.Flags().GetString("role")
	}
	if cmd.Flags().Changed(flagDescription) {
		device.Description, _ = cmd.Flags().GetString(flagDescription)
	}
}

// applyPositionUpdate handles --position/--face changes, including swap logic.
func applyPositionUpdate(cmd *cli.Command, inventory *devicetypes.Inventory, id uuid.UUID, device *devicetypes.CaniDeviceType) error {
	if !cmd.Flags().Changed(flagPosition) && !cmd.Flags().Changed("face") {
		return nil
	}
	newPos := device.RackPosition
	newFace := device.Face
	if cmd.Flags().Changed(flagPosition) {
		newPos, _ = cmd.Flags().GetInt(flagPosition)
	}
	if cmd.Flags().Changed("face") {
		newFace, _ = cmd.Flags().GetString("face")
	}
	// device.Rack is a derived field; look up the rack via Parent.
	rack := findDeviceRack(inventory, device)
	if rack == nil {
		return fmt.Errorf("device %s is not assigned to a rack", id)
	}
	doSwap, _ := cmd.Flags().GetBool("swap")
	return moveOrSwap(rack, inventory, id, device, newPos, newFace, doSwap)
}

// applyParentUpdate handles a --parent change.
func applyParentUpdate(cmd *cli.Command, inventory *devicetypes.Inventory, device *devicetypes.CaniDeviceType) error {
	if !cmd.Flags().Changed("parent") {
		return nil
	}
	parentRef, _ := cmd.Flags().GetString("parent")
	parentID, perr := resolveParent(inventory, parentRef)
	if perr != nil {
		return fmt.Errorf("resolving parent: %w", perr)
	}
	device.Parent = parentID
	return nil
}

// applyTagsAndMetadata handles --tag and --metadata changes.
func applyTagsAndMetadata(cmd *cli.Command, device *devicetypes.CaniDeviceType) error {
	if cmd.Flags().Changed("tag") {
		tags, _ := cmd.Flags().GetStringArray("tag")
		device.Tags = tags
	}
	if cmd.Flags().Changed("metadata") {
		pairs, _ := cmd.Flags().GetStringArray("metadata")
		if err := applyProviderMetadata(&device.ProviderMetadata, pairs); err != nil {
			return err
		}
	}
	return nil
}

// applyPrimaryIPs handles --primary-ipv4 and --primary-ipv6 changes.
func applyPrimaryIPs(cmd *cli.Command, inventory *devicetypes.Inventory, device *devicetypes.CaniDeviceType) error {
	if cmd.Flags().Changed("primary-ipv4") {
		ref, _ := cmd.Flags().GetString("primary-ipv4")
		ipID, rerr := resolveIPAddress(inventory, ref)
		if rerr != nil {
			return fmt.Errorf("resolving --primary-ipv4: %w", rerr)
		}
		device.PrimaryIPv4 = ipID
	}
	if cmd.Flags().Changed("primary-ipv6") {
		ref, _ := cmd.Flags().GetString("primary-ipv6")
		ipID, rerr := resolveIPAddress(inventory, ref)
		if rerr != nil {
			return fmt.Errorf("resolving --primary-ipv6: %w", rerr)
		}
		device.PrimaryIPv6 = ipID
	}
	return nil
}

func applySetToDevice(cmd *cli.Command, device *devicetypes.CaniDeviceType) error {
	sets, _ := cmd.Flags().GetStringArray("set")
	pairs, err := parseSetFlags(sets)
	if err != nil {
		return err
	}
	for k, v := range pairs {
		switch k {
		case "name":
			device.Name = v
		case "status":
			device.Status = v
		case "role":
			device.Role = v
		case flagDescription:
			device.Description = v
		case "rack_position":
			n, nerr := strconv.Atoi(v)
			if nerr != nil {
				return fmt.Errorf("invalid rack_position value: %s", v)
			}
			device.RackPosition = n
		case "face":
			device.Face = v
		case "serial":
			device.Serial = v
		case "asset_tag":
			device.AssetTag = v
		case "tag":
			device.Tags = append(device.Tags, v)
		case "parent":
			return fmt.Errorf("use --parent flag instead of --set parent=...")
		default:
			return fmt.Errorf("unknown device field: %s", k)
		}
	}
	return nil
}
