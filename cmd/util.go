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
	"os"

	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// MergeProviderFlags creates a new init command
// Initilizing a session is where all the information needed to interact with the inventory system(s) is gathered
// Plugin authors can call this to create their own flags based on their custom business logic
// A few common flags are set here, but the rest is up to the plugin author
func MergeProviderFlags(bootstrapCmd *cobra.Command, providerCmd *cobra.Command) (err error) {
	providerFlagset := &pflag.FlagSet{}

	// get the appropriate flagset from the provider's crafted command
	providerFlagset = providerCmd.Flags()

	if err != nil {
		return err
	}

	// add the provider flags to the command
	bootstrapCmd.Flags().AddFlagSet(providerFlagset)

	return nil
}

func MergeProviderCommand(bootstrapCmd *cobra.Command) (err error) {
	// each provider can craft their own commands
	// since this runs during init(), the domain object is not yet set up, switch statements are used to call the necessary functions
	log.Debug().Msgf("Merging '%s %s' command with provider command", bootstrapCmd.Parent().Name(), bootstrapCmd.Name())
	providerCmd := &cobra.Command{}
	providerCmd, err = domain.NewProviderCmd(bootstrapCmd)
	if err != nil {
		log.Error().Msgf("unable to get cmd from provider: %v", err)
		os.Exit(1)
	}
	// the provider command should be the same as the bootstrap command, allowing it to override the bootstrap cmd
	providerCmd.Use = bootstrapCmd.Name()

	// all flags should be set in init().
	// You can set flags after the fact, but it is much easier to work with everything up front
	// this will set existing variables for each provider
	err = MergeProviderFlags(bootstrapCmd, providerCmd)
	if err != nil {
		log.Error().Msgf("unable to get flags from provider: %v", err)
		os.Exit(1)
	}

	// Now the provider command has CANI's settings and those set by the provider
	// It may seem redundant to run this again, but in order to do things like MarkFlagsRequiredTogether(),
	// it is necessary to have all of the flags available during init, which is what the MergeProviderFlags will do
	err = domain.UpdateProviderCmd(bootstrapCmd)
	if err != nil {
		log.Error().Msgf("unable to get cmd from provider: %v", err)
		os.Exit(1)
	}
	return nil
}
