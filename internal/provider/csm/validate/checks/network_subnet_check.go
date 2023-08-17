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
	"bytes"
	"fmt"
	"net"

	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

const (
	CidrConsistent   common.ValidationCheck = "cidr-consistent"
	CidrValid        common.ValidationCheck = "cidr-valid"
	CidrContainsCidr common.ValidationCheck = "cidr-contains-other-cidr"
	CidrContainsIp   common.ValidationCheck = "cidr-contains-ip"
	IpStartAndEnd    common.ValidationCheck = "ip-start-and-end"
	VlanConflict     common.ValidationCheck = "vlan-conflict"
	NameConflict     common.ValidationCheck = "name-conflict"
)

type NetworkSubnet struct {
	Network     *sls_client.Network
	Props       map[string]interface{}
	Subnet      map[string]interface{}
	ComponentId string
	Cidr        string
	CidrIpNet   *net.IPNet
	SubnetCidr  string
	SubnetName  string
	Vlan        string
}

type NetworkSubnetCheck struct {
	slsStateExtended *common.SlsStateExtended
}

func NewNetworkSubnetCheck(slsStateExtended *common.SlsStateExtended) *NetworkSubnetCheck {
	networkSubnetCheck := NetworkSubnetCheck{
		slsStateExtended: slsStateExtended,
	}
	return &networkSubnetCheck
}

func (c *NetworkSubnetCheck) Validate(results *common.ValidationResults) {
	subnets := make([]NetworkSubnet, 0)
	vlans := make(map[string][]*NetworkSubnet)
	names := make(map[string][]*NetworkSubnet)
	for name, network := range c.slsStateExtended.SlsState.Networks {
		props, _ := common.GetMap(network.ExtraProperties)
		s, _ := common.GetSliceOfMaps(props["Subnets"])
		n := c.slsStateExtended.SlsState.Networks[name]
		cidr := ""
		if len(network.IPRanges) > 0 {
			cidr = network.IPRanges[0]
		}
		subnetCidr, _ := common.GetString(props, "CIDR")
		componentId := fmt.Sprintf("/Networks/%s", network.Name)
		checkCidrConsistancy(results, componentId, network.Name, cidr, subnetCidr)
		cidrIpNet := parseAndCheckCidr(results, componentId, network.Name, cidr)
		for _, subnet := range s {
			subnetName, _ := common.GetString(subnet, "Name")
			subnetCidr, _ := common.GetString(subnet, "CIDR")
			vlan, _ := common.GetString(subnet, "VlanID")
			ns := NetworkSubnet{
				Network:     &n,
				Props:       props,
				ComponentId: componentId,
				Cidr:        cidr,
				CidrIpNet:   cidrIpNet,
				Subnet:      subnet,
				SubnetName:  subnetName,
				SubnetCidr:  subnetCidr,
				Vlan:        vlan,
			}
			subnets = append(subnets, ns)
			vlans[vlan] = append(vlans[vlan], &ns)
			names[subnetName] = append(vlans[subnetName], &ns)
		}
	}

	for _, subnet := range subnets {
		switch subnet.Network.Name {
		case "HMN_MTN":
			fallthrough
		case "HMN_RVR":
			fallthrough
		case "NMN_MTN":
			fallthrough
		case "HNN_RVR":
			_, subnetCidrIpNet, err := net.ParseCIDR(subnet.SubnetCidr)
			if err != nil {
				results.Fail(
					CidrValid,
					subnet.ComponentId,
					fmt.Sprintf("Invalid CIDR: %s, details: %v.", subnet.SubnetCidr, err))
			}

			checkDhcpStartAndEnd(results, &subnet, subnetCidrIpNet)
			checkSubnetCidr(results, &subnet, subnetCidrIpNet)
			checkVlanUniqueness(results, vlans, &subnet)
			checkNameUniqueness(results, names, &subnet)
		}
	}
}

func parseAndCheckCidr(results *common.ValidationResults, componentId, networkName, cidr string) *net.IPNet {
	if cidr != "" {
		_, cidrIpNet, err := net.ParseCIDR(cidr)
		if err != nil {
			results.Fail(
				CidrValid,
				componentId,
				fmt.Sprintf("Invalid CIDR: %s, details: %v.", cidr, err))
		}
		return cidrIpNet
	}
	return nil
}

