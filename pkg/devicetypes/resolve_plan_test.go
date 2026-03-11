package devicetypes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestWriteAndReadPlan(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plan.json")

	plan := &ResolvePlan{
		Assignments: []PlanAssignment{
			{
				OrphanID:   uuid.New(),
				OrphanName: "dev1",
				OrphanKind: "device",
				ParentID:   uuid.New(),
				ParentName: "rack1",
				ParentKind: "rack",
			},
		},
	}

	if err := WritePlan(path, plan); err != nil {
		t.Fatalf("WritePlan: %v", err)
	}

	loaded, err := ReadPlan(path)
	if err != nil {
		t.Fatalf("ReadPlan: %v", err)
	}

	if len(loaded.Assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(loaded.Assignments))
	}
	a := loaded.Assignments[0]
	if a.OrphanName != "dev1" {
		t.Errorf("expected OrphanName 'dev1', got %q", a.OrphanName)
	}
	if a.ParentName != "rack1" {
		t.Errorf("expected ParentName 'rack1', got %q", a.ParentName)
	}
}

func TestReadPlanMissingFile(t *testing.T) {
	_, err := ReadPlan("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestApplyPlanDevice(t *testing.T) {
	inv := NewInventory()

	rackID := uuid.New()
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "rack1", UHeight: 42}

	devID := uuid.New()
	inv.Devices[devID] = &CaniDeviceType{ID: devID, Name: "server1"}

	plan := &ResolvePlan{
		Assignments: []PlanAssignment{
			{
				OrphanID:   devID,
				OrphanName: "server1",
				OrphanKind: "device",
				ParentID:   rackID,
				ParentName: "rack1",
				ParentKind: "rack",
			},
		},
	}

	changes, err := ApplyPlan(inv, plan)
	if err != nil {
		t.Fatalf("ApplyPlan: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if inv.Devices[devID].Parent != rackID {
		t.Errorf("device parent not set")
	}
}

func TestApplyPlanRack(t *testing.T) {
	inv := NewInventory()

	locID := uuid.New()
	inv.Locations[locID] = &CaniLocationType{ID: locID, Name: "site1", LocationType: "site"}

	rackID := uuid.New()
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "orphan-rack", UHeight: 42}

	plan := &ResolvePlan{
		Assignments: []PlanAssignment{
			{
				OrphanID:   rackID,
				OrphanName: "orphan-rack",
				OrphanKind: "rack",
				ParentID:   locID,
				ParentName: "site1",
				ParentKind: "location",
			},
		},
	}

	changes, err := ApplyPlan(inv, plan)
	if err != nil {
		t.Fatalf("ApplyPlan: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if inv.Racks[rackID].Location != locID {
		t.Errorf("rack location not set")
	}
}

func TestApplyPlanMissingDevice(t *testing.T) {
	inv := NewInventory()
	plan := &ResolvePlan{
		Assignments: []PlanAssignment{
			{
				OrphanID:   uuid.New(),
				OrphanName: "ghost",
				OrphanKind: "device",
				ParentID:   uuid.New(),
				ParentName: "rack1",
				ParentKind: "rack",
			},
		},
	}

	_, err := ApplyPlan(inv, plan)
	if err == nil {
		t.Fatal("expected error for missing device")
	}
}

func TestPlanFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roundtrip.json")

	devID := uuid.New()
	rackID := uuid.New()
	locID := uuid.New()

	plan := &ResolvePlan{
		Assignments: []PlanAssignment{
			{
				OrphanID: devID, OrphanName: "dev-a",
				OrphanKind: "device",
				ParentID:   rackID, ParentName: "rack-a",
				ParentKind: "rack",
			},
			{
				OrphanID: uuid.New(), OrphanName: "rack-b",
				OrphanKind: "rack",
				ParentID:   locID, ParentName: "site-1",
				ParentKind: "location",
			},
		},
	}

	if err := WritePlan(path, plan); err != nil {
		t.Fatalf("WritePlan: %v", err)
	}

	// Verify file is valid JSON by reading raw.
	raw, _ := os.ReadFile(path)
	if len(raw) == 0 {
		t.Fatal("plan file is empty")
	}

	loaded, err := ReadPlan(path)
	if err != nil {
		t.Fatalf("ReadPlan: %v", err)
	}

	if len(loaded.Assignments) != 2 {
		t.Fatalf("expected 2 assignments, got %d", len(loaded.Assignments))
	}

	// Verify UUIDs survived round-trip.
	if loaded.Assignments[0].OrphanID != devID {
		t.Errorf("orphan ID mismatch: got %s, want %s", loaded.Assignments[0].OrphanID, devID)
	}
	if loaded.Assignments[0].ParentID != rackID {
		t.Errorf("parent ID mismatch: got %s, want %s", loaded.Assignments[0].ParentID, rackID)
	}
}
