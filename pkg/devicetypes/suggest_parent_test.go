package devicetypes

import (
	"testing"

	"github.com/google/uuid"
)

func TestSuggestParentsDeviceToRack(t *testing.T) {
	inv := NewInventory()

	rackID := uuid.New()
	inv.Racks[rackID] = &CaniRackType{
		ID: rackID, Name: "x3000", UHeight: 42,
	}

	orphan := OrphanItem{
		ID:   uuid.New(),
		Name: "x3000c0s1b0",
		Kind: "device",
	}

	suggestions := SuggestParents(inv, orphan, 5)
	if len(suggestions) == 0 {
		t.Fatal("expected at least one suggestion")
	}
	if suggestions[0].ID != rackID {
		t.Errorf("expected rack %s as top suggestion, got %s", rackID, suggestions[0].ID)
	}
	if suggestions[0].Kind != "rack" {
		t.Errorf("expected Kind 'rack', got %q", suggestions[0].Kind)
	}
}

func TestSuggestParentsDevicesOnlySuggestRacks(t *testing.T) {
	inv := NewInventory()

	rackID := uuid.New()
	inv.Racks[rackID] = &CaniRackType{
		ID: rackID, Name: "rack1", UHeight: 42,
	}

	// Chassis device exists but must NOT appear as a suggestion.
	chassisID := uuid.New()
	inv.Devices[chassisID] = &CaniDeviceType{
		ID:           chassisID,
		Name:         "chassis1",
		HardwareType: "Chassis",
		Parent:       rackID,
		Rack:         rackID,
	}

	orphan := OrphanItem{
		ID:           uuid.New(),
		Name:         "blade1",
		Kind:         "device",
		HardwareType: "blade",
	}

	suggestions := SuggestParents(inv, orphan, 10)
	for _, s := range suggestions {
		if s.Kind == "device" {
			t.Errorf("device %q should not be suggested as parent for a device", s.Name)
		}
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion (the rack), got %d", len(suggestions))
	}
	if suggestions[0].ID != rackID {
		t.Errorf("expected rack %s, got %s", rackID, suggestions[0].ID)
	}
}

func TestSuggestParentsRackToLocation(t *testing.T) {
	inv := NewInventory()

	locID := uuid.New()
	inv.Locations[locID] = &CaniLocationType{
		ID: locID, Name: "datacenter1", LocationType: "site",
	}

	orphan := OrphanItem{
		ID:   uuid.New(),
		Name: "rack-orphan",
		Kind: "rack",
	}

	suggestions := SuggestParents(inv, orphan, 5)
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}
	if suggestions[0].ID != locID {
		t.Errorf("expected location %s, got %s", locID, suggestions[0].ID)
	}
	if suggestions[0].Kind != "location" {
		t.Errorf("expected Kind 'location', got %q", suggestions[0].Kind)
	}
}

func TestSuggestParentsNilInventory(t *testing.T) {
	suggestions := SuggestParents(nil, OrphanItem{Kind: "device"}, 5)
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for nil inventory, got %d", len(suggestions))
	}
}

func TestNameSimilarity(t *testing.T) {
	tests := []struct {
		a, b string
		min  int
	}{
		{"x3000c0s1b0", "x3000", 5},
		{"blade1", "blade2", 5},
		{"unrelated", "different", 0},
		{"", "something", 0},
	}
	for _, tt := range tests {
		score := nameSimilarity(tt.a, tt.b)
		if score < tt.min {
			t.Errorf("nameSimilarity(%q, %q) = %d, want >= %d", tt.a, tt.b, score, tt.min)
		}
	}
}

func TestSearchParentCandidates(t *testing.T) {
	inv := NewInventory()

	rackID := uuid.New()
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "alpha-rack"}

	locID := uuid.New()
	inv.Locations[locID] = &CaniLocationType{
		ID: locID, Name: "alpha-site", LocationType: "site",
	}

	// Chassis device exists but must NOT appear in device search.
	chassisID := uuid.New()
	inv.Devices[chassisID] = &CaniDeviceType{
		ID: chassisID, Name: "alpha-chassis", HardwareType: "Chassis",
	}

	// Search for device parents (racks only).
	results := SearchParentCandidates(inv, "alpha", "device", 10)
	if len(results) != 1 {
		t.Fatalf("expected 1 rack result, got %d", len(results))
	}
	if results[0].Name != "alpha-rack" {
		t.Errorf("expected 'alpha-rack', got %q", results[0].Name)
	}
	for _, r := range results {
		if r.Kind == "device" {
			t.Errorf("device %q should not appear in search for device parents", r.Name)
		}
	}

	// Search for rack parents (locations only).
	results = SearchParentCandidates(inv, "alpha", "rack", 10)
	if len(results) != 1 {
		t.Fatalf("expected 1 location result, got %d", len(results))
	}
	if results[0].Name != "alpha-site" {
		t.Errorf("expected 'alpha-site', got %q", results[0].Name)
	}
}

func TestProviderPrefixScore(t *testing.T) {
	meta1 := map[string]any{
		"csm": map[string]any{"xname": "x3000c0s1b0"},
	}
	meta2 := map[string]any{
		"csm": map[string]any{"xname": "x3000c0"},
	}
	score := providerPrefixScore(meta1, meta2)
	if score != 30 {
		t.Errorf("expected 30 for matching xname prefix, got %d", score)
	}

	score = providerPrefixScore(nil, meta2)
	if score != 0 {
		t.Errorf("expected 0 for nil metadata, got %d", score)
	}
}
