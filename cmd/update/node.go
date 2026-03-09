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
package update

import (
	"github.com/spf13/cobra"
)

// UpdateNodeCmd represents the node update command
var UpdateNodeCmd = &cobra.Command{
	Use:   "node PROVIDER",
	Short: "Update nodes in the inventory.",
	Long:  `Update nodes in the inventory.`,
	Args:  cobra.ExactArgs(1),
	RunE:  updateNode, // Update a node when this sub-command is called
}

// updateNode updates a node to the inventory
func updateNode(cmd *cobra.Command, args []string) (err error) {
	// // Remove the node from the inventory using domain methods
	// if cmd.Flags().Changed("uuid") {
	// 	// parse the passed in uuid
	// 	u, err := uuid.Parse(nodeUuid)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// get the inventory
	// 	inv, err := root.D.List()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// if the hardware exists, extract the location ordinals for the user
	// 	if n, ok := inv.Hardware[u]; ok {
	// 		cabinet = n.LocationPath.GetOrdinalPath()[1]
	// 		chassis = n.LocationPath.GetOrdinalPath()[2]
	// 		blade = n.LocationPath.GetOrdinalPath()[3]
	// 		nodecard = n.LocationPath.GetOrdinalPath()[4]
	// 		node = n.LocationPath.GetOrdinalPath()[5]
	// 	}
	// }

	// result, err := root.D.UpdateNode(cmd, args, cabinet, chassis, blade, nodecard, node)
	// if errors.Is(err, provider.ErrDataValidationFailure) {
	// 	// TODO the following should probably suggest commands to fix the issue?
	// 	log.Error().Msgf("Inventory data validation errors encountered")
	// 	for id, failedValidation := range result.ProviderValidationErrors {
	// 		log.Error().Msgf("  %s: %s", id, failedValidation.Hardware.LocationPath.String())
	// 		sort.Strings(failedValidation.Errors)
	// 		for _, validationError := range failedValidation.Errors {
	// 			log.Error().Msgf("    - %s", validationError)
	// 		}
	// 	}

	// 	return err
	// } else if err != nil {
	// 	return err
	// }

	// // TODO need a better identify, perhaps its UUID, or its location path?
	// log.Printf("Updated node")
	return nil
}
