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
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// loadPrefixes exports CaniPrefix records to Nautobot.
// Prefixes are sorted by prefix length (shortest first) to ensure parent
// prefixes exist before their children.
// Returns a map of cani prefix ID → Nautobot prefix ID for downstream use.
func (e *Exporter) loadPrefixes(
	ctx context.Context,
	inventory *devicetypes.Inventory,
	locationMap map[uuid.UUID]uuid.UUID,
	vlanMap map[uuid.UUID]uuid.UUID,
	result *LoadResult,
) (map[uuid.UUID]uuid.UUID, error) {
	created := make(map[uuid.UUID]uuid.UUID)

	if len(inventory.Prefixes) == 0 {
		return created, nil
	}

	clog.Header("Phase 8: Prefixes (%d)", len(inventory.Prefixes))

	// Sort by prefix length ascending (containers/wider first, then narrower)
	ordered := sortPrefixesByLength(inventory.Prefixes)

	// Resolve namespace once for all prefixes
	ns, err := e.Cache.GetOrCreateNamespace("Global")
	if err != nil {
		return created, fmt.Errorf("failed to resolve namespace: %w", err)
	}

	for _, prefix := range ordered {
		if prefix == nil || prefix.Prefix == "" {
			continue
		}

		// Check if prefix already exists
		existing, err := e.Cache.LookupPrefix(prefix.Prefix)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("prefix %s: lookup error: %v", prefix.Prefix, err))
			continue
		}
		if existing != nil {
			created[prefix.ID] = existing.ID
			setExternalID(&prefix.ExternalIDs, "nautobot", existing.ID)
			result.PrefixesSkipped++
			continue
		}

		nautobotID, err := e.createPrefix(ctx, prefix, ns.ID, vlanMap, created, result)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("prefix %s: create error: %v", prefix.Prefix, err))
			continue
		}

		created[prefix.ID] = nautobotID
		setExternalID(&prefix.ExternalIDs, "nautobot", nautobotID)
		e.Cache.CachePrefix(prefix.Prefix, &CachedItem{
			ID:   nautobotID,
			Name: prefix.Prefix,
		})
		result.PrefixesCreated++
	}

	clog.Info("  Prefixes created: %d", result.PrefixesCreated)
	return created, nil
}

// createPrefix creates a single prefix in Nautobot.
func (e *Exporter) createPrefix(
	ctx context.Context,
	prefix *devicetypes.CaniPrefix,
	namespaceID uuid.UUID,
	vlanMap map[uuid.UUID]uuid.UUID,
	createdPrefixes map[uuid.UUID]uuid.UUID,
	result *LoadResult,
) (uuid.UUID, error) {
	// Resolve status
	statusName := prefix.Status
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

	req := nautobotapi.WritablePrefixRequest{
		Prefix: prefix.Prefix,
		Status: makeIDRef(statusItem.ID),
	}

	// Set namespace
	nsRef := makeIDRef(namespaceID)
	req.Namespace = &nsRef

	// Set type
	if prefix.Type != "" {
		prefixType := mapPrefixType(prefix.Type)
		req.Type = &prefixType
	}

	// Set description
	if prefix.Description != "" {
		req.Description = &prefix.Description
	}

	// Note: Location is intentionally omitted for prefixes because Nautobot
	// restricts which location types may associate with prefixes. The location
	// type "Section" (used by cani) does not support prefix associations.

	// Resolve parent prefix
	if prefix.Parent != uuid.Nil {
		if parentNID, ok := createdPrefixes[prefix.Parent]; ok {
			parentRef := makePrefixParentRef(parentNID)
			req.Parent = &parentRef
		}
	}

	// Resolve VLAN
	if prefix.VLAN != uuid.Nil {
		if vlanNID, ok := vlanMap[prefix.VLAN]; ok {
			ref := makeTenantRef(vlanNID)
			req.Vlan = ref
		}
	}

	// Resolve role
	if prefix.Role != "" {
		roleItem, err := e.Cache.GetRole(prefix.Role)
		if err == nil && roleItem != nil {
			ref := makeTenantRef(roleItem.ID)
			req.Role = ref
		}
	}

	if e.Options.DryRun {
		clog.DryRun("Would create prefix: %s", prefix.Prefix)
		return uuid.New(), nil
	}

	// Use raw API call because the generated response parser has UUID format
	// parsing issues with Nautobot's response (base64 vs string UUID mismatch).
	httpResp, err := e.Client.IpamPrefixesCreate(ctx, &nautobotapi.IpamPrefixesCreateParams{}, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}
	defer httpResp.Body.Close()

	body, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusCreated {
		return uuid.Nil, fmt.Errorf("unexpected status %d: %s", httpResp.StatusCode, string(body))
	}

	// Extract ID from response JSON
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

	clog.Created("  + Prefix: %s (%s)", prefix.Prefix, prefix.Type)
	return nautobotID, nil
}

// sortPrefixesByLength returns prefixes sorted by prefix length ascending
// (wider/container prefixes first, so parents are created before children).
func sortPrefixesByLength(prefixes map[uuid.UUID]*devicetypes.CaniPrefix) []*devicetypes.CaniPrefix {
	result := make([]*devicetypes.CaniPrefix, 0, len(prefixes))
	for _, p := range prefixes {
		if p != nil {
			result = append(result, p)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].PrefixLen < result[j].PrefixLen
	})
	return result
}

// mapPrefixType converts a cani PrefixType to a Nautobot PrefixTypeChoices.
func mapPrefixType(t devicetypes.PrefixType) nautobotapi.PrefixTypeChoices {
	switch t {
	case devicetypes.PrefixTypeContainer:
		return nautobotapi.PrefixTypeChoicesContainer
	case devicetypes.PrefixTypeNetwork:
		return nautobotapi.PrefixTypeChoicesNetwork
	case devicetypes.PrefixTypePool:
		return nautobotapi.PrefixTypeChoicesPool
	default:
		return nautobotapi.PrefixTypeChoicesNetwork
	}
}
