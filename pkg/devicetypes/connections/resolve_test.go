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
package connections

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func testInventoryWithDevices(names ...string) *devicetypes.Inventory {
	inv := devicetypes.NewInventory()
	for _, name := range names {
		id := uuid.New()
		inv.Devices[id] = &devicetypes.CaniDeviceType{
			ID:   id,
			Name: name,
			Interfaces: []devicetypes.InterfaceSpec{
				{ID: uuid.New(), Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT},
				{ID: uuid.New(), Name: "eth1", Type: devicetypes.InterfacesElemTypeA1000BaseT},
			},
		}
	}
	return inv
}

func TestResolveConnectionMap_SingleConnection(t *testing.T) {
	inv := testInventoryWithDevices("node-01", "switch-01")

	cm := &ConnectionMap{
		Version: "v1",
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "node-01", Port: "eth0"},
				B: Endpoint{Device: "switch-01", Port: "eth1"},
			},
		},
	}

	resolved, errs := ResolveConnectionMap(cm, inv)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(resolved) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(resolved))
	}
	if resolved[0].APort != "eth0" || resolved[0].BPort != "eth1" {
		t.Errorf("unexpected ports: %s -> %s", resolved[0].APort, resolved[0].BPort)
	}
}

func TestResolveConnectionMap_PatternExpansion(t *testing.T) {
	inv := testInventoryWithDevices("node-01", "node-02", "node-03", "switch-01")

	cm := &ConnectionMap{
		Version: "v1",
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "node-{01..03}", Port: "eth0"},
				B: Endpoint{Device: "switch-01", Port: "eth{1..3}"},
			},
		},
	}

	resolved, errs := ResolveConnectionMap(cm, inv)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(resolved) != 3 {
		t.Fatalf("expected 3 connections, got %d", len(resolved))
	}

	expectedPorts := []string{"eth1", "eth2", "eth3"}
	for i, conn := range resolved {
		if conn.APort != "eth0" {
			t.Errorf("connection[%d]: expected A port eth0, got %s", i, conn.APort)
		}
		if conn.BPort != expectedPorts[i] {
			t.Errorf("connection[%d]: expected B port %s, got %s", i, expectedPorts[i], conn.BPort)
		}
	}
}

func TestResolveConnectionMap_DeviceNotFound(t *testing.T) {
	inv := testInventoryWithDevices("node-01")

	cm := &ConnectionMap{
		Version: "v1",
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "node-01", Port: "eth0"},
				B: Endpoint{Device: "nonexistent", Port: "eth0"},
			},
		},
	}

	resolved, errs := ResolveConnectionMap(cm, inv)
	if len(errs) == 0 {
		t.Fatal("expected errors for missing device")
	}
	if len(resolved) != 0 {
		t.Errorf("expected 0 resolved connections, got %d", len(resolved))
	}
}

func TestResolveConnectionMap_CableDefaults(t *testing.T) {
	inv := testInventoryWithDevices("node-01", "switch-01")

	cm := &ConnectionMap{
		Version:       "v1",
		CableDefaults: &CableDefaults{Type: "cat6a", Status: "connected", Color: "blue"},
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "node-01", Port: "eth0"},
				B: Endpoint{Device: "switch-01", Port: "eth0"},
			},
		},
	}

	resolved, errs := ResolveConnectionMap(cm, inv)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if resolved[0].Cable.Type != "cat6a" {
		t.Errorf("expected cable type cat6a, got %s", resolved[0].Cable.Type)
	}
	if resolved[0].Cable.Color != "blue" {
		t.Errorf("expected cable color blue, got %s", resolved[0].Cable.Color)
	}
}

func TestResolveConnectionMap_PerEntryOverride(t *testing.T) {
	inv := testInventoryWithDevices("node-01", "switch-01")

	length := 3.0
	cm := &ConnectionMap{
		Version:       "v1",
		CableDefaults: &CableDefaults{Type: "cat6a", Color: "blue"},
		Connections: []ConnectionEntry{
			{
				A:     Endpoint{Device: "node-01", Port: "eth0"},
				B:     Endpoint{Device: "switch-01", Port: "eth0"},
				Cable: &CableProps{Type: "dac-passive", Length: &length},
			},
		},
	}

	resolved, errs := ResolveConnectionMap(cm, inv)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	c := resolved[0].Cable
	if c.Type != "dac-passive" {
		t.Errorf("expected type override dac-passive, got %s", c.Type)
	}
	if c.Color != "blue" {
		t.Errorf("expected inherited color blue, got %s", c.Color)
	}
	if c.Length == nil || *c.Length != 3.0 {
		t.Errorf("expected length 3.0, got %v", c.Length)
	}
}

func TestResolveConnectionMap_PatternLengthMismatch(t *testing.T) {
	inv := testInventoryWithDevices("node-01", "node-02", "switch-01")

	cm := &ConnectionMap{
		Version: "v1",
		Connections: []ConnectionEntry{
			{
				A: Endpoint{Device: "node-{01..02}", Port: "eth0"},
				B: Endpoint{Device: "switch-01", Port: "eth{1..3}"},
			},
		},
	}

	_, errs := ResolveConnectionMap(cm, inv)
	if len(errs) == 0 {
		t.Fatal("expected error for pattern length mismatch")
	}
}

func TestZipCount(t *testing.T) {
	tests := []struct {
		name    string
		lengths []int
		want    int
		wantErr bool
	}{
		{"all ones", []int{1, 1, 1, 1}, 1, false},
		{"one expanded", []int{4, 1, 1, 1}, 4, false},
		{"two matching", []int{4, 1, 4, 1}, 4, false},
		{"mismatch", []int{3, 1, 4, 1}, 0, true},
		{"all zero", []int{0, 0, 0, 0}, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ZipCount(tt.lengths...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ZipCount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ZipCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestBroadcast(t *testing.T) {
	result := Broadcast([]string{"x"}, 3)
	if len(result) != 3 || result[0] != "x" || result[2] != "x" {
		t.Errorf("Broadcast single: got %v", result)
	}

	result = Broadcast([]string{"a", "b", "c"}, 3)
	if len(result) != 3 || result[0] != "a" {
		t.Errorf("Broadcast passthrough: got %v", result)
	}

	result = Broadcast(nil, 2)
	if len(result) != 2 || result[0] != "" {
		t.Errorf("Broadcast nil: got %v", result)
	}
}
