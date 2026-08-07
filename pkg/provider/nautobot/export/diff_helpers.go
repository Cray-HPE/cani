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

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// fetchFullDeviceByID retrieves the full Device object from the Nautobot API by UUID.
func (e *Exporter) fetchFullDeviceByID(ctx context.Context, id uuid.UUID) (*nautobotapi.Device, error) {
	resp, err := e.Client.DcimDevicesRetrieveWithResponse(ctx, id, &nautobotapi.DcimDevicesRetrieveParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch device %s: %w", id, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch device %s: status %d", id, resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("device %s not found", id)
	}
	return resp.JSON200, nil
}

// fetchFullDevice retrieves the full Device object from the Nautobot API by name.
func (e *Exporter) fetchFullDevice(ctx context.Context, name string) (*nautobotapi.Device, error) {
	nameFilter := []string{name}
	resp, err := e.Client.DcimDevicesListWithResponse(ctx, &nautobotapi.DcimDevicesListParams{
		Name: &nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch full device %s: %w", name, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch full device %s: status %d", name, resp.StatusCode())
	}
	if resp.JSON200 == nil || resp.JSON200.Results == nil || len(resp.JSON200.Results) == 0 {
		return nil, fmt.Errorf("device %s not found", name)
	}
	d := resp.JSON200.Results[0]
	return &d, nil
}

// ptrStr dereferences a *string, returning "" for nil.
func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// orNone returns "(none)" when s is empty.
func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

// printDeviceDiffs prints the field diffs for a device using colored output.
func printDeviceDiffs(deviceName string, diffs []FieldDiff) {
	if len(diffs) == 0 {
		return
	}
	clog.Changed("  %s: %d field(s) would change with --merge:", deviceName, len(diffs))
	for _, d := range diffs {
		clog.Diff(d.Field, d.RemoteVal, d.LocalVal)
	}
}

// mergedCustomFields combines explicit custom fields and flattened provider metadata.
func mergedCustomFields(explicit map[string]any, flat map[string]any) map[string]interface{} {
	cf := make(map[string]interface{}, len(explicit)+len(flat))
	for k, v := range explicit {
		cf[k] = v
	}
	for k, v := range flat {
		cf[k] = v
	}
	return cf
}

// customFieldsDrifted returns true if any local custom field value differs from remote.
func customFieldsDrifted(local map[string]interface{}, remote *map[string]interface{}) bool {
	remoteMap := map[string]interface{}{}
	if remote != nil {
		remoteMap = *remote
	}
	for k, v := range local {
		localStr := fmt.Sprintf("%v", v)
		remoteVal, exists := remoteMap[k]
		remoteStr := ""
		if exists {
			remoteStr = fmt.Sprintf("%v", remoteVal)
		}
		if localStr != remoteStr {
			return true
		}
	}
	return false
}
