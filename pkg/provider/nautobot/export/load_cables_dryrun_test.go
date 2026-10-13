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

// cabledNautobot is an emptyNautobot that also holds one device whose eth0
// already carries a cable.
type cabledNautobot struct {
	emptyNautobot
	deviceName string
}

func (s *cabledNautobot) handler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "dcim/devices") && r.URL.Query().Get("name") == s.deviceName:
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"count":1,"results":[{"id":%q,"name":%q,"display":%q}]}`,
			uuid.New().String(), s.deviceName, s.deviceName)
	case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "dcim/interfaces"):
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"count":1,"results":[{"id":%q,"name":"eth0","display":"eth0","cable":{"id":%q}}]}`,
			uuid.New().String(), uuid.New().String())
	default:
		s.emptyNautobot.handler(w, r)
	}
}

// TestDryRunSkipsPlannedCableWhoseRealEndIsCabled verifies that a dry run
// previewing a cable between a planned device and one already in Nautobot
// still applies the real run's "interface already cabled" check to the
// existing end, counting the cable as conflicted rather than created.
//
// Why it matters: the planned-device preview answers before the interface
// lookups, so without this check a dry run promised to create a cable the
// real run would skip, and the two runs disagreed on the plan.
// Inputs: a fake Nautobot holding rt-switch-1 whose eth0 carries a cable; a
// planned rt-node-1 cabled to it; dry run on.
// Outputs: no error, CablesConflicted 1, CablesCreated 0, no writes.
// Data choice: only one end is planned, so the preview path runs and the
// other end has a real interface to check; a cable on it is the conflict the
// real run reports.
func TestDryRunSkipsPlannedCableWhoseRealEndIsCabled(t *testing.T) {
	server := &cabledNautobot{deviceName: "rt-switch-1"}
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	inv := devicetypes.NewInventory()
	node, sw, cableID := uuid.New(), uuid.New(), uuid.New()
	inv.Devices[node] = &devicetypes.CaniDeviceType{ID: node, Name: "rt-node-1"}
	inv.Devices[sw] = &devicetypes.CaniDeviceType{ID: sw, Name: "rt-switch-1"}
	cable := &devicetypes.CaniCableType{
		ID: cableID, Label: "rt-node-1-to-rt-switch-1",
		TerminationADevice: node, TerminationAPort: "eth0", TerminationBDevice: sw, TerminationBPort: "eth0",
	}
	e.Options.DryRun = true
	e.Cache.PlanDevice("rt-node-1")
	result := &LoadResult{}

	err := e.createCaniCableType(context.Background(), cableID, cable, inv, map[string]uuid.UUID{}, result)

	if err != nil {
		t.Fatalf("dry run must not fail on a cabled end, got %v", err)
	}
	if result.CablesConflicted != 1 || result.CablesCreated != 0 {
		t.Fatalf("conflicted=%d created=%d, want 1 and 0", result.CablesConflicted, result.CablesCreated)
	}
	if writes := server.writes(); len(writes) != 0 {
		t.Fatalf("dry run must not write, saw %v", writes)
	}
}

// TestDryRunPreviewsCableBetweenPlannedDevices verifies that a dry run
// previews a cable whose ends are devices the run has only planned, while a
// real run against the same empty Nautobot still reports the missing device.
//
// Why it matters: planned devices have no Nautobot ID, so the cable phase
// used to fail every dry run into an empty Nautobot with exit 1 even though
// everything else was previewed correctly.
// Inputs: an empty fake Nautobot; two devices joined by one cable; the
// device-ID map empty as it is after a dry-run device phase.
// Outputs: with the devices planned under dry run, no error and one cable
// counted as created; without dry run, the "cannot find Nautobot device ID"
// error and no cable counted.
// Data choice: the same cable and inventory in both runs isolate the planned
// check as the only difference.
func TestDryRunPreviewsCableBetweenPlannedDevices(t *testing.T) {
	server := &emptyNautobot{}
	e, cleanup := newExporterWithServer(t, server.handler)
	defer cleanup()
	inv := devicetypes.NewInventory()
	a, b, cableID := uuid.New(), uuid.New(), uuid.New()
	inv.Devices[a] = &devicetypes.CaniDeviceType{ID: a, Name: "rt-node-1"}
	inv.Devices[b] = &devicetypes.CaniDeviceType{ID: b, Name: "rt-node-2"}
	cable := &devicetypes.CaniCableType{
		ID: cableID, Label: "rt-node-1-to-rt-node-2",
		TerminationADevice: a, TerminationAPort: "eth0", TerminationBDevice: b, TerminationBPort: "eth0",
	}

	e.Options.DryRun = true
	e.Cache.PlanDevice("rt-node-1")
	e.Cache.PlanDevice("rt-node-2")
	planned := &LoadResult{}
	if err := e.createCaniCableType(context.Background(), cableID, cable, inv, map[string]uuid.UUID{}, planned); err != nil {
		t.Fatalf("dry run must preview the cable, got %v", err)
	}
	if planned.CablesCreated != 1 {
		t.Fatalf("CablesCreated = %d, want 1", planned.CablesCreated)
	}

	e.Options.DryRun = false
	real := &LoadResult{}
	err := e.createCaniCableType(context.Background(), cableID, cable, inv, map[string]uuid.UUID{}, real)
	if err == nil || !strings.Contains(err.Error(), "cannot find Nautobot device ID for rt-node-1") {
		t.Fatalf("a real run must still report the missing device, got %v", err)
	}
	if real.CablesCreated != 0 || len(server.writes()) != 0 {
		t.Fatalf("nothing may be created or written: created=%d writes=%v", real.CablesCreated, server.writes())
	}
}
