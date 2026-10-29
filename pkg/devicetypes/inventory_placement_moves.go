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
package devicetypes

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ModulePlacement is where a module goes: a bay on a parent device, or a
// storage location as a spare. Exactly one of ParentDevice and Location is
// set; Bay accompanies ParentDevice.
type ModulePlacement struct {
	ParentDevice uuid.UUID
	Bay          string
	Location     uuid.UUID
}

// placementSnapshot holds what a placement change overwrites, so a refused
// change can be undone without leaving renamed ports behind.
type placementSnapshot struct {
	placement  ModulePlacement
	interfaces []InterfaceSpec
}

func snapshotPlacement(mod *CaniModuleType) placementSnapshot {
	return placementSnapshot{
		placement:  ModulePlacement{ParentDevice: mod.ParentDevice, Bay: mod.ModuleBayName, Location: mod.Location},
		interfaces: mod.Interfaces,
	}
}

// PlaceModule moves a module to the placement as one unit: installed in a bay
// its parent declares with no authored location, or stored at a location that
// exists and admits modules with no parent. The bay is stored in the spelling
// the parent declares. When the rebuilt inventory rejects the result (an
// occupied bay, for example) the previous placement is restored and the error
// returned, as AddModule does for a new module.
func (inv *Inventory) PlaceModule(id uuid.UUID, target ModulePlacement) error {
	mod := inv.Modules[id]
	if mod == nil {
		return fmt.Errorf("module %s not found", id)
	}
	resolved, err := inv.resolvePlacement(mod, target)
	if err != nil {
		return err
	}
	before := snapshotPlacement(mod)
	inv.applyPlacement(mod, resolved)
	if err := inv.commitPlacements(id); err != nil {
		inv.restorePlacement(mod, before)
		inv.RebuildDerivedState()
		return err
	}
	return nil
}

// SwapModules stores an installed module at a location and installs a spare
// in the bay it leaves, as one change: the two placements are validated
// together and both are restored when either is refused. A nil storeAt
// stores the outgoing module at its parent's location. The modules must be
// the same hardware type when both record one, and the outgoing module must
// name its bay when the parent declares bays, so the spare is never installed
// into a bay named after the module it replaces.
func (inv *Inventory) SwapModules(installedID, spareID, storeAt uuid.UUID) error {
	installed, spare := inv.Modules[installedID], inv.Modules[spareID]
	if err := checkSwapPair(installed, installedID, spare, spareID); err != nil {
		return err
	}
	bay, stock, err := inv.swapTargets(installed, storeAt)
	if err != nil {
		return err
	}
	outgoing, incoming := snapshotPlacement(installed), snapshotPlacement(spare)
	inv.applyPlacement(installed, stock)
	inv.applyPlacement(spare, bay)
	if err := inv.commitPlacements(installedID, spareID); err != nil {
		inv.restorePlacement(spare, incoming)
		inv.restorePlacement(installed, outgoing)
		inv.RebuildDerivedState()
		return err
	}
	return nil
}

// checkSwapPair rejects a pair that is not one installed module and one spare
// of the same hardware type.
func checkSwapPair(installed *CaniModuleType, installedID uuid.UUID, spare *CaniModuleType, spareID uuid.UUID) error {
	if installed == nil {
		return fmt.Errorf("module %s not found", installedID)
	}
	if spare == nil {
		return fmt.Errorf("module %s not found", spareID)
	}
	if installedID == spareID {
		return fmt.Errorf("module %q (%s) cannot be swapped with itself", installed.Name, installedID)
	}
	if installed.ParentDevice == uuid.Nil {
		return fmt.Errorf("module %q (%s) is not installed in a device", installed.Name, installedID)
	}
	if !spare.IsSpare() {
		return fmt.Errorf("module %q (%s) is not a spare stored at a location", spare.Name, spareID)
	}
	if a, b := moduleTypeKey(installed), moduleTypeKey(spare); a != "" && b != "" && !strings.EqualFold(a, b) {
		return fmt.Errorf("module %q is a %s and spare %q is a %s; swap like for like", installed.Name, a, spare.Name, b)
	}
	return nil
}

