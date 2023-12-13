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
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// MergeProviderFlags creates a new init command
// Initilizing a session is where all the information needed to interact with the inventory system(s) is gathered
// Plugin authors can call this to create their own flags based on their custom business logic
// A few common flags are set here, but the rest is up to the plugin author
func MergeProviderFlags(providerCmd *cobra.Command, caniCmd *cobra.Command) (err error) {
	caniFlagset := &pflag.FlagSet{}

	// get the appropriate flagset from the provider's crafted command
	caniFlagset = caniCmd.Flags()

	// add the provider flags to the command
	if providerCmd != nil {
		providerCmd.Flags().AddFlagSet(caniFlagset)
	}

	return nil
}

func DebugFlags(f *pflag.Flag) {
	// cmd.Flags().VisitAll(debugFlags)
	log.Info().Msgf("flag: %+v", f.Name)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func GetProviderCommand(cmd *cobra.Command, args []string) (providerCmd *cobra.Command, err error) {
	GetActiveDomain()
	log.Debug().Msgf("Getting %s-defined '%s %s' command", ActiveDomain.Provider, cmd.Parent().Name(), cmd.Parent().Name())

	for _, c := range cmd.Commands() {
		if contains(c.Aliases, ActiveDomain.Provider) {
			log.Debug().Msgf("Found %s-defined '%s %s' command", ActiveDomain.Provider, c.Parent().Parent().Name(), c.Parent().Name())
			log.Info().Msgf("GetProviderCommand args %+v", args)
			providerCmd = c
		}
	}

	return providerCmd, nil
}

// func MergeProviderCommand(caniCmd *cobra.Command) (err error) {
// 	for _, p := range domain.GetProviders() {
// 		// each provider can craft their own commands
// 		// since this runs during init(), the domain object is not yet set up, switch statements are used to call the necessary functions
// 		providerCmd, err := domain.NewProviderCmd(caniCmd, p.Slug())
// 		if err != nil {
// 			log.Error().Msgf("unable to get cmd from provider: %v", err)
// 			os.Exit(1)
// 		}

// 		// all flags should be set in init().
// 		// You can set flags after the fact, but it is much easier to work with everything up front
// 		// this will set existing variables for each provider
// 		log.Debug().Msgf("Merging '%s %s' command with %s command", caniCmd.Parent().Name(), caniCmd.Name(), p.Slug())
// 		err = MergeProviderFlags(providerCmd, caniCmd)
// 		if err != nil {
// 			log.Error().Msgf("unable to get flags from provider: %v", err)
// 			os.Exit(1)
// 		}

// 		// the provider command should be the same as the bootstrap command, allowing it to override the bootstrap cmd
// 		providerCmd.Use = caniCmd.Name()
// 		// providerCmd.RunE = caniCmd.RunE

// 	}
// 	return nil
// }

func RegisterProviderCommand(
	p provider.InventoryProvider,
	caniCmd *cobra.Command,
	caniRunE func(cmd *cobra.Command, args []string) (err error)) {

	log.Debug().Msgf("Registering '%s %s' command from %s", caniCmd.Parent().Name(), caniCmd.Name(), p.Slug())
	// Get the provider-defined command
	providerCmd, err := domain.NewProviderCmd(caniCmd, p.Slug())
	if err != nil {
		log.Error().Msgf("unable to get provider init command: %v", err)
		os.Exit(1)
	}

	// // Merge the provider's flags into the cani command
	// err = MergeProviderFlags(providerCmd, caniCmd)
	// if err != nil {
	// 	log.Error().Msgf("unable to get cmd '%s' from provider: %v", caniCmd.Name(), err)
	// 	os.Exit(1)
	// }

	// TODO: execute subcommand automatically so it is seamless to the user
	// at present, the user must run commands + provider name, like:
	//   cani add cabinet csm
	//   cani list blade hpengi
	providerCmd.Hidden = false
	// set the provider command's use to that of the cani command
	providerCmd.Use = p.Slug()
	// but add the provider as an alias so it can be keyed off of during command execution
	providerCmd.Aliases = append(providerCmd.Aliases, caniCmd.Name())
	// add it as a sub-command the cani command so it can be called during runtime
	caniCmd.AddCommand(providerCmd)
}

func ExecuteProviderRunE(cmd *cobra.Command, args []string) (err error) {
	// get a provider command, merged together with
	providerCmd, err := GetProviderCommand(cmd, args)
	if err != nil {
		return err
	}

	// run the prerun cmd if defined by the provider
	if providerCmd.PreRunE != nil {
		err := providerCmd.PreRunE(providerCmd, args)
		if err != nil {
			return err
		}
	}

	// loop through the subcommands to find the appropriate provider command
	// then execute it
	for _, c := range cmd.Commands() {
		for _, a := range c.Aliases {
			if a == ActiveDomain.Provider {
				log.Debug().Msgf("Running %+v-provided '%s %s' RunE function", a, c.Parent().Parent().Name(), c.Name())
				err = providerCmd.RunE(providerCmd, args)
				// show the provider's help instead of the default cani one
				if err != nil {
					log.Error().Msgf("%+v", err)
					providerCmd.Help()
				}
			}
		}
	}

	return nil
}
