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

// routingCable is a cable with an assigned wiring-diagram lane.
type routingCable struct {
	topU      int        // higher U position (rendered first)
	botU      int        // lower U position
	lane      int        // vertical lane column in the wiring channel
	labelA    string     // "device:port" for A termination
	labelB    string     // "device:port" for B termination
	topIsA    bool       // true when A-end sits at topU
	group     cableGroup // semantic classification
	interRack bool       // true when cable crosses racks
	dimmed    bool       // true when filtered out (rendered gray)

	// Endpoint annotation fields (Approach 1: port-pair annotations).
	portAtTop  string // raw port name at topU device
	portAtBot  string // raw port name at botU device
	remoteRack string // remote rack name (empty for intra-rack)
	localU     int    // U position of the local device (for inter-rack annotation placement)
	localIsA   bool   // true when A-end is the local device (outgoing)
}

// RenderRoutingView renders minimap-width racks with a wiring diagram.
// Without a RackFilter it renders one rack at a time sequentially.
func RenderRoutingView(inv *devicetypes.Inventory, opts CompactRenderOptions) error {
	if inv == nil {
		return fmt.Errorf("inventory is nil")
	}

	rackViews := buildCompactRackViews(inv, opts.RackFilter)
	if len(rackViews) == 0 {
		fmt.Println("No racks found in inventory")
		return nil
	}

	grids := make(map[uuid.UUID]map[int]MinimapSlot, len(rackViews))
	for _, rv := range rackViews {
		grids[rv.RackID] = buildMinimapGrid(inv, rv)
	}

	if opts.Verbose >= 1 {
		printRoutingLegend(opts)
	}

	// Show all cables regardless of verbose level.
	allOpts := opts
	allOpts.Verbose = 2

	for i, rv := range rackViews {
		single := []*CompactRackView{rv}
		cables := collectRoutingCables(inv, single, allOpts)
		// Inter-rack cables enter from above the rack, not at a false U.
		for j := range cables {
			if cables[j].interRack {
				cables[j].topU = rv.Height + 2
			}
		}
		assignRoutingLanes(cables)
		renderRoutingDiagram(single, grids, cables, rv.Height, opts)
		if i < len(rackViews)-1 {
			fmt.Println()
		}
	}
	return nil
}

// collectRoutingCables gathers cables whose endpoints are in visible racks.
func collectRoutingCables(inv *devicetypes.Inventory, rackViews []*CompactRackView, opts CompactRenderOptions) []routingCable {
	// Map each device to its rack ID for inter-rack detection.
	deviceRack := make(map[uuid.UUID]uuid.UUID)
	visible := make(map[uuid.UUID]bool)
	for _, rv := range rackViews {
		for _, devID := range rv.Rack.Devices {
			visible[devID] = true
			deviceRack[devID] = rv.RackID
		}
	}

	var out []routingCable
	for _, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		if opts.CableType != "" && !strings.Contains(strings.ToLower(cable.Slug), strings.ToLower(opts.CableType)) {
			continue
		}
		if !visible[cable.TerminationADevice] && !visible[cable.TerminationBDevice] {
			continue
		}
		devA := inv.Devices[cable.TerminationADevice]
		devB := inv.Devices[cable.TerminationBDevice]
		if devA == nil || devB == nil || devA.RackPosition <= 0 || devB.RackPosition <= 0 {
			continue
		}

		rackA := deviceRack[cable.TerminationADevice]
		rackB := deviceRack[cable.TerminationBDevice]
		sameRack := rackA == rackB

		if sameRack && opts.Verbose < 2 {
			continue
		}

		uA, uB := devA.RackPosition, devB.RackPosition
		rc := routingCable{
			labelA:    truncateName(devA.Name, 20) + ":" + cable.TerminationAPort,
			labelB:    truncateName(devB.Name, 20) + ":" + cable.TerminationBPort,
			group:     classifyCable(cable.Slug, cable.TerminationAPort, cable.TerminationBPort),
			interRack: !sameRack,
		}
		if uA >= uB {
			rc.topU, rc.botU, rc.topIsA = uA, uB, true
			rc.portAtTop = cable.TerminationAPort
			rc.portAtBot = cable.TerminationBPort
		} else {
			rc.topU, rc.botU, rc.topIsA = uB, uA, false
			rc.portAtTop = cable.TerminationBPort
			rc.portAtBot = cable.TerminationAPort
		}
		if !sameRack {
			// Inter-rack: fix endpoints so botU = local device,
			// portAtBot = local port, portAtTop = remote port.
			if visible[cable.TerminationADevice] {
				rc.localU = devA.RackPosition
				rc.botU = devA.RackPosition
				rc.portAtBot = cable.TerminationAPort
				rc.portAtTop = cable.TerminationBPort
				rc.localIsA = true
				if rr := inv.Racks[devB.Rack]; rr != nil {
					rc.remoteRack = rr.Name
				}
			} else {
				rc.localU = devB.RackPosition
				rc.botU = devB.RackPosition
				rc.portAtBot = cable.TerminationBPort
				rc.portAtTop = cable.TerminationAPort
				rc.localIsA = false
				if rr := inv.Racks[devA.Rack]; rr != nil {
					rc.remoteRack = rr.Name
				}
			}
		}
		out = append(out, rc)
	}
	return out
}

