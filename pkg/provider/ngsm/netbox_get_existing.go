/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
package ngsm

import (
	"context"
	"io"
	"net/url"
	"strconv"

	"github.com/netbox-community/go-netbox/v3"
	"github.com/rs/zerolog/log"
)

const (
	maxLimit = 1000
)

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
			if resp != nil {
				body, _ := io.ReadAll(resp.Body)
				defer resp.Body.Close()
				log.Error().Msgf("%v %+v: %+v", string(body), resp.Status, err)
			}
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
		paginated, _, err := req.Execute()
		if err != nil {
			// log.Error().Msgf("%+v: %+v", resp.Status, err)
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

	return existingDevices, nil
}
