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
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHighlightCSSHasABlockForBothEndsOfACable(t *testing.T) {
	page := buildHTMLPage(cablingInventory(), "", time.Unix(0, 0))
	css := string(page.PeerCSS)
	server := "body:has(#p-" + cabServerID.String() + ":popover-open) {\n"
	if !strings.Contains(css, server) {
		t.Fatalf("expected a highlight block for the server, got:\n%s", css)
	}
	block := css[strings.Index(css, server):]
	block = block[:strings.Index(block, "\n}\n")+3]
	for _, want := range []string{
		`& [data-dev="` + cabServerID.String() + `"][data-dev] { --hl: solid; --hl-self: 1; }`,
		`[data-dev="` + cab6300ID.String() + `"]`,
		`[data-dev="` + cab8325ID.String() + `"]`,
		`{ --hl: dashed; --hl-badge: "link"; }`,
	} {
		if !strings.Contains(block, want) {
			t.Errorf("server block lacks %q:\n%s", want, block)
		}
	}
	if strings.Contains(block, `[data-dev="`+cabEvilID.String()+`"]`) {
		t.Errorf("the server is not cabled to the hostile device")
	}
	switchBlock := "body:has(#p-" + cab6300ID.String() + ":popover-open) {"
	if !strings.Contains(css, switchBlock) {
		t.Errorf("expected a block for the switch end as well")
	}
	idle := "body:has(#p-" + cabIdleID.String() + ":popover-open) {\n  & [data-dev=\"" + cabIdleID.String() + "\"][data-dev] { --hl: solid; --hl-self: 1; }\n}\n"
	if !strings.Contains(css, idle) {
		t.Errorf("a device without cables still outlines itself, and only itself:\n%s", css)
	}
}

func TestHighlightCSSLightsUpRowsOnlyForTheSubject(t *testing.T) {
	css := string(highlightCSS([]htmlDevice{{ID: cabServerID.String(), PeerIDs: []string{cab6300ID.String()}}}))
	subject, peers, found := strings.Cut(css, "\n  & :is(")
	if !found {
		t.Fatalf("expected a subject line followed by a peer line:\n%s", css)
	}
	if !strings.HasSuffix(subject, `{ --hl: solid; --hl-self: 1; }`) {
		t.Errorf("rows and the own table light up through --hl-self on the subject line, got:\n%s", subject)
	}
	if strings.Contains(peers, "--hl-self") {
		t.Errorf("peer rows must not light up; the peer line only outlines and badges:\n%s", peers)
	}
}

func TestHighlightCSSIsWhitelistedAndUUIDOnly(t *testing.T) {
	page := buildHTMLPage(cablingInventory(), "", time.Unix(0, 0))
	css := string(page.PeerCSS)
	if !regexp.MustCompile(`^[-a-z0-9 :#.\[\]="(),&{};\n]*$`).MatchString(css) {
		t.Errorf("generated CSS contains characters outside the whitelist:\n%s", css)
	}
	ids := regexp.MustCompile(`#p-([^:]+):popover-open|data-dev="([^"]+)"`).FindAllStringSubmatch(css, -1)
	if len(ids) == 0 {
		t.Fatal("expected ids in the generated CSS")
	}
	for _, m := range ids {
		id := m[1] + m[2]
		if u, err := uuid.Parse(id); err != nil || u.String() != id {
			t.Errorf("selector id %q is not a canonical UUID", id)
		}
	}
	if strings.Contains(css, "evil") || strings.Contains(css, "<") {
		t.Errorf("device names must never reach the generated CSS")
	}
}

func TestHighlightCSSSkipsInvalidIDs(t *testing.T) {
	devices := []htmlDevice{
		{ID: "not-a-uuid", PeerIDs: []string{cab6300ID.String()}},
		{ID: cabServerID.String(), PeerIDs: []string{"<script>", cab6300ID.String()}},
	}
	css := string(highlightCSS(devices))
	if strings.Contains(css, "not-a-uuid") || strings.Contains(css, "<script>") {
		t.Errorf("invalid ids leaked into the CSS:\n%s", css)
	}
	if strings.Count(css, "body:has(") != 1 {
		t.Errorf("expected exactly one block, got:\n%s", css)
	}
}
