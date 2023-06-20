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
/*
MIT License

(C) Copyright 2023 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/

package checks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

const (
	cabinetExistsForHardware common.ValidationCheck = "cabinet-exists-for-hardware"
)

type HardwareCabinetCheck struct {
	hardware map[string]sls_client.Hardware
}

func NewHardwareCabinetCheck(hardware map[string]sls_client.Hardware) *HardwareCabinetCheck {
	hardwareCabinetCheck := HardwareCabinetCheck{
		hardware: hardware,
	}
	return &hardwareCabinetCheck
}

func (c *HardwareCabinetCheck) Validate(results *common.ValidationResults) {
	pattern := regexp.MustCompile("^x[0-9]{1,4}")
	for _, h := range c.hardware {
		if xnametypes.Cabinet == h.TypeString {
			continue
		}
		if !strings.HasPrefix(h.Xname, "x") {
			// Only check for hardware that is present withing a cabinet (starts with x),
			// and not for example a CDU (starts with d).
			continue
		}

		componentId := fmt.Sprintf("/Hardware/%s", h.Xname)
		matches := pattern.FindAllString(h.Xname, 1)

		if len(matches) > 0 {
			cabinetXname := matches[0]
			_, found := c.hardware[cabinetXname]
			if found {
				results.Pass(
					cabinetExistsForHardware,
					componentId,
					fmt.Sprintf("The cabinet %s exits for %s %s", cabinetXname, h.Xname, h.TypeString))
			} else {
				results.Fail(
					cabinetExistsForHardware,
					componentId,
					fmt.Sprintf("Missing cabinet %s for %s %s", cabinetXname, h.Xname, h.TypeString))
			}
		} else {
			results.Fail(
				cabinetExistsForHardware,
				componentId,
				fmt.Sprintf("Could not determine cabinet xname for %s %s", h.Xname, h.TypeString))
		}
	}
}
