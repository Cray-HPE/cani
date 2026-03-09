package transform

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
	"github.com/google/uuid"
)

// processNodes creates a device for each Node, enriched with SMD data.
func processNodes(
	cls []CsmClassification,
	result *devicetypes.TransformResult,
	byXname map[string]uuid.UUID,
	existing *devicetypes.Inventory,
) {
	for _, cl := range cls {
		if cl.CaniType != devicetypes.TypeNode {
			continue
		}
		devID := resolveExistingID(cl.Xname.Raw, existing)
		byXname[cl.Xname.Raw] = devID

		parentID := resolveParent(
			cl.Xname.Parent(), devicetypes.TypeNodeCard,
			cl.Hardware.Class, result, byXname, existing,
		)
		meta := buildNodeMetadata(cl)
		name := cl.Xname.Raw
		if len(meta.Aliases) > 0 {
			name = meta.Aliases[0]
		}

		result.Devices[devID] = &devicetypes.CaniDeviceType{
			ID:               devID,
			Name:             name,
			Type:             devicetypes.TypeNode,
			Status:           "active",
			Role:             meta.Role,
			Parent:           parentID,
			ProviderMetadata: map[string]any{"csm": toProviderMetadata(meta)},
		}
		if slug := resolveSlug(cl); slug != "" {
			_ = devicetypes.ApplyDeviceType(result.Devices[devID], slug)
		}
	}
}

// buildNodeMetadata extracts node metadata from SLS and SMD.
func buildNodeMetadata(cl CsmClassification) CsmMetadata {
	meta := CsmMetadata{
		Xname: cl.Xname.Raw,
		Class: cl.Hardware.Class,
	}
	if cl.Hardware.ExtraProperties != nil {
		ep, err := import_.DecodeExtraProperties[import_.SlsNodeExtraProperties](
			cl.Hardware.ExtraProperties,
		)
		if err == nil {
			meta.NID = ep.NID
			meta.Role = ep.Role
			meta.SubRole = ep.SubRole
			meta.Aliases = ep.Aliases
		}
	}
	if cl.Smd != nil {
		meta.State = cl.Smd.State
		if cl.Smd.NID != 0 {
			meta.NID = cl.Smd.NID
		}
		if cl.Smd.Role != "" {
			meta.Role = cl.Smd.Role
		}
		if cl.Smd.SubRole != "" {
			meta.SubRole = cl.Smd.SubRole
		}
	}
	return meta
}

// processMgmtSwitches creates a device for each MgmtSwitch.
func processMgmtSwitches(
	cls []CsmClassification,
	result *devicetypes.TransformResult,
	byXname map[string]uuid.UUID,
	existing *devicetypes.Inventory,
) {
	for _, cl := range cls {
		if cl.CaniType != devicetypes.TypeMgmtSwitch {
			continue
		}
		devID := resolveExistingID(cl.Xname.Raw, existing)
		byXname[cl.Xname.Raw] = devID
		parentID := resolveParent(
			cl.Xname.Parent(), devicetypes.TypeChassis,
			cl.Hardware.Class, result, byXname, existing,
		)
		dev := newDevice(devID, cl, byXname)
		dev.Parent = parentID
		dev.Name = aliasOrXname(cl, dev.Name)
		result.Devices[devID] = dev
	}
}

// processNetworkSwitches creates devices for HSN and HL switches.
func processNetworkSwitches(
	cls []CsmClassification,
	result *devicetypes.TransformResult,
	byXname map[string]uuid.UUID,
	existing *devicetypes.Inventory,
) {
	for _, cl := range cls {
		if cl.CaniType != devicetypes.TypeHsnSwitch &&
			cl.CaniType != devicetypes.TypeSwitch {
			continue
		}
		devID := resolveExistingID(cl.Xname.Raw, existing)
		byXname[cl.Xname.Raw] = devID
		parentID := resolveParent(
			cl.Xname.Parent(), devicetypes.TypeChassis,
			cl.Hardware.Class, result, byXname, existing,
		)
		dev := newDevice(devID, cl, byXname)
		dev.Parent = parentID
		dev.Name = aliasOrXname(cl, dev.Name)
		result.Devices[devID] = dev
	}
}

// aliasOrXname returns the first alias from extra properties if present.
func aliasOrXname(cl CsmClassification, fallback string) string {
	if cl.Hardware.ExtraProperties == nil {
		return fallback
	}
	// Try MgmtSwitch aliases
	ep, err := import_.DecodeExtraProperties[import_.SlsMgmtSwitchExtraProperties](
		cl.Hardware.ExtraProperties,
	)
	if err == nil && len(ep.Aliases) > 0 {
		return ep.Aliases[0]
	}
	// Try MgmtHLSwitch aliases
	hlep, err := import_.DecodeExtraProperties[import_.SlsMgmtHLSwitchExtraProperties](
		cl.Hardware.ExtraProperties,
	)
	if err == nil && len(hlep.Aliases) > 0 {
		return hlep.Aliases[0]
	}
	return fallback
}

