package cmd

import (
	"github.com/spf13/cobra"
)

// SessionCmd represents the session command
var SessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Interact with a session.",
	Long:  `Interact with a session.`,
	RunE:  runSession,
}

// runSession is the main entrypoint for the cani session command
func runSession(cmd *cobra.Command, args []string) error {
	return nil
}
