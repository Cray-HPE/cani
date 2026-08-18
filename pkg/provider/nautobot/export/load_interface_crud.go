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
	"context"
	"fmt"
	"net/http"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// statusRef resolves the named Nautobot status and builds a status reference.
func (e *Exporter) statusRef(name string) (nautobotapi.BulkWritableCableRequestStatus, error) {
	var status nautobotapi.BulkWritableCableRequestStatus
	statusItem, err := e.Cache.GetStatus(name)
	if err != nil {
		return status, fmt.Errorf("failed to get %s status: %w", name, err)
	}
	var statusIDUnion nautobotapi.BulkWritableCableRequestStatusId
	if err := statusIDUnion.FromBulkWritableCableRequestStatusId0(statusItem.ID); err != nil {
		return status, fmt.Errorf(errFmtCreateStatusID, err)
	}
	status.Id = &statusIDUnion
	return status, nil
}

// createInterface creates a single interface on a device
func (e *Exporter) createInterface(ctx context.Context, deviceID uuid.UUID, iface interfaceSpec, result *LoadResult) error {
	if e.Options.DryRun {
		clog.DryRun("Would create interface: %s", iface.Name)
		result.IfacesCreated++
		return nil
	}

	// Build device reference with proper union type for ID
	var deviceIDUnion nautobotapi.BulkWritableCableRequestStatusId
	if err := deviceIDUnion.FromBulkWritableCableRequestStatusId0(deviceID); err != nil {
		return fmt.Errorf("failed to create device ID: %w", err)
	}
	deviceRef := &nautobotapi.BulkWritableCircuitRequestTenant{
		Id: &deviceIDUnion,
	}

	// Build status reference - get "Active" status
	status, err := e.statusRef("Active")
	if err != nil {
		return err
	}

	// Build interface request
	ifaceType := nautobotapi.InterfaceTypeChoices(iface.Type)
	mgmtOnly := iface.MgmtOnly
	req := nautobotapi.WritableInterfaceRequest{
		Device:   deviceRef,
		Name:     iface.Name,
		Type:     ifaceType,
		Status:   status,
		MgmtOnly: &mgmtOnly,
	}
	req.Tags = e.Cache.resolveTagRefs(iface.Tags)

	if iface.Mac != "" {
		mac := iface.Mac
		req.MacAddress = &mac
	}

	if iface.Description != "" {
		desc := iface.Description
		req.Description = &desc
	}

	role, err := e.roleRef(iface.Role)
	if err != nil {
		return err
	}
	req.Role = role

	resp, err := e.Client.DcimInterfacesCreateWithResponse(ctx, &nautobotapi.DcimInterfacesCreateParams{}, req)
	if err != nil {
		return fmt.Errorf(errFmtAPIError, err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return fmt.Errorf(errFmtUnexpectedStatus, resp.StatusCode(), string(resp.Body))
	}

	// Cache the newly created interface for cable creation
	if resp.JSON201 != nil && resp.JSON201.Id != nil {
		cachedItem := &CachedItem{
			ID:      uuid.UUID(*resp.JSON201.Id),
			Name:    iface.Name,
			Display: iface.Name,
		}
		e.Cache.CacheInterface(deviceID, iface.Name, cachedItem)
	}

	result.IfacesCreated++
	return nil
}

// updateInterface updates an existing interface in Nautobot
func (e *Exporter) updateInterface(ctx context.Context, interfaceID uuid.UUID, deviceID uuid.UUID, iface interfaceSpec, result *LoadResult) error {
	if e.Options.DryRun {
		clog.DryRun("Would update interface: %s", iface.Name)
		return nil
	}

	// Build status reference - get "Active" status
	status, err := e.statusRef("Active")
	if err != nil {
		return err
	}

	// Build device reference (required by Nautobot API)
	var deviceIDUnion nautobotapi.BulkWritableCableRequestStatusId
	if err := deviceIDUnion.FromBulkWritableCableRequestStatusId0(deviceID); err != nil {
		return fmt.Errorf("failed to create device ID: %w", err)
	}
	device := &nautobotapi.BulkWritableCircuitRequestTenant{
		Id: &deviceIDUnion,
	}

	// Build patch request - update type, status, and device
	ifaceType := nautobotapi.InterfaceTypeChoices(iface.Type)
	mgmtOnly := iface.MgmtOnly
	req := nautobotapi.PatchedWritableInterfaceRequest{
		Device:   device,
		Type:     &ifaceType,
		Status:   &status,
		MgmtOnly: &mgmtOnly,
	}
	req.Tags = e.Cache.resolveTagRefs(iface.Tags)

	if iface.Mac != "" {
		mac := iface.Mac
		req.MacAddress = &mac
	}

	// Send description unconditionally so an emptied local value clears it in
	// Nautobot; the inventory is authoritative on reconcile.
	desc := iface.Description
	req.Description = &desc

	role, err := e.roleRef(iface.Role)
	if err != nil {
		return err
	}
	req.Role = role

	resp, err := e.Client.DcimInterfacesPartialUpdateWithResponse(ctx, interfaceID, &nautobotapi.DcimInterfacesPartialUpdateParams{}, req)
	if err != nil {
		return fmt.Errorf(errFmtAPIError, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf(errFmtUnexpectedStatus, resp.StatusCode(), string(resp.Body))
	}

	return nil
}
