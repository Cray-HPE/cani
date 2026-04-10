package devicetypes

import "github.com/google/uuid"

// ExpandChildren recursively creates child devices for each device-bay
// that has a default slug defined. Each child gets a new UUID, status
// Staged, and its Parent set to the parent device's ID. The parent's
// Children slice is updated accordingly.
//
// Returns a flat map of all newly created devices (keyed by UUID),
// including nested descendants.
func ExpandChildren(parent *CaniDeviceType) map[uuid.UUID]*CaniDeviceType {
	result := make(map[uuid.UUID]*CaniDeviceType)
	expandBays(parent, result)
	return result
}

// expandBays iterates over the device-bays of parent, instantiates
// the default child for each bay, and recurses into the child.
func expandBays(parent *CaniDeviceType, acc map[uuid.UUID]*CaniDeviceType) {
	for _, bay := range parent.DeviceBays {
		if bay.Default == nil {
			continue
		}
		slugs := bay.Default.Slugs()
		if len(slugs) == 0 {
			continue
		}
		child, err := NewDeviceFromSlug(slugs[0])
		if err != nil {
			continue
		}
		child.Parent = parent.ID
		parent.Children = append(parent.Children, child.ID)
		acc[child.ID] = child
		expandBays(child, acc)
	}
}
