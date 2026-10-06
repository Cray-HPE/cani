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
// IP-to-interface assignments via the join table API. An address takes its
// namespace from its parent prefix, so an address without a containing
// prefix in the inventory is reported instead of being filed under Global.
// Returns a map of cani address ID → Nautobot address ID for the primary-IP
// phase, which must not depend on external IDs a dry run does not record.
func (e *Exporter) loadIPAddresses(
	ctx context.Context,
	inventory *devicetypes.Inventory,
	prefixMap map[uuid.UUID]uuid.UUID,
	deviceMap map[string]uuid.UUID,
	result *LoadResult,
) (map[uuid.UUID]uuid.UUID, error) {
	loaded := make(map[uuid.UUID]uuid.UUID)
	if len(inventory.IPAddresses) == 0 {
		return loaded, nil
	}

	clog.Header("Phase 9: IP Addresses (%d)", len(inventory.IPAddresses))

	for _, addr := range inventory.IPAddresses {
		if addr == nil || addr.Address == "" {
			continue
		}
		parent := inventory.Prefixes[addr.Parent]
		if parent == nil {
			result.Errors = append(result.Errors, fmt.Sprintf(
				"ip %s: no containing prefix in namespace %q; add the prefix before exporting",
				addr.Address, inventory.IPAddressNamespace(addr)))
			continue
		}
		nautobotID, err := e.loadIPAddress(ctx, addr, parent.EffectiveNamespace(), prefixMap, result)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("ip %s/%s: %v", parent.EffectiveNamespace(), addr.Address, err))
			continue
		}
		loaded[addr.ID] = nautobotID
		e.recordExternalID(&addr.ExternalIDs, nautobotID)
		e.assignIPToInterfaces(ctx, nautobotID, addr, inventory, deviceMap, result)
	}

	clog.Info("  IP addresses created: %d", result.IPAddressesCreated)
	return loaded, nil
}

// loadIPAddress finds or creates one address in the namespace and returns
// its Nautobot ID. Finding an existing address counts as skipped.
func (e *Exporter) loadIPAddress(
	ctx context.Context,
	addr *devicetypes.CaniIPAddress,
	namespace string,
	prefixMap map[uuid.UUID]uuid.UUID,
	result *LoadResult,
) (uuid.UUID, error) {
	ns, exists, err := e.ensureNamespace(namespace)
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve namespace: %w", err)
	}
	var existing *CachedItem
	if exists {
		if existing, err = e.Cache.LookupIPAddress(namespace, addr.Address); err != nil {
			return uuid.Nil, fmt.Errorf("lookup error: %w", err)
		}
	}
	if existing != nil {
		result.IPAddressesSkipped++
		return existing.ID, nil
	}
	nautobotID, err := e.createIPAddress(ctx, addr, ns.ID, prefixMap, result)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create error: %w", err)
	}
	e.Cache.CacheIPAddress(namespace, addr.Address, &CachedItem{ID: nautobotID, Name: addr.Address})
	result.IPAddressesCreated++
	return nautobotID, nil
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
	}
	if err := setRefID(&req.Status, statusItem.ID); err != nil {
		return uuid.Nil, fmt.Errorf("set IP address status reference: %w", err)
	}

	// Resolve parent prefix first
	if addr.Parent != uuid.Nil {
		if parentNID, ok := prefixMap[addr.Parent]; ok {
			if err := setRefID(&req.Parent, parentNID); err != nil {
				return uuid.Nil, fmt.Errorf("set IP address parent reference: %w", err)
			}
		}
	}

	// Always set namespace — Nautobot requires at least one of parent or namespace.
	if err := setRefID(&req.Namespace, namespaceID); err != nil {
		return uuid.Nil, fmt.Errorf("set IP address namespace reference: %w", err)
	}

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
			if err := setRefID(&req.Role, roleItem.ID); err != nil {
				return uuid.Nil, fmt.Errorf("set IP address role reference: %w", err)
			}
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

// mapIPAddressType converts a cani IPAddressType to a Nautobot IPAddressTypeChoices.
func mapIPAddressType(t devicetypes.IPAddressType) nautobotapi.IPAddressTypeChoices {
	switch t {
	case devicetypes.IPAddressTypeHost:
		return nautobotapi.IPAddressTypeChoicesHost
	case devicetypes.IPAddressTypeDHCP:
		return nautobotapi.IPAddressTypeChoicesDhcp
	case devicetypes.IPAddressTypeSLAAC:
		return nautobotapi.IPAddressTypeChoicesSlaac
	default:
		return nautobotapi.IPAddressTypeChoicesHost
	}
}
