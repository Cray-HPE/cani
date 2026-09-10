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

// MergeCables merges incoming cables by UUID, provider identity, label, or
// unordered endpoint pair. The inventory map key remains the authoritative ID.
func (inv *Inventory) MergeCables(incoming map[uuid.UUID]*CaniCableType) map[uuid.UUID]uuid.UUID {
	if inv.Cables == nil {
		inv.Cables = make(map[uuid.UUID]*CaniCableType)
	}
	remap := make(map[uuid.UUID]uuid.UUID, len(incoming))
	for incomingID, cable := range incoming {
		if cable == nil {
			continue
		}
		if _, ok := inv.Cables[incomingID]; ok {
			cable.ID = incomingID
			inv.Cables[incomingID] = cable
			remap[incomingID] = incomingID
			continue
		}
		if existingID, ok := inv.findMatchingCable(cable); ok {
			cable.ID = existingID
			inv.Cables[existingID] = cable
			remap[incomingID] = existingID
			continue
		}
		cable.ID = incomingID
		inv.Cables[incomingID] = cable
		remap[incomingID] = incomingID
	}
	return remap
}

func (inv *Inventory) findMatchingCable(incoming *CaniCableType) (uuid.UUID, bool) {
	for existingID, existing := range inv.Cables {
		if existing == nil {
			continue
		}
		if sharedExternalID(existing.ExternalIDs, incoming.ExternalIDs) ||
			(!conflictingExternalID(existing.ExternalIDs, incoming.ExternalIDs) &&
				((incoming.Label != "" && existing.Label == incoming.Label) || sameCableEndpoints(existing, incoming))) {
			return existingID, true
		}
	}
	return uuid.Nil, false
}

// mergeCableByLabel retains the legacy focused helper while preserving the
// inventory UUID when a label match replaces the cable contents.
func (inv *Inventory) mergeCableByLabel(cable *CaniCableType) bool {
	if cable == nil || cable.Label == "" {
		return false
	}
	for existingID, existing := range inv.Cables {
		if existing != nil && existing.Label == cable.Label {
			cable.ID = existingID
			inv.Cables[existingID] = cable
			return true
		}
	}
	return false
}

func sharedExternalID(existing, incoming map[string]uuid.UUID) bool {
	for source, incomingID := range incoming {
		if incomingID != uuid.Nil && existing[source] == incomingID {
			return true
		}
	}
	return false
}

func conflictingExternalID(existing, incoming map[string]uuid.UUID) bool {
	for source, incomingID := range incoming {
		if existingID := existing[source]; incomingID != uuid.Nil && existingID != uuid.Nil && existingID != incomingID {
			return true
		}
	}
	return false
}

func sameCableEndpoints(left, right *CaniCableType) bool {
	leftA, leftB := cableEndpoint(left.TerminationA, left.TerminationADevice, left.TerminationAPort),
		cableEndpoint(left.TerminationB, left.TerminationBDevice, left.TerminationBPort)
	rightA, rightB := cableEndpoint(right.TerminationA, right.TerminationADevice, right.TerminationAPort),
		cableEndpoint(right.TerminationB, right.TerminationBDevice, right.TerminationBPort)
	if leftA == "" || leftB == "" || rightA == "" || rightB == "" {
		return false
	}
	return leftA == rightA && leftB == rightB || leftA == rightB && leftB == rightA
}

func cableEndpoint(interfaceID, deviceID uuid.UUID, port string) string {
	if interfaceID != uuid.Nil {
		return "interface:" + interfaceID.String()
	}
	if deviceID != uuid.Nil && port != "" {
		return fmt.Sprintf("device:%s:%s", deviceID, port)
	}
	return ""
}
