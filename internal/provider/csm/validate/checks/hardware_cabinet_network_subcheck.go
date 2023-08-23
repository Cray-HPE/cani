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

const (
	hmnMtnId = "/Networks/HMN_MTN"
	nmnMtnId = "/Networks/NMN_MTN"
	hmnRvrId = "/Networks/HMN_RVR"
	nmnRvrId = "/Networks/NMN_RVR"
)

type HardwareCabinetNetworkSubCheck struct {
	HmnRvr map[string]sls_client.NetworkIpv4Subnet
	HmnMtn map[string]sls_client.NetworkIpv4Subnet
	NmnRvr map[string]sls_client.NetworkIpv4Subnet
	NmnMtn map[string]sls_client.NetworkIpv4Subnet
}

func NewHardwareCabinetNetworkSubCheck(networks map[string]sls_client.Network) *HardwareCabinetNetworkSubCheck {
	subnets := &HardwareCabinetNetworkSubCheck{}

	hmnMtn, found := networks["HMN_MTN"]
	if found {
		subnets.HmnMtn = mapNetworkSubnets(&hmnMtn)
	} else {
		subnets.HmnMtn = make(map[string]sls_client.NetworkIpv4Subnet)
	}

	hmnRvr, found := networks["HMN_RVR"]
	if found {
		subnets.HmnRvr = mapNetworkSubnets(&hmnRvr)
	} else {
		subnets.HmnRvr = make(map[string]sls_client.NetworkIpv4Subnet)
	}

	nmnMtn, found := networks["NMN_MTN"]
	if found {
		subnets.NmnMtn = mapNetworkSubnets(&nmnMtn)
	} else {
		subnets.NmnMtn = make(map[string]sls_client.NetworkIpv4Subnet)
	}

	nmnRvr, found := networks["NMN_RVR"]
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

	cn, foundCn := common.GetMap(networks, "cn")
	if isRiver(hardware) {
		for n := range networks {
			if n != "cn" && n != "ncn" {
				results.Fail(
					CabinetNetworkCheck,
					componentId,
					fmt.Sprintf("%s %s invalid network catagory, %s. Allowed categories are cn and ncn", hardware.Xname, hardware.TypeString, n))
			}
		}
		if foundCn {
			validateCabinetNetwork(results, hardware, c.HmnRvr, hmnRvrId, c.NmnRvr, nmnRvrId, cn)
		} else {
			results.Fail(
				CabinetNetworkCheck,
				componentId,
				fmt.Sprintf("%s %s missing required network cn", hardware.Xname, hardware.TypeString))
		}
		ncn, foundNcn := common.GetMap(networks, "ncn")
		if foundNcn {
			validateCabinetNetwork(results, hardware, c.HmnRvr, hmnRvrId, c.NmnRvr, nmnRvrId, ncn)
		}
	} else {
		for n := range networks {
			if n != "cn" {
				results.Fail(
					CabinetNetworkCheck,
					componentId,
					fmt.Sprintf("%s %s invalid network catagory, %s. The only allowed category is cn", hardware.Xname, hardware.TypeString, n))
			}
		}
		if foundCn {
			validateCabinetNetwork(results, hardware, c.HmnMtn, hmnMtnId, c.NmnMtn, nmnMtnId, cn)
		} else {
			results.Fail(
				CabinetNetworkCheck,
				componentId,
				fmt.Sprintf("%s %s missing required network cn", hardware.Xname, hardware.TypeString))
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
	hmnSubnets map[string]sls_client.NetworkIpv4Subnet,
	hmnId string,
	nmnSubnets map[string]sls_client.NetworkIpv4Subnet,
	nmnId string,
	hardwareNetwork map[string]interface{}) {

	componentId := fmt.Sprintf("/Hardware/%s", hardware.Xname)

	// HMN
	hmn, found := common.GetMap(hardwareNetwork, "HMN")
	if found {
		validateHardwareSubnetAgainstNetwork(results, hmn, hmnSubnets, componentId, "HMN", hmnId)
	} else {
		results.Fail(
			CabinetNetworkCheck,
			componentId,
			fmt.Sprintf("%s %s is missing the HMN network", hardware.Xname, hardware.TypeString))
	}

	// NMN
	nmn, found := common.GetMap(hardwareNetwork, "NMN")
	if found {
		validateHardwareSubnetAgainstNetwork(results, nmn, nmnSubnets, componentId, "NMN", nmnId)
	} else {
		results.Fail(
			CabinetNetworkCheck,
			componentId,
			fmt.Sprintf("%s %s is missing the NMN network", hardware.Xname, hardware.TypeString))
	}

	// check for networks other than NMN and HMN
	for key := range hardwareNetwork {
		if key != "HMN" && key != "NMN" {
			results.Fail(
				CabinetNetworkCheck,
				componentId,
				fmt.Sprintf("%s %s has an invalid network %s. Allowed networks are HMN and NMN", hardware.Xname, hardware.TypeString, key))
		}
	}
}

// Validates the network info from the cabinet to the a subnet in the networks
//
// hardwareNetwork is
//
//	{
//	  "CIDR": "10.104.0.0/22",
//	  "Gateway": "10.104.0.1",
//	  "VLan": 3000
//	}
//
// networkName is one of HMN or NMN
// networkId is one of /Networks/HMN_MTN, /Networks/NMN_MTN, /Networks/HMN_RVR, or /Networks/NMN_RVR
func validateHardwareSubnetAgainstNetwork(
	results *common.ValidationResults,
	hardwareNetwork map[string]interface{},
	subnets map[string]sls_client.NetworkIpv4Subnet,
	hardwareId, networkName, networkId string) {

	cidr, _ := common.GetString(hardwareNetwork, "CIDR")
	subnet, found := subnets[cidr]
	if found {
		gateway, _ := common.GetString(hardwareNetwork, "Gateway")
		if gateway != subnet.Gateway {
			results.Fail(
				CabinetNetworkCheck,
				hardwareId,
				fmt.Sprintf("The cabinet %s Gateway %s for CIDR %s did not match the gateway in %s with the same CIDR",
					networkName, gateway, cidr, networkId))
		}
		vlanStr, _ := common.GetString(hardwareNetwork, "VLan")
		vlan, _ := common.ToInt(vlanStr)
		if vlan != int64(subnet.VlanID) {
			results.Fail(
				CabinetNetworkCheck,
				hardwareId,
				fmt.Sprintf("The cabinet %s vlan %s for CIDR %s did not match the vlan in %s with the same CIDR",
					networkName, vlanStr, cidr, networkId))
		}
	} else {
		if cidr != subnet.CIDR {
			results.Fail(
				CabinetNetworkCheck,
				hardwareId,
				fmt.Sprintf("The cabinet %s CIDR %s was not found in %s", networkName, cidr, networkId))
		}
	}

}
