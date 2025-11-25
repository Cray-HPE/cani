/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
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
package show

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/spf13/cobra"
)

// ListCablesCmd represents the cables list command
var ListCablesCmd = &cobra.Command{
	Use:   "cables",
	Short: "List cables in the inventory.",
	Long: `List cables in the inventory with their type and connection status.

Examples:
  cani show cables                    # Show all cables
  cani show cables --format table     # Show as table
  cani show cables --format json      # Show as JSON
  cani show cables --unconnected      # Show only unconnected cables`,
	RunE: listCables,
}

func init() {
	ListCablesCmd.Flags().Bool("unconnected", false, "Show only unconnected cables")
}

// CableDisplay represents a cable for display
type CableDisplay struct {
	Name          string `json:"name"`
	CableType     string `json:"cableType"`
	Network       string `json:"network"`
	ATermination  string `json:"aTermination"`
	BTermination  string `json:"bTermination"`
	Connected     bool   `json:"connected"`
	ProductNumber string `json:"productNumber,omitempty"`
}

// listCables lists cables in the inventory
func listCables(cmd *cobra.Command, args []string) error {
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inv, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load device store: %w", err)
	}

	// Filter to cables only
	var cables []CableDisplay
	for _, device := range inv.Devices {
		if device == nil {
			continue
		}

		// Check if this is a cable (has CableTypeSlug or hardware type is cable)
		if device.IsCable() {
			cable := CableDisplay{
				Name: device.Name,
				// CableType: device.Slug,
				Connected: false, // Will be updated based on metadata
			}

			// Extract additional info from metadata
			if device.ProviderMetadata != nil {
				if network, ok := device.ProviderMetadata["Network"].(string); ok {
					cable.Network = network
				}
				if pn, ok := device.ProviderMetadata["HPEProductNumber"].(string); ok {
					cable.ProductNumber = pn
				}
			}

			// Infer network from cable type if not set
			if cable.Network == "" {
				cable.Network = inferNetworkFromCableType(cable.CableType)
			}

			cables = append(cables, cable)
		}
	}

	// Apply unconnected filter if set
	unconnectedOnly, _ := cmd.Flags().GetBool("unconnected")
	if unconnectedOnly {
		var filtered []CableDisplay
		for _, c := range cables {
			if !c.Connected {
				filtered = append(filtered, c)
			}
		}
		cables = filtered
	}

	// Output based on format
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(cables)

	case "table":
		fallthrough
	default:
		return printCablesTable(cables)
	}
}

// printCablesTable prints cables in a formatted table
func printCablesTable(cables []CableDisplay) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "NAME\tTYPE\tNETWORK\tA TERMINATION\tB TERMINATION\tCONNECTED")
	fmt.Fprintln(w, "----\t----\t-------\t-------------\t-------------\t---------")

	for _, c := range cables {
		aterm := c.ATermination
		if aterm == "" {
			aterm = "-"
		}
		bterm := c.BTermination
		if bterm == "" {
			bterm = "-"
		}
		network := c.Network
		if network == "" {
			network = "-"
		}
		connected := "No"
		if c.Connected {
			connected = "Yes"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			truncate(c.Name, 40),
			c.CableType,
			network,
			aterm,
			bterm,
			connected,
		)
	}

	fmt.Printf("\nTotal: %d cable(s)\n", len(cables))
	return nil
}

// inferNetworkFromCableType infers the network type from cable type slug
func inferNetworkFromCableType(cableType string) string {
	switch cableType {
	case "cat5", "cat5e", "cat6", "cat6a":
		return "mgmt"
	case "dac-passive", "dac-active", "aoc", "mmf-om3", "mmf-om4", "smf":
		return "data"
	case "power":
		return "power"
	default:
		return ""
	}
}

// truncate truncates a string to max length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
