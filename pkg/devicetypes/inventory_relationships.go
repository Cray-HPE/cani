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
package devicetypes

import (
	"fmt"
	"log"

	"github.com/google/uuid"
)

// RelationshipResult holds the outcome of a relationship verification pass.
// Fixed lists auto-corrections applied, Warnings lists non-fatal issues,
// and Errors lists broken references that could not be resolved.
// Orphans collects items with no parent assigned.
type RelationshipResult struct {
	Fixed    []string
	Warnings []string
	Errors   []error
	Orphans  []OrphanItem
}

// HasErrors returns true when unresolvable relationship errors exist.
func (r *RelationshipResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasOrphans returns true when orphaned items were detected.
func (r *RelationshipResult) HasOrphans() bool {
	return len(r.Orphans) > 0
}

// merge combines another RelationshipResult into this one.
func (r *RelationshipResult) merge(other *RelationshipResult) {
	if other == nil {
		return
	}
	r.Fixed = append(r.Fixed, other.Fixed...)
	r.Warnings = append(r.Warnings, other.Warnings...)
	r.Errors = append(r.Errors, other.Errors...)
	r.Orphans = append(r.Orphans, other.Orphans...)
}

// logSummary logs all fixes, warnings, and errors.
func (r *RelationshipResult) logSummary() {
	if Debug {
		for _, f := range r.Fixed {
			log.Printf("Fixed: %s", f)
		}
		for _, w := range r.Warnings {
			log.Printf("Warning: %s", w)
		}
	}
	for _, e := range r.Errors {
		log.Printf("Error: %s", e)
	}
}

// uuidSetFromSlice builds a set from a UUID slice.
func uuidSetFromSlice(s []uuid.UUID) map[uuid.UUID]bool {
	m := make(map[uuid.UUID]bool, len(s))
	for _, id := range s {
		m[id] = true
	}
	return m
}

// uuidSetsEqual returns true when both sets contain the same elements.
func uuidSetsEqual(a, b map[uuid.UUID]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for id := range a {
		if !b[id] {
			return false
		}
	}
	return true
}

// rebuildLocationRelationships rebuilds Location.Children from each
// location's Parent field, reporting only actual changes.
func (inv *Inventory) rebuildLocationRelationships() *RelationshipResult {
	res := &RelationshipResult{}

	// Compute expected children from Parent pointers.
	expected := make(map[uuid.UUID]map[uuid.UUID]bool)
	for id, loc := range inv.Locations {
		if loc == nil || loc.Parent == uuid.Nil {
			continue
		}
		if _, ok := inv.Locations[loc.Parent]; !ok {
			res.Errors = append(res.Errors,
				fmt.Errorf("location %q (%s): parent location %s not found",
					loc.Name, id, loc.Parent))
			continue
		}
		if expected[loc.Parent] == nil {
			expected[loc.Parent] = make(map[uuid.UUID]bool)
		}
		expected[loc.Parent][id] = true
	}

	// Reconcile each location's Children list.
	for id, loc := range inv.Locations {
		if loc == nil {
			continue
		}
		current := uuidSetFromSlice(loc.Children)
		exp := expected[id]
		if uuidSetsEqual(current, exp) {
			continue
		}
		loc.Children = nil
		for childID := range exp {
			loc.AddChild(childID)
			if !current[childID] {
				child := inv.Locations[childID]
				if child != nil {
					res.Fixed = append(res.Fixed,
						fmt.Sprintf("location %q added as child of %q", child.Name, loc.Name))
				}
			}
		}
	}
	return res
}

// rebuildRackRelationships rebuilds Location.Racks from each rack's
// Location field, reporting only actual changes.
func (inv *Inventory) rebuildRackRelationships() *RelationshipResult {
	res := &RelationshipResult{}

	// Compute expected rack lists from each rack's Location.
	expected := make(map[uuid.UUID]map[uuid.UUID]bool)
	for id, rack := range inv.Racks {
		if rack == nil {
			continue
		}
		if rack.Location == uuid.Nil {
			res.Warnings = append(res.Warnings,
				fmt.Sprintf("rack %q (%s): no location assigned", rack.Name, id))
			res.Orphans = append(res.Orphans, OrphanItem{
				ID:               rack.ID,
				Name:             rack.Name,
				Kind:             "rack",
				DeviceType:       string(rack.Type),
				Model:            rack.Model,
				Manufacturer:     rack.Manufacturer,
				ProviderMetadata: rack.ProviderMetadata,
			})
			continue
		}
		if _, ok := inv.Locations[rack.Location]; !ok {
			res.Errors = append(res.Errors,
				fmt.Errorf("rack %q (%s): location %s not found",
					rack.Name, id, rack.Location))
			continue
		}
		if expected[rack.Location] == nil {
			expected[rack.Location] = make(map[uuid.UUID]bool)
		}
		expected[rack.Location][id] = true
	}

	// Reconcile each location's Racks list.
	for id, loc := range inv.Locations {
		if loc == nil {
			continue
		}
		current := uuidSetFromSlice(loc.Racks)
		exp := expected[id]
		if uuidSetsEqual(current, exp) {
			continue
		}
		loc.Racks = nil
		for rackID := range exp {
			loc.AddRack(rackID)
			if !current[rackID] {
				rack := inv.Racks[rackID]
				if rack != nil {
					res.Fixed = append(res.Fixed,
						fmt.Sprintf("rack %q added to location %q", rack.Name, loc.Name))
				}
			}
		}
	}
	return res
}

// clearDeviceReverseLists resets all derived fields on racks and devices
// so they can be rebuilt from Parent fields.
func (inv *Inventory) clearDeviceReverseLists() {
	for _, rack := range inv.Racks {
		if rack != nil {
			rack.Devices = nil
			rack.OccupiedSlots = nil
		}
	}
	for _, device := range inv.Devices {
		if device != nil {
			device.Children = nil
			device.Rack = uuid.Nil
			device.Location = uuid.Nil
			device.ParentDevice = uuid.Nil
		}
	}
}

// linkDeviceToRack links a device to its parent rack and returns a fix message.
// Returns true if the parent was a rack.
func (inv *Inventory) linkDeviceToRack(id uuid.UUID, device *CaniDeviceType) (bool, string) {
	rack, ok := inv.Racks[device.Parent]
	if !ok {
		return false, ""
	}
	rack.addDevice(id)
	device.Rack = device.Parent
	device.Location = rack.Location

	// Rebuild OccupiedSlots from the device's stored position.
	if device.RackPosition > 0 {
		height := device.UHeight
		if height < 1 {
			height = 1
		}
		rack.PlaceDevice(id, device.RackPosition, height, device.Face, device.IsFullDepth)
	}

	return true, fmt.Sprintf("device %q added to rack %q", device.Name, rack.Name)
}

// linkDeviceToParentDevice links a device to its parent device and returns a fix message.
// Returns true if the parent was a device.
func (inv *Inventory) linkDeviceToParentDevice(id uuid.UUID, device *CaniDeviceType) (bool, string) {
	parent, ok := inv.Devices[device.Parent]
	if !ok {
		return false, ""
	}
	if !containsUUID(parent.Children, id) {
		parent.Children = append(parent.Children, id)
	}
	device.ParentDevice = device.Parent
	device.Rack = parent.Rack
	device.Location = parent.Location

	// Also register with the rack so rack views can enumerate all
	// mounted devices, not just direct rack children.
	if rack, ok := inv.Racks[parent.Rack]; ok {
		rack.addDevice(id)
	}

	return true, fmt.Sprintf("device %q added as child of device %q", device.Name, parent.Name)
}

// rebuildDeviceRelationships clears and rebuilds Rack.Devices,
// Device.Children, and explicit FK fields (Rack, Location, ParentDevice)
// from each device's Parent field.
//
// The rebuild uses a multi-phase approach so that parent devices are
// always resolved before their children, regardless of map iteration
// order, making the result deterministic and idempotent.
func (inv *Inventory) rebuildDeviceRelationships() *RelationshipResult {
	res := &RelationshipResult{}

	// Snapshot derived fields for change detection.
	type deviceLink struct {
		rack, location, parentDevice uuid.UUID
	}
	old := make(map[uuid.UUID]deviceLink)
	for id, d := range inv.Devices {
		if d != nil {
			old[id] = deviceLink{d.Rack, d.Location, d.ParentDevice}
		}
	}

	inv.clearDeviceReverseLists()

	// Phase 1: link devices whose parent is a rack.
	for id, device := range inv.Devices {
		if device == nil || device.Parent == uuid.Nil {
			continue
		}
		inv.linkDeviceToRack(id, device)
	}

	// Phase 2: iteratively link devices whose parent is another device.
	// Repeats until no progress is made, handling arbitrary nesting
	// depth regardless of map iteration order.
	for {
		progress := false
		for id, device := range inv.Devices {
			if device == nil || device.Parent == uuid.Nil {
				continue
			}
			if inv.deviceAlreadyLinked(device) {
				continue
			}
			if _, isRack := inv.Racks[device.Parent]; isRack {
				continue
			}
			parent, ok := inv.Devices[device.Parent]
			if !ok || parent == nil {
				continue
			}
			if !inv.deviceAlreadyLinked(parent) {
				continue
			}
			if linked, _ := inv.linkDeviceToParentDevice(id, device); linked {
				progress = true
			}
		}
		if !progress {
			break
		}
	}

	// Phase 2b: link devices whose parent is an orphan device (no rack
	// ancestor). This handles device-bay hierarchies added without a
	// rack placement — children still need ParentDevice set so they
	// are not reported as errors.
	for {
		progress := false
		for id, device := range inv.Devices {
			if device == nil || device.Parent == uuid.Nil {
				continue
			}
			if inv.deviceAlreadyLinked(device) {
				continue
			}
			parent, ok := inv.Devices[device.Parent]
			if !ok || parent == nil {
				continue
			}
			device.ParentDevice = device.Parent
			if !containsUUID(parent.Children, id) {
				parent.Children = append(parent.Children, id)
			}
			progress = true
		}
		if !progress {
			break
		}
	}

	// Phase 3: report only actual changes and problems.
	for id, device := range inv.Devices {
		if device == nil {
			continue
		}
		if device.Parent == uuid.Nil {
			res.Warnings = append(res.Warnings,
				fmt.Sprintf("device %q (%s): no parent assigned (orphan)",
					device.Name, id))
			res.Orphans = append(res.Orphans, OrphanItem{
				ID:               device.ID,
				Name:             device.Name,
				Kind:             "device",
				DeviceType:       string(device.Type),
				Model:            device.Model,
				Manufacturer:     device.Manufacturer,
				ProviderMetadata: device.ProviderMetadata,
			})
			continue
		}
		if !inv.deviceAlreadyLinked(device) {
			res.Errors = append(res.Errors,
				fmt.Errorf("device %q (%s): parent %s not found in racks or devices",
					device.Name, id, device.Parent))
			continue
		}
		prev := old[id]
		if device.Rack == prev.rack && device.Location == prev.location &&
			device.ParentDevice == prev.parentDevice {
			continue
		}
		if device.ParentDevice != uuid.Nil {
			parent := inv.Devices[device.ParentDevice]
			if parent != nil {
				res.Fixed = append(res.Fixed,
					fmt.Sprintf("device %q added as child of device %q",
						device.Name, parent.Name))
			}
		} else if device.Rack != uuid.Nil {
			rack := inv.Racks[device.Rack]
			if rack != nil {
				res.Fixed = append(res.Fixed,
					fmt.Sprintf("device %q added to rack %q",
						device.Name, rack.Name))
			}
		}
	}

	return res
}

// deviceAlreadyLinked returns true when a device has been resolved to
// a rack (directly or through an ancestor device).
func (inv *Inventory) deviceAlreadyLinked(device *CaniDeviceType) bool {
	return device.Rack != uuid.Nil || device.ParentDevice != uuid.Nil
}

// validateModuleRelationships verifies every module's ParentDevice
// exists in the device map.
func (inv *Inventory) validateModuleRelationships() *RelationshipResult {
	res := &RelationshipResult{}
	for id, mod := range inv.Modules {
		if mod == nil {
			continue
		}
		if mod.ParentDevice == uuid.Nil {
			res.Warnings = append(res.Warnings,
				fmt.Sprintf("module %q (%s): no parent device assigned",
					mod.Name, id))
			continue
		}
		if _, ok := inv.Devices[mod.ParentDevice]; !ok {
			res.Errors = append(res.Errors,
				fmt.Errorf("module %q (%s): parent device %s not found",
					mod.Name, id, mod.ParentDevice))
		}
	}
	return res
}

// rebuildFruRelationships clears and rebuilds Device.Frus and Module.Frus
// from each FRU's Device/Parent field, and validates references.
func (inv *Inventory) rebuildFruRelationships() *RelationshipResult {
	res := &RelationshipResult{}

	// Clear existing reverse lists.
	for _, dev := range inv.Devices {
		if dev != nil {
			dev.Frus = nil
		}
	}
	for _, mod := range inv.Modules {
		if mod != nil {
			mod.Frus = nil
		}
	}

	for id, fru := range inv.Frus {
		if fru == nil {
			continue
		}
		if fru.Device == uuid.Nil && fru.Parent == uuid.Nil {
			res.Warnings = append(res.Warnings,
				fmt.Sprintf("fru %q (%s): no device or parent assigned",
					fru.Name, id))
			continue
		}

		// Validate and link by Device field.
		if fru.Device != uuid.Nil {
			if dev, ok := inv.Devices[fru.Device]; ok {
				if !containsUUID(dev.Frus, id) {
					dev.Frus = append(dev.Frus, id)
				}
			} else if mod, ok := inv.Modules[fru.Device]; ok {
				if !containsUUID(mod.Frus, id) {
					mod.Frus = append(mod.Frus, id)
				}
			} else {
				res.Errors = append(res.Errors,
					fmt.Errorf("fru %q (%s): device %s not found in devices or modules",
						fru.Name, id, fru.Device))
			}
		}

		// Validate Parent field.
		if fru.Parent != uuid.Nil && !inv.parentExists(fru.Parent) {
			res.Errors = append(res.Errors,
				fmt.Errorf("fru %q (%s): parent %s not found",
					fru.Name, id, fru.Parent))
		}
	}
	return res
}

// rebuildCableRelationships resolves cable TerminationA/B interface UUIDs
// from device + port name and removes cables whose device references are
// invalid (e.g. after a device was removed).
//
// Must run AFTER rebuildInterfaceRelationships (needs the rebuilt interface
// index) and BEFORE validateCableRelationships (which validates the resolved
// UUIDs).
func (inv *Inventory) rebuildCableRelationships() *RelationshipResult {
	res := &RelationshipResult{}
	for cableID, cable := range inv.Cables {
		if cable == nil {
			continue
		}

		// Remove cables whose device references are invalid.
		aDeviceOK := cable.TerminationADevice == uuid.Nil || inv.Devices[cable.TerminationADevice] != nil
		bDeviceOK := cable.TerminationBDevice == uuid.Nil || inv.Devices[cable.TerminationBDevice] != nil
		if !aDeviceOK || !bDeviceOK {
			res.Fixed = append(res.Fixed,
				fmt.Sprintf("cable %q (%s): removed (endpoint device deleted)",
					cable.Label, cableID))
			delete(inv.Cables, cableID)
			continue
		}

		// Resolve TerminationA interface UUID from device + port name.
		if cable.TerminationADevice != uuid.Nil && cable.TerminationAPort != "" {
			if ifaceID := inv.findInterfaceIDByPort(cable.TerminationADevice, cable.TerminationAPort); ifaceID != uuid.Nil {
				if cable.TerminationA != ifaceID {
					cable.TerminationA = ifaceID
					res.Fixed = append(res.Fixed,
						fmt.Sprintf("cable %q (%s): resolved termination A interface for port %q",
							cable.Label, cableID, cable.TerminationAPort))
				}
			}
		}

		// Resolve TerminationB interface UUID from device + port name.
		if cable.TerminationBDevice != uuid.Nil && cable.TerminationBPort != "" {
			if ifaceID := inv.findInterfaceIDByPort(cable.TerminationBDevice, cable.TerminationBPort); ifaceID != uuid.Nil {
				if cable.TerminationB != ifaceID {
					cable.TerminationB = ifaceID
					res.Fixed = append(res.Fixed,
						fmt.Sprintf("cable %q (%s): resolved termination B interface for port %q",
							cable.Label, cableID, cable.TerminationBPort))
				}
			}
		}
	}
	return res
}

// findInterfaceIDByPort finds an interface UUID on a device (or its modules)
// by matching the port name. Returns uuid.Nil if not found.
func (inv *Inventory) findInterfaceIDByPort(deviceID uuid.UUID, portName string) uuid.UUID {
	device := inv.Devices[deviceID]
	if device == nil {
		return uuid.Nil
	}
	for i := range device.Interfaces {
		if device.Interfaces[i].Name == portName {
			return device.Interfaces[i].ID
		}
	}
	// Fall back to module interfaces.
	for _, mod := range inv.Modules {
		if mod == nil || mod.ParentDevice != deviceID {
			continue
		}
		for i := range mod.Interfaces {
			if mod.Interfaces[i].Name == portName {
				return mod.Interfaces[i].ID
			}
		}
	}
	return uuid.Nil
}

// validateCableRelationships verifies cable termination devices and
// interface references exist.
func (inv *Inventory) validateCableRelationships() *RelationshipResult {
	res := &RelationshipResult{}
	for id, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		res.merge(inv.validateCableEnd(id, cable, "A",
			cable.TerminationADevice, cable.TerminationA))
		res.merge(inv.validateCableEnd(id, cable, "B",
			cable.TerminationBDevice, cable.TerminationB))
	}
	return res
}

