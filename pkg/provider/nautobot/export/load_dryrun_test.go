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

	"github.com/google/uuid"
)

// emptyNautobot imitates a Nautobot that holds nothing except one status
// whose content types are incomplete, and records every request it receives.
// Every lookup therefore misses, so each create path the options enable is
// attempted, and the status lookup finds something to patch.
type emptyNautobot struct {
	mu       sync.Mutex
	requests []string
}

func (s *emptyNautobot) handler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.requests = append(s.requests, r.Method+" "+r.URL.Path)
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method != http.MethodGet:
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":"`+uuid.New().String()+`","name":"created","display":"created"}`)
	case strings.Contains(r.URL.Path, "extras/statuses"):
		_, _ = io.WriteString(w, `{"count":1,"results":[{"id":"`+uuid.New().String()+
			`","name":"Active","display":"Active","content_types":[]}]}`)
	default:
		_, _ = io.WriteString(w, emptyListJSON)
	}
}

// writes returns every request that was not a GET.
func (s *emptyNautobot) writes() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var writes []string
	for _, request := range s.requests {
		if !strings.HasPrefix(request, http.MethodGet+" ") {
			writes = append(writes, request)
		}
	}
	return writes
}

// TestLoadDryRunIssuesNoWrites verifies a full dry-run export against an empty
// Nautobot, with every auto-create option on, sends nothing but reads.
//
// Why it matters: --dry-run promises no change to Nautobot, yet the lookup
// cache created statuses, roles, locations, location types, tags,
// manufacturers and device types, and patched content types, on its own
// whenever it found them missing.
// Inputs: the spine-leaf fixture with one tagged device; a server that holds
// only a status with no content types; DryRun with every create option
// enabled. Outputs: no write request.
// Data choice: an empty Nautobot makes every lookup miss, so every create path
// the options enable is attempted; the tag reaches the one create the options
// do not gate, and the incomplete status reaches the content-type patch.
func TestLoadDryRunIssuesNoWrites(t *testing.T) {
	// Arrange.
	ensureTypesLoaded(t)
	resetIPAMCaches()
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

	// Act.
	_ = e.Load(inv)

	// Assert.
	if writes := server.writes(); len(writes) != 0 {
		t.Errorf("dry run wrote to Nautobot (%d requests): %v", len(writes), writes)
	}
}
