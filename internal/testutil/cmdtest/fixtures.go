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

package cmdtest

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// DeviceSlug is a device type present in the embedded hardware library. Tests
// use a real slug so template expansion (interfaces, bays, power ports) runs
// the same way it does for users.
const DeviceSlug = "cray-gigabyte-r272-z30-00"

// DeviceSlugUHeight is the u_height DeviceSlug declares in the library. Seeded
// devices must match it or rack occupancy in tests would not reflect reality.
const DeviceSlugUHeight = 2

// InventoryWithRack returns an inventory holding one location and one empty
// 42U rack with the given name, which is the minimum needed to place a device.
func InventoryWithRack(rackName string) (*devicetypes.Inventory, uuid.UUID) {
	inventory := devicetypes.NewInventory()
	locationID := inventory.EnsureLocation()

	rackID := uuid.New()
	inventory.Racks[rackID] = &devicetypes.CaniRackType{
		ID:       rackID,
		Name:     rackName,
		Slug:     "hpe-eia-cabinet",
		UHeight:  42,
		Location: locationID,
		Type:     devicetypes.TypeRack,
	}

	return inventory, rackID
}

// AddDevice puts a device into the inventory at the given rack position and
// returns its ID, for tests that need an existing device to update or remove.
//
// UHeight matches the library entry for DeviceSlug so seeded devices occupy the
// same number of rack units as one added through the real command path.
func AddDevice(inventory *devicetypes.Inventory, rackID uuid.UUID, name string, position int) uuid.UUID {
	deviceID := uuid.New()
	inventory.Devices[deviceID] = &devicetypes.CaniDeviceType{
		ID:           deviceID,
		Name:         name,
		Slug:         DeviceSlug,
		Parent:       rackID,
		RackPosition: position,
		Face:         devicetypes.FaceFront,
		UHeight:      DeviceSlugUHeight,
		Type:         devicetypes.TypeNode,
	}
	return deviceID
}

// AddChildDevice nests a device under another device. Parent is the single
// forward FK for both rack and device parentage, so this is what drives the
// Children reverse index that remove cascades over.
func AddChildDevice(inventory *devicetypes.Inventory, parentDeviceID uuid.UUID, name string) uuid.UUID {
	childID := uuid.New()
	inventory.Devices[childID] = &devicetypes.CaniDeviceType{
		ID:      childID,
		Name:    name,
		Slug:    DeviceSlug,
		Parent:  parentDeviceID,
		UHeight: 1,
		Type:    devicetypes.TypeNode,
	}
	return childID
}
