/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Cray-HPE/cani/internal/util/uuidutil"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type DatastoreJSONCSM struct {
	inventoryLock sync.RWMutex
	inventory     *Inventory
	dataFilePath  string
	logFilePath   string
}

func NewDatastoreJSONCSM(dataFilePath string, logfilepath string, provider Provider) (Datastore, error) {
	datastore := &DatastoreJSONCSM{
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
		// A system ordinal is required for the top-level system object and is arbitrarily set to 0
		ordinal := 0
		// Create a config with default values since one does not exist
		datastore.inventory = &Inventory{
			SchemaVersion: SchemaVersionV1Alpha1,
			Provider:      provider,
			Hardware: map[uuid.UUID]Hardware{
				// NOTE: At present, we only allow ONE system in the inventory, but leaving the door open for multiple systems
				system: {
					Type:            hardwaretypes.System, // The top-level object is a hardwaretypes.System
					ID:              system,               // ID is the same as the key for the top-level system object to prevent a uuid.Nil
					Parent:          uuid.Nil,             // Parent should be nil to prevent illegitimate children
					LocationOrdinal: &ordinal,
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

func NewDatastoreInMemoryCSM(provider Provider) (*DatastoreJSONCSM, error) {
	datastore := &DatastoreJSONCSM{}

	// Generate a UUID for a new top-level "System" object
	system := uuid.New()
	// A system ordinal is required for the top-level system object and is arbitrarily set to 0
	ordinal := 0
	// Create a config with default values since one does not exist
	datastore.inventory = &Inventory{
		SchemaVersion: SchemaVersionV1Alpha1,
		Provider:      provider,
		Hardware: map[uuid.UUID]Hardware{
			// NOTE: At present, we only allow ONE system in the inventory, but leaving the door open for multiple systems
			system: {
				Type:            hardwaretypes.System, // The top-level object is a hardwaretypes.System
				ID:              system,               // ID is the same as the key for the top-level system object to prevent a uuid.Nil
				Parent:          uuid.Nil,             // Parent should be nil to prevent illegitimate children
				LocationOrdinal: &ordinal,
			},
		},
	}

	return datastore, nil
}

func (dj *DatastoreJSONCSM) GetSchemaVersion() (SchemaVersion, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return dj.inventory.SchemaVersion, nil
}

// SetExternalInventoryProvider sets the external inventory provider
func (dj *DatastoreJSONCSM) SetInventoryProvider(provider Provider) error {
	dj.inventoryLock.Lock()
	defer dj.inventoryLock.Unlock()

	dj.inventory.Provider = provider

	return nil
}

// GetExternalInventoryProvider gets the external inventory provider
func (dj *DatastoreJSONCSM) InventoryProvider() (Provider, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return dj.inventory.Provider, nil
}

// Flush writes the current inventory to the datastore
func (dj *DatastoreJSONCSM) Flush() error {
	if dj.dataFilePath == "" {
		// If running in in-memory mode there is nothing to flush
		return nil
	}

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

func (dj *DatastoreJSONCSM) Clone() (Datastore, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	result, err := NewDatastoreInMemoryCSM(dj.inventory.Provider)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to create in memory datastore"), err)
	}

	// Deep copy the hardware information into the datastore
	// TODO this is a hack
	raw, err := json.Marshal(dj.inventory.Hardware)
	if err != nil {
		return nil, err
	}
	result.inventory.Hardware = nil
	if err := json.Unmarshal(raw, &result.inventory.Hardware); err != nil {
		return nil, err
	}

	if err := result.calculateDerivedFields(); err != nil {
		return nil, err
	}

	return result, nil
}

func (dj *DatastoreJSONCSM) Merge(otherDJ Datastore) error {
	dj.inventoryLock.Lock()
	defer dj.inventoryLock.Unlock()

	otherAllHardware, err := otherDJ.List()
	if err != nil {
		return errors.Join(fmt.Errorf("failed to retrieve inventory from other datastore"), err)
	}

	// Identify hardware to remove
	hardwareToDelete := []uuid.UUID{}
	for id := range dj.inventory.Hardware {
		if _, exists := otherAllHardware.Hardware[id]; !exists {
			hardwareToDelete = append(hardwareToDelete, id)
		}
	}
	// Remove deleted hardware
	for _, id := range hardwareToDelete {
		delete(dj.inventory.Hardware, id)
	}

	// Update or add hardware
	for id, otherHardware := range otherAllHardware.Hardware {
		dj.inventory.Hardware[id] = otherHardware
	}

	return nil
}

func (dj *DatastoreJSONCSM) calculateDerivedFields() (err error) {
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
func (dj *DatastoreJSONCSM) Validate() (map[uuid.UUID]ValidateResult, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	log.Warn().Msg("DatastoreJSONCSM's Validate was called. This is not fully implemented")
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

	// Build up a lookup map of location paths to hardware UUIDs
	foundLocationPaths := map[string][]uuid.UUID{}
	for _, hardware := range dj.inventory.Hardware {
		if len(hardware.LocationPath) == 0 {
			continue
		}

		if hardware.LocationPath[0].HardwareType != hardwaretypes.System {
			// Skip any piece of hardware that does not begin with System type, as it is not complete
			continue
		}

		key := hardware.LocationPath.String()
		foundLocationPaths[key] = append(foundLocationPaths[key], hardware.ID)
	}

	// Verify all location paths have only one Hardware object present
	for _, hardwareIDs := range foundLocationPaths {
		if len(hardwareIDs) == 1 {
			continue
		}

		for _, hardwareID := range hardwareIDs {
			if _, exists := validationResults[hardwareID]; !exists {
				validationResults[hardwareID] = ValidateResult{Hardware: dj.inventory.Hardware[hardwareID]}
			}

			// Add the validation error and push it back into the map
			validationResult := validationResults[hardwareID]
			validationResult.Errors = append(validationResult.Errors,
				fmt.Sprintf("Location path not unique shared by: %s", uuidutil.Join(hardwareIDs, ", ", hardwareID)),
			)
			validationResults[hardwareID] = validationResult
		}
	}

	if len(validationResults) > 0 {
		return validationResults, ErrDatastoreValidationFailure
	}

	return nil, nil
}

// Add adds a new hardware object to the inventory
func (dj *DatastoreJSONCSM) Add(hardware *Hardware) error {
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
func (dj *DatastoreJSONCSM) Get(id uuid.UUID) (Hardware, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	if hardware, exists := dj.inventory.Hardware[id]; exists {
		return hardware, nil
	}

	dj.logTransaction("GET", id.String(), nil, nil)

	return Hardware{}, ErrHardwareNotFound
}

// Update updates a hardware object in the inventory
func (dj *DatastoreJSONCSM) Update(hardware *Hardware) error {
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
	if oldHardware.Parent != hardware.Parent {
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
func (dj *DatastoreJSONCSM) Remove(id uuid.UUID, recursion bool) error {
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
func (dj *DatastoreJSONCSM) List() (Inventory, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return *dj.inventory, nil
}

// GetLocation will follow the parent links up to the root node, which is signaled when a NIL parent UUID is found
// This will either return a partial location path, or a full path up to a cabinet or CDU
// TODO THIS NEEDS UNIT TESTS
func (dj *DatastoreJSONCSM) GetLocation(hardware Hardware) (LocationPath, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return dj.getLocation(hardware)
}

// getLocation is just like GetLocation except it doesn't attempt to acquire the inventory RWMutex, so it can be
// composed into other DatastoreJSONCSM functions that make changes to the inventory structure.
func (dj *DatastoreJSONCSM) getLocation(hardware Hardware) (LocationPath, error) {

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
func (dj *DatastoreJSONCSM) GetAtLocation(path LocationPath) (Hardware, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	if len(path) == 0 {
		return Hardware{}, ErrEmptyLocationPath
	}
	log.Trace().Msgf("GetAtLocation: Location Path: %s", path.String())

	//
	// Traverse the tree to see if the hardware exists at the given location
	//
	currentHardware, err := dj.getSystemZero()
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
			childHardware, ok := dj.inventory.Hardware[childID]
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
func (dj *DatastoreJSONCSM) GetChildren(id uuid.UUID) ([]Hardware, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return dj.getChildren(id)
}

// getChildren returns the children of a given hardware object
// This function depends on the cached/derived children data
func (dj *DatastoreJSONCSM) getChildren(id uuid.UUID) ([]Hardware, error) {
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

func (dj *DatastoreJSONCSM) GetDescendants(id uuid.UUID) ([]Hardware, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	results := []Hardware{}
	callback := func(h Hardware) error {
		results = append(results, h)
		return nil
	}

	if err := dj.traverseByLocation(id, callback); err != nil {
		return nil, err
	}

	return results, nil
}

func (dj *DatastoreJSONCSM) traverseByLocation(rootID uuid.UUID, callback func(h Hardware) error) error {
	queue := []uuid.UUID{rootID}

	for len(queue) != 0 {
		// Pull next ID from the queue
		hardwareID := queue[0]
		queue = queue[1:]

		// Retrieve the hardware object
		hardware, exists := dj.inventory.Hardware[hardwareID]
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
func (dj *DatastoreJSONCSM) logTransaction(operation string, key string, value interface{}, err error) {
	if dj.dataFilePath == "" {
		// If running in in-memory mode there is currently no place to log
		return
	}

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

// GetSystem returns the system that the given hardware object is a part of
func (dj *DatastoreJSONCSM) GetSystem(hardware Hardware) (Hardware, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return Hardware{}, errors.New("not yet implemented")
	return dj.getSystem(hardware)
}

// getSystem will follow the parent links up to the root node and errors if not found
// this is not in use until multiple systems are implemented
func (dj *DatastoreJSONCSM) getSystem(hardware Hardware) (Hardware, error) {
	// Follow the parent links up to the root node
	currentHardwareID := hardware.ID
	for currentHardwareID != uuid.Nil {
		currentHardware, exists := dj.inventory.Hardware[currentHardwareID]
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
func (dj *DatastoreJSONCSM) GetSystemZero() (Hardware, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	return dj.getSystemZero()
}

// getSystem finds the System type in the inventory and returns it
// it does not currently detect multiple systems and may grab the wrong one if multiple exist
func (dj *DatastoreJSONCSM) getSystemZero() (Hardware, error) {
	// Assume one system
	for _, hw := range dj.inventory.Hardware {
		system, exists := dj.inventory.Hardware[hw.ID]
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

func (dj *DatastoreJSONCSM) Search(filter SearchFilter) (map[uuid.UUID]Hardware, error) {
	dj.inventoryLock.RLock()
	defer dj.inventoryLock.RUnlock()

	// Build up lookup maps based on the filter
	wantedTypes := map[hardwaretypes.HardwareType]bool{}
	for _, wantedType := range filter.Types {
		wantedTypes[wantedType] = true
	}

	wantedStatus := map[HardwareStatus]bool{}
	for _, status := range filter.Status {
		wantedStatus[status] = true
	}

	return dj.inventory.FilterHardware(func(h Hardware) (bool, error) {
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
