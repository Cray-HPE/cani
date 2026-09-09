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
package imprt

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestRemapCableTerminations verifies cable endpoint device UUIDs are re-pointed
// through the merge remap (the fix for merge-import idempotency), while
// unmapped endpoints and interface UUIDs are left untouched.
func TestRemapCableTerminations(t *testing.T) {
	ephemeralA, existingA := uuid.New(), uuid.New()
	ephemeralB, existingB := uuid.New(), uuid.New()
	ifaceA, ifaceB := uuid.New(), uuid.New()
	unmapped := uuid.New()

	cableID := uuid.New()
	cables := map[uuid.UUID]*devicetypes.CaniCableType{
		cableID: {
			ID:                 cableID,
			TerminationADevice: ephemeralA,
			TerminationBDevice: ephemeralB,
			TerminationA:       ifaceA,
			TerminationB:       ifaceB,
		},
	}

	remapCableTerminations(cables, map[uuid.UUID]uuid.UUID{
		ephemeralA: existingA,
		ephemeralB: existingB,
		unmapped:   uuid.New(),
	})

	got := cables[cableID]
	if got.TerminationADevice != existingA {
		t.Errorf("TerminationADevice = %s, want %s", got.TerminationADevice, existingA)
	}
	if got.TerminationBDevice != existingB {
		t.Errorf("TerminationBDevice = %s, want %s", got.TerminationBDevice, existingB)
	}
	// Interface UUIDs are stable and must not be rewritten.
	if got.TerminationA != ifaceA || got.TerminationB != ifaceB {
		t.Errorf("interface terminations were modified: A=%s B=%s", got.TerminationA, got.TerminationB)
	}
}

// TestRemapCableTerminationsNoRemap verifies an empty remap is a no-op.
func TestRemapCableTerminationsNoRemap(t *testing.T) {
	devA, devB := uuid.New(), uuid.New()
	cableID := uuid.New()
	cables := map[uuid.UUID]*devicetypes.CaniCableType{
		cableID: {ID: cableID, TerminationADevice: devA, TerminationBDevice: devB},
	}

	remapCableTerminations(cables, nil)

	if cables[cableID].TerminationADevice != devA || cables[cableID].TerminationBDevice != devB {
		t.Error("expected terminations unchanged with an empty remap")
	}
}
