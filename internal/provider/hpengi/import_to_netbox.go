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
	"context"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/pkg/canu"
	nb "github.com/Cray-HPE/cani/pkg/netbox"
	"github.com/netbox-community/go-netbox/v3"
	"github.com/rs/zerolog/log"
)

const maxLimit = 1000 // how many items to get per page in netbox

// paddleAndSlsToNetbox translates the paddle and sls data and posts it netbox
func (hpengi *Hpengi) paddleAndSlsToNetbox() (err error) {
	// setup a new client
	client, ctx, err := nb.NewClient()
	if err != nil {
		return err
	}
	hpengi.NetboxClient = client
	hpengi.context = ctx

	// get the existing device types
	// create a map of devicetype requests and execute them
	existingDeviceTypes, err := getExistingDeviceTypes(client, ctx)
	if err != nil {
		return err
	}
	err = hpengi.createDeviceTypesRequests(existingDeviceTypes) // FIXME: hardcoded dir, logic from previous command
	if err != nil {
		return err
	}

	existingModuleTypes, err := getExistingModuleTypes(client, ctx)
	if err != nil {
		return err
	}
	err = hpengi.createModuleTypesRequests(existingModuleTypes) // FIXME: hardcoded dir, logic from previous command
	if err != nil {
		return err
	}

	// get the existing racks
	// create a map of rack requests and execute them
	existingRacks, err := getExistingRacks(client, ctx)
	if err != nil {
		return err
	}
	createRacksRequests, err := hpengi.createRackRequests(existingRacks)
	if err != nil {
		return err
	}
	log.Info().Msgf("Need to create %+v new racks", len(createRacksRequests))
	err = hpengi.executeCreateRackRequests(createRacksRequests)
	if err != nil {
		return err
	}

	// create devices now that the devicetypes and racks are created
	createDeviceRequests, err := hpengi.createDevicesRequests()
	if err != nil {
		return err
	}
	log.Info().Msgf("Need to create %+v new devices", len(createDeviceRequests))
	err = hpengi.executeCreateDevicesRequests(createDeviceRequests)
	if err != nil {
		return err
	}
	return nil
}

// getExistingDeviceTypes gets the existing racks from netbox
func getExistingDeviceTypes(client *netbox.APIClient, ctx context.Context) (existingDeviceTypes map[string]netbox.DeviceType, err error) {
	limit := maxLimit
	offset := 0
	existingDeviceTypes = make(map[string]netbox.DeviceType, 0)
	existingConcat := make([]netbox.DeviceType, 0)
	for {
		req := client.DcimAPI.DcimDeviceTypesList(ctx).Limit(int32(limit)).Offset(int32(offset))
		paginated, resp, err := req.Execute()
		if err != nil {
			log.Error().Msgf("%+v: %+v", resp.Status, err)
			return existingDeviceTypes, err
		}

		// append the results to to the slice
		existingConcat = append(existingConcat, paginated.Results...)

		// no more pages to get
		if paginated.Next.Get() == nil {
			break
		}

		// parse the next url to get the limit and offset
		var next *url.URL
		next, err = url.Parse(paginated.GetNext())
		if err != nil {
			return existingDeviceTypes, err
		}
		limitQ := next.Query().Get("limit")
		offsetQ := next.Query().Get("offset")
		limit, err = strconv.Atoi(limitQ)
		if err != nil {
			return existingDeviceTypes, err
		}
		offset, err = strconv.Atoi(offsetQ)
		if err != nil {
			return existingDeviceTypes, err
		}
	}

	// add the entires to the map after the loop is over and all pages are fetched
	for _, existing := range existingConcat {
		existingDeviceTypes[existing.Slug] = existing
	}

	return existingDeviceTypes, nil
}

