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
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// loadIPAddresses exports CaniIPAddress records to Nautobot and creates
// IP-to-interface assignments via the join table API.
func (e *Exporter) loadIPAddresses(
	ctx context.Context,
	inventory *devicetypes.Inventory,
	prefixMap map[uuid.UUID]uuid.UUID,
	deviceMap map[string]uuid.UUID,
	result *LoadResult,
) error {
	if len(inventory.IPAddresses) == 0 {
		return nil
	}

	clog.Header("Phase 9: IP Addresses (%d)", len(inventory.IPAddresses))

	// Resolve namespace once
	ns, err := e.Cache.GetOrCreateNamespace("Global")
	if err != nil {
		return fmt.Errorf("failed to resolve namespace: %w", err)
	}

	for _, addr := range inventory.IPAddresses {
		if addr == nil || addr.Address == "" {
			continue
		}

		// Check if IP address already exists
		existing, err := e.Cache.LookupIPAddress(addr.Address)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("ip %s: lookup error: %v", addr.Address, err))
			continue
		}
		if existing != nil {
			setExternalID(&addr.ExternalIDs, "nautobot", existing.ID)
			result.IPAddressesSkipped++
			// Still attempt interface assignments for existing IPs
			e.assignIPToInterfaces(ctx, existing.ID, addr, inventory, deviceMap, result)
			continue
		}

		nautobotID, err := e.createIPAddress(ctx, addr, ns.ID, prefixMap, result)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("ip %s: create error: %v", addr.Address, err))
			continue
		}

		setExternalID(&addr.ExternalIDs, "nautobot", nautobotID)
		e.Cache.CacheIPAddress(addr.Address, &CachedItem{
			ID:   nautobotID,
			Name: addr.Address,
		})
		result.IPAddressesCreated++

		// Assign to interfaces
		e.assignIPToInterfaces(ctx, nautobotID, addr, inventory, deviceMap, result)
	}

	clog.Info("  IP addresses created: %d", result.IPAddressesCreated)
	return nil
}

// createIPAddress creates a single IP address in Nautobot.
func (e *Exporter) createIPAddress(
	ctx context.Context,
	addr *devicetypes.CaniIPAddress,
	namespaceID uuid.UUID,
	prefixMap map[uuid.UUID]uuid.UUID,
	result *LoadResult,
) (uuid.UUID, error) {
	// Resolve status
	statusName := addr.Status
	if statusName == "" {
		statusName = e.Options.DefaultStatus
	}
	if statusName == "" {
		statusName = "Active"
	}
	statusItem, err := e.Cache.GetStatus(statusName)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to resolve status %q: %w", statusName, err)
	}

	req := nautobotapi.IPAddressRequest{
		Address: addr.Address,
		Status:  makeIDRef(statusItem.ID),
	}

	// Resolve parent prefix first
	if addr.Parent != uuid.Nil {
		if parentNID, ok := prefixMap[addr.Parent]; ok {
			parentRef := makeIPParentRef(parentNID)
			req.Parent = &parentRef
		}
	}

	// Always set namespace — Nautobot requires at least one of parent or namespace.
	nsRef := makeIPNamespaceRef(namespaceID)
	req.Namespace = &nsRef

	// Set type
	if addr.Type != "" {
		ipType := mapIPAddressType(addr.Type)
		req.Type = &ipType
	}

	// Set description
	if addr.Description != "" {
		req.Description = &addr.Description
	}

	// Set DNS name
	if addr.DNSName != "" {
		req.DnsName = &addr.DNSName
	}

	// Resolve role (IPRole maps to a Nautobot Role object)
	if addr.IPRole != "" {
		roleItem, err := e.Cache.GetRole(string(addr.IPRole))
		if err == nil && roleItem != nil {
			ref := makeTenantRef(roleItem.ID)
			req.Role = ref
		}
	}

	if e.Options.DryRun {
		clog.DryRun("Would create IP address: %s", addr.Address)
		return uuid.New(), nil
	}

	resp, err := e.Client.IpamIpAddressesCreate(ctx, &nautobotapi.IpamIpAddressesCreateParams{}, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return uuid.Nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var respObj struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &respObj); err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse response: %w", err)
	}
	nautobotID, err := uuid.Parse(respObj.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse UUID from response: %w", err)
	}

	clog.Created("  + IP: %s", addr.Address)
	return nautobotID, nil
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
		ifaceRef := makeTenantRef(nautobotIface.ID)
		assignReq := nautobotapi.IPAddressToInterfaceRequest{
			IpAddress: makeIDRef(ipNautobotID),
			Interface: ifaceRef,
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

// mapIPAddressType converts a cani IPAddressType to a Nautobot IPAddressTypeChoices.
func mapIPAddressType(t devicetypes.IPAddressType) nautobotapi.IPAddressTypeChoices {
	switch t {
	case devicetypes.IPAddressTypeHost:
		return nautobotapi.Host
	case devicetypes.IPAddressTypeDHCP:
		return nautobotapi.Dhcp
	case devicetypes.IPAddressTypeSLAAC:
		return nautobotapi.Slaac
	default:
		return nautobotapi.Host
	}
}
