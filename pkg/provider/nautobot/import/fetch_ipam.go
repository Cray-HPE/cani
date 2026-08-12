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
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// FetchVRFs retrieves all VRFs from the Nautobot API.
func FetchVRFs(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.VRF, error) {
	return paginate(ctx, "vrfs", func(ctx context.Context, offset int) (pageResult[nautobotapi.VRF], error) {
		resp, err := client.IpamVrfsListWithResponse(ctx, &nautobotapi.IpamVrfsListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.VRF]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "vrfs",
			func(b *nautobotapi.PaginatedVRFList) ([]nautobotapi.VRF, *string) {
				return b.Results, b.Next
			})
	})
}

// FetchVRFDeviceAssignments retrieves all VRF-device assignment records.
func FetchVRFDeviceAssignments(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.VRFDeviceAssignment, error) {
	return paginate(ctx, "vrf-device-assignments", func(ctx context.Context, offset int) (pageResult[nautobotapi.VRFDeviceAssignment], error) {
		resp, err := client.IpamVrfDeviceAssignmentsListWithResponse(ctx, &nautobotapi.IpamVrfDeviceAssignmentsListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.VRFDeviceAssignment]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "vrf-device-assignments",
			func(b *nautobotapi.PaginatedVRFDeviceAssignmentList) ([]nautobotapi.VRFDeviceAssignment, *string) {
				return b.Results, b.Next
			})
	})
}

// FetchVLANs retrieves all VLANs from the Nautobot API.
func FetchVLANs(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.VLAN, error) {
	return paginate(ctx, "vlans", func(ctx context.Context, offset int) (pageResult[nautobotapi.VLAN], error) {
		resp, err := client.IpamVlansListWithResponse(ctx, &nautobotapi.IpamVlansListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.VLAN]{}, err
		}
		return stdPage(resp.StatusCode(), resp.JSON200, "vlans",
			func(b *nautobotapi.PaginatedVLANList) ([]nautobotapi.VLAN, *string) {
				return b.Results, b.Next
			})
	})
}

// The generated client types Nautobot's network/broadcast/host address fields
// as []byte, which Go's encoding/json decodes from base64. Nautobot returns them
// as plain dotted-decimal strings (e.g. "10.0.0.0"), so decoding the generated
// Prefix/IPAddress directly fails with "illegal base64 data". These shadow types
// re-map those keys to json.RawMessage (which the transform ignores) so the rest
// of the object decodes normally. Field names must match the embedded struct so
// the shallower field wins during unmarshal.
type prefixNoBinary struct {
	nautobotapi.Prefix
	Network   json.RawMessage `json:"network"`
	Broadcast json.RawMessage `json:"broadcast"`
}

type ipAddressNoBinary struct {
	nautobotapi.IPAddress
	Host json.RawMessage `json:"host"`
}

// decodeListPage reads a raw list response body into dest, closing the body and
// verifying a 200 status.
func decodeListPage(resp *http.Response, dest any) error {
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dest)
}

// FetchPrefixes retrieves all prefixes from the Nautobot API.
func FetchPrefixes(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Prefix, error) {
	return paginate(ctx, "prefixes", func(ctx context.Context, offset int) (pageResult[nautobotapi.Prefix], error) {
		resp, err := client.IpamPrefixesList(ctx, &nautobotapi.IpamPrefixesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.Prefix]{}, err
		}
		var page struct {
			Next    *string          `json:"next"`
			Results []prefixNoBinary `json:"results"`
		}
		if err := decodeListPage(resp, &page); err != nil {
			return pageResult[nautobotapi.Prefix]{}, err
		}
		items := make([]nautobotapi.Prefix, len(page.Results))
		for i := range page.Results {
			items[i] = page.Results[i].Prefix
		}
		done := page.Next == nil || *page.Next == ""
		return pageResult[nautobotapi.Prefix]{Items: items, Done: done}, nil
	})
}

// FetchIPAddresses retrieves all IP addresses from the Nautobot API.
func FetchIPAddresses(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.IPAddress, error) {
	return paginate(ctx, "ip addresses", func(ctx context.Context, offset int) (pageResult[nautobotapi.IPAddress], error) {
		resp, err := client.IpamIpAddressesList(ctx, &nautobotapi.IpamIpAddressesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return pageResult[nautobotapi.IPAddress]{}, err
		}
		var page struct {
			Next    *string             `json:"next"`
			Results []ipAddressNoBinary `json:"results"`
		}
		if err := decodeListPage(resp, &page); err != nil {
			return pageResult[nautobotapi.IPAddress]{}, err
		}
		items := make([]nautobotapi.IPAddress, len(page.Results))
		for i := range page.Results {
			items[i] = page.Results[i].IPAddress
		}
		done := page.Next == nil || *page.Next == ""
		return pageResult[nautobotapi.IPAddress]{Items: items, Done: done}, nil
	})
}
