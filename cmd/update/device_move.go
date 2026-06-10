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
	"log"

	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// resolveParent tries to resolve a string as a rack UUID/name first,
// then as a device UUID/name. Returns the resolved UUID.
func resolveParent(inv *devicetypes.Inventory, ref string) (uuid.UUID, error) {
	if id, err := resolve.Rack(inv, ref); err == nil {
		return id, nil
	}
	if id, err := resolve.Device(inv, ref); err == nil {
		return id, nil
	}
	return uuid.Nil, fmt.Errorf("%q not found as rack or device", ref)
}

// findDeviceRack walks up from device.Parent (and ancestor devices)
// to find the containing rack. Returns nil if the device has no rack.
func findDeviceRack(inv *devicetypes.Inventory, device *devicetypes.CaniDeviceType) *devicetypes.CaniRackType {
	cur := device.Parent
	for i := 0; i < 10; i++ {
		if rack, ok := inv.Racks[cur]; ok {
			return rack
		}
		parent, ok := inv.Devices[cur]
		if !ok || parent == nil {
			return nil
		}
		cur = parent.Parent
	}
	return nil
}

// moveOrSwap places a device at newPos/newFace. If the target slot is occupied
// and swap is true, it atomically swaps the two devices' positions. If the slot
// is occupied and swap is false, it returns an error suggesting --swap.
func moveOrSwap(
	rack *devicetypes.CaniRackType,
	inv *devicetypes.Inventory,
	id uuid.UUID,
	device *devicetypes.CaniDeviceType,
	newPos int,
	newFace string,
	swap bool,
) error {
	if newFace == "" {
		newFace = devicetypes.FaceFront
	}
	occupant := rack.GetSlotOccupant(newPos, newFace)
	if occupant != uuid.Nil && occupant != id {
		if !swap {
			name := occupant.String()
			if d, ok := inv.Devices[occupant]; ok && d.Name != "" {
				name = d.Name
			}
			return fmt.Errorf("cannot place device at U%d (occupied by %s); use --swap to swap positions", newPos, name)
		}
		return doSwapDevices(rack, inv, id, device, occupant, newPos, newFace)
	}
	// Target is free — simple move.
	rack.RemoveDevice(id)
	height := device.UHeight
	if height < 1 {
		height = 1
	}
	if !rack.PlaceDevice(id, newPos, height, newFace, device.IsFullDepth) {
		return fmt.Errorf("cannot place device at U%d (slot occupied or out of bounds)", newPos)
	}
	device.RackPosition = newPos
	device.Face = newFace
	return nil
}

// doSwapDevices performs the atomic position swap between two devices.
func doSwapDevices(
	rack *devicetypes.CaniRackType,
	inv *devicetypes.Inventory,
	id uuid.UUID,
	device *devicetypes.CaniDeviceType,
	occupantID uuid.UUID,
	newPos int,
	newFace string,
) error {
	occupantDev := inv.Devices[occupantID]
	if occupantDev == nil {
		return fmt.Errorf("occupant device %s not found in inventory", occupantID)
	}
	oldPos := device.RackPosition
	oldFace := device.Face
	if oldFace == "" {
		oldFace = devicetypes.FaceFront
	}
	if err := rack.SwapDevices(id, occupantID); err != nil {
		return fmt.Errorf("swap failed: %w", err)
	}
	device.RackPosition = newPos
	device.Face = newFace
	occupantDev.RackPosition = oldPos
	occupantDev.Face = oldFace
	log.Printf("Swapped positions: %s (%s) → U%d, %s (%s) → U%d",
		id, device.Name, newPos, occupantID, occupantDev.Name, oldPos)
	return nil
}

// resolveIPAddress resolves a CIDR string or UUID to an IP address UUID in the inventory.
func resolveIPAddress(inv *devicetypes.Inventory, ref string) (uuid.UUID, error) {
	// Try as UUID first
	if id, err := uuid.Parse(ref); err == nil {
		if _, ok := inv.IPAddresses[id]; ok {
			return id, nil
		}
	}
	// Try as CIDR match
	for _, addr := range inv.IPAddresses {
		if addr.Address == ref {
			return addr.ID, nil
		}
	}
	return uuid.Nil, fmt.Errorf("ip address %q not found in inventory", ref)
}
