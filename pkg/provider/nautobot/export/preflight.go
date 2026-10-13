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
package export

import (
	"fmt"
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// catalogRefs holds, by the name each phase will look up, the reference data
// an inventory needs to exist in Nautobot before anything is written.
type catalogRefs struct {
	locationTypes map[string]bool
	deviceTypes   map[string]bool
	moduleTypes   map[string]bool
	statuses      map[string]bool
	roles         map[string]bool
}

// createFlags is a snapshot of the cache's create_* options.
type createFlags struct {
	deviceTypes, locationTypes, statuses, roles bool
}

// refuseMissingRefs runs the pre-flight and, when it finds anything, records
// each gap, prints the summary and returns the error that stops Load before
// its first phase.
func (e *Exporter) refuseMissingRefs(inventory *devicetypes.Inventory, result *LoadResult) error {
	missing := e.preflight(inventory)
	if len(missing) == 0 {
		return nil
	}
	for _, gap := range missing {
		result.Errors = append(result.Errors, "pre-flight: "+gap)
	}
	e.printLoadSummary(result)
	return fmt.Errorf("pre-flight found %d missing reference object(s); nothing was exported", len(missing))
}

// preflight looks up every catalog object the inventory refers to before the
// first phase writes, with the create_* options switched off so the lookups
// cannot write either, and returns the ones that are missing and that no
// phase would create. A non-empty result refuses the whole export, so a
// missing status can no longer leave half the objects behind.
func (e *Exporter) preflight(inventory *devicetypes.Inventory) []string {
	refs := collectCatalogRefs(inventory, e.Options.DefaultStatus, e.Options.DefaultRole)
	var missing []string
	e.Cache.lookupOnly(func(creates createFlags) {
		missing = append(missing, checkRefs(refs.locationTypes, creates.locationTypes, e.lookupLocationType)...)
		missing = append(missing, checkRefs(refs.deviceTypes, creates.deviceTypes, func(slug string) error {
			_, err := e.Cache.GetDeviceType(slug)
			return err
		})...)
		missing = append(missing, checkRefs(refs.moduleTypes, e.Options.CreateModuleTypes, e.lookupModuleType)...)
		missing = append(missing, checkRefs(refs.statuses, creates.statuses, func(name string) error {
			_, err := e.Cache.GetStatus(name)
			return err
		})...)
		missing = append(missing, checkRefs(refs.roles, creates.roles, func(name string) error {
			_, err := e.Cache.GetRole(name)
			return err
		})...)
	})
	sort.Strings(missing)
	return missing
}

// lookupLocationType fails the way the location phase would when the type is
// absent, without the content-type reconciliation that phase performs.
func (e *Exporter) lookupLocationType(slug string) error {
	name := locationTypeDisplayName(slug, parentDef(slug))
	lt, err := e.Cache.findLocationType(name)
	if err != nil {
		return err
	}
	if lt == nil {
		return errLocationTypeNotFound(name)
	}
	return nil
}

// lookupModuleType fails the way the module phase would when the type is
// absent; the lookup itself never creates, so no option needs switching off.
func (e *Exporter) lookupModuleType(model string) error {
	moduleType, err := e.findModuleType(e.Cache.ctx, model)
	if err != nil {
		return err
	}
	if moduleType == nil {
		return errModuleTypeNotFound(model)
	}
	return nil
}

// checkRefs looks every name up and collects the failures, unless the phase
// that needs the objects would create them anyway.
func checkRefs(names map[string]bool, willCreate bool, lookup func(string) error) []string {
	if willCreate {
		return nil
	}
	var missing []string
	for name := range names {
		if err := lookup(name); err != nil {
			missing = append(missing, err.Error())
		}
	}
	return missing
}

// lookupOnly runs fn with every create_* option off and hands it the options
// as they were; they are restored afterwards.
func (c *LookupCache) lookupOnly(fn func(createFlags)) {
	saved := createFlags{c.createDeviceTypes, c.createLocationTypes, c.createStatuses, c.createRoles}
	c.createDeviceTypes, c.createLocationTypes, c.createStatuses, c.createRoles = false, false, false, false
	defer func() {
		c.createDeviceTypes, c.createLocationTypes, c.createStatuses, c.createRoles =
			saved.deviceTypes, saved.locationTypes, saved.statuses, saved.roles
	}()
	fn(saved)
}

// collectCatalogRefs gathers the names the loaders resolve, applying the same
// fallbacks they do: locations, racks, modules and IPAM objects default their
// status to the provider default and then "Active"; a device without a
// status or role is left for the device phase to report; interfaces are
// Active with the role their spec declares or infers; a cable is Connected
// unless it says Planned.
func collectCatalogRefs(inventory *devicetypes.Inventory, defaultStatus, defaultRole string) catalogRefs {
	refs := catalogRefs{
		locationTypes: map[string]bool{}, deviceTypes: map[string]bool{}, moduleTypes: map[string]bool{},
		statuses: map[string]bool{}, roles: map[string]bool{},
	}
	for _, loc := range inventory.Locations {
		refs.locationTypes[firstSet(loc.LocationType, "Site")] = true
		refs.statuses[firstSet(loc.Status, defaultStatus, "Active")] = true
	}
	for _, rack := range inventory.Racks {
		refs.statuses[firstSet(rack.Status, defaultStatus, "Active")] = true
	}
	refs.collectDevices(inventory, defaultStatus, defaultRole)
	refs.collectModules(inventory, defaultStatus)
	refs.collectCables(inventory)
	refs.collectIPAM(inventory, defaultStatus)
	return refs
}

func (refs catalogRefs) collectDevices(inventory *devicetypes.Inventory, defaultStatus, defaultRole string) {
	for _, device := range inventory.Devices {
		if devicetypes.ClassifyForNautobot(string(device.Type)) != devicetypes.CategoryDevice {
			continue
		}
		refs.deviceTypes[device.Slug] = true
		refs.statuses[firstSet(device.Status, defaultStatus)] = true
		refs.roles[firstSet(device.GetRole(), defaultRole)] = true
		refs.collectInterfaces(getDeviceInterfaceSpecs(device))
	}
	delete(refs.deviceTypes, "")
}

// collectInterfaces records what the interface phase resolves for the specs
// it would create: the Active status and each declared or inferred role. A
// spec of a type Nautobot cannot store is skipped there and so here.
func (refs catalogRefs) collectInterfaces(specs []interfaceSpec) {
	for _, spec := range specs {
		if !isValidNautobotInterfaceType(spec.Type) {
			continue
		}
		refs.statuses["Active"] = true
		refs.roles[spec.Role] = true
	}
}

// collectModules records each module's status with the module phase's
// fallbacks, the model it looks the module type up by, and its ports'
// references; a module's own role is optional in that phase, so a missing
// one is not a refusal.
func (refs catalogRefs) collectModules(inventory *devicetypes.Inventory, defaultStatus string) {
	for _, module := range inventory.Modules {
		if module == nil || module.Name == "" {
			continue
		}
		refs.statuses[firstSet(module.Status, defaultStatus, "Active")] = true
		refs.moduleTypes[moduleTypeModel(module)] = true
		refs.collectInterfaces(moduleInterfaceSpecs(module))
	}
}

// collectCables records the status the cable phase gives each cable it would
// create; one without both ends is skipped there and so here. A cable the
// phase would skip as already in Nautobot still requires its status: telling
// the two apart needs the per-device interface lookups the phase itself does,
// and Nautobot ships Connected and Planned, so the only export this refuses
// that the phase would complete has every cable in place and a stock status
// deleted.
func (refs catalogRefs) collectCables(inventory *devicetypes.Inventory) {
	for _, cable := range inventory.Cables {
		if cable == nil || cable.TerminationADevice == uuid.Nil || cable.TerminationBDevice == uuid.Nil {
			continue
		}
		refs.statuses[cableStatusName(cable)] = true
	}
}

func (refs catalogRefs) collectIPAM(inventory *devicetypes.Inventory, defaultStatus string) {
	for _, vlan := range inventory.VLANs {
		refs.statuses[firstSet(vlan.Status, defaultStatus, "Active")] = true
		refs.roles[vlan.Role] = true
	}
	for _, prefix := range inventory.Prefixes {
		refs.statuses[firstSet(prefix.Status, defaultStatus, "Active")] = true
		refs.roles[prefix.Role] = true
	}
	for _, addr := range inventory.IPAddresses {
		refs.statuses[firstSet(addr.Status, defaultStatus, "Active")] = true
		refs.roles[string(addr.IPRole)] = true
	}
	delete(refs.statuses, "")
	delete(refs.roles, "")
}

// firstSet returns the first non-empty value.
func firstSet(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
