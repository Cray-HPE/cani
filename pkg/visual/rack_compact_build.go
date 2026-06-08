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
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

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
	deviceToRackIdx, deviceToPosition := buildDeviceRackMaps(inv, rackViews)

	var cables []InterRackCable
	for _, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		if cableTypeFilter != "" && !strings.Contains(strings.ToLower(cable.Slug), strings.ToLower(cableTypeFilter)) {
			continue
		}
		if c, ok := interRackCableFor(inv, cable, deviceToRackIdx, deviceToPosition, includeAll); ok {
			cables = append(cables, c)
		}
	}

	return cables
}

// buildDeviceRackMaps maps every device to its rack column index and rack U
// position across the given rack views.
func buildDeviceRackMaps(inv *devicetypes.Inventory, rackViews []*CompactRackView) (map[uuid.UUID]int, map[uuid.UUID]int) {
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
	return deviceToRackIdx, deviceToPosition
}

// interRackCableFor builds an InterRackCable for a cable whose endpoints are
// both known, returning ok=false for unknown endpoints or (by default)
// intra-rack cables.
func interRackCableFor(inv *devicetypes.Inventory, cable *devicetypes.CaniCableType, deviceToRackIdx, deviceToPosition map[uuid.UUID]int, includeAll bool) (InterRackCable, bool) {
	deviceA := cable.TerminationADevice
	deviceB := cable.TerminationBDevice

	rackIdxA, okA := deviceToRackIdx[deviceA]
	rackIdxB, okB := deviceToRackIdx[deviceB]
	if !okA || !okB {
		return InterRackCable{}, false
	}

	// Only include inter-rack cables by default
	if !includeAll && rackIdxA == rackIdxB {
		return InterRackCable{}, false
	}

	return InterRackCable{
		Cable:       cable,
		RackAIndex:  rackIdxA,
		RackBIndex:  rackIdxB,
		PositionA:   deviceToPosition[deviceA],
		PositionB:   deviceToPosition[deviceB],
		DeviceAName: deviceName(inv, deviceA),
		DeviceBName: deviceName(inv, deviceB),
		PortA:       cable.TerminationAPort,
		PortB:       cable.TerminationBPort,
	}, true
}

// deviceName returns the device's name, or an empty string when unknown.
func deviceName(inv *devicetypes.Inventory, id uuid.UUID) string {
	if d := inv.Devices[id]; d != nil {
		return d.Name
	}
	return ""
}

// filterCablesForRow returns cables whose endpoints both fall in [rowStart, rowEnd).
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
