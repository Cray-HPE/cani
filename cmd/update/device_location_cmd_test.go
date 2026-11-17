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

package update

import (
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// campusInventory returns an inventory with a campus that holds two rooms,
// rack-01 in room-1 with the device dev-00 in it, an empty rack-02 in
// room-2, and an annex outside the campus.
func campusInventory() *devicetypes.Inventory {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	campus, room2, annex, rack2 := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, room := range inventory.Locations {
		room.Name, room.Parent = "room-1", campus
	}
	inventory.Locations[campus] = &devicetypes.CaniLocationType{ID: campus, Name: "campus", LocationType: "site"}
	inventory.Locations[room2] = &devicetypes.CaniLocationType{ID: room2, Name: "room-2", LocationType: "site", Parent: campus}
	inventory.Locations[annex] = &devicetypes.CaniLocationType{ID: annex, Name: "annex", LocationType: "site"}
	inventory.Racks[rack2] = &devicetypes.CaniRackType{ID: rack2, Name: "rack-02", UHeight: 42, Location: room2, Type: devicetypes.TypeRack}
	cmdtest.AddDevice(inventory, rackID, "dev-00", 1)
	return inventory
}

// updateAndReload runs `update <argv>` and returns the inventory as the next
// command would load it.
func updateAndReload(t *testing.T, inventory *devicetypes.Inventory, argv ...string) (*devicetypes.Inventory, error) {
	t.Helper()
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)
	runErr := harness.Execute(t, append([]string{"update"}, argv...)...)
	reloaded, err := harness.Store.Load()
	if err != nil {
		t.Fatalf("reload inventory: %v", err)
	}
	return reloaded, runErr
}

// deviceLocationName returns the name of the location dev-00 is in.
func deviceLocationName(t *testing.T, inventory *devicetypes.Inventory) string {
	t.Helper()
	for _, device := range inventory.Devices {
		if device.Name != "dev-00" {
			continue
		}
		if location := inventory.Locations[device.Location]; location != nil {
			return location.Name
		}
		return ""
	}
	t.Fatal("dev-00 not found")
	return ""
}

// TestUpdateDeviceLocationHoldsUntilTheDeviceOrRackMoves verifies the life
// of a location a racked device records: `update device --location` sets one
// that holds the rack and refuses one that does not, and the device follows
// its rack again once it is moved to another rack or the rack itself moves.
//
// Why it matters: the campus still holds every room in it, so a location
// recorded there would outlive a move unless the move clears it; Nautobot
// puts the devices of a rack in the rack's location whenever the rack is
// saved, and an operator who moves a rack expects its devices to go too.
// Inputs: dev-00 in rack-01 in room-1 under a campus. In order: --location
// campus, --location annex, --parent rack-02, --location campus again, then
// `update rack rack-02 --location room-1`.
// Outputs: campus; a refusal naming the annex with campus kept; room-2;
// campus; room-1.
// Data choice: both rooms are under the campus, the case in which a
// recorded location would survive a move by still holding the new rack.
func TestUpdateDeviceLocationHoldsUntilTheDeviceOrRackMoves(t *testing.T) {
	// Arrange.
	inventory := campusInventory()

	// Act and assert, one command at a time.
	inventory, err := updateAndReload(t, inventory, "device", "dev-00", "--location", "campus")
	if got := deviceLocationName(t, inventory); err != nil || got != "campus" {
		t.Fatalf("--location campus: error = %v, location = %q; want campus", err, got)
	}

	inventory, err = updateAndReload(t, inventory, "device", "dev-00", "--location", "annex")
	if got := deviceLocationName(t, inventory); err == nil || !strings.Contains(err.Error(), "annex") || got != "campus" {
		t.Errorf("--location annex: error = %v, location = %q; want a refusal naming the annex and campus kept", err, got)
	}

	inventory, err = updateAndReload(t, inventory, "device", "dev-00", "--parent", "rack-02")
	if got := deviceLocationName(t, inventory); err != nil || got != "room-2" {
		t.Errorf("--parent rack-02: error = %v, location = %q; want the new rack's room-2", err, got)
	}

	inventory, err = updateAndReload(t, inventory, "device", "dev-00", "--location", "campus")
	if got := deviceLocationName(t, inventory); err != nil || got != "campus" {
		t.Fatalf("--location campus in rack-02: error = %v, location = %q; want campus", err, got)
	}

	inventory, err = updateAndReload(t, inventory, "rack", "rack-02", "--location", "room-1")
	if got := deviceLocationName(t, inventory); err != nil || got != "room-1" {
		t.Errorf("update rack --location room-1: error = %v, device location = %q; want the device moved with its rack", err, got)
	}
}
