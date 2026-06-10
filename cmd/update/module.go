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

// newModuleCommand creates the "update module" subcommand.
func newModuleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module <uuid-or-name>",
		Short: "Update a module in the inventory.",
		Long:  "Update a module's fields by UUID or name.",
		Args:  cobra.ExactArgs(1),
		RunE:  updateModule,
	}

	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("status", "", "New status")
	cmd.Flags().String("role", "", "New role")
	cmd.Flags().String(flagDescription, "", "Description")
	cmd.Flags().String("bay", "", "Module bay name")

	return cmd
}

func updateModule(cmd *cobra.Command, args []string) error {
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	id, err := resolve.Module(inventory, args[0])
	if err != nil {
		return err
	}

	mod := inventory.Modules[id]

	if cmd.Flags().Changed("name") {
		mod.Name, _ = cmd.Flags().GetString("name")
	}
	if cmd.Flags().Changed("status") {
		mod.Status, _ = cmd.Flags().GetString("status")
	}
	if cmd.Flags().Changed("role") {
		mod.Role, _ = cmd.Flags().GetString("role")
	}
	if cmd.Flags().Changed(flagDescription) {
		mod.Description, _ = cmd.Flags().GetString(flagDescription)
	}
	if cmd.Flags().Changed("bay") {
		mod.ModuleBayName, _ = cmd.Flags().GetString("bay")
	}

	if err := applySetToModule(cmd, mod); err != nil {
		return err
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("Updated module %s (%s)", id, mod.Name)
	return nil
}

func applySetToModule(cmd *cobra.Command, mod *devicetypes.CaniModuleType) error {
	sets, _ := cmd.Flags().GetStringArray("set")
	pairs, err := parseSetFlags(sets)
	if err != nil {
		return err
	}
	for k, v := range pairs {
		switch k {
		case "name":
			mod.Name = v
		case "status":
			mod.Status = v
		case "role":
			mod.Role = v
		case flagDescription:
			mod.Description = v
		case "module_bay_name":
			mod.ModuleBayName = v
		case "serial":
			mod.Serial = v
		case "asset_tag":
			mod.AssetTag = v
		default:
			return fmt.Errorf("unknown module field: %s", k)
		}
	}
	return nil
}
