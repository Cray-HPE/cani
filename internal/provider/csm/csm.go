package csm

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/hashicorp/go-retryablehttp"
)

type NewOpts struct {
	InsecureSkipVerify bool
	APIGatewayToken    string

	BaseUrlSLS string
	BaseUrlHSM string

	ValidRoles    []string
	ValidSubRoles []string
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

func New(opts NewOpts, hardwareLibrary *hardwaretypes.Library) (*CSM, error) {
	csm := &CSM{
		hardwareLibrary: hardwareLibrary,
	}

	//
	// Create Clients
	//

	// Setup HTTP client
	httpClient := retryablehttp.NewClient()
	httpClient.HTTPClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: opts.InsecureSkipVerify},
	}

	slsClientConfiguration := &sls_client.Configuration{
		BasePath:   opts.BaseUrlSLS,
		HTTPClient: httpClient.StandardClient(),
		UserAgent:  "cani",
		DefaultHeader: map[string]string{
			"Content-Type": "application/json",
		},
	}

	hsmClientConfiguration := &hsm_client.Configuration{
		BasePath:   opts.BaseUrlHSM,
		HTTPClient: httpClient.StandardClient(),
		UserAgent:  "cani",
		DefaultHeader: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// Add in the API token if provided
	if opts.APIGatewayToken != "" {
		slsClientConfiguration.DefaultHeader["Authorization"] = fmt.Sprintf("Bearer %s", opts.APIGatewayToken)
		hsmClientConfiguration.DefaultHeader["Authorization"] = fmt.Sprintf("Bearer %s", opts.APIGatewayToken)
	}

	csm.slsClient = sls_client.NewAPIClient(slsClientConfiguration)
	csm.hsmClient = hsm_client.NewAPIClient(hsmClientConfiguration)

	// Load system specific config data
	csm.ValidRoles = opts.ValidRoles
	csm.ValidSubRoles = opts.ValidSubRoles

	return csm, nil
}
