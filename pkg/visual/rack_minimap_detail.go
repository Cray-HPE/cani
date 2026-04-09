/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
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
package visual

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// detailAnnotationWidth is the max chars available for right-side annotations.
const detailAnnotationWidth = 120

// RenderMinimapDetailAll renders each rack one at a time with detail annotations.
func RenderMinimapDetailAll(inv *devicetypes.Inventory, opts CompactRenderOptions) error {
	if inv == nil {
		return fmt.Errorf("inventory is nil")
	}

	rackViews := buildCompactRackViews(inv, opts.RackFilter)
	if len(rackViews) == 0 {
		fmt.Println("No racks found in inventory")
		return nil
	}

	if opts.Verbose >= 1 {
		printMinimapLegend(opts)
	}

	for i, rv := range rackViews {
		if err := RenderMinimapDetail(inv, rv, opts); err != nil {
			return err
		}
		if i < len(rackViews)-1 {
			fmt.Println()
		}
	}
	return nil
}

// RenderMinimapDetail renders a single rack with right-side annotations.
func RenderMinimapDetail(inv *devicetypes.Inventory, rv *CompactRackView, opts CompactRenderOptions) error {
	if rv == nil || inv == nil {
		return fmt.Errorf("rack view or inventory is nil")
	}

	cf := newColorFuncs(opts.NoColor)
	grid := buildMinimapGrid(inv, rv)
	cableMap := buildDeviceCableMap(inv, rv)

	// Header
	fmt.Printf("    %s\n", cf.bold(rv.Rack.Name))
	fmt.Println("    ┌──┐")

	for u := rv.Height; u >= 1; u-- {
		slot := grid[u]
		c1 := cf.colorize(string(slot.Char1), slot.Color1)
		c2 := cf.colorize(string(slot.Char2), slot.Color2)

		annotation := buildAnnotation(inv, rv, u, cableMap, cf)
		fmt.Printf("%2d  │%s%s│ %s\n", u, c1, c2, annotation)
	}

	fmt.Println("    └──┘")
	printDetailSummary(inv, rv, cf)
	return nil
}

// buildAnnotation creates the annotation text for a single U row.
// Start rows use white text for name/height and grey for supplemental info.
// Continuation rows use grey text with a down-arrow prefix.
func buildAnnotation(inv *devicetypes.Inventory, rv *CompactRackView, u int, cableMap map[uuid.UUID][]cableInfo, cf colorFuncs) string {
	dev := findDeviceAtU(inv, rv, u)
	if dev == nil {
		return ""
	}

	uHeight := dev.GetUHeight()
	isStart := dev.RackPosition == u

	if !isStart {
		// Continuation row: grey "↓ name Nu"
		return cf.gray(fmt.Sprintf("↓ %s %du", dev.Name, uHeight))
	}

	// Start row: white "name Nu" + grey supplemental
	whitePart := cf.white(fmt.Sprintf("%s %du", dev.Name, uHeight))
	greyPart := buildSupplemental(inv, dev, cableMap, cf)

	if greyPart != "" {
		return whitePart + " " + cf.gray(greyPart)
	}
	return whitePart
}

// buildSupplemental builds the grey supplemental text for a device start row.
// Format: (x/yM type1, type2, ...) (x/yI)
func buildSupplemental(inv *devicetypes.Inventory, dev *devicetypes.CaniDeviceType, cableMap map[uuid.UUID][]cableInfo, cf colorFuncs) string {
	var parts []string

	// Module bays
	totalBays := len(dev.ModuleBays)
	if totalBays > 0 {
		populated, breakdown := countDeviceModules(inv, dev.ID, totalBays)
		modStr := fmt.Sprintf("(%d/%dM", populated, totalBays)
		if breakdown != "" {
			modStr += " " + breakdown
		}
		modStr += ")"
		parts = append(parts, modStr)
	}

	// Interfaces
	totalIntf := len(dev.Interfaces)
	if totalIntf > 0 {
		connected := countConnectedInterfaces(dev)
		parts = append(parts, fmt.Sprintf("(%d/%dI)", connected, totalIntf))
	}

	return strings.Join(parts, " ")
}

// countDeviceModules returns (populated count, type breakdown string) for a device.
// breakdown is e.g. "2 gpu, 1 nic"
func countDeviceModules(inv *devicetypes.Inventory, deviceID uuid.UUID, totalBays int) (int, string) {
	if inv == nil || inv.Modules == nil {
		return 0, ""
	}

	typeCounts := make(map[string]int)
	populated := 0
	for _, mod := range inv.Modules {
		if mod == nil || mod.ParentDevice != deviceID {
			continue
		}
		populated++
		hwType := strings.ToLower(string(mod.Type))
		if hwType == "" {
			hwType = "module"
		}
		typeCounts[hwType]++
	}

	if populated == 0 {
		return 0, ""
	}

	// Sort type names for deterministic output
	types := make([]string, 0, len(typeCounts))
	for t := range typeCounts {
		types = append(types, t)
	}
	sort.Strings(types)

	var breakdownParts []string
	for _, t := range types {
		breakdownParts = append(breakdownParts, fmt.Sprintf("%d %s", typeCounts[t], t))
	}
	return populated, strings.Join(breakdownParts, ", ")
}

