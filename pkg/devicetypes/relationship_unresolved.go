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
package devicetypes

import (
	"log"
	"slices"
)

// maxUnresolvedLogged caps the unresolved conflicts logged without --debug,
// so a legacy inventory with many duplicates does not bury command output.
const maxUnresolvedLogged = 10

// LogUnresolved logs the conflicts the operator must resolve, in a stable
// order: all of them with --debug, otherwise the first maxUnresolvedLogged
// and a count of the rest.
func (r *RelationshipResult) LogUnresolved() {
	if r == nil {
		return
	}
	items := slices.Sorted(slices.Values(r.Unresolved))
	shown := items
	if !Debug && len(shown) > maxUnresolvedLogged {
		shown = shown[:maxUnresolvedLogged]
	}
	for _, item := range shown {
		log.Printf("Warning: %s", item)
	}
	if hidden := len(items) - len(shown); hidden > 0 {
		log.Printf("Warning: %d more conflicts to resolve; enable debug output to list them", hidden)
	}
}
