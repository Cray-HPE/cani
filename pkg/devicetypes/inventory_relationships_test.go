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
