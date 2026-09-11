/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024, 2026 Hewlett Packard Enterprise Development LP
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
	"fmt"
	"net/http"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// GetOrCreateTag looks up a tag by name, creating it if not found.
func (c *LookupCache) GetOrCreateTag(name string) (*CachedItem, error) {
	c.tagsMu.RLock()
	if item, ok := c.tags[name]; ok {
		c.tagsMu.RUnlock()
		return item, nil
	}
	c.tagsMu.RUnlock()

	c.tagsMu.Lock()
	defer c.tagsMu.Unlock()

	// Double-check after acquiring write lock.
	if item, ok := c.tags[name]; ok {
		return item, nil
	}

	// Try to find existing tag.
	nameFilter := []string{name}
	resp, err := c.client.ExtrasTagsListWithResponse(c.ctx, &nautobotapi.ExtrasTagsListParams{
		Name: &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup tag %s: %w", name, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup tag %s: status %d", name, resp.StatusCode())
	}

	if resp.JSON200 != nil && len(resp.JSON200.Results) > 0 {
		t := resp.JSON200.Results[0]
		item := c.reconcileTagContentTypes(t)
		c.tags[name] = item
		return item, nil
	}

	// Tag not found — create it. Content types must cover every object cani
	// tags (device, rack, interface, inventory item); an empty list makes the
	// tag unassignable and Nautobot rejects the write with "Related object not
	// found using the provided attributes".
	clog.Detail("[nautobot] Creating tag: %s", name)
	createResp, err := c.client.ExtrasTagsCreateWithResponse(c.ctx,
		&nautobotapi.ExtrasTagsCreateParams{},
		nautobotapi.TagRequest{
			Name:         name,
			ContentTypes: taggableContentTypes(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag %s: %w", name, err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("failed to create tag %s: status %d: %s",
			name, createResp.StatusCode(), string(createResp.Body))
	}

	if createResp.JSON201 != nil {
		item := &CachedItem{
			ID:   toUUID(createResp.JSON201.Id),
			Name: createResp.JSON201.Name,
		}
		c.tags[name] = item
		clog.Created("[nautobot] Created tag: %s (ID: %s)", name, item.ID)
		return item, nil
	}

	return nil, fmt.Errorf("failed to create tag %s: no response body", name)
}

// reconcileTagContentTypes ensures an existing tag advertises every content
// type cani assigns tags to. Tags created without the right content_types
// (including those written by older cani versions) cannot be attached to
// devices, racks, interfaces, or FRUs. Missing types are patched in; on any
// patch error the original tag is returned so the export can still proceed.
func (c *LookupCache) reconcileTagContentTypes(t nautobotapi.Tag) *CachedItem {
	item := &CachedItem{ID: toUUID(t.Id), Name: t.Name}
	missing := missingContentTypes(t.ContentTypes, taggableContentTypes())
	if len(missing) == 0 {
		return item
	}
	updated := append(t.ContentTypes, missing...)
	patched, err := c.UpdateTagContentTypes(toUUID(t.Id), t.Name, updated)
	if err != nil {
		clog.Warn("[nautobot] WARNING: failed to update tag '%s' content types: %v", t.Name, err)
		return item
	}
	return patched
}

// UpdateTagContentTypes patches an existing tag to include additional content types.
func (c *LookupCache) UpdateTagContentTypes(id uuid.UUID, name string, contentTypes []string) (*CachedItem, error) {
	clog.Detail("[nautobot] Updating tag '%s' content types to: %v", name, contentTypes)

	patchResp, err := c.client.ExtrasTagsPartialUpdateWithResponse(c.ctx,
		id,
		&nautobotapi.ExtrasTagsPartialUpdateParams{},
		nautobotapi.PatchedTagRequest{
			ContentTypes: &contentTypes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update tag %s: %w", name, err)
	}
	if patchResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to update tag %s: status %d: %s", name, patchResp.StatusCode(), string(patchResp.Body))
	}

	if patchResp.JSON200 != nil {
		item := &CachedItem{
			ID:   toUUID(patchResp.JSON200.Id),
			Name: patchResp.JSON200.Name,
		}
		clog.Created("[nautobot] Updated tag: %s (ID: %s) with content_types: %v", name, item.ID, contentTypes)
		return item, nil
	}

	return nil, fmt.Errorf("failed to update tag %s: no response body", name)
}
