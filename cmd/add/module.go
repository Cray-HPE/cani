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

// newModuleCommand creates the "add module" subcommand.
func newModuleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module <slug-or-part-number>",
		Short: "Add module(s) to the inventory.",
		Long:  "Add one or more modules to the inventory by slug or part number.",
		Args:  validSlugOrPartNumber(NounModule),
		RunE:  addModule,
	}

	cmd.Flags().String("device", "", "Parent device UUID or name")
	cmd.Flags().String("bay", "", "Module bay name on the parent device")

	return cmd
}

func addModule(cmd *cobra.Command, args []string) error {
	qty, _ := cmd.Flags().GetInt("qty")
	if qty < 1 {
		qty = 1
	}

	result, err := lookupBySlugOrPart(NounModule, args[0])
	if err != nil {
		return err
	}

	deviceArg, _ := cmd.Flags().GetString("device")
	bayName, _ := cmd.Flags().GetString("bay")

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	for range qty {
		mod := *result.Module
		mod.ID = uuid.New()
		mod.ModuleBayName = bayName

		if deviceArg != "" {
			if did, derr := uuid.Parse(deviceArg); derr == nil {
				mod.ParentDevice = did
			}
		}

		if err := inventory.AddModule(&mod); err != nil {
			return fmt.Errorf("failed to add module: %w", err)
		}
		log.Printf("Added module %s (%s)", mod.ID, mod.Name)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("%d module(s) added", qty)
	return nil
}