// getExistingModuleTypes gets the existing racks from netbox
func getExistingModuleTypes(client *netbox.APIClient, ctx context.Context) (existing map[string]netbox.ModuleType, err error) {
	limit := maxLimit
	offset := 0
	existing = make(map[string]netbox.ModuleType, 0)
	existingConcat := make([]netbox.ModuleType, 0)
	for {
		req := client.DcimAPI.DcimModuleTypesList(ctx).Limit(int32(limit)).Offset(int32(offset))
		paginated, resp, err := req.Execute()
		if err != nil {
			log.Error().Msgf("%+v: %+v", resp.Status, err)
			return existing, err
		}

		// append the results to to the slice
		existingConcat = append(existingConcat, paginated.Results...)

		// no more pages to get
		if paginated.Next.Get() == nil {
			break
		}

		// parse the next url to get the limit and offset
		var next *url.URL
		next, err = url.Parse(paginated.GetNext())
		if err != nil {
			return existing, err
		}
		limitQ := next.Query().Get("limit")
		offsetQ := next.Query().Get("offset")
		limit, err = strconv.Atoi(limitQ)
		if err != nil {
			return existing, err
		}
		offset, err = strconv.Atoi(offsetQ)
		if err != nil {
			return existing, err
		}
	}

	// add the entires to the map after the loop is over and all pages are fetched
	for _, e := range existingConcat {
		existing[e.GetDisplay()] = e
	}

	return existing, nil
}

// getExistingRacks gets the existing racks from netbox
func getExistingRacks(client *netbox.APIClient, ctx context.Context) (existingRacks map[string]netbox.Rack, err error) {
	limit := maxLimit
	offset := 0
	existingRacks = make(map[string]netbox.Rack, 0)
	existingConcat := make([]netbox.Rack, 0)
	for {
		req := client.DcimAPI.DcimRacksList(ctx).Limit(int32(limit)).Offset(int32(offset))
		paginated, resp, err := req.Execute()
		if err != nil {
			log.Error().Msgf("%+v: %+v", resp.Status, err)
			return existingRacks, err
		}

		// append the results to to the slice
		existingConcat = append(existingConcat, paginated.Results...)

		// no more pages to get
		if paginated.Next.Get() == nil {
			break
		}

		// parse the next url to get the limit and offset
		var next *url.URL
		next, err = url.Parse(paginated.GetNext())
		if err != nil {
			return existingRacks, err
		}
		limitQ := next.Query().Get("limit")
		offsetQ := next.Query().Get("offset")
		limit, err = strconv.Atoi(limitQ)
		if err != nil {
			return existingRacks, err
		}
		offset, err = strconv.Atoi(offsetQ)
		if err != nil {
			return existingRacks, err
		}
	}

	// add the entires to the map after the loop is over and all pages are fetched
	for _, existing := range existingConcat {
		existingRacks[existing.GetName()] = existing
	}

	return existingRacks, nil
}

// getExistingDevices gets the existing racks from netbox
func getExistingDevices(client *netbox.APIClient, ctx context.Context) (existingDevices map[string]netbox.DeviceWithConfigContext, err error) {
	limit := maxLimit
	offset := 0
	existingDevices = make(map[string]netbox.DeviceWithConfigContext, 0)
	existingConcat := make([]netbox.DeviceWithConfigContext, 0)
	for {
		req := client.DcimAPI.DcimDevicesList(ctx).Limit(int32(limit)).Offset(int32(offset))
		paginated, resp, err := req.Execute()
		if err != nil {
			log.Error().Msgf("%+v: %+v", resp.Status, err)
			return existingDevices, err
		}

		// append the results to to the slice
		existingConcat = append(existingConcat, paginated.Results...)

		// no more pages to get
		if paginated.Next.Get() == nil {
			break
		}

		// parse the next url to get the limit and offset
		var next *url.URL
		next, err = url.Parse(paginated.GetNext())
		if err != nil {
			return existingDevices, err
		}
		limitQ := next.Query().Get("limit")
		offsetQ := next.Query().Get("offset")
		limit, err = strconv.Atoi(limitQ)
		if err != nil {
			return existingDevices, err
		}
		offset, err = strconv.Atoi(offsetQ)
		if err != nil {
			return existingDevices, err
		}
	}

	// add the entires to the map after the loop is over and all pages are fetched
	for _, existing := range existingConcat {
		existingDevices[existing.GetName()] = existing
	}
	log.Info().Msgf("Found %+v existing devices", len(existingDevices))

	return existingDevices, nil
}

