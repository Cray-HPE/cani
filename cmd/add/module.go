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
	"os"
	"strconv"

	"github.com/Cray-HPE/cani/internal/util/nameexpand"
	"github.com/Cray-HPE/cani/internal/util/placement"
	"github.com/Cray-HPE/cani/internal/util/validate"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// newModuleCommand creates the "add module" subcommand.
func newModuleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module <slug-or-part-number>",
		Short: "Add module(s) to the inventory.",
		Long: `Add one or more modules to the inventory by slug or part number.

Supports strategy-based placement across multiple devices:
  --device '%{FILL}'    Fill available bays per device before moving to next
  --device hpe-xd670    Target all devices matching this slug
  --device <name|uuid>  Target a single device

Bays are auto-filtered by the module's hardware type (e.g. gpu modules
go into GPU bays). Use --bay-filter to override.

Template variables for --name: %{DEVICE}, %{BAY}, %{SEQ}`,
		Args: validSlugOrPartNumber(NounModule),
		RunE: addModule,
	}

	cmd.Flags().String("device", "", "Parent device UUID, name, slug, or strategy (%{FILL})")
	cmd.Flags().String("bay", "", "Module bay name on the parent device")
	cmd.Flags().String("bay-filter", "", "Filter bays by name substring (overrides auto-filter)")
	cmd.Flags().String("name", "", "Module name, expansion pattern, or template (%{DEVICE}, %{BAY}, %{SEQ})")
	cmd.Flags().String("location", "", "Location filter for device selection (name or UUID)")
	cmd.Flags().Bool("dry-run", false, "Show placement plan without committing changes")

	return cmd
}

func addModule(cmd *cobra.Command, args []string) error {
	qty, _ := cmd.Flags().GetInt("qty")

	result, err := lookupBySlugOrPart(NounModule, args[0])
	if err != nil {
		return err
	}

	deviceArg, _ := cmd.Flags().GetString("device")
	nameArg, _ := cmd.Flags().GetString("name")

	_, isStrategy := placement.ParseStrategy(deviceArg)
	if isStrategy {
		// When --qty is not explicitly set, pass 0 to fill all available bays.
		if !cmd.Flags().Changed("qty") {
			qty = 0
		}
		return addModuleStrategy(cmd, result, qty, nameArg, deviceArg)
	}
	if qty < 1 {
		qty = 1
	}
	return addModuleLiteral(cmd, result, qty, nameArg, deviceArg)
}

