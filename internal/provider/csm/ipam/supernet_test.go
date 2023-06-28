package ipam

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/Cray-HPE/cani/internal/provider/csm/sls"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/stretchr/testify/suite"
	"inet.af/netaddr"
)

type IsSupernetHackedSuite struct {
	suite.Suite

	slsNetworks map[string]sls_client.Network
}

func (suite *IsSupernetHackedSuite) SetupTest() {
	// Load SLS state
	slsStateRaw, err := ioutil.ReadFile(testSLSFile)
	suite.NoError(err)

	var slsState sls_client.SlsState
	err = json.Unmarshal(slsStateRaw, &slsState)
	suite.NoError(err)

	suite.slsNetworks = slsState.Networks
}

func (suite *IsSupernetHackedSuite) TestHMN_BootstrapDHCP() {
	network := suite.slsNetworks["HMN"]
	subnet, _, err := sls.LookupSubnet(network, "bootstrap_dhcp")
	suite.NoError(err)

	correctedSubnetCIDR, err := IsSupernetHacked(network, subnet)
	suite.NoError(err)

	expectedSubnetCIDR := netaddr.MustParseIPPrefix("10.254.1.0/24")
	suite.Equal(&expectedSubnetCIDR, correctedSubnetCIDR)
}

func (suite *IsSupernetHackedSuite) TestHMN_NetworkHardware() {
	network := suite.slsNetworks["HMN"]
	subnet, _, err := sls.LookupSubnet(network, "network_hardware")
	suite.NoError(err)

	correctedSubnetCIDR, err := IsSupernetHacked(network, subnet)
	suite.NoError(err)

	expectedSubnetCIDR := netaddr.MustParseIPPrefix("10.254.0.0/24")
	suite.Equal(&expectedSubnetCIDR, correctedSubnetCIDR)
}

func (suite *IsSupernetHackedSuite) TestNMN_BootstrapDHCP() {
	network := suite.slsNetworks["NMN"]
	subnet, _, err := sls.LookupSubnet(network, "bootstrap_dhcp")
	suite.NoError(err)

	correctedSubnetCIDR, err := IsSupernetHacked(network, subnet)
	suite.NoError(err)

	expectedSubnetCIDR := netaddr.MustParseIPPrefix("10.252.1.0/24")
	suite.Equal(&expectedSubnetCIDR, correctedSubnetCIDR)
}

func (suite *IsSupernetHackedSuite) TestNMN_NetworkHardware() {
	network := suite.slsNetworks["NMN"]
	subnet, _, err := sls.LookupSubnet(network, "network_hardware")
	suite.NoError(err)

	correctedSubnetCIDR, err := IsSupernetHacked(network, subnet)
	suite.NoError(err)

	expectedSubnetCIDR := netaddr.MustParseIPPrefix("10.252.0.0/24")
	suite.Equal(&expectedSubnetCIDR, correctedSubnetCIDR)
}

func (suite *IsSupernetHackedSuite) TestMTL_BootstrapDHCP() {
	network := suite.slsNetworks["MTL"]
	subnet, _, err := sls.LookupSubnet(network, "bootstrap_dhcp")
	suite.NoError(err)

	correctedSubnetCIDR, err := IsSupernetHacked(network, subnet)
	suite.NoError(err)

	expectedSubnetCIDR := netaddr.MustParseIPPrefix("10.1.1.0/24")
	suite.Equal(&expectedSubnetCIDR, correctedSubnetCIDR)
}

func (suite *IsSupernetHackedSuite) TestMTL_NetworkHardware() {
	network := suite.slsNetworks["MTL"]
	subnet, _, err := sls.LookupSubnet(network, "network_hardware")
	suite.NoError(err)

	correctedSubnetCIDR, err := IsSupernetHacked(network, subnet)
	suite.NoError(err)

	expectedSubnetCIDR := netaddr.MustParseIPPrefix("10.1.0.0/24")
	suite.Equal(&expectedSubnetCIDR, correctedSubnetCIDR)
}

func (suite *IsSupernetHackedSuite) TestCMN_MetalLBStaticPool() {
	network := suite.slsNetworks["CMN"]
	subnet, _, err := sls.LookupSubnet(network, "cmn_metallb_static_pool")
	suite.NoError(err)

	correctedSubnetCIDR, err := IsSupernetHacked(network, subnet)
	suite.EqualError(err, "allocating an IP address on the (CMN) network is not currently supported")
	suite.Nil(correctedSubnetCIDR)
}

func (suite *IsSupernetHackedSuite) TestCMN_MetalLBDynamicPool() {
	network := suite.slsNetworks["CMN"]
	subnet, _, err := sls.LookupSubnet(network, "cmn_metallb_address_pool")
	suite.NoError(err)

	correctedSubnetCIDR, err := IsSupernetHacked(network, subnet)
	suite.EqualError(err, "allocating an IP address on the (CMN) network is not currently supported")
	suite.Nil(correctedSubnetCIDR)
}

func (suite *IsSupernetHackedSuite) TestCMN_NetworkHardware() {
	network := suite.slsNetworks["CMN"]
	subnet, _, err := sls.LookupSubnet(network, "network_hardware")
	suite.NoError(err)

	correctedSubnetCIDR, err := IsSupernetHacked(network, subnet)
	suite.EqualError(err, "allocating an IP address on the (CMN) network is not currently supported")
	suite.Nil(correctedSubnetCIDR)
}

func (suite *IsSupernetHackedSuite) TestCAN_MetalLBDynamicPool() {
	network := suite.slsNetworks["CAN"]
	subnet, _, err := sls.LookupSubnet(network, "can_metallb_address_pool")
	suite.NoError(err)

	correctedSubnetCIDR, err := IsSupernetHacked(network, subnet)
	suite.EqualError(err, "allocating an IP address on the (CAN) network is not currently supported")
	suite.Nil(correctedSubnetCIDR)
}

func TestIsSupernetHackedSuite(t *testing.T) {
	suite.Run(t, new(IsSupernetHackedSuite))
}
