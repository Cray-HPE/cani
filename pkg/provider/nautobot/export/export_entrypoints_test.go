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
	"errors"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// ---------- PrintSummary (smoke) ----------

// TestPrintSummaryDoesNotPanic verifies PrintSummary renders a fully populated
// LoadResult without panicking.
//
// Why it matters: the summary is the operator's end-of-run report; a nil-deref
// or format panic there would mask an otherwise successful export.
// Inputs: a LoadResult with every counter and slice field populated.
// Outputs: none asserted beyond no panic (function only logs).
// Data choice: filling every field exercises all conditional print branches at
// once. Note: this is a smoke test and asserts no output content.
func TestPrintSummaryDoesNotPanic(t *testing.T) {
	result := &LoadResult{
		LocationsCreated:   []string{"loc1"},
		LocationsSkipped:   []string{"loc2"},
		RacksCreated:       []string{"rack1"},
		Created:            []string{"dev1", "dev2"},
		Updated:            []string{"dev3"},
		Skipped:            []string{"dev4"},
		IfacesCreated:      5,
		IfacesSkipped:      2,
		CablesCreated:      3,
		VLANsCreated:       1,
		PrefixesCreated:    2,
		IPAddressesCreated: 4,
	}
	// Should not panic.
	PrintSummary(result)
}

// ---------- NewNautobotClient ----------

// TestNewNautobotClientErrors verifies the client constructor rejects an empty
// URL and an empty token.
//
// Why it matters: a client built without a base URL or auth token cannot reach
// Nautobot; failing at construction gives a clear error instead of opaque 401s
// or malformed requests later.
// Inputs: ("","token") then ("http://example.com","").
// Outputs: an error in both cases.
// Data choice: the two calls isolate each required-field guard independently.
func TestNewNautobotClientErrors(t *testing.T) {
	_, err := NewNautobotClient("", "token")
	if err == nil {
		t.Error("expected error for empty URL")
	}

	_, err = NewNautobotClient("http://example.com", "")
	if err == nil {
		t.Error("expected error for empty token")
	}
}

// TestNewNautobotClientSuccess verifies a non-nil client is returned when both
// URL and token are supplied.
//
// Why it matters: this is the happy path every export run depends on; the
// constructor must succeed with valid inputs and not require a live server.
// Inputs: a valid URL and token.
// Outputs: a non-nil *NautobotClient and nil error.
// Data choice: example.com with a dummy token proves construction is offline
// and does not validate connectivity here.
func TestNewNautobotClientSuccess(t *testing.T) {
	client, err := NewNautobotClient("http://example.com", "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

// ---------- ValidateInventory (additional edge cases) ----------

// TestValidateInventoryNil verifies ValidateInventory rejects a nil inventory.
//
// Why it matters: export must refuse to run against a nil inventory with a clear
// error instead of panicking on the first map access.
// Inputs: nil. Outputs: a non-nil error.
// Data choice: nil is the most basic precondition failure and the first guard
// in the function.
func TestValidateInventoryNil(t *testing.T) {
	if err := ValidateInventory(nil); err == nil {
		t.Error("expected error for nil inventory")
	}
}

// TestValidateInventoryEmpty verifies an inventory with zero devices is rejected.
//
// Why it matters: there is nothing to export from an empty inventory, so it
// should fail validation rather than make a no-op run look successful.
// Inputs: an Inventory with an empty Devices map.
// Outputs: a non-nil error.
// Data choice: an explicitly empty (non-nil) map distinguishes this from the
// nil-inventory case above.
func TestValidateInventoryEmpty(t *testing.T) {
	if err := ValidateInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
	}); err == nil {
		t.Error("expected error for empty devices")
	}
}

// TestValidateInventoryAllNilDevices verifies an inventory whose only entries
// are nil device pointers is rejected.
//
// Why it matters: a map sized >0 but holding only nils has no exportable devices;
// validation must catch this so export does not report success with nothing done.
// Inputs: a Devices map with a single nil entry.
// Outputs: a non-nil error.
// Data choice: one nil entry makes len(Devices)>0 yet validCount==0, hitting the
// "no exportable devices" branch.
func TestValidateInventoryAllNilDevices(t *testing.T) {
	if err := ValidateInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			uuid.New(): nil,
		},
	}); err == nil {
		t.Error("expected error for all nil devices")
	}
}

// TestValidateInventoryOnlySystem verifies an inventory containing only
// "system"-typed devices is rejected.
//
// Why it matters: system objects are organizational, not real Nautobot devices;
// an inventory with only systems has nothing to create and must fail validation.
// Inputs: a single device of type "system".
// Outputs: a non-nil error.
// Data choice: the lone system device exercises the type=="system" skip that
// drives validCount to zero.
func TestValidateInventoryOnlySystem(t *testing.T) {
	if err := ValidateInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			uuid.New(): {Name: "sys", Type: "system"},
		},
	}); err == nil {
		t.Error("expected error for only system devices")
	}
}

// TestValidateInventoryOnlyUnnamed verifies an inventory of only unnamed devices
// is rejected.
//
// Why it matters: Nautobot devices need a name; if name generation has not run,
// validation should stop the export before it attempts nameless creates.
// Inputs: a single node device with an empty Name.
// Outputs: a non-nil error.
// Data choice: an empty-name node hits the Name=="" skip, the unnamed analog of
// the only-system case.
func TestValidateInventoryOnlyUnnamed(t *testing.T) {
	if err := ValidateInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			uuid.New(): {Name: "", Type: "node"},
		},
	}); err == nil {
		t.Error("expected error for only unnamed devices")
	}
}

