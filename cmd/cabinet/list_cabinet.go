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
	"text/tabwriter"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
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
		tabwidth := 4         // width of tab characters (equivalent number of spaces)
		padding := 1          // padding added to a cell before computing its width
		padchar := byte('\t') // ASCII char used for padding

		w := tabwriter.NewWriter(os.Stdout, minwidth, tabwidth, padding, padchar, tabwriter.AlignRight)
		defer w.Flush()
		// t := template.New("cabinets")
		// t, _ = t.Parse(tpl)
		// if err := t.Execute(w, filtered); err != nil {
		// 	return err
		// }
		// fmt.Fprintln(w, "Cabinet\tSlug\tHMN VLAN\tLocation")

		var hmnVlan float64
		fmt.Fprintf(w, "%s\t%s\t%v\t%s\n",
			"UUID",
			"Type",
			"HMN VLAN",
			"Location")

		for _, hw := range filtered {
			if _, exists := hw.ProviderProperties[string(inventory.CSMProvider)]; exists {
				for k, v := range hw.ProviderProperties[string(inventory.CSMProvider)].(map[string]interface{}) {
					if _, e := hw.ProviderProperties[string(inventory.CSMProvider)].(map[string]interface{})[k]; e {
						hmnVlan = v.(float64)
					}

				}
			}
			fmt.Fprintf(w, "%s\t%s\t%v\t%s\n",
				hw.ID.String(),
				hw.DeviceTypeSlug,
				hmnVlan,
				hw.LocationPath.String())
		}

	}
	return nil
}
