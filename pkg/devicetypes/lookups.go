package devicetypes

import (
	"fmt"
	"sort"
	"strings"
)

// minFuzzyLen is the minimum query length for fuzzy substring matching.
// Shorter queries can still match via exact part number / slug steps.
const minFuzzyLen = 3

// minAcceptScore is the lowest score considered a valid fuzzy match.
// Scores below this threshold are treated as "no match".
const minAcceptScore = 30

// MatchResult pairs a lookup result with a confidence score.
type MatchResult struct {
	Slug  string
	Score int
}

// ScoreTierLabel returns a short human-readable label for a match score.
func ScoreTierLabel(score int) string {
	switch {
	case score >= 100:
		return "exact match"
	case score >= 90:
		return "slug match"
	case score >= 75:
		return "multi-token match"
	case score >= 70:
		return "token match"
	case score >= 60:
		return "partial token match"
	case score >= 50:
		return "model substring"
	case score >= 40:
		return "slug substring"
	case score >= 30:
		return "name substring"
	case score >= 15:
		return "hardware-type fallback"
	default:
		return "no match"
	}
}

// Lookup searches the device type library for a match against the given query
// string. It checks in order: exact part number, exact slug, case-insensitive
// slug, then scored fuzzy match on slug, model, manufacturer, and name.
// Returns an error if no match is found or the query is empty.
func Lookup(query string) (CaniDeviceType, error) {
	dt, _ := LookupScored(query)
	if dt.Slug == "" {
		return CaniDeviceType{}, fmt.Errorf("no device type found for %q", query)
	}
	return dt, nil
}

// LookupScored is like Lookup but also returns a confidence score.
// Score 100 = exact part number or slug, lower = fuzzy. Returns zero-value
// with score 0 when nothing matches.
func LookupScored(query string) (CaniDeviceType, int) {
	if query == "" {
		return CaniDeviceType{}, 0
	}

	// 1. Exact part number match.
	if dt, ok := GetByPartNumber(query); ok {
		return dt, 100
	}

	// 2. Exact slug match.
	if dt, ok := GetBySlug(query); ok {
		return dt, 100
	}

	// 3. Case-insensitive slug match.
	lower := strings.ToLower(query)
	for slug, dt := range allDeviceTypes {
		if strings.ToLower(slug) == lower {
			return dt, 95
		}
	}

	// 4. Identification model match (case-insensitive).
	for _, dt := range allDeviceTypes {
		for _, id := range dt.Identifications {
			if strings.EqualFold(id.Model, query) {
				return dt, 100
			}
		}
	}

	// 5. Scored fuzzy match (requires minimum query length).
	if len(query) < minFuzzyLen {
		return CaniDeviceType{}, 0
	}
	return fuzzyMatch(query)
}

// LookupModule searches the module type library for a match against the
// given query string. Returns an error if no match is found.
func LookupModule(query string) (CaniModuleType, error) {
	mt, _ := LookupModuleScored(query)
	if mt.Slug == "" {
		return CaniModuleType{}, fmt.Errorf("no module type found for %q", query)
	}
	return mt, nil
}

// LookupModuleScored is like LookupModule but also returns a confidence score.
func LookupModuleScored(query string) (CaniModuleType, int) {
	if query == "" {
		return CaniModuleType{}, 0
	}

	// 1. Exact part number match.
	if mt, ok := GetModuleTypeByPartNumber(query); ok {
		return mt, 100
	}

	// 2. Exact slug match.
	if mt, ok := GetModuleBySlug(query); ok {
		return mt, 100
	}

	// 3. Case-insensitive slug match.
	lower := strings.ToLower(query)
	for slug, mt := range allModuleTypes {
		if strings.ToLower(slug) == lower {
			return mt, 95
		}
	}

	// 4. Scored fuzzy match.
	if len(query) < minFuzzyLen {
		return CaniModuleType{}, 0
	}
	return fuzzyMatchModule(query)
}

// ── Scoring helpers ─────────────────────────────────────────────────

