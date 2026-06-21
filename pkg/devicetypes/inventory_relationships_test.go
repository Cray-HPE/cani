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

// +-----------------------------------------------+------------------------------------------------+--------------------------------------------------+
// | Function                                       | Happy-path test                                | Failure test                                     |
// +-----------------------------------------------+------------------------------------------------+--------------------------------------------------+
// | HasErrors                                      | TestHasErrorsReturnsFalseWhenEmpty              | TestHasErrorsReturnsTrueWithErrors               |
// | merge                                          | TestMergeCombinesResults                        | TestMergeWithNilIsNoop                           |
// | logSummary                                     | TestLogSummaryLogsAllEntries                    | TestLogSummaryEmptyResultNoOutput                |
// | rebuildLocationRelationships                    | TestRebuildLocationRelationshipsHappyPath       | TestRebuildLocationRelationshipsMissingParent     |
// | rebuildRackRelationships                        | TestRebuildRackRelationshipsHappyPath           | TestRebuildRackRelationshipsMissingLocation       |
// | clearDeviceReverseLists                         | TestClearDeviceReverseListsResetsFields         | TestClearDeviceReverseListsEmptyInventory         |
// | linkDeviceToRack                                | TestLinkDeviceToRackHappyPath                  | TestLinkDeviceToRackParentNotRack                |
// | linkDeviceToParentDevice                        | TestLinkDeviceToParentDeviceHappyPath           | TestLinkDeviceToParentDeviceNotFound             |
// | rebuildDeviceRelationships                      | TestRebuildDeviceRelationshipsHappyPath         | TestRebuildDeviceRelationshipsOrphanParent        |
// | validateModuleRelationships                     | TestValidateModuleRelationshipsHappyPath        | TestValidateModuleRelationshipsMissingDevice      |
// | validateFruRelationships                        | TestValidateFruRelationshipsHappyPath           | TestValidateFruRelationshipsMissingDevice         |
// | validateCableRelationships                      | TestValidateCableRelationshipsHappyPath         | TestValidateCableRelationshipsMissingDevice       |
// | validateCableEnd                                | TestValidateCableEndHappyPath                  | TestValidateCableEndMissingDeviceAndInterface     |
// | rebuildInterfaceRelationships                   | TestRebuildInterfaceRelationshipsHappyPath      | TestRebuildInterfaceRelationshipsEmptyInventory   |
// | detectCircularLocationRefs                      | TestDetectCircularLocationRefsNoCycle           | TestDetectCircularLocationRefsCycleDetected       |
// | hasLocationCycle                                | TestHasLocationCycleNoCycle                     | TestHasLocationCycleCycleExists                  |
// | detectCircularDeviceRefs                        | TestDetectCircularDeviceRefsNoCycle             | TestDetectCircularDeviceRefsCycleDetected         |
// | hasDeviceCycle                                  | TestHasDeviceCycleNoCycle                       | TestHasDeviceCycleCycleExists                    |
// +-----------------------------------------------+------------------------------------------------+--------------------------------------------------+

package devicetypes

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// HasErrors
// ---------------------------------------------------------------------------

func TestHasErrorsReturnsFalseWhenEmpty(t *testing.T) {
	r := &RelationshipResult{}
	if r.HasErrors() {
		t.Fatal("expected HasErrors to return false for empty result")
	}
}

func TestHasErrorsReturnsTrueWithErrors(t *testing.T) {
	r := &RelationshipResult{
		Errors: []error{
			fmt.Errorf("broken reference"),
		},
	}
	if !r.HasErrors() {
		t.Fatal("expected HasErrors to return true when errors exist")
	}
}

// ---------------------------------------------------------------------------
// merge
// ---------------------------------------------------------------------------

func TestMergeCombinesResults(t *testing.T) {
	a := &RelationshipResult{
		Fixed:    []string{"fix-a"},
		Warnings: []string{"warn-a"},
		Errors:   []error{fmt.Errorf("err-a")},
	}
	b := &RelationshipResult{
		Fixed:    []string{"fix-b"},
		Warnings: []string{"warn-b"},
		Errors:   []error{fmt.Errorf("err-b")},
	}
	a.merge(b)

	if len(a.Fixed) != 2 {
		t.Fatalf("expected 2 fixes, got %d", len(a.Fixed))
	}
	if len(a.Warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d", len(a.Warnings))
	}
	if len(a.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(a.Errors))
	}
}

func TestMergeWithNilIsNoop(t *testing.T) {
	a := &RelationshipResult{
		Fixed: []string{"fix-a"},
	}
	a.merge(nil)

	if len(a.Fixed) != 1 {
		t.Fatalf("expected 1 fix after nil merge, got %d", len(a.Fixed))
	}
}

// ---------------------------------------------------------------------------
// logSummary
// ---------------------------------------------------------------------------

