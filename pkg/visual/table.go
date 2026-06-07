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
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// Column widths for fixed-width table output.
const (
	ColName      = 30
	ColType      = 15
	ColModel     = 30
	ColStatus    = 10
	ColLocation  = 20
	ColRack      = 20
	ColDevice    = 20
	ColBay       = 15
	ColUPos      = 6
	ColCount     = 8
	ColLabel     = 25
	ColCableTyp  = 15
	ColTerm      = 25
	ColSerial    = 20
	ColConnected = 12
	ColRole      = 14
	ColMac       = 18
)

// Pad right-pads s with spaces to width n.
func Pad(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

// Trunc truncates s to max characters, appending "…" if needed.
func Trunc(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return s[:max]
	}
	return s[:max-1] + "…"
}

// Col formats a value into a fixed-width column.
func Col(s string, width int) string {
	return Pad(Trunc(s, width), width)
}

// ResolveLocationName looks up a location name by UUID.
func ResolveLocationName(id uuid.UUID, inv *devicetypes.Inventory) string {
	if id == uuid.Nil || inv == nil || inv.Locations == nil {
		return ""
	}
	if loc, ok := inv.Locations[id]; ok {
		return loc.Name
	}
	return ""
}

// ResolveRackName looks up a rack name by UUID.
func ResolveRackName(id uuid.UUID, inv *devicetypes.Inventory) string {
	if id == uuid.Nil || inv == nil || inv.Racks == nil {
		return ""
	}
	if r, ok := inv.Racks[id]; ok {
		return r.Name
	}
	return ""
}

// ResolveDeviceName looks up a device name by UUID.
func ResolveDeviceName(id uuid.UUID, inv *devicetypes.Inventory) string {
	if id == uuid.Nil || inv == nil {
		return ""
	}
	if inv.Devices != nil {
		if d, ok := inv.Devices[id]; ok {
			return d.Name
		}
	}
	if inv.Modules != nil {
		if m, ok := inv.Modules[id]; ok {
			return m.Name
		}
	}
	return ""
}

// FormatTermination builds a "device:port" string from a device UUID and port name.
func FormatTermination(deviceID uuid.UUID, port string, inv *devicetypes.Inventory) string {
	name := ResolveDeviceName(deviceID, inv)
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

// IntOrDash returns the integer as a string, or "-" if zero.
func IntOrDash(n int) string {
	if n == 0 {
		return "-"
	}
	return strconv.Itoa(n)
}

// JoinNonEmpty joins non-empty strings with sep.
func JoinNonEmpty(parts []string, sep string) string {
	var filtered []string
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) == 0 {
		return ""
	}
	result := filtered[0]
	for _, p := range filtered[1:] {
		result += sep + p
	}
	return result
}

// PrintLocationTable renders locations as a fixed-width table.
func PrintLocationTable(locations []*devicetypes.CaniLocationType, inv *devicetypes.Inventory) {
	header := Col("NAME", ColName) + "  " +
		Col("TYPE", ColType) + "  " +
		Col("STATUS", ColStatus) + "  " +
		Col("CHILDREN", ColCount) + "  " +
		Col("RACKS", ColCount)
	sep := Col(strings.Repeat("-", ColName), ColName) + "  " +
		Col(strings.Repeat("-", ColType), ColType) + "  " +
		Col(strings.Repeat("-", ColStatus), ColStatus) + "  " +
		Col(strings.Repeat("-", ColCount), ColCount) + "  " +
		Col(strings.Repeat("-", ColCount), ColCount)

	fmt.Println(header)
	fmt.Println(sep)
	for _, loc := range locations {
		fmt.Println(
			Col(loc.Name, ColName) + "  " +
				Col(loc.LocationType, ColType) + "  " +
				Col(loc.Status, ColStatus) + "  " +
				Col(strconv.Itoa(len(loc.Children)), ColCount) + "  " +
				Col(strconv.Itoa(len(loc.Racks)), ColCount),
		)
	}
	fmt.Printf("\nTotal: %d location(s)\n", len(locations))
}

// PrintRackTable renders racks as a fixed-width table.
func PrintRackTable(racks []*devicetypes.CaniRackType, inv *devicetypes.Inventory) {
	header := Col("NAME", ColName) + "  " +
		Col("MODEL", ColModel) + "  " +
		Col("STATUS", ColStatus) + "  " +
		Col("LOCATION", ColLocation) + "  " +
		Col("DEVICES", ColCount)
	sep := Col(strings.Repeat("-", ColName), ColName) + "  " +
		Col(strings.Repeat("-", ColModel), ColModel) + "  " +
		Col(strings.Repeat("-", ColStatus), ColStatus) + "  " +
		Col(strings.Repeat("-", ColLocation), ColLocation) + "  " +
		Col(strings.Repeat("-", ColCount), ColCount)

	fmt.Println(header)
	fmt.Println(sep)
	for _, r := range racks {
		locName := ResolveLocationName(r.Location, inv)
		fmt.Println(
			Col(r.Name, ColName) + "  " +
				Col(r.Model, ColModel) + "  " +
				Col(r.Status, ColStatus) + "  " +
				Col(locName, ColLocation) + "  " +
				Col(strconv.Itoa(len(r.Devices)), ColCount),
		)
	}
	fmt.Printf("\nTotal: %d rack(s)\n", len(racks))
}

