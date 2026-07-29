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

// EnsureCustomFields ensures all custom field definitions from the inventory
// metadata catalog exist in Nautobot. For select/multi-select fields, it also
// ensures the choices exist. This should be called before exporting objects
// that may carry values for these fields.
func (e *Exporter) EnsureCustomFields(ctx context.Context, inventory *devicetypes.Inventory) error {
	defs := inventory.ListCustomFields()
	if len(defs) == 0 {
		return nil
	}

	clog.Header("Custom Fields: ensuring %d definition(s) exist in Nautobot", len(defs))

	for _, def := range defs {
		cfID, err := e.ensureCustomField(ctx, def)
		if err != nil {
			return fmt.Errorf("custom field %q: %w", def.Key, err)
		}

		if len(def.Choices) > 0 {
			if err := e.ensureCustomFieldChoices(ctx, cfID, def); err != nil {
				return fmt.Errorf("custom field %q choices: %w", def.Key, err)
			}
		}
	}
	return nil
}

// ensureCustomField looks up a custom field by label and creates it if missing.
// Returns the Nautobot UUID of the custom field.
func (e *Exporter) ensureCustomField(ctx context.Context, def devicetypes.CustomFieldDefinition) (uuid.UUID, error) {
	labelFilter := []string{def.Label}
	resp, err := e.Client.ExtrasCustomFieldsListWithResponse(ctx,
		&nautobotapi.ExtrasCustomFieldsListParams{
			Label: &labelFilter,
		},
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to list custom fields: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return uuid.Nil, fmt.Errorf("failed to list custom fields: status %d", resp.StatusCode())
	}

	// Check if already exists
	if resp.JSON200 != nil && len(resp.JSON200.Results) > 0 {
		cf := resp.JSON200.Results[0]
		id := toUUID(cf.Id)
		clog.Skipped("Custom field %q already exists (ID: %s)", def.Label, id)
		return id, nil
	}

	// Create the custom field
	fieldType := nautobotapi.CustomFieldTypeChoices(def.Type)
	req := nautobotapi.WritableCustomFieldRequest{
		Label:        def.Label,
		Key:          &def.Key,
		Type:         &fieldType,
		ContentTypes: def.ContentTypes,
		Default:      def.Default,
	}

	if def.Description != "" {
		req.Description = &def.Description
	}
	if def.Required {
		req.Required = &def.Required
	}
	if def.Weight != 0 {
		req.Weight = &def.Weight
	}

	createResp, err := e.Client.ExtrasCustomFieldsCreateWithResponse(ctx,
		&nautobotapi.ExtrasCustomFieldsCreateParams{},
		req,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create custom field: %w", err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		return uuid.Nil, fmt.Errorf("failed to create custom field: status %d: %s",
			createResp.StatusCode(), string(createResp.Body))
	}

	if createResp.JSON201 == nil {
		return uuid.Nil, fmt.Errorf("failed to create custom field: no response body")
	}

	id := toUUID(createResp.JSON201.Id)
	clog.Created("[nautobot] Created custom field: %s (key=%s, type=%s, ID: %s)",
		def.Label, def.Key, def.Type, id)
	return id, nil
}

// ensureCustomFieldChoices ensures all choices for a select/multi-select custom
// field exist in Nautobot.
func (e *Exporter) ensureCustomFieldChoices(ctx context.Context, cfID uuid.UUID, def devicetypes.CustomFieldDefinition) error {
	// List existing choices for this field
	cfFilter := []string{cfID.String()}
	resp, err := e.Client.ExtrasCustomFieldChoicesListWithResponse(ctx,
		&nautobotapi.ExtrasCustomFieldChoicesListParams{
			CustomField: &cfFilter,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to list choices: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to list choices: status %d", resp.StatusCode())
	}

	// Build set of existing choice values
	existing := make(map[string]bool)
	if resp.JSON200 != nil {
		for _, choice := range resp.JSON200.Results {
			existing[choice.Value] = true
		}
	}

	// Create missing choices
	for i, value := range def.Choices {
		if existing[value] {
			clog.Skipped("Custom field %q choice %q already exists", def.Key, value)
			continue
		}

		var cfIDUnion nautobotapi.BulkWritableCableRequestStatusId
		if err := cfIDUnion.FromBulkWritableCableRequestStatusId0(cfID); err != nil {
			return fmt.Errorf("failed to create custom field ID reference: %w", err)
		}

		weight := (i + 1) * 100
		choiceReq := nautobotapi.CustomFieldChoiceRequest{
			CustomField: nautobotapi.BulkWritableCableRequestStatus{Id: &cfIDUnion},
			Value:       value,
			Weight:      &weight,
		}

		choiceResp, err := e.Client.ExtrasCustomFieldChoicesCreateWithResponse(ctx,
			&nautobotapi.ExtrasCustomFieldChoicesCreateParams{},
			choiceReq,
		)
		if err != nil {
			return fmt.Errorf("failed to create choice %q: %w", value, err)
		}
		if choiceResp.StatusCode() != http.StatusCreated {
			return fmt.Errorf("failed to create choice %q: status %d: %s",
				value, choiceResp.StatusCode(), string(choiceResp.Body))
		}
		clog.Created("[nautobot] Created custom field choice: %s -> %q", def.Key, value)
	}

	return nil
}
