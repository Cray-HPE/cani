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
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// loadPrefixes exports CaniPrefix records to Nautobot.
// Prefixes are sorted by prefix length (shortest first) to ensure parent
// prefixes exist before their children. Each prefix is resolved and created
// inside its own namespace; a lookup or namespace failure is reported rather
// than redirected to Global.
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

	for _, prefix := range ordered {
		if prefix == nil || prefix.Prefix == "" {
			continue
		}
		remote, err := e.loadPrefix(ctx, prefix, inventory, locationMap, vlanMap, created, result)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("prefix %s/%s: %v", prefix.EffectiveNamespace(), prefix.Prefix, err))
			continue
		}
		created[prefix.ID] = remote.ID
		e.recordExternalID(&prefix.ExternalIDs, remote.ID)
		e.ensureVRFPrefixAssignments(ctx, prefix, remote, inventory, result)
	}

	clog.Info("  Prefixes created: %d", result.PrefixesCreated)
	return created, nil
}

// loadPrefix finds or creates one prefix in its namespace and returns its
// cache entry. Finding an existing prefix counts as skipped; a prefix a dry
// run would create is Planned.
func (e *Exporter) loadPrefix(
	ctx context.Context,
	prefix *devicetypes.CaniPrefix,
	inventory *devicetypes.Inventory,
	locationMap map[uuid.UUID]uuid.UUID,
	vlanMap map[uuid.UUID]uuid.UUID,
	created map[uuid.UUID]uuid.UUID,
	result *LoadResult,
) (*CachedItem, error) {
	namespace := prefix.EffectiveNamespace()
	ns, exists, err := e.ensureNamespace(namespace)
	if err != nil {
		return nil, fmt.Errorf("resolve namespace: %w", err)
	}
	var existing *CachedItem
	if exists {
		if existing, err = e.Cache.LookupPrefix(namespace, prefix.Prefix); err != nil {
			return nil, fmt.Errorf("lookup error: %w", err)
		}
	}
	if existing != nil {
		result.PrefixesSkipped++
		return existing, nil
	}

	var nautobotLocID uuid.UUID
	if prefix.Location != uuid.Nil {
		nautobotLocID = locationMap[prefix.Location]
	}
	nautobotID, err := e.createPrefix(ctx, prefix, ns.ID, nautobotLocID, vlanMap, created, result)
	if err != nil {
		return nil, fmt.Errorf("create error: %w", err)
	}
	item := &CachedItem{ID: nautobotID, Name: prefix.Prefix, Planned: e.Options.DryRun}
	e.Cache.CachePrefix(namespace, prefix.Prefix, item)
	result.PrefixesCreated++
	return item, nil
}

// createPrefix creates a single prefix in Nautobot.
func (e *Exporter) createPrefix(
	ctx context.Context,
	prefix *devicetypes.CaniPrefix,
	namespaceID uuid.UUID,
	locationID uuid.UUID,
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
	}
	if err := setRefID(&req.Status, statusItem.ID); err != nil {
		return uuid.Nil, fmt.Errorf("set prefix status reference: %w", err)
	}

	// Set namespace
	if err := setRefID(&req.Namespace, namespaceID); err != nil {
		return uuid.Nil, fmt.Errorf("set prefix namespace reference: %w", err)
	}

	// Set type
	if prefix.Type != "" {
		prefixType := mapPrefixType(prefix.Type)
		req.Type = &prefixType
	}

	// Set description
	if prefix.Description != "" {
		req.Description = &prefix.Description
	}

	// Scope to the location when it maps to a known Nautobot location.
	if locationID != uuid.Nil {
		if err := setRefID(&req.Location, locationID); err != nil {
			return uuid.Nil, fmt.Errorf("set prefix location reference: %w", err)
		}
	}

	// Resolve parent prefix
	if prefix.Parent != uuid.Nil {
		if parentNID, ok := createdPrefixes[prefix.Parent]; ok {
			if err := setRefID(&req.Parent, parentNID); err != nil {
				return uuid.Nil, fmt.Errorf("set parent prefix reference: %w", err)
			}
		}
	}

	// Resolve VLAN
	if prefix.VLAN != uuid.Nil {
		if vlanNID, ok := vlanMap[prefix.VLAN]; ok {
			if err := setRefID(&req.Vlan, vlanNID); err != nil {
				return uuid.Nil, fmt.Errorf("set prefix VLAN reference: %w", err)
			}
		}
	}

	// Resolve role
	if prefix.Role != "" {
		roleItem, err := e.Cache.GetRole(prefix.Role)
		if err == nil && roleItem != nil {
			if err := setRefID(&req.Role, roleItem.ID); err != nil {
				return uuid.Nil, fmt.Errorf("set prefix role reference: %w", err)
			}
		}
	}

	if e.Options.DryRun {
		clog.DryRun("Would create prefix: %s", prefix.Prefix)
		return uuid.New(), nil
	}

	nautobotID, err := e.postPrefixWithFallback(ctx, prefix, &req)
	if err != nil {
		return uuid.Nil, err
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
