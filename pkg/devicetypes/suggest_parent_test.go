package devicetypes

import (
	"strings"
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
		ID:     chassisID,
		Name:   "chassis1",
		Type:   "Chassis",
		Parent: rackID,
		Rack:   rackID,
	}

	orphan := OrphanItem{
		ID:         uuid.New(),
		Name:       "blade1",
		Kind:       "device",
		DeviceType: "blade",
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
		ID: chassisID, Name: "alpha-chassis", Type: "Chassis",
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

	// Sub-maps present on both sides but values share no 3-char prefix: the
	// scorer must iterate every provider/key and fall through to 0.
	noShare1 := map[string]any{"csm": map[string]any{"xname": "aaa111"}}
	noShare2 := map[string]any{"csm": map[string]any{"xname": "zzz999"}}
	if score := providerPrefixScore(noShare1, noShare2); score != 0 {
		t.Errorf("expected 0 for non-matching xname prefixes, got %d", score)
	}
}

// TestClampScore verifies clampScore bounds a raw score into the 0–100 range.
//
// Why it matters: scores are summed from several heuristics and surfaced to the
// operator as a confidence percentage, so values must never exceed 100 or drop
// below 0 regardless of how many bonuses accumulate.
// Inputs: a value above 100, a negative value, and three in-range values
// including both endpoints. Outputs: 100 for the over-max, 0 for the negative,
// and the unchanged value otherwise. Data choice: testing both clamp branches
// plus the endpoints proves the comparisons are inclusive and one-sided.
func TestClampScore(t *testing.T) {
	cases := map[int]int{150: 100, -5: 0, 50: 50, 0: 0, 100: 100}
	for in, want := range cases {
		if got := clampScore(in); got != want {
			t.Errorf("clampScore(%d) = %d, want %d", in, got, want)
		}
	}
}

// TestMetaString verifies metaString returns string values directly, formats
// non-string values, and returns empty for a missing key.
//
// Why it matters: provider metadata is heterogeneous (any-typed), so the prefix
// scorer relies on metaString to coerce numbers and missing keys into
// comparable strings without panicking.
// Inputs: a map with a string value, an integer value, and a probe for an
// absent key. Outputs: the raw string, the integer rendered as text, and "".
// Data choice: the int case forces the fmt.Sprintf fallback that a string-only
// test would miss, and the absent key covers the not-ok guard.
func TestMetaString(t *testing.T) {
	m := map[string]any{"s": "hello", "n": 42}
	if got := metaString(m, "s"); got != "hello" {
		t.Errorf("metaString string = %q, want hello", got)
	}
	if got := metaString(m, "n"); got != "42" {
		t.Errorf("metaString int = %q, want 42", got)
	}
	if got := metaString(m, "missing"); got != "" {
		t.Errorf("metaString missing = %q, want empty", got)
	}
}

// TestExtractProviderSub verifies extractProviderSub returns the whole map for an
// empty provider, the nested map for a provider key, and nil when the key is
// absent or not a map.
//
// Why it matters: the prefix scorer walks both provider-scoped ("csm") and
// flat ("") metadata shapes, so this accessor must correctly disambiguate them
// to avoid comparing the wrong fields.
// Inputs: a metadata map queried with "" (flat), "csm" (nested map present),
// "redfish" (absent), and a "bad" key whose value is a non-map. Outputs: the
// original map, the nested map, and nil for the last two. Data choice: the
// non-map value exercises the failed type-assertion branch distinct from the
// simple missing-key branch.
func TestExtractProviderSub(t *testing.T) {
	sub := map[string]any{"xname": "x1"}
	meta := map[string]any{"csm": sub, "bad": "not-a-map"}

	if got := extractProviderSub(meta, ""); got["csm"] == nil {
		t.Error("empty provider should return the whole metadata map")
	}
	if got := extractProviderSub(meta, "csm"); got["xname"] != "x1" {
		t.Errorf("csm sub = %v, want xname=x1", got)
	}
	if got := extractProviderSub(meta, "redfish"); got != nil {
		t.Errorf("absent provider = %v, want nil", got)
	}
	if got := extractProviderSub(meta, "bad"); got != nil {
		t.Errorf("non-map provider = %v, want nil", got)
	}
}