// validateCableEnd checks one end of a cable for reference integrity.
func (inv *Inventory) validateCableEnd(
	cableID uuid.UUID,
	cable *CaniCableType,
	side string,
	deviceRef uuid.UUID,
	ifaceRef uuid.UUID,
) *RelationshipResult {
	res := &RelationshipResult{}
	if deviceRef != uuid.Nil {
		if _, ok := inv.Devices[deviceRef]; !ok {
			res.Errors = append(res.Errors,
				fmt.Errorf("cable %q (%s): termination %s device %s not found",
					cable.Label, cableID, side, deviceRef))
		}
	}
	if ifaceRef != uuid.Nil {
		if iface, _ := inv.GetInterfaceByID(ifaceRef); iface == nil {
			res.Errors = append(res.Errors,
				fmt.Errorf("cable %q (%s): termination %s interface %s not found",
					cable.Label, cableID, side, ifaceRef))
		}
	}
	return res
}

// rebuildInterfaceRelationships clears and rebuilds Inventory.Interfaces
// from each device's and module's embedded Interfaces slices,
// reporting only newly indexed interfaces.
func (inv *Inventory) rebuildInterfaceRelationships() *RelationshipResult {
	res := &RelationshipResult{}

	// Snapshot existing interface IDs for change detection.
	oldIfaces := make(map[uuid.UUID]bool, len(inv.Interfaces))
	for id := range inv.Interfaces {
		oldIfaces[id] = true
	}

	inv.Interfaces = make(map[uuid.UUID]*InterfaceInstance)

	for deviceID, device := range inv.Devices {
		if device == nil {
			continue
		}
		for i := range device.Interfaces {
			iface := &device.Interfaces[i]
			inst := &InterfaceInstance{
				ID:            iface.ID,
				Name:          iface.Name,
				InterfaceType: iface.Type,
				DeviceID:      deviceID,
				ObjectMeta:    ObjectMeta{Status: string(StatusActive)},
				MgmtOnly:      iface.MgmtOnly != nil && *iface.MgmtOnly,
			}
			inv.Interfaces[iface.ID] = inst
			if !oldIfaces[iface.ID] {
				res.Fixed = append(res.Fixed,
					fmt.Sprintf("interface %q (%s) indexed from device %q",
						iface.Name, iface.ID, device.Name))
			}
		}
	}

	for _, mod := range inv.Modules {
		if mod == nil {
			continue
		}
		for i := range mod.Interfaces {
			iface := &mod.Interfaces[i]
			inst := &InterfaceInstance{
				ID:            iface.ID,
				Name:          iface.Name,
				InterfaceType: iface.Type,
				DeviceID:      mod.ParentDevice,
				ObjectMeta:    ObjectMeta{Status: string(StatusActive)},
				MgmtOnly:      iface.MgmtOnly != nil && *iface.MgmtOnly,
			}
			inv.Interfaces[iface.ID] = inst
			if !oldIfaces[iface.ID] {
				res.Fixed = append(res.Fixed,
					fmt.Sprintf("interface %q (%s) indexed from module %q",
						iface.Name, iface.ID, mod.Name))
			}
		}
	}

	return res
}

