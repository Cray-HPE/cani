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
	"context"
	"fmt"
	"net/http"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// reconcileCustomFieldDrift compares the remote custom field definition against
// the local intent and PATCHes any drifted fields (label, type, content_types,
// required, default, weight, description).
func (e *Exporter) reconcileCustomFieldDrift(ctx context.Context, id uuid.UUID, remote nautobotapi.CustomField, local devicetypes.CustomFieldDefinition) error {
	var drifted bool
	req := nautobotapi.PatchedWritableCustomFieldRequest{}

	if remote.Label != local.Label {
		req.Label = &local.Label
		drifted = true
	}
	if remote.Type.Value != nil && string(*remote.Type.Value) != local.Type {
		t := nautobotapi.CustomFieldTypeChoices(local.Type)
		req.Type = &t
		drifted = true
	}
	if !stringSlicesEqual(remote.ContentTypes, local.ContentTypes) {
		req.ContentTypes = &local.ContentTypes
		drifted = true
	}
	remoteRequired := remote.Required != nil && *remote.Required
	if remoteRequired != local.Required {
		req.Required = &local.Required
		drifted = true
	}
	remoteWeight := 0
	if remote.Weight != nil {
		remoteWeight = *remote.Weight
	}
	if remoteWeight != local.Weight {
		req.Weight = &local.Weight
		drifted = true
	}
	remoteDesc := ""
	if remote.Description != nil {
		remoteDesc = *remote.Description
	}
	if remoteDesc != local.Description {
		req.Description = &local.Description
		drifted = true
	}

	if !drifted {
		return nil
	}

	if e.Options.DryRun {
		clog.DryRun("Would update drifted custom field definition: %s (key=%s)", local.Label, local.Key)
		return nil
	}

	resp, err := e.Client.ExtrasCustomFieldsPartialUpdateWithResponse(ctx, id,
		&nautobotapi.ExtrasCustomFieldsPartialUpdateParams{}, req)
	if err != nil {
		return fmt.Errorf("PATCH custom field %q: %w", local.Key, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("PATCH custom field %q: status %d: %s", local.Key, resp.StatusCode(), string(resp.Body))
	}
	clog.Changed("Reconciled drifted custom field: %s (key=%s)", local.Label, local.Key)
	return nil
}

// stringSlicesEqual reports whether two string slices contain the same elements
// (order-insensitive).
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[string]int, len(a))
	for _, s := range a {
		counts[s]++
	}
	for _, s := range b {
		counts[s]--
		if counts[s] < 0 {
			return false
		}
	}
	return true
}
