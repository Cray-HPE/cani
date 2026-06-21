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
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// loadDeviceServer answers the requests a single-device export makes:
//   - the device-create POST returns the supplied deviceID,
//   - the interface bulk POST returns one created interface,
//   - every list lookup (device-by-name, interface prefetch, role) is empty so
//     the orchestrator takes the create branch.
//
// calls counts every request so a test can assert the network was (or was not)
// touched.
func loadDeviceServer(deviceID, ifaceID uuid.UUID, calls *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*calls++
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "dcim/interfaces"):
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `[{"id":%q,"name":"eth0","display":"eth0"}]`, ifaceID.String())
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "dcim/devices"):
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"id":%q,"name":"compute-001","display":"compute-001"}`, deviceID.String())
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":0,"results":[]}`))
		}
	}
}

// -----------------------------------------------------------------------------
// Load — top-level orchestrator
// -----------------------------------------------------------------------------

// TestLoad_EmptyInventorySucceedsWithoutNetwork verifies Load on an empty
// inventory returns nil and issues zero HTTP requests.
//
// Why it matters: exporting nothing must be a cheap no-op — the orchestrator
// should not touch the Nautobot API when there are no locations, racks, or
// devices to sync, keeping re-runs and partial inventories side-effect free.
// Inputs: an empty *devicetypes.Inventory; Active status is pre-seeded so the
// interface phase resolves from cache. Outputs: nil error and call count 0.
// Data choice: seedActiveStatus removes the one lookup an empty run would
// otherwise need, isolating the "no work" guarantee.
func TestLoad_EmptyInventorySucceedsWithoutNetwork(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusOK, emptyListJSON))
	defer cleanup()
	// Seeding Active lets the interface phase resolve its status from cache, so
	// an empty inventory needs no network at all.
	seedActiveStatus(t, e)

	if err := e.Load(&devicetypes.Inventory{}); err != nil {
		t.Fatalf("Load(empty) error = %v", err)
	}
	if calls != 0 {
		t.Errorf("expected no HTTP for an empty inventory, got %d calls", calls)
	}
}

// TestLoad_AggregatesPhaseErrors verifies Load keeps running after a phase
// failure and returns a single aggregated "errors during sync" error.
//
// Why it matters: a real export touches many objects across phases; one failed
// rack lookup must not abort the whole sync, so failures are collected and
// reported together instead of fatally short-circuiting later phases.
// Inputs: an inventory with one rack; the handler returns 500 for dcim/racks and
// 200 elsewhere. Outputs: a non-nil error whose message contains
// "errors during sync".
// Data choice: a single rack plus a path-scoped 500 forces exactly one phase
// error, making the aggregation behaviour unambiguous.
func TestLoad_AggregatesPhaseErrors(t *testing.T) {
	// The rack lookup fails, so Phase 1 records an error and Load surfaces it as
	// an aggregated failure after running the remaining phases.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "dcim/racks") {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"detail":"boom"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(emptyListJSON))
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)

	inv := &devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			uuid.New(): {Name: "rack-1"},
		},
	}

	err := e.Load(inv)
	if err == nil || !strings.Contains(err.Error(), "errors during sync") {
		t.Fatalf("expected an aggregated sync error, got %v", err)
	}
}

// TestLoad_CreatesDeviceAndStampsExternalID verifies Load creates a new device
// and writes the returned Nautobot UUID back into
// device.ExternalIDs["nautobot"].
//
// Why it matters: stamping the external ID is what makes export idempotent — a
// later run resolves the device by its recorded UUID instead of creating a
// duplicate in Nautobot.
// Inputs: an inventory with one mappable device carrying an eth0 interface, refs
// and Active status seeded; the fake server returns fixed device/interface IDs
// on POST. Outputs: device.ExternalIDs[nautobot] equals the created device ID.
// Data choice: a device plus one interface exercises the create branch through
// the interface phase while keeping the asserted ID round-trip exact.
func TestLoad_CreatesDeviceAndStampsExternalID(t *testing.T) {
	deviceID := uuid.New()
	ifaceID := uuid.New()
	var calls int
	e, cleanup := newExporterWithServer(t, loadDeviceServer(deviceID, ifaceID, &calls))
	defer cleanup()

	// Seed the references Load's internal mapper resolves, plus Active status.
	seedDeviceRefs(t, e)
	e.Options.DefaultLocation = "DC1"
	e.Options.DefaultStatus = "Active"
	e.Options.DefaultRole = "Compute"

	device := newMappableDevice("compute-001")
	device.Interfaces = []devicetypes.InterfaceSpec{
		{Name: "eth0", Type: devicetypes.InterfacesElemType("1000base-t")},
	}
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			uuid.New(): device,
		},
	}

	if err := e.Load(inv); err != nil {
		t.Fatalf("Load(single device) error = %v", err)
	}

	// Load stamps the new Nautobot ID back onto the device so later exports
	// resolve it directly.
	if got := device.ExternalIDs[externalIDKeyNautobot]; got != deviceID {
		t.Errorf("device external ID = %s, want %s", got, deviceID)
	}
}
