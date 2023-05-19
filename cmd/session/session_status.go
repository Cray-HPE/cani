package session

import (
	"fmt"
	"os"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionStatusCmd represents the session status command
var SessionStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "View session status.",
	Long:  `View session status.`,
	RunE:  showSession,
}

// showSession shows the status of the session
func showSession(cmd *cobra.Command, args []string) error {
	ds := root.Conf.Session.DomainOptions.DatastorePath
	provider := root.Conf.Session.DomainOptions.Provider

	// If the session is active, check that the datastore exists
	if root.Conf.Session.Active {
		_, err := os.Stat(ds)
		if err != nil {
			return fmt.Errorf("Session is ACTIVE with provider '%s' but datastore '%s' does not exist", provider, ds)
		}
		log.Info().Msgf("Session is ACTIVE")
	} else {
		log.Info().Msgf("Session is INACTIVE")
	}

	return nil
}
