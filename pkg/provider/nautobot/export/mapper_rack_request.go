/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
package export

import (
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// MapToWritableRackRequest converts a CaniDeviceType (rack) to a WritableRackRequest
func (m *DeviceMapper) MapToWritableRackRequest(device *devicetypes.CaniDeviceType) (*nautobotapi.WritableRackRequest, error) {
	if device == nil {
		return nil, errors.New(errMsgDeviceNil)
	}

	// Resolve location
	location, err := m.resolveLocation(device)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve location for rack %s: %w", device.Name, err)
	}

	// Resolve status
	status, err := m.resolveStatus(device)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve status for rack %s: %w", device.Name, err)
	}

	// Build the request
	req := &nautobotapi.WritableRackRequest{
		Name: device.Name,
	}
	if err := setRefID(&req.Location, location.ID); err != nil {
		return nil, fmt.Errorf("set location reference for rack %s: %w", device.Name, err)
	}
	if err := setRefID(&req.Status, status.ID); err != nil {
		return nil, fmt.Errorf("set status reference for rack %s: %w", device.Name, err)
	}

	// Set rack height (default to 48U if not specified)
	uHeight := 48
	if device.ProviderMetadata != nil {
		if h, _ := device.GetProviderMeta("u_height"); h != nil {
			if hInt, ok := h.(int); ok && hInt > 0 {
				uHeight = hInt
			}
		}
	}
	req.UHeight = &uHeight

	// Map optional fields
	if flat := device.FlattenProviderMetadata(); len(flat) > 0 {
		customFields := make(map[string]interface{}, len(flat))
		for k, v := range flat {
			if k != "u_height" && k != "rack_position" {
				customFields[k] = v
			}
		}
		if len(customFields) > 0 {
			req.CustomFields = toNautobotCustomFields(customFields)
		}
	}

	return req, nil
}
