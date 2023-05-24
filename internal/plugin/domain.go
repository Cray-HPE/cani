package plugin

import (
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/provider/csm"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
)

// Plugin defines domain logic that plugs into cani
type Plugin struct {
	// library has all of the types of hardware this plugin is compatible with
	library *hardwaretypes.Library
	// datastore is the inventory datastore
	datastore inventory.Datastore
	// provider is the external inventory provider
	provider provider.InventoryProvider
}

// NewOpts defines the options for creating a new domain
type NewOpts struct {
	DatastorePath string      `yaml:"datastore_path"`
	LogFilePath   string      `yaml:"log_file_path"`
	Provider      string      `yaml:"provider"`
	CsmOptions    csm.NewOpts `yaml:"csm_options"`
}

// New returns a new domain using the provided options
func New(opts *NewOpts) (*Plugin, error) {
	var err error
	plugin := &Plugin{}

	// Load the hardware type library
	// TODO make this be able to be loaded from a directory
	plugin.library, err = hardwaretypes.NewEmbeddedLibrary()
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to load embedded hardware type library"),
			err,
		)
	}

	// Load the datastore
	plugin.datastore, err = inventory.NewDatastoreJSON(opts.DatastorePath, opts.LogFilePath, inventory.ProviderCSM)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to load inventory datastore from file"),
			err,
		)
	}

	// Setup External inventory provider
	// TODO how does the initial inventory data for a session get created, if it uses domain logic
	// as it will fail when we get to this point.
	providerName, err := plugin.datastore.GetProvider()
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to retrieve external inventory provider type"),
			err,
		)
	}
	switch providerName {
	case inventory.ProviderCSM:
		plugin.provider, err = csm.New(opts.CsmOptions)
		if err != nil {
			return nil, errors.Join(
				fmt.Errorf("failed to initialize CSM external inventory provider"),
				err,
			)
		}
	default:
		return nil, fmt.Errorf("unknown external inventory provider provided (%s)", providerName)
	}
	return plugin, nil
}
