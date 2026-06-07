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
	"net/http"
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// loadLocations exports CaniLocationType records to Nautobot as Location objects.
// It walks the location tree top-down (roots first, then children) so parent
// FKs are always resolvable. Each location's LocationType field drives the
// Nautobot LocationType, replacing the old hardcoded "Site" default.
func (e *Exporter) loadLocations(
	ctx context.Context,
	inventory *devicetypes.Inventory,
	result *LoadResult,
) (map[uuid.UUID]uuid.UUID, error) {
	created := make(map[uuid.UUID]uuid.UUID) // cani ID → Nautobot ID

	if len(inventory.Locations) == 0 {
		return created, nil
	}

	// Build ordered list: roots first, then children.
	ordered := topologicalSortLocations(inventory.Locations)

	for _, loc := range ordered {
		if loc == nil || loc.Name == "" {
			continue
		}

		// Always verify against Nautobot rather than trusting cached
		// ExternalIDs, which may be stale after a Nautobot reset.
		existing, err := e.Cache.LookupLocation(loc.Name)
		if err == nil && existing != nil {
			created[loc.ID] = existing.ID
			setExternalID(&loc.ExternalIDs, "nautobot", existing.ID)
			result.LocationsSkipped = append(result.LocationsSkipped, loc.Name)
			clog.Skipped("Skipped location (already exists): %s", loc.Name)
			continue
		}

		nautobotID, err := e.createLocationFromCani(ctx, loc, created, result)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("location %s: create error: %v", loc.Name, err))
			continue
		}
		created[loc.ID] = nautobotID
		setExternalID(&loc.ExternalIDs, "nautobot", nautobotID)
	}

	return created, nil
}

// createLocationFromCani creates a single Nautobot Location from a CaniLocationType.
func (e *Exporter) createLocationFromCani(
	ctx context.Context,
	loc *devicetypes.CaniLocationType,
	createdMap map[uuid.UUID]uuid.UUID,
	result *LoadResult,
) (uuid.UUID, error) {
	// Resolve LocationType from the CaniLocationType field.
	locTypeName := loc.LocationType
	if locTypeName == "" {
		return uuid.Nil, fmt.Errorf("location %q has no locationType set", loc.Name)
	}
	locType, err := e.Cache.GetOrCreateLocationType(locTypeName, parentDef(locTypeName))
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to resolve location type %q: %w", locTypeName, err)
	}

	// Resolve status.
	statusName := loc.Status
	if statusName == "" {
		statusName = e.Options.DefaultStatus
	}
	if statusName == "" {
		statusName = "Active"
	}
	status, err := e.Cache.GetStatus(statusName)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to resolve status %q: %w", statusName, err)
	}

	// Build references.
	locTypeRef := makeStatusRef(locType.ID)
	statusRef := makeStatusRef(status.ID)

	req := nautobotapi.LocationRequest{
		Name:         loc.Name,
		LocationType: locTypeRef,
		Status:       statusRef,
	}

	// Map parent FK if present.
	if loc.Parent != uuid.Nil {
		parentNautobotID, ok := createdMap[loc.Parent]
		if !ok {
			return uuid.Nil, fmt.Errorf("parent %s not yet created in Nautobot (ordering bug?)", loc.Parent)
		}
		parentRef := makeTenantRef(parentNautobotID)
		req.Parent = parentRef
	}

	// Map optional fields.
	if loc.Facility != "" {
		req.Facility = &loc.Facility
	}
	if loc.Description != "" {
		req.Description = &loc.Description
	}
	if loc.PhysicalAddress != "" {
		req.PhysicalAddress = &loc.PhysicalAddress
	}
	if loc.ShippingAddress != "" {
		req.ShippingAddress = &loc.ShippingAddress
	}
	if loc.ContactName != "" {
		req.ContactName = &loc.ContactName
	}
	if loc.ContactPhone != "" {
		req.ContactPhone = &loc.ContactPhone
	}
	if loc.ContactEmail != "" {
		req.ContactEmail = &loc.ContactEmail
	}
	if loc.TimeZone != "" {
		req.TimeZone = &loc.TimeZone
	}
	if loc.Latitude != "" {
		req.Latitude = &loc.Latitude
	}
	if loc.Longitude != "" {
		req.Longitude = &loc.Longitude
	}
	if loc.Asn != nil {
		req.Asn = loc.Asn
	}
	if loc.Comments != "" {
		req.Comments = &loc.Comments
	}
	if len(loc.CustomFields) > 0 {
		cf := map[string]interface{}{}
		for k, v := range loc.CustomFields {
			cf[k] = v
		}
		req.CustomFields = &cf
	}

	if e.Options.DryRun {
		clog.DryRun("Would create location: %s (type: %s)", loc.Name, locTypeName)
		result.LocationsCreated = append(result.LocationsCreated, loc.Name+" (dry-run)")
		// Cache so downstream phases (racks, devices) can find the location by name.
		e.Cache.CacheLocation(loc.Name, &CachedItem{Name: loc.Name})
		return uuid.Nil, nil
	}

	resp, err := e.Client.DcimLocationsCreateWithResponse(ctx,
		&nautobotapi.DcimLocationsCreateParams{}, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}
	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return uuid.Nil, fmt.Errorf("unexpected status %d: %s",
			resp.StatusCode(), string(resp.Body))
	}

	var nautobotID uuid.UUID
	if resp.JSON201 != nil && resp.JSON201.Id != nil {
		nautobotID = toUUID(resp.JSON201.Id)
	}

	// Cache the location so rack/device resolvers can find it by name.
	e.Cache.CacheLocation(loc.Name, &CachedItem{
		ID:   nautobotID,
		Name: loc.Name,
	})

	clog.Created("Created location: %s (type: %s, ID: %s)", loc.Name, locTypeName, nautobotID)
	result.LocationsCreated = append(result.LocationsCreated, loc.Name)
	return nautobotID, nil
}

