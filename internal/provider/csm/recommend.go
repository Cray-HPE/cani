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
package csm

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (csm *CSM) RecommendHardware(inv inventory.Inventory, deviceTypeSlug string) (recommended provider.HardwareRecommendations, err error) {
	// loop through the existing inventory to check for vlans
	log.Debug().Msg("Checking existing hardware to find recommendations")
	deviceType, exists := csm.hardwareLibrary.DeviceTypes[deviceTypeSlug]
	if !exists {
		return recommended, err
	}
	log.Debug().Msgf("Recommending a %v", deviceType.HardwareType)

	switch deviceType.HardwareType {
	case hardwaretypes.Cabinet:
		r, err := csm.recommendCabinet(inv, deviceTypeSlug)
		if err != nil {
			return recommended, err
		}
		recommended = r

	case hardwaretypes.NodeBlade:
		// Get the cabinet recommendations
		r, err := csm.recommendBlade(inv, deviceTypeSlug)
		if err != nil {
			return recommended, err
		}
		recommended = r

	default:
		// This function only handles cabinets and blades
	}

	// return the recommendations
	return recommended, nil
}

func (csm *CSM) recommendCabinet(inv inventory.Inventory, deviceTypeSlug string) (recommended provider.HardwareRecommendations, err error) {
	// slice to track existing vlans
	var existingVlans = []int{}
	// slice to track existing cabinets
	var existingCabinets = []int{}

	// loop through the existing inventory to check for vlans
	log.Debug().Msg("Checking existing hardware to find recommendations")
	for _, cHardware := range inv.Hardware {
		switch cHardware.Type {
		case hardwaretypes.Cabinet:
			log.Debug().Msgf("Checking %s (%s)", cHardware.Type, cHardware.ID.String())
			log.Debug().Msgf("Decoding csm properties %+v", cHardware.ProviderMetadata)

			metadata, err := DecodeProviderMetadata(cHardware)
			if err != nil {
				return provider.HardwareRecommendations{}, errors.Join(fmt.Errorf("failed to decode CSM metadata from cabinet (%v)", cHardware.ID), err)
			}

			if metadata.Cabinet != nil && metadata.Cabinet.HMNVlan != nil {
				// add it to the slice that tracks existing vlans
				existingVlans = append(existingVlans, *metadata.Cabinet.HMNVlan)
			}

			// add the ordinal to the existing cabinets slice for choosing a new one later
			existingCabinets = append(existingCabinets, *cHardware.LocationOrdinal)

		default:
			// This function only handles cabinets
			continue
		}
	}

	var chosenOrdinal int
	chosenOrdinal, err = DetermineStartingOrdinalFromSlug(deviceTypeSlug, *csm.hardwareLibrary)
	if err != nil {
		return recommended, err
	}
	log.Debug().Msgf("chosenOrdinal %d (%s)", chosenOrdinal, deviceTypeSlug)
	// Set the cabinet location
	if len(existingCabinets) == 0 {
		// there are no cabinets yet, so set it to the provider default
		recommended.CabinetOrdinal = chosenOrdinal
		log.Debug().Msgf("No cabinets found, using %d", recommended.CabinetOrdinal)
	} else {
		// set the recommended cabinet number
		recommended.CabinetOrdinal = nextAvailableInt(existingCabinets, chosenOrdinal)
		log.Debug().Msgf("Existing cabinets found (%v), using %d", existingCabinets, recommended.CabinetOrdinal)
	}

	// Determine the hardware class based off the slug
	// This is needed to assign an approriate VLAN from the ranges defined above
	class, err := DetermineHardwareClassFromSlug(deviceTypeSlug, *csm.hardwareLibrary)
	if err != nil {
		return recommended, err
	}

	// Set the metadata vlan
	var startingVlan, chosenVlan int
	startingVlan, err = DetermineStartingVlanFromSlug(deviceTypeSlug, *csm.hardwareLibrary)
	if err != nil {
		return recommended, err
	}
	if len(existingCabinets) == 0 {
		// choose a starting vlan based on the class
		chosenVlan = startingVlan
		log.Debug().Msgf("No cabinet VLANs found, using %d for %s %s", chosenVlan, class, hardwaretypes.Cabinet)
	} else {
		// set the recommended vlan by finding an available one from the existing
		chosenVlan = nextAvailableInt(existingVlans, startingVlan)

	}

	// set the provider metadata
	recommended.ProviderMetadata = map[string]interface{}{
		// there are no vlans yet, and presumably no cabinets, so set it to 1
		ProviderMetadataVlanId: chosenVlan,
	}

	// return the recommendations
	return recommended, nil
}

