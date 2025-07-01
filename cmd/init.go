/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	"log"

	"github.com/Cray-HPE/cani/cmd/add"
	"github.com/Cray-HPE/cani/cmd/alpha"
	"github.com/Cray-HPE/cani/cmd/remove"
	"github.com/Cray-HPE/cani/cmd/serve"
	"github.com/Cray-HPE/cani/cmd/session"
	"github.com/Cray-HPE/cani/cmd/show"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/spf13/cobra"
)

var (
	rootCmd *cobra.Command
	cfgFile string
)

func Init() {
	// initialize the root command
	rootCmd = newRootCommand()

	// build core verbs
	sessionCmd := session.NewCommand()
	addCmd := add.NewCommand()
	removeCmd := remove.NewCommand() // reusing the add package for remove functionality
	serveCmd := serve.NewCommand()
	showCmd := show.NewCommand()

	// at present, all commands are under the alpha command since this is still a work in progress
	alphaCmd := alpha.NewCommand()
	rootCmd.AddCommand(alphaCmd)
	alphaCmd.AddCommand(
		sessionCmd,
		addCmd,
		removeCmd,
		showCmd,
		serveCmd,
	)

	// now for each verb (except session init), ask each provider to decorate it
	for _, caniCmd := range rootCmd.Commands() {
		for _, p := range provider.GetProviders() {
			if providerCmd, err := p.NewProviderCmd(caniCmd); err == nil {
				if providerCmd == nil {
					// this provider doesn’t customize that verb
					continue
				}
				// log.Printf("Merging in %s command from %s", providerCmd.Name(), p.Slug())
				// mergeProviderCommand(caniCmd, providerCmd)
			}
		}
	}

}

// mergeProviderCommand copies metadata, flags, RunE/Args and sub-commands
// from providerCmd into baseCmd
func mergeProviderCommand(baseCmd *cobra.Command, providerCmd *cobra.Command) {
	if providerCmd.Short != "" {
		baseCmd.Short = providerCmd.Short
	}
	if providerCmd.Long != "" {
		baseCmd.Long = providerCmd.Long
	}
	if providerCmd.RunE != nil {
		baseCmd.RunE = providerCmd.RunE
	}
	if providerCmd.Args != nil {
		baseCmd.Args = providerCmd.Args
	}

	// merge flags
	baseCmd.Flags().AddFlagSet(providerCmd.Flags())
	baseCmd.PersistentFlags().AddFlagSet(providerCmd.PersistentFlags())

	// merge valid args, examples, aliases, annotations
	baseCmd.ValidArgs = providerCmd.ValidArgs
	baseCmd.Example = providerCmd.Example
	baseCmd.Aliases = append(baseCmd.Aliases, providerCmd.Aliases...)
	if baseCmd.Annotations == nil {
		baseCmd.Annotations = map[string]string{}
	}
	for k, v := range providerCmd.Annotations {
		baseCmd.Annotations[k] = v
	}

	// and finally merge any children
	for _, sc := range providerCmd.Commands() {
		log.Printf("Merging %s sub-command into %s", sc.Name(), baseCmd.Name())
		baseCmd.AddCommand(sc)
	}
}
