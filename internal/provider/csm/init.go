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
package csm

import "github.com/spf13/cobra"

var (
	k8sPodsCidr     string
	k8sServicesCidr string
	kubeconfig      string
	caCertPath      string
	insecure        bool
	secretName      string
	clientId        string
	clientSecret    string
	providerHost    string
	tokenUsername   string
	tokenPassword   string
	useSimulation   bool
	vlanId          int

	// import properties
	csvFile string

	// export arguments
	csvHeaders        string
	csvComponentTypes string
	csvAllTypes       bool
	csvListOptions    bool
	exportFormat      string
	ignoreValidation  bool

	// add blade
	role    string
	subrole string
	nid     int
	alias   string
)

func NewSessionInitCommand() (cmd *cobra.Command, err error) {
	// cmd represents the session init command
	cmd = &cobra.Command{}
	cmd.Long = `Query SLS and HSM.  Validate the data against a schema before allowing an import into CANI.`
	// ValidArgs:    DO NOT CONFIGURE.  This is set by cani's cmd pkg
	// Args:         DO NOT CONFIGURE.  This is set by cani's cmd pkg
	// RunE:         DO NOT CONFIGURE.  This is set by cani's cmd pkg
	// Session init flags
	cmd.Flags().String("csm-url-sls", "", "(CSM Provider) Base URL for the System Layout Service (SLS)")
	cmd.Flags().String("csm-url-hsm", "", "(CSM Provider) Base URL for the Hardware State Manager (HSM)")
	cmd.Flags().BoolVarP(&insecure, "csm-insecure-https", "k", false, "(CSM Provider) Allow insecure connections when using HTTPS to CSM services")
	cmd.Flags().BoolVarP(&useSimulation, "csm-simulator", "S", false, "(CSM Provider) Use simulation environment URLs")

	// These three pieces are needed for the CSM provider to get a token
	cmd.Flags().StringVar(&providerHost, "csm-api-host", "api-gw-service.local", "(CSM Provider) Host or FQDN for authentation and APIs")
	// cmd.MarkFlagRequired("csm-api-host")
	cmd.Flags().StringVar(&tokenUsername, "csm-keycloak-username", "", "(CSM Provider) Keycloak username")
	// cmd.MarkFlagRequired("csm-keycloak-username")
	cmd.Flags().StringVar(&tokenPassword, "csm-keycloak-password", "", "(CSM Provider) Keycloak password")
	// cmd.MarkFlagRequired("csm-keycloak-password")
	cmd.MarkFlagsRequiredTogether("csm-api-host", "csm-keycloak-username", "csm-keycloak-password")
	// TODO the API token, do we save ito the file?

	cmd.Flags().StringVar(&k8sPodsCidr, "csm-k8s-pods-cidr", "10.32.0.0/12", "(CSM Provider) CIDR used by kubernetes for pods")
	cmd.Flags().StringVar(&k8sServicesCidr, "csm-k8s-services-cidr", "10.16.0.0/12", "(CSM Provider) CIDR used by kubernetes for services")
	// Less secure auth methods for CSM that follow existing patterns, but to discourage use, mark them hidden
	cmd.Flags().StringVar(&kubeconfig, "csm-kube-config", "", "(CSM Provider) Path to the kube config file") // /etc/kubernetes/admin.conf
	cmd.Flags().MarkHidden("kube-config")
	cmd.Flags().StringVar(&caCertPath, "csm-ca-cert", "", "Path to the CA certificate file") // /etc/pki/trust/anchors/platform-ca-certs.crt"
	cmd.Flags().MarkHidden("csm-ca-cert")
	cmd.Flags().StringVar(&secretName, "csm-secret-name", "admin-client-auth", "(CSM Provider) secret name")
	cmd.Flags().MarkHidden("csm-secret-name")
	cmd.Flags().StringVar(&clientId, "csm-client-id", "", "(CSM Provider) Client ID")
	cmd.Flags().MarkHidden("csm-client-id")
	cmd.Flags().StringVar(&clientSecret, "csm-client-secret", "", "(CSM Provider) Client Secret")
	cmd.Flags().MarkHidden("csm-client-secret")

	return cmd, nil
}

func NewAddCabinetCommand() (cmd *cobra.Command, err error) {
	// cmd represents the session init command
	cmd = &cobra.Command{}
	cmd.Flags().Int("vlan-id", -1, "Vlan ID for the cabinet.")

	return cmd, nil
}

// UpdateAddCabinetCommand is run during init and allows the provider to set additional options for CANI flags
// such as marking certain options mutually exclusive with the auto flag
func UpdateAddCabinetCommand(caniCmd *cobra.Command) error {
	caniCmd.MarkFlagsRequiredTogether("cabinet", "vlan-id")
	caniCmd.MarkFlagsMutuallyExclusive("auto")
	return nil
}

func NewAddNodeCommand() (cmd *cobra.Command, err error) {
	// cmd represents for cani alpha add node
	cmd = &cobra.Command{}
	cmd.Flags().StringVar(&role, "role", "", "Role of the node")
	cmd.Flags().StringVar(&subrole, "subrole", "", "Subrole of the node")
	cmd.Flags().IntVar(&nid, "nid", 0, "NID of the node")
	cmd.Flags().StringVar(&alias, "alias", "", "Alias of the node")

	return cmd, nil
}

func NewUpdateNodeCommand() (cmd *cobra.Command, err error) {
	// cmd represents for cani alpha update node
	cmd = &cobra.Command{}
	cmd.Flags().String("role", "", "Role of the node")
	cmd.Flags().String("subrole", "", "Subrole of the node")
	cmd.Flags().Int("nid", 0, "NID of the node")
	cmd.Flags().StringSlice("alias", []string{}, "Comma-separated aliases of the node")

	return cmd, nil
}

// UpdateUpdateNodeCommand
func UpdateUpdateNodeCommand(caniCmd *cobra.Command) error {

	return nil
}

func NewExportCommand() (cmd *cobra.Command, err error) {
	// cmd represents cani alpha export
	cmd = &cobra.Command{}
	cmd.Flags().StringVar(
		&csvHeaders, "headers", "Type,Vlan,Role,SubRole,Status,Nid,Alias,Name,ID,Location", "Comma separated list of fields to get")
	cmd.Flags().StringVarP(
		&csvComponentTypes, "type", "t", "Node,Cabinet", "Comma separated list of the types of components to output")
	cmd.Flags().BoolVarP(&csvAllTypes, "all", "a", false, "List all components. This overrides the --type option")
	cmd.Flags().BoolVarP(&csvListOptions, "list-fields", "L", false, "List details about the fields in the CSV")
	cmd.Flags().StringVar(&exportFormat, "format", "csv", "Format option: [csv, sls-json]")
	cmd.Flags().BoolVar(&ignoreValidation, "ignore-validation", false, "Skip validating the sls data. This only applies to the sls-json format.")

	return cmd, nil
}

func NewImportCommand() (cmd *cobra.Command, err error) {
	// cmd represents cani alpha import
	cmd = &cobra.Command{}

	return cmd, nil
}