// moduleTypeKey is the hardware type a module records, or "" for a module
// imported without one.
func moduleTypeKey(mod *CaniModuleType) string {
	for _, key := range []string{mod.Slug, mod.Model, mod.PartNumber} {
		if key != "" {
			return key
		}
	}
	return ""
}

// swapTargets resolves the bay the spare takes over and the location the
// outgoing module is stored at.
func (inv *Inventory) swapTargets(installed *CaniModuleType, storeAt uuid.UUID) (bay, stock ModulePlacement, err error) {
	parent := inv.Devices[installed.ParentDevice]
	if parent == nil {
		return bay, stock, fmt.Errorf("parent device %s of module %q not found", installed.ParentDevice, installed.Name)
	}
	if installed.ModuleBayName == "" && len(moduleBaysOf(parent)) > 0 {
		return bay, stock, fmt.Errorf("module %q has no named bay in %q; set one with `update module --bay` before swapping",
			installed.Name, parent.Name)
	}
	if storeAt == uuid.Nil {
		storeAt = parent.Location
	}
	if storeAt == uuid.Nil {
		return bay, stock, fmt.Errorf("device %q has no location to store %q at; name one", parent.Name, installed.Name)
	}
	stock, err = inv.resolveStoreTarget(ModulePlacement{Location: storeAt})
	bay = ModulePlacement{ParentDevice: parent.ID, Bay: installed.ModuleBayName}
	return bay, stock, err
}

// resolvePlacement checks the target names exactly one placement and
// resolves it against the inventory.
func (inv *Inventory) resolvePlacement(mod *CaniModuleType, target ModulePlacement) (ModulePlacement, error) {
	switch {
	case target.ParentDevice != uuid.Nil && target.Location != uuid.Nil:
		return target, fmt.Errorf("module %q (%s): a module is installed in a device or stored at a location, not both", mod.Name, mod.ID)
	case target.ParentDevice != uuid.Nil:
		return inv.resolveInstallTarget(mod, target)
	case target.Location != uuid.Nil:
		return inv.resolveStoreTarget(target)
	}
	return target, fmt.Errorf("module %q (%s): name a parent device or a location", mod.Name, mod.ID)
}

// resolveInstallTarget requires a declared bay on a parent that declares bays
// and stores the bay in the parent's spelling.
func (inv *Inventory) resolveInstallTarget(mod *CaniModuleType, target ModulePlacement) (ModulePlacement, error) {
	parent := inv.Devices[target.ParentDevice]
	if parent == nil {
		return target, fmt.Errorf("device %s not found", target.ParentDevice)
	}
	if target.Bay == "" && len(moduleBaysOf(parent)) > 0 {
		return target, fmt.Errorf("device %q declares module bays; name the bay to install %q in", parent.Name, mod.Name)
	}
	bay, err := inv.ResolveModuleBayName(parent.ID, target.Bay)
	if err != nil {
		return target, err
	}
	return ModulePlacement{ParentDevice: parent.ID, Bay: bay}, nil
}

// resolveStoreTarget requires a location that exists and admits modules.
func (inv *Inventory) resolveStoreTarget(target ModulePlacement) (ModulePlacement, error) {
	loc := inv.Locations[target.Location]
	if loc == nil {
		return target, fmt.Errorf("location %s not found", target.Location)
	}
	if err := loc.ValidateContentType("module"); err != nil {
		return target, err
	}
	return ModulePlacement{Location: loc.ID}, nil
}

// applyPlacement writes the placement as one unit and resolves the port names
// the new bay gives the module.
func (inv *Inventory) applyPlacement(mod *CaniModuleType, p ModulePlacement) {
	mod.ParentDevice, mod.ModuleBayName, mod.Location = p.ParentDevice, p.Bay, p.Location
	inv.ResolveModuleNames(mod)
}

// restorePlacement undoes a refused placement, including the port names the
// attempted bay gave the module; the caller rebuilds the derived state.
func (inv *Inventory) restorePlacement(mod *CaniModuleType, before placementSnapshot) {
	inv.applyPlacement(mod, before.placement)
	mod.Interfaces = before.interfaces
}

// commitPlacements rebuilds the inventory and returns the errors about the
// placed modules, or every error when none names them.
func (inv *Inventory) commitPlacements(ids ...uuid.UUID) error {
	return inv.VerifyParentChildRelationships().errorsAbout(ids...)
}
