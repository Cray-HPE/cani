/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Cray-HPE/cani/cmd/blade"
	"github.com/Cray-HPE/cani/cmd/config"
	"github.com/Cray-HPE/cani/cmd/inventory"
	"github.com/Cray-HPE/cani/cmd/taxonomy"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "cani",
	Short: "From subfloor to top-of-rack, manage your HPC cluster's inventory!",
	Long:  `From subfloor to top-of-rack, manage your HPC cluster's inventory!`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			blade.EnableDebug()
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
		}
		setupLogging()
		return nil
	},
}

var (
	debug      bool
	simulation bool
	cfgFile    string
	conf       *config.Config
	spec       bool
	dbFile     string
	// the database is exported so it can be used in the subcommands
	Db *inventory.Database
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Create or load a yaml config and the database
	cobra.OnInitialize(initConfig, initDb)

	RootCmd.AddCommand(addCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(removeCmd)
	RootCmd.AddCommand(versionCmd)

	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, "Path to the configuration file")
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "additional debug output")
	RootCmd.PersistentFlags().BoolVarP(&simulation, "simulation", "S", false, "Use simulation mode for hsm-simulation-environment")
	RootCmd.PersistentFlags().StringVar(&dbFile, "database", dbFile, "JSON database file")

}

// setupLogging sets up the global logger
func setupLogging() {
	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		// enable debug output globally
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled")
		// include file and line number in debug output
		log.Logger = log.With().Caller().Logger()
	}
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	homeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)
	if cfgFile != "" {
		// global debug cannot be run during init() so check for debug flag here
		if debug {
			log.Debug().Msg(fmt.Sprintf("Using user-defined config file: %s", cfgFile))
		}
	} else {
		// Set a default configuration file
		cfgFile = filepath.Join(homeDir, taxonomy.CfgDir, taxonomy.CfgFile)
		if debug {
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
	conf, err = config.LoadConfig(cfgFile, conf)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Error loading config file: %s", err))
		os.Exit(1)
	}
	// Set up other global flags or settings based on the loaded configuration
}

func initDb() {
	homeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)
	if dbFile != "" {
		// global debug cannot be run during init() so check for debug flag here
		if debug {
			log.Debug().Msg(fmt.Sprintf("Using user-defined database file: %s", dbFile))
		}
	} else {
		// Set a default database file
		dbFile = filepath.Join(homeDir, taxonomy.CfgDir, inventory.DbPath)
		if debug {
			log.Debug().Msg(fmt.Sprintf("Using default database file %s", dbFile))
		}
	}
	// Initialize the database file if it does not exist
	Db, err = inventory.InitDb(dbFile)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Error initializing database file: %s", err))
		os.Exit(1)
	}

	// Load the configuration file
	Db, err = inventory.LoadDb(dbFile, Db)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Error loading database file: %s", err))
		os.Exit(1)
	}

	if debug {
		log.Debug().Msg(fmt.Sprintf("Loaded %s", dbFile))
	}

	defer inventory.CloseTransactionLog()
}

// GetDbPointer returns a pointer to the database
func GetDbPointer() *inventory.Database {
	return Db
}
