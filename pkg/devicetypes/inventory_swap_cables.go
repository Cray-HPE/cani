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

// cableEnd addresses one termination of a cable for in-place edits.
type cableEnd struct {
	cable *CaniCableType
	side  string
}

// fields returns the owner, port name and interface of the end.
func (e cableEnd) fields() (owner *uuid.UUID, port *string, iface *uuid.UUID) {
	if e.side == "A" {
		return &e.cable.TerminationADevice, &e.cable.TerminationAPort, &e.cable.TerminationA
	}
	return &e.cable.TerminationBDevice, &e.cable.TerminationBPort, &e.cable.TerminationB
}

// cableEndState is what a cable end held before a swap edited it, so a
// refused swap can put the cable back.
type cableEndState struct {
	end   cableEnd
	owner uuid.UUID
	port  string
	iface uuid.UUID
}

// state records the end's current owner, port name and interface.
func (e cableEnd) state() cableEndState {
	owner, port, iface := e.fields()
	return cableEndState{end: e, owner: *owner, port: *port, iface: *iface}
}

// onModule reports whether the end terminates on one of the module's ports:
// it names the module as its owner, or it names the parent device and
// resolved to one of the module's interfaces.
func (e cableEnd) onModule(moduleID, parent uuid.UUID, ports map[uuid.UUID]bool) bool {
	if e.cable == nil {
		return false
	}
	owner, _, iface := e.fields()
	return *owner == moduleID || (parent != uuid.Nil && *owner == parent && ports[*iface])
}

// cableEndsOnModule returns every cable end that terminates on one of the
// module's ports, in cable ID order. An end addresses a module port either
// by naming the module as its owner, or by naming the parent device with the
// port name the bay gives the module; the latter is what `add cable
// --a-device <device>` and connection imports record, and is matched by the
// interface the end resolved to. parent is the device the module is
// installed in, taken before the module is moved.
func (inv *Inventory) cableEndsOnModule(mod *CaniModuleType, parent uuid.UUID) []cableEndState {
	ports := make(map[uuid.UUID]bool, len(mod.Interfaces))
	for _, iface := range mod.Interfaces {
		ports[iface.ID] = true
	}
	var ends []cableEndState
	for _, id := range sortedIDs(inv.Cables) {
		for _, side := range []string{"A", "B"} {
			end := cableEnd{cable: inv.Cables[id], side: side}
			if end.onModule(mod.ID, parent, ports) {
				ends = append(ends, end.state())
			}
		}
	}
	return ends
}

// restoreCableEnds puts every end back to the state it held before the swap.
func restoreCableEnds(states []cableEndState) {
	for _, s := range states {
		owner, port, iface := s.end.fields()
		*owner, *port, *iface = s.owner, s.port, s.iface
	}
}

// CablesOnModule returns the cables with an end on one of the module's
// ports, whether the end names the module or its parent device as owner.
func (inv *Inventory) CablesOnModule(id uuid.UUID) []*CaniCableType {
	mod := inv.Modules[id]
	if mod == nil {
		return nil
	}
	seen := make(map[*CaniCableType]bool)
	var cables []*CaniCableType
	for _, s := range inv.cableEndsOnModule(mod, mod.ParentDevice) {
		if !seen[s.end.cable] {
			seen[s.end.cable] = true
			cables = append(cables, s.end.cable)
		}
	}
	return cables
}

// moveCablesToSpare re-terminates every cable end on the outgoing module's
// ports at the spare's port of the same name, so the cables stay in the bay
// as they do when a card is replaced in place. An end that names the
// outgoing module as its owner is re-owned by the spare; an end that names
// the parent device keeps its owner and port name. Called after both
// placements are applied, so the spare's port names are resolved for the
// bay. Nothing is changed when the spare lacks a port some cabled end needs.
func moveCablesToSpare(ends []cableEndState, outgoing, spare *CaniModuleType) error {
	byName := make(map[string]uuid.UUID, len(spare.Interfaces))
	for _, iface := range spare.Interfaces {
		byName[iface.Name] = iface.ID
	}
	for _, s := range ends {
		if _, ok := byName[s.port]; !ok {
			return fmt.Errorf("cable %q (%s): spare %q has no port %q to take over from %q; disconnect the cable or swap like for like",
				s.end.cable.Label, s.end.cable.ID, spare.Name, s.port, outgoing.Name)
		}
	}
	for _, s := range ends {
		owner, _, iface := s.end.fields()
		if *owner == outgoing.ID {
			*owner = spare.ID
		}
		*iface = byName[s.port]
	}
	return nil
}
