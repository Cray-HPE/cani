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

package hardwaretypes

import "strings"

type HardwareType string

// TODO should give a description of each of these
const (
	HardwareTypeCabinet                        HardwareType = "Cabinet"
	HardwareTypeChassis                        HardwareType = "Chassis"
	HardwareTypeChassisManagementModule        HardwareType = "ChassisManagementModule"
	HardwareTypeCabinetEnvironmentalController HardwareType = "CabinetEnvironmentalController"
	HardwareTypeNodeBlade                      HardwareType = "NodeBlade"
	HardwareTypeNodeCard                       HardwareType = "NodeCard"       // TODO Change to enclosure?
	HardwareTypeNodeController                 HardwareType = "NodeController" // A Node BMC is a child of a node card
	HardwareTypeNode                           HardwareType = "Node"
	HardwareTypeManagementSwitchEnclosure      HardwareType = "ManagementSwitchEnclosure"
	HardwareTypeManagementSwitch               HardwareType = "ManagementSwitch"
	HardwareTypeManagementSwitchController     HardwareType = "ManagementSwitchController"
	HardwareTypeHighSpeedSwitchEnclosure       HardwareType = "HighSpeedSwitchEnclosure"
	HardwareTypeHighSpeedSwitch                HardwareType = "HighSpeedSwitch"
	HardwareTypeHighSpeedSwitchController      HardwareType = "HighSpeedSwitchController"
	HardwareTypeCabinetPDUController           HardwareType = "CabinetPDUController"
	HardwareTypeCabinetPDU                     HardwareType = "CabinetPDU"
	HardwareTypeCoolingDistributionUnit        HardwareType = "CoolingDistributionUnit"

	// TODO NEED TO COMEBACK ON IF SWITCHES NEED TO BE SEPARATE FOR HSN AND MANAGEMENT
)

type HardwareTypePath []HardwareType

func (htp HardwareTypePath) Key() string {
	elements := []string{}
	for _, element := range htp {
		elements = append(elements, string(element))
	}
	return strings.Join(elements, ".")
}

type Airflow string

const (
	AirflowFrontToRear Airflow = "front-to-rear"
	AirflowRearToFront Airflow = "rear-to-front"
	AirflowLeftToRight Airflow = "left-to-right"
	AirflowRightToLeft Airflow = "right-to-left"
	AirflowSideToRear  Airflow = "side-to-rear"
	AirflowPassive     Airflow = "passive"
)

type WeightUnit string

const (
	WeightUnitKiloGram WeightUnit = "kg"
	WeightUnitGram     WeightUnit = "g"
	WeightUnitPound    WeightUnit = "lb"
	WeightUnitOunce    WeightUnit = "oz"
)

type SubDeviceRole string

const (
	SubDeviceRoleParent SubDeviceRole = "parent"
	SubDeviceRoleChild  SubDeviceRole = "child"
)

type DeviceType struct {
	Manufacturer string       `yaml:"manufacturer"`
	Model        string       `yaml:"model"`
	HardwareType HardwareType `yaml:"hardware-type"`
	Slug         string       `yaml:"slug"`

	PartNumber  *string     `yaml:"part_number"`
	UHeight     *float64    `yaml:"u_height"`
	IsFullDepth *bool       `yaml:"is_full_depth"`
	Weight      *float64    `yaml:"weight"`
	WeightUnit  *WeightUnit `yaml:"weight_unit"`

	FrontImage bool `yaml:"front_image"`
	RearImage  bool `yaml:"rear_image"`

	SubDeviceRole SubDeviceRole `yaml:"subdevice_role"`

	// TODO
	// ConsolePorts       []ConsolePort       `yaml:"console-ports"`
	// ConsoleServerPorts []ConsoleServerPort `yaml:"console-server-ports"`
	// PowerPowers        []PowerPower        `yaml:"power-ports"`
	// PowerOutlets       []PowerOutlets      `yaml:"power-outlets"`

	DeviceBays []DeviceBay `yaml:"device-bays"`
}

type DeviceBay struct {
	Name    string           `yaml:"name"`
	Allowed *AllowedHardware `yaml:"allowed"`
	Default *DefaultHardware `yaml:"default"`
}

type AllowedHardware struct {
	HardwareTypes []HardwareType `yaml:"hardware-type"`
	Slug          []string       `yaml:"slug"`
}

type DefaultHardware struct {
	Slug string `yaml:"slug"`
}

// TODO

// type ConsolePort struct {
// }

// type ConsoleServerPort struct {
// }

// type PowerPower struct {
// }
// type PowerOutlets struct {
// }

// type Interface struct {
// }

// type FrontPort struct {
// }

// type RearPort struct {
// }

// type ModuleBay struct {
// }

// type InventoryItem struct {
// }
