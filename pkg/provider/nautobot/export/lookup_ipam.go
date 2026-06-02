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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// ipamCache extends LookupCache with IPAM-specific caches.
// Fields and methods are attached via composition on LookupCache.
var (
	namespaces   = make(map[string]*CachedItem)
	namespacesMu sync.RWMutex

	vlans   = make(map[string]*CachedItem) // keyed by "vid:locationID"
	vlansMu sync.RWMutex

	prefixes   = make(map[string]*CachedItem) // keyed by CIDR string
	prefixesMu sync.RWMutex

	ipAddresses   = make(map[string]*CachedItem) // keyed by address string
	ipAddressesMu sync.RWMutex
)

// GetOrCreateNamespace looks up a Nautobot Namespace by name. If it doesn't
// exist and auto-creation is not disabled, it creates a new one.
func (c *LookupCache) GetOrCreateNamespace(name string) (*CachedItem, error) {
	namespacesMu.RLock()
	if item, ok := namespaces[name]; ok {
		namespacesMu.RUnlock()
		return item, nil
	}
	namespacesMu.RUnlock()

	namespacesMu.Lock()
	defer namespacesMu.Unlock()

	// Double-check
	if item, ok := namespaces[name]; ok {
		return item, nil
	}

	nameFilter := []string{name}
	resp, err := c.client.IpamNamespacesListWithResponse(c.ctx, &nautobotapi.IpamNamespacesListParams{
		Name: &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup namespace %s: %w", name, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup namespace %s: status %d", name, resp.StatusCode())
	}

	if resp.JSON200 != nil && resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		ns := (resp.JSON200.Results)[0]
		item := &CachedItem{
			ID:      toUUID(ns.Id),
			Name:    ns.Name,
			Display: *ns.Display,
		}
		namespaces[name] = item
		return item, nil
	}

	// Namespace not found — create it
	createResp, err := c.client.IpamNamespacesCreateWithResponse(c.ctx, &nautobotapi.IpamNamespacesCreateParams{}, nautobotapi.NamespaceRequest{
		Name: name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create namespace %s: %w", name, err)
	}
	if createResp.StatusCode() != http.StatusCreated && createResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create namespace %s: status %d: %s", name, createResp.StatusCode(), string(createResp.Body))
	}
	if createResp.JSON201 != nil {
		item := &CachedItem{
			ID:      toUUID(createResp.JSON201.Id),
			Name:    createResp.JSON201.Name,
			Display: *createResp.JSON201.Display,
		}
		namespaces[name] = item
		return item, nil
	}

	return nil, fmt.Errorf("namespace %s: unexpected empty response after create", name)
}

// LookupVLAN looks up an existing VLAN by VID and location. Returns nil, nil if not found.
func (c *LookupCache) LookupVLAN(vid int, locationName string) (*CachedItem, error) {
	key := strconv.Itoa(vid) + ":" + locationName
	vlansMu.RLock()
	if item, ok := vlans[key]; ok {
		vlansMu.RUnlock()
		return item, nil
	}
	vlansMu.RUnlock()

	vidFilter := []int{vid}
	resp, err := c.client.IpamVlansListWithResponse(c.ctx, &nautobotapi.IpamVlansListParams{
		Vid: &vidFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup VLAN %d: %w", vid, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup VLAN %d: status %d", vid, resp.StatusCode())
	}

	if resp.JSON200 != nil && resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		vlan := (resp.JSON200.Results)[0]
		item := &CachedItem{
			ID:      toUUID(vlan.Id),
			Name:    vlan.Name,
			Display: *vlan.Display,
		}
		vlansMu.Lock()
		vlans[key] = item
		vlansMu.Unlock()
		return item, nil
	}

	return nil, nil
}

// CacheVLAN stores a VLAN in the lookup cache after creation.
func (c *LookupCache) CacheVLAN(vid int, locationName string, item *CachedItem) {
	key := strconv.Itoa(vid) + ":" + locationName
	vlansMu.Lock()
	vlans[key] = item
	vlansMu.Unlock()
}

