/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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

// +----------------------------------------------+------------------------------------------------+--------------------------------------------------+
// | Function                                      | Happy-path test                                | Failure test                                     |
// +----------------------------------------------+------------------------------------------------+--------------------------------------------------+
// | RebuildProviderKeyIndex                        | TestRebuildIndexPopulatesFromDevices            | TestRebuildIndexEmptyInventory                   |
// | lookupProviderKey                              | TestLookupProviderKeyFindsMatch                | TestLookupProviderKeyNoMatch                     |
// | indexDevice / unindexDevice                    | TestIndexDeviceAddsEntries                     | TestUnindexDeviceRemovesEntries                  |
// | toIndexValue                                   | TestToIndexValueStringTypes                    | TestToIndexValueNilReturnsEmpty                  |
// | FindDeviceByProviderKey (indexed)              | TestFindDeviceByProviderKeyUsesIndex           | TestFindDeviceByProviderKeyFallsBack             |
// | MergeDevicesStrict index maintenance            | TestMergeDevicesStrictMaintainsIndex           | (covered by merge tests)                         |
// +----------------------------------------------+------------------------------------------------+--------------------------------------------------+

package devicetypes

import (
	"testing"

	"github.com/google/uuid"
)

// ---------- RebuildProviderKeyIndex ----------

func TestRebuildIndexPopulatesFromDevices(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Devices[id] = &CaniDeviceType{
		ID:   id,
		Name: "server1",
		ProviderMetadata: map[string]any{
			"redfish": map[string]any{
				"redfish_uuid": "abc-123",
				"bmc_fqdn":     "bmc1.example.com",
			},
		},
	}

	inv.RebuildProviderKeyIndex()

	got := inv.lookupProviderKey("redfish", "redfish_uuid", "abc-123")
	if got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}

	got = inv.lookupProviderKey("redfish", "bmc_fqdn", "bmc1.example.com")
	if got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

func TestRebuildIndexEmptyInventory(t *testing.T) {
	inv := NewInventory()
	inv.RebuildProviderKeyIndex()

	got := inv.lookupProviderKey("redfish", "redfish_uuid", "abc")
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}

// ---------- lookupProviderKey ----------

