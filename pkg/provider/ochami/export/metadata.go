package export

import "fmt"

// extractString returns a string value from a metadata map, or "" if absent/wrong type.
func extractString(meta map[string]any, key string) string {
	v, ok := meta[key]
	if !ok {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case fmt.Stringer:
		return s.String()
	default:
		return ""
	}
}

// extractIntPtr returns an *int from a metadata map, handling float64 (YAML round-trip),
// int, and string representations. Returns nil if absent or unparseable.
func extractIntPtr(meta map[string]any, key string) *int {
	v, ok := meta[key]
	if !ok {
		return nil
	}
	switch n := v.(type) {
	case int:
		return &n
	case float64:
		i := int(n)
		return &i
	case string:
		var i int
		if _, err := fmt.Sscanf(n, "%d", &i); err == nil {
			return &i
		}
		return nil
	default:
		return nil
	}
}

// extractStringSlice returns a []string from a metadata map, handling both
// []string and []interface{} (common after YAML/JSON round-trip).
func extractStringSlice(meta map[string]any, key string) []string {
	v, ok := meta[key]
	if !ok {
		return nil
	}
	switch s := v.(type) {
	case []string:
		return s
	case []any:
		out := make([]string, 0, len(s))
		for _, e := range s {
			out = append(out, fmt.Sprintf("%v", e))
		}
		return out
	default:
		return nil
	}
}
