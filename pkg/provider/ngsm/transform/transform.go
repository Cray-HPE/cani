/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
package transform

import (
	"log"
	"path/filepath"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/ngsm/extract"
	"github.com/google/uuid"
)

// Transform transforms devices in the queue into CANI's format
func Transform(existing devicetypes.Inventory, queues extract.Queues) (transformed map[uuid.UUID]*devicetypes.CaniDeviceType, err error) {
	transformed, err = transform(existing, queues)
	if err != nil {
		return nil, err
	}
	return transformed, nil
}

// extractDevicesFromBom extracts devices from the BOM files into CANI's inventory format
func transform(existing devicetypes.Inventory, queues extract.Queues) (transformed map[uuid.UUID]*devicetypes.CaniDeviceType, err error) {
	transformed = make(map[uuid.UUID]*devicetypes.CaniDeviceType)

	devices, err := transformDevicesFromQueue(existing, queues)
	if err != nil {
		return transformed, err
	}
	for _, device := range devices {
		d := device // create a copy of d for this iteration
		transformed[d.ID] = d
	}

	log.Println("")
	log.Printf("  %d devices existing in current inventory", len(existing.Devices)-len(existing.Systems()))
	log.Printf("  %d devices Transformed (not yet Loaded)", len(transformed))
	log.Println("")

	return transformed, nil
}

func transformDevicesFromQueue(existing devicetypes.Inventory, queues extract.Queues) (devices map[string]*devicetypes.CaniDeviceType, err error) {
	devices = make(map[string]*devicetypes.CaniDeviceType)
	racks := make([]*devicetypes.CaniDeviceType, 0)
	// First, check if the BOM has one or more racks since those need to exist to set the parent ID for devices
	for _, q := range queues {
		if len(q.RacksToCreate) > 0 {
			for _, row := range q.RacksToCreate {
				row.HardwareType = devicetypes.All()[row.DeviceTypeSlug].Type
				if filepath.Base(q.Bom) == row.Source {
					newRack := row.NewDeviceFromRow()
					systems := existing.Systems()
					newRack.Parent = systems[0].ID // TODO: logic for choosing a system other than 0
					racks = append(racks, &newRack)
					devices[row.NetboxName] = &newRack
				}
			}
		}
	}

	for _, q := range queues {
		if len(q.DevicesToCreate) > 0 {
			for _, row := range q.DevicesToCreate {
				row.HardwareType = devicetypes.All()[row.DeviceTypeSlug].Type
				if filepath.Base(q.Bom) == row.Source {
					newDevice := row.NewDeviceFromRow()
					deviceMeta := newDevice.ProviderMetadata["ngsm"].(map[string]any)
					if len(q.RacksToCreate) > 0 {
						rackMeta := racks[0].ProviderMetadata["ngsm"].(map[string]any)
						// if rack's source does not match the row source, do not add it as a parent to the device
						if deviceMeta["Source"] != rackMeta["Source"] {
							log.Printf("Parent will not be set for %s in %s, no racks detected.", row.NetboxName, row.Source)
							// newDevice.Parent = uuid.Nil // No parent if the rack source does not match
						} else {
							// Otherwise, set the parent to the first rack in the list
							newDevice.Parent = racks[0].ID // TODO: logic for choosing a rack other than 0
						}
					} else {
						log.Printf("No racks detected in %s, cannot set parent for %s", row.Source, row.NetboxName)
					}
					devices[row.NetboxName] = &newDevice
				}
			}
		}
	}
	return devices, nil
}
