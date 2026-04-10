/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func TestGenerateDeviceNames(t *testing.T) {
	tests := []struct {
		name         string
		inventory    *devicetypes.Inventory
		checkDevice  uuid.UUID
		expectedName string
	}{
		{
			name: "unnamed device gets name from serial",
			inventory: func() *devicetypes.Inventory {
				id := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
				return &devicetypes.Inventory{
					Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
						id: {
							ID:           id,
							Name:         "",
							Type: devicetypes.Type("node"),
							Serial:       "SN123",
						},
					},
				}
			}(),
			checkDevice:  uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			expectedName: "cani-SN123",
		},
		{
			name: "already-named device is not changed",
			inventory: func() *devicetypes.Inventory {
				id := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
				return &devicetypes.Inventory{
					Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
						id: {
							ID:           id,
							Name:         "my-server",
							Type: devicetypes.Type("node"),
						},
					},
				}
			}(),
			checkDevice:  uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			expectedName: "my-server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generateDeviceNames(tt.inventory)
			got := tt.inventory.Devices[tt.checkDevice].Name
			if got != tt.expectedName {
				t.Errorf("generateDeviceNames() device name = %q, want %q", got, tt.expectedName)
			}
		})
	}
}

func TestDisambiguateDeviceNames(t *testing.T) {
	t.Run("duplicate names get suffixed", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				id1: {ID: id1, Name: "server", Type: devicetypes.Type("node"), Serial: "SN-A"},
				id2: {ID: id2, Name: "server", Type: devicetypes.Type("node"), Serial: "SN-B"},
			},
		}

		disambiguateDeviceNames(inv)

		name1 := inv.Devices[id1].Name
		name2 := inv.Devices[id2].Name
		if name1 == name2 {
			t.Errorf("names should be unique after disambiguation: %q == %q", name1, name2)
		}
	})

	t.Run("unique names are not changed", func(t *testing.T) {
		id1 := uuid.New()
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				id1: {ID: id1, Name: "unique-server", Type: devicetypes.Type("node")},
			},
		}

		disambiguateDeviceNames(inv)

		if inv.Devices[id1].Name != "unique-server" {
			t.Errorf("unique name should not change, got %q", inv.Devices[id1].Name)
		}
	})
}

func TestSetExternalID(t *testing.T) {
	t.Run("initializes nil map and sets value", func(t *testing.T) {
		var m map[string]uuid.UUID
		id := uuid.New()

		setExternalID(&m, "nautobot", id)

		if m == nil {
			t.Fatal("expected map to be initialized")
		}
		if m["nautobot"] != id {
			t.Errorf("setExternalID() = %s, want %s", m["nautobot"], id)
		}
	})

	t.Run("overwrites existing value", func(t *testing.T) {
		oldID := uuid.New()
		newID := uuid.New()
		m := map[string]uuid.UUID{"nautobot": oldID}

		setExternalID(&m, "nautobot", newID)

		if m["nautobot"] != newID {
			t.Errorf("setExternalID() = %s, want %s", m["nautobot"], newID)
		}
	})
}

func TestContainsInfiniband(t *testing.T) {
	tests := []struct {
		name     string
		device   *devicetypes.CaniDeviceType
		expected bool
	}{
		{
			name:     "model containing NDR returns true",
			device:   &devicetypes.CaniDeviceType{Model: "Quantum-2 NDR Switch"},
			expected: true,
		},
		{
			name:     "model containing HDR returns true",
			device:   &devicetypes.CaniDeviceType{Model: "ConnectX-7 HDR Adapter"},
			expected: true,
		},
		{
			name:     "model containing MCX returns true",
			device:   &devicetypes.CaniDeviceType{Model: "MCX75310AAS-NEAT"},
			expected: true,
		},
		{
			name:     "plain server model returns false",
			device:   &devicetypes.CaniDeviceType{Model: "ProLiant DL380"},
			expected: false,
		},
		{
			name:     "empty model returns false",
			device:   &devicetypes.CaniDeviceType{Model: ""},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsInfiniband(tt.device)
			if got != tt.expected {
				t.Errorf("containsInfiniband() = %t, want %t", got, tt.expected)
			}
		})
	}
}