// topologicalSortLocations returns locations ordered so that parents come
// before children, enabling parent FK resolution during sequential creation.
// Roots and siblings are sorted by name for deterministic output.
func topologicalSortLocations(locs map[uuid.UUID]*devicetypes.CaniLocationType) []*devicetypes.CaniLocationType {
	// Build child → parent and parent → children maps.
	children := make(map[uuid.UUID][]uuid.UUID)
	var roots []uuid.UUID
	for id, loc := range locs {
		if loc == nil {
			continue
		}
		if loc.Parent == uuid.Nil {
			roots = append(roots, id)
		} else {
			children[loc.Parent] = append(children[loc.Parent], id)
		}
	}

	// Sort roots by location name for deterministic ordering.
	sort.Slice(roots, func(i, j int) bool {
		return locs[roots[i]].Name < locs[roots[j]].Name
	})

	// Sort each child list by name so siblings are deterministic.
	for parentID, kids := range children {
		sort.Slice(kids, func(i, j int) bool {
			return locs[kids[i]].Name < locs[kids[j]].Name
		})
		children[parentID] = kids
	}

	// BFS from roots.
	var ordered []*devicetypes.CaniLocationType
	queue := make([]uuid.UUID, 0, len(roots))
	queue = append(queue, roots...)

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if loc, ok := locs[id]; ok && loc != nil {
			ordered = append(ordered, loc)
		}
		for _, childID := range children[id] {
			queue = append(queue, childID)
		}
	}
	return ordered
}

// makeTenantRef creates a BulkWritableCircuitRequestTenant reference from a UUID.
func makeTenantRef(id uuid.UUID) *nautobotapi.BulkWritableCircuitRequestTenant {
	idUnion := nautobotapi.BulkWritableCableRequestStatusId{}
	idUnion.FromBulkWritableCableRequestStatusId0(id)
	return &nautobotapi.BulkWritableCircuitRequestTenant{
		Id: &idUnion,
	}
}
