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

package add

import (
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// campusInventory returns an inventory whose rack-01 stands in a room under
// a campus, next to an annex that holds neither.
func campusInventory() *devicetypes.Inventory {
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	campus, annex := uuid.New(), uuid.New()
	for _, room := range inventory.Locations {
		room.Name, room.Parent = "room", campus
	}
	inventory.Locations[campus] = &devicetypes.CaniLocationType{ID: campus, Name: "campus", LocationType: "site"}
	inventory.Locations[annex] = &devicetypes.CaniLocationType{ID: annex, Name: "annex", LocationType: "site"}
	return inventory
}

// addDeviceAndReload runs `add device` with the given arguments and returns
// the inventory as the next command would load it.
func addDeviceAndReload(t *testing.T, inventory *devicetypes.Inventory, argv ...string) (*devicetypes.Inventory, error) {
	t.Helper()
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)
	runErr := harness.Execute(t, append([]string{"add", "device", cmdtest.DeviceSlug}, argv...)...)
	reloaded, err := harness.Store.Load()
	if err != nil {
		t.Fatalf("reload inventory: %v", err)
	}
	return reloaded, runErr
}

// locationNameOf returns the name of the location the named device is in.
func locationNameOf(t *testing.T, inventory *devicetypes.Inventory, device string) string {
	t.Helper()
	location := inventory.Locations[deviceNamed(t, inventory, device).Location]
	if location == nil {
		return ""
	}
	return location.Name
}

// TestAddDeviceLocationPlacesTheDevice verifies what --location does on an
// add that names its rack or has none: a racked device keeps a location that
// holds its rack, a device with no rack is placed in the location, and a
// location that does not hold the rack is refused.
//
// Why it matters: --location was read only to filter racks for a placement
// strategy and ignored otherwise, so the one way to place a device was a
// provider-metadata key that a single exporter understood.
// Inputs: rack-01 in a room under a campus, and an unrelated annex; a device
// added to the rack with --location campus, one with --location annex, one
// with no location, and one with no rack and --location annex.
// Outputs: campus; a refusal naming the annex and nothing added; the room;
// the annex with no parent.
// Data choice: the annex is a real location, so the refusal is about holding
// the rack and the acceptance for the rackless device is the same value.
func TestAddDeviceLocationPlacesTheDevice(t *testing.T) {
	// Arrange.
	inventory := campusInventory()

	// Act.
	inventory, heldErr := addDeviceAndReload(t, inventory, "--rack", "rack-01", "--position", "10", "--name", "in-campus", "--location", "campus")
	inventory, strayErr := addDeviceAndReload(t, inventory, "--rack", "rack-01", "--position", "14", "--name", "in-annex", "--location", "annex")
	inventory, derivedErr := addDeviceAndReload(t, inventory, "--rack", "rack-01", "--position", "18", "--name", "in-room")
	inventory, looseErr := addDeviceAndReload(t, inventory, "--name", "loose", "--location", "annex")

	// Assert.
	if heldErr != nil || locationNameOf(t, inventory, "in-campus") != "campus" {
		t.Errorf("--location campus: error = %v, location = %q; want campus", heldErr, locationNameOf(t, inventory, "in-campus"))
	}
	if strayErr == nil || !strings.Contains(strayErr.Error(), "annex") {
		t.Errorf("--location annex on a racked device: error = %v, want a refusal naming the annex", strayErr)
	}
	for _, device := range inventory.Devices {
		if device.Name == "in-annex" {
			t.Error("the refused device was added")
		}
	}
	if derivedErr != nil || locationNameOf(t, inventory, "in-room") != "room" {
		t.Errorf("no --location: error = %v, location = %q; want the rack's room", derivedErr, locationNameOf(t, inventory, "in-room"))
	}
	loose := deviceNamed(t, inventory, "loose")
	if looseErr != nil || loose.Parent != uuid.Nil || locationNameOf(t, inventory, "loose") != "annex" {
		t.Errorf("rackless device: error = %v, parent = %s, location = %q; want no parent and the annex",
			looseErr, loose.Parent, locationNameOf(t, inventory, "loose"))
	}
}