func TestLogSummaryLogsAllEntries(t *testing.T) {
	r := &RelationshipResult{
		Fixed:    []string{"fixed-item"},
		Warnings: []string{"warning-item"},
		Errors:   []error{fmt.Errorf("error-item")},
	}
	// logSummary writes to the default logger; verify no panic.
	r.logSummary()
}

func TestLogSummaryEmptyResultNoOutput(t *testing.T) {
	r := &RelationshipResult{}
	// Should not panic on empty slices.
	r.logSummary()
}

// ---------------------------------------------------------------------------
// rebuildLocationRelationships
// ---------------------------------------------------------------------------

func TestRebuildLocationRelationshipsHappyPath(t *testing.T) {
	parentID := uuid.New()
	childID := uuid.New()

	inv := NewInventory()
	inv.Locations[parentID] = &CaniLocationType{Name: "parent"}
	inv.Locations[childID] = &CaniLocationType{Name: "child", Parent: parentID}

	res := inv.rebuildLocationRelationships()
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if len(inv.Locations[parentID].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(inv.Locations[parentID].Children))
	}
	if inv.Locations[parentID].Children[0] != childID {
		t.Fatal("child ID mismatch")
	}
}

func TestRebuildLocationRelationshipsMissingParent(t *testing.T) {
	childID := uuid.New()
	missingParent := uuid.New()

	inv := NewInventory()
	inv.Locations[childID] = &CaniLocationType{Name: "orphan", Parent: missingParent}

	res := inv.rebuildLocationRelationships()
	if !res.HasErrors() {
		t.Fatal("expected errors for missing parent location")
	}
}

// ---------------------------------------------------------------------------
// rebuildRackRelationships
// ---------------------------------------------------------------------------

func TestRebuildRackRelationshipsHappyPath(t *testing.T) {
	locID := uuid.New()
	rackID := uuid.New()

	inv := NewInventory()
	inv.Locations[locID] = &CaniLocationType{Name: "datacenter"}
	inv.Racks[rackID] = &CaniRackType{Name: "rack-1", Location: locID}

	res := inv.rebuildRackRelationships()
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if len(inv.Locations[locID].Racks) != 1 {
		t.Fatalf("expected 1 rack, got %d", len(inv.Locations[locID].Racks))
	}
}

func TestRebuildRackRelationshipsMissingLocation(t *testing.T) {
	rackID := uuid.New()
	missingLoc := uuid.New()

	inv := NewInventory()
	inv.Racks[rackID] = &CaniRackType{Name: "rack-lost", Location: missingLoc}

	res := inv.rebuildRackRelationships()
	if !res.HasErrors() {
		t.Fatal("expected errors for missing location")
	}
}

// ---------------------------------------------------------------------------
// clearDeviceReverseLists
// ---------------------------------------------------------------------------

func TestClearDeviceReverseListsResetsFields(t *testing.T) {
	rackID := uuid.New()
	deviceID := uuid.New()
	locID := uuid.New()
	parentDevID := uuid.New()

	inv := NewInventory()
	inv.Racks[rackID] = &CaniRackType{
		Name:    "r1",
		Devices: []uuid.UUID{deviceID},
	}
	inv.Devices[deviceID] = &CaniDeviceType{
		Name:         "d1",
		Children:     []uuid.UUID{uuid.New()},
		Rack:         rackID,
		Location:     locID,
		ParentDevice: parentDevID,
	}

	inv.clearDeviceReverseLists()

	if len(inv.Racks[rackID].Devices) != 0 {
		t.Fatal("expected rack devices to be cleared")
	}
	dev := inv.Devices[deviceID]
	if len(dev.Children) != 0 || dev.Rack != uuid.Nil ||
		dev.Location != uuid.Nil || dev.ParentDevice != uuid.Nil {
		t.Fatal("expected all device reverse fields to be cleared")
	}
}

func TestClearDeviceReverseListsEmptyInventory(t *testing.T) {
	inv := NewInventory()
	// Should not panic on empty maps.
	inv.clearDeviceReverseLists()
}

// ---------------------------------------------------------------------------
// linkDeviceToRack
// ---------------------------------------------------------------------------

func TestLinkDeviceToRackHappyPath(t *testing.T) {
	locID := uuid.New()
	rackID := uuid.New()
	deviceID := uuid.New()

	inv := NewInventory()
	inv.Racks[rackID] = &CaniRackType{Name: "r1", Location: locID}
	device := &CaniDeviceType{Name: "d1", Parent: rackID}

	linked, msg := inv.linkDeviceToRack(deviceID, device)
	if !linked {
		t.Fatal("expected device to be linked to rack")
	}
	if msg == "" {
		t.Fatal("expected non-empty fix message")
	}
	if device.Rack != rackID {
		t.Fatal("device.Rack not set")
	}
	if device.Location != locID {
		t.Fatal("device.Location not set")
	}
}

