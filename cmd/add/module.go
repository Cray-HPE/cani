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

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/nameexpand"
	"github.com/Cray-HPE/cani/internal/util/placement"
	"github.com/Cray-HPE/cani/internal/util/validate"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

const (
	errAddModule     = "failed to add module: %w"
	errSaveInventory = "failed to save inventory: %w"
)

// newModuleCommand creates the "add module" subcommand.
func newModuleCommand() *cli.Command {
	cmd := &cli.Command{
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
	cmd.Flags().String(flagBayFilter, "", "Filter bays by name substring (overrides auto-filter)")
	cmd.Flags().String("name", "", "Module name, expansion pattern, or template (%{DEVICE}, %{BAY}, %{SEQ})")
	cmd.Flags().String(flagLocation, "", "Location filter for device selection (name or UUID)")
	cmd.Flags().Bool(flagDryRun, false, "Show placement plan without committing changes")

	return cmd
}

func addModule(cmd *cli.Command, args []string) error {
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

// moduleAddOpts holds the parsed flags shared by the literal placement helpers.
type moduleAddOpts struct {
	qty          int
	nameArg      string
	deviceArg    string
	bayName      string
	bayFilterArg string
	prefix       string
	start        int
	padWidth     int
	statusArg    string
	serialArg    string
	dryRun       bool
}

// loadModuleInventory sets the device store from flags and loads the inventory.
func loadModuleInventory(cmd *cli.Command) (*devicetypes.Inventory, error) {
	if err := datastores.SetDeviceStore(cmd, nil); err != nil {
		return nil, fmt.Errorf("failed to set device store: %w", err)
	}
	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load inventory: %w", err)
	}
	return inventory, nil
}

// normalizeModuleStatus validates a status against the inventory, returning the
// canonical form. An empty status is returned unchanged.
func normalizeModuleStatus(statusArg string, inventory *devicetypes.Inventory) (string, error) {
	if statusArg == "" {
		return "", nil
	}
	return validate.StatusWithInventory(statusArg, inventory)
}

// applyModuleStatusSerial copies optional status and serial onto a module.
func applyModuleStatusSerial(mod *devicetypes.CaniModuleType, statusArg, serialArg string) {
	if statusArg != "" {
		mod.Status = statusArg
	}
	if serialArg != "" {
		mod.Serial = serialArg
	}
}

// commitPlannedModules creates and saves a module for each placement entry.
func commitPlannedModules(inventory *devicetypes.Inventory, base *devicetypes.CaniModuleType, entries []placement.ModulePlacementEntry, names []string, statusArg, serialArg string) error {
	for i, e := range entries {
		mod := *base
		mod.ID = uuid.New()
		mod.ParentDevice = e.DeviceID
		mod.ModuleBayName = e.BayName
		if i < len(names) {
			mod.Name = names[i]
		}
		applyModuleStatusSerial(&mod, statusArg, serialArg)
		if err := inventory.AddModule(&mod); err != nil {
			return fmt.Errorf(errAddModule, err)
		}
		log.Printf("Added module %s (%s) in %s bay %s", mod.ID, mod.Name, e.DeviceName, e.BayName)
	}
	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf(errSaveInventory, err)
	}
	return nil
}

// addModuleStrategy handles multi-device auto-placement with %{FILL}.
func addModuleStrategy(cmd *cli.Command, result *lookupResult, qty int, nameArg, deviceArg string) error {
	bayFilterArg, _ := cmd.Flags().GetString(flagBayFilter)
	locationArg, _ := cmd.Flags().GetString(flagLocation)
	dryRun, _ := cmd.Flags().GetBool(flagDryRun)
	statusArg, _ := cmd.Flags().GetString("status")
	serialArg, _ := cmd.Flags().GetString("serial")

	strategy, _ := placement.ParseStrategy(deviceArg)

	if nameArg != "" && !nameexpand.IsTemplate(nameArg) {
		return fmt.Errorf("strategy placement requires template naming (%%{DEVICE}, %%{BAY}, etc.) or no --name flag")
	}

	inventory, err := loadModuleInventory(cmd)
	if err != nil {
		return err
	}
	if statusArg, err = normalizeModuleStatus(statusArg, inventory); err != nil {
		return err
	}

	devices := resolveTargetDevices(inventory, locationArg)
	if len(devices) == 0 {
		return fmt.Errorf("no devices found for module placement")
	}

	bayFilter := resolveBayFilter(bayFilterArg, string(result.Module.Type))
	entries, err := placement.PlanModules(devices, inventory, bayFilter, qty, strategy)
	if err != nil {
		return err
	}

	names := resolveModuleTemplateNames(nameArg, entries)
	if dryRun {
		placement.PrintModulePlan(os.Stdout, entries, names)
		return nil
	}

	if err := commitPlannedModules(inventory, result.Module, entries, names, statusArg, serialArg); err != nil {
		return err
	}

	log.Printf("%d module(s) added via %s strategy", len(entries), strategy)
	return nil
}

// addModuleLiteral handles the original single-device flow and also
// supports device lookup by slug (all matching devices).
func addModuleLiteral(cmd *cli.Command, result *lookupResult, qty int, nameArg, deviceArg string) error {
	opts := parseModuleOpts(cmd, qty, nameArg, deviceArg)

	inventory, err := loadModuleInventory(cmd)
	if err != nil {
		return err
	}
	if opts.statusArg, err = normalizeModuleStatus(opts.statusArg, inventory); err != nil {
		return err
	}

	devices := resolveLiteralDevices(inventory, deviceArg)
	if len(devices) > 1 {
		return addModuleMultiDevice(inventory, result.Module, devices, opts)
	}
	return addModuleSingleDevice(inventory, result.Module, devices, opts)
}

