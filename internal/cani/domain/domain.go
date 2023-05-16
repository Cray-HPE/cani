package domain

import (
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/cani/inventory"
	hardware_type_library "github.com/Cray-HPE/cani/pkg/hardware-type-library"
)

type Domain struct {
	hardwareTypeLibrary *hardware_type_library.Library
	datastore           inventory.Datastore
}

func New() (*Domain, error) {
	var err error
	domain := &Domain{}

	// Load the hardware type library
	// TODO make this be able to be loaded from a directory
	domain.hardwareTypeLibrary, err = hardware_type_library.NewEmbeddedLibrary()
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to load embedded hardware type library"),
			err,
		)
	}

	// Load the datastore
	domain.datastore, err = inventory.NewDatastoreJSON("cani_db.json")
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to load inventory datastore from file"),
			err,
		)
	}

	return domain, nil
}
