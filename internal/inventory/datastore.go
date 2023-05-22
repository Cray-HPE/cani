package inventory

import (
	"errors"

	"github.com/google/uuid"
)

var ErrHardwareNotFound = errors.New("hardware not found")
var ErrHardwareParentNotFound = errors.New("hardware parent not found")
var ErrHardwareUUIDConflict = errors.New("hardware uuid already exists")
var ErrHardwareMissingLocationOrdinal = errors.New("hardware missing location ordinal")

type Datastore interface {
	GetSchemaVersion() (SchemaVersion, error)
	SetExternalInventoryProvider(provider ExternalInventoryProvider) error
	GetExternalInventoryProvider() (ExternalInventoryProvider, error)
	Flush() error

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

	// TODO for search properties
}
