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
package add

import (
	"fmt"
	"os"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/nameexpand"
	"github.com/Cray-HPE/cani/pkg/devicetypes/connections"
	"gopkg.in/yaml.v3"
)

// newGenerateCommand creates the "add connections generate" subcommand.
func newGenerateCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "generate <pattern>",
		Short: "Generate a connection map from a topology pattern.",
		Long: `Generate a connection map YAML file from a named topology pattern.

Supported patterns: star, leaf-spine, ring

Examples:
  cani alpha add connections generate star --hub switch-01 --spokes "node-{01..48}" > topo.yaml
  cani alpha add connections generate leaf-spine --leaves "leaf-{01..04}" --spines "spine-{01..02}"
  cani alpha add connections generate ring --devices "sw-{01..08}"`,
		Args: cli.ExactArgs(1),
		RunE: generateTopology,
	}

	// Star flags
	cmd.Flags().String("hub", "", "Hub device name (star topology)")
	cmd.Flags().String("hub-ports", "", "Port pattern on hub (e.g. eth{1..48})")
	cmd.Flags().String("spokes", "", "Spoke device pattern (e.g. node-{01..48})")
	cmd.Flags().String("spoke-port", "eth0", "Port name on each spoke")

	// Leaf-spine flags
	cmd.Flags().StringSlice("leaves", nil, "Leaf switch names or pattern (leaf-spine topology)")
	cmd.Flags().StringSlice("spines", nil, "Spine switch names or pattern (leaf-spine topology)")
	cmd.Flags().Int("uplinks-per-leaf", 1, "Number of uplinks from each leaf to each spine")

	// Ring flags
	cmd.Flags().StringSlice("devices", nil, "Ordered device names for ring topology")
	cmd.Flags().String("port-a", "eth0", "Port for forward ring link")
	cmd.Flags().String("port-b", "eth1", "Port for reverse ring link")

	// Common
	cmd.Flags().String("cable-type", "", "Default cable type for generated connections")
	cmd.Flags().String("cable-color", "", "Default cable color")

	return cmd
}

func generateTopology(cmd *cli.Command, args []string) error {
	pattern := args[0]
	cableType, _ := cmd.Flags().GetString("cable-type")
	cableColor, _ := cmd.Flags().GetString("cable-color")

	var entries []connections.ConnectionEntry
	var err error

	switch pattern {
	case "star":
		entries, err = generateStarFromFlags(cmd)
	case "leaf-spine":
		entries, err = generateLeafSpineFromFlags(cmd)
	case "ring":
		entries, err = generateRingFromFlags(cmd)
	default:
		return fmt.Errorf("unknown topology pattern: %s (supported: star, leaf-spine, ring)", pattern)
	}
	if err != nil {
		return err
	}

	cm := connections.ConnectionMap{
		Version:     "v1",
		Connections: entries,
	}
	if cableType != "" || cableColor != "" {
		cm.CableDefaults = &connections.CableDefaults{
			Type:  cableType,
			Color: cableColor,
		}
	}

	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	return enc.Encode(cm)
}

func generateStarFromFlags(cmd *cli.Command) ([]connections.ConnectionEntry, error) {
	hub, _ := cmd.Flags().GetString("hub")
	hubPorts, _ := cmd.Flags().GetString("hub-ports")
	spokes, _ := cmd.Flags().GetString("spokes")
	spokePort, _ := cmd.Flags().GetString("spoke-port")

	return connections.GenerateStarTopology(connections.TopologyStarParams{
		Hub:       hub,
		HubPorts:  hubPorts,
		Spokes:    spokes,
		SpokePort: spokePort,
	})
}

func generateLeafSpineFromFlags(cmd *cli.Command) ([]connections.ConnectionEntry, error) {
	leaves, _ := cmd.Flags().GetStringSlice("leaves")
	spines, _ := cmd.Flags().GetStringSlice("spines")
	uplinks, _ := cmd.Flags().GetInt("uplinks-per-leaf")

	leaves = expandStringSlice(leaves)
	spines = expandStringSlice(spines)

	return connections.GenerateLeafSpineTopology(connections.TopologyLeafSpineParams{
		Leaves:         leaves,
		Spines:         spines,
		UplinksPerLeaf: uplinks,
	})
}

func generateRingFromFlags(cmd *cli.Command) ([]connections.ConnectionEntry, error) {
	devices, _ := cmd.Flags().GetStringSlice("devices")
	portA, _ := cmd.Flags().GetString("port-a")
	portB, _ := cmd.Flags().GetString("port-b")

	devices = expandStringSlice(devices)

	return connections.GenerateRingTopology(connections.TopologyRingParams{
		Devices: devices,
		PortA:   portA,
		PortB:   portB,
	})
}

// expandStringSlice expands brace patterns in a string slice.
func expandStringSlice(s []string) []string {
	var result []string
	for _, item := range s {
		expanded, err := nameexpand.Expand(item)
		if err != nil || len(expanded) == 0 {
			result = append(result, item)
		} else {
			result = append(result, expanded...)
		}
	}
	return result
}
