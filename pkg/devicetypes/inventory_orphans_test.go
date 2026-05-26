package devicetypes

import (
	"testing"

	"github.com/google/uuid"
)

func TestOrphanDevices(t *testing.T) {
	rackID := uuid.New()
	inv := NewInventory()
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "rack1"}

	parentedID := uuid.New()
	inv.Devices[parentedID] = &CaniDeviceType{
		ID: parentedID, Name: "dev-parented", Parent: rackID,
	}

	orphanID1 := uuid.New()
	inv.Devices[orphanID1] = &CaniDeviceType{
		ID: orphanID1, Name: "blade1",
	}

	orphanID2 := uuid.New()
	inv.Devices[orphanID2] = &CaniDeviceType{
		ID: orphanID2, Name: "blade2", Type: "blade",
	}

	orphans := inv.OrphanDevices()

	if len(orphans) != 2 {
		t.Fatalf("expected 2 orphan devices, got %d", len(orphans))
	}
	// Sorted by name
	if orphans[0].Name != "blade1" {
		t.Errorf("expected first orphan 'blade1', got %q", orphans[0].Name)
	}
	if orphans[1].Name != "blade2" {
		t.Errorf("expected second orphan 'blade2', got %q", orphans[1].Name)
	}
	if orphans[1].DeviceType != "blade" {
		t.Errorf("expected HardwareType 'blade', got %q", orphans[1].DeviceType)
	}
	for _, o := range orphans {
		if o.Kind != "device" {
			t.Errorf("expected Kind 'device', got %q", o.Kind)
		}
	}
}

func TestOrphanRacks(t *testing.T) {
	locID := uuid.New()
	inv := NewInventory()
	inv.Locations[locID] = &CaniLocationType{
		ID: locID, Name: "site1", LocationType: "site",
	}

	parentedRack := uuid.New()
	inv.Racks[parentedRack] = &CaniRackType{
		ID: parentedRack, Name: "rack-parented", Location: locID,
	}

	orphanRack := uuid.New()
	inv.Racks[orphanRack] = &CaniRackType{
		ID: orphanRack, Name: "rack-orphan",
	}

	orphans := inv.OrphanRacks()

	if len(orphans) != 1 {
		t.Fatalf("expected 1 orphan rack, got %d", len(orphans))
	}
	if orphans[0].Name != "rack-orphan" {
		t.Errorf("expected 'rack-orphan', got %q", orphans[0].Name)
	}
	if orphans[0].Kind != "rack" {
		t.Errorf("expected Kind 'rack', got %q", orphans[0].Kind)
	}
}

func TestOrphanDevicesEmpty(t *testing.T) {
	inv := NewInventory()
	orphans := inv.OrphanDevices()
	if len(orphans) != 0 {
		t.Errorf("expected 0 orphans in empty inventory, got %d", len(orphans))
	}
}

func TestOrphanRacksEmpty(t *testing.T) {
	inv := NewInventory()
	orphans := inv.OrphanRacks()
	if len(orphans) != 0 {
		t.Errorf("expected 0 orphans in empty inventory, got %d", len(orphans))
	}
}
