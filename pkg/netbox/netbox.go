package netbox

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/netbox-community/go-netbox/v3"
	"github.com/rs/zerolog/log"
)

// docker compose -f ../../csminv/netbox-docker/docker-compose.yml down --remove-orphans
// docker compose -f ../../csminv/netbox-docker/docker-compose.yml up -d
// docker volume remove netbox-docker_netbox-postgres-data

func createManufacturer(client *netbox.APIClient, ctx context.Context, manufacturer string) error {
	m := netbox.NewManufacturerRequest(manufacturer, stringToSlug(manufacturer))
	m.AdditionalProperties = make(map[string]interface{}, 0)
	m.CustomFields = make(map[string]interface{}, 0)
	// tags := []netbox.NestedTagRequest{}
	// m.SetTags(tags))
	mfg := client.DcimAPI.DcimManufacturersCreate(ctx).ManufacturerRequest(*m)
	_, resp, err := mfg.Execute()
	if err != nil {
		if resp.StatusCode == 201 {
			return nil
		}

		if resp.StatusCode == 400 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			log.Error().Msgf("%+v", string(body))
			return err
		}
		return err
	}
	return nil
}

func (dt DeviceType) getManufacturer(client *netbox.APIClient, ctx context.Context) (netbox.Manufacturer, error) {
	existingManufacturers, err := getExistingVendors(client, ctx)
	if err != nil {
		return netbox.Manufacturer{}, err
	}
	for _, existingManufacturer := range existingManufacturers {
		if strings.ToLower(existingManufacturer.Name) == strings.ToLower(dt.Manufacturer) {
			return existingManufacturer, nil
		}
	}
	return netbox.Manufacturer{}, nil
}

func createDeviceType(client *netbox.APIClient, ctx context.Context, dt DeviceType) error {
	req := netbox.WritableDeviceTypeRequest{}
	// required:
	//   - manufacturer
	mf, err := dt.getManufacturer(client, ctx)
	if err != nil {
		return err
	}
	req.SetManufacturer(mf.Id)
	//   - model
	req.SetModel(dt.Model)
	//   - slug
	req.SetSlug(stringToSlug(dt.Model))
	//	 - u_height
	if dt.UHeight != nil {
		req.SetUHeight(*dt.UHeight)
	}

	// optional:
	if dt.Comments != nil {
		req.SetComments(*dt.Comments)
	}
	if dt.PartNumber != nil {
		req.SetPartNumber(*dt.PartNumber)
	}
	if dt.IsFullDepth != nil {
		req.SetIsFullDepth(*dt.IsFullDepth)
	}
	// unit and weight required together
	if dt.WeightUnit != nil {
		wu := netbox.DeviceTypeWeightUnitValue(*dt.WeightUnit)
		req.SetWeightUnit(wu)
		if dt.Weight != nil {
			req.SetWeight(*dt.Weight)
		}
	}

	req.AdditionalProperties = make(map[string]interface{}, 0)
	req.CustomFields = make(map[string]interface{}, 0)
	// req.SetFrontImage(dt.FrontImage)
	// req.SetRearImage(dt.RearImage)
	// req.SetDefaultPlatform()
	// req.SetWeightUnit()
	// req.SetSubdeviceRole(*dt.SubdeviceRole)
	// req.SetTags()
	// req.SetAirflow(dt.Airflow)

	d := client.DcimAPI.DcimDeviceTypesCreate(ctx).WritableDeviceTypeRequest(req)
	_, resp, err := d.Execute()
	if err != nil {

		// still created
		if resp.StatusCode == 201 {
			// no value given for required property device_count
			return nil
		}

		if resp.StatusCode == 400 {
			log.Warn().Msgf("DeviceType already exists (ignoring): %+v", dt.Model)
			return nil
		}

		return err
	}

	return nil
}

func getExistingVendors(client *netbox.APIClient, ctx context.Context) ([]netbox.Manufacturer, error) {
	req := client.DcimAPI.DcimManufacturersList(ctx).Limit(int32(9999)).Offset(int32(0))
	existingManufacturers, resp, err := req.Execute()
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to get existing manufacturers: %+v", resp.Status), err)
	}

	return existingManufacturers.Results, nil
}

func CreateDeviceTypeMap(manufacturersToImport map[string]string, deviceTypesToImport map[string]DeviceType, ignoreDirs []string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print(err)
			return nil
		}
		if info.IsDir() {
			dir := filepath.Base(path)
			for _, d := range ignoreDirs {
				if d == dir {
					return filepath.SkipDir
				}
			}
		} else {
			dt, err := UnmarshalDeviceType(path)
			if err != nil {
				return err
			}

			// add the manufacturer to the map
			// it is required to add a devicetype
			sanitized := strings.ToLower(dt.Manufacturer)
			_, manufacturerExists := manufacturersToImport[sanitized]
			if !manufacturerExists {
				manufacturersToImport[sanitized] = dt.Manufacturer
			}

			// add the device type to the map
			_, deviceTypeExists := deviceTypesToImport[dt.Model]
			if !deviceTypeExists {
				deviceTypesToImport[dt.Model] = dt
			}
		}
		return nil
	}
}

// NewClient creates a new netbox client using the environment variables
// This maintains parity/compatiblity with Device-Type-Librery-Import python repo
func NewClient() (*netbox.APIClient, context.Context, error) {
	// use the certificates in the http client
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
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
	token := "0123456789abcdef0123456789abcdef01234567"
	host := stripProtocolsAndSpecialChars("127.0.0.1:8000")
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