// addModuleStrategy handles multi-device auto-placement with %{FILL}.
func addModuleStrategy(cmd *cobra.Command, result *lookupResult, qty int, nameArg, deviceArg string) error {
	bayFilterArg, _ := cmd.Flags().GetString("bay-filter")
	locationArg, _ := cmd.Flags().GetString("location")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	statusArg, _ := cmd.Flags().GetString("status")
	serialArg, _ := cmd.Flags().GetString("serial")

	strategy, _ := placement.ParseStrategy(deviceArg)

	isTemplate := nameexpand.IsTemplate(nameArg)
	if !isTemplate && nameArg != "" {
		return fmt.Errorf("strategy placement requires template naming (%%{DEVICE}, %%{BAY}, etc.) or no --name flag")
	}

	if err := datastores.SetDeviceStore(cmd, nil); err != nil {
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

	devices := resolveTargetDevices(inventory, locationArg)
	if len(devices) == 0 {
		return fmt.Errorf("no devices found for module placement")
	}

	bayFilter := resolveBayFilter(bayFilterArg, result.Module.HardwareType)

	entries, err := placement.PlanModules(devices, inventory, bayFilter, qty, strategy)
	if err != nil {
		return err
	}

	names := resolveModuleTemplateNames(nameArg, entries)

	if dryRun {
		placement.PrintModulePlan(os.Stdout, entries, names)
		return nil
	}

	for i, e := range entries {
		mod := *result.Module
		mod.ID = uuid.New()
		mod.ParentDevice = e.DeviceID
		mod.ModuleBayName = e.BayName
		if i < len(names) {
			mod.Name = names[i]
		}
		if statusArg != "" {
			mod.Status = statusArg
		}
		if serialArg != "" {
			mod.Serial = serialArg
		}

		if err := inventory.AddModule(&mod); err != nil {
			return fmt.Errorf("failed to add module: %w", err)
		}
		log.Printf("Added module %s (%s) in %s bay %s", mod.ID, mod.Name, e.DeviceName, e.BayName)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("%d module(s) added via %s strategy", len(entries), strategy)
	return nil
}

// addModuleLiteral handles the original single-device flow and also
// supports device lookup by slug (all matching devices).
func addModuleLiteral(cmd *cobra.Command, result *lookupResult, qty int, nameArg, deviceArg string) error {
	bayName, _ := cmd.Flags().GetString("bay")
	bayFilterArg, _ := cmd.Flags().GetString("bay-filter")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	statusArg, _ := cmd.Flags().GetString("status")
	serialArg, _ := cmd.Flags().GetString("serial")
	prefix, _ := cmd.Flags().GetString("prefix")
	start, _ := cmd.Flags().GetInt("start")
	padWidth, _ := cmd.Flags().GetInt("pad-width")

	if err := datastores.SetDeviceStore(cmd, nil); err != nil {
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

	// Resolve target device(s): try name/UUID first, then slug match.
	var devices []*devicetypes.CaniDeviceType
	if deviceArg != "" {
		if dev := inventory.FindDeviceByNameOrID(deviceArg); dev != nil {
			devices = []*devicetypes.CaniDeviceType{dev}
		} else if slugDevices := inventory.DevicesBySlug(deviceArg); len(slugDevices) > 0 {
			devices = slugDevices
		}
	}

	// If multiple devices matched by slug and we have template naming,
	// delegate to strategy-like flow using FILL.
	if len(devices) > 1 {
		isTemplate := nameexpand.IsTemplate(nameArg)
		bayFilter := resolveBayFilter(bayFilterArg, result.Module.HardwareType)

		entries, planErr := placement.PlanModules(devices, inventory, bayFilter, qty, placement.StrategyFill)
		if planErr != nil {
			return planErr
		}

		var names []string
		if isTemplate {
			names = resolveModuleTemplateNames(nameArg, entries)
		}

		if dryRun {
			placement.PrintModulePlan(os.Stdout, entries, names)
			return nil
		}

		for i, e := range entries {
			mod := *result.Module
			mod.ID = uuid.New()
			mod.ParentDevice = e.DeviceID
			mod.ModuleBayName = e.BayName
			if names != nil && i < len(names) {
				mod.Name = names[i]
			}
			if statusArg != "" {
				mod.Status = statusArg
			}
			if serialArg != "" {
				mod.Serial = serialArg
			}
			if err := inventory.AddModule(&mod); err != nil {
				return fmt.Errorf("failed to add module: %w", err)
			}
			log.Printf("Added module %s (%s) in %s bay %s", mod.ID, mod.Name, e.DeviceName, e.BayName)
		}

		if err := datastores.Datastore.Save(inventory); err != nil {
			return fmt.Errorf("failed to save inventory: %w", err)
		}
		log.Printf("%d module(s) added", len(entries))
		return nil
	}

	// Single-device literal flow.
	names, err := nameexpand.ResolveNames(nameArg, prefix, start, padWidth, qty)
	if err != nil {
		return fmt.Errorf("name resolution failed: %w", err)
	}

	for i := range qty {
		mod := *result.Module
		mod.ID = uuid.New()
		mod.ModuleBayName = bayName

		if len(devices) == 1 {
			mod.ParentDevice = devices[0].ID
		}

		if names != nil {
			mod.Name = names[i]
		}
		if statusArg != "" {
			mod.Status = statusArg
		}
		if serialArg != "" {
			mod.Serial = serialArg
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

// resolveTargetDevices finds all devices in the inventory, optionally
// filtered by location.
func resolveTargetDevices(inventory *devicetypes.Inventory, locationArg string) []*devicetypes.CaniDeviceType {
	var devices []*devicetypes.CaniDeviceType
	if locationArg != "" {
		loc := inventory.FindLocationByNameOrID(locationArg)
		if loc == nil {
			return nil
		}
		racks := inventory.RacksByLocation(loc.ID)
		for _, rack := range racks {
			devices = append(devices, inventory.GetDevicesInRack(rack.ID)...)
		}
	} else {
		for _, dev := range inventory.Devices {
			if dev != nil {
				devices = append(devices, dev)
			}
		}
	}
	return devices
}

// resolveBayFilter returns the bay filter to use. If an explicit filter
// was provided, use it; otherwise auto-detect from the module hardware type.
func resolveBayFilter(explicit, hardwareType string) string {
	if explicit != "" {
		return explicit
	}
	return placement.BayFilterForHardwareType(hardwareType)
}

// resolveModuleTemplateNames expands template patterns for each placement entry.
func resolveModuleTemplateNames(nameArg string, entries []placement.ModulePlacementEntry) []string {
	if nameArg == "" || !nameexpand.IsTemplate(nameArg) {
		return nil
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		vars := map[string]string{
			"DEVICE": e.DeviceName,
			"BAY":    e.BayName,
			"SEQ":    strconv.Itoa(i + 1),
		}
		name, err := nameexpand.ExpandTemplate(nameArg, vars)
		if err != nil {
			log.Printf("warning: template expansion failed for entry %d: %v", i, err)
			continue
		}
		names[i] = name
	}
	return names
}
