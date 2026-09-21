package devicetypes

import (
	"slices"

	"github.com/google/uuid"
)

func (inv *Inventory) removeDeviceTree(id uuid.UUID) {
	device := inv.Devices[id]
	if device == nil {
		return
	}
	inv.unlinkDeviceFromParent(device, id)
	inv.unindexDevice(id, device)
	inv.removeCablesForDevice(id)
	inv.removeOwnedInterfaces(device.Interfaces)
	inv.removeModulesForDevice(id)
	inv.removeOwnedFrus(id)
	inv.detachDeviceAssignments(id)
	delete(inv.Devices, id)
	for childID, child := range inv.Devices {
		if child != nil && child.Parent == id {
			inv.removeDeviceTree(childID)
		}
	}
}

func (inv *Inventory) removeModule(id uuid.UUID) {
	module := inv.Modules[id]
	inv.removeCablesForDevice(id)
	inv.removeOwnedInterfaces(module.Interfaces)
	inv.removeOwnedFrus(id)
	delete(inv.Modules, id)
}

func (inv *Inventory) removeOwnedFrus(ownerID uuid.UUID) {
	for fruID, fru := range inv.Frus {
		if fru == nil || (fru.Device != ownerID && fru.Parent != ownerID) {
			continue
		}
		delete(inv.Frus, fruID)
		inv.removeOwnedFrus(fruID)
	}
}

func (inv *Inventory) detachDeviceAssignments(deviceID uuid.UUID) {
	for _, vrf := range inv.VRFs {
		if vrf != nil {
			vrf.Devices = removeUUID(vrf.Devices, deviceID)
		}
	}
	for _, device := range inv.Devices {
		if device != nil && device.BMCParent == deviceID {
			device.BMCParent = uuid.Nil
		}
	}
}

func (inv *Inventory) removeOwnedInterfaces(specs []InterfaceSpec) {
	removed := make(map[uuid.UUID]bool)
	for _, spec := range specs {
		if spec.ID != uuid.Nil {
			removed[spec.ID] = true
		}
	}
	inv.removeInterfaceReferences(removed)
}

func (inv *Inventory) removeInterfaceReferences(removed map[uuid.UUID]bool) {
	for interfaceID := range removed {
		delete(inv.Interfaces, interfaceID)
	}
	for _, address := range inv.IPAddresses {
		if address != nil {
			address.Interfaces = slices.DeleteFunc(address.Interfaces, func(interfaceID uuid.UUID) bool {
				return removed[interfaceID]
			})
		}
	}
	for cableID, cable := range inv.Cables {
		if cable != nil && inv.cableUsesRemovedInterface(cable, removed) {
			_ = inv.RemoveCable(cableID)
		}
	}
}

func (inv *Inventory) cableUsesRemovedInterface(cable *CaniCableType, removed map[uuid.UUID]bool) bool {
	return removed[cable.TerminationA] || removed[cable.TerminationB] ||
		removed[inv.FindInterfaceIDByPort(cable.TerminationADevice, cable.TerminationAPort)] ||
		removed[inv.FindInterfaceIDByPort(cable.TerminationBDevice, cable.TerminationBPort)]
}

func (inv *Inventory) unlinkCableReferences(cableID uuid.UUID) {
	for _, device := range inv.Devices {
		if device != nil {
			unlinkCableSpecs(device.Interfaces, cableID)
		}
	}
	for _, module := range inv.Modules {
		if module != nil {
			unlinkCableSpecs(module.Interfaces, cableID)
		}
	}
	inv.unlinkInterfaceCableReferences(cableID)
}

func (inv *Inventory) unlinkInterfaceCableReferences(cableID uuid.UUID) {
	for _, iface := range inv.Interfaces {
		if iface != nil && iface.ConnectedCable != nil && *iface.ConnectedCable == cableID {
			iface.ConnectedCable = nil
		}
	}
}

func unlinkCableSpecs(specs []InterfaceSpec, cableID uuid.UUID) {
	for index := range specs {
		if specs[index].ConnectedCable != nil && *specs[index].ConnectedCable == cableID {
			specs[index].ConnectedCable = nil
		}
	}
}
