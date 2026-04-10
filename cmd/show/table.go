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
package show

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// Column widths for fixed-width table output.
const (
	colName     = 30
	colType     = 15
	colModel    = 30
	colStatus   = 10
	colLocation = 20
	colRack     = 20
	colDevice   = 20
	colBay      = 15
	colUPos     = 6
	colCount    = 8
	colLabel    = 25
	colCableTyp = 15
	colTerm     = 25
)

// pad right-pads s with spaces to width n.
func pad(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

// trunc truncates s to max characters, appending "…" if needed.
func trunc(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return s[:max]
	}
	return s[:max-1] + "…"
}

// col formats a value into a fixed-width column.
func col(s string, width int) string {
	return pad(trunc(s, width), width)
}

// resolveLocationName looks up a location name by UUID.
func resolveLocationName(id uuid.UUID, inv *devicetypes.Inventory) string {
	if id == uuid.Nil || inv == nil || inv.Locations == nil {
		return ""
	}
	if loc, ok := inv.Locations[id]; ok {
		return loc.Name
	}
	return ""
}

// resolveRackName looks up a rack name by UUID.
func resolveRackName(id uuid.UUID, inv *devicetypes.Inventory) string {
	if id == uuid.Nil || inv == nil || inv.Racks == nil {
		return ""
	}
	if r, ok := inv.Racks[id]; ok {
		return r.Name
	}
	return ""
}

// resolveDeviceName looks up a device name by UUID.
func resolveDeviceName(id uuid.UUID, inv *devicetypes.Inventory) string {
	if id == uuid.Nil || inv == nil || inv.Devices == nil {
		return ""
	}
	if d, ok := inv.Devices[id]; ok {
		return d.Name
	}
	return ""
}

// printLocationTable renders locations as a fixed-width table.
func printLocationTable(locations []*devicetypes.CaniLocationType, inv *devicetypes.Inventory) {
	header := col("NAME", colName) + "  " +
		col("TYPE", colType) + "  " +
		col("STATUS", colStatus) + "  " +
		col("CHILDREN", colCount) + "  " +
		col("RACKS", colCount)
	sep := col(strings.Repeat("-", colName), colName) + "  " +
		col(strings.Repeat("-", colType), colType) + "  " +
		col(strings.Repeat("-", colStatus), colStatus) + "  " +
		col(strings.Repeat("-", colCount), colCount) + "  " +
		col(strings.Repeat("-", colCount), colCount)

	fmt.Println(header)
	fmt.Println(sep)
	for _, loc := range locations {
		fmt.Println(
			col(loc.Name, colName) + "  " +
				col(loc.LocationType, colType) + "  " +
				col(loc.Status, colStatus) + "  " +
				col(strconv.Itoa(len(loc.Children)), colCount) + "  " +
				col(strconv.Itoa(len(loc.Racks)), colCount),
		)
	}
	fmt.Printf("\nTotal: %d location(s)\n", len(locations))
}

// printRackTable renders racks as a fixed-width table.
func printRackTable(racks []*devicetypes.CaniRackType, inv *devicetypes.Inventory) {
	header := col("NAME", colName) + "  " +
		col("MODEL", colModel) + "  " +
		col("STATUS", colStatus) + "  " +
		col("LOCATION", colLocation) + "  " +
		col("DEVICES", colCount)
	sep := col(strings.Repeat("-", colName), colName) + "  " +
		col(strings.Repeat("-", colModel), colModel) + "  " +
		col(strings.Repeat("-", colStatus), colStatus) + "  " +
		col(strings.Repeat("-", colLocation), colLocation) + "  " +
		col(strings.Repeat("-", colCount), colCount)

	fmt.Println(header)
	fmt.Println(sep)
	for _, r := range racks {
		locName := resolveLocationName(r.Location, inv)
		fmt.Println(
			col(r.Name, colName) + "  " +
				col(r.Model, colModel) + "  " +
				col(r.Status, colStatus) + "  " +
				col(locName, colLocation) + "  " +
				col(strconv.Itoa(len(r.Devices)), colCount),
		)
	}
	fmt.Printf("\nTotal: %d rack(s)\n", len(racks))
}

