// MIT License
//
// (C) Copyright [2023] Hewlett Packard Enterprise Development LP
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
//
//go:generate go run ./generator

package csm

import (
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

var typeMapping = map[xnametypes.HMSType]hardwaretypes.HardwareTypePath{
	xnametypes.Cabinet: {
		hardwaretypes.HardwareTypeCabinet,
	},
	xnametypes.CEC: {
		hardwaretypes.HardwareTypeCabinet,
		hardwaretypes.HardwareTypeCabinetEnvironmentalController,
	},
	xnametypes.CabinetPDUController: {
		hardwaretypes.HardwareTypeCabinet,
		hardwaretypes.HardwareTypeCabinetPDUController,
	},
	xnametypes.CabinetPDU: {
		hardwaretypes.HardwareTypeCabinet,
		hardwaretypes.HardwareTypeCabinetPDUController,
		hardwaretypes.HardwareTypePDU,
	},
	xnametypes.Chassis: {
		hardwaretypes.HardwareTypeCabinet,
		hardwaretypes.HardwareTypeChassis,
	},
	xnametypes.ChassisBMC: {
		hardwaretypes.HardwareTypeCabinet,
		hardwaretypes.HardwareTypeChassis,
		hardwaretypes.HardwareTypeChassisManagementModule,
	},
	xnametypes.ComputeModule: {
		hardwaretypes.HardwareTypeCabinet,
		hardwaretypes.HardwareTypeChassis,
		hardwaretypes.HardwareTypeNodeBlade,
	},
	xnametypes.NodeBMC: {
		hardwaretypes.HardwareTypeCabinet,
		hardwaretypes.HardwareTypeChassis,
		hardwaretypes.HardwareTypeNodeBlade,
		hardwaretypes.HardwareTypeNodeCard,
	},
	xnametypes.Node: {
		hardwaretypes.HardwareTypeCabinet,
		hardwaretypes.HardwareTypeChassis,
		hardwaretypes.HardwareTypeNodeBlade,
		hardwaretypes.HardwareTypeNodeCard,
		hardwaretypes.HardwareTypeNode,
	},

	xnametypes.RouterModule: {
		hardwaretypes.HardwareTypeCabinet,
		hardwaretypes.HardwareTypeChassis,
		hardwaretypes.HardwareTypeHighSpeedSwitch,
	},
	xnametypes.RouterBMC: {
		hardwaretypes.HardwareTypeCabinet,
		hardwaretypes.HardwareTypeChassis,
		hardwaretypes.HardwareTypeHighSpeedSwitch,
		hardwaretypes.HardwareTypeHighSpeedSwitchBMC,
	},

	// TODO additional context is required to determine between these two types
	// xnametypes.MgmtSwitch: {
	// 	hardwaretypes.HardwareTypeCabinet,
	// 	hardwaretypes.HardwareTypeChassis,
	// 	hardwaretypes.HardwareTypeManagementSwitch,
	// },
	// xnametypes.MgmtHLSwitch: {
	// 	hardwaretypes.HardwareTypeCabinet,
	// 	hardwaretypes.HardwareTypeChassis,
	// 	hardwaretypes.HardwareTypeManagementSwitch,
	// },

	xnametypes.CDU: {
		hardwaretypes.HardwareTypeCoolingDistributionUnit,
	},
	xnametypes.CDUMgmtSwitch: {
		hardwaretypes.HardwareTypeCoolingDistributionUnit,
		hardwaretypes.HardwareTypeManagementSwitch,
	},
}

func buildhardwaretypestoHMSTypeMap() map[string]xnametypes.HMSType {
	// Build lookup date from Hardware type path to hms-xname type
	// TODO add a check to make sure that there is no overlapping data
	// But need to take in account that MgmtSwitch and MgmtHLSwitch have the same
	// hardware path, but the differance is in the switches roll.
	result := map[string]xnametypes.HMSType{}
	for hmsType, hardwaretypesTypePath := range typeMapping {
		result[hardwaretypesTypePath.Key()] = hmsType
	}
	return result
}

var hardwaretypesToHMSType = buildhardwaretypestoHMSTypeMap()

func GetHMSType(locationPath inventory.LocationPath) xnametypes.HMSType {
	hmsType, exists := hardwaretypesToHMSType[locationPath.GetHardwareTypePath().Key()]
	if !exists {
		return xnametypes.HMSTypeInvalid
	}

	return hmsType
}

func GetHardwareTypePath(hmsType xnametypes.HMSType) (hardwaretypes.HardwareTypePath, bool) {
	locationPath, exists := typeMapping[hmsType]
	return locationPath, exists
}
