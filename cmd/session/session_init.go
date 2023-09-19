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
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionInitCmd represents the session init command
var SessionInitCmd = &cobra.Command{
	Use:                "init",
	Short:              "Initialize and start a session. Will perform an import of system's inventory format.",
	Long:               `Initialize and start a session. Will perform an import of system's inventory format.`,
	Args:               validProvider,
	ValidArgs:          validArgs,
	SilenceUsage:       true, // Errors are more important than the usage
	RunE:               startSession,
	PersistentPostRunE: writeSession,
}

var (
	validArgs = []string{"csm"}
)

// startSession starts a session if one does not exist
func startSession(cmd *cobra.Command, args []string) (err error) {
	if useSimulation {
		log.Warn().Msg("Using simulation mode")
		root.Conf.Session.DomainOptions.CsmOptions.UseSimulation = true
	} else {
		slsUrl, _ := cmd.Flags().GetString("csm-url-sls")
		if slsUrl != "" {
			root.Conf.Session.DomainOptions.CsmOptions.BaseUrlSLS = slsUrl
		} else {
			root.Conf.Session.DomainOptions.CsmOptions.BaseUrlSLS = fmt.Sprintf("https://%s/apis/sls/v1", providerHost)
		}
		hsmUrl, _ := cmd.Flags().GetString("csm-url-hsm")
		if hsmUrl != "" {
			root.Conf.Session.DomainOptions.CsmOptions.BaseUrlHSM = hsmUrl
		} else {
			root.Conf.Session.DomainOptions.CsmOptions.BaseUrlHSM = fmt.Sprintf("https://%s/apis/smd/hsm/v2", providerHost)
		}
		root.Conf.Session.DomainOptions.CsmOptions.InsecureSkipVerify, _ = cmd.Flags().GetBool("csm-insecure-https")
	}
	if insecure {
		root.Conf.Session.DomainOptions.CsmOptions.InsecureSkipVerify = true
	}
	root.Conf.Session.DomainOptions.CsmOptions.SecretName = secretName
	root.Conf.Session.DomainOptions.CsmOptions.K8sPodsCidr = k8sPodsCidr
	root.Conf.Session.DomainOptions.CsmOptions.K8sServicesCidr = k8sServicesCidr
	root.Conf.Session.DomainOptions.CsmOptions.KubeConfig = kubeconfig
	root.Conf.Session.DomainOptions.CsmOptions.CaCertPath = caCertPath
	root.Conf.Session.DomainOptions.CsmOptions.ClientID = clientId
	root.Conf.Session.DomainOptions.CsmOptions.ClientSecret = clientSecret
	root.Conf.Session.DomainOptions.CsmOptions.ProviderHost = strings.TrimRight(providerHost, "/") // Remove trailing slash if present
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
	root.Conf.Session.DomainOptions.Provider = args[0]
	err = root.Domain.SetConfigOptions(cmd.Context(), root.Conf.Session.DomainOptions)
	if err != nil {
		return errors.Join(err,
			errors.New("External inventory is unstable. Unable to get provider specific config options. Fix issues before starting another session."))
	}

	// Validate the external inventory
	result, err := root.Domain.Validate(cmd.Context(), false, ignoreExternalValidation)
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
			errors.New("External inventory is unstable.  Fix issues before starting another session."))
	}

	// Commit the external inventory
	if err := root.Domain.Import(cmd.Context()); err != nil {
		return err
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
