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
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

type bayMoveFixture struct {
	exporter  *Exporter
	inventory *devicetypes.Inventory
	ids       map[string]uuid.UUID
	server    *bayMoveServer
	parent    *devicetypes.CaniDeviceType
	child     *devicetypes.CaniDeviceType
}

func newBayMoveFixture(t *testing.T) *bayMoveFixture {
	t.Helper()
	inv, ids, childID := bayExportInventory()
	f := &bayMoveFixture{
		inventory: inv, ids: ids,
		server: &bayMoveServer{devices: ids, bays: make(map[uuid.UUID]remoteDeviceBay)},
	}
	for _, device := range inv.Devices {
		device.Type = "switch"
		device.ExternalIDs = map[string]uuid.UUID{externalIDKeyNautobot: ids[device.Name]}
		if device.ParentBay == "" {
			f.parent = device
			device.Type = "chassis"
		} else {
			f.child = device
		}
	}
	f.addBay(ids[f.parent.Name], "Bay 1", childID)
	var cleanup func()
	f.exporter, cleanup = newExporterWithServer(t, f.server.handle)
	t.Cleanup(cleanup)
	f.exporter.Options.Merge = true
	f.exporter.Cache.SetContext(t.Context())
	return f
}

func (f *bayMoveFixture) addBay(parent uuid.UUID, name string, installed uuid.UUID) {
	id := uuid.New()
	f.server.bays[id] = remoteDeviceBay{id: id, parent: parent, name: name, installed: installed}
}

func (f *bayMoveFixture) addDevice(name, bay string) *devicetypes.CaniDeviceType {
	id := uuid.New()
	device := &devicetypes.CaniDeviceType{
		ID: uuid.New(), Name: name, Type: "switch", Parent: f.parent.ID, ParentBay: bay,
	}
	device.ExternalIDs = map[string]uuid.UUID{externalIDKeyNautobot: id}
	f.inventory.Devices[device.ID], f.ids[name] = device, id
	return device
}

func (f *bayMoveFixture) run(t *testing.T) *LoadResult {
	t.Helper()
	result := &LoadResult{}
	if err := f.exporter.prepareDeviceMoves(t.Context(), f.inventory, result); err != nil {
		t.Fatalf("prepareDeviceMoves: %v", err)
	}
	f.exporter.loadDeviceBays(t.Context(), f.inventory, f.ids, result)
	if len(result.Errors) != 0 {
		t.Fatalf("loadDeviceBays: %v", result.Errors)
	}
	return result
}

func (f *bayMoveFixture) assertAssignments(t *testing.T) {
	t.Helper()
	bays, _ := f.server.snapshot()
	actual := make(map[uuid.UUID]deviceBayTarget)
	for _, bay := range bays {
		if bay.installed != uuid.Nil {
			actual[bay.installed] = deviceBayTarget{parent: bay.parent, name: bay.name}
		}
	}
	want, err := desiredDeviceBays(f.inventory, f.ids)
	if err != nil {
		t.Fatal(err)
	}
	for id, target := range want {
		if actual[id] != target {
			t.Errorf("device %s: assignment=%+v, want %+v", id, actual[id], target)
		}
	}
}

func TestReconcileDeviceBaysLifecycle(t *testing.T) {
	for _, tc := range []struct {
		name      string
		setup     func(*bayMoveFixture)
		cleared   int
		installed int
	}{
		{"move between bays", func(f *bayMoveFixture) {
			f.child.ParentBay = "Bay 2"
			f.addBay(f.ids[f.parent.Name], "Bay 2", uuid.Nil)
		}, 1, 1},
		{"move between parents", func(f *bayMoveFixture) {
			parent := f.addDevice("other-chassis", "")
			parent.Parent = uuid.Nil
			f.child.Parent = parent.ID
			f.addBay(f.ids[parent.Name], "Bay 1", uuid.Nil)
		}, 1, 1},
		{"detach", func(f *bayMoveFixture) {
			f.child.Parent, f.child.ParentBay = uuid.Nil, ""
		}, 1, 0},
		{"replace managed occupant", func(f *bayMoveFixture) {
			f.child.Parent, f.child.ParentBay = uuid.Nil, ""
			f.addDevice("replacement", "Bay 1")
		}, 1, 1},
		{"swap", func(f *bayMoveFixture) {
			f.child.ParentBay = "Bay 2"
			other := f.addDevice("other-child", "Bay 1")
			f.addBay(f.ids[f.parent.Name], "Bay 2", f.ids[other.Name])
		}, 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newBayMoveFixture(t)
			tc.setup(f)
			f.assertLifecycle(t, tc.cleared, tc.installed)
		})
	}
}

