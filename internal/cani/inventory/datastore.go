package inventory

import (
	"errors"

	"github.com/google/uuid"
)

var ErrHardwareNotFound = errors.New("hardware not found")
var ErrHardwareParentNotFound = errors.New("hardware parent not found")
var ErrHardwareUUIDConflict = errors.New("hardware uuid already exists")

type Datastore interface {
	GetSchemaVersion() (SchemaVersion, error)
	SetExternalInventoryProvider(provider ExternalInventoryProvider) error
	GetExternalInventoryProvider() (ExternalInventoryProvider, error)
	Flush() error

	// Crud operations
	Add(hardware *Hardware) error
	Get(uuid.UUID) (Hardware, error)
	Update(hardware *Hardware) error
	Remove(uuid uuid.UUID) error

	// Graph functions
	GetLocation(hardware Hardware) ([]LocationToken, error)
	GetAtLocation(path []LocationToken) (Hardware, error)
	GetChildren(id uuid.UUID) ([]Hardware, error)

	// TODO for search properties
}
