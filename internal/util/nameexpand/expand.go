/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package nameexpand

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Expand expands a bash-style brace pattern into a list of names.
// Supports numeric ranges ({1..4}), letter ranges ({a..d}), and
// zero-padded numeric ranges ({001..100}). Multiple brace groups
// produce a cartesian product ({a..b}{1..2} -> a1, a2, b1, b2).
// Returns an error if the pattern is malformed.
func Expand(pattern string) ([]string, error) {
	segments, err := parseSegments(pattern)
	if err != nil {
		return nil, err
	}
	return cartesian(segments), nil
}

// segment represents either a literal string or a list of expanded values.
type segment struct {
	values []string
}

// parseSegments splits a pattern into alternating literal and brace-expanded segments.
func parseSegments(pattern string) ([]segment, error) {
	var segments []segment
	i := 0

	for i < len(pattern) {
		braceStart := strings.Index(pattern[i:], "{")
		if braceStart == -1 {
			segments = append(segments, segment{values: []string{pattern[i:]}})
			break
		}

		braceStart += i
		if braceStart > i {
			segments = append(segments, segment{values: []string{pattern[i:braceStart]}})
		}

		braceEnd := strings.Index(pattern[braceStart:], "}")
		if braceEnd == -1 {
			return nil, fmt.Errorf("unclosed brace at position %d", braceStart)
		}
		braceEnd += braceStart

		inner := pattern[braceStart+1 : braceEnd]
		expanded, err := expandRange(inner)
		if err != nil {
			return nil, fmt.Errorf("invalid range {%s}: %w", inner, err)
		}
		segments = append(segments, segment{values: expanded})
		i = braceEnd + 1
	}

	if len(segments) == 0 {
		return []segment{{values: []string{pattern}}}, nil
	}
	return segments, nil
}

// expandRange expands the inner content of a brace group.
// Supports comma-separated lists ({1,4,7} or {a,b,c}) and
// ranges ({1..4} or {a..d}).
func expandRange(inner string) ([]string, error) {
	if strings.Contains(inner, ",") {
		return expandCommaList(inner)
	}

	parts := strings.SplitN(inner, "..", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("expected START..END or comma-separated format")
	}

	start := parts[0]
	end := parts[1]

	if start == "" || end == "" {
		return nil, fmt.Errorf("empty start or end value")
	}

	if isNumeric(start) && isNumeric(end) {
		return expandNumericRange(start, end)
	}
	if isAlpha(start) && isAlpha(end) {
		return expandAlphaRange(start, end)
	}
	return nil, fmt.Errorf("cannot mix numeric and alphabetic ranges: %q..%q", start, end)
}

// expandCommaList splits a comma-separated list into its elements.
// Each element is returned as-is (no padding or transformation).
func expandCommaList(inner string) ([]string, error) {
	parts := strings.Split(inner, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, fmt.Errorf("empty element in comma-separated list")
		}
		result = append(result, p)
	}
	return result, nil
}

// expandNumericRange expands "1..4" or "01..04" into a list of strings,
// preserving zero-padding based on the wider of the two operands.
func expandNumericRange(startStr, endStr string) ([]string, error) {
	startVal, err := strconv.Atoi(startStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start number %q: %w", startStr, err)
	}
	endVal, err := strconv.Atoi(endStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end number %q: %w", endStr, err)
	}

	if startVal > endVal {
		return nil, fmt.Errorf("start (%d) must not be greater than end (%d)", startVal, endVal)
	}

	padWidth := len(startStr)
	if len(endStr) > padWidth {
		padWidth = len(endStr)
	}

	var result []string
	for v := startVal; v <= endVal; v++ {
		result = append(result, padNumber(v, padWidth))
	}
	return result, nil
}

// expandAlphaRange expands "a..d" into [a, b, c, d] or "A..D" into [A, B, C, D].
// Supports multi-character ranges like "y..ab" using spreadsheet-style column naming.
func expandAlphaRange(startStr, endStr string) ([]string, error) {
	startVal := alphaToIndex(startStr)
	endVal := alphaToIndex(endStr)

	if startVal > endVal {
		return nil, fmt.Errorf("start (%q) must not be greater than end (%q)", startStr, endStr)
	}

	upper := unicode.IsUpper(rune(startStr[0]))

	var result []string
	for v := startVal; v <= endVal; v++ {
		result = append(result, indexToAlpha(v, upper))
	}
	return result, nil
}

// alphaToIndex converts a column-style alpha string to a zero-based index.
// "a" -> 0, "b" -> 1, ..., "z" -> 25, "aa" -> 26, "ab" -> 27, etc.
func alphaToIndex(s string) int {
	s = strings.ToLower(s)
	idx := 0
	for _, c := range s {
		idx = idx*26 + int(c-'a') + 1
	}
	return idx - 1
}

// indexToAlpha converts a zero-based index back to column-style alpha string.
// 0 -> "a", 25 -> "z", 26 -> "aa", 27 -> "ab", etc.
func indexToAlpha(idx int, upper bool) string {
	var result []byte
	idx++ // convert to 1-based
	for idx > 0 {
		idx--
		if upper {
			result = append([]byte{byte('A' + idx%26)}, result...)
		} else {
			result = append([]byte{byte('a' + idx%26)}, result...)
		}
		idx /= 26
	}
	return string(result)
}

// cartesian computes the cartesian product of all segments by concatenating one
// value from each segment in order.
func cartesian(segments []segment) []string {
	result := []string{""}
	for _, seg := range segments {
		var next []string
		for _, prefix := range result {
			for _, val := range seg.values {
				next = append(next, prefix+val)
			}
		}
		result = next
	}
	return result
}

// padNumber formats n with at least width digits, zero-padded on the left.
func padNumber(n, width int) string {
	s := strconv.Itoa(n)
	for len(s) < width {
		s = "0" + s
	}
	return s
}

// isNumeric returns true if every character in s is a digit.
func isNumeric(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return len(s) > 0
}

// isAlpha returns true if every character in s is a letter.
func isAlpha(s string) bool {
	for _, c := range s {
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return len(s) > 0
}
