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
	"fmt"
	"path/filepath"
	"strings"

	"github.com/netbox-community/go-netbox/v3"
	"github.com/rs/zerolog/log"
)

// createDeviceReqFromRow creates a device request from a row
// at present, some information is hardcoded and some assumptions made
func (ngsm *Ngsm) createDeviceRequestFromRow(row Row, existingRacks map[string]netbox.Rack) (req netbox.ApiDcimDevicesCreateRequest, err error) {
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
	// set the notes to the comments field in netbox
	device.SetComments(fmt.Sprintf("sourced from: %s\nimported from row: %d\n%s", row.source, row.row, row.Notes))

	device.SetDeviceType(row.netboxDeviceTypeID)

	device.SetName(row.netboxName)

	for _, existingRack := range existingRacks {
		// if the rack name contains the source file name, associate the device with
		// the rack, even if it is an unracked device
		if strings.Contains(existingRack.GetName(), filepath.Base(row.source)) {
			log.Debug().Msgf("Associating %s to rack: %s (%d)", row.netboxName, existingRack.GetName(), existingRack.GetId())
			// TODO: get elevation somehow
			// TODO: check for more racks if rack is full
			device.SetRack(existingRack.GetId())
		}
	}

	// create the request using the crafted rack
	req = ngsm.NetboxClient.DcimAPI.DcimDevicesCreate(ngsm.context).WritableDeviceWithConfigContextRequest(*device)

	return req, nil
}

// createRackRequestFromRow creates a rack request from a row
func (ngsm *Ngsm) createRackRequestFromRow(row Row) (req netbox.ApiDcimRacksCreateRequest, err error) {
	// adding a rack requires:
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
	rack.SetName(row.netboxName)

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
	var height int32 = 42 // FIXME hardcoding for now

	// set the starting unit
	rack.SetStartingUnit(1) // FIXME hardcoding for now
	rack.SetUHeight(height) // FIXME hardcoding for now

	// create the request using the crafted rack
	req = ngsm.NetboxClient.DcimAPI.DcimRacksCreate(ngsm.context).WritableRackRequest(*rack)

	return req, nil
}
