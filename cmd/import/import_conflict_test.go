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
	"fmt"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestMergeTransformResultNamesCascadeWhenThePrefixHoldsAddresses verifies the
// remove command an import names for a conflicting root prefix carries
// --cascade exactly when the plain removal would be refused.
//
// Why it matters: the printed remedy is followed literally; for a root prefix
// that holds addresses the plain command fails with a second error, so the
// hint must name the cascade up front, and must not name it when a plain
// removal would do.
// Inputs: the prefix conflict fixture, once with an address under the
// existing prefix and once without.
// Outputs: "cani alpha remove prefix <record> --cascade" with the address,
// no --cascade without it.
// Data choice: the same conflict with and without the held address isolates
// the flag to the model's refusal.
func TestMergeTransformResultNamesCascadeWhenThePrefixHoldsAddresses(t *testing.T) {
	for name, holdsAddress := range map[string]bool{"root prefix holding an address": true, "empty prefix": false} {
		t.Run(name, func(t *testing.T) { assertCascadeHint(t, holdsAddress) })
	}
}

// assertCascadeHint merges a conflicting prefix, with or without an address
// under the existing record, and checks the remedy names --cascade only when
// the address is there.
func assertCascadeHint(t *testing.T, holdsAddress bool) {
	t.Helper()
	inventory, result, record := conflictingIPAM(devicetypes.IPAMKindPrefix)
	if holdsAddress {
		address := uuid.New()
		inventory.IPAddresses[address] = &devicetypes.CaniIPAddress{ID: address, Host: "10.0.0.5", Address: "10.0.0.5/24", Parent: record}
	}

	err := mergeTransformResult(&etlContext{inventory: inventory}, result)

	command := fmt.Sprintf("cani alpha remove prefix %s", record)
	if err == nil || !strings.Contains(err.Error(), command) {
		t.Fatalf("mergeTransformResult() error = %v; want it to name %q", err, command)
	}
	if got := strings.Contains(err.Error(), command+" --cascade"); got != holdsAddress {
		t.Errorf("error names --cascade = %v, want %v: %v", got, holdsAddress, err)
	}
}
