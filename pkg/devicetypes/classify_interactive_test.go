package devicetypes

// Test coverage for classify_interactive.go
//
// | Function                 | Happy-path test                  | Failure / edge test            |
// |--------------------------|----------------------------------|--------------------------------|
// | searchSlugs              | TestSearchSlugsCaseInsensitive   | TestSearchSlugsNoResults       |
// | colorFuncs               | TestColorFuncsNoColor            | TestColorFuncsWithColor        |
// | selectSuggestion         | TestSelectSuggestion             | TestSelectSuggestion           |
// | resolveSlugInput         | TestResolveSlugInput             | TestResolveSlugInput           |
// | readTypeSelection        | TestReadTypeSelection            | TestReadTypeSelectionReadError |
// | promptSearch             | TestPromptSearch                 | TestPromptSearch               |
// | handleSearchSelection    | TestHandleSearchSelection        | -                              |
// | printUnclassifiedDevice  | TestPrintUnclassifiedDevice      | -                              |
// | printTypeSuggestions     | TestPrintTypeSuggestions         | TestPrintTypeSuggestions       |
// | printProviderMetadata    | TestPrintProviderMetadata        | TestPrintProviderMetadata      |
// | PromptForDeviceType      | TestPromptForDeviceTypeSkip      | -                              |

