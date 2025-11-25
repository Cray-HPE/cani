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

	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

// newLocationCommand creates the "update location" subcommand.
func newLocationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "location <uuid-or-name>",
		Short: "Update a location in the inventory.",
		Long:  "Update a location's fields by UUID or name.",
		Args:  cobra.ExactArgs(1),
		RunE:  updateLocation,
	}

	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("status", "", "New status")
	cmd.Flags().String("type", "", "Location type (site, building, floor, room)")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("facility", "", "Facility name")
	cmd.Flags().String("address", "", "Physical address")

	return cmd
}

func updateLocation(cmd *cobra.Command, args []string) error {
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	id, err := resolve.Location(inventory, args[0])
	if err != nil {
		return err
	}

	loc := inventory.Locations[id]

	// Apply typed flags
	if cmd.Flags().Changed("name") {
		loc.Name, _ = cmd.Flags().GetString("name")
	}
	if cmd.Flags().Changed("status") {
		loc.Status, _ = cmd.Flags().GetString("status")
	}
	if cmd.Flags().Changed("type") {
		loc.LocationType, _ = cmd.Flags().GetString("type")
	}
	if cmd.Flags().Changed("description") {
		loc.Description, _ = cmd.Flags().GetString("description")
	}
	if cmd.Flags().Changed("facility") {
		loc.Facility, _ = cmd.Flags().GetString("facility")
	}
	if cmd.Flags().Changed("address") {
		loc.PhysicalAddress, _ = cmd.Flags().GetString("address")
	}

	// Apply generic --set pairs
	if err := applySetToLocation(cmd, loc); err != nil {
		return err
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Updated location %s (%s)", id, loc.Name)
	return nil
}

func applySetToLocation(cmd *cobra.Command, loc *devicetypes.CaniLocationType) error {
	sets, _ := cmd.Flags().GetStringArray("set")
	pairs, err := parseSetFlags(sets)
	if err != nil {
		return err
	}
	for k, v := range pairs {
		switch k {
		case "name":
			loc.Name = v
		case "status":
			loc.Status = v
		case "location_type":
			loc.LocationType = v
		case "description":
			loc.Description = v
		case "facility":
			loc.Facility = v
		case "physical_address":
			loc.PhysicalAddress = v
		default:
			return fmt.Errorf("unknown location field: %s", k)
		}
	}
	return nil
}
