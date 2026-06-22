package csm

import "github.com/Cray-HPE/cani/internal/cli"

// Options holds the provider's configuration options.
// These are written to the config file with YAML comments preserved.
type Options struct {
	// ProviderHost is the FQDN or host:port of the API gateway.
	ProviderHost string `yaml:"provider_host" head_comment:"FQDN or host:port of the API gateway"`
	// InsecureSkipVerify disables TLS certificate verification.
	InsecureSkipVerify bool `yaml:"insecure" head_comment:"Disable TLS certificate verification"`
	// CaCertPath is an optional path to a PEM-encoded CA certificate.
	CaCertPath string `yaml:"ca_cert" head_comment:"Path to PEM-encoded CA certificate"`
	// UseSimulation enables simulation mode.
	UseSimulation bool `yaml:"use_simulation" head_comment:"Use simulation mode (localhost:8443, no auth)"`
	// K8sPodsCidr is the CIDR used by Kubernetes for pods.
	K8sPodsCidr string `yaml:"k8s_pods_cidr" head_comment:"CIDR used by Kubernetes for pods"`
	// K8sServicesCidr is the CIDR used by Kubernetes for services.
	K8sServicesCidr string `yaml:"k8s_services_cidr" head_comment:"CIDR used by Kubernetes for services"`
	// SecretName is the Kubernetes secret name for auth credentials.
	SecretName string `yaml:"secret_name" head_comment:"Kubernetes secret name for auth credentials"`
}

// ImportOptions holds import-specific configuration.
// These options correspond to CLI flags for the import command.
type ImportOptions struct {
	SlsFile string `yaml:"sls_file" head_comment:"Path to SLS dumpstate JSON file"`
	SmdFile string `yaml:"smd_file" head_comment:"Path to SMD state components JSON file"`
	// TokenUsername is the Keycloak username for password-credentials auth.
	TokenUsername string `yaml:"-"`
	// TokenPassword is the Keycloak password for password-credentials auth.
	TokenPassword string `yaml:"-"`
	// APIGatewayToken is a pre-existing bearer token. If set, Keycloak auth is skipped.
	APIGatewayToken string `yaml:"token"`
	// ClientID is the OAuth2 client ID (hidden, alternative auth).
	ClientID string `yaml:"-"`
	// ClientSecret is the OAuth2 client secret (hidden, alternative auth).
	ClientSecret string `yaml:"-"`
}

// ExportOptions holds export-specific configuration.
// These options correspond to CLI flags for the export command.
type ExportOptions struct {
	// TokenUsername is the Keycloak username for password-credentials auth.
	TokenUsername string `yaml:"-"`
	// TokenPassword is the Keycloak password for password-credentials auth.
	TokenPassword string `yaml:"-"`
	// APIGatewayToken is a pre-existing bearer token. If set, Keycloak auth is skipped.
	APIGatewayToken string `yaml:"token"`
	// ClientID is the OAuth2 client ID (hidden, alternative auth).
	ClientID string `yaml:"-"`
	// ClientSecret is the OAuth2 client secret (hidden, alternative auth).
	ClientSecret string `yaml:"-"`
	// Commit pushes changes to the remote system.
	Commit bool `yaml:"-"`
}

// --- HasOptions interface implementation ---

// GetDefaultOptions returns the default configuration options for this provider.
// These are auto-populated in the config file if they do not exist.
func (p *Csm) GetDefaultOptions() map[string]any {
	return map[string]any{
		"provider_host":     "",
		"insecure":          false,
		"ca_cert":           "",
		"use_simulation":    false,
		"k8s_pods_cidr":     "10.32.0.0/12",
		"k8s_services_cidr": "10.16.0.0/12",
		"secret_name":       "admin-client-auth",
	}
}

// GetOptionsStruct returns the configuration struct for comment extraction.
// This enables YAML serialization with preserved field ordering and comments.
func (p *Csm) GetOptionsStruct() interface{} {
	return &Options{}
}

// --- HasImportOptions interface implementation ---

// GetImportOptionsStruct returns the import options struct for reflection.
func (p *Csm) GetImportOptionsStruct() interface{} {
	return map[string]any{}
}

// GetImportDefaults returns default import configuration options.
func (p *Csm) GetImportDefaults() map[string]any {
	return map[string]any{
		"sls_file": "",
		"smd_file": "",
	}
}

// BindImportFlags binds CLI flags to Viper for the import command.
// This enables precedence: CLI flags > env vars > config file > defaults.
func (p *Csm) BindImportFlags(cmd *cli.Command) error {
	// Flags are read directly in the import package via cmd.Flags().GetString()
	return nil
}

// --- HasExportOptions interface implementation ---

// GetExportOptionsStruct returns the export options struct for reflection.
func (p *Csm) GetExportOptionsStruct() interface{} {
	return map[string]any{}
}

// GetExportDefaults returns default export configuration options.
func (p *Csm) GetExportDefaults() map[string]any {
	return map[string]any{}
}

// BindExportFlags binds CLI flags to Viper for the export command.
// This enables precedence: CLI flags > env vars > config file > defaults.
func (p *Csm) BindExportFlags(cmd *cli.Command) error {
	// Auth flags are read directly via cmd.Flags() in the export package.
	return nil
}
