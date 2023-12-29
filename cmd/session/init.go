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
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	dryrun                   bool
	commit                   bool
	ignoreExternalValidation bool
	ignoreValidationMessage  = "Ignore validation failures. Use this to allow unconventional configurations."
	forceInit                bool

	ProviderInitCmds = map[string]*cobra.Command{}
	// BootstapCmd is used to start a session with a specific provider and allows the provider to define
	// how the real init command is defined using their custom business logic
	SessionInitCmd = &cobra.Command{
		Use:       "init PROVIDER",
		Short:     taxonomy.InitShort,
		Long:      taxonomy.InitLong,
		ValidArgs: taxonomy.SupportedProviders, // supported providers are defined in the taxonomy
		Args:      validProvider,               // validate the arg with more contextual help dialogs
		RunE:      initSessionWithProviderCmd,
	}
)

func init() {
	// Define the bare minimum needed to determine who the provider for the session will be
	SessionInitCmd.Flags().BoolVar(&ignoreExternalValidation, "ignore-validation", false, ignoreValidationMessage)
	SessionInitCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Overwrite the existing session with a new session")
	SessionInitCmd.Flags().BoolP("insecure", "k", false, "Allow insecure connections when using HTTPS")
	SessionInitCmd.Flags().BoolP("use-simulator", "S", false, "Use simulation environtment settings")

	for _, p := range domain.GetProviders() {
		// Create a provider "init" command
		providerCmd, err := domain.NewSessionInitCommand(p.Slug())
		if err != nil {
			log.Error().Msgf("unable to get provider init command: %v", err)
			os.Exit(1)
		}

		// Merge cani's default flags into the provider command
		err = root.MergeProviderFlags(providerCmd, SessionInitCmd)
		if err != nil {
			log.Error().Msgf("unable to get flags from provider: %v", err)
			os.Exit(1)
		}

		// set its use to the provider name (to be used as an arg to the "init" command)
		providerCmd.Use = p.Slug()
		// run cani's initialization function
		providerCmd.RunE = initSessionWithProviderCmd

		// add it as a sub-command to "init" so when an arg is passed, it will call the appropriate provider command
		SessionInitCmd.AddCommand(providerCmd)

	}

	// Add session commands to root commands
	root.SessionCmd.AddCommand(SessionInitCmd)
	root.SessionCmd.AddCommand(SessionApplyCmd)
	root.SessionCmd.AddCommand(SessionStatusCmd)
	root.SessionCmd.AddCommand(SessionSummaryCmd)

	// Session stop flags
	SessionApplyCmd.Flags().BoolVarP(&commit, "commit", "c", false, "Commit changes to session")
	SessionApplyCmd.Flags().BoolVarP(&dryrun, "dryrun", "d", false, "Perform dryrun, and do not make changes to the system")
	SessionApplyCmd.Flags().BoolVar(&ignoreExternalValidation, "ignore-validation", false, ignoreValidationMessage)
}