func (hpengi *Hpengi) createDeviceTypesRequests(existing map[string]netbox.DeviceType) (err error) {
	// Loading the environment variables from '.env' file.
	log.Info().Msgf("%+v", "Loading config from .env file")
	ie, err := nb.LoadEnvs()
	if err != nil {
		log.Error().Msgf("%+v", err)
		os.Exit(1)
	}

	nogit := strings.TrimSuffix(ie.RepoUrl, ".git")
	dir2 := filepath.Base(nogit)
	devicetypespath := filepath.Join(".", dir2, "device-types")

	// create a map of manufacturers to import in case they do not already exist
	manufacturersToImport := make(map[string]string, 0)
	deviceTypesToImport := make(map[string]nb.DeviceType, 0)
	err = filepath.Walk(devicetypespath, nb.CreateDeviceTypeMap(manufacturersToImport, deviceTypesToImport, []string{}))
	if err != nil {
		return err
	}

	// manufactureres need to exist in order to create device types
	err = nb.CreateManufacturers(hpengi.NetboxClient, hpengi.context, manufacturersToImport)
	if err != nil {
		return err
	}

	// create the device types
	err = nb.CreateDeviceTypes(hpengi.NetboxClient, hpengi.context, deviceTypesToImport)
	if err != nil {
		return err
	}

	return nil

}

func (hpengi *Hpengi) createModuleTypesRequests(existing map[string]netbox.ModuleType) (err error) {
	// Loading the environment variables from '.env' file.
	log.Info().Msgf("%+v", "Loading config from .env file")
	ie, err := nb.LoadEnvs()
	if err != nil {
		log.Error().Msgf("%+v", err)
		os.Exit(1)
	}

	nogit := strings.TrimSuffix(ie.RepoUrl, ".git")
	dir2 := filepath.Base(nogit)
	moduletypespath := filepath.Join(".", dir2, "module-types")

	// create a map of manufacturers to import in case they do not already exist
	manufacturersToImport := make(map[string]string, 0)

	// create the module types
	moduleTypesToImport := make(map[string]nb.ModuleType, 0)
	err = filepath.Walk(moduletypespath, nb.CreateModuleTypeMap(manufacturersToImport, moduleTypesToImport, []string{}))
	if err != nil {
		return err
	}

	// manufactureres need to exist in order to create device types
	err = nb.CreateManufacturers(hpengi.NetboxClient, hpengi.context, manufacturersToImport)
	if err != nil {
		return err
	}

	err = nb.CreateModuleTypes(hpengi.NetboxClient, hpengi.context, moduleTypesToImport)
	if err != nil {
		return err
	}

	return nil

}

func (hpengi *Hpengi) createModuleTypes(existing map[string]netbox.ModuleType) (err error) {
	// Loading the environment variables from '.env' file.
	log.Info().Msgf("%+v", "Loading config from .env file")
	ie, err := nb.LoadEnvs()
	if err != nil {
		log.Error().Msgf("%+v", err)
		os.Exit(1)
	}

	nogit := strings.TrimSuffix(ie.RepoUrl, ".git")
	dir2 := filepath.Base(nogit)
	path := filepath.Join(".", dir2, "module-types")

	manufacturersToImport := map[string]string{}
	// create a map of manufacturers to import in case they do not already exist
	moduleTypesToImport := make(map[string]nb.ModuleType, 0)
	err = filepath.Walk(path, nb.CreateModuleTypeMap(manufacturersToImport, moduleTypesToImport, []string{}))
	if err != nil {
		return err
	}
	return nil
}

