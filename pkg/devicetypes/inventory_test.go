package devicetypes

// Test coverage for inventory.go
//
// | Function       | Happy-path test                              | Failure test                                  |
// |----------------|----------------------------------------------|-----------------------------------------------|
// | NewInventory   | TestNewInventoryReturnsInitializedMaps        | TestNewInventoryMapsAreEmpty                  |

import (
	"testing"

	"github.com/google/uuid"
)

// ---------- NewInventory ----------

func TestNewInventoryReturnsInitializedMaps(t *testing.T) {
	inv := NewInventory()
	if inv == nil {
		t.Fatal("expected non-nil inventory")
	}
	if inv.Locations == nil {
		t.Error("expected Locations map to be initialized")
	}
	if inv.Racks == nil {
		t.Error("expected Racks map to be initialized")
	}
	if inv.Devices == nil {
		t.Error("expected Devices map to be initialized")
	}
	if inv.Modules == nil {
		t.Error("expected Modules map to be initialized")
	}
	if inv.Cables == nil {
		t.Error("expected Cables map to be initialized")
	}
	if inv.Frus == nil {
		t.Error("expected Frus map to be initialized")
	}
	if inv.Interfaces == nil {
		t.Error("expected Interfaces map to be initialized")
	}
}

func TestNewInventoryMapsAreEmpty(t *testing.T) {
	inv := NewInventory()

	// A freshly created inventory must have zero entries in every map.
	// Inserting an item and then checking a *different* inventory proves
	// that inventories are independent (no shared state).
	other := NewInventory()
	other.Locations[uuid.New()] = &CaniLocationType{}

	if len(inv.Locations) != 0 {
		t.Errorf("expected 0 locations, got %d", len(inv.Locations))
	}
	if len(inv.Racks) != 0 {
		t.Errorf("expected 0 racks, got %d", len(inv.Racks))
	}
	if len(inv.Devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(inv.Devices))
	}
	if len(inv.Modules) != 0 {
		t.Errorf("expected 0 modules, got %d", len(inv.Modules))
	}
	if len(inv.Cables) != 0 {
		t.Errorf("expected 0 cables, got %d", len(inv.Cables))
	}
	if len(inv.Frus) != 0 {
		t.Errorf("expected 0 frus, got %d", len(inv.Frus))
	}
	if len(inv.Interfaces) != 0 {
		t.Errorf("expected 0 interfaces, got %d", len(inv.Interfaces))
	}
}

// ---------- EnsureUniqueDeviceNames ----------

func TestEnsureUniqueDeviceNamesSuffixes(t *testing.T) {
	tr := &TransformResult{
		Devices: map[uuid.UUID]*CaniDeviceType{
			uuid.New(): {Name: "server"},
			uuid.New(): {Name: "server"},
			uuid.New(): {Name: "server"},
		},
	}

	tr.EnsureUniqueDeviceNames()

	seen := make(map[string]bool)
	for _, d := range tr.Devices {
		if seen[d.Name] {
			t.Errorf("duplicate name after EnsureUniqueDeviceNames: %q", d.Name)
		}
		seen[d.Name] = true
	}
	if len(seen) != 3 {
		t.Errorf("expected 3 unique names, got %d", len(seen))
	}
}

func TestEnsureUniqueDeviceNamesNoDuplicates(t *testing.T) {
	tr := &TransformResult{
		Devices: map[uuid.UUID]*CaniDeviceType{
			uuid.New(): {Name: "alpha"},
			uuid.New(): {Name: "bravo"},
			uuid.New(): {Name: "charlie"},
		},
	}

	tr.EnsureUniqueDeviceNames()

	// Names should be unchanged when all unique.
	names := make(map[string]bool)
	for _, d := range tr.Devices {
		names[d.Name] = true
	}
	for _, want := range []string{"alpha", "bravo", "charlie"} {
		if !names[want] {
			t.Errorf("expected name %q to remain unchanged", want)
		}
	}
}
