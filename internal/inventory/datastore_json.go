package inventory

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type DatastoreJSON struct {
	inventoryLock sync.RWMutex
	inventory     *Inventory
	dataFilePath  string
	logFilePath   string
}

func NewDatastoreJSON(dataFilePath string, logfilepath string, provider Provider) (*DatastoreJSON, error) {
	datastore := &DatastoreJSON{
		dataFilePath: dataFilePath,
		logFilePath:  logfilepath,
	}

	if _, err := os.Stat(dataFilePath); os.IsNotExist(err) {
		// Write a default config file if it doesn't exist
		log.Info().Msgf("%s does not exist, creating default datastore", dataFilePath)

		// Create the directory if it doesn't exist
		dbDir := filepath.Dir(dataFilePath)
		if _, err := os.Stat(dbDir); os.IsNotExist(err) {
			err = os.Mkdir(dbDir, 0755)
			if err != nil {
				return nil, errors.Join(fmt.Errorf("error creating datastore directory"), err)
			}
		}

		// Create a config with default values since one does not exist
		datastore.inventory = &Inventory{
			SchemaVersion: SchemaVersionV1Alpha1,
			Provider:      provider,
			Hardware:      map[uuid.UUID]Hardware{},
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

	// When loading the JSON inventory ideally we should recompute the derived fields, as
	// we don't know the if the inventory structure was left in an inconsistent state, and might
	// cause issues when operating with inconsistent cached data.
	// 1. An external actor could have manually modified the inventory structure
	// 2. CANI could have crashed or left the datastore in an inconsistent state.
	if err := datastore.calculateDerivedFields(); err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to calculate inventory derived fields"),
			err,
		)
	}

	if err := datastore.Flush(); err != nil {
		return nil, err
	}

	return datastore, nil
}

func (dj *DatastoreJSON) GetSchemaVersion() (SchemaVersion, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return dj.inventory.SchemaVersion, nil
}

// SetExternalInventoryProvider sets the external inventory provider
func (dj *DatastoreJSON) SetInventoryProvider(provider Provider) error {
	dj.inventoryLock.Lock()
	defer dj.inventoryLock.Unlock()

	dj.inventory.Provider = provider

	return nil
}

// GetExternalInventoryProvider gets the external inventory provider
func (dj *DatastoreJSON) InventoryProvider() (Provider, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return dj.inventory.Provider, nil
}

// Flush writes the current inventory to the datastore
func (dj *DatastoreJSON) Flush() error {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	// convert the Inventory struct to a JSON-formatted byte slice.
	data, err := json.MarshalIndent(dj.inventory, "", "  ")
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

func (dj *DatastoreJSON) calculateDerivedFields() (err error) {
	//
	// Update location path
	//
	for _, hardware := range dj.inventory.Hardware {

		// GetLocation relies on the Parent UUID field, so it should be safe to compute based off it
		hardware.LocationPath, err = dj.getLocation(hardware)
		if err != nil {
			return fmt.Errorf("failed to calculate location path of (%s)", hardware.ID)
		}

		// Push the updated hardware object back into the map
		dj.inventory.Hardware[hardware.ID] = hardware
	}

	//
	// Update children book keeping
	//

	// Clear out old children data
	for _, hardware := range dj.inventory.Hardware {
		hardware.Children = nil

		// Push the updated hardware object back into the map
		dj.inventory.Hardware[hardware.ID] = hardware
	}

	// Build up the children data
	for _, hardware := range dj.inventory.Hardware {

		if hardware.Parent != uuid.Nil {
			parent, ok := dj.inventory.Hardware[hardware.Parent]
			if !ok {
				// This should not happen
				return fmt.Errorf("unable to find parent hardware object with ID (%s) of (%s)", hardware.Parent, hardware.ID)
			}

			// Add this hardware object as a child of the parent
			parent.Children = append(parent.Children, hardware.ID)

			// Push the updated parent back into the map
			dj.inventory.Hardware[parent.ID] = parent

		}
	}

	// Sort the children data so we have deterministic results
	for _, hardware := range dj.inventory.Hardware {
		sort.Slice(hardware.Children, func(i, j int) bool {
			return hardware.Children[i].ID() < hardware.Children[j].ID()
		})

		// Push the updated hardware back into the map
		dj.inventory.Hardware[hardware.ID] = hardware
	}

	return nil
}

// Validate validates the current inventory
func (dj *DatastoreJSON) Validate() error {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	log.Warn().Msg("DatastoreJSON's Validate was called. This is not currently implemented")

	// Verify inventory map key matches hardware UUID
	// TODO

	// Verify all parent IDs are valid
	// TOOD

	// TODO think of other checks

	// TODO for right now say everything is ok
	return nil
}

// Add adds a new hardware object to the inventory
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
	if dj.inventory.Hardware == nil {
		log.Warn().Msg("Initializing inventory map")
		dj.inventory.Hardware = map[uuid.UUID]Hardware{}
	}
	dj.inventory.Hardware[hardware.ID] = *hardware

	dj.logTransaction("ADD", hardware.ID.String(), nil, nil)

	// Update derived fields
	if err := dj.calculateDerivedFields(); err != nil {
		return errors.Join(
			fmt.Errorf("failed to calculate inventory derived fields"),
			err,
		)
	}

	return nil
}