// PrintDeviceTable renders devices as a fixed-width table.
func PrintDeviceTable(devices []*devicetypes.CaniDeviceType, inv *devicetypes.Inventory, tf TreeFilter) {
	header := Col("NAME", ColName) + "  " +
		Col("TYPE", ColType) + "  " +
		Col("MODEL", ColModel) + "  " +
		Col("STATUS", ColStatus) + "  " +
		Col("RACK", ColRack) + "  " +
		Col("U-POS", ColUPos)
	sep := Col(strings.Repeat("-", ColName), ColName) + "  " +
		Col(strings.Repeat("-", ColType), ColType) + "  " +
		Col(strings.Repeat("-", ColModel), ColModel) + "  " +
		Col(strings.Repeat("-", ColStatus), ColStatus) + "  " +
		Col(strings.Repeat("-", ColRack), ColRack) + "  " +
		Col(strings.Repeat("-", ColUPos), ColUPos)

	if tf.Roles {
		header += "  " + Col("ROLE", ColRole)
		sep += "  " + Col(strings.Repeat("-", ColRole), ColRole)
	}

	fmt.Println(header)
	fmt.Println(sep)
	for _, d := range devices {
		rackName := ResolveRackName(d.Rack, inv)
		uPos := ""
		if d.RackPosition > 0 {
			uPos = strconv.Itoa(d.RackPosition)
		}
		row := Col(d.Name, ColName) + "  " +
			Col(string(d.Type), ColType) + "  " +
			Col(d.Model, ColModel) + "  " +
			Col(d.Status, ColStatus) + "  " +
			Col(rackName, ColRack) + "  " +
			Col(uPos, ColUPos)
		if tf.Roles {
			row += "  " + Col(d.GetRole(), ColRole)
		}
		fmt.Println(row)
	}
	fmt.Printf("\nTotal: %d device(s)\n", len(devices))
}

// PrintModuleTable renders modules as a fixed-width table.
func PrintModuleTable(modules []*devicetypes.CaniModuleType, inv *devicetypes.Inventory) {
	header := Col("NAME", ColName) + "  " +
		Col("TYPE", ColType) + "  " +
		Col("MODEL", ColModel) + "  " +
		Col("STATUS", ColStatus) + "  " +
		Col("DEVICE", ColDevice) + "  " +
		Col("BAY", ColBay)
	sep := Col(strings.Repeat("-", ColName), ColName) + "  " +
		Col(strings.Repeat("-", ColType), ColType) + "  " +
		Col(strings.Repeat("-", ColModel), ColModel) + "  " +
		Col(strings.Repeat("-", ColStatus), ColStatus) + "  " +
		Col(strings.Repeat("-", ColDevice), ColDevice) + "  " +
		Col(strings.Repeat("-", ColBay), ColBay)

	fmt.Println(header)
	fmt.Println(sep)
	for _, m := range modules {
		devName := ResolveDeviceName(m.ParentDevice, inv)
		fmt.Println(
			Col(m.Name, ColName) + "  " +
				Col(string(m.Type), ColType) + "  " +
				Col(m.Model, ColModel) + "  " +
				Col(m.Status, ColStatus) + "  " +
				Col(devName, ColDevice) + "  " +
				Col(m.ModuleBayName, ColBay),
		)
	}
	fmt.Printf("\nTotal: %d module(s)\n", len(modules))
}

// PrintCableTable renders cables as a fixed-width table.
func PrintCableTable(cables []*devicetypes.CaniCableType, inv *devicetypes.Inventory) {
	header := Col("LABEL", ColLabel) + "  " +
		Col("TYPE", ColCableTyp) + "  " +
		Col("STATUS", ColStatus) + "  " +
		Col("A TERMINATION", ColTerm) + "  " +
		Col("B TERMINATION", ColTerm)
	sep := Col(strings.Repeat("-", ColLabel), ColLabel) + "  " +
		Col(strings.Repeat("-", ColCableTyp), ColCableTyp) + "  " +
		Col(strings.Repeat("-", ColStatus), ColStatus) + "  " +
		Col(strings.Repeat("-", ColTerm), ColTerm) + "  " +
		Col(strings.Repeat("-", ColTerm), ColTerm)

	fmt.Println(header)
	fmt.Println(sep)
	for _, c := range cables {
		aTerm := FormatTermination(c.TerminationADevice, c.TerminationAPort, inv)
		bTerm := FormatTermination(c.TerminationBDevice, c.TerminationBPort, inv)
		fmt.Println(
			Col(c.Label, ColLabel) + "  " +
				Col(c.CableType, ColCableTyp) + "  " +
				Col(c.Status, ColStatus) + "  " +
				Col(aTerm, ColTerm) + "  " +
				Col(bTerm, ColTerm),
		)
	}
	fmt.Printf("\nTotal: %d cable(s)\n", len(cables))
}
