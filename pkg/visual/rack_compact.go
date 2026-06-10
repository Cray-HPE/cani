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

// Device type symbols for compact display
const (
	SymSwitch  = 'S' // switch
	SymNode    = 'N' // node/server/blade
	SymBlade   = 'B' // blade (alias for node)
	SymPDU     = 'P' // PDU
	SymCDU     = 'C' // CDU
	SymChassis = '#' // chassis
	SymUnknown = '?' // unknown device type
	SymEmpty   = '·' // empty slot

	// Cable drawing characters
	CableHoriz    = "─"
	CableVertical = "│"
	CableLeft     = "┤"
	CableRight    = "├"
	CableCross    = "┼"
)

// CompactRenderOptions controls compact rack visualization
type CompactRenderOptions struct {
	NoColor     bool                   // Disable ANSI colors
	RackFilter  string                 // Filter to specific rack name
	Columns     int                    // Number of rack columns before wrapping (0 = auto)
	Verbose     int                    // 0 = no legend, 1 = legend, 2 = all cables
	CableType   string                 // Filter cables by type (e.g., "dac", "cat6")
	Detail      bool                   // Show single-rack detail with annotations
	ShowLabels  bool                   // Show A/B termination labels on routing view
	Interactive bool                   // Enter interactive toggle mode for routing view
	Inventory   *devicetypes.Inventory // Full inventory for module/cable lookups
}

// CompactRackView holds pre-computed rack data for compact rendering
type CompactRackView struct {
	Rack    *devicetypes.CaniRackType
	RackID  uuid.UUID
	Height  int
	Devices map[int]*devicetypes.CaniDeviceType // U position -> device (start position only)
	Index   int                                 // Column index for cable drawing
}

// InterRackCable represents a cable between two racks
type InterRackCable struct {
	Cable       *devicetypes.CaniCableType
	RackAIndex  int // Column index of rack A
	RackBIndex  int // Column index of rack B
	PositionA   int // U position in rack A
	PositionB   int // U position in rack B
	DeviceAName string
	DeviceBName string
	PortA       string
	PortB       string
}

// compactRackWidth is the column width (name + box) used when rendering racks.
const compactRackWidth = 12

// rowColors holds the ANSI color closures used by the compact renderers.
type rowColors struct {
	green, yellow, cyan, gray, bold, red func(string) string
}

// newRowColors returns color closures. When noColor is true every closure is
// the identity function, otherwise each wraps its argument in the matching
// ANSI escape sequence.
func newRowColors(noColor bool) rowColors {
	id := func(s string) string { return s }
	c := rowColors{green: id, yellow: id, cyan: id, gray: id, bold: id, red: id}
	if noColor {
		return c
	}
	c.green = func(s string) string { return ColorGreen + s + ColorReset }
	c.yellow = func(s string) string { return ColorYellow + s + ColorReset }
	c.cyan = func(s string) string { return ColorCyan + s + ColorReset }
	c.gray = func(s string) string { return ColorGray + s + ColorReset }
	c.bold = func(s string) string { return ColorBold + s + ColorReset }
	c.red = func(s string) string { return ColorRed + s + ColorReset }
	return c
}

// RenderCompactRacks renders all racks in a compact ASCII format
func RenderCompactRacks(inv *devicetypes.Inventory, opts CompactRenderOptions) error {
	if inv == nil {
		return fmt.Errorf("inventory is nil")
	}

	// Build rack views
	rackViews := buildCompactRackViews(inv, opts.RackFilter)
	if len(rackViews) == 0 {
		fmt.Println("No racks found in inventory")
		return nil
	}

	// Find inter-rack cables
	cables := findInterRackCables(inv, rackViews, opts.CableType, opts.Verbose >= 2)

	// Determine columns
	columns := opts.Columns
	if columns <= 0 {
		columns = len(rackViews) // Show all racks in one row by default
		if columns > 6 {
			columns = 6 // Cap at 6 for readability
		}
	}

	// Print legend if verbose
	if opts.Verbose >= 1 {
		printCompactLegend(opts)
	}

	// Find max height across all racks
	maxHeight := 0
	for _, rv := range rackViews {
		if rv.Height > maxHeight {
			maxHeight = rv.Height
		}
	}

	// Render racks in rows
	for rowStart := 0; rowStart < len(rackViews); rowStart += columns {
		rowEnd := rowStart + columns
		if rowEnd > len(rackViews) {
			rowEnd = len(rackViews)
		}
		rowRacks := rackViews[rowStart:rowEnd]

		// Filter cables for this row
		rowCables := filterCablesForRow(cables, rowStart, rowEnd)

		renderCompactRow(rowRacks, rowCables, maxHeight, opts)
		fmt.Println() // Blank line between rows
	}

	return nil
}

