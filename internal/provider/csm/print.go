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
package csm

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// PrintRecommendations implements the InventoryProvider interface
// it prints the hardware to stdout based on the command
func (csm *CSM) PrintRecommendations(cmd *cobra.Command, args []string, recommendations provider.HardwareRecommendations) error {
	log.Info().Msgf("Suggested cabinet number: %d", recommendations.CabinetOrdinal)
	log.Info().Msgf("Suggested VLAN ID: %d", recommendations.ProviderMetadata["HMNVlan"])

	return nil
}

// PrintHardware implements the InventoryProvider interface
// it prints the hardware to stdout based on the command
func (csm *CSM) PrintHardware(cmd *cobra.Command, args []string, filtered map[uuid.UUID]inventory.Hardware) (err error) {
	switch cmd.Parent().Parent().Name() {
	case "add":
		err = csm.printHardwareForAddCommand(cmd, args, filtered)
	case "list":
		err = csm.printHardwareForListCommand(cmd, args, filtered)
	case "update":
		err = csm.printHardwareForUpdateCommand(cmd, args, filtered)
	case "remove":
		err = csm.printHardwareForRemoveCommand(cmd, args, filtered)
	default:
		log.Warn().Msgf("No print function for command '%s %s %s'", cmd.Name(), cmd.Parent().Name(), cmd.Parent().Parent().Name())
	}
	if err != nil {
		return err
	}

	return nil
}

// printHardwareForAddCommand prints the hardware for the add command
// based on the type of hardware
func (csm *CSM) printHardwareForAddCommand(cmd *cobra.Command, args []string, filtered map[uuid.UUID]inventory.Hardware) (err error) {
	for _, hw := range filtered {
		switch hw.Type {
		case hardwaretypes.Cabinet:
			err = printForAddCabinetCommand(cmd, args, hw)
		// case hardwaretypes.Chassis:
		// 	err = printForAddChassisCommand(cmd, args, hw)
		case hardwaretypes.NodeBlade:
			err = printForAddBladeCommand(cmd, args, hw)
		case hardwaretypes.Node:
			err = printForAddNodeCommand(cmd, args, hw)
		}
	}
	if err != nil {
		return err
	}
	return nil
}

// printHardwareForListCommand prints the hardware for the list command
func (csm *CSM) printHardwareForListCommand(cmd *cobra.Command, args []string, filtered map[uuid.UUID]inventory.Hardware) (err error) {
	err = printForListCommand(cmd, args, filtered)
	if err != nil {
		return err
	}
	return nil
}

// printHardwareForUpdateCommand
func (csm *CSM) printHardwareForUpdateCommand(cmd *cobra.Command, args []string, filtered map[uuid.UUID]inventory.Hardware) (err error) {

	return nil
}

// printHardwareForRemoveCommand
func (csm *CSM) printHardwareForRemoveCommand(cmd *cobra.Command, args []string, filtered map[uuid.UUID]inventory.Hardware) (err error) {

	return nil
}

// printForAddCabinetCommand prints the hardware for the add cabinet command
func printForAddCabinetCommand(cmd *cobra.Command, args []string, hw inventory.Hardware) error {
	log.Info().Str("status", "SUCCESS").Msgf("%s was successfully %s to be added to the system", hardwaretypes.Cabinet, hw.Status)
	log.Info().Msgf("UUID: %s", hw.ID)
	log.Info().Msgf("Cabinet Number: %d", *hw.LocationOrdinal)
	log.Info().Msgf("VLAN ID: %d", hw.ProviderMetadata["csm"]["Cabinet"].(map[string]interface{})["HMNVlan"])

	return nil
}

// printForAddBladeCommand prints the hardware for the add blade command
func printForAddBladeCommand(cmd *cobra.Command, args []string, hw inventory.Hardware) error {
	log.Info().Msgf("UUID: %s", hw.ID)

	cabinet, _ := hw.LocationPath.GetOrdinal(hardwaretypes.Cabinet)
	log.Info().Msgf("Cabinet: %d", cabinet)

	chassis, _ := hw.LocationPath.GetOrdinal(hardwaretypes.Chassis)
	log.Info().Msgf("Chassis: %d", chassis)

	blade, _ := hw.LocationPath.GetOrdinal(hardwaretypes.NodeBlade)
	log.Info().Msgf("Blade: %d", blade)

	return nil
}

// printForAddNodeCommand prints the hardware for the add node command
func printForAddNodeCommand(cmd *cobra.Command, args []string, hw inventory.Hardware) error {
	log.Info().Msgf("UUID: %s", hw.ID)

	cabinet, _ := hw.LocationPath.GetOrdinal(hardwaretypes.Cabinet)
	log.Info().Msgf("Cabinet: %d", cabinet)

	chassis, _ := hw.LocationPath.GetOrdinal(hardwaretypes.Chassis)
	log.Info().Msgf("Chassis: %d", chassis)

	blade, _ := hw.LocationPath.GetOrdinal(hardwaretypes.NodeBlade)
	log.Info().Msgf("Blade: %d", blade)

	node, _ := hw.LocationPath.GetOrdinal(hardwaretypes.Node)
	log.Info().Msgf("Node: %d", node)

	return nil
}

