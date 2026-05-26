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
package devicetypes

import "fmt"

// AnyResult is a discriminated union returned by LookupAny.
// Exactly one of the typed fields is non-nil.
type AnyResult struct {
	Category Category
	Device   *CaniDeviceType
	Rack     *CaniRackType
	Module   *CaniModuleType
	Cable    *CaniCableType
}

// LookupAny searches all registries (rack, device, module, cable) for a
// matching slug or part number. It returns the first exact match found,
// checking in the order: rack, device, module, cable. Returns an error
// if nothing matches across any registry.
func LookupAny(key string) (*AnyResult, error) {
	if key == "" {
		return nil, fmt.Errorf("slug or part number required")
	}

	// Rack registry
	if r, err := NewRackFromSlug(key); err == nil {
		return &AnyResult{Category: CategoryRack, Rack: r}, nil
	}
	if r, err := NewRackFromPartNumber(key); err == nil {
		return &AnyResult{Category: CategoryRack, Rack: r}, nil
	}

	// Device registry
	if d, err := NewDeviceFromSlug(key); err == nil {
		return &AnyResult{Category: CategoryDevice, Device: d}, nil
	}
	if d, err := NewDeviceFromPartNumber(key); err == nil {
		return &AnyResult{Category: CategoryDevice, Device: d}, nil
	}

	// Module registry
	if m, err := NewModuleFromSlug(key); err == nil {
		return &AnyResult{Category: CategoryModule, Module: m}, nil
	}
	if m, err := NewModuleFromPartNumber(key); err == nil {
		return &AnyResult{Category: CategoryModule, Module: m}, nil
	}

	// Cable registry
	if c, err := NewCableFromSlug(key); err == nil {
		return &AnyResult{Category: CategoryCable, Cable: c}, nil
	}
	if c, err := NewCableFromPartNumber(key); err == nil {
		return &AnyResult{Category: CategoryCable, Cable: c}, nil
	}

	return nil, fmt.Errorf("no hardware type found for slug or part number %q", key)
}