// printDeviceTable renders devices as a fixed-width table.
func printDeviceTable(devices []*devicetypes.CaniDeviceType, inv *devicetypes.Inventory) {
	header := col("NAME", colName) + "  " +
		col("TYPE", colType) + "  " +
		col("MODEL", colModel) + "  " +
		col("STATUS", colStatus) + "  " +
		col("RACK", colRack) + "  " +
		col("U-POS", colUPos)
	sep := col(strings.Repeat("-", colName), colName) + "  " +
		col(strings.Repeat("-", colType), colType) + "  " +
		col(strings.Repeat("-", colModel), colModel) + "  " +
		col(strings.Repeat("-", colStatus), colStatus) + "  " +
		col(strings.Repeat("-", colRack), colRack) + "  " +
		col(strings.Repeat("-", colUPos), colUPos)

	fmt.Println(header)
	fmt.Println(sep)
	for _, d := range devices {
		rackName := resolveRackName(d.Rack, inv)
		uPos := ""
		if d.RackPosition > 0 {
			uPos = strconv.Itoa(d.RackPosition)
		}
		fmt.Println(
			col(d.Name, colName) + "  " +
				col(string(d.Type), colType) + "  " +
				col(d.Model, colModel) + "  " +
				col(d.Status, colStatus) + "  " +
				col(rackName, colRack) + "  " +
				col(uPos, colUPos),
		)
	}
	fmt.Printf("\nTotal: %d device(s)\n", len(devices))
}

// printModuleTable renders modules as a fixed-width table.
func printModuleTable(modules []*devicetypes.CaniModuleType, inv *devicetypes.Inventory) {
	header := col("NAME", colName) + "  " +
		col("TYPE", colType) + "  " +
		col("MODEL", colModel) + "  " +
		col("STATUS", colStatus) + "  " +
		col("DEVICE", colDevice) + "  " +
		col("BAY", colBay)
	sep := col(strings.Repeat("-", colName), colName) + "  " +
		col(strings.Repeat("-", colType), colType) + "  " +
		col(strings.Repeat("-", colModel), colModel) + "  " +
		col(strings.Repeat("-", colStatus), colStatus) + "  " +
		col(strings.Repeat("-", colDevice), colDevice) + "  " +
		col(strings.Repeat("-", colBay), colBay)

	fmt.Println(header)
	fmt.Println(sep)
	for _, m := range modules {
		devName := resolveDeviceName(m.ParentDevice, inv)
		fmt.Println(
			col(m.Name, colName) + "  " +
				col(string(m.Type), colType) + "  " +
				col(m.Model, colModel) + "  " +
				col(m.Status, colStatus) + "  " +
				col(devName, colDevice) + "  " +
				col(m.ModuleBayName, colBay),
		)
	}
	fmt.Printf("\nTotal: %d module(s)\n", len(modules))
}

// printCableTableFromTypes renders cables (CaniCableType) as a fixed-width table.
func printCableTableFromTypes(cables []*devicetypes.CaniCableType, inv *devicetypes.Inventory) {
	header := col("LABEL", colLabel) + "  " +
		col("TYPE", colCableTyp) + "  " +
		col("STATUS", colStatus) + "  " +
		col("A TERMINATION", colTerm) + "  " +
		col("B TERMINATION", colTerm)
	sep := col(strings.Repeat("-", colLabel), colLabel) + "  " +
		col(strings.Repeat("-", colCableTyp), colCableTyp) + "  " +
		col(strings.Repeat("-", colStatus), colStatus) + "  " +
		col(strings.Repeat("-", colTerm), colTerm) + "  " +
		col(strings.Repeat("-", colTerm), colTerm)

	fmt.Println(header)
	fmt.Println(sep)
	for _, c := range cables {
		aTerm := formatTermination(c.TerminationADevice, c.TerminationAPort, inv)
		bTerm := formatTermination(c.TerminationBDevice, c.TerminationBPort, inv)
		fmt.Println(
			col(c.Label, colLabel) + "  " +
				col(c.CableType, colCableTyp) + "  " +
				col(c.Status, colStatus) + "  " +
				col(aTerm, colTerm) + "  " +
				col(bTerm, colTerm),
		)
	}
	fmt.Printf("\nTotal: %d cable(s)\n", len(cables))
}

// formatTermination builds a "device:port" string from a device UUID and port name.
func formatTermination(deviceID uuid.UUID, port string, inv *devicetypes.Inventory) string {
	name := resolveDeviceName(deviceID, inv)
	if name == "" && port == "" {
		return "-"
	}
	if port == "" {
		return name
	}
	if name == "" {
		return port
	}
	return name + ":" + port
}
