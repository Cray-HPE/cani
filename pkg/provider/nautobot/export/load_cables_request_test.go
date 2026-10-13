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
package export

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// requestEnds returns two resolved cable ends for request-building tests.
func requestEnds() (cableEnd, cableEnd) {
	return cableEnd{device: &devicetypes.CaniDeviceType{Name: "spine-01"}, port: "eth0", iface: &CachedItem{ID: uuid.New()}},
		cableEnd{device: &devicetypes.CaniDeviceType{Name: "leaf-01"}, port: "eth1", iface: &CachedItem{ID: uuid.New()}}
}

// TestCableRequestCarriesOptionalFields verifies that the create request
// takes the cable's explicit termination types, length, unit and colour, and
// the resolved interfaces as termination IDs.
//
// Why it matters: these fields reach Nautobot only through this request.
// Inputs: a cable with every optional field set, a named colour and unit "ft".
// Outputs: each value in the request, the colour as hex, the IDs of the ends.
// Data choice: a named colour shows the hex conversion runs.
func TestCableRequestCarriesOptionalFields(t *testing.T) {
	length := 3.0
	a, b := requestEnds()
	cable := &devicetypes.CaniCableType{
		Label: "spine-to-leaf", TerminationAType: "dcim.frontport", TerminationBType: "dcim.rearport",
		Length: &length, LengthUnit: "ft", Color: "green",
	}

	req, err := cableRequest(cable, a, b, uuid.New())

	if err != nil {
		t.Fatalf("cableRequest() error = %v", err)
	}
	if *req.TerminationAType != "dcim.frontport" || *req.TerminationBType != "dcim.rearport" {
		t.Fatalf("termination types = %s/%s, want the cable's own", *req.TerminationAType, *req.TerminationBType)
	}
	if uuid.UUID(*req.TerminationAId) != a.iface.ID || uuid.UUID(*req.TerminationBId) != b.iface.ID {
		t.Fatal("termination IDs must be the resolved interfaces")
	}
	if *req.Length != 3 || *req.LengthUnit != nautobotapi.LengthUnitEnumFt {
		t.Fatalf("length = %d %v, want 3 ft", *req.Length, *req.LengthUnit)
	}
	if *req.Color != colorNameToHex("green") {
		t.Fatalf("color = %s, want the hex of green", *req.Color)
	}
}

// TestCableRequestDefaultsBareCable verifies that a cable with no optional
// fields gets interface terminations and sends no length, unit or colour.
//
// Why it matters: an empty string or zero sent for these would be an invalid
// choice or a wrong value in Nautobot, not an omission.
// Inputs: a cable with only a label. Outputs: dcim.interface on both ends,
// nil length, unit and colour.
// Data choice: the label is the one field a cable always has.
func TestCableRequestDefaultsBareCable(t *testing.T) {
	a, b := requestEnds()

	req, err := cableRequest(&devicetypes.CaniCableType{Label: "bare"}, a, b, uuid.New())

	if err != nil {
		t.Fatalf("cableRequest() error = %v", err)
	}
	if *req.TerminationAType != "dcim.interface" || *req.TerminationBType != "dcim.interface" {
		t.Fatalf("termination types = %s/%s, want dcim.interface", *req.TerminationAType, *req.TerminationBType)
	}
	if req.Length != nil || req.LengthUnit != nil || req.Color != nil {
		t.Fatalf("a bare cable must not send length, unit or colour: %v %v %v", req.Length, req.LengthUnit, req.Color)
	}
}

// TestLengthUnitDropsUnknownUnits verifies the unit mapping: Nautobot's four
// units map and anything else, the empty string included, stays unset.
//
// Why it matters: an unknown unit sent as a choice fails the whole create.
// Inputs: m, cm, ft, in, "", "furlong". Outputs: the four enum values, then
// nil twice.
// Data choice: "furlong" is a unit, just not one Nautobot has.
func TestLengthUnitDropsUnknownUnits(t *testing.T) {
	want := map[string]nautobotapi.LengthUnitEnum{
		"m": nautobotapi.LengthUnitEnumM, "cm": nautobotapi.LengthUnitEnumCm,
		"ft": nautobotapi.LengthUnitEnumFt, "in": nautobotapi.LengthUnitEnumIn,
	}
	for unit, enum := range want {
		if got := lengthUnit(unit); got == nil || *got != enum {
			t.Fatalf("lengthUnit(%q) = %v, want %v", unit, got, enum)
		}
	}
	for _, unit := range []string{"", "furlong"} {
		if got := lengthUnit(unit); got != nil {
			t.Fatalf("lengthUnit(%q) = %v, want nil", unit, *got)
		}
	}
}

// TestCableEndsResolveModuleTerminations verifies that a termination naming a
// module resolves to the module's parent device, and that a termination
// naming nothing known is an error that names both ends.
//
// Why it matters: cables on module ports (NIC, HSN) are stored against the
// module in the inventory but belong to the parent device in Nautobot.
// Inputs: a device, a module in it, a cable from the module to the device;
// then a cable from the device to a random ID.
// Outputs: both ends of the first resolve to the device; the second errors.
// Data choice: the same device on both ends keeps the fixture to one device.
func TestCableEndsResolveModuleTerminations(t *testing.T) {
	inv := devicetypes.NewInventory()
	devID, modID := uuid.New(), uuid.New()
	inv.Devices[devID] = &devicetypes.CaniDeviceType{ID: devID, Name: "node-01"}
	inv.Modules[modID] = &devicetypes.CaniModuleType{ID: modID, Name: "nic-1", ParentDevice: devID}

	a, b, err := cableEnds(inv, &devicetypes.CaniCableType{TerminationADevice: modID, TerminationAPort: "p1", TerminationBDevice: devID, TerminationBPort: "eth0"})

	if err != nil || a.device != inv.Devices[devID] || b.device != inv.Devices[devID] || a.port != "p1" || b.port != "eth0" {
		t.Fatalf("ends = %+v %+v err = %v, want both on node-01", a, b, err)
	}
	if _, _, err := cableEnds(inv, &devicetypes.CaniCableType{TerminationADevice: devID, TerminationBDevice: uuid.New()}); err == nil {
		t.Fatal("an unknown termination must be an error")
	}
}
