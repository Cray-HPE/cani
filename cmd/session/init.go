package session

import (
	root "github.com/Cray-HPE/cani/cmd"
)

var (
	commit bool
)

func init() {
	// Add session commands to root commands
	root.SessionCmd.AddCommand(SessionStartCmd)
	root.SessionCmd.AddCommand(SessionStopCmd)
	root.SessionCmd.AddCommand(SessionStatusCmd)
	root.SessionCmd.AddCommand(SessionReconcileCmd)

	// Session stop flags
	SessionStopCmd.Flags().BoolVarP(&commit, "commit", "c", false, "Commit changes to session")

}
