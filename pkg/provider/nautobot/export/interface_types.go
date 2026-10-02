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
package export

import "strings"

// nautobotInterfaceTypes is the InterfaceTypeChoices enum of the pinned
// Nautobot client (pkg/nautobot/nautobot_api.go). TestNautobotInterfaceTypes
// MatchGeneratedEnum fails when the two drift after a client regeneration.
var nautobotInterfaceTypes = map[string]struct{}{
	"1000base-kx": {}, "1000base-t": {}, "1000base-x-gbic": {}, "1000base-x-sfp": {},
	"100base-fx": {}, "100base-lfx": {}, "100base-t1": {}, "100base-tx": {}, "100gbase-kp4": {},
	"100gbase-kr2": {}, "100gbase-kr4": {}, "100gbase-x-cfp": {}, "100gbase-x-cfp2": {},
	"100gbase-x-cfp4": {}, "100gbase-x-cpak": {}, "100gbase-x-cxp": {}, "100gbase-x-dsfp": {},
	"100gbase-x-qsfp28": {}, "100gbase-x-qsfpdd": {}, "100gbase-x-sfpdd": {}, "10g-epon": {},
	"10gbase-cx4": {}, "10gbase-kr": {}, "10gbase-kx4": {}, "10gbase-t": {}, "10gbase-x-sfpp": {},
	"10gbase-x-x2": {}, "10gbase-x-xenpak": {}, "10gbase-x-xfp": {}, "128gfc-qsfp28": {},
	"1600gbase-x-osfp": {}, "1600gbase-x-osfp-xd": {}, "16gfc-sfpp": {}, "1gfc-sfp": {},
	"2.5gbase-t": {}, "200gbase-x-cfp2": {}, "200gbase-x-qsfp56": {}, "200gbase-x-qsfpdd": {},
	"25gbase-kr": {}, "25gbase-x-sfp28": {}, "2gfc-sfp": {}, "32gfc-sfp28": {}, "32gfc-sfpp": {},
	"400gbase-x-cdfp": {}, "400gbase-x-cfp2": {}, "400gbase-x-cfp8": {}, "400gbase-x-osfp": {},
	"400gbase-x-osfp-rhs": {}, "400gbase-x-qsfp112": {}, "400gbase-x-qsfpdd": {}, "40gbase-kr4": {},
	"40gbase-x-qsfpp": {}, "4gfc-sfp": {}, "50gbase-kr": {}, "50gbase-x-sfp28": {},
	"50gbase-x-sfp56": {}, "5gbase-t": {}, "64gfc-qsfpp": {}, "64gfc-sfpdd": {}, "64gfc-sfpp": {},
	"800gbase-x-osfp": {}, "800gbase-x-osfp-xd": {}, "800gbase-x-qsfpdd": {}, "8gfc-sfpp": {},
	"bridge": {}, "cdma": {}, "cisco-flexstack": {}, "cisco-flexstack-plus": {},
	"cisco-stackwise": {}, "cisco-stackwise-160": {}, "cisco-stackwise-1t": {},
	"cisco-stackwise-320": {}, "cisco-stackwise-480": {}, "cisco-stackwise-80": {},
	"cisco-stackwise-plus": {}, "da15": {}, "da26": {}, "da31": {}, "db25": {}, "db44": {},
	"db60": {}, "dc37": {}, "dc62": {}, "dc79": {}, "dd100": {}, "dd50": {}, "dd78": {}, "de15": {},
	"de19": {}, "de9": {}, "df104": {}, "docsis": {}, "e1": {}, "e3": {}, "epon": {},
	"extreme-summitstack": {}, "extreme-summitstack-128": {}, "extreme-summitstack-256": {},
	"extreme-summitstack-512": {}, "gpon": {}, "gsm": {}, "ieee802.11a": {}, "ieee802.11ac": {},
	"ieee802.11ad": {}, "ieee802.11ax": {}, "ieee802.11ay": {}, "ieee802.11g": {}, "ieee802.11n": {},
	"ieee802.15.1": {}, "infiniband-ddr": {}, "infiniband-edr": {}, "infiniband-fdr": {},
	"infiniband-fdr10": {}, "infiniband-hdr": {}, "infiniband-ndr": {}, "infiniband-qdr": {},
	"infiniband-sdr": {}, "infiniband-xdr": {}, "juniper-vcp": {}, "lag": {}, "lte": {},
	"ng-pon2": {}, "other": {}, "other-wireless": {}, "sonet-oc12": {}, "sonet-oc192": {},
	"sonet-oc1920": {}, "sonet-oc3": {}, "sonet-oc3840": {}, "sonet-oc48": {}, "sonet-oc768": {},
	"t1": {}, "t3": {}, "tunnel": {}, "virtual": {}, "xdsl": {}, "xg-pon": {}, "xgs-pon": {},
}

// interfaceTypeAliases maps library spellings Nautobot lacks to the enum
// value Nautobot uses for the same port.
var interfaceTypeAliases = map[string]string{
	"1gbase-t": ifaceType1000BaseT,
}

// isValidNautobotInterfaceType reports whether Nautobot accepts the type as-is.
func isValidNautobotInterfaceType(ifaceType string) bool {
	_, ok := nautobotInterfaceTypes[ifaceType]
	return ok
}

// mapInterfaceType returns the Nautobot interface type for a library type.
// Types Nautobot accepts pass through unchanged; only spellings Nautobot lacks
// are rewritten, and anything unknown is returned as-is for
// supportedInterfaceSpecs to report. An empty type defaults to 1000base-t.
func mapInterfaceType(ifaceType string) string {
	lower := strings.ToLower(strings.TrimSpace(ifaceType))
	if lower == "" {
		return ifaceType1000BaseT
	}
	if isValidNautobotInterfaceType(lower) {
		return lower
	}
	if mapped, ok := interfaceTypeAliases[lower]; ok {
		return mapped
	}
	return ifaceType
}

// supportedInterfaceSpecs drops the specs whose type Nautobot cannot store,
// warning once per interface and counting them in result.IfacesUnsupported.
// It runs before dry-run reporting and before any create, so both describe
// the same set of interfaces.
func supportedInterfaceSpecs(specs []interfaceSpec, owner string, result *LoadResult) []interfaceSpec {
	kept := make([]interfaceSpec, 0, len(specs))
	for _, spec := range specs {
		if !isValidNautobotInterfaceType(spec.Type) {
			clog.Warn("Skipping interface %s on %s: Nautobot has no interface type %q", spec.Name, owner, spec.Type)
			result.IfacesUnsupported++
			continue
		}
		kept = append(kept, spec)
	}
	return kept
}
