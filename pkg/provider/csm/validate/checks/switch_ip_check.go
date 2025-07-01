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
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

const (
	SwitchIPCheck common.ValidationCheck = "switch-has-ips"
)

type SwitchIpCheck struct {
	slsStateExtended *common.SlsStateExtended
}

func NewSwitchIpCheck(slsStateExtended *common.SlsStateExtended) *SwitchIpCheck {
	switchIpCheck := SwitchIpCheck{
		slsStateExtended: slsStateExtended,
	}
	return &switchIpCheck
}

func (c *SwitchIpCheck) Validate(results *common.ValidationResults) {
	switchTypes := [2]string{
		xnametypes.MgmtSwitch.String(),
		xnametypes.MgmtHLSwitch.String(),
	}
	networksToCheck := make([]*sls_client.Network, 0)
	for _, network := range c.slsStateExtended.SlsState.Networks {
		n := network
		switch network.Name {
		case "MTL":
			fallthrough
		case "HMN":
			fallthrough
		case "NMN":
			fallthrough
		case "CMN":
			networksToCheck = append(networksToCheck, &n)
		}
	}

	for _, switchType := range switchTypes {
		list, found := c.slsStateExtended.TypeToHardware[switchType]
		if found {
			for _, hardware := range list {
				aliases := getAliases(hardware)
				for _, alias := range aliases {
					for _, network := range networksToCheck {
						found := false
						for _, subnet := range network.ExtraProperties.Subnets {
							for _, res := range subnet.IPReservations {
								if res.Name == alias {
									if res.IPAddress != "" {
										found = true
										break
									}
								}
							}
							if found {
								break
							}
						}
						if found {
							results.Pass(
								SwitchIPCheck,
								hardware.Xname,
								fmt.Sprintf("Switch %s has IP in network %s", alias, network.Name))
						} else {
							results.Fail(
								SwitchIPCheck,
								hardware.Xname,
								fmt.Sprintf("Switch %s does not have IP in network %s", alias, network.Name))
						}
					}
				}
			}
		}
	}
}

func getAliases(hardware *sls_client.Hardware) []string {
	aliases, _ := common.GetSliceOfStrings(hardware.ExtraProperties, "Aliases")
	return aliases
}
