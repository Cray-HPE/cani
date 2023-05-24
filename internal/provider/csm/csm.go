package csm

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
)

// NewOpts defines the options for creating a new CSM provider
type NewOpts struct {
	InsecureSkipVerify bool
	APIGatewayToken    string

	BaseUrlSLS string
	BaseUrlHSM string
}

// CSM is the CSM provider
type CSM struct {
	// Clients
	slsClient *sls_client.APIClient
	hsmClient *hsm_client.APIClient
}

// NodeMetadata is the metadata for a node, required by SLS
type NodeMetadata struct {
	Role                 string
	SubRole              string
	Nid                  string
	Alias                string
	AdditionalProperties map[string]interface{}
}

// New returns a new CSM provider using the provided options
func New(opts NewOpts) (*CSM, error) {
	// Create the CSM provider
	csm := &CSM{}

	// Setup HTTP client
	httpClient := retryablehttp.NewClient()
	httpClient.HTTPClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: opts.InsecureSkipVerify},
	}

	// Setup SLS client
	slsClientConfiguration := &sls_client.Configuration{
		BasePath:   opts.BaseUrlSLS,
		HTTPClient: httpClient.StandardClient(),
		UserAgent:  "cani",
		DefaultHeader: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// Setup HSM client
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

	// Instantiate the clients
	csm.slsClient = sls_client.NewAPIClient(slsClientConfiguration)
	csm.hsmClient = hsm_client.NewAPIClient(hsmClientConfiguration)
	return csm, nil
}

// ValidateExternal validates the external services of the inventory provider are correct
func (csm *CSM) ValidateExternal() error {
	log.Warn().Msg("CSM Provider's ValidateExternal was called. This is not currently implemented")
	return nil
}

// ValidateInternal valiates the representation of the inventory data into the destination inventory system is consistent.
// TODO perhaps this should just happen during Reconcile
func (csm *CSM) ValidateInternal() error {
	log.Warn().Msg("CSM Provider's ValidateInternal was called. This is not currently implemented")

	return nil
}

// Import imports external inventory data into CANI's inventory format
func (csm *CSM) Import() error {
	return fmt.Errorf("todo")

}

// BuildHardwareMetadata
func (csm *CSM) BuildHardwareMetadata(hw *inventory.Hardware, rawProperties map[string]interface{}) error {
	switch hw.Type {
	case hardwaretypes.HardwareTypeNode:
		// TODO do something interesting with the raw data, and convert it/validate it
		properties := NodeMetadata{} // Create an empty one
		if _, exists := hw.ProviderProperties["csm"]; exists {
			// If one exists set it.
			// TODO Depending on how the data is stored/unmarshalled this might be a map[string]interface{}, so using the mapstructure library might be required to get it into the struct form
			// https://github.com/Cray-HPE/cani/blob/develop/internal/provider/csm/sls/hardware.go
			// https://github.com/mitchellh/mapstructure

			properties = hw.ProviderProperties["csm"].(NodeMetadata)
		}

		if role, exists := rawProperties["role"]; exists {
			properties.Role = role.(string)
		}

		hw.ProviderProperties["csm"] = properties

		return nil
	}

	return fmt.Errorf("todo")
}
