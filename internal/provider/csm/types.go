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

//
// Mapping between CANI Inventory Hardware types to CSM Xnames
//

// XnameOrdinal is the mapping between the ordinal withing an xname to a hardware type in a location path
type XnameOrdinal struct {
	HardwareType              hardwaretypes.HardwareType
	HardwarePathLocationIndex int
}
type XnameConverter struct {
	XnameOrdinalMapping []XnameOrdinal
	PropertyMatcher     func(cHardware inventory.Hardware) (bool, error) // IF nil, match all
}

func (xc *XnameConverter) GetHardwareTypePath() hardwaretypes.HardwareTypePath {
	result := hardwaretypes.HardwareTypePath{}
	for _, e := range xc.XnameOrdinalMapping {
		result = append(result, e.HardwareType)
	}
	return result
}

func (xc *XnameConverter) GetOrdinalIndexMapping() (result []int) {
	for _, xnameOrdinal := range xc.XnameOrdinalMapping {
		if xnameOrdinal.HardwarePathLocationIndex < 0 {
			continue
		}

		result = append(result, xnameOrdinal.HardwarePathLocationIndex)
	}

	return result
}

func (xc *XnameConverter) Match(cHardware inventory.Hardware, locationPath inventory.LocationPath) (bool, error) {
	// First check to see if this has a matching hardware type path
	if xc.GetHardwareTypePath().Key() != locationPath.GetHardwareTypePath().Key() {
		return false, nil
	}

	// Next check to see extra properties match
	if xc.PropertyMatcher != nil {
		return xc.PropertyMatcher(cHardware)
	}

	// If we get to this point this is a match!
	return true, nil
}

var enhancedTypeConverters = map[xnametypes.HMSType]XnameConverter{
	xnametypes.Cabinet: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
		},
	},
	xnametypes.CEC: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
			{hardwaretypes.HardwareTypeCabinetEnvironmentalController, 1},
		},
	},
	xnametypes.CabinetPDUController: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
			{hardwaretypes.HardwareTypeCabinetPDUController, 1},
		},
	},
	xnametypes.CabinetPDU: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
			{hardwaretypes.HardwareTypeCabinetPDUController, 1},
			{hardwaretypes.HardwareTypeCabinetPDU, 2},
		},
	},
	xnametypes.Chassis: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
			{hardwaretypes.HardwareTypeChassis, 1},
		},
	},
	xnametypes.ChassisBMC: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
			{hardwaretypes.HardwareTypeChassis, 1},
			{hardwaretypes.HardwareTypeChassis, 2},
		},
	},
	xnametypes.ComputeModule: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
			{hardwaretypes.HardwareTypeChassis, 1},
			{hardwaretypes.HardwareTypeNodeBlade, 2},
		},
	},
	xnametypes.NodeEnclosure: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
			{hardwaretypes.HardwareTypeChassis, 1},
			{hardwaretypes.HardwareTypeNodeBlade, 2},
			{hardwaretypes.HardwareTypeNodeCard, 3},
		},
	},
	xnametypes.NodeBMC: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
			{hardwaretypes.HardwareTypeChassis, 1},
			{hardwaretypes.HardwareTypeNodeBlade, 2},
			{hardwaretypes.HardwareTypeNodeCard, 3},
			{hardwaretypes.HardwareTypeNodeController, -1},
		},
	},
	xnametypes.Node: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
			{hardwaretypes.HardwareTypeChassis, 1},
			{hardwaretypes.HardwareTypeNodeBlade, 2},
			{hardwaretypes.HardwareTypeNodeCard, 3},
			{hardwaretypes.HardwareTypeNode, 4},
		},
	},
	xnametypes.RouterModule: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
			{hardwaretypes.HardwareTypeChassis, 1},
			{hardwaretypes.HardwareTypeHighSpeedSwitchEnclosure, 2},
		},
	},
	xnametypes.RouterBMC: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
			{hardwaretypes.HardwareTypeChassis, 1},
			{hardwaretypes.HardwareTypeHighSpeedSwitchEnclosure, 2},
			{hardwaretypes.HardwareTypeHighSpeedSwitch, -1},
			{hardwaretypes.HardwareTypeHighSpeedSwitchController, 3},
		},
	},

	xnametypes.MgmtSwitch: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
			{hardwaretypes.HardwareTypeChassis, 1},
			{hardwaretypes.HardwareTypeManagementSwitchEnclosure, 2},
			{hardwaretypes.HardwareTypeManagementSwitch, -1},
		},
		PropertyMatcher: func(cHardware inventory.Hardware) (bool, error) {
			// Decode the properties into a struct
			// TODO

			// Check for assigned switch role
			// TODO if LeafBMC switch return true

			// TODO For right now just do not match
			return false, nil
		},
	},

	xnametypes.MgmtHLSwitchEnclosure: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
			{hardwaretypes.HardwareTypeChassis, 1},
			{hardwaretypes.HardwareTypeManagementSwitchEnclosure, 2},
		},
		PropertyMatcher: func(cHardware inventory.Hardware) (bool, error) {
			// Decode the properties into a struct
			// TODO

			// Check for assigned switch role
			// TODO if not LeafBMC switch return true

			// TODO For right now just do not match
			return false, nil
		},
	},
	xnametypes.MgmtHLSwitch: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCabinet, 0},
			{hardwaretypes.HardwareTypeChassis, 1},
			{hardwaretypes.HardwareTypeManagementSwitchEnclosure, 2},
			{hardwaretypes.HardwareTypeManagementSwitch, 3},
		},
		PropertyMatcher: func(cHardware inventory.Hardware) (bool, error) {
			// Decode the properties into a struct
			// TODO

			// Check for assigned switch role
			// TODO if not LeafBMC switch return true

			// TODO For right now just do not match
			return false, nil
		},
	},

	xnametypes.CDU: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCoolingDistributionUnit, 0},
		},
	},
	xnametypes.CDUMgmtSwitch: {
		XnameOrdinalMapping: []XnameOrdinal{
			{hardwaretypes.HardwareTypeCoolingDistributionUnit, 0},
			{hardwaretypes.HardwareTypeManagementSwitchEnclosure, 1},
			{hardwaretypes.HardwareTypeManagementSwitch, -1},
		},
	},
}

func GetXnameTypeConverters() map[xnametypes.HMSType]XnameConverter {
	return enhancedTypeConverters
}

func GetHMSType(cHardware inventory.Hardware, locationPath inventory.LocationPath) (xnametypes.HMSType, error) {
	for hmsType, enhancedTypeConverter := range enhancedTypeConverters {
		match, err := enhancedTypeConverter.Match(cHardware, locationPath)
		if err != nil {
			return xnametypes.HMSTypeInvalid, err
		}

		if match {
			return hmsType, nil
		}
	}

	// This piece of hardware does not have a corresponding xname
	return xnametypes.HMSTypeInvalid, nil
}
