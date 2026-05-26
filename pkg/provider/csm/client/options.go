package client

// Options holds connection and authentication settings for CSM APIs.
type Options struct {
	// ProviderHost is the FQDN or host:port of the API gateway.
	ProviderHost string
	// InsecureSkipVerify disables TLS certificate verification.
	InsecureSkipVerify bool
	// CaCertPath is an optional path to a PEM-encoded CA certificate.
	CaCertPath string
	// TokenUsername is the Keycloak username for password-credentials auth.
	TokenUsername string
	// TokenPassword is the Keycloak password for password-credentials auth.
	TokenPassword string
	// APIGatewayToken is a pre-existing bearer token. If set, Keycloak
	// auth is skipped.
	APIGatewayToken string
	// ClientID is the OAuth2 client ID (alternative auth method).
	ClientID string
	// ClientSecret is the OAuth2 client secret (alternative auth method).
	ClientSecret string
	// BaseURLSLS overrides the default SLS API base URL.
	BaseURLSLS string
	// BaseURLHSM overrides the default HSM API base URL.
	BaseURLHSM string
	// K8sPodsCidr is the CIDR used by Kubernetes for pods.
	K8sPodsCidr string
	// K8sServicesCidr is the CIDR used by Kubernetes for services.
	K8sServicesCidr string
	// KubeConfig is the path to the Kubernetes config file.
	KubeConfig string
	// SecretName is the Kubernetes secret name for auth credentials.
	SecretName string
	// UseSimulation enables simulation mode: sets ProviderHost to
	// localhost:8443, skips auth, and forces InsecureSkipVerify.
	UseSimulation bool
}

// applyDefaults fills in computed fields from the provided options.
func (o *Options) applyDefaults() {
	if o.UseSimulation {
		o.InsecureSkipVerify = true
		if o.ProviderHost == "" || o.ProviderHost == "api-gw-service.local" {
			o.ProviderHost = "localhost:8443"
		}
	}

	if o.BaseURLSLS == "" {
		o.BaseURLSLS = "https://" + o.ProviderHost + "/apis/sls/v1"
	}
	if o.BaseURLHSM == "" {
		o.BaseURLHSM = "https://" + o.ProviderHost + "/apis/smd/hsm/v2"
	}
}
