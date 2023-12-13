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
	"fmt"

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

// RegisterProviderCommand adds a provider-defined command as a subcommand to a specific cani command
// this is meant to be run during init() so each provider has their commands available for use
// TODO: execute subcommand automatically so it is seamless to the user
// at present, the user must run commands + provider name, like:
//
//	cani add cabinet csm
//	cani list blade hpengi
//
// this requires hiding the provider sub command and dynamically executing it, as opposed to making the user type it in
func RegisterProviderCommand(p provider.InventoryProvider, caniCmd *cobra.Command) (err error) {
	log.Debug().Msgf("Registering '%s %s' command from %s", caniCmd.Parent().Name(), caniCmd.Name(), p.Slug())
	// Get the provider-defined command
	providerCmd, err := domain.NewProviderCmd(caniCmd, p.Slug())
	if err != nil {
		return fmt.Errorf("unable to get provider init() command: %v", err)
	}

	providerCmd.Hidden = false
	// set the provider command's use to that of the cani command
	providerCmd.Use = p.Slug()
	// but add the provider as an alias so it can be keyed off of during command execution
	providerCmd.Aliases = append(providerCmd.Aliases, caniCmd.Name())
	// add it as a sub-command the cani command so it can be called during runtime
	caniCmd.AddCommand(providerCmd)

	return nil
}
