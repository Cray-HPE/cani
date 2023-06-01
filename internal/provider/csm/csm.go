package csm

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/csm/sls"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
)

type NewOpts struct {
	ImportPath         string
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

// Import external inventory data into CANI's inventory format
func (csm *CSM) Import(ctx context.Context, datastore inventory.Datastore, path string) error {
	log.Debug().Msg("Importing inventory data")
	_, err := csm.importFromPath(path)
	if err != nil {
		return fmt.Errorf("failed to import inventory data from %s: %v", path, err)
	}

	return nil
}

// importFromPath reads from a URL or a file
func (csm *CSM) importFromPath(src string) ([]byte, error) {
	// Parse the string as a URL
	u, err := url.Parse(src)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to parse source string: %v", err)
	}

	var dumpstate sls.Dumpstate
	var dumpstateBytes []byte
	// Check if it's a valid URL (has a scheme like http or https)
	if u.Scheme == "http" || u.Scheme == "https" {
		// Get the dumpate from SLS
		imported, resp, err := csm.slsClient.DumpstateApi.DumpstateGet(context.Background())
		if err != nil {
			return []byte{}, fmt.Errorf("SLS dumpstate failed. %v\n", err)
		}
		if err != nil {
			return []byte{}, fmt.Errorf("failed to get URL: %v", err)
		}
		defer resp.Body.Close()

		dumpstate = sls.Dumpstate{
			Hardware: imported.Hardware,
			Networks: imported.Networks,
		}

		dumpstateBytes, err = json.Marshal(dumpstate)
		if err != nil {
			return []byte{}, fmt.Errorf("failed to marshal dumpstate: %v", err)
		}
	}

	// It's not a URL, treat it as a file path
	dumpstateBytes, err = os.ReadFile(src)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to read file: %v", err)
	}

	return dumpstateBytes, nil
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
