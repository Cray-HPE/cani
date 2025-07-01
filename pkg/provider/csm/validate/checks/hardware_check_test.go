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
	"testing"

	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

func TestGetProps(t *testing.T) {
	h := &sls_client.Hardware{}
	networks := make(map[string]map[string]sls_client.HardwareExtraPropertiesCabinetNetworks)
	cn := make(map[string]sls_client.HardwareExtraPropertiesCabinetNetworks)
	cn["HMN"] = sls_client.HardwareExtraPropertiesCabinetNetworks{
		CIDR:    "10.104.0.0/12",
		Gateway: "10.104.0.1",
		VLan:    3000}
	networks["cn"] = cn
	h.ExtraProperties = sls_client.HardwareExtraPropertiesCabinet{Networks: networks}

	props := getProps(h)
	cidr, _ := common.GetString(props, "Networks", "cn", "HMN", "CIDR")
	if cidr != "10.104.0.0/12" {
		t.Errorf("Unexpected cidr. expected: %s, actual %s ",
			"10.104.0.0/12", cidr)
	}
}
