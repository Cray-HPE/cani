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
package ipam

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/provider/csm/sls"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/stretchr/testify/suite"
	"inet.af/netaddr"
)

const (
	testSLSFile = "../../../../testdata/fixtures/sls/valid-mug.json"
)

//
// ExistingIPAddressesSuite
//

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

func (suite *ExistingIPAddressesSuite) TestExistingIPAddresses_CAN_BootstrapDHCP() {
	subnet, _, err := sls.LookupSubnet(suite.slsState.Networks["CAN"], "bootstrap_dhcp")
	suite.NoError(err)

	existingIPAddresses, err := ExistingIPAddresses(subnet)
	suite.NoError(err)

	// Build up expected IP address set
	expectedIPAddressesBuilder := &netaddr.IPSetBuilder{}
	for i := 129; i <= 152; i++ {
		expectedIPAddressesBuilder.Add(netaddr.MustParseIP(fmt.Sprintf("10.102.162.%d", i)))
	}
	expectedIPAddresses, err := expectedIPAddressesBuilder.IPSet()

	suite.NoError(err)
	suite.Equal(expectedIPAddresses, existingIPAddresses)
}

func (suite *ExistingIPAddressesSuite) TestExistingIPAddresses_CMN_BootstrapDHCP() {
	subnet, _, err := sls.LookupSubnet(suite.slsState.Networks["CMN"], "bootstrap_dhcp")
	suite.NoError(err)

	existingIPAddresses, err := ExistingIPAddresses(subnet)
	suite.NoError(err)

	// Build up expected IP address set
	expectedIPAddressesBuilder := &netaddr.IPSetBuilder{}
	expectedIPAddressesBuilder.Add(netaddr.MustParseIP("10.102.162.1")) // Gateway
	for i := 18; i <= 37; i++ {
		expectedIPAddressesBuilder.Add(netaddr.MustParseIP(fmt.Sprintf("10.102.162.%d", i)))
	}
	expectedIPAddresses, err := expectedIPAddressesBuilder.IPSet()

	suite.NoError(err)
	suite.Equal(expectedIPAddresses, existingIPAddresses)
}

