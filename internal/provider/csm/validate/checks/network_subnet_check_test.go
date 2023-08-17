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

package checks

import (
	"net"
	"testing"

	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

func TestParseAndCheckCidr(t *testing.T) {
	componentId := "/Networks/NMN_RVR"
	networkName := "NMN_RVR"

	// test empty string
	results := common.NewValidationResults()
	cidr := ""
	ipNet := parseAndCheckCidr(results, componentId, networkName, cidr)
	if ipNet != nil {
		t.Errorf("Expected nil for empty CIDR, CIDR net: %v", ipNet)
	}
	if len(results.GetResults()) != 0 {
		t.Errorf("Expected empty results for an empty CIDR, results: %v", results.GetResults())
	}

	// test bad cidr
	results = common.NewValidationResults()
	cidr = "junk"
	ipNet = parseAndCheckCidr(results, componentId, networkName, cidr)
	if ipNet != nil {
		t.Errorf("Expected nil for bad  CIDR: %s, CIDR net: %v", cidr, ipNet)
	}
	if len(results.GetResults()) != 1 {
		t.Errorf("Expected one result for an bad CIDR: %s, results: %v", cidr, results.GetResults())
	}

	// test good cidr
	results = common.NewValidationResults()
	cidr = "10.106.0.0/22"
	ipNet = parseAndCheckCidr(results, componentId, networkName, cidr)
	if ipNet == nil {
		t.Errorf("Expected nil for good CIDR, CIDR: %s, net: %v", cidr, ipNet)
	}
	if len(results.GetResults()) != 0 {
		t.Errorf("Expected one result for an good CIDR: %s results: %v", cidr, results.GetResults())
	}
}

func TestCheckCidrConsistancy(t *testing.T) {
	// test matching cidr
	results := common.NewValidationResults()
	networkName := "HMN_MTN"
	componentId := "/Networks/" + networkName
	networkCidr := "10.106.0.0/22"
	subnetCidr := "10.106.0.0/22"
	checkCidrConsistancy(results, componentId, networkName, networkCidr, subnetCidr)
	if len(results.GetResults()) != 1 {
		t.Fatalf(
			"Expected one result for name: %s, networkCidr: %s, subnetCidr: %s, results: %v",
			networkName, networkCidr, subnetCidr, results.GetResults())
	}
	if results.GetResults()[0].Result != common.Pass {
		t.Errorf(
			"Expected passing result for name: %s, networkCidr: %s, subnetCidr: %s, result: %v",
			networkName, networkCidr, subnetCidr, results.GetResults()[0])
	}

	// test cidrs that don't match
	results = common.NewValidationResults()
	networkName = "HMN_MTN"
	componentId = "/Networks/" + networkName
	networkCidr = "10.106.0.0/22"
	subnetCidr = "10.106.0.0/21"
	checkCidrConsistancy(results, componentId, networkName, networkCidr, subnetCidr)
	if len(results.GetResults()) != 1 {
		t.Fatalf(
			"Expected one result for name: %s, networkCidr: %s, subnetCidr: %s, results: %v",
			networkName, networkCidr, subnetCidr, results.GetResults())
	}
	if results.GetResults()[0].Result != common.Fail {
		t.Errorf(
			"Expected failing result for name: %s, networkCidr: %s, subnetCidr: %s, result: %v",
			networkName, networkCidr, subnetCidr, results.GetResults()[0])
	}

	// test network type that is not checked by the checker
	results = common.NewValidationResults()
	networkName = "NMNLB"
	componentId = "/Networks/" + networkName
	networkCidr = "10.106.0.0/22"
	subnetCidr = "10.106.0.0/22"
	checkCidrConsistancy(results, componentId, networkName, networkCidr, subnetCidr)
	if len(results.GetResults()) != 0 {
		t.Errorf(
			"Expected zero results for name: %s, networkCidr: %s, subnetCidr: %s, results: %v",
			networkName, networkCidr, subnetCidr, results.GetResults())
	}
}

func TestCheckVlanUniqueness(t *testing.T) {
	// test with unique vlans
	vlan := "1000"
	vlan2 := "1001"
	subnet := newNetworkSubnetWithVlan(vlan)
	subnet2 := newNetworkSubnetWithVlan(vlan2)
	vlans := make(map[string][]*NetworkSubnet)
	vlans[vlan] = append(vlans[vlan], subnet)
	vlans[vlan2] = append(vlans[vlan2], subnet2)
	results := common.NewValidationResults()
	checkVlanUniqueness(results, vlans, subnet)
	if len(results.GetResults()) != 0 {
		t.Fatalf(
			"Expected zero results for vlan: %s, results: %v",
			vlan, results.GetResults())
	}

	// test with conflicting vlans
	vlan = "2000"
	subnet = newNetworkSubnetWithVlan(vlan)
	subnet2 = newNetworkSubnetWithVlan(vlan)
	vlans = make(map[string][]*NetworkSubnet)
	vlans[vlan] = append(vlans[vlan], subnet)
	vlans[vlan] = append(vlans[vlan], subnet2)
	results = common.NewValidationResults()
	checkVlanUniqueness(results, vlans, subnet)
	if len(results.GetResults()) != 1 {
		t.Fatalf(
			"Expected one result for conflicting vlans. vlan: %s, results: %v",
			vlan, results.GetResults())
	}
	if results.GetResults()[0].Result != common.Fail {
		t.Errorf(
			"Expected failing result for conflicting vlans. vlan: %s, result: %v",
			vlan, results.GetResults()[0])
	}

	// test with the vlan not in the map
	vlan = "3000"
	vlan2 = "3001"
	subnet = newNetworkSubnetWithVlan(vlan)
	subnet2 = newNetworkSubnetWithVlan(vlan2)
	vlans = make(map[string][]*NetworkSubnet)
	vlans[vlan2] = append(vlans[vlan2], subnet2)
	results = common.NewValidationResults()
	checkVlanUniqueness(results, vlans, subnet)
	if len(results.GetResults()) != 0 {
		t.Fatalf(
			"Expected zero results for no mapped vlans. vlan: %s, results: %v",
			vlan, results.GetResults())
	}
}

func TestCheckNameUniqueness(t *testing.T) {
	// test with unique names
	name := "cabinet_1000"
	name2 := "cabinet_1001"
	subnet := newNetworkSubnetWithName(name)
	subnet2 := newNetworkSubnetWithName(name2)
	names := make(map[string][]*NetworkSubnet)
	names[name] = append(names[name], subnet)
	names[name2] = append(names[name2], subnet2)
	results := common.NewValidationResults()
	checkNameUniqueness(results, names, subnet)
	if len(results.GetResults()) != 0 {
		t.Fatalf(
			"Expected zero results for name: %s, results: %v",
			name, results.GetResults())
	}

	// test with conflicting names
	name = "cabinet_2000"
	subnet = newNetworkSubnetWithName(name)
	subnet2 = newNetworkSubnetWithName(name)
	names = make(map[string][]*NetworkSubnet)
	names[name] = append(names[name], subnet)
	names[name] = append(names[name], subnet2)
	results = common.NewValidationResults()
	checkNameUniqueness(results, names, subnet)
	if len(results.GetResults()) != 1 {
		t.Fatalf(
			"Expected one result for conflicting names. name: %s, results: %v",
			name, results.GetResults())
	}
	if results.GetResults()[0].Result != common.Fail {
		t.Errorf(
			"Expected failing result for conflicting names. name: %s, result: %v",
			name, results.GetResults()[0])
	}

	// test with the name not in the map
	name = "cabinet_3000"
	name2 = "cabinet_3001"
	subnet = newNetworkSubnetWithName(name)
	subnet2 = newNetworkSubnetWithName(name2)
	names = make(map[string][]*NetworkSubnet)
	names[name2] = append(names[name2], subnet2)
	results = common.NewValidationResults()
	checkNameUniqueness(results, names, subnet)
	if len(results.GetResults()) != 0 {
		t.Fatalf(
			"Expected zero results for no mapped names. name: %s, results: %v",
			name, results.GetResults())
	}
}

func TestCheckSubnetCidr(t *testing.T) {
	// test with good cidrs
	cidr := "10.100.0.0/17"
	subnetCidr := "10.100.0.0/22"
	subnet, subnetCidrIpNet := newNetworkSubnetWithCidr(t, cidr, subnetCidr)
	results := common.NewValidationResults()
	checkSubnetCidr(results, subnet, subnetCidrIpNet)
	if len(results.GetResults()) != 0 {
		t.Fatalf(
			"Expected zero results for cidr check. cidr: %s, subnet cidr: %s, results: %v",
			cidr, subnetCidr, results.GetResults())
	}

	// test with bad cidrs
	cidr = "10.100.0.0/17"
	subnetCidr = "10.106.0.0/22"
	subnet, subnetCidrIpNet = newNetworkSubnetWithCidr(t, cidr, subnetCidr)
	results = common.NewValidationResults()
	checkSubnetCidr(results, subnet, subnetCidrIpNet)
	if len(results.GetResults()) != 1 {
		t.Fatalf(
			"Expected one result for cidr check. cidr: %s, subnet cidr: %s, results: %v",
			cidr, subnetCidr, results.GetResults())
	}
	if results.GetResults()[0].Result != common.Fail {
		t.Errorf(
			"Expected failing result for cidr check. cidr: %s, subnet cidr: %s, result: %v",
			cidr, subnetCidr, results.GetResults()[0])
	}
}

func TestCheckSubnetStartAndEnd(t *testing.T) {
	// test with start and end IPs in order
	subnetCidr := "10.100.0.0/22"
	startIp := "10.100.0.10"
	endIp := "10.100.3.254"
	subnet, subnetCidrIpNet := newNetworkSubnetWithDhcpStartAndEnd(t, subnetCidr, startIp, endIp)
	results := common.NewValidationResults()
	checkDhcpStartAndEnd(results, subnet, subnetCidrIpNet)
	if len(results.GetResults()) != 0 {
		t.Fatalf(
			"Expected zero results for start and end ip check. subnet cidr: %s, startIp: %s, endIp: %s, results: %v",
			subnetCidr, startIp, endIp, results.GetResults())
	}

	// test with out of order IPs
	subnetCidr = "10.100.0.0/22"
	startIp = "10.100.3.254"
	endIp = "10.100.0.10"
	subnet, subnetCidrIpNet = newNetworkSubnetWithDhcpStartAndEnd(t, subnetCidr, startIp, endIp)
	results = common.NewValidationResults()
	checkDhcpStartAndEnd(results, subnet, subnetCidrIpNet)
	if len(results.GetResults()) != 1 {
		t.Fatalf(
			"Expected one result for start and end ip check. subnet cidr: %s, startIp: %s, endIp: %s, results: %v",
			subnetCidr, startIp, endIp, results.GetResults())
	}
	if results.GetResults()[0].Result != common.Fail {
		t.Errorf(
			"Expected failing result for start and end ip check. subnet cidr: %s, startIp:%s, endIp: %s, result: %v",
			subnetCidr, startIp, endIp, results.GetResults()[0])
	}

	// test with bad end IP
	subnetCidr = "10.100.0.0/22"
	startIp = "10.100.3.254"
	endIp = ""
	subnet, subnetCidrIpNet = newNetworkSubnetWithDhcpStartAndEnd(t, subnetCidr, startIp, endIp)
	results = common.NewValidationResults()
	checkDhcpStartAndEnd(results, subnet, subnetCidrIpNet)
	if len(results.GetResults()) != 2 {
		t.Fatalf(
			"Expected two results for start and end ip check. subnet cidr: %s, startIp: %s, endIp: %s, results: %v",
			subnetCidr, startIp, endIp, results.GetResults())
	}
	if results.GetResults()[0].Result != common.Fail || results.GetResults()[1].Result != common.Fail {
		t.Errorf(
			"Expected failing result for start and end ip check. subnet cidr: %s, startIp:%s, endIp: %s, result0: %v, result1: %v",
			subnetCidr, startIp, endIp, results.GetResults()[0], results.GetResults()[1])
	}

	// test with bad start IP
	subnetCidr = "10.100.0.0/22"
	startIp = ""
	endIp = "10.100.3.254"
	subnet, subnetCidrIpNet = newNetworkSubnetWithDhcpStartAndEnd(t, subnetCidr, startIp, endIp)
	results = common.NewValidationResults()
	checkDhcpStartAndEnd(results, subnet, subnetCidrIpNet)
	if len(results.GetResults()) != 1 {
		t.Fatalf(
			"Expected one result for start and end ip check. subnet cidr: %s, startIp: %s, endIp: %s, results: %v",
			subnetCidr, startIp, endIp, results.GetResults())
	}
	if results.GetResults()[0].Result != common.Fail {
		t.Errorf(
			"Expected failing result for start and end ip check. subnet cidr: %s, startIp:%s, endIp: %s, result: %v",
			subnetCidr, startIp, endIp, results.GetResults()[0])
	}

	// test with unparse-able IPs
	subnetCidr = "10.100.0.0/22"
	startIp = "start_junk"
	endIp = "end_junk"
	subnet, subnetCidrIpNet = newNetworkSubnetWithDhcpStartAndEnd(t, subnetCidr, startIp, endIp)
	results = common.NewValidationResults()
	checkDhcpStartAndEnd(results, subnet, subnetCidrIpNet)
	if len(results.GetResults()) != 2 {
		t.Fatalf(
			"Expected two results for start and end ip check. subnet cidr: %s, startIp: %s, endIp: %s, results: %v",
			subnetCidr, startIp, endIp, results.GetResults())
	}
	if results.GetResults()[0].Result != common.Fail || results.GetResults()[1].Result != common.Fail {
		t.Errorf(
			"Expected failing result for start and end ip check. subnet cidr: %s, startIp:%s, endIp: %s, result0: %v, result1: %v",
			subnetCidr, startIp, endIp, results.GetResults()[0], results.GetResults()[1])
	}
}

func newNetworkSubnetWithVlan(vlan string) *NetworkSubnet {
	network := sls_client.Network{
		Name: "Network_for_vlan_" + vlan,
	}
	return &NetworkSubnet{
		Network: &network,
		Vlan:    vlan,
	}
}

func newNetworkSubnetWithName(name string) *NetworkSubnet {
	network := sls_client.Network{
		Name: "Network_for_" + name,
	}
	return &NetworkSubnet{
		Network:    &network,
		SubnetName: name,
	}
}

func newNetworkSubnetWithCidr(t *testing.T, cidr string, subnetCidr string) (*NetworkSubnet, *net.IPNet) {
	network := sls_client.Network{
		Name: "Network_for_cidr_" + cidr,
	}

	_, cidrIpNet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf(
			"Expected cidr to be parseable. subnetCidr: %s, error: %v",
			cidr, err)
	}

	_, subnetCidrIpNet, err := net.ParseCIDR(subnetCidr)
	if err != nil {
		t.Fatalf(
			"Expected cidr to be parseable. subnetCidr: %s, error: %v",
			subnetCidr, err)
	}

	subnet := &NetworkSubnet{
		Network:    &network,
		Cidr:       cidr,
		CidrIpNet:  cidrIpNet,
		SubnetCidr: subnetCidr,
	}

	return subnet, subnetCidrIpNet
}

func newNetworkSubnetWithDhcpStartAndEnd(t *testing.T, subnetCidr, startIp, endIp string) (*NetworkSubnet, *net.IPNet) {
	network := sls_client.Network{
		Name: "Network_for_cidr_" + subnetCidr,
	}

	_, subnetCidrIpNet, err := net.ParseCIDR(subnetCidr)
	if err != nil {
		t.Fatalf(
			"Expected cidr to be parseable. subnetCidr: %s, error: %v",
			subnetCidr, err)
	}

	subnetProps := map[string]interface{}{
		"DHCPStart": startIp,
		"DHCPEnd":   endIp,
	}

	subnet := &NetworkSubnet{
		Network:    &network,
		Subnet:     subnetProps,
		SubnetCidr: subnetCidr,
	}

	return subnet, subnetCidrIpNet
}
