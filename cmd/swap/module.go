/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
package swap

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

const (
	flagLocation = "location"
	flagDryRun   = "dry-run"
)

// newModuleCommand creates the "swap module" subcommand.
func newModuleCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "module <installed-uuid-or-name> <spare-uuid-or-name>",
		Short: "Replace an installed module with a spare.",
		Long: `Install a spare module in the bay of an installed module and store the
installed module at a location, as one change.

The spare takes over the bay and the port names the bay gives it. The replaced
module is stored at --location, or at its parent device's location, and keeps
its name, serial, asset tag and cables. Both modules must be the same hardware
type when both record one. The next export with --merge moves both modules in
the provider.`,
		Args: cli.ExactArgs(2),
		RunE: swapModule,
	}

	cmd.Flags().String(flagLocation, "", "Location to store the replaced module at (name or UUID; default: the parent device's location)")
	cmd.Flags().Bool(flagDryRun, false, "Show the resulting placements without saving")

	return cmd
}

func swapModule(cmd *cli.Command, args []string) error {
	if err := store.Setup(cmd); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}
	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	installedID, err := resolve.Module(inventory, args[0])
	if err != nil {
		return err
	}
	spareID, err := resolve.Module(inventory, args[1])
	if err != nil {
		return err
	}
	storeAt, err := resolveStoreLocation(cmd, inventory)
	if err != nil {
		return err
	}

	// The loaded inventory is detached from the datastore, so a dry run can
	// apply the swap for validation and report the resulting placements.
	if err := inventory.SwapModules(installedID, spareID, storeAt); err != nil {
		return fmt.Errorf("failed to swap modules: %w", err)
	}
	installed, spare := inventory.Modules[installedID], inventory.Modules[spareID]
	parent, stock := inventory.Devices[spare.ParentDevice], inventory.Locations[installed.Location]

	if dryRun, _ := cmd.Flags().GetBool(flagDryRun); dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "Would install %s in %s bay %s\nWould store %s at location %s\n",
			spare.Name, parent.Name, spare.EffectiveModuleBayName(), installed.Name, stock.Name)
		return nil
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}
	log.Printf("Swapped modules: %s (%s) installed in %s bay %s, %s (%s) stored at %s",
		spare.Name, spareID, parent.Name, spare.EffectiveModuleBayName(), installed.Name, installedID, stock.Name)
	// Cable terminations name the module as their owner, so they follow it into storage.
	if n := len(inventory.GetCablesForDevice(installedID)); n > 0 {
		log.Printf("%d cable(s) remain on the ports of %s", n, installed.Name)
	}
	return nil
}

// resolveStoreLocation resolves --location, or returns uuid.Nil so the model
// stores the replaced module at its parent's location.
func resolveStoreLocation(cmd *cli.Command, inventory *devicetypes.Inventory) (uuid.UUID, error) {
	locationArg, _ := cmd.Flags().GetString(flagLocation)
	if locationArg == "" {
		return uuid.Nil, nil
	}
	loc := inventory.FindLocationByNameOrID(locationArg)
	if loc == nil {
		return uuid.Nil, fmt.Errorf("location %q not found", locationArg)
	}
	return loc.ID, nil
}
