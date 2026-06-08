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
	deviceRack, visible := buildVisibilityMaps(rackViews)

	var out []routingCable
	for _, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		if opts.CableType != "" && !strings.Contains(strings.ToLower(cable.Slug), strings.ToLower(opts.CableType)) {
			continue
		}
		if rc, ok := routingCableFor(inv, cable, deviceRack, visible, opts.Verbose); ok {
			out = append(out, rc)
		}
	}
	return out
}

// buildVisibilityMaps maps each visible device to its rack ID and records the
// set of visible devices.
func buildVisibilityMaps(rackViews []*CompactRackView) (map[uuid.UUID]uuid.UUID, map[uuid.UUID]bool) {
	deviceRack := make(map[uuid.UUID]uuid.UUID)
	visible := make(map[uuid.UUID]bool)
	for _, rv := range rackViews {
		for _, devID := range rv.Rack.Devices {
			visible[devID] = true
			deviceRack[devID] = rv.RackID
		}
	}
	return deviceRack, visible
}

// routingCableFor builds a routingCable for a cable with at least one visible,
// rack-positioned endpoint. Intra-rack cables are skipped when verbose < 2.
func routingCableFor(inv *devicetypes.Inventory, cable *devicetypes.CaniCableType, deviceRack map[uuid.UUID]uuid.UUID, visible map[uuid.UUID]bool, verbose int) (routingCable, bool) {
	if !visible[cable.TerminationADevice] && !visible[cable.TerminationBDevice] {
		return routingCable{}, false
	}
	devA := inv.Devices[cable.TerminationADevice]
	devB := inv.Devices[cable.TerminationBDevice]
	if devA == nil || devB == nil || devA.RackPosition <= 0 || devB.RackPosition <= 0 {
		return routingCable{}, false
	}

	sameRack := deviceRack[cable.TerminationADevice] == deviceRack[cable.TerminationBDevice]
	if sameRack && verbose < 2 {
		return routingCable{}, false
	}

	rc := routingCable{
		labelA:    truncateName(devA.Name, 20) + ":" + cable.TerminationAPort,
		labelB:    truncateName(devB.Name, 20) + ":" + cable.TerminationBPort,
		group:     classifyCable(cable.Slug, cable.TerminationAPort, cable.TerminationBPort),
		interRack: !sameRack,
	}
	setRoutingEndpoints(&rc, cable, devA.RackPosition, devB.RackPosition)
	if !sameRack {
		setInterRackEndpoints(&rc, inv, cable, devA, devB, visible)
	}
	return rc, true
}

// setRoutingEndpoints orders the cable's endpoints top/bottom by U position.
func setRoutingEndpoints(rc *routingCable, cable *devicetypes.CaniCableType, uA, uB int) {
	if uA >= uB {
		rc.topU, rc.botU, rc.topIsA = uA, uB, true
		rc.portAtTop = cable.TerminationAPort
		rc.portAtBot = cable.TerminationBPort
		return
	}
	rc.topU, rc.botU, rc.topIsA = uB, uA, false
	rc.portAtTop = cable.TerminationBPort
	rc.portAtBot = cable.TerminationAPort
}

// setInterRackEndpoints fixes endpoints so botU/portAtBot describe the local
// device, portAtTop the remote port, and remoteRack the far rack name.
func setInterRackEndpoints(rc *routingCable, inv *devicetypes.Inventory, cable *devicetypes.CaniCableType, devA, devB *devicetypes.CaniDeviceType, visible map[uuid.UUID]bool) {
	if visible[cable.TerminationADevice] {
		rc.localU = devA.RackPosition
		rc.botU = devA.RackPosition
		rc.portAtBot = cable.TerminationAPort
		rc.portAtTop = cable.TerminationBPort
		rc.localIsA = true
		if rr := inv.Racks[devB.Rack]; rr != nil {
			rc.remoteRack = rr.Name
		}
		return
	}
	rc.localU = devB.RackPosition
	rc.botU = devB.RackPosition
	rc.portAtBot = cable.TerminationBPort
	rc.portAtTop = cable.TerminationAPort
	rc.localIsA = false
	if rr := inv.Racks[devA.Rack]; rr != nil {
		rc.remoteRack = rr.Name
	}
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
