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
package export

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// assignIPToInterfaces creates IPAddressToInterface join records for each
// interface the IP is assigned to.
func (e *Exporter) assignIPToInterfaces(
	ctx context.Context,
	ipNautobotID uuid.UUID,
	addr *devicetypes.CaniIPAddress,
	inventory *devicetypes.Inventory,
	deviceMap map[string]uuid.UUID,
	result *LoadResult,
) {
	if len(addr.Interfaces) == 0 {
		return
	}

	for _, ifaceID := range addr.Interfaces {
		// Resolve the interface from inventory
		iface, ok := inventory.Interfaces[ifaceID]
		if !ok || iface == nil {
			continue
		}

		// Find the device's Nautobot ID
		device, deviceOk := inventory.Devices[iface.DeviceID]
		if !deviceOk || device == nil {
			continue
		}
		deviceNautobotID, ok := deviceMap[device.Name]
		if !ok || deviceNautobotID == uuid.Nil {
			continue
		}

		// Find the interface in Nautobot
		nautobotIface, err := e.Cache.GetInterfaceByDeviceAndName(deviceNautobotID, iface.Name)
		if err != nil || nautobotIface == nil {
			// Try fuzzy match
			nautobotIface, err = e.Cache.GetInterfaceByDeviceAndNameFuzzy(deviceNautobotID, iface.Name)
			if err != nil || nautobotIface == nil {
				clog.Detail("  IP %s: interface %s on %s not found in Nautobot", addr.Address, iface.Name, device.Name)
				continue
			}
		}

		if e.Options.DryRun {
			clog.DryRun("Would assign IP %s to interface %s:%s", addr.Address, device.Name, iface.Name)
			continue
		}

		// Create the IP-to-interface assignment
		assignReq := nautobotapi.IPAddressToInterfaceRequest{}
		if err := setRefID(&assignReq.IpAddress, ipNautobotID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("ip %s: set assignment IP reference: %v", addr.Address, err))
			continue
		}
		if err := setRefID(&assignReq.Interface, nautobotIface.ID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("ip %s: set assignment interface reference: %v", addr.Address, err))
			continue
		}

		httpResp, err := e.Client.IpamIpAddressToInterfaceCreate(
			ctx,
			&nautobotapi.IpamIpAddressToInterfaceCreateParams{},
			assignReq,
		)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("ip %s: interface assignment error: %v", addr.Address, err))
			continue
		}
		respBody, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		if httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusOK {
			// 400 with "already exists" is acceptable (idempotency)
			if httpResp.StatusCode == http.StatusBadRequest {
				clog.Detail("  IP %s already assigned to %s:%s", addr.Address, device.Name, iface.Name)
				continue
			}
			result.Errors = append(result.Errors,
				fmt.Sprintf("ip %s: assignment to %s:%s returned %d: %s",
					addr.Address, device.Name, iface.Name, httpResp.StatusCode, string(respBody)))
		}
	}
}
