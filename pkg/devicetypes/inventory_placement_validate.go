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
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// bayClaim identifies one bay on one parent.
type bayClaim struct {
	Parent uuid.UUID
	Bay    string
}

// claimant is the object holding a bay and whether it named the bay itself
// rather than falling back to the legacy module-name default.
type claimant struct {
	id       uuid.UUID
	authored bool
}

// validatePlacement checks that child devices and modules occupy bays their
// parents declare, that no bay is claimed twice, and that a module is either
// installed in a parent or stored at a location but not placed in two ways.
// Missing legacy bay identity is reported, never guessed.
func (inv *Inventory) validatePlacement() *RelationshipResult {
	res := &RelationshipResult{}
	inv.validateDeviceBayPlacement(res)
	inv.validateModulePlacement(res)
	return res
}

func (inv *Inventory) validateDeviceBayPlacement(res *RelationshipResult) {
	claims := make(map[bayClaim]claimant)
	for _, id := range sortedIDs(inv.Devices) {
		device := inv.Devices[id]
		if device == nil {
			continue
		}
		parent := inv.Devices[device.Parent]
		if parent == nil {
			continue
		}
		owner := fmt.Sprintf("device %q (%s)", device.Name, id)
		if device.ParentBay == "" {
			res.Warnings = append(res.Warnings, fmt.Sprintf(
				"%s: installed in parent %q without a named device bay; the placement cannot be validated", owner, parent.Name))
			continue
		}
		if !declaresDeviceBay(parent, device.ParentBay) {
			res.Errors = append(res.Errors, fmt.Errorf("%s: device bay %q is not declared on parent %q",
				owner, device.ParentBay, parent.Name))
		}
		claimBay(res, claims, bayClaim{parent.ID, device.ParentBay}, claimant{id, true}, owner, "device bay", parent.Name)
	}
}

func (inv *Inventory) validateModulePlacement(res *RelationshipResult) {
	claims := make(map[bayClaim]claimant)
	for _, id := range sortedIDs(inv.Modules) {
		mod := inv.Modules[id]
		if mod == nil {
			continue
		}
		owner := fmt.Sprintf("module %q (%s)", mod.Name, id)
		parent := inv.Devices[mod.ParentDevice]
		if parent == nil {
			// A missing parent device is reported by validateModuleRelationships.
			if mod.ParentDevice == uuid.Nil {
				inv.validateSpare(res, mod, owner)
			}
			continue
		}
		inv.validateInstalledModule(res, claims, id, mod, parent, owner)
	}
}

// validateSpare checks a module with no parent device: one stored at a
// location requires that location to exist, as Nautobot protects it from
// deletion; one with no location either is reported. A bay name on a spare
// describes no placement, since the bay belongs to the parent the module
// left or never had, so it is reported rather than read.
func (inv *Inventory) validateSpare(res *RelationshipResult, mod *CaniModuleType, owner string) {
	if mod.Location == uuid.Nil {
		res.Warnings = append(res.Warnings, fmt.Sprintf("%s: neither installed in a device nor stored at a location", owner))
		return
	}
	if inv.Locations[mod.Location] == nil {
		res.Errors = append(res.Errors, fmt.Errorf("%s: stored at location %s, which does not exist", owner, mod.Location))
	}
	if mod.ModuleBayName != "" {
		res.Warnings = append(res.Warnings, fmt.Sprintf(
			"%s: stored at a location but names module bay %q; a spare has no bay and the name is ignored until the module is installed",
			owner, mod.ModuleBayName))
	}
}

// validateInstalledModule checks an installed module's bay and that any
// authored location agrees with the parent's. A module with neither a bay
// nor a name to stand in for one has no bay for export to place it in, so it
// is reported with the remedy and claims no bay. An authored location that
// differs from the parent's is stale, since the parent's location governs an
// installed module, and is reported as a warning so an inventory whose
// device moved stays writable.
func (inv *Inventory) validateInstalledModule(
	res *RelationshipResult, claims map[bayClaim]claimant, id uuid.UUID,
	mod *CaniModuleType, parent *CaniDeviceType, owner string,
) {
	if mod.Location != uuid.Nil && parent.Location != uuid.Nil && mod.Location != parent.Location {
		res.Warnings = append(res.Warnings, fmt.Sprintf("%s: installed in %q at location %s but authored at location %s; the parent's location governs",
			owner, parent.Name, parent.Location, mod.Location))
	}
	bay := mod.EffectiveModuleBayName()
	switch {
	case bay == "":
		res.Warnings = append(res.Warnings, fmt.Sprintf(
			"%s: installed in %q without a module bay or a name to stand in for one; export refuses the module until a bay is set with `update module --bay`",
			owner, parent.Name))
		return
	case mod.ModuleBayName == "" && len(moduleBaysOf(parent)) > 0:
		res.Warnings = append(res.Warnings, fmt.Sprintf(
			"%s: installed in %q without a named module bay; the module name %q is used as the bay, "+
				"and export refuses the module until a bay is set with `update module --bay`", owner, parent.Name, bay))
	case !declaresModuleBay(parent, bay):
		res.Errors = append(res.Errors, fmt.Errorf("%s: module bay %q is not declared on parent %q", owner, bay, parent.Name))
	}
	claimBay(res, claims, bayClaim{parent.ID, bay}, claimant{id, mod.ModuleBayName != ""}, owner, "module bay", parent.Name)
}

// claimBay records a bay claim. Bay names compare case-insensitively, so two
// spellings of one bay are one claim. A second claim is an error when both
// claimants named the bay; when either relies on the legacy module-name
// default it is a warning, so inventories written before bays were named stay
// writable.
func claimBay(res *RelationshipResult, claims map[bayClaim]claimant, claim bayClaim, c claimant, owner, kind, parentName string) {
	key := bayClaim{claim.Parent, strings.ToLower(claim.Bay)}
	first, taken := claims[key]
	if !taken {
		claims[key] = c
		return
	}
	msg := fmt.Sprintf("%s: %s %q on %q is already occupied by %s", owner, kind, claim.Bay, parentName, first.id)
	if first.authored && c.authored {
		res.Errors = append(res.Errors, errors.New(msg))
		return
	}
	res.Warnings = append(res.Warnings, msg+"; a module name is standing in for an unnamed bay")
}

// declaresDeviceBay reports whether the parent (or its library type) declares
// the bay, or declares no bays at all (ad hoc hardware without a library
// definition). Names compare case-insensitively, as InstallInBay admits them.
func declaresDeviceBay(parent *CaniDeviceType, bay string) bool {
	bays := deviceBaysOf(parent)
	if len(bays) == 0 {
		return true
	}
	for _, spec := range bays {
		if strings.EqualFold(spec.Name, bay) {
			return true
		}
	}
	return false
}

// declaresModuleBay reports whether the parent (or its library type) declares
// the bay, or declares no bays at all. Names compare case-insensitively, as
// ResolveModuleNames admits them when resolving the bay position.
func declaresModuleBay(parent *CaniDeviceType, bay string) bool {
	bays := moduleBaysOf(parent)
	if len(bays) == 0 {
		return true
	}
	for _, spec := range bays {
		if strings.EqualFold(spec.Name, bay) {
			return true
		}
	}
	return false
}
