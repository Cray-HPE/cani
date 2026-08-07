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

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// mergeVLAN fetches the remote VLAN, compares against local intent,
// and PATCHes any drifted fields. Returns true if the VLAN was updated.
func (e *Exporter) mergeVLAN(ctx context.Context, vlan *devicetypes.CaniVLAN, nautobotID uuid.UUID) (bool, error) {
	remote, err := e.fetchVLANByID(ctx, nautobotID)
	if err != nil {
		return false, err
	}

	req, drifted := buildVLANPatch(vlan, remote)
	if !drifted {
		return false, nil
	}

	if e.Options.DryRun {
		clog.DryRun("Would merge VLAN %d: %s", vlan.VID, vlan.Name)
		return true, nil
	}

	resp, err := e.Client.IpamVlansPartialUpdateWithResponse(ctx, nautobotID,
		&nautobotapi.IpamVlansPartialUpdateParams{}, *req)
	if err != nil {
		return false, fmt.Errorf("API error: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return false, fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	clog.Changed("Merged VLAN %d: %s", vlan.VID, vlan.Name)
	return true, nil
}

// fetchVLANByID retrieves a full VLAN from Nautobot by UUID.
func (e *Exporter) fetchVLANByID(ctx context.Context, id uuid.UUID) (*nautobotapi.VLAN, error) {
	resp, err := e.Client.IpamVlansRetrieveWithResponse(ctx, id, &nautobotapi.IpamVlansRetrieveParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VLAN %s: %w", id, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch VLAN %s: status %d", id, resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("VLAN %s not found", id)
	}
	return resp.JSON200, nil
}

// buildVLANPatch compares local VLAN intent against the remote state
// and returns a PATCH request with drifted fields.
func buildVLANPatch(vlan *devicetypes.CaniVLAN, remote *nautobotapi.VLAN) (*nautobotapi.PatchedVLANRequest, bool) {
	req := &nautobotapi.PatchedVLANRequest{}
	drifted := false

	if vlan.Name != remote.Name {
		req.Name = &vlan.Name
		drifted = true
	}
	if vlan.Description != "" && vlan.Description != ptrStr(remote.Description) {
		req.Description = &vlan.Description
		drifted = true
	}

	// Compare custom fields (explicit + flattened provider metadata)
	localCF := mergedCustomFields(vlan.CustomFields, vlan.FlattenProviderMetadata())
	if len(localCF) > 0 && customFieldsDrifted(localCF, remote.CustomFields) {
		req.CustomFields = &localCF
		drifted = true
	}

	return req, drifted
}
