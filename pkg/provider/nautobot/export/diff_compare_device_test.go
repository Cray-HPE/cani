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
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// ---------- compareDeviceFields ----------

// TestCompareDeviceFieldsSerial verifies a differing serial number produces a
// "serial" FieldDiff carrying the local and remote values.
//
// Why it matters: --merge updates rely on this diff to know which fields to
// PATCH in Nautobot; a missed serial diff would leave stale serials remotely.
// Inputs: local device serial "NEW-SN" vs remote "OLD-SN".
// Outputs: a diff list containing {serial, NEW-SN, OLD-SN}.
// Data choice: distinct new/old serials and otherwise-empty objects isolate the
// serial comparison branch.
func TestCompareDeviceFieldsSerial(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Serial: "NEW-SN",
	}
	oldSerial := "OLD-SN"
	remote := &nautobotapi.Device{
		Serial: &oldSerial,
	}

	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	diffs := compareDeviceFields(dev, remote, mapper)

	found := false
	for _, d := range diffs {
		if d.Field == "serial" && d.LocalVal == "NEW-SN" && d.RemoteVal == "OLD-SN" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected serial diff, got: %+v", diffs)
	}
}

// TestCompareDeviceFieldsAssetTag verifies a differing asset tag produces an
// "asset_tag" FieldDiff carrying the local and remote values.
//
// Why it matters: asset tags are tracked for inventory/audit; export's merge
// path must detect and propagate changes so Nautobot stays in sync.
// Inputs: local asset tag "NEW-TAG" vs remote "OLD-TAG".
// Outputs: a diff list containing {asset_tag, NEW-TAG, OLD-TAG}.
// Data choice: distinct new/old tags on otherwise-empty objects isolate the
// asset-tag branch, mirroring the serial test.
func TestCompareDeviceFieldsAssetTag(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		AssetTag: "NEW-TAG",
	}
	oldTag := "OLD-TAG"
	remote := &nautobotapi.Device{
		AssetTag: &oldTag,
	}

	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	diffs := compareDeviceFields(dev, remote, mapper)

	found := false
	for _, d := range diffs {
		if d.Field == "asset_tag" && d.LocalVal == "NEW-TAG" && d.RemoteVal == "OLD-TAG" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected asset_tag diff, got: %+v", diffs)
	}
}

// TestCompareDeviceFieldsNoChanges verifies that two empty device objects
// produce no field diffs.
//
// Why it matters: export should skip (not update) devices that match Nautobot,
// avoiding needless PATCHes; any spurious diff would cause churn on every run.
// Inputs: an empty local CaniDeviceType and an empty remote Device.
// Outputs: an empty diff slice.
// Data choice: empty-vs-empty exercises the all-fields-skipped path, since each
// comparison branch is guarded by a non-empty/non-nil check.
func TestCompareDeviceFieldsNoChanges(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{}
	remote := &nautobotapi.Device{}

	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	diffs := compareDeviceFields(dev, remote, mapper)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs for empty device, got: %+v", diffs)
	}
}

// ---------- compareDeviceFields with position/face diffs ----------

// TestCompareDeviceFieldsPositionFace verifies that compareDeviceFields reports
// both a "position" and a "face" diff when the local rack position and face
// differ from the remote Nautobot device.
//
// Why it matters: rack position and orientation must be reconciled so a moved
// or reoriented device is PATCHed; missing either diff would leave Nautobot out
// of sync with the physical layout.
// Inputs: local RackPosition=10/Face="rear" vs remote Position=5/Face="front".
// Outputs: a diff slice containing both "position" and "face" entries.
// Data choice: deliberately mismatched values on both axes guarantee both
// comparison branches fire.
func TestCompareDeviceFieldsPositionFace(t *testing.T) {
	pos := 5
	dev := &devicetypes.CaniDeviceType{
		RackPosition: 10,
		Face:         "rear",
	}
	remote := &nautobotapi.Device{
		Position: &pos,
	}
	setNBValue(&remote.Face, "front")

	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	diffs := compareDeviceFields(dev, remote, mapper)

	posFound := false
	faceFound := false
	for _, d := range diffs {
		if d.Field == "position" {
			posFound = true
		}
		if d.Field == "face" {
			faceFound = true
		}
	}
	if !posFound {
		t.Error("expected position diff")
	}
	if !faceFound {
		t.Error("expected face diff")
	}
}

