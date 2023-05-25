package csm

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/csm/validate"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mitchellh/mapstructure"
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
func (csm *CSM) ValidateExternal(ctx context.Context) error {
	// Get the dumpate from SLS
	slsState, reps, err := csm.slsClient.DumpstateApi.DumpstateGet(context.Background())
	if err != nil {
		return fmt.Errorf("SLS dumpstate failed. %v\n", err)
	}

	// Validate the dumpstate returned from SLS
	err = validate.Validate(&slsState, reps)
	if err != nil {
		return fmt.Errorf("Validation failed. %v\n", err)
	}
	return nil
}

// Validate the representation of the inventory data into the destination inventory system
// is consistent.
// TODO perhaps this should just happen during Reconcile
func (csm *CSM) ValidateInternal(ctx context.Context) error {
	log.Warn().Msg("CSM Provider's ValidateInternal was called. This is not currently implemented")

	return nil
}

// Import external inventory data into CANI's inventory format
func (csm *CSM) Import(ctx context.Context, datastore inventory.Datastore) error {
	return fmt.Errorf("todo")

}

func (csm *CSM) BuildHardwareMetadata(cHardware *inventory.Hardware, rawProperties map[string]interface{}) error {
	if cHardware.ProviderProperties == nil {
		cHardware.ProviderProperties = map[string]interface{}{}
	}

	switch cHardware.Type {
	case hardwaretypes.HardwareTypeNode:
		// TODO do something interesting with the raw data, and convert it/validate it
		properties := NodeMetadata{} // Create an empty one
		if _, exists := cHardware.ProviderProperties["csm"]; exists {
			// If one exists set it.
			// TODO Depending on how the data is stored/unmarshalled this might be a map[string]interface{}, so using the mapstructure library might be required to get it into the struct form
			// https://github.com/Cray-HPE/cani/blob/develop/internal/provider/csm/sls/hardware.go
			// https://github.com/mitchellh/mapstructure

			if err := mapstructure.Decode(cHardware.ProviderProperties["csm"], &properties); err != nil {
				return err
			}
		}
		// Make changes to the node metadata
		// The keys of rawProperties need to match what is defined in ./cmd/node/update_node.go
		if roleRaw, exists := rawProperties["role"]; exists {
			properties.Role = StringPtr(roleRaw.(string))
		}
		if subroleRaw, exists := rawProperties["subrole"]; exists {
			properties.SubRole = StringPtr(subroleRaw.(string))
		}
		if nidRaw, exists := rawProperties["nid"]; exists {
			properties.Nid = IntPtr(nidRaw.(int))
		}
		if aliasRaw, exists := rawProperties["alias"]; exists {
			properties.Alias = StringPtr(aliasRaw.(string))
		}

		cHardware.ProviderProperties["csm"] = properties

		return nil
	default:
		// This hardware type doesn't have metadata for it right now
		return nil
	}

}