func TestLinkDeviceToRackParentNotRack(t *testing.T) {
	deviceID := uuid.New()

	inv := NewInventory()
	device := &CaniDeviceType{Name: "d1", Parent: uuid.New()}

	linked, msg := inv.linkDeviceToRack(deviceID, device)
	if linked {
		t.Fatal("expected device NOT to be linked when parent is not a rack")
	}
	if msg != "" {
		t.Fatalf("expected empty message, got %q", msg)
	}
}

// ---------------------------------------------------------------------------
// linkDeviceToParentDevice
// ---------------------------------------------------------------------------

func TestLinkDeviceToParentDeviceHappyPath(t *testing.T) {
	rackID := uuid.New()
	locID := uuid.New()
	parentDevID := uuid.New()
	childDevID := uuid.New()

	inv := NewInventory()
	inv.Devices[parentDevID] = &CaniDeviceType{
		Name:     "parent-dev",
		Rack:     rackID,
		Location: locID,
	}

	child := &CaniDeviceType{Name: "child-dev", Parent: parentDevID}

	linked, msg := inv.linkDeviceToParentDevice(childDevID, child)
	if !linked {
		t.Fatal("expected device to be linked to parent device")
	}
	if msg == "" {
		t.Fatal("expected non-empty fix message")
	}
	if child.ParentDevice != parentDevID {
		t.Fatal("child.ParentDevice not set")
	}
	if child.Rack != rackID || child.Location != locID {
		t.Fatal("child did not inherit rack/location from parent")
	}
}

func TestLinkDeviceToParentDeviceNotFound(t *testing.T) {
	childDevID := uuid.New()

	inv := NewInventory()
	child := &CaniDeviceType{Name: "child-dev", Parent: uuid.New()}

	linked, msg := inv.linkDeviceToParentDevice(childDevID, child)
	if linked {
		t.Fatal("expected link to fail when parent device missing")
	}
	if msg != "" {
		t.Fatalf("expected empty message, got %q", msg)
	}
}

// ---------------------------------------------------------------------------
// rebuildDeviceRelationships
// ---------------------------------------------------------------------------

func TestRebuildDeviceRelationshipsHappyPath(t *testing.T) {
	locID := uuid.New()
	rackID := uuid.New()
	deviceID := uuid.New()

	inv := NewInventory()
	inv.Locations[locID] = &CaniLocationType{Name: "loc"}
	inv.Racks[rackID] = &CaniRackType{Name: "rack", Location: locID}
	inv.Devices[deviceID] = &CaniDeviceType{Name: "dev", Parent: rackID}

	res := inv.rebuildDeviceRelationships()
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if inv.Devices[deviceID].Rack != rackID {
		t.Fatal("device rack not set after rebuild")
	}
}

func TestRebuildDeviceRelationshipsOrphanParent(t *testing.T) {
	deviceID := uuid.New()
	missingParent := uuid.New()

	inv := NewInventory()
	inv.Devices[deviceID] = &CaniDeviceType{Name: "dev", Parent: missingParent}

	res := inv.rebuildDeviceRelationships()
	if !res.HasErrors() {
		t.Fatal("expected errors for device with missing parent")
	}
}

// ---------------------------------------------------------------------------
// validateModuleRelationships
// ---------------------------------------------------------------------------

func TestValidateModuleRelationshipsHappyPath(t *testing.T) {
	deviceID := uuid.New()
	moduleID := uuid.New()

	inv := NewInventory()
	inv.Devices[deviceID] = &CaniDeviceType{Name: "dev"}
	inv.Modules[moduleID] = &CaniModuleType{Name: "mod", ParentDevice: deviceID}

	res := inv.validateModuleRelationships()
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
}

func TestValidateModuleRelationshipsMissingDevice(t *testing.T) {
	moduleID := uuid.New()
	missingDev := uuid.New()

	inv := NewInventory()
	inv.Modules[moduleID] = &CaniModuleType{Name: "mod", ParentDevice: missingDev}

	res := inv.validateModuleRelationships()
	if !res.HasErrors() {
		t.Fatal("expected errors for module with missing parent device")
	}
}

// ---------------------------------------------------------------------------
// rebuildFruRelationships
// ---------------------------------------------------------------------------

func TestRebuildFruRelationshipsHappyPath(t *testing.T) {
	deviceID := uuid.New()
	fruID := uuid.New()

	inv := NewInventory()
	inv.Devices[deviceID] = &CaniDeviceType{ID: deviceID, Name: "dev"}
	inv.Frus[fruID] = &CaniFruType{Name: "fru", Device: deviceID}

	res := inv.rebuildFruRelationships()
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if len(inv.Devices[deviceID].Frus) != 1 {
		t.Fatalf("expected 1 fru on device, got %d", len(inv.Devices[deviceID].Frus))
	}
	if inv.Devices[deviceID].Frus[0] != fruID {
		t.Fatalf("expected fru %s, got %s", fruID, inv.Devices[deviceID].Frus[0])
	}
}

