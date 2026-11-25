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
package visual

import (
	"sort"
	"strconv"
	"strings"
)

// formatBytes renders a byte count with three significant digits: binary
// units for memory (32 GiB), decimal units for drives (1.92 TB).
func formatBytes(n float64, binary bool) string {
	if n <= 0 {
		return ""
	}
	step, units := 1000.0, []string{"B", "kB", "MB", "GB", "TB", "PB"}
	if binary {
		step, units = 1024.0, []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}
	}
	i := 0
	for n >= step && i < len(units)-1 {
		n /= step
		i++
	}
	decimals := 2
	if n >= 100 {
		decimals = 0
	} else if n >= 10 {
		decimals = 1
	}
	s := strconv.FormatFloat(n, 'f', decimals, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return s + " " + units[i]
}

// joinUnique joins the non-empty parts with spaces, dropping a part that is
// already contained in a kept one and replacing kept parts that a later part
// contains ("32 GiB" + "32 GiB DDR4" -> "32 GiB DDR4").
func joinUnique(parts ...string) string {
	var kept []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || containsFold(kept, part) {
			continue
		}
		var next []string
		for _, k := range kept {
			if !strings.Contains(strings.ToLower(part), strings.ToLower(k)) {
				next = append(next, k)
			}
		}
		kept = append(next, part)
	}
	return strings.Join(kept, " ")
}

// containsFold reports whether any kept part contains s, ignoring case.
func containsFold(kept []string, s string) bool {
	for _, k := range kept {
		if strings.Contains(strings.ToLower(k), strings.ToLower(s)) {
			return true
		}
	}
	return false
}

// orElse returns a when it is non-empty, else b.
func orElse(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// sortedKeys returns the keys of m in sorted order.
func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