// assignRoutingLanes uses greedy interval colouring so non-overlapping cables share lanes.
func assignRoutingLanes(cables []routingCable) {
	sort.Slice(cables, func(i, j int) bool {
		if cables[i].topU != cables[j].topU {
			return cables[i].topU > cables[j].topU
		}
		return cables[i].botU > cables[j].botU
	})
	for i := range cables {
		used := make(map[int]bool)
		for j := 0; j < i; j++ {
			if cables[j].topU >= cables[i].botU && cables[j].botU <= cables[i].topU {
				used[cables[j].lane] = true
			}
		}
		for lane := 0; ; lane++ {
			if !used[lane] {
				cables[i].lane = lane
				break
			}
		}
	}
}

// renderRoutingDiagram draws the full wiring diagram to stdout.
func renderRoutingDiagram(
	racks []*CompactRackView,
	grids map[uuid.UUID]map[int]MinimapSlot,
	cables []routingCable,
	maxHeight int,
	opts CompactRenderOptions,
) {
	cf := newColorFuncs(opts.NoColor)
	lanes := maxRoutingLane(cables) + 1

	// Rack name headers — right-justified, truncated to 4 chars to match rack column width.
	fmt.Print(strings.Repeat(" ", minimapUCol))
	for _, rv := range racks {
		name := rv.Rack.Name
		if len(name) > 4 {
			name = name[len(name)-4:]
		}
		fmt.Printf(" %4s", cf.bold(name))
	}
	if lanes > 0 {
		if entryStr := buildRoutingLaneStr(maxHeight+2, cables, lanes, cf); entryStr != "" {
			fmt.Print(" " + entryStr)
		}
	}
	fmt.Println()

	// Top border
	fmt.Print(strings.Repeat(" ", minimapUCol))
	for range racks {
		fmt.Print(" ┌──┐")
	}
	if lanes > 0 {
		if borderStr := buildRoutingLaneStr(maxHeight+1, cables, lanes, cf); borderStr != "" {
			fmt.Print(" " + borderStr)
		}
	}
	fmt.Println()

	for u := maxHeight; u >= 1; u-- {
		if uHasEndpoint(u, cables) {
			fmt.Print(cf.bold(fmt.Sprintf("%2d", u)) + "  ")
		} else {
			fmt.Printf("%2d  ", u)
		}
		for _, rv := range racks {
			if u > rv.Height {
				fmt.Print("     ")
				continue
			}
			slot := grids[rv.RackID][u]
			c1 := cf.colorize(string(slot.Char1), slot.Color1)
			c2 := cf.colorize(string(slot.Char2), slot.Color2)
			fmt.Printf(" │%s%s│", c1, c2)
		}

		lStr := buildRoutingLaneStr(u, cables, lanes, cf)
		if lStr != "" {
			fmt.Print(" " + lStr)
			if opts.ShowLabels {
				if ann := buildColoredAnnotation(u, cables, cf); ann != "" {
					fmt.Print("  " + ann)
				}
			}
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

// maxRoutingLane returns the highest assigned lane, or -1 if empty.
func maxRoutingLane(cables []routingCable) int {
	m := -1
	for _, c := range cables {
		if c.lane > m {
			m = c.lane
		}
	}
	return m
}

// buildRoutingLaneStr produces the lane-character segment for one U row.
func buildRoutingLaneStr(u int, cables []routingCable, lanes int, cf colorFuncs) string {
	if lanes <= 0 {
		return ""
	}
	type laneCell struct {
		ch    rune
		color string
	}
	cells := make([]laneCell, lanes)
	active := false
	for _, c := range cables {
		if u < c.botU || u > c.topU {
			continue
		}
		active = true
		var ch rune
		switch {
		case c.topU == c.botU:
			ch = '─'
		case u == c.topU:
			if c.interRack && c.localIsA {
				ch = '┌' // outgoing: originates here, exits up
			} else {
				ch = '┐' // incoming or intra-rack start
			}
		case u == c.botU:
			ch = '┘' // connection point at device
		default:
			ch = '│'
		}
		color := groupColorKey(c.group)
		if c.dimmed {
			color = "gray"
		}
		cells[c.lane] = laneCell{ch, color}
	}
	if !active {
		return ""
	}
	var sb strings.Builder
	for _, lc := range cells {
		if lc.ch == 0 || lc.ch == ' ' {
			sb.WriteByte(' ')
		} else {
			sb.WriteString(cf.colorize(string(lc.ch), lc.color))
		}
	}
	return sb.String()
}
