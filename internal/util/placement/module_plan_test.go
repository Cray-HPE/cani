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
package placement

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// +---------------------------------------+--------------------------------------------------+--------------------------------------------------+
// | Function                              | Happy-path test                                  | Failure test                                     |
// +---------------------------------------+--------------------------------------------------+--------------------------------------------------+
// | PlanModules (FILL)                    | TestPlanModulesFillSingleDevice                  | TestPlanModulesNoDevices                         |
// | PlanModules (FILL)                    | TestPlanModulesFillMultiDevice                   | TestPlanModulesNotEnoughBays                     |
// | BayFilterForHardwareType              | TestBayFilterGPU                                 | TestBayFilterUnknown                             |
// | PrintModulePlan                       | TestPrintModulePlan                              |                                                  |
// +---------------------------------------+--------------------------------------------------+--------------------------------------------------+

func makeTestDevice(name, slug string, bays []devicetypes.ModuleBaySpec) *devicetypes.CaniDeviceType {
	return &devicetypes.CaniDeviceType{
		ID:         uuid.New(),
		Name:       name,
		Slug:       slug,
		ModuleBays: bays,
	}
}

func makeTestInventory(devices []*devicetypes.CaniDeviceType) *devicetypes.Inventory {
	inv := devicetypes.NewInventory()
	for _, d := range devices {
		inv.Devices[d.ID] = d
	}
	return inv
}

func gpuBays(count int) []devicetypes.ModuleBaySpec {
	bays := make([]devicetypes.ModuleBaySpec, count)
	for i := range count {
		bays[i] = devicetypes.ModuleBaySpec{
			Name:     "GPU " + strings.Repeat("", 0) + string(rune('0'+i)),
			Position: "GPU" + string(rune('0'+i)),
		}
	}
	return bays
}

func TestPlanModulesFillSingleDevice(t *testing.T) {
	dev := makeTestDevice("dev-a", "hpe-xd670", gpuBays(4))
	inv := makeTestInventory([]*devicetypes.CaniDeviceType{dev})

	entries, err := PlanModules([]*devicetypes.CaniDeviceType{dev}, inv, "gpu", 3, StrategyFill)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.DeviceID != dev.ID {
			t.Errorf("expected device %s, got %s", dev.ID, e.DeviceID)
		}
	}
}

func TestPlanModulesFillMultiDevice(t *testing.T) {
	devA := makeTestDevice("dev-a", "hpe-xd670", gpuBays(2))
	devB := makeTestDevice("dev-b", "hpe-xd670", gpuBays(2))
	inv := makeTestInventory([]*devicetypes.CaniDeviceType{devA, devB})

	entries, err := PlanModules(
		[]*devicetypes.CaniDeviceType{devA, devB}, inv, "gpu", 3, StrategyFill)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	// First two should be in dev-a, third in dev-b.
	if entries[0].DeviceName != "dev-a" || entries[1].DeviceName != "dev-a" {
		t.Error("expected first two entries in dev-a")
	}
	if entries[2].DeviceName != "dev-b" {
		t.Error("expected third entry in dev-b")
	}
}

func TestPlanModulesFillSkipsOccupied(t *testing.T) {
	dev := makeTestDevice("dev-a", "hpe-xd670", gpuBays(3))
	inv := makeTestInventory([]*devicetypes.CaniDeviceType{dev})

	// Occupy bay "GPU 0".
	modID := uuid.New()
	inv.Modules[modID] = &devicetypes.CaniModuleType{
		ID: modID, ParentDevice: dev.ID, ModuleBayName: "GPU 0",
	}

	entries, err := PlanModules([]*devicetypes.CaniDeviceType{dev}, inv, "gpu", 2, StrategyFill)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].BayName == "GPU 0" {
		t.Error("expected GPU 0 to be skipped (occupied)")
	}
}

func TestPlanModulesNoDevices(t *testing.T) {
	inv := devicetypes.NewInventory()
	_, err := PlanModules(nil, inv, "", 1, StrategyFill)
	if err == nil {
		t.Fatal("expected error for no devices")
	}
}

func TestPlanModulesNotEnoughBays(t *testing.T) {
	dev := makeTestDevice("dev-a", "hpe-xd670", gpuBays(2))
	inv := makeTestInventory([]*devicetypes.CaniDeviceType{dev})

	_, err := PlanModules([]*devicetypes.CaniDeviceType{dev}, inv, "gpu", 5, StrategyFill)
	if err == nil {
		t.Fatal("expected error for not enough bays")
	}
	if !strings.Contains(err.Error(), "not enough free bays") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPlanModulesFillAllZeroQty(t *testing.T) {
	devA := makeTestDevice("dev-a", "hpe-xd670", gpuBays(3))
	devB := makeTestDevice("dev-b", "hpe-xd670", gpuBays(2))
	inv := makeTestInventory([]*devicetypes.CaniDeviceType{devA, devB})

	// qty=0 means fill all available bays.
	entries, err := PlanModules(
		[]*devicetypes.CaniDeviceType{devA, devB}, inv, "gpu", 0, StrategyFill)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 5 {
		t.Fatalf("expected 5 entries (3+2), got %d", len(entries))
	}
}

func TestBayFilterGPU(t *testing.T) {
	got := BayFilterForHardwareType("gpu")
	if got != "gpu" {
		t.Errorf("expected 'gpu', got %q", got)
	}
}

func TestBayFilterPSU(t *testing.T) {
	got := BayFilterForHardwareType("psu")
	if got != "psu" {
		t.Errorf("expected 'psu', got %q", got)
	}
}

func TestBayFilterUnknown(t *testing.T) {
	got := BayFilterForHardwareType("unknown")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestPrintModulePlan(t *testing.T) {
	entries := []ModulePlacementEntry{
		{DeviceName: "dev-a", BayName: "GPU 0", BayPosition: "GPU0"},
		{DeviceName: "dev-a", BayName: "GPU 1", BayPosition: "GPU1"},
	}
	names := []string{"gpu-a-0", "gpu-a-1"}

	var buf bytes.Buffer
	PrintModulePlan(&buf, entries, names)

	out := buf.String()
	if !strings.Contains(out, "dev-a") {
		t.Error("expected device name in output")
	}
	if !strings.Contains(out, "GPU 0") {
		t.Error("expected bay name in output")
	}
	if !strings.Contains(out, "gpu-a-0") {
		t.Error("expected module name in output")
	}
}
