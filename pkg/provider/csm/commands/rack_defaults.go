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
	Class           string `json:"class,omitempty" yaml:"Class,omitempty"`
	Ordinal         int    `json:"ordinal,omitempty" yaml:"Ordinal,omitempty"`
	StartingHmnVlan int    `json:"startingHmnVlan,omitempty" yaml:"StartingHmnVlan,omitempty"`
	EndingHmnVlan   int    `json:"endingHmnVlan,omitempty" yaml:"EndingHmnVlan,omitempty"`
}

// DecodeRackCSMDefaults extracts the CSM-specific defaults from the
// generic ProviderDefaults map on a CaniRackType.  Returns nil when
// CSM defaults are absent.
func DecodeRackCSMDefaults(pd map[string]any) *RackProviderDefaultsCSM {
	csmRaw, ok := pd["csm"]
	if !ok {
		return nil
	}
	sub, ok := csmRaw.(map[string]any)
	if !ok {
		return nil
	}
	d := &RackProviderDefaultsCSM{}
	if v, ok := sub["Class"].(string); ok {
		d.Class = v
	}
	if v := toInt(sub["Ordinal"]); v != 0 {
		d.Ordinal = v
	}
	if v := toInt(sub["StartingHmnVlan"]); v != 0 {
		d.StartingHmnVlan = v
	}
	if v := toInt(sub["EndingHmnVlan"]); v != 0 {
		d.EndingHmnVlan = v
	}
	return d
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
