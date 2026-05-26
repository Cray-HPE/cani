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
	"github.com/google/uuid"
)

// MinimapSlot holds the 2-character representation for one U position.
type MinimapSlot struct {
	Char1  rune // front/full-depth indicator
	Char2  rune // rear/module indicator
	Color1 string
	Color2 string
}

const (
	symDevice   = 'D' // full-depth device start
	symDeviceHW = 'd' // half-width device start
	symModFull  = 'M' // has populated modules (full-depth)
	symModHalf  = 'm' // has populated modules (half-width)
	symCont     = '*' // multi-U continuation / no modules
	symDot      = '·' // empty
)

// minimap layout constants (80-char terminal)
const (
	minimapUCol    = 4 // "48  " width
	minimapPerRack = 5 // "│XX│ " = 4 chars + 1 gap
	minimapMaxCols = 80
)

// RenderMinimapRacks renders all racks as ultra-compact 2-char-wide columns.
func RenderMinimapRacks(inv *devicetypes.Inventory, opts CompactRenderOptions) error {
	if inv == nil {
		return fmt.Errorf("inventory is nil")
	}

	rackViews := buildCompactRackViews(inv, opts.RackFilter)
	if len(rackViews) == 0 {
		fmt.Println("No racks found in inventory")
		return nil
	}

	// Pre-compute minimap grids per rack
	grids := make(map[uuid.UUID]map[int]MinimapSlot, len(rackViews))
	for _, rv := range rackViews {
		grids[rv.RackID] = buildMinimapGrid(inv, rv)
	}

	columns := minimapColumnsForRow(len(rackViews), opts.Columns)

	if opts.Verbose >= 1 {
		printMinimapLegend(opts)
	}

	maxHeight := 0
	for _, rv := range rackViews {
		if rv.Height > maxHeight {
			maxHeight = rv.Height
		}
	}

	for rowStart := 0; rowStart < len(rackViews); rowStart += columns {
		rowEnd := rowStart + columns
		if rowEnd > len(rackViews) {
			rowEnd = len(rackViews)
		}
		renderMinimapRow(rackViews[rowStart:rowEnd], grids, maxHeight, opts)
		fmt.Println()
	}

	return nil
}

// minimapColumnsForRow decides how many racks fit in one 80-char row.
func minimapColumnsForRow(totalRacks, requestedCols int) int {
	if requestedCols > 0 {
		return requestedCols
	}
	available := minimapMaxCols - minimapUCol
	cols := available / minimapPerRack
	if cols < 1 {
		cols = 1
	}
	if cols > totalRacks {
		cols = totalRacks
	}
	return cols
}

// renderMinimapRow renders one horizontal row of minimap racks.
func renderMinimapRow(racks []*CompactRackView, grids map[uuid.UUID]map[int]MinimapSlot, maxHeight int, opts CompactRenderOptions) {
	cf := newColorFuncs(opts.NoColor)

	// Rack name headers — truncate to 4 chars, space-padded to 5
	fmt.Print(strings.Repeat(" ", minimapUCol))
	for _, rv := range racks {
		name := abbreviateRackName(rv.Rack.Name, 4)
		fmt.Printf(" %-4s", cf.bold(name))
	}
	fmt.Println()

	// Top border
	fmt.Print(strings.Repeat(" ", minimapUCol))
	for range racks {
		fmt.Print(" ┌──┐")
	}
	fmt.Println()

	// Each U from top to bottom
	for u := maxHeight; u >= 1; u-- {
		fmt.Printf("%2d  ", u)
		for _, rv := range racks {
			if u > rv.Height {
				fmt.Print("     ") // rack too short
				continue
			}
			slot := grids[rv.RackID][u]
			c1 := cf.colorize(string(slot.Char1), slot.Color1)
			c2 := cf.colorize(string(slot.Char2), slot.Color2)
			fmt.Printf(" │%s%s│", c1, c2)
		}
		fmt.Println()
	}

	// Bottom border
	fmt.Print(strings.Repeat(" ", minimapUCol))
	for range racks {
		fmt.Print(" └──┘")
	}
	fmt.Println()
}

