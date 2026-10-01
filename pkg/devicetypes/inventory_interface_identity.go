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
	"bytes"
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// interfaceOwner identifies the device or module whose embedded interface
// specs are being indexed. deviceID is always the (parent) device.
type interfaceOwner struct {
	deviceID uuid.UUID
	ownerID  uuid.UUID
	kind     string
	name     string
}

// sortedIDs returns the map keys in byte order so index rebuilds visit owners
// deterministically and the same owner keeps a contested interface ID.
func sortedIDs[T any](items map[uuid.UUID]*T) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	slices.SortFunc(ids, func(a, b uuid.UUID) int { return bytes.Compare(a[:], b[:]) })
	return ids
}

// indexInterfaceSpecs adds the owner's interface specs to Inventory.Interfaces,
// stamping missing IDs and re-identifying any spec whose ID is already taken
// by a spec indexed earlier in the same rebuild.
func (inv *Inventory) indexInterfaceSpecs(specs []InterfaceSpec, owner interfaceOwner,
	oldIfaces map[uuid.UUID]bool, result *RelationshipResult) {
	for i := range specs {
		spec := &specs[i]
		if spec.ID == uuid.Nil {
			spec.ID = uuid.New()
		}
		if _, taken := inv.Interfaces[spec.ID]; taken {
			inv.reidentifyInterface(spec, owner, result)
		}
		inv.Interfaces[spec.ID] = interfaceInstanceFromSpec(spec, owner.deviceID)
		if !oldIfaces[spec.ID] {
			result.Fixed = append(result.Fixed,
				fmt.Sprintf("interface %q (%s) indexed from %s %q",
					spec.Name, spec.ID, owner.kind, owner.name))
		}
	}
}

// reidentifyInterface gives a spec a fresh ID when another owner already holds
// its current one, and moves the owner's cable terminations to the new ID.
func (inv *Inventory) reidentifyInterface(spec *InterfaceSpec, owner interfaceOwner, result *RelationshipResult) {
	oldID := spec.ID
	spec.ID = uuid.New()
	inv.rebindCableTerminations(owner, spec.Name, oldID, spec.ID)
	result.Fixed = append(result.Fixed,
		fmt.Sprintf("interface %q on %s %q: id %s was shared with another owner, re-identified as %s",
			spec.Name, owner.kind, owner.name, oldID, spec.ID))
}

// rebindCableTerminations repoints cable ends that address the owner directly
// by the port name and the old interface ID at the new ID. Ends addressed
// through a module's parent device keep resolving to the owner that kept the
// ID, which is also the module FindInterfaceIDByPort visits first.
func (inv *Inventory) rebindCableTerminations(owner interfaceOwner, port string, oldID, newID uuid.UUID) {
	for _, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		if cable.TerminationA == oldID && cable.TerminationAPort == port && cable.TerminationADevice == owner.ownerID {
			cable.TerminationA = newID
		}
		if cable.TerminationB == oldID && cable.TerminationBPort == port && cable.TerminationBDevice == owner.ownerID {
			cable.TerminationB = newID
		}
	}
}
