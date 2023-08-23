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

	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
)

var ErrHardwareNotFound = errors.New("hardware not found")
var ErrHardwareParentNotFound = errors.New("hardware parent not found")
var ErrHardwareUUIDConflict = errors.New("hardware uuid already exists")
var ErrHardwareMissingLocationOrdinal = errors.New("hardware missing location ordinal")
var ErrEmptyLocationPath = errors.New("empty location path provided")
var ErrDatastoreValidationFailure = errors.New("datastore validation failure")

// SearchFilter works as follows
// - Each field is a different category to filter on
// - When is there a match?
//   - A hardware object must match one element in a category
//   - If a category is empty then all hardware objects will match
//   - A must match across all categories
type SearchFilter struct {
	Types  []hardwaretypes.HardwareType
	Status []HardwareStatus
}

type Datastore interface {
	GetSchemaVersion() (SchemaVersion, error)
	SetInventoryProvider(provider Provider) error
	InventoryProvider() (Provider, error)
	Flush() error
	Validate() (map[uuid.UUID]ValidateResult, error)

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
	GetDescendants(id uuid.UUID) ([]Hardware, error)
	GetSystemZero() (Hardware, error)              // TODO replace this when multiple systems are supported
	GetSystem(hardware Hardware) (Hardware, error) // Not yet implemented until multiple systems are supported

	// TODO for search properties
	Search(filter SearchFilter) (map[uuid.UUID]Hardware, error)

	// Clone creates a in-memory version of the datastore to perform location operations
	Clone() (Datastore, error)

	// Merge the contents of the remote datastore (most likely a in-memory one with changes)
	Merge(Datastore) error
}

type ValidateResult struct {
	Hardware Hardware
	Errors   []string
}
