/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
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
	providerHost  string
	tokenUsername string
	tokenPassword string
	useSimulation bool
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
	SessionStartCmd.Flags().BoolVarP(&useSimulation, "csm-simulator", "S", false, "(CSM Provider) Use simulation environment URLs")

	// These three pieces are needed for the CSM provider to get a token
	SessionStartCmd.Flags().StringVar(&providerHost, "csm-api-host", "api-gw-service-nmn.local", "(CSM Provider) Host or FQDN for authentation and APIs")
	// SessionStartCmd.MarkFlagRequired("csm-api-host")
	SessionStartCmd.Flags().StringVar(&tokenUsername, "csm-keycloak-username", "", "(CSM Provider) Keycloak username")
	// SessionStartCmd.MarkFlagRequired("csm-keycloak-username")
	SessionStartCmd.Flags().StringVar(&tokenPassword, "csm-keycloak-password", "", "(CSM Provider) Keycloak password")
	// SessionStartCmd.MarkFlagRequired("csm-keycloak-password")
	SessionStartCmd.MarkFlagsRequiredTogether("csm-api-host", "csm-keycloak-username", "csm-keycloak-password")
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
