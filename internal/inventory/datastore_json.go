/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

func NewDatastoreJSON(dataFilePath string, logfilepath string, provider Provider) (Datastore, error) {
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

		// Generate a UUID for a new top-level "System" object
		system := uuid.New()
		// Create a config with default values since one does not exist
		datastore.inventory = &Inventory{
			SchemaVersion: SchemaVersionV1Alpha1,
			Provider:      provider,
			Hardware: map[uuid.UUID]Hardware{
				// NOTE: At present, we only allow ONE system in the inventory, but leaving the door open for multiple systems
				system: {
					Type: hardwaretypes.System, // The top-level object is a hardwaretypes.System
					ID:   system,               // ID is the same as the key for the top-level system object to prevent a uuid.Nil
				},
			},
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

		inventoryRaw, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(inventoryRaw, &datastore.inventory); err != nil {
			return nil, err
		}
	}

	if err := datastore.Flush(); err != nil {
		return nil, err
	}

	return datastore, nil
}

func NewDatastoreJSONInMemory(provider Provider) (*DatastoreJSON, error) {
	datastore := &DatastoreJSON{}

	// Generate a UUID for a new top-level "System" object
	system := uuid.New()
	// Create a config with default values since one does not exist
	datastore.inventory = &Inventory{
		SchemaVersion: SchemaVersionV1Alpha1,
		Provider:      provider,
		Hardware: map[uuid.UUID]Hardware{
			// NOTE: At present, we only allow ONE system in the inventory, but leaving the door open for multiple systems
			system: {
				Type: hardwaretypes.System, // The top-level object is a hardwaretypes.System
				ID:   system,               // ID is the same as the key for the top-level system object to prevent a uuid.Nil
			},
		},
	}

	return datastore, nil
}

func (ds *DatastoreJSON) GetSchemaVersion() (SchemaVersion, error) {
	ds.inventoryLock.RLock()
	defer ds.inventoryLock.RUnlock()

	return ds.inventory.SchemaVersion, nil
}

// SetExternalInventoryProvider sets the external inventory provider
func (ds *DatastoreJSON) SetInventoryProvider(provider Provider) error {
	ds.inventoryLock.Lock()
	defer ds.inventoryLock.Unlock()

	ds.inventory.Provider = provider

	return nil
}

// GetExternalInventoryProvider gets the external inventory provider
func (ds *DatastoreJSON) InventoryProvider() (Provider, error) {
	ds.inventoryLock.RLock()
	defer ds.inventoryLock.RUnlock()

	return ds.inventory.Provider, nil
}

