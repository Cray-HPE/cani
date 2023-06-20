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
package session

import (
	"errors"
	"fmt"
	"os"
	"sort"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionStopCmd represents the session stop command
var SessionStopCmd = &cobra.Command{
	Use:                "stop",
	Short:              "Stop a session.",
	Long:               `Stop a session.`,
	SilenceUsage:       true, // Errors are more important than the usage
	RunE:               stopSession,
	PersistentPostRunE: writeSession,
}

// stopSession stops a session if one exists
func stopSession(cmd *cobra.Command, args []string) error {
	ds := root.Conf.Session.DomainOptions.DatastorePath
	providerName := root.Conf.Session.DomainOptions.Provider
	d, err := domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	if root.Conf.Session.Active {
		// Check that the datastore exists before proceeding since we cannot continue without it
		_, err := os.Stat(ds)
		if err != nil {
			return fmt.Errorf("Session is STOPPED with provider '%s' but datastore '%s' does not exist", providerName, ds)
		}
		log.Info().Msgf("Session is STOPPED")
	} else {
		log.Info().Msgf("Session with provider '%s' and datastore '%s' is already STOPPED", providerName, ds)
	}

	if !commit {
		// Prompt user to commit changes if the commit flag is not set
		commit, err = promptForCommit(ds)
		if err != nil {
			return err
		}
	}
	if commit {
		log.Info().Msgf("Committing changes to session")

		// Commit the external inventory
		result, err := d.Commit(cmd.Context())
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
	}

	// "Deactivate" the session if the function has made it this far
	root.Conf.Session.Active = false

	if err := SessionSummaryCmd.RunE(cmd, []string{}); err != nil {
		return err
	}

	return nil
}

func promptForCommit(path string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Would you like to reconcile and commit %s", path),
		IsConfirm: true,
	}

	_, err := prompt.Run()

	if err != nil {
		if err == promptui.ErrAbort {
			// User chose not to overwrite the file
			return false, nil
		}
		// An error occurred
		return false, err
	}

	// User chose to overwrite the file
	return true, nil
}