// printForListCommand prints the hardware for the list command
func printForListCommand(cmd *cobra.Command, args []string, filtered map[uuid.UUID]inventory.Hardware) (err error) {
	format, _ := cmd.Flags().GetString("format")

	switch format {
	case "json":
		log.Info().Msgf("list print json")
		// Convert the filtered inventory into a formatted JSON string
		inventoryJSON, err := json.MarshalIndent(filtered, "", "  ")
		if err != nil {
			return errors.New(fmt.Sprintf("Error marshaling inventory to JSON: %v", err))
		}
		// Print the inventory
		fmt.Println(string(inventoryJSON))

	case "pretty":
		err = prettyPrintForListCommand(cmd, args, filtered)
	}
	if err != nil {
		return err
	}

	return nil
}

// prettyPrintForListCommand prints the hardware for the list command
// in a pretty, human-readable format and based on the type of hardware
func prettyPrintForListCommand(cmd *cobra.Command, args []string, filtered map[uuid.UUID]inventory.Hardware) (err error) {
	switch cmd.Parent().Name() {
	case "cabinet":
		err = prettyPrintForListCabinet(cmd, args, filtered)
	case "chassis":
		err = prettyPrintForListChassis(cmd, args, filtered)
	case "blade":
		err = prettyPrintForListBlade(cmd, args, filtered)
	case "node":
		err = prettyPrintForListNode(cmd, args, filtered)
	default:
	}
	if err != nil {
		return err
	}

	return nil
}

// prettyPrintForListCabinet pretty prints cabinets for the list command
func prettyPrintForListCabinet(cmd *cobra.Command, args []string, filtered map[uuid.UUID]inventory.Hardware) (err error) {
	format, _ := cmd.Flags().GetString("format")
	sortBy, _ := cmd.Flags().GetString("sort")

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
		providerColumns := listCabinetMetadataColumns()

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
			providerValues, err := listCabinetMetadataRow(hw)
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

// prettyPrintForListChassis pretty prints chassis for the list command
func prettyPrintForListChassis(cmd *cobra.Command, args []string, filtered map[uuid.UUID]inventory.Hardware) (err error) {
	log.Warn().Msgf("not yet implemented")
	return nil
}

// prettyPrintForListBlade pretty prints blades for the list command
func prettyPrintForListBlade(cmd *cobra.Command, args []string, filtered map[uuid.UUID]inventory.Hardware) (err error) {
	sortBy, _ := cmd.Flags().GetString("sort")
	minwidth := 0         // minimal cell width including any padding
	tabwidth := 8         // width of tab characters (equivalent number of spaces)
	padding := 1          // padding added to a cell before computing its width
	padchar := byte('\t') // ASCII char used for padding

	w := tabwriter.NewWriter(os.Stdout, minwidth, tabwidth, padding, padchar, tabwriter.AlignRight)
	defer w.Flush()

	fmt.Fprintf(w, "%s\t%s\t%s\t%v\n",
		"UUID",
		"Status",
		"Type",
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
		fmt.Fprintf(w, "%s\t%s\t%s\t%v\n",
			filtered[hw].ID.String(),
			filtered[hw].Status,
			filtered[hw].DeviceTypeSlug,
			filtered[hw].LocationPath.String())
	}
	return nil
}

// prettyPrintForListNode pretty prints nodes for the list command
func prettyPrintForListNode(cmd *cobra.Command, args []string, filtered map[uuid.UUID]inventory.Hardware) (err error) {
	// format, _ := cmd.Flags().GetString("format")
	sortBy, _ := cmd.Flags().GetString("sort")

	minwidth := 0         // minimal cell width including any padding
	tabwidth := 8         // width of tab characters (equivalent number of spaces)
	padding := 1          // padding added to a cell before computing its width
	padchar := byte('\t') // ASCII char used for padding

	w := tabwriter.NewWriter(os.Stdout, minwidth, tabwidth, padding, padchar, tabwriter.AlignRight)
	defer w.Flush()

	fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%s\t%s\t%s\t%s\n",
		"UUID",
		"Status",
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
		var nodeMetadata NodeMetadata

		// If metadata exists decode it
		if _, exists := filtered[hw].ProviderMetadata[inventory.CSMProvider]; exists {
			csmMetadata, err := DecodeProviderMetadata(filtered[hw])
			if err != nil {
				return err
			}

			if csmMetadata.Node != nil {
				nodeMetadata = *csmMetadata.Node
			}
		}

		// convert properties to strings and set nil values for easy printing
		pp := nodeMetadata.Pretty()

		fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%v\t%v\t%v\t%v\n",
			filtered[hw].ID.String(),
			filtered[hw].Status,
			filtered[hw].DeviceTypeSlug,
			pp.Role,
			pp.SubRole,
			pp.Alias,
			pp.Nid,
			filtered[hw].LocationPath.String())
	}
	return nil
}
