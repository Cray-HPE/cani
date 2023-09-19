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
package blade

import (
	"errors"
	"fmt"
	"sort"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/tui"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// AddBladeCmd represents the blade add command
var AddBladeCmd = &cobra.Command{
	Use:               "blade",
	Short:             "Add blades to the inventory.",
	Long:              `Add blades to the inventory.`,
	PersistentPreRunE: root.DatastoreExists, // A session must be active to write to a datastore
	SilenceUsage:      true,                 // Errors are more important than the usage
	Args:              validHardware,        // Hardware can only be valid if defined in the hardware library
	RunE:              addBlade,             // Add a blade when this sub-command is called
}

// addBlade adds a blade to the inventory
func addBlade(cmd *cobra.Command, args []string) error {
	if auto {
		recommendations, err := root.Domain.Recommend(args[0])
		if err != nil {
			return err
		}
		log.Info().Msgf("Querying inventory to suggest cabinet, chassis, and blade for this %s", hardwaretypes.NodeBlade)
		cabinet = recommendations.CabinetOrdinal
		chassis = recommendations.ChassisOrdinal
		blade = recommendations.BladeOrdinal
		log.Debug().Msgf("Provider recommendations: %+v", recommendations)
		log.Info().Msgf("Suggested %s number: %d", hardwaretypes.Cabinet, cabinet)
		log.Info().Msgf("Suggested %s number: %d", hardwaretypes.Chassis, chassis)
		log.Info().Msgf("Suggested %s number: %d", hardwaretypes.NodeBlade, blade)
		if accept {
			auto = true
		} else {
			// Prompt the user to confirm the suggestions
			auto, err = tui.CustomConfirmation(
				fmt.Sprintf("Would you like to accept the recommendations and add the %s", hardwaretypes.NodeBlade))
			if err != nil {
				return err
			}
		}

		// If the user chose not to accept the suggestions, exit
		if !auto {
			log.Warn().Msgf("Aborted %s add", hardwaretypes.NodeBlade)
			fmt.Printf("\nAuto-generated values can be overridden by re-running the command with explicit values:\n")
			fmt.Printf("\n\tcani alpha add %s %s --cabinet %d --chassis %d --blade %d\n\n", cmd.Name(), args[0], cabinet, chassis, blade)

			return nil
		}
	}

	// Add the blade from the inventory using domain methods
	result, err := root.Domain.AddBlade(cmd.Context(), args[0], cabinet, chassis, blade)
	if errors.Is(err, provider.ErrDataValidationFailure) {
		// TODO this validation error print logic could be shared

		// TODO the following should probably suggest commands to fix the issue?
		log.Error().Msgf("Inventory data validation errors encountered")
		// Merge datastore and provider validation errors
		failedValidations := map[uuid.UUID][]string{}
		locationPathStrings := map[uuid.UUID]string{}
		for _, result := range result.DatastoreValidationErrors {
			failedValidations[result.Hardware.ID] = append(failedValidations[result.Hardware.ID], result.Errors...)

			locationPathStrings[result.Hardware.ID] = result.Hardware.LocationPath.String()
		}
		for _, result := range result.ProviderValidationErrors {
			failedValidations[result.Hardware.ID] = append(failedValidations[result.Hardware.ID], result.Errors...)

			locationPathStrings[result.Hardware.ID] = result.Hardware.LocationPath.String()
		}

		// Provider validation errors
		for id, errorStrings := range failedValidations {
			log.Error().Msgf("  %s: %s", id, locationPathStrings[id])
			sort.Strings(errorStrings)
			for _, validationError := range errorStrings {
				log.Error().Msgf("    - %s", validationError)
			}
		}

		return err
	} else if err != nil {
		return err
	}

	// Use a map to track already added nodes.
	newNodes := []domain.HardwareLocationPair{}

	for _, result := range result.AddedHardware {
		// If the type is a Node
		if result.Hardware.Type == hardwaretypes.NodeBlade {
			log.Debug().Msg(result.Location.String())
			log.Debug().Msgf("This %s also contains a %s (added %s)",
				hardwaretypes.NodeBlade,
				hardwaretypes.Node,
				result.Hardware.ID.String())
			log.Debug().Msgf("This %s also contains a %s (%s) added at %s",
				hardwaretypes.NodeBlade,
				hardwaretypes.Node,
				result.Hardware.ID.String(),
				result.Location)
			// Add the node to the map
			newNodes = append(newNodes, result)
			if root.Conf.Session.DomainOptions.Provider == string(inventory.CSMProvider) {
				log.Info().Str("status", "SUCCESS").Msgf("%s was successfully staged to be added to the system", hardwaretypes.NodeBlade)
				log.Info().Msgf("UUID: %s", result.Hardware.ID)
				log.Info().Msgf("Cabinet: %d", cabinet)
				log.Info().Msgf("Chassis: %d", chassis)
				log.Info().Msgf("Blade: %d", blade)
			}
		}
	}

	return nil
}
