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
package imprt

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// remapDeviceParents rewrites device parent fields after location and rack
// merges retain existing inventory UUIDs.
func remapDeviceParents(
	devices map[uuid.UUID]*devicetypes.CaniDeviceType,
	locationRemap, rackRemap map[uuid.UUID]uuid.UUID,
) {
	for _, device := range devices {
		if device == nil {
			continue
		}
		if device.Parent != uuid.Nil {
			if mapped, ok := rackRemap[device.Parent]; ok {
				device.Parent = mapped
			} else if mapped, ok := locationRemap[device.Parent]; ok {
				device.Parent = mapped
			}
		}
		if mapped, ok := rackRemap[device.Rack]; ok {
			device.Rack = mapped
		}
	}
}

// remapDeviceReferences rewrites every foreign key that targets a device.
func remapDeviceReferences(
	inventory *devicetypes.Inventory,
	result *devicetypes.TransformResult,
	deviceRemap map[uuid.UUID]uuid.UUID,
) {
	if len(deviceRemap) == 0 {
		return
	}
	remapMergedDevices(inventory, result.Devices, deviceRemap)
	remapModuleDevices(result.Modules, deviceRemap)
	remapFruDevices(result.Frus, deviceRemap)
	remapCableDevices(result.Cables, deviceRemap)
}

func remapMergedDevices(
	inventory *devicetypes.Inventory,
	devices map[uuid.UUID]*devicetypes.CaniDeviceType,
	deviceRemap map[uuid.UUID]uuid.UUID,
) {
	for incomingID, device := range devices {
		if device == nil {
			continue
		}
		if mapped, ok := deviceRemap[device.Parent]; ok {
			device.Parent = mapped
		}
		resolvedID, ok := deviceRemap[incomingID]
		if !ok || resolvedID == incomingID {
			continue
		}
		resolved := inventory.Devices[resolvedID]
		if resolved == nil {
			continue
		}
		if device.Parent != uuid.Nil {
			resolved.Parent = device.Parent
		}
		if device.Rack != uuid.Nil {
			resolved.Rack = device.Rack
		}
	}
}

func remapModuleDevices(modules map[uuid.UUID]*devicetypes.CaniModuleType, deviceRemap map[uuid.UUID]uuid.UUID) {
	for _, module := range modules {
		if module != nil {
			if mapped, ok := deviceRemap[module.ParentDevice]; ok {
				module.ParentDevice = mapped
			}
		}
	}
}

func remapFruDevices(frus map[uuid.UUID]*devicetypes.CaniFruType, deviceRemap map[uuid.UUID]uuid.UUID) {
	for _, fru := range frus {
		if fru != nil {
			if mapped, ok := deviceRemap[fru.Device]; ok {
				fru.Device = mapped
			}
		}
	}
}

func remapCableDevices(cables map[uuid.UUID]*devicetypes.CaniCableType, deviceRemap map[uuid.UUID]uuid.UUID) {
	for _, cable := range cables {
		if cable == nil {
			continue
		}
		if mapped, ok := deviceRemap[cable.TerminationADevice]; ok {
			cable.TerminationADevice = mapped
		}
		if mapped, ok := deviceRemap[cable.TerminationBDevice]; ok {
			cable.TerminationBDevice = mapped
		}
	}
}

// remapIPAMLocations rewrites imported IPAM location references after
// locations are merged by name and retain their existing inventory UUIDs.
func remapIPAMLocations(result *devicetypes.TransformResult, locationRemap map[uuid.UUID]uuid.UUID) {
	for _, vlan := range result.VLANs {
		if vlan != nil {
			if mapped, ok := locationRemap[vlan.Location]; ok {
				vlan.Location = mapped
			}
		}
	}
	for _, prefix := range result.Prefixes {
		if prefix != nil {
			if mapped, ok := locationRemap[prefix.Location]; ok {
				prefix.Location = mapped
			}
		}
	}
}

// remapPrefixVLANs rewrites prefix VLAN references using the UUIDs retained by
// VLAN natural-key merges.
func remapPrefixVLANs(prefixes map[uuid.UUID]*devicetypes.CaniPrefix, vlanRemap map[uuid.UUID]uuid.UUID) {
	for _, prefix := range prefixes {
		if prefix != nil {
			if mapped, ok := vlanRemap[prefix.VLAN]; ok {
				prefix.VLAN = mapped
			}
		}
	}
}
