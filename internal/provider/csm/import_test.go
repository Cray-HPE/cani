package csm

import (
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
