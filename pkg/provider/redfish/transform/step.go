package transform

import (
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/redfish/import"
	"github.com/Cray-HPE/cani/pkg/visual"
)

const (
	sourceLibrary = "(library)"
	sourceRedfish = "(redfish)"

	// targetTypeCaniDevice labels Redfish field mappings whose target is a CaniDeviceType.
	targetTypeCaniDevice = "CaniDeviceType"
)

// stepInput bundles arguments for building step-through display info.
type stepInput struct {
	Num, Total int
	Root       import_.ServiceRoot
	Dev        *devicetypes.CaniDeviceType
	LibSlug    string
	MatchQuery string
	MatchScore int
}

// buildRootStepInfo constructs a NodeStepInfo showing raw Redfish fields and
// their mappings to CANI types. Used for step-through display.
func buildRootStepInfo(in stepInput) visual.NodeStepInfo {
	mappings := buildFieldMappings(in.Root, in.Dev, in.LibSlug)

	return visual.NodeStepInfo{
		NodeNum:         in.Num,
		Total:           in.Total,
		RawName:         in.Root.Product,
		RawType:         "server",
		RawUUID:         in.Root.UUID,
		FruCount:        0, // ServiceRoot has no FRU data
		Mappings:        mappings,
		LibMatch:        in.LibSlug,
		MatchQuery:      in.MatchQuery,
		MatchScore:      in.MatchScore,
		LibModel:        in.Dev.Model,
		LibManufacturer: in.Dev.Manufacturer,
	}
}

// buildFieldMappings creates field mappings for the step display.
func buildFieldMappings(
	root import_.ServiceRoot,
	dev *devicetypes.CaniDeviceType,
	libSlug string,
) []visual.FieldMapping {
	mappings := []visual.FieldMapping{
		{
			SourceField: "Product",
			SourceValue: root.Product,
			TargetType:  targetTypeCaniDevice,
			TargetField: "Name",
			TargetValue: dev.Name,
		},
		{
			SourceField: "Vendor",
			SourceValue: root.Vendor,
			TargetType:  targetTypeCaniDevice,
			TargetField: "Manufacturer",
			TargetValue: dev.Manufacturer,
		},
		{
			SourceField: "UUID",
			SourceValue: root.UUID,
			TargetType:  targetTypeCaniDevice,
			TargetField: "ProviderMetadata[redfish][redfish_uuid]",
			TargetValue: root.UUID,
		},
		{
			SourceField: "RedfishVersion",
			SourceValue: root.RedfishVersion,
			TargetType:  targetTypeCaniDevice,
			TargetField: "ProviderMetadata[redfish][redfish_version]",
			TargetValue: root.RedfishVersion,
		},
	}

	// Redfish OEM metadata mappings (single arrow, not derived).
	if fqdn := root.ManagerFQDN(); fqdn != "" {
		mappings = append(mappings, visual.FieldMapping{
			SourceField: sourceRedfish,
			SourceValue: "Manager.FQDN",
			TargetType:  targetTypeCaniDevice,
			TargetField: "ProviderMetadata[redfish][bmc_fqdn]",
			TargetValue: fqdn,
		})
	}
	if host := root.ManagerHostName(); host != "" {
		mappings = append(mappings, visual.FieldMapping{
			SourceField: sourceRedfish,
			SourceValue: "Manager.HostName",
			TargetType:  targetTypeCaniDevice,
			TargetField: "ProviderMetadata[redfish][bmc_hostname]",
			TargetValue: host,
		})
	}
	if bmc := root.ManagerType(); bmc != "" {
		mappings = append(mappings, visual.FieldMapping{
			SourceField: sourceRedfish,
			SourceValue: "Manager.ManagerType",
			TargetType:  targetTypeCaniDevice,
			TargetField: "ProviderMetadata[redfish][bmc_type]",
			TargetValue: bmc,
		})
	}
	if fw := root.ManagerFirmwareVersion(); fw != "" {
		mappings = append(mappings, visual.FieldMapping{
			SourceField: sourceRedfish,
			SourceValue: "Manager.ManagerFirmwareVersion",
			TargetType:  targetTypeCaniDevice,
			TargetField: "ProviderMetadata[redfish][bmc_firmware]",
			TargetValue: fw,
		})
	}

	// Library enrichment mappings.
	if libSlug != "" {
		mappings = append(mappings, visual.FieldMapping{
			SourceField: sourceLibrary,
			SourceValue: libSlug,
			TargetType:  targetTypeCaniDevice,
			TargetField: "Slug",
			TargetValue: dev.Slug,
			IsDerived:   true,
		})
		if dev.Model != "" {
			mappings = append(mappings, visual.FieldMapping{
				SourceField: sourceLibrary,
				SourceValue: libSlug,
				TargetType:  targetTypeCaniDevice,
				TargetField: "Model",
				TargetValue: dev.Model,
				IsDerived:   true,
			})
		}
		if dev.PartNumber != "" {
			mappings = append(mappings, visual.FieldMapping{
				SourceField: sourceLibrary,
				SourceValue: libSlug,
				TargetType:  targetTypeCaniDevice,
				TargetField: "PartNumber",
				TargetValue: dev.PartNumber,
				IsDerived:   true,
			})
		}
		if dev.UHeight > 0 {
			mappings = append(mappings, visual.FieldMapping{
				SourceField: sourceLibrary,
				SourceValue: libSlug,
				TargetType:  targetTypeCaniDevice,
				TargetField: "UHeight",
				TargetValue: fmt.Sprintf("%d", dev.UHeight),
				IsDerived:   true,
			})
		}
	}

	return mappings
}
