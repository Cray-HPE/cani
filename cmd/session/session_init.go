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
	"path/filepath"
	"sort"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/cmd/config"
	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionInitCmd represents the session init command
var SessionInitCmd = &cobra.Command{
	Use:          "init",
	Short:        "Initialize and start a session. Will perform an import of system's inventory format.",
	Long:         `Initialize and start a session. Will perform an import of system's inventory format.`,
	Args:         validProvider,
	ValidArgs:    taxonomy.SupportedProviders,
	SilenceUsage: true, // Errors are more important than the usage
	RunE:         startSession,
}

// startSession starts a session if one does not exist
func startSession(cmd *cobra.Command, args []string) (err error) {
	// Create a domain object to interact with the datastore and the provider
	root.D, err = domain.New(cmd, args)
	if err != nil {
		return errors.Join(err,
			errors.New("external inventory is unstable"),
			errors.New("unable to get provider specific config options"),
			errors.New("fix issues before starting another session"))
	}

	// Set the paths needed for starting a session
	root.D.CustomHardwareTypesDir = config.CustomDir
	root.D.DatastorePath = filepath.Join(config.ConfigDir, taxonomy.DsFile)
	root.D.LogFilePath = filepath.Join(config.ConfigDir, taxonomy.LogFile)

	// Setup the domain now that the minimum required options are set
	// This allows the provider to define their own logic and keeps it out
	// of the 'cmd' package
	err = root.D.SetupDomain(cmd, args)
	if err != nil {
		return err
	}

	// Create a config object that will be written to the config file
	root.Conf = &config.Config{
		Session: &config.Session{
			Domains: map[string]*domain.Domain{},
		},
	}

	// If a session is already active, there is nothing to do but the user may want to overwrite the existing session
	if root.D.Active {
		log.Info().Msgf("Session is already ACTIVE.")
		ds := root.D.DatastorePath
		// Check if the json file exists
		if _, err := os.Stat(ds); err == nil {
			// If the json file exists, prompt user for overwrite
			overwrite, err := promptForOverwrite(ds)
			if err != nil {
				return err
			}
			if !overwrite {
				// User chose not to overwrite the file
				os.Exit(0)
			} else {
				err = os.Remove(ds)
				if err != nil {
					return err
				}
			}
		}
	}

	// Validate the external inventory before attempting an import
	result, err := root.D.Validate(cmd.Context(), false, ignoreExternalValidation)
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
		return errors.Join(err,
			errors.New("external inventory is unstable"),
			errors.New("fix issues before starting another session"))
	}

	// Import the external inventory
	if err := root.D.Import(cmd.Context()); err != nil {
		return err
	}

	// "Activate" the session
	root.D.Active = true

	// add this provider to the config with the assembled domain object
	root.Conf.Session.Domains[args[0]] = root.D

	// write the config to the file
	err = root.WriteSession(cmd, args)
	if err != nil {
		return err
	}

	log.Info().Msgf("Session is now ACTIVE with provider %s and datastore %s", root.D.Provider, root.D.DatastorePath)
	return nil
}

func promptForOverwrite(path string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("File %s already exists. Keep session active but overwrite the datastore", path),
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
