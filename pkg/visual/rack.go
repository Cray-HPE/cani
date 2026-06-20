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
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

const (
	// Box drawing characters
	boxTopLeft     = "┌"
	boxTopRight    = "┐"
	boxBottomLeft  = "└"
	boxBottomRight = "┘"
	boxHorizontal  = "─"
	boxVertical    = "│"
	boxTeeDown     = "┬"
	boxTeeUp       = "┴"
	boxTeeRight    = "├"
	boxTeeLeft     = "┤"
	boxCross       = "┼"

	// Device markers
	markerOccupied  = "█"
	markerContinued = "▓"
	markerEmpty     = "░"

	// Layout constants (80 char terminal width)
	terminalWidth  = 80
	uColumnWidth   = 5  // "U48 " or " U1 "
	separatorWidth = 1  // "│"
	contentWidth   = 70 // Device name area
	paddingWidth   = 2  // Space around content
)

// Shared format string and placeholder values.
const (
	fmtQuint   = "%s%s%s%s%s\n"
	fmtTriple  = "%s%s%s\n"
	valUnknown = "unknown"
)

// rackPalette holds the color functions used when rendering racks.
type rackPalette struct {
	green, gray, cyan, yellow, red, bold func(string) string
}

// newRackPalette builds a rackPalette honoring the NoColor option.
func newRackPalette(noColor bool) rackPalette {
	mk := func(code string) func(string) string {
		return func(s string) string {
			if noColor {
				return s
			}
			return code + s + ColorReset
		}
	}
	return rackPalette{
		green:  mk(ColorGreen),
		gray:   mk(ColorGray),
		cyan:   mk(ColorCyan),
		yellow: mk(ColorYellow),
		red:    mk(ColorRed),
		bold:   mk(ColorBold),
	}
}

// BuildRackVisualization creates a RackView from inventory for a specific rack
func BuildRackVisualization(inv *devicetypes.Inventory, rackID uuid.UUID) (*RackView, error) {
	rack, ok := inv.Racks[rackID]
	if !ok {
		return nil, fmt.Errorf("rack %s not found in inventory", rackID)
	}

	// Determine rack height
	height := getRackHeight(rack)

	rackView := NewRackView(rack, height)

	// Process devices mounted directly in the rack (Parent == rackID).
	for _, device := range inv.GetDevicesInRack(rackID) {
		if device == nil {
			continue
		}

		// Get rack position from dedicated field
		position := device.RackPosition

		// Get u_height: first check Properties, then look up from device type definition
		uHeight := getDeviceUHeight(device)

		if position == 0 {
			// Device has no assigned position
			rackView.AddUnpositionedDevice(device)
		} else {
			rackView.AddDevice(device, position, uHeight)
		}
	}

	return rackView, nil
}

// FindAllRacks returns all racks in the inventory.
func FindAllRacks(inv *devicetypes.Inventory) []*devicetypes.CaniRackType {
	var racks []*devicetypes.CaniRackType
	for _, rack := range inv.Racks {
		if rack != nil {
			racks = append(racks, rack)
		}
	}
	return racks
}

// RenderAllRacks renders all racks (or filtered by name) to stdout
func RenderAllRacks(inv *devicetypes.Inventory, opts RenderOptions) error {
	return RenderAllRacksTo(os.Stdout, inv, opts)
}

// RenderAllRacksTo renders all racks to a specific writer
func RenderAllRacksTo(w io.Writer, inv *devicetypes.Inventory, opts RenderOptions) error {
	racks := FindAllRacks(inv)

	if len(racks) == 0 {
		// No racks - fall back to device list and cable view
		return RenderRacklessInventory(w, inv, opts)
	}

	racks = filterRacksByName(racks, opts.RackFilter)
	if len(racks) == 0 {
		fmt.Fprintf(w, "No racks matching '%s' found in inventory.\n", opts.RackFilter)
		return nil
	}

	// Render each rack
	for i, rack := range racks {
		rackView, err := BuildRackVisualization(inv, rack.ID)
		if err != nil {
			fmt.Fprintf(w, "Error building visualization for rack %s: %v\n", rack.Name, err)
			continue
		}

		if err := RenderRackASCII(w, rackView, opts); err != nil {
			return err
		}

		// Add separator between racks
		if i < len(racks)-1 {
			fmt.Fprintln(w)
		}
	}

	return nil
}

