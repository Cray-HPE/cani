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
	NoColor    bool                   // Disable ANSI colors
	RackFilter string                 // Filter to specific rack name
	Columns    int                    // Number of rack columns before wrapping (0 = auto)
	Verbose    int                    // 0 = no legend, 1 = legend, 2 = all cables
	CableType  string                 // Filter cables by type (e.g., "dac", "cat6")
	Detail     bool                   // Show single-rack detail with annotations
	Inventory  *devicetypes.Inventory // Full inventory for module/cable lookups
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

// buildCompactRackViews creates CompactRackView for each rack
func buildCompactRackViews(inv *devicetypes.Inventory, filter string) []*CompactRackView {
	var views []*CompactRackView

	// Sort racks by name for consistent ordering
	var rackIDs []uuid.UUID
	for id := range inv.Racks {
		rackIDs = append(rackIDs, id)
	}
	sort.Slice(rackIDs, func(i, j int) bool {
		return inv.Racks[rackIDs[i]].Name < inv.Racks[rackIDs[j]].Name
	})

	idx := 0
	for _, rackID := range rackIDs {
		rack := inv.Racks[rackID]
		if rack == nil {
			continue
		}

		// Apply filter
		if filter != "" && !strings.Contains(strings.ToLower(rack.Name), strings.ToLower(filter)) {
			continue
		}

		view := &CompactRackView{
			Rack:    rack,
			RackID:  rackID,
			Height:  rack.UHeight,
			Devices: make(map[int]*devicetypes.CaniDeviceType),
			Index:   idx,
		}

		// Map devices to their starting U positions
		for _, deviceID := range rack.Devices {
			device := inv.Devices[deviceID]
			if device == nil || device.RackPosition <= 0 {
				continue
			}
			view.Devices[device.RackPosition] = device
		}

		views = append(views, view)
		idx++
	}

	return views
}

// findInterRackCables finds cables that connect devices in different racks
func findInterRackCables(inv *devicetypes.Inventory, rackViews []*CompactRackView, cableTypeFilter string, includeAll bool) []InterRackCable {
	var cables []InterRackCable

	// Build device to rack index map
	deviceToRackIdx := make(map[uuid.UUID]int)
	deviceToPosition := make(map[uuid.UUID]int)
	for _, rv := range rackViews {
		for _, deviceID := range rv.Rack.Devices {
			deviceToRackIdx[deviceID] = rv.Index
			if device := inv.Devices[deviceID]; device != nil {
				deviceToPosition[deviceID] = device.RackPosition
			}
		}
	}

	for _, cable := range inv.Cables {
		if cable == nil {
			continue
		}

		// Apply cable type filter
		if cableTypeFilter != "" && !strings.Contains(strings.ToLower(cable.Slug), strings.ToLower(cableTypeFilter)) {
			continue
		}

		deviceA := cable.TerminationADevice
		deviceB := cable.TerminationBDevice

		rackIdxA, okA := deviceToRackIdx[deviceA]
		rackIdxB, okB := deviceToRackIdx[deviceB]

		if !okA || !okB {
			continue
		}

		// Only include inter-rack cables by default
		if !includeAll && rackIdxA == rackIdxB {
			continue
		}

		// Get device names
		deviceAName := ""
		deviceBName := ""
		if d := inv.Devices[deviceA]; d != nil {
			deviceAName = d.Name
		}
		if d := inv.Devices[deviceB]; d != nil {
			deviceBName = d.Name
		}

		cables = append(cables, InterRackCable{
			Cable:       cable,
			RackAIndex:  rackIdxA,
			RackBIndex:  rackIdxB,
			PositionA:   deviceToPosition[deviceA],
			PositionB:   deviceToPosition[deviceB],
			DeviceAName: deviceAName,
			DeviceBName: deviceBName,
			PortA:       cable.TerminationAPort,
			PortB:       cable.TerminationBPort,
		})
	}

	return cables
}