func (suite *ExistingIPAddressesSuite) TestExistingIPAddresses_CMN_NetworkHardware() {
	subnet, _, err := sls.LookupSubnet(suite.slsState.Networks["CMN"], "network_hardware")
	suite.NoError(err)

	existingIPAddresses, err := ExistingIPAddresses(subnet)
	suite.NoError(err)

	// Build up expected IP address set
	expectedIPAddressesBuilder := &netaddr.IPSetBuilder{}
	expectedIPAddressesBuilder.Add(netaddr.MustParseIP("10.102.162.1")) // Gateway
	for i := 2; i <= 4; i++ {
		expectedIPAddressesBuilder.Add(netaddr.MustParseIP(fmt.Sprintf("10.102.162.%d", i)))
	}
	expectedIPAddresses, err := expectedIPAddressesBuilder.IPSet()

	suite.NoError(err)
	suite.Equal(expectedIPAddresses, existingIPAddresses)
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

//
// FindNextAvailableIPSuite
//

type FindNextAvailableIPSuite struct {
	suite.Suite
}

func (suite *FindNextAvailableIPSuite) TestAllocateIP() {
	subnet := sls_client.NetworkIpv4Subnet{
		Name:    "test_subnet",
		CIDR:    "10.0.0.0/24",
		Gateway: "10.0.0.1",
		IPReservations: []sls_client.NetworkIpReservation{
			{Name: "foo", IPAddress: "10.0.0.2"},
			{Name: "bar", IPAddress: "10.0.0.3"},
		},
	}
	network := sls_client.Network{
		Name: "test_network",
		ExtraProperties: &sls_client.NetworkExtraProperties{
			CIDR:    "10.0.0.0/22",
			Subnets: []sls_client.NetworkIpv4Subnet{subnet},
		},
	}

	ipAddress, err := FindNextAvailableIP(network, subnet)
	suite.NoError(err)

	expectedIPaddress := netaddr.MustParseIP("10.0.0.4")
	suite.Equal(expectedIPaddress, ipAddress)
}

func (suite *FindNextAvailableIPSuite) TestFullSubnet() {
	subnet := sls_client.NetworkIpv4Subnet{
		Name:           "test_subnet",
		CIDR:           "10.0.0.0/24",
		Gateway:        "10.0.0.1",
		IPReservations: []sls_client.NetworkIpReservation{},
	}

	// Fill subnet with IP address reservations
	for i := 2; i < 255; i++ {
		subnet.IPReservations = append(subnet.IPReservations, sls_client.NetworkIpReservation{
			Name:      fmt.Sprintf("reservation_%d", i),
			IPAddress: fmt.Sprintf("10.0.0.%d", i),
		})
	}

	network := sls_client.Network{
		Name: "test_network",
		ExtraProperties: &sls_client.NetworkExtraProperties{
			CIDR:    "10.0.0.0/22",
			Subnets: []sls_client.NetworkIpv4Subnet{subnet},
		},
	}

	ipAddress, err := FindNextAvailableIP(network, subnet)
	suite.EqualError(err, "subnet has no available IPs")
	suite.Equal(netaddr.IP{}, ipAddress)
}

func (suite *FindNextAvailableIPSuite) TestInvalidSubnetCIDR() {
	subnet := sls_client.NetworkIpv4Subnet{
		Name:           "test_subnet",
		CIDR:           "not-a-cidr",
		Gateway:        "10.0.0.1",
		IPReservations: []sls_client.NetworkIpReservation{},
	}

	network := sls_client.Network{
		Name: "test_network",
		ExtraProperties: &sls_client.NetworkExtraProperties{
			CIDR:    "10.0.0.0/22",
			Subnets: []sls_client.NetworkIpv4Subnet{subnet},
		},
	}

	ipAddress, err := FindNextAvailableIP(network, subnet)
	suite.Equal(netaddr.IP{}, ipAddress)

	expectedErrorStrings := []string{
		"failed to parse subnet CIDR (not-a-cidr)",
		"netaddr.ParseIPPrefix(\"not-a-cidr\"): no '/'",
	}
	suite.EqualError(err, strings.Join(expectedErrorStrings, "\n"))
}

func TestFindNextAvailableIPSuite(t *testing.T) {
	suite.Run(t, new(FindNextAvailableIPSuite))
}

//
// AdvanceIPSuite
//

type AdvanceIPSuite struct {
	suite.Suite
}

func (suite *AdvanceIPSuite) TestAdvanceZero() {
	startingIP := netaddr.MustParseIP("10.254.0.10")

	ip, err := AdvanceIP(startingIP, 0)
	suite.NoError(err)

	expectedIP := netaddr.MustParseIP("10.254.0.10")
	suite.Equal(expectedIP, ip)
}

func (suite *AdvanceIPSuite) TestAdvanceOne() {
	startingIP := netaddr.MustParseIP("10.254.0.10")

	ip, err := AdvanceIP(startingIP, 1)
	suite.NoError(err)

	expectedIP := netaddr.MustParseIP("10.254.0.11")
	suite.Equal(expectedIP, ip)
}

func (suite *AdvanceIPSuite) TestAdvanceTen() {
	startingIP := netaddr.MustParseIP("10.254.0.10")

	ip, err := AdvanceIP(startingIP, 10)
	suite.NoError(err)

	expectedIP := netaddr.MustParseIP("10.254.0.20")
	suite.Equal(expectedIP, ip)
}

func (suite *AdvanceIPSuite) TestIPV6() {
	ip := netaddr.MustParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334")

	_, err := AdvanceIP(ip, 1)
	suite.EqualError(err, "IPv6 is not supported")
}

func (suite *AdvanceIPSuite) TestEmptyIP() {
	ip := netaddr.IP{}

	_, err := AdvanceIP(ip, 1)
	suite.EqualError(err, "empty IP address provided")
}

func TestAdvanceIPSuite(t *testing.T) {
	suite.Run(t, new(AdvanceIPSuite))
}

//
// SplitNetworkSuite
//

type SplitNetworkSuite struct {
	suite.Suite
}

func (suite *SplitNetworkSuite) TestCabinetSplitScenario() {
	network := netaddr.MustParseIPPrefix("10.254.0.0/17")

	subnets, err := SplitNetwork(network, 22)
	suite.NoError(err)

	expectedSubnets := []netaddr.IPPrefix{}
	for i := 0; i <= 124; i = i + 4 {
		subnet := netaddr.MustParseIPPrefix(fmt.Sprintf("10.254.%d.0/22", i))
		expectedSubnets = append(expectedSubnets, subnet)
	}
	suite.Equal(expectedSubnets, subnets)
}

func (suite *SplitNetworkSuite) TestSplitInHalf() {
	network := netaddr.MustParseIPPrefix("10.0.0.0/23")

	subnets, err := SplitNetwork(network, 24)
	suite.NoError(err)

	expectedSubnets := []netaddr.IPPrefix{
		netaddr.MustParseIPPrefix("10.0.0.0/24"),
		netaddr.MustParseIPPrefix("10.0.1.0/24"),
	}
	suite.Equal(expectedSubnets, subnets)
}

func (suite *SplitNetworkSuite) TestSubnetLargerThanNetworkBeingSplit() {
	network := netaddr.MustParseIPPrefix("10.0.0.0/24")

	subnets, err := SplitNetwork(network, 16)
	suite.EqualError(err, "provided subnet mask bits /16 is larger than starting network subnet mask /24")
	suite.Empty(subnets)
}

func (suite *SplitNetworkSuite) TestSameSubnetSize() {
	network := netaddr.MustParseIPPrefix("10.0.0.0/16")

	subnets, err := SplitNetwork(network, 16)
	suite.NoError(err)

	expectedSubnets := []netaddr.IPPrefix{
		netaddr.MustParseIPPrefix("10.0.0.0/16"),
	}
	suite.Equal(expectedSubnets, subnets)
}

func (suite *SplitNetworkSuite) TestInvalidSubnets() {
	network := netaddr.MustParseIPPrefix("10.0.0.0/16")

	// Build up subnet mask bits.
	// Basically all of the values of unint8 that are not between 16 and 30
	invalidSubnetMaskOneBits := []uint8{}
	for i := uint8(0); i < uint8(16); i++ {
		invalidSubnetMaskOneBits = append(invalidSubnetMaskOneBits, i)
	}
	for i := uint8(31); i < uint8(255); i++ {
		invalidSubnetMaskOneBits = append(invalidSubnetMaskOneBits, i)
	}

	for _, subnetMaskOneBits := range invalidSubnetMaskOneBits {
		subnets, err := SplitNetwork(network, subnetMaskOneBits)
		suite.EqualError(err, fmt.Sprintf("invalid subnet mask provided /%d", subnetMaskOneBits))
		suite.Empty(subnets)
	}

}

func TestSplitNetworkSuite(t *testing.T) {
	suite.Run(t, new(SplitNetworkSuite))
}

//
// FindNextAvailableSubnetSuite
//

type FindNextAvailableSubnetSuite struct {
	suite.Suite

	slsState sls_client.SlsState
}

func (suite *FindNextAvailableSubnetSuite) SetupTest() {
	// Load SLS state
	slsStateRaw, err := ioutil.ReadFile(testSLSFile)
	suite.NoError(err)

	err = json.Unmarshal(slsStateRaw, &suite.slsState)
	suite.NoError(err)
}

func (suite *FindNextAvailableSubnetSuite) TestAllocate_HMN_MTN() {
	networkExtraProperties := *suite.slsState.Networks["HMN_MTN"].ExtraProperties

	subnet, err := FindNextAvailableSubnet(networkExtraProperties)
	suite.NoError(err)

	expectedSubnet := netaddr.MustParseIPPrefix("11.254.0.0/22")
	suite.Equal(expectedSubnet, subnet)
}

func (suite *FindNextAvailableSubnetSuite) TestAllocate_HMN_RVR() {
	networkExtraProperties := *suite.slsState.Networks["HMN_RVR"].ExtraProperties

	subnet, err := FindNextAvailableSubnet(networkExtraProperties)
	suite.NoError(err)

	expectedSubnet := netaddr.MustParseIPPrefix("10.107.8.0/22")
	suite.Equal(expectedSubnet, subnet)
}
func (suite *FindNextAvailableSubnetSuite) TestAllocate_NMN_MTN() {
	networkExtraProperties := *suite.slsState.Networks["NMN_MTN"].ExtraProperties

	subnet, err := FindNextAvailableSubnet(networkExtraProperties)
	suite.NoError(err)

	expectedSubnet := netaddr.MustParseIPPrefix("11.252.0.0/22")
	suite.Equal(expectedSubnet, subnet)
}

func (suite *FindNextAvailableSubnetSuite) TestAllocate_NMN_RVR() {
	networkExtraProperties := *suite.slsState.Networks["NMN_RVR"].ExtraProperties

	subnet, err := FindNextAvailableSubnet(networkExtraProperties)
	suite.NoError(err)

	expectedSubnet := netaddr.MustParseIPPrefix("10.106.8.0/22")
	suite.Equal(expectedSubnet, subnet)
}

func (suite *FindNextAvailableSubnetSuite) TestNearFullNetwork() {
	networkExtraProperties := sls_client.NetworkExtraProperties{
		CIDR: "10.254.0.0/21",
		Subnets: []sls_client.NetworkIpv4Subnet{
			{CIDR: "10.254.0.0/22"},
		},
	}

	subnet, err := FindNextAvailableSubnet(networkExtraProperties)
	suite.NoError(err)

	expectedSubnet := netaddr.MustParseIPPrefix("10.254.4.0/22")
	suite.Equal(expectedSubnet, subnet)
}

func (suite *FindNextAvailableSubnetSuite) TestNetworkTooSmall() {
	networkExtraProperties := sls_client.NetworkExtraProperties{
		CIDR: "10.254.0.0/24",
	}

	subnet, err := FindNextAvailableSubnet(networkExtraProperties)

	expectedErrorStrings := []string{
		"failed to split network CIDR (10.254.0.0/24)",
		"provided subnet mask bits /22 is larger than starting network subnet mask /24",
	}
	suite.EqualError(err, strings.Join(expectedErrorStrings, "\n"))
	suite.Empty(subnet)
}

func (suite *FindNextAvailableSubnetSuite) TestNoRoom() {
	networkExtraProperties := sls_client.NetworkExtraProperties{
		CIDR: "10.254.0.0/21",
		Subnets: []sls_client.NetworkIpv4Subnet{
			{CIDR: "10.254.0.0/22"},
			{CIDR: "10.254.4.0/22"},
		},
	}

	subnet, err := FindNextAvailableSubnet(networkExtraProperties)
	suite.EqualError(err, "network space has been exhausted")
	suite.Empty(subnet)
}

func (suite *FindNextAvailableSubnetSuite) TestInvalidNetworkCIDR() {
	networkExtraProperties := sls_client.NetworkExtraProperties{
		CIDR: "10.254.0.0/16",
		Subnets: []sls_client.NetworkIpv4Subnet{
			{CIDR: "10.254.4.0/22"},
			{CIDR: "not-a-cidr"},
		},
	}

	subnet, err := FindNextAvailableSubnet(networkExtraProperties)

	expectedErrorStrings := []string{
		"failed to parse subnet CIDR (not-a-cidr)",
		"netaddr.ParseIPPrefix(\"not-a-cidr\"): no '/'",
	}
	suite.EqualError(err, strings.Join(expectedErrorStrings, "\n"))
	suite.Empty(subnet)
}

func (suite *FindNextAvailableSubnetSuite) TestInvalidSubnetCIDR() {
	networkExtraProperties := sls_client.NetworkExtraProperties{
		CIDR: "not-a-cidr",
	}

	subnet, err := FindNextAvailableSubnet(networkExtraProperties)

	expectedErrorStrings := []string{
		"failed to parse network CIDR (not-a-cidr)",
		"netaddr.ParseIPPrefix(\"not-a-cidr\"): no '/'",
	}
	suite.EqualError(err, strings.Join(expectedErrorStrings, "\n"))
	suite.Empty(subnet)
}

func (suite *FindNextAvailableSubnetSuite) TestEmptyNetworkExtraProperties() {
	networkExtraProperties := sls_client.NetworkExtraProperties{}

	subnet, err := FindNextAvailableSubnet(networkExtraProperties)

	expectedErrorStrings := []string{
		"failed to parse network CIDR ()",
		"netaddr.ParseIPPrefix(\"\"): no '/'",
	}
	suite.EqualError(err, strings.Join(expectedErrorStrings, "\n"))
	suite.Empty(subnet)
}

func TestFindNextAvailableSubnetSuite(t *testing.T) {
	suite.Run(t, new(FindNextAvailableSubnetSuite))
}

//
// AllocateCabinetSubnetSuite
//

type AllocateCabinetSubnetSuite struct {
	suite.Suite

	slsNetworks map[string]sls_client.Network
}

func (suite *AllocateCabinetSubnetSuite) SetupTest() {
	// Load SLS state
	slsStateRaw, err := ioutil.ReadFile(testSLSFile)
	suite.NoError(err)

	var slsState sls_client.SlsState
	err = json.Unmarshal(slsStateRaw, &slsState)
	suite.NoError(err)

	suite.slsNetworks = slsState.Networks
}

func (suite *AllocateCabinetSubnetSuite) TestAllocateNMN_MTN() {
	network := suite.slsNetworks["NMN_MTN"]
	network.ExtraProperties.Subnets = nil

	// Allocate a subnet into a network without subnets
	subnet, err := AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 1000}, nil)
	suite.NoError(err)

	expectedSubnet := sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_1000",
		CIDR:      "11.252.0.0/22",
		Gateway:   "11.252.0.1",
		DHCPStart: "11.252.0.10",
		DHCPEnd:   "11.252.3.254",
		VlanID:    2000,
	}
	suite.Equal(expectedSubnet, subnet)

	// Add a second cabinet subnet
	network.ExtraProperties.Subnets = append(network.ExtraProperties.Subnets, expectedSubnet)

	subnet, err = AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 1001}, nil)
	suite.NoError(err)
	expectedSubnet = sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_1001",
		CIDR:      "11.252.4.0/22",
		Gateway:   "11.252.4.1",
		DHCPStart: "11.252.4.10",
		DHCPEnd:   "11.252.7.254",
		VlanID:    2001,
	}
	suite.Equal(expectedSubnet, subnet)
}

