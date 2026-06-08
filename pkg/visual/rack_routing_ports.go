/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
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
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// abbreviatePort shortens well-known port names for compact display.
//
//	"HSN 0"  → "h0"    "MGMT 0" → "M0"    "mgmt0" → "m0"
//	"mgmt1"  → "m1"    "iLO"    → "iL"    "1/1/3"  → as-is
//	bare int → as-is
func abbreviatePort(port string) string {
	low := strings.ToLower(port)
	switch {
	case strings.HasPrefix(low, "hsn "):
		return "h" + port[4:]
	case strings.HasPrefix(low, "mgmt "):
		return strings.ToUpper(port[:1]) + port[5:]
	case strings.HasPrefix(low, "mgmt"):
		return "m" + port[4:]
	case low == "ilo":
		return "iL"
	}
	return port
}

// abbreviateRack shortens a rack name for inline display.
// Keeps the last 5 chars (e.g. "x3516") to preserve the x-prefix.
func abbreviateRack(name string) string {
	if len(name) <= 5 {
		return name
	}
	return name[len(name)-5:]
}

// portPair is a local:remote port mapping.
type portPair struct {
	local  string
	remote string
}

// parsedPort holds prefix+number decomposition for range compression.
type parsedPort struct {
	lPre string
	lNum int
	rPre string
	rNum int
	ok   bool
}

// compressPortPairs tries to compress consecutive port ranges.
// e.g. h0:1 h1:2 h2:3 h3:4 → "h0-3:1-4"
// Falls back to "local:remote,local:remote" when ranges don't compress.
func compressPortPairs(pairs []portPair) string {
	if len(pairs) == 0 {
		return ""
	}
	if len(pairs) == 1 {
		return pairs[0].local + ":" + pairs[0].remote
	}

	// Try range compression: split each port into prefix+number.
	pp := make([]parsedPort, len(pairs))
	allParsed := true
	for i, p := range pairs {
		lPre, lNum, lok := splitPortNum(p.local)
		rPre, rNum, rok := splitPortNum(p.remote)
		pp[i] = parsedPort{lPre, lNum, rPre, rNum, lok && rok}
		if !pp[i].ok {
			allParsed = false
		}
	}

	if allParsed {
		// Sort by local number for range detection.
		sort.Slice(pp, func(i, j int) bool { return pp[i].lNum < pp[j].lNum })
		if runs := findRuns(pp); runs != "" {
			return runs
		}
	}

	// Fallback: comma-separated pairs.
	items := make([]string, len(pairs))
	for i, p := range pairs {
		items[i] = p.local + ":" + p.remote
	}
	return strings.Join(items, ",")
}

// splitPortNum splits "h3" → ("h", 3, true), "49" → ("", 49, true).
func splitPortNum(port string) (string, int, bool) {
	// Find where trailing digits start.
	i := len(port)
	for i > 0 && port[i-1] >= '0' && port[i-1] <= '9' {
		i--
	}
	if i == len(port) {
		return port, 0, false // no trailing number
	}
	n, err := strconv.Atoi(port[i:])
	if err != nil {
		return port, 0, false
	}
	return port[:i], n, true
}

// findRuns detects consecutive runs in sorted parsed pairs and formats
// them as compressed ranges. Returns "" if no clean runs are found.
func findRuns(pp []parsedPort) string {
	if len(pp) == 0 {
		return ""
	}
	// Check all share the same prefix pair.
	lPre, rPre := pp[0].lPre, pp[0].rPre
	for _, p := range pp {
		if p.lPre != lPre || p.rPre != rPre {
			return "" // mixed prefixes, can't compress
		}
	}
	// Check local numbers are consecutive.
	for i := 1; i < len(pp); i++ {
		if pp[i].lNum != pp[i-1].lNum+1 {
			return "" // gap in local sequence
		}
	}
	// Build range string.
	first, last := pp[0], pp[len(pp)-1]
	lRange := formatRange(lPre, first.lNum, last.lNum)
	rRange := formatRange(rPre, first.rNum, last.rNum)
	return lRange + ":" + rRange
}

// formatRange formats a prefix+number range: "h", 0, 3 → "h0-3".
func formatRange(prefix string, lo, hi int) string {
	if lo == hi {
		return fmt.Sprintf("%s%d", prefix, lo)
	}
	return fmt.Sprintf("%s%d-%d", prefix, lo, hi)
}
