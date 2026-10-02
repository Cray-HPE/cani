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

// ensureVRFPrefixAssignments links a prefix to each VRF it is a member of.
// It is a find-or-create keyed on (vrf, prefix) so repeated exports create no
// duplicate assignment records, and it runs for existing prefixes too so a
// membership added locally reaches Nautobot. A legacy VRF name that never
// resolved into a membership is reported rather than guessed.
func (e *Exporter) ensureVRFPrefixAssignments(
	ctx context.Context,
	prefix *devicetypes.CaniPrefix,
	remote *CachedItem,
	inventory *devicetypes.Inventory,
	result *LoadResult,
) {
	owner := prefix.EffectiveNamespace() + "/" + prefix.Prefix
	if prefix.VRF != "" {
		result.Errors = append(result.Errors, fmt.Sprintf(
			"prefix %s: legacy VRF name %q does not resolve to one VRF in the namespace; set vrfs explicitly",
			owner, prefix.VRF))
	}
	for _, vrfID := range prefix.VRFs {
		vrf := inventory.VRFs[vrfID]
		if vrf == nil {
			result.Errors = append(result.Errors, fmt.Sprintf("prefix %s: VRF %s not found in inventory", owner, vrfID))
			continue
		}
		cached := e.Cache.LookupVRF(vrf.EffectiveNamespace(), vrf.Name)
		if cached == nil {
			result.Errors = append(result.Errors, fmt.Sprintf("prefix %s: VRF %q not exported; assignment skipped", owner, vrf.Name))
			continue
		}
		if err := e.ensureVRFPrefixAssignment(ctx, cached, remote, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("prefix %s: VRF %q assignment: %v", owner, vrf.Name, err))
			continue
		}
		clog.Detail("  Prefix %s: member of VRF %s", owner, vrf.Name)
	}
}

// ensureVRFPrefixAssignment creates one VRF-prefix assignment unless it exists.
// A dry run still asks Nautobot whether the assignment exists when both ends
// do, so its summary separates the links it would create from those in place.
func (e *Exporter) ensureVRFPrefixAssignment(ctx context.Context, vrf, prefix *CachedItem, result *LoadResult) error {
	exists, err := e.vrfPrefixAssignmentExists(vrf, prefix)
	if err != nil {
		return err
	}
	if exists {
		result.VRFPrefixAssignmentsSkipped++
		return nil
	}
	if e.Options.DryRun {
		clog.DryRun("Would assign prefix %s to VRF %s", prefix.Name, vrf.Name)
	} else if err := e.createVRFPrefixAssignment(ctx, vrf.ID, prefix.ID); err != nil {
		return err
	}
	result.VRFPrefixAssignmentsCreated++
	return nil
}

// vrfPrefixAssignmentExists reports whether Nautobot already links the VRF
// and the prefix, from the VRF's listed assignments. An end a dry run only
// plans cannot be assigned yet, so no lookup is made for it.
func (e *Exporter) vrfPrefixAssignmentExists(vrf, prefix *CachedItem) (bool, error) {
	if vrf.Planned || prefix.Planned {
		return false, nil
	}
	return e.Cache.VRFPrefixAssigned(vrf.ID, prefix.ID)
}

// createVRFPrefixAssignment POSTs one VRF-prefix assignment and records the
// link in the cache.
func (e *Exporter) createVRFPrefixAssignment(ctx context.Context, vrfID, prefixID uuid.UUID) error {
	req := nautobotapi.VRFPrefixAssignmentRequest{}
	if err := setRefID(&req.Vrf, vrfID); err != nil {
		return fmt.Errorf("set assignment VRF reference: %w", err)
	}
	if err := setRefID(&req.Prefix, prefixID); err != nil {
		return fmt.Errorf("set assignment prefix reference: %w", err)
	}
	createResp, err := e.Client.IpamVrfPrefixAssignmentsCreateWithResponse(ctx,
		&nautobotapi.IpamVrfPrefixAssignmentsCreateParams{}, req)
	if err != nil {
		return fmt.Errorf("API error: %w", err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("unexpected status %d: %s", createResp.StatusCode(), string(createResp.Body))
	}
	e.Cache.CacheVRFPrefixAssignment(vrfID, prefixID)
	return nil
}