func (f *bayMoveFixture) assertLifecycle(t *testing.T, cleared, installed int) {
	t.Helper()
	result := f.run(t)
	if result.DeviceBaysCleared != cleared || result.DeviceBaysInstalled != installed {
		t.Fatalf("counts=%+v; want %d clears, %d installs", result.DeviceBayCounts, cleared, installed)
	}
	f.assertAssignments(t)
	_, requests := f.server.snapshot()
	var writes []capturedRequest
	for _, request := range requests {
		if request.method == http.MethodPatch {
			writes = append(writes, request)
		}
	}
	for _, request := range writes[:cleared] {
		if string(request.body) != `{"installed_device":null}` {
			t.Errorf("install before all clears: %s", request.body)
		}
	}
}

func TestReconcileDeviceBaysIdempotent(t *testing.T) {
	f := newBayMoveFixture(t)
	f.child.ParentBay = "Bay 2"
	f.addBay(f.ids[f.parent.Name], "Bay 2", uuid.Nil)
	f.run(t)
	_, before := f.server.snapshot()
	result := f.run(t)
	_, after := f.server.snapshot()
	if result.DeviceBaysCleared != 0 || result.DeviceBaysInstalled != 0 || result.DeviceBaysSkipped != 1 {
		t.Fatalf("second export counts=%+v", result.DeviceBayCounts)
	}
	if countBayWrites(after) != countBayWrites(before) {
		t.Fatal("idempotent export issued new writes")
	}
}

func TestReconcileDeviceBaysDryRun(t *testing.T) {
	f := newBayMoveFixture(t)
	f.child.ParentBay = "Bay 2"
	f.addBay(f.ids[f.parent.Name], "Bay 2", uuid.Nil)
	f.exporter.Options.DryRun = true
	before, _ := f.server.snapshot()
	result := f.run(t)
	after, requests := f.server.snapshot()
	if !slices.Equal(before, after) || countBayWrites(requests) != 0 {
		t.Fatalf("dry run changed remote state: requests=%v", requests)
	}
	if result.DeviceBaysCleared != 1 || result.DeviceBaysInstalled != 1 {
		t.Fatalf("dry-run counts=%+v", result.DeviceBayCounts)
	}
}

func TestReconcileDeviceBaysProtectsUnmanagedOccupants(t *testing.T) {
	f := newBayMoveFixture(t)
	f.child.ParentBay = "Bay 2"
	f.addBay(f.ids[f.parent.Name], "Bay 2", uuid.New())
	before, _ := f.server.snapshot()
	err := f.exporter.prepareDeviceMoves(t.Context(), f.inventory, &LoadResult{})
	after, requests := f.server.snapshot()
	if err == nil || !strings.Contains(err.Error(), "which is not moving") {
		t.Fatalf("expected unmanaged conflict, got %v", err)
	}
	if !slices.Equal(before, after) || countBayWrites(requests) != 0 {
		t.Fatal("destination conflict cleared a source bay")
	}
}

