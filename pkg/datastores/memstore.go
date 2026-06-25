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

import "github.com/Cray-HPE/cani/pkg/devicetypes"

// MemStore is an in-memory DeviceStore that holds a single inventory for the
// duration of a batch session. Load returns the held inventory and Save retains
// it, so a run of many commands shares one inventory and incurs no per-command
// disk I/O. The batch runner persists the result once when the session ends.
type MemStore struct {
	inv *devicetypes.Inventory
}

// Load returns the in-memory inventory, rebuilding derived reverse indices and
// foreign keys first so each command observes the same consistent view the disk
// store would produce after unmarshalling.
func (m *MemStore) Load() (*devicetypes.Inventory, error) {
	if m.inv == nil {
		m.inv = devicetypes.NewInventory()
	}
	m.inv.RebuildProviderKeyIndex()
	m.inv.RebuildDerivedState()
	return m.inv, nil
}

// Save retains the inventory in memory; the batch runner flushes it to disk
// once the session ends.
func (m *MemStore) Save(inventory *devicetypes.Inventory) error {
	m.inv = inventory
	return nil
}

// Inventory returns the inventory currently held by the session for final
// persistence by the batch runner.
func (m *MemStore) Inventory() *devicetypes.Inventory {
	return m.inv
}
