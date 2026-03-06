/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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

// Package resolve provides UUID-then-name resolution for inventory items.
package resolve

import (
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// Location resolves a positional arg to a location UUID.
// It first tries uuid.Parse; on failure it searches by name (case-insensitive).
func Location(inv *devicetypes.Inventory, arg string) (uuid.UUID, error) {
	if id, err := uuid.Parse(arg); err == nil {
		if _, ok := inv.Locations[id]; ok {
			return id, nil
		}
		return uuid.Nil, fmt.Errorf("location %s not found", id)
	}
	return findByName(arg, locationNames(inv))
}

// Rack resolves a positional arg to a rack UUID.
func Rack(inv *devicetypes.Inventory, arg string) (uuid.UUID, error) {
	if id, err := uuid.Parse(arg); err == nil {
		if _, ok := inv.Racks[id]; ok {
			return id, nil
		}
		return uuid.Nil, fmt.Errorf("rack %s not found", id)
	}
	return findByName(arg, rackNames(inv))
}

// Device resolves a positional arg to a device UUID.
func Device(inv *devicetypes.Inventory, arg string) (uuid.UUID, error) {
	if id, err := uuid.Parse(arg); err == nil {
		if _, ok := inv.Devices[id]; ok {
			return id, nil
		}
		return uuid.Nil, fmt.Errorf("device %s not found", id)
	}
	return findByName(arg, deviceNames(inv))
}

// Module resolves a positional arg to a module UUID.
func Module(inv *devicetypes.Inventory, arg string) (uuid.UUID, error) {
	if id, err := uuid.Parse(arg); err == nil {
		if _, ok := inv.Modules[id]; ok {
			return id, nil
		}
		return uuid.Nil, fmt.Errorf("module %s not found", id)
	}
	return findByName(arg, moduleNames(inv))
}

// Cable resolves a positional arg to a cable UUID.
func Cable(inv *devicetypes.Inventory, arg string) (uuid.UUID, error) {
	if id, err := uuid.Parse(arg); err == nil {
		if _, ok := inv.Cables[id]; ok {
			return id, nil
		}
		return uuid.Nil, fmt.Errorf("cable %s not found", id)
	}
	return findByName(arg, cableLabels(inv))
}

// nameEntry pairs a name with its UUID for lookup.
type nameEntry struct {
	id   uuid.UUID
	name string
}

// findByName does case-insensitive name matching. Returns an error on
// zero or multiple matches.
func findByName(arg string, entries []nameEntry) (uuid.UUID, error) {
	lower := strings.ToLower(arg)
	var matches []nameEntry
	for _, e := range entries {
		if strings.ToLower(e.name) == lower {
			matches = append(matches, e)
		}
	}
	switch len(matches) {
	case 0:
		return uuid.Nil, fmt.Errorf("no item found matching %q", arg)
	case 1:
		return matches[0].id, nil
	default:
		lines := make([]string, 0, len(matches))
		for _, m := range matches {
			lines = append(lines, fmt.Sprintf("  %s  %s", m.id, m.name))
		}
		return uuid.Nil, fmt.Errorf(
			"multiple items match %q; use a UUID:\n%s",
			arg, strings.Join(lines, "\n"),
		)
	}
}

func locationNames(inv *devicetypes.Inventory) []nameEntry {
	entries := make([]nameEntry, 0, len(inv.Locations))
	for id, loc := range inv.Locations {
		if loc != nil {
			entries = append(entries, nameEntry{id: id, name: loc.Name})
		}
	}
	return entries
}

func rackNames(inv *devicetypes.Inventory) []nameEntry {
	entries := make([]nameEntry, 0, len(inv.Racks))
	for id, r := range inv.Racks {
		if r != nil {
			entries = append(entries, nameEntry{id: id, name: r.Name})
		}
	}
	return entries
}

func deviceNames(inv *devicetypes.Inventory) []nameEntry {
	entries := make([]nameEntry, 0, len(inv.Devices))
	for id, d := range inv.Devices {
		if d != nil {
			entries = append(entries, nameEntry{id: id, name: d.Name})
		}
	}
	return entries
}

func moduleNames(inv *devicetypes.Inventory) []nameEntry {
	entries := make([]nameEntry, 0, len(inv.Modules))
	for id, m := range inv.Modules {
		if m != nil {
			entries = append(entries, nameEntry{id: id, name: m.Name})
		}
	}
	return entries
}

func cableLabels(inv *devicetypes.Inventory) []nameEntry {
	entries := make([]nameEntry, 0, len(inv.Cables))
	for id, c := range inv.Cables {
		if c != nil {
			entries = append(entries, nameEntry{id: id, name: c.Label})
		}
	}
	return entries
}
