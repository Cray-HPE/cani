package inventory

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type DatastoreJSON struct {
	inventoryLock sync.RWMutex
	inventory     *Inventory
	dataFilePath  string
}

func NewDatastoreJSON(dataFilePath string) (*DatastoreJSON, error) {
	datastore := &DatastoreJSON{
		dataFilePath: dataFilePath,
	}

	if _, err := os.Stat(dataFilePath); os.IsNotExist(err) {
		// Write a default config file if it doesn't exist
		log.Info().Msgf("%s does not exist, creating default database", dataFilePath)

		// Create the directory if it doesn't exist
		dbDir := filepath.Dir(dataFilePath)
		if _, err := os.Stat(dbDir); os.IsNotExist(err) {
			err = os.Mkdir(dbDir, 0755)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Error creating database directory: %s", err))
			}
		}

		// Create a config with default values since one does not exist
		datastore.inventory = &Inventory{
			SchemaVersion: SchemaVersionV1Alpha1,
			Hardware:      map[uuid.UUID]Hardware{},
		}

		if err := datastore.Flush(); err != nil {
			return nil, err
		}
	} else {
		// Load from datastore

		// Create the directory if it doesn't exist
		cfgDir := filepath.Dir(dataFilePath)
		os.MkdirAll(cfgDir, os.ModePerm)

		file, err := os.Open(dataFilePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		inventoryRaw, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(inventoryRaw, &datastore.inventory); err != nil {
			return nil, err
		}
	}

	return datastore, nil
}

func (dj *DatastoreJSON) GetSchemaVersion() (SchemaVersion, error) {
	return dj.inventory.SchemaVersion, nil
}

func (dj *DatastoreJSON) Flush() error {
	// convert the cfg struct to a JSON-formatted byte slice.

	data, err := json.Marshal(dj.inventory)
	if err != nil {
		return err
	}

	// write the byte slice to a file
	err = ioutil.WriteFile(dj.dataFilePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (dj *DatastoreJSON) Add(hardware *Hardware) error {
	dj.inventoryLock.Lock()
	defer dj.inventoryLock.Unlock()

	// Check to see if the hardware object has a UUID, if not create one
	if hardware.ID == uuid.Nil {
		hardware.ID = uuid.New()
	}

	// Check to see if this UUID is unique
	if _, exists := dj.inventory.Hardware[hardware.ID]; exists {
		return ErrHardwareUUIDConflict
	}

	// Check to see if parent UUID exists
	if hardware.Parent != uuid.Nil {
		if _, exists := dj.inventory.Hardware[hardware.Parent]; !exists {
			return ErrHardwareParentNotFound
		}
	}

	// Add it to the inventory map
	dj.inventory.Hardware[hardware.ID] = *hardware

	return nil
}

func (dj *DatastoreJSON) Get(id uuid.UUID) (Hardware, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	if hardware, exists := dj.inventory.Hardware[id]; exists {
		return hardware, nil
	}

	return Hardware{}, ErrHardwareNotFound
}

func (dj *DatastoreJSON) Update(hardware *Hardware) error {
	dj.inventoryLock.Lock()
	defer dj.inventoryLock.Unlock()

	// TODO this is verify similar to add, except for the set UUID at the start

	// Check to see if this UUID is unique
	if _, exists := dj.inventory.Hardware[hardware.ID]; exists {
		return ErrHardwareUUIDConflict
	}

	// Check to see if parent UUID exists
	if hardware.Parent != uuid.Nil {
		if _, exists := dj.inventory.Hardware[hardware.Parent]; !exists {
			return ErrHardwareParentNotFound
		}
	}

	// Add it to the inventory map
	dj.inventory.Hardware[hardware.ID] = *hardware

	return nil
}

func (dj *DatastoreJSON) Remove(id uuid.UUID) error {
	dj.inventoryLock.Lock()
	defer dj.inventoryLock.Unlock()

	// Check to see if this UUID exists
	if _, exists := dj.inventory.Hardware[id]; !exists {
		return ErrHardwareNotFound
	}

	// Check to see if any piece of hardware has this device as a parent
	// as you should not be able to remove a piece of hardware without either
	// delinking it or removing its children
	if children, err := dj.getChildren(id); err != nil {
		return err
	} else if len(children) != 0 {
		childrenIDs := []string{}
		for _, child := range children {
			childrenIDs = append(childrenIDs, child.ID.String())
		}
		return fmt.Errorf("unable to remove (%s) as it is the parent of [%s]", strings.Join(childrenIDs, ","))
	}

	// Remove the hardware!
	delete(dj.inventory.Hardware, id)

	return nil
}

// Graph functions
func (dj *DatastoreJSON) GetLocation(hardware Hardware) ([]LocationToken, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return nil, fmt.Errorf("todo")

}
func (dj *DatastoreJSON) GetAtLocation(path []LocationToken) (Hardware, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return Hardware{}, fmt.Errorf("todo")
}

func (dj *DatastoreJSON) GetChildren(id uuid.UUID) ([]Hardware, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return dj.getChildren(id)
}

func (dj *DatastoreJSON) getChildren(id uuid.UUID) ([]Hardware, error) {
	var results []Hardware

	// For right we need to iterate over the map, as we don't have any
	// book keeping to keep track of child hardware
	for _, hardware := range dj.inventory.Hardware {
		if hardware.Parent == id {
			results = append(results, hardware)
		}
	}

	return results, nil
}
