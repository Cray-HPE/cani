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
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// TestAddStampsSerialAndAssetTag verifies the persistent --serial and
// --asset-tag flags reach a device added by placement and a module added
// into a bay, and that an unset flag leaves the field empty.
//
// Why it matters: the flags are declared once on `add` and applied by every
// noun; a noun that reads one flag but not the other silently drops the
// asset tag an operator typed, which is the identity a lab tracks spares by.
// Inputs: `add device <slug> --serial SN-1 --asset-tag AT-1` into rack-01,
// then `add module <slug> --device cn-01 --bay PCIe1 --asset-tag AT-2`.
// Outputs: the device records both values; the module records the asset tag
// and no serial.
// Data choice: the module omits --serial to prove the two flags are applied
// independently rather than as a pair.
func TestAddStampsSerialAndAssetTag(t *testing.T) {
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)
	err := harness.Run(t, map[string][]string{
		"rack": {"rack-01"}, "position": {"10"}, "face": {"front"}, "name": {"cn-01"},
		flagSerial: {"SN-1"}, flagAssetTag: {"AT-1"},
	}, cmdtest.DeviceSlug)
	if err != nil {
		t.Fatalf("add device: unexpected error: %v", err)
	}
	device := deviceNamed(t, harness.Inventory(), "cn-01")
	if device.Serial != "SN-1" || device.AssetTag != "AT-1" {
		t.Errorf("device identity = {%q %q}, want {SN-1 AT-1}", device.Serial, device.AssetTag)
	}
	device.ModuleBays = []devicetypes.ModuleBaySpec{{Name: "PCIe1"}}

	harness = cmdtest.New(t, NewCommand(), newModuleCommand(), harness.Inventory())
	err = harness.Run(t, map[string][]string{
		"device": {"cn-01"}, "bay": {"PCIe1"}, "name": {"nic-01"}, flagAssetTag: {"AT-2"},
	}, "hpe-2p-25gbe-pcie-nic")
	if err != nil {
		t.Fatalf("add module: unexpected error: %v", err)
	}
	for _, mod := range harness.Inventory().Modules {
		if mod.Name != "nic-01" {
			continue
		}
		if mod.Serial != "" || mod.AssetTag != "AT-2" {
			t.Errorf("module identity = {%q %q}, want {\"\" AT-2}", mod.Serial, mod.AssetTag)
		}
		return
	}
	t.Fatal("module nic-01 not found")
}

// deviceNamed returns the named device from the inventory.
func deviceNamed(t *testing.T, inventory *devicetypes.Inventory, name string) *devicetypes.CaniDeviceType {
	t.Helper()
	for _, device := range inventory.Devices {
		if device.Name == name {
			return device
		}
	}
	t.Fatalf("device %q not found", name)
	return nil
}
