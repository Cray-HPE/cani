// MIT License
//
// (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package sls

import (
	"sort"

	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/antihax/optional"
)

func NewHardware(xname xnames.Xname, class sls_client.HardwareClass, extraProperties interface{}) sls_client.Hardware {
	return sls_client.Hardware{
		Xname:           xname.String(),
		Class:           class,
		ExtraProperties: extraProperties,

		// Calculate derived fields
		Parent:     xname.ParentInterface().String(),
		TypeString: xname.Type(),
		Type:       sls_client.HardwareType(sls_common.HMSTypeToHMSStringType(xname.Type())), // The main lookup table is in the SLS package, TODO should maybe move that into this package
	}
}

func NewHardwarePostOpts(hardware sls_client.Hardware) *sls_client.HardwareApiHardwarePostOpts {
	return &sls_client.HardwareApiHardwarePostOpts{
		Body: optional.NewInterface(sls_client.HardwarePost{
			Xname:           hardware.Xname,
			Class:           &hardware.Class,
			ExtraProperties: &hardware.ExtraProperties,
		}),
	}
}

func NewHardwareXnamePutOpts(hardware sls_client.Hardware) *sls_client.HardwareApiHardwareXnamePutOpts {
	return &sls_client.HardwareApiHardwareXnamePutOpts{
		Body: optional.NewInterface(sls_client.HardwarePut{
			Class:           &hardware.Class,
			ExtraProperties: &hardware.ExtraProperties,
		}),
	}
}

func SortHardware(hardware []sls_client.Hardware) {
	sort.Slice(hardware, func(i, j int) bool {
		return hardware[i].Xname < hardware[j].Xname
	})
}

func SortHardwareReverse(hardware []sls_client.Hardware) {
	sort.Slice(hardware, func(i, j int) bool {
		return hardware[i].Xname > hardware[j].Xname
	})
}
