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
	"net"

	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

type NetworkIpRangeCheck struct {
	slsStateExtended *common.SlsStateExtended
	k8sPodsCidr      string
	k8sServicesCidr  string
}

func NewNetworkIpRangeCheck(slsStateExtended *common.SlsStateExtended, k8sPodsCidr string, k8sServicesCidr string) *NetworkIpRangeCheck {
	networkIpRangeCheck := NetworkIpRangeCheck{
		slsStateExtended: slsStateExtended,
		k8sPodsCidr:      k8sPodsCidr,
		k8sServicesCidr:  k8sServicesCidr,
	}
	return &networkIpRangeCheck
}

func (c *NetworkIpRangeCheck) Validate(results *common.ValidationResults) {
	passFailResults := make(map[string]bool)
	ipRangeMap := make(map[string]*sls_client.Network)
	ipRanges := make([]string, 0)

	k8sPodsCidr, _ := parseConfigFileCidr(results, "k8spodscidr", c.k8sPodsCidr)
	k8sServicesCidr, _ := parseConfigFileCidr(results, "k8sservicescidr", c.k8sServicesCidr)

	for _, network := range c.slsStateExtended.SlsState.Networks {
		n := network
		for _, r := range network.IPRanges {
			ipRangeMap[r] = &n
			ipRanges = append(ipRanges, r)
			passFailResults[r] = true
		}
	}
	for i, ipRange1 := range ipRanges {
		name1 := ipRangeMap[ipRange1].Name
		_, net1, _ := net.ParseCIDR(ipRange1)
		checkK8sCidr(results, "CSM", k8sPodsCidr, "k8spodscidr", net1, name1)
		checkK8sCidr(results, "CSM", k8sServicesCidr, "k8sservicescidr", net1, name1)
		if !net1.IP.IsUnspecified() {
			for j := i + 1; j < len(ipRanges); j++ {
				ipRange2 := ipRanges[j]
				_, net2, err := net.ParseCIDR(ipRange2)
				if err != nil {
					results.Fail(
						common.IPRangeConflictCheck,
						name1,
						fmt.Sprintf("%s invalid IP range: %s", name1, ipRange2))
					passFailResults[ipRange2] = false
					continue
				}

				// only bican can have 0.0.0.0/0
				// SystemDefaultRoute one of CAN, CHN
				// check for wrong fields
				// schema check on ExtraProperties
				if !net2.IP.IsUnspecified() && (net1.Contains(net2.IP) || net2.Contains(net1.IP)) {
					name2 := ipRangeMap[ipRange2].Name
					results.Fail(
						common.IPRangeConflictCheck,
						name1,
						fmt.Sprintf("%s %s overlaps with %s %s", name1, ipRange1, name2, ipRange2))
					results.Fail(
						common.IPRangeConflictCheck,
						name2,
						fmt.Sprintf("%s %s overlaps with %s %s", name2, ipRange2, name1, ipRange1))
					passFailResults[ipRange1] = false
					passFailResults[ipRange2] = false
				}
			}
		}
	}
	for ipRange, pass := range passFailResults {
		name := ipRangeMap[ipRange].Name
		if pass {
			results.Pass(
				common.IPRangeConflictCheck,
				name,
				fmt.Sprintf("IP range for %s %s does not overlap with any other network", name, ipRange))
		}
	}
}

func checkK8sCidr(results *common.ValidationResults, componentId string, k8sCidr *net.IPNet, k8sCidrName string, cidr *net.IPNet, cidrName string) {
	if k8sCidr == nil || cidr == nil {
		return
	}
	if k8sCidr.IP.IsUnspecified() || cidr.IP.IsUnspecified() {
		return
	}

	if k8sCidr.Contains(cidr.IP) || cidr.Contains(k8sCidr.IP) {
		results.Fail(
			common.IPRangeConflictWithK8sCheck,
			componentId,
			fmt.Sprintf("%s %s overlaps with %s %s", k8sCidrName, k8sCidr.String(), cidrName, cidr.String()))
	}
}

func parseConfigFileCidr(results *common.ValidationResults, componentId, cidr string) (*net.IPNet, error) {
	_, cidrIpNet, err := net.ParseCIDR(cidr)
	if err != nil {
		results.Fail(
			common.IPRangeConflictCheck,
			componentId,
			fmt.Sprintf("The configuration file contains an invalid CIDR: %s. %s", cidr, err))
	}

	return cidrIpNet, err
}
