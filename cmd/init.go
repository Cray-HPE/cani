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
	"os"
	"path/filepath"

	"github.com/Cray-HPE/cani/cmd/config"
	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	// Create or load a yaml config and the database
	cobra.OnInitialize(setupLogging, initConfig)

	RootCmd.AddCommand(AlphaCmd)
	RootCmd.AddCommand(MakeDocsCmd)
	RootCmd.AddCommand(MakeProviderCmd)
	AlphaCmd.AddCommand(AddCmd)
	AlphaCmd.AddCommand(ListCmd)
	AlphaCmd.AddCommand(RemoveCmd)
	AlphaCmd.AddCommand(SessionCmd)
	AlphaCmd.AddCommand(UpdateCmd)
	AlphaCmd.AddCommand(ValidateCmd)

	AlphaCmd.AddCommand(ExportCmd)
	err := MergeProviderCommand(ExportCmd)
	if err != nil {
		log.Error().Msgf("%+v", err)
		os.Exit(1)
	}

	AlphaCmd.AddCommand(ImportCmd)
	err = MergeProviderCommand(ImportCmd)
	if err != nil {
		log.Error().Msgf("%+v", err)
		os.Exit(1)
	}

	// Global root command flags
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, "Path to the configuration file")
	RootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "D", false, "additional debug output")
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "additional verbose output")
}

// setupLogging sets up the global logger
func setupLogging() {
	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	// Fancy, human-friendly console logger (but slower)
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out: os.Stderr,
			// When not in a terminal disable color
			NoColor: !term.IsTerminal(int(os.Stderr.Fd())),
		},
	)
	if Debug {
		// enable debug output globally
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled")
		// include file and line number in debug output
		if Verbose {
			log.Logger = log.With().Caller().Logger()
		}
	}
}

func initConfig() {
	homeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)
	if cfgFile != "" {
		// global debug cannot be run during init() so check for debug flag here
		if Debug {
			log.Debug().Msg(fmt.Sprintf("Using user-defined config file: %s", cfgFile))
		}
	} else {
		// Set a default configuration file
		cfgFile = filepath.Join(homeDir, taxonomy.CfgPath)
		if Debug {
			log.Debug().Msg(fmt.Sprintf("Using default config file %s", cfgFile))
		}
	}
	// Initialize the configuration file if it does not exist
	err = config.InitConfig(cfgFile)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Error initializing config file: %s", err))
		os.Exit(1)
	}

	// Load the configuration file
	Conf, err = config.LoadConfig(cfgFile)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Error loading config file: %s", err))
		os.Exit(1)
	}

}

func setupDomain(cmd *cobra.Command, args []string) (err error) {
	log.Debug().Msgf("Setting %s provider", cmd.Name())
	log.Debug().Msgf("Setting up domain for command: %s", cmd.Name())
	// if cmd.Name() != "init" || cmd.Name() != "status" {
	log.Debug().Msgf("Checking for active domains")
	// Find an active session
	activeDomains := []*domain.Domain{}
	activeProviders := []string{}
	for p, d := range Conf.Session.Domains {
		if d.Active {
			log.Debug().Msgf("Provider '%s' is ACTIVE", p)
			activeDomains = append(activeDomains, d)
			activeProviders = append(activeProviders, p)
		} else {
			log.Debug().Msgf("Provider '%s' is inactive", p)
		}
	}

	if cmd.Name() != "init" {
		// Error if no sessions are active
		if len(activeProviders) == 0 {
			// These commands are special because they validate hardware in the args
			// so SetupDomain is called manually
			// The timing of events works out such that simply returning the error
			// will exit without the message
			if cmd.Name() == "status" {
				log.Info().Msgf("No active session.")
				return nil
			} else {
				log.Error().Msgf("No active session.")
				return err
			}
		}

		// Check that only one session is active
		if len(activeProviders) > 1 {
			for _, p := range activeProviders {
				err := fmt.Errorf("currently active: %v", p)
				log.Error().Msgf("%v", err)
			}
			log.Error().Msgf("only one session may be active at a time")
			return err
		}
		activeDomain := activeDomains[0]

		log.Debug().Msgf("Active provider is: %s", activeDomain.Provider)
		D = activeDomain
		err = D.SetupDomain(cmd, args, Conf.Session.Domains)
		if err != nil {
			return err
		}
		HwLibrary, err := hardwaretypes.NewEmbeddedLibrary(D.CustomHardwareTypesDir)
		if err != nil {
			return err
		}

		// Get the list of supported hardware types
		CabinetTypes = HwLibrary.GetDeviceTypesByHardwareType(hardwaretypes.Cabinet)
		BladeTypes = HwLibrary.GetDeviceTypesByHardwareType(hardwaretypes.NodeBlade)
		NodeTypes = HwLibrary.GetDeviceTypesByHardwareType(hardwaretypes.NodeBlade)
		MgmtSwitchTypes = HwLibrary.GetDeviceTypesByHardwareType(hardwaretypes.ManagementSwitch)
		HsnSwitchTypes = HwLibrary.GetDeviceTypesByHardwareType(hardwaretypes.HighSpeedSwitch)
	}
	return nil
}
