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
	htl "github.com/Cray-HPE/cani/pkg/hardware-type-library"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

var typeMapping = map[xnametypes.HMSType]htl.HardwareTypePath{
	xnametypes.Cabinet: {
		htl.HardwareTypeCabinet,
	},
	xnametypes.CEC: {
		htl.HardwareTypeCabinet,
		htl.HardwareTypeCabinetEnvironmentalController,
	},
	xnametypes.CabinetPDUController: {
		htl.HardwareTypeCabinet,
		htl.HardwareTypeCabinetPDUController,
	},
	xnametypes.CabinetPDU: {
		htl.HardwareTypeCabinet,
		htl.HardwareTypeCabinetPDUController,
		htl.HardwareTypePDU,
	},
	xnametypes.Chassis: {
		htl.HardwareTypeCabinet,
		htl.HardwareTypeChassis,
	},
	xnametypes.ChassisBMC: {
		htl.HardwareTypeCabinet,
		htl.HardwareTypeChassis,
		htl.HardwareTypeChassisManagementModule,
	},
	xnametypes.ComputeModule: {
		htl.HardwareTypeCabinet,
		htl.HardwareTypeChassis,
		htl.HardwareTypeNodeBlade,
	},
	xnametypes.NodeBMC: {
		htl.HardwareTypeCabinet,
		htl.HardwareTypeChassis,
		htl.HardwareTypeNodeBlade,
		htl.HardwareTypeNodeCard,
	},
	xnametypes.Node: {
		htl.HardwareTypeCabinet,
		htl.HardwareTypeChassis,
		htl.HardwareTypeNodeBlade,
		htl.HardwareTypeNodeCard,
		htl.HardwareTypeNode,
	},

	xnametypes.RouterModule: {
		htl.HardwareTypeCabinet,
		htl.HardwareTypeChassis,
		htl.HardwareTypeHighSpeedSwitch,
	},
	xnametypes.RouterBMC: {
		htl.HardwareTypeCabinet,
		htl.HardwareTypeChassis,
		htl.HardwareTypeHighSpeedSwitch,
		htl.HardwareTypeHighSpeedSwitchBMC,
	},

	// TODO additional context is required to determine between these two types
	// xnametypes.MgmtSwitch: {
	// 	htl.HardwareTypeCabinet,
	// 	htl.HardwareTypeChassis,
	// 	htl.HardwareTypeManagementSwitch,
	// },
	// xnametypes.MgmtHLSwitch: {
	// 	htl.HardwareTypeCabinet,
	// 	htl.HardwareTypeChassis,
	// 	htl.HardwareTypeManagementSwitch,
	// },

	xnametypes.CDU: {
		htl.HardwareTypeCoolingDistributionUnit,
	},
	xnametypes.CDUMgmtSwitch: {
		htl.HardwareTypeCoolingDistributionUnit,
		htl.HardwareTypeManagementSwitch,
	},
}

func buildHTLtoHMSTypeMap() map[string]xnametypes.HMSType {
	// Build lookup date from Hardware type path to hms-xname type
	// TODO add a check to make sure that there is no overlapping data
	// But need to take in account that MgmtSwitch and MgmtHLSwitch have the same
	// hardware path, but the differance is in the switches roll.
	result := map[string]xnametypes.HMSType{}
	for hmsType, htlTypePath := range typeMapping {
		result[htlTypePath.Key()] = hmsType
	}
	return result
}

var htlToHMSType = buildHTLtoHMSTypeMap()

func GetHMSType(locationPath inventory.LocationPath) xnametypes.HMSType {
	hmsType, exists := htlToHMSType[locationPath.GetHardwareTypePath().Key()]
	if !exists {
		return xnametypes.HMSTypeInvalid
	}

	return hmsType
}

func GetHardwareTypePath(hmsType xnametypes.HMSType) (htl.HardwareTypePath, bool) {
	locationPath, exists := typeMapping[hmsType]
	return locationPath, exists
}
