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
	"io"

	"github.com/netbox-community/go-netbox/v3"
	"github.com/rs/zerolog/log"
)

// executeCreateRackRequests executes the create rack requests, usually from a
// queue.  This allows for fewer API calls by grouping them together
func (ngsm *Ngsm) executeCreateRackRequests(reqs map[string]netbox.ApiDcimRacksCreateRequest) (err error) {
	// execute all the rack requests, creating them in netbox
	for name, req := range reqs {
		log.Info().Msgf("Creating a rack: %+v", name)
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

// executeCreateDevicesRequests executes the create rack requests, usually from a
// queue.  This allows for fewer API calls by grouping them together
func (ngsm *Ngsm) executeCreateDevicesRequests(reqs map[string]netbox.ApiDcimDevicesCreateRequest) (err error) {
	// execute all the rack requests, creating them in netbox
	for name, req := range reqs {
		log.Info().Msgf("Creating a device: %+v", name)
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
