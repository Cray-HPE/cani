package netbox

import (
	"context"
	"crypto/tls"
	"encoding/json"
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
	d := netbox.WritableDeviceTypeRequest{}
	log.Info().Msgf("len(dt.Interfaces) %+v", len(dt.Interfaces))
	// required:
	//   - manufacturer
	mf, err := dt.getManufacturer(client, ctx)
	if err != nil {
		return err
	}
	d.SetManufacturer(mf.Id)
	//   - model
	d.SetModel(dt.Model)
	//   - slug
	d.SetSlug(dt.Slug)
	//	 - u_height
	if dt.UHeight != nil {
		d.SetUHeight(*dt.UHeight)
	}

	// optional:
	if dt.Comments != nil {
		d.SetComments(*dt.Comments)
	}
	if dt.PartNumber != nil {
		d.SetPartNumber(*dt.PartNumber)
	}
	if dt.IsFullDepth != nil {
		d.SetIsFullDepth(*dt.IsFullDepth)
	}
	// unit and weight required together
	if dt.WeightUnit != nil {
		wu := netbox.DeviceTypeWeightUnitValue(*dt.WeightUnit)
		d.SetWeightUnit(wu)
		if dt.Weight != nil {
			d.SetWeight(*dt.Weight)
		}
	}
	d.AdditionalProperties = make(map[string]interface{}, 0)
	d.CustomFields = make(map[string]interface{}, 0)
	// req.SetFrontImage(dt.FrontImage)
	// req.SetRearImage(dt.RearImage)
	// req.SetDefaultPlatform()
	// req.SetWeightUnit()
	// req.SetSubdeviceRole(*dt.SubdeviceRole)
	// req.SetTags()
	// req.SetAirflow(dt.Airflow)

	req := client.DcimAPI.DcimDeviceTypesCreate(ctx).WritableDeviceTypeRequest(d)
	createdDeviceType, resp, err := req.Execute()
	if err != nil {
		body, _ := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if resp.StatusCode == 201 {
			log.Debug().Msgf("%+v", resp.Status) // still created
		} else if resp.StatusCode == 400 {
			log.Error().Msgf("%+v", string(body))
			return err
		}
		// if it was created, there seems to be a bug where the devicetype returned by req.Execute() is empty
		// do some stupid parsing of the response body to get the id
		getId := make(map[string]interface{}, 0)
		err = json.Unmarshal(body, &getId) // Convert to a map
		if err != nil {
			return err
		}
		cid := getId["id"].(float64)        // created device id
		createdDeviceType.SetId(int32(cid)) // set it
	}

	// Create interface templates and assign it to the created device type
	for _, i := range dt.Interfaces {
		iftmpl := netbox.NewWritableInterfaceTemplateRequestWithDefaults()
		// required:
		//  - name
		iftmpl.SetName(i.Name)
		//  - type
		itv := netbox.NewInterfaceTypeWithDefaults()
		val := netbox.InterfaceTypeValue(i.Type)
		// lbl := netbox.InterfaceTypeLabel(i.Type)
		itv.SetValue(val)
		// itv.SetLabel(lbl)
		iftmpl.SetType(val)
		iftmpl.SetDeviceType(createdDeviceType.GetId())
		// ndt := netbox.NewNestedDeviceTypeWithDefaults()
		// ndt.SetDisplay(dt.Model)
		r := client.DcimAPI.DcimInterfaceTemplatesCreate(ctx).WritableInterfaceTemplateRequest(*iftmpl)
		// r := client.DcimAPI.DcimInterfacesCreate(ctx).WritableInterfaceRequest(*iftmpl)
		_, resp, err := r.Execute()
		if err != nil {

			body, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			// still created
			if resp.StatusCode == 201 {
				return nil
			}

			if resp.StatusCode == 400 {
				log.Error().Msgf("could not make interface template: %+v", string(body))
				return err
			}

			return err
		}
	}
	return nil
}

func createModuleType(client *netbox.APIClient, ctx context.Context, mt ModuleType) error {
	existingMfg, err := getExistingVendors(client, ctx)
	if err != nil {
		return err
	}

	module := netbox.NewWritableModuleTypeRequestWithDefaults()
	for _, m := range existingMfg {
		if strings.ToLower(m.Name) == strings.ToLower(mt.Manufacturer) {
			module.SetManufacturer(m.Id)
			module.SetModel(mt.Model)
		}
	}
	_, ok := module.GetManufacturerOk()
	if !ok {
		return fmt.Errorf("manufacturer not found: %+v", mt.Manufacturer)
	}

	req := client.DcimAPI.DcimModuleTypesCreate(ctx).WritableModuleTypeRequest(*module)
	_, resp, err := req.Execute()
	if err != nil {

		body, _ := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		// still created
		if resp.StatusCode == 201 {
			return nil
		}

		if resp.StatusCode == 400 {
			log.Warn().Msgf("%+v", string(body))
			return err
		}

		return err
	}
	return nil
}

func createInterfaces(client *netbox.APIClient, ctx context.Context, dt DeviceType) error {
	for _, i := range dt.Interfaces {
		req := netbox.WritableInterfaceRequest{}
		req.SetName(i.Name)
		d := client.DcimAPI.DcimInterfacesCreate(ctx).WritableInterfaceRequest(req)
		_, resp, err := d.Execute()
		if err != nil {

			body, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			// still created
			if resp.StatusCode == 201 {
				// no value given for required property device_count
				return nil
			}

			if resp.StatusCode == 400 {
				log.Warn().Msgf("DeviceType already exists (ignoring): %+v (%+v)", dt.Model, stringToSlug(dt.Model))
				log.Warn().Msgf("%+v", string(body))
				return nil
			}

			return err
		}
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

func CreateModuleTypeMap(manufacturersToImport map[string]string, moduleTypesToImport map[string]ModuleType, ignoreDirs []string) filepath.WalkFunc {
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
			mt, err := UnmarshalModuleType(path)
			if err != nil {
				return err
			}

			// add the manufacturer to the map
			// it is required to add a devicetype
			sanitized := strings.ToLower(mt.Manufacturer)
			_, manufacturerExists := manufacturersToImport[sanitized]
			if !manufacturerExists {
				manufacturersToImport[sanitized] = mt.Manufacturer
			}

			// add the module type to the map
			_, moduleTypeExists := moduleTypesToImport[mt.Model]
			if !moduleTypeExists {
				moduleTypesToImport[mt.Model] = mt
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
	token := os.Getenv("NETBOX_TOKEN")
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
