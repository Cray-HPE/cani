package session

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionStopCmd represents the session stop command
var SessionStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a session.",
	Long:  `Stop a session.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stopSession(cmd, args)
		return nil
	},
}

func stopSession(cmd *cobra.Command, args []string) error {
	// Sanity check that the datastore exists
	ds := Conf.Session.DomainOptions.DatastorePath
	provider := Conf.Session.DomainOptions.Provider

	if Conf.Session.Active {
		// "Deactivate" the session
		Conf.Session.Active = false

		_, err := os.Stat(ds)
		if err != nil {
			return fmt.Errorf("Session is STOPPED with provider '%s' but datastore '%s' does not exist", provider, ds)
		}

		log.Info().Msgf("Session is STOPPED")
		log.Info().Msgf("Next step: commit changes from '%s'", ds)
	} else {
		log.Info().Msgf("Session with provider '%s' and datastore '%s' is already STOPPED", provider, ds)
	}

	return nil
}
