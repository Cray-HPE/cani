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

import "github.com/google/uuid"

// CaniObject is the contract every entity stored in an Inventory satisfies.
// It is the minimum needed to identify an entity and check it for internal
// consistency, so a caller can walk the whole inventory without knowing which
// collection an entity came from.
type CaniObject interface {
	// Validate checks the instance for internal consistency.
	Validate() error

	// GetID returns the unique identifier.
	GetID() uuid.UUID
}

// CaniType extends CaniObject for the hardware entities that resolve against
// the device-type library and carry a lifecycle status: CaniDeviceType,
// CaniRackType, CaniLocationType, CaniModuleType, CaniCableType and
// CaniFruType. IPAM entities and interfaces are CaniObjects but not CaniTypes,
// since neither a slug nor a hardware status is meaningful for them.
type CaniType interface {
	CaniObject

	// GetSlug returns the hardware library slug (or type identifier for locations).
	GetSlug() string

	// GetStatus returns the current status string.
	GetStatus() string
}

// Compile-time interface satisfaction checks.
var (
	_ CaniType = (*CaniDeviceType)(nil)
	_ CaniType = (*CaniRackType)(nil)
	_ CaniType = (*CaniLocationType)(nil)
	_ CaniType = (*CaniModuleType)(nil)
	_ CaniType = (*CaniCableType)(nil)
	_ CaniType = (*CaniFruType)(nil)

	_ CaniObject = (*CaniInterface)(nil)
	_ CaniObject = (*CaniPrefix)(nil)
	_ CaniObject = (*CaniIPAddress)(nil)
	_ CaniObject = (*CaniVLAN)(nil)
	_ CaniObject = (*CaniVRF)(nil)
)
