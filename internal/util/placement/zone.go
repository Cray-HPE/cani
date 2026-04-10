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
package placement

import (
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// Zone identifies a vertical region inside a rack.
type Zone string

const (
	ZoneTop    Zone = "top"
	ZoneMiddle Zone = "middle"
	ZoneBottom Zone = "bottom"

	// DefaultTopZoneHeight is the default number of U reserved at the
	// top of a rack for switches and top-of-rack controllers.
	DefaultTopZoneHeight = 4

	// DefaultBottomZoneHeight is the default number of U reserved at
	// the bottom. Zero means the bottom zone is unused by default.
	DefaultBottomZoneHeight = 0
)

// ZoneBounds describes the U-slot ranges for each rack zone.
// StartU and EndU are both inclusive and 1-based.
type ZoneBounds struct {
	Top    URange
	Middle URange
	Bottom URange
}

// URange is an inclusive range of rack units.
type URange struct {
	StartU int
	EndU   int
}

// Height returns the number of U-slots in the range, or 0 if empty.
// A zero-value URange (both fields 0) is considered empty.
func (r URange) Height() int {
	if r.StartU <= 0 || r.EndU < r.StartU {
		return 0
	}
	return r.EndU - r.StartU + 1
}

// ComputeZoneBounds calculates zone boundaries for a rack of the given
// height. topHeight and bottomHeight specify how many U to reserve.
// Pass 0 for either to leave that zone empty.
func ComputeZoneBounds(uHeight, topHeight, bottomHeight int) ZoneBounds {
	if topHeight < 0 {
		topHeight = 0
	}
	if bottomHeight < 0 {
		bottomHeight = 0
	}

	var b ZoneBounds

	// Bottom zone: U1 … bottomHeight
	if bottomHeight > 0 && bottomHeight < uHeight {
		b.Bottom = URange{StartU: 1, EndU: bottomHeight}
	}

	// Top zone: (uHeight - topHeight + 1) … uHeight
	if topHeight > 0 && topHeight < uHeight {
		b.Top = URange{
			StartU: uHeight - topHeight + 1,
			EndU:   uHeight,
		}
	}

	// Middle zone: everything between bottom and top
	midStart := bottomHeight + 1
	midEnd := uHeight - topHeight
	if midStart < 1 {
		midStart = 1
	}
	if midEnd > uHeight {
		midEnd = uHeight
	}
	if midStart <= midEnd {
		b.Middle = URange{StartU: midStart, EndU: midEnd}
	}

	return b
}

// ZoneForHardwareType returns the default zone for a device based on
// its hardware-type string. Switches go to top, PDUs/CDUs to bottom,
// and everything else (servers, blades, chassis, etc.) to middle.
func ZoneForHardwareType(hwType string) Zone {
	switch devicetypes.Type(strings.ToLower(hwType)) {
	case devicetypes.TypeSwitch, devicetypes.TypeMgmtSwitch, devicetypes.TypeHsnSwitch:
		return ZoneTop
	case devicetypes.TypeCabinetPDU, devicetypes.TypeCDU:
		return ZoneBottom
	default:
		return ZoneMiddle
	}
}

// ParseZone converts a user-provided string to a Zone constant.
// Returns an error for unrecognised values.
func ParseZone(s string) (Zone, error) {
	switch Zone(strings.ToLower(strings.TrimSpace(s))) {
	case ZoneTop:
		return ZoneTop, nil
	case ZoneMiddle:
		return ZoneMiddle, nil
	case ZoneBottom:
		return ZoneBottom, nil
	default:
		return "", fmt.Errorf("unknown zone %q (valid: top, middle, bottom)", s)
	}
}

// RangeForZone returns the URange for the requested zone.
func (b ZoneBounds) RangeForZone(z Zone) URange {
	switch z {
	case ZoneTop:
		return b.Top
	case ZoneBottom:
		return b.Bottom
	default:
		return b.Middle
	}
}

// FindSlotInZone finds the next available starting U-position within
// the given zone range. It scans top-to-bottom inside the zone.
// Returns 0 if no space is available.
func FindSlotInZone(rack *devicetypes.CaniRackType, height int, face string, isFullDepth bool, zr URange) int {
	if rack == nil || height < 1 || zr.Height() < height {
		return 0
	}
	for startU := zr.EndU - height + 1; startU >= zr.StartU; startU-- {
		if rack.CanFitDevice(startU, height, face, isFullDepth) {
			return startU
		}
	}
	return 0
}

// ResolveZoneBounds returns ZoneBounds for a rack, using per-rack
// fields when positive and falling back to package-level defaults.
// A value of zero means "use default"; set TopZoneHeight or
// BottomZoneHeight to a positive value to override.
func ResolveZoneBounds(rack *devicetypes.CaniRackType) ZoneBounds {
	top := DefaultTopZoneHeight
	if rack.TopZoneHeight > 0 {
		top = rack.TopZoneHeight
	}
	bottom := DefaultBottomZoneHeight
	if rack.BottomZoneHeight > 0 {
		bottom = rack.BottomZoneHeight
	}
	return ComputeZoneBounds(rack.UHeight, top, bottom)
}
