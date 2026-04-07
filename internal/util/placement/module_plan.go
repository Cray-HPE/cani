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
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// ModulePlacementEntry describes a single module placement in a device bay.
type ModulePlacementEntry struct {
	DeviceID    uuid.UUID
	DeviceName  string
	BayName     string
	BayPosition string
	DeviceIndex int // 0-based index into the device list
	BayIndex    int // 0-based index into the bay list for this device
}

// PlanModules computes placement entries for qty modules across the given
// devices. Devices are sorted by name for deterministic results.
//
// The bayFilter is passed to Inventory.AvailableModuleBays to select
// which bays are eligible (e.g. "gpu" for GPU-type bays only).
//
// With StrategyFill, bays in each device are filled before moving to the
// next device. Returns an error if there are not enough free bays.
func PlanModules(
	devices []*devicetypes.CaniDeviceType,
	inv *devicetypes.Inventory,
	bayFilter string,
	qty int,
	strategy Strategy,
) ([]ModulePlacementEntry, error) {
	if len(devices) == 0 {
		return nil, fmt.Errorf("no candidate devices available")
	}

	sorted := sortDevicesByName(devices)

	switch strategy {
	case StrategyFill:
		return planModulesFill(sorted, inv, bayFilter, qty)
	default:
		return nil, fmt.Errorf("unsupported module placement strategy: %s", strategy)
	}
}

// planModulesFill packs bays in each device before moving to the next.
// A qty <= 0 means "fill all available bays".
func planModulesFill(
	devices []*devicetypes.CaniDeviceType,
	inv *devicetypes.Inventory,
	bayFilter string,
	qty int,
) ([]ModulePlacementEntry, error) {
	fillAll := qty <= 0
	var entries []ModulePlacementEntry
	placed := 0

	for devIdx, dev := range devices {
		if !fillAll && placed >= qty {
			break
		}
		bays := inv.AvailableModuleBays(dev.ID, bayFilter)
		for bayIdx, bay := range bays {
			if !fillAll && placed >= qty {
				break
			}
			entries = append(entries, ModulePlacementEntry{
				DeviceID:    dev.ID,
				DeviceName:  dev.Name,
				BayName:     bay.Name,
				BayPosition: bay.Position,
				DeviceIndex: devIdx,
				BayIndex:    bayIdx,
			})
			placed++
		}
	}

	if !fillAll && placed < qty {
		return nil, fmt.Errorf(
			"not enough free bays: need %d, found %d across %d device(s)",
			qty, placed, len(devices))
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no free bays available across %d device(s)", len(devices))
	}
	return entries, nil
}

// BayFilterForHardwareType returns a bay-name filter string that matches
// bays appropriate for the given module hardware type. This convention
// mirrors how device-type YAML names bays (e.g. "GPU 0", "PSU1", "NIC 0").
//
// Returns empty string if no convention mapping exists, which means all
// bays are eligible.
func BayFilterForHardwareType(hwType string) string {
	mapping := map[string]string{
		"gpu":         "gpu",
		"psu":         "psu",
		"nic":         "nic",
		"memory":      "dimm",
		"cpu":         "cpu",
		"transceiver": "sfp",
	}
	return mapping[strings.ToLower(hwType)]
}

// PrintModulePlan writes a module placement table for dry-run review.
func PrintModulePlan(w io.Writer, entries []ModulePlacementEntry, names []string) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "#\tDevice\tBay\tPosition\tName")
	fmt.Fprintln(tw, "-\t------\t---\t--------\t----")
	for i, e := range entries {
		name := ""
		if i < len(names) {
			name = names[i]
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n",
			i+1, e.DeviceName, e.BayName, e.BayPosition, name)
	}
	tw.Flush()
}

// sortDevicesByName returns a copy sorted alphabetically by Name.
func sortDevicesByName(devices []*devicetypes.CaniDeviceType) []*devicetypes.CaniDeviceType {
	sorted := make([]*devicetypes.CaniDeviceType, len(devices))
	copy(sorted, devices)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}
