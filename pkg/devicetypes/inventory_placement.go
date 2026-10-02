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
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
)

// EffectiveModuleBayName returns the bay a module occupies on its parent:
// the authored ModuleBayName, or the module's own name as the legacy default.
func (m *CaniModuleType) EffectiveModuleBayName() string {
	if m == nil {
		return ""
	}
	if m.ModuleBayName != "" {
		return m.ModuleBayName
	}
	return m.Name
}

// IsSpare reports whether the module is stored at a location rather than
// installed in a parent device.
func (m *CaniModuleType) IsSpare() bool {
	return m != nil && m.ParentDevice == uuid.Nil && m.Location != uuid.Nil
}

// errorsAbout returns the errors that name id, or every error when none does,
// so a caller rejecting one object reports that object's problems first while
// an inventory that is broken elsewhere still refuses the write.
func (r *RelationshipResult) errorsAbout(id uuid.UUID) error {
	if r == nil || len(r.Errors) == 0 {
		return nil
	}
	var own []error
	for _, err := range r.Errors {
		if strings.Contains(err.Error(), id.String()) {
			own = append(own, err)
		}
	}
	if len(own) == 0 {
		return r.Err()
	}
	return errors.Join(own...)
}

// ResolveModuleBayName returns the declared name of the module bay that name
// selects on the parent device, matching case-insensitively as InstallInBay
// does for device bays, so an authored bay carries the library's casing. An
// empty name, an unknown parent, or a parent that declares no bays returns
// name unchanged; a bay the parent does not declare is an error that lists
// the declared bays.
func (inv *Inventory) ResolveModuleBayName(parentID uuid.UUID, name string) (string, error) {
	parent := inv.Devices[parentID]
	bays := moduleBaysOf(parent)
	if name == "" || len(bays) == 0 {
		return name, nil
	}
	names := make([]string, 0, len(bays))
	for _, spec := range bays {
		if strings.EqualFold(spec.Name, name) {
			return spec.Name, nil
		}
		names = append(names, spec.Name)
	}
	return "", fmt.Errorf("device %q has no module bay %q (bays: %s)", parent.Name, name, strings.Join(names, ", "))
}

// ReconcileModuleBays canonicalises the bay of every installed module that
// remap names, so a provider's spelling of a declared bay is stored as the
// parent declares it. A bay the parent does not declare is cleared, leaving
// the module installed without a named bay (the state inventories written
// before bays were named are in), and reported for the operator to set with
// `update module --bay`; the provider's own placement descriptor remains
// wherever the provider recorded it. The report is returned in a stable
// order. Nothing is guessed and nothing is dropped.
func (inv *Inventory) ReconcileModuleBays(remap map[uuid.UUID]uuid.UUID) []string {
	incomingIDs := make([]uuid.UUID, 0, len(remap))
	for id := range remap {
		incomingIDs = append(incomingIDs, id)
	}
	slices.SortFunc(incomingIDs, func(a, b uuid.UUID) int { return bytes.Compare(a[:], b[:]) })
	seen := make(map[uuid.UUID]bool, len(remap))
	var unresolved []string
	for _, incomingID := range incomingIDs {
		id := remap[incomingID]
		mod := inv.Modules[id]
		if seen[id] || mod == nil || mod.ParentDevice == uuid.Nil || mod.ModuleBayName == "" {
			continue
		}
		seen[id] = true
		bay, err := inv.ResolveModuleBayName(mod.ParentDevice, mod.ModuleBayName)
		if err == nil {
			mod.ModuleBayName = bay
			continue
		}
		unresolved = append(unresolved, fmt.Sprintf(
			"module %q (%s): %v; imported without a bay, set one with `update module --bay`", mod.Name, id, err))
		mod.ModuleBayName = ""
	}
	return unresolved
}
