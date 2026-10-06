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
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// emptyNautobot imitates a Nautobot that holds nothing except one status
// whose content types are incomplete, and records every request it receives.
// Every lookup therefore misses, so each create path the options enable is
// attempted, and the status lookup finds something to patch. A list filtered
// by a namespace answers 400, as Nautobot does for a namespace it does not
// hold.
type emptyNautobot struct {
	mu       sync.Mutex
	requests []string
}

func (s *emptyNautobot) handler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.requests = append(s.requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method != http.MethodGet:
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":"`+uuid.New().String()+`","name":"created","display":"created"}`)
	case strings.Contains(r.URL.Path, "extras/statuses"):
		_, _ = io.WriteString(w, `{"count":1,"results":[{"id":"`+uuid.New().String()+
			`","name":"Active","display":"Active","content_types":[]}]}`)
	case r.URL.Query().Has("namespace"):
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"namespace":["Select a valid choice."]}`)
	default:
		_, _ = io.WriteString(w, emptyListJSON)
	}
}

// writes returns every request that was not a GET.
func (s *emptyNautobot) writes() []string {
	return s.requestsMatching(func(request string) bool { return !strings.HasPrefix(request, http.MethodGet+" ") })
}

// requestsMatching returns the recorded requests the predicate accepts.
func (s *emptyNautobot) requestsMatching(accept func(string) bool) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var matches []string
	for _, request := range s.requests {
		if accept(request) {
			matches = append(matches, request)
		}
	}
	return matches
}

// addTaggedTenantIPAM adds a tagged VRF, a member prefix and an address in a
// namespace an empty Nautobot does not have, and returns their IDs. The
// fixture's JSON carries no IPAM collections, so they are initialised here.
func addTaggedTenantIPAM(inv *devicetypes.Inventory) (vrfID, prefixID, addressID uuid.UUID) {
	if inv.VRFs == nil {
		inv.VRFs = map[uuid.UUID]*devicetypes.CaniVRF{}
	}
	if inv.Prefixes == nil {
		inv.Prefixes = map[uuid.UUID]*devicetypes.CaniPrefix{}
	}
	if inv.IPAddresses == nil {
		inv.IPAddresses = map[uuid.UUID]*devicetypes.CaniIPAddress{}
	}
	vrfID, prefixID, addressID = uuid.New(), uuid.New(), uuid.New()
	tags := devicetypes.ObjectMeta{Tags: []string{"dry-run-probe"}}
	inv.VRFs[vrfID] = &devicetypes.CaniVRF{ID: vrfID, Name: "blue", Namespace: "tenant-z", ObjectMeta: tags}
	inv.Prefixes[prefixID] = &devicetypes.CaniPrefix{
		ID: prefixID, Prefix: "10.9.0.0/24", PrefixLen: 24, Namespace: "tenant-z", VRFs: []uuid.UUID{vrfID},
	}
	inv.IPAddresses[addressID] = &devicetypes.CaniIPAddress{
		ID: addressID, Host: "10.9.0.5", Address: "10.9.0.5/24", Parent: prefixID,
	}
	return vrfID, prefixID, addressID
}

// recordedExternalIDs lists the IPAM records that carry a Nautobot external
// ID, and the racks and locations that carry an entry for one at all: a dry
// run against an empty Nautobot creates every object, so none may be stamped,
// not even with a nil ID.
func recordedExternalIDs(inv *devicetypes.Inventory) []string {
	var recorded []string
	for _, vrf := range inv.VRFs {
		if vrf.ExternalIDs[externalIDKeyNautobot] != uuid.Nil {
			recorded = append(recorded, "vrf "+vrf.Name)
		}
	}
	for _, prefix := range inv.Prefixes {
		if prefix.ExternalIDs[externalIDKeyNautobot] != uuid.Nil {
			recorded = append(recorded, "prefix "+prefix.Prefix)
		}
	}
	for _, addr := range inv.IPAddresses {
		if addr.ExternalIDs[externalIDKeyNautobot] != uuid.Nil {
			recorded = append(recorded, "ip "+addr.Address)
		}
	}
	for _, rack := range inv.Racks {
		if _, stamped := rack.ExternalIDs[externalIDKeyNautobot]; stamped {
			recorded = append(recorded, "rack "+rack.Name)
		}
	}
	for _, loc := range inv.Locations {
		if _, stamped := loc.ExternalIDs[externalIDKeyNautobot]; stamped {
			recorded = append(recorded, "location "+loc.Name)
		}
	}
	return recorded
}

// TestLoadDryRunIssuesNoWrites verifies a full dry-run export against an empty
// Nautobot, with every auto-create option on, sends nothing but reads and
// leaves no Nautobot ID on the IPAM records.
//
// Why it matters: --dry-run promises no change to Nautobot, yet the lookup
// cache created statuses, roles, locations, location types, tags,
// manufacturers, device types and namespaces, and patched content types, on
// its own whenever it found them missing; the IPAM loaders also stored the
// placeholder IDs a dry run invents, which the next import rejected as a
// conflicting source identity once saved.
// Inputs: the spine-leaf fixture with one tagged device plus a tagged VRF, a
// member prefix and an address in a new namespace; a server that holds only a
// status with no content types and answers 400 to any namespace filter;
// DryRun with every create option enabled. Outputs: no write request and no
// external ID on the VRF, prefix or address.
// Data choice: an empty Nautobot makes every lookup miss, so every create path
// the options enable is attempted; the tag reaches the one create the options
// do not gate, the incomplete status reaches the content-type patch, and the
// 400 is what a dry run that merely suppressed the POST would still trip over.
func TestLoadDryRunIssuesNoWrites(t *testing.T) {
	// Arrange.
	ensureTypesLoaded(t)
	server := &emptyNautobot{}
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	for _, enable := range []func(bool){
		e.Cache.SetCreateDeviceTypes, e.Cache.SetCreateLocationTypes, e.Cache.SetCreateLocations,
		e.Cache.SetCreateStatuses, e.Cache.SetCreateRoles,
	} {
		enable(true)
	}
	e.Options.DryRun = true
	e.Options.DefaultStatus = "Active"
	e.Options.DefaultRole = "Leaf"
	inv := loadSpineLeafInventory(t)
	for _, device := range inv.Devices {
		device.Tags = []string{"dry-run-probe"}
		break
	}
	addTaggedTenantIPAM(inv)

	// Act.
	_ = e.Load(inv)

	// Assert.
	if writes := server.writes(); len(writes) != 0 {
		t.Errorf("dry run wrote to Nautobot (%d requests): %v", len(writes), writes)
	}
	if recorded := recordedExternalIDs(inv); len(recorded) != 0 {
		t.Errorf("dry run recorded Nautobot IDs on %v", recorded)
	}
}
