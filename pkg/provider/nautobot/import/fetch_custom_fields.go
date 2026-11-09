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
package imprt

import (
	"context"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// FetchCustomFields retrieves every custom-field definition. The values on
// imported objects are only as portable as their definitions: a fresh
// destination has none, so the export must be able to recreate them before
// it sends a value.
func FetchCustomFields(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.CustomField, error) {
	return paginate(ctx, "custom fields", func(ctx context.Context, offset int) (pageResult[nautobotapi.CustomField], error) {
		resp, err := client.ExtrasCustomFieldsListWithResponse(ctx, &nautobotapi.ExtrasCustomFieldsListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.CustomField]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "custom fields",
			func(b *nautobotapi.PaginatedCustomFieldList) ([]nautobotapi.CustomField, *string) {
				return b.Results, b.Next
			})
	})
}

// FetchCustomFieldChoices retrieves the choices of every select and
// multi-select custom field, which the definition itself does not carry.
func FetchCustomFieldChoices(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.CustomFieldChoice, error) {
	return paginate(ctx, "custom field choices", func(ctx context.Context, offset int) (pageResult[nautobotapi.CustomFieldChoice], error) {
		resp, err := client.ExtrasCustomFieldChoicesListWithResponse(ctx, &nautobotapi.ExtrasCustomFieldChoicesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.CustomFieldChoice]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "custom field choices",
			func(b *nautobotapi.PaginatedCustomFieldChoiceList) ([]nautobotapi.CustomFieldChoice, *string) {
				return b.Results, b.Next
			})
	})
}
