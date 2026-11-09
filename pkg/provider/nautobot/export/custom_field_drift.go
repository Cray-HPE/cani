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
// the local intent and PATCHes any drifted fields (label, content_types,
// required, weight, description). Two of Nautobot's rules shape what is sent:
// a definition's type cannot change once created, so a type mismatch is
// reported and left alone; and removing a content type deletes every value of
// the field on that object type, so content types are only ever added.
func (e *Exporter) reconcileCustomFieldDrift(ctx context.Context, id uuid.UUID, remote nautobotapi.CustomField, local devicetypes.CustomFieldDefinition) error {
	reportCustomFieldTypeDrift(remote, local)
	req, drifted := buildCustomFieldPatch(remote, local)
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

// buildCustomFieldPatch collects the mutable definition fields whose remote
// value differs from the local one.
func buildCustomFieldPatch(remote nautobotapi.CustomField, local devicetypes.CustomFieldDefinition) (nautobotapi.PatchedWritableCustomFieldRequest, bool) {
	var drifted bool
	req := nautobotapi.PatchedWritableCustomFieldRequest{}

	if remote.Label != local.Label {
		req.Label = &local.Label
		drifted = true
	}
	if added := mergedContentTypes(remote.ContentTypes, local.ContentTypes, local.Key); added != nil {
		req.ContentTypes = &added
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
	return req, drifted
}

// reportCustomFieldTypeDrift warns when the catalog and Nautobot disagree on
// a definition's type. Nautobot refuses to change a type once created, so
// the values this export sends may be rejected; the operator has to recreate
// the field or change the catalog.
func reportCustomFieldTypeDrift(remote nautobotapi.CustomField, local devicetypes.CustomFieldDefinition) {
	if remote.Type.Value == nil || string(*remote.Type.Value) == local.Type {
		return
	}
	clog.Warn("custom field %q: Nautobot defines it as %s but the inventory catalog says %s; a custom field's type cannot be changed once created, so values of this field may be rejected",
		local.Key, string(*remote.Type.Value), local.Type)
}

// mergedContentTypes returns remote ∪ local when local names a content type
// Nautobot lacks, and nil when there is nothing to add. A content type that
// only Nautobot has is reported and kept: removing it would delete the
// field's values from every object of that type.
func mergedContentTypes(remote, local []string, key string) []string {
	have, wanted := stringSet(remote), stringSet(local)
	merged := append([]string(nil), remote...)
	for _, ct := range local {
		if !have[ct] {
			merged = append(merged, ct)
			have[ct] = true
		}
	}
	for _, ct := range remote {
		if !wanted[ct] {
			clog.Warn("custom field %q: Nautobot also applies it to %s, which the inventory catalog does not list; left in place because removing it would delete those values", key, ct)
		}
	}
	if len(merged) == len(remote) {
		return nil
	}
	return merged
}

// stringSet indexes items for membership tests.
func stringSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}
