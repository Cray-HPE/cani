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
package transform

import (
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// BuildLocationTypeKeyMap maps each Nautobot LocationType UUID to the cani
// location-type key a CaniLocationType stores. Export writes the registered
// definition's display name ("Data Center") for the key ("dc"), so import
// reverses that through the registry and falls back to a slug of the name
// for types cani has no definition for.
func BuildLocationTypeKeyMap(raw []nautobotapi.LocationType) map[uuid.UUID]string {
	keys := make(map[uuid.UUID]string, len(raw))
	for _, lt := range raw {
		if lt.Id != nil {
			keys[uuid.UUID(*lt.Id)] = LocationTypeKey(lt.Name)
		}
	}
	return keys
}

// LocationTypeKey returns the cani key for a Nautobot location type name:
// the registered definition whose name or slug matches (case-insensitive),
// otherwise the name lower-cased with spaces as hyphens.
func LocationTypeKey(name string) string {
	for slug, def := range devicetypes.AllLocationTypes() {
		if strings.EqualFold(def.Name, name) || strings.EqualFold(slug, name) {
			return slug
		}
	}
	return strings.ToLower(strings.Join(strings.Fields(name), "-"))
}
