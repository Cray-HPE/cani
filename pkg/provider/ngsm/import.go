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
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// Import
func (ngsm *Ngsm) Import(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	log.Warn().Msgf("Importing to netbox is not yet implemented")

	return nil
}

func (ngsm *Ngsm) getUnknownDeviceTypeId() (id *int32, err error) {
	// get the device type id from the map
	req := ngsm.NetboxClient.DcimAPI.DcimDeviceTypesList(ngsm.context).Limit(int32(1)).Offset(int32(0)).Q("unknown")
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
