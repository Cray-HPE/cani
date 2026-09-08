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
	"encoding/json"
	"testing"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/transform"
	"github.com/google/uuid"
)

// TestGeneratedShapeConformance pins the generated request shapes consumed by
// the provider adapters. A regenerated client that changes one of these shapes
// should fail here before request construction can silently drift.
func TestGeneratedShapeConformance(t *testing.T) {
	t.Run("scalar reference", func(t *testing.T) {
		id := uuid.New()
		var req nautobotapi.WritableDeviceRequest
		mustSetRefID(t, &req.DeviceType, id)
		if got := transform.RefUUID(req.DeviceType.Id); got != id {
			t.Fatalf("device type ID = %s, want %s", got, id)
		}
	})

	t.Run("pointer reference", func(t *testing.T) {
		id := uuid.New()
		var req nautobotapi.LocationRequest
		mustSetRefID(t, &req.Parent, id)
		if req.Parent == nil {
			t.Fatal("parent reference is nil")
		}
		if got := transform.RefUUID(req.Parent.Id); got != id {
			t.Fatalf("parent ID = %s, want %s", got, id)
		}
	})

	t.Run("reference slice", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}
		var req nautobotapi.WritableDeviceRequest
		if err := setRefSlice(&req.Tags, ids); err != nil {
			t.Fatalf("setRefSlice() error = %v", err)
		}
		if req.Tags == nil || len(*req.Tags) != len(ids) {
			t.Fatalf("tags length = %d, want %d", refSliceLen(req.Tags), len(ids))
		}
		for i := range ids {
			if got := transform.RefUUID((*req.Tags)[i].Id); got != ids[i] {
				t.Errorf("tag %d ID = %s, want %s", i, got, ids[i])
			}
		}
	})

	t.Run("custom fields", func(t *testing.T) {
		req := nautobotapi.WritableDeviceRequest{
			CustomFields: toNautobotCustomFields(map[string]interface{}{
				"tier":  "gold",
				"clear": nil,
			}),
		}
		blob, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("marshal custom fields: %v", err)
		}
		var payload map[string]map[string]interface{}
		if err := json.Unmarshal(blob, &payload); err != nil {
			t.Fatalf("unmarshal custom fields: %v", err)
		}
		if got := payload["custom_fields"]["tier"]; got != "gold" {
			t.Errorf("custom_fields.tier = %v, want gold", got)
		}
		if got, ok := payload["custom_fields"]["clear"]; !ok || got != nil {
			t.Errorf("custom_fields.clear = %v (present %t), want explicit null", got, ok)
		}
	})

	t.Run("explicit null patch field", func(t *testing.T) {
		blob, err := json.Marshal(interfacePatch{})
		if err != nil {
			t.Fatalf("marshal interface patch: %v", err)
		}
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(blob, &payload); err != nil {
			t.Fatalf("unmarshal interface patch: %v", err)
		}
		if got := string(payload["role"]); got != "null" {
			t.Fatalf("role = %s, want explicit null", got)
		}
	})
}

func refSliceLen[T any](refs *[]T) int {
	if refs == nil {
		return 0
	}
	return len(*refs)
}
