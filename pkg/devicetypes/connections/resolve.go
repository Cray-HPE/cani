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
package connections

import (
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/internal/util/nameexpand"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// ResolvedConnection is a single fully-resolved cable ready to be added
// to the inventory.
type ResolvedConnection struct {
	ADevice uuid.UUID
	APort   string
	BDevice uuid.UUID
	BPort   string
	Cable   CableProps
}

// ResolveConnectionMap expands patterns and resolves device names in a
// ConnectionMap against the inventory, returning a flat list of concrete
// cable connections. Errors are collected per-entry so the caller can
// report all problems at once.
func ResolveConnectionMap(cm *ConnectionMap, inv *devicetypes.Inventory) ([]ResolvedConnection, []error) {
	var resolved []ResolvedConnection
	var errs []error

	for i, entry := range cm.Connections {
		conns, err := resolveEntry(entry, cm.CableDefaults, inv)
		if err != nil {
			errs = append(errs, fmt.Errorf("connection[%d]: %w", i, err))
			continue
		}
		resolved = append(resolved, conns...)
	}
	return resolved, errs
}

// resolveEntry expands a single ConnectionEntry into one or more
// ResolvedConnections using brace expansion + zip semantics.
func resolveEntry(entry ConnectionEntry, defaults *CableDefaults, inv *devicetypes.Inventory) ([]ResolvedConnection, error) {
	aDevices := expandPattern(entry.A.Device)
	aPorts := expandPattern(entry.A.Port)
	bDevices := expandPattern(entry.B.Device)
	bPorts := expandPattern(entry.B.Port)

	count, err := ZipCount(len(aDevices), len(aPorts), len(bDevices), len(bPorts))
	if err != nil {
		return nil, err
	}

	aDevices = Broadcast(aDevices, count)
	aPorts = Broadcast(aPorts, count)
	bDevices = Broadcast(bDevices, count)
	bPorts = Broadcast(bPorts, count)

	result := make([]ResolvedConnection, 0, count)
	for i := range count {
		aDev := inv.FindDeviceByNameOrID(aDevices[i])
		if aDev == nil {
			return nil, fmt.Errorf("device not found: %s", aDevices[i])
		}
		bDev := inv.FindDeviceByNameOrID(bDevices[i])
		if bDev == nil {
			return nil, fmt.Errorf("device not found: %s", bDevices[i])
		}
		result = append(result, ResolvedConnection{
			ADevice: aDev.ID,
			APort:   aPorts[i],
			BDevice: bDev.ID,
			BPort:   bPorts[i],
			Cable:   mergeProps(entry.Cable, defaults),
		})
	}
	return result, nil
}

// expandPattern expands a brace pattern or returns a single-element slice.
func expandPattern(s string) []string {
	if s == "" {
		return []string{""}
	}
	if strings.Contains(s, "{") && strings.Contains(s, "}") {
		expanded, err := nameexpand.Expand(s)
		if err == nil && len(expanded) > 0 {
			return expanded
		}
	}
	return []string{s}
}

// ZipCount determines the aligned count from multiple list lengths.
// All lengths > 1 must agree; lengths of 0 or 1 are broadcast.
func ZipCount(lengths ...int) (int, error) {
	maxLen := 1
	for _, l := range lengths {
		if l > 1 {
			if maxLen > 1 && l != maxLen {
				return 0, fmt.Errorf("pattern length mismatch: %d vs %d", maxLen, l)
			}
			maxLen = l
		}
	}
	return maxLen, nil
}

// Broadcast replicates a single-element slice to length n.
func Broadcast(s []string, n int) []string {
	if len(s) == n {
		return s
	}
	out := make([]string, n)
	if len(s) == 1 {
		for i := range out {
			out[i] = s[0]
		}
	}
	return out
}

// mergeProps combines per-entry cable props with defaults.
// Per-entry values override defaults.
func mergeProps(entry *CableProps, defaults *CableDefaults) CableProps {
	var p CableProps
	if defaults != nil {
		p.Type = defaults.Type
		p.Status = defaults.Status
		p.Color = defaults.Color
		p.LengthUnit = defaults.LengthUnit
	}
	if entry != nil {
		if entry.Type != "" {
			p.Type = entry.Type
		}
		if entry.Label != "" {
			p.Label = entry.Label
		}
		if entry.Color != "" {
			p.Color = entry.Color
		}
		if entry.Length != nil {
			p.Length = entry.Length
		}
		if entry.LengthUnit != "" {
			p.LengthUnit = entry.LengthUnit
		}
		if entry.Status != "" {
			p.Status = entry.Status
		}
	}
	return p
}
