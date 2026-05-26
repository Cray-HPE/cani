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
)

// TopologyStarParams configures a star (hub-spoke) topology.
type TopologyStarParams struct {
	Hub       string // Hub device name
	HubPorts  string // Port pattern on hub (e.g. "eth{1..48}")
	Spokes    string // Spoke device pattern (e.g. "node-{01..48}")
	SpokePort string // Port name on each spoke (e.g. "eth0")
}

// GenerateStarTopology creates connections for a star topology where
// every spoke connects to a single hub. Hub ports and spoke devices
// are expanded and zipped.
func GenerateStarTopology(p TopologyStarParams) ([]ConnectionEntry, error) {
	if p.Hub == "" || p.Spokes == "" {
		return nil, fmt.Errorf("hub and spokes are required")
	}
	spokePort := p.SpokePort
	if spokePort == "" {
		spokePort = "eth0"
	}
	hubPorts := p.HubPorts
	if hubPorts == "" {
		hubPorts = "auto"
	}
	return []ConnectionEntry{
		{
			A: Endpoint{Device: p.Spokes, Port: spokePort},
			B: Endpoint{Device: p.Hub, Port: hubPorts},
		},
	}, nil
}

// TopologyLeafSpineParams configures a leaf-spine fabric topology.
type TopologyLeafSpineParams struct {
	Leaves         []string // Leaf switch names
	Spines         []string // Spine switch names
	UplinksPerLeaf int      // Number of uplinks from each leaf to each spine
	LeafUplinkPort string   // Port pattern on leaves (e.g. "eth49")
	SpinePort      string   // Port pattern on spines (e.g. "eth1")
}

// GenerateLeafSpineTopology creates connections for a leaf-spine fabric.
// Each leaf connects to every spine with UplinksPerLeaf links.
func GenerateLeafSpineTopology(p TopologyLeafSpineParams) ([]ConnectionEntry, error) {
	if len(p.Leaves) == 0 || len(p.Spines) == 0 {
		return nil, fmt.Errorf("at least one leaf and one spine are required")
	}
	uplinks := p.UplinksPerLeaf
	if uplinks < 1 {
		uplinks = 1
	}

	var entries []ConnectionEntry
	spinePort := 1
	for _, leaf := range p.Leaves {
		leafPort := 49 // conventional first uplink port
		if p.LeafUplinkPort != "" {
			leafPort = 0 // will use explicit pattern
		}
		for _, spine := range p.Spines {
			for u := range uplinks {
				aPort := p.LeafUplinkPort
				if aPort == "" {
					aPort = fmt.Sprintf("eth%d", leafPort+u)
				}
				bPort := p.SpinePort
				if bPort == "" {
					bPort = fmt.Sprintf("eth%d", spinePort)
				}
				entries = append(entries, ConnectionEntry{
					A: Endpoint{Device: leaf, Port: aPort},
					B: Endpoint{Device: spine, Port: bPort},
				})
				spinePort++
			}
			if p.LeafUplinkPort == "" {
				leafPort += uplinks
			}
		}
	}
	return entries, nil
}

// TopologyRingParams configures a ring topology.
type TopologyRingParams struct {
	Devices []string // Ordered device names forming the ring
	PortA   string   // Port used for the "next" link
	PortB   string   // Port used for the "prev" link
}

// GenerateRingTopology creates connections that form a ring where each
// device connects to its neighbor. The last device connects back to
// the first.
func GenerateRingTopology(p TopologyRingParams) ([]ConnectionEntry, error) {
	if len(p.Devices) < 2 {
		return nil, fmt.Errorf("ring topology requires at least 2 devices")
	}
	portA := p.PortA
	if portA == "" {
		portA = "eth0"
	}
	portB := p.PortB
	if portB == "" {
		portB = "eth1"
	}

	entries := make([]ConnectionEntry, len(p.Devices))
	for i := range p.Devices {
		next := (i + 1) % len(p.Devices)
		entries[i] = ConnectionEntry{
			A: Endpoint{Device: p.Devices[i], Port: portA},
			B: Endpoint{Device: p.Devices[next], Port: portB},
		}
	}
	return entries, nil
}
