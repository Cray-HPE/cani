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

// ipAssignment is one interface an address is assigned to, with the Nautobot
// interface it resolves to; target is nil when a dry run only planned the
// module that owns the port.
type ipAssignment struct {
	iface  *devicetypes.CaniInterface
	device *devicetypes.CaniDeviceType
	target *CachedItem
}

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
	for _, ifaceID := range addr.Interfaces {
		assignment, err := e.resolveIPAssignment(inventory, ifaceID, addr.Address, deviceMap)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("ip %s: %v", addr.Address, err))
			continue
		}
		if assignment == nil {
			continue
		}
		device, iface := assignment.device, assignment.iface
		if e.Options.DryRun {
			clog.DryRun("Would assign IP %s to interface %s:%s", addr.Address, device.Name, iface.Name)
			continue
		}
		nautobotIface := assignment.target

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

// resolveIPAssignment finds the Nautobot interface an address is assigned
// to: on its module when the port is module-owned, otherwise on its device
// (exact, then fuzzy). An interface the run cannot assign to (unknown or
// unexported device, interface missing in Nautobot) is logged and returned
// nil. A module-owned port whose module this run did not confirm is an
// error, not a device lookup, so the address is never attached to a
// same-named port of the parent device; a module a dry run only planned
// returns an assignment with no target for the preview.
func (e *Exporter) resolveIPAssignment(
	inventory *devicetypes.Inventory, ifaceID uuid.UUID, address string, deviceMap map[string]uuid.UUID,
) (*ipAssignment, error) {
	iface := inventory.Interfaces[ifaceID]
	if iface == nil {
		return nil, nil
	}
	device := inventory.Devices[iface.DeviceID]
	if device == nil {
		return nil, nil
	}
	deviceNautobotID, ok := deviceMap[device.Name]
	if !ok || deviceNautobotID == uuid.Nil {
		return nil, nil
	}
	assignment := &ipAssignment{iface: iface, device: device}
	module := owningModule(inventory, iface.OwnerID(), ifaceID)
	if e.Options.DryRun && module != nil && e.Cache.ModulePlanned(module.ID) {
		return assignment, nil
	}
	moduleID, err := e.confirmedModuleID(module, iface.Name, device.Name)
	if err != nil {
		return nil, err
	}
	assignment.target, err = e.resolveOwnedInterface(moduleID, deviceNautobotID, iface.Name)
	if err != nil || assignment.target == nil {
		clog.Detail("  IP %s: interface %s on %s not found in Nautobot", address, iface.Name, device.Name)
		return nil, nil
	}
	return assignment, nil
}
