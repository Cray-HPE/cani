package devicetypes

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestPromptForParentSelectByNumber(t *testing.T) {
	inv := NewInventory()
	rackID := uuid.New()
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "rack1"}

	orphan := OrphanItem{
		ID:   uuid.New(),
		Name: "blade1",
		Kind: "device",
	}

	suggestions := []ParentSuggestion{
		{ID: rackID, Name: "rack1", Kind: "rack", Score: 50, Reason: "test"},
	}

	// Simulate user typing "1\n"
	input := strings.NewReader("1\n")
	var output bytes.Buffer

	opts := ClassifyOptions{
		NoColor: true,
		Writer:  &output,
		Reader:  input,
	}

	got, err := PromptForParent(inv, orphan, suggestions, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != rackID {
		t.Errorf("expected %s, got %s", rackID, got)
	}
	if !strings.Contains(output.String(), "rack1") {
		t.Error("output should contain 'rack1'")
	}
}

func TestPromptForParentSkip(t *testing.T) {
	orphan := OrphanItem{
		ID:   uuid.New(),
		Name: "blade1",
		Kind: "device",
	}

	input := strings.NewReader("k\n")
	var output bytes.Buffer

	opts := ClassifyOptions{
		NoColor: true,
		Writer:  &output,
		Reader:  input,
	}

	got, err := PromptForParent(nil, orphan, nil, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != uuid.Nil {
		t.Errorf("expected Nil UUID on skip, got %s", got)
	}
}

func TestPromptForParentUUIDInput(t *testing.T) {
	targetID := uuid.New()

	orphan := OrphanItem{
		ID:   uuid.New(),
		Name: "blade1",
		Kind: "device",
	}

	// Type a raw UUID
	input := strings.NewReader(targetID.String() + "\n")
	var output bytes.Buffer

	opts := ClassifyOptions{
		NoColor: true,
		Writer:  &output,
		Reader:  input,
	}

	got, err := PromptForParent(nil, orphan, nil, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != targetID {
		t.Errorf("expected %s, got %s", targetID, got)
	}
}

func TestScoreTier(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{80, "strong"},
		{70, "strong"},
		{50, "moderate"},
		{40, "moderate"},
		{20, "weak"},
		{0, "weak"},
	}
	for _, tt := range tests {
		got := scoreTier(tt.score)
		if got != tt.want {
			t.Errorf("scoreTier(%d) = %q, want %q", tt.score, got, tt.want)
		}
	}
}

// invWithRacks returns an inventory holding two named racks usable as device
// parents in search tests.
func invWithRacks(t *testing.T) (*Inventory, uuid.UUID, uuid.UUID) {
	t.Helper()
	inv := NewInventory()
	idA := uuid.New()
	idB := uuid.New()
	inv.Racks[idA] = &CaniRackType{ID: idA, Name: "rack-alpha"}
	inv.Racks[idB] = &CaniRackType{ID: idB, Name: "rack-beta"}
	return inv, idA, idB
}

// TestPromptParentSearchSelect verifies promptParentSearch returns the chosen
// candidate's UUID when a query matches and the operator selects by number.
//
// Why it matters: search is the escape hatch when none of the scored
// suggestions fit, so selecting a searched candidate must yield the exact
// parent UUID that reparenting will use.
// Inputs: an inventory with racks "rack-alpha"/"rack-beta", orphanKind
// "device", and a reader supplying the query "rack" then selection "1".
// Outputs: the UUID of the first (alphabetically sorted) matching rack and no
// error. Data choice: two racks sharing the "rack" prefix force the sort-and-
// index path that a single-result search would not exercise.
func TestPromptParentSearchSelect(t *testing.T) {
	inv, idA, _ := invWithRacks(t)
	var sb strings.Builder
	r := bufio.NewReader(strings.NewReader("rack\n1\n"))

	got, err := promptParentSearch(inv, "device", &sb, r, true)
	if err != nil {
		t.Fatalf("promptParentSearch error: %v", err)
	}
	if got != idA {
		t.Errorf("search select = %s, want %s (rack-alpha)", got, idA)
	}
}

// TestPromptParentSearchEdges verifies promptParentSearch returns uuid.Nil for
// an empty query, a no-match query, a skip selection, and an out-of-range
// selection.
//
// Why it matters: each non-selecting path must leave the orphan unparented
// rather than guessing, so the function has to converge on uuid.Nil for every
// inconclusive branch.
// Inputs: a rack inventory queried with "" (empty), "zzz" (no match),
// "rack" then "k" (skip), and "rack" then "99" (out of range). Outputs:
// uuid.Nil and no error for all four. Data choice: these four readers map one-
// to-one onto the function's early-return branches after the search executes.
func TestPromptParentSearchEdges(t *testing.T) {
	inv, _, _ := invWithRacks(t)

	run := func(input string) uuid.UUID {
		t.Helper()
		var sb strings.Builder
		r := bufio.NewReader(strings.NewReader(input))
		got, err := promptParentSearch(inv, "device", &sb, r, true)
		if err != nil {
			t.Fatalf("promptParentSearch(%q) error: %v", input, err)
		}
		return got
	}

	if got := run("\n"); got != uuid.Nil {
		t.Errorf("empty query = %s, want Nil", got)
	}
	if got := run("zzz\n"); got != uuid.Nil {
		t.Errorf("no-match query = %s, want Nil", got)
	}
	if got := run("rack\nk\n"); got != uuid.Nil {
		t.Errorf("skip selection = %s, want Nil", got)
	}
	if got := run("rack\n99\n"); got != uuid.Nil {
		t.Errorf("out-of-range selection = %s, want Nil", got)
	}
}

// TestPromptParentSearchReadErrors verifies promptParentSearch surfaces an error
// when the input stream ends before the query or before the selection is read.
//
// Why it matters: a truncated stdin must abort reparenting loudly rather than
// silently leaving the orphan unattached or hanging.
// Inputs: an empty reader (EOF on the query read) and a reader with a matching
// query but no selection line (EOF on the selection read). Outputs: a non-nil
// error and uuid.Nil in both cases. Data choice: the two readers isolate the
// two distinct ReadString error sites in the function.
func TestPromptParentSearchReadErrors(t *testing.T) {
	inv, _, _ := invWithRacks(t)

	var sb strings.Builder
	r := bufio.NewReader(strings.NewReader(""))
	if _, err := promptParentSearch(inv, "device", &sb, r, true); err == nil {
		t.Error("expected error on EOF before query, got nil")
	}

	var sb2 strings.Builder
	r2 := bufio.NewReader(strings.NewReader("rack"))
	if _, err := promptParentSearch(inv, "device", &sb2, r2, true); err == nil {
		t.Error("expected error on EOF before selection, got nil")
	}
}

// TestPromptForParentSearchPath verifies PromptForParent routes the "s" command
// through promptParentSearch and confirms the searched selection.
//
// Why it matters: this exercises the search branch of the top-level prompt that
// numeric- and skip-input tests bypass, ensuring the wiring between the menu
// loop and the search helper resolves to a parent UUID.
// Inputs: a rack inventory, an orphan device, no scored suggestions, and a
// reader supplying "s", then query "rack", then selection "1". Outputs: the
// first matching rack's UUID, no error, and a "selected" confirmation in the
// output. Data choice: passing nil suggestions forces the menu straight to the
// search path so the assertion targets only that branch.
func TestPromptForParentSearchPath(t *testing.T) {
	inv, idA, _ := invWithRacks(t)
	orphan := OrphanItem{ID: uuid.New(), Name: "blade1", Kind: "device"}
	var output bytes.Buffer
	opts := ClassifyOptions{NoColor: true, Writer: &output, Reader: strings.NewReader("s\nrack\n1\n")}

	got, err := PromptForParent(inv, orphan, nil, opts)
	if err != nil {
		t.Fatalf("PromptForParent error: %v", err)
	}
	if got != idA {
		t.Errorf("search path = %s, want %s (rack-alpha)", got, idA)
	}
	if !strings.Contains(output.String(), "selected") {
		t.Errorf("expected 'selected' confirmation, got %q", output.String())
	}
}
