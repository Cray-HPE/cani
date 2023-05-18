package session

import (
	"fmt"
	"os"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionStopCmd represents the session stop command
var SessionStopCmd = &cobra.Command{
	Use:                "stop",
	Short:              "Stop a session.",
	Long:               `Stop a session.`,
	RunE:               stopSession,
	PersistentPostRunE: writeSession,
}

// stopSession stops a session if one exists
func stopSession(cmd *cobra.Command, args []string) error {
	ds := root.Conf.Session.DomainOptions.DatastorePath
	provider := root.Conf.Session.DomainOptions.Provider

	if root.Conf.Session.Active {
		// "Deactivate" the session
		root.Conf.Session.Active = false

		// Check that the datastore exists before proceeding since we cannot continue without it
		_, err := os.Stat(ds)
		if err != nil {
			return fmt.Errorf("Session is STOPPED with provider '%s' but datastore '%s' does not exist", provider, ds)
		}
		log.Info().Msgf("Session is STOPPED")
	} else {
		log.Info().Msgf("Session with provider '%s' and datastore '%s' is already STOPPED", provider, ds)
	}

	return nil
}
