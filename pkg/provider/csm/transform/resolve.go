package transform

import (
	"log"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// resolveExistingID checks whether a device with matching CSM xname
// already exists in the inventory. Returns the existing UUID if
// found, otherwise generates a new UUID. This makes repeated imports
// idempotent: the same SLS data always produces the same device UUIDs.
func resolveExistingID(xname string, existing *devicetypes.Inventory) uuid.UUID {
	if existing == nil || xname == "" {
		return uuid.New()
	}

	checks := []devicetypes.ProviderKeyCheck{
		{Key: "xname", Value: xname},
	}

	if match := existing.FindDeviceByProviderKeys("csm", checks); match != nil {
		log.Printf("Matched existing device %s (%s) by CSM xname", match.Name, match.ID)
		return match.ID
	}

	return uuid.New()
}

// resolveExistingRackID checks whether a rack with the same name already
// exists in the inventory. Returns the existing UUID if found, otherwise
// generates a new UUID.
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
