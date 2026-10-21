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
package export

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// childDeviceInventory returns an inventory holding a rack, a chassis device
// in that rack, and the given device whose parent is the chassis.
func childDeviceInventory(child *devicetypes.CaniDeviceType) *devicetypes.Inventory {
	rackID, chassisID := uuid.New(), uuid.New()
	child.Parent = chassisID
	return &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{rackID: {ID: rackID, Name: "R0179"}},
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			chassisID: {ID: chassisID, Name: "chassis-15", Parent: rackID, RackPosition: 15},
			child.ID:  child,
		},
	}
}

// TestIsChildDevice verifies child detection from both the subdevice role and
// a device (non-rack) parent.
//
// Why it matters: Nautobot rejects rack position and face on child devices, so
// misclassifying one fails the export of the device and all of its cables.
// Inputs: a rack-mounted device, a device inside a chassis, and a child-role
// device with no inventory. Outputs: false, true, true.
func TestIsChildDevice(t *testing.T) {
	rackID := uuid.New()
	racked := &devicetypes.CaniDeviceType{ID: uuid.New(), Parent: rackID}
	inv := &devicetypes.Inventory{
		Racks:   map[uuid.UUID]*devicetypes.CaniRackType{rackID: {ID: rackID}},
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{racked.ID: racked},
	}
	if isChildDevice(racked, inv) {
		t.Error("rack-mounted device reported as child")
	}

	inChassis := &devicetypes.CaniDeviceType{ID: uuid.New()}
	if !isChildDevice(inChassis, childDeviceInventory(inChassis)) {
		t.Error("device inside a chassis not reported as child")
	}

	childRole := &devicetypes.CaniDeviceType{SubdeviceRole: "child"}
	if !isChildDevice(childRole, nil) {
		t.Error("subdevice_role child not reported as child")
	}
}

// TestSetRackSlot verifies position and face are set for rack-mounted devices
// and left unset for child devices.
//
// Why it matters: this is the fix for Nautobot's "Child device types cannot be
// assigned to a rack face" error. Inputs: a racked device at U15 rear and a
// child carrying the same values. Outputs: position/face set only for the
// racked device.
func TestSetRackSlot(t *testing.T) {
	rackID := uuid.New()
	racked := &devicetypes.CaniDeviceType{ID: uuid.New(), Parent: rackID, RackPosition: 15, Face: "rear"}
	inv := &devicetypes.Inventory{
		Racks:   map[uuid.UUID]*devicetypes.CaniRackType{rackID: {ID: rackID}},
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{racked.ID: racked},
	}
	req := &nautobotapi.WritableDeviceRequest{}
	setRackSlot(racked, inv, &req.Position, &req.Face)
	if req.Position == nil || *req.Position != 15 || req.Face == nil {
		t.Errorf("racked device: position=%v face=%v, want 15 and set", req.Position, req.Face)
	}

	child := &devicetypes.CaniDeviceType{ID: uuid.New(), RackPosition: 15, Face: "front"}
	childReq := &nautobotapi.PatchedWritableDeviceRequest{}
	setRackSlot(child, childDeviceInventory(child), &childReq.Position, &childReq.Face)
	if childReq.Position != nil || childReq.Face != nil {
		t.Errorf("child device: position=%v face=%v, want both unset", childReq.Position, childReq.Face)
	}
}

// TestCompareDeviceFieldsSkipsChildPlacement verifies the export diff never
// reports position or face for a child device, while a rack-mounted device
// with the same values still does.
//
// Why it matters: Nautobot stores no position or face for child devices, so
// comparing them would flag every child as changed on every export. Inputs:
// a child in a chassis and a racked device, both at U15 front, against a
// remote device with neither field. Outputs: no placement diffs for the
// child; position and face diffs for the racked device.
func TestCompareDeviceFieldsSkipsChildPlacement(t *testing.T) {
	e, done := newExporterWithServer(t, interfaceServer(&capturedRequest{}, nil))
	defer done()
	mapper := NewDeviceMapper(e.Cache, &MapperOpts{})

	child := &devicetypes.CaniDeviceType{ID: uuid.New(), RackPosition: 15, Face: "front"}
	inv := childDeviceInventory(child)
	rackID := uuid.Nil
	for id := range inv.Racks {
		rackID = id
	}
	racked := &devicetypes.CaniDeviceType{ID: uuid.New(), Parent: rackID, RackPosition: 15, Face: "front"}
	inv.Devices[racked.ID] = racked
	mapper.SetInventory(inv)

	placement := func(dev *devicetypes.CaniDeviceType) map[string]bool {
		got := map[string]bool{}
		for _, d := range compareDeviceFields(dev, &nautobotapi.Device{}, mapper) {
			if d.Field == "position" || d.Field == "face" {
				got[d.Field] = true
			}
		}
		return got
	}
	if got := placement(child); len(got) != 0 {
		t.Errorf("child placement diffs = %v, want none", got)
	}
	if got := placement(racked); !got["position"] || !got["face"] {
		t.Errorf("racked placement diffs = %v, want position and face", got)
	}
}
