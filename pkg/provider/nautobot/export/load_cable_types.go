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
package export

import (
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// cableTypeMap maps CableCategory strings to their best-fit Nautobot CableTypeChoices.
var cableTypeMap = map[string]nautobotapi.CableTypeChoices{
	// Copper / Ethernet
	"cat3":  nautobotapi.CableTypeChoicesCat3,
	"cat5":  nautobotapi.CableTypeChoicesCat5,
	"cat5e": nautobotapi.CableTypeChoicesCat5e,
	"cat6":  nautobotapi.CableTypeChoicesCat6,
	"cat6a": nautobotapi.CableTypeChoicesCat6a,
	"cat7":  nautobotapi.CableTypeChoicesCat7,
	"cat7a": nautobotapi.CableTypeChoicesCat7a,
	"cat8":  nautobotapi.CableTypeChoicesCat8,
	// Coaxial
	"coaxial": nautobotapi.CableTypeChoicesCoaxial,
	// Direct-attach copper
	"dac-passive": nautobotapi.CableTypeChoicesDacPassive,
	"dac-active":  nautobotapi.CableTypeChoicesDacActive,
	// Active optical cable
	"aoc": nautobotapi.CableTypeChoicesAoc,
	// Fiber – multimode
	"mmf":     nautobotapi.CableTypeChoicesMmf,
	"mmf-om1": nautobotapi.CableTypeChoicesMmfOm1,
	"mmf-om2": nautobotapi.CableTypeChoicesMmfOm2,
	"mmf-om3": nautobotapi.CableTypeChoicesMmfOm3,
	"mmf-om4": nautobotapi.CableTypeChoicesMmfOm4,
	"mmf-om5": nautobotapi.CableTypeChoicesMmfOm5,
	// Fiber – singlemode
	"smf":     nautobotapi.CableTypeChoicesSmf,
	"smf-os1": nautobotapi.CableTypeChoicesSmfOs1,
	"smf-os2": nautobotapi.CableTypeChoicesSmfOs2,
	// MRJ21
	"mrj21-trunk": nautobotapi.CableTypeChoicesMrj21Trunk,
	// Power & other
	"power": nautobotapi.CableTypeChoicesPower,
	"other": nautobotapi.CableTypeChoicesOther,
}

// connectorToCableType maps common ConnectorType values to a reasonable
// CableTypeChoices default when CableType and CableCategory are empty.
var connectorToCableType = map[string]nautobotapi.CableTypeChoices{
	"rj45":   nautobotapi.CableTypeChoicesCat6,
	"rj-45":  nautobotapi.CableTypeChoicesCat6,
	"lc":     nautobotapi.CableTypeChoicesSmfOs2,
	"sc":     nautobotapi.CableTypeChoicesMmfOm4,
	"mpo":    nautobotapi.CableTypeChoicesMmfOm4,
	"mtp":    nautobotapi.CableTypeChoicesMmfOm4,
	"sfp":    nautobotapi.CableTypeChoicesDacPassive,
	"sfp+":   nautobotapi.CableTypeChoicesDacPassive,
	"sfp28":  nautobotapi.CableTypeChoicesDacPassive,
	"qsfp":   nautobotapi.CableTypeChoicesDacPassive,
	"qsfp+":  nautobotapi.CableTypeChoicesDacPassive,
	"qsfp28": nautobotapi.CableTypeChoicesDacPassive,
}

// resolveCableType determines the Nautobot CableTypeChoices for a cable.
// Priority: explicit CableType field → CableCategory lookup → ConnectorType
// heuristic → slug-based fallback.
func resolveCableType(cable *devicetypes.CaniCableType) *nautobotapi.CableTypeChoices {
	// 1. Explicit CableType field (already a Nautobot enum string)
	if cable.CableType != "" {
		if choice, ok := cableTypeMap[cable.CableType]; ok {
			return &choice
		}
	}

	// 2. CableCategory lookup (e.g. "cat6a", "mmf-om4", "dac-passive")
	if cable.CableCategory != "" {
		if choice, ok := cableTypeMap[strings.ToLower(cable.CableCategory)]; ok {
			return &choice
		}
	}

	// 3. ConnectorType heuristic (e.g. "rj45" → cat6, "lc" → smf-os2)
	if cable.ConnectorType != "" {
		if choice, ok := connectorToCableType[strings.ToLower(cable.ConnectorType)]; ok {
			return &choice
		}
	}

	// 4. Legacy slug-based fallback
	slug := strings.ToLower(cable.Slug)
	switch {
	case strings.Contains(slug, "cat"):
		c := nautobotapi.CableTypeChoicesCat5e
		return &c
	case strings.Contains(slug, "dac"):
		c := nautobotapi.CableTypeChoicesDacPassive
		return &c
	case strings.Contains(slug, "aoc"):
		c := nautobotapi.CableTypeChoicesAoc
		return &c
	case strings.Contains(slug, "fiber") || strings.Contains(slug, "mmf"):
		c := nautobotapi.CableTypeChoicesMmfOm4
		return &c
	}

	return nil
}

// colorNameToHex converts common color names to 6-digit hex codes expected by
// Nautobot. If the value already looks like a hex code it is returned as-is.
func colorNameToHex(name string) string {
	namedColors := map[string]string{
		"black":  "000000",
		"white":  "ffffff",
		"red":    "ff0000",
		"green":  "00ff00",
		"blue":   "0000ff",
		"yellow": "ffff00",
		"orange": "ff8800",
		"purple": "800080",
		"grey":   "808080",
		"gray":   "808080",
		"brown":  "795548",
		"pink":   "e91e63",
		"teal":   "009688",
		"cyan":   "00bcd4",
	}

	lower := strings.ToLower(strings.TrimSpace(name))
	// Strip leading '#' if present
	lower = strings.TrimPrefix(lower, "#")

	if hex, ok := namedColors[lower]; ok {
		return hex
	}
	// Assume it's already a hex code
	return lower
}
