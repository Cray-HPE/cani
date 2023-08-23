package inventory

import (
	"sort"
	"testing"

	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type DatastoreJSONSearchTestSuite struct {
	suite.Suite

	datastore Datastore
}

func (suite *DatastoreJSONSearchTestSuite) LenHardware(allHardware map[uuid.UUID]Hardware, length int) {
	// Mark this as a helper function to it shows the right line number in to the
	// go test output.
	suite.T().Helper()

	// Build up a slice of location paths, as it makes the test output a bit more readable
	locationPaths := []string{}
	for _, hardware := range allHardware {
		locationPaths = append(locationPaths, hardware.LocationPath.String())
	}
	sort.Strings(locationPaths)

	suite.Len(locationPaths, length)
}

func (suite *DatastoreJSONSearchTestSuite) SetupTest() {
	var err error
	suite.datastore, err = NewDatastoreJSON("../../testdata/fixtures/cani/configs/canitestdb_valid_system_import.json", "", CSMProvider)
	suite.NoError(err)
}

func (suite *DatastoreJSONSearchTestSuite) TestEmptyFilter() {
	// Empty filter should match nothing
	hardware, err := suite.datastore.Search(SearchFilter{})
	suite.NoError(err)
	suite.LenHardware(hardware, 120)
}

func (suite *DatastoreJSONSearchTestSuite) TestSingleTypeSystem() {
	hardware, err := suite.datastore.Search(SearchFilter{
		Types: []hardwaretypes.HardwareType{
			hardwaretypes.System,
		},
	})
	suite.NoError(err)
	suite.LenHardware(hardware, 1)
}

func (suite *DatastoreJSONSearchTestSuite) TestSingleTypeCabinet() {
	hardware, err := suite.datastore.Search(SearchFilter{
		Types: []hardwaretypes.HardwareType{
			hardwaretypes.Cabinet,
		},
	})
	suite.NoError(err)
	suite.LenHardware(hardware, 1)
}

func (suite *DatastoreJSONSearchTestSuite) TestSingleTypeChassis() {
	hardware, err := suite.datastore.Search(SearchFilter{
		Types: []hardwaretypes.HardwareType{
			hardwaretypes.Chassis,
		},
	})
	suite.NoError(err)
	suite.LenHardware(hardware, 2)
}

func (suite *DatastoreJSONSearchTestSuite) TestSingleTypeNode() {
	hardware, err := suite.datastore.Search(SearchFilter{
		Types: []hardwaretypes.HardwareType{
			hardwaretypes.Node,
		},
	})
	suite.NoError(err)
	suite.LenHardware(hardware, 64)
}

func (suite *DatastoreJSONSearchTestSuite) TestMultipleTypeSystemNode() {
	hardware, err := suite.datastore.Search(SearchFilter{
		Types: []hardwaretypes.HardwareType{
			hardwaretypes.System,
			hardwaretypes.Node,
		},
	})
	suite.NoError(err)
	suite.LenHardware(hardware, 1+64)
}

func (suite *DatastoreJSONSearchTestSuite) TestMultipleTypeCabinetChassis() {
	hardware, err := suite.datastore.Search(SearchFilter{
		Types: []hardwaretypes.HardwareType{
			hardwaretypes.Cabinet,
			hardwaretypes.Chassis,
		},
	})
	suite.NoError(err)
	suite.LenHardware(hardware, 1+2)
}

func (suite *DatastoreJSONSearchTestSuite) TestSingleStatusEmpty() {
	hardware, err := suite.datastore.Search(SearchFilter{
		Status: []HardwareStatus{
			HardwareStatusEmpty,
		},
	})
	suite.NoError(err)
	suite.LenHardware(hardware, 108)
}

func (suite *DatastoreJSONSearchTestSuite) TestSingleStatusStaged() {
	hardware, err := suite.datastore.Search(SearchFilter{
		Status: []HardwareStatus{
			HardwareStatusStaged,
		},
	})
	suite.NoError(err)
	suite.LenHardware(hardware, 5)
}

func (suite *DatastoreJSONSearchTestSuite) TestSingleStatusProvisioned() {
	hardware, err := suite.datastore.Search(SearchFilter{
		Status: []HardwareStatus{
			HardwareStatusProvisioned,
		},
	})
	suite.NoError(err)
	suite.LenHardware(hardware, 6)
}

func (suite *DatastoreJSONSearchTestSuite) TestMultipleStatusEmptyStaged() {
	hardware, err := suite.datastore.Search(SearchFilter{
		Status: []HardwareStatus{
			HardwareStatusEmpty,
			HardwareStatusStaged,
		},
	})
	suite.NoError(err)
	suite.LenHardware(hardware, 113)
}

func (suite *DatastoreJSONSearchTestSuite) TestMultipleStatusProvisionedStaged() {
	hardware, err := suite.datastore.Search(SearchFilter{
		Status: []HardwareStatus{
			HardwareStatusProvisioned,
			HardwareStatusStaged,
		},
	})
	suite.NoError(err)
	suite.LenHardware(hardware, 11)
}

func (suite *DatastoreJSONSearchTestSuite) TestSingleStatusMultipleType() {
	hardware, err := suite.datastore.Search(SearchFilter{
		Status: []HardwareStatus{
			HardwareStatusStaged,
		},
		Types: []hardwaretypes.HardwareType{
			hardwaretypes.NodeCard,
			hardwaretypes.Node,
		},
	})
	suite.NoError(err)
	// 2 Node
	// 1 NodeCard
	suite.LenHardware(hardware, 3)
}

func (suite *DatastoreJSONSearchTestSuite) TestSingleTypeMultipleStatus() {
	hardware, err := suite.datastore.Search(SearchFilter{
		Status: []HardwareStatus{
			HardwareStatusStaged,
			HardwareStatusEmpty,
		},
		Types: []hardwaretypes.HardwareType{
			hardwaretypes.Node,
		},
	})
	suite.NoError(err)
	// 62 empty
	// 2 staged
	suite.LenHardware(hardware, 2+62)
}

func (suite *DatastoreJSONSearchTestSuite) TestMultipleTypeMultipleStatus() {
	hardware, err := suite.datastore.Search(SearchFilter{
		Status: []HardwareStatus{
			HardwareStatusStaged,
			HardwareStatusEmpty,
		},
		Types: []hardwaretypes.HardwareType{
			hardwaretypes.Node,
			hardwaretypes.NodeCard,
		},
	})
	suite.NoError(err)
	// Status counts
	// 93 empty
	// 3 staged
	// Type counts
	// 64 Node
	// 32 NodeCard
	suite.LenHardware(hardware, 93+3)
}

func TestDatastoreJSONSearchTestSuite(t *testing.T) {
	suite.Run(t, new(DatastoreJSONSearchTestSuite))
}
