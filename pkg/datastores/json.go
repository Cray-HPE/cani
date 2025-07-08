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
package datastores

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/internal/core"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// JSONStore handles inventory persistence
type JSONStore struct {
	Path string
}

// NewJSONStore creates a new store with the config path
func NewJSONStore() *JSONStore {
	// Use the config directory but with our inventory filename
	return &JSONStore{
		Path: filepath.Join(filepath.Dir(config.Cfg.Path), filepath.Base(config.Cfg.Datastore)),
	}
}

// Load retrieves inventory from disk
func (s *JSONStore) Load() (*devicetypes.Inventory, error) {
	inventory := &devicetypes.Inventory{
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
	}

	// Check if file exists
	if _, err := os.Stat(s.Path); os.IsNotExist(err) {
		// No inventory yet, return empty
		return inventory, nil
	}

	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, fmt.Errorf("reading inventory file: %w", err)
	}

	if err := json.Unmarshal(data, inventory); err != nil {
		return nil, fmt.Errorf("parsing inventory: %w", err)
	}

	if len(inventory.Devices) == 0 {
		log.Printf("No devices found in inventory, starting with an empty inventory")
	} else {
		log.Printf("Loaded %d devices from inventory", len(inventory.Devices))
	}

	if inventory.Systems() == nil || len(inventory.Systems()) == 0 {
		log.Printf("No systems found in inventory, adding a default system")
		system := devicetypes.NewSystem()
		inventory.Devices[system.ID] = system
	} else {
		log.Printf("Found %d systems in inventory", len(inventory.Systems()))
	}

	log.Printf("Loaded inventory from %s", s.Path)
	return inventory, nil
}

// Save writes inventory to disk
func (s *JSONStore) Save(inventory *devicetypes.Inventory) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(s.Path), 0755); err != nil {
		return fmt.Errorf("creating inventory directory: %w", err)
	}

	data, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding inventory: %w", err)
	}

	if err := os.WriteFile(s.Path, data, 0644); err != nil {
		return fmt.Errorf("writing inventory file: %w", err)
	}

	return nil
}

// Create adds new devices to the store
func (s *JSONStore) Create(devices map[uuid.UUID]*devicetypes.CaniDeviceType) error {
	// Load existing inventory
	inventory, err := s.Load()
	if err != nil {
		return err
	}

	// Ensure devices have parents assigned
	if err := s.ensureDevicesHaveParents(inventory, devices); err != nil {
		return err
	}

	// Add new devices
	for id, device := range devices {
		if _, exists := inventory.Devices[id]; exists {
			return fmt.Errorf("device with ID %s already exists", id)
		}
		inventory.Devices[id] = device
	}

	inventory.VerifyParentChildRelationships()

	// Save updated inventory
	return s.Save(inventory)
}

// Read retrieves devices from the store
func (s *JSONStore) Read(ids []uuid.UUID) (map[uuid.UUID]*devicetypes.CaniDeviceType, error) {
	// Load inventory
	inventory, err := s.Load()
	if err != nil {
		return nil, err
	}

	// If no IDs provided, return all devices
	if len(ids) == 0 {
		return inventory.Devices, nil
	}

	// Return only requested devices
	result := make(map[uuid.UUID]*devicetypes.CaniDeviceType)
	for _, id := range ids {
		if device, exists := inventory.Devices[id]; exists {
			result[id] = device
		}
	}

	return result, nil
}

// Update updates existing devices in the store
func (s *JSONStore) Update(devices map[uuid.UUID]*devicetypes.CaniDeviceType) error {
	// Load inventory
	inventory, err := s.Load()
	if err != nil {
		return err
	}

	// Check if parent reassignment is needed
	reassignParents := false
	for _, device := range devices {
		if _, exists := inventory.Devices[device.ID]; exists {
			// Only check devices that exist
			if device.Parent == uuid.Nil && device.Type != "system" {
				reassignParents = true
				break
			}
		}
	}

	// Ask if user wants to reassign parents
	if reassignParents {
		shouldReassign, err := core.PromptForConfirmation("Some devices don't have parents. Reassign them?")
		if err != nil {
			return err
		}
		if shouldReassign {
			if err := s.ensureDevicesHaveParents(inventory, devices); err != nil {
				return err
			}
		}
	}

	// Update devices
	for id, device := range devices {
		if _, exists := inventory.Devices[id]; !exists {
			return fmt.Errorf("device with ID %s does not exist", id)
		}
		inventory.Devices[id] = device
	}

	// Save updated inventory
	return s.Save(inventory)
}

