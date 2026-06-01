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
	"os"
	"os/exec"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// cableGroup classifies cables into semantic categories.
type cableGroup int

const (
	groupMgmt cableGroup = iota // management: copper, iLO, BMC
	groupHSN                    // HPC fabric: InfiniBand, NDR, OSFP
	groupNet                    // network fabric: Ethernet switches, SFP+, QSFP28
)

// groupLabel returns a display name for the group.
func groupLabel(g cableGroup) string {
	switch g {
	case groupMgmt:
		return "MGMT"
	case groupHSN:
		return "HSN"
	case groupNet:
		return "NET"
	default:
		return "OTHER"
	}
}

// groupColorKey returns the ANSI color key for a cable group.
func groupColorKey(g cableGroup) string {
	switch g {
	case groupMgmt:
		return "yellow"
	case groupHSN:
		return "green"
	default:
		return "magenta"
	}
}

// classifyCable determines the cable group from slug and port names.
//
// Groups:
//   - mgmt: copper Ethernet (Cat5/6, RJ45), iLO/BMC, mgmt ports
//   - hsn:  HPC/InfiniBand fabric (NDR, OSFP, 400G, MPO, HSN ports)
//   - net:  Ethernet switch fabric (SFP+, QSFP28, DAC, AOC)
func classifyCable(slug, portA, portB string) cableGroup {
	s := strings.ToLower(slug)
	pA := strings.ToLower(portA)
	pB := strings.ToLower(portB)

	// Management: copper + BMC/iLO
	if strings.Contains(s, "cat5") || strings.Contains(s, "cat6") || strings.Contains(s, "rj45") ||
		strings.Contains(pA, "mgmt") || strings.Contains(pA, "ilo") ||
		strings.Contains(pB, "mgmt") || strings.Contains(pB, "ilo") {
		return groupMgmt
	}

	// HSN: InfiniBand / HPC fabric
	if strings.Contains(s, "osfp") || strings.Contains(s, "ndr") ||
		strings.Contains(s, "400g") || strings.Contains(s, "mpo") ||
		strings.Contains(s, "infiniband") || strings.Contains(s, "-ib-") ||
		strings.Contains(pA, "hsn") || strings.Contains(pB, "hsn") {
		return groupHSN
	}

	// Network fabric: Ethernet switch interconnects
	if strings.Contains(s, "sfp") || strings.Contains(s, "qsfp") ||
		strings.Contains(s, "dac") || strings.Contains(s, "aoc") ||
		strings.Contains(s, "10g") || strings.Contains(s, "100g") ||
		strings.Contains(s, "25g") || strings.Contains(s, "40g") {
		return groupNet
	}

	return groupNet // default: assume network fabric
}

// interactiveState holds toggles for the interactive routing view.
type interactiveState struct {
	showMgmt   bool
	showHSN    bool
	showNet    bool
	interOnly  bool // when true, hide intra-rack cables
	showLabels bool
	rackIdx    int // current rack index
	rackCount  int // total number of racks
}

// RunInteractiveRouting enters a raw-mode loop for toggle-based cable filtering.
func RunInteractiveRouting(inv *devicetypes.Inventory, opts CompactRenderOptions) error {
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

	// Pre-compute cables per rack (all cables, filtering via toggles).
	allOpts := opts
	allOpts.Verbose = 2
	type rackCables struct {
		view   *CompactRackView
		cables []routingCable
	}
	perRack := make([]rackCables, len(rackViews))
	for i, rv := range rackViews {
		single := []*CompactRackView{rv}
		c := collectRoutingCables(inv, single, allOpts)
		for j := range c {
			if c[j].interRack {
				c[j].topU = rv.Height + 2
			}
		}
		assignRoutingLanes(c)
		perRack[i] = rackCables{rv, c}
	}

	state := interactiveState{
		showMgmt:   true,
		showHSN:    true,
		showNet:    true,
		showLabels: opts.ShowLabels,
		rackIdx:    0,
		rackCount:  len(rackViews),
	}

	restore, err := enableRawMode()
	if err != nil {
		return fmt.Errorf("failed to enable raw mode: %w", err)
	}
	defer restore()

	rc := perRack[state.rackIdx]
	renderFrame([]*CompactRackView{rc.view}, grids, rc.cables, rc.view.Height, opts, state)

	buf := make([]byte, 1)
	for {
		if _, err := os.Stdin.Read(buf); err != nil {
			break
		}
		switch buf[0] {
		case '1', 'm':
			state.showMgmt = !state.showMgmt
		case '2', 'h':
			state.showHSN = !state.showHSN
		case '3', 'o':
			state.showNet = !state.showNet
		case '4', 'i':
			state.interOnly = !state.interOnly
		case 'a':
			state.showMgmt, state.showHSN, state.showNet, state.interOnly = true, true, true, false
		case 'l':
			state.showLabels = !state.showLabels
		case 'n':
			if state.rackIdx < state.rackCount-1 {
				state.rackIdx++
			}
		case 'N':
			if state.rackIdx > 0 {
				state.rackIdx--
			}
		case 'q', 3: // 'q' or Ctrl-C
			fmt.Print("\033[2J\033[H") // clear before exit
			return nil
		default:
			continue
		}
		rc = perRack[state.rackIdx]
		renderFrame([]*CompactRackView{rc.view}, grids, rc.cables, rc.view.Height, opts, state)
	}
	return nil
}

