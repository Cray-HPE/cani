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

// cableDestNameMaxLen bounds destination device names shown in the cable view.
const cableDestNameMaxLen = 15

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
	rackDeviceIDs := rackDeviceIDSet(rv)

	for _, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		if cableTypeFilter != "" && !strings.Contains(strings.ToLower(cable.Slug), strings.ToLower(cableTypeFilter)) {
			continue
		}
		addRackCableEndpoints(cables, inv, cable, rackDeviceIDs, includeAll)
	}
	return cables
}

// rackDeviceIDSet returns the set of device IDs installed in a rack.
func rackDeviceIDSet(rv *CompactRackView) map[uuid.UUID]bool {
	rackDeviceIDs := make(map[uuid.UUID]bool)
	for _, deviceID := range rv.Rack.Devices {
		rackDeviceIDs[deviceID] = true
	}
	return rackDeviceIDs
}

// addRackCableEndpoints appends cable endpoints for any in-rack termination of a
// cable, keyed by the in-rack device's U position. Intra-rack cables are skipped
// unless includeAll is set.
func addRackCableEndpoints(cables map[int][]CableEndpoint, inv *devicetypes.Inventory, cable *devicetypes.CaniCableType, rackDeviceIDs map[uuid.UUID]bool, includeAll bool) {
	aInRack := rackDeviceIDs[cable.TerminationADevice]
	bInRack := rackDeviceIDs[cable.TerminationBDevice]

	if !aInRack && !bInRack {
		return
	}
	if aInRack && bInRack && !includeAll {
		return
	}

	devA := inv.Devices[cable.TerminationADevice]
	devB := inv.Devices[cable.TerminationBDevice]
	if devA == nil || devB == nil {
		return
	}

	if aInRack {
		ep := makeCableEndpoint(devA, devB, cable.TerminationAPort, cable.TerminationBPort)
		cables[devA.RackPosition] = append(cables[devA.RackPosition], ep)
	}
	if bInRack && !aInRack {
		ep := makeCableEndpoint(devB, devA, cable.TerminationBPort, cable.TerminationAPort)
		cables[devB.RackPosition] = append(cables[devB.RackPosition], ep)
	}
}

// makeCableEndpoint builds a CableEndpoint anchored on src pointing at dst.
func makeCableEndpoint(src, dst *devicetypes.CaniDeviceType, srcPort, dstPort string) CableEndpoint {
	return CableEndpoint{
		U:        src.RackPosition,
		Port:     srcPort,
		DestName: truncateName(dst.Name, cableDestNameMaxLen),
		DestPort: dstPort,
		DestU:    dst.RackPosition,
		GoingUp:  dst.RackPosition > src.RackPosition,
	}
}

// renderRackWithCables renders a single rack with chronyc-style cable branching
func renderRackWithCables(rv *CompactRackView, cables map[int][]CableEndpoint, opts CompactRenderOptions) {
	colors := newRowColors(opts.NoColor)
	contentWidth := compactRackWidth - 2
	cableGrid := buildCableGrid(rv.Height, cables)

	fmt.Printf("    %s\n", colors.bold(rv.Rack.Name))
	fmt.Printf("    ┌%s┐\n", strings.Repeat("─", contentWidth))

	for u := rv.Height; u >= 1; u-- {
		content := renderRackCell(rv, u, contentWidth, false, colors)
		cableVis := buildCableVisualization(u, cables, cableGrid, colors.cyan)
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

	vis := newCableVisRow(grid[u])
	drawStartingCables(vis, cables[u], grid[u], u)
	drawPassThroughCables(vis, grid[u], u)

	result := string(vis)
	if labels := cableLabels(cables[u]); len(labels) > 0 {
		result += " " + strings.Join(labels, ", ")
	}
	return cyan(strings.TrimRight(result, " "))
}

// newCableVisRow allocates a blank visualization row wide enough for the widest
// column used by the routes passing through this U.
func newCableVisRow(routes []CableRoute) []rune {
	maxCol := 0
	for _, r := range routes {
		if r.Column > maxCol {
			maxCol = r.Column
		}
	}

	width := (maxCol + 1) * 2
	vis := make([]rune, width)
	for i := range vis {
		vis[i] = ' '
	}
	return vis
}

// drawStartingCables draws the horizontal branch for every cable that starts at
// this U row.
func drawStartingCables(vis []rune, endpoints []CableEndpoint, routes []CableRoute, u int) {
	for _, ep := range endpoints {
		for _, route := range routes {
			if route.StartU == u && route.Endpoint.DestName == ep.DestName && route.Endpoint.DestPort == ep.DestPort {
				drawStartingCable(vis, route, ep)
				break
			}
		}
	}
}

// drawStartingCable draws a single starting cable's horizontal run and its
// terminating up/down corner.
func drawStartingCable(vis []rune, route CableRoute, ep CableEndpoint) {
	for col := 0; col <= route.Column; col++ {
		pos := col * 2
		if pos >= len(vis) {
			continue
		}
		switch {
		case col < route.Column:
			vis[pos] = '─'
		case ep.GoingUp:
			vis[pos] = '/'
		default:
			vis[pos] = '\\'
		}
	}
}

// drawPassThroughCables draws vertical segments and corners for cables that pass
// through or terminate at this U row.
func drawPassThroughCables(vis []rune, routes []CableRoute, u int) {
	for _, route := range routes {
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
}

// cableLabels returns the "dest:port" labels for cables starting at a U row.
func cableLabels(endpoints []CableEndpoint) []string {
	var labels []string
	for _, ep := range endpoints {
		labels = append(labels, fmt.Sprintf("%s:%s", ep.DestName, ep.DestPort))
	}
	return labels
}
