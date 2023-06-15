// MIT License
//
// (C) Copyright 2022 Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package sls

import (
	"fmt"
	"reflect"
	"sort"

	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/rs/zerolog/log"
)

// Hardware present in A that is missing from B
// Set subtract operation
func HardwareSubtract(a, b sls_client.SlsState) ([]sls_client.Hardware, error) {
	var missingHardware []sls_client.Hardware

	// Build up a lookup map for the hardware in set B
	bHardwareMap := map[string]sls_client.Hardware{}
	for _, hardware := range b.Hardware {
		// Verify new hardware
		if _, present := bHardwareMap[hardware.Xname]; present {
			return nil, fmt.Errorf("found duplicate xname %v in set B", hardware.Xname)
		}

		bHardwareMap[hardware.Xname] = hardware
	}

	// Iterate over Set A to identify hardware not present in set B
	for _, hardware := range a.Hardware {
		if _, present := bHardwareMap[hardware.Xname]; present {
			continue
		}

		missingHardware = append(missingHardware, hardware)
	}

	// Sort the slice to make it look nice, and have a deterministic order
	sort.Slice(missingHardware, func(i, j int) bool {
		return missingHardware[i].Xname < missingHardware[j].Xname
	})

	return missingHardware, nil
}

type GenericHardwarePair struct {
	Xname     string
	HardwareA sls_client.Hardware
	HardwareB sls_client.Hardware
}

// Identify hardware
// Note when comparing hardware network information like IP address and subnets are not considered.
// TODO make striping of networking information toggable
func HardwareUnion(a, b sls_client.SlsState) (identicalHardware []sls_client.Hardware, differingContents []GenericHardwarePair, err error) {
	// Build up a lookup map for the hardware in set B
	bHardwareMap := map[string]sls_client.Hardware{}
	for _, hardware := range b.Hardware {
		// Verify new hardware
		if _, present := bHardwareMap[hardware.Xname]; present {
			return nil, nil, fmt.Errorf("found duplicate xname %v in set B", hardware.Xname)
		}

		bHardwareMap[hardware.Xname] = hardware
	}

	// Iterate over Set A to identify hardware present in set B
	for _, hardwareA := range a.Hardware {
		hardwareB, present := bHardwareMap[hardwareA.Xname]
		if !present {
			continue
		}

		hardwarePair := GenericHardwarePair{
			Xname:     hardwareA.Xname,
			HardwareA: hardwareA,
			HardwareB: hardwareB,
		}

		// See if the hardware class between the 2 hardware objects is different
		if hardwareA.Class != hardwareB.Class {
			differingContents = append(differingContents, hardwarePair)

			// Don't bother checking differences in extra properties as there are already differences.
			continue
		}

		// Next check to see if the extra properties between the two hardware objects
		// We are ignoring fields like IPv4 fields and Model during the comparison
		// as that is something that we don't know when generating from the CCJ or CSI

		extraPropertiesA, err := hardwareA.DecodeExtraProperties()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode extra properties on (%s): %w", hardwareA.Xname, err)
		}
		extraPropertiesA = stripIpInformationFromHardware(extraPropertiesA)

		extraPropertiesB, err := hardwareB.DecodeExtraProperties()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode extra properties on (%s): %w", hardwareB.Xname, err)
		}
		extraPropertiesB = stripIpInformationFromHardware(extraPropertiesB)

		if !reflect.DeepEqual(extraPropertiesA, extraPropertiesB) {
			log.Info().Msgf("Hardware A: %v", extraPropertiesA)
			log.Info().Msgf("Hardware B: %v", extraPropertiesB)
			panic("NOOOO")
			differingContents = append(differingContents, hardwarePair)
			continue
		}

		// If we made it here, then these 2 hardware objects must be identical
		identicalHardware = append(identicalHardware, hardwareA)
	}

	// Sort the slices to make it look nice, and have a deterministic order
	sort.Slice(identicalHardware, func(i, j int) bool {
		return identicalHardware[i].Xname < identicalHardware[j].Xname
	})
	sort.Slice(differingContents, func(i, j int) bool {
		return differingContents[i].Xname < differingContents[j].Xname
	})

	return
}

func stripIpInformationFromHardware(extraPropertiesRaw interface{}) interface{} {
	// Helper command to build up the switch
	// grep -R "type HardwareExtraProperties" -R ./pkg/sls-client/ | grep -v HardwareExtraPropertiesCabinetNetworks | awk '{print $2 ":" }' | sort | sed -e 's/^/case sls_client./'
	switch ep := extraPropertiesRaw.(type) {
	case sls_client.HardwareExtraPropertiesBmcNic:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesCabPduNic:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesCabPduPwrConnector:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesCabinet:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""

		ep.Networks = nil
		ep.Model = ""
		// TODO deal with this at somepoint
		// if cabinetKind := csi.CabinetKind(ep.Model); cabinetKind.IsModel() {
		// 	ep.Model = ""
		// }
		return ep
	case sls_client.HardwareExtraPropertiesCduMgmtSwitch:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesChassisBmc:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesCompmod:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesCompmodPowerConnector:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesHsnConnector:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesMgmtHlSwitch:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""

		ep.IP4addr = ""
		ep.IP6addr = ""
		ep.Model = "" // Not guaranteed that the system was installed with information about the switch model.
		return ep
	case sls_client.HardwareExtraPropertiesMgmtSwitch:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""

		ep.IP4addr = ""
		ep.IP6addr = ""
		ep.Model = "" // Not guaranteed that the system was installed with information about the switch model.
		return ep
	case sls_client.HardwareExtraPropertiesMgmtSwitchConnector:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""

		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesNcard:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesNode:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesNodeHsnNic:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesNodeNic:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesRtrBmc:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesRtrBmcNic:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesRtrmod:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	case sls_client.HardwareExtraPropertiesSystem:
		ep.CaniId = ""
		ep.CaniSlsSchemaVersion = ""
		ep.CaniLastModified = ""
		return ep
	}

	return extraPropertiesRaw
}
