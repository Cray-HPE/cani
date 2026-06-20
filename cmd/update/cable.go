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

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// newCableCommand creates the "update cable" subcommand.
func newCableCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "cable <uuid-or-label>",
		Short: "Update a cable in the inventory.",
		Long:  "Update a cable's fields by UUID or label.",
		Args:  cli.ExactArgs(1),
		RunE:  updateCable,
	}

	cmd.Flags().String("label", "", "New label")
	cmd.Flags().String("status", "", "New status")
	cmd.Flags().String("color", "", "Cable color")
	cmd.Flags().String(flagDescription, "", "Description")

	return cmd
}

func updateCable(cmd *cli.Command, args []string) error {
	if err := store.Setup(cmd); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	id, err := resolve.Cable(inventory, args[0])
	if err != nil {
		return err
	}

	cable := inventory.Cables[id]

	if cmd.Flags().Changed("label") {
		cable.Label, _ = cmd.Flags().GetString("label")
	}
	if cmd.Flags().Changed("status") {
		cable.Status, _ = cmd.Flags().GetString("status")
	}
	if cmd.Flags().Changed("color") {
		cable.Color, _ = cmd.Flags().GetString("color")
	}
	if cmd.Flags().Changed(flagDescription) {
		cable.Description, _ = cmd.Flags().GetString(flagDescription)
	}

	if err := applySetToCable(cmd, cable); err != nil {
		return err
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Updated cable %s (%s)", id, cable.Label)
	return nil
}

func applySetToCable(cmd *cli.Command, cable *devicetypes.CaniCableType) error {
	sets, _ := cmd.Flags().GetStringArray("set")
	pairs, err := parseSetFlags(sets)
	if err != nil {
		return err
	}
	for k, v := range pairs {
		switch k {
		case "label":
			cable.Label = v
		case "status":
			cable.Status = v
		case "color":
			cable.Color = v
		case flagDescription:
			cable.Description = v
		default:
			return fmt.Errorf("unknown cable field: %s", k)
		}
	}
	return nil
}
