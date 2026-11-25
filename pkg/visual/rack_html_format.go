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
	"fmt"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// sortedPositionsDesc returns the start positions of a rack's devices, top first.
func sortedPositionsDesc(devices map[int]*devicetypes.CaniDeviceType) []int {
	positions := make([]int, 0, len(devices))
	for pos := range devices {
		positions = append(positions, pos)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(positions)))
	return positions
}

// htmlKind maps a hardware type onto the five color-coded tile kinds.
func htmlKind(t devicetypes.Type) string {
	switch t {
	case devicetypes.TypeBlade, devicetypes.TypeNode:
		return "server"
	case devicetypes.TypeChassis:
		return "chassis"
	case devicetypes.TypeSwitch, devicetypes.TypeMgmtSwitch:
		return "switch"
	case devicetypes.TypeHsnSwitch:
		return "hsn"
	default:
		return "other"
	}
}

// htmlClasses returns the CSS modifier classes of a tile or child cell. An
// empty status is rendered as "unknown" so it picks up the unknown status ink.
func htmlClasses(d htmlDevice) string {
	status := cssToken(d.Status)
	if status == "" {
		status = "unknown"
	}
	classes := []string{"kind-" + d.Kind, "status-" + status}
	for _, tag := range strings.Split(d.TagList, ", ") {
		if tag != "" {
			classes = append(classes, "tag-"+cssToken(tag))
		}
	}
	if d.HalfDepth {
		classes = append(classes, "half")
	}
	if len(d.Children) > 0 {
		classes = append(classes, "parent")
	}
	return strings.Join(classes, " ")
}

// cssToken lowercases s and replaces anything outside [a-z0-9] with '-'.
func cssToken(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	return b.String()
}

// htmlTooltip is the native hover text of a tile: name, model, position, status.
func htmlTooltip(d htmlDevice) string {
	parts := []string{d.Name}
	if d.Model != "" {
		parts = append(parts, d.Model)
	}
	position := d.Position
	if d.ParentLink != nil {
		position = fmt.Sprintf("in %s (%s)", d.ParentLink.Name, d.Position)
	}
	parts = append(parts, strings.TrimSuffix(position+", "+d.Status, ", "))
	return strings.Join(parts, "\n")
}

// formatURange renders "U3-U4" or "U7" for a single unit.
func formatURange(start, end int) string {
	if start == end {
		return fmt.Sprintf("U%d", start)
	}
	return fmt.Sprintf("U%d-U%d", start, end)
}

// childColumns picks the child grid width so a chassis of h U shows its
// children in at most h rows (a 2U chassis with 4 nodes becomes 2x2).
func childColumns(children, height int) int {
	if children == 0 {
		return 0
	}
	if height < 1 {
		height = 1
	}
	return (children + height - 1) / height
}

// hasTag reports whether tags contains tag.
func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// htmlIPv4 resolves a primary IPv4 reference to its address string.
func htmlIPv4(inv *devicetypes.Inventory, id uuid.UUID) string {
	if id == uuid.Nil || inv.IPAddresses == nil {
		return ""
	}
	ip := inv.IPAddresses[id]
	if ip == nil {
		return ""
	}
	if ip.Address != "" {
		return ip.Address
	}
	return ip.Host
}

// htmlLocationPath renders a rack's location chain root first, e.g. "Geneva / MDC".
func htmlLocationPath(inv *devicetypes.Inventory, id uuid.UUID) string {
	const maxDepth = 8
	var names []string
	for depth := 0; depth < maxDepth && id != uuid.Nil; depth++ {
		loc := inv.Locations[id]
		if loc == nil {
			break
		}
		names = append([]string{loc.Name}, names...)
		id = loc.Parent
	}
	return strings.Join(names, " / ")
}
