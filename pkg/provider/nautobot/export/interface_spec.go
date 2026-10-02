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
package export

import "github.com/Cray-HPE/cani/pkg/devicetypes"

// specFromInterface converts a persisted interface spec, whether it belongs to
// a device or to a module, into the exporter's wire-level spec. Both owners
// go through here so they carry the same attribute set to Nautobot.
func specFromInterface(iface devicetypes.InterfaceSpec) interfaceSpec {
	ifaceType := mapInterfaceType(string(iface.Type))
	mgmtOnly := iface.MgmtOnly != nil && *iface.MgmtOnly
	role := iface.Role
	if role == "" {
		role = devicetypes.InferInterfaceRole(iface.Name, iface.Type, mgmtOnly)
	}
	return interfaceSpec{
		Name:         iface.Name,
		Type:         ifaceType,
		Speed:        getSpeedForType(ifaceType),
		Role:         role,
		MgmtOnly:     mgmtOnly,
		Mac:          iface.MacAddress,
		Label:        iface.Label,
		Tags:         iface.Tags,
		Lag:          iface.Lag,
		Mode:         iface.Mode,
		UntaggedVLAN: iface.UntaggedVLAN,
		TaggedVLANs:  iface.TaggedVLANs,
		VRF:          iface.VRF,
		Description:  iface.Description,
	}
}

// optionalString returns a pointer to value, or nil when it is empty so the
// request field is omitted rather than cleared.
func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
