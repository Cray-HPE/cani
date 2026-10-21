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
	"strings"

	"github.com/google/uuid"
)

// InstallInBay places child into a device bay of the parent device. An empty
// bay selects the first free bay that allows the child. A named bay must
// exist on the parent, be free (or already hold child), and allow the child's
// slug when the bay restricts slugs. pending holds devices that are not yet in
// the inventory but whose bay assignments must be respected (may be nil).
// Rack position and face are cleared because a child's placement comes from
// its parent.
func (inv *Inventory) InstallInBay(child *CaniDeviceType, parentID uuid.UUID, bay string, pending map[uuid.UUID]*CaniDeviceType) error {
	parent, ok := inv.Devices[parentID]
	if !ok || parent == nil {
		return fmt.Errorf("parent device %s not found", parentID)
	}
	bays := deviceBaysOf(parent)
	if len(bays) == 0 {
		return fmt.Errorf("device %q has no device bays", parent.Name)
	}
	occupied := inv.bayOccupants(parentID, pending)
	spec, err := selectBay(parent, bays, child, bay, occupied)
	if err != nil {
		return err
	}
	child.Parent = parent.ID
	child.ParentBay = spec.Name
	child.RackPosition = 0
	child.Face = ""
	return nil
}

// HasDeviceBays reports whether the device (or its library type) defines
// device bays that children can be installed into.
func (c *CaniDeviceType) HasDeviceBays() bool {
	return c != nil && len(deviceBaysOf(c)) > 0
}

// deviceBaysOf returns the parent's device bays, falling back to its library
// device type for instances saved without bay specs.
func deviceBaysOf(parent *CaniDeviceType) []DeviceBaySpec {
	if len(parent.DeviceBays) > 0 || parent.Slug == "" {
		return parent.DeviceBays
	}
	if tmpl, ok := GetBySlug(parent.Slug); ok {
		return tmpl.DeviceBays
	}
	return nil
}

// bayOccupants maps each occupied bay name on parentID to its device.
func (inv *Inventory) bayOccupants(parentID uuid.UUID, pending map[uuid.UUID]*CaniDeviceType) map[string]*CaniDeviceType {
	occupied := make(map[string]*CaniDeviceType)
	for _, devices := range []map[uuid.UUID]*CaniDeviceType{inv.Devices, pending} {
		for _, d := range devices {
			if d != nil && d.Parent == parentID && d.ParentBay != "" {
				occupied[d.ParentBay] = d
			}
		}
	}
	return occupied
}

func otherBayOccupant(occupied map[string]*CaniDeviceType, bay string, childID uuid.UUID) *CaniDeviceType {
	for name, device := range occupied {
		if strings.EqualFold(name, bay) && device.ID != childID {
			return device
		}
	}
	return nil
}

// selectBay returns the named bay, or the first free bay allowing child when
// name is empty.
func selectBay(parent *CaniDeviceType, bays []DeviceBaySpec, child *CaniDeviceType, name string, occupied map[string]*CaniDeviceType) (DeviceBaySpec, error) {
	if name == "" {
		return firstFreeBay(parent, bays, child, occupied)
	}
	for _, b := range bays {
		if !strings.EqualFold(b.Name, name) {
			continue
		}
		if occ := otherBayOccupant(occupied, b.Name, child.ID); occ != nil {
			return DeviceBaySpec{}, fmt.Errorf("bay %q on %q is occupied by %q", b.Name, parent.Name, occ.Name)
		}
		return b, bayAllows(b, child)
	}
	return DeviceBaySpec{}, fmt.Errorf("device %q has no bay %q (bays: %s)", parent.Name, name, bayNames(bays))
}

// firstFreeBay returns the first unoccupied bay that allows child.
func firstFreeBay(parent *CaniDeviceType, bays []DeviceBaySpec, child *CaniDeviceType, occupied map[string]*CaniDeviceType) (DeviceBaySpec, error) {
	for _, b := range bays {
		occ := otherBayOccupant(occupied, b.Name, child.ID)
		if occ == nil && bayAllows(b, child) == nil {
			return b, nil
		}
	}
	return DeviceBaySpec{}, fmt.Errorf("no free bay on %q for %q (bays: %s)", parent.Name, child.Slug, bayNames(bays))
}

// bayAllows reports an error when the bay restricts slugs and child's slug is
// not among them.
func bayAllows(b DeviceBaySpec, child *CaniDeviceType) error {
	if b.Allowed == nil {
		return nil
	}
	allowed := b.Allowed.Slugs()
	if len(allowed) == 0 || slices.Contains(allowed, child.Slug) {
		return nil
	}
	return fmt.Errorf("bay %q does not allow %q (allowed: %s)", b.Name, child.Slug, strings.Join(allowed, ", "))
}

// bayNames lists bay names for error messages.
func bayNames(bays []DeviceBaySpec) string {
	names := make([]string, 0, len(bays))
	for _, b := range bays {
		names = append(names, b.Name)
	}
	return strings.Join(names, ", ")
}
