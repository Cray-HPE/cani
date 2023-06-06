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

	// TODO for search properties

	// Clone creates a in-memory version of the datastore to perform location operations
	// TODO This can be kind of as a primitive to start a database transactions
	Clone() (Datastore, error)

	// Merge the contents of the remote datastore (most likely a in-memory one with changes)
	// TODO This can be kind of as a primitive to end a database transactions
	Merge(Datastore) error
}
