package session

import (
	"github.com/Cray-HPE/cani/internal/cani/domain"
	"github.com/spf13/cobra"
)

// SessionStopCmd represents the session stop command
var SessionStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a session.",
	Long:  `Stop a session.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		postRun(cmd, args)
		return nil
	},
}

func postRun(cmd *cobra.Command, args []string) error {
	// Save the crafted session data to the datastore
	domain.Data.SessionActive = false
	return nil
}
