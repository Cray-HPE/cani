package session

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionStatusCmd represents the session status command
var SessionStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "View session status.",
	Long:  `View session status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := sessionShow(cmd, args)
		if err != nil {
			return err
		}
		return nil
	},
}

func sessionShow(cmd *cobra.Command, args []string) error {
	if Conf.Session.Active {
		ds := Conf.Session.DomainOptions.DatastorePath
		provider := Conf.Session.DomainOptions.Provider
		_, err := os.Stat(ds)
		if err != nil {
			return fmt.Errorf("Session is ACTIVE with provider '%s' but datastore '%s' does not exist", provider, ds)
		}

		log.Info().Msgf("Session with provider '%s' and datastore '%s' is ACTIVE", provider, ds)
	} else {
		log.Info().Msgf("Session with provider '%s' and datastore '%s' is INACTIVE", provider, ds)
	}

	return nil
}