// scoreFields computes a match score for a query against a library entry's
// text fields. Higher is better.
//
// Scoring:
//
//	100  exact model match (case-insensitive)
//	 90  exact slug match (case-insensitive)
//	 70  query equals a token in slug or model (split on -_. and space)
//	 50  query is a substring of model
//	 40  query is a substring of slug
//	 30  query is a substring of name
//	 20  query is a substring of manufacturer ONLY — rejected (returns 0)
func scoreFields(query, slug, model, manufacturer, name string) int {
	s := scoreFieldsCore(query, slug, model, manufacturer, name)
	if s > 0 {
		return s
	}
	// Try stripping common manufacturer prefixes from query.
	stripped := stripMfrPrefix(query)
	if stripped != "" && stripped != query {
		return scoreFieldsCore(stripped, slug, model, manufacturer, name)
	}
	return 0
}

// mfrPrefixes lists manufacturer prefixes (longest first) to strip from
// queries before fuzzy matching. Order matters: longest prefixes first.
var mfrPrefixes = []string{
	"hpe cray ",
	"hpe ",
	"cray ",
	"hp ",
}

// stripMfrPrefix removes a leading manufacturer prefix from query.
// Returns the remainder, or "" if no prefix matches or remainder is empty.
func stripMfrPrefix(query string) string {
	lower := strings.ToLower(query)
	for _, pfx := range mfrPrefixes {
		if strings.HasPrefix(lower, pfx) {
			remainder := strings.TrimSpace(query[len(pfx):])
			if remainder != "" {
				return remainder
			}
		}
	}
	return ""
}

// scoreFieldsCore computes the raw match score without prefix stripping.
func scoreFieldsCore(query, slug, model, manufacturer, name string) int {
	lower := strings.ToLower(query)

	// Exact field matches.
	if strings.EqualFold(model, query) {
		return 100
	}
	if strings.EqualFold(slug, query) {
		return 90
	}

	// Token match: query matches a whole token in slug or model.
	if tokenMatch(slug, lower) || tokenMatch(model, lower) {
		return 70
	}

	// Multi-token match: decompose query at letter↔digit boundaries and
	// check how many sub-tokens match tokens in slug or model.
	if s := multiTokenScore(lower, slug, model); s > 0 {
		return s
	}

	// Substring matches, ranked by field specificity.
	slugLower := strings.ToLower(slug)
	modelLower := strings.ToLower(model)
	nameLower := strings.ToLower(name)
	mfrLower := strings.ToLower(manufacturer)

	inModel := strings.Contains(modelLower, lower)
	inSlug := strings.Contains(slugLower, lower)
	inName := strings.Contains(nameLower, lower)
	inMfr := strings.Contains(mfrLower, lower)

	if inModel {
		return 50
	}
	if inSlug {
		return 40
	}
	if inName {
		return 30
	}
	// Manufacturer-only match is too broad (e.g. "Gigabyte" matches all
	// Gigabyte devices). Reject it.
	if inMfr {
		return 0
	}
	return 0
}

// tokenMatch reports whether lower appears as a complete token in s when s is
// split on common delimiters (-, _, ., space).
func tokenMatch(s, lower string) bool {
	for _, tok := range tokenize(s) {
		if strings.ToLower(tok) == lower {
			return true
		}
	}
	return false
}

// tokenize splits s on -, _, ., and space.
func tokenize(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == '.' || r == ' '
	})
}

// tokenizeCamelNum splits s on delimiter characters AND at letter↔digit
// boundaries. For example "dl360gen11" → ["dl360", "gen11"] and
// "mgmtsw0" → ["mgmtsw", "0"]. This allows compound device names
// without delimiters to be decomposed into matchable sub-tokens.
func tokenizeCamelNum(s string) []string {
	// First split on standard delimiters.
	parts := tokenize(s)
	var result []string
	for _, part := range parts {
		result = append(result, splitAlphaNum(part)...)
	}
	return result
}