// countConnectedInterfaces returns the number of interfaces with a connected cable.
func countConnectedInterfaces(dev *devicetypes.CaniDeviceType) int {
	if dev == nil {
		return 0
	}
	count := 0
	for _, iface := range dev.Interfaces {
		if iface.ConnectedCable != nil {
			count++
		}
	}
	return count
}

// findDeviceAtU finds the device occupying a U position (start or continuation).
func findDeviceAtU(inv *devicetypes.Inventory, rv *CompactRackView, u int) *devicetypes.CaniDeviceType {
	// Check OccupiedSlots first
	if rv.Rack.OccupiedSlots != nil {
		if faces, ok := rv.Rack.OccupiedSlots[u]; ok {
			for _, devID := range faces {
				if dev := inv.Devices[devID]; dev != nil {
					return dev
				}
			}
		}
	}

	// Fallback: walk Devices map
	if dev, ok := rv.Devices[u]; ok {
		return dev
	}
	for startU, dev := range rv.Devices {
		if dev == nil {
			continue
		}
		uHeight := dev.GetUHeight()
		if uHeight < 1 {
			uHeight = 1
		}
		if startU < u && u < startU+uHeight {
			return dev
		}
	}
	return nil
}

// cableInfo holds minimal cable termination data for annotation display.
type cableInfo struct {
	localPort  string
	remoteName string
	remotePort string
	cableSlug  string
}

// buildDeviceCableMap groups cable terminations by device UUID for a rack.
func buildDeviceCableMap(inv *devicetypes.Inventory, rv *CompactRackView) map[uuid.UUID][]cableInfo {
	result := make(map[uuid.UUID][]cableInfo)
	if inv == nil || inv.Cables == nil {
		return result
	}

	rackDevIDs := make(map[uuid.UUID]bool, len(rv.Rack.Devices))
	for _, devID := range rv.Rack.Devices {
		rackDevIDs[devID] = true
	}

	for _, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		addCableTermination(inv, cable, cable.TerminationADevice, cable.TerminationAPort,
			cable.TerminationBDevice, cable.TerminationBPort, rackDevIDs, result)
		addCableTermination(inv, cable, cable.TerminationBDevice, cable.TerminationBPort,
			cable.TerminationADevice, cable.TerminationAPort, rackDevIDs, result)
	}
	return result
}

// addCableTermination adds one side of a cable to the result map if localDev is in the rack.
func addCableTermination(inv *devicetypes.Inventory, cable *devicetypes.CaniCableType,
	localDev uuid.UUID, localPort string, remoteDev uuid.UUID, remotePort string,
	rackDevIDs map[uuid.UUID]bool, result map[uuid.UUID][]cableInfo) {

	if !rackDevIDs[localDev] {
		return
	}
	remoteName := "unknown"
	if rd := inv.Devices[remoteDev]; rd != nil {
		remoteName = rd.Name
		if len(remoteName) > 15 {
			remoteName = remoteName[:14] + "…"
		}
	}
	slug := cable.Slug
	if slug == "" {
		slug = "cable"
	}
	result[localDev] = append(result[localDev], cableInfo{
		localPort:  localPort,
		remoteName: remoteName,
		remotePort: remotePort,
		cableSlug:  slug,
	})
}

// printDetailSummary prints a summary line below the detail rack view.
func printDetailSummary(inv *devicetypes.Inventory, rv *CompactRackView, cf colorFuncs) {
	deviceCount := len(rv.Rack.Devices)
	occupied := countOccupiedU(rv)
	empty := rv.Height - occupied
	moduleCount := countModulesInRack(inv, rv)

	fmt.Printf("  %s: %d devices, %d/%d U occupied, %d U empty",
		cf.bold("Summary"), deviceCount, occupied, rv.Height, empty)
	if moduleCount > 0 {
		fmt.Printf(", %d modules", moduleCount)
	}
	fmt.Println()
}

// countOccupiedU counts the number of U positions with a device.
func countOccupiedU(rv *CompactRackView) int {
	if rv.Rack.OccupiedSlots != nil {
		return len(rv.Rack.OccupiedSlots)
	}
	count := 0
	for _, dev := range rv.Devices {
		if dev == nil {
			continue
		}
		h := dev.GetUHeight()
		if h < 1 {
			h = 1
		}
		count += h
	}
	return count
}

// countModulesInRack counts modules whose parent device is in this rack.
func countModulesInRack(inv *devicetypes.Inventory, rv *CompactRackView) int {
	if inv == nil || inv.Modules == nil {
		return 0
	}
	rackDevIDs := make(map[uuid.UUID]bool, len(rv.Rack.Devices))
	for _, devID := range rv.Rack.Devices {
		rackDevIDs[devID] = true
	}
	count := 0
	for _, mod := range inv.Modules {
		if mod != nil && rackDevIDs[mod.ParentDevice] {
			count++
		}
	}
	return count
}