func TestMapInterfaceType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "1000base-t passes through",
			input:    "1000base-t",
			expected: "1000base-t",
		},
		{
			name:     "10gbase-x-sfpp maps correctly",
			input:    "10gbase-x-sfpp",
			expected: "10gbase-x-sfpp",
		},
		{
			name:     "400gbase-x-osfp maps correctly",
			input:    "400gbase-x-osfp",
			expected: "400gbase-x-osfp",
		},
		{
			name:     "infiniband-ndr maps correctly",
			input:    "infiniband-ndr",
			expected: "infiniband-ndr",
		},
		{
			name:     "unknown type passes through as-is",
			input:    "custom-type",
			expected: "custom-type",
		},
		{
			name:     "empty string defaults to 1000base-t",
			input:    "",
			expected: "1000base-t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapInterfaceType(tt.input)
			if got != tt.expected {
				t.Errorf("mapInterfaceType(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetSpeedForType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "100base-tx returns 100Mbps",
			input:    "100base-tx",
			expected: 100000,
		},
		{
			name:     "1000base-t returns 1Gbps",
			input:    "1000base-t",
			expected: 1000000,
		},
		{
			name:     "400gbase-x-osfp returns 400Gbps",
			input:    "400gbase-x-osfp",
			expected: 400000000,
		},
		{
			name:     "infiniband-hdr returns 200Gbps",
			input:    "infiniband-hdr",
			expected: 200000000,
		},
		{
			name:     "unknown type defaults to 1Gbps",
			input:    "unknown-type",
			expected: 1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSpeedForType(tt.input)
			if got != tt.expected {
				t.Errorf("getSpeedForType(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetDeviceInterfaceSpecs(t *testing.T) {
	t.Run("device with explicit interfaces uses those", func(t *testing.T) {
		device := &devicetypes.CaniDeviceType{
			Type: devicetypes.Type("node"),
			Interfaces: []devicetypes.InterfaceSpec{
				{Name: "eth0", Type: "1000base-t"},
				{Name: "eth1", Type: "10gbase-x-sfpp"},
			},
		}

		specs := getDeviceInterfaceSpecs(device)

		if len(specs) != 2 {
			t.Fatalf("expected 2 specs, got %d", len(specs))
		}
		if specs[0].Name != "eth0" {
			t.Errorf("specs[0].Name = %q, want %q", specs[0].Name, "eth0")
		}
		if specs[1].Type != "10gbase-x-sfpp" {
			t.Errorf("specs[1].Type = %q, want %q", specs[1].Type, "10gbase-x-sfpp")
		}
	})

	t.Run("PDU device without interfaces gets defaults", func(t *testing.T) {
		device := &devicetypes.CaniDeviceType{
			Type: devicetypes.Type("cabinet-pdu"),
		}

		specs := getDeviceInterfaceSpecs(device)

		if len(specs) != 1 {
			t.Fatalf("expected 1 spec for PDU, got %d", len(specs))
		}
		if specs[0].Name != "mgmt0" {
			t.Errorf("specs[0].Name = %q, want %q", specs[0].Name, "mgmt0")
		}
	})
}

func TestResolveCableType(t *testing.T) {
	t.Run("explicit cable type resolves", func(t *testing.T) {
		cable := &devicetypes.CaniCableType{CableType: "cat6"}
		got := resolveCableType(cable)
		if got == nil {
			t.Error("expected non-nil cable type for cat6")
		}
	})

	t.Run("cable category resolves", func(t *testing.T) {
		cable := &devicetypes.CaniCableType{CableCategory: "dac-passive"}
		got := resolveCableType(cable)
		if got == nil {
			t.Error("expected non-nil cable type for dac-passive category")
		}
	})

	t.Run("connector type resolves", func(t *testing.T) {
		cable := &devicetypes.CaniCableType{ConnectorType: "rj45"}
		got := resolveCableType(cable)
		if got == nil {
			t.Error("expected non-nil cable type for rj45 connector")
		}
	})

	t.Run("unknown cable returns nil", func(t *testing.T) {
		cable := &devicetypes.CaniCableType{
			CableType:     "unknown-xyz",
			CableCategory: "unknown-abc",
			ConnectorType: "unknown-123",
			Slug:          "no-match",
		}
		got := resolveCableType(cable)
		if got != nil {
			t.Error("expected nil for unrecognized cable type")
		}
	})
}

func TestColorNameToHex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "named color converts to hex",
			input:    "red",
			expected: "ff0000",
		},
		{
			name:     "case-insensitive named color",
			input:    "Blue",
			expected: "0000ff",
		},
		{
			name:     "hex passthrough without hash",
			input:    "aa11bb",
			expected: "aa11bb",
		},
		{
			name:     "hex passthrough with hash stripped",
			input:    "#00ff00",
			expected: "00ff00",
		},
		{
			name:     "unknown name passes through as lowercase",
			input:    "Magenta",
			expected: "magenta",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorNameToHex(tt.input)
			if got != tt.expected {
				t.Errorf("colorNameToHex(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
