package netbox

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/caarlos0/env"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/joho/godotenv"
	"github.com/netbox-community/go-netbox/v3"
	"github.com/rs/zerolog/log"
)

// ImportEnvs is the struct that holds the environment variables
// This maintains parity/compatiblity with Device-Type-Librery-Import python repo
type ImportEnvs struct {
	NetboxUrl       string `env:"NETBOX_URL"`
	NetboxToken     string `env:"NETBOX_TOKEN"`
	RepoUrl         string `env:"REPO_URL"`
	RepoBranch      string `env:"REPO_BRANCH"`
	IgnoreSslErrors bool   `env:"IGNORE_SSL_ERRORS"`
	Slugs           string `env:"SLUGS"`
}

// LoadEnvs loads the environment variables from the .env file and parses them into a struct
// This maintains parity/compatiblity with Device-Type-Librery-Import python repo
func LoadEnvs() (ImportEnvs, error) {
	// Loading the environment variables from '.env' file.
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Msgf("unable to load .env file: %e", err)
	}

	// create a config object and parse the environment variables into it
	cfg := ImportEnvs{}
	err = env.Parse(&cfg)
	if err != nil {
		log.Fatal().Msgf("unable to parse environment variables: %e", err)
	}

	return cfg, nil
}

// NewClient creates a new netbox client using the environment variables
// This maintains parity/compatiblity with Device-Type-Librery-Import python repo
func (ie ImportEnvs) NewClient() (*netbox.APIClient, context.Context, error) {
	// use the certificates in the http client
	tlsConfig := &tls.Config{
		InsecureSkipVerify: ie.IgnoreSslErrors,
	}

	// use TLS config in transport
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Setup our HTTP transport and client
	httpClient := retryablehttp.NewClient()
	httpClient.HTTPClient.Transport = tr
	httpClient.Logger = nil
	c := httpClient.StandardClient()

	// create the netbox config
	token := ie.NetboxToken
	host := stripProtocolsAndSpecialChars(ie.NetboxUrl)
	nbcfg := netbox.NewConfiguration()
	nbcfg.Host = host
	nbcfg.HTTPClient = c
	nbcfg.DefaultHeader["Authorization"] = fmt.Sprintf("Token %s", token)
	nbcfg.Debug = false
	nbcfg.Scheme = "http"

	ctx := context.Background()
	client := netbox.NewAPIClient(nbcfg)

	return client, ctx, nil
}

func CreateManufacturers(client *netbox.APIClient, ctx context.Context, manufacturersToImport map[string]string) error {
	// get existing manufacturers
	req := client.DcimAPI.DcimManufacturersList(ctx).Limit(int32(9999)).Offset(int32(0))
	existingManufacturers, resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("failed to get existing manufacturers: %+v: %+v", resp.Status, err)
	}

	log.Info().Msgf("Found %+v importable manufacturers ", len(manufacturersToImport))
	log.Info().Msgf("Found %+v existing manufacturers", len(existingManufacturers.Results))

	// // FIXME: parse next url for offset and create a new req
	// if existingManufacturers.Next.Get() != nil {
	// 	log.Warn().Msgf("There are more manufacturers than the API is returning")
	// 	log.Warn().Msgf("Getting next not yet implemented")
	// }

	// as long as there are existing manufacturers,
	// we need to check if the manufacturers to import already exists to avoid duplicates
	if *existingManufacturers.Count != 0 {
		for _, existingManufacturer := range existingManufacturers.Results {
			// if the manufacturer does not exist in the ones that need to be imported, add it to the map
			_, ok := manufacturersToImport[strings.ToLower(existingManufacturer.Name)]
			if ok {
				log.Debug().Msgf("Existing manufacturer will not be re-created: %+v (%s)", existingManufacturer.Name, existingManufacturer.Slug)
				delete(manufacturersToImport, strings.ToLower(existingManufacturer.Name))
			}
		}
	}

	// loop through manufacturers that do not exist and create them
	for key, name := range manufacturersToImport {
		log.Info().Msgf("Creating manufacturer: %+v (%+v)", key, name)
		err = createManufacturer(client, ctx, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateDeviceTypes(client *netbox.APIClient, ctx context.Context, deviceTypesToImport map[string]DeviceType) error {
	const maxLimit = 1000 // 1000 is the max limit for this endpoint
	var existingDeviceTypes = []netbox.DeviceType{}

	limit := maxLimit
	offset := 0
	for {
		req := client.DcimAPI.DcimDeviceTypesList(ctx).Limit(int32(limit)).Offset(int32(offset))
		paginatedDeviceTypes, resp, err := req.Execute()
		if err != nil {
			return fmt.Errorf("failed to get existing DeviceTypes: %+v: %+v", resp.Status, err)
		}

		existingDeviceTypes = append(existingDeviceTypes, paginatedDeviceTypes.Results...)

		// no more pages to get
		if paginatedDeviceTypes.Next.Get() == nil {
			break
		}

		var next *url.URL
		next, err = url.Parse(paginatedDeviceTypes.GetNext())
		if err != nil {
			return err
		}
		limitQ := next.Query().Get("limit")
		offsetQ := next.Query().Get("offset")
		limit, err = strconv.Atoi(limitQ)
		if err != nil {
			return err
		}
		offset, err = strconv.Atoi(offsetQ)
		if err != nil {
			return err
		}
	}

	log.Info().Msgf("Found %+v importable DeviceTypes", len(deviceTypesToImport))
	log.Info().Msgf("Found %+v existing DeviceTypes", len(existingDeviceTypes))

	// as long as there are existing manufacturers,
	// we need to check if the manufacturers to import already exists to avoid duplicates
	if len(existingDeviceTypes) != 0 {
		for _, existingDeviceType := range existingDeviceTypes {
			// if the manufacturer does not exist in the ones that need to be imported, add it to the map
			_, ok := deviceTypesToImport[existingDeviceType.Model]
			if ok {
				log.Debug().Msgf("Existing DeviceType will not be re-created: %+v (%s)", existingDeviceType.Model, existingDeviceType.GetPartNumber())
				delete(deviceTypesToImport, existingDeviceType.Model)
			}
		}
	}

	// loop through devicetypes that do not exist and create them
	for _, deviceType := range deviceTypesToImport {
		log.Info().Msgf("Creating DeviceType: %+v (%+v)", deviceType.Model, deviceType.PartNumber)
		err := createDeviceType(client, ctx, deviceType)
		if err != nil {
			return err
		}
	}

	return nil
}
