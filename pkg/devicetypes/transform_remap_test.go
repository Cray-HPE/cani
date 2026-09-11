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

func TestTransformResultRemapReferencesRemapsVRFDevices(t *testing.T) {
	incomingDeviceID, existingDeviceID := uuid.New(), uuid.New()
	vrfID := uuid.New()
	result := &TransformResult{
		VRFs: map[uuid.UUID]*CaniVRF{
			vrfID: {ID: vrfID, Name: "BLUE", Devices: []uuid.UUID{incomingDeviceID}},
		},
	}

	result.RemapReferences(nil, ReferenceRemaps{
		Devices: map[uuid.UUID]uuid.UUID{incomingDeviceID: existingDeviceID},
	})

	if got := result.VRFs[vrfID].Devices; len(got) != 1 || got[0] != existingDeviceID {
		t.Fatalf("VRF devices = %v, want [%s]", got, existingDeviceID)
	}
}

func TestTransformResultRemapReferencesIsIdempotent(t *testing.T) {
	incomingDeviceID, existingDeviceID := uuid.New(), uuid.New()
	vrfID := uuid.New()
	result := &TransformResult{
		VRFs: map[uuid.UUID]*CaniVRF{
			vrfID: {ID: vrfID, Name: "BLUE", Devices: []uuid.UUID{incomingDeviceID}},
		},
	}
	remaps := ReferenceRemaps{Devices: map[uuid.UUID]uuid.UUID{incomingDeviceID: existingDeviceID}}

	result.RemapReferences(nil, remaps)
	result.RemapReferences(nil, remaps)

	if got := result.VRFs[vrfID].Devices; len(got) != 1 || got[0] != existingDeviceID {
		t.Fatalf("VRF devices = %v, want [%s]", got, existingDeviceID)
	}
}

func TestMergeTransformResultPreservesDynamicMetadataTypes(t *testing.T) {
	deviceID := uuid.New()
	inventory := NewInventory()
	inventory.Devices[deviceID] = &CaniDeviceType{
		ID: deviceID, Name: "server-1",
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{
			"example": map[string]any{"ordinal": 7},
		}},
	}

	if _, err := inventory.MergeTransformResult(&TransformResult{}); err != nil {
		t.Fatalf("MergeTransformResult() error = %v", err)
	}
	metadata, ok := inventory.Devices[deviceID].ProviderMetadata["example"].(map[string]any)
	if !ok {
		t.Fatalf("provider metadata type = %T, want map[string]any", inventory.Devices[deviceID].ProviderMetadata["example"])
	}
	if _, ok := metadata["ordinal"].(int); !ok {
		t.Fatalf("ordinal type = %T, want int", metadata["ordinal"])
	}
}