func TestPrepareDeviceMovesSkipsUnresolvedAndRemovedDevices(t *testing.T) {
	for _, removed := range []bool{false, true} {
		t.Run(fmt.Sprintf("removed=%v", removed), func(t *testing.T) {
			f := newBayMoveFixture(t)
			if removed {
				delete(f.inventory.Devices, f.child.ID)
			} else {
				f.child.ExternalIDs = nil
				delete(f.server.devices, f.child.Name)
			}
			assertBayPreparationPreservesState(t, f)
		})
	}
}

func assertBayPreparationPreservesState(t *testing.T, f *bayMoveFixture) {
	t.Helper()
	before, _ := f.server.snapshot()
	if err := f.exporter.prepareDeviceMoves(t.Context(), f.inventory, &LoadResult{}); err != nil {
		t.Fatal(err)
	}
	after, requests := f.server.snapshot()
	if !slices.Equal(before, after) || countBayWrites(requests) != 0 {
		t.Fatal("unresolved or removed device was detached")
	}
}

func TestPrepareDeviceMovesRequiresMerge(t *testing.T) {
	f := newBayMoveFixture(t)
	f.exporter.Options.Merge = false
	f.child.ParentBay = ""
	if err := f.exporter.prepareDeviceMoves(t.Context(), f.inventory, &LoadResult{}); err != nil {
		t.Fatal(err)
	}
	_, requests := f.server.snapshot()
	if len(requests) != 0 {
		t.Fatalf("non-merge preparation made requests: %v", requests)
	}
}

func TestPrepareDeviceMovesResolvesManagedDeviceByName(t *testing.T) {
	f := newBayMoveFixture(t)
	f.child.ExternalIDs = nil
	f.child.Parent, f.child.ParentBay = uuid.Nil, ""
	result := f.run(t)
	if result.DeviceBaysCleared != 1 {
		t.Fatalf("name-resolved device was not detached: %+v", result.DeviceBayCounts)
	}
	f.assertAssignments(t)
}

func TestPrepareDeviceMovesDiscoveryFailures(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*bayMoveServer)
		want  string
	}{
		{"source status", func(s *bayMoveServer) { s.listStatus = http.StatusForbidden }, "status 403"},
		{"destination status", func(s *bayMoveServer) { s.destinationStatus = http.StatusInternalServerError }, "status 500"},
		{"invalid JSON", func(s *bayMoveServer) { s.invalidJSON = true }, "device bay listing"},
		{"invalid bay", func(s *bayMoveServer) { s.invalidBay = true }, "invalid bay"},
		{"empty next page", func(s *bayMoveServer) { s.emptyNext = true }, "empty page"},
		{"identity status", func(s *bayMoveServer) { s.deviceStatus = http.StatusInternalServerError }, "checking device"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newBayMoveFixture(t)
			f.child.ParentBay = "Bay 2"
			f.addBay(f.ids[f.parent.Name], "Bay 2", uuid.Nil)
			tc.setup(f.server)
			err := f.exporter.prepareDeviceMoves(t.Context(), f.inventory, &LoadResult{})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want %q", err, tc.want)
			}
			_, requests := f.server.snapshot()
			if countBayWrites(requests) != 0 {
				t.Fatal("discovery failure caused a write")
			}
		})
	}
}

func TestPrepareDeviceMovesClearFailure(t *testing.T) {
	f := newBayMoveFixture(t)
	f.child.ParentBay = ""
	f.server.patchStatus = http.StatusInternalServerError
	result := &LoadResult{}
	err := f.exporter.prepareDeviceMoves(t.Context(), f.inventory, result)
	if err == nil || !strings.Contains(err.Error(), "clear device bay") || !strings.Contains(err.Error(), "status 500") {
		t.Fatalf("unexpected clear error: %v", err)
	}
	if result.DeviceBaysCleared != 0 {
		t.Fatalf("failed clear counted as success: %+v", result.DeviceBayCounts)
	}
}

