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
package validate

import (
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// -----------------------------------------------------------------------------
// Status — builtin-only validation
// -----------------------------------------------------------------------------

// TestStatus_NormalizesValidInputCaseInsensitively covers the happy path: a
// recognised status is accepted regardless of case and returned in canonical
// Title-case form.
func TestStatus_NormalizesValidInputCaseInsensitively(t *testing.T) {
	got, err := Status("active")
	if err != nil {
		t.Fatalf("Status(active) returned error: %v", err)
	}
	if got != "Active" {
		t.Errorf("Status(active) = %q, want %q", got, "Active")
	}
}

// TestStatus_RejectsUnknownValue covers the failing path for an unrecognised
// status.
func TestStatus_RejectsUnknownValue(t *testing.T) {
	got, err := Status("totally-bogus")
	if err == nil {
		t.Fatal("expected an error for an unknown status")
	}
	if got != "" {
		t.Errorf("expected empty canonical value on error, got %q", got)
	}
}

// TestStatus_RejectsStaged proves the user-facing validator excludes the
// internal-only "Staged" status even though it is otherwise a known value.
func TestStatus_RejectsStaged(t *testing.T) {
	if _, err := Status("staged"); err == nil {
		t.Error("expected Status(staged) to be rejected for user input")
	}
}

// -----------------------------------------------------------------------------
// StatusWithInventory — builtin plus custom catalog statuses
// -----------------------------------------------------------------------------

// TestStatusWithInventory_ResolvesBuiltinBeforeCustom confirms a builtin status
// is accepted without consulting the inventory catalog.
func TestStatusWithInventory_ResolvesBuiltinBeforeCustom(t *testing.T) {
	got, err := StatusWithInventory("active", &devicetypes.Inventory{})
	if err != nil {
		t.Fatalf("StatusWithInventory(active) returned error: %v", err)
	}
	if got != "Active" {
		t.Errorf("StatusWithInventory(active) = %q, want %q", got, "Active")
	}
}

// TestStatusWithInventory_ResolvesCustomStatus covers the path where a value
// is not a builtin but is registered in the inventory metadata catalog. The
// returned value preserves the catalog's canonical casing.
func TestStatusWithInventory_ResolvesCustomStatus(t *testing.T) {
	inv := &devicetypes.Inventory{
		Metadata: &devicetypes.InventoryMetadata{
			Statuses: []devicetypes.MetadataEntry{{Name: "Burned-In"}},
		},
	}

	got, err := StatusWithInventory("burned-in", inv)
	if err != nil {
		t.Fatalf("StatusWithInventory(custom) returned error: %v", err)
	}
	if got != "Burned-In" {
		t.Errorf("StatusWithInventory(burned-in) = %q, want %q", got, "Burned-In")
	}
}

// TestStatusWithInventory_InvalidErrorListsCustomOptions verifies the error for
// an unknown value names the offending input and surfaces the available custom
// statuses so the user can correct it.
func TestStatusWithInventory_InvalidErrorListsCustomOptions(t *testing.T) {
	inv := &devicetypes.Inventory{
		Metadata: &devicetypes.InventoryMetadata{
			Statuses: []devicetypes.MetadataEntry{{Name: "Burned-In"}},
		},
	}

	_, err := StatusWithInventory("bogus", inv)
	if err == nil {
		t.Fatal("expected an error for an unknown status")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error %q should echo the invalid input", err)
	}
	if !strings.Contains(err.Error(), "Burned-In") {
		t.Errorf("error %q should list the available custom status", err)
	}
}

// TestStatusWithInventory_NilInventoryFallsBackToBuiltinError ensures a nil
// inventory is handled gracefully and the error omits any custom-status clause.
func TestStatusWithInventory_NilInventoryFallsBackToBuiltinError(t *testing.T) {
	_, err := StatusWithInventory("bogus", nil)
	if err == nil {
		t.Fatal("expected an error for an unknown status with nil inventory")
	}
	if strings.Contains(err.Error(), "custom status") {
		t.Errorf("error %q should not mention custom statuses when inventory is nil", err)
	}
}
