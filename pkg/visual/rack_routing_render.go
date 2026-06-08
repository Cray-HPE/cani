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

	"github.com/google/uuid"
)

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

	printRoutingHeader(racks, cables, maxHeight, lanes, cf)
	printRoutingTopBorder(racks, cables, maxHeight, lanes, cf)

	for u := maxHeight; u >= 1; u-- {
		renderRoutingURow(u, racks, grids, cables, lanes, opts, cf)
	}

	printRoutingBottomBorder(racks)
}

// printRoutingHeader prints the right-justified rack-name headers (truncated to
// 4 chars) plus the cable entry lane above the racks.
func printRoutingHeader(racks []*CompactRackView, cables []routingCable, maxHeight, lanes int, cf colorFuncs) {
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
}

// printRoutingTopBorder prints the rack top borders plus the lane segment that
// sits just above the first U row.
func printRoutingTopBorder(racks []*CompactRackView, cables []routingCable, maxHeight, lanes int, cf colorFuncs) {
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
}

// renderRoutingURow renders a single U row: the U label, each rack's minimap
// slot, and the wiring-lane segment with optional annotations.
func renderRoutingURow(u int, racks []*CompactRackView, grids map[uuid.UUID]map[int]MinimapSlot, cables []routingCable, lanes int, opts CompactRenderOptions, cf colorFuncs) {
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

// printRoutingBottomBorder prints the rack bottom borders.
func printRoutingBottomBorder(racks []*CompactRackView) {
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
		color := groupColorKey(c.group)
		if c.dimmed {
			color = "gray"
		}
		cells[c.lane] = laneCell{routingLaneChar(c, u), color}
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

// routingLaneChar selects the box-drawing rune for a cable at U row u.
func routingLaneChar(c routingCable, u int) rune {
	switch {
	case c.topU == c.botU:
		return '─'
	case u == c.topU:
		if c.interRack && c.localIsA {
			return '┌' // outgoing: originates here, exits up
		}
		return '┐' // incoming or intra-rack start
	case u == c.botU:
		return '┘' // connection point at device
	default:
		return '│'
	}
}