// Get returns a hardware object from the inventory
func (dj *DatastoreJSON) Get(id uuid.UUID) (Hardware, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	if hardware, exists := dj.inventory.Hardware[id]; exists {
		return hardware, nil
	}

	dj.logTransaction("GET", id.String(), nil, nil)

	return Hardware{}, ErrHardwareNotFound
}

// Update updates a hardware object in the inventory
func (dj *DatastoreJSON) Update(hardware *Hardware) error {
	dj.inventoryLock.Lock()
	defer dj.inventoryLock.Unlock()

	// Check to see if this UUID exists
	oldHardware, exists := dj.inventory.Hardware[hardware.ID]
	if !exists {
		return ErrHardwareNotFound
	}

	// Check to see if parent UUID exists
	if hardware.Parent != uuid.Nil {
		if _, exists := dj.inventory.Hardware[hardware.Parent]; !exists {
			return ErrHardwareParentNotFound
		}
	}

	// Add it to the inventory map
	dj.inventory.Hardware[hardware.ID] = *hardware

	dj.logTransaction("UPDATE", hardware.ID.String(), nil, nil)

	// Update derived fields if the parent ID is different than the old value
	if oldHardware.Parent != hardware.ID {
		log.Debug().Msgf("Detected parent ID change for (%s) from (%s) to (%s)", hardware.ID, oldHardware.Parent, hardware.ID)
		if err := dj.calculateDerivedFields(); err != nil {
			return errors.Join(
				fmt.Errorf("failed to calculate inventory derived fields"),
				err,
			)
		}
	}

	return nil
}

// Remove removes a hardware object from the inventory
func (dj *DatastoreJSON) Remove(id uuid.UUID, recursion bool) error {
	dj.inventoryLock.Lock()
	defer dj.inventoryLock.Unlock()

	// Check to see if this UUID exists
	if _, exists := dj.inventory.Hardware[id]; !exists {
		return ErrHardwareNotFound
	}

	// Check to see if any piece of hardware has this device as a parent
	// as you should not be able to remove a piece of hardware without either
	// delinking it or removing its children
	// FIXME: https://github.com/Cray-HPE/cani/pull/28#discussion_r1199347499
	if children, err := dj.getChildren(id); err != nil {
		return err
	} else if len(children) != 0 {
		childrenIDs := []string{}
		for _, child := range children {
			childrenIDs = append(childrenIDs, child.ID.String())
		}
		// If recursion is true, remove the children as well
		if recursion {
			for _, child := range children {
				delete(dj.inventory.Hardware, child.ID)
			}
		} else {
			return fmt.Errorf("unable to remove (%s) as it is the parent of [%s]", id.String(), strings.Join(childrenIDs, ","))
		}
	}

	// Remove the hardware!
	delete(dj.inventory.Hardware, id)

	dj.logTransaction("REMOVE", id.String(), nil, nil)

	// Update all derived fields, as we allow for recursive removal and trying to keep this logic simple for now
	// If this is too much of a performance penalty we can make this more smart
	if err := dj.calculateDerivedFields(); err != nil {
		return errors.Join(
			fmt.Errorf("failed to calculate inventory derived fields"),
			err,
		)
	}

	return nil
}

// List returns the entire inventory
func (dj *DatastoreJSON) List() (Inventory, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return *dj.inventory, nil
}

// GetLocation will follow the parent links up to the root node, which is signaled when a NIL parent UUID is found
// This will either return a partial location path, or a full path up to a cabinet or CDU
// TODO THIS NEEDS UNIT TESTS
func (dj *DatastoreJSON) GetLocation(hardware Hardware) (LocationPath, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return dj.getLocation(hardware)
}

// getLocation is just like GetLocation except it doesn't attempt to acquire the inventory RWMutex, so it can be
// composed into other DatastoreJSON functions that make changes to the inventory structure.
func (dj *DatastoreJSON) getLocation(hardware Hardware) (LocationPath, error) {

	locationPath := LocationPath{}

	// Follow the parent links up to the root node
	currentHardwareID := hardware.ID
	for currentHardwareID != uuid.Nil {
		currentHardware, exists := dj.inventory.Hardware[currentHardwareID]
		if !exists {
			return nil, errors.Join(
				fmt.Errorf("unable to find ancestor (%s) of (%s)", currentHardwareID, hardware.ID),
				ErrHardwareNotFound,
			)
		}

		// The inventory structure allows for hardware to have no location, and this is a valid state.
		// Such as information has been obtained from a node, but it is missing geolocation
		if currentHardware.LocationOrdinal == nil {
			return nil, errors.Join(
				fmt.Errorf("missing location ordinal in ancestor (%s) of (%s)", currentHardwareID, hardware.ID),
				ErrHardwareMissingLocationOrdinal,
			)
		}

		// Build up an element in the location path.
		// Since the tree is being traversed bottom up, need to add each location token to the front of the slice
		locationPath = append([]LocationToken{{
			HardwareType: currentHardware.Type,
			Ordinal:      *currentHardware.LocationOrdinal,
		}}, locationPath...)

		// Go the parent node next
		currentHardwareID = currentHardware.Parent
	}

	return locationPath, nil

}

