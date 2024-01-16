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
package ngsm

import (
	nb "github.com/Cray-HPE/cani/pkg/netbox"
	"github.com/netbox-community/go-netbox/v3"
	"github.com/rs/zerolog/log"
)

// importBoms imports the HBOMs into netbox
func (ngsm *Ngsm) importBoms() (err error) {
	// setup a new client for communicating with netbox
	ngsm.NetboxClient, ngsm.context, err = nb.NewClient()
	if err != nil {
		return err
	}

	// create the rack requests first since they are parents for devices
	createRacksRequests, err := ngsm.createRackRequestsFromBom()
	if err != nil {
		return err
	}

	log.Info().Msgf("Need to create %+v new racks", len(createRacksRequests))

	// excute the queued requests for racks
	err = ngsm.executeCreateRackRequests(createRacksRequests)
	if err != nil {
		return err
	}

	// create devices now that any new racks are created
	createDeviceRequests, err := ngsm.createDeviceRequestsFromBom()
	if err != nil {
		return err
	}

	log.Info().Msgf("Need to create %+v new devices", len(createDeviceRequests))

	// excute the queued requests for devices
	err = ngsm.executeCreateDevicesRequests(createDeviceRequests)
	if err != nil {
		return err
	}
	return nil
}

// expertBomToNetbox converts the BOMs to netbox requests
func (ngsm *Ngsm) createDeviceRequestsFromBom() (reqs map[string]netbox.ApiDcimDevicesCreateRequest, err error) {
	reqs = make(map[string]netbox.ApiDcimDevicesCreateRequest)

	// get existing racks from netbox so we can use their IDs to associate devices
	// with them it is possible some were created during the import, so we need to
	// check again here
	_, existingRacks, _, err := ngsm.getExistingFromNetbox()
	if err != nil {
		return reqs, err
	}

	for _, q := range ngsm.Boms {
		// add any updated racks to the queue
		q.existingRacks = existingRacks

		for _, row := range q.devicesToCreate {
			// add the request to the map for execution later
			_, exists := reqs[row.netboxName]
			if !exists {
				_, ok := q.existingRacks[row.netboxName]
				if ok {
					log.Info().Msgf("need to set row to use existing rack: %+v", q.existingRacks[row.netboxName])
				}
				// create a device request from the information in the row
				req, err := ngsm.createDeviceRequestFromRow(row, existingRacks)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to create request for device %+v", row.netboxName)
					return reqs, err
				}
				reqs[row.netboxName] = req
			} else {
				// do not try to create if the device already exists
				log.Info().Msgf("%+v already exists in netbox", row.netboxName)
			}
		}
	}
	return reqs, nil
}

// createRackRequestsFromBom create a map of rack creation requests
func (ngsm *Ngsm) createRackRequestsFromBom() (reqs map[string]netbox.ApiDcimRacksCreateRequest, err error) {
	reqs = make(map[string]netbox.ApiDcimRacksCreateRequest)
	for _, q := range ngsm.Boms {
		for _, row := range q.racksToCreate {
			// FIXME, name must be unique per location
			// so need more checks when creating the rack request from the row and set the location
			// without this check, the rack will be created in the default location and multiples may be created
			_, exists := q.existingRacks[row.netboxName]
			if !exists {
				req, err := ngsm.createRackRequestFromRow(row)
				if err != nil {
					return reqs, err
				}
				reqs[row.netboxName] = req
			}
		}
	}
	return reqs, nil
}
