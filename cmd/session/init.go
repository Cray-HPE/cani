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
	dryrun                   bool
	commit                   bool
	ignoreExternalValidation bool
	k8sPodsCidr              string
	k8sServicesCidr          string
	kubeconfig               string
	caCertPath               string
	insecure                 bool
	secretName               string
	clientId                 string
	clientSecret             string
	providerHost             string
	tokenUsername            string
	tokenPassword            string
	useSimulation            bool
)

func init() {
	// temporary varables for shared messages
	ignoreValidationMessage :=
		"(CSM Provider) Ignore validation failures. Use this to allow unconventional SLS configurations."

	// Add session commands to root commands
	root.SessionCmd.AddCommand(SessionInitCmd)
	root.SessionCmd.AddCommand(SessionApplyCmd)
	root.SessionCmd.AddCommand(SessionStatusCmd)
	root.SessionCmd.AddCommand(SessionSummaryCmd)

	// Session init flags
	SessionInitCmd.Flags().String("csm-url-sls", "", "(CSM Provider) Base URL for the System Layout Service (SLS)")
	SessionInitCmd.Flags().String("csm-url-hsm", "", "(CSM Provider) Base URL for the Hardware State Manager (HSM)")
	SessionInitCmd.Flags().BoolVarP(&insecure, "csm-insecure-https", "k", false, "(CSM Provider) Allow insecure connections when using HTTPS to CSM services")
	SessionInitCmd.Flags().BoolVarP(&useSimulation, "csm-simulator", "S", false, "(CSM Provider) Use simulation environment URLs")

	// These three pieces are needed for the CSM provider to get a token
	SessionInitCmd.Flags().StringVar(&providerHost, "csm-api-host", "api-gw-service.local", "(CSM Provider) Host or FQDN for authentation and APIs")
	// SessionInitCmd.MarkFlagRequired("csm-api-host")
	SessionInitCmd.Flags().StringVar(&tokenUsername, "csm-keycloak-username", "", "(CSM Provider) Keycloak username")
	// SessionInitCmd.MarkFlagRequired("csm-keycloak-username")
	SessionInitCmd.Flags().StringVar(&tokenPassword, "csm-keycloak-password", "", "(CSM Provider) Keycloak password")
	// SessionInitCmd.MarkFlagRequired("csm-keycloak-password")
	SessionInitCmd.MarkFlagsRequiredTogether("csm-api-host", "csm-keycloak-username", "csm-keycloak-password")
	// TODO the API token, do we save ito the file?

	SessionInitCmd.Flags().StringVar(&k8sPodsCidr, "csm-k8s-pods-cidr", "10.32.0.0/12", "(CSM Provider) CIDR used by kubernetes for pods")
	SessionInitCmd.Flags().StringVar(&k8sServicesCidr, "csm-k8s-services-cidr", "10.16.0.0/12", "(CSM Provider) CIDR used by kubernetes for services")
	// Less secure auth methods for CSM that follow existing patterns, but to discourage use, mark them hidden
	SessionInitCmd.Flags().StringVar(&kubeconfig, "csm-kube-config", "", "(CSM Provider) Path to the kube config file") // /etc/kubernetes/admin.conf
	SessionInitCmd.Flags().MarkHidden("kube-config")
	SessionInitCmd.Flags().StringVar(&caCertPath, "csm-ca-cert", "", "Path to the CA certificate file") // /etc/pki/trust/anchors/platform-ca-certs.crt"
	SessionInitCmd.Flags().MarkHidden("csm-ca-cert")
	SessionInitCmd.Flags().StringVar(&secretName, "csm-secret-name", "admin-client-auth", "(CSM Provider) secret name")
	SessionInitCmd.Flags().MarkHidden("csm-secret-name")
	SessionInitCmd.Flags().StringVar(&clientId, "csm-client-id", "", "(CSM Provider) Client ID")
	SessionInitCmd.Flags().MarkHidden("csm-client-id")
	SessionInitCmd.Flags().StringVar(&clientSecret, "csm-client-secret", "", "(CSM Provider) Client Secret")
	SessionInitCmd.Flags().MarkHidden("csm-client-secret")
	SessionInitCmd.Flags().BoolVar(&ignoreExternalValidation, "ignore-validation", false, ignoreValidationMessage)

	// Session stop flags
	SessionApplyCmd.Flags().BoolVarP(&commit, "commit", "c", false, "Commit changes to session")
	SessionApplyCmd.Flags().BoolVarP(&dryrun, "dryrun", "d", false, "Perform dryrun, and do not make changes to the system")
	SessionApplyCmd.Flags().BoolVar(&ignoreExternalValidation, "ignore-validation", false, ignoreValidationMessage)
}
