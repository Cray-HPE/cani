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
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// renderCompactRow renders a row of racks with cables
func renderCompactRow(racks []*CompactRackView, cables []InterRackCable, maxHeight int, opts CompactRenderOptions) {
	if len(racks) == 0 {
		return
	}

	colors := newRowColors(opts.NoColor)
	interRackMap, intraRackMap := buildCableMap(racks, cables)
	rackWidth := compactRackWidth
	contentWidth := rackWidth - 2 // Width inside the box borders

	printRackNames(racks, rackWidth, colors.bold)
	printRackTopBorder(racks, rackWidth)

	for u := maxHeight; u >= 1; u-- {
		renderURow(racks, u, contentWidth, rackWidth, interRackMap, intraRackMap, colors)
	}

	printRackBottomBorder(racks, rackWidth)
}

// printRackNames prints the rack name header row above the boxes.
func printRackNames(racks []*CompactRackView, rackWidth int, bold func(string) string) {
	fmt.Print("   ") // U column space
	for _, rv := range racks {
		name := rv.Rack.Name
		if len(name) > rackWidth-2 {
			name = name[:rackWidth-2]
		}
		fmt.Printf(" %-*s", rackWidth, bold(name))
	}
	fmt.Println()
}

// printRackTopBorder prints the top border line for a row of racks.
func printRackTopBorder(racks []*CompactRackView, rackWidth int) {
	fmt.Print("   ") // U column space
	for range racks {
		fmt.Printf(" ┌%s┐ ", strings.Repeat("─", rackWidth-2))
	}
	fmt.Println()
}

// printRackBottomBorder prints the bottom border line for a row of racks.
func printRackBottomBorder(racks []*CompactRackView, rackWidth int) {
	fmt.Print("   ") // U column space
	for range racks {
		fmt.Printf(" └%s┘ ", strings.Repeat("─", rackWidth-2))
	}
	fmt.Println()
}

// renderURow renders a single U row across every rack in the row.
func renderURow(racks []*CompactRackView, u, contentWidth, rackWidth int, interRackMap, intraRackMap map[string]bool, colors rowColors) {
	fmt.Printf("%2d ", u) // U number

	for i, rv := range racks {
		if u > rv.Height {
			// This rack doesn't extend to this U
			fmt.Printf(" %s ", strings.Repeat(" ", rackWidth))
			continue
		}

		intraKey := fmt.Sprintf("%d-%d", i, u)
		content := renderRackCell(rv, u, contentWidth, intraRackMap[intraKey], colors)
		fmt.Printf(" │%s│", content)

		// Draw cable connections between racks
		if i < len(racks)-1 {
			cableKey := fmt.Sprintf("%d-%d-%d", i, i+1, u)
			if interRackMap[cableKey] {
				fmt.Print(colors.cyan(CableHoriz))
			} else {
				fmt.Print(" ")
			}
		}
	}
	fmt.Println()
}

// renderRackCell returns the rendered content for one rack's cell at U position u.
func renderRackCell(rv *CompactRackView, u, contentWidth int, hasIntraCable bool, colors rowColors) string {
	if device := rv.Devices[u]; device != nil {
		return deviceCellContent(device, contentWidth, hasIntraCable, colors)
	}
	if dev := findSpanningDevice(rv, u); dev != nil {
		return deviceCellContent(dev, contentWidth, hasIntraCable, colors)
	}
	return emptyCellContent(contentWidth, hasIntraCable, colors)
}

// deviceCellContent renders a device symbol cell, optionally with an intra-rack
// cable line.
func deviceCellContent(device *devicetypes.CaniDeviceType, contentWidth int, hasIntraCable bool, colors rowColors) string {
	content := colorizeDevice(string(getDeviceSymbol(device)), device, colors.red, colors.green, colors.yellow, colors.cyan)
	if hasIntraCable {
		return content + colors.cyan("│") + strings.Repeat(" ", contentWidth-2)
	}
	return content + strings.Repeat(" ", contentWidth-1)
}

// emptyCellContent renders an empty cell, optionally with an intra-rack cable
// line drawn through it.
func emptyCellContent(contentWidth int, hasIntraCable bool, colors rowColors) string {
	if hasIntraCable {
		return colors.gray(string(SymEmpty)) + colors.cyan("│") + colors.gray(strings.Repeat(string(SymEmpty), contentWidth-2))
	}
	return colors.gray(strings.Repeat(string(SymEmpty), contentWidth))
}

// findSpanningDevice returns the multi-U device occupying position u (when u is
// not the device's start position), or nil if none spans it.
func findSpanningDevice(rv *CompactRackView, u int) *devicetypes.CaniDeviceType {
	for startU, dev := range rv.Devices {
		if dev == nil {
			continue
		}
		uHeight := getDeviceUHeight(dev)
		if uHeight <= 0 {
			uHeight = 1
		}
		if startU < u && u < startU+uHeight {
			return dev
		}
	}
	return nil
}

// buildCableMap creates a map of cables at each U position between adjacent racks
// Also returns intra-rack cable map for cables within the same rack
func buildCableMap(racks []*CompactRackView, cables []InterRackCable) (map[string]bool, map[string]bool) {
	interRackMap := make(map[string]bool)
	intraRackMap := make(map[string]bool) // key: "rackIdx-lowU-highU"

	for _, cable := range cables {
		rackARelIdx, rackBRelIdx := resolveRelIndices(racks, cable)
		if rackARelIdx < 0 || rackBRelIdx < 0 {
			continue
		}

		if rackARelIdx == rackBRelIdx {
			markIntraRackCable(intraRackMap, rackARelIdx, cable.PositionA, cable.PositionB)
			continue
		}

		markInterRackCable(interRackMap, rackARelIdx, rackBRelIdx, cable.PositionA, cable.PositionB)
	}

	return interRackMap, intraRackMap
}

// resolveRelIndices maps a cable's absolute rack indices to their row-relative
// positions within racks, returning -1 for any endpoint not present.
func resolveRelIndices(racks []*CompactRackView, cable InterRackCable) (int, int) {
	a, b := -1, -1
	for i, rv := range racks {
		if rv.Index == cable.RackAIndex {
			a = i
		}
		if rv.Index == cable.RackBIndex {
			b = i
		}
	}
	return a, b
}

// markIntraRackCable marks every U position spanned by an intra-rack cable.
func markIntraRackCable(intraRackMap map[string]bool, rackIdx, posA, posB int) {
	lowU, highU := posA, posB
	if lowU > highU {
		lowU, highU = highU, lowU
	}
	for u := lowU; u <= highU; u++ {
		intraRackMap[fmt.Sprintf("%d-%d", rackIdx, u)] = true
	}
}

// markInterRackCable marks the adjacent-rack gaps a cable crosses, drawn at the
// average U position of its two endpoints.
func markInterRackCable(interRackMap map[string]bool, relA, relB, posA, posB int) {
	if relA > relB {
		relA, relB = relB, relA
		posA, posB = posB, posA
	}

	avgPos := (posA + posB) / 2
	if avgPos < 1 {
		avgPos = 1
	}

	for r := relA; r < relB; r++ {
		interRackMap[fmt.Sprintf("%d-%d-%d", r, r+1, avgPos)] = true
	}
}