func TestPrepareDeviceMovesRetriesPartialClear(t *testing.T) {
	f := newBayMoveFixture(t)
	f.child.ParentBay = "Bay 2"
	other := f.addDevice("other-child", "Bay 1")
	f.addBay(f.ids[f.parent.Name], "Bay 2", f.ids[other.Name])
	f.server.patchStatus, f.server.patchFailAt = http.StatusInternalServerError, 2
	result := &LoadResult{}
	err := f.exporter.prepareDeviceMoves(t.Context(), f.inventory, result)
	if err == nil || result.DeviceBaysCleared != 1 {
		t.Fatalf("partial clear: error=%v, counts=%+v", err, result.DeviceBayCounts)
	}
	_, requests := f.server.snapshot()
	for _, request := range requests {
		if request.method == http.MethodPatch && string(request.body) != `{"installed_device":null}` {
			t.Fatalf("partial clear failure attempted an install: %s", request.body)
		}
	}
	f.server.mu.Lock()
	f.server.patchStatus = 0
	f.server.mu.Unlock()
	retry := f.run(t)
	if retry.DeviceBaysCleared != 1 || retry.DeviceBaysInstalled != 2 {
		t.Fatalf("retry counts=%+v", retry.DeviceBayCounts)
	}
	f.assertAssignments(t)
}

func TestReconcileDeviceBaysRetriesFailedInstallation(t *testing.T) {
	f := newBayMoveFixture(t)
	f.child.ParentBay = "Bay 2"
	f.addBay(f.ids[f.parent.Name], "Bay 2", uuid.Nil)
	f.server.patchStatus, f.server.patchFailAt = http.StatusInternalServerError, 2
	emptyBays, _ := f.server.snapshot()
	for i := range emptyBays {
		emptyBays[i].installed = uuid.Nil
	}
	result := &LoadResult{}
	if err := f.exporter.prepareDeviceMoves(t.Context(), f.inventory, result); err != nil {
		t.Fatalf("prepareDeviceMoves: %v", err)
	}

	f.exporter.loadDeviceBays(t.Context(), f.inventory, f.ids, result)

	assertDeviceBayErrorContext(t, result, f.child.Name, `"Bay 2"`, f.parent.Name, "device bay update", "status 500")
	if result.DeviceBayCounts != (DeviceBayCounts{DeviceBaysCleared: 1}) {
		t.Fatalf("failed installation counts=%+v, want only one successful clear", result.DeviceBayCounts)
	}
	afterFailure, requests := f.server.snapshot()
	if !slices.Equal(afterFailure, emptyBays) || countBayWrites(requests) != 2 {
		t.Fatalf("failed install: bays=%+v requests=%v, want empty bays after a clear and failed install", afterFailure, requests)
	}

	retry := f.run(t)
	if retry.DeviceBayCounts != (DeviceBayCounts{DeviceBaysInstalled: 1}) {
		t.Fatalf("retry counts=%+v, want one install and no repeated clear", retry.DeviceBayCounts)
	}
	f.assertAssignments(t)
	afterRetry, beforeRepeat := f.server.snapshot()

	repeat := f.run(t)
	if repeat.DeviceBayCounts != (DeviceBayCounts{DeviceBaysSkipped: 1}) {
		t.Fatalf("repeat counts=%+v, want one already-installed skip", repeat.DeviceBayCounts)
	}
	afterRepeat, afterRequests := f.server.snapshot()
	if !slices.Equal(afterRepeat, afterRetry) || countBayWrites(afterRequests) != countBayWrites(beforeRepeat) {
		t.Fatal("idempotent export changed remote assignments or issued new writes")
	}
}

func TestManagedDeviceBaysPaginationAndBatching(t *testing.T) {
	for _, count := range []int{2, deviceBayPageSize + 1} {
		t.Run(fmt.Sprintf("devices=%d", count), func(t *testing.T) {
			f := newBayMoveFixture(t)
			f.server.bays = make(map[uuid.UUID]remoteDeviceBay)
			targets := make(map[uuid.UUID]deviceBayTarget)
			for i := range count {
				id := uuid.New()
				targets[id] = deviceBayTarget{}
				f.addBay(f.ids[f.parent.Name], fmt.Sprintf("Bay %d", i), id)
			}
			f.server.pageSize = 1
			assertManagedBayCount(t, f.exporter, targets)
		})
	}
}

