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
	"strings"
	"testing"

	"github.com/google/uuid"
)

// orderedID returns a UUID whose byte order is fixed by n, so a test controls
// which of two objects sortedIDs visits first.
func orderedID(n byte) uuid.UUID {
	var id uuid.UUID
	id[15] = n
	return id
}

// siblingNICs returns the fixture chassis with two installed modules whose
// single ports are both named "Port 1", the first sorting before the second,
// and a switch to cable them to.
func siblingNICs(t *testing.T) (inv *Inventory, chassis *CaniDeviceType, first, second *CaniModuleType, sw *CaniDeviceType) {
	t.Helper()
	inv, chassis = chassisWithBays()
	first = &CaniModuleType{ID: orderedID(1), Name: "nic-first", Model: "nic", ParentDevice: chassis.ID, ModuleBayName: "slot1",
		Interfaces: []InterfaceSpec{{ID: uuid.New(), Name: "Port 1"}}}
	second = &CaniModuleType{ID: orderedID(2), Name: "nic-second", Model: "nic", ParentDevice: chassis.ID, ModuleBayName: "slot2",
		Interfaces: []InterfaceSpec{{ID: uuid.New(), Name: "Port 1"}}}
	inv.Modules[first.ID], inv.Modules[second.ID] = first, second
	sw = &CaniDeviceType{ID: uuid.New(), Name: "switch", Parent: chassis.Parent,
		Interfaces: []InterfaceSpec{{ID: uuid.New(), Name: "swp1"}}}
	inv.Devices[sw.ID] = sw
	if err := inv.RebuildDerivedState().Err(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return inv, chassis, first, second, sw
}

// TestCableEndKeepsRecordedPortOnSiblingModule verifies a device-addressed
// cable end whose recorded interface is a port on the second of two sibling
// modules with same-named ports validates, while a recorded interface on
// another device is still refused.
//
// Why it matters: every library module type names its ports literally, so
// two identical NICs in one server share "Port 1"; re-resolving the end by
// name picked whichever module sorted first and refused the cable, and with
// it every add, place and swap, whenever the cabled NIC sorted second.
// Inputs: the sibling fixture with a cable from the chassis "Port 1" recorded
// on the second module's port, then a cable recorded on the switch's own
// port. Outputs: no error for the first; one "does not match port" error for
// the second.
// Data choice: fixed UUIDs make the cabled module the one a name lookup does
// not find, so the test cannot pass by luck of UUID order.
func TestCableEndKeepsRecordedPortOnSiblingModule(t *testing.T) {
	inv, chassis, first, second, sw := siblingNICs(t)
	if got := inv.FindInterfaceIDByPort(chassis.ID, "Port 1"); got != first.Interfaces[0].ID {
		t.Fatalf("name lookup resolved %s, want the first module's port so the second is the contested one", got)
	}
	cable := &CaniCableType{ID: uuid.New(), Label: "to-second",
		TerminationADevice: chassis.ID, TerminationAPort: "Port 1", TerminationA: second.Interfaces[0].ID,
		TerminationBDevice: sw.ID, TerminationBPort: "swp1", TerminationB: sw.Interfaces[0].ID}
	inv.Cables[cable.ID] = cable

	if err := inv.RebuildDerivedState().Err(); err != nil {
		t.Fatalf("cable on the second module's port: %v", err)
	}

	cable.TerminationA = sw.Interfaces[0].ID
	err := inv.RebuildDerivedState().Err()
	if err == nil || !strings.Contains(err.Error(), `does not match port "Port 1"`) {
		t.Fatalf("cable recorded on another device's port: error = %v, want a port mismatch", err)
	}
}
