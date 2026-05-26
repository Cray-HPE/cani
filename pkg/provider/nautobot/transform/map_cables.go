package transform

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// MapCables converts Nautobot Cable objects to CANI cables.
// It requires:
//   - deviceMap: Nautobot device UUID → CANI device UUID
//   - ifaceMap: Nautobot interface UUID → (Nautobot device UUID, interface name)
func MapCables(
	raw []nautobotapi.Cable,
	deviceMap map[uuid.UUID]uuid.UUID,
	ifaceMap map[uuid.UUID]ifaceRef,
) map[uuid.UUID]*devicetypes.CaniCableType {
	result := make(map[uuid.UUID]*devicetypes.CaniCableType, len(raw))

	for _, cable := range raw {
		nbID := directUUID(cable.Id)
		if nbID == uuid.Nil {
			continue
		}
		caniID := uuid.New()

		caniCable := &devicetypes.CaniCableType{
			ID:         caniID,
			Label:      strVal(cable.Label),
			ObjectMeta: devicetypes.ObjectMeta{Status: strVal(cable.Status.Url)},
			Color:      strVal(cable.Color),
		}

		// Cable type.
		if cable.Type != nil && cable.Type.Value != nil {
			caniCable.CableType = string(*cable.Type.Value)
		}

		// Length.
		if cable.Length != nil {
			l := float64(*cable.Length)
			caniCable.Length = &l
		}
		if cable.LengthUnit != nil && cable.LengthUnit.Value != nil {
			caniCable.LengthUnit = string(*cable.LengthUnit.Value)
		}

		// Resolve termination A.
		termAIfaceID := uuid.UUID(cable.TerminationAId)
		if ref, ok := ifaceMap[termAIfaceID]; ok {
			if caniDevID, ok2 := deviceMap[ref.deviceID]; ok2 {
				caniCable.TerminationADevice = caniDevID
				caniCable.TerminationAPort = ref.name
			}
		}
		caniCable.TerminationAType = cable.TerminationAType

		// Resolve termination B.
		termBIfaceID := uuid.UUID(cable.TerminationBId)
		if ref, ok := ifaceMap[termBIfaceID]; ok {
			if caniDevID, ok2 := deviceMap[ref.deviceID]; ok2 {
				caniCable.TerminationBDevice = caniDevID
				caniCable.TerminationBPort = ref.name
			}
		}
		caniCable.TerminationBType = cable.TerminationBType

		if cable.CustomFields != nil {
			caniCable.CustomFields = *cable.CustomFields
		}

		result[caniID] = caniCable
	}

	return result
}

// ifaceRef holds the device and name for an interface, used during cable mapping.
type ifaceRef struct {
	deviceID uuid.UUID // Nautobot device UUID
	name     string
}

// BuildInterfaceMap creates a lookup from Nautobot interface UUID → (device UUID, name).
func BuildInterfaceMap(ifaces []nautobotapi.Interface) map[uuid.UUID]ifaceRef {
	m := make(map[uuid.UUID]ifaceRef, len(ifaces))
	for _, iface := range ifaces {
		ifaceID := directUUID(iface.Id)
		if ifaceID == uuid.Nil {
			continue
		}
		devID := uuid.Nil
		if iface.Device != nil {
			devID = tenantRefID(iface.Device)
		}
		m[ifaceID] = ifaceRef{deviceID: devID, name: iface.Name}
	}
	return m
}
