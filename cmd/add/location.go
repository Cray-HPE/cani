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

	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// newLocationCommand creates the "add location" subcommand.
func newLocationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "location [name]",
		Short: "Add a location to the inventory.",
		Long:  "Add a location (site, building, floor, room) to the inventory.",
		Args:  cobra.ExactArgs(1),
		RunE:  addLocation,
	}

	cmd.Flags().String("type", "site", "Location type (site, building, floor, room)")
	cmd.Flags().String("parent", "", "Parent location UUID or name")

	return cmd
}

func addLocation(cmd *cobra.Command, args []string) error {
	name := args[0]
	locType, _ := cmd.Flags().GetString("type")

	loc := &devicetypes.CaniLocationType{
		ID:           uuid.New(),
		Name:         name,
		LocationType: locType,
		Status:       "active",
	}

	parentArg, _ := cmd.Flags().GetString("parent")
	if parentArg != "" {
		if pid, err := uuid.Parse(parentArg); err == nil {
			loc.Parent = pid
		}
	}

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	if err := inventory.AddLocation(loc); err != nil {
		return fmt.Errorf("failed to add location: %w", err)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Added location %s (%s, type=%s)", loc.ID, loc.Name, loc.LocationType)
	return nil
}