import (
	"bufio"
	"os"
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

// ---------- interactive I/O helpers ----------

// withTestSlugs registers a few deterministic device-type slugs on top of the
// embedded library and removes exactly those slugs when the test finishes, so
// the shared registry is left untouched for other tests. It returns the slugs
// it registered so callers can drive lookups by slug.
func withTestSlugs(t *testing.T) []string {
	t.Helper()
	slugs := []string{"acme-server", "acme-switch", "globex-blade"}
	for _, s := range slugs {
		RegisterDeviceType(CaniDeviceType{Slug: s, Manufacturer: "Acme", Model: s})
	}
	t.Cleanup(func() {
		for _, s := range slugs {
			delete(allDeviceTypes, s)
		}
	})
	return slugs
}

// noColorFuncs returns the identity coloring closures used by helpers under
// test, matching what colorFuncs(true) produces.
func noColorFuncs() (cyan, yellow, green, gray, bold func(string) string) {
	return colorFuncs(true)
}

// TestSelectSuggestion verifies selectSuggestion maps a 1-based numeric choice
// to its slug and rejects out-of-range or non-numeric input.
//
// Why it matters: the menu lets operators pick a suggestion by number, so an
// off-by-one or weak bounds check would classify a device as the wrong type.
// Inputs: a two-entry suggestion list with choices "1", "2", "3" (over),
// "0" (under), and "x" (non-numeric). Outputs: the matching slug with ok=true
// for valid choices and ok=false otherwise.
// Data choice: the boundary values 0 and len+1 plus a non-numeric string cover
// both edges of the range check and the Atoi failure branch.
func TestSelectSuggestion(t *testing.T) {
	suggestions := []MatchResult{{Slug: "first", Score: 90}, {Slug: "second", Score: 80}}
	cases := []struct {
		input    string
		wantSlug string
		wantOK   bool
	}{
		{"1", "first", true},
		{"2", "second", true},
		{"3", "", false},
		{"0", "", false},
		{"x", "", false},
	}
	for _, tc := range cases {
		slug, ok := selectSuggestion(tc.input, suggestions)
		if slug != tc.wantSlug || ok != tc.wantOK {
			t.Errorf("selectSuggestion(%q) = (%q,%v), want (%q,%v)", tc.input, slug, ok, tc.wantSlug, tc.wantOK)
		}
	}
}

// TestResolveSlugInput verifies resolveSlugInput resolves a numeric menu choice,
// a directly typed registered slug, and reports no match otherwise — printing a
// confirmation line on success.
//
// Why it matters: the prompt accepts either a number or a literal slug, so this
// function is the single junction that turns raw input into a confirmed device
// type and must not silently accept unknown slugs.
// Inputs: a one-entry suggestion list plus a registered "acme-server" slug,
// queried with "1" (numeric), "acme-server" (direct slug), and "nope"
// (unknown). Outputs: the resolved slug with ok=true and a confirmation written
// to the buffer for the first two, and ("",false) with no output for the third.
// Data choice: using a registered slug exercises the GetBySlug branch that a
// numeric-only test would miss, and "nope" proves unknown input is rejected.
func TestResolveSlugInput(t *testing.T) {
	withTestSlugs(t)
	_, _, green, _, _ := noColorFuncs()
	suggestions := []MatchResult{{Slug: "from-menu", Score: 90}}

	var sb strings.Builder
	if slug, ok := resolveSlugInput(&sb, "1", suggestions, green); !ok || slug != "from-menu" {
		t.Errorf("numeric resolveSlugInput = (%q,%v), want (from-menu,true)", slug, ok)
	}
	if !strings.Contains(sb.String(), "from-menu") {
		t.Errorf("expected confirmation to mention slug, got %q", sb.String())
	}

	sb.Reset()
	if slug, ok := resolveSlugInput(&sb, "acme-server", suggestions, green); !ok || slug != "acme-server" {
		t.Errorf("direct-slug resolveSlugInput = (%q,%v), want (acme-server,true)", slug, ok)
	}

	sb.Reset()
	if slug, ok := resolveSlugInput(&sb, "nope", suggestions, green); ok || slug != "" {
		t.Errorf("unknown resolveSlugInput = (%q,%v), want (\"\",false)", slug, ok)
	}
	if sb.String() != "" {
		t.Errorf("expected no output for unknown input, got %q", sb.String())
	}
}

// TestReadTypeSelection verifies readTypeSelection routes skip, numeric, search,
// and retry-after-invalid input to the correct outcome.
//
// Why it matters: this loop is the heart of the classification prompt; mis-
// routing any token would either skip a device that should be classified or
// accept an invalid choice.
// Inputs: a one-entry suggestion list driven by four separate readers — "k"
// (skip), "1" (numeric select), "garbage" then "1" (invalid then select), and
// "s" then a search query then "1" (search path). Outputs: an empty slug for
// skip and the suggestion/search slug otherwise, with an "Invalid input"
// message emitted on the retry case.
// Data choice: feeding "garbage" before a valid number forces the loop to
// iterate, proving the invalid-input branch and re-prompt actually execute.
func TestReadTypeSelection(t *testing.T) {
	slugs := withTestSlugs(t)
	_, _, green, _, _ := noColorFuncs()
	suggestions := []MatchResult{{Slug: slugs[0], Score: 90}}

	read := func(input string) (string, string) {
		t.Helper()
		var sb strings.Builder
		r := bufio.NewReader(strings.NewReader(input))
		got, err := readTypeSelection(&sb, r, suggestions, true, green)
		if err != nil {
			t.Fatalf("readTypeSelection(%q) error: %v", input, err)
		}
		return got, sb.String()
	}

	if got, _ := read("k\n"); got != "" {
		t.Errorf("skip input = %q, want empty", got)
	}
	if got, _ := read("1\n"); got != slugs[0] {
		t.Errorf("numeric input = %q, want %q", got, slugs[0])
	}
	got, out := read("garbage\n1\n")
	if got != slugs[0] {
		t.Errorf("retry input = %q, want %q", got, slugs[0])
	}
	if !strings.Contains(out, "Invalid input") {
		t.Errorf("expected 'Invalid input' message, got %q", out)
	}
	if got, _ := read("s\nacme\n1\n"); got == "" {
		t.Error("search input returned empty, want a slug")
	}
}

// TestReadTypeSelectionReadError verifies readTypeSelection surfaces an error
// when the input stream ends before a selection is made.
//
// Why it matters: a closed or truncated stdin must fail loudly rather than
// silently classifying a device, so the caller can abort the import.
// Inputs: an empty reader (immediate EOF). Outputs: a non-nil error and an
// empty slug. Data choice: an empty string is the simplest trigger for the
// ReadString error branch that all other inputs avoid.
func TestReadTypeSelectionReadError(t *testing.T) {
	_, _, green, _, _ := noColorFuncs()
	r := bufio.NewReader(strings.NewReader(""))
	got, err := readTypeSelection(nil, r, nil, true, green)
	if err == nil {
		t.Fatal("expected error on EOF, got nil")
	}
	if got != "" {
		t.Errorf("got slug %q on error, want empty", got)
	}
}

// TestPromptSearch verifies promptSearch covers the empty-query, no-match,
// numeric-selection, skip-selection, and out-of-range-selection paths.
//
// Why it matters: free-text search is the fallback when no suggestion fits, so
// each branch determines whether an operator can reach the right type or is
// dropped back to a skip.
// Inputs: a registry of acme/globex slugs driven by five readers — empty query,
// a no-match query, a matching query then "1", a matching query then "k", and a
// matching query then an out-of-range index. Outputs: a resolved slug only for
// the numeric-in-range case and an empty slug for the rest.
// Data choice: pairing one matching query with three different second-line
// responses isolates the selection-handling branches from the search itself.
func TestPromptSearch(t *testing.T) {
	withTestSlugs(t)

	search := func(input string) string {
		t.Helper()
		var sb strings.Builder
		r := bufio.NewReader(strings.NewReader(input))
		got, err := promptSearch(&sb, r, true)
		if err != nil {
			t.Fatalf("promptSearch(%q) error: %v", input, err)
		}
		return got
	}

	if got := search("\n"); got != "" {
		t.Errorf("empty query = %q, want empty", got)
	}
	if got := search("zzznomatch\n"); got != "" {
		t.Errorf("no-match query = %q, want empty", got)
	}
	if got := search("acme\n1\n"); got == "" {
		t.Error("matched query + select 1 returned empty, want a slug")
	}
	if got := search("acme\nk\n"); got != "" {
		t.Errorf("matched query + skip = %q, want empty", got)
	}
	if got := search("acme\n99\n"); got != "" {
		t.Errorf("matched query + out-of-range = %q, want empty", got)
	}
}

// TestHandleSearchSelection verifies handleSearchSelection prints a confirmation
// when promptSearch yields a slug.
//
// Why it matters: this wrapper is what the main loop calls for the search path,
// and the confirmation line is the operator's only feedback that the choice
// registered.
// Inputs: a matching query then "1" against the test registry. Outputs: a
// non-empty slug and a buffer containing the check-mark confirmation.
// Data choice: reusing the registered "acme" prefix guarantees a deterministic
// single-step search-and-select without depending on the embedded library.
func TestHandleSearchSelection(t *testing.T) {
	withTestSlugs(t)
	_, _, green, _, _ := noColorFuncs()
	var sb strings.Builder
	r := bufio.NewReader(strings.NewReader("acme\n1\n"))
	slug, err := handleSearchSelection(&sb, r, true, green)
	if err != nil {
		t.Fatalf("handleSearchSelection error: %v", err)
	}
	if slug == "" {
		t.Fatal("expected a slug, got empty")
	}
	if !strings.Contains(sb.String(), slug) {
		t.Errorf("expected confirmation to contain %q, got %q", slug, sb.String())
	}
}

// TestPrintUnclassifiedDevice verifies printUnclassifiedDevice renders every
// optional field that is populated.
//
// Why it matters: this header is the operator's snapshot of the device being
// classified, so dropped fields would hide the very signals (model, role,
// provider hints) needed to choose a type.
// Inputs: a device with Name, DeviceType, Status, Role, Model, Manufacturer,
// ChildrenCount, and a CSM provider-metadata map populated. Outputs: a buffer
// mentioning the name, model, manufacturer, child count, and provider key.
// Data choice: populating every optional field at once asserts that none of the
// conditional print branches is skipped.
func TestPrintUnclassifiedDevice(t *testing.T) {
	cyan, _, _, gray, bold := noColorFuncs()
	device := UnclassifiedDevice{
		Name:          "x3000c0s1b0n0",
		DeviceType:    "node",
		Status:        "Active",
		Role:          "Compute",
		Model:         "R272",
		Manufacturer:  "Gigabyte",
		ChildrenCount: 2,
		ProviderMetadata: map[string]any{
			"csm": map[string]any{"xname": "x3000c0s1b0n0", "role": "Compute"},
		},
	}
	var sb strings.Builder
	printUnclassifiedDevice(&sb, device, cyan, gray, bold)
	out := sb.String()
	for _, want := range []string{"x3000c0s1b0n0", "R272", "Gigabyte", "Compute", "csm"} {
		if !strings.Contains(out, want) {
			t.Errorf("printUnclassifiedDevice output missing %q:\n%s", want, out)
		}
	}
}

// TestPrintTypeSuggestions verifies printTypeSuggestions renders a numbered list
// when suggestions exist and a placeholder when the list is empty.
//
// Why it matters: the suggestion list is the primary fast path for
// classification, and the empty-state message tells the operator to fall back
// to search.
// Inputs: a two-entry suggestion list, then an empty list. Outputs: output
// containing "[1]" and the first slug for the populated case, and a
// "No suggestions found" notice for the empty case.
// Data choice: distinct scores exercise the ScoreTierLabel formatting while the
// empty slice covers the early-return branch.
func TestPrintTypeSuggestions(t *testing.T) {
	_, yellow, _, gray, bold := noColorFuncs()

	var sb strings.Builder
	printTypeSuggestions(&sb, []MatchResult{{Slug: "acme-server", Score: 95}, {Slug: "acme-switch", Score: 60}}, yellow, gray, bold)
	out := sb.String()
	if !strings.Contains(out, "[1]") || !strings.Contains(out, "acme-server") {
		t.Errorf("expected numbered suggestion list, got %q", out)
	}

	var empty strings.Builder
	printTypeSuggestions(&empty, nil, yellow, gray, bold)
	if !strings.Contains(empty.String(), "No suggestions found") {
		t.Errorf("expected empty-state message, got %q", empty.String())
	}
}

// TestPrintProviderMetadata verifies printProviderMetadata renders selected keys
// for map-typed provider entries and skips empty maps, non-map values, and
// entries with no recognized keys.
//
// Why it matters: provider metadata (CSM xname, class, role) is decisive for
// classification, but the renderer must defend against the heterogeneous
// any-typed values that arrive from different providers.
// Inputs: a nil map; a CSM map with xname/class; a provider whose value is a
// plain string (non-map); and a provider map containing only unrecognized keys.
// Outputs: empty output for the nil, non-map, and unrecognized-key cases, and a
// line containing "xname=" for the CSM map.
// Data choice: the four shapes correspond one-to-one with the function's guard
// branches (empty, type-assert failure, no-parts, and the happy path).
func TestPrintProviderMetadata(t *testing.T) {
	_, _, _, gray, _ := noColorFuncs()

	var empty strings.Builder
	printProviderMetadata(&empty, nil, gray)
	if empty.String() != "" {
		t.Errorf("nil metadata produced output: %q", empty.String())
	}

	var ok strings.Builder
	printProviderMetadata(&ok, map[string]any{
		"csm": map[string]any{"xname": "x1", "class": "River"},
	}, gray)
	if !strings.Contains(ok.String(), "xname=x1") {
		t.Errorf("expected xname in output, got %q", ok.String())
	}

	var nonMap strings.Builder
	printProviderMetadata(&nonMap, map[string]any{"csm": "not-a-map"}, gray)
	if nonMap.String() != "" {
		t.Errorf("non-map provider value produced output: %q", nonMap.String())
	}

	var noKeys strings.Builder
	printProviderMetadata(&noKeys, map[string]any{"csm": map[string]any{"other": "v"}}, gray)
	if noKeys.String() != "" {
		t.Errorf("provider with no recognized keys produced output: %q", noKeys.String())
	}
}

// TestPromptForDeviceTypeSkip verifies the end-to-end prompt prints the device
// header and returns an empty slug when the operator skips.
//
// Why it matters: this is the public entry point used by the classify command,
// so it must wire the writer/reader options together and honor a skip without
// error.
// Inputs: a NoColor prompt over a buffered reader supplying "k", for a device
// with a name and a model. Outputs: an empty slug, no error, and output
// containing the unclassified-device header and the device name.
// Data choice: the skip token is the only input that exercises the full render
// path while keeping the result independent of suggestion scoring.
func TestPromptForDeviceTypeSkip(t *testing.T) {
	withTestSlugs(t)
	var sb strings.Builder
	device := UnclassifiedDevice{Name: "unknown-node", Model: "ZZ9"}
	opts := ClassifyOptions{NoColor: true, Writer: &sb, Reader: strings.NewReader("k\n")}

	slug, err := PromptForDeviceType(device, opts)
	if err != nil {
		t.Fatalf("PromptForDeviceType error: %v", err)
	}
	if slug != "" {
		t.Errorf("skip returned slug %q, want empty", slug)
	}
	if !strings.Contains(sb.String(), "Unclassified Device") || !strings.Contains(sb.String(), "unknown-node") {
		t.Errorf("expected header and device name in output, got %q", sb.String())
	}
}

// TestClassifyOptionsWriterReaderDefaults verifies the writer and reader
// accessors fall back to the process standard streams when unset.
//
// Why it matters: non-interactive callers construct a zero-value ClassifyOptions
// and expect prompts to reach the real terminal, so the nil-fallback path must
// return os.Stdout/os.Stdin rather than a nil stream that would panic on write.
// Inputs: a zero-value ClassifyOptions. Outputs: writer() == os.Stdout and
// reader() == os.Stdin. Data choice: leaving both fields nil is the only way to
// exercise the default branch that the buffer-backed prompt tests bypass.
func TestClassifyOptionsWriterReaderDefaults(t *testing.T) {
	var opts ClassifyOptions
	if opts.writer() != os.Stdout {
		t.Error("writer() should default to os.Stdout when Writer is nil")
	}
	if opts.reader() != os.Stdin {
		t.Error("reader() should default to os.Stdin when Reader is nil")
	}
}
