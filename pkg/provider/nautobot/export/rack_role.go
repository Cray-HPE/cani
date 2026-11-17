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
	"slices"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// GetRoleFor returns the role named name once objects of contentType may use
// it. GetRole finds or creates a role with the content types devices,
// interfaces and IPAM objects need; a kind outside that set, such as a rack,
// has its content type added here when create_roles allows changing roles.
// Without that option a role that does not list the content type is an
// error, so the caller can leave the role off instead of sending a reference
// Nautobot refuses together with the whole object. A role a dry run only
// planned is returned as planned.
func (c *LookupCache) GetRoleFor(name, contentType string) (*CachedItem, error) {
	role, err := c.GetRole(name)
	if err != nil {
		return nil, err
	}
	if role.Planned || c.roleKindKnown(name, contentType) {
		return role, nil
	}
	types, err := c.roleContentTypes(role)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(types, contentType) {
		if !c.createRoles {
			return nil, fmt.Errorf("role %q does not list content type %s in Nautobot", name, contentType)
		}
		if _, err := c.UpdateRoleContentTypes(role.ID, name, append(types, contentType)); err != nil {
			return nil, err
		}
	}
	c.rememberRoleKind(name, contentType)
	return role, nil
}

// roleContentTypes reads the content types Nautobot lists for a role.
func (c *LookupCache) roleContentTypes(role *CachedItem) ([]string, error) {
	resp, err := c.client.ExtrasRolesRetrieveWithResponse(c.ctx, role.ID, &nautobotapi.ExtrasRolesRetrieveParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to read role %s: %w", role.Name, err)
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to read role %s: status %d", role.Name, resp.StatusCode())
	}
	return resp.JSON200.ContentTypes, nil
}

// roleKindKnown reports whether this run already confirmed that the role
// lists the content type, so a rack row costs one read per role, not per rack.
func (c *LookupCache) roleKindKnown(name, contentType string) bool {
	c.roleKindsMu.Lock()
	defer c.roleKindsMu.Unlock()
	return c.roleKinds[roleKindKey(name, contentType)]
}

func (c *LookupCache) rememberRoleKind(name, contentType string) {
	c.roleKindsMu.Lock()
	defer c.roleKindsMu.Unlock()
	if c.roleKinds == nil {
		c.roleKinds = make(map[string]bool)
	}
	c.roleKinds[roleKindKey(name, contentType)] = true
}

func roleKindKey(name, contentType string) string {
	return contentType + "\x00" + name
}

// setRackRole sends the rack's role when Nautobot holds one racks may use.
// A role that cannot be resolved is named in a warning, left off the request
// and counted with the objects not exported as authored, so the rack and the
// devices in it are still created.
func (e *Exporter) setRackRole(req *nautobotapi.WritableRackRequest, rack *devicetypes.CaniRackType, result *LoadResult) {
	if rack.Role == "" {
		return
	}
	role, err := e.Cache.GetRoleFor(rack.Role, contentTypeRack)
	if err == nil && !role.Planned {
		err = setRefID(&req.Role, role.ID)
	}
	if err != nil {
		result.RackRolesSkipped++
		clog.Warn("rack %s: role %q not sent: %v", rack.Name, rack.Role, err)
	}
}
