package export

import "testing"

func TestExtractString(t *testing.T) {
	tests := []struct {
		name string
		meta map[string]any
		key  string
		want string
	}{
		{"present", map[string]any{"xname": "x3000c0s1b0"}, "xname", "x3000c0s1b0"},
		{"missing", map[string]any{"other": "val"}, "xname", ""},
		{"nil map", nil, "xname", ""},
		{"wrong type", map[string]any{"xname": 123}, "xname", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractString(tt.meta, tt.key)
			if got != tt.want {
				t.Errorf("extractString(%v, %q) = %q, want %q", tt.meta, tt.key, got, tt.want)
			}
		})
	}
}

func TestExtractIntPtr(t *testing.T) {
	tests := []struct {
		name    string
		meta    map[string]any
		key     string
		wantNil bool
		wantVal int
	}{
		{"int", map[string]any{"nid": 42}, "nid", false, 42},
		{"float64", map[string]any{"nid": float64(99)}, "nid", false, 99},
		{"string", map[string]any{"nid": "7"}, "nid", false, 7},
		{"bad string", map[string]any{"nid": "abc"}, "nid", true, 0},
		{"missing", map[string]any{}, "nid", true, 0},
		{"nil map", nil, "nid", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractIntPtr(tt.meta, tt.key)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %d", *got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil, got nil")
			}
			if *got != tt.wantVal {
				t.Errorf("got %d, want %d", *got, tt.wantVal)
			}
		})
	}
}

func TestExtractStringSlice(t *testing.T) {
	tests := []struct {
		name string
		meta map[string]any
		key  string
		want []string
	}{
		{"[]string", map[string]any{"host_aliases": []string{"a", "b"}}, "host_aliases", []string{"a", "b"}},
		{"[]any", map[string]any{"host_aliases": []any{"c", "d"}}, "host_aliases", []string{"c", "d"}},
		{"missing", map[string]any{}, "host_aliases", nil},
		{"nil map", nil, "host_aliases", nil},
		{"wrong type", map[string]any{"host_aliases": 123}, "host_aliases", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractStringSlice(tt.meta, tt.key)
			if tt.want == nil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
