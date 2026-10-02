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

import "github.com/google/uuid"

// RecordModule records the Nautobot ID the module phase created or found for
// an inventory module, so the cable and IP phases resolve module-owned ports
// through an ID this run confirmed. A persisted external ID is not consulted
// for lookups: it may name a module Nautobot no longer has, and Nautobot
// answers a filter on such an ID with 400.
func (c *LookupCache) RecordModule(inventoryID, nautobotID uuid.UUID) {
	c.modulesMu.Lock()
	defer c.modulesMu.Unlock()
	c.exportedModules[inventoryID] = nautobotID
}

// ExportedModuleID returns the Nautobot ID this run confirmed for the
// inventory module, or uuid.Nil when the module was not created or found.
func (c *LookupCache) ExportedModuleID(inventoryID uuid.UUID) uuid.UUID {
	c.modulesMu.RLock()
	defer c.modulesMu.RUnlock()
	return c.exportedModules[inventoryID]
}

// PlanModule records that a dry run would create the inventory module, as
// PlanDevice does for a device, so a cable on one of its ports is previewed
// rather than looked up on a module that does not exist yet.
func (c *LookupCache) PlanModule(inventoryID uuid.UUID) {
	c.modulesMu.Lock()
	defer c.modulesMu.Unlock()
	c.plannedModules[inventoryID] = true
}

// ModulePlanned reports whether a dry run stands in for the inventory module.
func (c *LookupCache) ModulePlanned(inventoryID uuid.UUID) bool {
	c.modulesMu.RLock()
	defer c.modulesMu.RUnlock()
	return c.plannedModules[inventoryID]
}