func (suite *AllocateCabinetSubnetSuite) TestAllocateHMN_MTN() {
	network := suite.slsNetworks["HMN_MTN"]
	network.ExtraProperties.Subnets = nil

	// Allocate a subnet into a network without subnets
	subnet, err := AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 1000}, nil)
	suite.NoError(err)

	expectedSubnet := sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_1000",
		CIDR:      "11.254.0.0/22",
		Gateway:   "11.254.0.1",
		DHCPStart: "11.254.0.10",
		DHCPEnd:   "11.254.3.254",
		VlanID:    3000,
	}
	suite.Equal(expectedSubnet, subnet)

	// Add a second cabinet subnet
	network.ExtraProperties.Subnets = append(network.ExtraProperties.Subnets, expectedSubnet)

	subnet, err = AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 1001}, nil)
	suite.NoError(err)
	expectedSubnet = sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_1001",
		CIDR:      "11.254.4.0/22",
		Gateway:   "11.254.4.1",
		DHCPStart: "11.254.4.10",
		DHCPEnd:   "11.254.7.254",
		VlanID:    3001,
	}
	suite.Equal(expectedSubnet, subnet)
}

func (suite *AllocateCabinetSubnetSuite) TestAllocateNMN_MTN_VlanOverride() {
	network := suite.slsNetworks["NMN_MTN"]
	network.ExtraProperties.Subnets = nil

	// Allocate a subnet into a network without subnets
	vlanOverride := int32(1234)
	subnet, err := AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 1000}, &vlanOverride)
	suite.NoError(err)

	expectedSubnet := sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_1000",
		CIDR:      "11.252.0.0/22",
		Gateway:   "11.252.0.1",
		DHCPStart: "11.252.0.10",
		DHCPEnd:   "11.252.3.254",
		VlanID:    1234,
	}
	suite.Equal(expectedSubnet, subnet)

	// Add a second cabinet subnet
	network.ExtraProperties.Subnets = append(network.ExtraProperties.Subnets, expectedSubnet)

	vlanOverride = int32(2234)
	subnet, err = AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 1001}, &vlanOverride)
	suite.NoError(err)
	expectedSubnet = sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_1001",
		CIDR:      "11.252.4.0/22",
		Gateway:   "11.252.4.1",
		DHCPStart: "11.252.4.10",
		DHCPEnd:   "11.252.7.254",
		VlanID:    2234,
	}
	suite.Equal(expectedSubnet, subnet)
}

