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

package cmdtest_test

import (
	"maps"
	"testing"

	"github.com/Cray-HPE/cani/cmd/add"
	"github.com/Cray-HPE/cani/cmd/update"
	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// TestMetadataCapabilityReceivesParsedPairs verifies add and update dispatch metadata.
//
// Why it matters: generic flags must reach the provider with the right target.
// Inputs: two metadata pairs and an inventory containing a different device.
// Outputs: one callback with exact parsed values and metadata on the saved target.
// Data choice: an equals sign inside a value detects lossy flag parsing, while
// distinct device names distinguish updating the intended object from another.
func TestMetadataCapabilityReceivesParsedPairs(t *testing.T) {
	cases := []struct {
		name    string
		command func() *cli.Command
		ref     string
		target  string
		flags   map[string][]string
	}{
		{"add", add.NewCommand, cmdtest.DeviceSlug, "cn-02", map[string][]string{"rack": {"rack-01"}, "name": {"cn-02"}}},
		{"update", update.NewCommand, "cn-01", "cn-99", map[string][]string{"name": {"cn-99"}}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			capabilities := useCapabilities(t)
			capabilities.applyMetadata = func(target *map[string]any, values map[string]string) {
				if *target == nil {
					*target = make(map[string]any)
				}
				(*target)[capabilities.Slug()] = maps.Clone(values)
			}
			inventory, rackID := cmdtest.InventoryWithRack("rack-01")
			cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
			parent := testCase.command()
			harness := cmdtest.New(t, parent, subcommand(t, parent, "device"), inventory)
			flags := maps.Clone(testCase.flags)
			flags["metadata"] = []string{"owner=ops", "note=a=b"}

			if err := harness.Run(t, flags, testCase.ref); err != nil {
				t.Fatal(err)
			}

			device := namedDevice(t, harness.Inventory(), testCase.target)
			assertMetadataResult(t, capabilities, device)
		})
	}
}

func assertMetadataResult(t *testing.T, capabilities *capableProvider, device *devicetypes.CaniDeviceType) {
	t.Helper()
	assertCallCount(t, "ApplyMetadata", len(capabilities.calls.metadata), 1)
	call := capabilities.calls.metadata[0]
	want := map[string]string{"owner": "ops", "note": "a=b"}
	if !maps.Equal(call.values, want) {
		t.Errorf("metadata arguments = %v, want %v", call.values, want)
	}
	got, ok := device.ProviderMetadata[capabilities.Slug()].(map[string]string)
	if !ok || !maps.Equal(got, want) {
		t.Errorf("saved target metadata = %v, want %v", got, want)
	}
}

// TestProviderDeviceUpdateFlags verifies contributed flags reach runtime dispatch.
//
// Why it matters: registration alone cannot prove a provider's flags are applied.
// Inputs: a synthetic provider flag and an existing device selected by name.
// Outputs: one registration and one application with the same command and target.
// Data choice: a non-default value distinguishes forwarded input from defaults.
func TestProviderDeviceUpdateFlags(t *testing.T) {
	capabilities := useCapabilities(t)
	inventory, rackID := cmdtest.InventoryWithRack("rack-01")
	deviceID := cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
	parent := update.NewCommand()
	harness := cmdtest.New(t, parent, subcommand(t, parent, "device"), inventory)

	if err := harness.Run(t, map[string][]string{capabilityFlag: {"forwarded"}}, "cn-01"); err != nil {
		t.Fatal(err)
	}

	assertCallCount(t, "RegisterDeviceUpdateFlags", len(capabilities.calls.registration), 1)
	assertCallCount(t, "ApplyDeviceUpdateFlags", len(capabilities.calls.updates), 1)
	call := capabilities.calls.updates[0]
	if call.command != capabilities.calls.registration[0] || call.device != harness.Inventory().Devices[deviceID] {
		t.Error("flag application received a different command or device")
	}
	value, err := call.command.Flags().GetString(capabilityFlag)
	if err != nil || value != "forwarded" || !call.command.Flags().Changed(capabilityFlag) {
		t.Errorf("provider flag = %q, error = %v; want changed value forwarded", value, err)
	}
}

// TestAutoStagingDispatchesCapabilities verifies staging preference and fallback.
//
// Why it matters: the generic add path must use optional provider capabilities
// and describe the actual staged device, not merely register those capabilities.
// Inputs: a new-device staging provider or an existing-device fallback, with --auto.
// Outputs: exact staging calls, one save and a description of the saved device.
// Data choice: two devices at separate rack positions distinguish new creation
// from restaging and prove a successful rack stager short-circuits the fallback.
func TestAutoStagingDispatchesCapabilities(t *testing.T) {
	cases := []struct {
		name      string
		createNew bool
	}{
		{"new in rack", true},
		{"existing fallback", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			capabilities := useCapabilities(t)
			inventory, rackID := cmdtest.InventoryWithRack("rack-01")
			existingID := cmdtest.AddDevice(inventory, rackID, "cn-01", 10)
			capabilities.stageNew = func(inventory *devicetypes.Inventory, _ string) bool {
				if !testCase.createNew {
					return false
				}
				deviceID := cmdtest.AddDevice(inventory, rackID, "staged-new", 12)
				inventory.Devices[deviceID].Status = string(devicetypes.StatusStaged)
				return true
			}
			capabilities.stageExisting = func(inventory *devicetypes.Inventory, _ string) bool {
				inventory.Devices[existingID].Status = string(devicetypes.StatusStaged)
				return true
			}
			command := add.NewCommand()
			command.Flags().AddFlagSet(command.PersistentFlags())
			harness := cmdtest.New(t, &cli.Command{Use: "alpha"}, command, inventory)

			if err := harness.Run(t, map[string][]string{"auto": {"true"}, "accept": {"true"}}, cmdtest.DeviceSlug); err != nil {
				t.Fatal(err)
			}

			assertStagingResult(t, capabilities.calls, harness, testCase.createNew)
		})
	}
}

func assertStagingResult(t *testing.T, calls capabilityLog, harness *cmdtest.Harness, createNew bool) {
	t.Helper()
	wantExisting, target := 1, "cn-01"
	if createNew {
		wantExisting, target = 0, "staged-new"
	}
	assertCallCount(t, "StageNewInRack", len(calls.newStages), 1)
	assertCallCount(t, "StageExisting", len(calls.existing), wantExisting)
	assertCallCount(t, "DescribeStagedDevice", len(calls.descriptions), 1)
	assertStageArguments(t, calls.newStages, harness.Inventory())
	assertStageArguments(t, calls.existing, harness.Inventory())
	device := namedDevice(t, harness.Inventory(), target)
	if calls.descriptions[0] != device || device.Status != string(devicetypes.StatusStaged) {
		t.Error("description did not receive the saved, staged device")
	}
	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}
}

func assertStageArguments(t *testing.T, calls []stageCall, inventory *devicetypes.Inventory) {
	t.Helper()
	for _, call := range calls {
		if call.inventory != inventory || call.slug != cmdtest.DeviceSlug {
			t.Errorf("staging arguments = %p, %q; want saved inventory %p, %q", call.inventory, call.slug, inventory, cmdtest.DeviceSlug)
		}
	}
}

func namedDevice(t *testing.T, inventory *devicetypes.Inventory, name string) *devicetypes.CaniDeviceType {
	t.Helper()
	for _, device := range inventory.Devices {
		if device.Name == name {
			return device
		}
	}
	t.Fatalf("device %q not found", name)
	return nil
}