// filterRacksByName returns racks whose name contains filter (case-insensitive).
// An empty filter returns the input unchanged.
func filterRacksByName(racks []*devicetypes.CaniRackType, filter string) []*devicetypes.CaniRackType {
	if filter == "" {
		return racks
	}
	var filtered []*devicetypes.CaniRackType
	for _, rack := range racks {
		if strings.Contains(strings.ToLower(rack.Name), strings.ToLower(filter)) {
			filtered = append(filtered, rack)
		}
	}
	return filtered
}

// RenderRackASCII renders a single rack to a writer
func RenderRackASCII(w io.Writer, rv *RackView, opts RenderOptions) error {
	c := newRackPalette(opts.NoColor)

	// Calculate widths
	innerWidth := terminalWidth - 2 // Account for left and right borders

	// Rack title
	title := truncateString(rv.Rack.Name, innerWidth-2)
	titlePadded := centerString(title, innerWidth)

	// Print top border
	fmt.Fprintf(w, fmtTriple, boxTopLeft, strings.Repeat(boxHorizontal, innerWidth), boxTopRight)

	// Print title
	fmt.Fprintf(w, fmtTriple, boxVertical, c.bold(titlePadded), boxVertical)

	// Print header separator
	fmt.Fprintf(w, fmtQuint,
		boxTeeRight,
		strings.Repeat(boxHorizontal, uColumnWidth),
		boxTeeDown,
		strings.Repeat(boxHorizontal, innerWidth-uColumnWidth-1),
		boxTeeLeft)

	// Print U positions from top to bottom
	for u := rv.Height; u >= 1; u-- {
		uLabel := fmt.Sprintf(" U%-3d", u)
		content := slotContent(rv.GetSlot(u), c)
		contentPadded := padRight(content, innerWidth-uColumnWidth-1, opts.NoColor)

		fmt.Fprintf(w, fmtQuint,
			boxVertical,
			c.cyan(uLabel),
			boxVertical,
			contentPadded,
			boxVertical)
	}

	// Print bottom border
	fmt.Fprintf(w, fmtQuint,
		boxBottomLeft,
		strings.Repeat(boxHorizontal, uColumnWidth),
		boxTeeUp,
		strings.Repeat(boxHorizontal, innerWidth-uColumnWidth-1),
		boxBottomRight)

	// Print summary
	fmt.Fprintf(w, "  %s: %d devices, %d/%d U occupied, %d U empty\n",
		c.bold("Summary"),
		rv.DeviceCount(),
		rv.OccupiedCount(),
		rv.Height,
		rv.EmptyCount())

	printUnpositionedDevices(w, rv, c)
	printRackCables(w, rv, opts, c)

	return nil
}

// slotContent renders the content cell for a single rack U slot.
func slotContent(slot *RackSlot, c rackPalette) string {
	if slot == nil {
		return c.gray(fmt.Sprintf("%s [EMPTY]", markerEmpty))
	}
	if slot.IsContinued {
		return colorizeDevice(fmt.Sprintf("%s (continued)", markerContinued), slot.Device, c.red, c.green, c.yellow, c.cyan)
	}

	deviceName := truncateString(slot.Device.Name, contentWidth-10)
	uHeight := slot.Device.GetUHeight()
	if uHeight == 0 {
		uHeight = 1
	}
	if uHeight > 1 {
		return colorizeDevice(fmt.Sprintf("%s %s (%dU)", markerOccupied, deviceName, uHeight), slot.Device, c.red, c.green, c.yellow, c.cyan)
	}
	return colorizeDevice(fmt.Sprintf("%s %s", markerOccupied, deviceName), slot.Device, c.red, c.green, c.yellow, c.cyan)
}

// printUnpositionedDevices lists devices without an assigned U position.
func printUnpositionedDevices(w io.Writer, rv *RackView, c rackPalette) {
	if len(rv.UnpositionedDevices) == 0 {
		return
	}
	fmt.Fprintf(w, "\n  %s:\n", c.yellow("Unpositioned Devices"))
	for _, device := range rv.UnpositionedDevices {
		fmt.Fprintf(w, "    %s %s\n", c.yellow("•"), device.Name)
	}
}

// printRackCables prints cable connections for devices in the rack.
func printRackCables(w io.Writer, rv *RackView, opts RenderOptions, c rackPalette) {
	if !opts.ShowCables || opts.Inventory == nil || opts.Inventory.Cables == nil {
		return
	}
	rackCables := findCablesForRack(rv, opts.Inventory)
	if len(rackCables) == 0 {
		return
	}
	fmt.Fprintf(w, "\n  %s (%d cables):\n", c.cyan("Cable Connections"), len(rackCables))
	for _, cable := range rackCables {
		printRackCableLine(w, cable, opts.Inventory, c)
	}
}

