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

import "html/template"

// htmlPage is the root data handed to the rack elevation template.
type htmlPage struct {
	Title       string
	Location    string
	RackCount   int
	DeviceCount int
	CableCount  int
	Generated   string
	Racks       []htmlRack
	Switches    []htmlSwitchTable
	ThemeCSS    template.CSS
	CSS         template.CSS
	PeerCSS     template.CSS
}

// htmlRack is one rack column of the elevation. Rows is the flat contents
// table under the frame (devices top first, chassis nodes after their
// chassis); NodeCount is the number of those chassis nodes.
type htmlRack struct {
	Name        string
	Height      int
	Units       []htmlUnit
	Devices     []htmlDevice
	Rows        []htmlListRow
	DeviceCount int
	NodeCount   int
	Occupied    int
}

// htmlUnit is one U row: its U number and its grid row (1 = top of rack).
type htmlUnit struct {
	U   int
	Row int
}

// htmlDevice is one device tile, or one chassis child cell, plus the
// detail-card fields shared by both.
type htmlDevice struct {
	ID           string
	Name         string
	Model        string
	Manufacturer string
	Type         string
	Kind         string
	Status       string
	Position     string
	TagList      string
	Serial       string
	IPv4         string
	Note         string
	Tooltip      string
	Classes      string
	HalfDepth    bool
	Reachable    bool
	Verified     bool
	Unverified   bool
	UStart       int
	UEnd         int
	Height       int
	Row          int
	Cols         int
	Children     []htmlDevice

	Mgmt            *htmlMgmt
	Ports           []htmlPort
	PortsCabled     int
	PortsTotal      int
	Hardware        []htmlHardwareGroup
	HardwareSummary string
	ParentLink      *htmlRef
	ChildLinks      []htmlRef
	PeerIDs         []string
	Rack            string
	CablingTable    bool
}

// htmlRef points at another device's detail card (a chassis, a node, a peer).
type htmlRef struct {
	ID   string
	Name string
	Mgmt *htmlMgmt
}

// htmlMgmt is a validated management-UI link: "iLO 10.6.84.81" -> https://10.6.84.81/.
type htmlMgmt struct {
	Label string
	Host  string
	URL   string
}

// htmlPort is one row of a device's ports table. Via* names the chassis port
// an uncabled node iLO uses instead of its own cable.
type htmlPort struct {
	Name     string
	Hint     string
	PeerID   string
	PeerName string
	PeerPort string
	ViaID    string
	ViaName  string
	ViaPort  string
	VLAN     string
	IP       string
	IPNote   string
	IPURL    string
	MAC      string
	Cabled   bool
}

// htmlHardwareGroup is one summary line of the hardware section:
// "16x 32 GiB DDR4 Samsung (512 GiB)" plus the per-item serial table.
type htmlHardwareGroup struct {
	Count int
	Title string
	Meta  string
	Total string
	Items []htmlHardwareItem
}

// htmlHardwareItem is one FRU or module inside a hardware group.
type htmlHardwareItem struct {
	Slot       string
	PartNumber string
	Serial     string
}

// HasSerials reports whether the nested serial table has anything to show.
func (g htmlHardwareGroup) HasSerials() bool {
	for _, it := range g.Items {
		if it.Serial != "" {
			return true
		}
	}
	return false
}

// htmlSwitchTable is the printable per-switch cabling table.
type htmlSwitchTable struct {
	ID        string
	Name      string
	Where     string
	Cabled    int
	Total     int
	HasVLAN   bool
	Rows      []htmlSwitchRow
	Free      string
	FreeCount int
}

// htmlSwitchRow is one cabled switch port.
type htmlSwitchRow struct {
	Port        string
	Description string
	PeerID      string
	PeerName    string
	PeerPort    string
	VLAN        string
	PeerIP      string
	PeerIPURL   string
}