func (hpengi *Hpengi) createRackRequests(existing map[string]netbox.Rack) (racks map[string]netbox.ApiDcimRacksCreateRequest, err error) {
	// create a map of existing racks for comparison
	if len(existing) == 0 {
		log.Info().Msgf("No racks to create")
	}

	log.Info().Msgf("Found %+v existing racks", len(existing))
	// get and create racks
	racks = make(map[string]netbox.ApiDcimRacksCreateRequest, 0)
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
		req := hpengi.NetboxClient.DcimAPI.DcimRacksCreate(hpengi.context).WritableRackRequest(*rack)

		// create the rack request if the rack does not exist
		_, ok := existing[*topo.Location.Rack]
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

	return racks, nil
}

func (hpengi *Hpengi) createDevicesRequests() (devices map[string]netbox.ApiDcimDevicesCreateRequest, err error) {
	existingDeviceTypes, err := getExistingDeviceTypes(hpengi.NetboxClient, hpengi.context)
	if err != nil {
		return devices, err
	}
	existingRacks, err := getExistingRacks(hpengi.NetboxClient, hpengi.context)
	if err != nil {
		return devices, err
	}
	existingDevices, err := getExistingDevices(hpengi.NetboxClient, hpengi.context)
	if err != nil {
		return devices, err
	}

	// create devices requests
	devices = make(map[string]netbox.ApiDcimDevicesCreateRequest, 0)
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
		// check the arch first, and if that fails, check the model
		// finally, set an unknown device type so it can be added and someone can fix it later
		var existingDeviceType netbox.DeviceType
		var ok bool
		existingDeviceType, ok = existingDeviceTypes[*topo.Architecture]
		if ok {
			log.Trace().Msgf("Found a matching device type (%+v) for %+v", existingDeviceType.GetId(), *topo.Architecture)
			device.SetDeviceType(existingDeviceType.GetId())
		} else {
			existingDeviceType, ok2 := existingDeviceTypes[*topo.Model]
			if ok2 {
				log.Trace().Msgf("Found a secondary matching device type (%+v) for %+v", existingDeviceType.GetId(), *topo.Model)
				device.SetDeviceType(existingDeviceType.GetId())
			} else {
				log.Trace().Msgf("No matching device type (%+v) for %+v", existingDeviceType.GetId(), *topo.Model)
				unknown, err := hpengi.getUnknownDeviceTypeId()
				if err != nil {
					return devices, err
				}

				device.SetDeviceType(*unknown) // FIXME need to create unknown if it does not exist
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
				return devices, err
			}
			u = float64(e)
		}

		if *topo.Architecture != "pdu" && *topo.Architecture != "cmm" && *topo.Architecture != "cec" {
			// set the elevation
			device.SetPosition(float64(u))
			face, err := netbox.NewRackFaceFromValue("front")
			if err != nil {
				return devices, err
			}
			device.SetFace(*face)
		}

		// // get the device interfaces
		// limit := maxLimit
		// offset := 0
		// existingInterfaces := make(map[int32]netbox.Interface, 0)
		// existingConcat := make([]netbox.Interface, 0)
		// for {
		// 	req := hpengi.NetboxClient.DcimAPI.DcimInterfacesList(hpengi.context).Limit(int32(limit)).Offset(int32(offset))
		// 	paginated, resp, err := req.Execute()
		// 	if err != nil {
		// 		log.Error().Msgf("%+v: %+v", resp.Status, err)
		// 		return devices, err
		// 	}

		// 	// append the results to to the slice
		// 	existingConcat = append(existingConcat, paginated.Results...)

		// 	// no more pages to get
		// 	if paginated.Next.Get() == nil {
		// 		break
		// 	}

		// 	// parse the next url to get the limit and offset
		// 	var next *url.URL
		// 	next, err = url.Parse(paginated.GetNext())
		// 	if err != nil {
		// 		return devices, err
		// 	}
		// 	limitQ := next.Query().Get("limit")
		// 	offsetQ := next.Query().Get("offset")
		// 	limit, err = strconv.Atoi(limitQ)
		// 	if err != nil {
		// 		return devices, err
		// 	}
		// 	offset, err = strconv.Atoi(offsetQ)
		// 	if err != nil {
		// 		return devices, err
		// 	}
		// }

		// // add the entires to the map after the loop is over and all pages are fetched
		// for _, existing := range existingConcat {
		// 	log.Info().Msgf("interface %+v: %+v", existing.GetId(), existing.GetName())
		// 	existingInterfaces[existing.GetId()] = existing
		// }

		log.Trace().Msgf("Creating request for %+v", *topo.CommonName)
		// create the request using the crafted rack
		req := hpengi.NetboxClient.DcimAPI.DcimDevicesCreate(hpengi.context).WritableDeviceWithConfigContextRequest(*device)

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
	return devices, nil
}

