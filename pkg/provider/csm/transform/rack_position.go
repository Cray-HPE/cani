package transform

import (
	"fmt"
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// rackDeviceRef pairs a device UUID with its pointer for position assignment.
type rackDeviceRef struct {
	id  uuid.UUID
	dev *devicetypes.CaniDeviceType
}

// assignRackPositions sets RackPosition on devices that should be
// visible in the rack view. Switches are placed at the top of the
// rack and blades are placed below them, starting from the bottom.
func assignRackPositions(result *devicetypes.TransformResult) {
	// Build a lookup: rack name → rackID.
	rackByName := make(map[string]uuid.UUID)
	for id, rack := range result.Racks {
		rackByName[rack.Name] = id
	}

	// Group rack-visible devices by their cabinet xname.
	switches := make(map[string][]rackDeviceRef)
	blades := make(map[string][]rackDeviceRef)

	for id, dev := range result.Devices {
		cabinet := cabinetForDevice(dev)
		if cabinet == "" {
			continue
		}
		switch dev.Type {
		case devicetypes.TypeMgmtSwitch, devicetypes.TypeSwitch, devicetypes.TypeHsnSwitch:
			switches[cabinet] = append(switches[cabinet], rackDeviceRef{id, dev})
		case devicetypes.TypeBlade:
			blades[cabinet] = append(blades[cabinet], rackDeviceRef{id, dev})
		}
	}

	// Assign positions per rack.
	for cabName, rackID := range rackByName {
		rack := result.Racks[rackID]
		height := rack.UHeight
		if height == 0 {
			height = 42
		}
		placeSwitches(switches[cabName], height)
		placeBlades(blades[cabName])
	}
}

// placeSwitches assigns positions from the top of the rack downward.
func placeSwitches(refs []rackDeviceRef, rackHeight int) {
	sortByXname(refs)
	pos := rackHeight
	for i := range refs {
		h := refs[i].dev.UHeight
		if h == 0 {
			h = 1
		}
		refs[i].dev.RackPosition = pos - h + 1
		pos -= h
	}
}

// placeBlades assigns positions from the bottom of the rack upward.
func placeBlades(refs []rackDeviceRef) {
	sortByXname(refs)
	pos := 1
	for i := range refs {
		h := refs[i].dev.UHeight
		if h == 0 {
			h = 2
		}
		refs[i].dev.RackPosition = pos
		pos += h
	}
}

// cabinetForDevice extracts the cabinet xname from provider metadata.
// Returns "" when the xname is unavailable.
func cabinetForDevice(dev *devicetypes.CaniDeviceType) string {
	xname := xnameFromMetadata(dev)
	if xname == "" {
		return ""
	}
	parsed := ParseXname(xname)
	if parsed.Cabinet == 0 && parsed.Type == "" {
		return ""
	}
	return fmt.Sprintf("x%d", parsed.Cabinet)
}

// xnameFromMetadata extracts the xname string from CSM provider metadata.
func xnameFromMetadata(dev *devicetypes.CaniDeviceType) string {
	if dev.ProviderMetadata == nil {
		return ""
	}
	csm, ok := dev.ProviderMetadata["csm"]
	if !ok {
		return ""
	}
	md, ok := csm.(map[string]any)
	if !ok {
		return ""
	}
	x, _ := md["xname"].(string)
	return x
}

// sortByXname sorts device refs by xname for deterministic ordering.
func sortByXname(refs []rackDeviceRef) {
	sort.Slice(refs, func(i, j int) bool {
		return xnameFromMetadata(refs[i].dev) < xnameFromMetadata(refs[j].dev)
	})
}
