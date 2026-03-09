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
package transform

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/logcolor"
	"github.com/google/uuid"
)

var clog = logcolor.New("[nautobot] ", false)

// Transform transforms devices in the queue into CANI's format
func Transform(existing devicetypes.Inventory) (transformed map[uuid.UUID]*devicetypes.CaniDeviceType, err error) {
	transformed, err = transform(existing)
	if err != nil {
		return nil, err
	}
	return transformed, nil
}

// extractDevicesFromBom extracts devices from the BOM files into CANI's inventory format
func transform(existing devicetypes.Inventory) (transformed map[uuid.UUID]*devicetypes.CaniDeviceType, err error) {
	transformed = make(map[uuid.UUID]*devicetypes.CaniDeviceType)

	for _, device := range existing.Devices {
		d := device // create a copy of d for this iteration
		transformed[d.ID] = d
	}

	clog.Plain("")
	clog.Detail("  %d devices existing in current inventory", len(existing.Devices))
	clog.Detail("  %d devices Transformed (not yet Loaded)", len(transformed))
	clog.Plain("")

	return transformed, nil
}
