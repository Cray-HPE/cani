/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
package commands

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// RackProviderDefaultsCSM holds CSM-specific defaults for a rack type.
type RackProviderDefaultsCSM struct {
	Class           string `json:"class,omitempty" yaml:"class,omitempty"`
	Ordinal         int    `json:"ordinal,omitempty" yaml:"ordinal,omitempty"`
	StartingHmnVlan int    `json:"startingHmnVlan,omitempty" yaml:"starting_hmn_vlan,omitempty"`
	EndingHmnVlan   int    `json:"endingHmnVlan,omitempty" yaml:"ending_hmn_vlan,omitempty"`
}

// rackDefaultsBySlug maps a rack-type slug to the CSM cabinet numbering and
// HMN VLAN policy for that hardware.  These defaults are CSM configuration,
// not portable hardware data, so they live here rather than in the shared
// device-type library.
//
// hpe-eia-chassis is intentionally absent: its library entry spelled the key
// starting_cabinet, which the old decoder never read, so it never applied.
var rackDefaultsBySlug = map[string]RackProviderDefaultsCSM{
	"hpe-eia-cabinet":                    {Class: "River", Ordinal: 3000, StartingHmnVlan: 1513, EndingHmnVlan: 1769},
	"hpe-ex2000":                         {Class: "Hill", Ordinal: 9000, StartingHmnVlan: 3000, EndingHmnVlan: 3999},
	"hpe-ex2500-1-liquid-cooled-chassis": {Class: "Hill", Ordinal: 8000, StartingHmnVlan: 3000, EndingHmnVlan: 3999},
	"hpe-ex2500-2-liquid-cooled-chassis": {Class: "Hill", Ordinal: 8000, StartingHmnVlan: 3000, EndingHmnVlan: 3999},
	"hpe-ex2500-3-liquid-cooled-chassis": {Class: "Hill", Ordinal: 8000, StartingHmnVlan: 3000, EndingHmnVlan: 3999},
	"hpe-ex3000":                         {Class: "Mountain", Ordinal: 1000, StartingHmnVlan: 3000, EndingHmnVlan: 3999},
	"hpe-ex4000":                         {Class: "Mountain", Ordinal: 1000, StartingHmnVlan: 3000, EndingHmnVlan: 3999},
}

// LookupRackDefaults returns the CSM defaults for a rack-type slug, or nil
// when the slug has no CSM cabinet policy.
func LookupRackDefaults(slug string) *RackProviderDefaultsCSM {
	d, ok := rackDefaultsBySlug[slug]
	if !ok {
		return nil
	}
	return &d
}

// BayOrdinal extracts the provider-specific ordinal from a DeviceBaySpec's
// Extra map.  Returns 0 when not present.
func BayOrdinal(bay devicetypes.DeviceBaySpec) int {
	return toInt(bay.Extra["ordinal"])
}

// toInt converts a numeric interface{} (int or float64 from YAML/JSON
// decoding) to int.  Returns 0 for non-numeric values.
func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	}
	return 0
}
