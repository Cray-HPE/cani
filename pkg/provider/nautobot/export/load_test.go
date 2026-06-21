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

// TestGenerateDeviceNames verifies generateDeviceNames assigns a cani-prefixed
// name derived from the serial to unnamed devices while leaving already-named
// devices untouched.
//
// Why it matters: Nautobot requires every device to have a name, so the
// exporter must synthesize stable names for inventory entries that lack one
// without clobbering names operators already set.
// Inputs: an Inventory with one unnamed node (serial "SN123") and one already
// named "my-server". Outputs: in-place names "cani-SN123" and "my-server".
// Data choice: a populated serial exercises the highest-priority naming source,
// while the pre-named device guards the no-overwrite branch.
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
							ID:     id,
							Name:   "",
							Type:   devicetypes.Type("node"),
							Serial: "SN123",
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
							ID:   id,
							Name: "my-server",
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

// TestDisambiguateDeviceNames verifies disambiguateDeviceNames suffixes
// colliding device names to make them unique and leaves singleton names alone.
//
// Why it matters: Nautobot enforces name uniqueness per location+tenant, so two
// inventory devices sharing a name must be split before export or one create
// will fail.
// Inputs: subtests with two nodes both named "server" (serials SN-A/SN-B) and a
// lone "unique-server". Outputs: the duplicates become distinct names; the
// unique name is unchanged.
// Data choice: distinct serials let the function append serial-based suffixes,
// its primary disambiguation strategy.
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

// TestSetExternalID verifies setExternalID lazily allocates a nil map and
// overwrites an existing provider entry.
//
// Why it matters: the exporter records the Nautobot UUID it assigns under the
// "nautobot" external-ID key so later phases and re-runs can correlate local
// inventory with remote records; that map is often nil on first write.
// Inputs: subtests passing a nil map, then a map with a stale "nautobot" UUID.
// Outputs: an initialized map and the key updated to the new UUID.
// Data choice: the nil-map and existing-key cases cover both branches of the
// pointer-to-map helper.
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

// TestContainsInfiniband verifies containsInfiniband flags models whose name
// contains InfiniBand markers (NDR, HDR, MCX) and rejects ordinary models.
//
// Why it matters: InfiniBand hardware drives extra ib* interfaces during export,
// so detecting it from the model string controls which interfaces get created
// in Nautobot.
// Inputs: device models such as "Quantum-2 NDR Switch", "ConnectX-7 HDR",
// "MCX75310AAS-NEAT", "ProLiant DL380", and "". Outputs: the matching bool.
// Data choice: each true case targets one substring marker, while the ProLiant
// and empty models confirm there are no false positives.
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

// TestMapInterfaceType verifies mapInterfaceType normalizes devicetypes
// interface strings to Nautobot enum values, passes unknown values through, and
// defaults empty input to 1000base-t.
//
// Why it matters: Nautobot rejects interface creates whose type is not a known
// enum, so the exporter must translate library types into accepted values.
// Inputs: representative types (1000base-t, 10gbase-x-sfpp, 400gbase-x-osfp,
// infiniband-ndr), an unknown "custom-type", and "". Outputs: the mapped string.
// Data choice: the samples span copper, SFP+, OSFP, and InfiniBand families plus
// the unknown and empty edge cases the default branch handles.
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

// TestGetSpeedForType verifies getSpeedForType returns the correct Kbps speed
// per interface type and falls back to 1Gbps for unknown types.
//
// Why it matters: Nautobot stores interface speed alongside type, so exported
// interfaces need accurate speeds to reflect real link capacity.
// Inputs: types 100base-tx, 1000base-t, 400gbase-x-osfp, infiniband-hdr, and an
// unknown type. Outputs: speeds in Kbps (e.g. 400000000 for 400G, 200000000 for
// HDR) and the 1000000 default.
// Data choice: the cases cover the fast-ethernet, gigabit, 400G, and InfiniBand
// rungs plus the default branch.
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

// TestGetDeviceInterfaceSpecs verifies getDeviceInterfaceSpecs prefers a
// device's explicitly instantiated interfaces and otherwise falls back to
// type-based defaults.
//
// Why it matters: these specs become the interfaces created in Nautobot, so the
// exporter must honor library-provided interfaces yet still give bare devices
// (like a PDU) a sane default management port.
// Inputs: subtests with a node carrying eth0/eth1 interfaces and a cabinet-pdu
// with none. Outputs: the two explicit specs, or a single "mgmt0" spec for the
// PDU.
// Data choice: a node with two interfaces checks order/type preservation, while
// the PDU exercises the simplest single-interface default branch.
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

// TestResolveCableType verifies resolveCableType resolves a Nautobot cable type
// from an explicit type, a category, or a connector, and returns nil when none
// match.
//
// Why it matters: Nautobot cables carry a typed enum, and the exporter must
// derive it from whichever cabling hint the inventory provides without guessing
// for genuinely unknown cables.
// Inputs: subtests with CableType "cat6", CableCategory "dac-passive",
// ConnectorType "rj45", and an all-unknown cable. Outputs: a non-nil cable type
// for the first three, nil for the last.
// Data choice: each populated field targets one resolution tier in priority
// order, and the all-unknown cable confirms the nil fallthrough.
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

// TestColorNameToHex verifies colorNameToHex maps known color names to 6-digit
// hex, strips a leading '#', and lowercases pass-through values.
//
// Why it matters: Nautobot stores cable colors as hex codes, so human-friendly
// color names from the inventory must be converted before export.
// Inputs: "red", "Blue", "aa11bb", "#00ff00", and "Magenta". Outputs: hex such
// as "ff0000"/"0000ff", the unchanged/normalized hex, and lowercased "magenta".
// Data choice: the cases cover a named color, case-insensitivity, bare hex,
// hash-prefixed hex, and an unknown name that falls through as lowercase.
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
