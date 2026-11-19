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
)

// rackLists returns the "Rack contents" blocks of a rendered page in order.
func rackLists(t *testing.T, out string) []string {
	t.Helper()
	var blocks []string
	for _, part := range strings.Split(out, `<details class="rack-list">`)[1:] {
		end := strings.Index(part, "</details>")
		if end < 0 {
			t.Fatalf("rack-list block without closing tag")
		}
		blocks = append(blocks, part[:end])
	}
	return blocks
}

var rackListRow = regexp.MustCompile(`<tr( class="rl-child")?>\s*<th scope="row" class="mono">([^<]*)</th>\s*<td class="rl-dev"><span class="rl-name">([^<]*)</span>`)

func TestRenderRackHTMLRailsAndSlotsAreDecorative(t *testing.T) {
	out := renderHTML(t, "")
	// Rack-A has 10 U, Rack-B 4 U: 14 rows of two rails and one slot each.
	for _, want := range []string{
		`<span class="u rail-l" style="--row:\d+" aria-hidden="true">\d+</span>`,
		`<span class="slot" style="--row:\d+" aria-hidden="true"></span>`,
		`<span class="u rail-r" style="--row:\d+" aria-hidden="true">\d+</span>`,
	} {
		if n := len(regexp.MustCompile(want).FindAllString(out, -1)); n != 14 {
			t.Errorf("expected 14 matches of %s, got %d", want, n)
		}
	}
	for _, m := range regexp.MustCompile(`<([a-z]+)[^>]*\saria-hidden="true"`).FindAllStringSubmatch(out, -1) {
		if m[1] != "span" {
			t.Errorf("aria-hidden must only hide decorative spans, found <%s>", m[1])
		}
	}
}

func TestRenderRackHTMLRackListFollowsTheElevation(t *testing.T) {
	out := renderHTML(t, "Rack-A")
	lists := rackLists(t, out)
	if len(lists) != 1 {
		t.Fatalf("expected one rack-list for one rack, got %d", len(lists))
	}
	list := lists[0]
	if strings.Contains(out, `<details class="rack-list" open`) {
		t.Errorf("the rack list must start collapsed")
	}
	for _, want := range []string{
		`<caption class="vh">Contents of Rack-A, top of rack first</caption>`,
		`<span class="rl-meta">3 devices, 1 chassis node, top first</span>`,
		`<span class="rl-kind">Chassis</span> <span class="rl-status">Active</span>`,
		`<span class="rl-kind">Switch</span> <span class="rl-status">Offline</span>`,
		`<span class="rl-model">HPE DL380</span>`,
		`<span class="rl-name">cn001</span> <span class="rl-model">XL225n</span><span class="vh">, in apollo-chassis</span>`,
	} {
		if !strings.Contains(list, want) {
			t.Errorf("expected the rack list to contain %q", want)
		}
	}
	if n := strings.Count(list, `<th scope="col">`); n != 4 {
		t.Errorf("expected 4 column headers, got %d", n)
	}
	var got []string
	for _, m := range rackListRow.FindAllStringSubmatch(list, -1) {
		got = append(got, strings.TrimSpace(m[1]+" "+m[2]+" "+m[3]))
	}
	want := []string{"U8-U9 apollo-chassis", `class="rl-child" U8-U9 cn001`, "U7 leaf-sw", "U3-U4 node-2u"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("rows must follow the rack top first with chassis nodes after their chassis:\n got %q\nwant %q", got, want)
	}
}

func TestRenderRackHTMLRackListManagementLinks(t *testing.T) {
	list := rackLists(t, renderCabling(t))[0]
	if n := len(rackListRow.FindAllString(list, -1)); n != 9 {
		t.Errorf("expected 8 device rows and 1 chassis node row, got %d", n)
	}
	for _, want := range []string{
		`<a class="rl-mgmt" href="https://10.0.84.5/" target="_blank" rel="noopener noreferrer">iLO<span class="vh"> 10.0.84.5</span></a>`,
		`<a class="rl-mgmt" href="http://10.0.84.14:8000/" target="_blank" rel="noopener noreferrer">BMC<span class="vh"> 10.0.84.14:8000</span></a>`,
		`<a class="rl-mgmt" href="https://10.0.83.240/" target="_blank" rel="noopener noreferrer">Web UI<span class="vh"> 10.0.83.240</span></a>`,
		`<td><span class="weak">&#8212;</span></td>`,
	} {
		if !strings.Contains(list, want) {
			t.Errorf("expected the rack list to contain %q", want)
		}
	}
	if strings.Contains(list, "javascript:") {
		t.Errorf("a hostile bmc_url must not reach the rack list")
	}
	if strings.Contains(list, `evil"]<`) || !strings.Contains(list, `evil&#34;]&lt;`) {
		t.Errorf("device names must be escaped in the rack list")
	}
}

func TestRenderRackHTMLRackListEscapesNames(t *testing.T) {
	list := rackLists(t, renderHTML(t, "Rack-B"))[0]
	if !strings.Contains(list, `<span class="rl-name">&lt;script&gt;alert(1)&lt;/script&gt;</span>`) {
		t.Errorf("expected the escaped device name in the rack list")
	}
	if !strings.Contains(list, `<span class="rl-model">Evil &amp; Co</span>`) {
		t.Errorf("expected the escaped model in the rack list")
	}
	if regexp.MustCompile(`(?i)<script`).MatchString(list) || regexp.MustCompile(`(?i)\son[a-z]+=`).MatchString(list) {
		t.Errorf("the rack list must ship without scripts or inline event handlers")
	}
}
