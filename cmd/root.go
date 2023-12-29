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

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Cray-HPE/cani/cmd/config"
	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:               taxonomy.App,
	Short:             taxonomy.ShortDescription,
	Long:              taxonomy.LongDescription,
	PersistentPreRunE: setupDomain, // the domain object is needed for all provider operations, load it early from root so it is available to all subcommands
	// RunE:               runRoot,
	Version:            version(),
	PersistentPostRunE: WriteSession, // write any changes made back to the config
}

var (
	// cfgFile is the global configuration file path (from --config)
	cfgFile string
	// Debug is a global flag that enables debug logging
	Debug bool
	// Verbose is a global flag that enables verbose logging
	Verbose bool
	// This is the active domain being used
	D *domain.Domain
	// Conf is everything in the config file (all sessions for all providers)
	Conf *config.Config
	// Hardware library should also be accessible globally
	HwLibrary *hardwaretypes.Library
	// List of hardware types that are blades
	CabinetTypes, BladeTypes, NodeTypes, MgmtSwitchTypes, HsnSwitchTypes []hardwaretypes.DeviceType
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// runRoot is the main entrypoint for the cani command
func runRoot(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmd.Help()
	}

	return nil
}

// WriteSession writes the session configuration back to the config file
func WriteSession(cmd *cobra.Command, args []string) error {
	if cmd.Parent().Name() == "init" {
		// Write the configuration back to the file
		cfgFile := cmd.Root().PersistentFlags().Lookup("config").Value.String()
		log.Debug().Msgf("Writing session to config %s", cfgFile)
		err := config.WriteConfig(cfgFile, Conf)
		if err != nil {
			return err
		}
	}
	return nil
}
