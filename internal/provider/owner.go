/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
package provider

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// ForInventory returns the providers allowed to act on inv.
//
// Generic CRUD (add, remove, update, show) offers optional interfaces such as
// MetadataApplier and RackPostAddHook to providers. Those hooks mutate the
// inventory, so only the provider that owns it may run: applying every
// registered provider's hooks would file a user's metadata under a provider
// that is not in use, or stamp another provider's naming scheme onto an object.
//
// An inventory that does not record an owner predates provider stamping, so
// every provider is returned to preserve existing behaviour. An inventory owned
// by a provider that is not registered gets none, since no registered provider
// can speak for it.
func ForInventory(inv *devicetypes.Inventory) []Provider {
	if inv == nil || inv.Provider == "" {
		return All()
	}
	if p := GetProvider(inv.Provider); p != nil {
		return []Provider{p}
	}
	return nil
}
