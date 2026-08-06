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

	"github.com/google/uuid"
)

// rebuildCableRelationships resolves missing cable TerminationA/B interface
// UUIDs from device + port name. Explicit termination UUIDs are preserved so
// validateCableRelationships can report contradictory endpoint fields.
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

		// Resolve TerminationA interface UUID from device + port name.
		if cable.TerminationA == uuid.Nil && cable.TerminationADevice != uuid.Nil && cable.TerminationAPort != "" {
			if ifaceID := inv.FindInterfaceIDByPort(cable.TerminationADevice, cable.TerminationAPort); ifaceID != uuid.Nil {
				cable.TerminationA = ifaceID
				res.Fixed = append(res.Fixed,
					fmt.Sprintf("cable %q (%s): resolved termination A interface for port %q",
						cable.Label, cableID, cable.TerminationAPort))
			}
		}

		// Resolve TerminationB interface UUID from device + port name.
		if cable.TerminationB == uuid.Nil && cable.TerminationBDevice != uuid.Nil && cable.TerminationBPort != "" {
			if ifaceID := inv.FindInterfaceIDByPort(cable.TerminationBDevice, cable.TerminationBPort); ifaceID != uuid.Nil {
				cable.TerminationB = ifaceID
				res.Fixed = append(res.Fixed,
					fmt.Sprintf("cable %q (%s): resolved termination B interface for port %q",
						cable.Label, cableID, cable.TerminationBPort))
			}
		}
	}
	return res
}

// FindInterfaceIDByPort finds an interface UUID on a device (or its modules)
// by matching the port name. Also handles the case where the ID refers
// directly to a module. Returns uuid.Nil if not found.
func (inv *Inventory) FindInterfaceIDByPort(deviceID uuid.UUID, portName string) uuid.UUID {
	// Check if the ID refers directly to a module.
	if mod, ok := inv.Modules[deviceID]; ok && mod != nil {
		for i := range mod.Interfaces {
			if mod.Interfaces[i].Name == portName {
				return mod.Interfaces[i].ID
			}
		}
		return uuid.Nil
	}

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

// validateCableRelationships verifies each populated cable endpoint is a
// coherent device or module, port name, and interface UUID tuple.
func (inv *Inventory) validateCableRelationships() *RelationshipResult {
	res := &RelationshipResult{}
	for id, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		res.merge(inv.validateCableEnd(id, cable, "A",
			cable.TerminationADevice, cable.TerminationAPort, cable.TerminationA))
		res.merge(inv.validateCableEnd(id, cable, "B",
			cable.TerminationBDevice, cable.TerminationBPort, cable.TerminationB))
	}
	return res
}

// cableEndpointOwner describes a device or module used by a cable endpoint.
func (inv *Inventory) cableEndpointOwner(deviceRef uuid.UUID) (string, string, bool) {
	if device := inv.Devices[deviceRef]; device != nil {
		return "device", displayEndpointName(device.Name, deviceRef), true
	}
	if module := inv.Modules[deviceRef]; module != nil {
		return "module", displayEndpointName(module.Name, deviceRef), true
	}
	return "device or module", deviceRef.String(), false
}

func displayEndpointName(name string, id uuid.UUID) string {
	if name != "" {
		return name
	}
	return id.String()
}

// validateCableEnd checks one end of a cable for reference integrity.
func (inv *Inventory) validateCableEnd(
	cableID uuid.UUID,
	cable *CaniCableType,
	side string,
	deviceRef uuid.UUID,
	portName string,
	ifaceRef uuid.UUID,
) *RelationshipResult {
	res := &RelationshipResult{}
	if deviceRef == uuid.Nil && portName == "" && ifaceRef == uuid.Nil {
		return res
	}
	if deviceRef == uuid.Nil {
		res.Errors = append(res.Errors, fmt.Errorf(
			"cable %q (%s): termination %s device or module is required for populated endpoint",
			cable.Label, cableID, side))
		return res
	}

	ownerKind, ownerName, ownerExists := inv.cableEndpointOwner(deviceRef)
	if !ownerExists {
		res.Errors = append(res.Errors, fmt.Errorf(
			"cable %q (%s): termination %s %s %s not found",
			cable.Label, cableID, side, ownerKind, ownerName))
		return res
	}
	if portName == "" {
		res.Errors = append(res.Errors, fmt.Errorf(
			"cable %q (%s): termination %s port is required for %s %q",
			cable.Label, cableID, side, ownerKind, ownerName))
		return res
	}

	resolvedID := inv.FindInterfaceIDByPort(deviceRef, portName)
	if resolvedID == uuid.Nil {
		res.Errors = append(res.Errors, fmt.Errorf(
			"cable %q (%s): termination %s port %q not found on %s %q",
			cable.Label, cableID, side, portName, ownerKind, ownerName))
		return res
	}
	if ifaceRef == uuid.Nil {
		res.Errors = append(res.Errors, fmt.Errorf(
			"cable %q (%s): termination %s interface is unresolved for port %q on %s %q",
			cable.Label, cableID, side, portName, ownerKind, ownerName))
		return res
	}
	if iface, _ := inv.GetInterfaceByID(ifaceRef); iface == nil {
		res.Errors = append(res.Errors, fmt.Errorf(
			"cable %q (%s): termination %s interface %s not found for port %q on %s %q",
			cable.Label, cableID, side, ifaceRef, portName, ownerKind, ownerName))
		return res
	}
	if ifaceRef != resolvedID {
		res.Errors = append(res.Errors, fmt.Errorf(
			"cable %q (%s): termination %s interface %s does not match port %q on %s %q (expected %s)",
			cable.Label, cableID, side, ifaceRef, portName, ownerKind, ownerName, resolvedID))
	}
	return res
}
