/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/
package cmd

import (
	"errors"
	"sort"

	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ValidateCmd represents the validate command
var ValidateCmd = &cobra.Command{
	Use:          "validate",
	Short:        "Validate assets in the inventory.",
	Long:         `Validate assets in the inventory.`,
	SilenceUsage: true, // Errors are more important than the usage
	RunE:         validateInventory,
}

func validateInventory(cmd *cobra.Command, args []string) error {
	log.Warn().Msg("This may fail in the HMS Simulator without Network information.")
	if Conf.Session.Active {
		// Create a domain object to interact with the datastore
		d, err := domain.New(Conf.Session.DomainOptions)
		if err != nil {
			return err
		}
		// Validate the external inventory
		result, err := d.Validate(cmd.Context())
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
	} else {
		return errors.New("No active session.  Domain options needed to validate inventory.")
	}
	return nil
}
