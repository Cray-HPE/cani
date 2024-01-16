/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package hpengi

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/hpcm"
	"github.com/Cray-HPE/cani/pkg/canu"
	nb "github.com/Cray-HPE/cani/pkg/netbox"
	"github.com/google/uuid"
	"github.com/netbox-community/go-netbox/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ImportInit imports the external inventory data into CANI's inventory format
func (hpengi *Hpengi) ImportInit(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	log.Warn().Msgf("ImportInit partially implemented")
	// copy the datastore and add set provider metadata
	ds, err := setupTempDatastore(datastore)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("cm-config") {
		log.Info().Msgf("Translating %T", hpengi.Hpcm.CmConfig)
		f, _ := cmd.Flags().GetString("cm-config")
		cm, err := hpcm.LoadCmConfig(f)
		if err != nil {
			return err
		}

		translated, err = cm.TranslateCmHardwareToCaniHw()
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("sls-dumpstate") {
		// get a map of translated hardware from the hpcm config
		translated, err = hpengi.translateSlsDumpstateToCaniHw(cmd, args)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("paddle") {
		log.Info().Msgf("Translating %T", hpengi.Paddle)
		// get a map of translated hardware from the hpcm config
		// translated, err = hpengi.translatePaddleToCaniHw(cmd, args)
		// if err != nil {
		// 	return err
		// }
	}

	if cmd.Flags().Changed("cmdb") {
		// translate external inventory data to cani hardware entries
		translated, err = hpengi.Hpcm.Translate(cmd, args)
		if err != nil {
			return err
		}
	}

	// all flags will return a map of translated hardware
	// loop through that, and add it each to the datastore
	for _, hw := range translated {
		err = ds.Add(hw)
		if err != nil {
			return err
		}
	}

	// merge the datastore if all is successful
	err = datastore.Merge(ds)
	if err != nil {
		return err
	}

	return datastore.Flush()
}

// Import
func (hpengi *Hpengi) Import(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	log.Warn().Msgf("Import partially implemented")
	// // copy the datastore and add set provider metadata
	// ds, err := setupTempDatastore(datastore)
	// if err != nil {
	// 	return err
	// }

	if cmd.Flags().Changed("sls-dumpstate") {
		dump, err := hpengi.getSlsDumpstate(cmd, args)
		if err != nil {
			return err
		}
		hpengi.SlsInput = dump
	}

	if cmd.Flags().Changed("paddle") {
		ccj, err := hpengi.getPaddle(cmd, args)
		if err != nil {
			return err
		}
		hpengi.Paddle = ccj
	}

	// translate the hpcm fields to cani fields
	err := hpengi.paddleAndSlsToNetbox()
	if err != nil {
		return err
	}
	// all flags will return a map of translated hardware
	// loop through that, and add it each to the datastore
	// for _, hw := range translated {
	// 	err = ds.Add(hw)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// // merge the datastore if all is successful
	// err = datastore.Merge(ds)
	// if err != nil {
	// 	return err
	// }

	// return datastore.Flush()
	return nil
}

// translateSlsDumpstateToCaniHw
func (hpengi *Hpengi) translateSlsDumpstateToCaniHw(cmd *cobra.Command, args []string) (translated map[uuid.UUID]*inventory.Hardware, err error) {
	log.Info().Msgf("%+v", hpengi.SlsInput.Hardware)
	// log.Debug().Msgf("Translating %s hardware to %s format (%+v)", "CANU", taxonomy.App, *n.CommonName)
	// hw := &inventory.Hardware{}
	// hw.Name = slsHw.Xname
	// // translate the hpcm fields to cani fields
	// err = translatePaddleHardwareToCaniHw(n, *hpengi.Paddle, hw)
	// if err != nil {
	// 	return translated, err
	// }

	// translate the hpcm fields to cani fields
	// err = hpengi.translateSlsToNetbox()
	// if err != nil {
	// 	return translated, err
	// }

	// // add the hardware to the map if it does not exist
	// _, exists := translated[hw.ID]
	// if exists {
	// 	return translated, fmt.Errorf("Hardware already exists: %s", hw.ID)
	// }
	// translated[hw.ID] = hw

	// return the map of translated hpcm --> cani hardware
	return translated, nil
}

func (hpengi *Hpengi) paddleAndSlsToNetbox() (err error) {
	const format = "  %-*v: %-*v %-*v: %v"
	const fieldWidth = 15
	const valueWidth = 30
	// log.Debug().Msgf(format, fieldWidth, "EXISTING FIELD", valueWidth, "VALUE", fieldWidth, "FINCH FIELD", "VALUE")
	// log.Debug().Msgf(format, fieldWidth, "---------", valueWidth, "---------", fieldWidth, "---------", "---------")

	client, ctx, err := nb.NewClient()
	if err != nil {
		return err
	}

	req := client.DcimAPI.DcimRacksList(ctx).Limit(int32(9999)).Offset(int32(0))
	er, resp, err := req.Execute()
	if err != nil {
		log.Error().Msgf("%+v: %+v", resp.Status, err)
		return err
	}

	// create a map of existing racks for comparison
	// the api does not block the creation of a rack with the same name if only the required fields are set
	existingRacks := map[string]netbox.Rack{}
	if len(er.Results) >= 0 {
		for _, existingRack := range er.Results {
			existingRacks[existingRack.Name] = existingRack
		}
	}
	log.Info().Msgf("%+v existing racks detected", len(existingRacks))

	// get and create racks
	racks := make(map[string]netbox.ApiDcimRacksCreateRequest, 0)
	for _, topo := range hpengi.Paddle.Topology {
		// set the rack it is in
		// required:
		//    - site
		//    - name
		//    - status
		//    - width
		//    - starting unit
		//    - height in rack units

		// create the rack request
		rack := netbox.NewWritableRackRequestWithDefaults()

		// set the site
		rack.SetSite(1) // FIXME hardcoding for now

		// set the name
		rack.SetName(*topo.Location.Rack)

		// set the status
		status := netbox.NewRackStatusWithDefaults()
		status.SetLabel("Active")
		status.SetValue("active")
		rack.SetStatus(status.GetValue()) // FIXME hardcoding for now

		// set the width
		width := netbox.NewRackWidthWithDefaults()
		width.SetLabel("19 inches")
		width.SetValue(19)
		rack.SetWidth(netbox.PatchedWritableRackRequestWidth(width.GetValue()))

		// set the height in rack units
		reCdu := regexp.MustCompile(`^d.*`)
		rePdu := regexp.MustCompile(`^p.*`)
		var height, startingU int32
		switch {
		case reCdu.MatchString(*topo.Location.Rack):
			height = 42
			startingU = 1
		case rePdu.MatchString(*topo.Location.Elevation):
			height = 3
			startingU = 1
		default:
			height = 48
			startingU = 1
		}

		// set the starting unit
		rack.SetStartingUnit(startingU) // FIXME hardcoding for now
		rack.SetUHeight(height)         // FIXME hardcoding for now

		// create the request using the crafted rack
		req := client.DcimAPI.DcimRacksCreate(ctx).WritableRackRequest(*rack)

		// create the rack request if the rack does not exist
		_, ok := existingRacks[*topo.Location.Rack]
		if !ok {
			// add the request to the map for execution later
			_, exists := racks[*topo.Location.Rack]
			if !exists {
				log.Trace().Msgf("Crafting a rack request: %+v", *topo.Location.Rack)
				racks[*topo.Location.Rack] = req
			}
		}
	}
	log.Debug().Msgf("%+v pending rack requests", len(racks))

	// execute all the rack requests, creating them in netbox
	for name, req := range racks {
		log.Info().Msgf("Executing a create rack request: %+v", name)
		// FIXME: Get existing racks.  API does not reject same rack name
		_, resp, err := req.Execute()
		if err != nil {
			// 201 is created
			if resp.StatusCode == 201 {
				continue
			}
			log.Error().Msgf("%+v", resp.Status)
			return err
		}
	}

	// get existing devices
	existingDevices := map[string]netbox.DeviceWithConfigContext{}
	req2 := client.DcimAPI.DcimDevicesList(ctx).Limit(int32(9999)).Offset(int32(0))
	ed, resp, err := req2.Execute()
	if err != nil {
		log.Error().Msgf("%+v: %+v", resp.Status, err)
		return err
	}

	for _, existingDevice := range ed.Results {
		existingDevices[existingDevice.GetName()] = existingDevice
	}
	log.Info().Msgf("%+v existing devices detected", len(existingDevices))
	/// paginated

	limit := 1000
	offset := 0
	existingDeviceTypesMap := make(map[string]netbox.DeviceType, 0)
	existingDeviceTypesConcat := make([]netbox.DeviceType, 0)
	for {
		req := client.DcimAPI.DcimDeviceTypesList(ctx).Limit(int32(limit)).Offset(int32(offset))
		paginatedDeviceTypes, resp, err := req.Execute()
		if err != nil {
			log.Error().Msgf("%+v: %+v", resp.Status, err)
			return err
		}

		existingDeviceTypesConcat = append(existingDeviceTypesConcat, paginatedDeviceTypes.Results...)

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

	// paginated

	for _, existingDeviceType := range existingDeviceTypesConcat {
		existingDeviceTypesMap[existingDeviceType.GetModel()] = existingDeviceType
	}
	log.Info().Msgf("%+v existing devices types detected", len(existingDeviceTypesMap))

	// get and create devices
	devices := make(map[string]netbox.ApiDcimDevicesCreateRequest, 0)
	for _, topo := range hpengi.Paddle.Topology {
		// adding a device requires:
		//    - site
		//    - status
		//    - device role
		//    - device type

		// create the device request
		device := netbox.NewWritableDeviceWithConfigContextRequestWithDefaults()

		// set the site
		device.SetSite(1) // FIXME hardcoding for now

		// set the status
		status := netbox.NewDeviceStatusWithDefaults()
		status.SetLabel("Active")
		status.SetValue("active")
		device.SetStatus(status.GetValue()) // FIXME hardcoding for now

		// set the role
		device.SetRole(1) // FIXME hardcoding for now

		// // trim the part number from strings like: 8325_JL706A, resulting in the model: JL706A
		// model := ""
		// partNumber := ""
		// m := strings.Split(*topo.Model, "_")
		// if len(m) > 1 {
		// 	model = m[0]
		// 	partNumber = m[1]
		// } else {
		// 	model = m[0]
		// }

		// get the device type id from the map
		existingDeviceType, ok := existingDeviceTypesMap[*topo.Architecture]
		if ok {
			log.Info().Msgf("Adding device of type: %+v", existingDeviceType.GetModel())
			device.SetDeviceType(existingDeviceType.GetId())
		} else {
			existingDeviceType2, ok2 := existingDeviceTypesMap[*topo.Model]
			if ok2 {
				device.SetDeviceType(existingDeviceType2.GetId())
			} else {
				log.Warn().Msgf("Adding device of type: %+v", "HARDCODE")
				device.SetDeviceType(13283) // FIXME hardcoding for now
			}
		}

		// also set the name
		device.SetName(*topo.CommonName)

		// set the rack it is in
		existingRack, ok := existingRacks[*topo.Location.Rack]
		if ok {
			device.SetRack(existingRack.Id)
		}

		var u float64
		switch *topo.Architecture {
		case "pdu", "cmm", "cec":
			u = 0.0
		default:
			// most devices will just be in a u	position
			elevation := strings.TrimPrefix(*topo.Location.Elevation, "u")
			e, err := strconv.Atoi(elevation)
			if err != nil {
				return err
			}
			u = float64(e)
		}
		if *topo.Architecture != "pdu" && *topo.Architecture != "cmm" && *topo.Architecture != "cec" {
			// set the elevation
			device.SetPosition(float64(u))
			face, err := netbox.NewRackFaceFromValue("front")
			if err != nil {
				return err
			}
			device.SetFace(*face)
		}

		// create the request using the crafted rack
		req := client.DcimAPI.DcimDevicesCreate(ctx).WritableDeviceWithConfigContextRequest(*device)

		// create the rack request if the rack does not exist
		_, deviceok := existingDevices[*topo.CommonName]
		if !deviceok {
			// add the request to the map for execution later
			_, exists := devices[*topo.CommonName]
			if !exists {
				log.Trace().Msgf("Crafting a new device request: %+v", *topo.CommonName)
				devices[*topo.CommonName] = req
			}
		}
	}
	log.Debug().Msgf("%+v pending device requests", len(devices))
	// log.Debug().Msgf(format, fieldWidth, "CommonName", valueWidth, *topo.CommonName, fieldWidth, "Name", *topo.CommonName)
	// log.Debug().Msgf(format, fieldWidth, "Vendor", valueWidth, *topo.Vendor, fieldWidth, "Manufacturer", *topo.Vendor)
	// log.Debug().Msgf(format, fieldWidth, "Model", valueWidth, *topo.Model, fieldWidth, "Model", model)
	// log.Debug().Msgf(format, fieldWidth, "Model", valueWidth, *topo.Model, fieldWidth, "PartNumber", partNumber)
	// log.Debug().Msgf(format, fieldWidth, "", valueWidth, *topo.Model, fieldWidth, "PartNumber", partNumber)
	// log.Debug().Msgf(format, fieldWidth, "Architecture", valueWidth, *topo.Architecture, fieldWidth, "Model", *topo.Architecture)
	log.Debug().Msgf("")
	// execute all the rack requests, creating them in netbox
	for name, req := range devices {
		log.Info().Msgf("Executing a create device request: %+v", name)
		_, resp, err := req.Execute()
		if err != nil {
			body, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			log.Warn().Msgf("Trying to add %+v: %+v", name, string(body))
			// 201 is created
			if resp.StatusCode == 201 {
				continue
			}
			log.Error().Msgf("%+v", resp.Status)
			return err
		}
	}

	return nil
}

// translatePaddleToCaniHw
// func (hpengi *Hpengi) translatePaddleToCaniHw(cmd *cobra.Command, args []string) (translated map[uuid.UUID]*inventory.Hardware, err error) {
// 	translated = make(map[uuid.UUID]*inventory.Hardware, 0)

// 	// use the certificates in the http client
// 	tlsConfig := &tls.Config{
// 		InsecureSkipVerify: true,
// 	}

// 	// use TLS config in transport
// 	tr := &http.Transport{
// 		TLSClientConfig: tlsConfig,
// 	}

// 	// Setup our HTTP transport and client
// 	httpClient := retryablehttp.NewClient()
// 	httpClient.HTTPClient.Transport = tr

// 	c := httpClient.StandardClient()

// 	// token := "47c21acc6425474c66b98d20bf247a52853eee35"
// 	token := "0123456789abcdef0123456789abcdef01234567"
// 	cfg := &netbox.Configuration{
// 		// Host:       "demo.netbox.dev",
// 		Host:       "127.0.0.1:8000",
// 		Scheme:     "http",
// 		HTTPClient: c,
// 		DefaultHeader: map[string]string{
// 			"Authorization": fmt.Sprintf("Token %s", token),
// 		},
// 		UserAgent: taxonomy.App,
// 		Debug:     false,
// 	}
// 	if os.Getenv("NB_DEBUG") != "" {
// 		cfg.Debug = true
// 	}

// 	cfg.Servers = netbox.ServerConfigurations{
// 		{
// 			// URL: fmt.Sprintf("%s://%s", cfg.Scheme, cfg.Host), Description: "Netbox Demo Server", Variables: map[string]netbox.ServerVariable{},
// 			URL: fmt.Sprintf("%s://%s", cfg.Scheme, cfg.Host), Description: "Nautobot container", Variables: map[string]netbox.ServerVariable{},
// 		},
// 	}
// 	ctx := context.Background()
// 	client := netbox.NewAPIClient(cfg) // or: client := netbox.NewAPIClientFor(fmt.Sprintf("https://%s", cfg.Host), token)

// 	dt := client.DcimAPI.DcimDeviceTypesList(ctx)
// 	dts, resp, err := dt.Execute()
// 	if err != nil {
// 		log.Error().Msgf("%+v: %+v", resp.Status, err)
// 		return translated, err
// 	}
// 	if len(dts.Results) >= 0 {
// 		for _, d := range dts.Results {
// 			log.Info().Msgf("%+v", d)
// 		}
// 	}

// 	for _, n := range hpengi.Paddle.Topology {
// 		log.Debug().Msgf("Translating %s hardware to %s format (%+v)", "CANU", taxonomy.App, *n.CommonName)
// 		hw := &inventory.Hardware{}

// 		// // translate the hpcm fields to cani fields
// 		// err = translatePaddleHardwareToCaniHw(n, *hpengi.Paddle, hw)
// 		// if err != nil {
// 		// 	return translated, err
// 		// }

// 		// translate the hpcm fields to cani fields
// 		err = translatePaddleHardwareToNetboxHw(n, *hpengi.Paddle, hw)
// 		if err != nil {
// 			return translated, err
// 		}

// 		// // add the hardware to the map if it does not exist
// 		// _, exists := translated[hw.ID]
// 		// if exists {
// 		// 	return translated, fmt.Errorf("Hardware already exists: %s", slsHw.ID)
// 		// }
// 		// translated[hw.ID] = hw
// 	}

// 	// return the map of translated hpcm --> cani hardware
// 	return translated, nil
// }

// func (hpengi *Hpengi) translateCmConfigToCaniHw(cmd *cobra.Command, args []string) (translated map[uuid.UUID]*inventory.Hardware, err error) {
// 	translated = make(map[uuid.UUID]*inventory.Hardware, 0)
// 	for _, d := range hpengi.CmConfig.Discover {
// 		log.Debug().Msgf("Translating %s hardware to %s format (%+v)", "HPCM", taxonomy.App, d.Hostname1)
// 		hw := &inventory.Hardware{}

// 		// translate the hpcm fields to cani fields
// 		err = translateCmHardwareToCaniHw(d, hpengi.CmConfig, hw)
// 		if err != nil {
// 			return translated, err
// 		}

// 		// add the hardware to the map if it does not exist
// 		_, exists := translated[hw.ID]
// 		if exists {
// 			return translated, fmt.Errorf("Hardware already exists: %s", slsHw.ID)
// 		}
// 		translated[hw.ID] = hw
// 	}

// 	// return the map of translated hpcm --> cani hardware
// 	return translated, nil
// }

func translatePaddleHardwareToNetboxHw(n canu.PaddleTopologyElem, ccj canu.Paddle, hw *inventory.Hardware) (err error) {
	const format = "  %-*v: %-*v %-*v: %v"
	const fieldWidth = 15
	const valueWidth = 30
	log.Debug().Msgf("Importing %+v", *n.CommonName)
	// existingRoles := client.DcimAPI.DcimDeviceRolesList(ctx)
	// roles, resp, err := existingRoles.Execute()
	// if err != nil {
	// 	log.Error().Msgf("%+v: %+v", resp.Status, err)
	// 	return err
	// }
	// for _, r := range roles.Results {
	// 	log.Debug().Msgf("%+v", r)
	// }
	// log.Debug().Msgf("%+v", roles)

	// req := netbox.WritableDeviceRoleRequest{
	// 	Name: *n.Type,
	// 	Slug: strings.ToLower(*n.Type),
	// }

	// deviceRole, resp, err := client.DcimAPI.DcimDeviceRolesCreate(ctx).WritableDeviceRoleRequest(req).Execute()
	// if err != nil {
	// 	return err
	// }

	// log.Info().Msgf("%+v: %+v", resp.Status, deviceRole)

	// name := netbox.NullableString{}
	// name.Set(n.CommonName)
	// nbDevice := netbox.WritableDeviceWithConfigContextRequest{
	// 	Name: name,
	// }

	// devs, resp, err := client.DcimAPI.DcimDevicesCreate(ctx).WritableDeviceWithConfigContextRequest(nbDevice).Execute()
	// if err != nil {
	// 	return err
	// }
	// log.Info().Msgf("%+v", resp.StatusCode)

	// log.Debug().Msgf("%+v", devs.DeviceType)
	log.Debug().Msgf("")

	return nil
}

func translatePaddleHardwareToCaniHw(n canu.PaddleTopologyElem, ccj canu.Paddle, hw *inventory.Hardware) (err error) {
	const format = "  %-*v: %-*v %-*v: %v"
	const fieldWidth = 15
	const valueWidth = 30
	log.Debug().Msgf(format, fieldWidth, "CANU FIELD", valueWidth, "VALUE", fieldWidth, "CANI FIELD", "VALUE")
	log.Debug().Msgf(format, fieldWidth, "---------", valueWidth, "---------", fieldWidth, "---------", "---------")

	// create a uuid for the new hardware since the paddle file does not have one
	u := uuid.New()
	hw.ID = u
	log.Debug().Msgf(format, fieldWidth, "ID", valueWidth, *n.Id, fieldWidth, "UUID", hw.ID)

	// hw.Architecture = *n.Architecture
	// log.Debug().Msgf(format, fieldWidth, "Architecture", valueWidth, *n.Architecture, fieldWidth, "Architecture", "")

	hw.Model = *n.Model
	log.Debug().Msgf(format, fieldWidth, "Model", valueWidth, *n.Model, fieldWidth, "Model", hw.Model)

	hw.Name = *n.CommonName
	log.Debug().Msgf(format, fieldWidth, "Name", valueWidth, *n.CommonName, fieldWidth, "Name", hw.Name)

	hw.Vendor = strings.ToTitle(*n.Vendor)
	log.Debug().Msgf(format, fieldWidth, "Vendor", valueWidth, *n.Vendor, fieldWidth, "Vendor", hw.Vendor)

	hw.Type, err = n.PaddleTypeToCaniHardwareType()
	if err != nil {
		return err
	}
	log.Debug().Msgf(format, fieldWidth, "Type", valueWidth, *n.Type, fieldWidth, "Type", hw.Type)

	model := strings.ToLower(*n.Model)
	hw.DeviceTypeSlug = strings.Replace(model, "_", "-", -1)
	log.Debug().Msgf(format, fieldWidth, "DeviceTypeSlug", valueWidth, *n.Model, fieldWidth, "DeviceTypeSlug", hw.DeviceTypeSlug)

	// dt, err := n.PaddleToNetboxDeviceType()
	// if err != nil {
	// 	return err
	// }

	// hw.Device = dt
	log.Debug().Msgf(format, fieldWidth, "RackElevation", valueWidth, n.RackElevation, fieldWidth, "UHeight", "hw.Device.UHeight")
	log.Debug().Msgf(format, fieldWidth, "Ports", valueWidth, "-----------", fieldWidth, "FrontPorts", "----------")
	for _, p := range n.Ports {
		if p.DestinationNodeId != nil {
			log.Debug().Msgf(format, fieldWidth, "DestNodeId", valueWidth, *p.DestinationNodeId, fieldWidth, "Id", "hw.Device.FrontPorts[i].Id")
		}
		if p.Slot != nil {
			log.Debug().Msgf(format, fieldWidth, "Slot", valueWidth, *p.Slot, fieldWidth, "Label", "*hw.Device.FrontPorts[i].Type.Label")
		}
		if p.Speed != nil {
			log.Debug().Msgf(format, fieldWidth, "Speed", valueWidth, *p.Speed, fieldWidth, "InterfaceSpeed", "*hw.Device.Interfaces[i].Speed.Get()")
		}
	}
	// hw.LocationPath, err := n.PaddleLocToCaniLoc()
	// if err != nil {
	// 	return err
	// }

	// if n.Location.SubLocation != nil {
	// 	log.Debug().Msgf(format, fieldWidth, "SubLocation", valueWidth, *n.Location.SubLocation, fieldWidth, "Location", "")
	// }
	// if n.Location.Rack != nil {
	// 	log.Debug().Msgf(format, fieldWidth, "Rack", valueWidth, *n.Location.Rack, fieldWidth, "-", "")
	// }
	// if n.Location.Parent != nil {
	// 	log.Debug().Msgf(format, fieldWidth, "Parent", valueWidth, *n.Location.Parent, fieldWidth, "-", "")
	// }
	// if n.Location.Elevation != nil {
	// 	log.Debug().Msgf(format, fieldWidth, "Elevation", valueWidth, *n.Location.Elevation, fieldWidth, "-", "")
	// }
	// if n.RackElevation != nil {
	// 	log.Debug().Msgf(format, fieldWidth, "RackElevation", valueWidth, *n.RackElevation, fieldWidth, "-", "")
	// }
	// if n.RackNumber != nil {
	// 	log.Debug().Msgf(format, fieldWidth, "RackNumber", valueWidth, *n.RackNumber, fieldWidth, "-", "")
	// }

	log.Debug().Msgf("")

	return nil
}

// func translateCmHardwareToCaniHw(d hpcm.Discover, cm hpcm.HpcmConfig, hw *inventory.Hardware) (err error) {
// 	// create a uuid for the new hardware
// 	u := uuid.New()
// 	log.Debug().Msgf("  Unique Identifier:  --> %s: %+v", "ID", u)
// 	hw.ID = u

// 	// Convert HPCM type to cani hardwaretypes
// 	t := hpcmTypeToCaniType(d, cm)
// 	if err != nil {
// 		return err
// 	}
// 	hw.Type = t
// 	log.Debug().Msgf("  type: %s --> %s: %s", d.Type, "Type", t)

// 	// Convert HPCM template name to cani device type slug
// 	s := hpcmTemplateNameToCaniSlug(d, cm)
// 	if err != nil {
// 		return err
// 	}
// 	hw.DeviceTypeSlug = s
// 	log.Debug().Msgf("  template_name: %s --> %s: %s", d.TemplateName, "DeviceTypeSlug", s)

// 	// Convert HPCM card type to cani vendor
// 	v := hpcmCardTypeToCaniVendor(d, cm)
// 	if err != nil {
// 		return err
// 	}
// 	hw.Vendor = v
// 	log.Debug().Msgf("  card_type: %s --> %s: %s", d.CardType, "Vendor", v)

// 	// Convert HPCM card type to cani vendor
// 	lp := hpcmGeoCaniLocationPath(d, cm)
// 	if err != nil {
// 		return err
// 	}
// 	hw.LocationPath = lp
// 	log.Debug().Msgf("  *_nr: %d->%d->%d->%d --> %s: %s",
// 		d.RackNr,
// 		d.Chassis,
// 		d.Tray,
// 		d.NodeNr,
// 		"LocationPath", lp.String())

// 	// these fields map 1:1 and are not necessarily required, so just fill them
// 	hw.LocationOrdinal = &d.NodeNr
// 	log.Debug().Msgf("  node_nr: %d --> %s: %d", d.NodeNr, "LocationOrdinal", *hw.LocationOrdinal)
// 	hw.Architecture = d.Architecture
// 	log.Debug().Msgf("  %s: %s %s %s: %s", "architecture", d.Architecture, "-->", "Architecture", hw.Architecture)
// 	hw.Model = d.TemplateName
// 	// log.Debug().Msgf("  %s: %s %s %s: %s", "template_name", d.TemplateName, "-->", "Model", hw.Model)
// 	hw.Name = d.Hostname1
// 	log.Debug().Msgf("  %s: %s %s %s: %s", "hostname1", d.Hostname1, "-->", "Name", hw.Name)
// 	log.Debug().Msgf("")

// 	return nil
// }

// // hpcmGeoCaniLocationPath
// func hpcmGeoCaniLocationPath(d hpcm.Discover, cm hpcm.HpcmConfig) (lp inventory.LocationPath) {
// 	lp = inventory.LocationPath{
// 		inventory.LocationToken{HardwareType: hardwaretypes.Cabinet, Ordinal: d.RackNr},
// 		inventory.LocationToken{HardwareType: hardwaretypes.Chassis, Ordinal: d.Chassis},
// 		inventory.LocationToken{HardwareType: hardwaretypes.NodeBlade, Ordinal: d.Tray},
// 		inventory.LocationToken{HardwareType: hardwaretypes.Node, Ordinal: d.NodeNr},
// 	}
// 	return lp
// }

// // hpcmCardTypeToCaniVendor
// func hpcmCardTypeToCaniVendor(d hpcm.Discover, cm hpcm.HpcmConfig) (v string) {
// 	switch d.CardType {
// 	case "iLo":
// 		v = "HPE"
// 	case "Intel":
// 		v = "Intel"
// 	case "IPMI":
// 		v = "IPMI"
// 	default:
// 		v = "Unknown"
// 	}

// 	return v
// }

// // hpcmTemplateNameToCaniSlug
// func hpcmTemplateNameToCaniSlug(d hpcm.Discover, cm hpcm.HpcmConfig) (t string) {
// 	switch d.Type {
// 	case "":
// 		// tpl := getDiscoverTemplate(d, cm)
// 		// t = d.TemplateName
// 	default:
// 		t = d.TemplateName
// 	}

// 	return t
// }

// // hpcmTypeToCaniType
// func hpcmTypeToCaniType(d hpcm.Discover, cm hpcm.HpcmConfig) (t hardwaretypes.HardwareType) {
// 	switch d.Type {
// 	case "":
// 		// tpl := getDiscoverTemplate(d, cm)
// 		t = hardwaretypes.NodeBlade
// 	case "leaf", "spine":
// 		t = hardwaretypes.ManagementSwitch
// 	}

// 	return t
// }

// // getDiscoverTemplate
// func getDiscoverTemplate(d hpcm.Discover, cm hpcm.HpcmConfig) (tpl hpcm.Template) {
// 	val, exists := cm.Templates[d.TemplateName]
// 	if exists {
// 		tpl = val
// 	}

// 	return tpl
// }

// setupTempDatastore
func setupTempDatastore(datastore inventory.Datastore) (temp inventory.Datastore, err error) {
	temp, err = datastore.Clone()
	if err != nil {
		return temp, errors.Join(fmt.Errorf("failed to clone datastore"), err)
	}

	// Get the parent system
	sys, err := datastore.GetSystemZero()
	if err != nil {
		return temp, err
	}
	// Set additional metadata
	p, err := datastore.InventoryProvider()
	if err != nil {
		return temp, err
	}
	// Set top-level meta to the "system"
	sysMeta := inventory.ProviderMetadataRaw{}
	sys.ProviderMetadata = make(map[inventory.Provider]inventory.ProviderMetadataRaw)
	sys.ProviderMetadata[p] = sysMeta

	// Add it to the datastore
	err = temp.Update(&sys)
	if err != nil {
		return temp, err
	}
	return temp, nil
}
