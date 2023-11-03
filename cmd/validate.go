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
package cmd

import (
	"errors"
	"sort"

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ValidateCmd represents the validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate assets in the inventory.",
	Long:  `Validate assets in the inventory.`,
	RunE:  validateInventory,
}

func validateInventory(cmd *cobra.Command, args []string) error {
	log.Warn().Msg("This may fail in the HMS Simulator without Network information.")
	if D.Active {

		// Validate the external inventory
		result, err := D.Validate(cmd.Context(), true, false)
		if errors.Is(err, provider.ErrDataValidationFailure) {
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
	} else {
		return errors.New("No active session.  Domain options needed to validate inventory.")
	}
	return nil
}