// GetAtLocation returns the hardware at the given location
// TODO THIS NEEDS UNIT TESTS
func (dj *DatastoreJSON) GetAtLocation(path LocationPath) (Hardware, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	if len(path) == 0 {
		return Hardware{}, ErrEmptyLocationPath
	}

	// TODO Currently do not have a system type in our inventory, which cabinets/CDUs would be the child of, so for now keep track of them
	topLevelHardware := map[hardwaretypes.HardwareType][]uuid.UUID{
		hardwaretypes.HardwareTypeCabinet:                 []uuid.UUID{},
		hardwaretypes.HardwareTypeCoolingDistributionUnit: []uuid.UUID{},
	}
	for _, hardware := range dj.inventory.Hardware {
		if hardware.Type == hardwaretypes.HardwareTypeCabinet || hardware.Type == hardwaretypes.HardwareTypeCoolingDistributionUnit {
			topLevelHardware[hardware.Type] = append(topLevelHardware[hardware.Type], hardware.ID)
		}
	}

	//
	// Traverse the tree to see if the hardware exists at the given location
	//

	var currentHardware Hardware
	for i, locationToken := range path {
		log.Debug().Msgf("GetAtLocation: Processing token %d of %d: '%s'", i+1, len(path), locationToken.String())
		log.Debug().Msgf("GetAtLocation: Current ID %s", currentHardware.ID)

		var children []uuid.UUID
		if i == 0 {
			// First find the children of "root"/top level object, but we currently lack that so use the topLevelHardware lookup
			children = topLevelHardware[path[0].HardwareType]
		} else {
			// Otherwise get the children like normal
			children = currentHardware.Children
		}

		// For each child of the current hardware object check to see if it
		foundMatch := false
		for _, childID := range children {
			log.Debug().Msgf("GetAtLocation: Visiting Child (%s)", childID)
			// Get the hardware
			childHardware, ok := dj.inventory.Hardware[childID]
			if !ok {
				// This should not happen
				return Hardware{}, errors.Join(
					fmt.Errorf("unable to find hardware object with ID (%s)", childID),
					ErrHardwareNotFound,
				)
			}

			if childHardware.LocationOrdinal == nil {
				log.Debug().Msgf("GetAtLocation: Child has no location ordinal set, skipping")
				continue
			}
			log.Debug().Msgf("GetAtLocation: Child location token: %s:%d", childHardware.Type, *childHardware.LocationOrdinal)

			// Check to see if the location token matches
			if childHardware.Type == locationToken.HardwareType && *childHardware.LocationOrdinal == locationToken.Ordinal {
				// Found a match!
				log.Debug().Msgf("GetAtLocation: Child has matching location token")
				currentHardware = childHardware
				foundMatch = true
				break
			}
		}

		if !foundMatch {
			// None of the children match
			return Hardware{}, ErrHardwareNotFound
		}

	}

	if currentHardware.ID == uuid.Nil {
		return Hardware{}, ErrHardwareNotFound
	}

	return currentHardware, nil
}

// GetChildren returns the children of a given hardware object
func (dj *DatastoreJSON) GetChildren(id uuid.UUID) ([]Hardware, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return dj.getChildren(id)
}

// getChildren returns the children of a given hardware object
// This function depends on the cached/derived children data
func (dj *DatastoreJSON) getChildren(id uuid.UUID) ([]Hardware, error) {
	hardware, exists := dj.inventory.Hardware[id]
	if !exists {
		return nil, ErrHardwareNotFound
	}

	// For right we need to iterate over the map, as we don't have any
	// book keeping to keep track of child hardware
	var results []Hardware
	for _, childID := range hardware.Children {
		childHardware, exists := dj.inventory.Hardware[childID]
		if !exists {
			// This should not happen
			return nil, fmt.Errorf("unable to find child hardware object with ID (%s) of (%s)", childHardware.ID, hardware.ID)
		}

		results = append(results, childHardware)
	}

	return results, nil
}

// logTransaction logs a transaction to logger
func (dj *DatastoreJSON) logTransaction(operation string, key string, value interface{}, err error) {
	tl, err = os.OpenFile(
		dj.logFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open transaction log file")
		return
	}
	logger = zerolog.New(tl).With().Timestamp().Logger()
	defer tl.Close()

	// Get the current timestamp
	timestamp := time.Now()

	// Determine the operation status
	status := "SUCCESS"
	if err != nil {
		status = "FAILED"
	}

	// Log the transaction
	logEvent := logger.With().
		Timestamp().
		Str("timestamp", timestamp.Format(time.RFC3339)).
		Str("operation", operation).
		Str("key", key).
		Interface("value", value).
		Str("status", status).
		Logger()

	if err != nil {
		logEvent.Err(err).Msg("Transaction")
	} else {
		logEvent.Info().Msg("Transaction")
	}

}