func TestRebuildFruRelationshipsMissingDevice(t *testing.T) {
	fruID := uuid.New()
	missingDev := uuid.New()

	inv := NewInventory()
	inv.Frus[fruID] = &CaniFruType{Name: "fru", Device: missingDev}

	res := inv.rebuildFruRelationships()
	if !res.HasErrors() {
		t.Fatal("expected errors for fru with missing device")
	}
}

func TestRebuildFruRelationshipsModuleParent(t *testing.T) {
	modID := uuid.New()
	fruID := uuid.New()

	inv := NewInventory()
	inv.Modules[modID] = &CaniModuleType{ID: modID, Name: "mod"}
	inv.Frus[fruID] = &CaniFruType{Name: "mod-fru", Device: modID, Parent: modID}

	res := inv.rebuildFruRelationships()
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if len(inv.Modules[modID].Frus) != 1 {
		t.Fatalf("expected 1 fru on module, got %d", len(inv.Modules[modID].Frus))
	}
}

// ---------------------------------------------------------------------------
// validateCableRelationships
// ---------------------------------------------------------------------------

func TestValidateCableRelationshipsHappyPath(t *testing.T) {
	devA := uuid.New()
	devB := uuid.New()
	ifaceA := uuid.New()
	ifaceB := uuid.New()
	cableID := uuid.New()

	inv := NewInventory()
	inv.Devices[devA] = &CaniDeviceType{
		Name: "dev-a",
		Interfaces: []InterfaceSpec{
			{ID: ifaceA, Name: "eth0", Type: InterfacesElemTypeA1000BaseT},
		},
	}
	inv.Devices[devB] = &CaniDeviceType{
		Name: "dev-b",
		Interfaces: []InterfaceSpec{
			{ID: ifaceB, Name: "eth0", Type: InterfacesElemTypeA1000BaseT},
		},
	}
	// Pre-populate interface index so GetInterfaceByID works.
	inv.Interfaces[ifaceA] = &CaniInterface{ID: ifaceA, DeviceID: devA}
	inv.Interfaces[ifaceB] = &CaniInterface{ID: ifaceB, DeviceID: devB}

	inv.Cables[cableID] = &CaniCableType{
		Label:              "cable-1",
		TerminationADevice: devA,
		TerminationA:       ifaceA,
		TerminationBDevice: devB,
		TerminationB:       ifaceB,
	}

	res := inv.validateCableRelationships()
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
}

func TestValidateCableRelationshipsMissingDevice(t *testing.T) {
	cableID := uuid.New()
	missingDev := uuid.New()

	inv := NewInventory()
	inv.Cables[cableID] = &CaniCableType{
		Label:              "cable-bad",
		TerminationADevice: missingDev,
	}

	res := inv.validateCableRelationships()
	if !res.HasErrors() {
		t.Fatal("expected errors for cable with missing termination device")
	}
}

// ---------------------------------------------------------------------------
// validateCableEnd
// ---------------------------------------------------------------------------

func TestValidateCableEndHappyPath(t *testing.T) {
	devID := uuid.New()
	ifaceID := uuid.New()
	cableID := uuid.New()

	inv := NewInventory()
	inv.Devices[devID] = &CaniDeviceType{
		Name: "dev",
		Interfaces: []InterfaceSpec{
			{ID: ifaceID, Name: "eth0", Type: InterfacesElemTypeA1000BaseT},
		},
	}
	inv.Interfaces[ifaceID] = &CaniInterface{ID: ifaceID, DeviceID: devID}
	cable := &CaniCableType{Label: "c1"}

	res := inv.validateCableEnd(cableID, cable, "A", devID, ifaceID)
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
}

func TestValidateCableEndMissingDeviceAndInterface(t *testing.T) {
	cableID := uuid.New()
	missingDev := uuid.New()
	missingIface := uuid.New()
	cable := &CaniCableType{Label: "c-bad"}

	inv := NewInventory()

	res := inv.validateCableEnd(cableID, cable, "B", missingDev, missingIface)
	if !res.HasErrors() {
		t.Fatal("expected errors for missing device and interface")
	}
	if len(res.Errors) != 2 {
		t.Fatalf("expected 2 errors (device + interface), got %d", len(res.Errors))
	}
}

// ---------------------------------------------------------------------------
// rebuildInterfaceRelationships
// ---------------------------------------------------------------------------

func TestRebuildInterfaceRelationshipsHappyPath(t *testing.T) {
	deviceID := uuid.New()
	ifaceID := uuid.New()

	inv := NewInventory()
	inv.Devices[deviceID] = &CaniDeviceType{
		Name: "dev",
		Interfaces: []InterfaceSpec{
			{ID: ifaceID, Name: "eth0", Type: InterfacesElemTypeA1000BaseT},
		},
	}

	res := inv.rebuildInterfaceRelationships()
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if len(inv.Interfaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(inv.Interfaces))
	}
	inst, ok := inv.Interfaces[ifaceID]
	if !ok {
		t.Fatal("expected interface to be indexed")
	}
	if inst.DeviceID != deviceID {
		t.Fatal("interface DeviceID mismatch")
	}
}