// printRackCableLine prints a single cable connection line for a rack.
func printRackCableLine(w io.Writer, cable *devicetypes.CaniCableType, inv *devicetypes.Inventory, c rackPalette) {
	ifaceA, deviceA := inv.GetInterfaceByID(cable.TerminationA)
	ifaceB, deviceB := inv.GetInterfaceByID(cable.TerminationB)

	deviceAName := valUnknown
	deviceBName := valUnknown
	portAName := "?"
	portBName := "?"
	if deviceA != nil {
		deviceAName = deviceA.Name
	}
	if deviceB != nil {
		deviceBName = deviceB.Name
	}
	if ifaceA != nil {
		portAName = ifaceA.Name
	}
	if ifaceB != nil {
		portBName = ifaceB.Name
	}

	fmt.Fprintf(w, "    %s %s\n",
		c.green("•"),
		fmt.Sprintf("%s [%s:%s] ←→ [%s:%s]",
			cable.Label,
			deviceAName, portAName,
			deviceBName, portBName))
}

// getRackHeight determines the height of a rack in U
func getRackHeight(rack *devicetypes.CaniRackType) int {
	// Prefer the rack's explicit height.
	if rack.UHeight > 0 {
		return rack.UHeight
	}

	// First try device type lookup
	if rack.Slug != "" {
		if dt, ok := devicetypes.GetBySlug(rack.Slug); ok && dt.UHeight > 0 {
			return dt.UHeight
		}
	}

	// Try to parse from slug or name (e.g., "48u", "42u")
	patterns := []string{rack.Slug, rack.Name}
	re := regexp.MustCompile(`(\d+)[uU]`)

	for _, s := range patterns {
		matches := re.FindStringSubmatch(strings.ToLower(s))
		if len(matches) >= 2 {
			if h, err := strconv.Atoi(matches[1]); err == nil && h > 0 {
				return h
			}
		}
	}

	// Default to 48U
	return 48
}

// getDeviceUHeight returns the u_height for a device.
func getDeviceUHeight(device *devicetypes.CaniDeviceType) int {
	return device.GetUHeight()
}

// getIntProperty safely extracts an int from a map[string]any
func getIntProperty(props map[string]any, key string, defaultVal int) int {
	if props == nil {
		return defaultVal
	}

	val, ok := props[key]
	if !ok {
		return defaultVal
	}

	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}

	return defaultVal
}

// truncateString truncates a string to a maximum length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// centerString centers a string within a given width
func centerString(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	leftPad := (width - len(s)) / 2
	rightPad := width - len(s) - leftPad
	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

// padRight pads a string to a given width (accounting for ANSI codes)
func padRight(s string, width int, noColor bool) string {
	// Calculate visible length (excluding ANSI codes)
	visibleLen := visibleLength(s)

	if visibleLen >= width {
		return s
	}

	return s + strings.Repeat(" ", width-visibleLen)
}

// visibleLength returns the visible length of a string, excluding ANSI escape codes
func visibleLength(s string) int {
	// Remove ANSI escape codes
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	clean := re.ReplaceAllString(s, "")
	return utf8.RuneCountInString(clean)
}

// findCablesForRack returns all cables where at least one termination is a device in the rack
func findCablesForRack(rv *RackView, inv *devicetypes.Inventory) []*devicetypes.CaniCableType {
	if inv == nil || inv.Cables == nil {
		return nil
	}

	// Build a set of device IDs in this rack
	rackDeviceIDs := make(map[uuid.UUID]bool)
	rackDeviceIDs[rv.Rack.ID] = true
	for _, deviceID := range rv.Rack.Devices {
		rackDeviceIDs[deviceID] = true
	}

	var cables []*devicetypes.CaniCableType
	for _, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		// Check if either termination's device is in this rack
		_, deviceA := inv.GetInterfaceByID(cable.TerminationA)
		_, deviceB := inv.GetInterfaceByID(cable.TerminationB)
		deviceAInRack := deviceA != nil && rackDeviceIDs[deviceA.ID]
		deviceBInRack := deviceB != nil && rackDeviceIDs[deviceB.ID]
		if deviceAInRack || deviceBInRack {
			cables = append(cables, cable)
		}
	}

	return cables
}

