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

import (
	"testing"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// TestBuildInterfaceMapRecordsModuleOwner verifies the index keeps the module
// of a module-owned interface whether or not Nautobot also names a device on
// it: Nautobot 3.2 sets the root device on a port of an installed module and
// leaves it empty on a port of a spare module.
//
// Why it matters: cable terminations on module ports can only be resolved
// through the module, so the index must keep that owner next to the device.
// Inputs: an installed-module port carrying Device and Module references and
// a spare-module port carrying only a Module reference. Outputs: both refs
// carry the module UUID and their name; only the first carries the device.
// Data choice: these are the two shapes Nautobot 3.2 returns for a module port.
func TestBuildInterfaceMapRecordsModuleOwner(t *testing.T) {
	installedID, spareID, deviceNBID, moduleNBID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	installed := nautobotapi.Interface{Id: oaPtr(installedID), Name: "p1"}
	setNBRef(&installed.Device, deviceNBID)
	setNBRef(&installed.Module, moduleNBID)
	spare := nautobotapi.Interface{Id: oaPtr(spareID), Name: "p2"}
	setNBRef(&spare.Module, moduleNBID)

	refs := BuildInterfaceMap([]nautobotapi.Interface{installed, spare})

	if ref := refs[installedID]; ref.moduleID != moduleNBID || ref.deviceID != deviceNBID || ref.name != "p1" {
		t.Errorf("installed ref = %+v, want module %s, device %s, name p1", ref, moduleNBID, deviceNBID)
	}
	if ref := refs[spareID]; ref.moduleID != moduleNBID || ref.deviceID != uuid.Nil || ref.name != "p2" {
		t.Errorf("spare ref = %+v, want module %s, no device, name p2", ref, moduleNBID)
	}
}

// TestMapCablesResolvesModuleOwnedTerminations verifies a cable end on a
// module-owned interface terminates on the CANI module, a cable end whose
// module was not imported is left empty, and a device-owned end is unchanged.
//
// Why it matters: the inventory records module ports against the module, so
// import must produce that shape even though Nautobot 3.2 also names the root
// device on a port of an installed module.
// Inputs: a cable from a module port to a device port, and a cable from a
// port on an unknown module to the same device port; both module ports carry
// the root device as Nautobot 3.2 returns them.
// Outputs: the first cable's A end names the CANI module and its port; the
// second cable's A end is empty while its B end still resolves.
// Data choice: giving the module ports the device of the device-owned B end
// proves the module, not the device, decides the owner, and that an end on an
// unknown module does not fall back to that device.
func TestMapCablesResolvesModuleOwnedTerminations(t *testing.T) {
	devNBID, devCaniID := uuid.New(), uuid.New()
	modNBID, modCaniID := uuid.New(), uuid.New()
	modPort, devPort, strayPort := uuid.New(), uuid.New(), uuid.New()
	ifaceMap := map[uuid.UUID]ifaceRef{
		modPort:   {deviceID: devNBID, moduleID: modNBID, name: "p1"},
		devPort:   {deviceID: devNBID, name: "eth0"},
		strayPort: {deviceID: devNBID, moduleID: uuid.New(), name: "p1"},
	}
	cableIDs := []uuid.UUID{uuid.New(), uuid.New()}
	raw := []nautobotapi.Cable{
		{Id: oaPtr(cableIDs[0]), TerminationAId: oaPtr(modPort), TerminationBId: oaPtr(devPort)},
		{Id: oaPtr(cableIDs[1]), TerminationAId: oaPtr(strayPort), TerminationBId: oaPtr(devPort)},
	}

	got := MapCables(raw, map[uuid.UUID]uuid.UUID{devNBID: devCaniID}, ifaceMap, map[uuid.UUID]uuid.UUID{modNBID: modCaniID}, nil)

	byNB := map[uuid.UUID]int{}
	for _, cable := range got {
		for i, id := range cableIDs {
			if cable.ExternalIDs[externalIDKeyNautobot] == id {
				byNB[id] = i
				switch i {
				case 0:
					if cable.TerminationADevice != modCaniID || cable.TerminationAPort != "p1" || cable.TerminationA != modPort {
						t.Errorf("module-owned end = device %v, port %q, iface %v; want module %v, p1, %v",
							cable.TerminationADevice, cable.TerminationAPort, cable.TerminationA, modCaniID, modPort)
					}
				case 1:
					if cable.TerminationADevice != uuid.Nil || cable.TerminationA != uuid.Nil {
						t.Errorf("unknown-module end = device %v, iface %v; want both Nil", cable.TerminationADevice, cable.TerminationA)
					}
				}
				if cable.TerminationBDevice != devCaniID || cable.TerminationBPort != "eth0" {
					t.Errorf("device-owned end = %v %q, want %v eth0", cable.TerminationBDevice, cable.TerminationBPort, devCaniID)
				}
			}
		}
	}
	if len(byNB) != 2 {
		t.Fatalf("mapped %d of 2 cables", len(byNB))
	}
}
