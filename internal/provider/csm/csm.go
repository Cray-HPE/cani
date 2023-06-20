package csm

import (
	"fmt"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"

	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

type NewOpts struct {
	UseSimulation      bool
	InsecureSkipVerify bool
	APIGatewayToken    string
	BaseUrlSLS         string
	BaseUrlHSM         string
	SecretName         string
	KubeConfig         string
	ClientID           string `json:"-" yaml:"-"` // omit credentials from cani.yml
	ClientSecret       string `json:"-" yaml:"-"` // omit credentials from cani.yml
	TokenHost          string
	TokenUsername      string `json:"-" yaml:"-"` // omit credentials from cani.yml
	TokenPassword      string `json:"-" yaml:"-"` // omit credentials from cani.yml
	CaCertPath         string
	ValidRoles         []string
	ValidSubRoles      []string
}

var DefaultValidRoles = []string{
	"Compute",
	"Service",
	"System",
	"Application",
	"Storage",
	"Management",
}
var DefaultValidSubRolesRoles = []string{
	"Worker",
	"Master",
	"Storage",
	"UAN",
	"Gateway",
	"LNETRouter",
	"Visualization",
	"UserDefined",
}

type CSM struct {
	// Clients
	slsClient *sls_client.APIClient
	hsmClient *hsm_client.APIClient
	// System Configuration data
	ValidRoles    []string
	ValidSubRoles []string

	hardwareLibrary *hardwaretypes.Library
}

func New(opts *NewOpts, hardwareLibrary *hardwaretypes.Library) (*CSM, error) {
	csm := &CSM{
		hardwareLibrary: hardwareLibrary,
	}

	// Setup HTTP client and context using csm options
	httpClient, _, err := opts.newClient()
	if err != nil {
		return nil, err
	}

	if opts.UseSimulation {
		if opts.BaseUrlSLS == "" {
			opts.BaseUrlSLS = "https://localhost:8443/apis/sls/v1"
		}
		if opts.BaseUrlHSM == "" {
			opts.BaseUrlHSM = "https://localhost:8443/apis/smd/hsm/v2"
		}
	}

	slsClientConfiguration := &sls_client.Configuration{
		BasePath:   opts.BaseUrlSLS,
		HTTPClient: httpClient.StandardClient(),
		UserAgent:  taxonomy.App,
		DefaultHeader: map[string]string{
			"Content-Type": "application/json",
		},
	}

	hsmClientConfiguration := &hsm_client.Configuration{
		BasePath:   opts.BaseUrlHSM,
		HTTPClient: httpClient.StandardClient(),
		UserAgent:  taxonomy.App,
		DefaultHeader: map[string]string{
			"Content-Type": "application/json",
		},
	}

	if opts.APIGatewayToken != "" {
		// Set the token for use in the clients
		slsClientConfiguration.DefaultHeader["Authorization"] = fmt.Sprintf("Bearer %s", opts.APIGatewayToken)
		hsmClientConfiguration.DefaultHeader["Authorization"] = fmt.Sprintf("Bearer %s", opts.APIGatewayToken)
	}

	// Set the clients
	csm.slsClient = sls_client.NewAPIClient(slsClientConfiguration)
	csm.hsmClient = hsm_client.NewAPIClient(hsmClientConfiguration)

	// Load system specific config data
	csm.ValidRoles = opts.ValidRoles
	csm.ValidSubRoles = opts.ValidSubRoles
	return csm, nil
}
