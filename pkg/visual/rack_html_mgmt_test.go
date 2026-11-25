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

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

func TestMgmtLinkFromILOInterface(t *testing.T) {
	inv := cablingInventory()
	link := mgmtLink(newHTMLIndex(inv), cablingDevice(t, inv, cabServerID))
	if link == nil || *link != (htmlMgmt{Label: "iLO", Host: "10.0.84.5", URL: "https://10.0.84.5/"}) {
		t.Errorf("iLO link = %+v", link)
	}
}

func TestMgmtLinkFromSwitchMgmtInterface(t *testing.T) {
	inv := cablingInventory()
	link := mgmtLink(newHTMLIndex(inv), cablingDevice(t, inv, cab6300ID))
	if link == nil || *link != (htmlMgmt{Label: "Web UI", Host: "10.0.83.240", URL: "https://10.0.83.240/"}) {
		t.Errorf("switch link = %+v", link)
	}
	if link := mgmtLink(newHTMLIndex(inv), cablingDevice(t, inv, cabIdleID)); link != nil {
		t.Errorf("a switch without a management address has no link, got %+v", link)
	}
}

func TestMgmtLinkOverrideWins(t *testing.T) {
	inv := cablingInventory()
	link := mgmtLink(newHTMLIndex(inv), cablingDevice(t, inv, cabBMCID))
	if link == nil || *link != (htmlMgmt{Label: "BMC", Host: "10.0.84.14:8000", URL: "http://10.0.84.14:8000/"}) {
		t.Errorf("bmc_url override = %+v", link)
	}
}

func TestMgmtLinkDropsHostileOverrides(t *testing.T) {
	inv := cablingInventory()
	idx := newHTMLIndex(inv)
	evil := cablingDevice(t, inv, cabEvilBMCID)
	if link := mgmtLink(idx, evil); link == nil || link.URL != "https://10.0.84.70/" {
		t.Errorf("javascript: override must be dropped in favour of the interface address, got %+v", link)
	}
	for _, raw := range []string{"ftp://10.0.0.1/", "//evil.example/", "10.0.84.14", "https://", "data:text/html,hi", " "} {
		dev := &devicetypes.CaniDeviceType{Type: devicetypes.TypeNode, ObjectMeta: devicetypes.ObjectMeta{CustomFields: map[string]any{"bmc_url": raw}}}
		if link := bmcURLOverride(dev); link != nil {
			t.Errorf("bmc_url %q must be rejected, got %+v", raw, link)
		}
	}
	if link := bmcURLOverride(&devicetypes.CaniDeviceType{ObjectMeta: devicetypes.ObjectMeta{CustomFields: map[string]any{"bmc_url": 42}}}); link != nil {
		t.Errorf("non-string bmc_url must be ignored")
	}
}

func TestIPURLOnlyFromParsedAddresses(t *testing.T) {
	cases := map[string]string{
		"10.6.84.81":           "https://10.6.84.81/",
		"fd00::1":              "https://[fd00::1]/",
		"fe80::1%en0":          "https://[fe80::1]/",
		"10.6.84.81/24":        "",
		"evil.example":         "",
		`10.6.84.81"><script>`: "",
		"javascript:alert(1)":  "",
		"":                     "",
	}
	for host, want := range cases {
		if got := ipURL(host); got != want {
			t.Errorf("ipURL(%q) = %q, want %q", host, got, want)
		}
	}
}

func TestRenderedManagementLinks(t *testing.T) {
	out := renderCabling(t)
	header := `<a class="mgmt" href="https://10.0.84.5/" target="_blank" rel="noopener noreferrer"><span class="mgmt-label">iLO</span> <span class="mono">10.0.84.5</span>`
	if !strings.Contains(out, header) {
		t.Errorf("expected the iLO link in the server card header")
	}
	if !strings.Contains(out, `href="http://10.0.84.14:8000/" target="_blank" rel="noopener noreferrer"><span class="mgmt-label">BMC</span>`) {
		t.Errorf("expected the bmc_url override link")
	}
	if strings.Contains(strings.ToLower(out), "javascript:") {
		t.Errorf("hostile bmc_url leaked into the page")
	}
	for _, m := range regexp.MustCompile(`<a [^>]*href="([^"]*)"[^>]*>`).FindAllStringSubmatch(out, -1) {
		href := m[1]
		if !strings.HasPrefix(href, "https://") && !strings.HasPrefix(href, "http://") && !strings.HasPrefix(href, "#sw-") {
			t.Errorf("unexpected link target %q", href)
		}
		if strings.HasPrefix(href, "http") && !strings.Contains(m[0], `rel="noopener noreferrer"`) {
			t.Errorf("external link without rel=noopener noreferrer: %s", m[0])
		}
	}
	// The switch table links the server's iLO address; the chassis card lists its node with the node's iLO.
	if !strings.Contains(out, `<td class="mono"><a href="https://10.0.84.5/" target="_blank" rel="noopener noreferrer">10.0.84.5</a></td>`) {
		t.Errorf("expected the 6300 table row to link the peer iLO address")
	}
	if !strings.Contains(out, `<li><button type="button" class="peer" popovertarget="p-`+cabChildID.String()+`" popovertargetaction="show">cn1</button><a class="mgmt" href="https://10.0.84.9/"`) {
		t.Errorf("expected the chassis card to list the node button with its iLO link")
	}
}