// recommendBlade recommends values for cabinet, chassis, and blade ordinals
// by collecting all of the possible ordinals for a given cabinet/chassis
// then checking to see if a slot is populated or not
func (csm *CSM) recommendBlade(inv inventory.Inventory, deviceTypeSlug string) (recommended provider.HardwareRecommendations, err error) {
	// struct for sorting a map by their ordinal location
	type ordinalmap struct {
		key      uuid.UUID
		ordinals []int
	}

	// a slot will be a nodebay and can be populated or not
	type slot struct {
		ordinals  []int
		populated bool
		id        uuid.UUID
	}

	// get all cabinets, chassis, and blades
	allCabinets := inv.FilterHardwareByType(hardwaretypes.Cabinet)
	if err != nil {
		return recommended, err
	}
	allChassis := inv.FilterHardwareByType(hardwaretypes.Chassis)
	if err != nil {
		return recommended, err
	}
	allBlades := inv.FilterHardwareByType(hardwaretypes.NodeBlade)
	if err != nil {
		return recommended, err
	}

	// map uuid of the hw types to their location path so they can be sorted
	cabinetMap := getOrdinals(allChassis)

	// sort the cabinets by their ordinals
	var sortedCabinetSlice []ordinalmap
	for i, cabinetLocations := range cabinetMap {
		l := ordinalmap{
			key:      i,
			ordinals: cabinetLocations,
		}
		sortedCabinetSlice = append(sortedCabinetSlice, l)
	}
	sort.Slice(sortedCabinetSlice, func(i, j int) bool {
		iSlice := sortedCabinetSlice[i].ordinals
		jSlice := sortedCabinetSlice[j].ordinals
		for x := 0; x < len(iSlice) && x < len(jSlice); x++ {
			if iSlice[x] < jSlice[x] {
				return true
			} else if iSlice[x] > jSlice[x] {
				return false
			}
		}
		return false // equal slices
	})

	// map uuid of the hw types to their location path so they can be sorted
	bladeMap := getOrdinals(allBlades)
	// create a sorted list of what exists and is or is not populated with a blade
	existing := []slot{}

	// for each chassis (now in order)
	for _, chass := range sortedCabinetSlice {
		// get the chassis and cabinet Hardware
		chassis := allChassis[chass.key]
		cabinet := allCabinets[chassis.Parent]

		// add each blade to a slice for sorting
		sortedBladeSlice := []ordinalmap{}
		for i, bladeLocations := range bladeMap {
			blade := allBlades[i]
			if blade.Parent == chassis.ID {
				l := ordinalmap{
					key:      i,
					ordinals: bladeLocations,
				}
				sortedBladeSlice = append(sortedBladeSlice, l)
			}
		}
		// compare slices element-by-element
		// move the the next element only if the current elements are equal
		sort.Slice(sortedBladeSlice, func(i, j int) bool {
			iSlice := sortedBladeSlice[i].ordinals
			jSlice := sortedBladeSlice[j].ordinals
			for x := 0; x < len(iSlice) && x < len(jSlice); x++ {
				if iSlice[x] < jSlice[x] {
					return true
				} else if iSlice[x] > jSlice[x] {
					return false
				}
			}
			return false // equal slices
		})

		// check the chassis specification to determine how many slots it should have
		chassisSpecs, exists := csm.hardwareLibrary.DeviceTypes[chassis.DeviceTypeSlug]
		if !exists {
			return recommended, err
		}

		// check to ensure the bay supports a blade
		var bladeSupported bool
		var bay hardwaretypes.DeviceBay
		// loop through all of the devices bays
		for _, b := range chassisSpecs.DeviceBays {
			bay = b
			// bays can have different things than just blades in them
			for _, t := range b.Allowed.Types {
				// so check if the bay supports a NodeBlade
				if t == hardwaretypes.NodeBlade {
					bladeSupported = true
					break
				} else {
					// ignore non-blade bays
					bladeSupported = false
					continue
				}
			}
			// if a blade is supported
			if bladeSupported {
				baySlot := slot{}
				baySlot.populated = false
				baySlot.ordinals = []int{0, *cabinet.LocationOrdinal, *chassis.LocationOrdinal, bay.Ordinal}
				for _, bld := range sortedBladeSlice {
					blade := allBlades[bld.key]
					// if the blade matches the bay, a blade is there
					if *blade.LocationOrdinal == bay.Ordinal {
						if blade.Status != inventory.HardwareStatusEmpty {
							baySlot.populated = true
						}
						baySlot.ordinals = blade.LocationPath.GetOrdinalPath()
						baySlot.id = blade.ID
					}
				}
				// append it to what exists in the chassis
				existing = append(existing, baySlot)
			}
		}
	}

	// loop throuigh what exists
	var nextAvailable *slot
	for e := range existing {
		log.Debug().Msgf("%s %d -> %s %d -> %s %s %d -> %s %s",
			hardwaretypes.Cabinet, existing[e].ordinals[1],
			hardwaretypes.Chassis, existing[e].ordinals[2],
			hardwaretypes.NodeBlade, "Bay", existing[e].ordinals[3],
			hardwaretypes.NodeBlade, existing[e].id.String())
		// if it is not populated, that is the next available
		if !existing[e].populated {
			nextAvailable = &existing[e]
			break
		}
	}
	if nextAvailable != nil {
		log.Debug().Msgf("Next available: %v", nextAvailable.ordinals)
		recommended.CabinetOrdinal = nextAvailable.ordinals[1]
		recommended.ChassisOrdinal = nextAvailable.ordinals[2]
		recommended.BladeOrdinal = nextAvailable.ordinals[3]
	} else {
		return recommended, fmt.Errorf("no available %s slots", string(hardwaretypes.NodeBlade))
	}

	// return the recommendations
	return recommended, nil
}