func TestLookupProviderKeyFindsMatch(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	inv.Devices[id] = &CaniDeviceType{
		ID:   id,
		Name: "node1",
		ProviderMetadata: map[string]any{
			"csm": map[string]any{"xname": "x1000c0s0b0n0"},
		},
	}
	inv.RebuildProviderKeyIndex()

	got := inv.lookupProviderKey("csm", "xname", "x1000c0s0b0n0")
	if got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

func TestLookupProviderKeyNoMatch(t *testing.T) {
	inv := NewInventory()
	inv.RebuildProviderKeyIndex()

	got := inv.lookupProviderKey("csm", "xname", "x9999")
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}

// ---------- indexDevice / unindexDevice ----------

func TestIndexDeviceAddsEntries(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	dev := &CaniDeviceType{
		ID:   id,
		Name: "test",
		ProviderMetadata: map[string]any{
			"hpcm": map[string]any{"hpcm_uuid": "hpcm-111"},
		},
	}

	inv.indexDevice(id, dev)

	got := inv.lookupProviderKey("hpcm", "hpcm_uuid", "hpcm-111")
	if got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

func TestUnindexDeviceRemovesEntries(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	dev := &CaniDeviceType{
		ID:   id,
		Name: "test",
		ProviderMetadata: map[string]any{
			"hpcm": map[string]any{"hpcm_uuid": "hpcm-222"},
		},
	}
	inv.indexDevice(id, dev)
	inv.unindexDevice(id, dev)

	got := inv.lookupProviderKey("hpcm", "hpcm_uuid", "hpcm-222")
	if got != uuid.Nil {
		t.Fatalf("expected uuid.Nil after unindex, got %s", got)
	}
}

// ---------- toIndexValue ----------

func TestToIndexValueStringTypes(t *testing.T) {
	tests := []struct {
		input any
		want  string
	}{
		{"hello", "hello"},
		{42, "42"},
		{int64(99), "99"},
		{3.14, "3.14"},
		{true, "true"},
		{false, "false"},
	}
	for _, tt := range tests {
		got := toIndexValue(tt.input)
		if got != tt.want {
			t.Errorf("toIndexValue(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestToIndexValueNilReturnsEmpty(t *testing.T) {
	if got := toIndexValue(nil); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
	// Slices/maps should also return empty.
	if got := toIndexValue([]string{"a"}); got != "" {
		t.Fatalf("expected empty for slice, got %q", got)
	}
}

// ---------- FindDeviceByProviderKey (indexed path) ----------

func TestFindDeviceByProviderKeyUsesIndex(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	dev := &CaniDeviceType{
		ID:   id,
		Name: "indexed-server",
		ProviderMetadata: map[string]any{
			"redfish": map[string]any{"redfish_uuid": "rf-indexed"},
		},
	}
	inv.Devices[id] = dev
	inv.RebuildProviderKeyIndex()

	got := inv.FindDeviceByProviderKey("redfish", "redfish_uuid", "rf-indexed")
	if got == nil {
		t.Fatal("expected to find device via index")
	}
	if got.ID != id {
		t.Fatalf("expected ID %s, got %s", id, got.ID)
	}
}

func TestFindDeviceByProviderKeyFallsBack(t *testing.T) {
	inv := NewInventory()
	id := uuid.New()
	dev := &CaniDeviceType{
		ID:   id,
		Name: "fallback-server",
		ProviderMetadata: map[string]any{
			"redfish": map[string]any{"redfish_uuid": "rf-fallback"},
		},
	}
	inv.Devices[id] = dev
	// Don't build the index — forces fallback to linear scan.

	got := inv.FindDeviceByProviderKey("redfish", "redfish_uuid", "rf-fallback")
	if got == nil {
		t.Fatal("expected to find device via fallback scan")
	}
	if got.ID != id {
		t.Fatalf("expected ID %s, got %s", id, got.ID)
	}
}

// ---------- MergeDevicesStrict index maintenance ----------

func TestMergeDevicesStrictMaintainsIndex(t *testing.T) {
	inv := NewInventory()
	inv.RebuildProviderKeyIndex()

	// Insert a new device via merge.
	id := uuid.New()
	dev := &CaniDeviceType{
		ID:   id,
		Name: "merge-test",
		ProviderMetadata: map[string]any{
			"csm": map[string]any{"xname": "x3000c0s0b0n0"},
		},
	}
	inv.MergeDevicesStrict(map[uuid.UUID]*CaniDeviceType{id: dev}, false)

	// Verify the index was updated.
	got := inv.FindDeviceByProviderKey("csm", "xname", "x3000c0s0b0n0")
	if got == nil {
		t.Fatal("expected device to be findable after merge")
	}
	if got.ID != id {
		t.Fatalf("expected ID %s, got %s", id, got.ID)
	}

	// Now merge an update that changes the xname.
	updated := &CaniDeviceType{
		ID:   id,
		Name: "merge-test",
		ProviderMetadata: map[string]any{
			"csm": map[string]any{"xname": "x3000c0s1b0n0"},
		},
	}
	inv.MergeDevicesStrict(map[uuid.UUID]*CaniDeviceType{id: updated}, false)

	// Old xname should no longer match.
	old := inv.FindDeviceByProviderKey("csm", "xname", "x3000c0s0b0n0")
	if old != nil {
		t.Fatal("old xname should not match after update")
	}

	// New xname should match.
	newDev := inv.FindDeviceByProviderKey("csm", "xname", "x3000c0s1b0n0")
	if newDev == nil {
		t.Fatal("new xname should match after update")
	}
}
