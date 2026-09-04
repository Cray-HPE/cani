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
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

func TestLoadPrimaryIPs_ContinuesAfterOneFailure(t *testing.T) {
	deviceA := uuid.New()
	deviceB := uuid.New()
	ipA := uuid.New()
	ipB := uuid.New()
	nautobotDevA := uuid.New()
	nautobotDevB := uuid.New()
	nautobotIPA := uuid.New()
	nautobotIPB := uuid.New()

	var patchedDevices []string

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPatch && strings.Contains(r.URL.Path, "dcim/devices") {
			// Fail for device A, succeed for device B
			if strings.Contains(r.URL.Path, nautobotDevA.String()) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, `{"detail":"server error"}`)
				return
			}
			patchedDevices = append(patchedDevices, r.URL.Path)
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
		deviceA: {
			Name:        "switch-a",
			PrimaryIPv4: ipA,
			ObjectMeta:  devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": nautobotDevA}},
		},
		deviceB: {
			Name:        "switch-b",
			PrimaryIPv4: ipB,
			ObjectMeta:  devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": nautobotDevB}},
		},
	}
	inv.IPAddresses = map[uuid.UUID]*devicetypes.CaniIPAddress{
		ipA: {ID: ipA, ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": nautobotIPA}}},
		ipB: {ID: ipB, ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": nautobotIPB}}},
	}

	err := e.loadPrimaryIPs(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error from failed device patch")
	}
	// Device B should still have been patched despite device A failing.
	if len(patchedDevices) == 0 {
		t.Error("expected at least one successful patch after the failure")
	}
}

func TestLoadPrimaryIPs_IndependentFamilyResolution(t *testing.T) {
	deviceID := uuid.New()
	ipv4ID := uuid.New()
	nautobotDev := uuid.New()
	nautobotIPv4 := uuid.New()

	var patchCount int

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch && strings.Contains(r.URL.Path, "dcim/devices") {
			patchCount++
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	inv := devicetypes.NewInventory()
	// Device has IPv4 resolvable but IPv6 points to a non-existent IP.
	inv.Devices = map[uuid.UUID]*devicetypes.CaniDeviceType{
		deviceID: {
			Name:        "dual-stack",
			PrimaryIPv4: ipv4ID,
			PrimaryIPv6: uuid.New(), // points to missing IP
			ObjectMeta:  devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": nautobotDev}},
		},
	}
	inv.IPAddresses = map[uuid.UUID]*devicetypes.CaniIPAddress{
		ipv4ID: {ID: ipv4ID, ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{"nautobot": nautobotIPv4}}},
	}

	err := e.loadPrimaryIPs(context.Background(), inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if patchCount != 1 {
		t.Errorf("expected 1 patch (IPv4 only), got %d", patchCount)
	}
}

// -----------------------------------------------------------------------------
// primaryIPsMatch — pure logic
// -----------------------------------------------------------------------------

// setPrimaryIP sets a device's primary IPv4 reference to the given UUID.
func setPrimaryIP(remote *nautobotapi.Device, id uuid.UUID) {
	setRefID(&remote.PrimaryIp4, id)
}

func TestPrimaryIPsMatch_TrueWhenSameIPv4(t *testing.T) {
	ipID := uuid.New()
	payload := map[string]any{"primary_ip4": map[string]string{"id": ipID.String()}}
	remote := &nautobotapi.Device{}
	setPrimaryIP(remote, ipID)
	if !primaryIPsMatch(payload, remote) {
		t.Error("expected match when IPv4 IDs are the same")
	}
}

func TestPrimaryIPsMatch_FalseWhenDifferentIPv4(t *testing.T) {
	payload := map[string]any{"primary_ip4": map[string]string{"id": uuid.New().String()}}
	remote := &nautobotapi.Device{}
	setPrimaryIP(remote, uuid.New())
	if primaryIPsMatch(payload, remote) {
		t.Error("expected mismatch when IPv4 IDs differ")
	}
}

func TestPrimaryIPsMatch_FalseWhenRemoteHasNoIPv4(t *testing.T) {
	payload := map[string]any{"primary_ip4": map[string]string{"id": uuid.New().String()}}
	remote := &nautobotapi.Device{}
	if primaryIPsMatch(payload, remote) {
		t.Error("expected mismatch when remote has no primary IPv4")
	}
}

func TestPrimaryIPsMatch_TrueWhenNoIPInPayload(t *testing.T) {
	remote := &nautobotapi.Device{}
	setPrimaryIP(remote, uuid.New())
	if !primaryIPsMatch(map[string]any{}, remote) {
		t.Error("empty payload should match anything")
	}
}
