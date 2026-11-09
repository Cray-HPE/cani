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

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
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

// assignIPToInterfaces creates an IPAddressToInterface join record for each
// interface the address is assigned to. An address that already existed in
// Nautobot may already carry some of them: those are read first and counted
// as skipped, so a repeated export duplicates nothing and a dry run against
// an existing address reports only what is missing. An address this run
// created or planned can have none and is not queried. Nautobot's join table
// is not change-logged, so this skip is the only record of the comparison.
func (e *Exporter) assignIPToInterfaces(
	ctx context.Context,
	ipNautobotID uuid.UUID,
	existed bool,
	addr *devicetypes.CaniIPAddress,
	inventory *devicetypes.Inventory,
	deviceMap map[string]uuid.UUID,
	result *LoadResult,
) {
	assigned := map[uuid.UUID]bool{}
	if existed {
		var err error
		if assigned, err = e.listIPAssignments(ctx, ipNautobotID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("ip %s: list interface assignments: %v", addr.Address, err))
			return
		}
	}
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
		if assignment.target != nil && assigned[assignment.target.ID] {
			clog.Detail("  IP %s already assigned to %s:%s", addr.Address, device.Name, iface.Name)
			result.IPAssignmentsSkipped++
			continue
		}
		if e.Options.DryRun {
			clog.DryRun("Would assign IP %s to interface %s:%s", addr.Address, device.Name, iface.Name)
			result.IPAssignmentsCreated++
			continue
		}
		if err := e.createIPAssignment(ctx, ipNautobotID, assignment.target.ID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("ip %s: assignment to %s:%s: %v", addr.Address, device.Name, iface.Name, err))
			continue
		}
		result.IPAssignmentsCreated++
	}
}

// listIPAssignments returns the Nautobot interface IDs an address is
// already assigned to. One page is enough: an address is on a handful of
// ports at most, the same bound the interface lookups use.
func (e *Exporter) listIPAssignments(ctx context.Context, ipNautobotID uuid.UUID) (map[uuid.UUID]bool, error) {
	filter, limit := []openapi_types.UUID{openapi_types.UUID(ipNautobotID)}, 1000
	resp, err := e.Client.IpamIpAddressToInterfaceListWithResponse(ctx, &nautobotapi.IpamIpAddressToInterfaceListParams{
		IpAddress: &filter, Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	assigned := make(map[uuid.UUID]bool, len(resp.JSON200.Results))
	for i := range resp.JSON200.Results {
		if iface := resp.JSON200.Results[i].Interface; iface != nil && iface.Id != nil {
			if id := refID(iface.Id); id != uuid.Nil {
				assigned[id] = true
			}
		}
	}
	return assigned, nil
}

// createIPAssignment posts one join record. Any status other than 201 is an
// error: a 400 here means a malformed reference, not an existing assignment,
// which listIPAssignments already ruled out.
func (e *Exporter) createIPAssignment(ctx context.Context, ipNautobotID, ifaceNautobotID uuid.UUID) error {
	req := nautobotapi.IPAddressToInterfaceRequest{}
	if err := setRefID(&req.IpAddress, ipNautobotID); err != nil {
		return fmt.Errorf("set assignment IP reference: %w", err)
	}
	if err := setRefID(&req.Interface, ifaceNautobotID); err != nil {
		return fmt.Errorf("set assignment interface reference: %w", err)
	}
	httpResp, err := e.Client.IpamIpAddressToInterfaceCreate(ctx, &nautobotapi.IpamIpAddressToInterfaceCreateParams{}, req)
	if err != nil {
		return fmt.Errorf("interface assignment error: %w", err)
	}
	respBody, _ := io.ReadAll(httpResp.Body)
	httpResp.Body.Close()
	if httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("returned %d: %s", httpResp.StatusCode, string(respBody))
	}
	return nil
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
