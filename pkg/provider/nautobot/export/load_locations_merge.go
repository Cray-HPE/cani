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

// mergeLocation fetches the remote location, compares against local intent,
// and PATCHes any drifted fields. Returns true if the location was updated.
func (e *Exporter) mergeLocation(ctx context.Context, loc *devicetypes.CaniLocationType, nautobotID uuid.UUID) (bool, error) {
	remote, err := e.fetchLocationByID(ctx, nautobotID)
	if err != nil {
		return false, err
	}

	req, drifted := buildLocationPatch(loc, remote)
	if !drifted {
		return false, nil
	}

	if e.Options.DryRun {
		clog.DryRun("Would merge location: %s", loc.Name)
		return true, nil
	}

	resp, err := e.Client.DcimLocationsPartialUpdateWithResponse(ctx, nautobotID,
		&nautobotapi.DcimLocationsPartialUpdateParams{}, *req)
	if err != nil {
		return false, fmt.Errorf("API error: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return false, fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	clog.Changed("Merged location: %s", loc.Name)
	return true, nil
}

// fetchLocationByID retrieves a full Location from Nautobot by UUID.
func (e *Exporter) fetchLocationByID(ctx context.Context, id uuid.UUID) (*nautobotapi.Location, error) {
	resp, err := e.Client.DcimLocationsRetrieveWithResponse(ctx, id, &nautobotapi.DcimLocationsRetrieveParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch location %s: %w", id, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch location %s: status %d", id, resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("location %s not found", id)
	}
	return resp.JSON200, nil
}

// buildLocationPatch compares local location intent against the remote state
// and returns a PATCH request with drifted fields.
func buildLocationPatch(loc *devicetypes.CaniLocationType, remote *nautobotapi.Location) (*nautobotapi.PatchedLocationRequest, bool) {
	req := &nautobotapi.PatchedLocationRequest{}
	drifted := false

	drifted = driftStr(loc.Description, remote.Description, &req.Description) || drifted
	drifted = driftStr(loc.Facility, remote.Facility, &req.Facility) || drifted
	drifted = driftStr(loc.PhysicalAddress, remote.PhysicalAddress, &req.PhysicalAddress) || drifted
	drifted = driftStr(loc.ShippingAddress, remote.ShippingAddress, &req.ShippingAddress) || drifted
	drifted = driftStr(loc.ContactName, remote.ContactName, &req.ContactName) || drifted
	drifted = driftStr(loc.ContactPhone, remote.ContactPhone, &req.ContactPhone) || drifted
	drifted = driftStr(loc.ContactEmail, remote.ContactEmail, &req.ContactEmail) || drifted
	drifted = driftStr(loc.TimeZone, remote.TimeZone, &req.TimeZone) || drifted
	drifted = driftStr(loc.Latitude, remote.Latitude, &req.Latitude) || drifted
	drifted = driftStr(loc.Longitude, remote.Longitude, &req.Longitude) || drifted
	drifted = driftStr(loc.Comments, remote.Comments, &req.Comments) || drifted

	// Compare custom fields (explicit + flattened provider metadata)
	localCF := mergedCustomFields(loc.CustomFields, loc.FlattenProviderMetadata())
	if len(localCF) > 0 && customFieldsDrifted(localCF, remote.CustomFields) {
		req.CustomFields = &localCF
		drifted = true
	}

	return req, drifted
}

// driftStr sets target when local differs from remote (including clearing).
func driftStr(local string, remote *string, target **string) bool {
	if local != ptrStr(remote) {
		*target = &local
		return true
	}
	return false
}