// nextAvailableInt finds the next available integer accounting for gaps
func nextAvailableInt(s []int, offset int) int {
	// slice must be sorted in order to work properly
	sort.Ints(s)

	// If the slice is empty, return the offset
	if len(s) == 0 || offset < s[0] {
		return offset
	}
	// Check if the offset is in the slice
	i := sort.Search(len(s), func(i int) bool { return s[i] >= offset })
	if i < len(s) && s[i] == offset {
		for ; i < len(s); i++ {
			// if it is the last element or there is a gap to the next one
			if i == len(s)-1 || s[i+1]-s[i] > 1 {
				return s[i] + 1
			}
		}
	}
	// the offset is not in the slice, so return it
	return offset
}

// getOrdinals returns a map of uuids to a slice of a hardware's ordinal path
func getOrdinals(hwMap map[uuid.UUID]inventory.Hardware) (ordinals map[uuid.UUID][]int) {
	// make a map of existing hw and the ordinals they contain
	ordinals = make(map[uuid.UUID][]int, 0)

	// for each harware in the given map
	for _, hw := range hwMap {
		// check if the hw is already in the map
		_, exists := ordinals[hw.ID]
		// if not in the map, add it
		if !exists {
			ordinals[hw.ID] = hw.LocationPath.GetOrdinalPath()
		}
	}

	return ordinals
}
