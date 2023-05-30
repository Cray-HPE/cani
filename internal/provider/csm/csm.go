package csm

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/provider/csm/validate"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
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

	// Load system specific config data
	csm.ValidRoles = opts.ValidRoles
	csm.ValidSubRoles = opts.ValidSubRoles

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

func joinUUIDs(ids []uuid.UUID, ignoreID uuid.UUID, sep string) string {
	idStrs := []string{}
	for _, id := range ids {
		if id == ignoreID {
			continue
		}

		idStrs = append(idStrs, id.String())
	}

	sort.Strings(idStrs)

	return strings.Join(idStrs, sep)
}

// Validate the representation of the inventory data into the destination inventory system
// is consistent.
// TODO perhaps this should just happen during Reconcile
func (csm *CSM) ValidateInternal(ctx context.Context, datastore inventory.Datastore) (map[uuid.UUID]provider.HardwareValidationResult, error) {
	log.Debug().Msg("Validating datastore contents against the CSM Provider")

	allHardware, err := datastore.List()
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to list hardware from the datastore"),
			err,
		)
	}

	// Build up the validation results map
	results := map[uuid.UUID]provider.HardwareValidationResult{}
	for _, cHardware := range allHardware.Hardware {
		results[cHardware.ID] = provider.HardwareValidationResult{
			Hardware: cHardware,
		}
	}

	validRoles := map[string]bool{}
	for _, role := range csm.ValidRoles {
		validRoles[role] = true
	}
	validSubRoles := map[string]bool{}
	for _, subRole := range csm.ValidSubRoles {
		validSubRoles[subRole] = true
	}

	//
	// Uniques checks
	// The following are checks that can be performed all the time
	// as it just verifies that the data being added is unique.
	// Ideally should be ran as early as possible
	//

	// Verify all specified Node metadata is valid
	nodeNIDLookup := map[int][]uuid.UUID{}
	nodeAliasLookup := map[string][]uuid.UUID{}
	for _, cHardware := range allHardware.Hardware {
		if cHardware.Type != hardwaretypes.Node {
			continue
		}

		log.Debug().Msgf("Validating %s: %v", cHardware.ID, cHardware)

		metadata, err := GetProviderMetadataT[NodeMetadata](cHardware)
		if err != nil {
			return nil, errors.Join(
				fmt.Errorf("failed to get provider metadata from hardware (%s)", cHardware.ID),
				err,
			)
		}

		// There is no metadata for this node
		if metadata == nil {
			log.Debug().Msgf("No metadata found for %s", cHardware.ID)
			continue
		}

		validationResult := results[cHardware.ID]

		// Verify all specified Roles are valid
		if metadata.Role != nil {
			if !validRoles[*metadata.Role] {
				validationResult.Errors = append(validationResult.Errors,
					fmt.Sprintf("Specified role (%s) is invalid, choose from: %s", *metadata.Role, strings.Join(csm.ValidRoles, ", ")),
				)
			}
		}

		// Verify all specified SubRoles are valid
		if metadata.SubRole != nil {
			if !validRoles[*metadata.SubRole] {
				validationResult.Errors = append(validationResult.Errors,
					fmt.Sprintf("Specified sub-role (%s) is invalid, choose from: %s", *metadata.SubRole, strings.Join(csm.ValidSubRoles, ", ")),
				)
			}
		}

		// Verify NID is valid
		if metadata.Nid != nil {
			nodeNIDLookup[*metadata.Nid] = append(nodeNIDLookup[*metadata.Nid], cHardware.ID)
			if *metadata.Nid <= 0 {
				validationResult.Errors = append(validationResult.Errors,
					fmt.Sprintf("Specified NID (%d) invalid, needs to be positive integer", *metadata.Nid),
				)
			}
		}

		// Verify Alias is valid
		if metadata.Alias != nil {
			nodeAliasLookup[*metadata.Alias] = append(nodeAliasLookup[*metadata.Alias], cHardware.ID)

			// TODO a regex here might be better
			if strings.Contains(*metadata.Alias, " ") {
				validationResult.Errors = append(validationResult.Errors,
					fmt.Sprintf("Specified alias (%d) is invalid, alias contains spaces", *metadata.Nid),
				)
			}
		}

		results[cHardware.ID] = validationResult
	}

	// Verify all specified NIDs are unique
	for nid, matchingHardware := range nodeNIDLookup {
		if len(matchingHardware) > 1 {
			// We found hardware with duplicate NIDs
			for _, id := range matchingHardware {
				validationResult := results[id]
				validationResult.Errors = append(validationResult.Errors,
					fmt.Sprintf("Specified NID (%d) is not unique, shared by: %s", nid, joinUUIDs(matchingHardware, id, ", ")),
				)
				results[id] = validationResult
			}
		}
	}

	// Verify all specified Aliases are unique
	for alias, matchingHardware := range nodeAliasLookup {
		if len(matchingHardware) > 1 {
			// We found hardware with duplicate NIDs
			for _, id := range matchingHardware {
				validationResult := results[id]
				validationResult.Errors = append(validationResult.Errors,
					fmt.Sprintf("Specified alias (%s) is not unique, shared by: %s", alias, joinUUIDs(matchingHardware, id, ", ")),
				)
				results[id] = validationResult
			}
		}
	}

	//
	// Missing data checks
	// These checks should be ran at reconcile time or via a command line options
	// to ensure all of the required data is present in the datastore before
	//

	// TODO

	//
	// Build results
	//
	resultsWithErrors := map[uuid.UUID]provider.HardwareValidationResult{}
	for id, result := range results {
		if len(result.Errors) > 0 {
			resultsWithErrors[id] = result
		}
	}

	if len(resultsWithErrors) > 0 {
		return resultsWithErrors, provider.ErrDataValidationFailure
	}

	return nil, nil
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