// filterCablesForRow returns cables that connect racks in the given row range
func filterCablesForRow(cables []InterRackCable, rowStart, rowEnd int) []InterRackCable {
	var filtered []InterRackCable
	for _, c := range cables {
		// Cable connects racks in this row
		if c.RackAIndex >= rowStart && c.RackAIndex < rowEnd &&
			c.RackBIndex >= rowStart && c.RackBIndex < rowEnd {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// renderCompactRow renders a row of racks with cables
func renderCompactRow(racks []*CompactRackView, cables []InterRackCable, maxHeight int, opts CompactRenderOptions) {
	if len(racks) == 0 {
		return
	}

	// Color functions
	green := func(s string) string { return s }
	yellow := func(s string) string { return s }
	cyan := func(s string) string { return s }
	gray := func(s string) string { return s }
	bold := func(s string) string { return s }
	red := func(s string) string { return s }

	if !opts.NoColor {
		green = func(s string) string { return ColorGreen + s + ColorReset }
		yellow = func(s string) string { return ColorYellow + s + ColorReset }
		cyan = func(s string) string { return ColorCyan + s + ColorReset }
		gray = func(s string) string { return ColorGray + s + ColorReset }
		bold = func(s string) string { return ColorBold + s + ColorReset }
		red = func(s string) string { return ColorRed + s + ColorReset }
	}

	// Build cable map: for each U, track which rack pairs have cables
	interRackMap, intraRackMap := buildCableMap(racks, cables)

	// Calculate rack column width (name + box)
	rackWidth := 12 // "│NNNNN│" + some padding

	// Print rack names
	fmt.Print("   ") // U column space
	for _, rv := range racks {
		name := rv.Rack.Name
		if len(name) > rackWidth-2 {
			name = name[:rackWidth-2]
		}
		fmt.Printf(" %-*s", rackWidth, bold(name))
	}
	fmt.Println()

	// Print top border
	fmt.Print("   ") // U column space
	for range racks {
		fmt.Printf(" ┌%s┐ ", strings.Repeat("─", rackWidth-2))
	}
	fmt.Println()

	// Print each U from top to bottom
	contentWidth := rackWidth - 2 // Width inside the box borders

	for u := maxHeight; u >= 1; u-- {
		// U number
		fmt.Printf("%2d ", u)

		// Each rack at this U
		for i, rv := range racks {
			if u > rv.Height {
				// This rack doesn't extend to this U
				fmt.Printf(" %s ", strings.Repeat(" ", rackWidth))
				continue
			}

			// Get device at this U (check if it's the start position)
			device := rv.Devices[u]
			var content string

			// Check if there's an intra-rack cable at this U position
			intraKey := fmt.Sprintf("%d-%d", i, u)
			hasIntraCable := intraRackMap[intraKey]

			if device != nil {
				// Device starts at this U - show symbol
				sym := getDeviceSymbol(device)
				content = string(sym)
				content = colorizeDevice(content, device, red, green, yellow, cyan)
				if hasIntraCable {
					content = content + cyan("│") + strings.Repeat(" ", contentWidth-2)
				} else {
					content = content + strings.Repeat(" ", contentWidth-1)
				}
			} else {
				// Check if this U is part of a multi-U device
				found := false
				for startU, dev := range rv.Devices {
					if dev == nil {
						continue
					}
					uHeight := getDeviceUHeight(dev)
					if uHeight <= 0 {
						uHeight = 1
					}
					if startU < u && u < startU+uHeight {
						sym := getDeviceSymbol(dev)
						content = string(sym)
						content = colorizeDevice(content, dev, red, green, yellow, cyan)
						if hasIntraCable {
							content = content + cyan("│") + strings.Repeat(" ", contentWidth-2)
						} else {
							content = content + strings.Repeat(" ", contentWidth-1)
						}
						found = true
						break
					}
				}
				if !found {
					if hasIntraCable {
						// Show vertical cable line through empty space
						content = gray(string(SymEmpty)) + cyan("│") + gray(strings.Repeat(string(SymEmpty), contentWidth-2))
					} else {
						content = gray(strings.Repeat(string(SymEmpty), contentWidth))
					}
				}
			}

			fmt.Printf(" │%s│", content)

			// Draw cable connections between racks
			if i < len(racks)-1 {
				cableKey := fmt.Sprintf("%d-%d-%d", i, i+1, u)
				if interRackMap[cableKey] {
					fmt.Print(cyan(CableHoriz))
				} else {
					fmt.Print(" ")
				}
			}
		}
		fmt.Println()
	}

	// Print bottom border
	fmt.Print("   ") // U column space
	for range racks {
		fmt.Printf(" └%s┘ ", strings.Repeat("─", rackWidth-2))
	}
	fmt.Println()
}

// buildCableMap creates a map of cables at each U position between adjacent racks
// Also returns intra-rack cable map for cables within the same rack
func buildCableMap(racks []*CompactRackView, cables []InterRackCable) (map[string]bool, map[string]bool) {
	interRackMap := make(map[string]bool)
	intraRackMap := make(map[string]bool) // key: "rackIdx-lowU-highU"

	for _, cable := range cables {
		// Normalize rack indices to row-relative
		rackARelIdx := -1
		rackBRelIdx := -1
		for i, rv := range racks {
			if rv.Index == cable.RackAIndex {
				rackARelIdx = i
			}
			if rv.Index == cable.RackBIndex {
				rackBRelIdx = i
			}
		}

		if rackARelIdx < 0 || rackBRelIdx < 0 {
			continue
		}

		// Handle intra-rack cables
		if rackARelIdx == rackBRelIdx {
			lowU := cable.PositionA
			highU := cable.PositionB
			if lowU > highU {
				lowU, highU = highU, lowU
			}
			// Mark all U positions between the two endpoints
			for u := lowU; u <= highU; u++ {
				key := fmt.Sprintf("%d-%d", rackARelIdx, u)
				intraRackMap[key] = true
			}
			continue
		}

		// Ensure A < B for consistent key
		if rackARelIdx > rackBRelIdx {
			rackARelIdx, rackBRelIdx = rackBRelIdx, rackARelIdx
			cable.PositionA, cable.PositionB = cable.PositionB, cable.PositionA
		}

		// Mark cables at both endpoint positions
		// For adjacent racks, draw at the average position
		avgPos := (cable.PositionA + cable.PositionB) / 2
		if avgPos < 1 {
			avgPos = 1
		}

		// Mark all racks between A and B
		for r := rackARelIdx; r < rackBRelIdx; r++ {
			key := fmt.Sprintf("%d-%d-%d", r, r+1, avgPos)
			interRackMap[key] = true
		}
	}

	return interRackMap, intraRackMap
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
	green := func(s string) string { return s }
	yellow := func(s string) string { return s }
	cyan := func(s string) string { return s }
	gray := func(s string) string { return s }
	red := func(s string) string { return s }

	if !opts.NoColor {
		green = func(s string) string { return ColorGreen + s + ColorReset }
		yellow = func(s string) string { return ColorYellow + s + ColorReset }
		cyan = func(s string) string { return ColorCyan + s + ColorReset }
		gray = func(s string) string { return ColorGray + s + ColorReset }
		red = func(s string) string { return ColorRed + s + ColorReset }
	}

	fmt.Println(" .-- Device type: " + green("S") + "=switch " + green("N") + "=node " + green("B") + "=blade " + yellow("P") + "=pdu " + cyan("C") + "=cdu " + gray("#") + "=chassis")
	fmt.Println("/ .- Status: " + green("green") + "=active " + yellow("yellow") + "=staged " + red("red") + "=decommissioned " + cyan("cyan") + "=other")
	fmt.Println("||  Cable: " + cyan("───") + " inter-rack  " + cyan("│") + " intra-rack")
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

// CableEndpoint represents one end of a cable for routing
type CableEndpoint struct {
	U        int    // U position in rack
	Port     string // Port name
	DestName string // Destination device name
	DestPort string // Destination port
	DestU    int    // Destination U position
	GoingUp  bool   // True if cable goes up (dest U > src U)
}

// CableRoute tracks a cable's path through the grid
type CableRoute struct {
	StartU   int
	EndU     int
	Column   int // Horizontal column for this cable's vertical segment
	Endpoint CableEndpoint
}

// RenderCompactRacksWithCables renders racks one per line with chronyc-style cable visualization
func RenderCompactRacksWithCables(inv *devicetypes.Inventory, opts CompactRenderOptions) error {
	if inv == nil {
		return fmt.Errorf("inventory is nil")
	}

	rackViews := buildCompactRackViews(inv, opts.RackFilter)
	if len(rackViews) == 0 {
		fmt.Println("No racks found in inventory")
		return nil
	}

	if opts.Verbose >= 1 {
		printCableLegend(opts)
	}

	for _, rv := range rackViews {
		rackCables := findRackCables(inv, rv, opts.CableType, opts.Verbose >= 2)
		renderRackWithCables(rv, rackCables, opts)
		fmt.Println()
	}

	return nil
}

// findRackCables finds all cables connected to devices in a rack
func findRackCables(inv *devicetypes.Inventory, rv *CompactRackView, cableTypeFilter string, includeAll bool) map[int][]CableEndpoint {
	cables := make(map[int][]CableEndpoint)

	rackDeviceIDs := make(map[uuid.UUID]bool)
	for _, deviceID := range rv.Rack.Devices {
		rackDeviceIDs[deviceID] = true
	}

	for _, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		if cableTypeFilter != "" && !strings.Contains(strings.ToLower(cable.Slug), strings.ToLower(cableTypeFilter)) {
			continue
		}

		aInRack := rackDeviceIDs[cable.TerminationADevice]
		bInRack := rackDeviceIDs[cable.TerminationBDevice]

		if !aInRack && !bInRack {
			continue
		}
		if aInRack && bInRack && !includeAll {
			continue
		}

		devA := inv.Devices[cable.TerminationADevice]
		devB := inv.Devices[cable.TerminationBDevice]
		if devA == nil || devB == nil {
			continue
		}

		if aInRack {
			ep := CableEndpoint{
				U:        devA.RackPosition,
				Port:     cable.TerminationAPort,
				DestName: truncateName(devB.Name, 15),
				DestPort: cable.TerminationBPort,
				DestU:    devB.RackPosition,
				GoingUp:  devB.RackPosition > devA.RackPosition,
			}
			cables[devA.RackPosition] = append(cables[devA.RackPosition], ep)
		}
		if bInRack && !aInRack {
			ep := CableEndpoint{
				U:        devB.RackPosition,
				Port:     cable.TerminationBPort,
				DestName: truncateName(devA.Name, 15),
				DestPort: cable.TerminationAPort,
				DestU:    devA.RackPosition,
				GoingUp:  devA.RackPosition > devB.RackPosition,
			}
			cables[devB.RackPosition] = append(cables[devB.RackPosition], ep)
		}
	}
	return cables
}

// truncateName shortens a name to maxLen characters
func truncateName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen-1] + "…"
}

// renderRackWithCables renders a single rack with chronyc-style cable branching
func renderRackWithCables(rv *CompactRackView, cables map[int][]CableEndpoint, opts CompactRenderOptions) {
	green := func(s string) string { return s }
	yellow := func(s string) string { return s }
	cyan := func(s string) string { return s }
	gray := func(s string) string { return s }
	bold := func(s string) string { return s }
	red := func(s string) string { return s }

	if !opts.NoColor {
		green = func(s string) string { return ColorGreen + s + ColorReset }
		yellow = func(s string) string { return ColorYellow + s + ColorReset }
		cyan = func(s string) string { return ColorCyan + s + ColorReset }
		gray = func(s string) string { return ColorGray + s + ColorReset }
		bold = func(s string) string { return ColorBold + s + ColorReset }
		red = func(s string) string { return ColorRed + s + ColorReset }
	}

	rackWidth := 12
	contentWidth := rackWidth - 2
	cableGrid := buildCableGrid(rv.Height, cables)

	fmt.Printf("    %s\n", bold(rv.Rack.Name))
	fmt.Printf("    ┌%s┐\n", strings.Repeat("─", contentWidth))

	for u := rv.Height; u >= 1; u-- {
		device := rv.Devices[u]
		var content string

		if device != nil {
			sym := getDeviceSymbol(device)
			content = colorizeDevice(string(sym), device, red, green, yellow, cyan)
			content = content + strings.Repeat(" ", contentWidth-1)
		} else {
			found := false
			for startU, dev := range rv.Devices {
				if dev == nil {
					continue
				}
				uHeight := getDeviceUHeight(dev)
				if uHeight <= 0 {
					uHeight = 1
				}
				if startU < u && u < startU+uHeight {
					sym := getDeviceSymbol(dev)
					content = colorizeDevice(string(sym), dev, red, green, yellow, cyan)
					content = content + strings.Repeat(" ", contentWidth-1)
					found = true
					break
				}
			}
			if !found {
				content = gray(strings.Repeat(string(SymEmpty), contentWidth))
			}
		}

		cableVis := buildCableVisualization(u, cables, cableGrid, cyan)
		fmt.Printf("%2d  │%s│%s\n", u, content, cableVis)
	}

	fmt.Printf("    └%s┘\n", strings.Repeat("─", contentWidth))
}

// buildCableGrid creates routing information for all cables
func buildCableGrid(height int, cables map[int][]CableEndpoint) [][]CableRoute {
	var routes []CableRoute
	for srcU, endpoints := range cables {
		for _, ep := range endpoints {
			routes = append(routes, CableRoute{StartU: srcU, EndU: ep.DestU, Endpoint: ep})
		}
	}

	// Sort routes by range (larger ranges get outer columns)
	sort.Slice(routes, func(i, j int) bool {
		return abs(routes[i].EndU-routes[i].StartU) > abs(routes[j].EndU-routes[j].StartU)
	})

	for i := range routes {
		routes[i].Column = i
	}

	grid := make([][]CableRoute, height+1)
	for _, route := range routes {
		lowU, highU := min(route.StartU, route.EndU), max(route.StartU, route.EndU)
		for u := lowU; u <= highU; u++ {
			if u >= 0 && u <= height {
				grid[u] = append(grid[u], route)
			}
		}
	}
	return grid
}

// buildCableVisualization creates the cable drawing string for a given U row
func buildCableVisualization(u int, cables map[int][]CableEndpoint, grid [][]CableRoute, cyan func(string) string) string {
	if u < 1 || u >= len(grid) || len(grid[u]) == 0 {
		return ""
	}

	maxCol := 0
	for _, r := range grid[u] {
		if r.Column > maxCol {
			maxCol = r.Column
		}
	}

	width := (maxCol + 1) * 2
	vis := make([]rune, width)
	for i := range vis {
		vis[i] = ' '
	}

	// Draw cables starting at this U
	for _, ep := range cables[u] {
		for _, route := range grid[u] {
			if route.StartU == u && route.Endpoint.DestName == ep.DestName && route.Endpoint.DestPort == ep.DestPort {
				for col := 0; col <= route.Column; col++ {
					pos := col * 2
					if pos < len(vis) {
						if col < route.Column {
							vis[pos] = '─'
						} else if ep.GoingUp {
							vis[pos] = '/'
						} else {
							vis[pos] = '\\'
						}
					}
				}
				break
			}
		}
	}

	// Draw vertical lines for cables passing through
	for _, route := range grid[u] {
		pos := route.Column * 2
		if pos >= len(vis) {
			continue
		}
		if route.StartU != u && route.EndU != u && vis[pos] == ' ' {
			vis[pos] = '│'
		} else if route.EndU == u && route.StartU != u {
			if route.StartU > u {
				vis[pos] = '\\'
			} else {
				vis[pos] = '/'
			}
		}
	}

	// Add destination labels for cables starting at this U
	var labels []string
	for _, ep := range cables[u] {
		labels = append(labels, fmt.Sprintf("%s:%s", ep.DestName, ep.DestPort))
	}

	result := string(vis)
	if len(labels) > 0 {
		result += " " + strings.Join(labels, ", ")
	}
	return cyan(strings.TrimRight(result, " "))
}

// printCableLegend prints the legend for cable routing view
func printCableLegend(opts CompactRenderOptions) {
	green := func(s string) string { return s }
	yellow := func(s string) string { return s }
	cyan := func(s string) string { return s }
	gray := func(s string) string { return s }
	red := func(s string) string { return s }

	if !opts.NoColor {
		green = func(s string) string { return ColorGreen + s + ColorReset }
		yellow = func(s string) string { return ColorYellow + s + ColorReset }
		cyan = func(s string) string { return ColorCyan + s + ColorReset }
		gray = func(s string) string { return ColorGray + s + ColorReset }
		red = func(s string) string { return ColorRed + s + ColorReset }
	}

	fmt.Println(" .-- Device: " + green("S") + "=switch " + green("N") + "=node " + green("B") + "=blade " + yellow("P") + "=pdu " + cyan("C") + "=cdu " + gray("#") + "=chassis")
	fmt.Println("/ .- Status: " + green("green") + "=active " + yellow("yellow") + "=staged " + red("red") + "=decommissioned " + cyan("cyan") + "=other")
	fmt.Println("||  Cable: " + cyan("─") + " horiz  " + cyan("/") + " up  " + cyan("\\") + " down  " + cyan("│") + " pass-through")
	if opts.Verbose >= 2 {
		fmt.Println("||  (showing all cables)")
	} else {
		fmt.Println("||  (showing inter-rack cables only, use -VV for all)")
	}
	fmt.Println()
}
