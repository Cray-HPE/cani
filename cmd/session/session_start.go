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
	"strings"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/cmd/config"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/provider/csm"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionStartCmd represents the session start command
var SessionStartCmd = &cobra.Command{
	Use:                "start",
	Short:              "Start a session.",
	Long:               `Start a session.`,
	Args:               validProvider,
	ValidArgs:          validArgs,
	SilenceUsage:       true, // Errors are more important than the usage
	RunE:               startSession,
	PersistentPostRunE: writeSession,
}

var (
	providerName string
	validArgs    = []string{"csm"}
)

// startSession starts a session if one does not exist
func startSession(cmd *cobra.Command, args []string) error {
	// TODO This is probably not the right way todo this, but hopefully this will be easy way...
	// Sorry Jacob
	if useSimulation {
		log.Warn().Msg("Using simulation mode")
		root.Conf.Session.DomainOptions.CsmOptions.UseSimulation = true
	} else {
		root.Conf.Session.DomainOptions.CsmOptions.BaseUrlSLS, _ = cmd.Flags().GetString("csm-url-sls")
		root.Conf.Session.DomainOptions.CsmOptions.BaseUrlHSM, _ = cmd.Flags().GetString("csm-url-hsm")
		root.Conf.Session.DomainOptions.CsmOptions.InsecureSkipVerify, _ = cmd.Flags().GetBool("csm-insecure-https")
	}
	if insecure {
		root.Conf.Session.DomainOptions.CsmOptions.InsecureSkipVerify = true
	}
	root.Conf.Session.DomainOptions.CsmOptions.SecretName = secretName
	root.Conf.Session.DomainOptions.CsmOptions.KubeConfig = kubeconfig
	root.Conf.Session.DomainOptions.CsmOptions.CaCertPath = caCertPath
	root.Conf.Session.DomainOptions.CsmOptions.ClientID = clientId
	root.Conf.Session.DomainOptions.CsmOptions.ClientSecret = clientSecret
	root.Conf.Session.DomainOptions.CsmOptions.TokenHost = strings.TrimRight(tokenUrl, "/") // Remove trailing slash if present
	root.Conf.Session.DomainOptions.CsmOptions.TokenUsername = tokenUsername
	root.Conf.Session.DomainOptions.CsmOptions.TokenPassword = tokenPassword

	// If a session is already active, there is nothing to do but the user may want to overwrite the existing session
	if root.Conf.Session.Active {
		log.Info().Msgf("Session is already ACTIVE.")
		ds := root.Conf.Session.DomainOptions.DatastorePath
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
			}
		}
	}

	// Create a domain object to interact with the datastore
	var err error
	root.Conf.Session.Domain, err = domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	// Perform provider plugin specific logic at session start
	switch root.Conf.Session.DomainOptions.Provider {
	case string(inventory.CSMProvider):
		// Need to get the systems Roles/SubRole data from the system
		// TODO CASMINST-6417

		// For now just use the defaults
		root.Conf.Session.DomainOptions.CsmOptions.ValidRoles = csm.DefaultValidRoles
		root.Conf.Session.DomainOptions.CsmOptions.ValidSubRoles = csm.DefaultValidSubRolesRoles
	}

	// Validate the external inventory
	result, err := root.Conf.Session.Domain.Validate(cmd.Context(), false)
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
			errors.New("External inventory is unstable.  Fix, and check with 'cani validate' before continuing."))
	}

	// "Activate" the session
	root.Conf.Session.Active = true

	ds := root.Conf.Session.DomainOptions.DatastorePath
	provider := root.Conf.Session.DomainOptions.Provider
	log.Info().Msgf("Session is now ACTIVE with provider %s and datastore %s", provider, ds)
	return nil
}

// writeSession writes the session configuration back to the config file
func writeSession(cmd *cobra.Command, args []string) error {
	// Write the configuration back to the file
	cfgFile := cmd.Root().PersistentFlags().Lookup("config").Value.String()
	err := config.WriteConfig(cfgFile, root.Conf)
	if err != nil {
		return err
	}
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
