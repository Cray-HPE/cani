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
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

var (
	htmlRackAID    = uuid.MustParse("00000000-0000-0000-0007-000000000001")
	htmlRackBID    = uuid.MustParse("00000000-0000-0000-0007-000000000002")
	htmlLocationID = uuid.MustParse("00000000-0000-0000-0007-000000000003")
	htmlNode2UID   = uuid.MustParse("00000000-0000-0000-0008-000000000001")
	htmlSwitchID   = uuid.MustParse("00000000-0000-0000-0008-000000000002")
	htmlChassisID  = uuid.MustParse("00000000-0000-0000-0008-000000000003")
	htmlChildID    = uuid.MustParse("00000000-0000-0000-0008-000000000004")
	htmlRackBDevID = uuid.MustParse("00000000-0000-0000-0008-000000000005")
	htmlIPID       = uuid.MustParse("00000000-0000-0000-0009-000000000001")
)

// htmlRackInventory builds two racks: Rack-A (10U) holds a 2U node at U3,
// a 1U switch at U7 and a 2U chassis at U8 with one child node; Rack-B (4U)
// holds a single device whose name carries a script tag.
func htmlRackInventory() *devicetypes.Inventory {
	inv := devicetypes.NewInventory()
	inv.Locations[htmlLocationID] = &devicetypes.CaniLocationType{ID: htmlLocationID, Name: "Lab 2"}
	inv.Racks[htmlRackAID] = &devicetypes.CaniRackType{
		ID: htmlRackAID, Name: "Rack-A", UHeight: 10, Location: htmlLocationID,
		Devices: []uuid.UUID{htmlNode2UID, htmlSwitchID, htmlChassisID},
	}
	inv.Racks[htmlRackBID] = &devicetypes.CaniRackType{
		ID: htmlRackBID, Name: "Rack-B", UHeight: 4, Location: htmlLocationID,
		Devices: []uuid.UUID{htmlRackBDevID},
	}
	inv.IPAddresses[htmlIPID] = &devicetypes.CaniIPAddress{ID: htmlIPID, Host: "10.0.0.5", Address: "10.0.0.5/24"}
	inv.Devices[htmlNode2UID] = &devicetypes.CaniDeviceType{
		ID: htmlNode2UID, Name: "node-2u", Model: "HPE DL380", Type: devicetypes.TypeNode,
		UHeight: 2, IsFullDepth: true, RackPosition: 3, Parent: htmlRackAID, PrimaryIPv4: htmlIPID,
		ObjectMeta: devicetypes.ObjectMeta{Status: "Active", Tags: []string{"verified", "reachable"}},
	}
	inv.Devices[htmlSwitchID] = &devicetypes.CaniDeviceType{
		ID: htmlSwitchID, Name: "leaf-sw", Model: "Aruba 8325", Type: devicetypes.TypeSwitch,
		UHeight: 1, RackPosition: 7, Parent: htmlRackAID,
		ObjectMeta: devicetypes.ObjectMeta{Status: "Offline", Tags: []string{"not-ours"}},
	}
	inv.Devices[htmlChassisID] = &devicetypes.CaniDeviceType{
		ID: htmlChassisID, Name: "apollo-chassis", Model: "Apollo 2000", Type: devicetypes.TypeChassis,
		UHeight: 2, IsFullDepth: true, RackPosition: 8, Parent: htmlRackAID, Children: []uuid.UUID{htmlChildID},
		ObjectMeta: devicetypes.ObjectMeta{Status: "Active"},
	}
	inv.Devices[htmlChildID] = &devicetypes.CaniDeviceType{
		ID: htmlChildID, Name: "cn001", Model: "XL225n", Type: devicetypes.TypeNode, Parent: htmlChassisID,
		ObjectMeta: devicetypes.ObjectMeta{Status: "Offline", Tags: []string{"unverified"}},
	}
	inv.Devices[htmlRackBDevID] = &devicetypes.CaniDeviceType{
		ID: htmlRackBDevID, Name: "<script>alert(1)</script>", Model: "Evil & Co", Type: devicetypes.TypeNode,
		UHeight: 1, RackPosition: 2, Parent: htmlRackBID,
		ObjectMeta: devicetypes.ObjectMeta{Status: "Active"},
	}
	return inv
}

