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

import "testing"

func TestGenerateStarTopology(t *testing.T) {
	entries, err := GenerateStarTopology(TopologyStarParams{
		Hub:       "switch-01",
		HubPorts:  "eth{1..4}",
		Spokes:    "node-{01..04}",
		SpokePort: "eth0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (with patterns), got %d", len(entries))
	}
	if entries[0].A.Device != "node-{01..04}" {
		t.Errorf("unexpected A device: %s", entries[0].A.Device)
	}
	if entries[0].B.Port != "eth{1..4}" {
		t.Errorf("unexpected B port: %s", entries[0].B.Port)
	}
}

func TestGenerateStarTopology_MissingParams(t *testing.T) {
	_, err := GenerateStarTopology(TopologyStarParams{})
	if err == nil {
		t.Fatal("expected error for empty params")
	}
}

func TestGenerateLeafSpineTopology(t *testing.T) {
	entries, err := GenerateLeafSpineTopology(TopologyLeafSpineParams{
		Leaves:         []string{"leaf-01", "leaf-02"},
		Spines:         []string{"spine-01", "spine-02"},
		UplinksPerLeaf: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	// 2 leaves x 2 spines x 1 uplink = 4 connections
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	if entries[0].A.Device != "leaf-01" || entries[0].B.Device != "spine-01" {
		t.Errorf("entry[0]: %s -> %s", entries[0].A.Device, entries[0].B.Device)
	}
	if entries[1].A.Device != "leaf-01" || entries[1].B.Device != "spine-02" {
		t.Errorf("entry[1]: %s -> %s", entries[1].A.Device, entries[1].B.Device)
	}
}

func TestGenerateLeafSpineTopology_MultipleUplinks(t *testing.T) {
	entries, err := GenerateLeafSpineTopology(TopologyLeafSpineParams{
		Leaves:         []string{"leaf-01"},
		Spines:         []string{"spine-01"},
		UplinksPerLeaf: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries for 2 uplinks, got %d", len(entries))
	}
	if entries[0].A.Port != "eth49" {
		t.Errorf("expected eth49, got %s", entries[0].A.Port)
	}
	if entries[1].A.Port != "eth50" {
		t.Errorf("expected eth50, got %s", entries[1].A.Port)
	}
}

func TestGenerateLeafSpineTopology_MissingParams(t *testing.T) {
	_, err := GenerateLeafSpineTopology(TopologyLeafSpineParams{})
	if err == nil {
		t.Fatal("expected error for empty params")
	}
}

func TestGenerateRingTopology(t *testing.T) {
	devices := []string{"sw-01", "sw-02", "sw-03", "sw-04"}
	entries, err := GenerateRingTopology(TopologyRingParams{
		Devices: devices,
		PortA:   "eth0",
		PortB:   "eth1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}
	last := entries[3]
	if last.A.Device != "sw-04" || last.B.Device != "sw-01" {
		t.Errorf("ring wrap: %s -> %s", last.A.Device, last.B.Device)
	}
}

func TestGenerateRingTopology_TooFew(t *testing.T) {
	_, err := GenerateRingTopology(TopologyRingParams{
		Devices: []string{"only-one"},
	})
	if err == nil {
		t.Fatal("expected error for < 2 devices")
	}
}
