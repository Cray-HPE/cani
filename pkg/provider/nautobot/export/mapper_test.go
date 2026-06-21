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

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// TestMakeStatusRef verifies a status UUID round-trips into the Nautobot status
// reference union type used by write requests.
//
// Why it matters: every device/cable/interface the export writes carries a
// status reference; if the UUID is mangled when wrapped in the API union the
// remote object would be created with the wrong (or no) status.
// Inputs: a uuid.UUID. Outputs: a status ref whose embedded ID must decode back
// to the original UUID.
// Data choice: a fixed all-ones UUID makes the round-trip assertion readable and
// uuid.Nil confirms the helper still produces a well-formed ref for the zero
// value.
func TestMakeStatusRef(t *testing.T) {
	tests := []struct {
		name string
		id   uuid.UUID
	}{
		{
			name: "creates ref from valid UUID",
			id:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		},
		{
			name: "creates ref from nil UUID",
			id:   uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := makeStatusRef(tt.id)
			if ref.Id == nil {
				t.Fatal("expected ref.Id to be non-nil")
			}
			got, err := ref.Id.AsBulkWritableCableRequestStatusId0()
			if err != nil {
				t.Fatalf("unexpected error extracting UUID: %v", err)
			}
			if uuid.UUID(got) != tt.id {
				t.Errorf("makeStatusRef() round-trip = %s, want %s", uuid.UUID(got), tt.id)
			}
		})
	}
}

// TestResolveFace verifies rack-face resolution maps "rear" to rear and defaults
// everything else (including empty and unknown values) to front.
//
// Why it matters: rack-mounted devices exported to Nautobot must declare a
// mounting face; defaulting unknown/empty input to front keeps the export from
// failing on incomplete cani data while preserving an explicit "rear".
// Inputs: a face string. Outputs: a *RackFace decoding to a FaceEnum.
// Data choice: "rear" (the one non-default), "" (missing data) and "top" (an
// invalid value) cover the explicit, empty, and fallback branches.
func TestResolveFace(t *testing.T) {
	tests := []struct {
		name     string
		face     string
		expected nautobotapi.FaceEnum
	}{
		{
			name:     "rear returns rear",
			face:     "rear",
			expected: nautobotapi.FaceEnumRear,
		},
		{
			name:     "empty defaults to front",
			face:     "",
			expected: nautobotapi.FaceEnumFront,
		},
		{
			name:     "unknown defaults to front",
			face:     "top",
			expected: nautobotapi.FaceEnumFront,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveFace(tt.face)
			if got == nil {
				t.Fatal("expected non-nil RackFace")
			}
			val, err := got.AsFaceEnum()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val != tt.expected {
				t.Errorf("resolveFace(%q) = %v, want %v", tt.face, val, tt.expected)
			}
		})
	}
}

// TestNewDeviceMapper verifies the mapper constructor stores its cache and
// defaults by reference.
//
// Why it matters: the DeviceMapper turns cani devices into Nautobot write
// requests; it must share the live lookup cache (so resolved IDs are reused) and
// honor the caller's default location/role/status.
// Inputs: a *LookupCache and *MapperOpts. Outputs: a wired *DeviceMapper.
// Data choice: one case provides populated defaults and one provides an empty
// struct to confirm the constructor stores whatever it is given without
// substituting its own values.
func TestNewDeviceMapper(t *testing.T) {
	tests := []struct {
		name     string
		defaults *MapperOpts
	}{
		{
			name: "creates mapper with defaults",
			defaults: &MapperOpts{
				DefaultLocation: "SiteA",
				DefaultRole:     "Generic",
				DefaultStatus:   "Active",
				Strict:          false,
			},
		},
		{
			name:     "creates mapper with empty defaults",
			defaults: &MapperOpts{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := NewNautobotClient("http://localhost/api", "token")
			cache := NewLookupCache(client)
			mapper := NewDeviceMapper(cache, tt.defaults)
			if mapper == nil {
				t.Fatal("expected non-nil mapper")
			}
			if mapper.cache != cache {
				t.Error("mapper.cache does not match input")
			}
			if mapper.defaults != tt.defaults {
				t.Error("mapper.defaults does not match input")
			}
		})
	}
}
