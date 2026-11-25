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
	"sort"
	"strings"
	"testing"
	"time"
)

func TestNaturalLessOrdersBreakoutsAndNumbers(t *testing.T) {
	ports := []string{"mgmt", "1/1/49", "1/1/2", "1/1/1:4", "1/1/10", "1/1/1:3", "1/1/1", "21", "3"}
	sort.Slice(ports, func(i, j int) bool { return naturalLess(ports[i], ports[j]) })
	want := "1/1/1 1/1/1:3 1/1/1:4 1/1/2 1/1/10 1/1/49 3 21 mgmt"
	if got := strings.Join(ports, " "); got != want {
		t.Errorf("natural order:\n got %s\nwant %s", got, want)
	}
}

func TestBuildSwitchTableRowsFollowPortOrder(t *testing.T) {
	inv := cablingInventory()
	idx := newHTMLIndex(inv)
	table := buildSwitchTable(idx, cablingDevice(t, inv, cab8325ID), "Rack-X U12")
	if table == nil {
		t.Fatal("expected a table for the cabled switch")
	}
	if table.Cabled != 2 || table.Total != 5 || !table.HasVLAN || table.Where != "Rack-X U12" {
		t.Errorf("summary: cabled %d of %d, vlan %v, where %q", table.Cabled, table.Total, table.HasVLAN, table.Where)
	}
	if len(table.Rows) != 2 || table.Rows[0].Port != "1/1/1:1" || table.Rows[1].Port != "1/1/2" {
		t.Fatalf("rows out of port order: %+v", table.Rows)
	}
	if r := table.Rows[0]; r.Description != "srv-eth" || r.PeerName != "srv" || r.PeerPort != "Slot 10 Port 1" || r.VLAN != "1083" || r.PeerIP != "10.0.83.5" || r.PeerIPURL != "" {
		t.Errorf("breakout access row: %+v", r)
	}
	if r := table.Rows[1]; r.Description != "" || r.PeerName != "sw6300" || r.VLAN != "trunk" || r.PeerIP != "" {
		t.Errorf("unlabeled trunk row keeps the port name only: %+v", r)
	}
	if table.Free != "1/1/1:2\u20131/1/1:4" || table.FreeCount != 3 {
		t.Errorf("free ports: %q (%d)", table.Free, table.FreeCount)
	}
}

func TestBuildSwitchTableFreeRangesAndMgmtLinks(t *testing.T) {
	inv := cablingInventory()
	table := buildSwitchTable(newHTMLIndex(inv), cablingDevice(t, inv, cab6300ID), "Rack-X U10")
	if table == nil {
		t.Fatal("expected a table")
	}
	if table.Free != "1/1/3\u20131/1/4" || table.FreeCount != 2 || table.Cabled != 3 || table.Total != 5 {
		t.Errorf("free %q (%d), cabled %d of %d", table.Free, table.FreeCount, table.Cabled, table.Total)
	}
	if r := table.Rows[0]; r.Port != "1/1/1" || r.Description != "srv-ilo" || r.PeerIP != "10.0.84.5" || r.PeerIPURL != "https://10.0.84.5/" {
		t.Errorf("iLO row must link the peer address: %+v", r)
	}
	if r := table.Rows[1]; r.Port != "1/1/2" || r.PeerName != `evil"]<` || r.PeerIP != "" {
		t.Errorf("row for the hostile device: %+v", r)
	}
}

func TestPortRanges(t *testing.T) {
	cases := map[string][]string{
		"":                                       nil,
		"1/1/5":                                  {"1/1/5"},
		"1/1/25\u20131/1/48, 1/1/51\u20131/1/52": {"1/1/25", "1/1/26", "1/1/27", "1/1/28", "1/1/29", "1/1/30", "1/1/31", "1/1/32", "1/1/33", "1/1/34", "1/1/35", "1/1/36", "1/1/37", "1/1/38", "1/1/39", "1/1/40", "1/1/41", "1/1/42", "1/1/43", "1/1/44", "1/1/45", "1/1/46", "1/1/47", "1/1/48", "1/1/51", "1/1/52"},
		"1/1/1:3\u20131/1/1:4, 1/1/2":            {"1/1/1:3", "1/1/1:4", "1/1/2"},
		"11\u201313, 20":                         {"11", "12", "13", "20"},
		"a, b":                                   {"a", "b"},
	}
	for want, ports := range cases {
		if got := portRanges(ports); got != want {
			t.Errorf("portRanges(%v) = %q, want %q", ports, got, want)
		}
	}
}

func TestSwitchWithoutCablesHasNoTable(t *testing.T) {
	inv := cablingInventory()
	if table := buildSwitchTable(newHTMLIndex(inv), cablingDevice(t, inv, cabIdleID), ""); table != nil {
		t.Errorf("expected no table for the idle switch, got %+v", table)
	}
	out := renderCabling(t)
	if strings.Contains(out, `id="sw-`+cabIdleID.String()+`"`) {
		t.Errorf("idle switch must not get a cabling table")
	}
	for _, id := range []string{cab6300ID.String(), cab8325ID.String()} {
		if !strings.Contains(out, `<details open id="sw-`+id+`" class="swt" data-dev="`+id+`">`) {
			t.Errorf("expected an open cabling table for switch %s", id)
		}
	}
}

func TestEveryCableAppearsOncePerSwitchEnd(t *testing.T) {
	page := buildHTMLPage(cablingInventory(), "", time.Unix(0, 0))
	rows := 0
	for _, table := range page.Switches {
		rows += len(table.Rows)
	}
	// iLO and NIC cables: one switch end each; trunk: two; hostile device: one.
	if rows != 5 {
		t.Errorf("expected 5 switch-side rows across all tables, got %d", rows)
	}
	if len(page.Switches) != 2 || page.Switches[0].Name != "sw8325" || page.Switches[1].Name != "sw6300" {
		t.Errorf("tables should follow rack order top-down, got %+v", page.Switches)
	}
	if page.CableCount != 4 {
		t.Errorf("expected 4 cables counted on the page, got %d", page.CableCount)
	}
}
