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
)

// Hardware categories in display order: CPUs, memory, drives, cards, rest.
const (
	hwCPU = iota
	hwMemory
	hwDisk
	hwCard
	hwOther
)

const mib = 1024 * 1024

// hwEntry is one FRU or module classified for grouping.
type hwEntry struct {
	category int
	kind     string
	title    string
	meta     string
	item     htmlHardwareItem
	bytes    float64
}

// hwGroup accumulates the FRUs or modules that share one summary line.
type hwGroup struct {
	category int
	kind     string
	group    htmlHardwareGroup
	bytes    float64
}

// buildHardware groups the FRUs (CPUs, DIMMs, drives) and modules (cards)
// of a device into summary lines and returns them with the section summary.
func buildHardware(inv *devicetypes.Inventory, dev *devicetypes.CaniDeviceType) ([]htmlHardwareGroup, string) {
	groups := map[string]*hwGroup{}
	for _, fru := range deviceFrus(inv, dev) {
		addHardware(groups, describeFru(dev, fru))
	}
	for _, mod := range deviceModules(inv, dev) {
		addHardware(groups, describeModule(mod))
	}
	ordered := sortedHardware(groups)
	out := make([]htmlHardwareGroup, 0, len(ordered))
	for _, g := range ordered {
		out = append(out, finishHardware(g))
	}
	return out, hardwareSummary(ordered)
}

// deviceFrus returns the FRUs installed in dev, sorted by name.
func deviceFrus(inv *devicetypes.Inventory, dev *devicetypes.CaniDeviceType) []*devicetypes.CaniFruType {
	listed := map[string]bool{}
	for _, id := range dev.Frus {
		listed[id.String()] = true
	}
	var frus []*devicetypes.CaniFruType
	for _, fru := range inv.Frus {
		if fru != nil && (fru.Device == dev.ID || listed[fru.ID.String()]) {
			frus = append(frus, fru)
		}
	}
	sort.Slice(frus, func(i, j int) bool { return naturalLess(frus[i].Name, frus[j].Name) })
	return frus
}

// deviceModules returns the modules plugged into dev, sorted by bay.
func deviceModules(inv *devicetypes.Inventory, dev *devicetypes.CaniDeviceType) []*devicetypes.CaniModuleType {
	var mods []*devicetypes.CaniModuleType
	for _, mod := range inv.Modules {
		if mod != nil && mod.ParentDevice == dev.ID {
			mods = append(mods, mod)
		}
	}
	sort.Slice(mods, func(i, j int) bool {
		return naturalLess(mods[i].ModuleBayName+" "+mods[i].Name, mods[j].ModuleBayName+" "+mods[j].Name)
	})
	return mods
}

// describeFru classifies one FRU: category, group title, muted detail,
// serial-table item and byte capacity.
func describeFru(dev *devicetypes.CaniDeviceType, fru *devicetypes.CaniFruType) hwEntry {
	e := hwEntry{kind: string(fru.Type), item: htmlHardwareItem{Slot: fruSlot(dev, fru), PartNumber: fru.PartNumber, Serial: fru.Serial}}
	switch fru.Type {
	case devicetypes.TypeCPU:
		e.category, e.title, e.meta = hwCPU, orElse(fru.Model, joinUnique(fru.Manufacturer, fru.Name)), cpuMeta(fru.CustomFields)
	case devicetypes.TypeMemory:
		e.bytes, _ = cfNumber(fru.CustomFields, "capacity_mib")
		e.bytes *= mib
		e.category, e.title, e.meta = hwMemory, joinUnique(formatBytes(e.bytes, true), fru.Model, fru.Manufacturer), memoryMeta(fru.CustomFields)
	case "disk":
		e.bytes, _ = cfNumber(fru.CustomFields, "capacity_bytes")
		e.category = hwDisk
		e.title = joinUnique(formatBytes(e.bytes, false), cfString(fru.CustomFields, "protocol"),
			cfString(fru.CustomFields, "media"), fru.Manufacturer, fru.Model)
	default:
		e.category, e.title = hwOther, orElse(joinUnique(fru.Manufacturer, fru.Model), fru.Name)
	}
	return e
}

// describeModule classifies one module as a card in its bay.
func describeModule(mod *devicetypes.CaniModuleType) hwEntry {
	return hwEntry{
		category: hwCard, kind: "card",
		title: orElse(joinUnique(mod.Manufacturer, mod.Model), mod.Name),
		item:  htmlHardwareItem{Slot: mod.ModuleBayName, PartNumber: mod.PartNumber, Serial: mod.Serial},
	}
}

