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

// TestGenerateStarTopology verifies a star topology emits one pattern-bearing
// connection linking the spoke pattern to the hub on the given ports.
//
// Why it matters: the generator emits declarative patterns (not pre-expanded
// links), so a star must produce a single entry that the resolver later expands.
// Inputs: a hub, hub-port pattern eth{1..4}, spoke pattern node-{01..04}, and
// spoke port eth0. Outputs: one entry whose A device is the spoke pattern and B
// port is the hub-port pattern.
// Data choice: 4-wide brace ranges on both ends confirm the generator passes
// patterns through verbatim rather than expanding them itself.
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

// TestGenerateStarTopology_MissingParams verifies the star generator errors when
// required hub or spokes are absent.
//
// Why it matters: a star with no hub or no spokes is meaningless, so the
// generator must reject it instead of emitting an empty or half-formed entry.
// Inputs: an empty TopologyStarParams. Outputs: a non-nil error.
// Data choice: the zero-value params omit both required fields at once, the
// minimal trigger for the required-params guard.
func TestGenerateStarTopology_MissingParams(t *testing.T) {
	_, err := GenerateStarTopology(TopologyStarParams{})
	if err == nil {
		t.Fatal("expected error for empty params")
	}
}

// TestGenerateLeafSpineTopology verifies a leaf-spine fabric connects every leaf
// to every spine, ordered leaf-major then spine.
//
// Why it matters: a fabric's correctness is its full leaf×spine mesh, so the
// generator must emit exactly that set in a predictable order for stable output.
// Inputs: two leaves, two spines, one uplink each. Outputs: four entries with
// entry[0] leaf-01→spine-01 and entry[1] leaf-01→spine-02.
// Data choice: a 2×2 fabric is the smallest that distinguishes leaf-major from
// spine-major ordering and proves the full cross product is produced.
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

// TestGenerateLeafSpineTopology_MultipleUplinks verifies multiple uplinks per
// leaf yield sequentially numbered leaf ports.
//
// Why it matters: redundant fabric uplinks must land on distinct, predictably
// numbered ports so generated cabling matches physical port allocation.
// Inputs: one leaf, one spine, two uplinks per leaf. Outputs: two entries whose
// A ports are eth49 and eth50.
// Data choice: two uplinks on a single leaf/spine isolates the per-uplink port
// increment from the default eth49 base without fabric-size noise.
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

// TestGenerateLeafSpineTopology_MissingParams verifies the fabric generator
// errors when leaves or spines are absent.
//
// Why it matters: a fabric needs at least one leaf and one spine, so empty input
// must be rejected rather than producing an empty connection set.
// Inputs: an empty TopologyLeafSpineParams. Outputs: a non-nil error.
// Data choice: the zero-value params omit both required slices at once, the
// minimal trigger for the required-params guard.
func TestGenerateLeafSpineTopology_MissingParams(t *testing.T) {
	_, err := GenerateLeafSpineTopology(TopologyLeafSpineParams{})
	if err == nil {
		t.Fatal("expected error for empty params")
	}
}

// TestGenerateRingTopology verifies a ring links each device to its successor and
// wraps the last device back to the first.
//
// Why it matters: a ring's defining property is the wrap-around closure, so the
// generator must connect the final device back to the first to form the loop.
// Inputs: four ordered devices with ports eth0/eth1. Outputs: four entries whose
// last entry links sw-04 back to sw-01.
// Data choice: four devices make the modulo wrap unambiguous and confirm the
// closing link is generated rather than a dangling chain.
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

// TestGenerateRingTopology_TooFew verifies a ring of fewer than two devices is
// rejected.
//
// Why it matters: a ring needs at least two devices to form a link, so a
// single-device request is invalid and must error rather than emit a self-loop.
// Inputs: a Devices slice with one entry. Outputs: a non-nil error.
// Data choice: exactly one device is the boundary case just below the two-device
// minimum the generator enforces.
func TestGenerateRingTopology_TooFew(t *testing.T) {
	_, err := GenerateRingTopology(TopologyRingParams{
		Devices: []string{"only-one"},
	})
	if err == nil {
		t.Fatal("expected error for < 2 devices")
	}
}

// ========== additional branch-coverage tests ==========

// TestGenerateStarTopology_Defaults verifies the star generator fills default
// ports when spoke port and hub ports are omitted.
//
// Why it matters: callers may give only hub and spokes, so the generator must
// supply sensible defaults (eth0 on the spoke, auto on the hub) rather than emit
// blank ports.
// Inputs: a hub and spoke pattern with no SpokePort or HubPorts. Outputs: one
// entry whose A port is eth0 and B port is auto.
// Data choice: leaving both port fields empty drives both default-assignment
// branches in a single entry.
func TestGenerateStarTopology_Defaults(t *testing.T) {
	entries, err := GenerateStarTopology(TopologyStarParams{
		Hub:    "switch-01",
		Spokes: "node-{01..04}",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].A.Port != "eth0" {
		t.Errorf("spoke port = %q, want default eth0", entries[0].A.Port)
	}
	if entries[0].B.Port != "auto" {
		t.Errorf("hub port = %q, want default auto", entries[0].B.Port)
	}
}

// TestGenerateLeafSpineTopology_ExplicitPorts verifies explicit leaf/spine ports
// are used verbatim and a zero uplink count defaults to one.
//
// Why it matters: operators often pin uplinks to specific ports, so an explicit
// LeafUplinkPort and SpinePort must override the numeric defaults while
// UplinksPerLeaf=0 still yields one link.
// Inputs: one leaf, one spine, UplinksPerLeaf 0, LeafUplinkPort eth49, SpinePort
// eth1. Outputs: one entry with A port eth49 and B port eth1.
// Data choice: combining the zero-uplink default with explicit ports exercises
// both the uplink-floor and explicit-port branches in one entry.
func TestGenerateLeafSpineTopology_ExplicitPorts(t *testing.T) {
	entries, err := GenerateLeafSpineTopology(TopologyLeafSpineParams{
		Leaves:         []string{"leaf-01"},
		Spines:         []string{"spine-01"},
		UplinksPerLeaf: 0,
		LeafUplinkPort: "eth49",
		SpinePort:      "eth1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].A.Port != "eth49" {
		t.Errorf("leaf port = %q, want eth49", entries[0].A.Port)
	}
	if entries[0].B.Port != "eth1" {
		t.Errorf("spine port = %q, want eth1", entries[0].B.Port)
	}
}

// TestGenerateRingTopology_DefaultPorts verifies the ring generator fills default
// ports when PortA and PortB are omitted.
//
// Why it matters: a ring described only by its devices must still produce usable
// links, so the generator defaults the two ring ports to eth0 and eth1.
// Inputs: two devices with no PortA or PortB. Outputs: two entries whose A port
// is eth0 and B port is eth1.
// Data choice: two devices with both ports omitted drives both default-port
// branches with the minimal valid ring.
func TestGenerateRingTopology_DefaultPorts(t *testing.T) {
	entries, err := GenerateRingTopology(TopologyRingParams{
		Devices: []string{"sw-01", "sw-02"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].A.Port != "eth0" {
		t.Errorf("port A = %q, want default eth0", entries[0].A.Port)
	}
	if entries[0].B.Port != "eth1" {
		t.Errorf("port B = %q, want default eth1", entries[0].B.Port)
	}
}
