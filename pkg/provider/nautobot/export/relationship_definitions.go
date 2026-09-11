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
	"sync"

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

// getOrCreateRelationship find-or-creates a relationship definition by key.
func (e *Exporter) getOrCreateRelationship(ctx context.Context, def relationshipDef) (uuid.UUID, error) {
	if id, ok := cachedRelationship(def.key); ok {
		return id, nil
	}
	existing, err := e.findRelationship(ctx, def.key)
	if err != nil {
		return uuid.Nil, err
	}
	if existing != uuid.Nil {
		cacheRelationship(def.key, existing)
		return existing, nil
	}
	if e.Options.DryRun {
		id := uuid.New()
		cacheRelationship(def.key, id)
		return id, nil
	}
	return e.createRelationship(ctx, def)
}

func (e *Exporter) createRelationship(ctx context.Context, def relationshipDef) (uuid.UUID, error) {
	key := def.key
	relType := def.relType
	req := nautobotapi.RelationshipRequest{
		Label: def.label, Key: &key, SourceType: def.srcType,
		DestinationType: def.dstType, Type: &relType,
	}
	resp, err := e.Client.ExtrasRelationshipsCreateWithResponse(
		ctx, &nautobotapi.ExtrasRelationshipsCreateParams{}, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}
	if resp.StatusCode() != http.StatusCreated || resp.JSON201 == nil {
		return uuid.Nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	id := toUUID(resp.JSON201.Id)
	cacheRelationship(def.key, id)
	clog.Created("  + relationship definition: %s", def.key)
	return id, nil
}

func (e *Exporter) findRelationship(ctx context.Context, key string) (uuid.UUID, error) {
	resp, err := e.Client.ExtrasRelationshipsListWithResponse(ctx, &nautobotapi.ExtrasRelationshipsListParams{})
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil {
		return uuid.Nil, fmt.Errorf("unexpected status %d", resp.StatusCode())
	}
	for _, relationship := range resp.JSON200.Results {
		if relationship.Key != nil && *relationship.Key == key {
			return toUUID(relationship.Id), nil
		}
	}
	return uuid.Nil, nil
}

func cachedRelationship(key string) (uuid.UUID, bool) {
	relationshipCacheMu.RLock()
	defer relationshipCacheMu.RUnlock()
	id, ok := relationshipCache[key]
	return id, ok
}

func cacheRelationship(key string, id uuid.UUID) {
	relationshipCacheMu.Lock()
	relationshipCache[key] = id
	relationshipCacheMu.Unlock()
}
