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
package visual

import (
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// chassisUplink describes the path of an uncabled node management port that
// reaches the network through its chassis management module (Apollo 2000
// Gen10 Plus): node iLO -> chassis "iLO port N" -> switch port.
func chassisUplink(idx *htmlIndex, dev *devicetypes.CaniDeviceType, spec *devicetypes.InterfaceSpec, port *htmlPort) {
	if !isMgmtOnly(spec) || dev.Parent == uuid.Nil {
		return
	}
	chassis := idx.inv.Devices[dev.Parent]
	if chassis == nil {
		return
	}
	for i := range chassis.Interfaces {
		uplink := &chassis.Interfaces[i]
		far, ok := idx.peer(chassis.ID, uplink.Name)
		if !ok || !strings.HasPrefix(uplink.Name, "iLO port") {
			continue
		}
		peerDev, peerSpec := idx.resolve(far)
		port.ViaID, port.ViaName, port.ViaPort = chassis.ID.String(), chassis.Name, uplink.Name
		port.PeerID, port.PeerName, port.PeerPort = peerIdentity(peerDev, far)
		port.VLAN = portVLAN(uplink, peerSpec)
		return
	}
}
