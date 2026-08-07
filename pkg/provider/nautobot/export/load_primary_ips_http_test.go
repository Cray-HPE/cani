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

	err := e.loadPrimaryIPs(context.Background(), inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if patchCalls != 0 {
		t.Errorf("expected 0 PATCH calls (already set), got %d", patchCalls)
	}
}

func TestLoadPrimaryIPs_DryRunDoesNotPatch(t *testing.T) {
	deviceID := uuid.New()
	ipID := uuid.New()
	nautobotDev := uuid.New()
	nautobotIP := uuid.New()

	var patchCalls int

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// GET device retrieve — return device with no primary IP set
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "dcim/devices/"+nautobotDev.String()) {
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
		ipID: {ID: ipID, ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": nautobotIP}}},
	}

	err := e.loadPrimaryIPs(context.Background(), inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if patchCalls != 0 {
		t.Errorf("dry-run should issue 0 PATCH calls, got %d", patchCalls)
	}
}
