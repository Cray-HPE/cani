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
	"net/netip"
	"net/url"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// bmcURLField is the device custom field that overrides the derived
// management link, e.g. a BMC web UI on a non-default port.
const bmcURLField = "bmc_url"

// mgmtLink resolves the management-UI link of a device. First match wins:
// the bmc_url custom field, an interface named iLO with a bound IP, any
// other management-only interface with a bound IP. Every emitted URL is
// rebuilt from a parsed address or a parsed http(s) URL, never copied raw.
func mgmtLink(idx *htmlIndex, dev *devicetypes.CaniDeviceType) *htmlMgmt {
	if link := bmcURLOverride(dev); link != nil {
		return link
	}
	if link := interfaceMgmt(idx, dev, isILO, "iLO"); link != nil {
		return link
	}
	return interfaceMgmt(idx, dev, isMgmtPort, mgmtLabel(dev.Type))
}

// bmcURLOverride validates the bmc_url custom field as an http(s) URL.
func bmcURLOverride(dev *devicetypes.CaniDeviceType) *htmlMgmt {
	raw := cfString(dev.CustomFields, bmcURLField)
	if raw == "" {
		return nil
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil
	}
	return &htmlMgmt{Label: mgmtLabel(dev.Type), Host: u.Host, URL: u.String()}
}

// interfaceMgmt builds a link from the first interface matching pick that has an IP.
func interfaceMgmt(idx *htmlIndex, dev *devicetypes.CaniDeviceType, pick func(*devicetypes.InterfaceSpec) bool, label string) *htmlMgmt {
	for i := range dev.Interfaces {
		spec := &dev.Interfaces[i]
		if !pick(spec) {
			continue
		}
		if ip := idx.ipOf(spec.ID); ip != nil {
			if u := ipURL(ipHost(ip)); u != "" {
				return &htmlMgmt{Label: label, Host: ipHost(ip), URL: u}
			}
		}
	}
	return nil
}

// mgmtURL is the link behind a port's IP when the port is a management port:
// the device override if any, else https:// on the address itself.
func (idx *htmlIndex) mgmtURL(dev *devicetypes.CaniDeviceType, spec *devicetypes.InterfaceSpec, ip *devicetypes.CaniIPAddress) string {
	if !isMgmtPort(spec) {
		return ""
	}
	if link := bmcURLOverride(dev); link != nil {
		return link.URL
	}
	return ipURL(ipHost(ip))
}

// isMgmtPort recognises a management port by its mgmt_only flag or, when the
// flag was never recorded, by the conventional names (iLO, mgmt, Management,
// BMC) that aruba-config relies on too.
func isMgmtPort(spec *devicetypes.InterfaceSpec) bool {
	if isMgmtOnly(spec) || isILO(spec) {
		return true
	}
	switch strings.ToLower(spec.Name) {
	case "mgmt", "management", "bmc":
		return true
	default:
		return false
	}
}

// ipURL turns a validated IP address into https://<ip>/; anything that is
// not an address yields "".
func ipURL(host string) string {
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return ""
	}
	if addr.Is6() {
		return "https://[" + addr.WithZone("").String() + "]/"
	}
	return "https://" + addr.String() + "/"
}

// isILO matches the HPE iLO interface by name.
func isILO(spec *devicetypes.InterfaceSpec) bool {
	return strings.EqualFold(spec.Name, "iLO")
}

// mgmtLabel names the management UI of a device kind: switches expose a web
// UI, everything else a BMC.
func mgmtLabel(t devicetypes.Type) string {
	switch t {
	case devicetypes.TypeSwitch, devicetypes.TypeMgmtSwitch, devicetypes.TypeHsnSwitch:
		return "Web UI"
	default:
		return "BMC"
	}
}

// cfString reads a string custom field; missing or non-string values are "".
func cfString(fields map[string]any, key string) string {
	s, _ := fields[key].(string)
	return s
}

// cfNumber reads a numeric custom field; JSON decoding yields float64 while
// in-process writers may store ints.
func cfNumber(fields map[string]any, key string) (float64, bool) {
	switch v := fields[key].(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	default:
		return 0, false
	}
}
