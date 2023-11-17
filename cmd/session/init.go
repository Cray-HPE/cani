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
	"os"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/provider/csm"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	dryrun                   bool
	commit                   bool
	ignoreExternalValidation bool
	ignoreValidationMessage  = "Ignore validation failures. Use this to allow unconventional configurations."
	csmInitCmd               = &cobra.Command{}

	ProviderCmd = &cobra.Command{}
	// BootstapCmd is used to start a session with a specific provider and allows the provider to define
	// how the real init command is defined using their custom business logic
	BootstrapCmd = &cobra.Command{
		Use:       "init PROVIDER",
		Short:     taxonomy.InitShort,
		Long:      taxonomy.InitLong,
		ValidArgs: taxonomy.SupportedProviders, // supported providers are defined in the taxonomy
		Args:      validProvider,               // validate the arg with more contextual help dialogs
		RunE:      initSessionWithProviderCmd,
	}
)

func init() {
	// init is run once, and this is where the flags get set
	// since flags vary by provider, create a variable for each
	var err error
	for _, provider := range taxonomy.SupportedProviders {
		switch provider {
		case taxonomy.CSM:
			csmInitCmd, err = csm.NewSessionInitCommand()
			csmInitCmd.Use = "init"
			ProviderCmd = csmInitCmd
		default:
			log.Debug().Msgf("skipping provider: %s", provider)
		}
		if err != nil {
			log.Error().Msgf("unable to get cmd from provider: %v", err)
			os.Exit(1)
		}
	}

	// Define the bare minimum needed to determine who the provider for the session will be
	BootstrapCmd.Flags().BoolVar(&ignoreExternalValidation, "ignore-validation", false, ignoreValidationMessage)

	// all flags should be set in init().  you can set flags after the fact, but it is much easier to work with everything up front
	// this will set existing variables for each provider
	err = mergeProviderFlags(BootstrapCmd, ProviderCmd)
	if err != nil {
		log.Error().Msgf("unable to get flags from provider: %v", err)
		os.Exit(1)
	}

	// Add session commands to root commands
	root.SessionCmd.AddCommand(BootstrapCmd)
	BootstrapCmd.AddCommand(csmInitCmd)
	root.SessionCmd.AddCommand(SessionApplyCmd)
	root.SessionCmd.AddCommand(SessionStatusCmd)
	root.SessionCmd.AddCommand(SessionSummaryCmd)

	// Session stop flags
	SessionApplyCmd.Flags().BoolVarP(&commit, "commit", "c", false, "Commit changes to session")
	SessionApplyCmd.Flags().BoolVarP(&dryrun, "dryrun", "d", false, "Perform dryrun, and do not make changes to the system")
	SessionApplyCmd.Flags().BoolVar(&ignoreExternalValidation, "ignore-validation", false, ignoreValidationMessage)
}
