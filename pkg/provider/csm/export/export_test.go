package export

import (
	"testing"
)

func TestSplitCSV(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "passing test with comma-separated values",
			input:    "name, type , status",
			expected: []string{"name", "type", "status"},
		},
		{
			name:     "failing test with empty string returns empty slice",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitCSV(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf("splitCSV() returned %d items, want %d", len(got), len(tt.expected))
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("splitCSV()[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}
