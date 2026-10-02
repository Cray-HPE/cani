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
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// ipamKey identifies a namespace-scoped IPAM object in the caches. A struct
// key keeps ("Global/blue", "green") apart from ("Global", "blue/green"), which
// one joined string could not.
type ipamKey struct {
	namespace string
	value     string
}

// String renders the key as "namespace/value" for messages.
func (k ipamKey) String() string {
	return k.namespace + "/" + k.value
}

// scopedKey builds the cache key for a namespace-scoped IPAM object.
func scopedKey(namespace, value string) ipamKey {
	return ipamKey{namespace: devicetypes.NamespaceOrDefault(namespace), value: value}
}

// hostOf strips the mask from an address so lookups match the host regardless
// of mask length, which is how Nautobot scopes address uniqueness.
func hostOf(address string) string {
	if idx := strings.IndexByte(address, '/'); idx >= 0 {
		return address[:idx]
	}
	return address
}

// GetOrCreateNamespace looks up a Nautobot Namespace by name. If it doesn't
// exist and auto-creation is not disabled, it creates a new one.
func (c *LookupCache) GetOrCreateNamespace(name string) (*CachedItem, error) {
	c.namespacesMu.RLock()
	if item, ok := c.namespaces[name]; ok {
		c.namespacesMu.RUnlock()
		return item, nil
	}
	c.namespacesMu.RUnlock()

	c.namespacesMu.Lock()
	defer c.namespacesMu.Unlock()

	// Double-check
	if item, ok := c.namespaces[name]; ok {
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
		c.namespaces[name] = item
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
		c.namespaces[name] = item
		return item, nil
	}

	return nil, fmt.Errorf("namespace %s: unexpected empty response after create", name)
}

// LookupVLAN looks up an existing VLAN by VID and location. Returns nil, nil if not found.
func (c *LookupCache) LookupVLAN(vid int, locationName string) (*CachedItem, error) {
	key := strconv.Itoa(vid) + ":" + locationName
	c.vlansMu.RLock()
	if item, ok := c.vlans[key]; ok {
		c.vlansMu.RUnlock()
		return item, nil
	}
	c.vlansMu.RUnlock()

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
		c.vlansMu.Lock()
		c.vlans[key] = item
		c.vlansMu.Unlock()
		return item, nil
	}

	return nil, nil
}

// CacheVLAN stores a VLAN in the lookup cache after creation.
func (c *LookupCache) CacheVLAN(vid int, locationName string, item *CachedItem) {
	key := strconv.Itoa(vid) + ":" + locationName
	c.vlansMu.Lock()
	c.vlans[key] = item
	c.vlansMu.Unlock()
}

// LookupPrefix looks up an existing prefix by namespace and CIDR. Returns
// nil, nil if not found and an error if the namespace holds several matches.
// Uses raw API call because the generated response parser has []byte fields
// (network, broadcast) that Go's JSON decoder tries to base64-decode, but
// Nautobot returns plain IP strings.
func (c *LookupCache) LookupPrefix(namespace, cidr string) (*CachedItem, error) {
	key := scopedKey(namespace, cidr)
	c.prefixesMu.RLock()
	if item, ok := c.prefixes[key]; ok {
		c.prefixesMu.RUnlock()
		return item, nil
	}
	c.prefixesMu.RUnlock()

	prefixFilter := []string{cidr}
	namespaceFilter := []string{devicetypes.NamespaceOrDefault(namespace)}
	httpResp, err := c.client.IpamPrefixesList(c.ctx, &nautobotapi.IpamPrefixesListParams{
		Prefix:    &prefixFilter,
		Namespace: &namespaceFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup prefix %s: %w", key, err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup prefix %s: status %d", key, httpResp.StatusCode)
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read prefix lookup response for %s: %w", key, err)
	}

	var result struct {
		Results []struct {
			ID      string `json:"id"`
			Prefix  string `json:"prefix"`
			Display string `json:"display"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse prefix lookup response for %s: %w", key, err)
	}

	switch len(result.Results) {
	case 0:
		return nil, nil
	case 1:
		p := result.Results[0]
		parsedID, err := uuid.Parse(p.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse prefix UUID for %s: %w", key, err)
		}
		item := &CachedItem{ID: parsedID, Name: p.Prefix, Display: p.Display}
		c.CachePrefix(namespace, cidr, item)
		return item, nil
	default:
		return nil, fmt.Errorf("prefix %s is ambiguous: %d matches in Nautobot", key, len(result.Results))
	}
}

// CachePrefix stores a prefix in the lookup cache after creation.
func (c *LookupCache) CachePrefix(namespace, cidr string, item *CachedItem) {
	c.prefixesMu.Lock()
	c.prefixes[scopedKey(namespace, cidr)] = item
	c.prefixesMu.Unlock()
}

// LookupIPAddress looks up an existing IP address by namespace and host.
// The mask is ignored because Nautobot keys address uniqueness on the host
// within a namespace; Nautobot's unmasked address filter matches any mask.
// Returns nil, nil if not found and an error if several hosts match.
// Uses raw API call because the generated response parser has a []byte field
// (host) that Go's JSON decoder tries to base64-decode, but Nautobot returns
// plain IP strings.
func (c *LookupCache) LookupIPAddress(namespace, address string) (*CachedItem, error) {
	host := hostOf(address)
	key := scopedKey(namespace, host)
	c.ipAddressesMu.RLock()
	if item, ok := c.ipAddresses[key]; ok {
		c.ipAddressesMu.RUnlock()
		return item, nil
	}
	c.ipAddressesMu.RUnlock()

	addrFilter := []string{host}
	namespaceFilter := []string{devicetypes.NamespaceOrDefault(namespace)}
	httpResp, err := c.client.IpamIpAddressesList(c.ctx, &nautobotapi.IpamIpAddressesListParams{
		Address:   &addrFilter,
		Namespace: &namespaceFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup IP address %s: %w", key, err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to lookup IP address %s: status %d", key, httpResp.StatusCode)
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read IP address lookup response for %s: %w", key, err)
	}

	var result struct {
		Results []struct {
			ID      string `json:"id"`
			Address string `json:"address"`
			Display string `json:"display"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse IP address lookup response for %s: %w", key, err)
	}

	switch len(result.Results) {
	case 0:
		return nil, nil
	case 1:
		ip := result.Results[0]
		parsedID, err := uuid.Parse(ip.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse IP address UUID for %s: %w", key, err)
		}
		item := &CachedItem{ID: parsedID, Name: ip.Address, Display: ip.Display}
		c.CacheIPAddress(namespace, address, item)
		return item, nil
	default:
		return nil, fmt.Errorf("IP address %s is ambiguous: %d matches in Nautobot", key, len(result.Results))
	}
}

// CacheIPAddress stores an IP address in the lookup cache after creation.
func (c *LookupCache) CacheIPAddress(namespace, address string, item *CachedItem) {
	c.ipAddressesMu.Lock()
	c.ipAddresses[scopedKey(namespace, hostOf(address))] = item
	c.ipAddressesMu.Unlock()
}
