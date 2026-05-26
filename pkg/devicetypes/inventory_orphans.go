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
package devicetypes

import (
	"sort"

	"github.com/google/uuid"
)

// OrphanItem describes an inventory item that has no parent assigned.
type OrphanItem struct {
	ID               uuid.UUID
	Name             string
	Kind             string // "device" or "rack"
	DeviceType       string
	Model            string
	Manufacturer     string
	ProviderMetadata map[string]any
}

// OrphanDevices returns all devices whose Parent field is uuid.Nil.
// Results are sorted by name for deterministic output.
func (inv *Inventory) OrphanDevices() []OrphanItem {
	var result []OrphanItem
	for _, d := range inv.Devices {
		if d == nil || d.Parent != uuid.Nil {
			continue
		}
		result = append(result, OrphanItem{
			ID:               d.ID,
			Name:             d.Name,
			Kind:             "device",
			DeviceType:       string(d.Type),
			Model:            d.Model,
			Manufacturer:     d.Manufacturer,
			ProviderMetadata: d.ProviderMetadata,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// OrphanRacks returns all racks whose Location field is uuid.Nil.
// Results are sorted by name for deterministic output.
func (inv *Inventory) OrphanRacks() []OrphanItem {
	var result []OrphanItem
	for _, r := range inv.Racks {
		if r == nil || r.Location != uuid.Nil {
			continue
		}
		result = append(result, OrphanItem{
			ID:               r.ID,
			Name:             r.Name,
			Kind:             "rack",
			DeviceType:       string(r.Type),
			Model:            r.Model,
			Manufacturer:     r.Manufacturer,
			ProviderMetadata: r.ProviderMetadata,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
