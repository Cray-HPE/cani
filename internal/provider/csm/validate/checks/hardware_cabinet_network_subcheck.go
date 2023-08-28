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
	"fmt"

	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

type HardwareCabinetNetworkSubCheck struct {
	HmnRvr map[string]sls_client.NetworkIpv4Subnet
	HmnMtn map[string]sls_client.NetworkIpv4Subnet
	NmnRvr map[string]sls_client.NetworkIpv4Subnet
	NmnMtn map[string]sls_client.NetworkIpv4Subnet
}

func NewHardwareCabinetNetworkSubCheck(networks map[string]sls_client.Network) *HardwareCabinetNetworkSubCheck {
	subnets := &HardwareCabinetNetworkSubCheck{}

	hmnMtn, found := networks[HMN_MTN.String()]
	if found {
		subnets.HmnMtn = mapNetworkSubnets(&hmnMtn)
	} else {
		subnets.HmnMtn = make(map[string]sls_client.NetworkIpv4Subnet)
	}

	hmnRvr, found := networks[HMN_RVR.String()]
	if found {
		subnets.HmnRvr = mapNetworkSubnets(&hmnRvr)
	} else {
		subnets.HmnRvr = make(map[string]sls_client.NetworkIpv4Subnet)
	}

	nmnMtn, found := networks[NMN_MTN.String()]
	if found {
		subnets.NmnMtn = mapNetworkSubnets(&nmnMtn)
	} else {
		subnets.NmnMtn = make(map[string]sls_client.NetworkIpv4Subnet)
	}

	nmnRvr, found := networks[NMN_RVR.String()]
	if found {
		subnets.NmnRvr = mapNetworkSubnets(&nmnRvr)
	} else {
		subnets.NmnRvr = make(map[string]sls_client.NetworkIpv4Subnet)
	}

	return subnets
}

func (c *HardwareCabinetNetworkSubCheck) Validate(
	results *common.ValidationResults,
	hardware *sls_client.Hardware,
	props map[string]interface{}) {

	componentId := fmt.Sprintf("/Hardware/%s", hardware.Xname)
	networks, found := common.GetMap(props, "Networks")
	if !found {
		results.Fail(
			CabinetNetworkCheck,
			componentId,
			fmt.Sprintf("%s %s must have Networks defined", hardware.Xname, hardware.TypeString))
	}

	cnNetwork, foundCn := common.GetMap(networks, cn.String())
	if isRiver(hardware) {
		for n := range networks {
			if n != cn.String() && n != ncn.String() {
				results.Fail(
					CabinetNetworkCheck,
					componentId,
					fmt.Sprintf("%s %s invalid network catagory, %s. Allowed categories are %s and %s",
						hardware.Xname, hardware.TypeString, n, cn, ncn))
			}
		}
		if foundCn {
			validateCabinetNetwork(results, hardware, cn, c.HmnRvr, hmnRvrId, c.NmnRvr, nmnRvrId, cnNetwork)
		} else {
			results.Fail(
				CabinetNetworkCheck,
				componentId,
				fmt.Sprintf("%s %s missing the %s network", hardware.Xname, hardware.TypeString, cn))
		}
		ncnNetwork, foundNcn := common.GetMap(networks, ncn.String())
		if foundNcn {
			validateCabinetNetwork(results, hardware, ncn, c.HmnRvr, hmnRvrId, c.NmnRvr, nmnRvrId, ncnNetwork)
		} else {
			results.Fail(
				CabinetNetworkCheck,
				componentId,
				fmt.Sprintf("%s %s missing the %s network", hardware.Xname, hardware.TypeString, ncn))
		}
	} else {
		for n := range networks {
			if n != cn.String() {
				results.Fail(
					CabinetNetworkCheck,
					componentId,
					fmt.Sprintf("%s %s invalid network catagory, %s. The only allowed category is %s", hardware.Xname, hardware.TypeString, n, cn))
			}
		}
		if foundCn {
			validateCabinetNetwork(results, hardware, cn, c.HmnMtn, hmnMtnId, c.NmnMtn, nmnMtnId, cnNetwork)
		} else {
			results.Fail(
				CabinetNetworkCheck,
				componentId,
				fmt.Sprintf("%s %s missing required network %s", hardware.Xname, hardware.TypeString, cn))
		}
	}
}

