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
	"fmt"
	"net/http"
	"strings"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// assignmentPageSize is the number of VRF-prefix assignments requested per page.
const assignmentPageSize = 100

// CacheVRF stores a VRF under its namespace and name. vrfsByName lets an
// unqualified interface VRF name keep working while it is unique across the
// exported namespaces.
func (c *LookupCache) CacheVRF(namespace, name string, item *CachedItem) {
	key := scopedKey(namespace, name)
	c.vrfsMu.Lock()
	defer c.vrfsMu.Unlock()
	if _, exists := c.vrfs[key]; !exists {
		c.vrfsByName[name] = append(c.vrfsByName[name], key)
	}
	c.vrfs[key] = item
}

// LookupVRF returns the VRF cached under a namespace and name, or nil. A
// loader that holds the VRF record uses it instead of a reference string, so
// a name that itself contains a slash still resolves.
func (c *LookupCache) LookupVRF(namespace, name string) *CachedItem {
	c.vrfsMu.RLock()
	defer c.vrfsMu.RUnlock()
	return c.vrfs[scopedKey(namespace, name)]
}

// LookupCachedVRF resolves a VRF reference from the cache. A reference is
// either "namespace/name" or a bare name; a bare name resolves only while
// exactly one cached namespace carries it, otherwise the caller gets an
// error naming the qualifier it must supply. The last slash separates the
// qualifier, so a VRF whose own name contains a slash cannot be referenced
// this way (the model's reference grammar shares that limit). nil, nil means
// not cached.
func (c *LookupCache) LookupCachedVRF(ref string) (*CachedItem, error) {
	c.vrfsMu.RLock()
	defer c.vrfsMu.RUnlock()
	if idx := strings.LastIndexByte(ref, '/'); idx >= 0 {
		return c.vrfs[scopedKey(ref[:idx], ref[idx+1:])], nil
	}
	keys := c.vrfsByName[ref]
	switch len(keys) {
	case 0:
		return nil, nil
	case 1:
		return c.vrfs[keys[0]], nil
	default:
		qualified := make([]string, len(keys))
		for i, key := range keys {
			qualified[i] = key.String()
		}
		return nil, fmt.Errorf("VRF %q exists in %d namespaces (%s); qualify it as namespace/%s",
			ref, len(keys), strings.Join(qualified, ", "), ref)
	}
}

// VRFPrefixAssigned reports whether Nautobot links the prefix to the VRF. A
// VRF's assignments are listed once, when it is first asked about, so an
// export pays one paginated read per VRF rather than one lookup per
// membership.
func (c *LookupCache) VRFPrefixAssigned(vrfID, prefixID uuid.UUID) (bool, error) {
	assigned, err := c.vrfPrefixAssignments(vrfID)
	if err != nil {
		return false, err
	}
	return assigned[prefixID], nil
}

// CacheVRFPrefixAssignment adds a link this run created to the VRF's listed
// assignments. A VRF not listed yet needs nothing: its first listing will
// include the link.
func (c *LookupCache) CacheVRFPrefixAssignment(vrfID, prefixID uuid.UUID) {
	c.vrfsMu.Lock()
	defer c.vrfsMu.Unlock()
	if assigned, listed := c.vrfPrefixes[vrfID]; listed {
		assigned[prefixID] = true
	}
}

// vrfPrefixAssignments returns the IDs of the prefixes assigned to a VRF,
// listing them from Nautobot the first time. The offset advances by the
// results a page returned, not the size requested, so a server that caps
// pages below the request cannot make the listing skip assignments.
func (c *LookupCache) vrfPrefixAssignments(vrfID uuid.UUID) (map[uuid.UUID]bool, error) {
	c.vrfsMu.RLock()
	assigned, listed := c.vrfPrefixes[vrfID]
	c.vrfsMu.RUnlock()
	if listed {
		return assigned, nil
	}
	assigned = map[uuid.UUID]bool{}
	for offset := 0; ; {
		page, err := c.listVRFPrefixAssignments(vrfID, offset)
		if err != nil {
			return nil, err
		}
		for _, assignment := range page.Results {
			assigned[refID(assignment.Prefix.Id)] = true
		}
		if page.Next == nil || *page.Next == "" || len(page.Results) == 0 {
			break
		}
		offset += len(page.Results)
	}
	c.vrfsMu.Lock()
	defer c.vrfsMu.Unlock()
	c.vrfPrefixes[vrfID] = assigned
	return assigned, nil
}

// listVRFPrefixAssignments fetches one page of a VRF's prefix assignments.
func (c *LookupCache) listVRFPrefixAssignments(vrfID uuid.UUID, offset int) (*nautobotapi.PaginatedVRFPrefixAssignmentList, error) {
	vrfFilter := []string{vrfID.String()}
	limit := assignmentPageSize
	resp, err := c.client.IpamVrfPrefixAssignmentsListWithResponse(c.ctx,
		&nautobotapi.IpamVrfPrefixAssignmentsListParams{Vrf: &vrfFilter, Limit: &limit, Offset: &offset})
	if err != nil {
		return nil, fmt.Errorf("list prefix assignments of VRF %s: %w", vrfID, err)
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil {
		return nil, fmt.Errorf("list prefix assignments of VRF %s: unexpected status %d: %s",
			vrfID, resp.StatusCode(), string(resp.Body))
	}
	return resp.JSON200, nil
}
