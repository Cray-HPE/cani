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

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/internal/util/validate"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// newRackUpdateCommand creates the "update rack" subcommand.
func newRackUpdateCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "rack <uuid-or-name>",
		Short: "Update a rack in the inventory.",
		Long:  "Update a rack's fields by UUID or name.",
		Args:  cli.ExactArgs(1),
		RunE:  updateRack,
	}

	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("status", "", "New status")
	cmd.Flags().String("role", "", "New role")
	cmd.Flags().String(flagDescription, "", "Description")
	cmd.Flags().Int(flagUHeight, 0, "Rack unit height")
	cmd.Flags().String(flagLocation, "", "Parent location UUID or name")

	return cmd
}

func updateRack(cmd *cli.Command, args []string) error {
	if err := store.Setup(cmd); err != nil {
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

	if err := applyRackScalarFlags(cmd, rack); err != nil {
		return err
	}
	if err := applyRackLocation(cmd, inventory, rack); err != nil {
		return err
	}
	if err := applyRackTagsMetadata(cmd, rack); err != nil {
		return err
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

// applyRackScalarFlags applies the simple scalar rack flags when set.
func applyRackScalarFlags(cmd *cli.Command, rack *devicetypes.CaniRackType) error {
	if cmd.Flags().Changed("name") {
		rack.Name, _ = cmd.Flags().GetString("name")
	}
	if cmd.Flags().Changed("status") {
		s, _ := cmd.Flags().GetString("status")
		normalized, err := validate.Status(s)
		if err != nil {
			return err
		}
		rack.Status = normalized
	}
	if cmd.Flags().Changed("role") {
		rack.Role, _ = cmd.Flags().GetString("role")
	}
	if cmd.Flags().Changed(flagDescription) {
		rack.Description, _ = cmd.Flags().GetString(flagDescription)
	}
	if cmd.Flags().Changed(flagUHeight) {
		rack.UHeight, _ = cmd.Flags().GetInt(flagUHeight)
	}
	return nil
}

// applyRackLocation resolves and applies the --location flag when set.
func applyRackLocation(cmd *cli.Command, inventory *devicetypes.Inventory, rack *devicetypes.CaniRackType) error {
	if !cmd.Flags().Changed(flagLocation) {
		return nil
	}
	locRef, _ := cmd.Flags().GetString(flagLocation)
	locID, lerr := resolve.Location(inventory, locRef)
	if lerr != nil {
		return fmt.Errorf("resolving location: %w", lerr)
	}
	rack.Location = locID
	return nil
}

// applyRackTagsMetadata applies the --tag and --metadata flags when set.
func applyRackTagsMetadata(cmd *cli.Command, rack *devicetypes.CaniRackType) error {
	if cmd.Flags().Changed("tag") {
		tags, _ := cmd.Flags().GetStringArray("tag")
		rack.Tags = tags
	}
	if cmd.Flags().Changed("metadata") {
		pairs, _ := cmd.Flags().GetStringArray("metadata")
		if err := applyProviderMetadata(&rack.ProviderMetadata, pairs); err != nil {
			return err
		}
	}
	return nil
}

func applySetToRack(cmd *cli.Command, rack *devicetypes.CaniRackType) error {
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
			normalized, err := validate.Status(v)
			if err != nil {
				return err
			}
			rack.Status = normalized
		case "role":
			rack.Role = v
		case flagDescription:
			rack.Description = v
		case "u_height":
			n, nerr := strconv.Atoi(v)
			if nerr != nil {
				return fmt.Errorf("invalid u_height value: %s", v)
			}
			rack.UHeight = n
		case "tag":
			rack.Tags = append(rack.Tags, v)
		case flagLocation:
			return fmt.Errorf("use --location flag instead of --set location=...")
		default:
			return fmt.Errorf("unknown rack field: %s", k)
		}
	}
	return nil
}
