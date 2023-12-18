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
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/tui"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// AddCabinetCmd represents the cabinet add command
var AddCabinetCmd = &cobra.Command{
	Use:     "cabinet",
	Short:   "Add cabinets to the inventory.",
	Long:    `Add cabinets to the inventory.`,
	PreRunE: validHardware, // Hardware can only be valid if defined in the hardware library
	RunE:    addCabinet,    // Add a cabinet when this sub-command is called
}

// addCabinet adds a cabinet to the inventory
func addCabinet(cmd *cobra.Command, args []string) (err error) {
	var recommendations = provider.HardwareRecommendations{}
	recommendations, err = root.D.Recommend(cmd, args, auto)
	if err != nil {
		return err
	}

	if auto {
		// get hardware recommendations from the provider
		log.Info().Msgf("Querying inventory to suggest %s", hardwaretypes.Cabinet)
		// Prompt the user to confirm the suggestions
		if !accept {
			accept, err = tui.CustomConfirmation(fmt.Sprintf("Would you like to accept the recommendations and add the %s", hardwaretypes.Cabinet))
			if err != nil {
				return err
			}
		}

		// If the user chose not to accept the suggestions, exit
		if !accept {
			log.Warn().Msgf("Aborted %s add", hardwaretypes.Cabinet)
			return nil
		}

		// log the provider recommendations to the screen
		recommendations.Print()
	}

	// Add the cabinet to the inventory using domain methods
	result, err := root.D.AddCabinet(cmd, args, recommendations)
	if errors.Is(err, provider.ErrDataValidationFailure) {
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

	var filtered = make(map[uuid.UUID]inventory.Hardware, 0)
	for _, result := range result.AddedHardware {
		if result.Hardware.Type == hardwaretypes.Cabinet {
			log.Debug().Msgf("%s added at %s with parent %s (%s)", result.Hardware.Type, result.Location.String(), hardwaretypes.System, result.Hardware.Parent)
			log.Info().Str("status", "SUCCESS").Msgf("%s %d was successfully staged to be added to the system", hardwaretypes.Cabinet, recommendations.CabinetOrdinal)
			filtered[result.Hardware.ID] = result.Hardware
		}
	}

	root.D.PrintHardware(cmd, args, filtered)
	return nil
}
