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

func TestNormalizeMAC(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"colon lowercase", "aa:bb:cc:dd:ee:ff", "aa:bb:cc:dd:ee:ff", false},
		{"colon uppercase", "AA:BB:CC:DD:EE:FF", "aa:bb:cc:dd:ee:ff", false},
		{"hyphen", "aa-bb-cc-dd-ee-ff", "aa:bb:cc:dd:ee:ff", false},
		{"dotted", "aabb.ccdd.eeff", "aa:bb:cc:dd:ee:ff", false},
		{"empty", "", "", false},
		{"whitespace", "   ", "", false},
		{"surrounding whitespace", "  aa:bb:cc:dd:ee:ff  ", "aa:bb:cc:dd:ee:ff", false},
		{"too short", "aa:bb:cc", "", true},
		{"non-hex", "zz:zz:zz:zz:zz:zz", "", true},
		{"garbage", "not-a-mac", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeMAC(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NormalizeMAC(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeMAC(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("NormalizeMAC(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// macTestInventory builds an inventory with a single device and module, each
// carrying one interface spec, plus the derived interface instance map.
func macTestInventory() (*Inventory, uuid.UUID, uuid.UUID) {
	inv := NewInventory()

	deviceID := uuid.New()
	inv.Devices[deviceID] = &CaniDeviceType{
		ID:   deviceID,
		Name: "node-01",
		Interfaces: []InterfaceSpec{
			{ID: uuid.New(), Name: "iLO", Type: InterfacesElemTypeA1000BaseT},
		},
	}

	moduleID := uuid.New()
	inv.Modules[moduleID] = &CaniModuleType{
		ID:           moduleID,
		Name:         "gpu-0",
		ParentDevice: deviceID,
		Interfaces: []InterfaceSpec{
			{ID: uuid.New(), Name: "eth0", Type: InterfacesElemTypeA1000BaseT},
		},
	}

	inv.rebuildInterfaceRelationships()
	return inv, deviceID, moduleID
}

func TestSetInterfaceMACByID_Device(t *testing.T) {
	inv, deviceID, _ := macTestInventory()

	if err := inv.SetInterfaceMACByID(deviceID, "iLO", "AA-BB-CC-DD-EE-FF"); err != nil {
		t.Fatalf("SetInterfaceMACByID: %v", err)
	}

	spec := inv.findInterfaceSpecOnOwner(deviceID, "iLO")
	if spec == nil {
		t.Fatal("interface spec not found")
	}
	if spec.MacAddress != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("spec MAC = %q, want %q", spec.MacAddress, "aa:bb:cc:dd:ee:ff")
	}

	// The already-indexed instance should be mirrored.
	inst, ok := inv.Interfaces[spec.ID]
	if !ok {
		t.Fatal("interface instance not indexed")
	}
	if inst.MacAddress != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("instance MAC = %q, want %q", inst.MacAddress, "aa:bb:cc:dd:ee:ff")
	}
}

func TestSetInterfaceMACByID_Module(t *testing.T) {
	inv, _, moduleID := macTestInventory()

	if err := inv.SetInterfaceMACByID(moduleID, "eth0", "aa:bb:cc:dd:ee:01"); err != nil {
		t.Fatalf("SetInterfaceMACByID: %v", err)
	}

	spec := inv.findInterfaceSpecOnOwner(moduleID, "eth0")
	if spec == nil || spec.MacAddress != "aa:bb:cc:dd:ee:01" {
		t.Errorf("module interface MAC = %v, want aa:bb:cc:dd:ee:01", spec)
	}
}

func TestSetInterfaceMAC_ByName(t *testing.T) {
	inv, _, _ := macTestInventory()

	if err := inv.SetInterfaceMAC("node-01", "iLO", "aa:bb:cc:dd:ee:02"); err != nil {
		t.Fatalf("SetInterfaceMAC: %v", err)
	}

	spec := inv.findInterfaceSpecOnOwner(inv.FindConnectableByNameOrID("node-01"), "iLO")
	if spec == nil || spec.MacAddress != "aa:bb:cc:dd:ee:02" {
		t.Errorf("interface MAC = %v, want aa:bb:cc:dd:ee:02", spec)
	}
}

func TestSetInterfaceMAC_Errors(t *testing.T) {
	inv, deviceID, _ := macTestInventory()

	if err := inv.SetInterfaceMAC("nonexistent", "iLO", "aa:bb:cc:dd:ee:ff"); err == nil {
		t.Error("expected error for unknown owner")
	}

	if err := inv.SetInterfaceMACByID(deviceID, "nope", "aa:bb:cc:dd:ee:ff"); err == nil {
		t.Error("expected error for unknown interface")
	}

	if err := inv.SetInterfaceMACByID(deviceID, "iLO", "not-a-mac"); err == nil {
		t.Error("expected error for invalid MAC")
	}

	// An invalid MAC must not mutate the spec.
	spec := inv.findInterfaceSpecOnOwner(deviceID, "iLO")
	if spec.MacAddress != "" {
		t.Errorf("spec MAC = %q, want empty after failed set", spec.MacAddress)
	}
}

func TestRebuildInterfaceRelationships_CopiesMAC(t *testing.T) {
	inv, deviceID, _ := macTestInventory()

	spec := inv.findInterfaceSpecOnOwner(deviceID, "iLO")
	spec.MacAddress = "aa:bb:cc:dd:ee:ff"

	// Rebuild discards and regenerates the instance map; MAC must survive.
	inv.rebuildInterfaceRelationships()

	inst, ok := inv.Interfaces[spec.ID]
	if !ok {
		t.Fatal("interface instance not indexed after rebuild")
	}
	if inst.MacAddress != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("instance MAC after rebuild = %q, want %q", inst.MacAddress, "aa:bb:cc:dd:ee:ff")
	}
}
