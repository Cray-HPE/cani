package session

import (
	root "github.com/Cray-HPE/cani/cmd"
)

func init() {
	// Add session commands to root commands
	root.SessionCmd.AddCommand(SessionStartCmd)
	root.SessionCmd.AddCommand(SessionStopCmd)
	root.SessionCmd.AddCommand(SessionStatusCmd)

	// Session start flags
}
