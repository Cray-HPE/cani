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
	"testing"

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// TestTransform_LegacyPathCopiesExistingDevices verifies that with no raw API
// data stored, Transform takes the legacy path and copies existing inventory
// devices straight through.
func TestTransform_LegacyPathCopiesExistingDevices(t *testing.T) {
	id := uuid.New()
	existing := devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id: {ID: id, Name: "node-1"},
		},
	}

	p := New()
	res, err := p.Transform(context.Background(), existing)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}
	if res == nil {
		t.Fatal("expected a non-nil result")
	}
	if _, ok := res.Devices[id]; !ok {
		t.Errorf("expected device %s to be copied into the result", id)
	}
}

// TestTransform_UsesRawDataWhenPresent verifies that storing raw API data
// switches Transform to the raw mapping path and produces the mapped entity.
func TestTransform_UsesRawDataWhenPresent(t *testing.T) {
	locID := openapi_types.UUID(uuid.New())

	p := New()
	p.rawLocations = []nautobotapi.Location{
		{Id: &locID, Name: "Site-A"},
	}

	res, err := p.Transform(context.Background(), devicetypes.Inventory{})
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}
	if res == nil {
		t.Fatal("expected a non-nil result from the raw-data path")
	}
	if len(res.Locations) != 1 {
		t.Errorf("len(Locations) = %d, want 1 mapped location", len(res.Locations))
	}
}
