package csm

import (
	"testing"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/stretchr/testify/suite"
)

func buildCabinet(hardwareTypeLibrary hardwaretypes.Library, cabinetDeviceType string, chassisPopulation map[int]map[int]string) ([]inventory.Hardware, error) {
	allHardware := []inventory.Hardware{}

	// Build cabinet hardware

	// Build up blade hardware
	for chassisOrdinal, slots := range chassisPopulation {
		// Get chassis ID
		for bladeOrdinal, bladeDeviceType := range slots {
			
		}
	}

	return allHardware, nil
}

type DetermineHardwareClassSuite struct {
	suite.Suite

	hardwareTypeLibrary *hardwaretypes.Library
	data                inventory.Inventory
}

func (suite *DetermineHardwareClassSuite) SetupSuite() {
	var err error
	suite.hardwareTypeLibrary, err = hardwaretypes.NewEmbeddedLibrary()
	suite.NoError(err)

	// Generate a inventory of hardware
	datastore, err := inventory.NewDatastoreInMemory(inventory.CSMProvider)
	suite.NoError(err)

	system, err := datastore.GetSystemZero()
	suite.NoError(err)

	// Add a Mountain Cabinet
	// mountainCabinetHardware, err := suite.hardwareTypeLibrary.GetDefaultHardwareBuildOut("hpe-ex4000", 1000, system.ID)
	// suite.NoError(err)
	// // for

	// mountainBladeHardware, err := suite.hardwareTypeLibrary.GetDefaultHardwareBuildOut("hpe-crayex-ex420-compute-blade")
	mountainCabinetHardware := buildCabinet("hpe-ex4000", map[int]map[int]string{
		1: {
			0: "hpe-crayex-ex420-compute-blade",
		},
	})
	for _, hardware := range mountainCabinetHardware {
		err = datastore.Add(&hardware)
		suite.NoError(err)
	}

	// Add a Hill Cabinet

	// Add a blade in the hill cabinet

	// Add a River Cabinet

	// Add a blade in the river cabinet

}

func (suite *DetermineHardwareClassSuite) ClassMountain() {

}

func (suite *DetermineHardwareClassSuite) ClassHill() {

}

func (suite *DetermineHardwareClassSuite) ClassRiver() {

}

func TestDetermineHardwareClassSuite(t *testing.T) {
	suite.Run(t, new(DetermineHardwareClassSuite))
}
