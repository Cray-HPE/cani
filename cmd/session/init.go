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

	// Session start flags
	// TODO need a quick simulation environment flag
	// TODO the API token, do we save ito the file?
	SessionStartCmd.Flags().String("csm-url-sls", "https://api-gw-service-nmn.local/apis/sls/v1", "CSM Provider: Base URL for the System Layout Service (SLS)")
	SessionStartCmd.Flags().String("csm-url-hsm", "https://api-gw-service-nmn.local/apis/smd/hsm/v2", "CSM Provider: Base URL for the Hardware State Manager (HSM)")
	SessionStartCmd.Flags().Bool("csm-insecure-https", false, "CSM Provider: Allow insecure connections when using HTTPS to CSM services")
	SessionStartCmd.Flags().Bool("csm-sim-urls", false, "CSM Provider: Use simulation environment URLs")

	// Session stop flags
	SessionStopCmd.Flags().BoolVarP(&commit, "commit", "c", false, "Commit changes to session")

}
