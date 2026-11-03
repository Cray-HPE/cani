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
	"io"
	"log"
	"os"
	"strconv"

	"github.com/Cray-HPE/cani/internal/util/nameexpand"
	"github.com/Cray-HPE/cani/internal/util/placement"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

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

	if err := commitPlannedModules(inventory, base, entries, names, opts.statusArg, opts.ident); err != nil {
		return err
	}
	log.Printf("%d module(s) added", len(entries))
	return nil
}

// addModuleSingleDevice adds qty modules to a single (or no) parent device,
// or stores them as spares. A named bay is stored in the spelling the parent
// declares, so the datastore and the ModuleBay export creates carry one name
// per bay. A dry run prints the placements and saves nothing, as the
// strategy and multi-device paths do.
func addModuleSingleDevice(inventory *devicetypes.Inventory, base *devicetypes.CaniModuleType, devices []*devicetypes.CaniDeviceType, opts moduleAddOpts) error {
	if len(devices) == 1 {
		bay, err := inventory.ResolveModuleBayName(devices[0].ID, opts.bayName)
		if err != nil {
			return err
		}
		opts.bayName = bay
	}
	names, err := resolveSingleDeviceNames(opts, devices)
	if err != nil {
		return err
	}
	if opts.dryRun {
		printSingleDevicePlan(os.Stdout, inventory, base, devices, names, opts)
		return nil
	}

	for i := range opts.qty {
		mod := base.NewInstance()
		mod.ModuleBayName = opts.bayName
		mod.Location = opts.spareLocation
		if len(devices) == 1 {
			mod.ParentDevice = devices[0].ID
		}
		if names != nil {
			mod.Name = names[i]
		}
		applyModuleStatusIdentity(mod, opts.statusArg, opts.ident)
		if err := inventory.AddModule(mod); err != nil {
			return fmt.Errorf(errAddModule, err)
		}
		log.Printf("Added module %s (%s)", mod.ID, mod.Label())
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf(errSaveInventory, err)
	}
	log.Printf("%d module(s) added", opts.qty)
	return nil
}

// printSingleDevicePlan prints one line per module addModuleSingleDevice
// would add: installed in the device and bay, stored at the spare location,
// or added with no parent. An unnamed module is described as its Label
// would read, by type and the identity the flags stamp on it.
func printSingleDevicePlan(w io.Writer, inventory *devicetypes.Inventory, base *devicetypes.CaniModuleType,
	devices []*devicetypes.CaniDeviceType, names []string, opts moduleAddOpts) {
	for i := range opts.qty {
		label := plannedModuleLabel(base, names, i, opts.ident)
		switch {
		case opts.spareLocation != uuid.Nil:
			fmt.Fprintf(w, "Would store module %s at location %s\n", label, inventory.Locations[opts.spareLocation].Name)
		case len(devices) == 1 && opts.bayName != "":
			fmt.Fprintf(w, "Would add module %s to %s bay %s\n", label, devices[0].Name, opts.bayName)
		case len(devices) == 1:
			fmt.Fprintf(w, "Would add module %s to %s\n", label, devices[0].Name)
		default:
			fmt.Fprintf(w, "Would add module %s with no parent device\n", label)
		}
	}
}

// plannedModuleLabel is the name the i-th planned module gets, or the label
// an unnamed instance would carry: its type and the serial or asset tag.
func plannedModuleLabel(base *devicetypes.CaniModuleType, names []string, i int, ident identity) string {
	if names != nil {
		return names[i]
	}
	planned := *base
	planned.ID = uuid.Nil
	ident.apply(&planned.Serial, &planned.AssetTag)
	return planned.Label()
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
