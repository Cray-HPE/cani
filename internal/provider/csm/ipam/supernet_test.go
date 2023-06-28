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

func (suite *IsSupernetHackedSuite) TestFoo() {
	network := suite.slsNetworks["HMN"]
	subnet, _, err := sls.LookupSubnet(network, "bootstrap_dhcp")
	suite.NoError(err)

	correctedSubnetCIDR, err := IsSupernetHacked(network, subnet)
	suite.NoError(err)

	expectedSubnetCIDR := netaddr.MustParseIPPrefix("10.254.0.0/24")
	suite.Equal(expectedSubnetCIDR, correctedSubnetCIDR)
}

func TestIsSupernetHackedSuite(t *testing.T) {
	suite.Run(t, new(IsSupernetHackedSuite))
}