// Delete removes devices from the store
func (s *JSONStore) Delete(ids []uuid.UUID) error {
	// Load inventory
	inventory, err := s.Load()
	if err != nil {
		return err
	}

	// Validate that all IDs exist
	var nonExistentIDs []uuid.UUID
	for _, id := range ids {
		if _, exists := inventory.Devices[id]; !exists {
			nonExistentIDs = append(nonExistentIDs, id)
		}
	}

	if len(nonExistentIDs) > 0 {
		return fmt.Errorf("the following device IDs do not exist: %v", nonExistentIDs)
	}

	// Check if we're deleting systems or racks with children
	systemsWithChildren := make(map[uuid.UUID][]string) // Map of system ID to child names
	for _, id := range ids {
		log.Printf("Checking device %s for children", id)
		if device, exists := inventory.Devices[id]; exists && (device.Type == devicetypes.System || device.Type == devicetypes.Rack) {
			// Find children of this system or rack
			childNames := []string{}
			for _, potentialChild := range inventory.Devices {
				if potentialChild.Parent == id {
					childNames = append(childNames, potentialChild.Name)
				}
			}

			if len(childNames) > 0 {
				systemsWithChildren[id] = childNames
			}
		}
	}

	// Warn about orphaned children
	if len(systemsWithChildren) > 0 {
		fmt.Println("Warning: The following systems have child devices that will be orphaned:")
		for systemID, children := range systemsWithChildren {
			system := inventory.Devices[systemID]
			fmt.Printf("- System '%s' has %d children: %s\n", system.Name, len(children), strings.Join(children, ", "))
		}

		confirm, err := core.PromptForConfirmation("Do you want to proceed and delete these systems?")
		if err != nil {
			return err
		}
		if !confirm {
			return fmt.Errorf("operation canceled by user")
		}

		// Ask if orphaned devices should be reassigned
		reassign, err := core.PromptForConfirmation("Do you want to reassign orphaned devices to another system?")
		if err != nil {
			return err
		}

		if reassign {
			log.Printf("Reassigning orphaned devices to a system...")
			// Gather all orphaned devices
			orphanedDevices := make(map[uuid.UUID]*devicetypes.CaniDeviceType)
			for systemID := range systemsWithChildren {
				for devID, device := range inventory.Devices {
					if device.Parent == systemID {
						orphanedDevices[devID] = device
					}
				}
			}

			// Create temporary inventory without deleted systems
			tempInventory := &devicetypes.Inventory{
				Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
			}

			for id, device := range inventory.Devices {
				// Skip devices being deleted
				shouldSkip := false
				for _, deleteID := range ids {
					if id == deleteID {
						shouldSkip = true
						break
					}
				}

				if !shouldSkip {
					tempInventory.Devices[id] = device
				}
			}

			// Run parent reassignment on orphaned devices
			if err := s.ensureDevicesHaveParents(tempInventory, orphanedDevices); err != nil {
				return err
			}

			// Update original inventory with reassigned devices
			for id, device := range orphanedDevices {
				inventory.Devices[id] = device
			}
		}
	}

	// Delete devices
	for _, id := range ids {
		log.Printf("Deleting device with ID %s", id)

		// Before deleting, remove this device from its parent's children list
		if device, exists := inventory.Devices[id]; exists {
			if device.Parent != uuid.Nil {
				if parent, parentExists := inventory.Devices[device.Parent]; parentExists {
					// Remove this device ID from parent's children slice
					updatedChildren := make([]uuid.UUID, 0, len(parent.Children))
					for _, childID := range parent.Children {
						if childID != id {
							updatedChildren = append(updatedChildren, childID)
						}
					}
					parent.Children = updatedChildren
					log.Printf("Removed device %s from parent %s's children list", id, device.Parent)
				}
			}
		}

		delete(inventory.Devices, id)
	}

	// Check if we deleted all systems
	if len(inventory.Systems()) == 0 {
		log.Printf("Adding a system since all were deleted")
		system := devicetypes.NewSystem()
		inventory.Devices[system.ID] = system
	}

	// Save updated inventory
	return s.Save(inventory)
}

