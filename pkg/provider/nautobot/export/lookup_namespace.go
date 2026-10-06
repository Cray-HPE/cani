/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
)

// ensureNamespace resolves the namespace an object exports into. exists is
// false only for a namespace a dry run plans: nothing can exist in it yet and
// Nautobot rejects lookups filtered by an unknown namespace, so callers treat
// everything in it as absent instead of querying.
func (e *Exporter) ensureNamespace(name string) (ns *CachedItem, exists bool, err error) {
	ns, err = e.Cache.ResolveNamespace(name)
	if err != nil {
		return nil, false, err
	}
	return ns, !ns.Planned, nil
}

// ResolveNamespace returns the Nautobot namespace with the given name,
// creating a missing one. A dry run plans the create instead and caches a
// Planned placeholder with a nil ID, so the lookup is not repeated either.
func (c *LookupCache) ResolveNamespace(name string) (*CachedItem, error) {
	c.namespacesMu.RLock()
	item, found := c.namespaces[name]
	c.namespacesMu.RUnlock()
	if found {
		return item, nil
	}

	c.namespacesMu.Lock()
	defer c.namespacesMu.Unlock()
	return c.resolveNamespaceLocked(name)
}

// resolveNamespaceLocked finishes ResolveNamespace under the write lock so
// concurrent callers create a missing namespace only once.
func (c *LookupCache) resolveNamespaceLocked(name string) (*CachedItem, error) {
	if item, found := c.namespaces[name]; found {
		return item, nil
	}
	item, err := c.findNamespace(name)
	if err != nil {
		return nil, err
	}
	switch {
	case item != nil:
	case c.dryRun:
		item = plannedCreate("namespace", name)
	default:
		if item, err = c.createNamespace(name); err != nil {
			return nil, err
		}
	}
	c.namespaces[name] = item
	return item, nil
}

// findNamespace looks a namespace up by name; nil means it does not exist.
func (c *LookupCache) findNamespace(name string) (*CachedItem, error) {
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
	if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
		return nil, nil
	}
	ns := resp.JSON200.Results[0]
	return &CachedItem{ID: toUUID(ns.Id), Name: ns.Name, Display: *ns.Display}, nil
}

// createNamespace creates a namespace in Nautobot.
func (c *LookupCache) createNamespace(name string) (*CachedItem, error) {
	resp, err := c.client.IpamNamespacesCreateWithResponse(c.ctx, &nautobotapi.IpamNamespacesCreateParams{},
		nautobotapi.NamespaceRequest{Name: name})
	if err != nil {
		return nil, fmt.Errorf("failed to create namespace %s: %w", name, err)
	}
	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create namespace %s: status %d: %s", name, resp.StatusCode(), string(resp.Body))
	}
	if resp.JSON201 == nil {
		return nil, fmt.Errorf("namespace %s: unexpected empty response after create", name)
	}
	return &CachedItem{ID: toUUID(resp.JSON201.Id), Name: resp.JSON201.Name, Display: *resp.JSON201.Display}, nil
}
