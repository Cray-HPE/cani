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
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

const (
	chassisBmcCheck common.ValidationCheck = "chassis-bmc"
)

type HardwareChassisBmcCheck struct {
	hardware       map[string]sls_client.Hardware
	typeToHardware map[string][]*sls_client.Hardware
}

func NewHardwareChassisBmcCheck(hardware map[string]sls_client.Hardware, typeToHardware map[string][]*sls_client.Hardware) *HardwareChassisBmcCheck {
	hardwareChassisBmcCheck := HardwareChassisBmcCheck{
		hardware:       hardware,
		typeToHardware: typeToHardware,
	}
	return &hardwareChassisBmcCheck
}

func (c *HardwareChassisBmcCheck) Validate(results *common.ValidationResults) {
	chassis := c.typeToHardware[xnametypes.Chassis.String()]
	for _, h := range chassis {
		if isMountain(h) {
			bmcXname := h.Xname + "b0"
			c.validateChassisBmc(results, h, xnametypes.Chassis, bmcXname, xnametypes.ChassisBMC)
		}
	}

	chassisBmc := c.typeToHardware[xnametypes.ChassisBMC.String()]
	for _, h := range chassisBmc {
		if isMountain(h) {
			bmcXname := xnames.FromString(h.Xname)
			chassisXname := bmcXname.ParentInterface().String()
			c.validateChassisBmc(results, h, xnametypes.ChassisBMC, chassisXname, xnametypes.Chassis)
		}
	}
}

func (c *HardwareChassisBmcCheck) validateChassisBmc(
	results *common.ValidationResults, fromHardware *sls_client.Hardware, fromType xnametypes.HMSType, toXname string, toType xnametypes.HMSType) {
	componentId := fmt.Sprintf("/Hardware/%s", fromHardware.Xname)
	if ftype := xnametypes.GetHMSType(fromHardware.Xname); ftype != fromType {
		results.Fail(
			chassisBmcCheck,
			componentId,
			fmt.Sprintf("The Xname %s is for the type %s but it instead has the TypeString %s",
				fromHardware.Xname, ftype.String(), fromHardware.TypeString))
		return
	}
	toX := xnames.FromString(toXname)
	toHardware, found := c.hardware[toX.String()]
	if found {
		if toHardware.Class != fromHardware.Class {
			results.Fail(
				chassisBmcCheck,
				componentId,
				fmt.Sprintf("The classes do not match, %s %s has a class of %s and %s %s has a class of %s",
					toHardware.Xname, toHardware.TypeString, toHardware.Class,
					fromHardware.Xname, fromHardware.TypeString, fromHardware.Class))
		} else if toHardware.TypeString != toType {
			results.Fail(
				chassisBmcCheck,
				componentId,
				fmt.Sprintf("%s should be of type %s, but is instead of type %s",
					toHardware.Xname, toType.String(), toHardware.TypeString))
		} else {
			results.Pass(
				chassisBmcCheck,
				componentId,
				fmt.Sprintf("%s %s exists for %s %s",
					toHardware.Xname, toHardware.TypeString, fromHardware.Xname, fromHardware.TypeString))
		}
	} else {
		results.Fail(
			chassisBmcCheck,
			componentId,
			fmt.Sprintf("The Hardware entry %s %s is missing, every %s needs a corresponding %s",
				toX.String(), toType.String(), fromHardware.TypeString, toType.String()))
	}
}
