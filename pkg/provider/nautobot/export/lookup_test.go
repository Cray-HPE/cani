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

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func TestToUUID(t *testing.T) {
	tests := []struct {
		name     string
		input    *openapi_types.UUID
		expected uuid.UUID
	}{
		{
			name: "valid pointer returns UUID",
			input: func() *openapi_types.UUID {
				u := openapi_types.UUID(uuid.MustParse("11111111-1111-1111-1111-111111111111"))
				return &u
			}(),
			expected: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		},
		{
			name:     "nil pointer returns nil UUID",
			input:    nil,
			expected: uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toUUID(tt.input)
			if got != tt.expected {
				t.Errorf("toUUID() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestNewLookupCache(t *testing.T) {
	t.Run("initializes all cache maps", func(t *testing.T) {
		client, _ := NewNautobotClient("http://localhost/api", "token")
		cache := NewLookupCache(client)

		if cache == nil {
			t.Fatal("expected non-nil cache")
		}
		if cache.deviceTypes == nil {
			t.Error("deviceTypes map not initialized")
		}
		if cache.locations == nil {
			t.Error("locations map not initialized")
		}
		if cache.statuses == nil {
			t.Error("statuses map not initialized")
		}
		if cache.roles == nil {
			t.Error("roles map not initialized")
		}
		if cache.devices == nil {
			t.Error("devices map not initialized")
		}
		if cache.manufacturers == nil {
			t.Error("manufacturers map not initialized")
		}
		if cache.interfaces == nil {
			t.Error("interfaces map not initialized")
		}
		if cache.tags == nil {
			t.Error("tags map not initialized")
		}
	})

	t.Run("sets client reference", func(t *testing.T) {
		client, _ := NewNautobotClient("http://localhost/api", "token")
		cache := NewLookupCache(client)

		if cache.client != client {
			t.Error("cache.client does not match input")
		}
	})
}

func TestInterfaceCacheKey(t *testing.T) {
	tests := []struct {
		name      string
		deviceID  uuid.UUID
		ifaceName string
		expected  string
	}{
		{
			name:      "builds correct key format",
			deviceID:  uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			ifaceName: "eth0",
			expected:  "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa:eth0",
		},
		{
			name:      "handles empty interface name",
			deviceID:  uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			ifaceName: "",
			expected:  "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interfaceCacheKey(tt.deviceID, tt.ifaceName)
			if got != tt.expected {
				t.Errorf("interfaceCacheKey() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNormalizeInterfaceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strips port prefix",
			input:    "port1",
			expected: "1",
		},
		{
			name:     "strips eth prefix from ethernet (eth matched first)",
			input:    "ethernet0",
			expected: "ernet0",
		},
		{
			name:     "strips eth prefix",
			input:    "eth0",
			expected: "0",
		},
		{
			name:     "lowercases and strips GigabitEthernet",
			input:    "GigabitEthernet1",
			expected: "1",
		},
		{
			name:     "plain number unchanged",
			input:    "42",
			expected: "42",
		},
		{
			name:     "strips leading separators after prefix",
			input:    "port-1",
			expected: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeInterfaceName(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeInterfaceName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractPortNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple number",
			input:    "1",
			expected: "1",
		},
		{
			name:     "port prefix",
			input:    "port1",
			expected: "1",
		},
		{
			name:     "hierarchical returns first number",
			input:    "1/0/1",
			expected: "1",
		},
		{
			name:     "no numbers returns empty string",
			input:    "mgmt",
			expected: "",
		},
		{
			name:     "GigabitEthernet extracts first number",
			input:    "GigabitEthernet1/0/1",
			expected: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPortNumber(tt.input)
			if got != tt.expected {
				t.Errorf("extractPortNumber(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFindNameByID(t *testing.T) {
	t.Run("nil UUID returns none", func(t *testing.T) {
		client, _ := NewNautobotClient("http://localhost/api", "token")
		cache := NewLookupCache(client)

		got := cache.FindNameByID("location", uuid.Nil)

		if got != "(none)" {
			t.Errorf("FindNameByID(nil) = %q, want %q", got, "(none)")
		}
	})

	t.Run("cached item found by ID", func(t *testing.T) {
		client, _ := NewNautobotClient("http://localhost/api", "token")
		cache := NewLookupCache(client)

		id := uuid.New()
		cache.CacheLocation("SiteA", &CachedItem{ID: id, Name: "SiteA"})

		got := cache.FindNameByID("location", id)

		if got != "SiteA" {
			t.Errorf("FindNameByID() = %q, want %q", got, "SiteA")
		}
	})

	t.Run("uncached ID returns UUID string", func(t *testing.T) {
		client, _ := NewNautobotClient("http://localhost/api", "token")
		cache := NewLookupCache(client)

		id := uuid.MustParse("33333333-3333-3333-3333-333333333333")
		got := cache.FindNameByID("location", id)

		if got != id.String() {
			t.Errorf("FindNameByID() = %q, want %q", got, id.String())
		}
	})
}

func TestCacheLocation(t *testing.T) {
	t.Run("stores and retrieves location", func(t *testing.T) {
		client, _ := NewNautobotClient("http://localhost/api", "token")
		cache := NewLookupCache(client)

		id := uuid.New()
		item := &CachedItem{ID: id, Name: "DC-East"}
		cache.CacheLocation("DC-East", item)

		cache.locationsMu.RLock()
		got, ok := cache.locations["DC-East"]
		cache.locationsMu.RUnlock()

		if !ok {
			t.Fatal("expected location to be cached")
		}
		if got.ID != id {
			t.Errorf("cached ID = %s, want %s", got.ID, id)
		}
	})

	t.Run("overwrites existing cached location", func(t *testing.T) {
		client, _ := NewNautobotClient("http://localhost/api", "token")
		cache := NewLookupCache(client)

		oldID := uuid.New()
		newID := uuid.New()
		cache.CacheLocation("Site", &CachedItem{ID: oldID, Name: "Site"})
		cache.CacheLocation("Site", &CachedItem{ID: newID, Name: "Site"})

		cache.locationsMu.RLock()
		got := cache.locations["Site"]
		cache.locationsMu.RUnlock()

		if got.ID != newID {
			t.Errorf("cached ID = %s, want %s (overwritten)", got.ID, newID)
		}
	})
}
