package cmd

import (
	"github.com/Cray-HPE/cani/cmd/session"
	"github.com/spf13/cobra"
)

var (
	ds       string
	provider string
)

// sessionCmd represents the session command
var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Interact with a session.",
	Long:  `Interact with a session.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	sessionCmd.AddCommand(session.SessionStartCmd)
	sessionCmd.AddCommand(session.SessionStatusCmd)
	sessionCmd.AddCommand(session.SessionStopCmd)

	sessionCmd.PersistentFlags().StringVarP(&ds, "datastore", "d", ds, "Path to datastore")
}