func checkCidrConsistancy(results *common.ValidationResults, componentId, networkName, networkCidr, subnetCidr string) {
	switch networkName {
	case "HMN_MTN":
		fallthrough
	case "HMN_RVR":
		fallthrough
	case "NMN_MTN":
		fallthrough
	case "HNN_RVR":
		if networkCidr == subnetCidr {
			results.Pass(
				CidrConsistent,
				componentId,
				fmt.Sprintf("In IPRanges the first value '%s' matches the CIDR '%s' in ExtraProperties.", networkCidr, subnetCidr))
		} else {
			results.Fail(
				CidrConsistent,
				componentId,
				fmt.Sprintf("In IPRanges the first value '%s' does not match the CIDR '%s' in ExtraProperties.", networkCidr, subnetCidr))
		}
	}
}

func checkVlanUniqueness(results *common.ValidationResults, vlans map[string][]*NetworkSubnet, subnet *NetworkSubnet) {
	if subnet.Vlan != "" {
		vlanSubnets := vlans[subnet.Vlan]
		if len(vlanSubnets) > 1 {
			description := descriptionOfNetworkCidrs(vlanSubnets, subnet.Network.Name)
			results.Fail(
				VlanConflict,
				subnet.ComponentId,
				fmt.Sprintf("Vlan %s conflicts with a vlan in %s.", subnet.Vlan, description))
		}
	}
}

func checkNameUniqueness(results *common.ValidationResults, names map[string][]*NetworkSubnet, subnet *NetworkSubnet) {
	if subnet.SubnetName != "" {
		nameSubnets := names[subnet.SubnetName]
		if len(nameSubnets) > 1 {
			description := descriptionOfNetworkCidrs(nameSubnets, subnet.Network.Name)
			results.Fail(
				NameConflict,
				subnet.ComponentId,
				fmt.Sprintf("Subnet name %s conflicts with the names of these subnets %s.", subnet.Cidr, description))
		}
	}
}

// check that the subnet CIDR is contained in the ExtraProperties CIDR
func checkSubnetCidr(results *common.ValidationResults, subnet *NetworkSubnet, subnetCidrIpNet *net.IPNet) {
	if subnet.CidrIpNet != nil && subnetCidrIpNet != nil {
		if !subnet.CidrIpNet.Contains(subnetCidrIpNet.IP) {
			results.Fail(
				CidrContainsCidr,
				subnet.ComponentId,
				fmt.Sprintf("CIDR: %s is not a sub CIDR of: %s.", subnet.SubnetCidr, subnet.Cidr))
		}
	}
}

func checkDhcpStartAndEnd(results *common.ValidationResults, subnet *NetworkSubnet, subnetCidrIpNet *net.IPNet) {
	dhcpStart, _ := common.GetString(subnet.Subnet, "DHCPStart")
	dhcpStartIp := net.ParseIP(dhcpStart)
	if !subnetCidrIpNet.Contains(dhcpStartIp) {
		results.Fail(
			CidrContainsCidr,
			subnet.ComponentId,
			fmt.Sprintf("DHCPStart %s is not CIDR %s.", dhcpStart, subnet.SubnetCidr))
	}

	dhcpEnd, _ := common.GetString(subnet.Subnet, "DHCPEnd")
	dhcpEndIp := net.ParseIP(dhcpEnd)
	if !subnetCidrIpNet.Contains(dhcpEndIp) {
		results.Fail(
			CidrContainsCidr,
			subnet.ComponentId,
			fmt.Sprintf("DHCPEnd %s is not CIDR %s.", dhcpEnd, subnet.SubnetCidr))
	}

	if bytes.Compare(dhcpStartIp.To16(), dhcpEndIp.To16()) > 0 {
		results.Fail(
			IpStartAndEnd,
			subnet.ComponentId,
			fmt.Sprintf("DHCPStart %s is after DHCPEnd %s.", dhcpStart, dhcpEnd))
	}
}

func descriptionOfNetworkCidrs(subnets []*NetworkSubnet, excludeNetwork ...string) string {
	description := ""
	haveAddedFirst := false
	for _, subnet := range subnets {
		if contains(subnet.Network.Name, excludeNetwork) {
			continue
		}
		if haveAddedFirst {
			description = description + ", "
		}
		description += fmt.Sprintf("%s %s", subnet.Network.Name, subnet.SubnetCidr)
		haveAddedFirst = true
	}
	return description
}
