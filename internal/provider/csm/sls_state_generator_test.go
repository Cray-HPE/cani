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
	"testing"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/stretchr/testify/suite"
)

func buildCabinet(hardwareTypeLibrary hardwaretypes.Library, datastore inventory.Datastore, cabinetDeviceType string, cabinetOrdinal int, chassisPopulation map[int]map[int]string) error {
	system, err := datastore.GetSystemZero()
	if err != nil {
		return err
	}

	// Build cabinet hardware
	cabinetHardwareBuildOut, err := inventory.GenerateDefaultHardwareBuildOut(&hardwareTypeLibrary, cabinetDeviceType, cabinetOrdinal, system)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to build cabinet hardware for device type %s with ordinal %d", cabinetDeviceType, cabinetOrdinal), err)
	}
	for _, hardwareBuildOut := range cabinetHardwareBuildOut {
		hardware := inventory.NewHardwareFromBuildOut(hardwareBuildOut, inventory.HardwareStatusProvisioned)
		if err := datastore.Add(&hardware); err != nil {
			return err
		}
	}

	// Build up blade hardware
	for chassisOrdinal, slots := range chassisPopulation {
		// fmt.Println(slots)

		// Get chassis ID
		chassisLocationPath := inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: cabinetOrdinal},
			{HardwareType: hardwaretypes.Chassis, Ordinal: chassisOrdinal},
		}

		chassis, err := datastore.GetAtLocation(chassisLocationPath)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to find chassis at %v", chassisLocationPath), err)
		}
		// fmt.Println(chassis)

		// Build up hardware for each slot
		for bladeOrdinal, bladeDeviceType := range slots {
			cabinetHardwareBuildOut, err := inventory.GenerateDefaultHardwareBuildOut(&hardwareTypeLibrary, bladeDeviceType, bladeOrdinal, chassis)
			if err != nil {
				return errors.Join(fmt.Errorf("failed to build cabinet hardware for blade type (%s) with ordinal (%d) in chassis (%v)", bladeDeviceType, bladeOrdinal, chassisLocationPath), err)
			}
			// fmt.Println(cabinetHardwareBuildOut)
			for _, hardwareBuildOut := range cabinetHardwareBuildOut {
				hardware := inventory.NewHardwareFromBuildOut(hardwareBuildOut, inventory.HardwareStatusProvisioned)
				if err := datastore.Add(&hardware); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type DetermineHardwareClassSuite struct {
	suite.Suite

	hardwareTypeLibrary *hardwaretypes.Library
	datastore           inventory.Datastore
}

func (suite *DetermineHardwareClassSuite) getHardwareInCabinet(cabinetOrdinal int) []inventory.Hardware {
	inventoryData, err := suite.datastore.List()
	suite.NoError(err)

	hardwareToCheck := []inventory.Hardware{}
	for _, hardware := range inventoryData.Hardware {
		if len(hardware.LocationPath) < 2 {
			continue
		}

		if hardware.LocationPath[1].HardwareType == hardwaretypes.Cabinet && hardware.LocationPath[1].Ordinal == cabinetOrdinal {
			hardwareToCheck = append(hardwareToCheck, hardware)
		}
	}

	return hardwareToCheck
}

func (suite *DetermineHardwareClassSuite) SetupTest() {
	var err error
	suite.hardwareTypeLibrary, err = hardwaretypes.NewEmbeddedLibrary("")
	suite.NoError(err)

	// Generate a inventory of hardware
	suite.datastore, err = inventory.NewDatastoreInMemory(inventory.CSMProvider)
	suite.NoError(err)

	// Build up the different versions of the supported Mountain cabinets
	err = buildCabinet(*suite.hardwareTypeLibrary, suite.datastore, "hpe-ex3000", 1000, map[int]map[int]string{
		1: {
			0: "hpe-crayex-ex420-compute-blade",
		},
	})
	suite.NoError(err)

	err = buildCabinet(*suite.hardwareTypeLibrary, suite.datastore, "hpe-ex4000", 1001, map[int]map[int]string{
		1: {
			0: "hpe-crayex-ex420-compute-blade",
		},
	})
	suite.NoError(err)

	// Build up the different versions of the supported Hill cabinets
	err = buildCabinet(*suite.hardwareTypeLibrary, suite.datastore, "hpe-ex2000", 9000, map[int]map[int]string{
		3: {
			1: "hpe-crayex-ex420-compute-blade",
		},
	})
	suite.NoError(err)

	err = buildCabinet(*suite.hardwareTypeLibrary, suite.datastore, "hpe-ex2500-1-liquid-cooled-chassis", 8000, map[int]map[int]string{
		0: {
			1: "hpe-crayex-ex420-compute-blade",
		},
	})
	suite.NoError(err)

	err = buildCabinet(*suite.hardwareTypeLibrary, suite.datastore, "hpe-ex2500-2-liquid-cooled-chassis", 8001, map[int]map[int]string{
		1: {
			1: "hpe-crayex-ex420-compute-blade",
		},
	})
	suite.NoError(err)

	err = buildCabinet(*suite.hardwareTypeLibrary, suite.datastore, "hpe-ex2500-3-liquid-cooled-chassis", 8002, map[int]map[int]string{
		2: {
			1: "hpe-crayex-ex420-compute-blade",
		},
	})
	suite.NoError(err)

	// Build up the different versions of the supported River cabinets
	err = buildCabinet(*suite.hardwareTypeLibrary, suite.datastore, "hpe-eia-cabinet", 3000, map[int]map[int]string{})
	suite.NoError(err)
}

func (suite *DetermineHardwareClassSuite) TestClassMountainEX3000() {
	//
	// Find hardware in the 1001 cabinet
	//
	hardwareToCheck := suite.getHardwareInCabinet(1000)
	suite.Len(hardwareToCheck, 28) // Verify the expected number of items was found

	//
	// Verify all hardware in cabinet 1001 has class Mountain
	//
	inventoryData, err := suite.datastore.List()
	suite.NoError(err)
	for _, hardware := range hardwareToCheck {
		// Determine the cabinet class
		class, err := DetermineHardwareClass(hardware, inventoryData, *suite.hardwareTypeLibrary)
		suite.NoError(err)
		suite.Equal(sls_client.HardwareClassMountain, class)
	}
}

func (suite *DetermineHardwareClassSuite) TestClassMountainEX4000() {
	//
	// Find hardware in the 1001 cabinet
	//
	hardwareToCheck := suite.getHardwareInCabinet(1001)
	suite.Len(hardwareToCheck, 28) // Verify the expected number of items was found

	//
	// Verify all hardware in cabinet 1001 has class Mountain
	//
	inventoryData, err := suite.datastore.List()
	suite.NoError(err)
	for _, hardware := range hardwareToCheck {
		// Determine the cabinet class
		class, err := DetermineHardwareClass(hardware, inventoryData, *suite.hardwareTypeLibrary)
		suite.NoError(err)
		suite.Equal(sls_client.HardwareClassMountain, class)
	}
}

func (suite *DetermineHardwareClassSuite) TestClassHillEX2000() {
	//
	// Find hardware in the 9000 cabinet
	//
	hardwareToCheck := suite.getHardwareInCabinet(9000)
	suite.Len(hardwareToCheck, 15) // Verify the expected number of items was found

	//
	// Verify all hardware in cabinet 9000 has class Hill
	//
	inventoryData, err := suite.datastore.List()
	suite.NoError(err)
	for _, hardware := range hardwareToCheck {
		// Determine the cabinet class
		class, err := DetermineHardwareClass(hardware, inventoryData, *suite.hardwareTypeLibrary)
		suite.NoError(err)
		suite.Equal(sls_client.HardwareClassHill, class)
	}
}

func (suite *DetermineHardwareClassSuite) TestClassHillEX2500_1Chassis() {
	//
	// Find hardware in the 8000 cabinet
	//
	hardwareToCheck := suite.getHardwareInCabinet(8000)
	suite.Len(hardwareToCheck, 13) // Verify the expected number of items was found

	//
	// Verify all hardware in cabinet 8000 has class Hill
	//
	inventoryData, err := suite.datastore.List()
	suite.NoError(err)
	for _, hardware := range hardwareToCheck {
		// Determine the cabinet class
		class, err := DetermineHardwareClass(hardware, inventoryData, *suite.hardwareTypeLibrary)
		suite.NoError(err)
		suite.Equal(sls_client.HardwareClassHill, class)
	}
}

func (suite *DetermineHardwareClassSuite) TestClassHillEX2500_2Chassis() {
	//
	// Find hardware in the 8001 cabinet
	//
	hardwareToCheck := suite.getHardwareInCabinet(8001)
	suite.Len(hardwareToCheck, 15) // Verify the expected number of items was found

	//
	// Verify all hardware in cabinet 8001 has class Hill
	//
	inventoryData, err := suite.datastore.List()
	suite.NoError(err)
	for _, hardware := range hardwareToCheck {
		// Determine the cabinet class
		class, err := DetermineHardwareClass(hardware, inventoryData, *suite.hardwareTypeLibrary)
		suite.NoError(err)
		suite.Equal(sls_client.HardwareClassHill, class)
	}
}

func (suite *DetermineHardwareClassSuite) TestClassHillEX2500_3Chassis() {
	//
	// Find hardware in the 8002 cabinet
	//
	hardwareToCheck := suite.getHardwareInCabinet(8002)
	suite.Len(hardwareToCheck, 17) // Verify the expected number of items was found

	//
	// Verify all hardware in cabinet 8002 has class Hill
	//
	inventoryData, err := suite.datastore.List()
	suite.NoError(err)
	for _, hardware := range hardwareToCheck {
		// Determine the cabinet class
		class, err := DetermineHardwareClass(hardware, inventoryData, *suite.hardwareTypeLibrary)
		suite.NoError(err)
		suite.Equal(sls_client.HardwareClassHill, class)
	}
}

func (suite *DetermineHardwareClassSuite) TestClassRiver() {
	//
	// Find hardware in the 3000 cabinet
	//
	hardwareToCheck := suite.getHardwareInCabinet(3000)
	suite.Len(hardwareToCheck, 2) // Verify the expected number of items was found

	//
	// Verify all hardware in cabinet 3000 has class River
	//
	inventoryData, err := suite.datastore.List()
	suite.NoError(err)
	for _, hardware := range hardwareToCheck {
		// Determine the cabinet class
		class, err := DetermineHardwareClass(hardware, inventoryData, *suite.hardwareTypeLibrary)
		suite.NoError(err)
		suite.Equal(sls_client.HardwareClassRiver, class)
	}
}

func TestDetermineHardwareClassSuite(t *testing.T) {
	suite.Run(t, new(DetermineHardwareClassSuite))
}
