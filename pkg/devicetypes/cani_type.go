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

// CaniType is the shared interface implemented by all inventory types.
// It provides a uniform API for validation, identification, and status
// across CaniDeviceType, CaniRackType, CaniLocationType, CaniModuleType,
// CaniCableType, and CaniFruType.
type CaniType interface {
	// Validate checks the instance for internal consistency.
	Validate() error

	// GetID returns the unique identifier.
	GetID() uuid.UUID

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
)
