package devicetypes

import (
	"fmt"

	"github.com/google/uuid"
)

func (inv *Inventory) validateIPAMRelationships() *RelationshipResult {
	result := &RelationshipResult{}
	inv.validateVLANRelationships(result)
	inv.validatePrefixRelationships(result)
	inv.validateIPAddressRelationships(result)
	inv.validateVRFRelationships(result)
	inv.validateDeviceIPAMRelationships(result)
	return result
}

func (inv *Inventory) rebuildInterfaceIPAddresses() {
	for addressID, address := range inv.IPAddresses {
		if address == nil {
			continue
		}
		for _, interfaceID := range address.Interfaces {
			iface := inv.Interfaces[interfaceID]
			if iface != nil && !containsUUID(iface.IPAddresses, addressID) {
				iface.IPAddresses = append(iface.IPAddresses, addressID)
			}
		}
	}
}

func (inv *Inventory) validateVLANRelationships(result *RelationshipResult) {
	for vlanID, vlan := range inv.VLANs {
		if vlan == nil {
			continue
		}
		owner := fmt.Sprintf("VLAN %q (%s)", vlan.Name, vlanID)
		validateOptionalIPAMReference(result, owner, "location", vlan.Location, inv.Locations[vlan.Location] != nil)
	}
}

func (inv *Inventory) validatePrefixRelationships(result *RelationshipResult) {
	for prefixID, prefix := range inv.Prefixes {
		if prefix == nil {
			continue
		}
		owner := fmt.Sprintf("prefix %q (%s)", prefix.Prefix, prefixID)
		validateOptionalIPAMReference(result, owner, "location", prefix.Location, inv.Locations[prefix.Location] != nil)
		validateOptionalIPAMReference(result, owner, "VLAN", prefix.VLAN, inv.VLANs[prefix.VLAN] != nil)
		validateOptionalIPAMReference(result, owner, "parent prefix", prefix.Parent, inv.Prefixes[prefix.Parent] != nil)
		if inv.hasPrefixCycle(prefixID) {
			result.Errors = append(result.Errors, fmt.Errorf("%s: circular parent prefix reference detected", owner))
		}
	}
}

func (inv *Inventory) validateIPAddressRelationships(result *RelationshipResult) {
	for addressID, address := range inv.IPAddresses {
		if address == nil {
			continue
		}
		owner := fmt.Sprintf("IP address %q (%s)", address.Address, addressID)
		validateOptionalIPAMReference(result, owner, "parent prefix", address.Parent, inv.Prefixes[address.Parent] != nil)
		validateOptionalIPAMReference(result, owner, "NAT inside IP", address.NATInside, inv.IPAddresses[address.NATInside] != nil)
		validateIPAMAssignments(result, owner, "interface", address.Interfaces, inv.Interfaces)
	}
}

func (inv *Inventory) validateVRFRelationships(result *RelationshipResult) {
	for vrfID, vrf := range inv.VRFs {
		if vrf == nil {
			continue
		}
		owner := fmt.Sprintf("VRF %q (%s)", vrf.Name, vrfID)
		validateIPAMAssignments(result, owner, "device", vrf.Devices, inv.Devices)
	}
}

func (inv *Inventory) validateDeviceIPAMRelationships(result *RelationshipResult) {
	for deviceID, device := range inv.Devices {
		if device == nil {
			continue
		}
		owner := fmt.Sprintf("device %q (%s)", device.Name, deviceID)
		validateOptionalIPAMReference(result, owner, "primary IPv4", device.PrimaryIPv4, inv.IPAddresses[device.PrimaryIPv4] != nil)
		validateOptionalIPAMReference(result, owner, "primary IPv6", device.PrimaryIPv6, inv.IPAddresses[device.PrimaryIPv6] != nil)
		validateIPAMAssignments(result, owner, "VLAN", device.AssignedVLANs, inv.VLANs)
	}
}

func validateOptionalIPAMReference(result *RelationshipResult, owner, relation string, targetID uuid.UUID, exists bool) {
	if targetID != uuid.Nil && !exists {
		result.Errors = append(result.Errors, fmt.Errorf("%s: %s %s not found", owner, relation, targetID))
	}
}

func validateIPAMAssignments[Item any](result *RelationshipResult, owner, relation string, ids []uuid.UUID, items map[uuid.UUID]*Item) {
	for _, targetID := range ids {
		if items[targetID] == nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %s %s not found", owner, relation, targetID))
		}
	}
}

func (inv *Inventory) hasPrefixCycle(prefixID uuid.UUID) bool {
	visited := make(map[uuid.UUID]bool)
	for prefixID != uuid.Nil {
		prefix := inv.Prefixes[prefixID]
		if prefix == nil {
			return false
		}
		if visited[prefixID] {
			return true
		}
		visited[prefixID] = true
		prefixID = prefix.Parent
	}
	return false
}
