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
package export

import (
	"context"
	"net/http"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// unsupportedTypeInventory returns an inventory holding one device with a
// Nautobot-valid eth0 and an nvlink port Nautobot has no type for.
func unsupportedTypeInventory() (*devicetypes.Inventory, map[string]uuid.UUID) {
	dev := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "gpu-node",
		Type: "node",
		Interfaces: []devicetypes.InterfaceSpec{
			{Name: "eth0", Type: "1000base-t"},
			{Name: "nv0", Type: "nvlink"},
		},
	}
	inv := devicetypes.NewInventory()
	inv.Devices[dev.ID] = dev
	return inv, map[string]uuid.UUID{dev.Name: uuid.New()}
}

// TestLoadInterfaces_DryRunSkipsUnsupportedType verifies that a dry run
// reports only the interfaces a real export would create and counts the
// unsupported port, without issuing any POST.
//
// Why it matters: a dry run that promised an nvlink interface would mislead
// the operator, because the real run cannot create it; both must agree.
// Inputs: a device with eth0 (1000base-t) and nv0 (nvlink), DryRun enabled.
// Outputs: IfacesCreated == 1, IfacesUnsupported == 1, zero interface POSTs.
// Data choice: nvlink is the GPU interconnect the shipped H100 module
// defines, the one library type Nautobot lacks in practice.
func TestLoadInterfaces_DryRunSkipsUnsupportedType(t *testing.T) {
	inv, createdDeviceIDs := unsupportedTypeInventory()
	var posts int
	e, cleanup := newExporterWithServer(t, moduleIfaceServer(&posts, http.StatusCreated, `{}`))
	defer cleanup()
	e.Options.DryRun = true

	result := &LoadResult{}
	if err := e.loadInterfaces(context.Background(), inv, createdDeviceIDs, result); err != nil {
		t.Fatalf("loadInterfaces() error = %v", err)
	}

	if result.IfacesCreated != 1 || result.IfacesUnsupported != 1 {
		t.Errorf("IfacesCreated = %d, IfacesUnsupported = %d, want 1 and 1", result.IfacesCreated, result.IfacesUnsupported)
	}
	if posts != 0 {
		t.Errorf("interface POSTs = %d, want 0 in dry-run", posts)
	}
}

// TestLoadInterfaces_CreatesOnlySupportedTypes verifies the device path drops
// an unsupported type before the bulk POST, so the payload carries only the
// interfaces Nautobot can store and the omission is counted.
//
// Why it matters: before the pre-flight filter one unsupported type failed
// the whole bulk batch and surfaced as a per-interface error after the
// fallback; filtering first keeps the batch intact and the report explicit.
// Inputs: a device with eth0 (1000base-t) and nv0 (nvlink) plus a seeded
// Active status. Outputs: a POST body with eth0 only, IfacesCreated == 1,
// IfacesUnsupported == 1.
// Data choice: the same two-port device as the dry-run test, so the two
// tests prove dry-run and real export describe the same interfaces.
func TestLoadInterfaces_CreatesOnlySupportedTypes(t *testing.T) {
	inv, createdDeviceIDs := unsupportedTypeInventory()
	rec := &capturedRequest{}
	created := []map[string]string{{"id": uuid.NewString(), "name": "eth0"}}
	e, cleanup := newExporterWithServer(t, interfaceServer(rec, created))
	defer cleanup()
	seedActiveStatus(t, e)

	result := &LoadResult{}
	if err := e.loadInterfaces(context.Background(), inv, createdDeviceIDs, result); err != nil {
		t.Fatalf("loadInterfaces() error = %v", err)
	}

	sent := decodeSentInterfaces(t, rec.body)
	if len(sent) != 1 || sent[0].Name != "eth0" {
		t.Errorf("POST body = %+v, want eth0 only", sent)
	}
	if result.IfacesCreated != 1 || result.IfacesUnsupported != 1 {
		t.Errorf("IfacesCreated = %d, IfacesUnsupported = %d, want 1 and 1", result.IfacesCreated, result.IfacesUnsupported)
	}
}
