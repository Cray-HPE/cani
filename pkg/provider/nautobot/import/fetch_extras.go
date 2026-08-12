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

// FetchStatuses retrieves all statuses from the Nautobot API.
func FetchStatuses(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Status, error) {
	return paginate(ctx, "statuses", func(ctx context.Context, offset int) (pageResult[nautobotapi.Status], error) {
		resp, err := client.ExtrasStatusesListWithResponse(ctx, &nautobotapi.ExtrasStatusesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.Status]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "statuses",
			func(b *nautobotapi.PaginatedStatusList) ([]nautobotapi.Status, *string) {
				return b.Results, b.Next
			})
	})
}

// FetchRoles retrieves all roles from the Nautobot API.
func FetchRoles(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Role, error) {
	return paginate(ctx, "roles", func(ctx context.Context, offset int) (pageResult[nautobotapi.Role], error) {
		resp, err := client.ExtrasRolesListWithResponse(ctx, &nautobotapi.ExtrasRolesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.Role]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "roles",
			func(b *nautobotapi.PaginatedRoleList) ([]nautobotapi.Role, *string) {
				return b.Results, b.Next
			})
	})
}
