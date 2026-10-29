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
package export

import (
	"context"
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// cableEnd is one termination of a cable: the inventory device that owns the
// port, directly or through a module, and once looked up its Nautobot
// device ID and interface. module is the inventory module that owns the
// port, nil for a device-owned port; moduleID is the Nautobot ID this run
// created or found for that module, else Nil.
type cableEnd struct {
	device   *devicetypes.CaniDeviceType
	module   *devicetypes.CaniModuleType
	port     string
	moduleID uuid.UUID
	deviceID uuid.UUID
	iface    *CachedItem
}

// createCaniCableType creates a cable between two interfaces in Nautobot,
// unless Nautobot already has it or either interface already carries one.
func (e *Exporter) createCaniCableType(ctx context.Context, cableID uuid.UUID, cable *devicetypes.CaniCableType, inventory *devicetypes.Inventory, deviceIDMap map[string]uuid.UUID, result *LoadResult) error {
	if cable.TerminationADevice == uuid.Nil || cable.TerminationBDevice == uuid.Nil {
		clog.Detail("INFO: Cable %s has incomplete terminations (status=%s), skipping", cable.Label, cable.Status)
		return nil
	}
	a, b, err := cableEnds(inventory, cable)
	if err != nil {
		return err
	}
	e.resolveEndModules(&a, &b)
	if e.previewPlannedCable(cable, a, b, result) {
		return nil
	}
	if err := e.resolveCableEnds(&a, &b, deviceIDMap); err != nil {
		return err
	}
	if e.skipPresentCable(cable, a, b, result) {
		return nil
	}
	if e.Options.DryRun {
		clog.DryRun("Would create cable: %s (%s)", cable.Label, cableEndpoints(a, b))
		result.CablesCreated++
		return nil
	}
	statusID, err := e.statusRef(cableStatusName(cable))
	if err != nil {
		return err
	}
	req, err := cableRequest(cable, a, b, statusID)
	if err != nil {
		return err
	}
	return e.postCable(ctx, cable, a, b, req, result)
}

// cableEnds resolves both terminations to inventory devices and owning
// modules; a termination on a module belongs to the module's parent device.
func cableEnds(inventory *devicetypes.Inventory, cable *devicetypes.CaniCableType) (a, b cableEnd, err error) {
	a = cableEnd{device: terminationDevice(inventory, cable.TerminationADevice), port: cable.TerminationAPort,
		module: owningModule(inventory, cable.TerminationADevice, cable.TerminationA)}
	b = cableEnd{device: terminationDevice(inventory, cable.TerminationBDevice), port: cable.TerminationBPort,
		module: owningModule(inventory, cable.TerminationBDevice, cable.TerminationB)}
	if a.device == nil || b.device == nil {
		return a, b, fmt.Errorf("cable endpoints reference unknown devices (A=%s, B=%s)", cable.TerminationADevice, cable.TerminationBDevice)
	}
	return a, b, nil
}

// owningModule returns the inventory module that owns a termination: the
// module ownerID names directly, or otherwise the module the resolved
// interface records as its owner, which covers a cable that addresses the
// parent device by one of its modules' port names. nil when the port is
// device-owned.
func owningModule(inventory *devicetypes.Inventory, ownerID, ifaceID uuid.UUID) *devicetypes.CaniModuleType {
	if module := inventory.Modules[ownerID]; module != nil {
		return module
	}
	if iface := inventory.Interfaces[ifaceID]; iface != nil && iface.ModuleID != uuid.Nil {
		return inventory.Modules[iface.ModuleID]
	}
	return nil
}

// resolveEndModules sets each module-owned end's Nautobot module ID to the
// one this run created or found. A module the run only planned or failed to
// export leaves Nil, so no lookup is attempted against an ID Nautobot may
// not have.
func (e *Exporter) resolveEndModules(a, b *cableEnd) {
	for _, end := range []*cableEnd{a, b} {
		if end.module != nil {
			end.moduleID = e.Cache.ExportedModuleID(end.module.ID)
		}
	}
}

// exportedModuleID returns the Nautobot ID this run confirmed for the module
// that owns a termination, or uuid.Nil for a device-owned port or a module
// the run did not create or find.
func (e *Exporter) exportedModuleID(inventory *devicetypes.Inventory, ownerID, ifaceID uuid.UUID) uuid.UUID {
	module := owningModule(inventory, ownerID, ifaceID)
	if module == nil {
		return uuid.Nil
	}
	return e.Cache.ExportedModuleID(module.ID)
}

func terminationDevice(inventory *devicetypes.Inventory, id uuid.UUID) *devicetypes.CaniDeviceType {
	if device := inventory.Devices[id]; device != nil {
		return device
	}
	if module := inventory.Modules[id]; module != nil {
		return inventory.Devices[module.ParentDevice]
	}
	return nil
}

// resolveCableEnds finds both ends' Nautobot devices and then both
// interfaces, in that order, so a missing device is reported before a
// missing interface on the other end.
func (e *Exporter) resolveCableEnds(a, b *cableEnd, deviceIDMap map[string]uuid.UUID) error {
	for _, end := range []*cableEnd{a, b} {
		if err := e.resolveEndDevice(end, deviceIDMap); err != nil {
			return err
		}
	}
	for _, end := range []*cableEnd{a, b} {
		if err := e.resolveEndInterface(end); err != nil {
			return err
		}
	}
	return nil
}

// resolveEndDevice takes the Nautobot device ID this run created for the
// end's device, or looks the device up by name when an earlier run created it.
func (e *Exporter) resolveEndDevice(end *cableEnd, deviceIDMap map[string]uuid.UUID) error {
	if id, ok := deviceIDMap[end.device.Name]; ok {
		end.deviceID = id
		return nil
	}
	device, err := e.Cache.GetDeviceByName(end.device.Name)
	if err != nil || device == nil {
		return fmt.Errorf("cannot find Nautobot device ID for %s", end.device.Name)
	}
	end.deviceID = device.ID
	return nil
}

// resolveEndInterface finds the end's interface through
// resolveOwnedInterface and reports a miss with the end's device name.
func (e *Exporter) resolveEndInterface(end *cableEnd) error {
	iface, err := e.resolveOwnedInterface(end.moduleID, end.deviceID, end.port)
	if err != nil || iface == nil {
		return fmt.Errorf(errFmtFindInterface, end.port, end.device.Name, err)
	}
	end.iface = iface
	return nil
}

// resolveOwnedInterface finds a Nautobot interface by its owner. A
// module-owned port (moduleID set) is looked up only among its module's ports,
// by exact and then fuzzy name, so it matches neither a sibling module's port
// nor a same-named port of the parent device. Any other port is looked up
// among the device's own ports by fuzzy name, which absorbs naming
// differences between the inventory and Nautobot; a miss drops the device's
// interface prefetch and retries once. Returns nil, nil when the interface
// does not exist.
func (e *Exporter) resolveOwnedInterface(moduleID, deviceID uuid.UUID, port string) (*CachedItem, error) {
	if moduleID != uuid.Nil {
		return e.Cache.GetInterfaceByModuleAndNameFuzzy(moduleID, port)
	}
	iface, err := e.Cache.GetInterfaceByDeviceAndNameFuzzy(deviceID, port)
	if err != nil || iface != nil {
		return iface, err
	}
	e.Cache.InvalidateInterfacePrefetch(deviceID)
	return e.Cache.GetInterfaceByDeviceAndNameFuzzy(deviceID, port)
}

// skipPresentCable reports a cable Nautobot already has between the two
// interfaces as skipped, and one whose interface already carries another
// cable as conflicted; the latter usually means the inventory assigns two
// cables to one physical port.
func (e *Exporter) skipPresentCable(cable *devicetypes.CaniCableType, a, b cableEnd, result *LoadResult) bool {
	existing, err := e.Cache.GetCableByTerminations(a.iface.ID, b.iface.ID)
	if err != nil {
		clog.Warn("Warning: failed to check for existing cable between %s:%s and %s:%s: %v",
			a.device.Name, a.port, b.device.Name, b.port, err)
	}
	if existing != nil {
		clog.Skipped("Skipped cable (already exists): %s (%s)", cable.Label, cableEndpoints(a, b))
		result.CablesSkipped++
		return true
	}
	for _, end := range []cableEnd{a, b} {
		if end.iface.CableID != uuid.Nil {
			clog.Skipped("Skipped cable (interface already cabled): %s:%s", end.device.Name, end.port)
			result.CablesConflicted++
			return true
		}
	}
	return false
}

// cableEndpoints renders both ends as "device:port <-> device:port".
func cableEndpoints(a, b cableEnd) string {
	return fmt.Sprintf("%s:%s <-> %s:%s", a.device.Name, a.port, b.device.Name, b.port)
}