// TestRackDetail verifies rackDetail renders model/type, capacity, and resolved
// location context for a rack candidate.
//
// Why it matters: this string is the operator's at-a-glance summary when
// choosing a parent rack, so each populated attribute must appear and a missing
// location must be omitted gracefully.
// Inputs: a rack with Model, UHeight, one occupying device, and a Location that
// resolves in the inventory; a second rack with no Model but a Type set; and a
// rack whose Location is not present in the inventory. Outputs: a detail string
// containing the model, "U"/"occupied", and "location:" for the first; the Type
// string for the second; and no location fragment for the third. Data choice:
// the Model-vs-Type pair exercises the if/else-if, and the unresolved location
// covers the lookup-miss branch.
func TestRackDetail(t *testing.T) {
	inv := NewInventory()
	locID := uuid.New()
	inv.Locations[locID] = &CaniLocationType{ID: locID, Name: "dc1"}

	rack := &CaniRackType{Model: "RackModel", UHeight: 42, Devices: []uuid.UUID{uuid.New()}, Location: locID}
	d := rackDetail(inv, rack)
	for _, want := range []string{"RackModel", "42U", "1/42 occupied", "location: dc1"} {
		if !strings.Contains(d, want) {
			t.Errorf("rackDetail = %q, missing %q", d, want)
		}
	}

	typed := rackDetail(inv, &CaniRackType{Type: "standard-rack"})
	if !strings.Contains(typed, "standard-rack") {
		t.Errorf("rackDetail (type only) = %q, want it to contain the Type", typed)
	}

	unresolved := rackDetail(inv, &CaniRackType{Model: "M", Location: uuid.New()})
	if strings.Contains(unresolved, "location:") {
		t.Errorf("rackDetail with unresolved location should omit location, got %q", unresolved)
	}
}

// TestLocationDetail verifies locationDetail concatenates every populated
// attribute of a location candidate.
//
// Why it matters: this summary helps an operator pick the right site/room when
// reparenting a rack, so type, rack count, facility, and description must all
// surface.
// Inputs: a location with LocationType, one rack, Facility, and Description set,
// then an empty location. Outputs: a string containing all four attributes and
// the rack count, and an empty string for the empty location. Data choice:
// populating every optional field at once asserts none of the append branches is
// skipped, while the empty case covers the all-absent path.
func TestLocationDetail(t *testing.T) {
	loc := &CaniLocationType{
		LocationType: "site",
		Racks:        []uuid.UUID{uuid.New()},
		Facility:     "fac-1",
		Description:  "primary site",
	}
	d := locationDetail(loc)
	for _, want := range []string{"site", "1 racks", "fac-1", "primary site"} {
		if !strings.Contains(d, want) {
			t.Errorf("locationDetail = %q, missing %q", d, want)
		}
	}

	if got := locationDetail(&CaniLocationType{}); got != "" {
		t.Errorf("empty locationDetail = %q, want empty", got)
	}
}

// TestScoreRackToLocation verifies scoreRackToLocation awards name-similarity and
// existing-rack bonuses with matching reasons, and scores an unrelated pair at
// zero.
//
// Why it matters: rack-to-location scoring drives the ranked parent list for
// orphan racks, so both bonus branches must fire together for a strong match and
// neither for an unrelated one.
// Inputs: an orphan named "datacenter-rack" against a location named
// "datacenter1" that already has a rack, then the same orphan against an
// unrelated empty location. Outputs: a positive score with both reason strings
// for the first, and a zero score with no reasons for the second. Data choice:
// the shared "datacenter" prefix guarantees name similarity while the pre-
// existing rack triggers the second bonus, isolating both additive branches.
func TestScoreRackToLocation(t *testing.T) {
	orphan := OrphanItem{Name: "datacenter-rack", Kind: "rack"}

	loc := &CaniLocationType{Name: "datacenter1", Racks: []uuid.UUID{uuid.New()}}
	score, reasons := scoreRackToLocation(orphan, loc)
	if score <= 0 {
		t.Errorf("expected positive score, got %d", score)
	}
	joined := joinReasons(reasons)
	if !strings.Contains(joined, "name similarity") || !strings.Contains(joined, "existing racks") {
		t.Errorf("reasons = %q, want both name-similarity and existing-racks", joined)
	}

	zero, zeroReasons := scoreRackToLocation(OrphanItem{Name: "zzz"}, &CaniLocationType{Name: "aaa"})
	if zero != 0 || len(zeroReasons) != 0 {
		t.Errorf("unrelated pair = (%d, %v), want (0, [])", zero, zeroReasons)
	}
}

// TestSuggestDeviceParentsMinScoreAndNilSkip verifies suggestDeviceParents skips
// nil rack entries and applies the minimum score of 1 to a rack that earns no
// heuristic points.
//
// Why it matters: the prompt shows every rack so an operator can always choose
// one, which requires a floor score; meanwhile nil map entries must never
// produce a phantom suggestion.
// Inputs: an inventory containing a nil rack entry and one real rack that scores
// zero (no name overlap, zero UHeight). Outputs: exactly one suggestion whose
// score is the clamped minimum of 1. Data choice: a zero-scoring rack isolates
// the score<1 floor branch, and the nil entry covers the continue guard.
func TestSuggestDeviceParentsMinScoreAndNilSkip(t *testing.T) {
	inv := NewInventory()
	inv.Racks[uuid.New()] = nil
	realID := uuid.New()
	inv.Racks[realID] = &CaniRackType{ID: realID, Name: "zzz-no-match"}

	out := suggestDeviceParents(inv, OrphanItem{Name: "aaa-orphan", Kind: "device"})
	if len(out) != 1 {
		t.Fatalf("expected 1 suggestion (nil skipped), got %d", len(out))
	}
	if out[0].Score != 1 {
		t.Errorf("min-score rack = %d, want clamped floor 1", out[0].Score)
	}
}