// enableRawMode puts the terminal into raw mode and returns a restore function.
func enableRawMode() (func(), error) {
	cmd := exec.Command("stty", "-g")
	cmd.Stdin = os.Stdin
	saved, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	raw := exec.Command("stty", "-icanon", "-echo", "min", "1")
	raw.Stdin = os.Stdin
	if err := raw.Run(); err != nil {
		return nil, err
	}

	restore := func() {
		r := exec.Command("stty", strings.TrimSpace(string(saved)))
		r.Stdin = os.Stdin
		_ = r.Run()
	}
	return restore, nil
}

// renderFrame clears the screen, draws the diagram with current filters, and prints the status bar.
func renderFrame(
	racks []*CompactRackView,
	grids map[uuid.UUID]map[int]MinimapSlot,
	cables []routingCable,
	maxHeight int,
	opts CompactRenderOptions,
	state interactiveState,
) {
	applyDimming(cables, state)

	cf := newColorFuncs(opts.NoColor)
	lanes := maxRoutingLane(cables) + 1
	visible := countVisible(cables)

	// Clear screen + home cursor
	fmt.Print("\033[2J\033[H")

	// Rack name headers — right-justified, truncated to 4 chars.
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

	// Top border — also show inter-rack cable entry glyphs.
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
			if state.showLabels {
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

	renderStatusBar(cf, state, visible, len(cables))
}

// applyDimming sets the dimmed flag on each cable based on toggle state.
func applyDimming(cables []routingCable, state interactiveState) {
	for i := range cables {
		groupOK := false
		switch cables[i].group {
		case groupMgmt:
			groupOK = state.showMgmt
		case groupHSN:
			groupOK = state.showHSN
		case groupNet:
			groupOK = state.showNet
		}
		if state.interOnly && !cables[i].interRack {
			groupOK = false
		}
		cables[i].dimmed = !groupOK
	}
}

// countVisible returns the number of non-dimmed cables.
func countVisible(cables []routingCable) int {
	n := 0
	for _, c := range cables {
		if !c.dimmed {
			n++
		}
	}
	return n
}

// renderStatusBar prints the interactive toggle bar.
func renderStatusBar(cf colorFuncs, state interactiveState, visible, total int) {
	fmt.Println()
	fmt.Print("  ")
	fmt.Print(toggleStr(cf, "[1]", "MGMT", "blue", state.showMgmt))
	fmt.Print("  ")
	fmt.Print(toggleStr(cf, "[2]", "HSN", "green", state.showHSN))
	fmt.Print("  ")
	fmt.Print(toggleStr(cf, "[3]", "NET", "magenta", state.showNet))
	fmt.Print("  ")
	fmt.Print(toggleStr(cf, "[4]", "inter-only", "yellow", state.interOnly))
	fmt.Print("  ")
	fmt.Print(toggleStr(cf, "[l]", "labels", "white", state.showLabels))
	fmt.Printf("  [a]ll  [q]uit")
	fmt.Printf("  (%d/%d cables)", visible, total)
	fmt.Printf("  rack %d/%d [n]ext [N]prev\n", state.rackIdx+1, state.rackCount)
}

// toggleStr formats a toggle indicator: ■ when on, □ when off.
func toggleStr(cf colorFuncs, key, label, color string, on bool) string {
	indicator := "□"
	if on {
		indicator = cf.colorize("■", color)
	}
	return cf.bold(key) + " " + indicator + " " + label
}
