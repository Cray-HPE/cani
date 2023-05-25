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

	// TODO for search properties
}
