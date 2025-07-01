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
	"os"

	"github.com/gocarina/gocsv"
	"github.com/netbox-community/go-netbox/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// loadBoms loads the bom files into the expert bom map
func (ngsm *Ngsm) loadBoms(cmd *cobra.Command, args []string) (err error) {
	// get existing decive types, racks, and devices from netbox
	// we need these to compare against the bom files to see if we need to create
	// new objects in netbox or of they already exist
	// this provides idempotency so we don't create duplicates
	existingDeviceTypes, existingRacks, existingDevices, err := ngsm.getExistingFromNetbox()
	if err != nil {
		return err
	}

	log.Info().Msgf("Found %+v existing racks", len(existingRacks))
	log.Info().Msgf("Found %+v existing device types", len(existingDeviceTypes))
	log.Info().Msgf("Found %+v existing devices", len(existingDevices))

	// multiple bom flags can be passed, so iterate over them
	// this allows more than one bom to be parsed at a time
	// a customer may have one bom from the initial order, which may later be
	// supplemented with additional orders, so this can combine them all
	boms, _ := cmd.Flags().GetStringArray("bom")
	for _, bom := range boms {
		// get all the rows from the current bom
		rows, err := ngsm.getRowsFromBom(bom)
		if err != nil {
			return err
		}

		// make a queue for each bom.  these will be combined into a single Queue
		// which will create the everything in netbox using fewer API requests
		q, err := newQueue(bom, existingDeviceTypes, existingRacks, existingDevices)
		if err != nil {
			return err
		}

		// add the rows to the queue
		for n, row := range rows {
			// parse each row and sanitize the cells
			r, err := row.Sanitize()
			if err != nil {
				return err
			}

			// set the row number, +1 for 0 based index, +1 for header row
			// TODO: there is likely a better way to do this
			row.SetRow(n + 2)

			// add the row to the queue if it is valid
			err = q.AddRow(r)
			if err != nil {
				return err
			}
		}

		// add the queue to the Queue
		ngsm.Boms[bom] = q
	}
	return nil
}

// getExistingFromNetbox gets the existing device types, racks, and devices from
// netbox.  these are needed for comparison when creating new netbox objects so
// we are not needlessly trying to create duplicates. The devicetypes in
// particular are needed to get the device type id so the requests can be queued
// up and executed at once instead of one at a time this reduces the number of
// requests to netbox and speeds up the process
func (ngsm *Ngsm) getExistingFromNetbox() (existingDeviceTypes map[string]netbox.DeviceType, existingRacks map[string]netbox.Rack, existingDevices map[string]netbox.DeviceWithConfigContext, err error) {
	// get device types
	existingDeviceTypes, err = getExistingDeviceTypes(ngsm.NetboxClient, ngsm.context)
	if err != nil {
		return existingDeviceTypes, existingRacks, existingDevices, err
	}
	// get racks
	existingRacks, err = getExistingRacks(ngsm.NetboxClient, ngsm.context)
	if err != nil {
		return existingDeviceTypes, existingRacks, existingDevices, err
	}
	// get devices
	existingDevices, err = getExistingDevices(ngsm.NetboxClient, ngsm.context)
	if err != nil {
		return existingDeviceTypes, existingRacks, existingDevices, err
	}
	// TODO: get module types
	return existingDeviceTypes, existingRacks, existingDevices, nil
}

// getRowsFromBom unmarshals the rows from the bom file into a slice of rows
// this retains the rows order, which helps in adding things to the queue
//
//		TODO: gosv offers other unmarshal options, which may be helpful in
//		parsing the bom files, which are not completely standard and may have
//	  extra info at the top, instead of a header row
func (ngsm *Ngsm) getRowsFromBom(bom string) (rows []*Row, err error) {
	// open the bom file
	data, err := os.OpenFile(bom, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return rows, err
	}
	defer data.Close()

	// unmashal the csv into a slice of rows
	// TODO: investigate other unmarshal options
	rows = []*Row{}
	err = gocsv.UnmarshalFile(data, &rows)
	if err != nil {
		return rows, err
	}
	return rows, nil
}
