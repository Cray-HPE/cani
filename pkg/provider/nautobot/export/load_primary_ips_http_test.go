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
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func TestLoadPrimaryIPs_SkipsWhenAlreadySet(t *testing.T) {
	deviceID := uuid.New()
	ipID := uuid.New()
	nautobotDev := uuid.New()
	nautobotIP := uuid.New()

	var patchCalls int

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// GET device retrieve — return device with matching primary_ip4
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "dcim/devices/"+nautobotDev.String()) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"id":%q,"name":"sw1","primary_ip4":{"id":%q},"device_type":{"id":%q},"location":{"id":%q},"status":{"id":%q}}`,
				nautobotDev, nautobotIP, uuid.New(), uuid.New(), uuid.New())
			return
		}
		if r.Method == http.MethodPatch {
			patchCalls++
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	inv := devicetypes.NewInventory()
	inv.Devices = map[uuid.UUID]*devicetypes.CaniDeviceType{
		deviceID: {
			Name:        "sw1",
			PrimaryIPv4: ipID,
			ObjectMeta:  devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": nautobotDev}},
		},
	}
	inv.IPAddresses = map[uuid.UUID]*devicetypes.CaniIPAddress{
		ipID: {ID: ipID, ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": nautobotIP}}},
	}

	err := e.loadPrimaryIPs(context.Background(), inv, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if patchCalls != 0 {
		t.Errorf("expected 0 PATCH calls (already set), got %d", patchCalls)
	}
}

// TestLoadPrimaryIPs_DryRunDoesNotPatch verifies a dry run previews a primary
// IP the same run planned without patching the device.
//
// Why it matters: a dry run records no external IDs, so the primary-IP phase
// must resolve a planned address from the IDs the IP phase returned, or the
// preview would report the address as missing.
// Inputs: a device with a Nautobot ID and a primary IPv4 whose address carries
// no external ID but is in the per-run map; a server whose device has no
// primary IP set. Outputs: one device GET (the comparison happened) and no
// PATCH.
// Data choice: an address known only through the map is exactly what a dry run
// produces for an address it would create.
func TestLoadPrimaryIPs_DryRunDoesNotPatch(t *testing.T) {
	deviceID := uuid.New()
	ipID := uuid.New()
	nautobotDev := uuid.New()
	nautobotIP := uuid.New()

	var patchCalls, deviceFetches int

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// GET device retrieve — return device with no primary IP set
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "dcim/devices/"+nautobotDev.String()) {
			deviceFetches++
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"id":%q,"name":"sw1","device_type":{"id":%q},"location":{"id":%q},"status":{"id":%q}}`,
				nautobotDev, uuid.New(), uuid.New(), uuid.New())
			return
		}
		if r.Method == http.MethodPatch {
			patchCalls++
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	e.Options.DryRun = true

	inv := devicetypes.NewInventory()
	inv.Devices = map[uuid.UUID]*devicetypes.CaniDeviceType{
		deviceID: {
			Name:        "sw1",
			PrimaryIPv4: ipID,
			ObjectMeta:  devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": nautobotDev}},
		},
	}
	inv.IPAddresses = map[uuid.UUID]*devicetypes.CaniIPAddress{
		ipID: {ID: ipID},
	}

	// The address is one this dry run planned: it has no recorded external ID
	// and is known only through the IDs the IP phase returned.
	err := e.loadPrimaryIPs(context.Background(), inv, map[uuid.UUID]uuid.UUID{ipID: nautobotIP})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if patchCalls != 0 {
		t.Errorf("dry-run should issue 0 PATCH calls, got %d", patchCalls)
	}
	if deviceFetches != 1 {
		t.Errorf("device fetches = %d, want 1: the planned address must resolve so the dry run compares remote state", deviceFetches)
	}
}
