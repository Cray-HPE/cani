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

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// loadFrus exports CaniFruType records to Nautobot as InventoryItem objects.
// FRUs are processed in two passes: first those without a parent FRU (top-level),
// then nested FRUs, so parent FK references are always resolvable.
func (e *Exporter) loadFrus(
	ctx context.Context,
	inventory *devicetypes.Inventory,
	createdDeviceIDs map[string]uuid.UUID,
	result *LoadResult,
) error {
	if len(inventory.Frus) == 0 {
		return nil
	}

	// Map cani FRU ID → Nautobot InventoryItem ID.
	createdFruIDs := make(map[uuid.UUID]uuid.UUID)

	// Order: top-level FRUs first (Parent == Nil), then nested.
	ordered := topologicalSortFrus(inventory.Frus)

	for _, fru := range ordered {
		if fru == nil || fru.Name == "" {
			continue
		}

		nautobotID, err := e.createFruFromCani(ctx, fru, inventory, createdDeviceIDs, createdFruIDs, result)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("fru %s: create error: %v", fru.Name, err))
			continue
		}
		createdFruIDs[fru.ID] = nautobotID
	}
	return nil
}

// createFruFromCani creates a single Nautobot InventoryItem from a CaniFruType.
func (e *Exporter) createFruFromCani(
	ctx context.Context,
	fru *devicetypes.CaniFruType,
	inventory *devicetypes.Inventory,
	createdDeviceIDs map[string]uuid.UUID,
	createdFruIDs map[uuid.UUID]uuid.UUID,
	result *LoadResult,
) (uuid.UUID, error) {
	// Resolve parent device.
	parentDevice := inventory.Devices[fru.Device]
	if parentDevice == nil {
		return uuid.Nil, fmt.Errorf("parent device %s not found in inventory", fru.Device)
	}
	parentNautobotID, ok := createdDeviceIDs[parentDevice.Name]
	if !ok {
		return uuid.Nil, fmt.Errorf("parent device %s not found in Nautobot", parentDevice.Name)
	}

	// Idempotency: check if an inventory item with this name already exists on the device.
	deviceFilter := []string{parentNautobotID.String()}
	nameFilter := []string{fru.Name}
	existResp, err := e.Client.DcimInventoryItemsListWithResponse(ctx,
		&nautobotapi.DcimInventoryItemsListParams{
			Device: &deviceFilter,
			Name:   &nameFilter,
		})
	if err == nil && existResp.StatusCode() == http.StatusOK &&
		existResp.JSON200 != nil && existResp.JSON200.Count > 0 {
		existing := existResp.JSON200.Results[0]
		existingID := toUUID(existing.Id)
		clog.Skipped("Skipped inventory item (already exists): %s on %s",
			fru.Name, parentDevice.Name)
		result.FrusSkipped++
		return existingID, nil
	}

	// Build device reference.
	deviceRef := makeStatusRef(parentNautobotID)

	req := nautobotapi.InventoryItemRequest{
		Name:   fru.Name,
		Device: deviceRef,
	}

	// Map PartNumber → PartId.
	if fru.PartNumber != "" {
		req.PartId = &fru.PartNumber
	}

	// Map Manufacturer → FK lookup.
	if fru.Manufacturer != "" {
		mfr, err := e.Cache.GetOrCreateManufacturer(fru.Manufacturer)
		if err == nil && mfr != nil {
			req.Manufacturer = makeTenantRef(mfr.ID)
		}
	}

	// Map parent FRU → InventoryItem.Parent FK.
	if fru.Parent != uuid.Nil {
		if parentFruNautobotID, ok := createdFruIDs[fru.Parent]; ok {
			req.Parent = makeTenantRef(parentFruNautobotID)
		}
	}

	// Map optional fields.
	if fru.Serial != "" {
		req.Serial = &fru.Serial
	}
	if fru.AssetTag != "" {
		req.AssetTag = &fru.AssetTag
	}
	if fru.Description != "" {
		req.Description = &fru.Description
	}
	if fru.Label != "" {
		req.Label = &fru.Label
	}
	if fru.Discovered {
		req.Discovered = &fru.Discovered
	}

	// Map Tags — convert string tag names to Nautobot status-ref objects.
	if len(fru.Tags) > 0 {
		tags := make([]nautobotapi.BulkWritableCableRequestStatus, 0, len(fru.Tags))
		for _, tagName := range fru.Tags {
			tag, err := e.Cache.GetOrCreateTag(tagName)
			if err == nil && tag != nil {
				tags = append(tags, makeStatusRef(tag.ID))
			}
		}
		if len(tags) > 0 {
			req.Tags = &tags
		}
	}

	// Map CustomFields
	if len(fru.CustomFields) > 0 {
		cf := map[string]interface{}(fru.CustomFields)
		req.CustomFields = &cf
	}

	if e.Options.DryRun {
		clog.DryRun("Would create inventory item: %s (device: %s)",
			fru.Name, parentDevice.Name)
		result.FrusCreated++
		return uuid.Nil, nil
	}

	resp, err := e.Client.DcimInventoryItemsCreateWithResponse(ctx,
		&nautobotapi.DcimInventoryItemsCreateParams{}, req)
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

	clog.Created("Created inventory item: %s (device: %s, ID: %s)",
		fru.Name, parentDevice.Name, nautobotID)
	result.FrusCreated++
	return nautobotID, nil
}

// topologicalSortFrus returns FRUs ordered so that top-level items come
// before nested ones, enabling parent FK resolution during sequential creation.
func topologicalSortFrus(frus map[uuid.UUID]*devicetypes.CaniFruType) []*devicetypes.CaniFruType {
	children := make(map[uuid.UUID][]uuid.UUID)
	var roots []uuid.UUID
	for id, fru := range frus {
		if fru == nil {
			continue
		}
		if fru.Parent == uuid.Nil {
			roots = append(roots, id)
		} else {
			children[fru.Parent] = append(children[fru.Parent], id)
		}
	}

	var ordered []*devicetypes.CaniFruType
	queue := make([]uuid.UUID, 0, len(roots))
	queue = append(queue, roots...)
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if fru, ok := frus[id]; ok && fru != nil {
			ordered = append(ordered, fru)
		}
		for _, childID := range children[id] {
			queue = append(queue, childID)
		}
	}

	// Detect unreachable FRUs (cycles or orphaned parent references).
	if len(ordered) < len(frus) {
		var missing []string
		visited := make(map[uuid.UUID]bool, len(ordered))
		for _, fru := range ordered {
			visited[fru.ID] = true
		}
		for id, fru := range frus {
			if fru != nil && !visited[id] {
				missing = append(missing, fru.Name+" ("+id.String()+")")
			}
		}
		clog.Warn("WARNING: %d FRU(s) unreachable (cycle or orphan parent): %v",
			len(missing), missing)
	}

	return ordered
}
