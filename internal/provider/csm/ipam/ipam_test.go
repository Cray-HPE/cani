package ipam

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/provider/csm/sls"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/stretchr/testify/suite"
	"inet.af/netaddr"
)

const (
	testSLSFile = "../../../../testdata/fixtures/sls/valid-mug.json"
)

type ExistingIPAddressesSuite struct {
	suite.Suite

	slsState sls_client.SlsState
}

func (suite *ExistingIPAddressesSuite) SetupTest() {
	// Load SLS state
	slsStateRaw, err := ioutil.ReadFile(testSLSFile)
	suite.NoError(err)

	err = json.Unmarshal(slsStateRaw, &suite.slsState)
	suite.NoError(err)

}

func (suite *ExistingIPAddressesSuite) TestExistingIPAddresses_CAN() {

}

func (suite *ExistingIPAddressesSuite) TestExistingIPAddresses_CMN() {

}

func (suite *ExistingIPAddressesSuite) TestExistingIPAddresses_NMN_NetworkHardware() {
	subnet, _, err := sls.LookupSubnet(suite.slsState.Networks["NMN"], "network_hardware")
	suite.NoError(err)

	existingIPAddresses, err := ExistingIPAddresses(subnet)
	suite.NoError(err)

	// Build up expected IP address set
	expectedIPAddressesBuilder := &netaddr.IPSetBuilder{}
	for i := 1; i <= 4; i++ {
		expectedIPAddressesBuilder.Add(netaddr.MustParseIP(fmt.Sprintf("10.252.0.%d", i)))
	}
	expectedIPAddresses, err := expectedIPAddressesBuilder.IPSet()

	suite.NoError(err)
	suite.Equal(expectedIPAddresses, existingIPAddresses)

}

func (suite *ExistingIPAddressesSuite) TestExistingIPAddresses_NMN_BootstrapDHCP() {
	subnet, _, err := sls.LookupSubnet(suite.slsState.Networks["NMN"], "bootstrap_dhcp")
	suite.NoError(err)

	existingIPAddresses, err := ExistingIPAddresses(subnet)
	suite.NoError(err)

	// Build up expected IP address set
	expectedIPAddressesBuilder := &netaddr.IPSetBuilder{}
	expectedIPAddressesBuilder.Add(netaddr.MustParseIP("10.252.0.1")) // Gateway
	for i := 2; i <= 22; i++ {
		expectedIPAddressesBuilder.Add(netaddr.MustParseIP(fmt.Sprintf("10.252.1.%d", i)))
	}
	expectedIPAddresses, err := expectedIPAddressesBuilder.IPSet()

	suite.NoError(err)
	suite.Equal(expectedIPAddresses, existingIPAddresses)
}

func (suite *ExistingIPAddressesSuite) TestExistingIPAddresses_NMN_UAIMacVLAN() {
	subnet, _, err := sls.LookupSubnet(suite.slsState.Networks["NMN"], "uai_macvlan")
	suite.NoError(err)

	existingIPAddresses, err := ExistingIPAddresses(subnet)
	suite.NoError(err)

	// Build up expected IP address set
	expectedIPAddressesBuilder := &netaddr.IPSetBuilder{}
	expectedIPAddressesBuilder.Add(netaddr.MustParseIP("10.252.0.1")) // Gateway
	for i := 2; i <= 6; i++ {
		expectedIPAddressesBuilder.Add(netaddr.MustParseIP(fmt.Sprintf("10.252.2.%d", i)))
	}
	expectedIPAddresses, err := expectedIPAddressesBuilder.IPSet()

	suite.NoError(err)
	suite.Equal(expectedIPAddresses, existingIPAddresses)
}

func (suite *ExistingIPAddressesSuite) TestExistingIPAddresses_NMN_RVR_Cabinet3000() {
	subnet, _, err := sls.LookupSubnet(suite.slsState.Networks["NMN_RVR"], "cabinet_3000")
	suite.NoError(err)

	existingIPAddresses, err := ExistingIPAddresses(subnet)
	suite.NoError(err)

	// Build up expected IP address set
	expectedIPAddressesBuilder := &netaddr.IPSetBuilder{}
	expectedIPAddressesBuilder.Add(netaddr.MustParseIP("10.106.0.1")) // Gateway
	expectedIPAddresses, err := expectedIPAddressesBuilder.IPSet()
	suite.NoError(err)
	suite.Equal(expectedIPAddresses, existingIPAddresses)
}

