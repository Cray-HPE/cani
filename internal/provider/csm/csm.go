package csm

import (
	"crypto/tls"
	"fmt"
	"net/http"

	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
)

type NewOpts struct {
	InsecureSkipVerify bool
	APIGatewayToken    string

	BaseUrlSLS string
	BaseUrlHSM string
}

type CSM struct {

	// Clients
	slsClient *sls_client.APIClient
	hsmClient *hsm_client.APIClient
	BladeMeta *BladeMeta
}

type BladeMeta struct {
	Role                 string
	SubRole              string
	Nid                  string
	Alias                string
	AdditionalProperties map[string]interface{}
}

func New(opts NewOpts) (*CSM, error) {
	csm := &CSM{}

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
	return csm, nil
}

// Validate the external services of the inventory provider are correct
func (csm *CSM) ValidateExternal() error {
	log.Warn().Msg("CSM Provider's ValidateExternal was called. This is not currently implemented")
	return nil
}

// Validate the representation of the inventory data into the destination inventory system
// is consistent.
// TODO perhaps this should just happen during Reconcile
func (csm *CSM) ValidateInternal() error {
	log.Warn().Msg("CSM Provider's ValidateInternal was called. This is not currently implemented")

	return nil
}

// Import external inventory data into CANI's inventory format
func (csm *CSM) Import() error {
	return fmt.Errorf("todo")

}