// getDeviceSymbol returns the single-character symbol for a device type
func getDeviceSymbol(device *devicetypes.CaniDeviceType) rune {
	if device == nil {
		return SymEmpty
	}

	hwType := strings.ToLower(string(device.Type))
	switch {
	case strings.Contains(hwType, "switch") || hwType == "hsn-switch" || hwType == "mgmt-switch":
		return SymSwitch
	case strings.Contains(hwType, "node") || strings.Contains(hwType, "server"):
		return SymNode
	case strings.Contains(hwType, "blade"):
		return SymBlade
	case strings.Contains(hwType, "pdu") || hwType == "cabinet-pdu":
		return SymPDU
	case strings.Contains(hwType, "cdu"):
		return SymCDU
	case strings.Contains(hwType, "chassis"):
		return SymChassis
	default:
		return SymUnknown
	}
}

// colorizeDevice applies color based on device status.
func colorizeDevice(s string, dev *devicetypes.CaniDeviceType, red, green, yellow, cyan func(string) string) string {
	if dev == nil {
		return s
	}
	switch StatusColor(dev.Status) {
	case "red":
		return red(s)
	case "green":
		return green(s)
	case "yellow":
		return yellow(s)
	default:
		return cyan(s)
	}
}

// printCompactLegend prints the legend header for compact view
func printCompactLegend(opts CompactRenderOptions) {
	c := newRowColors(opts.NoColor)

	fmt.Println(" .-- Device type: " + c.green("S") + "=switch " + c.green("N") + "=node " + c.green("B") + "=blade " + c.yellow("P") + "=pdu " + c.cyan("C") + "=cdu " + c.gray("#") + "=chassis")
	fmt.Println("/ .- Status: " + c.green("green") + "=active " + c.yellow("yellow") + "=staged " + c.red("red") + "=decommissioned " + c.cyan("cyan") + "=other")
	fmt.Println("||  Cable: " + c.cyan("───") + " inter-rack  " + c.cyan("│") + " intra-rack")
	if opts.Verbose >= 2 {
		fmt.Println("||  (showing all cables)")
	} else {
		fmt.Println("||  (showing inter-rack cables only, use -VV for all)")
	}
	fmt.Println()
}

// printCableLegend prints the legend for cable routing view
func printCableLegend(opts CompactRenderOptions) {
	c := newRowColors(opts.NoColor)

	fmt.Println(" .-- Device: " + c.green("S") + "=switch " + c.green("N") + "=node " + c.green("B") + "=blade " + c.yellow("P") + "=pdu " + c.cyan("C") + "=cdu " + c.gray("#") + "=chassis")
	fmt.Println("/ .- Status: " + c.green("green") + "=active " + c.yellow("yellow") + "=staged " + c.red("red") + "=decommissioned " + c.cyan("cyan") + "=other")
	fmt.Println("||  Cable: " + c.cyan("─") + " horiz  " + c.cyan("/") + " up  " + c.cyan("\\") + " down  " + c.cyan("│") + " pass-through")
	if opts.Verbose >= 2 {
		fmt.Println("||  (showing all cables)")
	} else {
		fmt.Println("||  (showing inter-rack cables only, use -VV for all)")
	}
	fmt.Println()
}

// min returns the minimum of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two ints
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// abs returns absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// truncateName shortens a name to maxLen characters
func truncateName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen-1] + "…"
}
