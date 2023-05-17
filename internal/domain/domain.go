package domain

import (
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/provider/csm"
	hardware_type_library "github.com/Cray-HPE/cani/pkg/hardware-type-library"
)

var (
	Data *Domain
)

type Domain struct {
	SessionActive       bool
	hardwareTypeLibrary *hardware_type_library.Library
	datastore           inventory.Datastore

	externalInventoryProvider provider.InventoryProvider
}

type NewOpts struct {
	DatastorePath string
	Provider      string
	EIPCSMOpts    csm.NewOpts
}

func New(opts *NewOpts) (*Domain, error) {
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
	domain.datastore, err = inventory.NewDatastoreJSON(opts.DatastorePath)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to load inventory datastore from file"),
			err,
		)
	}

	// Setup External inventory provider
	// TODO how does the initial inventory data for a session get created, if it uses domain logic
	// as it will fail when we get to this point.
	externalInventoryProviderName, err := domain.datastore.GetExternalInventoryProvider()
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to retrieve external inventory provider type"),
			err,
		)
	}
	switch externalInventoryProviderName {
	case inventory.ExternalInventoryProviderCSM:
		domain.externalInventoryProvider, err = csm.New(opts.EIPCSMOpts)
		if err != nil {
			return nil, errors.Join(
				fmt.Errorf("failed to initialize CSM external inventory provider"),
				err,
			)
		}
	default:
		return nil, fmt.Errorf("unknown external inventory provider provided (%s)", externalInventoryProviderName)
	}
	return domain, nil
}
