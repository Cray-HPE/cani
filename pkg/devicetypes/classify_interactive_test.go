package devicetypes

// Test coverage for classify_interactive.go (non-I/O helpers)
//
// | Function     | Happy-path test                          | Failure test                          |
// |--------------|------------------------------------------|---------------------------------------|
// | searchSlugs  | TestSearchSlugsCaseInsensitive            | TestSearchSlugsNoResults              |
// | colorFuncs   | TestColorFuncsNoColor                    | TestColorFuncsWithColor               |

import (
	"strings"
	"testing"
)

// ---------- searchSlugs ----------

func TestSearchSlugsCaseInsensitive(t *testing.T) {
	// Register a device type so GetAllSlugs has at least this slug.
	RegisterDeviceType(CaniDeviceType{
		Slug:         "test-classify-interactive-device",
		Model:        "Interactive Test",
		Manufacturer: "TestCo",
	})
	defer func() { delete(allDeviceTypes, "test-classify-interactive-device") }()

	// Search with different casing.
	results := searchSlugs("CLASSIFY-INTERACTIVE", 10)
	found := false
	for _, slug := range results {
		if slug == "test-classify-interactive-device" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("searchSlugs(\"CLASSIFY-INTERACTIVE\") did not find expected slug, got %v", results)
	}
}

func TestSearchSlugsNoResults(t *testing.T) {
	results := searchSlugs("zzz-absolutely-no-match-zzz", 10)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d: %v", len(results), results)
	}
}

// ---------- colorFuncs ----------

func TestColorFuncsNoColor(t *testing.T) {
	cyan, yellow, green, gray, bold := colorFuncs(true)

	// When noColor is true, all functions should return the input unchanged.
	input := "hello"
	for name, fn := range map[string]func(string) string{
		"cyan": cyan, "yellow": yellow, "green": green,
		"gray": gray, "bold": bold,
	} {
		got := fn(input)
		if got != input {
			t.Errorf("colorFuncs(noColor=true).%s(%q) = %q, want %q", name, input, got, input)
		}
	}
}

func TestColorFuncsWithColor(t *testing.T) {
	cyan, yellow, green, gray, bold := colorFuncs(false)

	// When noColor is false, functions should wrap with ANSI escape codes.
	input := "test"
	for name, fn := range map[string]func(string) string{
		"cyan": cyan, "yellow": yellow, "green": green,
		"gray": gray, "bold": bold,
	} {
		got := fn(input)
		if !strings.Contains(got, input) {
			t.Errorf("colorFuncs(noColor=false).%s(%q) = %q, missing input text", name, input, got)
		}
		if !strings.Contains(got, "\033[") {
			t.Errorf("colorFuncs(noColor=false).%s(%q) = %q, missing ANSI escape", name, input, got)
		}
	}
}
