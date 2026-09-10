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

import "github.com/google/uuid"

// ReferenceRemaps records incoming UUIDs and the inventory UUIDs retained by
// natural-key merges. TransformResult owns the reference graph that consumes
// these maps so command orchestration cannot omit a model relationship.
type ReferenceRemaps struct {
	Locations   map[uuid.UUID]uuid.UUID
	Racks       map[uuid.UUID]uuid.UUID
	Devices     map[uuid.UUID]uuid.UUID
	Modules     map[uuid.UUID]uuid.UUID
	Frus        map[uuid.UUID]uuid.UUID
	Cables      map[uuid.UUID]uuid.UUID
	VLANs       map[uuid.UUID]uuid.UUID
	Prefixes    map[uuid.UUID]uuid.UUID
	IPAddresses map[uuid.UUID]uuid.UUID
	VRFs        map[uuid.UUID]uuid.UUID
}

// RemapReferences rewrites every persisted UUID relationship represented by a
// TransformResult. It may be called after each merge stage as more remaps become
// available; applying the same remap more than once is harmless.
func (result *TransformResult) RemapReferences(inventory *Inventory, remaps ReferenceRemaps) {
	if result == nil {
		return
	}
	result.remapLocations(inventory, remaps)
	result.remapRacks(remaps)
	result.remapDevices(inventory, remaps)
	result.remapModules(inventory, remaps)
	result.remapFrus(inventory, remaps)
	result.remapCables(remaps)
	result.remapIPAM(remaps)
}

// DerivePrefixParents computes the most-specific parent for each transformed
// prefix across both the existing inventory and the current transform result.
func (result *TransformResult) DerivePrefixParents(existing map[uuid.UUID]*CaniPrefix) {
	if result == nil {
		return
	}
	candidates := make(map[uuid.UUID]*CaniPrefix, len(existing)+len(result.Prefixes))
	for id, prefix := range existing {
		candidates[id] = prefix
	}
	for id, prefix := range result.Prefixes {
		candidates[id] = prefix
	}
	for _, prefix := range result.Prefixes {
		if prefix != nil {
			prefix.Parent = FindParentPrefix(prefix, candidates)
		}
	}
}

// DeriveIPAddressParents assigns each transformed address to the most-specific
// prefix retained in the inventory.
func (result *TransformResult) DeriveIPAddressParents(prefixes map[uuid.UUID]*CaniPrefix) {
	if result == nil {
		return
	}
	for _, address := range result.IPAddresses {
		if address != nil {
			address.Parent = FindParentPrefixForIP(address, prefixes)
		}
	}
}

func (result *TransformResult) remapLocations(inventory *Inventory, remaps ReferenceRemaps) {
	for incomingID, location := range result.Locations {
		if location == nil {
			continue
		}
		location.Parent = remapUUID(location.Parent, remaps.Locations)
		if resolvedID, ok := remaps.Locations[incomingID]; ok {
			location.ID = resolvedID
			if inventory != nil && resolvedID != incomingID {
				if resolved := inventory.Locations[resolvedID]; resolved != nil && location.Parent != uuid.Nil {
					resolved.Parent = location.Parent
				}
			}
		}
	}
}

func (result *TransformResult) remapRacks(remaps ReferenceRemaps) {
	for _, rack := range result.Racks {
		if rack != nil {
			rack.Location = remapUUID(rack.Location, remaps.Locations)
		}
	}
}

func (result *TransformResult) remapDevices(inventory *Inventory, remaps ReferenceRemaps) {
	for incomingID, device := range result.Devices {
		if device == nil {
			continue
		}
		device.Parent = remapUUID(device.Parent, remaps.Devices, remaps.Racks, remaps.Locations)
		device.Rack = remapUUID(device.Rack, remaps.Racks)
		device.Location = remapUUID(device.Location, remaps.Locations)
		device.ParentDevice = remapUUID(device.ParentDevice, remaps.Devices)
		device.BMCParent = remapUUID(device.BMCParent, remaps.Devices)
		device.AssignedVLANs = remapUUIDs(device.AssignedVLANs, remaps.VLANs)
		device.PrimaryIPv4 = remapUUID(device.PrimaryIPv4, remaps.IPAddresses)
		device.PrimaryIPv6 = remapUUID(device.PrimaryIPv6, remaps.IPAddresses)
		device.Frus = remapUUIDs(device.Frus, remaps.Frus)
		remapInterfaceCables(device.Interfaces, remaps.Cables)
		result.syncResolvedDevice(inventory, incomingID, device, remaps.Devices)
	}
	for _, vrf := range result.VRFs {
		if vrf != nil {
			vrf.Devices = remapUUIDs(vrf.Devices, remaps.Devices)
		}
	}
}