// Flush writes the current inventory to the datastore
func (ds *DatastoreJSON) Flush() error {
	if ds.dataFilePath == "" {
		// If running in in-memory mode there is nothing to flush
		return nil
	}

	ds.inventoryLock.RLock()
	defer ds.inventoryLock.RUnlock()

	// convert the Inventory struct to a yaml-formatted byte slice.
	data, err := json.MarshalIndent(ds.inventory, "", "  ")
	if err != nil {
		return err
	}

	// write the byte slice to a file
	err = os.WriteFile(ds.dataFilePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (ds *DatastoreJSON) Clone() (Datastore, error) {
	ds.inventoryLock.RLock()
	defer ds.inventoryLock.RUnlock()

	result, err := NewDatastoreJSONInMemory(ds.inventory.Provider)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to create in memory datastore"), err)
	}

	// Deep copy the hardware information into the datastore
	// TODO this is a hack
	raw, err := json.Marshal(ds.inventory.Hardware)
	if err != nil {
		return nil, err
	}
	result.inventory.Hardware = nil
	if err := json.Unmarshal(raw, &result.inventory.Hardware); err != nil {
		return nil, err
	}

	return result, nil
}

func (ds *DatastoreJSON) Merge(other Datastore) error {
	ds.inventoryLock.Lock()
	defer ds.inventoryLock.Unlock()

	otherAllHardware, err := other.List()
	if err != nil {
		return errors.Join(fmt.Errorf("failed to retrieve inventory from other datastore"), err)
	}

	// Identify hardware to remove
	hardwareToDelete := []uuid.UUID{}
	for id := range ds.inventory.Hardware {
		if _, exists := otherAllHardware.Hardware[id]; !exists {
			hardwareToDelete = append(hardwareToDelete, id)
		}
	}
	// Remove deleted hardware
	for _, id := range hardwareToDelete {
		delete(ds.inventory.Hardware, id)
	}

	// Update or add hardware
	for id, otherHardware := range otherAllHardware.Hardware {
		ds.inventory.Hardware[id] = otherHardware
	}

	return nil
}

// Validate validates the current inventory
func (ds *DatastoreJSON) Validate() (map[uuid.UUID]ValidateResult, error) {
	ds.inventoryLock.RLock()
	defer ds.inventoryLock.RUnlock()

	log.Warn().Msg("DatastoreJSON's Validate was called. This is not fully implemented")
	validationResults := map[uuid.UUID]ValidateResult{}

	// Verify inventory map key matches hardware UUID
	// TODO

	// Verify all parent IDs are valid
	// TOOD

	// TODO think of other checks

	// TODO for right now say everything is ok

	//
	// Verify all complete location paths are unique.
	//

	if len(validationResults) > 0 {
		return validationResults, ErrDatastoreValidationFailure
	}

	return nil, nil
}

// Add adds a new hardware object to the inventory
func (ds *DatastoreJSON) Add(hardware *Hardware) error {
	ds.inventoryLock.Lock()
	defer ds.inventoryLock.Unlock()

	// Check to see if the hardware object has a UUID, if not create one
	if hardware.ID == uuid.Nil {
		hardware.ID = uuid.New()
	}

	// Check to see if this UUID is unique
	if _, exists := ds.inventory.Hardware[hardware.ID]; exists {
		return ErrHardwareUUIDConflict
	}

	// Check to see if parent UUID exists
	if hardware.Parent != uuid.Nil {
		if _, exists := ds.inventory.Hardware[hardware.Parent]; !exists {
			return ErrHardwareParentNotFound
		}
	}

	// Add it to the inventory map
	if ds.inventory.Hardware == nil {
		log.Warn().Msg("Initializing inventory map")
		ds.inventory.Hardware = map[uuid.UUID]Hardware{}
	}
	ds.inventory.Hardware[hardware.ID] = *hardware

	return nil
}

// Get returns a hardware object from the inventory
func (ds *DatastoreJSON) Get(id uuid.UUID) (Hardware, error) {
	ds.inventoryLock.RLock()
	defer ds.inventoryLock.RUnlock()

	if hardware, exists := ds.inventory.Hardware[id]; exists {
		return hardware, nil
	}

	ds.logTransaction("GET", id.String(), nil, nil)

	return Hardware{}, ErrHardwareNotFound
}

// Update updates a hardware object in the inventory
func (ds *DatastoreJSON) Update(hardware *Hardware) error {
	ds.inventoryLock.Lock()
	defer ds.inventoryLock.Unlock()

	// Check to see if parent UUID exists
	if hardware.Parent != uuid.Nil {
		if _, exists := ds.inventory.Hardware[hardware.Parent]; !exists {
			return ErrHardwareParentNotFound
		}
	}

	// Add it to the inventory map
	ds.inventory.Hardware[hardware.ID] = *hardware

	ds.logTransaction("UPDATE", hardware.ID.String(), nil, nil)

	return nil
}

// Remove removes a hardware object from the inventory
func (ds *DatastoreJSON) Remove(id uuid.UUID, recursion bool) error {
	ds.inventoryLock.Lock()
	defer ds.inventoryLock.Unlock()

	// Check to see if this UUID exists
	if _, exists := ds.inventory.Hardware[id]; !exists {
		return ErrHardwareNotFound
	}

	// Check to see if any piece of hardware has this device as a parent
	// as you should not be able to remove a piece of hardware without either
	// delinking it or removing its children
	// FIXME: https://github.com/Cray-HPE/cani/pull/28#discussion_r1199347499
	if children, err := ds.getChildren(id); err != nil {
		return err
	} else if len(children) != 0 {
		childrenIDs := []string{}
		for _, child := range children {
			childrenIDs = append(childrenIDs, child.ID.String())
		}
		// If recursion is true, remove the children as well
		if recursion {
			for _, child := range children {
				delete(ds.inventory.Hardware, child.ID)
			}
		} else {
			return fmt.Errorf("unable to remove (%s) as it is the parent of [%s]", id.String(), strings.Join(childrenIDs, ","))
		}
	}

	// Remove the hardware!
	delete(ds.inventory.Hardware, id)

	ds.logTransaction("REMOVE", id.String(), nil, nil)

	return nil
}

// List returns the entire inventory
func (ds *DatastoreJSON) List() (Inventory, error) {
	ds.inventoryLock.RLock()
	defer ds.inventoryLock.RUnlock()

	return *ds.inventory, nil
}

// GetLocation will follow the parent links up to the root node, which is signaled when a NIL parent UUID is found
// This will either return a partial location path, or a full path up to a cabinet or CDU
// TODO THIS NEEDS UNIT TESTS
func (ds *DatastoreJSON) GetLocation(hardware Hardware) (LocationPath, error) {
	ds.inventoryLock.RLock()
	defer ds.inventoryLock.RUnlock()

	return ds.getLocation(hardware)
}

// getLocation is just like GetLocation except it doesn't attempt to acquire the inventory RWMutex, so it can be
// composed into other DatastoreJSON functions that make changes to the inventory structure.
func (ds *DatastoreJSON) getLocation(hardware Hardware) (LocationPath, error) {

	locationPath := LocationPath{}

	// Follow the parent links up to the root node
	currentHardwareID := hardware.ID
	for currentHardwareID != uuid.Nil {
		currentHardware, exists := ds.inventory.Hardware[currentHardwareID]
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
		locationPath = append(locationPath, LocationToken{
			HardwareType: currentHardware.Type,
			Ordinal:      *currentHardware.LocationOrdinal,
		})

		// Go the parent node next
		currentHardwareID = currentHardware.Parent
	}

	// Reverse in place, since the tree was traversed bottom up
	// This is more efficient than building prepending the location path, due to not
	// needing to a lot of memory allocations and slice magic by adding an new element
	// to the start of the slice every time we visit a new location.
	//
	// For loop
	// Initial condition: Set i to beginning, and j to the end.
	// Check: Continue if i is before j
	// Advance: Move i forward, and j backward
	for i, j := 0, len(locationPath)-1; i < j; i, j = i+1, j-1 {
		locationPath[j], locationPath[i] = locationPath[i], locationPath[j]
	}

	return locationPath, nil

}

// GetAtLocation returns the hardware at the given location
// TODO THIS NEEDS UNIT TESTS
func (ds *DatastoreJSON) GetAtLocation(path LocationPath) (Hardware, error) {
	ds.inventoryLock.RLock()
	defer ds.inventoryLock.RUnlock()

	if len(path) == 0 {
		return Hardware{}, ErrEmptyLocationPath
	}
	log.Trace().Msgf("GetAtLocation: Location Path: %s", path.String())

	//
	// Traverse the tree to see if the hardware exists at the given location
	//
	currentHardware, err := ds.getSystemZero()
	if err != nil {
		return Hardware{}, err
	}

	// Base case - System 0
	if len(path) == 1 {
		if path[0].HardwareType == hardwaretypes.System && path[1].Ordinal == 0 {
			return currentHardware, nil
		} else {
			return Hardware{}, ErrHardwareNotFound
		}
	}

	// Vist rest of the path
	for i, locationToken := range path[1:] {
		log.Trace().Msgf("GetAtLocation: Processing token %d of %d: '%s'", i+1, len(path), locationToken.String())
		log.Trace().Msgf("GetAtLocation: Current ID %s", currentHardware.ID)

		// For each child of the current hardware object check to see if it
		foundMatch := false
		for _, childID := range currentHardware.Children {
			log.Trace().Msgf("GetAtLocation: Visiting Child (%s)", childID)
			// Get the hardware
			childHardware, ok := ds.inventory.Hardware[childID]
			if !ok {
				// This should not happen
				return Hardware{}, errors.Join(
					fmt.Errorf("unable to find hardware object with ID (%s)", childID),
					ErrHardwareNotFound,
				)
			}

			if childHardware.LocationOrdinal == nil {
				log.Trace().Msgf("GetAtLocation: Child has no location ordinal set, skipping")
				continue
			}
			log.Trace().Msgf("GetAtLocation: Child location token: %s:%d", childHardware.Type, *childHardware.LocationOrdinal)

			// Check to see if the location token matches
			if childHardware.Type == locationToken.HardwareType && *childHardware.LocationOrdinal == locationToken.Ordinal {
				// Found a match!
				log.Trace().Msgf("GetAtLocation: Child has matching location token")
				currentHardware = childHardware
				foundMatch = true
				break
			}
		}
		indent := strings.Repeat(" | ", i)

		if i == len(path)-2 { // Subtract 2 instead of 1 because we started from the second element of the slice (path[1:])
			// This is the last iteration of the loop
			if !foundMatch {
				log.Debug().Bool("exists", false).Msgf("%s --%s (%d)", indent, locationToken.HardwareType, locationToken.Ordinal)
			} else {
				log.Debug().Bool("exists", true).Str("uuid", currentHardware.ID.String()).Msgf("%s --%s (%d)", indent, locationToken.HardwareType, locationToken.Ordinal)
			}
		}

		if !foundMatch {
			return Hardware{}, ErrHardwareNotFound
		}
	}

	if currentHardware.ID == uuid.Nil {
		return Hardware{}, ErrHardwareNotFound
	}

	return currentHardware, nil
}

// GetChildren returns the children of a given hardware object
func (ds *DatastoreJSON) GetChildren(id uuid.UUID) ([]Hardware, error) {
	ds.inventoryLock.RLock()
	defer ds.inventoryLock.RUnlock()

	return ds.getChildren(id)
}

// getChildren returns the children of a given hardware object
// This function depends on the cached/derived children data
func (ds *DatastoreJSON) getChildren(id uuid.UUID) ([]Hardware, error) {
	hardware, exists := ds.inventory.Hardware[id]
	if !exists {
		return nil, ErrHardwareNotFound
	}

	// For right we need to iterate over the map, as we don't have any
	// book keeping to keep track of child hardware
	var results []Hardware
	for _, childID := range hardware.Children {
		childHardware, exists := ds.inventory.Hardware[childID]
		if !exists {
			// This should not happen
			return nil, fmt.Errorf("unable to find child hardware object with ID (%s) of (%s)", childHardware.ID, hardware.ID)
		}

		results = append(results, childHardware)
	}

	return results, nil
}

func (ds *DatastoreJSON) GetDescendants(id uuid.UUID) ([]Hardware, error) {
	ds.inventoryLock.RLock()
	defer ds.inventoryLock.RUnlock()

	results := []Hardware{}
	callback := func(h Hardware) error {
		results = append(results, h)
		return nil
	}

	if err := ds.traverseByLocation(id, callback); err != nil {
		return nil, err
	}

	return results, nil
}

func (ds *DatastoreJSON) traverseByLocation(rootID uuid.UUID, callback func(h Hardware) error) error {
	queue := []uuid.UUID{rootID}

	for len(queue) != 0 {
		// Pull next ID from the queue
		hardwareID := queue[0]
		queue = queue[1:]

		// Retrieve the hardware object
		hardware, exists := ds.inventory.Hardware[hardwareID]
		if !exists {
			// This should not happen
			return fmt.Errorf("unable to find hardware object with ID (%s)", hardware.ID)
		}

		// Visit the hardware object
		if err := callback(hardware); err != nil {
			return errors.Join(fmt.Errorf("callback failed on hardware object with ID (%s)", hardware.ID), err)
		}

		// Add the children to the queue
		queue = append(queue, hardware.Children...)
	}

	return nil
}

// logTransaction logs a transaction to logger
func (ds *DatastoreJSON) logTransaction(operation string, key string, value interface{}, err error) {
	if ds.dataFilePath == "" {
		// If running in in-memory mode there is currently no place to log
		return
	}

	tl, err = os.OpenFile(
		ds.logFilePath,
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

// GetSystem returns the system that the given hardware object is a part of
func (ds *DatastoreJSON) GetSystem(hardware Hardware) (Hardware, error) {
	ds.inventoryLock.RLock()
	defer ds.inventoryLock.RUnlock()

	return Hardware{}, errors.New("not yet implemented")
	return ds.getSystem(hardware)
}

// getSystem will follow the parent links up to the root node and errors if not found
// this is not in use until multiple systems are implemented
func (ds *DatastoreJSON) getSystem(hardware Hardware) (Hardware, error) {
	// Follow the parent links up to the root node
	currentHardwareID := hardware.ID
	for currentHardwareID != uuid.Nil {
		currentHardware, exists := ds.inventory.Hardware[currentHardwareID]
		if !exists {
			return Hardware{}, errors.Join(
				fmt.Errorf("unable to find ancestor (%s) of (%s)", currentHardwareID, hardware.ID),
				ErrHardwareNotFound,
			)
		}

		if currentHardware.Type == hardwaretypes.System {
			// return the system
			return currentHardware, nil
		}
		// Go the parent node next
		currentHardwareID = currentHardware.Parent
	}

	return Hardware{}, ErrHardwareNotFound
}

// GetSystemZero assumes one system exists and returns it
func (ds *DatastoreJSON) GetSystemZero() (Hardware, error) {
	ds.inventoryLock.RLock()
	defer ds.inventoryLock.RUnlock()

	return ds.getSystemZero()
}

// getSystem finds the System type in the inventory and returns it
// it does not currently detect multiple systems and may grab the wrong one if multiple exist
func (ds *DatastoreJSON) getSystemZero() (Hardware, error) {
	// Assume one system
	for _, hw := range ds.inventory.Hardware {
		system, exists := ds.inventory.Hardware[hw.ID]
		if !exists {
			return Hardware{}, errors.Join(
				fmt.Errorf("unable to find %s (%s)", hardwaretypes.System, hw.ID),
				ErrHardwareNotFound,
			)
		}

		if system.Type == hardwaretypes.System {
			// return the system
			return system, nil
		}
	}

	return Hardware{}, ErrHardwareNotFound
}

func (ds *DatastoreJSON) Search(filter SearchFilter) (map[uuid.UUID]Hardware, error) {
	ds.inventoryLock.RLock()
	defer ds.inventoryLock.RUnlock()

	// Build up lookup maps based on the filter
	wantedTypes := map[hardwaretypes.HardwareType]bool{}
	for _, wantedType := range filter.Types {
		wantedTypes[wantedType] = true
	}

	wantedStatus := map[HardwareStatus]bool{}
	for _, status := range filter.Status {
		wantedStatus[status] = true
	}

	return ds.inventory.FilterHardware(func(h Hardware) (bool, error) {
		matchType := false
		if len(wantedTypes) == 0 || wantedTypes[h.Type] {
			matchType = true
		}

		matchStatus := false
		if len(wantedStatus) == 0 || wantedStatus[h.Status] {
			matchStatus = true
		}

		return matchType && matchStatus, nil
	})
}
