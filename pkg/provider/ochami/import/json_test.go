package import_

import (
	"strings"
	"testing"
)

func TestParseJson(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectErr   bool
		errorMsg    string
		expectCount int
	}{
		{
			name:        "valid JSON returns records",
			path:        "../../../../testdata/fixtures/ochami/ochami_test_data.json",
			expectErr:   false,
			expectCount: 32,
		},
		{
			name:      "nonexistent file returns error",
			path:      "nonexistent.json",
			expectErr: true,
			errorMsg:  "failed to open Json file",
		},
		{
			name:        "empty JSON returns no records",
			path:        "../../../../testdata/fixtures/ochami/empty.json",
			expectErr:   false,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJson(tt.path)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q but got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(got) != tt.expectCount {
				t.Errorf("expected %d records, got %d", tt.expectCount, len(got))
			}
		})
	}
}
