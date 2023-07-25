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
	"time"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/csm/sls"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func DetermineHardwareClass(hardware inventory.Hardware, data inventory.Inventory, hardwareTypeLibrary hardwaretypes.Library) (sls_client.HardwareClass, error) {
	currentHardwareID := hardware.ID
	for currentHardwareID != uuid.Nil {
		currentHardware, exists := data.Hardware[currentHardwareID]
		if !exists {
			return "", errors.Join(
				fmt.Errorf("unable to find ancestor (%s) of (%s)", currentHardwareID, hardware.ID),
			)
		}

		deviceType, exists := hardwareTypeLibrary.DeviceTypes[currentHardware.DeviceTypeSlug]
		if !exists {
			return "", errors.Join(
				fmt.Errorf("unable to find device type (%s) for (%s)", currentHardware.DeviceTypeSlug, currentHardwareID),
			)
		}

		if deviceType.ProviderDefaults != nil && deviceType.ProviderDefaults.CSM != nil && deviceType.ProviderDefaults.CSM.Class != nil {
			classRaw := *deviceType.ProviderDefaults.CSM.Class
			switch classRaw {
			case "River":
				return sls_client.HardwareClassRiver, nil
			case "Mountain":
				return sls_client.HardwareClassMountain, nil
			case "Hill":
				return sls_client.HardwareClassHill, nil
			default:
				return "", fmt.Errorf("encountered unknown CSM hardware class (%s)", classRaw)
			}
		}

		// Go the parent node next
		currentHardwareID = currentHardware.Parent
	}

	return "", fmt.Errorf("unable to determine CSM Class of (%s)", hardware.ID)
}

func DetermineHardwareClassFromSlug(deviceTypeSlug string, hardwareTypeLibrary hardwaretypes.Library) (sls_client.HardwareClass, error) {
	deviceType, exists := hardwareTypeLibrary.DeviceTypes[deviceTypeSlug]
	if !exists {
		return "", errors.Join(
			fmt.Errorf("unable to find device type (%s)", deviceTypeSlug),
		)
	}

	if deviceType.ProviderDefaults != nil && deviceType.ProviderDefaults.CSM != nil && deviceType.ProviderDefaults.CSM.Class != nil {
		classRaw := *deviceType.ProviderDefaults.CSM.Class
		switch classRaw {
		case "River":
			return sls_client.HardwareClassRiver, nil
		case "Mountain":
			return sls_client.HardwareClassMountain, nil
		case "Hill":
			return sls_client.HardwareClassHill, nil
		default:
			return "", fmt.Errorf("encountered unknown CSM hardware class (%s)", classRaw)
		}
	}

	return "", fmt.Errorf("unable to determine CSM Class of (%s)", deviceTypeSlug)
}

func DetermineStartingOrdinalFromSlug(deviceTypeSlug string, hardwareTypeLibrary hardwaretypes.Library) (int, error) {
	deviceType, exists := hardwareTypeLibrary.DeviceTypes[deviceTypeSlug]
	if !exists {
		return 0, errors.Join(
			fmt.Errorf("unable to find device type (%s)", deviceTypeSlug),
		)
	}

	if deviceType.ProviderDefaults != nil && deviceType.ProviderDefaults.CSM != nil {
		return deviceType.ProviderDefaults.CSM.Ordinal, nil
	}

	return 0, fmt.Errorf("unable to determine CSM starting ordinal of (%s) %+v", deviceTypeSlug, deviceType.ProviderDefaults.CSM)
}

func DetermineStartingVlanFromSlug(deviceTypeSlug string, hardwareTypeLibrary hardwaretypes.Library) (int, error) {
	deviceType, exists := hardwareTypeLibrary.DeviceTypes[deviceTypeSlug]
	if !exists {
		return 0, errors.Join(
			fmt.Errorf("unable to find device type (%s)", deviceTypeSlug),
		)
	}

	if deviceType.ProviderDefaults != nil && deviceType.ProviderDefaults.CSM != nil {
		return deviceType.ProviderDefaults.CSM.StartingHmnVlan, nil
	}

	return 0, fmt.Errorf("unable to determine CSM starting VLAN of (%s) %+v", deviceTypeSlug, deviceType.ProviderDefaults.CSM)
}

