/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// sentIP is the minimal view of the IPAddressRequest body POSTed to
// /ipam/ip-addresses/.
type sentIP struct {
	Address   string     `json:"address"`
	Status    wireIDRef  `json:"status"`
	Namespace wireIDRef  `json:"namespace"`
	Parent    *wireIDRef `json:"parent"`
	Type      string     `json:"type"`
	DnsName   string     `json:"dns_name"`
}

// sentAssign is the minimal view of the IPAddressToInterfaceRequest body POSTed
// to /ipam/ip-address-to-interface/.
type sentAssign struct {
	IpAddress wireIDRef `json:"ip_address"`
	Interface wireIDRef `json:"interface"`
}

// -----------------------------------------------------------------------------
// loadIPAddresses — happy path: create the address and bind it to an interface
// (PASSING scenario)
// -----------------------------------------------------------------------------

// TestLoadIPAddresses_CreatesAddressAndAssignsToInterface verifies the full
// Phase-9 contract: an IP is created with status, parent prefix, and namespace
// resolved to remote Nautobot IDs (plus type and dns_name), then a separate
// join record links the new IP to its resolved interface.
//
// Why it matters: an IP is only useful in Nautobot when parented to a prefix
// and bound to its owning interface; this asserts the exact wire payload.
// Inputs: one device, one interface, one CaniIPAddress (10.0.1.5/24, Host).
// Outputs: one IP POST, one assignment POST, IPAddressesCreated == 1, external
// ID stamped back onto the inventory record.
// Data choice: the interface is pre-cached and prefixMap pre-seeded so the
// assertions isolate payload shape rather than lookup behaviour.
func TestLoadIPAddresses_CreatesAddressAndAssignsToInterface(t *testing.T) {
	// Arrange.
	resetIPAMCaches()

	ipNID := uuid.New() // remote ID assigned to the created IP address
	parentPrefixNID := uuid.New()
	ifaceNID := uuid.New()  // remote ID of the interface the IP binds to
	deviceNID := uuid.New() // remote ID of the owning device

	var ipPosts, assignPosts [][]byte
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path
		switch {
		// Order matters: "ip-address-to-interface" must be matched before the
		// broader "ip-addresses" substring.
		case strings.Contains(path, "ip-address-to-interface"):
			if r.Method == http.MethodPost {
				body, _ := io.ReadAll(r.Body)
				assignPosts = append(assignPosts, body)
				w.WriteHeader(http.StatusCreated)
				_, _ = io.WriteString(w, `{}`)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		case strings.Contains(path, "ip-addresses"):
			if r.Method == http.MethodPost {
				body, _ := io.ReadAll(r.Body)
				ipPosts = append(ipPosts, body)
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"` + ipNID.String() + `"}`))
				return
			}
			// GET lookup -> not found, so the exporter creates it.
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
		}
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	statusID := seedActiveStatus(t, e)
	nsID := uuid.New()
	seedGlobalNamespace(nsID)
	// Pre-cache the resolved interface so assignment skips its own lookup.
	e.Cache.CacheInterface(deviceNID, "mgmt0", &CachedItem{ID: ifaceNID, Name: "mgmt0"})

	deviceID := uuid.New()
	ifaceID := uuid.New()
	parentPrefixID := uuid.New()
	ipID := uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			deviceID: {ID: deviceID, Name: "server-1"},
		},
		Interfaces: map[uuid.UUID]*devicetypes.CaniInterface{
			ifaceID: {ID: ifaceID, Name: "mgmt0", DeviceID: deviceID},
		},
		IPAddresses: map[uuid.UUID]*devicetypes.CaniIPAddress{
			ipID: {
				ID:         ipID,
				Address:    "10.0.1.5/24",
				Type:       devicetypes.IPAddressTypeHost,
				DNSName:    "host.example.com",
				Parent:     parentPrefixID,
				Interfaces: []uuid.UUID{ifaceID},
				ObjectMeta: devicetypes.ObjectMeta{Status: "Active"},
			},
		},
	}

	prefixMap := map[uuid.UUID]uuid.UUID{parentPrefixID: parentPrefixNID}
	deviceMap := map[string]uuid.UUID{"server-1": deviceNID}
	result := &LoadResult{}

	// Act.
	err := e.loadIPAddresses(context.Background(), inv, prefixMap, deviceMap, result)

	// Assert.
	if err != nil {
		t.Fatalf("loadIPAddresses returned error: %v", err)
	}
	if len(ipPosts) != 1 {
		t.Fatalf("expected 1 IP create POST, got %d", len(ipPosts))
	}

	ip := decodeSentIP(t, ipPosts[0])
	if ip.Address != "10.0.1.5/24" {
		t.Errorf("address = %q, want 10.0.1.5/24", ip.Address)
	}
	if ip.Status.ID != statusID.String() {
		t.Errorf("status.id = %q, want %q", ip.Status.ID, statusID)
	}
	if ip.Namespace.ID != nsID.String() {
		t.Errorf("namespace.id = %q, want %q", ip.Namespace.ID, nsID)
	}
	if ip.Parent == nil || ip.Parent.ID != parentPrefixNID.String() {
		t.Errorf("parent.id = %v, want remote prefix ID %q", ip.Parent, parentPrefixNID)
	}
	if ip.Type != string(nautobotapi.Host) {
		t.Errorf("type = %q, want %q", ip.Type, nautobotapi.Host)
	}
	if ip.DnsName != "host.example.com" {
		t.Errorf("dns_name = %q, want host.example.com", ip.DnsName)
	}

	// The join record links the freshly-created IP to the resolved interface.
	if len(assignPosts) != 1 {
		t.Fatalf("expected 1 interface-assignment POST, got %d", len(assignPosts))
	}
	assign := decodeSentAssign(t, assignPosts[0])
	if assign.IpAddress.ID != ipNID.String() {
		t.Errorf("assignment ip_address.id = %q, want %q", assign.IpAddress.ID, ipNID)
	}
	if assign.Interface.ID != ifaceNID.String() {
		t.Errorf("assignment interface.id = %q, want %q", assign.Interface.ID, ifaceNID)
	}

	if result.IPAddressesCreated != 1 {
		t.Errorf("IPAddressesCreated = %d, want 1", result.IPAddressesCreated)
	}
	if got := inv.IPAddresses[ipID].ExternalIDs[externalIDKeyNautobot]; got != ipNID {
		t.Errorf("ExternalIDs[nautobot] = %s, want %s", got, ipNID)
	}
}

// -----------------------------------------------------------------------------
// loadIPAddresses — input validation (skip nil / empty addresses)
// -----------------------------------------------------------------------------

// TestLoadIPAddresses_SkipsNilAndEmptyAddresses verifies the loader's guard
// clauses: a nil record and a record with an empty Address are skipped without
// issuing any create POST.
//
// Why it matters: inventory can contain partially-populated or placeholder IP
// records; sending them would create invalid objects or 400s, so they must be
// dropped silently and not counted as created.
// Inputs: an IPAddresses map with one nil entry and one entry whose Address is
// "". Outputs: zero POSTs and IPAddressesCreated == 0.
// Data choice: nil and empty-string are the two distinct shapes the guard
// clause must catch.
func TestLoadIPAddresses_SkipsNilAndEmptyAddresses(t *testing.T) {
	// Arrange.
	resetIPAMCaches()

	var posts int
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			posts++
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	seedActiveStatus(t, e)
	seedGlobalNamespace(uuid.New())

	nilID := uuid.New()
	emptyID := uuid.New()
	inv := &devicetypes.Inventory{
		IPAddresses: map[uuid.UUID]*devicetypes.CaniIPAddress{
			nilID:   nil,
			emptyID: {ID: emptyID, Address: ""},
		},
	}
	result := &LoadResult{}

	// Act.
	err := e.loadIPAddresses(context.Background(), inv,
		map[uuid.UUID]uuid.UUID{}, map[string]uuid.UUID{}, result)

	// Assert.
	if err != nil {
		t.Fatalf("loadIPAddresses returned error: %v", err)
	}
	if posts != 0 {
		t.Errorf("expected no POSTs for nil/empty addresses, got %d", posts)
	}
	if result.IPAddressesCreated != 0 {
		t.Errorf("IPAddressesCreated = %d, want 0", result.IPAddressesCreated)
	}
}

// -----------------------------------------------------------------------------
// assignIPToInterfaces — guard when the interface is not in inventory
// -----------------------------------------------------------------------------

// TestAssignIPToInterfaces_SkipsUnknownInterface verifies that an IP whose
// Interfaces reference is absent from the inventory is ignored gracefully: no
// join POST is sent, no error is recorded, and nothing panics.
//
// Why it matters: dangling interface references can occur when an export is
// partial or an interface failed earlier; the assignment step must skip them
// rather than abort the whole IPAM phase.
// Inputs: a CaniIPAddress referencing a random interface UUID and an empty
// Interfaces inventory map. Outputs: result.Errors stays empty; the fake
// server fails the test if any POST arrives.
// Data choice: an empty inventory map guarantees the lookup miss that drives
// the skip branch.
func TestAssignIPToInterfaces_SkipsUnknownInterface(t *testing.T) {
	// Arrange: the handler fails the test if any assignment POST arrives.
	resetIPAMCaches()
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			t.Errorf("unexpected POST to %s for an unresolved interface", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	addr := &devicetypes.CaniIPAddress{
		ID:         uuid.New(),
		Address:    "10.0.1.9/24",
		Interfaces: []uuid.UUID{uuid.New()}, // references an interface not in inventory
	}
	inv := &devicetypes.Inventory{
		Interfaces: map[uuid.UUID]*devicetypes.CaniInterface{}, // empty
	}
	result := &LoadResult{}

	// Act / Assert (must not panic).
	e.assignIPToInterfaces(context.Background(), uuid.New(), addr, inv,
		map[string]uuid.UUID{}, result)

	if len(result.Errors) != 0 {
		t.Errorf("expected no errors for a skipped interface, got %v", result.Errors)
	}
}

func decodeSentIP(t *testing.T, body []byte) sentIP {
	t.Helper()
	var ip sentIP
	if err := json.Unmarshal(body, &ip); err != nil {
		t.Fatalf("decode IP payload: %v\nbody: %s", err, body)
	}
	return ip
}

func decodeSentAssign(t *testing.T, body []byte) sentAssign {
	t.Helper()
	var a sentAssign
	if err := json.Unmarshal(body, &a); err != nil {
		t.Fatalf("decode assignment payload: %v\nbody: %s", err, body)
	}
	return a
}
