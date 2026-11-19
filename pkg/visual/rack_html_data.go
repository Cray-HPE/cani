/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
	"strings"
	"time"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// htmlBuilder carries the shared joins and the page-level collections
// (switch tables, shown devices) while racks are converted one by one.
type htmlBuilder struct {
	idx      *htmlIndex
	switches []htmlSwitchTable
	shown    map[uuid.UUID]bool
}

// buildHTMLPage assembles the template data for every rack matching filter.
func buildHTMLPage(inv *devicetypes.Inventory, filter string, now time.Time) htmlPage {
	page := htmlPage{Title: "Rack elevation", Generated: now.Format("2006-01-02 15:04 MST")}
	b := &htmlBuilder{idx: newHTMLIndex(inv), shown: map[uuid.UUID]bool{}}
	var locations []string
	for _, rv := range buildCompactRackViews(inv, filter) {
		rack := b.buildRack(rv)
		page.Racks = append(page.Racks, rack)
		page.DeviceCount += rack.DeviceCount
		locations = appendUnique(locations, htmlLocationPath(inv, rv.Rack.Location))
	}
	page.RackCount = len(page.Racks)
	page.Location = strings.Join(locations, "; ")
	page.CableCount = b.idx.cableCount(b.shown)
	page.Switches = b.switches
	page.PeerCSS = highlightCSS(pageDevices(page.Racks))
	return page
}

// appendUnique appends s unless it is empty or already listed.
func appendUnique(list []string, s string) []string {
	if s == "" {
		return list
	}
	for _, have := range list {
		if have == s {
			return list
		}
	}
	return append(list, s)
}

// pageDevices flattens the rack-mounted devices of every rack.
func pageDevices(racks []htmlRack) []htmlDevice {
	var devices []htmlDevice
	for _, rack := range racks {
		devices = append(devices, rack.Devices...)
	}
	return devices
}

// buildRack converts a compact rack view into template rows and tiles and
// collects the cabling table of every switch in it, top of rack first.
func (b *htmlBuilder) buildRack(rv *CompactRackView) htmlRack {
	rack := htmlRack{
		Name:        rv.Rack.Name,
		Height:      rv.Height,
		DeviceCount: len(rv.Rack.Devices),
		Occupied:    countOccupiedU(rv),
	}
	for u := rv.Height; u >= 1; u-- {
		rack.Units = append(rack.Units, htmlUnit{U: u, Row: rv.Height - u + 1})
	}
	for _, pos := range sortedPositionsDesc(rv.Devices) {
		dev := b.buildDevice(rv.Devices[pos], rv)
		if dev.Height < 1 {
			continue
		}
		rack.Devices = append(rack.Devices, dev)
		b.addSwitchTable(rv.Devices[pos], rack.Name+" "+formatURange(dev.UStart, dev.UEnd))
	}
	rack.Rows = listRows(rack.Devices)
	rack.NodeCount = countNodes(rack.Devices)
	return rack
}

// addSwitchTable records the cabling table of a switch that has cables.
func (b *htmlBuilder) addSwitchTable(dev *devicetypes.CaniDeviceType, where string) {
	if !isSwitchKind(dev.Type) {
		return
	}
	if table := buildSwitchTable(b.idx, dev, where); table != nil {
		b.switches = append(b.switches, *table)
	}
}

// buildDevice builds a rack-mounted tile. RackPosition is the lowest
// occupied U and the device extends upward; the U range is clamped to the rack.
func (b *htmlBuilder) buildDevice(dev *devicetypes.CaniDeviceType, rv *CompactRackView) htmlDevice {
	d := b.newDevice(dev, rv.Rack.Name)
	d.UStart = dev.RackPosition
	d.UEnd = min(dev.RackPosition+dev.GetUHeight()-1, rv.Height)
	d.Height = d.UEnd - d.UStart + 1
	d.Row = rv.Height - d.UEnd + 1
	d.Position = formatURange(d.UStart, d.UEnd)
	if d.HalfDepth {
		d.Position += ", half depth"
	}
	for _, child := range ResolveDeviceChildren(dev.Children, b.idx.inv) {
		c := b.newDevice(child, rv.Rack.Name)
		c.Position = formatURange(d.UStart, d.UEnd)
		c.ParentLink = &htmlRef{ID: d.ID, Name: d.Name}
		c.Tooltip = htmlTooltip(c)
		c.Classes = htmlClasses(c)
		d.Children = append(d.Children, c)
		d.ChildLinks = append(d.ChildLinks, htmlRef{ID: c.ID, Name: c.Name, Mgmt: c.Mgmt})
	}
	d.Cols = childColumns(len(d.Children), d.Height)
	d.Tooltip = htmlTooltip(d)
	d.Classes = htmlClasses(d)
	return d
}

// newDevice copies the identity fields of a device and attaches the card
// content: management link, ports, hardware and cable peers.
func (b *htmlBuilder) newDevice(dev *devicetypes.CaniDeviceType, rack string) htmlDevice {
	b.shown[dev.ID] = true
	d := htmlDevice{
		ID:           dev.ID.String(),
		Name:         dev.Name,
		Model:        dev.Model,
		Manufacturer: dev.Manufacturer,
		Type:         string(dev.Type),
		Kind:         htmlKind(dev.Type),
		Status:       dev.Status,
		TagList:      strings.Join(dev.Tags, ", "),
		Serial:       dev.Serial,
		IPv4:         htmlIPv4(b.idx.inv, dev.PrimaryIPv4),
		Note:         dev.Description,
		HalfDepth:    !dev.IsFullDepth,
		Reachable:    hasTag(dev.Tags, "reachable"),
		Verified:     hasTag(dev.Tags, "verified"),
		Unverified:   hasTag(dev.Tags, "unverified"),
		Mgmt:         mgmtLink(b.idx, dev),
		PeerIDs:      b.idx.peersOf(dev.ID),
		PortsTotal:   len(dev.Interfaces),
		Rack:         rack,
	}
	d.Ports, d.PortsCabled = buildPorts(b.idx, dev)
	d.Hardware, d.HardwareSummary = buildHardware(b.idx.inv, dev)
	d.CablingTable = isSwitchKind(dev.Type) && d.PortsCabled > 0
	return d
}
