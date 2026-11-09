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
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// deviceCreateEchoingDefined answers a device POST the way Nautobot does when
// only some of the sent custom-field keys are defined for dcim.device: it
// echoes those keys back and silently omits the rest.
func deviceCreateEchoingDefined(defined []string, sent *map[string]any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost || !strings.Contains(r.URL.Path, "dcim/devices") {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, emptyListJSON)
			return
		}
		var req struct {
			CustomFields map[string]any `json:"custom_fields"`
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &req)
		*sent = req.CustomFields
		echoed := map[string]any{}
		for _, key := range defined {
			echoed[key] = req.CustomFields[key]
		}
		resp, _ := json.Marshal(map[string]any{"id": "6f1c6a1e-0000-4000-8000-000000000001", "custom_fields": echoed})
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(resp)
	}
}

// TestCreateDevice_CountsCustomFieldsNautobotDropped verifies the read-back
// on a real device create: a sent key the response does not echo is counted
// as lossy, while the keys that arrive are not.
//
// Why it matters: Nautobot answers 201 even for an undefined custom-field
// key it will never expose, so a green create log is not evidence the value arrived;
// the Story requires every object that did not arrive as authored to be
// accounted for.
// Inputs: a device with explicit {decommissioned: false, spare_slots: 0,
// tier} and provider {xname}; a fake Nautobot that has definitions for
// decommissioned, spare_slots and xname only.
// Outputs: the request carried all four keys, result.CustomFieldsDropped
// == 1 and the device is still recorded as created.
// Data choice: the dropped key is the plain string so the two falsy values
// are shown to round-trip through the real JSON encoding.
func TestCreateDevice_CountsCustomFieldsNautobotDropped(t *testing.T) {
	var sent map[string]any
	e, cleanup := newExporterWithServer(t, deviceCreateEchoingDefined([]string{"decommissioned", "spare_slots", "xname"}, &sent))
	defer cleanup()
	seedDeviceRefs(t, e)

	result := &LoadResult{}
	if err := e.createDevice(context.Background(), explicitCustomFieldDevice("compute-001"), newCrudMapper(e), result); err != nil {
		t.Fatalf("createDevice() error = %v", err)
	}

	if len(sent) != 4 || sent["decommissioned"] != false || sent["spare_slots"] != float64(0) {
		t.Errorf("request custom_fields = %v, want all four keys with false and 0 intact", sent)
	}
	if result.CustomFieldsDropped != 1 {
		t.Errorf("CustomFieldsDropped = %d, want 1 (tier has no definition)", result.CustomFieldsDropped)
	}
	if !containsName(result.Created, "compute-001") {
		t.Errorf("Created = %v, want compute-001", result.Created)
	}
}
