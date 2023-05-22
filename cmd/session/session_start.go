package session

import (
	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/cmd/config"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionStartCmd represents the session start command
var SessionStartCmd = &cobra.Command{
	Use:                "start",
	Short:              "Start a session.",
	Long:               `Start a session.`,
	Args:               validProvider,
	ValidArgs:          validArgs,
	RunE:               startSession,
	PersistentPostRunE: writeSession,
}

var (
	cfgFile   string
	ds        string
	dopts     *domain.NewOpts
	provider  string
	validArgs = []string{"csm"}
)

// startSession starts a session if one does not exist
func startSession(cmd *cobra.Command, args []string) error {
	ds := root.Conf.Session.DomainOptions.DatastorePath
	logfile := root.Conf.Session.DomainOptions.LogFilePath
	provider := root.Conf.Session.DomainOptions.Provider

	// If a session is already active, there is nothing to do
	if root.Conf.Session.Active {
		log.Info().Msgf("Session is already ACTIVE.  Nothing to do.")
		return nil
	}
	// If a session is not active, create one
	_, err := inventory.NewDatastoreJSON(ds, logfile)
	if err != nil {
		return err
	}

	root.Domain, err = domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}
	// "Activate" the session
	root.Conf.Session.Active = true

	log.Info().Msgf("Session is now ACTIVE with provider %s and datastore %s", provider, ds)
	return nil
}

// writeSession writes the session configuration back to the config file
func writeSession(cmd *cobra.Command, args []string) error {
	// Write the configuration back to the file
	cfgFile := cmd.Root().PersistentFlags().Lookup("config").Value.String()
	err := config.WriteConfig(cfgFile, root.Conf)
	if err != nil {
		return err
	}
	return nil
}