// ---------- compareDeviceFields with type/location/status/role diffs ----------

// TestCompareDeviceFieldsTypeDiff verifies that when the locally resolved
// device type has a different UUID than the remote device's DeviceType ref,
// compareDeviceFields emits a "device_type" diff whose LocalVal is the cached
// local type name ("LocalType").
//
// Why it matters: the export update path relies on these diffs to detect when a
// device's hardware type drifted from Nautobot and must be patched; a wrong
// field name or value would push bad updates or mask real drift.
// Inputs: a CaniDeviceType (Slug "local-type"), a remote Device with a
// different device-type ref, and a seeded mapper. Outputs: a []FieldDiff that
// includes the device_type entry.
// Data choice: distinct local/remote device-type UUIDs force the mismatch while
// location/status/role are left matching so only device_type diverges.
func TestCompareDeviceFieldsTypeDiff(t *testing.T) {
	cache := NewLookupCache(nil)
	localDtID := uuid.New()
	remoteDtID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["local-type"] = &CachedItem{ID: localDtID, Name: "LocalType"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["Site"] = &CachedItem{ID: uuid.New(), Name: "Site"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: uuid.New(), Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: uuid.New(), Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Site",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{Slug: "local-type"}

	// Build remote with a different device type ID
	remote := &nautobotapi.Device{}
	mustSetRefID(t, &remote.DeviceType, remoteDtID)
	mustSetRefID(t, &remote.Location, uuid.New())
	mustSetRefID(t, &remote.Status, uuid.New())
	mustSetRefID(t, &remote.Role, uuid.New())

	diffs := compareDeviceFields(dev, remote, mapper)
	found := false
	for _, d := range diffs {
		if d.Field == "device_type" {
			found = true
			if d.LocalVal != "LocalType" {
				t.Errorf("localVal = %q, want 'LocalType'", d.LocalVal)
			}
		}
	}
	if !found {
		t.Error("expected device_type diff")
	}
}

// TestCompareDeviceFieldsStatusDiff verifies that compareDeviceFields emits a
// "status" diff (LocalVal "Planned") when the locally resolved status UUID
// differs from the remote device's Status reference.
//
// Why it matters: lifecycle status (Planned/Active/...) is exported to Nautobot,
// so the diff engine must flag status drift for the update path to reconcile it.
// Inputs: a CaniDeviceType with Status "Planned", a remote Device whose Status
// ref is a different UUID, and a mapper seeded with that status/location/role.
// Outputs: a []FieldDiff containing the status entry.
// Data choice: only the status UUID is made to differ; location and role caches
// match their defaults so the status field is the sole asserted divergence.
func TestCompareDeviceFieldsStatusDiff(t *testing.T) {
	cache := NewLookupCache(nil)
	localStatusID := uuid.New()
	remoteStatusID := uuid.New()

	cache.statusesMu.Lock()
	cache.statuses["Planned"] = &CachedItem{ID: localStatusID, Name: "Planned"}
	cache.statusesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC"] = &CachedItem{ID: uuid.New(), Name: "DC"}
	cache.locationsMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: uuid.New(), Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC",
		DefaultStatus:   "Planned",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		ObjectMeta: devicetypes.ObjectMeta{Status: "Planned"},
	}

	remote := &nautobotapi.Device{}
	mustSetRefID(t, &remote.DeviceType, uuid.New())
	mustSetRefID(t, &remote.Location, uuid.New())
	mustSetRefID(t, &remote.Status, remoteStatusID)
	mustSetRefID(t, &remote.Role, uuid.New())

	diffs := compareDeviceFields(dev, remote, mapper)
	found := false
	for _, d := range diffs {
		if d.Field == "status" {
			found = true
			if d.LocalVal != "Planned" {
				t.Errorf("localVal = %q, want 'Planned'", d.LocalVal)
			}
		}
	}
	if !found {
		t.Error("expected status diff")
	}
}

// TestCompareDeviceFieldsRoleDiff verifies that compareDeviceFields emits a
// "role" diff (LocalVal "Compute") when the locally resolved role UUID differs
// from the remote device's Role reference.
//
// Why it matters: device role drives Nautobot's device classification, so the
// exporter must detect and report role drift before patching the remote device.
// Inputs: an empty-slug CaniDeviceType, a remote Device whose Role ref is a
// fresh UUID, and a mapper whose DefaultRole "Compute" is cached. Outputs: a
// []FieldDiff containing the role entry.
// Data choice: location/status caches match their defaults while the remote role
// gets an unrelated UUID, isolating role as the only mismatch.
func TestCompareDeviceFieldsRoleDiff(t *testing.T) {
	cache := NewLookupCache(nil)
	localRoleID := uuid.New()

	cache.rolesMu.Lock()
	cache.roles["Compute"] = &CachedItem{ID: localRoleID, Name: "Compute"}
	cache.rolesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC"] = &CachedItem{ID: uuid.New(), Name: "DC"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: uuid.New(), Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC",
		DefaultStatus:   "Active",
		DefaultRole:     "Compute",
	})

	dev := &devicetypes.CaniDeviceType{}

	remote := &nautobotapi.Device{}
	mustSetRefID(t, &remote.DeviceType, uuid.New())
	mustSetRefID(t, &remote.Location, uuid.New())
	mustSetRefID(t, &remote.Status, uuid.New())
	mustSetRefID(t, &remote.Role, uuid.New()) // different from localRoleID

	diffs := compareDeviceFields(dev, remote, mapper)
	found := false
	for _, d := range diffs {
		if d.Field == "role" {
			found = true
			if d.LocalVal != "Compute" {
				t.Errorf("localVal = %q, want 'Compute'", d.LocalVal)
			}
		}
	}
	if !found {
		t.Error("expected role diff")
	}
}

