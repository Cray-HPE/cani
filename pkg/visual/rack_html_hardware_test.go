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
	"strings"
	"testing"
)

func TestBuildHardwareGroupsFrusAndModules(t *testing.T) {
	inv := cablingInventory()
	groups, summary := buildHardware(inv, cablingDevice(t, inv, cabServerID))
	if summary != "2 CPUs · 528 GiB · 1 drive · 2 cards" {
		t.Errorf("summary = %q", summary)
	}
	if len(groups) != 5 {
		t.Fatalf("expected 5 groups (CPU, 2 DIMM populations, drive, GPUs), got %d: %+v", len(groups), groups)
	}
	if g := groups[0]; g.Count != 2 || g.Title != "EPYC 7543 32-Core" || g.Meta != "32 cores, 64 threads" || g.Items[0].Slot != "Proc 1" || g.Items[1].Serial != "CPU2" {
		t.Errorf("CPU group: %+v", g)
	}
	if g := groups[1]; g.Count != 16 || g.Title != "32 GiB DDR4 Acme" || g.Meta != "3200 MHz" || g.Total != "512 GiB" || !g.HasSerials() {
		t.Errorf("main DIMM population: %+v", g)
	}
	if g := groups[2]; g.Count != 1 || g.Title != "16 GiB DDR4 Acme" || g.Total != "" || g.Items[0].Slot != "PROC 2 DIMM 1" || g.Items[0].PartNumber != "PN-16 GiB DDR4" {
		t.Errorf("odd DIMM must stay its own line: %+v", g)
	}
	if g := groups[3]; g.Count != 1 || g.Title != "1.92 TB NVMe SSD Acme PM9A3" || g.Items[0].Slot != "Box 1 Bay 1" {
		t.Errorf("drive group: %+v", g)
	}
	if g := groups[4]; g.Count != 2 || g.Title != "NVIDIA A100 80GB" || g.Meta != "PCIe1, PCIe3" || g.Items[1].Serial != "GPU2" || g.Items[1].PartNumber != "900-21001" {
		t.Errorf("GPU group: %+v", g)
	}
}

func TestBuildHardwareIsAbsentWithoutData(t *testing.T) {
	inv := cablingInventory()
	groups, summary := buildHardware(inv, cablingDevice(t, inv, cab6300ID))
	if len(groups) != 0 || summary != "" {
		t.Errorf("switch without FRUs: groups %+v, summary %q", groups, summary)
	}
	out := renderCabling(t)
	if got := strings.Count(out, `<span class="sect-title">Hardware</span>`); got != 1 {
		t.Errorf("only the server has hardware, expected 1 hardware section, got %d", got)
	}
}

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		n      float64
		binary bool
		want   string
	}{
		{34359738368, true, "32 GiB"},
		{16384 * mib, true, "16 GiB"},
		{549755813888, true, "512 GiB"},
		{1649267441664, true, "1.5 TiB"},
		{1920383410176, false, "1.92 TB"},
		{960197124096, false, "960 GB"},
		{2000398934016, false, "2 TB"},
		{15360950534144, false, "15.4 TB"},
		{0, false, ""},
	}
	for _, c := range cases {
		if got := formatBytes(c.n, c.binary); got != c.want {
			t.Errorf("formatBytes(%v, %v) = %q, want %q", c.n, c.binary, got, c.want)
		}
	}
}

func TestCustomFieldNumbersAcceptIntAndFloat(t *testing.T) {
	fields := map[string]any{"i": 32, "f": float64(32768), "s": "text", "i64": int64(7)}
	for key, want := range map[string]float64{"i": 32, "f": 32768, "i64": 7} {
		if got, ok := cfNumber(fields, key); !ok || got != want {
			t.Errorf("cfNumber(%q) = %v, %v; want %v", key, got, ok, want)
		}
	}
	if _, ok := cfNumber(fields, "s"); ok {
		t.Errorf("a string field is not a number")
	}
	if _, ok := cfNumber(nil, "missing"); ok {
		t.Errorf("a missing field is not a number")
	}
	if cfString(fields, "s") != "text" || cfString(fields, "i") != "" {
		t.Errorf("cfString must only return string values")
	}
}

func TestJoinUniqueDropsRepeatedParts(t *testing.T) {
	cases := []struct {
		parts []string
		want  string
	}{
		{[]string{"32 GiB", "32 GiB DDR4", "Samsung"}, "32 GiB DDR4 Samsung"},
		{[]string{"1.92 TB", "NVMe", "SSD", "Samsung", "Samsung PM9A3"}, "1.92 TB NVMe SSD Samsung PM9A3"},
		{[]string{"NVIDIA", "A100 80GB", "A100 80GB"}, "NVIDIA A100 80GB"},
		{[]string{"", "  ", "x"}, "x"},
	}
	for _, c := range cases {
		if got := joinUnique(c.parts...); got != c.want {
			t.Errorf("joinUnique(%q) = %q, want %q", c.parts, got, c.want)
		}
	}
}

func TestRenderedHardwareSection(t *testing.T) {
	out := renderCabling(t)
	for _, want := range []string{
		`<span class="sect-meta">2 CPUs · 528 GiB · 1 drive · 2 cards</span>`,
		`<span class="hw-count">16&#215;</span> <span class="hw-text">32 GiB DDR4 Acme <span class="hw-meta">3200 MHz</span> <span class="hw-meta">(512 GiB)</span></span>`,
		`<details class="serials"><summary>serials</summary>`,
		`<td>PCIe3</td><td class="mono">900-21001</td><td class="mono">GPU2</td>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected hardware section to contain %q", want)
		}
	}
}
