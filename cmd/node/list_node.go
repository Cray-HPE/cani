/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/csm"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// ListNodeCmd represents the node list command
var ListNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "List nodes in the inventory.",
	Long:  `List nodes in the inventory.`,
	Args:  cobra.ArbitraryArgs,
	RunE:  listNode,
}

// listNode lists nodes in the inventory
func listNode(cmd *cobra.Command, args []string) error {
	// Create a domain object to interact with the datastore
	d, err := domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	// Get the entire inventory
	inv, err := d.List()
	if err != nil {
		return err
	}
	// Filter the inventory to only nodes
	filtered := make(map[uuid.UUID]inventory.Hardware, 0)
	for key, hw := range inv.Hardware {
		if hw.Type == hardwaretypes.Node {
			filtered[key] = hw
		}
	}

	switch format {
	case "json":
		// Convert the filtered inventory into a formatted JSON string
		inventoryJSON, err := json.MarshalIndent(filtered, "", "  ")
		if err != nil {
			return errors.New(fmt.Sprintf("Error marshaling inventory to JSON: %v", err))
		}
		// Print the inventory
		fmt.Println(string(inventoryJSON))

	case "pretty":

		minwidth := 0         // minimal cell width including any padding
		tabwidth := 8         // width of tab characters (equivalent number of spaces)
		padding := 1          // padding added to a cell before computing its width
		padchar := byte('\t') // ASCII char used for padding

		w := tabwriter.NewWriter(os.Stdout, minwidth, tabwidth, padding, padchar, tabwriter.AlignRight)
		defer w.Flush()

		fmt.Fprintf(w, "%s\t%s\t%v\t%s\t%s\t%s\t%s\n",
			"UUID",
			"Type",
			"Role",
			"SubRole",
			"Alias",
			"NID",
			"Location")

		// make keys slice to sort by values in the map
		keys := make([]uuid.UUID, 0, len(filtered))
		for key := range filtered {
			keys = append(keys, key)
		}

		// sort by what the user wants
		sort.Slice(keys, func(i, j int) bool {
			switch sortBy {
			case "location":
				return filtered[keys[i]].LocationPath.String() < filtered[keys[j]].LocationPath.String()

			case "type":
				return string(filtered[keys[i]].DeviceTypeSlug) < string(filtered[keys[j]].DeviceTypeSlug)

			case "uuid":
				return filtered[keys[i]].ID.String() < filtered[keys[j]].ID.String()

			}

			// default is sorted by loc
			return filtered[keys[i]].LocationPath.String() < filtered[keys[j]].LocationPath.String()
		})

		for _, hw := range keys {
			// Start with an empty Node metadata struct, just in case if this node doesn't have any
			// metadata set
			var nodeMetadata csm.NodeMetadata

			// If metadata exists decode it
			if _, exists := filtered[hw].ProviderMetadata[inventory.CSMProvider]; exists {
				csmMetadata, err := csm.DecodeProviderMetadata(filtered[hw])
				if err != nil {
					return err
				}

				if csmMetadata.Node != nil {
					nodeMetadata = *csmMetadata.Node
				}
			}

			// convert properties to strings and set nil values for easy printing
			pp := nodeMetadata.Pretty()

			fmt.Fprintf(w, "%s\t%s\t%v\t%v\t%v\t%v\t%v\n",
				filtered[hw].ID.String(),
				filtered[hw].DeviceTypeSlug,
				pp.Role,
				pp.SubRole,
				pp.Alias,
				pp.Nid,
				filtered[hw].LocationPath.String())
		}

	}
	return nil
}