func TestRebuildInterfaceRelationshipsEmptyInventory(t *testing.T) {
	inv := NewInventory()

	res := inv.rebuildInterfaceRelationships()
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if len(inv.Interfaces) != 0 {
		t.Fatalf("expected 0 interfaces, got %d", len(inv.Interfaces))
	}
}

// ---------------------------------------------------------------------------
// detectCircularLocationRefs
// ---------------------------------------------------------------------------

func TestDetectCircularLocationRefsNoCycle(t *testing.T) {
	rootID := uuid.New()
	childID := uuid.New()

	inv := NewInventory()
	inv.Locations[rootID] = &CaniLocationType{Name: "root"}
	inv.Locations[childID] = &CaniLocationType{Name: "child", Parent: rootID}

	res := inv.detectCircularLocationRefs()
	if res.HasErrors() {
		t.Fatalf("unexpected circular reference detected: %v", res.Errors)
	}
}

func TestDetectCircularLocationRefsCycleDetected(t *testing.T) {
	aID := uuid.New()
	bID := uuid.New()

	inv := NewInventory()
	inv.Locations[aID] = &CaniLocationType{Name: "a", Parent: bID}
	inv.Locations[bID] = &CaniLocationType{Name: "b", Parent: aID}

	res := inv.detectCircularLocationRefs()
	if !res.HasErrors() {
		t.Fatal("expected circular reference error")
	}
}

// ---------------------------------------------------------------------------
// hasLocationCycle
// ---------------------------------------------------------------------------

func TestHasLocationCycleNoCycle(t *testing.T) {
	rootID := uuid.New()
	childID := uuid.New()

	inv := NewInventory()
	inv.Locations[rootID] = &CaniLocationType{Name: "root"}
	inv.Locations[childID] = &CaniLocationType{Name: "child", Parent: rootID}

	if inv.hasLocationCycle(childID) {
		t.Fatal("expected no cycle for linear chain")
	}
}

func TestHasLocationCycleCycleExists(t *testing.T) {
	aID := uuid.New()
	bID := uuid.New()
	cID := uuid.New()

	inv := NewInventory()
	inv.Locations[aID] = &CaniLocationType{Name: "a", Parent: bID}
	inv.Locations[bID] = &CaniLocationType{Name: "b", Parent: cID}
	inv.Locations[cID] = &CaniLocationType{Name: "c", Parent: aID}

	if !inv.hasLocationCycle(aID) {
		t.Fatal("expected cycle to be detected in 3-node loop")
	}
}

// ---------------------------------------------------------------------------
// detectCircularDeviceRefs
// ---------------------------------------------------------------------------

func TestDetectCircularDeviceRefsNoCycle(t *testing.T) {
	rackID := uuid.New()
	deviceID := uuid.New()

	inv := NewInventory()
	inv.Racks[rackID] = &CaniRackType{Name: "rack"}
	inv.Devices[deviceID] = &CaniDeviceType{Name: "dev", Parent: rackID}

	res := inv.detectCircularDeviceRefs()
	if res.HasErrors() {
		t.Fatalf("unexpected circular reference detected: %v", res.Errors)
	}
}

func TestDetectCircularDeviceRefsCycleDetected(t *testing.T) {
	aID := uuid.New()
	bID := uuid.New()

	inv := NewInventory()
	inv.Devices[aID] = &CaniDeviceType{Name: "a", Parent: bID}
	inv.Devices[bID] = &CaniDeviceType{Name: "b", Parent: aID}

	res := inv.detectCircularDeviceRefs()
	if !res.HasErrors() {
		t.Fatal("expected circular reference error")
	}
}

// ---------------------------------------------------------------------------
// hasDeviceCycle
// ---------------------------------------------------------------------------

func TestHasDeviceCycleNoCycle(t *testing.T) {
	rackID := uuid.New()
	deviceID := uuid.New()

	inv := NewInventory()
	inv.Racks[rackID] = &CaniRackType{Name: "rack"}
	inv.Devices[deviceID] = &CaniDeviceType{Name: "dev", Parent: rackID}

	if inv.hasDeviceCycle(deviceID) {
		t.Fatal("expected no cycle when parent is a rack")
	}
}

func TestHasDeviceCycleCycleExists(t *testing.T) {
	aID := uuid.New()
	bID := uuid.New()
	cID := uuid.New()

	inv := NewInventory()
	inv.Devices[aID] = &CaniDeviceType{Name: "a", Parent: bID}
	inv.Devices[bID] = &CaniDeviceType{Name: "b", Parent: cID}
	inv.Devices[cID] = &CaniDeviceType{Name: "c", Parent: aID}

	if !inv.hasDeviceCycle(aID) {
		t.Fatal("expected cycle to be detected in 3-node device loop")
	}
}

