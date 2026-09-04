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
	"sync"

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

const (
	relKeyAssignedVLANs = "assigned_vlans"
	relKeyBMCDevice     = "bmc_device"
)

var (
	relationshipCache   = make(map[string]uuid.UUID)
	relationshipCacheMu sync.RWMutex
)

// relationshipDef describes a Nautobot relationship definition to find-or-create.
type relationshipDef struct {
	key     string
	label   string
	srcType string
	dstType string
	relType nautobotapi.RelationshipTypeChoices
}

var (
	assignedVLANsRel = relationshipDef{
		key: relKeyAssignedVLANs, label: "Assigned VLANs",
		srcType: contentTypeDevice, dstType: contentTypeVLAN,
		relType: nautobotapi.RelationshipTypeChoicesManyToMany,
	}
	bmcDeviceRel = relationshipDef{
		key: relKeyBMCDevice, label: "BMC Device",
		srcType: contentTypeDevice, dstType: contentTypeDevice,
		relType: nautobotapi.RelationshipTypeChoicesOneToOne,
	}
)

// loadRelationships is the final phase. It creates the assigned_vlans and
// bmc_device relationship associations described by device inventory, creating
// the relationship definitions on demand.
func (e *Exporter) loadRelationships(
	ctx context.Context,
	inventory *devicetypes.Inventory,
	createdVLANIDs map[uuid.UUID]uuid.UUID,
	result *LoadResult,
) error {
	if !relationshipsNeeded(inventory) {
		return nil
	}
	clog.Header("Phase 10: Relationships")

	for _, device := range inventory.Devices {
		if device == nil {
			continue
		}
		srcID := device.ExternalIDs[externalIDKeyNautobot]
		if srcID == uuid.Nil {
			continue
		}
		if len(device.AssignedVLANs) > 0 {
			e.assignVLANs(ctx, device, srcID, createdVLANIDs, result)
		}
		if device.BMCParent != uuid.Nil {
			e.assignBMC(ctx, device, srcID, inventory, result)
		}
	}
	return nil
}

// relationshipsNeeded reports whether any device carries relationship data.
func relationshipsNeeded(inventory *devicetypes.Inventory) bool {
	for _, d := range inventory.Devices {
		if d != nil && (len(d.AssignedVLANs) > 0 || d.BMCParent != uuid.Nil) {
			return true
		}
	}
	return false
}

// assignVLANs creates the assigned_vlans associations for one device.
func (e *Exporter) assignVLANs(
	ctx context.Context,
	device *devicetypes.CaniDeviceType,
	srcID uuid.UUID,
	createdVLANIDs map[uuid.UUID]uuid.UUID,
	result *LoadResult,
) {
	relID, err := e.getOrCreateRelationship(ctx, assignedVLANsRel)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("relationship assigned_vlans: %v", err))
		return
	}
	for _, vlanCaniID := range device.AssignedVLANs {
		vlanID, ok := createdVLANIDs[vlanCaniID]
		if !ok {
			continue
		}
		e.createAssociation(ctx, relKeyAssignedVLANs, relID, assocEndpoints{
			srcType: contentTypeDevice, srcID: srcID,
			dstType: contentTypeVLAN, dstID: vlanID,
		}, result)
	}
}

// assignBMC creates the bmc_device association linking a BMC to its parent.
func (e *Exporter) assignBMC(
	ctx context.Context,
	device *devicetypes.CaniDeviceType,
	srcID uuid.UUID,
	inventory *devicetypes.Inventory,
	result *LoadResult,
) {
	parent := inventory.Devices[device.BMCParent]
	if parent == nil {
		return
	}
	parentID := parent.ExternalIDs[externalIDKeyNautobot]
	if parentID == uuid.Nil {
		return
	}
	relID, err := e.getOrCreateRelationship(ctx, bmcDeviceRel)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("relationship bmc_device: %v", err))
		return
	}
	e.createAssociation(ctx, relKeyBMCDevice, relID, assocEndpoints{
		srcType: contentTypeDevice, srcID: srcID,
		dstType: contentTypeDevice, dstID: parentID,
	}, result)
}

// assocEndpoints identifies the source and destination objects of a
// relationship association.
type assocEndpoints struct {
	srcType string
	srcID   uuid.UUID
	dstType string
	dstID   uuid.UUID
}

