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

package cmdtest

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// uuidPattern matches any RFC 4122 identifier appearing in serialized output.
var uuidPattern = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

// Fingerprint renders an inventory as sorted, stable text for comparison.
//
// Identifiers are generated fresh on every run, so two semantically identical
// inventories never compare equal byte-for-byte. Fingerprint masks every UUID
// and sorts the result, leaving everything else intact.
//
// It serializes each entity in full rather than listing chosen fields, so a
// divergence in any serialized attribute — Interfaces, ModuleBays, Tags,
// CustomFields, ProviderMetadata — is caught. Picking fields by hand would
// silently stop covering whatever the schema gained next. Masking UUIDs does
// lose parentage, so a name-resolved relationship line is emitted alongside.
func Fingerprint(inventory *devicetypes.Inventory) string {
	if inventory == nil {
		return "<nil inventory>"
	}

	var lines []string

	for _, device := range inventory.Devices {
		lines = append(lines,
			fmt.Sprintf("rel  device %s parent=%s", device.Name, parentName(inventory, device)),
			"device "+device.Name+" "+encode(device))
	}
	for _, rack := range inventory.Racks {
		lines = append(lines, "rack "+rack.Name+" "+encode(rack))
	}
	for _, location := range inventory.Locations {
		lines = append(lines, "location "+location.Name+" "+encode(location))
	}
	for _, module := range inventory.Modules {
		lines = append(lines, "module "+module.Name+" "+encode(module))
	}
	for _, cable := range inventory.Cables {
		lines = append(lines, "cable "+cable.Label+" "+encode(cable))
	}
	for _, iface := range inventory.Interfaces {
		lines = append(lines, "interface "+iface.Name+" "+encode(iface))
	}

	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// MaskIDs replaces every identifier in text with a fixed placeholder.
//
// Identifiers are generated fresh on every run, so any rendering that carries
// them has to be masked before two runs can be compared. Callers comparing
// command output rather than an inventory need this directly.
func MaskIDs(text string) string {
	return uuidPattern.ReplaceAllString(text, "<id>")
}

// encode serializes one entity with every identifier masked.
func encode(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("<unencodable: %v>", err)
	}
	return MaskIDs(string(encoded))
}

// parentName resolves a device's parent FK to a name, recording the
// relationship that masking the UUID would otherwise erase.
func parentName(inventory *devicetypes.Inventory, device *devicetypes.CaniDeviceType) string {
	if rack, ok := inventory.Racks[device.Parent]; ok {
		return "rack:" + rack.Name
	}
	if parent, ok := inventory.Devices[device.Parent]; ok {
		return "device:" + parent.Name
	}
	return "none"
}
