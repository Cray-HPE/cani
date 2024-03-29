// This is custom code that is not generated by swagger-codegen
// The SLS Hardware extra proeprteis is context dependent with different structs allowed.
// The code generate does struct compoistion, and things are not well behaved when performing
// JSON unmarshalling if multiple of these structs have the same field which causes these fields
// to be ignored.
package sls_client

import (
	"fmt"

	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/mitchellh/mapstructure"
)

type HardwareExtraProperties interface{}

func (hardware *Hardware) DecodeExtraProperties() (result interface{}, err error) {
	// This can be filled out with types with some help of the following. Doesn't fully work, but gets you close
	// $ cat pkg/sls-common/types.go | grep '^type Comptype' | sort
	switch xnametypes.GetHMSType(hardware.Xname) {
	case xnametypes.NodeBMCNic:
		result = HardwareExtraPropertiesBmcNic{}
	case xnametypes.CDUMgmtSwitch:
		result = HardwareExtraPropertiesCduMgmtSwitch{}
	case xnametypes.CabinetPDUNic:
		result = HardwareExtraPropertiesCabPduNic{}
	case xnametypes.Cabinet:
		result = HardwareExtraPropertiesCabinet{}
	case xnametypes.Chassis:
		result = HardwareExtraPropertiesChassis{}
	case xnametypes.ChassisBMC:
		result = HardwareExtraPropertiesChassisBmc{}
	case xnametypes.ComputeModule:
		result = HardwareExtraPropertiesCompmod{}
	case xnametypes.NodePowerConnector:
		result = HardwareExtraPropertiesCompmodPowerConnector{}
	case xnametypes.NodeHsnNic:
		result = HardwareExtraPropertiesNodeHsnNic{}
	case xnametypes.HSNConnector:
		result = HardwareExtraPropertiesHsnConnector{}
	case xnametypes.MgmtHLSwitch:
		result = HardwareExtraPropertiesMgmtHlSwitch{}
	case xnametypes.MgmtSwitch:
		result = HardwareExtraPropertiesMgmtSwitch{}
	case xnametypes.MgmtSwitchConnector:
		result = HardwareExtraPropertiesMgmtSwitchConnector{}
	case xnametypes.Node:
		result = HardwareExtraPropertiesNode{}
	case xnametypes.NodeBMC:
		result = HardwareExtraPropertiesNcard{}
	case xnametypes.NodeNic:
		result = HardwareExtraPropertiesNodeNic{}
	case xnametypes.RouterBMC:
		result = HardwareExtraPropertiesRtrBmc{}
	case xnametypes.RouterBMCNic:
		result = HardwareExtraPropertiesRtrBmcNic{}
	case xnametypes.RouterModule:
		result = HardwareExtraPropertiesRtrmod{}
	default:
		// Not all SLS types have an associated struct. If EP is nil, then its not a problem.
		if hardware.ExtraProperties == nil {
			return nil, nil
		}

		return nil, fmt.Errorf("hardware object (%s) has unexpected properties of type (%T)", hardware.Xname, hardware.ExtraProperties)
	}

	// Decode the Raw extra properties into a give structure
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToIPHookFunc(),
		Result:     &result,
	})
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(hardware.ExtraProperties)

	return result, err
}
