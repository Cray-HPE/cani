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

	"github.com/google/uuid"
)

// A device's location follows three rules, which are Nautobot's:
//
//   - A device with no parent is placed directly in the location it records.
//   - A device mounted in a rack is in the rack's location unless it records
//     one that holds the rack's: the rack's own location or an ancestor of
//     it. Any other recorded location is replaced by the rack's on rebuild.
//   - A device installed in another device takes that device's location.
//
// A racked device stops recording a location of its own when its rack moves
// or it is given another parent, so it follows the rack again; Nautobot
// resets the devices in a rack the same way whenever the rack is saved.

// LocationHolds reports whether outer is inner or one of inner's ancestors.
func (inv *Inventory) LocationHolds(outer, inner uuid.UUID) bool {
	if outer == uuid.Nil {
		return false
	}
	// The hop limit ends the walk on a parent cycle, which validation reports.
	for hops := 0; inner != uuid.Nil && hops <= len(inv.Locations); hops++ {
		if inner == outer {
			return true
		}
		location := inv.Locations[inner]
		if location == nil {
			return false
		}
		inner = location.Parent
	}
	return false
}

// rackedDeviceLocation returns the location of a device mounted in rack: the
// one the device records when it holds the rack's, and the rack's otherwise.
func (inv *Inventory) rackedDeviceLocation(recorded uuid.UUID, rack *CaniRackType) uuid.UUID {
	if inv.LocationHolds(recorded, rack.Location) {
		return recorded
	}
	return rack.Location
}

// SetDeviceLocation records the location a device belongs to. The device
// need not be in the inventory yet, so an add can place what it is building.
// A location that does not exist, one that does not hold the device's rack,
// and any location for a device installed in another device are refused.
func (inv *Inventory) SetDeviceLocation(device *CaniDeviceType, locationID uuid.UUID) error {
	location := inv.Locations[locationID]
	if location == nil {
		return fmt.Errorf("location %s not found", locationID)
	}
	if device.Parent == uuid.Nil {
		device.Location = locationID
		return nil
	}
	rack := inv.Racks[device.Parent]
	if rack == nil {
		return fmt.Errorf("device %q is installed in another device and takes that device's location", device.Name)
	}
	if !inv.LocationHolds(locationID, rack.Location) {
		return fmt.Errorf("device %q is mounted in rack %q, so its location must be the rack's or one that holds it; %q is neither",
			device.Name, rack.Name, location.Name)
	}
	device.Location = locationID
	return nil
}

// MoveRack places a rack in a location and takes the devices mounted in it
// along: none keeps a location of its own, so each is in the rack's new
// location after the next rebuild.
func (inv *Inventory) MoveRack(rackID, locationID uuid.UUID) error {
	rack := inv.Racks[rackID]
	if rack == nil {
		return fmt.Errorf("rack %s not found", rackID)
	}
	if inv.Locations[locationID] == nil {
		return fmt.Errorf("location %s not found", locationID)
	}
	rack.Location = locationID
	inv.releaseRackedDevices(rackID)
	return nil
}

// releaseRackedDevices clears the location each device mounted in the rack
// records, so the next rebuild gives it the rack's.
func (inv *Inventory) releaseRackedDevices(rackID uuid.UUID) {
	for _, device := range inv.Devices {
		if device != nil && device.Parent == rackID {
			device.Location = uuid.Nil
		}
	}
}
