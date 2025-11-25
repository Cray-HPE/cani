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
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// newRackAddCommand creates the "add rack" subcommand.
func newRackAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rack <slug-or-part-number>",
		Short: "Add rack(s) to the inventory.",
		Long:  "Add one or more racks to the inventory by slug or part number.",
		Args:  validSlugOrPartNumber(NounRack),
		RunE:  addRack,
	}

	cmd.Flags().String("location", "", "Parent location UUID or name")

	return cmd
}

func addRack(cmd *cobra.Command, args []string) error {
	qty, _ := cmd.Flags().GetInt("qty")
	if qty < 1 {
		qty = 1
	}

	result, err := lookupBySlugOrPart(NounRack, args[0])
	if err != nil {
		return err
	}

	locationArg, _ := cmd.Flags().GetString("location")
	var locationID uuid.UUID
	if locationArg != "" {
		if pid, perr := uuid.Parse(locationArg); perr == nil {
			locationID = pid
		}
	}

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	if locationID == uuid.Nil {
		locationID = inventory.EnsureLocation()
	}

	for range qty {
		rack := *result.Rack // shallow copy
		rack.ID = uuid.New()
		rack.Location = locationID
		if rack.Name == "" && rack.Model != "" {
			rack.Name = rack.Model
		}
		if err := inventory.AddRack(&rack); err != nil {
			return fmt.Errorf("failed to add rack: %w", err)
		}
		inventory.AssignRacksToLocation(locationID)

		log.Printf("Added rack %s (%s)", rack.ID, rack.Name)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("%d rack(s) added", qty)
	return nil
}