func (suite *AllocateCabinetSubnetSuite) TestAllocateHMN_MTN_VlanOverride() {
	network := suite.slsNetworks["HMN_MTN"]
	network.ExtraProperties.Subnets = nil

	// Allocate a subnet into a network without subnets
	vlanOverride := int32(1234)
	subnet, err := AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 1000}, &vlanOverride)
	suite.NoError(err)

	expectedSubnet := sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_1000",
		CIDR:      "11.254.0.0/22",
		Gateway:   "11.254.0.1",
		DHCPStart: "11.254.0.10",
		DHCPEnd:   "11.254.3.254",
		VlanID:    1234,
	}
	suite.Equal(expectedSubnet, subnet)

	// Add a second cabinet subnet
	network.ExtraProperties.Subnets = append(network.ExtraProperties.Subnets, expectedSubnet)

	vlanOverride = int32(2234)
	subnet, err = AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 1001}, &vlanOverride)
	suite.NoError(err)
	expectedSubnet = sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_1001",
		CIDR:      "11.254.4.0/22",
		Gateway:   "11.254.4.1",
		DHCPStart: "11.254.4.10",
		DHCPEnd:   "11.254.7.254",
		VlanID:    2234,
	}
	suite.Equal(expectedSubnet, subnet)
}

func (suite *AllocateCabinetSubnetSuite) TestAllocateNMN_RVR() {
	network := suite.slsNetworks["NMN_RVR"]
	network.ExtraProperties.Subnets = nil

	// Allocate a subnet into a network without subnets
	subnet, err := AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 3000}, nil)
	suite.NoError(err)

	expectedSubnet := sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_3000",
		CIDR:      "10.106.0.0/22",
		Gateway:   "10.106.0.1",
		DHCPStart: "10.106.0.10",
		DHCPEnd:   "10.106.3.254",
		VlanID:    1770,
	}
	suite.Equal(expectedSubnet, subnet)

	// Add a second cabinet subnet
	network.ExtraProperties.Subnets = append(network.ExtraProperties.Subnets, expectedSubnet)

	subnet, err = AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 3001}, nil)
	suite.NoError(err)
	expectedSubnet = sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_3001",
		CIDR:      "10.106.4.0/22",
		Gateway:   "10.106.4.1",
		DHCPStart: "10.106.4.10",
		DHCPEnd:   "10.106.7.254",
		VlanID:    1771,
	}
	suite.Equal(expectedSubnet, subnet)
}

