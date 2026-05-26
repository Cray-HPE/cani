package commands

import "github.com/spf13/cobra"

// addAuthFlags adds CSM authentication flags common to import and export.
func addAuthFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("use-simulator", "S", false,
		"Use simulation mode (localhost:8443, no auth, skip TLS verification)")
	cmd.Flags().BoolP("insecure", "k", false,
		"Skip TLS certificate verification")
	cmd.Flags().String("csm-api-host", "api-gw-service.local",
		"FQDN or host:port of the CSM API gateway")
	cmd.Flags().String("csm-keycloak-username", "",
		"Keycloak username for authentication")
	cmd.Flags().String("csm-keycloak-password", "",
		"Keycloak password for authentication")
	cmd.Flags().String("csm-url-sls", "",
		"Override the SLS API base URL")
	cmd.Flags().String("csm-url-hsm", "",
		"Override the HSM API base URL")
	cmd.Flags().String("csm-ca-cert", "",
		"Path to a PEM-encoded CA certificate")
	cmd.Flags().String("csm-k8s-pods-cidr", "10.32.0.0/12",
		"CIDR used by Kubernetes for pods")
	cmd.Flags().String("csm-k8s-services-cidr", "10.16.0.0/12",
		"CIDR used by Kubernetes for services")

	// Hidden flags for less-common auth methods
	cmd.Flags().String("csm-kube-config", "",
		"Path to the Kubernetes config file")
	cmd.Flags().String("csm-secret-name", "admin-client-auth",
		"Kubernetes secret name for auth credentials")
	cmd.Flags().String("csm-client-id", "",
		"OAuth2 client ID")
	cmd.Flags().String("csm-client-secret", "",
		"OAuth2 client secret")
	_ = cmd.Flags().MarkHidden("csm-kube-config")
	_ = cmd.Flags().MarkHidden("csm-secret-name")
	_ = cmd.Flags().MarkHidden("csm-client-id")
	_ = cmd.Flags().MarkHidden("csm-client-secret")

	cmd.MarkFlagsRequiredTogether("csm-keycloak-username", "csm-keycloak-password", "csm-api-host")
}

func NewImportCommand(base *cobra.Command) (*cobra.Command, error) {
	cmd := &cobra.Command{}
	cmd.Flags().String("sls-file", "", "Path to SLS dumpstate JSON file")
	cmd.Flags().String("smd-file", "", "Path to SMD state components JSON file")
	cmd.Flags().Bool("ignore-validation", false,
		"Continue importing even if the external inventory fails validation")
	addAuthFlags(cmd)
	return cmd, nil
}

func NewExportCommand(base *cobra.Command) (*cobra.Command, error) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("commit", false, "Push changes to the remote CSM system")
	cmd.Flags().Bool("dryrun", false, "Perform a dry run without applying changes")
	cmd.Flags().String("headers", "Type,Vlan,Role,SubRole,Status,Nid,Alias,Name,ID",
		"Comma-separated list of CSV columns to include")
	cmd.Flags().StringP("type", "t", "Node,Cabinet",
		"Comma-separated list of hardware types to include")
	cmd.Flags().BoolP("all", "a", false,
		"Include all hardware types (overrides --type)")
	cmd.Flags().String("format", "csv",
		"Output format: csv, sls-json")
	cmd.Flags().Bool("ignore-validation", false,
		"Skip validation (only applies to sls-json format)")
	addAuthFlags(cmd)
	return cmd, nil
}

func NewShowCommand(base *cobra.Command) (*cobra.Command, error) {
	// TODO: Add show-specific flags or subcommands
	return &cobra.Command{}, nil
}

func NewAddCommand(base *cobra.Command) (*cobra.Command, error) {
	// TODO: Add add-specific flags or subcommands
	return nil, nil
}

func NewRemoveCommand(base *cobra.Command) (*cobra.Command, error) {
	// TODO: Add remove-specific flags or subcommands
	return &cobra.Command{}, nil
}

func NewUpdateCommand(base *cobra.Command) (*cobra.Command, error) {
	// Provider-specific logic belongs in import/export only.
	return nil, nil
}
