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
package cabinet

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ListCabinetCmd represents the cabinet list command
var ListCabinetCmd = &cobra.Command{
	Use:   "cabinet",
	Short: "List cabinets in the inventory.",
	Long:  `List cabinets in the inventory.`,
	Args:  cobra.ArbitraryArgs,
	RunE:  listCabinet,
}

// listCabinet lists cabinets in the inventory
func listCabinet(cmd *cobra.Command, args []string) error {
	// Get the entire inventory
	inv, err := root.D.List()
	if err != nil {
		return err
	}
	// Filter the inventory to only cabinets
	filtered := make(map[uuid.UUID]inventory.Hardware, 0)
	for key, hw := range inv.Hardware {
		if hw.Type == hardwaretypes.Cabinet {
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
		// 		tpl := `{{ printf "%.25s" CABINET }}
		//{{- range . }}
		// |{{ .ID }} | {{ .DeviceTypeSlug }} | {{ .DeviceTypeSlug }}  | {{ .LocationPath }} | {{ end }}
		// 		`

		minwidth := 0         // minimal cell width including any padding
		tabwidth := 8         // width of tab characters (equivalent number of spaces)
		padding := 1          // padding added to a cell before computing its width
		padchar := byte('\t') // ASCII char used for padding

		w := tabwriter.NewWriter(os.Stdout, minwidth, tabwidth, padding, padchar, tabwriter.AlignRight)
		defer w.Flush()

		// set the CANI columns
		caniColumns := []string{
			"UUID",
			"Status",
			"Type",
			"Location",
		}
		// Get columns set by the provider
		providerColumns := root.D.ListCabinetMetadataColumns()

		// combine CANI and provider columns
		columns := []string{}
		for _, col := range [][]string{caniColumns, providerColumns} {
			columns = append(columns, col...)
		}

		fmt.Fprint(
			w,
			fmt.Sprintf("%v%s", strings.Join(columns, "\t"), "\n"),
		)

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

		for _, u := range keys {
			hw, exists := filtered[u]
			if !exists {
				return err
			}
			// get the provider-specific fields
			providerValues, err := root.D.ListCabinetMetadataRow(hw)
			if err != nil {
				return err
			}

			// Set the fields CANI uses
			fields := []string{"%s", "%s", "%s"}
			// append any provider-specified ones, using a %+v to display them to avoid any typing issues at the cost of something ugly printing
			for _, n := range providerColumns {
				log.Debug().Msgf("Using provider-defined column: %+v", n)
				fields = append(fields, "%+v")
			}
			// print the table with CANI and provider columns/rows
			fmt.Fprint(
				w,
				fmt.Sprintf(strings.Join(fields, "\t"),
					filtered[u].ID.String(),
					filtered[u].Status,
					filtered[u].DeviceTypeSlug,
					filtered[u].LocationPath.String()),
				"\t",
				fmt.Sprintf(strings.Join(providerValues, "\t")),
				"\n",
			)
		}

	}
	return nil
}
