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
	"testing"

	"github.com/google/uuid"
)

// TestDeviceNewInstanceOwnsInterfaceSpecs verifies that two instances created
// from one device template keep independent interface specs: stamping an ID on
// one leaves the sibling and the template untouched.
//
// Why it matters: `cani add device --qty N` copies one template per instance;
// a shared Interfaces backing array made every sibling report the same
// interface UUID, collapsing the inventory interface index.
// Inputs: a template with one spec and two NewInstance copies; the first copy's
// spec ID is set. Outputs: the sibling and template specs still have a nil ID,
// and the instances have distinct non-nil IDs.
// Data choice: a nil template ID mirrors library YAML, which carries no IDs.
func TestDeviceNewInstanceOwnsInterfaceSpecs(t *testing.T) {
	template := &CaniDeviceType{
		Slug:       "tpl",
		Interfaces: []InterfaceSpec{{Name: "eth0", Type: InterfacesElemTypeA1000BaseT}},
		ObjectMeta: ObjectMeta{Tags: []string{"lib"}, ProviderMetadata: map[string]any{"k": "v"}},
	}
	first := template.NewInstance()
	second := template.NewInstance()

	first.Interfaces[0].ID = uuid.New()
	first.Tags = append(first.Tags, "mine")
	first.ProviderMetadata["k"] = "changed"

	if second.Interfaces[0].ID != uuid.Nil || template.Interfaces[0].ID != uuid.Nil {
		t.Fatalf("interface ID leaked: sibling=%s template=%s", second.Interfaces[0].ID, template.Interfaces[0].ID)
	}
	if first.ID == second.ID || first.ID == uuid.Nil {
		t.Fatalf("instances must have distinct IDs, got %s and %s", first.ID, second.ID)
	}
	if len(template.Tags) != 1 || template.ProviderMetadata["k"] != "v" {
		t.Fatalf("metadata leaked into template: tags=%v meta=%v", template.Tags, template.ProviderMetadata)
	}
}

// TestNewInstanceResetsInterfaceIdentity verifies NewInstance clears the
// interface ID and cable link copied from an already-indexed source so the new
// instance is indexed with its own identity.
//
// Why it matters: an instance cloned from another instance must not keep
// pointing at the source's interfaces or cables.
// Inputs: a module whose spec carries an ID and a ConnectedCable. Outputs: the
// clone's spec has a nil ID and no cable; the source keeps both.
// Data choice: a module exercises the second receiver; the cable pointer is the
// only pointer field on InterfaceSpec.
func TestNewInstanceResetsInterfaceIdentity(t *testing.T) {
	cableID := uuid.New()
	source := &CaniModuleType{
		Interfaces: []InterfaceSpec{{ID: uuid.New(), Name: "p1", ConnectedCable: &cableID, Tags: []string{"a"}}},
	}

	clone := source.NewInstance()
	clone.Interfaces[0].Tags[0] = "b"

	if clone.Interfaces[0].ID != uuid.Nil || clone.Interfaces[0].ConnectedCable != nil {
		t.Fatalf("clone kept source identity: id=%s cable=%v", clone.Interfaces[0].ID, clone.Interfaces[0].ConnectedCable)
	}
	if source.Interfaces[0].ID == uuid.Nil || source.Interfaces[0].ConnectedCable == nil || source.Interfaces[0].Tags[0] != "a" {
		t.Fatalf("source was modified: %+v", source.Interfaces[0])
	}
	if nilDevice := (*CaniDeviceType)(nil).NewInstance(); nilDevice != nil {
		t.Fatalf("nil receiver should yield nil, got %+v", nilDevice)
	}
}