func (suite *ExistingIPAddressesSuite) TestExistingIPAddresses_HMN_NetworkHardware() {
	subnet, _, err := sls.LookupSubnet(suite.slsState.Networks["HMN"], "network_hardware")
	suite.NoError(err)

	existingIPAddresses, err := ExistingIPAddresses(subnet)
	suite.NoError(err)

	// Build up expected IP address set
	expectedIPAddressesBuilder := &netaddr.IPSetBuilder{}
	expectedIPAddressesBuilder.Add(netaddr.MustParseIP("10.254.0.1")) // Gateway
	for i := 2; i <= 4; i++ {
		expectedIPAddressesBuilder.Add(netaddr.MustParseIP(fmt.Sprintf("10.254.0.%d", i)))
	}
	expectedIPAddresses, err := expectedIPAddressesBuilder.IPSet()
	suite.NoError(err)
	suite.Equal(expectedIPAddresses, existingIPAddresses)
}

func (suite *ExistingIPAddressesSuite) TestExistingIPAddresses_HMN_BootstrapDHCP() {
	subnet, _, err := sls.LookupSubnet(suite.slsState.Networks["HMN"], "bootstrap_dhcp")
	suite.NoError(err)

	existingIPAddresses, err := ExistingIPAddresses(subnet)
	suite.NoError(err)

	// Build up expected IP address set
	expectedIPAddressesBuilder := &netaddr.IPSetBuilder{}
	expectedIPAddressesBuilder.Add(netaddr.MustParseIP("10.254.0.1")) // Gateway
	for i := 2; i <= 40; i++ {
		expectedIPAddressesBuilder.Add(netaddr.MustParseIP(fmt.Sprintf("10.254.1.%d", i)))
	}
	expectedIPAddresses, err := expectedIPAddressesBuilder.IPSet()
	suite.NoError(err)
	suite.Equal(expectedIPAddresses, existingIPAddresses)
}

func (suite *ExistingIPAddressesSuite) TestExistingIPAddresses_HMN_RVR() {
	subnet, _, err := sls.LookupSubnet(suite.slsState.Networks["HMN_RVR"], "cabinet_3000")
	suite.NoError(err)

	existingIPAddresses, err := ExistingIPAddresses(subnet)
	suite.NoError(err)

	// Build up expected IP address set
	expectedIPAddressesBuilder := &netaddr.IPSetBuilder{}
	expectedIPAddressesBuilder.Add(netaddr.MustParseIP("10.107.0.1")) // Gateway
	expectedIPAddresses, err := expectedIPAddressesBuilder.IPSet()
	suite.NoError(err)
	suite.Equal(expectedIPAddresses, existingIPAddresses)
}

func (suite *ExistingIPAddressesSuite) TestExistingIPAddresses_InvalidGateway() {
	subnet := sls_client.NetworkIpv4Subnet{
		Gateway: "not valid IP address",
	}

	existingIPAddresses, err := ExistingIPAddresses(subnet)

	expectedErrorStrings := []string{
		"failed to parse gateway IP (not valid IP address)",
		"ParseIP(\"not valid IP address\"): unable to parse IP",
	}
	suite.EqualError(err, strings.Join(expectedErrorStrings, "\n"))
	suite.Empty(existingIPAddresses)
}

func (suite *ExistingIPAddressesSuite) TestExistingIPAddresses_InvalidIPAddressReservation() {
	subnet := sls_client.NetworkIpv4Subnet{
		Gateway: "10.0.0.1",
		IPReservations: []sls_client.NetworkIpReservation{
			{IPAddress: "10.0.0.2"},
			{IPAddress: "not valid IP address"},
		},
	}

	existingIPAddresses, err := ExistingIPAddresses(subnet)

	expectedErrorStrings := []string{
		"failed to parse IPReservation IP (not valid IP address)",
		"ParseIP(\"not valid IP address\"): unable to parse IP",
	}
	suite.EqualError(err, strings.Join(expectedErrorStrings, "\n"))
	suite.Empty(existingIPAddresses)
}

func TestExistingIPAddressesSuite(t *testing.T) {
	suite.Run(t, new(ExistingIPAddressesSuite))
}
