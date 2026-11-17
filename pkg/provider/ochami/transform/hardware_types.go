/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
package transform

import "strings"

// IdentityKeys names the keys in the OChaMi bucket that identify one
// physical device: the xname alone, which means the same as CSM's.
// Transform declares them on its result.
var IdentityKeys = []string{"xname"}

// hardwareTypeMap maps ochami deviceType strings to CANI canonical hardware types.
var hardwareTypeMap = map[string]string{
	"rack":         "rack",
	"chassis":      "chassis",
	"blade":        "blade",
	"node":         "node",
	"mgmt-switch":  "mgmt-switch",
	"hsn-switch":   "hsn-switch",
	"cabinet-pdu":  "pdu",
	"cdu":          "cdu",
	"cpu":          "cpu",
	"dimm":         "memory",
	"gpu":          "gpu",
	"nic":          "nic",
	"power-supply": "psu",
	"cable":        "cable",
	"bmc":          "bmc",
}

// normaliseHardwareType maps an ochami deviceType string to a CANI canonical hardware type.
func normaliseHardwareType(deviceType string) string {
	if canonical, ok := hardwareTypeMap[strings.ToLower(deviceType)]; ok {
		return canonical
	}
	return strings.ToLower(deviceType)
}