// RenderRacklessInventory renders inventory that has no racks - shows device summary and cables
func RenderRacklessInventory(w io.Writer, inv *devicetypes.Inventory, opts RenderOptions) error {
	c := newRackPalette(opts.NoColor)

	innerWidth := terminalWidth - 2

	// Print header
	fmt.Fprintf(w, fmtTriple, boxTopLeft, strings.Repeat(boxHorizontal, innerWidth), boxTopRight)
	title := centerString("Inventory Summary (No Racks Defined)", innerWidth)
	fmt.Fprintf(w, fmtTriple, boxVertical, c.bold(title), boxVertical)
	fmt.Fprintf(w, fmtTriple, boxBottomLeft, strings.Repeat(boxHorizontal, innerWidth), boxBottomRight)

	typeCounts, topLevelDevices := countDeviceTypes(inv)
	printDeviceTypeSummary(w, typeCounts, c)
	printTopLevelDevices(w, topLevelDevices, c)
	printRacklessCables(w, inv, opts, c)

	fmt.Fprintln(w)
	return nil
}

// countDeviceTypes tallies devices by hardware type and collects top-level devices.
func countDeviceTypes(inv *devicetypes.Inventory) (map[string]int, []*devicetypes.CaniDeviceType) {
	typeCounts := make(map[string]int)
	var topLevelDevices []*devicetypes.CaniDeviceType
	for _, device := range inv.Devices {
		if device == nil {
			continue
		}
		typeCounts[string(device.Type)]++
		if device.Parent == uuid.Nil {
			topLevelDevices = append(topLevelDevices, device)
		}
	}
	return typeCounts, topLevelDevices
}

// printDeviceTypeSummary prints the device-type counts section.
func printDeviceTypeSummary(w io.Writer, typeCounts map[string]int, c rackPalette) {
	fmt.Fprintf(w, "\n  %s:\n", c.cyan("Device Types"))
	for hwType, count := range typeCounts {
		if hwType == "" {
			hwType = "(untyped)"
		}
		fmt.Fprintf(w, "    %s %-20s %d\n", c.green("•"), hwType, count)
	}
}

// printTopLevelDevices prints devices that have no parent.
func printTopLevelDevices(w io.Writer, devices []*devicetypes.CaniDeviceType, c rackPalette) {
	if len(devices) == 0 {
		return
	}
	fmt.Fprintf(w, "\n  %s (%d devices):\n", c.cyan("Top-Level Devices"), len(devices))
	for _, device := range devices {
		hwType := string(device.Type)
		if hwType == "" {
			hwType = "device"
		}
		fmt.Fprintf(w, "    %s %s [%s]\n", c.green("•"), device.Name, c.yellow(hwType))
	}
}

// printRacklessCables prints cable connections (or a placeholder) when requested.
func printRacklessCables(w io.Writer, inv *devicetypes.Inventory, opts RenderOptions, c rackPalette) {
	if !opts.ShowCables {
		return
	}
	if len(inv.Cables) == 0 {
		fmt.Fprintf(w, "\n  %s: No cables defined in inventory\n", c.yellow("Cables"))
		return
	}
	fmt.Fprintf(w, "\n  %s (%d cables):\n", c.cyan("Cable Connections"), len(inv.Cables))
	for _, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		printRacklessCableLine(w, cable, inv, c)
	}
}

// printRacklessCableLine prints a single cable line for the rackless summary.
func printRacklessCableLine(w io.Writer, cable *devicetypes.CaniCableType, inv *devicetypes.Inventory, c rackPalette) {
	ifaceA, deviceA := inv.GetInterfaceByID(cable.TerminationA)
	ifaceB, deviceB := inv.GetInterfaceByID(cable.TerminationB)

	deviceAName := valUnknown
	deviceBName := valUnknown
	portAName := "?"
	portBName := "?"
	if deviceA != nil {
		deviceAName = truncateString(deviceA.Name, 30)
	}
	if deviceB != nil {
		deviceBName = truncateString(deviceB.Name, 30)
	}
	if ifaceA != nil {
		portAName = ifaceA.Name
	}
	if ifaceB != nil {
		portBName = ifaceB.Name
	}

	cableType := cable.Slug
	if cableType == "" {
		cableType = valUnknown
	}

	fmt.Fprintf(w, "    %s [%s] %s:%s ←→ %s:%s\n",
		c.green(cableType),
		c.yellow(cable.Status),
		deviceAName, portAName,
		deviceBName, portBName)
}
