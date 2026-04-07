package devicetypes

import (
	"strings"
	"testing"
)

func TestLookupByPartNumber(t *testing.T) {
	// Find any device type with a part number to use as test input.
	var pn string
	var want CaniDeviceType
	for _, dt := range deviceTypesByPartNum {
		pn = dt.PartNumber
		want = dt
		break
	}
	if pn == "" {
		t.Skip("no device types with part numbers loaded")
	}

	got, err := Lookup(pn)
	if err != nil {
		t.Fatalf("Lookup(%q) error: %v", pn, err)
	}
	if got.Slug != want.Slug {
		t.Errorf("Lookup(%q) slug = %q, want %q", pn, got.Slug, want.Slug)
	}
}

func TestLookupBySlug(t *testing.T) {
	var slug string
	var want CaniDeviceType
	for s, dt := range allDeviceTypes {
		slug = s
		want = dt
		break
	}
	if slug == "" {
		t.Skip("no device types loaded")
	}

	got, err := Lookup(slug)
	if err != nil {
		t.Fatalf("Lookup(%q) error: %v", slug, err)
	}
	if got.Slug != want.Slug {
		t.Errorf("Lookup(%q) slug = %q, want %q", slug, got.Slug, want.Slug)
	}
}

func TestLookupCaseInsensitiveSlug(t *testing.T) {
	var slug string
	var want CaniDeviceType
	for s, dt := range allDeviceTypes {
		slug = s
		want = dt
		break
	}
	if slug == "" {
		t.Skip("no device types loaded")
	}

	// Try uppercase variant of the slug.
	upper := ""
	for _, c := range slug {
		if c >= 'a' && c <= 'z' {
			upper += string(c - 32)
		} else {
			upper += string(c)
		}
	}

	got, err := Lookup(upper)
	if err != nil {
		t.Fatalf("Lookup(%q) error: %v", upper, err)
	}
	if got.Slug != want.Slug {
		t.Errorf("Lookup(%q) slug = %q, want %q", upper, got.Slug, want.Slug)
	}
}

func TestLookupFuzzyModel(t *testing.T) {
	// Find a device type with a non-empty model field.
	var model string
	for _, dt := range allDeviceTypes {
		if dt.Model != "" {
			model = dt.Model
			break
		}
	}
	if model == "" {
		t.Skip("no device types with models loaded")
	}

	got, err := Lookup(model)
	if err != nil {
		t.Fatalf("Lookup(%q) error: %v", model, err)
	}
	// The returned entry must have a matching model (multiple slugs may
	// share the same model string, so we check the model, not the slug).
	if got.Model != model {
		t.Errorf("Lookup(%q) model = %q, want %q", model, got.Model, model)
	}
}

func TestLookupNoMatch(t *testing.T) {
	_, err := Lookup("zzz-nonexistent-device-xyz-99999")
	if err == nil {
		t.Fatal("Lookup should return error for non-existent device")
	}
}

func TestLookupEmptyQuery(t *testing.T) {
	_, err := Lookup("")
	if err == nil {
		t.Fatal("Lookup should return error for empty query")
	}
}

func TestLookupShortQueryNoFuzzy(t *testing.T) {
	// A 2-character query like "NA" should NOT fuzzy-match anything.
	_, err := Lookup("NA")
	if err == nil {
		t.Error("Lookup(\"NA\") should fail — too short for fuzzy matching")
	}
}

func TestLookupManufacturerOnlyRejected(t *testing.T) {
	// scoreFields should return 0 when the query matches ONLY manufacturer
	// and not slug, model, or name.
	s := scoreFields("SomeMfr", "unrelated-slug", "different-model", "SomeMfr", "other-name")
	if s != 0 {
		t.Errorf("manufacturer-only scoreFields = %d, want 0", s)
	}
}

func TestLookupScoredExactSlug(t *testing.T) {
	var slug string
	for s := range allDeviceTypes {
		slug = s
		break
	}
	if slug == "" {
		t.Skip("no device types loaded")
	}
	dt, score := LookupScored(slug)
	if score < 95 {
		t.Errorf("LookupScored(%q) score = %d, want >= 95", slug, score)
	}
	if dt.Slug != slug {
		t.Errorf("LookupScored(%q) slug = %q, want %q", slug, dt.Slug, slug)
	}
}

