package devicetypes

import (
	"testing"

	"github.com/google/uuid"
)

// TestVerifyPopulatesOrphans verifies that VerifyParentChildRelationships
// populates the Orphans slice for devices and racks with no parent.
func TestVerifyPopulatesOrphans(t *testing.T) {
	inv := NewInventory()

	locID := uuid.New()
	inv.Locations[locID] = &CaniLocationType{
		ID: locID, Name: "site1", LocationType: "site",
	}

	// Parented rack
	rackID := uuid.New()
	inv.Racks[rackID] = &CaniRackType{
		ID: rackID, Name: "rack1", Location: locID, UHeight: 42,
	}

	// Orphan rack
	orphanRackID := uuid.New()
	inv.Racks[orphanRackID] = &CaniRackType{
		ID: orphanRackID, Name: "orphan-rack", UHeight: 42,
	}

	// Parented device
	parentedDevID := uuid.New()
	inv.Devices[parentedDevID] = &CaniDeviceType{
		ID: parentedDevID, Name: "dev-parented", Parent: rackID,
	}

	// Orphan device
	orphanDevID := uuid.New()
	inv.Devices[orphanDevID] = &CaniDeviceType{
		ID: orphanDevID, Name: "orphan-dev", HardwareType: "blade",
	}

	result := inv.VerifyParentChildRelationships()

	if !result.HasOrphans() {
		t.Fatal("expected HasOrphans() to be true")
	}

	// Check orphan counts (1 rack + 1 device = 2).
	deviceOrphans := 0
	rackOrphans := 0
	for _, o := range result.Orphans {
		switch o.Kind {
		case "device":
			deviceOrphans++
		case "rack":
			rackOrphans++
		}
	}
	if deviceOrphans != 1 {
		t.Errorf("expected 1 device orphan, got %d", deviceOrphans)
	}
	if rackOrphans != 1 {
		t.Errorf("expected 1 rack orphan, got %d", rackOrphans)
	}
}

// TestReparentDeviceViaParentField verifies that setting device.Parent and
// calling VerifyParentChildRelationships correctly rebuilds derived fields.
func TestReparentDeviceViaParentField(t *testing.T) {
	inv := NewInventory()

	locID := uuid.New()
	inv.Locations[locID] = &CaniLocationType{
		ID: locID, Name: "site1", LocationType: "site",
	}

	rackID := uuid.New()
	inv.Racks[rackID] = &CaniRackType{
		ID: rackID, Name: "rack1", Location: locID, UHeight: 42,
	}

	// Start as orphan
	devID := uuid.New()
	inv.Devices[devID] = &CaniDeviceType{
		ID: devID, Name: "blade1", HardwareType: "blade",
	}

	// First verify: device is orphan
	result := inv.VerifyParentChildRelationships()
	if !result.HasOrphans() {
		t.Fatal("expected orphans before reparenting")
	}

	// Reparent: set Parent to rack
	inv.Devices[devID].Parent = rackID

	// Second verify: device now linked
	result = inv.VerifyParentChildRelationships()
	if result.HasErrors() {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	dev := inv.Devices[devID]
	if dev.Rack != rackID {
		t.Errorf("expected Rack = %s, got %s", rackID, dev.Rack)
	}
	if dev.Location != locID {
		t.Errorf("expected Location = %s, got %s", locID, dev.Location)
	}

	// Rack should now list the device
	rack := inv.Racks[rackID]
	if !containsUUID(rack.Devices, devID) {
		t.Errorf("expected rack to contain device %s", devID)
	}

	// No more orphan devices
	deviceOrphans := 0
	for _, o := range result.Orphans {
		if o.Kind == "device" {
			deviceOrphans++
		}
	}
	if deviceOrphans != 0 {
		t.Errorf("expected 0 device orphans after reparent, got %d", deviceOrphans)
	}
}

// TestReparentRackViaLocationField verifies that setting rack.Location and
// calling VerifyParentChildRelationships correctly rebuilds the location's
// Racks list.
func TestReparentRackViaLocationField(t *testing.T) {
	inv := NewInventory()

	locID := uuid.New()
	inv.Locations[locID] = &CaniLocationType{
		ID: locID, Name: "site1", LocationType: "site",
	}

	rackID := uuid.New()
	inv.Racks[rackID] = &CaniRackType{
		ID: rackID, Name: "orphan-rack", UHeight: 42,
	}

	// First verify: rack is orphan
	result := inv.VerifyParentChildRelationships()
	rackOrphans := 0
	for _, o := range result.Orphans {
		if o.Kind == "rack" {
			rackOrphans++
		}
	}
	if rackOrphans != 1 {
		t.Fatalf("expected 1 rack orphan, got %d", rackOrphans)
	}

	// Reparent
	inv.Racks[rackID].Location = locID

	result = inv.VerifyParentChildRelationships()
	if result.HasErrors() {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	loc := inv.Locations[locID]
	if !containsUUID(loc.Racks, rackID) {
		t.Errorf("expected location to contain rack %s", rackID)
	}

	// No more orphan racks
	rackOrphans = 0
	for _, o := range result.Orphans {
		if o.Kind == "rack" {
			rackOrphans++
		}
	}
	if rackOrphans != 0 {
		t.Errorf("expected 0 rack orphans after reparent, got %d", rackOrphans)
	}
}

// TestHasOrphansFalseWhenAllParented verifies HasOrphans returns false
// when every item has a parent.
func TestHasOrphansFalseWhenAllParented(t *testing.T) {
	inv := NewInventory()

	locID := uuid.New()
	inv.Locations[locID] = &CaniLocationType{
		ID: locID, Name: "site1", LocationType: "site",
	}

	rackID := uuid.New()
	inv.Racks[rackID] = &CaniRackType{
		ID: rackID, Name: "rack1", Location: locID, UHeight: 42,
	}

	devID := uuid.New()
	inv.Devices[devID] = &CaniDeviceType{
		ID: devID, Name: "dev1", Parent: rackID,
	}

	result := inv.VerifyParentChildRelationships()
	if result.HasOrphans() {
		t.Error("expected HasOrphans() to be false when all items are parented")
	}
}
