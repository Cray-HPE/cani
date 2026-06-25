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
package datastores

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// TestMemStoreLoadSaveRoundTrip verifies MemStore returns the inventory it holds
// on Load and retains a newly saved one.
//
// Why it matters: a batch run shares one in-memory inventory across many
// commands; Load must hand back the live inventory (no disk read) and Save must
// keep it so later commands and the final flush observe accumulated changes.
// Inputs: an empty inventory, then a replacement. Outputs: the same pointers via
// Load and Inventory.
// Data choice: empty inventories isolate the store's pass-through behavior from
// any relationship rebuild side effects.
func TestMemStoreLoadSaveRoundTrip(t *testing.T) {
	inv := devicetypes.NewInventory()
	ms := &MemStore{inv: inv}

	got, err := ms.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != inv {
		t.Errorf("Load returned a different inventory than the one held")
	}

	replacement := devicetypes.NewInventory()
	if err := ms.Save(replacement); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if ms.Inventory() != replacement {
		t.Errorf("Inventory() did not return the saved inventory")
	}
}

// TestMemStoreLoadInitializesNilInventory verifies Load creates an inventory
// when the store holds none.
//
// Why it matters: starting a session from an empty datastore must still yield a
// usable inventory for the first command rather than a nil dereference.
// Inputs: a MemStore with a nil inventory. Outputs: a non-nil inventory.
// Data choice: the nil case is the boundary the guard exists to handle.
func TestMemStoreLoadInitializesNilInventory(t *testing.T) {
	ms := &MemStore{}
	got, err := ms.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got == nil {
		t.Error("Load returned nil inventory")
	}
}

// TestSessionStoreOverridesSetDeviceStore verifies that while a session is
// active SetDeviceStore keeps the in-memory store, and that ending the session
// clears it.
//
// Why it matters: each command re-dispatched inside a batch calls store.Setup ->
// SetDeviceStore; without the session override it would reopen the file and undo
// the load-once/save-once design.
// Inputs: a begun session and a SetDeviceStore("json") call during it. Outputs:
// the package Datastore identity and the cleared session after EndSession.
// Data choice: the "json" type is the only implemented backend, so it exercises
// the real override branch without constructing a disk store.
func TestSessionStoreOverridesSetDeviceStore(t *testing.T) {
	prev := Datastore
	t.Cleanup(func() {
		EndSession()
		Datastore = prev
	})

	ms := BeginSession(devicetypes.NewInventory())
	if Datastore != DeviceStore(ms) {
		t.Fatalf("BeginSession did not set Datastore to the session store")
	}

	if err := SetDeviceStore("json"); err != nil {
		t.Fatalf("SetDeviceStore during session: %v", err)
	}
	if Datastore != DeviceStore(ms) {
		t.Errorf("SetDeviceStore replaced the active session store")
	}

	EndSession()
	if sessionStore != nil {
		t.Errorf("EndSession did not clear the active session store")
	}
}