// subnets of the network mapped from the cidr to the subnet info
func mapNetworkSubnets(network *sls_client.Network) map[string]sls_client.NetworkIpv4Subnet {
	m := make(map[string]sls_client.NetworkIpv4Subnet)
	for _, subnet := range network.ExtraProperties.Subnets {
		m[subnet.CIDR] = subnet
	}
	return m
}

func isRiver(hardware *sls_client.Hardware) bool {
	return hardware.Class == "River"
}

func validateCabinetNetwork(
	results *common.ValidationResults,
	hardware *sls_client.Hardware,
	cabinetNetworkName CabinetNetworkName,
	hmnSubnets map[string]sls_client.NetworkIpv4Subnet,
	hmnId NetworkId,
	nmnSubnets map[string]sls_client.NetworkIpv4Subnet,
	nmnId NetworkId,
	hardwareNetwork map[string]interface{}) {

	componentId := fmt.Sprintf("/Hardware/%s", hardware.Xname)

	// HMN
	hmn, found := common.GetMap(hardwareNetwork, HMN.String())
	if found {
		validateHardwareSubnetAgainstNetwork(results, hmn, hmnSubnets, componentId, cabinetNetworkName, HMN, hmnId)
	} else {
		results.Fail(
			CabinetNetworkCheck,
			componentId,
			fmt.Sprintf("%s %s missing the HMN network", hardware.Xname, hardware.TypeString))
	}

	// NMN
	nmn, found := common.GetMap(hardwareNetwork, NMN.String())
	if found {
		validateHardwareSubnetAgainstNetwork(results, nmn, nmnSubnets, componentId, cabinetNetworkName, NMN, nmnId)
	} else {
		results.Fail(
			CabinetNetworkCheck,
			componentId,
			fmt.Sprintf("%s %s missing the NMN network", hardware.Xname, hardware.TypeString))
	}

	// check for networks other than NMN and HMN
	for key := range hardwareNetwork {
		if key != HMN.String() && key != NMN.String() {
			results.Fail(
				CabinetNetworkCheck,
				componentId,
				fmt.Sprintf("%s %s has an invalid network %s. Allowed networks are HMN and NMN", hardware.Xname, hardware.TypeString, key))
		}
	}
}

// Validates the network info from the cabinet to the a subnet in the networks
//
// example hardwareNetwork value:
//
//	{
//	 "CIDR": "10.104.0.0/22",
//	 "Gateway": "10.104.0.1",
//	 "VLan": 3000
//	}
func validateHardwareSubnetAgainstNetwork(
	results *common.ValidationResults,
	hardwareNetwork map[string]interface{},
	subnets map[string]sls_client.NetworkIpv4Subnet,
	hardwareId string,
	cabinetNetworkName CabinetNetworkName,
	networkName NetworkName,
	networkId NetworkId) {

	passed := true
	cidr, _ := common.GetString(hardwareNetwork, CIDR.String())
	subnet, found := subnets[cidr]
	if found {
		gateway, _ := common.GetString(hardwareNetwork, Gateway.String())
		if gateway != subnet.Gateway {
			passed = false
			results.Fail(
				CabinetNetworkCheck,
				hardwareId,
				fmt.Sprintf("The cabinet %s Gateway %s for CIDR %s did not match the gateway %s in %s",
					networkName, gateway, cidr, subnet.Gateway, networkId))
		}
		vlanStr, _ := common.GetString(hardwareNetwork, VLan.String())
		vlan, _ := common.ToInt(vlanStr)
		if vlan != int64(subnet.VlanID) {
			passed = false
			results.Fail(
				CabinetNetworkCheck,
				hardwareId,
				fmt.Sprintf("The cabinet %s vlan %s for CIDR %s did not match the vlan %d in %s",
					networkName, vlanStr, cidr, subnet.VlanID, networkId))
		}
	} else {
		if cidr != subnet.CIDR {
			passed = false
			results.Fail(
				CabinetNetworkCheck,
				hardwareId,
				fmt.Sprintf("The cabinet %s CIDR %s was not found in %s", networkName, cidr, networkId))
		}
	}

	if passed {
		results.Pass(
			CabinetNetworkCheck,
			hardwareId,
			fmt.Sprintf("The cabinet network %s with the CIDR %s matched the network %s", cabinetNetworkName, cidr, networkId))

	}
}