// detectCircularLocationRefs walks location parent chains to find cycles.
func (inv *Inventory) detectCircularLocationRefs() *RelationshipResult {
	res := &RelationshipResult{}
	for id, loc := range inv.Locations {
		if loc == nil {
			continue
		}
		if inv.hasLocationCycle(id) {
			res.Errors = append(res.Errors,
				fmt.Errorf("location %q (%s): circular parent reference detected",
					loc.Name, id))
		}
	}
	return res
}

// hasLocationCycle returns true if following Parent pointers from id
// leads back to id.
func (inv *Inventory) hasLocationCycle(id uuid.UUID) bool {
	visited := map[uuid.UUID]bool{}
	cur := id
	for {
		loc, ok := inv.Locations[cur]
		if !ok || loc == nil || loc.Parent == uuid.Nil {
			return false
		}
		if visited[cur] {
			return true
		}
		visited[cur] = true
		cur = loc.Parent
	}
}

// detectCircularDeviceRefs walks device parent chains to find cycles.
func (inv *Inventory) detectCircularDeviceRefs() *RelationshipResult {
	res := &RelationshipResult{}
	for id, device := range inv.Devices {
		if device == nil {
			continue
		}
		if inv.hasDeviceCycle(id) {
			res.Errors = append(res.Errors,
				fmt.Errorf("device %q (%s): circular parent reference detected",
					device.Name, id))
		}
	}
	return res
}

// hasDeviceCycle returns true if following device Parent pointers from
// id leads back to id (only follows device→device links).
func (inv *Inventory) hasDeviceCycle(id uuid.UUID) bool {
	visited := map[uuid.UUID]bool{}
	cur := id
	for {
		device, ok := inv.Devices[cur]
		if !ok || device == nil || device.Parent == uuid.Nil {
			return false
		}
		// Stop if parent is a rack (not a device cycle).
		if _, isRack := inv.Racks[device.Parent]; isRack {
			return false
		}
		if visited[cur] {
			return true
		}
		visited[cur] = true
		cur = device.Parent
	}
}