func renderHTML(t *testing.T, filter string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := RenderRackHTML(&buf, htmlRackInventory(), CompactRenderOptions{RackFilter: filter}); err != nil {
		t.Fatalf("RenderRackHTML: %v", err)
	}
	return buf.String()
}

func TestRenderRackHTMLListsDevicesAndChildren(t *testing.T) {
	out := renderHTML(t, "")
	for _, want := range []string{"node-2u", "leaf-sw", "apollo-chassis", "cn001", "Rack-A", "Rack-B", "Lab 2", "10.0.0.5/24"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q", want)
		}
	}
	if !strings.Contains(out, `class="bay kind-server status-offline tag-unverified half"`) {
		t.Errorf("expected the chassis child to be rendered as a bay cell with its own status and tag classes")
	}
}

func TestRenderRackHTMLDeviceSpansUpwardFromRackPosition(t *testing.T) {
	out := renderHTML(t, "Rack-A")
	// 2U device at U3 occupies U3-U4; in a 10U rack U4 is grid row 7 (U10 = row 1).
	want := `style="--row:7;--span:2" data-u-start="3" data-u-end="4"`
	if !strings.Contains(out, want) {
		t.Errorf("expected 2U device at position 3 to render %q", want)
	}
	// 1U switch at U7 is grid row 4 and spans one row.
	if !strings.Contains(out, `style="--row:4;--span:1" data-u-start="7" data-u-end="7"`) {
		t.Errorf("expected 1U device at position 7 to render at row 4")
	}
	if !strings.Contains(out, `data-u-start="8" data-u-end="9"`) {
		t.Errorf("expected 2U chassis at position 8 to occupy U8-U9")
	}
}

func TestRenderRackHTMLEscapesDeviceNames(t *testing.T) {
	out := renderHTML(t, "")
	if strings.Contains(out, "<script") {
		t.Errorf("device name must be HTML-escaped; found a raw <script tag")
	}
	if !strings.Contains(out, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Errorf("expected the escaped device name in the output")
	}
	if !strings.Contains(out, "Evil &amp; Co") {
		t.Errorf("expected the escaped model in the output")
	}
}

func TestRenderRackHTMLHasNoScriptsOrEventHandlers(t *testing.T) {
	out := renderHTML(t, "")
	if regexp.MustCompile(`(?i)<script`).MatchString(out) {
		t.Errorf("the page must ship without <script> tags")
	}
	if m := regexp.MustCompile(`(?i)\son[a-z]+=`).FindString(out); m != "" {
		t.Errorf("the page must ship without inline event handlers; found %q", m)
	}
	if !strings.Contains(out, `popovertarget="p-`+htmlChildID.String()+`"`) {
		t.Errorf("expected a declarative popover invoker for the chassis child")
	}
}

func TestRenderRackHTMLRackFilter(t *testing.T) {
	out := renderHTML(t, "Rack-B")
	if strings.Contains(out, "Rack-A") || strings.Contains(out, "node-2u") {
		t.Errorf("rack filter must exclude Rack-A and its devices")
	}
	if !strings.Contains(out, "<h2>Rack-B</h2>") {
		t.Errorf("rack filter must keep Rack-B")
	}
}

func TestRenderRackHTMLNilInventory(t *testing.T) {
	if err := RenderRackHTML(&bytes.Buffer{}, nil, CompactRenderOptions{}); err == nil {
		t.Errorf("expected an error for a nil inventory")
	}
}

func TestRenderRackHTMLEscapesInterfaceNamesAndLabels(t *testing.T) {
	out := renderCabling(t)
	if strings.Contains(out, "<b>x</b>") {
		t.Errorf("interface label must be escaped")
	}
	if !strings.Contains(out, "&lt;b&gt;x&lt;/b&gt;") {
		t.Errorf("expected the escaped interface label in a cabling table")
	}
	if strings.Contains(out, `evil"]<`) || !strings.Contains(out, "evil&#34;]&lt;") {
		t.Errorf("device name with quote, bracket and angle must be escaped everywhere")
	}
	if regexp.MustCompile(`(?i)<script`).MatchString(out) || regexp.MustCompile(`(?i)\son[a-z]+=`).MatchString(out) {
		t.Errorf("the page must ship without scripts or inline event handlers")
	}
}