func TestScoreFieldsExactModel(t *testing.T) {
	s := scoreFields("DL380 Gen11", "hpe-dl380-gen11", "DL380 Gen11", "HPE", "")
	if s != 100 {
		t.Errorf("exact model score = %d, want 100", s)
	}
}

func TestScoreFieldsTokenMatch(t *testing.T) {
	// "dl380" is a token in "hpe-dl380-gen11" when split on '-'.
	s := scoreFields("dl380", "hpe-dl380-gen11", "DL380 Gen 11", "HPE", "")
	if s != 70 {
		t.Errorf("token match score = %d, want 70", s)
	}
}

func TestScoreFieldsManufacturerOnly(t *testing.T) {
	// Matching only manufacturer should score 0.
	s := scoreFields("Gigabyte", "other-thing", "X999", "Gigabyte", "")
	if s != 0 {
		t.Errorf("manufacturer-only score = %d, want 0", s)
	}
}

func TestTokenize(t *testing.T) {
	got := tokenize("hpe-dl380_gen.11 plus")
	want := []string{"hpe", "dl380", "gen", "11", "plus"}
	if len(got) != len(want) {
		t.Fatalf("tokenize len = %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("tokenize[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestContainsFoldMatch(t *testing.T) {
	tests := []struct {
		s, sub string
		want   bool
	}{
		{"HPE ProLiant DL360", "proliant", true},
		{"HPE ProLiant DL360", "dl360", true},
		{"HPE ProLiant DL360", "xyz", false},
		{"", "anything", false},
		{"something", "", false},
	}
	for _, tt := range tests {
		got := containsFold(tt.s, tt.sub)
		if got != tt.want {
			t.Errorf("containsFold(%q, %q) = %v, want %v", tt.s, tt.sub, got, tt.want)
		}
	}
}

// --- LookupModule tests ---

func TestLookupModuleByPartNumber(t *testing.T) {
	var pn string
	var want CaniModuleType
	for _, mt := range moduleTypesByPartNum {
		pn = mt.PartNumber
		want = mt
		break
	}
	if pn == "" {
		t.Skip("no module types with part numbers loaded")
	}

	got, err := LookupModule(pn)
	if err != nil {
		t.Fatalf("LookupModule(%q) error: %v", pn, err)
	}
	if got.Slug != want.Slug {
		t.Errorf("LookupModule(%q) slug = %q, want %q", pn, got.Slug, want.Slug)
	}
}

func TestLookupModuleBySlug(t *testing.T) {
	var slug string
	var want CaniModuleType
	for s, mt := range allModuleTypes {
		slug = s
		want = mt
		break
	}
	if slug == "" {
		t.Skip("no module types loaded")
	}

	got, err := LookupModule(slug)
	if err != nil {
		t.Fatalf("LookupModule(%q) error: %v", slug, err)
	}
	if got.Slug != want.Slug {
		t.Errorf("LookupModule(%q) slug = %q, want %q", slug, got.Slug, want.Slug)
	}
}

func TestLookupModuleNoMatch(t *testing.T) {
	_, err := LookupModule("zzzzzzzzz")
	if err == nil {
		t.Fatal("LookupModule should return error for non-existent module")
	}
}

func TestLookupModuleEmptyQuery(t *testing.T) {
	_, err := LookupModule("")
	if err == nil {
		t.Fatal("LookupModule should return error for empty query")
	}
}

// --- Manufacturer-prefix stripping tests ---

func TestStripMfrPrefix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"HPE CRAY EX235A", "EX235A"},
		{"HPE Cray XD295v", "XD295v"},
		{"Cray XD220v", "XD220v"},
		{"HP ProLiant", "ProLiant"},
		{"hpe dl380", "dl380"},
		{"UnknownMfr Something", ""},
		{"HPE", ""},  // nothing left after prefix
		{"HPE ", ""}, // nothing left after trim
	}
	for _, tt := range tests {
		got := stripMfrPrefix(tt.input)
		if got != tt.want {
			t.Errorf("stripMfrPrefix(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestScoreFieldsMfrPrefixStripping(t *testing.T) {
	// "HPE CRAY EX235A" should match model token "EX235A" via prefix stripping.
	s := scoreFields("HPE CRAY EX235A", "hpe-crayex-ex235a-compute-blade", "EX235A AMD MI200 accelerator blade (Bard Peak)", "HPE", "")
	if s < 50 {
		t.Errorf("mfr-prefix stripped score = %d, want >= 50", s)
	}
}

func TestLookupScoredIdentificationMatch(t *testing.T) {
	// Register a test device with identifications.
	RegisterDeviceType(CaniDeviceType{
		Slug:         "test-ident-device",
		Model:        "Test Device X9000",
		Manufacturer: "TestCorp",
		Identifications: []Identification{
			{Manufacturer: "TestCorp", Model: "X9000-Alias"},
		},
	})
	defer func() {
		delete(allDeviceTypes, "test-ident-device")
	}()

	dt, score := LookupScored("X9000-Alias")
	if score != 100 {
		t.Errorf("identification match score = %d, want 100", score)
	}
	if dt.Slug != "test-ident-device" {
		t.Errorf("identification match slug = %q, want %q", dt.Slug, "test-ident-device")
	}
}

func TestScoreTierLabel(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{100, "exact match"},
		{95, "slug match"},
		{75, "multi-token match"},
		{70, "token match"},
		{60, "partial token match"},
		{50, "model substring"},
		{40, "slug substring"},
		{30, "name substring"},
		{15, "hardware-type fallback"},
		{0, "no match"},
		{10, "no match"},
	}
	for _, tt := range tests {
		got := ScoreTierLabel(tt.score)
		if got != tt.want {
			t.Errorf("ScoreTierLabel(%d) = %q, want %q", tt.score, got, tt.want)
		}
	}
}

// --- tokenizeCamelNum tests ---

func TestTokenizeCamelNum(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"dl360gen11", []string{"dl360", "gen11"}},
		{"hpe-dl380-gen11", []string{"hpe", "dl380", "gen11"}},
		{"mgmtsw0", []string{"mgmtsw0"}},
		{"fmn", []string{"fmn"}},
		{"x9000c1s7b0n0", []string{"x9000", "c1", "s7", "b0", "n0"}},
		{"", nil},
	}
	for _, tt := range tests {
		got := tokenizeCamelNum(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("tokenizeCamelNum(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range tt.want {
			if got[i] != tt.want[i] {
				t.Errorf("tokenizeCamelNum(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestSplitAlphaNum(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"dl360gen11", []string{"dl360", "gen11"}},
		{"abc", []string{"abc"}},
		{"123", []string{"123"}},
		{"a1b2", []string{"a1", "b2"}},
		{"", nil},
	}
	for _, tt := range tests {
		got := splitAlphaNum(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitAlphaNum(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range tt.want {
			if got[i] != tt.want[i] {
				t.Errorf("splitAlphaNum(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestMultiTokenScoreAllMatch(t *testing.T) {
	// "dl380gen11" → sub-tokens ["dl380", "gen11"]
	// slug "hpe-dl380-gen-11" has tokens ["hpe", "dl380", "gen", "11"]
	// "dl380" matches, but "gen11" doesn't (tokens are "gen" and "11" separately)
	// So this is a partial match (score 60) for that slug.
	// But model "DL380 Gen 11" has tokens ["DL380", "Gen", "11"] — "dl380" matches.
	s := multiTokenScore("dl380gen11", "hpe-dl380-gen-11", "DL380 Gen 11")
	if s < 60 {
		t.Errorf("multiTokenScore(dl380gen11) = %d, want >= 60", s)
	}
}

func TestMultiTokenScoreCompound(t *testing.T) {
	// Register a device type with matching tokens to verify multi-token scoring.
	RegisterDeviceType(CaniDeviceType{
		Slug:         "test-dl360-gen11",
		Model:        "ProLiant DL360 Gen11",
		Manufacturer: "HPE",
	})
	defer func() {
		delete(allDeviceTypes, "test-dl360-gen11")
	}()

	// "dl360" is a token in slug "test-dl360-gen11".
	s := scoreFields("dl360gen11", "test-dl360-gen11", "ProLiant DL360 Gen11", "HPE", "")
	if s < 60 {
		t.Errorf("scoreFields(dl360gen11) multi-token = %d, want >= 60", s)
	}
}

func TestFuzzyMatchAll(t *testing.T) {
	RegisterDeviceType(CaniDeviceType{
		Slug:         "test-fuzzy-a",
		Model:        "FuzzyTestModel Alpha",
		Manufacturer: "TestCorp",
	})
	RegisterDeviceType(CaniDeviceType{
		Slug:         "test-fuzzy-b",
		Model:        "FuzzyTestModel Beta",
		Manufacturer: "TestCorp",
	})
	defer func() {
		delete(allDeviceTypes, "test-fuzzy-a")
		delete(allDeviceTypes, "test-fuzzy-b")
	}()

	results := FuzzyMatchAll("FuzzyTestModel", 10)
	if len(results) < 2 {
		t.Errorf("FuzzyMatchAll returned %d results, want >= 2", len(results))
	}
}

func TestLookupScoredFMN(t *testing.T) {
	// "fmn" is 3 chars — with minFuzzyLen=3, this should now fuzzy-match
	// any device type containing "fmn" as a token.
	RegisterDeviceType(CaniDeviceType{
		Slug:         "test-proliant-dl325-gen11-fmn",
		Model:        "ProLiant DL325 Gen11 FMN",
		Manufacturer: "HPE",
	})
	defer func() {
		delete(allDeviceTypes, "test-proliant-dl325-gen11-fmn")
	}()

	dt, score := LookupScored("fmn")
	if score == 0 {
		t.Fatal("LookupScored(\"fmn\") returned score 0, want > 0")
	}
	// May match the embedded type or the test-registered one — both have fmn.
	if !strings.Contains(dt.Slug, "fmn") {
		t.Errorf("LookupScored(\"fmn\") slug = %q, want slug containing \"fmn\"", dt.Slug)
	}
	if score < 30 {
		t.Errorf("LookupScored(\"fmn\") score = %d, want >= 30", score)
	}
}

// ---------- LookupModuleScored ----------

func TestLookupModuleScoredReturnsScore(t *testing.T) {
	RegisterModuleType(CaniModuleType{
		Slug:         "test-scored-module",
		Model:        "Test Scored Module",
		Manufacturer: "TestCo",
		PartNumber:   "TSM-001",
	})
	defer func() { delete(allModuleTypes, "test-scored-module") }()

	mt, score := LookupModuleScored("TSM-001")
	if score < 100 {
		t.Errorf("LookupModuleScored(\"TSM-001\") score = %d, want >= 100", score)
	}
	if mt.Slug != "test-scored-module" {
		t.Errorf("slug = %q, want test-scored-module", mt.Slug)
	}
}

func TestLookupModuleScoredNoMatch(t *testing.T) {
	_, score := LookupModuleScored("zzz-no-such-module-zzz")
	if score != 0 {
		t.Errorf("expected score 0 for no match, got %d", score)
	}
}

// ---------- tokenizeCamelNum ----------

func TestTokenizeCamelNumSplitsLetterDigit(t *testing.T) {
	tokens := tokenizeCamelNum("dl360gen11")
	want := map[string]bool{"dl360": true, "gen11": true}
	for _, tok := range tokens {
		delete(want, tok)
	}
	if len(want) != 0 {
		t.Errorf("tokenizeCamelNum(\"dl360gen11\") missing tokens: %v, got %v", want, tokens)
	}
}

func TestTokenizeCamelNumPlainToken(t *testing.T) {
	tokens := tokenizeCamelNum("blade")
	if len(tokens) != 1 || tokens[0] != "blade" {
		t.Errorf("tokenizeCamelNum(\"blade\") = %v, want [blade]", tokens)
	}
}