func (result *TransformResult) syncResolvedDevice(
	inventory *Inventory,
	incomingID uuid.UUID,
	incoming *CaniDeviceType,
	deviceRemap map[uuid.UUID]uuid.UUID,
) {
	if inventory == nil {
		return
	}
	resolvedID, ok := deviceRemap[incomingID]
	if !ok || resolvedID == incomingID {
		return
	}
	resolved := inventory.Devices[resolvedID]
	if resolved == nil {
		return
	}
	if incoming.Parent != uuid.Nil {
		resolved.Parent = incoming.Parent
	}
	if incoming.Rack != uuid.Nil {
		resolved.Rack = incoming.Rack
	}
	if incoming.Location != uuid.Nil {
		resolved.Location = incoming.Location
	}
	if incoming.ParentDevice != uuid.Nil {
		resolved.ParentDevice = incoming.ParentDevice
	}
	if incoming.BMCParent != uuid.Nil {
		resolved.BMCParent = incoming.BMCParent
	}
	if len(incoming.AssignedVLANs) > 0 {
		resolved.AssignedVLANs = incoming.AssignedVLANs
	}
	if incoming.PrimaryIPv4 != uuid.Nil {
		resolved.PrimaryIPv4 = incoming.PrimaryIPv4
	}
	if incoming.PrimaryIPv6 != uuid.Nil {
		resolved.PrimaryIPv6 = incoming.PrimaryIPv6
	}
}

func (result *TransformResult) remapModules(inventory *Inventory, remaps ReferenceRemaps) {
	for incomingID, module := range result.Modules {
		if module == nil {
			continue
		}
		module.ParentDevice = remapUUID(module.ParentDevice, remaps.Devices)
		module.Location = remapUUID(module.Location, remaps.Locations)
		module.Frus = remapUUIDs(module.Frus, remaps.Frus)
		remapInterfaceCables(module.Interfaces, remaps.Cables)
		if resolvedID, ok := remaps.Modules[incomingID]; ok {
			module.ID = resolvedID
			if inventory != nil && resolvedID != incomingID {
				if resolved := inventory.Modules[resolvedID]; resolved != nil {
					if module.ParentDevice != uuid.Nil {
						resolved.ParentDevice = module.ParentDevice
					}
					if module.Location != uuid.Nil {
						resolved.Location = module.Location
					}
				}
			}
		}
	}
}

func (result *TransformResult) remapFrus(inventory *Inventory, remaps ReferenceRemaps) {
	for incomingID, fru := range result.Frus {
		if fru == nil {
			continue
		}
		fru.Device = remapUUID(fru.Device, remaps.Devices, remaps.Modules)
		fru.Parent = remapUUID(fru.Parent, remaps.Frus)
		if resolvedID, ok := remaps.Frus[incomingID]; ok {
			fru.ID = resolvedID
			if inventory != nil && resolvedID != incomingID {
				if resolved := inventory.Frus[resolvedID]; resolved != nil {
					if fru.Device != uuid.Nil {
						resolved.Device = fru.Device
					}
					if fru.Parent != uuid.Nil {
						resolved.Parent = fru.Parent
					}
				}
			}
		}
	}
}

func (result *TransformResult) remapCables(remaps ReferenceRemaps) {
	for _, cable := range result.Cables {
		if cable == nil {
			continue
		}
		cable.TerminationADevice = remapUUID(cable.TerminationADevice, remaps.Devices, remaps.Modules)
		cable.TerminationBDevice = remapUUID(cable.TerminationBDevice, remaps.Devices, remaps.Modules)
	}
}

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
	}
	for _, address := range result.IPAddresses {
		if address == nil {
			continue
		}
		address.Parent = remapUUID(address.Parent, remaps.Prefixes)
		address.NATInside = remapUUID(address.NATInside, remaps.IPAddresses)
	}
}

func remapUUID(id uuid.UUID, remaps ...map[uuid.UUID]uuid.UUID) uuid.UUID {
	if id == uuid.Nil {
		return id
	}
	for _, remap := range remaps {
		if mapped, ok := remap[id]; ok {
			return mapped
		}
	}
	return id
}

func remapUUIDs(ids []uuid.UUID, remap map[uuid.UUID]uuid.UUID) []uuid.UUID {
	for index, id := range ids {
		ids[index] = remapUUID(id, remap)
	}
	return ids
}

func remapInterfaceCables(interfaces []InterfaceSpec, cableRemap map[uuid.UUID]uuid.UUID) {
	for index := range interfaces {
		if interfaces[index].ConnectedCable == nil {
			continue
		}
		mapped := remapUUID(*interfaces[index].ConnectedCable, cableRemap)
		interfaces[index].ConnectedCable = &mapped
	}
}