// ---------------------------------------------------------------------------
// VerifyParentChildRelationships – idempotency
// ---------------------------------------------------------------------------

func TestVerifyIdempotentSameResultTwice(t *testing.T) {
	locID := uuid.New()
	rackID := uuid.New()
	chassisID := uuid.New()
	bladeID := uuid.New()

	inv := NewInventory()
	inv.Locations[locID] = &CaniLocationType{ID: locID, Name: "site"}
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "rack", Location: locID}
	inv.Devices[chassisID] = &CaniDeviceType{Name: "chassis", Parent: rackID}
	inv.Devices[bladeID] = &CaniDeviceType{Name: "blade", Parent: chassisID}

	r1 := inv.VerifyParentChildRelationships()
	if r1.HasErrors() {
		t.Fatalf("first call: unexpected errors: %v", r1.Errors)
	}

	// Snapshot derived fields after first call.
	bladeRack1 := inv.Devices[bladeID].Rack
	bladeLoc1 := inv.Devices[bladeID].Location
	bladeParent1 := inv.Devices[bladeID].ParentDevice

	r2 := inv.VerifyParentChildRelationships()
	if r2.HasErrors() {
		t.Fatalf("second call: unexpected errors: %v", r2.Errors)
	}
	if len(r2.Fixed) != 0 {
		t.Fatalf("second call: expected 0 fixes, got %d: %v", len(r2.Fixed), r2.Fixed)
	}

	if inv.Devices[bladeID].Rack != bladeRack1 {
		t.Fatalf("idempotency: blade Rack changed: %s → %s",
			bladeRack1, inv.Devices[bladeID].Rack)
	}
	if inv.Devices[bladeID].Location != bladeLoc1 {
		t.Fatalf("idempotency: blade Location changed: %s → %s",
			bladeLoc1, inv.Devices[bladeID].Location)
	}
	if inv.Devices[bladeID].ParentDevice != bladeParent1 {
		t.Fatalf("idempotency: blade ParentDevice changed: %s → %s",
			bladeParent1, inv.Devices[bladeID].ParentDevice)
	}
}

func TestVerifyIdempotentNestedDevices(t *testing.T) {
	locID := uuid.New()
	rackID := uuid.New()
	chassisID := uuid.New()
	bladeID := uuid.New()
	gpuID := uuid.New()

	inv := NewInventory()
	inv.Locations[locID] = &CaniLocationType{ID: locID, Name: "site"}
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "rack", Location: locID}
	inv.Devices[chassisID] = &CaniDeviceType{Name: "chassis", Parent: rackID}
	inv.Devices[bladeID] = &CaniDeviceType{Name: "blade", Parent: chassisID}
	inv.Devices[gpuID] = &CaniDeviceType{Name: "gpu", Parent: bladeID}

	res := inv.VerifyParentChildRelationships()
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}

	// All three levels must have inherited the rack and location.
	if inv.Devices[chassisID].Rack != rackID {
		t.Fatal("chassis rack not set")
	}
	if inv.Devices[bladeID].Rack != rackID {
		t.Fatal("blade rack not set")
	}
	if inv.Devices[gpuID].Rack != rackID {
		t.Fatal("gpu rack not set")
	}
	if inv.Devices[gpuID].Location != locID {
		t.Fatal("gpu location not set")
	}

	// Second call must produce identical derived fields and zero fixes.
	r2 := inv.VerifyParentChildRelationships()
	if len(r2.Fixed) != 0 {
		t.Fatalf("second call: expected 0 fixes, got %d: %v", len(r2.Fixed), r2.Fixed)
	}
	if inv.Devices[gpuID].Rack != rackID {
		t.Fatal("gpu rack not set after second call")
	}
	if inv.Devices[gpuID].Location != locID {
		t.Fatal("gpu location not set after second call")
	}
}

func TestVerifyIdempotentNoDuplicateChildren(t *testing.T) {
	locID := uuid.New()
	rackID := uuid.New()
	chassisID := uuid.New()

	inv := NewInventory()
	inv.Locations[locID] = &CaniLocationType{ID: locID, Name: "site"}
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "rack", Location: locID}
	inv.Devices[chassisID] = &CaniDeviceType{Name: "chassis", Parent: rackID}

	inv.VerifyParentChildRelationships()
	inv.VerifyParentChildRelationships()
	inv.VerifyParentChildRelationships()

	if len(inv.Locations[locID].Children) != 0 {
		t.Fatalf("expected 0 child locations, got %d", len(inv.Locations[locID].Children))
	}
	if len(inv.Locations[locID].Racks) != 1 {
		t.Fatalf("expected 1 rack, got %d", len(inv.Locations[locID].Racks))
	}
	if len(inv.Racks[rackID].Devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(inv.Racks[rackID].Devices))
	}
}

