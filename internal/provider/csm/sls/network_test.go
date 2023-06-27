package sls

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/stretchr/testify/suite"
)

const (
	testSLSFile = "../../../../testdata/fixtures/sls/valid-mug.json"
)

type LookupSubnetSuite struct {
	suite.Suite

	networkUnderTest sls_client.Network
}

func (suite *LookupSubnetSuite) SetupTest() {
	// Load SLS state
	slsStateRaw, err := ioutil.ReadFile(testSLSFile)
	suite.NoError(err)

	var slsState sls_client.SlsState
	err = json.Unmarshal(slsStateRaw, &slsState)
	suite.NoError(err)

	// Extract the wanted network
	suite.networkUnderTest = slsState.Networks["HMN"]
}

func (suite *LookupSubnetSuite) Test_NetworkHardware() {
	subnet, index, err := LookupSubnet(suite.networkUnderTest, "network_hardware")
	suite.NoError(err)
	suite.Equal(0, index)

	suite.Equal("10.254.0.0/17", subnet.CIDR)
	suite.Equal("HMN Management Network Infrastructure", subnet.FullName)
	suite.Equal("10.254.0.1", subnet.Gateway)
	suite.Len(subnet.IPReservations, 3)
	suite.Equal("network_hardware", subnet.Name)
	suite.Equal(int32(4), subnet.VlanID)
}

func (suite *LookupSubnetSuite) Test_BootstrapDHCP() {
	subnet, index, err := LookupSubnet(suite.networkUnderTest, "bootstrap_dhcp")
	suite.NoError(err)
	suite.Equal(1, index)

	suite.Equal("10.254.1.0/17", subnet.CIDR)
	suite.Equal("HMN Bootstrap DHCP Subnet", subnet.FullName)
	suite.Equal("10.254.0.1", subnet.Gateway)
	suite.Len(subnet.IPReservations, 39)
	suite.Equal("bootstrap_dhcp", subnet.Name)
	suite.Equal(int32(4), subnet.VlanID)
}

func (suite *LookupSubnetSuite) TestInvalid_NotFound() {
	_, _, err := LookupSubnet(suite.networkUnderTest, "does-not-exist")
	suite.EqualError(err, "subnet not found (does-not-exist)")
}

func (suite *LookupSubnetSuite) TestInvalid_EmptyNetwork() {
	emptyNetwork := sls_client.Network{}

	_, _, err := LookupSubnet(emptyNetwork, "some-subnet")
	suite.EqualError(err, "subnet not found (some-subnet)")
}

func (suite *LookupSubnetSuite) TestInvalid_NoSubnets() {
	emptyNetwork := sls_client.Network{
		ExtraProperties: &sls_client.NetworkExtraProperties{},
	}

	_, _, err := LookupSubnet(emptyNetwork, "some-subnet")
	suite.EqualError(err, "subnet not found (some-subnet)")
}

func (suite *LookupSubnetSuite) TestInvalid_MultipleSubnetsWithSameName() {
	malformedNetwokr := sls_client.Network{
		ExtraProperties: &sls_client.NetworkExtraProperties{
			Subnets: []sls_client.NetworkIpv4Subnet{
				{Name: "my-subnet"},
				{Name: "my-subnet"},
			},
		},
	}

	_, _, err := LookupSubnet(malformedNetwokr, "my-subnet")
	suite.EqualError(err, "found 2 subnets instead of just one")
}

func TestLookupSubnetSuite(t *testing.T) {
	suite.Run(t, new(LookupSubnetSuite))
}

type ReservationsByNameSuite struct {
	suite.Suite

	subnetUnderTest sls_client.NetworkIpv4Subnet
}

func (suite *ReservationsByNameSuite) SetupTest() {
	// Load SLS state
	slsStateRaw, err := ioutil.ReadFile(testSLSFile)
	suite.NoError(err)

	var slsState sls_client.SlsState
	err = json.Unmarshal(slsStateRaw, &slsState)
	suite.NoError(err)

	// Extract the wanted subnet
	suite.subnetUnderTest, _, err = LookupSubnet(slsState.Networks["HMN"], "network_hardware")
	suite.NoError(err)
}

func (suite *ReservationsByNameSuite) Test() {
	reservations := ReservationsByName(suite.subnetUnderTest)
	suite.NotEmpty(reservations)

	expectedReservations := map[string]sls_client.NetworkIpReservation{
		"sw-spine-001": {
			Name:      "sw-spine-001",
			IPAddress: "10.254.0.2",
			Comment:   "x3000c0h41s1",
		},
		"sw-spine-002": {
			Name:      "sw-spine-002",
			IPAddress: "10.254.0.3",
			Comment:   "x3000c0h42s1",
		},
		"sw-leaf-bmc-001": {
			Name:      "sw-leaf-bmc-001",
			IPAddress: "10.254.0.4",
			Comment:   "x3000c0w22",
		},
	}
	suite.Equal(expectedReservations, reservations)
}

func TestReservationsByNameSuite(t *testing.T) {
	suite.Run(t, new(ReservationsByNameSuite))
}
