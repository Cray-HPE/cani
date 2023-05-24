package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Cray-HPE/cani/cmd/config"
	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	// Create or load a yaml config and the database
	cobra.OnInitialize(initConfig, setupLogging)

	RootCmd.AddCommand(AddCmd)
	RootCmd.AddCommand(ListCmd)
	RootCmd.AddCommand(RemoveCmd)
	RootCmd.AddCommand(SessionCmd)
	RootCmd.AddCommand(UpdateCmd)
	RootCmd.AddCommand(VersionCmd)

	// Global root command flags
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, "Path to the configuration file")
	RootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "D", false, "additional debug output")
	RootCmd.PersistentFlags().BoolVarP(&Simulation, "simulation", "S", false, "Use simulation mode for hsm-simulation-environment")

	// Global add flags
	AddCmd.PersistentFlags().StringVarP(&vendor, "vendor", "m", "HPE", "Vendor")
	AddCmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Name")
	AddCmd.PersistentFlags().StringVarP(&u, "uuid", "u", "", "Specific UUID to use")
}

// setupLogging sets up the global logger
func setupLogging() {
	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if Debug {
		// enable debug output globally
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled")
		// include file and line number in debug output
		//log.Logger = log.With().Caller().Logger()
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

func loadConfigAndDomainOpts(cmd *cobra.Command, args []string) error {
	var err error
	Conf, err = config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}
	if Debug {
		log.Debug().Msgf("Loaded config file %s", cfgFile)
		log.Debug().Msgf("Session: %+v", Conf.Session.Active)
	}

	return nil
}