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
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// deviceBayFromJSON decodes a Nautobot device bay the way the API returns it.
func deviceBayFromJSON(t *testing.T, name string, parentNB uuid.UUID, installedNB *uuid.UUID) nautobotapi.DeviceBay {
	t.Helper()
	raw := map[string]any{"name": name, "device": map[string]any{"id": parentNB.String()}}
	if installedNB != nil {
		raw["installed_device"] = map[string]any{"id": installedNB.String()}
	}
	data, _ := json.Marshal(raw)
	var bay nautobotapi.DeviceBay
	if err := json.Unmarshal(data, &bay); err != nil {
		t.Fatalf("decode device bay: %v", err)
	}
	return bay
}

// TestApplyDeviceBays verifies installed children are re-parented to their
// chassis with the bay name, and that empty or unknown bays are ignored.
//
// Why it matters: without this, imported half-width switches sit directly in
// the rack and re-exporting them fails Nautobot's child-device validation.
// Inputs: a chassis and switch (the switch imported with a rack parent and a
// stale U position), one occupied bay, one empty bay, and one bay whose
// device was not imported. Outputs: one placement; switch parent = chassis,
// ParentBay "Bay 2", no position or face.
func TestApplyDeviceBays(t *testing.T) {
	chassisNB, switchNB, missingNB := uuid.New(), uuid.New(), uuid.New()
	chassis := &devicetypes.CaniDeviceType{ID: uuid.New(), Name: "chassis-15"}
	sw := &devicetypes.CaniDeviceType{ID: uuid.New(), Name: "leaf-002", Parent: uuid.New(), RackPosition: 15, Face: "front"}
	devices := map[uuid.UUID]*devicetypes.CaniDeviceType{chassis.ID: chassis, sw.ID: sw}
	deviceMap := map[uuid.UUID]uuid.UUID{chassisNB: chassis.ID, switchNB: sw.ID}

	bays := []nautobotapi.DeviceBay{
		deviceBayFromJSON(t, "Bay 1", chassisNB, nil),
		deviceBayFromJSON(t, "Bay 2", chassisNB, &switchNB),
		deviceBayFromJSON(t, "Bay 1", missingNB, &switchNB),
	}
	if placed := ApplyDeviceBays(bays, devices, deviceMap); placed != 1 {
		t.Fatalf("placed = %d, want 1", placed)
	}
	if sw.Parent != chassis.ID || sw.ParentBay != "Bay 2" || sw.RackPosition != 0 || sw.Face != "" {
		t.Errorf("switch = parent %s bay %q pos %d face %q; want chassis, Bay 2, 0, empty",
			sw.Parent, sw.ParentBay, sw.RackPosition, sw.Face)
	}
}
