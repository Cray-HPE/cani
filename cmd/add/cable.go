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

	"github.com/Cray-HPE/cani/internal/util/nameexpand"
	"github.com/Cray-HPE/cani/internal/util/validate"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// newCableCommand creates the "add cable" subcommand.
func newCableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cable <slug-or-part-number>",
		Short: "Add cable(s) to the inventory.",
		Long:  "Add one or more cables to the inventory by slug or part number.",
		Args:  validSlugOrPartNumber(NounCable),
		RunE:  addCable,
	}

	cmd.Flags().String("a-device", "", "Termination A device UUID or name")
	cmd.Flags().String("a-port", "", "Termination A port name")
	cmd.Flags().String("b-device", "", "Termination B device UUID or name")
	cmd.Flags().String("b-port", "", "Termination B port name")
	cmd.Flags().String("label", "", "Cable label")
	cmd.Flags().String("color", "", "Cable color")
	cmd.Flags().String("name", "", "Cable name or expansion pattern")

	return cmd
}

func addCable(cmd *cobra.Command, args []string) error {
	qty, _ := cmd.Flags().GetInt("qty")
	if qty < 1 {
		qty = 1
	}

	result, err := lookupBySlugOrPart(NounCable, args[0])
	if err != nil {
		return err
	}

	aDeviceArg, _ := cmd.Flags().GetString("a-device")
	aPort, _ := cmd.Flags().GetString("a-port")
	bDeviceArg, _ := cmd.Flags().GetString("b-device")
	bPort, _ := cmd.Flags().GetString("b-port")
	label, _ := cmd.Flags().GetString("label")
	color, _ := cmd.Flags().GetString("color")
	statusArg, _ := cmd.Flags().GetString("status")
	nameArg, _ := cmd.Flags().GetString("name")
	prefix, _ := cmd.Flags().GetString("prefix")
	start, _ := cmd.Flags().GetInt("start")
	padWidth, _ := cmd.Flags().GetInt("pad-width")

	names, err := nameexpand.ResolveNames(nameArg, prefix, start, padWidth, qty)
	if err != nil {
		return fmt.Errorf("name resolution failed: %w", err)
	}

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	if statusArg != "" {
		normalized, verr := validate.StatusWithInventory(statusArg, inventory)
		if verr != nil {
			return verr
		}
		statusArg = normalized
	}

	for i := range qty {
		cable := *result.Cable
		cable.ID = uuid.New()

		if names != nil {
			cable.Label = names[i]
		} else if label != "" {
			cable.Label = label
		}
		if color != "" {
			cable.Color = color
		}

		if aDeviceArg != "" {
			if aid, aerr := uuid.Parse(aDeviceArg); aerr == nil {
				cable.TerminationADevice = aid
			}
		}
		cable.TerminationAPort = aPort

		if bDeviceArg != "" {
			if bid, berr := uuid.Parse(bDeviceArg); berr == nil {
				cable.TerminationBDevice = bid
			}
		}
		cable.TerminationBPort = bPort

		if statusArg != "" {
			cable.Status = statusArg
		}

		if err := inventory.AddCable(&cable); err != nil {
			return fmt.Errorf("failed to add cable: %w", err)
		}
		log.Printf("Added cable %s (%s)", cable.ID, cable.Label)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("%d cable(s) added", qty)
	return nil
}
