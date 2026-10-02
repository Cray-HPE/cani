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
	"encoding/json"
	"slices"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// moduleBayFromJSON decodes a Nautobot module bay the way the API returns it;
// a nil parent decodes a bay nested in a module.
func moduleBayFromJSON(t *testing.T, name, position string, parentDevice *uuid.UUID) nautobotapi.ModuleBay {
	t.Helper()
	raw := map[string]any{"name": name, "position": position}
	if parentDevice != nil {
		raw["parent_device"] = map[string]any{"id": parentDevice.String()}
	}
	data, _ := json.Marshal(raw)
	var bay nautobotapi.ModuleBay
	if err := json.Unmarshal(data, &bay); err != nil {
		t.Fatalf("decode module bay: %v", err)
	}
	return bay
}

// TestApplyModuleBaysRecordsNautobotBaysOnDevices verifies each device
// receives the bays Nautobot declares for it, sorted by name with their
// positions, that a bay nested in a module or on an unknown device is
// ignored, and that applying the same bays twice records nothing new.
//
// Why it matters: the merge resolves a module's bay against the parent's
// declared bays; when the instance declares none, the library template is
// consulted and a bay Nautobot spells differently is cleared, after which
// the module can neither keep nor restore its bay.
// Inputs: one imported server, two of its bays listed out of name order (one
// with a position), one bay nested in a module, one bay on a device that was
// not imported. Outputs: two bays recorded as "PCIe1"/"Slot B" with the
// position kept, zero on the second application.
// Data choice: "Slot B" is a name the embedded library does not declare for
// any device, which is the case the library fallback gets wrong.
func TestApplyModuleBaysRecordsNautobotBaysOnDevices(t *testing.T) {
	serverNB, strangerNB := uuid.New(), uuid.New()
	server := &devicetypes.CaniDeviceType{ID: uuid.New(), Name: "srv-01"}
	devices := map[uuid.UUID]*devicetypes.CaniDeviceType{server.ID: server}
	deviceMap := map[uuid.UUID]uuid.UUID{serverNB: server.ID}
	bays := []nautobotapi.ModuleBay{
		moduleBayFromJSON(t, "Slot B", "2", &serverNB),
		moduleBayFromJSON(t, "PCIe1", "", &serverNB),
		moduleBayFromJSON(t, "nested", "", nil),
		moduleBayFromJSON(t, "Slot A", "", &strangerNB),
	}

	recorded := ApplyModuleBays(bays, devices, deviceMap)

	want := []devicetypes.ModuleBaySpec{{Name: "PCIe1"}, {Name: "Slot B", Position: "2"}}
	if recorded != 2 || !slices.Equal(server.ModuleBays, want) {
		t.Fatalf("recorded = %d, bays = %+v; want 2 and %+v", recorded, server.ModuleBays, want)
	}
	if again := ApplyModuleBays(bays, devices, deviceMap); again != 0 || len(server.ModuleBays) != 2 {
		t.Errorf("second application recorded %d (bays now %d), want 0 and 2", again, len(server.ModuleBays))
	}
}
