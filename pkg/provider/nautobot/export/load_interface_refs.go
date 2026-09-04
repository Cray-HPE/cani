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

import (
	"fmt"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// roleRef resolves an interface role name to its Nautobot UUID (uuid.Nil when
// the name is empty).
//
// The generated PatchedWritableInterfaceRequest.Role FK has no `omitempty`, so
// a nil ref serializes as `"role":null` and CLEARS the role in Nautobot. Any
// PATCH that means to keep an existing role must therefore re-send it via this
// helper; an enrichment PATCH that left Role nil is what dropped roles in
// FORGE-305.
func (e *Exporter) roleRef(name string) (uuid.UUID, error) {
	if name == "" {
		return uuid.Nil, nil
	}
	item, err := e.Cache.GetRole(name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve interface role %q: %w", name, err)
	}
	if item == nil {
		return uuid.Nil, fmt.Errorf("resolve interface role %q: not found", name)
	}
	return item.ID, nil
}

// vrfRef resolves a VRF name to its Nautobot UUID using the VRF cache populated
// by the VRF phase, or uuid.Nil when the name is empty or unknown.
func (e *Exporter) vrfRef(name string) uuid.UUID {
	if name == "" {
		return uuid.Nil
	}
	item, ok := e.Cache.LookupCachedVRF(name)
	if !ok || item == nil {
		clog.Warn("unresolved VRF reference %q, skipping", name)
		return uuid.Nil
	}
	return item.ID
}

// lagRef resolves the parent LAG interface on the same device to its Nautobot
// UUID, or uuid.Nil when the LAG name is empty or cannot be resolved.
func (e *Exporter) lagRef(deviceID uuid.UUID, lagName string) uuid.UUID {
	if lagName == "" {
		return uuid.Nil
	}
	lagIface, err := e.Cache.GetInterfaceByDeviceAndName(deviceID, lagName)
	if err != nil || lagIface == nil {
		clog.Warn("unresolved LAG reference %q on device %s, skipping", lagName, deviceID)
		return uuid.Nil
	}
	return lagIface.ID
}

// setPatchedInterfaceMode sets a switchport-mode on a patched interface request,
// leaving it unset when mode is empty or invalid.
func setPatchedInterfaceMode(req *nautobotapi.PatchedWritableInterfaceRequest, mode string) {
	if mode == "" {
		return
	}
	var m nautobotapi.PatchedWritableInterfaceRequest_Mode
	if err := m.FromInterfaceModeChoices(nautobotapi.InterfaceModeChoices(mode)); err != nil {
		return
	}
	req.Mode = &m
}

// untaggedVLANRef resolves a VLAN ID to its Nautobot UUID, or uuid.Nil.
func untaggedVLANRef(vid int, vidToVLAN map[int]uuid.UUID) uuid.UUID {
	if vid == 0 {
		return uuid.Nil
	}
	if vlanID, ok := vidToVLAN[vid]; ok {
		return vlanID
	}
	clog.Warn("unresolved untagged VLAN reference (VID %d), skipping", vid)
	return uuid.Nil
}

// resolveTaggedVLANRefs maps tagged VLAN IDs to Nautobot VLAN UUIDs, skipping
// any VLAN that was not created.
func resolveTaggedVLANRefs(vids []int, vidToVLAN map[int]uuid.UUID) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(vids))
	for _, vid := range vids {
		if vlanID, ok := vidToVLAN[vid]; ok {
			ids = append(ids, vlanID)
		} else {
			clog.Warn("unresolved tagged VLAN reference (VID %d), skipping", vid)
		}
	}
	return ids
}