// TestCompareDeviceFieldsLocationDiff verifies that compareDeviceFields emits a
// "location" diff (LocalVal "Room-Z") when the locally resolved location UUID
// differs from the remote device's Location reference.
//
// Why it matters: placing devices at the correct Nautobot location is core to
// the export, so drift in location assignment must be surfaced for reconciliation.
// Inputs: an empty CaniDeviceType, a remote Device whose Location ref is a fresh
// UUID, and a mapper with DefaultLocation "Room-Z" cached. Outputs: a []FieldDiff
// containing the location entry.
// Data choice: status/role caches match their defaults while the remote location
// UUID is unrelated, so location is the only field that diverges.
func TestCompareDeviceFieldsLocationDiff(t *testing.T) {
	cache := NewLookupCache(nil)
	localLocID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["Room-Z"] = &CachedItem{ID: localLocID, Name: "Room-Z"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: uuid.New(), Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: uuid.New(), Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "Room-Z",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{}

	remote := &nautobotapi.Device{}
	mustSetRefID(t, &remote.DeviceType, uuid.New())
	mustSetRefID(t, &remote.Location, uuid.New()) // different
	mustSetRefID(t, &remote.Status, uuid.New())
	mustSetRefID(t, &remote.Role, uuid.New())

	diffs := compareDeviceFields(dev, remote, mapper)
	found := false
	for _, d := range diffs {
		if d.Field == "location" {
			found = true
			if d.LocalVal != "Room-Z" {
				t.Errorf("localVal = %q, want 'Room-Z'", d.LocalVal)
			}
		}
	}
	if !found {
		t.Error("expected location diff")
	}
}

// TestCompareDeviceFieldsNoTypeDiffWhenMatching verifies that compareDeviceFields
// returns an empty diff slice when every locally resolved reference UUID equals
// the corresponding remote reference.
//
// Why it matters: the export update path must NOT report changes when local and
// remote already agree, otherwise it would issue needless PATCH churn to Nautobot.
// Inputs: a CaniDeviceType (Slug "shared"), a remote Device whose four refs all
// use one shared UUID, and a mapper whose caches all resolve to that same UUID.
// Outputs: an empty []FieldDiff.
// Data choice: reusing a single sharedID for every cache entry and remote ref
// makes all four ID comparisons equal, exercising the no-diff path precisely.
func TestCompareDeviceFieldsNoTypeDiffWhenMatching(t *testing.T) {
	cache := NewLookupCache(nil)
	sharedID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["shared"] = &CachedItem{ID: sharedID, Name: "Shared"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC"] = &CachedItem{ID: sharedID, Name: "DC"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: sharedID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: sharedID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{Slug: "shared"}

	remote := &nautobotapi.Device{}
	mustSetRefID(t, &remote.DeviceType, sharedID)
	mustSetRefID(t, &remote.Location, sharedID)
	mustSetRefID(t, &remote.Status, sharedID)
	mustSetRefID(t, &remote.Role, sharedID)

	diffs := compareDeviceFields(dev, remote, mapper)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs when all IDs match, got %+v", diffs)
	}
}
