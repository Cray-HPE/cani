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
package transform

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// MapCables converts Nautobot Cable objects to CANI cables.
// It requires:
//   - deviceMap: Nautobot device UUID → CANI device UUID
//   - ifaceMap: Nautobot interface UUID → (owner UUIDs, interface name)
//   - moduleMap: Nautobot module UUID → CANI module UUID, for a termination
//     on a module-owned interface, which the inventory records against its
//     module even though Nautobot 3.2 also names the root device on it
func MapCables(
	raw []nautobotapi.Cable,
	deviceMap map[uuid.UUID]uuid.UUID,
	ifaceMap map[uuid.UUID]ifaceRef,
	moduleMap map[uuid.UUID]uuid.UUID,
	statusNameMap map[uuid.UUID]string,
) map[uuid.UUID]*devicetypes.CaniCableType {
	result := make(map[uuid.UUID]*devicetypes.CaniCableType, len(raw))

	for _, cable := range raw {
		nbID := directUUID(cable.Id)
		if nbID == uuid.Nil {
			continue
		}
		caniID := uuid.New()

		caniCable := &devicetypes.CaniCableType{
			ID:    caniID,
			Label: strVal(cable.Label),
			ObjectMeta: devicetypes.ObjectMeta{
				Status:      resolveRefName(cable.Status, statusNameMap),
				ExternalIDs: map[string]uuid.UUID{externalIDKeyNautobot: nbID},
			},
			Color: strVal(cable.Color),
		}

		// Cable type.
		if cable.Type != nil && cable.Type.Value != nil {
			caniCable.CableType = string(*cable.Type.Value)
		}

		// Length.
		if cable.Length != nil {
			l := float64(*cable.Length)
			caniCable.Length = &l
		}
		if cable.LengthUnit != nil && cable.LengthUnit.Value != nil {
			caniCable.LengthUnit = string(*cable.LengthUnit.Value)
		}

		// Resolve termination A.
		termAIfaceID := directUUID(cable.TerminationAId)
		if ref, ok := ifaceMap[termAIfaceID]; ok {
			if ownerID, ok2 := terminationOwner(ref, deviceMap, moduleMap); ok2 {
				caniCable.TerminationA = termAIfaceID
				caniCable.TerminationADevice = ownerID
				caniCable.TerminationAPort = ref.name
			}
		}
		caniCable.TerminationAType = strVal(cable.TerminationAType)

		// Resolve termination B.
		termBIfaceID := directUUID(cable.TerminationBId)
		if ref, ok := ifaceMap[termBIfaceID]; ok {
			if ownerID, ok2 := terminationOwner(ref, deviceMap, moduleMap); ok2 {
				caniCable.TerminationB = termBIfaceID
				caniCable.TerminationBDevice = ownerID
				caniCable.TerminationBPort = ref.name
			}
		}
		caniCable.TerminationBType = strVal(cable.TerminationBType)

		if cable.CustomFields != nil {
			caniCable.CustomFields = convCustomFields(cable.CustomFields)
		}

		result[caniID] = caniCable
	}

	return result
}

// terminationOwner returns the CANI device or module a cable end terminates
// on. A module-owned interface resolves through moduleMap to the module
// itself, the owner the inventory records for a module port, before its
// device is considered: Nautobot 3.2 also sets the root device on a port of
// an installed module. An end whose module was not imported stays
// unresolved rather than falling back to that device. A device-owned
// interface resolves through deviceMap.
func terminationOwner(ref ifaceRef, deviceMap, moduleMap map[uuid.UUID]uuid.UUID) (uuid.UUID, bool) {
	if ref.moduleID != uuid.Nil {
		id, ok := moduleMap[ref.moduleID]
		return id, ok
	}
	if ref.deviceID != uuid.Nil {
		id, ok := deviceMap[ref.deviceID]
		return id, ok
	}
	return uuid.Nil, false
}

// ifaceRef holds the owner and name of an interface, used during cable
// mapping. moduleID is set for a module-owned interface and deviceID for any
// interface Nautobot places on a device: a device-owned port, or, since
// Nautobot 3.2, a port of a module installed in that device. A port of a
// spare module carries only moduleID.
type ifaceRef struct {
	deviceID uuid.UUID // Nautobot device UUID; Nil for a port of a spare module
	moduleID uuid.UUID // Nautobot module UUID; Nil for a device-owned interface
	name     string
}

// BuildInterfaceMap creates a lookup from Nautobot interface UUID → (owner, name).
func BuildInterfaceMap(ifaces []nautobotapi.Interface) map[uuid.UUID]ifaceRef {
	m := make(map[uuid.UUID]ifaceRef, len(ifaces))
	for _, iface := range ifaces {
		ifaceID := directUUID(iface.Id)
		if ifaceID == uuid.Nil {
			continue
		}
		ref := ifaceRef{name: iface.Name}
		if iface.Device != nil {
			ref.deviceID = tenantRefID(iface.Device)
		}
		if iface.Module != nil {
			ref.moduleID = refID(iface.Module)
		}
		m[ifaceID] = ref
	}
	return m
}