func assertManagedBayCount(t *testing.T, exporter *Exporter, targets map[uuid.UUID]deviceBayTarget) {
	t.Helper()
	bays, err := exporter.managedDeviceBays(t.Context(), targets)
	if err != nil || len(bays) != len(targets) {
		t.Fatalf("bays=%d, error=%v; want %d", len(bays), err, len(targets))
	}
	found := make(map[uuid.UUID]bool)
	for _, bay := range bays {
		found[bay.installed] = true
	}
	if len(found) != len(targets) {
		t.Fatalf("pagination duplicated or skipped devices: got %d, want %d", len(found), len(targets))
	}
}

func TestLoadClearsBayBeforeDeviceUpdate(t *testing.T) {
	f := newBayMoveFixture(t)
	seedDeviceRefs(t, f.exporter)
	f.exporter.Options.DefaultLocation = "DC1"
	f.exporter.Options.DefaultRole = "Compute"
	f.exporter.Options.DefaultStatus = "Active"
	for _, device := range f.inventory.Devices {
		device.Slug = "hpe-dl380"
		device.Status = "Active"
		device.Role = "Compute"
	}
	rack := &devicetypes.CaniRackType{ID: uuid.New(), Name: "rack-01"}
	f.inventory.Racks[rack.ID] = rack
	rackID := uuid.New()
	f.server.racks = map[string]uuid.UUID{rack.Name: rackID}
	f.child.Parent, f.child.ParentBay = rack.ID, ""
	f.child.RackPosition, f.child.Face = 1, "front"
	if err := f.exporter.Load(f.inventory); err != nil {
		t.Fatalf("Load: %v", err)
	}
	f.assertAssignments(t)
	_, requests := f.server.snapshot()
	assertBayClearBeforeDeviceUpdate(t, requests, f.ids[f.child.Name], rackID)
}

func TestLoadStopsAfterBayPreparationFailure(t *testing.T) {
	f := newBayMoveFixture(t)
	f.child.ParentBay = ""
	f.server.patchStatus = http.StatusInternalServerError
	err := f.exporter.Load(f.inventory)
	if err == nil || !strings.Contains(err.Error(), "device bay reconciliation") {
		t.Fatalf("Load error=%v, want bay preparation failure", err)
	}
	_, requests := f.server.snapshot()
	for _, request := range requests {
		if request.method != http.MethodGet && !strings.Contains(request.path, "device-bays") {
			t.Fatalf("write after preparation failure: %s %s", request.method, request.path)
		}
	}
}

func assertBayClearBeforeDeviceUpdate(t *testing.T, requests []capturedRequest, childID, rackID uuid.UUID) {
	t.Helper()
	clearIndex, updateIndex := -1, -1
	for i, request := range requests {
		if request.method != http.MethodPatch {
			continue
		}
		switch {
		case string(request.body) == `{"installed_device":null}`:
			clearIndex = i
		case strings.Contains(request.path, "/devices/"+childID.String()+"/"):
			updateIndex = i
			assertDevicePatchRack(t, request.body, rackID)
		}
	}
	if clearIndex < 0 || updateIndex <= clearIndex {
		t.Fatalf("bay clear index=%d, device update index=%d", clearIndex, updateIndex)
	}
}

func assertDevicePatchRack(t *testing.T, body []byte, rackID uuid.UUID) {
	t.Helper()
	var patch map[string]json.RawMessage
	if err := json.Unmarshal(body, &patch); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(patch["rack"]), rackID.String()) {
		t.Fatalf("device update did not assign target rack: %s", body)
	}
}

func countBayWrites(requests []capturedRequest) int {
	count := 0
	for _, request := range requests {
		if request.method != http.MethodGet {
			count++
		}
	}
	return count
}