func (suite *AllocateCabinetSubnetSuite) TestAllocateHMN_RVR() {
	network := suite.slsNetworks["HMN_RVR"]
	network.ExtraProperties.Subnets = nil

	// Allocate a subnet into a network without subnets
	subnet, err := AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 3000}, nil)
	suite.NoError(err)

	expectedSubnet := sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_3000",
		CIDR:      "10.107.0.0/22",
		Gateway:   "10.107.0.1",
		DHCPStart: "10.107.0.10",
		DHCPEnd:   "10.107.3.254",
		VlanID:    1513,
	}
	suite.Equal(expectedSubnet, subnet)

	// Add a second cabinet subnet
	network.ExtraProperties.Subnets = append(network.ExtraProperties.Subnets, expectedSubnet)

	subnet, err = AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 3001}, nil)
	suite.NoError(err)
	expectedSubnet = sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_3001",
		CIDR:      "10.107.4.0/22",
		Gateway:   "10.107.4.1",
		DHCPStart: "10.107.4.10",
		DHCPEnd:   "10.107.7.254",
		VlanID:    1514,
	}
	suite.Equal(expectedSubnet, subnet)
}

func (suite *AllocateCabinetSubnetSuite) TestAllocateNMN_RVR_VlanOverride() {
	network := suite.slsNetworks["NMN_RVR"]
	network.ExtraProperties.Subnets = nil

	// Allocate a subnet into a network without subnets
	vlanOverride := int32(1234)
	subnet, err := AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 3000}, &vlanOverride)
	suite.NoError(err)

	expectedSubnet := sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_3000",
		CIDR:      "10.106.0.0/22",
		Gateway:   "10.106.0.1",
		DHCPStart: "10.106.0.10",
		DHCPEnd:   "10.106.3.254",
		VlanID:    1234,
	}
	suite.Equal(expectedSubnet, subnet)

	// Add a second cabinet subnet
	network.ExtraProperties.Subnets = append(network.ExtraProperties.Subnets, expectedSubnet)

	vlanOverride = int32(2234)
	subnet, err = AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 3001}, &vlanOverride)
	suite.NoError(err)
	expectedSubnet = sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_3001",
		CIDR:      "10.106.4.0/22",
		Gateway:   "10.106.4.1",
		DHCPStart: "10.106.4.10",
		DHCPEnd:   "10.106.7.254",
		VlanID:    2234,
	}
	suite.Equal(expectedSubnet, subnet)
}

func (suite *AllocateCabinetSubnetSuite) TestAllocateHMN_RVR_VlanOverride() {
	network := suite.slsNetworks["HMN_RVR"]
	network.ExtraProperties.Subnets = nil

	// Allocate a subnet into a network without subnets
	vlanOverride := int32(1234)
	subnet, err := AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 3000}, &vlanOverride)
	suite.NoError(err)

	expectedSubnet := sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_3000",
		CIDR:      "10.107.0.0/22",
		Gateway:   "10.107.0.1",
		DHCPStart: "10.107.0.10",
		DHCPEnd:   "10.107.3.254",
		VlanID:    1234,
	}
	suite.Equal(expectedSubnet, subnet)

	// Add a second cabinet subnet
	network.ExtraProperties.Subnets = append(network.ExtraProperties.Subnets, expectedSubnet)

	vlanOverride = int32(2234)
	subnet, err = AllocateCabinetSubnet(network.Name, *network.ExtraProperties, xnames.Cabinet{Cabinet: 3001}, &vlanOverride)
	suite.NoError(err)
	expectedSubnet = sls_client.NetworkIpv4Subnet{
		Name:      "cabinet_3001",
		CIDR:      "10.107.4.0/22",
		Gateway:   "10.107.4.1",
		DHCPStart: "10.107.4.10",
		DHCPEnd:   "10.107.7.254",
		VlanID:    2234,
	}
	suite.Equal(expectedSubnet, subnet)
}

func (suite *AllocateCabinetSubnetSuite) TestInvalidNetworkCIDR() {
	slsNetworkEP := sls_client.NetworkExtraProperties{
		CIDR: "not-a-cidr",
	}

	subnet, err := AllocateCabinetSubnet("HMN_MTN", slsNetworkEP, xnames.Cabinet{Cabinet: 1000}, nil)
	suite.Empty(subnet)
	expectedErrorStrings := []string{
		"failed to allocate cabinet subnet for (x1000) in CIDR (not-a-cidr)",
		"failed to parse network CIDR (not-a-cidr)",
		"netaddr.ParseIPPrefix(\"not-a-cidr\"): no '/'",
	}
	suite.EqualError(err, strings.Join(expectedErrorStrings, "\n"))
}

func (suite *AllocateCabinetSubnetSuite) TestSubnetAlreadyExists() {
	slsNetworkEP := sls_client.NetworkExtraProperties{
		CIDR: "10.0.0.0/16",
		Subnets: []sls_client.NetworkIpv4Subnet{
			{
				Name:      "cabinet_1000",
				CIDR:      "10.0.0.0/24",
				DHCPStart: "10.0.0.10",
				DHCPEnd:   "10.0.0.254",
				VlanID:    1234,
			},
		},
	}

	subnet, err := AllocateCabinetSubnet("HMN_MTN", slsNetworkEP, xnames.Cabinet{Cabinet: 1000}, nil)
	suite.Empty(subnet)
	suite.EqualError(err, "subnet (cabinet_1000) already exists")
}

func (suite *AllocateCabinetSubnetSuite) TestUnsupportedNetworkNames() {
	for _, name := range []string{"CAN", "CMN", "MTL", "HMNLB", "NMNLB", "BICAN", "IDK"} {
		slsNetworkEP := sls_client.NetworkExtraProperties{
			CIDR: "10.0.0.0/16",
		}

		subnet, err := AllocateCabinetSubnet(name, slsNetworkEP, xnames.Cabinet{Cabinet: 1000}, nil)
		suite.Empty(subnet)
		suite.EqualError(err, fmt.Sprintf("unsupported network (%s) unable to allocate vlan for cabinet subnet", name))
	}
}