// createAssociation creates one relationship association, skipping it when an
// identical association already exists.
func (e *Exporter) createAssociation(
	ctx context.Context,
	relKey string,
	relID uuid.UUID,
	ep assocEndpoints,
	result *LoadResult,
) {
	if e.associationExists(ctx, relKey, ep.srcID, ep.dstID) {
		result.RelationshipsSkipped++
		return
	}
	if e.Options.DryRun {
		clog.DryRun("Would associate %s: %s -> %s", relKey, ep.srcID, ep.dstID)
		return
	}
	req := nautobotapi.RelationshipAssociationRequest{
		SourceType:      ep.srcType,
		SourceId:        openapi_types.UUID(ep.srcID),
		DestinationType: ep.dstType,
		DestinationId:   openapi_types.UUID(ep.dstID),
	}
	setRefID(&req.Relationship, relID)
	resp, err := e.Client.ExtrasRelationshipAssociationsCreateWithResponse(
		ctx, &nautobotapi.ExtrasRelationshipAssociationsCreateParams{}, req)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("association %s: API error: %v", relKey, err))
		return
	}
	if resp.StatusCode() != http.StatusCreated {
		result.Errors = append(result.Errors,
			fmt.Sprintf("association %s: unexpected status %d: %s", relKey, resp.StatusCode(), string(resp.Body)))
		return
	}
	result.RelationshipsCreated++
	clog.Created("  + relationship %s: %s -> %s", relKey, ep.srcID, ep.dstID)
}

// associationExists reports whether an identical association already exists.
func (e *Exporter) associationExists(ctx context.Context, relKey string, srcID, dstID uuid.UUID) bool {
	resp, err := e.Client.ExtrasRelationshipAssociationsListWithResponse(ctx,
		&nautobotapi.ExtrasRelationshipAssociationsListParams{
			Relationship:  &[]string{relKey},
			SourceId:      &[]openapi_types.UUID{openapi_types.UUID(srcID)},
			DestinationId: &[]openapi_types.UUID{openapi_types.UUID(dstID)},
		})
	if err != nil || resp.JSON200 == nil {
		return false
	}
	return len(resp.JSON200.Results) > 0
}

// getOrCreateRelationship find-or-creates a relationship definition by key,
// returning its Nautobot ID.
func (e *Exporter) getOrCreateRelationship(ctx context.Context, def relationshipDef) (uuid.UUID, error) {
	relationshipCacheMu.RLock()
	if id, ok := relationshipCache[def.key]; ok {
		relationshipCacheMu.RUnlock()
		return id, nil
	}
	relationshipCacheMu.RUnlock()

	existing, err := e.findRelationship(ctx, def.key)
	if err != nil {
		return uuid.Nil, err
	}
	if existing != uuid.Nil {
		e.cacheRelationship(def.key, existing)
		return existing, nil
	}

	if e.Options.DryRun {
		id := uuid.New()
		e.cacheRelationship(def.key, id)
		return id, nil
	}

	key := def.key
	relType := def.relType
	req := nautobotapi.RelationshipRequest{
		Label:           def.label,
		Key:             &key,
		SourceType:      def.srcType,
		DestinationType: def.dstType,
		Type:            &relType,
	}
	resp, err := e.Client.ExtrasRelationshipsCreateWithResponse(ctx, &nautobotapi.ExtrasRelationshipsCreateParams{}, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}
	if resp.StatusCode() != http.StatusCreated || resp.JSON201 == nil {
		return uuid.Nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	id := toUUID(resp.JSON201.Id)
	e.cacheRelationship(def.key, id)
	clog.Created("  + relationship definition: %s", def.key)
	return id, nil
}

// findRelationship returns the ID of an existing relationship with the given
// key, or uuid.Nil.
func (e *Exporter) findRelationship(ctx context.Context, key string) (uuid.UUID, error) {
	resp, err := e.Client.ExtrasRelationshipsListWithResponse(ctx, &nautobotapi.ExtrasRelationshipsListParams{})
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil {
		return uuid.Nil, fmt.Errorf("unexpected status %d", resp.StatusCode())
	}
	for _, r := range resp.JSON200.Results {
		if r.Key != nil && *r.Key == key {
			return toUUID(r.Id), nil
		}
	}
	return uuid.Nil, nil
}

// cacheRelationship stores a relationship ID by key.
func (e *Exporter) cacheRelationship(key string, id uuid.UUID) {
	relationshipCacheMu.Lock()
	relationshipCache[key] = id
	relationshipCacheMu.Unlock()
}
