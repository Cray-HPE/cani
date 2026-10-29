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
	"io"
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
		Use:   "module <installed> <spare>",
		Short: "Replace an installed module with a spare.",
		Long: `Install a spare module in the bay of an installed module and store the
installed module at a location, as one change. Each module is named by UUID,
name, serial number or asset tag.

The spare takes over the bay, the port names the bay gives it, and the cables
on those ports, as when a card is replaced in place: each cable on the
replaced module's ports is re-terminated at the spare's port of the same
name. The replaced module is stored at --location, or at its parent device's
location, with no cables, and keeps its name, serial and asset tag. Both
modules must be the same hardware type when both record one. The next export
with --merge moves both modules in the provider.`,
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
	// The cables on the outgoing module's ports are re-terminated on the spare.
	moved := len(inventory.CablesOnModule(installedID))
	if err := inventory.SwapModules(installedID, spareID, storeAt); err != nil {
		return fmt.Errorf("failed to swap modules: %w", err)
	}
	outcome := outcomeOf(inventory, installedID, spareID, moved)

	if dryRun, _ := cmd.Flags().GetBool(flagDryRun); dryRun {
		outcome.print(cmd.OutOrStdout())
		return nil
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}
	outcome.log()
	return nil
}

// swapOutcome is what the command reports once SwapModules has applied: the
// spare in the bay, the replaced module at its location, and the number of
// cables that moved from the replaced module's ports to the spare's.
type swapOutcome struct {
	installed, spare *devicetypes.CaniModuleType
	parent           *devicetypes.CaniDeviceType
	stock            *devicetypes.CaniLocationType
	moved            int
}

func outcomeOf(inventory *devicetypes.Inventory, installedID, spareID uuid.UUID, moved int) swapOutcome {
	installed, spare := inventory.Modules[installedID], inventory.Modules[spareID]
	return swapOutcome{
		installed: installed, spare: spare, moved: moved,
		parent: inventory.Devices[spare.ParentDevice],
		stock:  inventory.Locations[installed.Location],
	}
}

// print writes the dry-run report.
func (o swapOutcome) print(w io.Writer) {
	fmt.Fprintf(w, "Would install %s in %s bay %s\nWould store %s at location %s\n",
		o.spare.Name, o.parent.Name, o.spare.EffectiveModuleBayName(), o.installed.Name, o.stock.Name)
	if o.moved > 0 {
		fmt.Fprintf(w, "Would move %d cable(s) to the ports of %s\n", o.moved, o.spare.Name)
	}
}

// log records the saved swap.
func (o swapOutcome) log() {
	log.Printf("Swapped modules: %s (%s) installed in %s bay %s, %s (%s) stored at %s",
		o.spare.Name, o.spare.ID, o.parent.Name, o.spare.EffectiveModuleBayName(), o.installed.Name, o.installed.ID, o.stock.Name)
	if o.moved > 0 {
		log.Printf("%d cable(s) moved to the ports of %s", o.moved, o.spare.Name)
	}
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
