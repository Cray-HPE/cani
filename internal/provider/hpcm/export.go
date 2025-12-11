/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
package hpcm

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/spf13/cobra"
)

// openChamiBMC represents the minimal BMC entry expected by OpenCHAMI's nodes.yaml
type openChamiBMC struct {
	Xname string `yaml:"xname"`
	IP    string `yaml:"ip"`
	MAC   string `yaml:"mac"`
}

// openChamiNode represents the minimal node entry expected by OpenCHAMI's nodes.yaml
type openChamiNode struct {
	Xname       string   `yaml:"xname"`
	IP          string   `yaml:"ip"`
	BootMAC     string   `yaml:"boot_mac"`
	NID         *int     `yaml:"nid,omitempty"`
	Hostname    string   `yaml:"hostname,omitempty"`
	HostAliases []string `yaml:"host_aliases,omitempty"`
}

// openChamiPayload is the full document layout
type openChamiPayload struct {
	BMCs  []openChamiBMC  `yaml:"bmcs"`
	Nodes []openChamiNode `yaml:"nodes"`
}

func (hpcm *Hpcm) Export(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	switch exportFormat {
	case "openchami":
		return hpcm.exportOpenChami(cmd, args, datastore)
	default:
		return fmt.Errorf("the requested format, %s, is unsupported for HPCM provider", exportFormat)
	}
}

// exportOpenChami renders a nodes.yaml compatible with OpenCHAMI.
// It is provider-agnostic as long as LocationPath is populated and basic
// network metadata exists in ProviderMetadata or Properties.
func (hpcm *Hpcm) exportOpenChami(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	inv, err := datastore.List()
	if err != nil {
		return err
	}

	var payload openChamiPayload

	for id, hw := range inv.Hardware {
		if hw.Type != hardwaretypes.Node {
			continue
		}

		// Build xname from LocationPath; fall back to Name or ID
		nodeXname := buildXnameString(hw)
		if nodeXname == "" {
			nodeXname = hw.Name
		}
		if nodeXname == "" {
			nodeXname = id.String()
		}

		// Pull node metadata from HPCM provider metadata
		nodeIP := firstNonEmpty(
			extractStringFromMetadata(hw, "IP4addr"),
			extractStringFromProperties(hw, "IP4addr"),
		)
		nodeMAC := firstNonEmpty(
			extractStringFromMetadata(hw, "MACaddr"),
			extractStringFromProperties(hw, "MACaddr"),
		)
		nodeNID := extractIntPtrFromMetadata(hw, "Nid")
		nodeAliases := extractStringSliceFromMetadata(hw, "Alias")
		nodeHostname := ""
		if len(nodeAliases) > 0 {
			nodeHostname = nodeAliases[0]
		}

		payload.Nodes = append(payload.Nodes, openChamiNode{
			Xname:       nodeXname,
			IP:          nodeIP,
			BootMAC:     nodeMAC,
			NID:         nodeNID,
			Hostname:    nodeHostname,
			HostAliases: nodeAliases,
		})

		// Locate BMC child
		children, err := datastore.GetChildren(id)
		if err != nil {
			return err
		}
		for _, child := range children {
			if child.Type != hardwaretypes.NodeController {
				continue
			}

			bmcXname := buildXnameString(child)
			if bmcXname == "" {
				bmcXname = child.Name
			}
			if bmcXname == "" {
				bmcXname = child.ID.String()
			}

			bmcIP := firstNonEmpty(
				extractStringFromMetadata(child, "IP4addr"),
				extractStringFromProperties(child, "IP4addr"),
			)
			bmcMAC := firstNonEmpty(
				extractStringFromMetadata(child, "MACaddr"),
				extractStringFromProperties(child, "MACaddr"),
			)

			payload.BMCs = append(payload.BMCs, openChamiBMC{
				Xname: bmcXname,
				IP:    bmcIP,
				MAC:   bmcMAC,
			})
		}
	}

	// Sort for stable output
	sort.Slice(payload.Nodes, func(i, j int) bool { return payload.Nodes[i].Xname < payload.Nodes[j].Xname })
	sort.Slice(payload.BMCs, func(i, j int) bool { return payload.BMCs[i].Xname < payload.BMCs[j].Xname })

	out, err := yaml.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = cmd.OutOrStdout().Write(out)
	return err
}

// buildXnameString attempts to generate a Cray xname from the LocationPath.
// Returns an empty string if generation fails.
func buildXnameString(hw inventory.Hardware) string {
	if len(hw.LocationPath) == 0 {
		return ""
	}

	// For HPCM, we use the location path but can't call BuildXname directly
	// since it's a CSM-specific function. Instead, generate a simple hierarchy string.
	// In a real implementation, you might want to implement provider-specific xname building.
	var parts []string
	for _, loc := range hw.LocationPath {
		parts = append(parts, fmt.Sprintf("%s%d", strings.ToLower(string(loc.HardwareType))[:1], loc.Ordinal))
	}
	return strings.Join(parts, "")
}

func extractStringFromMetadata(hw inventory.Hardware, key string) string {
	meta, ok := hw.ProviderMetadata[inventory.HPCMProvider]
	if !ok {
		return ""
	}
	return lookupString(meta, key)
}

func extractStringFromProperties(hw inventory.Hardware, key string) string {
	if hw.Properties == nil {
		return ""
	}
	return lookupString(hw.Properties, key)
}

func extractStringSliceFromMetadata(hw inventory.Hardware, key string) []string {
	meta, ok := hw.ProviderMetadata[inventory.HPCMProvider]
	if !ok {
		return nil
	}
	return lookupStringSlice(meta, key)
}

func extractIntPtrFromMetadata(hw inventory.Hardware, key string) *int {
	meta, ok := hw.ProviderMetadata[inventory.HPCMProvider]
	if !ok {
		return nil
	}
	return lookupIntPtr(meta, key)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func lookupString(src interface{}, key string) string {
	m, ok := src.(map[string]interface{})
	if !ok {
		return ""
	}
	if val, exists := m[key]; exists {
		switch v := val.(type) {
		case string:
			return v
		case fmt.Stringer:
			return v.String()
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		case int:
			return strconv.Itoa(v)
		}
	}
	return ""
}

func lookupStringSlice(src interface{}, key string) []string {
	m, ok := src.(map[string]interface{})
	if !ok {
		return nil
	}
	val, exists := m[key]
	if !exists {
		return nil
	}
	switch v := val.(type) {
	case []string:
		return v
	case []interface{}:
		var out []string
		for _, e := range v {
			out = append(out, fmt.Sprintf("%v", e))
		}
		return out
	default:
		return []string{fmt.Sprintf("%v", v)}
	}
}

func lookupIntPtr(src interface{}, key string) *int {
	m, ok := src.(map[string]interface{})
	if !ok {
		return nil
	}
	if val, exists := m[key]; exists {
		switch v := val.(type) {
		case float64:
			i := int(v)
			return &i
		case int:
			return &v
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return &i
			}
		}
	}
	return nil
}