func DetermineEndingVlanFromSlug(deviceTypeSlug string, hardwareTypeLibrary hardwaretypes.Library) (int, error) {
	deviceType, exists := hardwareTypeLibrary.DeviceTypes[deviceTypeSlug]
	if !exists {
		return 0, errors.Join(
			fmt.Errorf("unable to find device type (%s)", deviceTypeSlug),
		)
	}

	if deviceType.ProviderDefaults != nil && deviceType.ProviderDefaults.CSM != nil {
		return deviceType.ProviderDefaults.CSM.EndingHmnVlan, nil
	}

	return 0, fmt.Errorf("unable to determine CSM ending VLAN of (%s) %+v", deviceTypeSlug, deviceType.ProviderDefaults.CSM)
}

func BuildExpectedHardwareState(hardwareTypeLibrary hardwaretypes.Library, datastore inventory.Datastore, slsNetworks map[string]sls_client.Network) (sls_client.SlsState, map[string]inventory.Hardware, error) {
	// Retrieve the CANI inventory data
	data, err := datastore.List()
	if err != nil {
		return sls_client.SlsState{}, nil, errors.Join(
			fmt.Errorf("failed to list hardware from the datastore"),
			err,
		)
	}

	// This is a lookup map that keeps track of what CANI hardware object generated a
	// piece of SLS hardware
	hardwareMapping := map[string]inventory.Hardware{}

	// Iterate over the CANI inventory data to build SLS data
	allHardware := map[string]sls_client.Hardware{}
	for _, cHardware := range data.Hardware {
		// Skip systems
		if cHardware.Type == hardwaretypes.System {
			continue
		}

		//
		// Build the SLS hardware representation
		//
		log.Debug().Any("cHardware", cHardware).Msg("Processing")
		locationPath, err := datastore.GetLocation(cHardware)
		if err != nil {
			return sls_client.SlsState{}, nil, errors.Join(
				fmt.Errorf("failed to get location of hardware (%s) from the datastore", cHardware.ID),
				err,
			)
		}

		slsClass, err := DetermineHardwareClass(cHardware, data, hardwareTypeLibrary)
		if err != nil {
			return sls_client.SlsState{}, nil, errors.Join(
				fmt.Errorf("failed to determine SLS class of hardware (%s)", cHardware.ID),
				err,
			)
		}

		hardware, err := BuildSLSHardware(cHardware, locationPath, slsClass, slsNetworks)
		// if err != nil && ignoreUnknownCANUHardwareArchitectures && strings.Contains(err.Error(), "unknown architecture type") {
		// 	log.Printf("WARNING %s", err.Error())
		// } else if err != nil {
		if err != nil {
			return sls_client.SlsState{}, nil, err
		}

		log.Debug().Any("hardware", hardware).Msg("Generated SLS hardware")

		// Ignore empty hardware
		if hardware.Xname == "" {
			continue
		}

		// Update CANI->SLS hardware mapping
		hardwareMapping[hardware.Xname] = cHardware

		// Verify cabinet exists (ignore CDUs)
		// TODO
		// if strings.HasPrefix(hardware.Xname, "x") {
		// 	cabinetXname, err := csi.CabinetForXname(hardware.Xname)
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	if !cabinetLookup.CabinetExists(cabinetXname) {
		// 		err := fmt.Errorf("unknown cabinet (%s)", cabinetXname)
		// 		panic(err)
		// 	}
		// }

		// Verify new hardware
		if _, present := allHardware[hardware.Xname]; present {
			err := fmt.Errorf("found duplicate xname %v", hardware.Xname)
			panic(err)
		}

		allHardware[hardware.Xname] = hardware

		//
		// Build up derived hardware
		//
		// TODO This is probably not needed as the CANI Inventory will have all that we need
		// if hardware.TypeString == xnametypes.ChassisBMC {
		// 	allHardware[hardware.Xname] = sls_client.NewGenericHardware(hardware.Parent, hardware.Class, nil)
		// }

		//
		// Build the MgmtSwitchConnector for the hardware
		//
		mgmtSwtichConnector, err := BuildSLSMgmtSwitchConnector(hardware, cHardware)
		if err != nil {
			panic(err)
		}

		// Ignore empty mgmtSwtichConnectors
		if mgmtSwtichConnector.Xname == "" {
			continue
		}

		if _, present := allHardware[mgmtSwtichConnector.Xname]; present {
			err := fmt.Errorf("found duplicate xname %v", mgmtSwtichConnector.Xname)
			panic(err)
		}

		allHardware[mgmtSwtichConnector.Xname] = mgmtSwtichConnector

	}

	// Build up and the SLS state
	return sls_client.SlsState{
		Hardware: allHardware,
	}, hardwareMapping, nil
}

func BuildSLSHardware(cHardware inventory.Hardware, locationPath inventory.LocationPath, class sls_client.HardwareClass, slsNetworks map[string]sls_client.Network) (sls_client.Hardware, error) {
	log.Debug().Stringer("locationPath", locationPath).Msg("LocationPath")

	// Get the physical location for the hardware
	xname, err := BuildXname(cHardware, locationPath)
	log.Debug().Any("xname", xname).Err(err).Msg("Build xname")
	if err != nil {
		return sls_client.Hardware{}, err
	} else if xname == nil {
		// This means that this piece of the hardware inventory can't be represented in SLS due to no xname, so just skip it
		return sls_client.Hardware{}, nil
	}

	// Get the class of the piece of hardware
	// Generally this will match the class of the containing cabinet, the exception is river hardware within a EX2500 cabinet.
	// TODO
	var extraProperties interface{}

	switch cHardware.Type {
	case hardwaretypes.Cabinet:
		var cabinetExtraProperties sls_client.HardwareExtraPropertiesCabinet

		//
		// Apply CANI Metadata
		//
		cabinetExtraProperties.CaniId = cHardware.ID.String()
		cabinetExtraProperties.CaniSlsSchemaVersion = "v1alpha1"
		cabinetExtraProperties.CaniLastModified = time.Now().UTC().String()

		//
		// Build cabinet metadata
		//
		cabinetExtraProperties.Networks = map[string]map[string]sls_client.HardwareExtraPropertiesCabinetNetworks{
			"cn": map[string]sls_client.HardwareExtraPropertiesCabinetNetworks{},
		}

		// Determine which SLS network contains the cabinet subnet
		hmnNetworkName := "HMN_MTN"
		nmnNetworkName := "NMN_MTN"
		if class == sls_client.HardwareClassRiver {
			hmnNetworkName = "HMN_RVR"
			nmnNetworkName = "NMN_RVR"
		}

		// Determine the subnet name, should be the same between the HMN_* and NMN_* networks
		subnetName := fmt.Sprintf("cabinet_%d", *cHardware.LocationOrdinal)

		// Find cabinet HMN subnet
		hmnNetwork, exists := slsNetworks[hmnNetworkName]
		if !exists {
			return sls_client.Hardware{}, fmt.Errorf("SLS Network (%s) does not exist", hmnNetworkName)
		}
		for _, subnet := range hmnNetwork.ExtraProperties.Subnets {
			if subnet.Name == subnetName {
				cabinetExtraProperties.Networks["cn"]["HMN"] = sls_client.HardwareExtraPropertiesCabinetNetworks{
					CIDR:    subnet.CIDR,
					Gateway: subnet.Gateway,
					VLan:    subnet.VlanID,
				}
			}
		}

		// Find cabinet NMN subnet
		nmnNetwork, exists := slsNetworks[nmnNetworkName]
		if !exists {
			return sls_client.Hardware{}, fmt.Errorf("SLS Network (%s) does not exist", nmnNetworkName)
		}
		for _, subnet := range nmnNetwork.ExtraProperties.Subnets {
			if subnet.Name == subnetName {
				cabinetExtraProperties.Networks["cn"]["NMN"] = sls_client.HardwareExtraPropertiesCabinetNetworks{
					CIDR:    subnet.CIDR,
					Gateway: subnet.Gateway,
					VLan:    subnet.VlanID,
				}
			}
		}

		if class == sls_client.HardwareClassRiver {
			// If this is a river cabinet, we need to make a entry for ncn network.
			cabinetExtraProperties.Networks["ncn"] = cabinetExtraProperties.Networks["cn"]
		}

		extraProperties = cabinetExtraProperties
	case hardwaretypes.Chassis:
		var chassisExtraProperties sls_client.HardwareExtraPropertiesChassis

		// Apply CANI Metadata
		chassisExtraProperties.CaniId = cHardware.ID.String()
		chassisExtraProperties.CaniSlsSchemaVersion = "v1alpha1"
		chassisExtraProperties.CaniLastModified = time.Now().UTC().String()

		extraProperties = chassisExtraProperties
	case hardwaretypes.ChassisManagementModule:
		var cmmExtraProperties sls_client.HardwareExtraPropertiesChassisBmc

		// Apply CANI Metadata
		cmmExtraProperties.CaniId = cHardware.ID.String()
		cmmExtraProperties.CaniSlsSchemaVersion = "v1alpha1"
		cmmExtraProperties.CaniLastModified = time.Now().UTC().String()

		extraProperties = cmmExtraProperties
	case hardwaretypes.NodeBlade:
		// Not represented in SLS
		return sls_client.Hardware{}, nil
	case hardwaretypes.NodeCard:
		// Not represented in SLS
		return sls_client.Hardware{}, nil
	case hardwaretypes.NodeController:
		// Not represented in SLS
		return sls_client.Hardware{}, nil
	case hardwaretypes.Node:
		metadata, err := GetProviderMetadataT[NodeMetadata](cHardware)
		if err != nil {
			return sls_client.Hardware{}, errors.Join(
				fmt.Errorf("failed to get provider metadata from hardware (%s)", cHardware.ID),
				err,
			)
		}

		var nodeExtraProperties sls_client.HardwareExtraPropertiesNode
		// Apply CANI Metadata
		nodeExtraProperties.CaniId = cHardware.ID.String()
		nodeExtraProperties.CaniSlsSchemaVersion = "v1alpha1"
		nodeExtraProperties.CaniLastModified = time.Now().UTC().String()

		// Logical metadata
		if metadata != nil {

			// In order to properly populate SLS several bits of information are required.
			// This information should have been collected when hardware was added to the inventory
			// - Role
			// - SubRole
			// - NID
			// - Alias/Common Name
			if metadata.Role != nil {
				nodeExtraProperties.Role = *metadata.Role
			}
			if metadata.SubRole != nil {
				nodeExtraProperties.Role = *metadata.SubRole
			}
			if metadata.Nid != nil {
				nodeExtraProperties.NID = int32(*metadata.Nid)
			}
			if metadata.Alias != nil {
				nodeExtraProperties.Aliases = metadata.Alias
			}

			log.Info().Any("nodeEp", nodeExtraProperties).Msgf("Generated Extra Properties for %s", xname.String())
		}
		extraProperties = nodeExtraProperties
	default:
		log.Warn().Msgf("Do not known how to handle %s", xname.String())
		return sls_client.Hardware{}, nil
	}

	return sls.NewHardware(xname, class, extraProperties), nil
}

func BuildSLSMgmtSwitchConnector(hardware sls_client.Hardware, cHardware inventory.Hardware) (sls_client.Hardware, error) {
	return sls_client.Hardware{}, nil
}