func (suite *AllocateCabinetSubnetSuite) TestVLANSpaceExhausted() {
	slsNetworkEP := sls_client.NetworkExtraProperties{
		CIDR: "10.0.0.0/21",
		Subnets: []sls_client.NetworkIpv4Subnet{
			{
				Name:      "cabinet_1",
				CIDR:      "10.0.0.0/22",
				DHCPStart: "10.0.0.10",
				DHCPEnd:   "10.0.3.254",
				VlanID:    1234,
			},
			{
				Name:      "cabinet_2",
				CIDR:      "10.0.4.0/22",
				DHCPStart: "10.0.4.10",
				DHCPEnd:   "10.0.4.254",
				VlanID:    1234,
			},
		},
	}

	subnet, err := AllocateCabinetSubnet("HMN_MTN", slsNetworkEP, xnames.Cabinet{Cabinet: 1000}, nil)
	suite.Empty(subnet)

	expectedErrorStrings := []string{
		"failed to allocate cabinet subnet for (x1000) in CIDR (10.0.0.0/21)",
		"network space has been exhausted",
	}
	suite.EqualError(err, strings.Join(expectedErrorStrings, "\n"))
}

func TestAllocateCabinetSubnetSuite(t *testing.T) {
	suite.Run(t, new(AllocateCabinetSubnetSuite))
}

//
// AllocateIPSuite
//

type AllocateIPSuite struct {
	suite.Suite

	slsNetworks map[string]sls_client.Network
}

func (suite *AllocateIPSuite) SetupSuite() {
	// Load SLS state
	slsStateRaw, err := ioutil.ReadFile(testSLSFile)
	suite.NoError(err)

	var slsState sls_client.SlsState
	err = json.Unmarshal(slsStateRaw, &slsState)
	suite.NoError(err)

	suite.slsNetworks = slsState.Networks
}

func (suite *AllocateIPSuite) TestAllocateIPInStaticSubnet() {
	subnet := sls_client.NetworkIpv4Subnet{
		Name:    "test_subnet",
		CIDR:    "10.0.0.0/24",
		Gateway: "10.0.0.1",
		IPReservations: []sls_client.NetworkIpReservation{
			{Name: "foo", IPAddress: "10.0.0.2"},
			{Name: "bar", IPAddress: "10.0.0.3"},
		},
	}
	network := sls_client.Network{
		Name: "test_network",
		ExtraProperties: &sls_client.NetworkExtraProperties{
			CIDR:    "10.0.0.0/22",
			Subnets: []sls_client.NetworkIpv4Subnet{subnet},
		},
	}

	ipReservation, err := AllocateIP(network, subnet, xnames.FromString("x3000c0w15"), "sw-leaf-bmc-001")
	suite.NoError(err)

	expectedIPReservation := sls_client.NetworkIpReservation{
		Comment:   "x3000c0w15",
		IPAddress: "10.0.0.4",
		Name:      "sw-leaf-bmc-001",
	}
	suite.Equal(expectedIPReservation, ipReservation)
}

func (suite *AllocateIPSuite) TestAllocateIPInSubnetWithDHCPRange() {
	subnet := sls_client.NetworkIpv4Subnet{
		Name:      "test_subnet",
		CIDR:      "10.0.0.0/24",
		Gateway:   "10.0.0.1",
		DHCPStart: "10.0.0.10",
		DHCPEnd:   "10.0.0.99",
		IPReservations: []sls_client.NetworkIpReservation{
			{Name: "foo", IPAddress: "10.0.0.2"},
			{Name: "bar", IPAddress: "10.0.0.3"},
		},
	}
	network := sls_client.Network{
		Name: "test_network",
		ExtraProperties: &sls_client.NetworkExtraProperties{
			CIDR:    "10.0.0.0/22",
			Subnets: []sls_client.NetworkIpv4Subnet{subnet},
		},
	}

	ipReservation, err := AllocateIP(network, subnet, xnames.FromString("x3000c0w15"), "sw-leaf-bmc-001")
	suite.NoError(err)

	expectedIPReservation := sls_client.NetworkIpReservation{
		Comment:   "x3000c0w15",
		IPAddress: "10.0.0.4",
		Name:      "sw-leaf-bmc-001",
	}
	suite.Equal(expectedIPReservation, ipReservation)
}

func (suite *AllocateIPSuite) TestFullStaticSubnet() {
	subnet := sls_client.NetworkIpv4Subnet{
		Name:           "test_subnet",
		CIDR:           "10.0.0.0/24",
		Gateway:        "10.0.0.1",
		IPReservations: []sls_client.NetworkIpReservation{},
	}

	// Fill subnet with IP address reservations
	for i := 2; i < 255; i++ {
		subnet.IPReservations = append(subnet.IPReservations, sls_client.NetworkIpReservation{
			Name:      fmt.Sprintf("reservation_%d", i),
			IPAddress: fmt.Sprintf("10.0.0.%d", i),
		})
	}

	network := sls_client.Network{
		Name: "test_network",
		ExtraProperties: &sls_client.NetworkExtraProperties{
			CIDR:    "10.0.0.0/22",
			Subnets: []sls_client.NetworkIpv4Subnet{subnet},
		},
	}

	ipAddress, err := FindNextAvailableIP(network, subnet)
	suite.EqualError(err, "subnet has no available IPs")
	suite.Equal(netaddr.IP{}, ipAddress)
}

func (suite *AllocateIPSuite) TestIPReservationAlreadyExists() {
	subnet := sls_client.NetworkIpv4Subnet{
		Name:    "test_subnet",
		CIDR:    "10.0.0.0/24",
		Gateway: "10.0.0.1",
		IPReservations: []sls_client.NetworkIpReservation{
			{Name: "foo", IPAddress: "10.0.0.2"},
			{Name: "bar", IPAddress: "10.0.0.3"},
			{Name: "sw-leaf-bmc-001", IPAddress: "10.0.0.4", Comment: "x3000c0w15"},
		},
	}
	network := sls_client.Network{
		Name: "test_network",
		ExtraProperties: &sls_client.NetworkExtraProperties{
			CIDR:    "10.0.0.0/22",
			Subnets: []sls_client.NetworkIpv4Subnet{subnet},
		},
	}

	ipReservation, err := AllocateIP(network, subnet, xnames.FromString("x3000c0w15"), "sw-leaf-bmc-001")
	suite.Empty(ipReservation)
	suite.EqualError(err, "ip reservation with name (sw-leaf-bmc-001) and xname (x3000c0w15) already exists with IP (10.0.0.4)")

}