// newDevice creates a CaniDeviceType with standard fields.
// It attempts to resolve a device-type slug from SLS metadata
// and the hardware library so that strict-mode imports succeed.
func newDevice(
	id uuid.UUID,
	cl CsmClassification,
	byXname map[string]uuid.UUID,
) *devicetypes.CaniDeviceType {
	meta := CsmMetadata{Xname: cl.Xname.Raw, Class: cl.Hardware.Class}
	dev := &devicetypes.CaniDeviceType{
		ID:               id,
		Name:             cl.Xname.Raw,
		Type:             cl.CaniType,
		Status:           "active",
		Parent:           byXname[cl.Xname.Parent()],
		ProviderMetadata: map[string]any{"csm": toProviderMetadata(meta)},
	}
	if slug := resolveSlug(cl); slug != "" {
		_ = devicetypes.ApplyDeviceType(dev, slug)
	}
	return dev
}

// resolveParent returns the UUID for a parent xname, creating implicit
// parents up the hierarchy if they don't exist yet.
func resolveParent(
	xname string,
	expectedType devicetypes.Type,
	class string,
	result *devicetypes.TransformResult,
	byXname map[string]uuid.UUID,
	existing *devicetypes.Inventory,
) uuid.UUID {
	if id, ok := byXname[xname]; ok && id != uuid.Nil {
		return id
	}
	return ensureImplicitParent(xname, expectedType, class, result, byXname, existing)
}

// ensureImplicitParent recursively creates missing parent devices.
func ensureImplicitParent(
	xname string,
	expectedType devicetypes.Type,
	class string,
	result *devicetypes.TransformResult,
	byXname map[string]uuid.UUID,
	existing *devicetypes.Inventory,
) uuid.UUID {
	if id, ok := byXname[xname]; ok && id != uuid.Nil {
		return id
	}
	parsed := ParseXname(xname)
	grandparent := parsed.Parent()
	var gpID uuid.UUID
	if grandparent != "" && grandparent != "s0" {
		gpID = resolveParent(
			grandparent, parentTypeFor(expectedType),
			class, result, byXname, existing,
		)
	}
	devID := resolveExistingID(xname, existing)
	byXname[xname] = devID
	meta := CsmMetadata{Xname: xname, Class: class}
	result.Devices[devID] = &devicetypes.CaniDeviceType{
		ID:               devID,
		Name:             xname,
		Type:             expectedType,
		Status:           "active",
		Parent:           gpID,
		ProviderMetadata: map[string]any{"csm": toProviderMetadata(meta)},
	}
	// Resolve a default slug for the implicit parent.
	implicitCl := CsmClassification{
		CaniType: expectedType,
		Xname:    parsed,
		Hardware: import_.SlsHardware{Xname: xname, Class: class},
	}
	if slug := resolveSlug(implicitCl); slug != "" {
		_ = devicetypes.ApplyDeviceType(result.Devices[devID], slug)
	}
	return devID
}

// processCabinetPDUs creates a device for each CabinetPDUController.
func processCabinetPDUs(
	cls []CsmClassification,
	result *devicetypes.TransformResult,
	byXname map[string]uuid.UUID,
	existing *devicetypes.Inventory,
) {
	for _, cl := range cls {
		if cl.CaniType != devicetypes.TypeCabinetPDU {
			continue
		}
		devID := resolveExistingID(cl.Xname.Raw, existing)
		byXname[cl.Xname.Raw] = devID
		parentID := resolveParent(
			cl.Xname.Parent(), devicetypes.TypeCabinet,
			cl.Hardware.Class, result, byXname, existing,
		)
		dev := newDevice(devID, cl, byXname)
		dev.Parent = parentID
		result.Devices[devID] = dev
	}
}

// parentTypeFor returns the CANI type of the parent for a given child type.
func parentTypeFor(childType devicetypes.Type) devicetypes.Type {
	switch childType {
	case devicetypes.TypeNode:
		return devicetypes.TypeNodeCard
	case devicetypes.TypeNodeCard:
		return devicetypes.TypeBlade
	case devicetypes.TypeBlade:
		return devicetypes.TypeChassis
	case devicetypes.TypeChassis:
		return devicetypes.TypeCabinet
	case devicetypes.TypeMgmtSwitch:
		return devicetypes.TypeChassis
	case devicetypes.TypeSwitch:
		return devicetypes.TypeChassis
	case devicetypes.TypeHsnSwitch:
		return devicetypes.TypeChassis
	case devicetypes.TypeCabinetPDU:
		return devicetypes.TypeCabinet
	case devicetypes.TypeCDU:
		return devicetypes.TypeCabinet
	default:
		return devicetypes.TypeChassis
	}
}
