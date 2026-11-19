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
	"html/template"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// peersOf returns the sorted, de-duplicated UUIDs of the devices cabled to dev.
func (idx *htmlIndex) peersOf(dev uuid.UUID) []string {
	seen := map[string]bool{}
	for local, far := range idx.cables {
		if local.dev == dev && far.dev != dev {
			seen[far.dev.String()] = true
		}
	}
	peers := make([]string, 0, len(seen))
	for id := range seen {
		peers = append(peers, id)
	}
	sort.Strings(peers)
	return peers
}

// highlightCSS emits one rule block per device: while its card is open, the
// device's own tile gets a solid outline and its cable peers, if any, a
// dashed one plus the "link" badge. The same block lights up the cabling
// rows that lead to the device (and frames its own cabling table) through
// --hl-self, because rows and tables carry data-dev too. The static
// stylesheet draws from the custom properties; only UUIDs that parse are
// written, so no escaping is needed.
func highlightCSS(devices []htmlDevice) template.CSS {
	var b strings.Builder
	for _, dev := range devices {
		writeHighlightRule(&b, dev)
		for _, child := range dev.Children {
			writeHighlightRule(&b, child)
		}
	}
	return template.CSS(b.String())
}

// writeHighlightRule appends the rule block of one device.
// The subject selector repeats the attribute so that, with two cards stacked,
// a device that is both a subject and a peer keeps the solid outline.
func writeHighlightRule(b *strings.Builder, dev htmlDevice) {
	if !isUUID(dev.ID) {
		return
	}
	b.WriteString("body:has(#p-" + dev.ID + ":popover-open) {\n")
	b.WriteString(`  & [data-dev="` + dev.ID + `"][data-dev] { --hl: solid; --hl-self: 1; }` + "\n")
	if peers := validUUIDs(dev.PeerIDs); len(peers) > 0 {
		selectors := make([]string, len(peers))
		for i, id := range peers {
			selectors[i] = `[data-dev="` + id + `"]`
		}
		b.WriteString("  & :is(" + strings.Join(selectors, ", ") + `) { --hl: dashed; --hl-badge: "link"; }` + "\n")
	}
	b.WriteString("}\n")
}

// validUUIDs keeps the ids that parse as UUIDs, in canonical lowercase form.
func validUUIDs(ids []string) []string {
	var out []string
	for _, id := range ids {
		if u, err := uuid.Parse(id); err == nil {
			out = append(out, u.String())
		}
	}
	return out
}

func isUUID(id string) bool {
	u, err := uuid.Parse(id)
	return err == nil && u.String() == id
}
