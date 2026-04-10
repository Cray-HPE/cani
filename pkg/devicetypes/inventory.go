/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
	"fmt"

	"github.com/google/uuid"
)

// Schema version constants for the inventory datastore.
const (
	SchemaVersionV1Alpha1 = "v1alpha1"
	SchemaVersionV1Alpha2 = "v1alpha2"
)

// Inventory represents the entire inventory of devices, racks, locations, etc.
// This is the structure stored in datastore.
type Inventory struct {
	SchemaVersion string `json:"schemaVersion" yaml:"schema_version"`
	Provider      string `json:"provider,omitempty" yaml:"provider,omitempty"`

	Locations  map[uuid.UUID]*CaniLocationType  `json:"locations"  yaml:"locations"`
	Racks      map[uuid.UUID]*CaniRackType      `json:"racks"      yaml:"racks"`
	Devices    map[uuid.UUID]*CaniDeviceType    `json:"devices"    yaml:"devices"`
	Modules    map[uuid.UUID]*CaniModuleType    `json:"modules"    yaml:"modules"`
	Cables     map[uuid.UUID]*CaniCableType     `json:"cables"     yaml:"cables"`
	Frus       map[uuid.UUID]*CaniFruType       `json:"frus"       yaml:"frus"`
	Interfaces map[uuid.UUID]*InterfaceInstance `json:"interfaces" yaml:"interfaces"`

	// Metadata stores the catalog of metadata definitions (roles,
	// statuses, tags) that individual inventory items reference.
	Metadata *InventoryMetadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// pkIndex is a transient (non-serialized) provider-key lookup cache.
	// Maps [provider][key][value] → device UUID for O(1) dedup lookups.
	// Rebuilt by RebuildProviderKeyIndex after datastore load.
	pkIndex providerKeyIndex `yaml:"-" json:"-"`
}

// TransformResult holds the combined output of the Transform step.
// Nil maps indicate the provider does not produce that type.
type TransformResult struct {
	Locations map[uuid.UUID]*CaniLocationType
	Racks     map[uuid.UUID]*CaniRackType
	Devices   map[uuid.UUID]*CaniDeviceType
	Modules   map[uuid.UUID]*CaniModuleType
	Cables    map[uuid.UUID]*CaniCableType
	Frus      map[uuid.UUID]*CaniFruType
}

// EnsureUniqueDeviceNames detects duplicate names within the transform
// result and appends an incrementing number to make each name unique.
// This runs before the result is merged into the inventory so that the
// provider never introduces collisions.
func (tr *TransformResult) EnsureUniqueDeviceNames() {
	if len(tr.Devices) == 0 {
		return
	}

	// Count how many times each name appears.
	nameCount := make(map[string]int)
	for _, d := range tr.Devices {
		if d != nil && d.Name != "" {
			nameCount[d.Name]++
		}
	}

	// For every duplicated name, assign an incrementing suffix.
	nameSeq := make(map[string]int) // next sequence number per base name
	for _, d := range tr.Devices {
		if d == nil || d.Name == "" {
			continue
		}
		if nameCount[d.Name] <= 1 {
			continue
		}
		nameSeq[d.Name]++
		d.Name = fmt.Sprintf("%s-%d", d.Name, nameSeq[d.Name])
	}
}

// NewInventory creates an Inventory with all maps initialized.
func NewInventory() *Inventory {
	return &Inventory{
		SchemaVersion: SchemaVersionV1Alpha2,
		Locations:     make(map[uuid.UUID]*CaniLocationType),
		Racks:         make(map[uuid.UUID]*CaniRackType),
		Devices:       make(map[uuid.UUID]*CaniDeviceType),
		Modules:       make(map[uuid.UUID]*CaniModuleType),
		Cables:        make(map[uuid.UUID]*CaniCableType),
		Frus:          make(map[uuid.UUID]*CaniFruType),
		Interfaces:    make(map[uuid.UUID]*InterfaceInstance),
		Metadata:      &InventoryMetadata{},
		pkIndex:       make(providerKeyIndex),
	}
}
