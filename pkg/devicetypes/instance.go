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
package devicetypes

import (
	"maps"
	"slices"

	"github.com/google/uuid"
)

// NewInstance returns an inventory instance of the device type: a copy with a
// fresh ID whose interface specs and metadata are independent of the source,
// so IDs stamped on this instance never reach the library template or sibling
// instances created from the same template.
func (c *CaniDeviceType) NewInstance() *CaniDeviceType {
	if c == nil {
		return nil
	}
	instance := *c
	instance.ID = uuid.New()
	instance.Interfaces = newInstanceInterfaces(c.Interfaces)
	instance.ObjectMeta = cloneObjectMeta(c.ObjectMeta)
	instance.AssignedVLANs = slices.Clone(c.AssignedVLANs)
	instance.Children = nil
	instance.Frus = nil
	return &instance
}

// NewInstance returns an inventory instance of the module type with a fresh ID
// and interface specs independent of the source (see CaniDeviceType.NewInstance).
func (m *CaniModuleType) NewInstance() *CaniModuleType {
	if m == nil {
		return nil
	}
	instance := *m
	instance.ID = uuid.New()
	instance.Interfaces = newInstanceInterfaces(m.Interfaces)
	instance.ObjectMeta = cloneObjectMeta(m.ObjectMeta)
	instance.Frus = nil
	return &instance
}

// newInstanceInterfaces copies interface specs for a new instance. Identity
// and cable state are cleared so the instance is indexed with its own IDs.
func newInstanceInterfaces(specs []InterfaceSpec) []InterfaceSpec {
	if specs == nil {
		return nil
	}
	clones := make([]InterfaceSpec, len(specs))
	for index, spec := range specs {
		clone := spec
		clone.ID = uuid.Nil
		clone.ConnectedCable = nil
		clone.Tags = slices.Clone(spec.Tags)
		clone.TaggedVLANs = slices.Clone(spec.TaggedVLANs)
		clone.CustomFields = cloneAnyMap(spec.CustomFields)
		clone.ExternalIDs = maps.Clone(spec.ExternalIDs)
		clone.ProviderMetadata = cloneAnyMap(spec.ProviderMetadata)
		clones[index] = clone
	}
	return clones
}