func TestRenderRackHTMLCablingChipAndDrawer(t *testing.T) {
	out := renderCabling(t)
	for _, want := range []string{
		`<input type="checkbox" id="show-cabling" checked>`,
		`body:has(#show-cabling:not(:checked)) .cabling { display: none; }`,
		`[popover] { display: none !important; transition: none; }`,
		`1 rack, 8 devices, 4 cables`,
		`<section class="cabling" aria-labelledby="cabling-title">`,
		`role="dialog" aria-labelledby="h-` + cabServerID.String() + `" tabindex="-1" autofocus`,
		`<p class="note"><span class="note-label">Note</span>Issue: fan 3 reports degraded.</p>`,
		`<p class="note warn"><span class="note-label">Unverified · Note</span>Not verified: no BMC answered.</p>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected page to contain %q", want)
		}
	}
	if strings.Contains(out, "anchor-name") || strings.Contains(out, "position-anchor") {
		t.Errorf("the drawer replaces anchored cards; no anchor positioning should remain")
	}
}

func TestRenderRackHTMLDataDevAttributesAreUUIDs(t *testing.T) {
	out := renderCabling(t)
	matches := regexp.MustCompile(`\sdata-dev="([^"]*)"`).FindAllStringSubmatch(out, -1)
	if len(matches) != 16 {
		t.Fatalf("expected data-dev on 8 tiles, 1 bay, 2 switch tables and 5 cabling rows, got %d", len(matches))
	}
	for _, m := range matches {
		if u, err := uuid.Parse(m[1]); err != nil || u.String() != m[1] {
			t.Errorf("data-dev %q is not a canonical UUID", m[1])
		}
	}
	if !strings.Contains(out, `class="bay kind-server status-active half" popovertarget="p-`+cabChildID.String()+`" data-dev="`+cabChildID.String()+`"`) {
		t.Errorf("expected the chassis child cell to carry its data-dev")
	}
}

func TestRenderRackHTMLCablingRowsCarryTheirPeer(t *testing.T) {
	out := renderCabling(t)
	for _, id := range []string{cab6300ID.String(), cab8325ID.String()} {
		if !strings.Contains(out, `<details open id="sw-`+id+`" class="swt" data-dev="`+id+`">`) {
			t.Errorf("switch table %s must carry the switch UUID", id)
		}
	}
	rows := map[string]int{}
	for _, m := range regexp.MustCompile(`<tr data-dev="([^"]*)">`).FindAllStringSubmatch(out, -1) {
		rows[m[1]]++
	}
	// iLO and NIC rows point at the server, the trunk row of each switch at
	// the other switch, the eth0 row at the hostile device.
	want := map[string]int{cabServerID.String(): 2, cab6300ID.String(): 1, cab8325ID.String(): 1, cabEvilID.String(): 1}
	for id, n := range want {
		if rows[id] != n {
			t.Errorf("expected %d rows with data-dev=%s, got %d", n, id, rows[id])
		}
	}
	if len(rows) != len(want) {
		t.Errorf("unexpected row peers: %v", rows)
	}
	if strings.Contains(out, `<tr data-dev="">`) {
		t.Errorf("rows without a resolvable peer must not carry an empty data-dev")
	}
}

func TestRenderRackHTMLGuideAndPane(t *testing.T) {
	out := renderCabling(t)
	for _, want := range []string{
		`<aside class="guide" aria-labelledby="guide-title">`,
		`<h2 id="guide-title">Click a device to open its card</h2>`,
		`@media (min-width: 82rem)`,
		`grid-template-columns: minmax(0, 1fr) var(--drawer-w)`,
		`.swt > summary::after { content: var(--hl-badge, none);`,
		`--hl-self: initial;`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected page to contain %q", want)
		}
	}
	if strings.Count(out, `<aside class="guide"`) != 1 {
		t.Errorf("the guide must be rendered exactly once")
	}
}
