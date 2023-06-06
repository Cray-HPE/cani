package session

import (
	root "github.com/Cray-HPE/cani/cmd"
)

var (
	commit        bool
	kubeconfig    string
	caCertPath    string
	insecure      bool
	secretName    string
	clientId      string
	clientSecret  string
	tokenUrl      string
	tokenUsername string
	tokenPassword string
)

func init() {
	// Add session commands to root commands
	root.SessionCmd.AddCommand(SessionStartCmd)
	root.SessionCmd.AddCommand(SessionStopCmd)
	root.SessionCmd.AddCommand(SessionStatusCmd)
	root.SessionCmd.AddCommand(SessionSummaryCmd)
	root.SessionCmd.AddCommand(SessionImportCmd)

	// Session start flags
	// TODO need a quick simulation environment flag
	SessionStartCmd.Flags().String("csm-url-sls", "https://api-gw-service-nmn.local/apis/sls/v1", "(CSM Provider) Base URL for the System Layout Service (SLS)")
	SessionStartCmd.Flags().String("csm-url-hsm", "https://api-gw-service-nmn.local/apis/smd/hsm/v2", "(CSM Provider) Base URL for the Hardware State Manager (HSM)")
	SessionStartCmd.Flags().BoolVarP(&insecure, "csm-insecure-https", "k", false, "(CSM Provider) Allow insecure connections when using HTTPS to CSM services")
	SessionStartCmd.Flags().Bool("csm-sim-urls", false, "(CSM Provider) Use simulation environment URLs")

	// These three pieces are needed for the CSM provider to get a token
	SessionStartCmd.Flags().StringVar(&tokenUrl, "csm-base-auth-url", "", "(CSM Provider) Base URL for the CSM authentication")
	SessionStartCmd.MarkFlagRequired("csm-base-auth-url")
	SessionStartCmd.Flags().StringVar(&tokenUsername, "csm-keycloak-username", "", "(CSM Provider) Keycloak username")
	SessionStartCmd.MarkFlagRequired("csm-keycloak-username")
	SessionStartCmd.Flags().StringVar(&tokenPassword, "csm-keycloak-password", "", "(CSM Provider) Keycloak password")
	SessionStartCmd.MarkFlagRequired("csm-keycloak-password")
	// TODO the API token, do we save ito the file?

	// Less secure auth methods for CSM that follow existing patterns, but to discourage use, mark them hidden
	SessionStartCmd.Flags().StringVar(&kubeconfig, "csm-kube-config", "", "(CSM Provider) Path to the kube config file") // /etc/kubernetes/admin.conf
	SessionStartCmd.Flags().MarkHidden("kube-config")
	SessionStartCmd.Flags().StringVar(&caCertPath, "csm-ca-cert", "", "Path to the CA certificate file") // /etc/pki/trust/anchors/platform-ca-certs.crt"
	SessionStartCmd.Flags().MarkHidden("csm-ca-cert")
	SessionStartCmd.Flags().StringVar(&secretName, "csm-secret-name", "admin-client-auth", "(CSM Provider) secret name")
	SessionStartCmd.Flags().MarkHidden("csm-secret-name")
	SessionStartCmd.Flags().StringVar(&clientId, "csm-client-id", "", "(CSM Provider) Client ID")
	SessionStartCmd.Flags().MarkHidden("csm-client-id")
	SessionStartCmd.Flags().StringVar(&clientSecret, "csm-client-secret", "", "(CSM Provider) Client Secret")
	SessionStartCmd.Flags().MarkHidden("csm-client-secret")

	// Session stop flags
	SessionStopCmd.Flags().BoolVarP(&commit, "commit", "c", false, "Commit changes to session")

}
