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
package add

import (
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/google/uuid"
)

// TestAddModuleSpareStoresAtLocation verifies --spare --location adds a module
// with a location and no parent device, and that --spare without --location
// is rejected.
//
// Why it matters: spares were previously impossible to author, so operators
// invented a parent; the location is what export needs for a spare.
// Inputs: an inventory with one location; add module hpe-2p-25gbe-pcie-nic
// --spare --location <name>, then --spare alone (run first, because flag
// values persist on the shared command). Outputs: one module whose Location is
// the site and whose ParentDevice is empty; one error; one save.
// Data choice: hpe-2p-25gbe-pcie-nic is a library module type so the add
// passes type validation without a parent device.
func TestAddModuleSpareStoresAtLocation(t *testing.T) {
	inventory, _ := cmdtest.InventoryWithRack("rack-01")
	var siteID uuid.UUID
	var siteName string
	for id, loc := range inventory.Locations {
		siteID, siteName = id, loc.Name
	}
	harness := cmdtest.New(t, NewCommand(), newModuleCommand(), inventory)

	err := harness.Run(t, map[string][]string{"spare": {"true"}}, "hpe-2p-25gbe-pcie-nic")
	if err == nil || !strings.Contains(err.Error(), "--spare requires --location") {
		t.Errorf("spare without location error = %v, want rejection", err)
	}

	if err := harness.Run(t, map[string][]string{"spare": {"true"}, "location": {siteName}}, "hpe-2p-25gbe-pcie-nic"); err != nil {
		t.Fatalf("add module --spare --location: unexpected error: %v", err)
	}
	if len(harness.Inventory().Modules) != 1 {
		t.Fatalf("modules = %d, want 1", len(harness.Inventory().Modules))
	}
	for _, mod := range harness.Inventory().Modules {
		if mod.ParentDevice != uuid.Nil || mod.Location != siteID || !mod.IsSpare() {
			t.Errorf("module = {parent %v, location %v}, want a spare at %v", mod.ParentDevice, mod.Location, siteID)
		}
	}
	if harness.Store.Saves != 1 {
		t.Errorf("Saves = %d, want 1", harness.Store.Saves)
	}
}