// fruSlot is the slot shown in the serial table: the recorded slot or
// socket, else the FRU name without the device-name prefix.
func fruSlot(dev *devicetypes.CaniDeviceType, fru *devicetypes.CaniFruType) string {
	if slot := cfString(fru.CustomFields, "slot"); slot != "" {
		return slot
	}
	if socket := cfString(fru.CustomFields, "socket"); socket != "" {
		return socket
	}
	return strings.TrimSpace(strings.TrimPrefix(fru.Name, dev.Name))
}

// cpuMeta renders "32 cores, 64 threads" from the CPU custom fields.
func cpuMeta(fields map[string]any) string {
	var parts []string
	if cores, ok := cfNumber(fields, "cores"); ok && cores > 0 {
		parts = append(parts, pluralize(int(cores), "core"))
	}
	if threads, ok := cfNumber(fields, "threads"); ok && threads > 0 {
		parts = append(parts, pluralize(int(threads), "thread"))
	}
	return strings.Join(parts, ", ")
}

// memoryMeta renders the DIMM speed when recorded.
func memoryMeta(fields map[string]any) string {
	if speed, ok := cfNumber(fields, "speed_mhz"); ok && speed > 0 {
		return fmt.Sprintf("%d MHz", int(speed))
	}
	return ""
}

// addHardware files one entry under its group, creating the group on first sight.
func addHardware(groups map[string]*hwGroup, e hwEntry) {
	key := fmt.Sprintf("%d|%s|%s", e.category, e.title, e.meta)
	g, ok := groups[key]
	if !ok {
		g = &hwGroup{category: e.category, kind: e.kind, group: htmlHardwareGroup{Title: e.title, Meta: e.meta}}
		groups[key] = g
	}
	g.group.Count++
	g.group.Items = append(g.group.Items, e.item)
	g.bytes += e.bytes
}

// sortedHardware orders groups by category, then larger groups first, then title.
func sortedHardware(groups map[string]*hwGroup) []*hwGroup {
	ordered := make([]*hwGroup, 0, len(groups))
	for _, g := range groups {
		ordered = append(ordered, g)
	}
	sort.Slice(ordered, func(i, j int) bool {
		a, b := ordered[i], ordered[j]
		if a.category != b.category {
			return a.category < b.category
		}
		if a.group.Count != b.group.Count {
			return a.group.Count > b.group.Count
		}
		return a.group.Title < b.group.Title
	})
	return ordered
}

// finishHardware fills the derived fields of a group: the memory total and,
// for cards, the bay list as the muted detail.
func finishHardware(g *hwGroup) htmlHardwareGroup {
	out := g.group
	if g.category == hwMemory && g.group.Count > 1 && g.bytes > 0 {
		out.Total = formatBytes(g.bytes, true)
	}
	if g.category == hwCard {
		var bays []string
		for _, it := range out.Items {
			if it.Slot != "" {
				bays = append(bays, it.Slot)
			}
		}
		out.Meta = strings.Join(bays, ", ")
	}
	return out
}

// hardwareSummary renders "2 CPUs · 512 GiB · 8 drives · 3 cards".
func hardwareSummary(groups []*hwGroup) string {
	counts := map[int]int{}
	var bytes float64
	others := map[string]int{}
	for _, g := range groups {
		counts[g.category] += g.group.Count
		if g.category == hwMemory {
			bytes += g.bytes
		}
		if g.category == hwOther {
			others[g.kind] += g.group.Count
		}
	}
	var parts []string
	if n := counts[hwCPU]; n > 0 {
		parts = append(parts, pluralize(n, "CPU"))
	}
	if counts[hwMemory] > 0 {
		parts = append(parts, orElse(formatBytes(bytes, true), pluralize(counts[hwMemory], "DIMM")))
	}
	if n := counts[hwDisk]; n > 0 {
		parts = append(parts, pluralize(n, "drive"))
	}
	if n := counts[hwCard]; n > 0 {
		parts = append(parts, pluralize(n, "card"))
	}
	for _, kind := range sortedKeys(others) {
		parts = append(parts, pluralize(others[kind], strings.ReplaceAll(kind, "-", " ")))
	}
	return strings.Join(parts, " · ")
}
