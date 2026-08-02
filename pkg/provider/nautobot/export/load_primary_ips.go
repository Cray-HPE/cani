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
	var errs []string

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

		// Resolve each family independently; warn but don't skip the other.
		payload := map[string]any{}

		if device.PrimaryIPv4 != uuid.Nil {
			ip := inventory.IPAddresses[device.PrimaryIPv4]
			if ip != nil {
				if id, ok := ip.ExternalIDs["nautobot"]; ok && id != uuid.Nil {
					payload["primary_ip4"] = map[string]string{"id": id.String()}
				} else {
					clog.Warn("device %s: primary IPv4 %s has no Nautobot ID, skipping", device.Name, device.PrimaryIPv4)
				}
			} else {
				clog.Warn("device %s: primary IPv4 %s not found in inventory, skipping", device.Name, device.PrimaryIPv4)
			}
		}

		if device.PrimaryIPv6 != uuid.Nil {
			ip := inventory.IPAddresses[device.PrimaryIPv6]
			if ip != nil {
				if id, ok := ip.ExternalIDs["nautobot"]; ok && id != uuid.Nil {
					payload["primary_ip6"] = map[string]string{"id": id.String()}
				} else {
					clog.Warn("device %s: primary IPv6 %s has no Nautobot ID, skipping", device.Name, device.PrimaryIPv6)
				}
			} else {
				clog.Warn("device %s: primary IPv6 %s not found in inventory, skipping", device.Name, device.PrimaryIPv6)
			}
		}

		if len(payload) == 0 {
			continue
		}

		if e.Options.DryRun {
			clog.DryRun("Would set primary IP on device: %s", device.Name)
			continue
		}
		body, err := json.Marshal(payload)
		if err != nil {
			errs = append(errs, fmt.Sprintf("device %s: marshal primary IP: %v", device.Name, err))
			continue
		}

		resp, err := e.Client.DcimDevicesPartialUpdateWithBodyWithResponse(
			ctx, deviceNautobotID,
			&nautobotapi.DcimDevicesPartialUpdateParams{},
			"application/json",
			bytes.NewReader(body))
		if err != nil {
			errs = append(errs, fmt.Sprintf("device %s: primary IP patch: %v", device.Name, err))
			continue
		}
		if resp.StatusCode() != http.StatusOK {
			errs = append(errs, fmt.Sprintf("device %s: primary IP patch: status %d: %s",
				device.Name, resp.StatusCode(), string(resp.Body)))
			continue
		}
		clog.Changed("Set primary IP on device: %s", device.Name)
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d primary IP error(s): %s", len(errs), errs[0])
	}
	return nil
}