// TestFindInterfaceIDByPort verifies FindInterfaceIDByPort resolves a port name
// to its interface UUID across a device's own interfaces, an attached module's
// interfaces, and a directly addressed module.
//
// Why it matters: cable termination resolution looks up ports by name, so the
// function must search a device's interfaces, then fall back to its modules,
// and also accept a module ID directly — returning uuid.Nil when nothing
// matches.
// Inputs: a device with port "eth0", a module (parented to that device) with
// port "mgmt0", and queries for "eth0", "mgmt0", a direct module-ID lookup of
// "mgmt0", and a missing port. Outputs: the matching interface UUIDs, and
// uuid.Nil for the miss and for an unknown device.
// Data choice: distinct port names on the device vs. the module prove the
// device-first / module-fallback ordering, and the direct module-ID query
// exercises the early module branch that a device-only test would skip.
func TestFindInterfaceIDByPort(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	modID := uuid.New()
	ethID := uuid.New()
	mgmtID := uuid.New()

	inv.Devices[devID] = &CaniDeviceType{
		ID:   devID,
		Name: "dev",
		Interfaces: []InterfaceSpec{
			{ID: ethID, Name: "eth0", Type: InterfacesElemTypeA1000BaseT},
		},
	}
	inv.Modules[modID] = &CaniModuleType{
		ID:           modID,
		Name:         "nic",
		ParentDevice: devID,
		Interfaces: []InterfaceSpec{
			{ID: mgmtID, Name: "mgmt0", Type: InterfacesElemTypeA1000BaseT},
		},
	}

	if got := inv.FindInterfaceIDByPort(devID, "eth0"); got != ethID {
		t.Errorf("device port lookup = %v, want %v", got, ethID)
	}
	if got := inv.FindInterfaceIDByPort(devID, "mgmt0"); got != mgmtID {
		t.Errorf("module-fallback lookup = %v, want %v", got, mgmtID)
	}
	if got := inv.FindInterfaceIDByPort(modID, "mgmt0"); got != mgmtID {
		t.Errorf("direct module lookup = %v, want %v", got, mgmtID)
	}
	if got := inv.FindInterfaceIDByPort(devID, "nope"); got != uuid.Nil {
		t.Errorf("missing port = %v, want uuid.Nil", got)
	}
	if got := inv.FindInterfaceIDByPort(uuid.New(), "eth0"); got != uuid.Nil {
		t.Errorf("unknown device = %v, want uuid.Nil", got)
	}
}

// TestLogSummaryDebugLogsFixedAndWarnings verifies logSummary emits Fixed and
// Warning entries only when Debug is enabled, while Errors are always logged.
//
// Why it matters: relationship rebuilds can produce large fix/warning lists;
// gating them behind Debug keeps normal runs quiet, and a regression that
// always or never logs would either spam operators or hide diagnostics.
// Inputs: a result carrying one Fixed, one Warning, and one Error, invoked once
// with Debug=true. Outputs: no panic and full traversal of the Debug-only
// loops. Data choice: enabling Debug here is the only way to execute the Fixed
// and Warnings branches that the existing Debug=false test cannot reach.
func TestLogSummaryDebugLogsFixedAndWarnings(t *testing.T) {
	orig := Debug
	Debug = true
	t.Cleanup(func() { Debug = orig })

	r := &RelationshipResult{
		Fixed:    []string{"fixed-item"},
		Warnings: []string{"warning-item"},
		Errors:   []error{fmt.Errorf("error-item")},
	}
	r.logSummary()
}

// TestValidateModuleRelationshipsNilAndNoParent verifies validateModuleRelationships
// skips nil module entries and warns (without erroring) on a module that has no
// parent device assigned.
//
// Why it matters: import data frequently contains modules not yet attached to a
// device; these must surface as warnings rather than hard errors, and nil map
// slots must never panic the validator.
// Inputs: an inventory with a nil module entry and a module whose ParentDevice
// is uuid.Nil. Outputs: no errors and exactly one warning. Data choice: pairing
// the nil entry with an unparented module covers both the continue guard and the
// warning branch that the happy-path and missing-device tests skip.
func TestValidateModuleRelationshipsNilAndNoParent(t *testing.T) {
	inv := NewInventory()
	inv.Modules[uuid.New()] = nil
	inv.Modules[uuid.New()] = &CaniModuleType{Name: "floating"}

	res := inv.validateModuleRelationships()
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if len(res.Warnings) != 1 {
		t.Fatalf("expected 1 warning for unparented module, got %d: %v", len(res.Warnings), res.Warnings)
	}
}