// buildMinimapGrid pre-computes the MinimapSlot for every U in a rack.
func buildMinimapGrid(inv *devicetypes.Inventory, rv *CompactRackView) map[int]MinimapSlot {
	grid := make(map[int]MinimapSlot, rv.Height)

	// Initialize all slots as empty
	for u := 1; u <= rv.Height; u++ {
		grid[u] = MinimapSlot{Char1: symDot, Char2: symDot, Color1: "gray", Color2: "gray"}
	}

	// Walk OccupiedSlots for face-aware rendering
	if rv.Rack.OccupiedSlots != nil {
		for u, faces := range rv.Rack.OccupiedSlots {
			slot := grid[u]
			for face, devID := range faces {
				dev := inv.Devices[devID]
				if dev == nil {
					continue
				}
				isStart := dev.RackPosition == u
				hasMods := deviceHasModules(inv, devID)
				isFull := dev.IsFullDepth || dev.Face == "" || dev.Face == devicetypes.FaceFull
				ch, color := slotChar(dev, isStart, hasMods)
				switch face {
				case devicetypes.FaceFull:
					slot.Char1, slot.Color1 = ch, color
					slot.Char2 = secondSlotChar(isStart, hasMods, isFull)
					slot.Color2 = color
				case devicetypes.FaceFront:
					slot.Char1, slot.Color1 = ch, color
				case devicetypes.FaceRear:
					slot.Char2, slot.Color2 = ch, color
				}
			}
			grid[u] = slot
		}
		return grid
	}

	// Fallback: iterate devices when OccupiedSlots is not populated
	for startU, dev := range rv.Devices {
		if dev == nil {
			continue
		}
		hasMods := deviceHasModules(inv, dev.ID)
		isFull := dev.IsFullDepth || dev.Face == "" || dev.Face == devicetypes.FaceFull
		uHeight := dev.GetUHeight()
		if uHeight < 1 {
			uHeight = 1
		}
		for u := startU; u < startU+uHeight && u <= rv.Height; u++ {
			slot := grid[u]
			isStart := u == startU
			ch, color := slotChar(dev, isStart, hasMods)
			if isFull {
				slot.Char1, slot.Color1 = ch, color
				slot.Char2 = secondSlotChar(isStart, hasMods, isFull)
				slot.Color2 = color
			} else if dev.Face == devicetypes.FaceRear {
				slot.Char2, slot.Color2 = ch, color
			} else {
				slot.Char1, slot.Color1 = ch, color
			}
			grid[u] = slot
		}
	}
	return grid
}

// slotChar returns the rune and color key for the first character of a device slot.
func slotChar(dev *devicetypes.CaniDeviceType, isStart, hasMods bool) (rune, string) {
	color := statusColor(dev)
	if !isStart {
		return symCont, color
	}
	if dev.IsFullDepth || dev.Face == "" || dev.Face == devicetypes.FaceFull {
		return symDevice, color
	}
	return symDeviceHW, color
}

// secondSlotChar returns the second character for a device slot.
// Start + hasMods: M (full) or m (half). Start + !hasMods: *. Continuation: *.
func secondSlotChar(isStart, hasMods, isFullDepth bool) rune {
	if !isStart {
		return symCont
	}
	if hasMods {
		if isFullDepth {
			return symModFull
		}
		return symModHalf
	}
	return symCont
}

// statusColor picks a color key based on device status.
func statusColor(dev *devicetypes.CaniDeviceType) string {
	if dev == nil {
		return "gray"
	}
	return StatusColor(dev.Status)
}

// deviceHasModules returns true if inv.Modules contains any module parented to deviceID.
func deviceHasModules(inv *devicetypes.Inventory, deviceID uuid.UUID) bool {
	if inv == nil {
		return false
	}
	for _, mod := range inv.Modules {
		if mod != nil && mod.ParentDevice == deviceID {
			return true
		}
	}
	return false
}

// abbreviateRackName shortens a rack name to fit maxLen characters.
func abbreviateRackName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen]
}

// printMinimapLegend prints the symbol legend for the minimap view.
func printMinimapLegend(opts CompactRenderOptions) {
	cf := newColorFuncs(opts.NoColor)
	fmt.Println("  Symbols: " +
		cf.green("D") + "=device(full) " +
		cf.green("d") + "=device(half) " +
		cf.green("M") + "=hasmodules(full) " +
		cf.green("m") + "=hasmodules(half) " +
		cf.green("*") + "=continuation " +
		cf.gray("·") + "=empty")
	fmt.Println("  Colors:  " +
		cf.green("green") + "=active " +
		cf.yellow("yellow") + "=staged " +
		cf.red("red") + "=decommissioned " +
		cf.cyan("cyan") + "=other")
	fmt.Println()
}

// colorFuncs bundles color helper closures.
type colorFuncs struct {
	red    func(string) string
	green  func(string) string
	yellow func(string) string
	cyan   func(string) string
	gray   func(string) string
	bold   func(string) string
	white  func(string) string
}

func newColorFuncs(noColor bool) colorFuncs {
	id := func(s string) string { return s }
	if noColor {
		return colorFuncs{id, id, id, id, id, id, id}
	}
	return colorFuncs{
		red:    func(s string) string { return ColorRed + s + ColorReset },
		green:  func(s string) string { return ColorGreen + s + ColorReset },
		yellow: func(s string) string { return ColorYellow + s + ColorReset },
		cyan:   func(s string) string { return ColorCyan + s + ColorReset },
		gray:   func(s string) string { return ColorGray + s + ColorReset },
		bold:   func(s string) string { return ColorBold + s + ColorReset },
		white:  func(s string) string { return ColorWhite + s + ColorReset },
	}
}

// colorize maps a named color key to its ANSI-wrapped string.
func (cf colorFuncs) colorize(s, colorKey string) string {
	switch colorKey {
	case "red":
		return cf.red(s)
	case "green":
		return cf.green(s)
	case "yellow":
		return cf.yellow(s)
	case "cyan":
		return cf.cyan(s)
	case "white":
		return cf.white(s)
	default:
		return cf.gray(s)
	}
}
