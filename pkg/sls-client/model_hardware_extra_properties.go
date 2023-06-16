// This is custom code that is not generated by swagger-codegen
// The SLS Hardware extra proeprteis is context dependent with different structs allowed.
// The code generate does struct compoistion, and things are not well behaved when performing
// JSON unmarshalling if multiple of these structs have the same field which causes these fields
// to be ignored.
package sls_client

import (
	"fmt"
	"time"

	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/google/uuid"
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

//
// Interface for setting CANI metadata
//

type CANIMetadata interface {
	SetCaniID(id uuid.UUID)
	SetCaniLastModified(ts time.Time)
	SetCaniSlsSchemaVersion(version string)

	GetCaniID() string
	GetCaniLastModified() string
	GetCaniSlsSchemaVersion() string
}

//
// HardwareExtraPropertiesBmcNic
//

func (ep *HardwareExtraPropertiesBmcNic) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}
func (ep *HardwareExtraPropertiesBmcNic) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesBmcNic) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesBmcNic) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesBmcNic) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesBmcNic) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesCabPduNic
//

func (ep *HardwareExtraPropertiesCabPduNic) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesCabPduNic) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesCabPduNic) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesCabPduNic) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesCabPduNic) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesCabPduNic) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesCabPduPwrConnector
//

func (ep *HardwareExtraPropertiesCabPduPwrConnector) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesCabPduPwrConnector) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesCabPduPwrConnector) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesCabPduPwrConnector) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesCabPduPwrConnector) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesCabPduPwrConnector) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesCabinet
//

func (ep *HardwareExtraPropertiesCabinet) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesCabinet) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesCabinet) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesCabinet) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesCabinet) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesCabinet) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesCduMgmtSwitch
//

func (ep *HardwareExtraPropertiesCduMgmtSwitch) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesCduMgmtSwitch) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesCduMgmtSwitch) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesCduMgmtSwitch) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesCduMgmtSwitch) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesCduMgmtSwitch) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesChassis
//

func (ep *HardwareExtraPropertiesChassis) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesChassis) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesChassis) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesChassis) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesChassis) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesChassis) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesChassisBmc
//

func (ep *HardwareExtraPropertiesChassisBmc) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesChassisBmc) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesChassisBmc) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesChassisBmc) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesChassisBmc) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesChassisBmc) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesCompmod
//

func (ep *HardwareExtraPropertiesCompmod) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesCompmod) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesCompmod) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesCompmod) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesCompmod) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesCompmod) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesCompmodPowerConnector
//

func (ep *HardwareExtraPropertiesCompmodPowerConnector) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesCompmodPowerConnector) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesCompmodPowerConnector) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesCompmodPowerConnector) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesCompmodPowerConnector) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesCompmodPowerConnector) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesHsnConnector
//

func (ep *HardwareExtraPropertiesHsnConnector) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesHsnConnector) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesHsnConnector) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesHsnConnector) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesHsnConnector) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesHsnConnector) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesMgmtSwitch
//

func (ep *HardwareExtraPropertiesMgmtSwitch) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesMgmtSwitch) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesMgmtSwitch) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesMgmtSwitch) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesMgmtSwitch) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesMgmtSwitch) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesMgmtSwitchConnector
//

func (ep *HardwareExtraPropertiesMgmtSwitchConnector) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesMgmtSwitchConnector) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesMgmtSwitchConnector) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesMgmtSwitchConnector) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesMgmtSwitchConnector) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesMgmtSwitchConnector) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesNcard
//

func (ep *HardwareExtraPropertiesNcard) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesNcard) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesNcard) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesNcard) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesNcard) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesNcard) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesNode
//

func (ep *HardwareExtraPropertiesNode) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesNode) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesNode) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesNode) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesNode) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesNode) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesNodeHsnNic
//

func (ep *HardwareExtraPropertiesNodeHsnNic) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesNodeHsnNic) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesNodeHsnNic) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesNodeHsnNic) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesNodeHsnNic) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesNodeHsnNic) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesNodeNic
//

func (ep *HardwareExtraPropertiesNodeNic) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesNodeNic) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesNodeNic) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesNodeNic) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesNodeNic) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesNodeNic) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesRtrBmc
//

func (ep *HardwareExtraPropertiesRtrBmc) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesRtrBmc) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesRtrBmc) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesRtrBmc) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesRtrBmc) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesRtrBmc) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesRtrBmcNic
//

func (ep *HardwareExtraPropertiesRtrBmcNic) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesRtrBmcNic) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesRtrBmcNic) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}
func (ep *HardwareExtraPropertiesRtrBmcNic) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesRtrBmcNic) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesRtrBmcNic) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesRtrmod
//

func (ep *HardwareExtraPropertiesRtrmod) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesRtrmod) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesRtrmod) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesRtrmod) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesRtrmod) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesRtrmod) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}

//
// HardwareExtraPropertiesSystem
//

func (ep *HardwareExtraPropertiesSystem) SetCaniID(id uuid.UUID) {
	ep.CaniId = id.String()
}

func (ep *HardwareExtraPropertiesSystem) SetCaniLastModified(ts time.Time) {
	ep.CaniLastModified = ts.UTC().String()
}

func (ep *HardwareExtraPropertiesSystem) SetCaniSlsSchemaVersion(version string) {
	ep.CaniSlsSchemaVersion = version
}

func (ep *HardwareExtraPropertiesSystem) GetCaniID() string {
	return ep.CaniId
}

func (ep *HardwareExtraPropertiesSystem) GetCaniLastModified() string {
	return ep.CaniLastModified
}

func (ep *HardwareExtraPropertiesSystem) GetCaniSlsSchemaVersion() string {
	return ep.CaniSlsSchemaVersion
}