// TestValidateInventoryValid verifies an inventory with one named node passes
// validation.
//
// Why it matters: this is the happy path gating every export; a false rejection
// here would block valid runs entirely.
// Inputs: a single node named "server-1".
// Outputs: nil error.
// Data choice: one minimally valid device (named, type node) is the smallest
// fixture that should pass.
func TestValidateInventoryValid(t *testing.T) {
	if err := ValidateInventory(&devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			uuid.New(): {Name: "server-1", Type: "node"},
		},
	}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- PrintSummary (all branches) ----------

// TestPrintSummaryAllBranches verifies that PrintSummary executes without
// panicking when given a LoadResult that exercises every counter and slice
// (created/updated/skipped/conflicts/errors across all object kinds).
//
// Why it matters: the summary is the operator-facing report of an export run;
// it must render all non-zero sections without crashing regardless of which
// counters are populated.
// Inputs: a LoadResult with every field non-empty. Outputs: none asserted
// beyond the absence of a panic (the rendered text is not captured).
// Data choice: populating all fields drives every print branch in one call.
// NOTE: this is a smoke test; it does not assert the rendered output.
func TestPrintSummaryAllBranches(t *testing.T) {
	result := &LoadResult{
		LocationsCreated:   []string{"loc1"},
		LocationsSkipped:   []string{"loc2"},
		RacksCreated:       []string{"rack1"},
		RacksSkipped:       2,
		Created:            []string{"dev1"},
		Updated:            []string{"dev2"},
		Skipped:            []string{"dev3"},
		Conflicts:          []ConflictInfo{{DeviceName: "dev3", Reason: "exists"}},
		IfacesCreated:      5,
		IfacesSkipped:      2,
		IfacesUnsupported:  1,
		ModulesCreated:     1,
		ModulesSkipped:     1,
		FrusCreated:        3,
		FrusSkipped:        1,
		CablesCreated:      4,
		CablesSkipped:      1,
		CablesConflicted:   1,
		VLANsCreated:       2,
		VLANsSkipped:       1,
		PrefixesCreated:    3,
		PrefixesSkipped:    1,
		IPAddressesCreated: 6,
		IPAddressesSkipped: 2,
		Errors:             []string{"something went wrong"},
	}
	// Should not panic.
	PrintSummary(result)
}

// ---------- NewExporter ----------

// TestNewExporterOptions verifies NewExporter builds a non-nil Exporter that
// carries the supplied ExporterOpts (DryRun and Merge) unchanged.
//
// Why it matters: export behavior (dry-run vs. write, merge-on-conflict) is
// governed by these options, so the constructor must wire them through
// faithfully.
// Inputs: a nil client, a cache, and opts with DryRun/Merge true. Outputs: an
// Exporter whose Options reflect those flags.
// Data choice: DryRun/Merge are asserted as representative toggles; the nil
// client is fine because no API call is made.
func TestNewExporterOptions(t *testing.T) {
	cache := NewLookupCache(nil)
	opts := &ExporterOpts{
		DryRun:          true,
		Merge:           true,
		DefaultLocation: "Site-1",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	}

	exp := NewExporter(nil, cache, opts)
	if exp == nil {
		t.Fatal("expected non-nil exporter")
	}
	if exp.Options.DryRun != true {
		t.Error("DryRun not set")
	}
	if exp.Options.Merge != true {
		t.Error("Merge not set")
	}
}

// ---------- ExporterOpts fields ----------

// TestExporterOptsAllFlags verifies an ExporterOpts literal stores its
// representative flags (DefaultLocation, CreateDeviceTypes, Strict).
//
// Why it matters: these flags gate whether the exporter creates missing
// Nautobot prerequisites and how strictly it treats conflicts.
// Inputs: an ExporterOpts with all create/strict flags set. Outputs: those
// fields read back with the assigned values.
// Data choice: asserting a string, a create-flag, and Strict samples the
// distinct flag categories without restating every field.
func TestExporterOptsAllFlags(t *testing.T) {
	opts := &ExporterOpts{
		DefaultLocation:     "DC1",
		DefaultRole:         "Leaf",
		DefaultStatus:       "Planned",
		Merge:               true,
		DryRun:              false,
		Strict:              true,
		CreateDeviceTypes:   true,
		CreateLocationTypes: true,
		CreateModuleTypes:   true,
		CreateLocations:     true,
		CreateStatuses:      true,
		CreateRoles:         true,
	}
	if opts.DefaultLocation != "DC1" {
		t.Errorf("DefaultLocation = %q", opts.DefaultLocation)
	}
	if !opts.CreateDeviceTypes {
		t.Error("CreateDeviceTypes should be true")
	}
	if !opts.Strict {
		t.Error("Strict should be true")
	}
}

// ---------- ErrDeviceUnclassified sentinel ----------

// TestErrDeviceUnclassifiedSentinel verifies resolveDeviceType returns the
// ErrDeviceUnclassified sentinel (matchable via errors.Is) for a device with no
// slug or model.
//
// Why it matters: callers branch on this sentinel to skip or warn about
// unclassifiable devices instead of pushing junk device types to Nautobot.
// Inputs: a non-strict mapper and an empty device. Outputs: an error that
// errors.Is matches against ErrDeviceUnclassified.
// Data choice: an empty device (no slug/model) is the only input that yields the
// sentinel.
func TestErrDeviceUnclassifiedSentinel(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{Strict: false})
	dev := &devicetypes.CaniDeviceType{} // no slug or model

	_, err := mapper.resolveDeviceType(dev)
	if !errors.Is(err, ErrDeviceUnclassified) {
		t.Errorf("expected ErrDeviceUnclassified, got %v", err)
	}
}
