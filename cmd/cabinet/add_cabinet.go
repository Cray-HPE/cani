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
	"errors"
	"fmt"
	"sort"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/provider/csm"
	"github.com/Cray-HPE/cani/internal/tui"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// AddCabinetCmd represents the cabinet add command
var AddCabinetCmd = &cobra.Command{
	Use:               "cabinet",
	Short:             "Add cabinets to the inventory.",
	Long:              `Add cabinets to the inventory.`,
	PersistentPreRunE: validFlagCombos, // Also ensures a session is active
	Args:              validHardware,   // Hardware can only be valid if defined in the hardware library
	SilenceUsage:      true,            // Errors are more important than the usage
	RunE:              addCabinet,      // Add a cabinet when this sub-command is called
}

// addCabinet adds a cabinet to the inventory
func addCabinet(cmd *cobra.Command, args []string) error {
	// Create a domain object to interact with the datastore
	d, err := domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	if auto {
		recommendations, err := d.Recommend(args[0])
		if err != nil {
			return err
		}
		log.Info().Msgf("Querying inventory to suggest cabinet number and VLAN ID")
		// set the vars to the recommendations
		cabinetNumber = recommendations.LocationOrdinal
		vlanId = recommendations.ProviderMetadata[csm.ProviderPropertyVlanId].(int)
		log.Debug().Msgf("Provider recommendations: %+v", recommendations)
		log.Info().Msgf("Suggested cabinet number: %d", cabinetNumber)
		log.Info().Msgf("Suggested VLAN ID: %d", vlanId)
		// Prompt the user to confirm the suggestions
		auto, err = tui.CustomConfirmation(
			fmt.Sprintf("Would you like to accept the recommendations and add the %s", hardwaretypes.Cabinet))
		if err != nil {
			return err
		}

		// If the user chose not to accept the suggestions, exit
		if !auto {
			log.Warn().Msgf("Aborted %s add", hardwaretypes.Cabinet)
			fmt.Printf("\nAuto-generated values can be overridden by re-running the command with explicit values:\n")
			fmt.Printf("\n\tcani alpha add %s %s --vlan-id %d --cabinet %d\n\n", cmd.Name(), args[0], vlanId, cabinetNumber)

			return nil
		}
	}

	// Push all the CLI flags that were provided into a generic map
	// TODO Need to figure out how to specify to unset something
	// Right now the build metadata function in the CSM provider will
	// unset options if nil is passed in.
	cabinetMetadata := map[string]interface{}{
		csm.ProviderPropertyVlanId: vlanId,
	}

	// Add the cabinet to the inventory using domain methods
	result, err := d.AddCabinet(cmd.Context(), args[0], cabinetNumber, cabinetMetadata)
	if errors.Is(err, provider.ErrDataValidationFailure) {
		// TODO the following should probably suggest commands to fix the issue?
		log.Error().Msgf("Inventory data validation errors encountered")
		for id, failedValidation := range result.ProviderValidationErrors {
			log.Error().Msgf("  %s: %s", id, failedValidation.Hardware.LocationPath.String())
			sort.Strings(failedValidation.Errors)
			for _, validationError := range failedValidation.Errors {
				log.Error().Msgf("    - %s", validationError)
			}
		}

		return err
	} else if err != nil {
		return err
	}

	log.Info().Str("status", "SUCCESS").Msgf("%s %d was successfully staged to be added to the system", hardwaretypes.Cabinet, cabinetNumber)

	// Use a map to track already added nodes.
	newNodes := []domain.HardwareLocationPair{}

	for _, result := range result.AddedHardware {
		// If the type is a Node
		if result.Hardware.Type == hardwaretypes.Cabinet {
			log.Debug().Msgf("%s added at %s with parent %s (%s)", result.Hardware.Type, result.Location.String(), hardwaretypes.System, result.Hardware.Parent)
			log.Info().Msgf("UUID: %s", result.Hardware.ID)
			log.Info().Msgf("Cabinet Number: %d", cabinetNumber)
			log.Info().Msgf("VLAN ID: %d", vlanId)
			// Add the node to the map
			newNodes = append(newNodes, result)
		}
	}

	return nil
}
