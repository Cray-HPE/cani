/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
package inventory

import (
	"errors"

	"github.com/google/uuid"
)

var ErrHardwareNotFound = errors.New("hardware not found")
var ErrHardwareParentNotFound = errors.New("hardware parent not found")
var ErrHardwareUUIDConflict = errors.New("hardware uuid already exists")
var ErrHardwareMissingLocationOrdinal = errors.New("hardware missing location ordinal")
var ErrEmptyLocationPath = errors.New("empty location path provided")

type Datastore interface {
	GetSchemaVersion() (SchemaVersion, error)
	SetInventoryProvider(provider Provider) error
	InventoryProvider() (Provider, error)
	Flush() error
	Validate() error

	// Crud operations
	Add(hardware *Hardware) error
	Get(uuid.UUID) (Hardware, error)
	Update(hardware *Hardware) error
	Remove(uuid uuid.UUID, recursion bool) error
	List() (Inventory, error)

	// Graph functions
	GetLocation(hardware Hardware) (LocationPath, error)
	GetAtLocation(path LocationPath) (Hardware, error)
	GetChildren(id uuid.UUID) ([]Hardware, error)
	GetSystemZero() (Hardware, error)              // TODO replace this when multiple systems are supported
	GetSystem(hardware Hardware) (Hardware, error) // Not yet implemented until multiple systems are supported

	// GetHardwareHierarchy(hardware Hardware) ([]Hardware, error)
	// TODO for search properties

	// Clone creates a in-memory version of the datastore to perform location operations
	Clone() (Datastore, error)

	// Merge the contents of the remote datastore (most likely a in-memory one with changes)
	Merge(Datastore) error
}