// splitAlphaNum splits a string at letter→digit and digit→letter boundaries.
// "dl360gen11" → ["dl", "360", "gen", "11"], which are then recombined
// into meaningful tokens by grouping adjacent alpha+num pairs.
func splitAlphaNum(s string) []string {
	if s == "" {
		return nil
	}
	var segments []string
	start := 0
	for i := 1; i < len(s); i++ {
		prevDigit := s[i-1] >= '0' && s[i-1] <= '9'
		curDigit := s[i] >= '0' && s[i] <= '9'
		if prevDigit != curDigit {
			segments = append(segments, s[start:i])
			start = i
		}
	}
	segments = append(segments, s[start:])

	// Recombine adjacent alpha+digit pairs into compound tokens.
	// "dl","360","gen","11" → "dl360","gen11"
	var tokens []string
	for i := 0; i < len(segments); i++ {
		isDigit := len(segments[i]) > 0 && segments[i][0] >= '0' && segments[i][0] <= '9'
		if !isDigit && i+1 < len(segments) {
			nextIsDigit := len(segments[i+1]) > 0 && segments[i+1][0] >= '0' && segments[i+1][0] <= '9'
			if nextIsDigit {
				tokens = append(tokens, segments[i]+segments[i+1])
				i++ // skip next
				continue
			}
		}
		tokens = append(tokens, segments[i])
	}
	return tokens
}

// multiTokenScore decomposes lowerQuery at letter↔digit boundaries and checks
// how many sub-tokens match tokens in slug or model. Returns:
//   - 75 if ALL sub-tokens (≥2) match
//   - 60 if at least one sub-token (≥ minFuzzyLen) matches
//   - 0 otherwise
func multiTokenScore(lowerQuery, slug, model string) int {
	subs := tokenizeCamelNum(lowerQuery)
	if len(subs) < 2 {
		return 0
	}

	matched := 0
	for _, sub := range subs {
		if len(sub) < minFuzzyLen {
			continue
		}
		if tokenMatch(slug, sub) || tokenMatch(model, sub) {
			matched++
		}
	}

	// Count sub-tokens that are long enough to be meaningful.
	meaningful := 0
	for _, sub := range subs {
		if len(sub) >= minFuzzyLen {
			meaningful++
		}
	}
	if meaningful == 0 {
		return 0
	}
	if matched == meaningful {
		return 75 // all meaningful sub-tokens matched
	}
	if matched > 0 {
		return 60 // partial sub-token match
	}
	return 0
}

// ── Fuzzy matchers ──────────────────────────────────────────────────

// fuzzyMatch searches for the best-scoring device type for the query.
func fuzzyMatch(query string) (CaniDeviceType, int) {
	var best CaniDeviceType
	bestScore := 0

	for _, dt := range allDeviceTypes {
		s := scoreFields(query, dt.Slug, dt.Model, dt.Manufacturer, dt.Name)
		if s > bestScore || (s == bestScore && s > 0 && len(dt.Slug) < len(best.Slug)) {
			bestScore = s
			best = dt
		}
	}
	if bestScore < minAcceptScore {
		return CaniDeviceType{}, 0
	}
	return best, bestScore
}

// FuzzyMatchAll returns all device types that score at or above minAcceptScore
// for the given query, sorted by descending score. This is used by SuggestTypes
// to gather multiple candidates instead of just the single best.
func FuzzyMatchAll(query string, max int) []MatchResult {
	if len(query) < minFuzzyLen {
		return nil
	}
	var results []MatchResult
	for _, dt := range allDeviceTypes {
		s := scoreFields(query, dt.Slug, dt.Model, dt.Manufacturer, dt.Name)
		if s >= minAcceptScore {
			results = append(results, MatchResult{Slug: dt.Slug, Score: s})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if max > 0 && len(results) > max {
		results = results[:max]
	}
	return results
}

// fuzzyMatchModule searches for the best-scoring module type for the query.
func fuzzyMatchModule(query string) (CaniModuleType, int) {
	var best CaniModuleType
	bestScore := 0

	for _, mt := range allModuleTypes {
		s := scoreFields(query, mt.Slug, mt.Model, mt.Manufacturer, mt.Name)
		if s > bestScore || (s == bestScore && s > 0 && len(mt.Slug) < len(best.Slug)) {
			bestScore = s
			best = mt
		}
	}
	if bestScore < minAcceptScore {
		return CaniModuleType{}, 0
	}
	return best, bestScore
}

// containsFold reports whether substr appears within s, case-insensitive.
// Returns false if substr is empty.
func containsFold(s, lowerSubstr string) bool {
	if lowerSubstr == "" {
		return false
	}
	return strings.Contains(strings.ToLower(s), lowerSubstr)
}
