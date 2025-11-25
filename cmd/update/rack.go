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
	"github.com/spf13/cobra"
)

// newRackUpdateCommand creates the "update rack" subcommand.
func newRackUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rack <uuid-or-name>",
		Short: "Update a rack in the inventory.",
		Long:  "Update a rack's fields by UUID or name.",
		Args:  cobra.ExactArgs(1),
		RunE:  updateRack,
	}

	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("status", "", "New status")
	cmd.Flags().String("role", "", "New role")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().Int("u-height", 0, "Rack unit height")
	cmd.Flags().String("location", "", "Parent location UUID or name")

	return cmd
}

func updateRack(cmd *cobra.Command, args []string) error {
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	id, err := resolve.Rack(inventory, args[0])
	if err != nil {
		return err
	}

	rack := inventory.Racks[id]

	if cmd.Flags().Changed("name") {
		rack.Name, _ = cmd.Flags().GetString("name")
	}
	if cmd.Flags().Changed("status") {
		rack.Status, _ = cmd.Flags().GetString("status")
	}
	if cmd.Flags().Changed("role") {
		rack.Role, _ = cmd.Flags().GetString("role")
	}
	if cmd.Flags().Changed("description") {
		rack.Description, _ = cmd.Flags().GetString("description")
	}
	if cmd.Flags().Changed("u-height") {
		rack.UHeight, _ = cmd.Flags().GetInt("u-height")
	}
	if cmd.Flags().Changed("location") {
		locRef, _ := cmd.Flags().GetString("location")
		locID, lerr := resolve.Location(inventory, locRef)
		if lerr != nil {
			return fmt.Errorf("resolving location: %w", lerr)
		}
		rack.Location = locID
	}

	if err := applySetToRack(cmd, rack); err != nil {
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

	log.Printf("Updated rack %s (%s)", id, rack.Name)
	return nil
}

func applySetToRack(cmd *cobra.Command, rack *devicetypes.CaniRackType) error {
	sets, _ := cmd.Flags().GetStringArray("set")
	pairs, err := parseSetFlags(sets)
	if err != nil {
		return err
	}
	for k, v := range pairs {
		switch k {
		case "name":
			rack.Name = v
		case "status":
			rack.Status = v
		case "role":
			rack.Role = v
		case "description":
			rack.Description = v
		case "u_height":
			n, nerr := strconv.Atoi(v)
			if nerr != nil {
				return fmt.Errorf("invalid u_height value: %s", v)
			}
			rack.UHeight = n
		case "location":
			return fmt.Errorf("use --location flag instead of --set location=...")
		default:
			return fmt.Errorf("unknown rack field: %s", k)
		}
	}
	return nil
}
