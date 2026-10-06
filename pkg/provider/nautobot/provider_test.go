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
package nautobot

import (
	"context"
	"reflect"
	"testing"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	imprt "github.com/Cray-HPE/cani/pkg/provider/nautobot/import"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/transform"
)

func TestNew(t *testing.T) {
	p := New()

	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	if p.Options == nil {
		t.Fatal("expected non-nil Options")
	}
	if p.Options.URL != "http://localhost:8081/api" {
		t.Errorf("URL = %q, want %q", p.Options.URL, "http://localhost:8081/api")
	}
	if p.Options.Import == nil {
		t.Error("expected non-nil Import opts")
	}
	if p.Options.Export == nil {
		t.Error("expected non-nil Export opts")
	}
}

func TestSlug(t *testing.T) {
	p := New()
	if got := p.Slug(); got != "nautobot" {
		t.Errorf("Slug() = %q, want %q", got, "nautobot")
	}
}

// TestClearRawData verifies ClearRawData discards a stored snapshot so the
// next Transform takes the legacy path until an import stores a new one.
//
// Why it matters: a stale snapshot would be re-mapped on a transform-only run.
// Inputs: a provider holding a snapshot with one location. Outputs: raw is nil
// after ClearRawData.
// Data choice: one populated collection is enough to show the whole snapshot
// goes, since it is held as a single value.
func TestClearRawData(t *testing.T) {
	p := New()
	p.raw = &transform.RawData{Locations: []nautobotapi.Location{{Name: "loc"}}}

	p.ClearRawData()

	if p.raw != nil {
		t.Error("snapshot not cleared")
	}
}

// TestSetRawData verifies SetRawData stores the import snapshot intact for
// Transform.
//
// Why it matters: every collection the transform maps must arrive unchanged.
// Inputs: a snapshot with one record in several collections. Outputs: the
// stored snapshot equals the input.
// Data choice: named records in distinct collections catch a copy that drops
// or swaps a field.
func TestSetRawData(t *testing.T) {
	p := New()

	d := imprt.RawData{
		Locations:                 []nautobotapi.Location{{Name: "site-a"}},
		Racks:                     []nautobotapi.Rack{{Name: "rack-1"}},
		Devices:                   []nautobotapi.Device{{}},
		DeviceTypes:               []nautobotapi.DeviceType{{Model: "DL380"}},
		Interfaces:                []nautobotapi.Interface{{Name: "eth0"}},
		Modules:                   []nautobotapi.Module{{}},
		ModuleBays:                []nautobotapi.ModuleBay{{Name: "bay-0"}},
		Cables:                    []nautobotapi.Cable{{}},
		InventoryItems:            []nautobotapi.InventoryItem{{Name: "gpu-0"}},
		Statuses:                  []nautobotapi.Status{{Name: "Active"}},
		Roles:                     []nautobotapi.Role{{Name: "Compute"}},
		VLANLocationAssignments:   []nautobotapi.VLANLocationAssignment{{}},
		PrefixLocationAssignments: []nautobotapi.PrefixLocationAssignment{{}},
		Namespaces:                []nautobotapi.Namespace{{Name: "Global"}},
	}

	p.SetRawData(d)

	if p.raw == nil || !reflect.DeepEqual(*p.raw, d) {
		t.Errorf("stored snapshot = %+v, want %+v", p.raw, d)
	}
}

func TestGetContext(t *testing.T) {
	p := New()

	t.Run("nil context returns nil", func(t *testing.T) {
		if p.GetContext() != nil {
			t.Error("expected nil context")
		}
	})

	t.Run("returns stored context", func(t *testing.T) {
		ctx := context.Background()
		p.ctx = ctx
		if p.GetContext() != ctx {
			t.Error("GetContext() did not return stored context")
		}
	})
}
