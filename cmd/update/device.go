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
	"log"
	"strconv"

	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// newDeviceCommand creates the "update device" subcommand.
func newDeviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "device <uuid-or-name>",
		Short: "Update a device in the inventory.",
		Long:  "Update a device's fields by UUID or name.",
		Args:  cobra.ExactArgs(1),
		RunE:  updateDevice,
	}

	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("status", "", "New status")
	cmd.Flags().String("role", "", "New role")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().Int("position", 0, "Rack U position")
	cmd.Flags().String("face", "", "Rack face (front, rear)")
	cmd.Flags().String("parent", "", "Parent UUID or name (rack or device)")

	return cmd
}

func updateDevice(cmd *cobra.Command, args []string) error {
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	id, err := resolve.Device(inventory, args[0])
	if err != nil {
		return err
	}

	device := inventory.Devices[id]

	if cmd.Flags().Changed("name") {
		device.Name, _ = cmd.Flags().GetString("name")
	}
	if cmd.Flags().Changed("status") {
		device.Status, _ = cmd.Flags().GetString("status")
	}
	if cmd.Flags().Changed("role") {
		device.Role, _ = cmd.Flags().GetString("role")
	}
	if cmd.Flags().Changed("description") {
		device.Description, _ = cmd.Flags().GetString("description")
	}
	if cmd.Flags().Changed("position") {
		device.RackPosition, _ = cmd.Flags().GetInt("position")
	}
	if cmd.Flags().Changed("face") {
		device.Face, _ = cmd.Flags().GetString("face")
	}
	if cmd.Flags().Changed("parent") {
		parentRef, _ := cmd.Flags().GetString("parent")
		parentID, perr := resolveParent(inventory, parentRef)
		if perr != nil {
			return fmt.Errorf("resolving parent: %w", perr)
		}
		device.Parent = parentID
	}

	if err := applySetToDevice(cmd, device); err != nil {
		return err
	}

	// Rebuild relationships so derived fields are updated.
	result := inventory.VerifyParentChildRelationships()
	if result.HasErrors() {
		return fmt.Errorf("relationship errors: %v", result.Errors)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Updated device %s (%s)", id, device.Name)
	return nil
}

// resolveParent tries to resolve a string as a rack UUID/name first,
// then as a device UUID/name. Returns the resolved UUID.
func resolveParent(inv *devicetypes.Inventory, ref string) (uuid.UUID, error) {
	if id, err := resolve.Rack(inv, ref); err == nil {
		return id, nil
	}
	if id, err := resolve.Device(inv, ref); err == nil {
		return id, nil
	}
	return uuid.Nil, fmt.Errorf("%q not found as rack or device", ref)
}

func applySetToDevice(cmd *cobra.Command, device *devicetypes.CaniDeviceType) error {
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
		case "description":
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
		case "parent":
			return fmt.Errorf("use --parent flag instead of --set parent=...")
		default:
			return fmt.Errorf("unknown device field: %s", k)
		}
	}
	return nil
}
