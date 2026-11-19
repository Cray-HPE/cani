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

// htmlListRow is one row of the "Rack contents" table under a rack: the
// textual twin of the elevation for reader views and screen readers, which
// drop or skip the tile buttons. A chassis node follows its chassis with
// Parent set and the chassis' U range.
type htmlListRow struct {
	Name   string
	Model  string
	Kind   string
	Status string
	Mgmt   *htmlMgmt
	U      string
	Parent string
}

// listRows flattens the rack-mounted devices (already top of rack first)
// and their chassis nodes into table rows.
func listRows(devices []htmlDevice) []htmlListRow {
	var rows []htmlListRow
	for _, dev := range devices {
		u := formatURange(dev.UStart, dev.UEnd)
		rows = append(rows, listRow(dev, u, ""))
		for _, child := range dev.Children {
			rows = append(rows, listRow(child, u, dev.Name))
		}
	}
	return rows
}

// listRow copies the fields the contents table shows.
func listRow(dev htmlDevice, u, parent string) htmlListRow {
	return htmlListRow{
		Name:   dev.Name,
		Model:  dev.Model,
		Kind:   dev.Kind,
		Status: dev.Status,
		Mgmt:   dev.Mgmt,
		U:      u,
		Parent: parent,
	}
}

// countNodes counts the chassis nodes of a rack's devices.
func countNodes(devices []htmlDevice) int {
	n := 0
	for _, dev := range devices {
		n += len(dev.Children)
	}
	return n
}
