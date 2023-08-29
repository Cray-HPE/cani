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
	"fmt"
	"testing"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/stretchr/testify/suite"
)

type FromXnameSuite struct {
	suite.Suite
}

func (suite *FromXnameSuite) TestSystem() {
	lp, err := FromXname(xnames.System{})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestSystemPointer() {
	lp, err := FromXname(&xnames.System{})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestCabinet() {
	lp, err := FromXname(xnames.Cabinet{
		Cabinet: 1000,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestCabinetPointer() {
	lp, err := FromXname(&xnames.Cabinet{
		Cabinet: 1000,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestChassis() {
	lp, err := FromXname(xnames.Chassis{
		Cabinet: 1000,
		Chassis: 2,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
		{HardwareType: hardwaretypes.Chassis, Ordinal: 2},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestChassisPointer() {
	lp, err := FromXname(&xnames.Chassis{
		Cabinet: 1000,
		Chassis: 2,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
		{HardwareType: hardwaretypes.Chassis, Ordinal: 2},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestChassisBMC() {
	lp, err := FromXname(xnames.ChassisBMC{
		Cabinet:    1000,
		Chassis:    2,
		ChassisBMC: 0,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
		{HardwareType: hardwaretypes.Chassis, Ordinal: 2},
		{HardwareType: hardwaretypes.ChassisManagementModule, Ordinal: 0},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestChassisBMCPointer() {
	lp, err := FromXname(&xnames.ChassisBMC{
		Cabinet:    1000,
		Chassis:    2,
		ChassisBMC: 0,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
		{HardwareType: hardwaretypes.Chassis, Ordinal: 2},
		{HardwareType: hardwaretypes.ChassisManagementModule, Ordinal: 0},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestComputeModule() {
	lp, err := FromXname(xnames.ComputeModule{
		Cabinet:       1000,
		Chassis:       2,
		ComputeModule: 7,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
		{HardwareType: hardwaretypes.Chassis, Ordinal: 2},
		{HardwareType: hardwaretypes.NodeBlade, Ordinal: 7},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestComputeModulePointer() {
	lp, err := FromXname(&xnames.ComputeModule{
		Cabinet:       1000,
		Chassis:       2,
		ComputeModule: 7,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
		{HardwareType: hardwaretypes.Chassis, Ordinal: 2},
		{HardwareType: hardwaretypes.NodeBlade, Ordinal: 7},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestNodeBMC() {
	lp, err := FromXname(xnames.NodeBMC{
		Cabinet:       1000,
		Chassis:       2,
		ComputeModule: 7,
		NodeBMC:       1,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
		{HardwareType: hardwaretypes.Chassis, Ordinal: 2},
		{HardwareType: hardwaretypes.NodeBlade, Ordinal: 7},
		{HardwareType: hardwaretypes.NodeCard, Ordinal: 1},
		{HardwareType: hardwaretypes.NodeController, Ordinal: 0},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestNodeBMCPointer() {
	lp, err := FromXname(&xnames.NodeBMC{
		Cabinet:       1000,
		Chassis:       2,
		ComputeModule: 7,
		NodeBMC:       1,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
		{HardwareType: hardwaretypes.Chassis, Ordinal: 2},
		{HardwareType: hardwaretypes.NodeBlade, Ordinal: 7},
		{HardwareType: hardwaretypes.NodeCard, Ordinal: 1},
		{HardwareType: hardwaretypes.NodeController, Ordinal: 0},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestNode() {
	lp, err := FromXname(xnames.Node{
		Cabinet:       1000,
		Chassis:       2,
		ComputeModule: 7,
		NodeBMC:       1,
		Node:          3,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
		{HardwareType: hardwaretypes.Chassis, Ordinal: 2},
		{HardwareType: hardwaretypes.NodeBlade, Ordinal: 7},
		{HardwareType: hardwaretypes.NodeCard, Ordinal: 1},
		{HardwareType: hardwaretypes.Node, Ordinal: 3},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestNodePointer() {
	lp, err := FromXname(&xnames.Node{
		Cabinet:       1000,
		Chassis:       2,
		ComputeModule: 7,
		NodeBMC:       1,
		Node:          3,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
		{HardwareType: hardwaretypes.Chassis, Ordinal: 2},
		{HardwareType: hardwaretypes.NodeBlade, Ordinal: 7},
		{HardwareType: hardwaretypes.NodeCard, Ordinal: 1},
		{HardwareType: hardwaretypes.Node, Ordinal: 3},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestRouterModule() {
	lp, err := FromXname(xnames.RouterModule{
		Cabinet:      1000,
		Chassis:      2,
		RouterModule: 7,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
		{HardwareType: hardwaretypes.Chassis, Ordinal: 2},
		{HardwareType: hardwaretypes.HighSpeedSwitchEnclosure, Ordinal: 7},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestRouterModulePointer() {
	lp, err := FromXname(&xnames.RouterModule{
		Cabinet:      1000,
		Chassis:      2,
		RouterModule: 7,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
		{HardwareType: hardwaretypes.Chassis, Ordinal: 2},
		{HardwareType: hardwaretypes.HighSpeedSwitchEnclosure, Ordinal: 7},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestRouterBMC() {
	lp, err := FromXname(xnames.RouterBMC{
		Cabinet:      1000,
		Chassis:      2,
		RouterModule: 7,
		RouterBMC:    0,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
		{HardwareType: hardwaretypes.Chassis, Ordinal: 2},
		{HardwareType: hardwaretypes.HighSpeedSwitchEnclosure, Ordinal: 7},
		{HardwareType: hardwaretypes.HighSpeedSwitchController, Ordinal: 0},
	}

	suite.Equal(expectedLP, lp)
}

func (suite *FromXnameSuite) TestRouterBMCPointer() {
	lp, err := FromXname(&xnames.RouterBMC{
		Cabinet:      1000,
		Chassis:      2,
		RouterModule: 7,
		RouterBMC:    0,
	})
	suite.NoError(err)

	expectedLP := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: 1000},
		{HardwareType: hardwaretypes.Chassis, Ordinal: 2},
		{HardwareType: hardwaretypes.HighSpeedSwitchEnclosure, Ordinal: 7},
		{HardwareType: hardwaretypes.HighSpeedSwitchController, Ordinal: 0},
	}

	suite.Equal(expectedLP, lp)
}

func TestFromXnameSuite(t *testing.T) {
	suite.Run(t, new(FromXnameSuite))
}

type IdentifyDeviceSlugTestSuite struct {
	suite.Suite

	csm *CSM
}

func (suite *IdentifyDeviceSlugTestSuite) SetupSuite() {
	hardwareTypeLibrary, err := hardwaretypes.NewEmbeddedLibrary("")
	suite.NoError(err)

	suite.csm = &CSM{
		hardwareLibrary: hardwareTypeLibrary,
	}
}

func (suite *IdentifyDeviceSlugTestSuite) TestWindom() {
	var testData = []struct {
		Manufacturer string
		Model        string
		PartNumber   string
		SKU          string
	}{
		{"HPE", "WNC", "101920703.D", ""},
		{"HPE", "WNC", "101920703.F", ""},
		{"HPE", "WNC", "101920704.A", ""},
		{"HPE", "WNC", "101920704.B", ""},
		{"HPE", "WindomNodeCard", "101920702.C", ""},
		{"HPE", "WindomNodeCard", "101920703.B", ""},
		{"HPE", "WindomNodeCard", "101920703.D", ""},
	}

	for _, test := range testData {
		msgString := fmt.Sprintf("Failed to identify device slug with values: Manufacturer %s, Model %s, PartNumber %s, SKU %s", test.Manufacturer, test.Model, test.PartNumber, test.SKU)

		deviceSlug, err := suite.csm.identifyDeviceSlug(test.Manufacturer, test.Model, test.PartNumber)
		suite.NoError(err, msgString)

		suite.Equal("hpe-crayex-ex425-compute-blade", deviceSlug, msgString)
	}
}

func (suite *IdentifyDeviceSlugTestSuite) TestCastle() {
	var testData = []struct {
		Manufacturer string
		Model        string
		PartNumber   string
		SKU          string
	}{
		{"HPE", "CNC", "P40731-002.B", ""},
		{"HPE", "CNC", "P40731-003.A", ""},
	}

	for _, test := range testData {
		msgString := fmt.Sprintf("Failed to identify device slug with values: Manufacturer %s, Model %s, PartNumber %s, SKU %s", test.Manufacturer, test.Model, test.PartNumber, test.SKU)

		deviceSlug, err := suite.csm.identifyDeviceSlug(test.Manufacturer, test.Model, test.PartNumber)
		suite.NoError(err, msgString)

		suite.Equal("hpe-crayex-ex420-compute-blade", deviceSlug, msgString)
	}
}

func (suite *IdentifyDeviceSlugTestSuite) TestGrizzlyPeak() {
	var testData = []struct {
		Manufacturer string
		Model        string
		PartNumber   string
		SKU          string
	}{
		{"HPE", "GrizzlyPkNodeCard", "102209502.B", ""},
	}

	for _, test := range testData {
		msgString := fmt.Sprintf("Failed to identify device slug with values: Manufacturer %s, Model %s, PartNumber %s, SKU %s", test.Manufacturer, test.Model, test.PartNumber, test.SKU)

		deviceSlug, err := suite.csm.identifyDeviceSlug(test.Manufacturer, test.Model, test.PartNumber)
		suite.NoError(err, msgString)

		suite.Equal("hpe-crayex-ex235n-compute-blade", deviceSlug, msgString)
	}
}

func (suite *IdentifyDeviceSlugTestSuite) TestBardPeak() {
	var testData = []struct {
		Manufacturer string
		Model        string
		PartNumber   string
		SKU          string
	}{
		{"HPE", "BardPeakNC", "P37085-001.A", ""},
		{"HPE", "BardPeakNC", "P37085-003.A", ""},
	}

	for _, test := range testData {
		msgString := fmt.Sprintf("Failed to identify device slug with values: Manufacturer %s, Model %s, PartNumber %s, SKU %s", test.Manufacturer, test.Model, test.PartNumber, test.SKU)

		deviceSlug, err := suite.csm.identifyDeviceSlug(test.Manufacturer, test.Model, test.PartNumber)
		suite.NoError(err, msgString)

		suite.Equal("hpe-crayex-ex235a-compute-blade", deviceSlug, msgString)
	}
}

func (suite *IdentifyDeviceSlugTestSuite) TestAntero() {
	var testData = []struct {
		Manufacturer string
		Model        string
		PartNumber   string
		SKU          string
	}{
		{"HPE", "HSSMEZZCARD", "P47016-002.1", ""},
		{"HPE", "HSSMEZZCARD", "P47016-003.A", ""},
	}

	for _, test := range testData {
		msgString := fmt.Sprintf("Failed to identify device slug with values: Manufacturer %s, Model %s, PartNumber %s, SKU %s", test.Manufacturer, test.Model, test.PartNumber, test.SKU)

		deviceSlug, err := suite.csm.identifyDeviceSlug(test.Manufacturer, test.Model, test.PartNumber)
		suite.NoError(err, msgString)

		suite.Equal("hpe-crayex-ex4252-compute-blade", deviceSlug, msgString)
	}
}

func (suite *IdentifyDeviceSlugTestSuite) TestBlancaPeak() {
	var testData = []struct {
		Manufacturer string
		Model        string
		PartNumber   string
		SKU          string
	}{
		{"HPE", "BlancaPeakNC", "P52291-002.B", ""},
		{"HPE", "BlancaPeakNC", "P52291-003.C", ""},
	}

	for _, test := range testData {
		msgString := fmt.Sprintf("Failed to identify device slug with values: Manufacturer %s, Model %s, PartNumber %s, SKU %s", test.Manufacturer, test.Model, test.PartNumber, test.SKU)

		deviceSlug, err := suite.csm.identifyDeviceSlug(test.Manufacturer, test.Model, test.PartNumber)
		suite.NoError(err, msgString)

		suite.Equal("hpe-crayex-ex254n-compute-blade", deviceSlug, msgString)
	}
}

func TestIdentifyDeviceSlugTestSuite(t *testing.T) {
	suite.Run(t, new(IdentifyDeviceSlugTestSuite))
}
