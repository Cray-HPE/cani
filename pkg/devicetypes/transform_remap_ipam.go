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
	"slices"

	"github.com/google/uuid"
)

// remapIPAM rewrites the references IPAM records hold to locations, VLANs,
// prefixes, VRFs, addresses and interfaces.
func (result *TransformResult) remapIPAM(remaps ReferenceRemaps) {
	for _, vlan := range result.VLANs {
		if vlan != nil {
			vlan.Location = remapUUID(vlan.Location, remaps.Locations)
		}
	}
	for _, prefix := range result.Prefixes {
		if prefix == nil {
			continue
		}
		prefix.Location = remapUUID(prefix.Location, remaps.Locations)
		prefix.VLAN = remapUUID(prefix.VLAN, remaps.VLANs)
		prefix.Parent = remapUUID(prefix.Parent, remaps.Prefixes)
		prefix.VRFs = remapUUIDs(prefix.VRFs, remaps.VRFs)
	}
	for _, address := range result.IPAddresses {
		if address == nil {
			continue
		}
		address.Parent = remapUUID(address.Parent, remaps.Prefixes)
		address.NATInside = remapUUID(address.NATInside, remaps.IPAddresses)
		address.Interfaces = remapUUIDs(address.Interfaces, remaps.Interfaces)
	}
}

// InterfaceRemapping is what InterfaceRemap found for the ports of devices
// and modules merged onto existing records.
type InterfaceRemapping struct {
	Remap    map[uuid.UUID]uuid.UUID // incoming spec ID → the same-named spec ID the merged owner kept
	Orphaned map[uuid.UUID]string    // incoming spec ID → the port and owner it named, for a port the kept record lacks
}

// InterfaceRemap maps each incoming interface spec ID to the ID of the
// same-named spec on the inventory record its owner resolved to. A device or
// module merged onto an existing record keeps that record's specs, so an
// address or cable on an incoming port has to follow the port to the ID the
// inventory kept. An incoming port the kept record has no port of that name
// for is orphaned instead, so the merge can drop what refers to it and say
// so rather than refuse the whole import. Owners the merge inserted as new
// keep their own specs and need no entry.
func (result *TransformResult) InterfaceRemap(inventory *Inventory, remaps ReferenceRemaps) InterfaceRemapping {
	mapping := InterfaceRemapping{Remap: make(map[uuid.UUID]uuid.UUID), Orphaned: make(map[uuid.UUID]string)}
	if result == nil || inventory == nil {
		return mapping
	}
	for incomingID, device := range result.Devices {
		if kept := keptOwner(inventory.Devices, remaps.Devices, incomingID); device != nil && kept != nil && kept != device {
			mapping.follow(device.Interfaces, kept.Interfaces, fmt.Sprintf("device %q", kept.Name))
		}
	}
	result.followModulePorts(inventory, remaps.Modules, mapping)
	return mapping
}

// followModulePorts is InterfaceRemap's walk over incoming modules.
func (result *TransformResult) followModulePorts(inventory *Inventory, remap map[uuid.UUID]uuid.UUID, mapping InterfaceRemapping) {
	for incomingID, module := range result.Modules {
		if kept := keptOwner(inventory.Modules, remap, incomingID); module != nil && kept != nil && kept != module {
			mapping.follow(module.Interfaces, kept.Interfaces, fmt.Sprintf("module %q", kept.Label()))
		}
	}
}

// keptOwner returns the record an incoming owner resolved to, or nil when
// the merge did not resolve it.
func keptOwner[Owner any](owners map[uuid.UUID]*Owner, remap map[uuid.UUID]uuid.UUID, incomingID uuid.UUID) *Owner {
	if resolvedID, ok := remap[incomingID]; ok {
		return owners[resolvedID]
	}
	return nil
}

// follow records, for each incoming spec, the ID of the kept spec with the
// same name when the two differ, and orphans an incoming spec whose name the
// kept record lacks. The first kept spec of a name wins, as it does for
// FindInterfaceIDByPort.
func (mapping InterfaceRemapping) follow(incoming, kept []InterfaceSpec, owner string) {
	byName := make(map[string]uuid.UUID, len(kept))
	for i := range kept {
		if _, seen := byName[kept[i].Name]; !seen {
			byName[kept[i].Name] = kept[i].ID
		}
	}
	for i := range incoming {
		keptID, ok := byName[incoming[i].Name]
		switch {
		case !ok:
			mapping.Orphaned[incoming[i].ID] = fmt.Sprintf("port %q on %s", incoming[i].Name, owner)
		case keptID != incoming[i].ID:
			mapping.Remap[incoming[i].ID] = keptID
		}
	}
}

// dropOrphanedAssignments takes the ports InterfaceRemap orphaned off every
// incoming address and names each assignment it drops, in a stable order.
// The record the owner merged onto has no port of that name, so the
// assignment has nothing to point at; dropping it lets the rest of the
// import merge where keeping it would fail the relationship validation and
// refuse everything. The source still holds the assignment.
func (result *TransformResult) dropOrphanedAssignments(orphaned map[uuid.UUID]string) []string {
	var dropped []string
	if len(orphaned) == 0 {
		return dropped
	}
	for _, address := range result.IPAddresses {
		if address == nil {
			continue
		}
		address.Interfaces = slices.DeleteFunc(address.Interfaces, func(id uuid.UUID) bool {
			port, ok := orphaned[id]
			if ok {
				dropped = append(dropped, fmt.Sprintf(
					"IP address %q: assignment to %s dropped; the inventory's record has no port of that name", address.Address, port))
			}
			return ok
		})
	}
	slices.Sort(dropped)
	return dropped
}