// findDevicesWithDanglingReferences finds devices that reference deleted devices as parents
func (s *JSONStore) findDevicesWithDanglingReferences(inventory *devicetypes.Inventory, deletedIDs []uuid.UUID) map[uuid.UUID]*devicetypes.CaniDeviceType {
	danglingDevices := make(map[uuid.UUID]*devicetypes.CaniDeviceType)

	for deviceID, device := range inventory.Devices {
		// Skip devices that are being deleted
		isBeingDeleted := false
		for _, deletedID := range deletedIDs {
			if deviceID == deletedID {
				isBeingDeleted = true
				break
			}
		}
		if isBeingDeleted {
			continue
		}

		// Check if this device's parent is being deleted
		for _, deletedID := range deletedIDs {
			if device.Parent == deletedID {
				danglingDevices[deviceID] = device
				break
			}
		}
	}

	return danglingDevices
}

// ensureSystemSelection handles system selection or creation
// It returns the selected system ID and a modified inventory
func (s *JSONStore) ensureSystemSelection(inventory *devicetypes.Inventory) (uuid.UUID, error) {
	systems := inventory.Systems()

	if len(systems) == 0 {
		// No systems exist, create one
		log.Printf("Adding a system since none exist in the inventory")
		system := devicetypes.NewSystem()
		inventory.Devices[system.ID] = system
		return system.ID, nil
	} else if len(systems) > 1 {
		// Multiple systems exist, let user choose one
		fmt.Println("Select a system for the operation:")
		for i, system := range systems {
			fmt.Printf("  %d. %s (%v)\n", i+1, system.Name, system.ID)
		}

		// Prompt for selection
		var selectedIndex int
		for {
			fmt.Print("Select a system (enter number): ")
			var input string
			fmt.Scanln(&input)

			// Parse input
			if i, err := strconv.Atoi(input); err == nil && i > 0 && i <= len(systems) {
				selectedIndex = i - 1
				break
			}
			fmt.Println("Invalid selection. Please try again.")
		}

		selectedSystem := systems[selectedIndex]
		log.Printf("Selected system: %s", selectedSystem.Name)
		return selectedSystem.ID, nil
	} else {
		// Only one system exists
		chosenSystem, err := core.PromptForConfirmation(
			fmt.Sprintf("Use system '%s' (%v)?", systems[0].Name, systems[0].ID))
		if err != nil {
			return uuid.Nil, err
		}
		if !chosenSystem {
			return uuid.Nil, fmt.Errorf("operation canceled by user")
		}
		return systems[0].ID, nil
	}
}

// ensureDevicesHaveParents ensures all non-system devices have valid parent IDs
func (s *JSONStore) ensureDevicesHaveParents(inventory *devicetypes.Inventory, devices map[uuid.UUID]*devicetypes.CaniDeviceType) error {
	// Check if any device needs a parent
	needParent := false
	for _, device := range devices {
		if device.Type == devicetypes.Rack && device.Parent == uuid.Nil {
			needParent = true
			break
		}
	}

	if !needParent {
		return nil // No parenting needed
	}

	// Get system selection
	systemID, err := s.ensureSystemSelection(inventory)
	if err != nil {
		return err
	}

	// Assign parent to devices that need it
	for _, device := range devices {
		if device.Type == devicetypes.Rack && device.Parent == uuid.Nil {
			device.Parent = systemID
			log.Printf("Set parent of %s to system %v", device.Name, systemID)
		}
	}

	return nil
}
