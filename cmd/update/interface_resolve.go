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
package update

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// resolveInterfaces finds the target interface(s) based on either:
//   - A positional UUID argument
//   - --device + --name flags (name supports glob patterns)
func resolveInterfaces(cmd *cobra.Command, args []string, inv *devicetypes.Inventory) ([]interfaceTarget, error) {
	// Case 1: Positional UUID argument
	if len(args) == 1 {
		id, err := uuid.Parse(args[0])
		if err != nil {
			return nil, fmt.Errorf("invalid interface UUID: %w", err)
		}
		iface, ok := inv.Interfaces[id]
		if !ok {
			return nil, fmt.Errorf("interface %s not found", id)
		}
		spec := findInterfaceSpec(inv, iface)
		return []interfaceTarget{{instance: iface, spec: spec}}, nil
	}

	// Case 2: --device/--module + --name
	deviceRef, _ := cmd.Flags().GetString("device")
	moduleRef, _ := cmd.Flags().GetString("module")
	namePattern, _ := cmd.Flags().GetString("name")

	if deviceRef == "" && moduleRef == "" {
		return nil, fmt.Errorf("either a positional UUID, --device, or --module flag is required")
	}
	if deviceRef != "" && moduleRef != "" {
		return nil, fmt.Errorf("--device and --module are mutually exclusive")
	}
	if namePattern == "" {
		return nil, fmt.Errorf("--name is required when using --device or --module")
	}

	// --module targets only the named module's own interfaces, which
	// disambiguates interface names shared with the parent device or
	// sibling modules (e.g. multiple "HSN 0" ports on one node).
	if moduleRef != "" {
		moduleID, err := resolve.Module(inv, moduleRef)
		if err != nil {
			return nil, fmt.Errorf("resolving --module: %w", err)
		}
		return collectModuleInterfaceTargets(inv, moduleID, namePattern)
	}

	deviceID, err := resolve.Device(inv, deviceRef)
	if err != nil {
		return nil, fmt.Errorf("resolving --device: %w", err)
	}
	return collectInterfaceTargets(inv, namePattern, func(iface *devicetypes.CaniInterface) bool {
		return belongsToDevice(inv, iface, deviceID)
	})
}

// collectModuleInterfaceTargets returns targets for the named module's own
// interfaces. Module interfaces are flattened into inv.Interfaces under the
// parent device's ID, so they cannot be distinguished by DeviceID; instead we
// walk the module's interface specs (which carry the canonical ID) and pair
// each with its top-level instance looked up by that ID.
func collectModuleInterfaceTargets(inv *devicetypes.Inventory, moduleID uuid.UUID, namePattern string) ([]interfaceTarget, error) {
	mod := inv.Modules[moduleID]
	if mod == nil {
		return nil, fmt.Errorf("module %s not found", moduleID)
	}
	var targets []interfaceTarget
	for i := range mod.Interfaces {
		spec := &mod.Interfaces[i]
		matched, merr := matchInterfaceName(namePattern, spec.Name)
		if merr != nil {
			return nil, fmt.Errorf("invalid --name pattern %q: %w", namePattern, merr)
		}
		if !matched && !strings.EqualFold(spec.Name, namePattern) {
			continue
		}
		instance := inv.Interfaces[spec.ID]
		if instance == nil {
			continue
		}
		targets = append(targets, interfaceTarget{instance: instance, spec: spec})
	}
	return targets, nil
}

// collectInterfaceTargets returns interface targets whose name matches the
// given pattern and that satisfy the supplied ownership predicate.
func collectInterfaceTargets(inv *devicetypes.Inventory, namePattern string, owns func(*devicetypes.CaniInterface) bool) ([]interfaceTarget, error) {
	var targets []interfaceTarget
	for _, iface := range inv.Interfaces {
		if iface == nil || !owns(iface) {
			continue
		}
		// Use matchInterfaceName so that '*' can match '/' characters
		// in interface names like "1/1/14".
		matched, merr := matchInterfaceName(namePattern, iface.Name)
		if merr != nil {
			return nil, fmt.Errorf("invalid --name pattern %q: %w", namePattern, merr)
		}
		if !matched && !strings.EqualFold(iface.Name, namePattern) {
			continue
		}
		spec := findInterfaceSpec(inv, iface)
		targets = append(targets, interfaceTarget{instance: iface, spec: spec})
	}
	return targets, nil
}

// belongsToDevice returns true if the interface belongs to the given device
// directly or via one of its child modules.
func belongsToDevice(inv *devicetypes.Inventory, iface *devicetypes.CaniInterface, deviceID uuid.UUID) bool {
	if iface.DeviceID == deviceID {
		return true
	}
	// Check if the interface's device is actually a module parented by this device
	for _, mod := range inv.Modules {
		if mod == nil {
			continue
		}
		if mod.ID == iface.DeviceID && mod.ParentDevice == deviceID {
			return true
		}
	}
	return false
}

// matchInterfaceName performs glob matching where '*' can match '/' characters.
// Standard filepath.Match treats '/' as a path separator that '*' cannot cross;
// this function works around that for interface names like "1/1/14".
func matchInterfaceName(pattern, name string) (bool, error) {
	p := strings.ReplaceAll(pattern, "/", "\x00")
	n := strings.ReplaceAll(name, "/", "\x00")
	return filepath.Match(p, n)
}

// findInterfaceSpec locates the InterfaceSpec on the parent device/module
// that corresponds to the given CaniInterface.
func findInterfaceSpec(inv *devicetypes.Inventory, iface *devicetypes.CaniInterface) *devicetypes.InterfaceSpec {
	if device, ok := inv.Devices[iface.DeviceID]; ok && device != nil {
		if spec := findSpecInList(device.Interfaces, iface); spec != nil {
			return spec
		}
	}
	if mod, ok := inv.Modules[iface.DeviceID]; ok && mod != nil {
		if spec := findSpecInList(mod.Interfaces, iface); spec != nil {
			return spec
		}
	}
	return nil
}

// findSpecInList returns the spec in specs that matches iface by name or ID.
func findSpecInList(specs []devicetypes.InterfaceSpec, iface *devicetypes.CaniInterface) *devicetypes.InterfaceSpec {
	for i := range specs {
		if specs[i].Name == iface.Name {
			return &specs[i]
		}
		if specs[i].ID == iface.ID {
			return &specs[i]
		}
	}
	return nil
}