func (suite *AllocateIPSuite) TestDuplicateAlias() {
	subnet := sls_client.NetworkIpv4Subnet{
		Name:    "test_subnet",
		CIDR:    "10.0.0.0/24",
		Gateway: "10.0.0.1",
		IPReservations: []sls_client.NetworkIpReservation{
			{Name: "foo", IPAddress: "10.0.0.2"},
			{Name: "bar", IPAddress: "10.0.0.3"},
			{Name: "sw-leaf-bmc-001", IPAddress: "10.0.0.4", Comment: "x3000c0w15"},
		},
	}

	network := sls_client.Network{
		Name: "test_network",
		ExtraProperties: &sls_client.NetworkExtraProperties{
			CIDR:    "10.0.0.0/22",
			Subnets: []sls_client.NetworkIpv4Subnet{subnet},
		},
	}

	ipReservation, err := AllocateIP(network, subnet, xnames.FromString("x3000c0w30"), "sw-leaf-bmc-001")
	suite.Empty(ipReservation)
	suite.EqualError(err, "ip reservation with name (sw-leaf-bmc-001) already exists on (x3000c0w15) with IP (10.0.0.4)")
}

func (suite *AllocateIPSuite) TestDuplicateXname() {
	subnet := sls_client.NetworkIpv4Subnet{
		Name:    "test_subnet",
		CIDR:    "10.0.0.0/24",
		Gateway: "10.0.0.1",
		IPReservations: []sls_client.NetworkIpReservation{
			{Name: "foo", IPAddress: "10.0.0.2"},
			{Name: "bar", IPAddress: "10.0.0.3"},
			{Name: "sw-leaf-bmc-001", IPAddress: "10.0.0.4", Comment: "x3000c0w15"},
		},
	}

	network := sls_client.Network{
		Name: "test_network",
		ExtraProperties: &sls_client.NetworkExtraProperties{
			CIDR:    "10.0.0.0/22",
			Subnets: []sls_client.NetworkIpv4Subnet{subnet},
		},
	}

	ipReservation, err := AllocateIP(network, subnet, xnames.FromString("x3000c0w15"), "sw-leaf-bmc-002")
	suite.Empty(ipReservation)
	suite.EqualError(err, "ip reservation with xname (x3000c0w15) already exists with name (sw-leaf-bmc-001) with IP (10.0.0.4)")
}

func (suite *AllocateIPSuite) TestFullStaticRange() {
	subnet := sls_client.NetworkIpv4Subnet{
		Name:      "test_subnet",
		CIDR:      "10.0.0.0/24",
		Gateway:   "10.0.0.1",
		DHCPStart: "10.0.0.10",
		DHCPEnd:   "10.0.0.99",
	}

	// Fill subnet with IP address reservations
	for i := 2; i < 10; i++ {
		subnet.IPReservations = append(subnet.IPReservations, sls_client.NetworkIpReservation{
			Name:      fmt.Sprintf("reservation_%d", i),
			IPAddress: fmt.Sprintf("10.0.0.%d", i),
		})
	}

	network := sls_client.Network{
		Name: "test_network",
		ExtraProperties: &sls_client.NetworkExtraProperties{
			CIDR:    "10.0.0.0/22",
			Subnets: []sls_client.NetworkIpv4Subnet{subnet},
		},
	}

	ipAddress, err := AllocateIP(network, subnet, xnames.FromString("x3000c0w15"), "sw-leaf-bmc-002")
	suite.EqualError(err, "ip reservation with xname (x3000c0w15) and IP 10.0.0.10 is outside the static IP address range, with starting DHCP IP of 10.0.0.10")
	suite.Empty(ipAddress)
}

func (suite *AllocateIPSuite) TestInvalidDHCPStart() {
	subnet := sls_client.NetworkIpv4Subnet{
		Name:      "test_subnet",
		CIDR:      "10.0.0.0/24",
		Gateway:   "10.0.0.1",
		DHCPStart: "invalid-ip",
		DHCPEnd:   "10.0.0.99",
	}

	network := sls_client.Network{
		Name: "test_network",
		ExtraProperties: &sls_client.NetworkExtraProperties{
			CIDR:    "10.0.0.0/22",
			Subnets: []sls_client.NetworkIpv4Subnet{subnet},
		},
	}

	ipAddress, err := AllocateIP(network, subnet, xnames.FromString("x3000c0w15"), "sw-leaf-bmc-002")
	suite.Empty(ipAddress)
	expectedErrorStrings := []string{
		"failed to parse DHCP Start IP (invalid-ip) address",
		"ParseIP(\"invalid-ip\"): unable to parse IP",
	}
	suite.EqualError(err, strings.Join(expectedErrorStrings, "\n"))
}

func (suite *AllocateIPSuite) TestAllocateIP_HMN_BootstrapDHCP() {
	// TODO This test is failing due to super net hack
	network := suite.slsNetworks["HMN"]
	subnet, _, err := sls.LookupSubnet(network, "bootstrap_dhcp")
	suite.NoError(err)

	ipReservation, err := AllocateIP(network, subnet, xnames.FromString("x3000c0w15"), "sw-leaf-bmc-100")
	suite.Empty(ipReservation)
	suite.EqualError(err, "ip reservation with xname (x3000c0w15) and IP 10.254.1.41 is outside the static IP address range, with starting DHCP IP of 10.254.1.41")

	// TODO The following would be true if the static range of the subnet was expanded
	// expectedIPReservation := sls_client.NetworkIpReservation{
	// 	Comment:   "x3000c0w15",
	// 	IPAddress: "10.254.1.41",
	// 	Name:      "sw-leaf-bmc-100",
	// }
	// suite.Equal(expectedIPReservation, ipReservation)
}