// TestRebuildCableRelationships verifies rebuildCableRelationships removes cables
// whose endpoint devices were deleted and resolves termination interface UUIDs
// from device + port names.
//
// Why it matters: after devices are removed or re-imported, cable terminations
// must be re-resolved or pruned so the topology stays consistent; stale cables
// would otherwise reference non-existent endpoints.
// Inputs: an inventory with a device that has an "eth0" interface plus a cable
// whose A-side names that device/port and whose B-side references a deleted
// device. Outputs: the stale cable removed and, for a second valid cable, the
// A-side TerminationA UUID resolved to the interface ID. Data choice: combining
// a delete case and a resolve case in one inventory exercises both the prune
// branch and the port-resolution branch of the loop.
func TestRebuildCableRelationships(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	ethID := uuid.New()
	inv.Devices[devID] = &CaniDeviceType{
		ID:   devID,
		Name: "dev",
		Interfaces: []InterfaceSpec{
			{ID: ethID, Name: "eth0", Type: InterfacesElemTypeA1000BaseT},
		},
	}

	staleID := uuid.New()
	inv.Cables[staleID] = &CaniCableType{
		Label:              "stale",
		TerminationADevice: devID,
		TerminationBDevice: uuid.New(), // deleted device -> cable pruned
	}

	resolveID := uuid.New()
	inv.Cables[resolveID] = &CaniCableType{
		Label:              "resolve",
		TerminationADevice: devID,
		TerminationAPort:   "eth0",
	}

	res := inv.rebuildCableRelationships()

	if _, ok := inv.Cables[staleID]; ok {
		t.Error("expected stale cable to be removed")
	}
	if inv.Cables[resolveID].TerminationA != ethID {
		t.Errorf("TerminationA = %v, want resolved %v", inv.Cables[resolveID].TerminationA, ethID)
	}
	if len(res.Fixed) == 0 {
		t.Error("expected Fixed entries for prune + resolve")
	}
}

// TestRebuildCableRelationshipsBSideAndNoChange verifies rebuildCableRelationships
// resolves the B-side termination and records no fix when a termination is
// already correct.
//
// Why it matters: the B-side resolution branch mirrors the A-side and must work
// independently, while an already-resolved cable should produce no spurious
// "fixed" noise on repeated rebuilds (idempotency).
// Inputs: an inventory with a device owning "eth1"; one cable whose B-side names
// that device/port (needs resolving) and whose A-side TerminationA is already
// set to the correct interface UUID. Outputs: TerminationB resolved to the
// interface ID and no Fixed entry for the already-correct A-side. Data choice:
// pre-setting TerminationA to the right value exercises the equality guard that
// skips the redundant fix, isolating it from the B-side resolution.
func TestRebuildCableRelationshipsBSideAndNoChange(t *testing.T) {
	inv := NewInventory()
	devID := uuid.New()
	eth1ID := uuid.New()
	inv.Devices[devID] = &CaniDeviceType{
		ID:   devID,
		Name: "dev",
		Interfaces: []InterfaceSpec{
			{ID: eth1ID, Name: "eth1", Type: InterfacesElemTypeA1000BaseT},
		},
	}

	cableID := uuid.New()
	inv.Cables[cableID] = &CaniCableType{
		Label:              "b-resolve",
		TerminationADevice: devID,
		TerminationAPort:   "eth1",
		TerminationA:       eth1ID, // already correct -> no fix
		TerminationBDevice: devID,
		TerminationBPort:   "eth1", // needs resolving
	}

	res := inv.rebuildCableRelationships()

	if inv.Cables[cableID].TerminationB != eth1ID {
		t.Errorf("TerminationB = %v, want resolved %v", inv.Cables[cableID].TerminationB, eth1ID)
	}
	for _, f := range res.Fixed {
		if strings.Contains(f, "termination A") {
			t.Errorf("unexpected fix for already-correct A-side: %q", f)
		}
	}
}

// TestUUIDSetsEqual verifies uuidSetsEqual reports equality only when both sets
// hold exactly the same UUIDs.
//
// Why it matters: relationship rebuilds compare expected vs. actual child sets
// to decide whether a change occurred; a faulty comparison would either report
// phantom changes or silently drop real ones.
// Inputs: two identical sets, a shorter set, and a same-length set with one
// differing element. Outputs: true for the identical pair and false for the
// length-mismatch and element-mismatch cases. Data choice: separating the
// length and membership differences exercises both early-return branches of the
// helper.
func TestUUIDSetsEqual(t *testing.T) {
	a := uuid.New()
	b := uuid.New()
	set := map[uuid.UUID]bool{a: true, b: true}

	if !uuidSetsEqual(set, map[uuid.UUID]bool{a: true, b: true}) {
		t.Error("identical sets should be equal")
	}
	if uuidSetsEqual(set, map[uuid.UUID]bool{a: true}) {
		t.Error("different-length sets should not be equal")
	}
	if uuidSetsEqual(set, map[uuid.UUID]bool{a: true, uuid.New(): true}) {
		t.Error("same-length sets with a different element should not be equal")
	}
}
