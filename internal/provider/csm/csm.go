package csm

import (
	"context"
	"fmt"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/mitchellh/mapstructure"
)

type NewOpts struct {
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
}

func New(opts *NewOpts) (*CSM, error) {
	csm := &CSM{}
	// Setup HTTP client and context using csm options
	httpClient, ctx, err := opts.newClient()
	if err != nil {
		return nil, err
	}

	slsClientConfiguration := &sls_client.Configuration{
		BasePath:   opts.BaseUrlSLS,
		HTTPClient: httpClient,
		UserAgent:  taxonomy.App,
		DefaultHeader: map[string]string{
			"Content-Type": "application/json",
		},
	}

	hsmClientConfiguration := &hsm_client.Configuration{
		BasePath:   opts.BaseUrlHSM,
		HTTPClient: httpClient,
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
	slsClientConfiguration.Ctx = ctx
	csm.slsClient = sls_client.NewAPIClient(slsClientConfiguration)
	hsmClientConfiguration.Ctx = ctx
	csm.hsmClient = hsm_client.NewAPIClient(hsmClientConfiguration)

	// Load system specific config data
	csm.ValidRoles = opts.ValidRoles
	csm.ValidSubRoles = opts.ValidSubRoles
	return csm, nil
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
	case hardwaretypes.Node:
		// TODO do something interesting with the raw data, and convert it/validate it
		properties := NodeMetadata{} // Create an empty one
		if _, exists := cHardware.ProviderProperties["csm"]; exists {
			// If one exists set it.
			if err := mapstructure.Decode(cHardware.ProviderProperties["csm"], &properties); err != nil {
				return err
			}
		}
		// Make changes to the node metadata
		// The keys of rawProperties need to match what is defined in ./cmd/node/update_node.go
		if roleRaw, exists := rawProperties["role"]; exists {
			if roleRaw == nil {
				properties.Role = nil
			} else {
				properties.Role = StringPtr(roleRaw.(string))
			}
		}
		if subroleRaw, exists := rawProperties["subrole"]; exists {
			if subroleRaw == nil {
				properties.SubRole = nil
			} else {
				properties.SubRole = StringPtr(subroleRaw.(string))
			}
		}
		if nidRaw, exists := rawProperties["nid"]; exists {
			if nidRaw == nil {
				properties.Nid = nil
			} else {
				properties.Nid = IntPtr(nidRaw.(int))
			}
		}
		if aliasRaw, exists := rawProperties["alias"]; exists {
			if aliasRaw == nil {
				properties.Alias = nil
			} else {
				properties.Alias = StringPtr(aliasRaw.(string))
			}
		}

		cHardware.ProviderProperties["csm"] = properties

		return nil
	default:
		// This hardware type doesn't have metadata for it right now
		return nil
	}

}
