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
	"github.com/Cray-HPE/cani/cmd/add"
	"github.com/Cray-HPE/cani/cmd/alpha"
	"github.com/Cray-HPE/cani/cmd/classify"
	"github.com/Cray-HPE/cani/cmd/export"
	imprt "github.com/Cray-HPE/cani/cmd/import"
	initcmd "github.com/Cray-HPE/cani/cmd/init"
	"github.com/Cray-HPE/cani/cmd/remove"
	"github.com/Cray-HPE/cani/cmd/serve"
	"github.com/Cray-HPE/cani/cmd/show"
	"github.com/Cray-HPE/cani/cmd/update"
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

	// add the init command at the root level (not under alpha)
	initProviderCmd := initcmd.NewCommand()
	rootCmd.AddCommand(initProviderCmd)

	// build core verbs
	importCmd := imprt.NewCommand()
	addCmd := add.NewCommand()
	removeCmd := remove.NewCommand()
	serveCmd := serve.NewCommand()
	showCmd := show.NewCommand()
	exportCmd := export.NewCommand()
	updateCmd := update.NewCommand()
	classifyCmd := classify.NewCommand()

	// at present, all commands are under the alpha command since this is still a work in progress
	alphaCmd := alpha.NewCommand()
	rootCmd.AddCommand(alphaCmd)
	alphaCmd.AddCommand(
		importCmd,
		addCmd,
		removeCmd,
		showCmd,
		serveCmd,
		exportCmd,
		updateCmd,
		classifyCmd,
	)

	// Let providers decorate import and export commands only.
	// Normal CRUD operations (add, remove, update, show) use
	// cmd/ + pkg/devicetypes + pkg/datastores without provider hooks.
	for _, caniCmd := range alphaCmd.Commands() {
		switch caniCmd.Name() {
		case "import", "export":
			for _, p := range provider.GetProviders() {
				if providerCmd, err := p.NewProviderCmd(caniCmd); err == nil {
					if providerCmd == nil {
						continue
					}
				}
			}
		}
	}
}
