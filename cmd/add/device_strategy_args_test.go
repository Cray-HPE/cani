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
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/google/uuid"
)

func TestAddDeviceStrategyRejectsNamedBay(t *testing.T) {
	for _, strategy := range []string{"%{FILL}", "%{SPREAD}"} {
		t.Run(strategy, func(t *testing.T) {
			assertStrategyArgsRejected(t, strategy, []string{"--bay", "Bay 1"}, "--bay cannot be combined")
		})
	}
}

func TestAddDeviceStrategyRejectsParent(t *testing.T) {
	for _, strategy := range []string{"%{FILL}", "%{SPREAD}"} {
		t.Run(strategy, func(t *testing.T) {
			assertStrategyArgsRejected(t, strategy, []string{"--parent", "rack-01"}, "--parent cannot be combined")
		})
	}
}

func TestAddDeviceStrategyRejectsParentAndBay(t *testing.T) {
	for _, strategy := range []string{"%{FILL}", "%{SPREAD}"} {
		t.Run(strategy, func(t *testing.T) {
			assertStrategyArgsRejected(t, strategy,
				[]string{"--parent", "chassis-15", "--bay", "Bay 1"}, "--parent cannot be combined")
		})
	}
}

func assertStrategyArgsRejected(t *testing.T, strategy string, flags []string, want string) {
	t.Helper()
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddChassisWithBays(inventory, rackID, "chassis-15", 15, "Bay 1", "Bay 2")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)
	before := len(inventory.Devices)
	args := append([]string{"add", "device", cmdtest.DeviceSlug, "--rack", strategy}, flags...)
	err := harness.Execute(t, args...)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("error=%v, want %q", err, want)
	}
	if harness.Store.Loads != 0 || harness.Store.Saves != 0 {
		t.Errorf("invalid strategy accessed datastore: loads=%d, saves=%d", harness.Store.Loads, harness.Store.Saves)
	}
	if len(harness.Inventory().Devices) != before {
		t.Error("invalid strategy changed inventory")
	}
}

func TestAddDeviceStrategyAcceptsAbsentParentAndBay(t *testing.T) {
	for _, strategy := range []string{"%{FILL}", "%{SPREAD}"} {
		for _, tc := range []struct {
			name  string
			flags []string
		}{
			{"defaults", nil},
			{"nil parent", []string{"--parent", uuid.Nil.String()}},
			{"empty parent", []string{"--parent", ""}},
			{"empty bay", []string{"--bay", ""}},
		} {
			t.Run(strategy+"/"+tc.name, func(t *testing.T) {
				assertRackStrategyPlacement(t, strategy, tc.flags)
			})
		}
	}
}

func assertRackStrategyPlacement(t *testing.T, strategy string, flags []string) {
	t.Helper()
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)
	args := append([]string{"add", "device", cmdtest.DeviceSlug, "--rack", strategy}, flags...)
	if err := harness.Execute(t, args...); err != nil {
		t.Fatalf("valid strategy rejected: %v", err)
	}
	if harness.Store.Saves != 1 || len(harness.Inventory().Devices) != 1 {
		t.Fatalf("saves=%d, devices=%d; want one device saved once", harness.Store.Saves, len(harness.Inventory().Devices))
	}
	for _, device := range harness.Inventory().Devices {
		if device.Parent != rackID || device.ParentBay != "" || device.RackPosition <= 0 {
			t.Fatalf("unexpected rack strategy placement: %+v", device)
		}
	}
}

func TestAddDeviceLiteralRackKeepsParentPrecedence(t *testing.T) {
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	cmdtest.AddChassisWithBays(inventory, rackID, "chassis-15", 15, "Bay 1")
	harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), inventory)
	err := harness.Execute(t, "add", "device", cmdtest.DeviceSlug,
		"--rack", "rack-01", "--parent", "chassis-15", "--position", "20", "--name", "rack-device")
	if err != nil {
		t.Fatalf("literal precedence changed: %v", err)
	}
	device := childrenOf(harness.Inventory(), rackID)["rack-device"]
	if device == nil || device.ParentBay != "" || device.RackPosition != 20 {
		t.Fatalf("literal rack must retain precedence: %+v", device)
	}
}

func TestAddDeviceListModeIgnoresPlacementFlags(t *testing.T) {
	const childEnv = "CANI_TEST_DEVICE_LIST_STRATEGY"
	if os.Getenv(childEnv) == "1" {
		harness := cmdtest.New(t, NewCommand(), newDeviceCommand(), nil)
		err := harness.Execute(t, "add", "device", "--list-supported-types",
			"--rack", "%{FILL}", "--parent", "chassis-15", "--bay", "Bay 1")
		if err != nil {
			t.Fatal(err)
		}
		if harness.Store.Loads != 0 || harness.Store.Saves != 0 {
			t.Fatal("type listing accessed datastore")
		}
		return
	}
	// Type listing exits the process, so isolate it from the test runner.
	process := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestAddDeviceListModeIgnoresPlacementFlags$")
	process.Env = append(os.Environ(), childEnv+"=1")
	output, err := process.CombinedOutput()
	if err != nil {
		t.Fatalf("type listing failed: %v\n%s", err, output)
	}
	if !bytes.Contains(output, []byte(cmdtest.DeviceSlug)) {
		t.Fatal("type listing did not print supported device types")
	}
}
