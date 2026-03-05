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

// BuildRackVisualization creates a RackView from inventory for a specific rack
func BuildRackVisualization(inv *devicetypes.Inventory, rackID uuid.UUID) (*RackView, error) {
	rack, ok := inv.Devices[rackID]
	if !ok {
		return nil, fmt.Errorf("rack %s not found in inventory", rackID)
	}

	if rack.HardwareType != string(devicetypes.Rack) {
		return nil, fmt.Errorf("device %s is not a rack (type: %s)", rack.Name, rack.HardwareType)
	}

	// Determine rack height
	height := getRackHeight(rack)

	rackView := NewRackView(rack, height)

	// Process children
	for _, childID := range rack.Children {
		device, ok := inv.Devices[childID]
		if !ok || device == nil {
			continue
		}

		// Skip nested racks
		if device.HardwareType == string(devicetypes.Rack) {
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

// FindAllRacks finds all rack-type devices in the inventory
func FindAllRacks(inv *devicetypes.Inventory) []*devicetypes.CaniDeviceType {
	var racks []*devicetypes.CaniDeviceType
	for _, device := range inv.Devices {
		if device != nil && device.HardwareType == string(devicetypes.Rack) {
			racks = append(racks, device)
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

	// Filter by rack name if specified
	if opts.RackFilter != "" {
		var filtered []*devicetypes.CaniDeviceType
		for _, rack := range racks {
			if strings.Contains(strings.ToLower(rack.Name), strings.ToLower(opts.RackFilter)) {
				filtered = append(filtered, rack)
			}
		}
		racks = filtered

		if len(racks) == 0 {
			fmt.Fprintf(w, "No racks matching '%s' found in inventory.\n", opts.RackFilter)
			return nil
		}
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

// RenderRackASCII renders a single rack to a writer
func RenderRackASCII(w io.Writer, rv *RackView, opts RenderOptions) error {
	// Color helper functions
	green := func(s string) string {
		if opts.NoColor {
			return s
		}
		return ColorGreen + s + ColorReset
	}
	gray := func(s string) string {
		if opts.NoColor {
			return s
		}
		return ColorGray + s + ColorReset
	}
	cyan := func(s string) string {
		if opts.NoColor {
			return s
		}
		return ColorCyan + s + ColorReset
	}
	yellow := func(s string) string {
		if opts.NoColor {
			return s
		}
		return ColorYellow + s + ColorReset
	}
	bold := func(s string) string {
		if opts.NoColor {
			return s
		}
		return ColorBold + s + ColorReset
	}

	// Calculate widths
	innerWidth := terminalWidth - 2 // Account for left and right borders

	// Rack title
	title := truncateString(rv.Rack.Name, innerWidth-2)
	titlePadded := centerString(title, innerWidth)

	// Print top border
	fmt.Fprintf(w, "%s%s%s\n", boxTopLeft, strings.Repeat(boxHorizontal, innerWidth), boxTopRight)

	// Print title
	fmt.Fprintf(w, "%s%s%s\n", boxVertical, bold(titlePadded), boxVertical)

	// Print header separator
	fmt.Fprintf(w, "%s%s%s%s%s\n",
		boxTeeRight,
		strings.Repeat(boxHorizontal, uColumnWidth),
		boxTeeDown,
		strings.Repeat(boxHorizontal, innerWidth-uColumnWidth-1),
		boxTeeLeft)

	// Print U positions from top to bottom
	for u := rv.Height; u >= 1; u-- {
		uLabel := fmt.Sprintf("U%-3d", u)
		slot := rv.GetSlot(u)

		var content string
		if slot == nil {
			// Empty slot
			content = gray(fmt.Sprintf("%s [EMPTY]", markerEmpty))
		} else if slot.IsContinued {
			// Continuation of multi-U device
			content = green(fmt.Sprintf("%s (continued)", markerContinued))
		} else {
			// Start of device
			deviceName := truncateString(slot.Device.Name, contentWidth-10)
			uHeight := slot.Device.GetUHeight()
			if uHeight == 0 {
				uHeight = 1
			}
			if uHeight > 1 {
				content = green(fmt.Sprintf("%s %s (%dU)", markerOccupied, deviceName, uHeight))
			} else {
				content = green(fmt.Sprintf("%s %s", markerOccupied, deviceName))
			}
		}

		// Pad content to fill the space
		contentPadded := padRight(content, innerWidth-uColumnWidth-1, opts.NoColor)

		fmt.Fprintf(w, "%s %s %s%s%s\n",
			boxVertical,
			cyan(uLabel),
			boxVertical,
			contentPadded,
			boxVertical)
	}

	// Print bottom border
	fmt.Fprintf(w, "%s%s%s%s%s\n",
		boxBottomLeft,
		strings.Repeat(boxHorizontal, uColumnWidth),
		boxTeeUp,
		strings.Repeat(boxHorizontal, innerWidth-uColumnWidth-1),
		boxBottomRight)

	// Print summary
	fmt.Fprintf(w, "  %s: %d devices, %d/%d U occupied, %d U empty\n",
		bold("Summary"),
		rv.DeviceCount(),
		rv.OccupiedCount(),
		rv.Height,
		rv.EmptyCount())

	// Print unpositioned devices if any
	if len(rv.UnpositionedDevices) > 0 {
		fmt.Fprintf(w, "\n  %s:\n", yellow("Unpositioned Devices"))
		for _, device := range rv.UnpositionedDevices {
			fmt.Fprintf(w, "    %s %s\n", yellow("•"), device.Name)
		}
	}

	// Print cable connections if requested
	if opts.ShowCables && opts.Inventory != nil && opts.Inventory.Cables != nil {
		// Find cables connected to devices in this rack
		rackCables := findCablesForRack(rv, opts.Inventory)
		if len(rackCables) > 0 {
			fmt.Fprintf(w, "\n  %s (%d cables):\n", cyan("Cable Connections"), len(rackCables))
			for _, cable := range rackCables {
				ifaceA, deviceA := opts.Inventory.GetInterfaceByID(cable.TerminationA)
				ifaceB, deviceB := opts.Inventory.GetInterfaceByID(cable.TerminationB)

				deviceAName := "unknown"
				deviceBName := "unknown"
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
					green("•"),
					fmt.Sprintf("%s [%s:%s] ←→ [%s:%s]",
						cable.Label,
						deviceAName, portAName,
						deviceBName, portBName))
			}
		}
	}

	return nil
}

// getRackHeight determines the height of a rack in U
func getRackHeight(rack *devicetypes.CaniDeviceType) int {
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
	return len(clean)
}

// findCablesForRack returns all cables where at least one termination is a device in the rack
func findCablesForRack(rv *RackView, inv *devicetypes.Inventory) []*devicetypes.CaniCableType {
	if inv == nil || inv.Cables == nil {
		return nil
	}

	// Build a set of device IDs in this rack
	rackDeviceIDs := make(map[uuid.UUID]bool)
	rackDeviceIDs[rv.Rack.ID] = true
	for _, childID := range rv.Rack.Children {
		rackDeviceIDs[childID] = true
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
	// Color helpers
	green := func(s string) string {
		if opts.NoColor {
			return s
		}
		return ColorGreen + s + ColorReset
	}
	cyan := func(s string) string {
		if opts.NoColor {
			return s
		}
		return ColorCyan + s + ColorReset
	}
	yellow := func(s string) string {
		if opts.NoColor {
			return s
		}
		return ColorYellow + s + ColorReset
	}
	bold := func(s string) string {
		if opts.NoColor {
			return s
		}
		return ColorBold + s + ColorReset
	}

	innerWidth := terminalWidth - 2

	// Print header
	fmt.Fprintf(w, "%s%s%s\n", boxTopLeft, strings.Repeat(boxHorizontal, innerWidth), boxTopRight)
	title := centerString("Inventory Summary (No Racks Defined)", innerWidth)
	fmt.Fprintf(w, "%s%s%s\n", boxVertical, bold(title), boxVertical)
	fmt.Fprintf(w, "%s%s%s\n", boxBottomLeft, strings.Repeat(boxHorizontal, innerWidth), boxBottomRight)

	// Count devices by hardware type
	typeCounts := make(map[string]int)
	var topLevelDevices []*devicetypes.CaniDeviceType

	for _, device := range inv.Devices {
		if device == nil {
			continue
		}
		typeCounts[device.HardwareType]++

		// Collect top-level devices (no parent or parent is nil UUID)
		if device.Parent == uuid.Nil {
			topLevelDevices = append(topLevelDevices, device)
		}
	}

	// Print device type summary
	fmt.Fprintf(w, "\n  %s:\n", cyan("Device Types"))
	for hwType, count := range typeCounts {
		if hwType == "" {
			hwType = "(untyped)"
		}
		fmt.Fprintf(w, "    %s %-20s %d\n", green("•"), hwType, count)
	}

	// Print top-level devices
	if len(topLevelDevices) > 0 {
		fmt.Fprintf(w, "\n  %s (%d devices):\n", cyan("Top-Level Devices"), len(topLevelDevices))
		for _, device := range topLevelDevices {
			hwType := device.HardwareType
			if hwType == "" {
				hwType = "device"
			}
			fmt.Fprintf(w, "    %s %s [%s]\n", green("•"), device.Name, yellow(hwType))
		}
	}

	// Print cables if requested
	if opts.ShowCables && inv.Cables != nil && len(inv.Cables) > 0 {
		fmt.Fprintf(w, "\n  %s (%d cables):\n", cyan("Cable Connections"), len(inv.Cables))

		for _, cable := range inv.Cables {
			if cable == nil {
				continue
			}

			ifaceA, deviceA := inv.GetInterfaceByID(cable.TerminationA)
			ifaceB, deviceB := inv.GetInterfaceByID(cable.TerminationB)

			deviceAName := "unknown"
			deviceBName := "unknown"
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
				cableType = "unknown"
			}

			fmt.Fprintf(w, "    %s [%s] %s:%s ←→ %s:%s\n",
				green(cableType),
				yellow(cable.Status),
				deviceAName, portAName,
				deviceBName, portBName)
		}
	} else if opts.ShowCables {
		fmt.Fprintf(w, "\n  %s: No cables defined in inventory\n", yellow("Cables"))
	}

	fmt.Fprintln(w)
	return nil
}
