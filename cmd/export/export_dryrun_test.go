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
	"bytes"
	"os"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// configDryRunProvider is a fake exporter whose dry run comes from its own
// configuration, which it reports through the ExportDryRunner capability.
type configDryRunProvider struct {
	fakeExportProvider
	dryRun bool
}

func (provider *configDryRunProvider) ExportDryRun() bool {
	return provider.dryRun
}

// TestRunExportHonoursTheProvidersOwnDryRun verifies the save is skipped when
// the provider, not the flag, reports a dry run, and still happens otherwise.
//
// Why it matters: a provider can be told to dry-run through its own config
// (Nautobot's export.dry_run), which the generic --dry-run flag never sees;
// saving after such a run would persist whatever the planning left behind.
// Inputs: --dry-run unset; a fake exporter implementing ExportDryRunner that
// stamps an external ID, once reporting a dry run and once not.
// Outputs: the file is unchanged when the provider reports a dry run and
// rewritten when it does not.
// Data choice: the same stamp in both runs makes the save the only variable.
func TestRunExportHonoursTheProvidersOwnDryRun(t *testing.T) {
	for name, dryRun := range map[string]bool{"provider dry run": true, "real run": false} {
		t.Run(name, func(t *testing.T) { assertSaveSkippedWhen(t, dryRun) })
	}
}

// assertSaveSkippedWhen runs an export whose provider reports the given dry
// run and checks the datastore was left alone exactly when it did.
func assertSaveSkippedWhen(t *testing.T, providerDryRun bool) {
	t.Helper()
	inventory, deviceID := connectedExportInventory()
	command, inventoryPath := prepareExportTest(t, inventory)
	command.Flags().Bool("dry-run", false, "")
	before, err := os.ReadFile(inventoryPath)
	if err != nil {
		t.Fatalf("reading inventory before export: %v", err)
	}
	provider := &configDryRunProvider{dryRun: providerDryRun}
	provider.mutate = func(inventory *devicetypes.Inventory) {
		inventory.Devices[deviceID].ExternalIDs = map[string]uuid.UUID{"fake-export": uuid.New()}
	}

	if err := runExport(command, nil, provider); err != nil {
		t.Fatalf("runExport error = %v", err)
	}

	after, err := os.ReadFile(inventoryPath)
	if err != nil {
		t.Fatalf("reading inventory after export: %v", err)
	}
	if unchanged := bytes.Equal(before, after); unchanged != providerDryRun {
		t.Errorf("inventory file unchanged = %v, want %v when the provider reports dry run = %v", unchanged, providerDryRun, providerDryRun)
	}
}