// parseModuleOpts reads the module placement flags into a struct.
func parseModuleOpts(cmd *cli.Command, qty int, nameArg, deviceArg string) moduleAddOpts {
	opts := moduleAddOpts{qty: qty, nameArg: nameArg, deviceArg: deviceArg}
	opts.bayName, _ = cmd.Flags().GetString("bay")
	opts.bayFilterArg, _ = cmd.Flags().GetString(flagBayFilter)
	opts.dryRun, _ = cmd.Flags().GetBool(flagDryRun)
	opts.statusArg, _ = cmd.Flags().GetString("status")
	opts.serialArg, _ = cmd.Flags().GetString("serial")
	opts.prefix, _ = cmd.Flags().GetString("prefix")
	opts.start, _ = cmd.Flags().GetInt("start")
	opts.padWidth, _ = cmd.Flags().GetInt("pad-width")
	return opts
}

// resolveLiteralDevices resolves a device argument by name/UUID, then slug.
func resolveLiteralDevices(inventory *devicetypes.Inventory, deviceArg string) []*devicetypes.CaniDeviceType {
	if deviceArg == "" {
		return nil
	}
	if dev := inventory.FindDeviceByNameOrID(deviceArg); dev != nil {
		return []*devicetypes.CaniDeviceType{dev}
	}
	if slugDevices := inventory.DevicesBySlug(deviceArg); len(slugDevices) > 0 {
		return slugDevices
	}
	return nil
}

// addModuleMultiDevice plans and commits modules across all matching devices.
func addModuleMultiDevice(inventory *devicetypes.Inventory, base *devicetypes.CaniModuleType, devices []*devicetypes.CaniDeviceType, opts moduleAddOpts) error {
	bayFilter := resolveBayFilter(opts.bayFilterArg, string(base.Type))
	entries, err := placement.PlanModules(devices, inventory, bayFilter, opts.qty, placement.StrategyFill)
	if err != nil {
		return err
	}

	var names []string
	if nameexpand.IsTemplate(opts.nameArg) {
		names = resolveModuleTemplateNames(opts.nameArg, entries)
	}

	if opts.dryRun {
		placement.PrintModulePlan(os.Stdout, entries, names)
		return nil
	}

	if err := commitPlannedModules(inventory, base, entries, names, opts.statusArg, opts.serialArg); err != nil {
		return err
	}
	log.Printf("%d module(s) added", len(entries))
	return nil
}

// addModuleSingleDevice adds qty modules to a single (or no) parent device.
func addModuleSingleDevice(inventory *devicetypes.Inventory, base *devicetypes.CaniModuleType, devices []*devicetypes.CaniDeviceType, opts moduleAddOpts) error {
	names, err := resolveSingleDeviceNames(opts, devices)
	if err != nil {
		return err
	}

	for i := range opts.qty {
		mod := *base
		mod.ID = uuid.New()
		mod.ModuleBayName = opts.bayName
		if len(devices) == 1 {
			mod.ParentDevice = devices[0].ID
		}
		if names != nil {
			mod.Name = names[i]
		}
		applyModuleStatusSerial(&mod, opts.statusArg, opts.serialArg)
		if err := inventory.AddModule(&mod); err != nil {
			return fmt.Errorf(errAddModule, err)
		}
		log.Printf("Added module %s (%s)", mod.ID, mod.Name)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf(errSaveInventory, err)
	}
	log.Printf("%d module(s) added", opts.qty)
	return nil
}

// resolveSingleDeviceNames resolves module names for the single-device flow,
// expanding deferred templates against the resolved device and bay context.
func resolveSingleDeviceNames(opts moduleAddOpts, devices []*devicetypes.CaniDeviceType) ([]string, error) {
	names, err := nameexpand.ResolveNames(opts.nameArg, opts.prefix, opts.start, opts.padWidth, opts.qty)
	if err != nil {
		return nil, fmt.Errorf("name resolution failed: %w", err)
	}
	if names == nil && nameexpand.IsTemplate(opts.nameArg) {
		deviceName := ""
		if len(devices) == 1 {
			deviceName = devices[0].Name
		}
		names, err = expandLiteralModuleNames(opts.nameArg, deviceName, opts.bayName, opts.start, opts.qty)
		if err != nil {
			return nil, err
		}
	}
	return names, nil
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

// expandLiteralModuleNames expands a template name (e.g. "CX7-%{DEVICE}") for
// the single-device add flow, producing one name per quantity using the
// resolved device name and bay as context.
func expandLiteralModuleNames(nameArg, deviceName, bayName string, start, qty int) ([]string, error) {
	names := make([]string, qty)
	for i := range qty {
		vars := map[string]string{
			"DEVICE": deviceName,
			"BAY":    bayName,
			"SEQ":    strconv.Itoa(start + i),
		}
		expanded, err := nameexpand.ExpandTemplate(nameArg, vars)
		if err != nil {
			return nil, fmt.Errorf("name template expansion failed: %w", err)
		}
		names[i] = expanded
	}
	return names, nil
}