func (suite *AllocateIPSuite) TestAllocateIP_HMN_NetworkHardware() {
	network := suite.slsNetworks["HMN"]
	subnet, _, err := sls.LookupSubnet(network, "network_hardware")
	suite.NoError(err)

	ipReservation, err := AllocateIP(network, subnet, xnames.FromString("x3000c0w15"), "sw-leaf-bmc-100")
	suite.NoError(err)

	expectedIPReservation := sls_client.NetworkIpReservation{
		Comment:   "x3000c0w15",
		IPAddress: "10.254.0.5",
		Name:      "sw-leaf-bmc-100",
	}
	suite.Equal(expectedIPReservation, ipReservation)
}

func (suite *AllocateIPSuite) TestAllocateIP_NMN_BootstrapDHCP() {
	// TODO This test is failing due to super net hack
	network := suite.slsNetworks["NMN"]
	subnet, _, err := sls.LookupSubnet(network, "bootstrap_dhcp")
	suite.NoError(err)

	ipReservation, err := AllocateIP(network, subnet, xnames.FromString("x3000c0w15"), "sw-leaf-bmc-100")
	suite.Empty(ipReservation)
	suite.EqualError(err, "ip reservation with xname (x3000c0w15) and IP 10.252.1.23 is outside the static IP address range, with starting DHCP IP of 10.252.1.23")

	// TODO The following would be true if the static range of the subnet was expanded
	// expectedIPReservation := sls_client.NetworkIpReservation{
	// 	Comment:   "x3000c0w15",
	// 	IPAddress: "10.252.1.23",
	// 	Name:      "sw-leaf-bmc-100",
	// }
	// suite.Equal(expectedIPReservation, ipReservation)
}

func (suite *AllocateIPSuite) TestAllocateIP_NMN_NetworkHardware() {
	network := suite.slsNetworks["NMN"]
	subnet, _, err := sls.LookupSubnet(network, "network_hardware")
	suite.NoError(err)

	ipReservation, err := AllocateIP(network, subnet, xnames.FromString("x3000c0w15"), "sw-leaf-bmc-100")
	suite.NoError(err)

	expectedIPReservation := sls_client.NetworkIpReservation{
		Comment:   "x3000c0w15",
		IPAddress: "10.252.0.5",
		Name:      "sw-leaf-bmc-100",
	}
	suite.Equal(expectedIPReservation, ipReservation)
}

func (suite *AllocateIPSuite) TestAllocateIP_CMN_BootstrapDHCP() {
	// TODO This test is failing due to super net hack
	network := suite.slsNetworks["CMN"]
	subnet, _, err := sls.LookupSubnet(network, "bootstrap_dhcp")
	suite.NoError(err)

	ipReservation, err := AllocateIP(network, subnet, xnames.FromString("x3000c0w15"), "sw-leaf-bmc-100")
	suite.Empty(ipReservation)
	expectedErrorStrings := []string{
		"failed to allocate ip for hardware (x3000c0w15) in subnet (10.102.162.16/25)",
		"failed to detect supernet hack on subnet bootstrap_dhcp in network CMN",
		"allocating an IP address on the (CMN) network is not currently supported",
	}
	suite.EqualError(err, strings.Join(expectedErrorStrings, "\n"))

	// Currently FindNextAvailableIP does not support allocate IPs on the CMN
	// suite.NoError(err)
	// expectedIPReservation := sls_client.NetworkIpReservation{
	// 	Comment:   "x3000c0w15",
	// 	IPAddress: "10.102.162.38",
	// 	Name:      "sw-leaf-bmc-100",
	// }
	// suite.Equal(expectedIPReservation, ipReservation)
}

func (suite *AllocateIPSuite) TestAllocateIP_CMN_NetworkHardware() {
	network := suite.slsNetworks["CMN"]
	subnet, _, err := sls.LookupSubnet(network, "network_hardware")
	suite.NoError(err)

	ipReservation, err := AllocateIP(network, subnet, xnames.FromString("x3000c0w15"), "sw-leaf-bmc-100")
	suite.Empty(ipReservation)
	expectedErrorStrings := []string{
		"failed to allocate ip for hardware (x3000c0w15) in subnet (10.102.162.0/25)",
		"failed to detect supernet hack on subnet network_hardware in network CMN",
		"allocating an IP address on the (CMN) network is not currently supported",
	}
	suite.EqualError(err, strings.Join(expectedErrorStrings, "\n"))

	// Currently FindNextAvailableIP does not support allocate IPs on the CMN
	// suite.NoError(err)
	// expectedIPReservation := sls_client.NetworkIpReservation{
	// 	Comment:   "x3000c0w15",
	// 	IPAddress: "10.102.162.5",
	// 	Name:      "sw-leaf-bmc-100",
	// }
	// suite.Equal(expectedIPReservation, ipReservation)
}

func TestAllocateIPSuite(t *testing.T) {
	suite.Run(t, new(AllocateIPSuite))
}

// //
// // FreeIPsInStaticRangeSuite
// //
//
// // TODO
//
// type FreeIPsInStaticRangeSuite struct {
// 	suite.Suite
// }
//
// func TestFreeIPsInStaticRangeSuite(t *testing.T) {
// 	suite.Run(t, new(FreeIPsInStaticRangeSuite))
// }
//
// //
// // ExpandSubnetStaticRangeSuite
// //
//
// // TODO
//
// type ExpandSubnetStaticRangeSuite struct {
// 	suite.Suite
// }
//
// func TestExpandSubnetStaticRangeSuite(t *testing.T) {
// 	suite.Run(t, new(ExpandSubnetStaticRangeSuite))
// }
