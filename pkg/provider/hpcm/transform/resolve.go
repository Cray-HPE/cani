package transform

import (
	"log"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// resolveExistingDeviceID checks whether a device with matching HPCM
// metadata already exists in the inventory. Returns the existing UUID
// if found, otherwise generates a new UUID. This makes repeated
// imports idempotent: the same HPCM data always produces the same
// device UUIDs.
func resolveExistingDeviceID(name, hpcmUUID string, existing *devicetypes.Inventory) uuid.UUID {
	if existing == nil {
		return uuid.New()
	}

	var checks []devicetypes.ProviderKeyCheck
	if hpcmUUID != "" {
		checks = append(checks, devicetypes.ProviderKeyCheck{
			Key: "hpcm_uuid", Value: hpcmUUID,
		})
	}

	if match := existing.FindDeviceByProviderKeys("hpcm", checks); match != nil {
		log.Printf("Matched existing device %s (%s) by HPCM provider key", match.Name, match.ID)
		return match.ID
	}

	// Fallback: check by device name.
	if name != "" {
		for _, dev := range existing.Devices {
			if dev != nil && dev.Name == name {
				log.Printf("Matched existing device %s (%s) by name", dev.Name, dev.ID)
				return dev.ID
			}
		}
	}

	return uuid.New()
}

// resolveExistingRackID checks whether a rack with the same name
// already exists in the inventory. Returns the existing UUID if
// found, otherwise generates a new UUID.
func resolveExistingRackID(name string, existing *devicetypes.Inventory) uuid.UUID {
	if existing == nil || name == "" {
		return uuid.New()
	}

	if rack := existing.FindRackByName(name); rack != nil {
		log.Printf("Matched existing rack %s (%s) by name", rack.Name, rack.ID)
		return rack.ID
	}

	return uuid.New()
}
