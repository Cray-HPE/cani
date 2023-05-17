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
	ds, sesh, err := getDatastoreAndSession(cmd, args)
	if err != nil {
		return err
	}

	// Check if the session file exists
	if _, err := os.Stat(sesh); os.IsNotExist(err) {
		// If the session file does not exist, there is no active session
		log.Info().Msg(fmt.Sprintf("%s is already INACTIVE", sesh))
		os.Exit(1)
	}

	// Delete the session file
	err = os.Remove(sesh)
	if err != nil {
		return err
	}

	// commit to the database
	log.Info().Msgf("%s stopped.", sesh)
	log.Info().Msgf("Push/commit %s to the datastore", ds)

	return nil
}
