package devicetypes

import (
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