func (hpengi *Hpengi) executeCreateDeviceTypesRequests(reqs map[string]netbox.ApiDcimDeviceTypesCreateRequest) (err error) {
	// execute all the rack requests, creating them in netbox
	for name, req := range reqs {
		log.Info().Msgf("Executing a create device type request: %+v", name)
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
	return nil
}

func (hpengi *Hpengi) executeCreateRackRequests(reqs map[string]netbox.ApiDcimRacksCreateRequest) (err error) {
	// execute all the rack requests, creating them in netbox
	for name, req := range reqs {
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
	return nil
}

func (hpengi *Hpengi) executeCreateDevicesRequests(reqs map[string]netbox.ApiDcimDevicesCreateRequest) (err error) {
	// execute all the rack requests, creating them in netbox
	for name, req := range reqs {
		log.Info().Msgf("Executing a create device request for: %+v", name)
		_, resp, err := req.Execute()
		if err != nil {
			body, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			if resp.StatusCode == 201 { // 201 is created
				continue
			}
			log.Error().Msgf("%+v: %+v", resp.Status, string(body))
			return err
		}
	}
	return nil
}

func (hpengi *Hpengi) getUnknownDeviceTypeId() (id *int32, err error) {
	// get the device type id from the map
	req := hpengi.NetboxClient.DcimAPI.DcimDeviceTypesList(hpengi.context).Limit(int32(1)).Offset(int32(0)).Q("unknown")
	paginated, resp, err := req.Execute()
	if err != nil {
		log.Error().Msgf("%+v: %+v", resp.Status, err)
		return nil, err
	}

	// check if the unknown device type exists and return its id
	for _, existing := range paginated.Results {
		if existing.GetModel() == "unknown" {
			i := existing.GetId()
			id = &i
			return &i, nil
		}
	}
	return nil, nil
}

func (hpengi *Hpengi) createDeviceInterface(ports []canu.PaddleTopologyElemPortsElem, d netbox.Device) (err error) {
	i := netbox.NewWritableInterfaceRequestWithDefaults()
	// required:
	//    - device
	i.SetDevice(d.GetId())
	//    - name
	i.SetName("eth0")
	//    - type
	t := netbox.NewInterfaceTypeWithDefaults()
	t.SetLabel("Ethernet 1000BASE-T (1GE)")
	t.SetValue("Ethernet 1000BASE-T (1GE)")
	i.SetType(t.GetValue())
	req := hpengi.NetboxClient.DcimAPI.DcimInterfacesCreate(hpengi.context).WritableInterfaceRequest(*i)
	_, resp, err := req.Execute()
	if err != nil {
		log.Error().Msgf("%+v: %+v", resp.Status, err)
		return err
	}

	return nil
}