// LookupPrefix looks up an existing prefix by CIDR string. Returns nil, nil if not found.
// Uses raw API call because the generated response parser has []byte fields
// (network, broadcast) that Go's JSON decoder tries to base64-decode, but
// Nautobot returns plain IP strings.
func (c *LookupCache) LookupPrefix(cidr string) (*CachedItem, error) {
	prefixesMu.RLock()
	if item, ok := prefixes[cidr]; ok {
		prefixesMu.RUnlock()
		return item, nil
	}
	prefixesMu.RUnlock()

	prefixFilter := []string{cidr}
	httpResp, err := c.client.IpamPrefixesList(c.ctx, &nautobotapi.IpamPrefixesListParams{
		Prefix: &prefixFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup prefix %s: %w", cidr, err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup prefix %s: status %d", cidr, httpResp.StatusCode)
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read prefix lookup response for %s: %w", cidr, err)
	}

	var result struct {
		Results []struct {
			ID      string `json:"id"`
			Prefix  string `json:"prefix"`
			Display string `json:"display"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse prefix lookup response for %s: %w", cidr, err)
	}

	if len(result.Results) > 0 {
		p := result.Results[0]
		parsedID, err := uuid.Parse(p.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse prefix UUID for %s: %w", cidr, err)
		}
		item := &CachedItem{
			ID:      parsedID,
			Name:    p.Prefix,
			Display: p.Display,
		}
		prefixesMu.Lock()
		prefixes[cidr] = item
		prefixesMu.Unlock()
		return item, nil
	}

	return nil, nil
}

// CachePrefix stores a prefix in the lookup cache after creation.
func (c *LookupCache) CachePrefix(cidr string, item *CachedItem) {
	prefixesMu.Lock()
	prefixes[cidr] = item
	prefixesMu.Unlock()
}

// LookupIPAddress looks up an existing IP address by address string. Returns nil, nil if not found.
// Uses raw API call because the generated response parser has a []byte field
// (host) that Go's JSON decoder tries to base64-decode, but Nautobot returns
// plain IP strings.
func (c *LookupCache) LookupIPAddress(address string) (*CachedItem, error) {
	ipAddressesMu.RLock()
	if item, ok := ipAddresses[address]; ok {
		ipAddressesMu.RUnlock()
		return item, nil
	}
	ipAddressesMu.RUnlock()

	addrFilter := []string{address}
	httpResp, err := c.client.IpamIpAddressesList(c.ctx, &nautobotapi.IpamIpAddressesListParams{
		Address: &addrFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup IP address %s: %w", address, err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup IP address %s: status %d", address, httpResp.StatusCode)
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read IP address lookup response for %s: %w", address, err)
	}

	var result struct {
		Results []struct {
			ID      string `json:"id"`
			Address string `json:"address"`
			Display string `json:"display"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse IP address lookup response for %s: %w", address, err)
	}

	if len(result.Results) > 0 {
		ip := result.Results[0]
		parsedID, err := uuid.Parse(ip.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse IP address UUID for %s: %w", address, err)
		}
		item := &CachedItem{
			ID:      parsedID,
			Name:    ip.Address,
			Display: ip.Display,
		}
		ipAddressesMu.Lock()
		ipAddresses[address] = item
		ipAddressesMu.Unlock()
		return item, nil
	}

	return nil, nil
}

// CacheIPAddress stores an IP address in the lookup cache after creation.
func (c *LookupCache) CacheIPAddress(address string, item *CachedItem) {
	ipAddressesMu.Lock()
	ipAddresses[address] = item
	ipAddressesMu.Unlock()
}

// makeIDRef creates a BulkWritableCableRequestStatus FK reference from a UUID.
// This is the generic pattern for Nautobot nested FK objects.
func makeIDRef(id uuid.UUID) nautobotapi.BulkWritableCableRequestStatus {
	idUnion := nautobotapi.BulkWritableCableRequestStatusId{}
	idUnion.FromBulkWritableCableRequestStatusId0(id)
	return nautobotapi.BulkWritableCableRequestStatus{
		Id: &idUnion,
	}
}

// makeLocationRef creates a BulkWritablePrefixRequestLocation FK reference from a UUID.
func makeLocationRef(id uuid.UUID) nautobotapi.BulkWritablePrefixRequestLocation {
	idUnion := nautobotapi.BulkWritableCableRequestStatusId{}
	idUnion.FromBulkWritableCableRequestStatusId0(id)
	return nautobotapi.BulkWritablePrefixRequestLocation{
		Id: &idUnion,
	}
}

// makePrefixParentRef creates a BulkWritablePrefixRequestParent FK reference from a UUID.
func makePrefixParentRef(id uuid.UUID) nautobotapi.BulkWritablePrefixRequestParent {
	idUnion := nautobotapi.BulkWritableCableRequestStatusId{}
	idUnion.FromBulkWritableCableRequestStatusId0(id)
	return nautobotapi.BulkWritablePrefixRequestParent{
		Id: &idUnion,
	}
}

// makeIPParentRef creates a BulkWritableIPAddressRequestParent FK reference from a UUID.
func makeIPParentRef(id uuid.UUID) nautobotapi.BulkWritableIPAddressRequestParent {
	idUnion := nautobotapi.BulkWritableCableRequestStatusId{}
	idUnion.FromBulkWritableCableRequestStatusId0(id)
	return nautobotapi.BulkWritableIPAddressRequestParent{
		Id: &idUnion,
	}
}

// makeIPNamespaceRef creates a BulkWritableIPAddressRequestNamespace FK reference from a UUID.
func makeIPNamespaceRef(id uuid.UUID) nautobotapi.BulkWritableIPAddressRequestNamespace {
	idUnion := nautobotapi.BulkWritableCableRequestStatusId{}
	idUnion.FromBulkWritableCableRequestStatusId0(id)
	return nautobotapi.BulkWritableIPAddressRequestNamespace{
		Id: &idUnion,
	}
}
