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
	"slices"
	"strings"

	"github.com/google/uuid"
)

// ModuleToken is the Nautobot device-type-library placeholder for the
// position of the module bay a module is installed in, as in "HSN {module}/0".
const ModuleToken = "{module}"

// ResolveModuleNames replaces {module} in the module's interface names and
// labels with the position of its module bay on the parent device, as
// Nautobot does when it installs a module. Unresolved templates are retained
// on each spec for subsequent moves and datastore round trips. The library
// supplies missing templates when the interface lists line up. A bay without
// a position uses its name; a module that is not in a bay is left unchanged.
// Registered instances keep matching cable port names and indexed interfaces
// in sync without changing interface UUIDs.
func (inv *Inventory) ResolveModuleNames(mod *CaniModuleType) {
	if mod == nil {
		return
	}
	position := inv.moduleBayPosition(mod.ParentDevice, mod.ModuleBayName)
	if position == "" {
		return
	}
	templates := moduleInterfaceTemplates(mod)
	// Copy first: an instance built by struct copy shares its slice with the
	// library type, which must keep its {module} placeholders.
	mod.Interfaces = slices.Clone(mod.Interfaces)
	owner := interfaceOwner{deviceID: mod.ParentDevice, ownerID: mod.ID}
	for i := range mod.Interfaces {
		iface := &mod.Interfaces[i]
		before := *iface
		iface.Name, iface.NameTemplate = resolveModuleValue(iface.Name, iface.NameTemplate, templates[i].Name, position)
		iface.Label, iface.LabelTemplate = resolveModuleValue(iface.Label, iface.LabelTemplate, templates[i].Label, position)
		// An unregistered copy may still carry another instance's IDs.
		if inv.Modules[mod.ID] == mod {
			inv.syncInterfaceRename(owner, before, iface)
		}
	}
}

func resolveModuleValue(value string, original *string, fallback, position string) (string, *string) {
	if original == nil {
		switch {
		case strings.Contains(value, ModuleToken):
			original = &value
		case strings.Contains(fallback, ModuleToken):
			original = &fallback
		default:
			return fallback, nil
		}
	}
	return strings.ReplaceAll(*original, ModuleToken, position), original
}

// moduleInterfaceTemplates returns the unresolved interface specs for mod:
// its library type's when they match one-to-one, otherwise its own.
func moduleInterfaceTemplates(mod *CaniModuleType) []InterfaceSpec {
	if tmpl, ok := GetModuleBySlug(mod.Slug); ok && len(tmpl.Interfaces) == len(mod.Interfaces) {
		return tmpl.Interfaces
	}
	return mod.Interfaces
}

// moduleBayPosition returns the position of the named module bay on the
// device, or the bay name when the bay has no position or is not declared.
func (inv *Inventory) moduleBayPosition(deviceID uuid.UUID, bay string) string {
	if bay == "" {
		return ""
	}
	for _, spec := range moduleBaysOf(inv.Devices[deviceID]) {
		if strings.EqualFold(spec.Name, bay) && spec.Position != "" {
			return spec.Position
		}
	}
	return bay
}

// moduleBaysOf returns the device's module bays, falling back to its library
// device type for instances saved without bay specs.
func moduleBaysOf(device *CaniDeviceType) []ModuleBaySpec {
	if device == nil {
		return nil
	}
	if len(device.ModuleBays) > 0 || device.Slug == "" {
		return device.ModuleBays
	}
	if tmpl, ok := GetBySlug(device.Slug); ok {
		return tmpl.ModuleBays
	}
	return nil
}
