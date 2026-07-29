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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// loadPrimaryIPs patches devices that have a primary IPv4/IPv6 address set.
// This runs after IP addresses are created so the Nautobot UUIDs are available.
func (e *Exporter) loadPrimaryIPs(ctx context.Context, inventory *devicetypes.Inventory) error {
	for _, device := range inventory.Devices {
		if device == nil {
			continue
		}
		if device.PrimaryIPv4 == uuid.Nil && device.PrimaryIPv6 == uuid.Nil {
			continue
		}

		deviceNautobotID, ok := device.ExternalIDs["nautobot"]
		if !ok || deviceNautobotID == uuid.Nil {
			continue
		}

		var ipv4NautobotID, ipv6NautobotID uuid.UUID
		changed := false

		if device.PrimaryIPv4 != uuid.Nil {
			ip := inventory.IPAddresses[device.PrimaryIPv4]
			if ip == nil {
				continue
			}
			id, ok := ip.ExternalIDs["nautobot"]
			if !ok || id == uuid.Nil {
				continue
			}
			ipv4NautobotID = id
			changed = true
		}

		if device.PrimaryIPv6 != uuid.Nil {
			ip := inventory.IPAddresses[device.PrimaryIPv6]
			if ip == nil {
				continue
			}
			id, ok := ip.ExternalIDs["nautobot"]
			if !ok || id == uuid.Nil {
				continue
			}
			ipv6NautobotID = id
			changed = true
		}

		if !changed {
			continue
		}

		if e.Options.DryRun {
			clog.DryRun("Would set primary IP on device: %s", device.Name)
			continue
		}

		// Use raw JSON body to avoid sending zero-value fields that trigger
		// Nautobot validation errors (e.g. "face without rack").
		payload := map[string]any{}
		if device.PrimaryIPv4 != uuid.Nil {
			payload["primary_ip4"] = map[string]string{"id": ipv4NautobotID.String()}
		}
		if device.PrimaryIPv6 != uuid.Nil {
			payload["primary_ip6"] = map[string]string{"id": ipv6NautobotID.String()}
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("device %s: marshal primary IP: %w", device.Name, err)
		}

		resp, err := e.Client.DcimDevicesPartialUpdateWithBodyWithResponse(
			ctx, deviceNautobotID,
			&nautobotapi.DcimDevicesPartialUpdateParams{},
			"application/json",
			bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("device %s: primary IP patch: %w", device.Name, err)
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("device %s: primary IP patch: status %d: %s",
				device.Name, resp.StatusCode(), string(resp.Body))
		}
		clog.Changed("Set primary IP on device: %s", device.Name)
	}
	return nil
}
