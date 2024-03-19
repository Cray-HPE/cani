package netbox

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

func UnmarshalModuleType(path string) (ModuleType, error) {
	moduleType := ModuleType{}

	// unmarshal the file into a devicetype
	data, err := os.ReadFile(path)
	if err != nil {
		return moduleType, err
	}

	err = yaml.Unmarshal(data, &moduleType)
	if err != nil {
		return moduleType, err
	}

	return moduleType, nil
}

type ModuleType struct {
	// Comments corresponds to the JSON schema field "comments".
	Comments *string `json:"comments,omitempty" yaml:"comments,omitempty" mapstructure:"comments,omitempty"`

	// ConsolePorts corresponds to the JSON schema field "console-ports".
	ConsolePorts []ModuleTypeConsolePortsElem `json:"console-ports,omitempty" yaml:"console-ports,omitempty" mapstructure:"console-ports,omitempty"`

	// ConsoleServerPorts corresponds to the JSON schema field "console-server-ports".
	ConsoleServerPorts []ModuleTypeConsoleServerPortsElem `json:"console-server-ports,omitempty" yaml:"console-server-ports,omitempty" mapstructure:"console-server-ports,omitempty"`

	// Description corresponds to the JSON schema field "description".
	Description *string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`

	// FrontPorts corresponds to the JSON schema field "front-ports".
	FrontPorts []ModuleTypeFrontPortsElem `json:"front-ports,omitempty" yaml:"front-ports,omitempty" mapstructure:"front-ports,omitempty"`

	// Interfaces corresponds to the JSON schema field "interfaces".
	Interfaces []ModuleTypeInterfacesElem `json:"interfaces,omitempty" yaml:"interfaces,omitempty" mapstructure:"interfaces,omitempty"`

	// Manufacturer corresponds to the JSON schema field "manufacturer".
	Manufacturer string `json:"manufacturer" yaml:"manufacturer" mapstructure:"manufacturer"`

	// Model corresponds to the JSON schema field "model".
	Model string `json:"model" yaml:"model" mapstructure:"model"`

	// PartNumber corresponds to the JSON schema field "part_number".
	PartNumber *string `json:"part_number,omitempty" yaml:"part_number,omitempty" mapstructure:"part_number,omitempty"`

	// PowerOutlets corresponds to the JSON schema field "power-outlets".
	PowerOutlets []ModuleTypePowerOutletsElem `json:"power-outlets,omitempty" yaml:"power-outlets,omitempty" mapstructure:"power-outlets,omitempty"`

	// PowerPorts corresponds to the JSON schema field "power-ports".
	PowerPorts []ModuleTypePowerPortsElem `json:"power-ports,omitempty" yaml:"power-ports,omitempty" mapstructure:"power-ports,omitempty"`

	// RearPorts corresponds to the JSON schema field "rear-ports".
	RearPorts []ModuleTypeRearPortsElem `json:"rear-ports,omitempty" yaml:"rear-ports,omitempty" mapstructure:"rear-ports,omitempty"`

	// Weight corresponds to the JSON schema field "weight".
	Weight *float64 `json:"weight,omitempty" yaml:"weight,omitempty" mapstructure:"weight,omitempty"`

	// WeightUnit corresponds to the JSON schema field "weight_unit".
	WeightUnit *ModuleTypeWeightUnit `json:"weight_unit,omitempty" yaml:"weight_unit,omitempty" mapstructure:"weight_unit,omitempty"`
}

type ModuleTypeConsolePortsElem struct {
	// Label corresponds to the JSON schema field "label".
	Label *string `json:"label,omitempty" yaml:"label,omitempty" mapstructure:"label,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// Poe corresponds to the JSON schema field "poe".
	Poe *bool `json:"poe,omitempty" yaml:"poe,omitempty" mapstructure:"poe,omitempty"`

	// Type corresponds to the JSON schema field "type".
	Type ModuleTypeConsolePortsElemType `json:"type" yaml:"type" mapstructure:"type"`
}

type ModuleTypeConsolePortsElemType string

const ModuleTypeConsolePortsElemTypeDb25 ModuleTypeConsolePortsElemType = "db-25"
const ModuleTypeConsolePortsElemTypeDe9 ModuleTypeConsolePortsElemType = "de-9"
const ModuleTypeConsolePortsElemTypeMiniDin8 ModuleTypeConsolePortsElemType = "mini-din-8"
const ModuleTypeConsolePortsElemTypeOther ModuleTypeConsolePortsElemType = "other"
const ModuleTypeConsolePortsElemTypeRj11 ModuleTypeConsolePortsElemType = "rj-11"
const ModuleTypeConsolePortsElemTypeRj12 ModuleTypeConsolePortsElemType = "rj-12"
const ModuleTypeConsolePortsElemTypeRj45 ModuleTypeConsolePortsElemType = "rj-45"
const ModuleTypeConsolePortsElemTypeUsbA ModuleTypeConsolePortsElemType = "usb-a"
const ModuleTypeConsolePortsElemTypeUsbB ModuleTypeConsolePortsElemType = "usb-b"
const ModuleTypeConsolePortsElemTypeUsbC ModuleTypeConsolePortsElemType = "usb-c"
const ModuleTypeConsolePortsElemTypeUsbMicroA ModuleTypeConsolePortsElemType = "usb-micro-a"
const ModuleTypeConsolePortsElemTypeUsbMicroAb ModuleTypeConsolePortsElemType = "usb-micro-ab"
const ModuleTypeConsolePortsElemTypeUsbMicroB ModuleTypeConsolePortsElemType = "usb-micro-b"
const ModuleTypeConsolePortsElemTypeUsbMiniA ModuleTypeConsolePortsElemType = "usb-mini-a"
const ModuleTypeConsolePortsElemTypeUsbMiniB ModuleTypeConsolePortsElemType = "usb-mini-b"

type ModuleTypeConsoleServerPortsElem struct {
	// Label corresponds to the JSON schema field "label".
	Label *string `json:"label,omitempty" yaml:"label,omitempty" mapstructure:"label,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// Type corresponds to the JSON schema field "type".
	Type ModuleTypeConsoleServerPortsElemType `json:"type" yaml:"type" mapstructure:"type"`
}

type ModuleTypeConsoleServerPortsElemType string

const ModuleTypeConsoleServerPortsElemTypeDb25 ModuleTypeConsoleServerPortsElemType = "db-25"
const ModuleTypeConsoleServerPortsElemTypeDe9 ModuleTypeConsoleServerPortsElemType = "de-9"
const ModuleTypeConsoleServerPortsElemTypeMiniDin8 ModuleTypeConsoleServerPortsElemType = "mini-din-8"
const ModuleTypeConsoleServerPortsElemTypeOther ModuleTypeConsoleServerPortsElemType = "other"
const ModuleTypeConsoleServerPortsElemTypeRj11 ModuleTypeConsoleServerPortsElemType = "rj-11"
const ModuleTypeConsoleServerPortsElemTypeRj12 ModuleTypeConsoleServerPortsElemType = "rj-12"
const ModuleTypeConsoleServerPortsElemTypeRj45 ModuleTypeConsoleServerPortsElemType = "rj-45"
const ModuleTypeConsoleServerPortsElemTypeUsbA ModuleTypeConsoleServerPortsElemType = "usb-a"
const ModuleTypeConsoleServerPortsElemTypeUsbB ModuleTypeConsoleServerPortsElemType = "usb-b"
const ModuleTypeConsoleServerPortsElemTypeUsbC ModuleTypeConsoleServerPortsElemType = "usb-c"
const ModuleTypeConsoleServerPortsElemTypeUsbMicroA ModuleTypeConsoleServerPortsElemType = "usb-micro-a"
const ModuleTypeConsoleServerPortsElemTypeUsbMicroAb ModuleTypeConsoleServerPortsElemType = "usb-micro-ab"
const ModuleTypeConsoleServerPortsElemTypeUsbMicroB ModuleTypeConsoleServerPortsElemType = "usb-micro-b"
const ModuleTypeConsoleServerPortsElemTypeUsbMiniA ModuleTypeConsoleServerPortsElemType = "usb-mini-a"
const ModuleTypeConsoleServerPortsElemTypeUsbMiniB ModuleTypeConsoleServerPortsElemType = "usb-mini-b"

type ModuleTypeFrontPortsElem struct {
	// Color corresponds to the JSON schema field "color".
	Color *string `json:"color,omitempty" yaml:"color,omitempty" mapstructure:"color,omitempty"`

	// Label corresponds to the JSON schema field "label".
	Label *string `json:"label,omitempty" yaml:"label,omitempty" mapstructure:"label,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// RearPort corresponds to the JSON schema field "rear_port".
	RearPort string `json:"rear_port" yaml:"rear_port" mapstructure:"rear_port"`

	// RearPortPosition corresponds to the JSON schema field "rear_port_position".
	RearPortPosition *int `json:"rear_port_position,omitempty" yaml:"rear_port_position,omitempty" mapstructure:"rear_port_position,omitempty"`

	// Type corresponds to the JSON schema field "type".
	Type ModuleTypeFrontPortsElemType `json:"type" yaml:"type" mapstructure:"type"`
}

type ModuleTypeFrontPortsElemType string

const ModuleTypeFrontPortsElemTypeA110Punch ModuleTypeFrontPortsElemType = "110-punch"
const ModuleTypeFrontPortsElemTypeA4P2C ModuleTypeFrontPortsElemType = "4p2c"
const ModuleTypeFrontPortsElemTypeA4P4C ModuleTypeFrontPortsElemType = "4p4c"
const ModuleTypeFrontPortsElemTypeA6P2C ModuleTypeFrontPortsElemType = "6p2c"
const ModuleTypeFrontPortsElemTypeA6P4C ModuleTypeFrontPortsElemType = "6p4c"
const ModuleTypeFrontPortsElemTypeA6P6C ModuleTypeFrontPortsElemType = "6p6c"
const ModuleTypeFrontPortsElemTypeA8P2C ModuleTypeFrontPortsElemType = "8p2c"
const ModuleTypeFrontPortsElemTypeA8P4C ModuleTypeFrontPortsElemType = "8p4c"
const ModuleTypeFrontPortsElemTypeA8P6C ModuleTypeFrontPortsElemType = "8p6c"
const ModuleTypeFrontPortsElemTypeA8P8C ModuleTypeFrontPortsElemType = "8p8c"
const ModuleTypeFrontPortsElemTypeBnc ModuleTypeFrontPortsElemType = "bnc"
const ModuleTypeFrontPortsElemTypeCs ModuleTypeFrontPortsElemType = "cs"
const ModuleTypeFrontPortsElemTypeF ModuleTypeFrontPortsElemType = "f"
const ModuleTypeFrontPortsElemTypeFc ModuleTypeFrontPortsElemType = "fc"
const ModuleTypeFrontPortsElemTypeGg45 ModuleTypeFrontPortsElemType = "gg45"
const ModuleTypeFrontPortsElemTypeLc ModuleTypeFrontPortsElemType = "lc"
const ModuleTypeFrontPortsElemTypeLcApc ModuleTypeFrontPortsElemType = "lc-apc"
const ModuleTypeFrontPortsElemTypeLcPc ModuleTypeFrontPortsElemType = "lc-pc"
const ModuleTypeFrontPortsElemTypeLcUpc ModuleTypeFrontPortsElemType = "lc-upc"
const ModuleTypeFrontPortsElemTypeLsh ModuleTypeFrontPortsElemType = "lsh"
const ModuleTypeFrontPortsElemTypeLshApc ModuleTypeFrontPortsElemType = "lsh-apc"
const ModuleTypeFrontPortsElemTypeLshPc ModuleTypeFrontPortsElemType = "lsh-pc"
const ModuleTypeFrontPortsElemTypeLshUpc ModuleTypeFrontPortsElemType = "lsh-upc"
const ModuleTypeFrontPortsElemTypeLx5 ModuleTypeFrontPortsElemType = "lx5"
const ModuleTypeFrontPortsElemTypeLx5Apc ModuleTypeFrontPortsElemType = "lx5-apc"
const ModuleTypeFrontPortsElemTypeLx5Pc ModuleTypeFrontPortsElemType = "lx5-pc"
const ModuleTypeFrontPortsElemTypeLx5Upc ModuleTypeFrontPortsElemType = "lx5-upc"
const ModuleTypeFrontPortsElemTypeMpo ModuleTypeFrontPortsElemType = "mpo"
const ModuleTypeFrontPortsElemTypeMrj21 ModuleTypeFrontPortsElemType = "mrj21"
const ModuleTypeFrontPortsElemTypeMtrj ModuleTypeFrontPortsElemType = "mtrj"
const ModuleTypeFrontPortsElemTypeN ModuleTypeFrontPortsElemType = "n"
const ModuleTypeFrontPortsElemTypeOther ModuleTypeFrontPortsElemType = "other"
const ModuleTypeFrontPortsElemTypeSc ModuleTypeFrontPortsElemType = "sc"
const ModuleTypeFrontPortsElemTypeScApc ModuleTypeFrontPortsElemType = "sc-apc"
const ModuleTypeFrontPortsElemTypeScPc ModuleTypeFrontPortsElemType = "sc-pc"
const ModuleTypeFrontPortsElemTypeScUpc ModuleTypeFrontPortsElemType = "sc-upc"
const ModuleTypeFrontPortsElemTypeSma905 ModuleTypeFrontPortsElemType = "sma-905"
const ModuleTypeFrontPortsElemTypeSma906 ModuleTypeFrontPortsElemType = "sma-906"
const ModuleTypeFrontPortsElemTypeSn ModuleTypeFrontPortsElemType = "sn"
const ModuleTypeFrontPortsElemTypeSplice ModuleTypeFrontPortsElemType = "splice"
const ModuleTypeFrontPortsElemTypeSt ModuleTypeFrontPortsElemType = "st"
const ModuleTypeFrontPortsElemTypeTera1P ModuleTypeFrontPortsElemType = "tera-1p"
const ModuleTypeFrontPortsElemTypeTera2P ModuleTypeFrontPortsElemType = "tera-2p"
const ModuleTypeFrontPortsElemTypeTera4P ModuleTypeFrontPortsElemType = "tera-4p"
const ModuleTypeFrontPortsElemTypeUrmP2 ModuleTypeFrontPortsElemType = "urm-p2"
const ModuleTypeFrontPortsElemTypeUrmP4 ModuleTypeFrontPortsElemType = "urm-p4"
const ModuleTypeFrontPortsElemTypeUrmP8 ModuleTypeFrontPortsElemType = "urm-p8"

type ModuleTypeInterfacesElem struct {
	// Label corresponds to the JSON schema field "label".
	Label *string `json:"label,omitempty" yaml:"label,omitempty" mapstructure:"label,omitempty"`

	// MgmtOnly corresponds to the JSON schema field "mgmt_only".
	MgmtOnly *bool `json:"mgmt_only,omitempty" yaml:"mgmt_only,omitempty" mapstructure:"mgmt_only,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// PoeMode corresponds to the JSON schema field "poe_mode".
	PoeMode *ModuleTypeInterfacesElemPoeMode `json:"poe_mode,omitempty" yaml:"poe_mode,omitempty" mapstructure:"poe_mode,omitempty"`

	// PoeType corresponds to the JSON schema field "poe_type".
	PoeType *ModuleTypeInterfacesElemPoeType `json:"poe_type,omitempty" yaml:"poe_type,omitempty" mapstructure:"poe_type,omitempty"`

	// Type corresponds to the JSON schema field "type".
	Type ModuleTypeInterfacesElemType `json:"type" yaml:"type" mapstructure:"type"`
}

type ModuleTypeInterfacesElemPoeMode string

const ModuleTypeInterfacesElemPoeModePd ModuleTypeInterfacesElemPoeMode = "pd"
const ModuleTypeInterfacesElemPoeModePse ModuleTypeInterfacesElemPoeMode = "pse"

type ModuleTypeInterfacesElemPoeType string

const ModuleTypeInterfacesElemPoeTypePassive24V2Pair ModuleTypeInterfacesElemPoeType = "passive-24v-2pair"
const ModuleTypeInterfacesElemPoeTypePassive24V4Pair ModuleTypeInterfacesElemPoeType = "passive-24v-4pair"
const ModuleTypeInterfacesElemPoeTypePassive48V2Pair ModuleTypeInterfacesElemPoeType = "passive-48v-2pair"
const ModuleTypeInterfacesElemPoeTypePassive48V4Pair ModuleTypeInterfacesElemPoeType = "passive-48v-4pair"
const ModuleTypeInterfacesElemPoeTypeType1Ieee8023Af ModuleTypeInterfacesElemPoeType = "type1-ieee802.3af"
const ModuleTypeInterfacesElemPoeTypeType2Ieee8023At ModuleTypeInterfacesElemPoeType = "type2-ieee802.3at"
const ModuleTypeInterfacesElemPoeTypeType3Ieee8023Bt ModuleTypeInterfacesElemPoeType = "type3-ieee802.3bt"
const ModuleTypeInterfacesElemPoeTypeType4Ieee8023Bt ModuleTypeInterfacesElemPoeType = "type4-ieee802.3bt"

type ModuleTypeInterfacesElemType string

const ModuleTypeInterfacesElemTypeA1000BaseKx ModuleTypeInterfacesElemType = "1000base-kx"
const ModuleTypeInterfacesElemTypeA1000BaseT ModuleTypeInterfacesElemType = "1000base-t"
const ModuleTypeInterfacesElemTypeA1000BaseXGbic ModuleTypeInterfacesElemType = "1000base-x-gbic"
const ModuleTypeInterfacesElemTypeA1000BaseXSfp ModuleTypeInterfacesElemType = "1000base-x-sfp"
const ModuleTypeInterfacesElemTypeA100BaseFx ModuleTypeInterfacesElemType = "100base-fx"
const ModuleTypeInterfacesElemTypeA100BaseLfx ModuleTypeInterfacesElemType = "100base-lfx"
const ModuleTypeInterfacesElemTypeA100BaseT1 ModuleTypeInterfacesElemType = "100base-t1"
const ModuleTypeInterfacesElemTypeA100BaseTx ModuleTypeInterfacesElemType = "100base-tx"
const ModuleTypeInterfacesElemTypeA100GbaseKp4 ModuleTypeInterfacesElemType = "100gbase-kp4"
const ModuleTypeInterfacesElemTypeA100GbaseKr2 ModuleTypeInterfacesElemType = "100gbase-kr2"
const ModuleTypeInterfacesElemTypeA100GbaseKr4 ModuleTypeInterfacesElemType = "100gbase-kr4"
const ModuleTypeInterfacesElemTypeA100GbaseXCfp ModuleTypeInterfacesElemType = "100gbase-x-cfp"
const ModuleTypeInterfacesElemTypeA100GbaseXCfp2 ModuleTypeInterfacesElemType = "100gbase-x-cfp2"
const ModuleTypeInterfacesElemTypeA100GbaseXCfp4 ModuleTypeInterfacesElemType = "100gbase-x-cfp4"
const ModuleTypeInterfacesElemTypeA100GbaseXCpak ModuleTypeInterfacesElemType = "100gbase-x-cpak"
const ModuleTypeInterfacesElemTypeA100GbaseXCxp ModuleTypeInterfacesElemType = "100gbase-x-cxp"
const ModuleTypeInterfacesElemTypeA100GbaseXDsfp ModuleTypeInterfacesElemType = "100gbase-x-dsfp"
const ModuleTypeInterfacesElemTypeA100GbaseXQsfp28 ModuleTypeInterfacesElemType = "100gbase-x-qsfp28"
const ModuleTypeInterfacesElemTypeA100GbaseXQsfpdd ModuleTypeInterfacesElemType = "100gbase-x-qsfpdd"
const ModuleTypeInterfacesElemTypeA100GbaseXSfpdd ModuleTypeInterfacesElemType = "100gbase-x-sfpdd"
const ModuleTypeInterfacesElemTypeA10GEpon ModuleTypeInterfacesElemType = "10g-epon"
const ModuleTypeInterfacesElemTypeA10GbaseCx4 ModuleTypeInterfacesElemType = "10gbase-cx4"
const ModuleTypeInterfacesElemTypeA10GbaseKr ModuleTypeInterfacesElemType = "10gbase-kr"
const ModuleTypeInterfacesElemTypeA10GbaseKx4 ModuleTypeInterfacesElemType = "10gbase-kx4"
const ModuleTypeInterfacesElemTypeA10GbaseT ModuleTypeInterfacesElemType = "10gbase-t"
const ModuleTypeInterfacesElemTypeA10GbaseXSfpp ModuleTypeInterfacesElemType = "10gbase-x-sfpp"
const ModuleTypeInterfacesElemTypeA10GbaseXX2 ModuleTypeInterfacesElemType = "10gbase-x-x2"
const ModuleTypeInterfacesElemTypeA10GbaseXXenpak ModuleTypeInterfacesElemType = "10gbase-x-xenpak"
const ModuleTypeInterfacesElemTypeA10GbaseXXfp ModuleTypeInterfacesElemType = "10gbase-x-xfp"
const ModuleTypeInterfacesElemTypeA128GfcQsfp28 ModuleTypeInterfacesElemType = "128gfc-qsfp28"
const ModuleTypeInterfacesElemTypeA16GfcSfpp ModuleTypeInterfacesElemType = "16gfc-sfpp"
const ModuleTypeInterfacesElemTypeA1GfcSfp ModuleTypeInterfacesElemType = "1gfc-sfp"
const ModuleTypeInterfacesElemTypeA200GbaseXCfp2 ModuleTypeInterfacesElemType = "200gbase-x-cfp2"
const ModuleTypeInterfacesElemTypeA200GbaseXQsfp56 ModuleTypeInterfacesElemType = "200gbase-x-qsfp56"
const ModuleTypeInterfacesElemTypeA200GbaseXQsfpdd ModuleTypeInterfacesElemType = "200gbase-x-qsfpdd"
const ModuleTypeInterfacesElemTypeA25GbaseKr ModuleTypeInterfacesElemType = "25gbase-kr"
const ModuleTypeInterfacesElemTypeA25GbaseT ModuleTypeInterfacesElemType = "2.5gbase-t"
const ModuleTypeInterfacesElemTypeA25GbaseXSfp28 ModuleTypeInterfacesElemType = "25gbase-x-sfp28"
const ModuleTypeInterfacesElemTypeA2GfcSfp ModuleTypeInterfacesElemType = "2gfc-sfp"
const ModuleTypeInterfacesElemTypeA32GfcSfp28 ModuleTypeInterfacesElemType = "32gfc-sfp28"
const ModuleTypeInterfacesElemTypeA400GbaseXCdfp ModuleTypeInterfacesElemType = "400gbase-x-cdfp"
const ModuleTypeInterfacesElemTypeA400GbaseXCfp2 ModuleTypeInterfacesElemType = "400gbase-x-cfp2"
const ModuleTypeInterfacesElemTypeA400GbaseXCfp8 ModuleTypeInterfacesElemType = "400gbase-x-cfp8"
const ModuleTypeInterfacesElemTypeA400GbaseXOsfp ModuleTypeInterfacesElemType = "400gbase-x-osfp"
const ModuleTypeInterfacesElemTypeA400GbaseXOsfpRhs ModuleTypeInterfacesElemType = "400gbase-x-osfp-rhs"
const ModuleTypeInterfacesElemTypeA400GbaseXQsfp112 ModuleTypeInterfacesElemType = "400gbase-x-qsfp112"
const ModuleTypeInterfacesElemTypeA400GbaseXQsfpdd ModuleTypeInterfacesElemType = "400gbase-x-qsfpdd"
const ModuleTypeInterfacesElemTypeA40GbaseKr4 ModuleTypeInterfacesElemType = "40gbase-kr4"
const ModuleTypeInterfacesElemTypeA40GbaseXQsfpp ModuleTypeInterfacesElemType = "40gbase-x-qsfpp"
const ModuleTypeInterfacesElemTypeA4GfcSfp ModuleTypeInterfacesElemType = "4gfc-sfp"
const ModuleTypeInterfacesElemTypeA50GbaseKr ModuleTypeInterfacesElemType = "50gbase-kr"
const ModuleTypeInterfacesElemTypeA50GbaseXSfp28 ModuleTypeInterfacesElemType = "50gbase-x-sfp28"
const ModuleTypeInterfacesElemTypeA50GbaseXSfp56 ModuleTypeInterfacesElemType = "50gbase-x-sfp56"
const ModuleTypeInterfacesElemTypeA5GbaseT ModuleTypeInterfacesElemType = "5gbase-t"
const ModuleTypeInterfacesElemTypeA64GfcQsfpp ModuleTypeInterfacesElemType = "64gfc-qsfpp"
const ModuleTypeInterfacesElemTypeA800GbaseXOsfp ModuleTypeInterfacesElemType = "800gbase-x-osfp"
const ModuleTypeInterfacesElemTypeA800GbaseXQsfpdd ModuleTypeInterfacesElemType = "800gbase-x-qsfpdd"
const ModuleTypeInterfacesElemTypeA8GfcSfpp ModuleTypeInterfacesElemType = "8gfc-sfpp"
const ModuleTypeInterfacesElemTypeBridge ModuleTypeInterfacesElemType = "bridge"
const ModuleTypeInterfacesElemTypeCdma ModuleTypeInterfacesElemType = "cdma"
const ModuleTypeInterfacesElemTypeCiscoFlexstack ModuleTypeInterfacesElemType = "cisco-flexstack"
const ModuleTypeInterfacesElemTypeCiscoFlexstackPlus ModuleTypeInterfacesElemType = "cisco-flexstack-plus"
const ModuleTypeInterfacesElemTypeCiscoStackwise ModuleTypeInterfacesElemType = "cisco-stackwise"
const ModuleTypeInterfacesElemTypeCiscoStackwise160 ModuleTypeInterfacesElemType = "cisco-stackwise-160"
const ModuleTypeInterfacesElemTypeCiscoStackwise1T ModuleTypeInterfacesElemType = "cisco-stackwise-1t"
const ModuleTypeInterfacesElemTypeCiscoStackwise320 ModuleTypeInterfacesElemType = "cisco-stackwise-320"
const ModuleTypeInterfacesElemTypeCiscoStackwise480 ModuleTypeInterfacesElemType = "cisco-stackwise-480"
const ModuleTypeInterfacesElemTypeCiscoStackwise80 ModuleTypeInterfacesElemType = "cisco-stackwise-80"
const ModuleTypeInterfacesElemTypeCiscoStackwisePlus ModuleTypeInterfacesElemType = "cisco-stackwise-plus"
const ModuleTypeInterfacesElemTypeDocsis ModuleTypeInterfacesElemType = "docsis"
const ModuleTypeInterfacesElemTypeE1 ModuleTypeInterfacesElemType = "e1"
const ModuleTypeInterfacesElemTypeE3 ModuleTypeInterfacesElemType = "e3"
const ModuleTypeInterfacesElemTypeEpon ModuleTypeInterfacesElemType = "epon"
const ModuleTypeInterfacesElemTypeExtremeSummitstack ModuleTypeInterfacesElemType = "extreme-summitstack"
const ModuleTypeInterfacesElemTypeExtremeSummitstack128 ModuleTypeInterfacesElemType = "extreme-summitstack-128"
const ModuleTypeInterfacesElemTypeExtremeSummitstack256 ModuleTypeInterfacesElemType = "extreme-summitstack-256"
const ModuleTypeInterfacesElemTypeExtremeSummitstack512 ModuleTypeInterfacesElemType = "extreme-summitstack-512"
const ModuleTypeInterfacesElemTypeGpon ModuleTypeInterfacesElemType = "gpon"
const ModuleTypeInterfacesElemTypeGsm ModuleTypeInterfacesElemType = "gsm"
const ModuleTypeInterfacesElemTypeIeee80211A ModuleTypeInterfacesElemType = "ieee802.11a"
const ModuleTypeInterfacesElemTypeIeee80211Ac ModuleTypeInterfacesElemType = "ieee802.11ac"
const ModuleTypeInterfacesElemTypeIeee80211Ad ModuleTypeInterfacesElemType = "ieee802.11ad"
const ModuleTypeInterfacesElemTypeIeee80211Ax ModuleTypeInterfacesElemType = "ieee802.11ax"
const ModuleTypeInterfacesElemTypeIeee80211Ay ModuleTypeInterfacesElemType = "ieee802.11ay"
const ModuleTypeInterfacesElemTypeIeee80211G ModuleTypeInterfacesElemType = "ieee802.11g"
const ModuleTypeInterfacesElemTypeIeee80211N ModuleTypeInterfacesElemType = "ieee802.11n"
const ModuleTypeInterfacesElemTypeIeee802151 ModuleTypeInterfacesElemType = "ieee802.15.1"
const ModuleTypeInterfacesElemTypeInfinibandDdr ModuleTypeInterfacesElemType = "infiniband-ddr"
const ModuleTypeInterfacesElemTypeInfinibandEdr ModuleTypeInterfacesElemType = "infiniband-edr"
const ModuleTypeInterfacesElemTypeInfinibandFdr ModuleTypeInterfacesElemType = "infiniband-fdr"
const ModuleTypeInterfacesElemTypeInfinibandFdr10 ModuleTypeInterfacesElemType = "infiniband-fdr10"
const ModuleTypeInterfacesElemTypeInfinibandHdr ModuleTypeInterfacesElemType = "infiniband-hdr"
const ModuleTypeInterfacesElemTypeInfinibandNdr ModuleTypeInterfacesElemType = "infiniband-ndr"
const ModuleTypeInterfacesElemTypeInfinibandQdr ModuleTypeInterfacesElemType = "infiniband-qdr"
const ModuleTypeInterfacesElemTypeInfinibandSdr ModuleTypeInterfacesElemType = "infiniband-sdr"
const ModuleTypeInterfacesElemTypeInfinibandXdr ModuleTypeInterfacesElemType = "infiniband-xdr"
const ModuleTypeInterfacesElemTypeJuniperVcp ModuleTypeInterfacesElemType = "juniper-vcp"
const ModuleTypeInterfacesElemTypeLag ModuleTypeInterfacesElemType = "lag"
const ModuleTypeInterfacesElemTypeLte ModuleTypeInterfacesElemType = "lte"
const ModuleTypeInterfacesElemTypeNgPon2 ModuleTypeInterfacesElemType = "ng-pon2"
const ModuleTypeInterfacesElemTypeOther ModuleTypeInterfacesElemType = "other"
const ModuleTypeInterfacesElemTypeOtherWireless ModuleTypeInterfacesElemType = "other-wireless"
const ModuleTypeInterfacesElemTypeSonetOc12 ModuleTypeInterfacesElemType = "sonet-oc12"
const ModuleTypeInterfacesElemTypeSonetOc192 ModuleTypeInterfacesElemType = "sonet-oc192"
const ModuleTypeInterfacesElemTypeSonetOc1920 ModuleTypeInterfacesElemType = "sonet-oc1920"
const ModuleTypeInterfacesElemTypeSonetOc3 ModuleTypeInterfacesElemType = "sonet-oc3"
const ModuleTypeInterfacesElemTypeSonetOc3840 ModuleTypeInterfacesElemType = "sonet-oc3840"
const ModuleTypeInterfacesElemTypeSonetOc48 ModuleTypeInterfacesElemType = "sonet-oc48"
const ModuleTypeInterfacesElemTypeSonetOc768 ModuleTypeInterfacesElemType = "sonet-oc768"
const ModuleTypeInterfacesElemTypeT1 ModuleTypeInterfacesElemType = "t1"
const ModuleTypeInterfacesElemTypeT3 ModuleTypeInterfacesElemType = "t3"
const ModuleTypeInterfacesElemTypeVirtual ModuleTypeInterfacesElemType = "virtual"
const ModuleTypeInterfacesElemTypeXdsl ModuleTypeInterfacesElemType = "xdsl"
const ModuleTypeInterfacesElemTypeXgPon ModuleTypeInterfacesElemType = "xg-pon"
const ModuleTypeInterfacesElemTypeXgsPon ModuleTypeInterfacesElemType = "xgs-pon"

type ModuleTypePowerOutletsElem struct {
	// FeedLeg corresponds to the JSON schema field "feed_leg".
	FeedLeg *ModuleTypePowerOutletsElemFeedLeg `json:"feed_leg,omitempty" yaml:"feed_leg,omitempty" mapstructure:"feed_leg,omitempty"`

	// Label corresponds to the JSON schema field "label".
	Label *string `json:"label,omitempty" yaml:"label,omitempty" mapstructure:"label,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// PowerPort corresponds to the JSON schema field "power_port".
	PowerPort *string `json:"power_port,omitempty" yaml:"power_port,omitempty" mapstructure:"power_port,omitempty"`

	// Type corresponds to the JSON schema field "type".
	Type ModuleTypePowerOutletsElemType `json:"type" yaml:"type" mapstructure:"type"`
}

type ModuleTypePowerOutletsElemFeedLeg string

const ModuleTypePowerOutletsElemFeedLegA ModuleTypePowerOutletsElemFeedLeg = "A"
const ModuleTypePowerOutletsElemFeedLegB ModuleTypePowerOutletsElemFeedLeg = "B"
const ModuleTypePowerOutletsElemFeedLegC ModuleTypePowerOutletsElemFeedLeg = "C"

type ModuleTypePowerOutletsElemType string

const ModuleTypePowerOutletsElemTypeCS6360C ModuleTypePowerOutletsElemType = "CS6360C"
const ModuleTypePowerOutletsElemTypeCS6364C ModuleTypePowerOutletsElemType = "CS6364C"
const ModuleTypePowerOutletsElemTypeCS8164C ModuleTypePowerOutletsElemType = "CS8164C"
const ModuleTypePowerOutletsElemTypeCS8264C ModuleTypePowerOutletsElemType = "CS8264C"
const ModuleTypePowerOutletsElemTypeCS8364C ModuleTypePowerOutletsElemType = "CS8364C"
const ModuleTypePowerOutletsElemTypeCS8464C ModuleTypePowerOutletsElemType = "CS8464C"
const ModuleTypePowerOutletsElemTypeDcTerminal ModuleTypePowerOutletsElemType = "dc-terminal"
const ModuleTypePowerOutletsElemTypeHardwired ModuleTypePowerOutletsElemType = "hardwired"
const ModuleTypePowerOutletsElemTypeHdotCx ModuleTypePowerOutletsElemType = "hdot-cx"
const ModuleTypePowerOutletsElemTypeIec603092PE4H ModuleTypePowerOutletsElemType = "iec-60309-2p-e-4h"
const ModuleTypePowerOutletsElemTypeIec603092PE6H ModuleTypePowerOutletsElemType = "iec-60309-2p-e-6h"
const ModuleTypePowerOutletsElemTypeIec603092PE9H ModuleTypePowerOutletsElemType = "iec-60309-2p-e-9h"
const ModuleTypePowerOutletsElemTypeIec603093PE4H ModuleTypePowerOutletsElemType = "iec-60309-3p-e-4h"
const ModuleTypePowerOutletsElemTypeIec603093PE6H ModuleTypePowerOutletsElemType = "iec-60309-3p-e-6h"
const ModuleTypePowerOutletsElemTypeIec603093PE9H ModuleTypePowerOutletsElemType = "iec-60309-3p-e-9h"
const ModuleTypePowerOutletsElemTypeIec603093PNE4H ModuleTypePowerOutletsElemType = "iec-60309-3p-n-e-4h"
const ModuleTypePowerOutletsElemTypeIec603093PNE6H ModuleTypePowerOutletsElemType = "iec-60309-3p-n-e-6h"
const ModuleTypePowerOutletsElemTypeIec603093PNE9H ModuleTypePowerOutletsElemType = "iec-60309-3p-n-e-9h"
const ModuleTypePowerOutletsElemTypeIec60309PNE4H ModuleTypePowerOutletsElemType = "iec-60309-p-n-e-4h"
const ModuleTypePowerOutletsElemTypeIec60309PNE6H ModuleTypePowerOutletsElemType = "iec-60309-p-n-e-6h"
const ModuleTypePowerOutletsElemTypeIec60309PNE9H ModuleTypePowerOutletsElemType = "iec-60309-p-n-e-9h"
const ModuleTypePowerOutletsElemTypeIec60320C13 ModuleTypePowerOutletsElemType = "iec-60320-c13"
const ModuleTypePowerOutletsElemTypeIec60320C15 ModuleTypePowerOutletsElemType = "iec-60320-c15"
const ModuleTypePowerOutletsElemTypeIec60320C19 ModuleTypePowerOutletsElemType = "iec-60320-c19"
const ModuleTypePowerOutletsElemTypeIec60320C21 ModuleTypePowerOutletsElemType = "iec-60320-c21"
const ModuleTypePowerOutletsElemTypeIec60320C5 ModuleTypePowerOutletsElemType = "iec-60320-c5"
const ModuleTypePowerOutletsElemTypeIec60320C7 ModuleTypePowerOutletsElemType = "iec-60320-c7"
const ModuleTypePowerOutletsElemTypeIec609061 ModuleTypePowerOutletsElemType = "iec-60906-1"
const ModuleTypePowerOutletsElemTypeItaE ModuleTypePowerOutletsElemType = "ita-e"
const ModuleTypePowerOutletsElemTypeItaF ModuleTypePowerOutletsElemType = "ita-f"
const ModuleTypePowerOutletsElemTypeItaG ModuleTypePowerOutletsElemType = "ita-g"
const ModuleTypePowerOutletsElemTypeItaH ModuleTypePowerOutletsElemType = "ita-h"
const ModuleTypePowerOutletsElemTypeItaI ModuleTypePowerOutletsElemType = "ita-i"
const ModuleTypePowerOutletsElemTypeItaJ ModuleTypePowerOutletsElemType = "ita-j"
const ModuleTypePowerOutletsElemTypeItaK ModuleTypePowerOutletsElemType = "ita-k"
const ModuleTypePowerOutletsElemTypeItaL ModuleTypePowerOutletsElemType = "ita-l"
const ModuleTypePowerOutletsElemTypeItaM ModuleTypePowerOutletsElemType = "ita-m"
const ModuleTypePowerOutletsElemTypeItaMultistandard ModuleTypePowerOutletsElemType = "ita-multistandard"
const ModuleTypePowerOutletsElemTypeItaN ModuleTypePowerOutletsElemType = "ita-n"
const ModuleTypePowerOutletsElemTypeItaO ModuleTypePowerOutletsElemType = "ita-o"
const ModuleTypePowerOutletsElemTypeNbr1413610A ModuleTypePowerOutletsElemType = "nbr-14136-10a"
const ModuleTypePowerOutletsElemTypeNbr1413620A ModuleTypePowerOutletsElemType = "nbr-14136-20a"
const ModuleTypePowerOutletsElemTypeNema1030R ModuleTypePowerOutletsElemType = "nema-10-30r"
const ModuleTypePowerOutletsElemTypeNema1050R ModuleTypePowerOutletsElemType = "nema-10-50r"
const ModuleTypePowerOutletsElemTypeNema115R ModuleTypePowerOutletsElemType = "nema-1-15r"
const ModuleTypePowerOutletsElemTypeNema1420R ModuleTypePowerOutletsElemType = "nema-14-20r"
const ModuleTypePowerOutletsElemTypeNema1430R ModuleTypePowerOutletsElemType = "nema-14-30r"
const ModuleTypePowerOutletsElemTypeNema1450R ModuleTypePowerOutletsElemType = "nema-14-50r"
const ModuleTypePowerOutletsElemTypeNema1460R ModuleTypePowerOutletsElemType = "nema-14-60r"
const ModuleTypePowerOutletsElemTypeNema1515R ModuleTypePowerOutletsElemType = "nema-15-15r"
const ModuleTypePowerOutletsElemTypeNema1520R ModuleTypePowerOutletsElemType = "nema-15-20r"
const ModuleTypePowerOutletsElemTypeNema1530R ModuleTypePowerOutletsElemType = "nema-15-30r"
const ModuleTypePowerOutletsElemTypeNema1550R ModuleTypePowerOutletsElemType = "nema-15-50r"
const ModuleTypePowerOutletsElemTypeNema1560R ModuleTypePowerOutletsElemType = "nema-15-60r"
const ModuleTypePowerOutletsElemTypeNema515R ModuleTypePowerOutletsElemType = "nema-5-15r"
const ModuleTypePowerOutletsElemTypeNema520R ModuleTypePowerOutletsElemType = "nema-5-20r"
const ModuleTypePowerOutletsElemTypeNema530R ModuleTypePowerOutletsElemType = "nema-5-30r"
const ModuleTypePowerOutletsElemTypeNema550R ModuleTypePowerOutletsElemType = "nema-5-50r"
const ModuleTypePowerOutletsElemTypeNema615R ModuleTypePowerOutletsElemType = "nema-6-15r"
const ModuleTypePowerOutletsElemTypeNema620R ModuleTypePowerOutletsElemType = "nema-6-20r"
const ModuleTypePowerOutletsElemTypeNema630R ModuleTypePowerOutletsElemType = "nema-6-30r"
const ModuleTypePowerOutletsElemTypeNema650R ModuleTypePowerOutletsElemType = "nema-6-50r"
const ModuleTypePowerOutletsElemTypeNemaL1030R ModuleTypePowerOutletsElemType = "nema-l10-30r"
const ModuleTypePowerOutletsElemTypeNemaL115R ModuleTypePowerOutletsElemType = "nema-l1-15r"
const ModuleTypePowerOutletsElemTypeNemaL1420R ModuleTypePowerOutletsElemType = "nema-l14-20r"
const ModuleTypePowerOutletsElemTypeNemaL1430R ModuleTypePowerOutletsElemType = "nema-l14-30r"
const ModuleTypePowerOutletsElemTypeNemaL1450R ModuleTypePowerOutletsElemType = "nema-l14-50r"
const ModuleTypePowerOutletsElemTypeNemaL1460R ModuleTypePowerOutletsElemType = "nema-l14-60r"
const ModuleTypePowerOutletsElemTypeNemaL1520R ModuleTypePowerOutletsElemType = "nema-l15-20r"
const ModuleTypePowerOutletsElemTypeNemaL1530R ModuleTypePowerOutletsElemType = "nema-l15-30r"
const ModuleTypePowerOutletsElemTypeNemaL1550R ModuleTypePowerOutletsElemType = "nema-l15-50r"
const ModuleTypePowerOutletsElemTypeNemaL1560R ModuleTypePowerOutletsElemType = "nema-l15-60r"
const ModuleTypePowerOutletsElemTypeNemaL2120R ModuleTypePowerOutletsElemType = "nema-l21-20r"
const ModuleTypePowerOutletsElemTypeNemaL2130R ModuleTypePowerOutletsElemType = "nema-l21-30r"
const ModuleTypePowerOutletsElemTypeNemaL2230R ModuleTypePowerOutletsElemType = "nema-l22-30r"
const ModuleTypePowerOutletsElemTypeNemaL515R ModuleTypePowerOutletsElemType = "nema-l5-15r"
const ModuleTypePowerOutletsElemTypeNemaL520R ModuleTypePowerOutletsElemType = "nema-l5-20r"
const ModuleTypePowerOutletsElemTypeNemaL530R ModuleTypePowerOutletsElemType = "nema-l5-30r"
const ModuleTypePowerOutletsElemTypeNemaL550R ModuleTypePowerOutletsElemType = "nema-l5-50r"
const ModuleTypePowerOutletsElemTypeNemaL615R ModuleTypePowerOutletsElemType = "nema-l6-15r"
const ModuleTypePowerOutletsElemTypeNemaL620R ModuleTypePowerOutletsElemType = "nema-l6-20r"
const ModuleTypePowerOutletsElemTypeNemaL630R ModuleTypePowerOutletsElemType = "nema-l6-30r"
const ModuleTypePowerOutletsElemTypeNemaL650R ModuleTypePowerOutletsElemType = "nema-l6-50r"
const ModuleTypePowerOutletsElemTypeNeutrikPowercon20A ModuleTypePowerOutletsElemType = "neutrik-powercon-20a"
const ModuleTypePowerOutletsElemTypeNeutrikPowercon32A ModuleTypePowerOutletsElemType = "neutrik-powercon-32a"
const ModuleTypePowerOutletsElemTypeNeutrikPowerconTrue1 ModuleTypePowerOutletsElemType = "neutrik-powercon-true1"
const ModuleTypePowerOutletsElemTypeNeutrikPowerconTrue1Top ModuleTypePowerOutletsElemType = "neutrik-powercon-true1-top"
const ModuleTypePowerOutletsElemTypeOther ModuleTypePowerOutletsElemType = "other"
const ModuleTypePowerOutletsElemTypeSafDGrid ModuleTypePowerOutletsElemType = "saf-d-grid"
const ModuleTypePowerOutletsElemTypeUbiquitiSmartpower ModuleTypePowerOutletsElemType = "ubiquiti-smartpower"
const ModuleTypePowerOutletsElemTypeUsbA ModuleTypePowerOutletsElemType = "usb-a"
const ModuleTypePowerOutletsElemTypeUsbC ModuleTypePowerOutletsElemType = "usb-c"
const ModuleTypePowerOutletsElemTypeUsbMicroB ModuleTypePowerOutletsElemType = "usb-micro-b"

type ModuleTypePowerPortsElem struct {
	// AllocatedDraw corresponds to the JSON schema field "allocated_draw".
	AllocatedDraw *int `json:"allocated_draw,omitempty" yaml:"allocated_draw,omitempty" mapstructure:"allocated_draw,omitempty"`

	// Label corresponds to the JSON schema field "label".
	Label *string `json:"label,omitempty" yaml:"label,omitempty" mapstructure:"label,omitempty"`

	// MaximumDraw corresponds to the JSON schema field "maximum_draw".
	MaximumDraw *int `json:"maximum_draw,omitempty" yaml:"maximum_draw,omitempty" mapstructure:"maximum_draw,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// Type corresponds to the JSON schema field "type".
	Type ModuleTypePowerPortsElemType `json:"type" yaml:"type" mapstructure:"type"`
}

type ModuleTypePowerPortsElemType string

const ModuleTypePowerPortsElemTypeCs6361C ModuleTypePowerPortsElemType = "cs6361c"
const ModuleTypePowerPortsElemTypeCs6365C ModuleTypePowerPortsElemType = "cs6365c"
const ModuleTypePowerPortsElemTypeCs8165C ModuleTypePowerPortsElemType = "cs8165c"
const ModuleTypePowerPortsElemTypeCs8265C ModuleTypePowerPortsElemType = "cs8265c"
const ModuleTypePowerPortsElemTypeCs8365C ModuleTypePowerPortsElemType = "cs8365c"
const ModuleTypePowerPortsElemTypeCs8465C ModuleTypePowerPortsElemType = "cs8465c"
const ModuleTypePowerPortsElemTypeDcTerminal ModuleTypePowerPortsElemType = "dc-terminal"
const ModuleTypePowerPortsElemTypeHardwired ModuleTypePowerPortsElemType = "hardwired"
const ModuleTypePowerPortsElemTypeIec603092PE4H ModuleTypePowerPortsElemType = "iec-60309-2p-e-4h"
const ModuleTypePowerPortsElemTypeIec603092PE6H ModuleTypePowerPortsElemType = "iec-60309-2p-e-6h"
const ModuleTypePowerPortsElemTypeIec603092PE9H ModuleTypePowerPortsElemType = "iec-60309-2p-e-9h"
const ModuleTypePowerPortsElemTypeIec603093PE4H ModuleTypePowerPortsElemType = "iec-60309-3p-e-4h"
const ModuleTypePowerPortsElemTypeIec603093PE6H ModuleTypePowerPortsElemType = "iec-60309-3p-e-6h"
const ModuleTypePowerPortsElemTypeIec603093PE9H ModuleTypePowerPortsElemType = "iec-60309-3p-e-9h"
const ModuleTypePowerPortsElemTypeIec603093PNE4H ModuleTypePowerPortsElemType = "iec-60309-3p-n-e-4h"
const ModuleTypePowerPortsElemTypeIec603093PNE6H ModuleTypePowerPortsElemType = "iec-60309-3p-n-e-6h"
const ModuleTypePowerPortsElemTypeIec603093PNE9H ModuleTypePowerPortsElemType = "iec-60309-3p-n-e-9h"
const ModuleTypePowerPortsElemTypeIec60309PNE4H ModuleTypePowerPortsElemType = "iec-60309-p-n-e-4h"
const ModuleTypePowerPortsElemTypeIec60309PNE6H ModuleTypePowerPortsElemType = "iec-60309-p-n-e-6h"
const ModuleTypePowerPortsElemTypeIec60309PNE9H ModuleTypePowerPortsElemType = "iec-60309-p-n-e-9h"
const ModuleTypePowerPortsElemTypeIec60320C14 ModuleTypePowerPortsElemType = "iec-60320-c14"
const ModuleTypePowerPortsElemTypeIec60320C16 ModuleTypePowerPortsElemType = "iec-60320-c16"
const ModuleTypePowerPortsElemTypeIec60320C20 ModuleTypePowerPortsElemType = "iec-60320-c20"
const ModuleTypePowerPortsElemTypeIec60320C22 ModuleTypePowerPortsElemType = "iec-60320-c22"
const ModuleTypePowerPortsElemTypeIec60320C6 ModuleTypePowerPortsElemType = "iec-60320-c6"
const ModuleTypePowerPortsElemTypeIec60320C8 ModuleTypePowerPortsElemType = "iec-60320-c8"
const ModuleTypePowerPortsElemTypeIec609061 ModuleTypePowerPortsElemType = "iec-60906-1"
const ModuleTypePowerPortsElemTypeItaC ModuleTypePowerPortsElemType = "ita-c"
const ModuleTypePowerPortsElemTypeItaE ModuleTypePowerPortsElemType = "ita-e"
const ModuleTypePowerPortsElemTypeItaEf ModuleTypePowerPortsElemType = "ita-ef"
const ModuleTypePowerPortsElemTypeItaF ModuleTypePowerPortsElemType = "ita-f"
const ModuleTypePowerPortsElemTypeItaG ModuleTypePowerPortsElemType = "ita-g"
const ModuleTypePowerPortsElemTypeItaH ModuleTypePowerPortsElemType = "ita-h"
const ModuleTypePowerPortsElemTypeItaI ModuleTypePowerPortsElemType = "ita-i"
const ModuleTypePowerPortsElemTypeItaJ ModuleTypePowerPortsElemType = "ita-j"
const ModuleTypePowerPortsElemTypeItaK ModuleTypePowerPortsElemType = "ita-k"
const ModuleTypePowerPortsElemTypeItaL ModuleTypePowerPortsElemType = "ita-l"
const ModuleTypePowerPortsElemTypeItaM ModuleTypePowerPortsElemType = "ita-m"
const ModuleTypePowerPortsElemTypeItaN ModuleTypePowerPortsElemType = "ita-n"
const ModuleTypePowerPortsElemTypeItaO ModuleTypePowerPortsElemType = "ita-o"
const ModuleTypePowerPortsElemTypeNbr1413610A ModuleTypePowerPortsElemType = "nbr-14136-10a"
const ModuleTypePowerPortsElemTypeNbr1413620A ModuleTypePowerPortsElemType = "nbr-14136-20a"
const ModuleTypePowerPortsElemTypeNema1030P ModuleTypePowerPortsElemType = "nema-10-30p"
const ModuleTypePowerPortsElemTypeNema1050P ModuleTypePowerPortsElemType = "nema-10-50p"
const ModuleTypePowerPortsElemTypeNema115P ModuleTypePowerPortsElemType = "nema-1-15p"
const ModuleTypePowerPortsElemTypeNema1420P ModuleTypePowerPortsElemType = "nema-14-20p"
const ModuleTypePowerPortsElemTypeNema1430P ModuleTypePowerPortsElemType = "nema-14-30p"
const ModuleTypePowerPortsElemTypeNema1450P ModuleTypePowerPortsElemType = "nema-14-50p"
const ModuleTypePowerPortsElemTypeNema1460P ModuleTypePowerPortsElemType = "nema-14-60p"
const ModuleTypePowerPortsElemTypeNema1515P ModuleTypePowerPortsElemType = "nema-15-15p"
const ModuleTypePowerPortsElemTypeNema1520P ModuleTypePowerPortsElemType = "nema-15-20p"
const ModuleTypePowerPortsElemTypeNema1530P ModuleTypePowerPortsElemType = "nema-15-30p"
const ModuleTypePowerPortsElemTypeNema1550P ModuleTypePowerPortsElemType = "nema-15-50p"
const ModuleTypePowerPortsElemTypeNema1560P ModuleTypePowerPortsElemType = "nema-15-60p"
const ModuleTypePowerPortsElemTypeNema515P ModuleTypePowerPortsElemType = "nema-5-15p"
const ModuleTypePowerPortsElemTypeNema520P ModuleTypePowerPortsElemType = "nema-5-20p"
const ModuleTypePowerPortsElemTypeNema530P ModuleTypePowerPortsElemType = "nema-5-30p"
const ModuleTypePowerPortsElemTypeNema550P ModuleTypePowerPortsElemType = "nema-5-50p"
const ModuleTypePowerPortsElemTypeNema615P ModuleTypePowerPortsElemType = "nema-6-15p"
const ModuleTypePowerPortsElemTypeNema620P ModuleTypePowerPortsElemType = "nema-6-20p"
const ModuleTypePowerPortsElemTypeNema630P ModuleTypePowerPortsElemType = "nema-6-30p"
const ModuleTypePowerPortsElemTypeNema650P ModuleTypePowerPortsElemType = "nema-6-50p"
const ModuleTypePowerPortsElemTypeNemaL1030P ModuleTypePowerPortsElemType = "nema-l10-30p"
const ModuleTypePowerPortsElemTypeNemaL115P ModuleTypePowerPortsElemType = "nema-l1-15p"
const ModuleTypePowerPortsElemTypeNemaL1420P ModuleTypePowerPortsElemType = "nema-l14-20p"
const ModuleTypePowerPortsElemTypeNemaL1430P ModuleTypePowerPortsElemType = "nema-l14-30p"
const ModuleTypePowerPortsElemTypeNemaL1450P ModuleTypePowerPortsElemType = "nema-l14-50p"
const ModuleTypePowerPortsElemTypeNemaL1460P ModuleTypePowerPortsElemType = "nema-l14-60p"
const ModuleTypePowerPortsElemTypeNemaL1520P ModuleTypePowerPortsElemType = "nema-l15-20p"
const ModuleTypePowerPortsElemTypeNemaL1530P ModuleTypePowerPortsElemType = "nema-l15-30p"
const ModuleTypePowerPortsElemTypeNemaL1550P ModuleTypePowerPortsElemType = "nema-l15-50p"
const ModuleTypePowerPortsElemTypeNemaL1560P ModuleTypePowerPortsElemType = "nema-l15-60p"
const ModuleTypePowerPortsElemTypeNemaL2120P ModuleTypePowerPortsElemType = "nema-l21-20p"
const ModuleTypePowerPortsElemTypeNemaL2130P ModuleTypePowerPortsElemType = "nema-l21-30p"
const ModuleTypePowerPortsElemTypeNemaL2230P ModuleTypePowerPortsElemType = "nema-l22-30p"
const ModuleTypePowerPortsElemTypeNemaL515P ModuleTypePowerPortsElemType = "nema-l5-15p"
const ModuleTypePowerPortsElemTypeNemaL520P ModuleTypePowerPortsElemType = "nema-l5-20p"
const ModuleTypePowerPortsElemTypeNemaL530P ModuleTypePowerPortsElemType = "nema-l5-30p"
const ModuleTypePowerPortsElemTypeNemaL550P ModuleTypePowerPortsElemType = "nema-l5-50p"
const ModuleTypePowerPortsElemTypeNemaL615P ModuleTypePowerPortsElemType = "nema-l6-15p"
const ModuleTypePowerPortsElemTypeNemaL620P ModuleTypePowerPortsElemType = "nema-l6-20p"
const ModuleTypePowerPortsElemTypeNemaL630P ModuleTypePowerPortsElemType = "nema-l6-30p"
const ModuleTypePowerPortsElemTypeNemaL650P ModuleTypePowerPortsElemType = "nema-l6-50p"
const ModuleTypePowerPortsElemTypeNeutrikPowercon20 ModuleTypePowerPortsElemType = "neutrik-powercon-20"
const ModuleTypePowerPortsElemTypeNeutrikPowercon32 ModuleTypePowerPortsElemType = "neutrik-powercon-32"
const ModuleTypePowerPortsElemTypeNeutrikPowerconTrue1 ModuleTypePowerPortsElemType = "neutrik-powercon-true1"
const ModuleTypePowerPortsElemTypeNeutrikPowerconTrue1Top ModuleTypePowerPortsElemType = "neutrik-powercon-true1-top"
const ModuleTypePowerPortsElemTypeOther ModuleTypePowerPortsElemType = "other"
const ModuleTypePowerPortsElemTypeSafDGrid ModuleTypePowerPortsElemType = "saf-d-grid"
const ModuleTypePowerPortsElemTypeUbiquitiSmartpower ModuleTypePowerPortsElemType = "ubiquiti-smartpower"
const ModuleTypePowerPortsElemTypeUsb3B ModuleTypePowerPortsElemType = "usb-3-b"
const ModuleTypePowerPortsElemTypeUsb3MicroB ModuleTypePowerPortsElemType = "usb-3-micro-b"
const ModuleTypePowerPortsElemTypeUsbA ModuleTypePowerPortsElemType = "usb-a"
const ModuleTypePowerPortsElemTypeUsbB ModuleTypePowerPortsElemType = "usb-b"
const ModuleTypePowerPortsElemTypeUsbC ModuleTypePowerPortsElemType = "usb-c"
const ModuleTypePowerPortsElemTypeUsbMicroA ModuleTypePowerPortsElemType = "usb-micro-a"
const ModuleTypePowerPortsElemTypeUsbMicroAb ModuleTypePowerPortsElemType = "usb-micro-ab"
const ModuleTypePowerPortsElemTypeUsbMicroB ModuleTypePowerPortsElemType = "usb-micro-b"
const ModuleTypePowerPortsElemTypeUsbMiniA ModuleTypePowerPortsElemType = "usb-mini-a"
const ModuleTypePowerPortsElemTypeUsbMiniB ModuleTypePowerPortsElemType = "usb-mini-b"

type ModuleTypeRearPortsElem struct {
	// Color corresponds to the JSON schema field "color".
	Color *string `json:"color,omitempty" yaml:"color,omitempty" mapstructure:"color,omitempty"`

	// Label corresponds to the JSON schema field "label".
	Label *string `json:"label,omitempty" yaml:"label,omitempty" mapstructure:"label,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// Poe corresponds to the JSON schema field "poe".
	Poe *bool `json:"poe,omitempty" yaml:"poe,omitempty" mapstructure:"poe,omitempty"`

	// Positions corresponds to the JSON schema field "positions".
	Positions *int `json:"positions,omitempty" yaml:"positions,omitempty" mapstructure:"positions,omitempty"`

	// Type corresponds to the JSON schema field "type".
	Type ModuleTypeRearPortsElemType `json:"type" yaml:"type" mapstructure:"type"`
}

type ModuleTypeRearPortsElemType string

const ModuleTypeRearPortsElemTypeA110Punch ModuleTypeRearPortsElemType = "110-punch"
const ModuleTypeRearPortsElemTypeA4P2C ModuleTypeRearPortsElemType = "4p2c"
const ModuleTypeRearPortsElemTypeA4P4C ModuleTypeRearPortsElemType = "4p4c"
const ModuleTypeRearPortsElemTypeA6P2C ModuleTypeRearPortsElemType = "6p2c"
const ModuleTypeRearPortsElemTypeA6P4C ModuleTypeRearPortsElemType = "6p4c"
const ModuleTypeRearPortsElemTypeA6P6C ModuleTypeRearPortsElemType = "6p6c"
const ModuleTypeRearPortsElemTypeA8P2C ModuleTypeRearPortsElemType = "8p2c"
const ModuleTypeRearPortsElemTypeA8P4C ModuleTypeRearPortsElemType = "8p4c"
const ModuleTypeRearPortsElemTypeA8P6C ModuleTypeRearPortsElemType = "8p6c"
const ModuleTypeRearPortsElemTypeA8P8C ModuleTypeRearPortsElemType = "8p8c"
const ModuleTypeRearPortsElemTypeBnc ModuleTypeRearPortsElemType = "bnc"
const ModuleTypeRearPortsElemTypeCs ModuleTypeRearPortsElemType = "cs"
const ModuleTypeRearPortsElemTypeF ModuleTypeRearPortsElemType = "f"
const ModuleTypeRearPortsElemTypeFc ModuleTypeRearPortsElemType = "fc"
const ModuleTypeRearPortsElemTypeGg45 ModuleTypeRearPortsElemType = "gg45"
const ModuleTypeRearPortsElemTypeLc ModuleTypeRearPortsElemType = "lc"
const ModuleTypeRearPortsElemTypeLcApc ModuleTypeRearPortsElemType = "lc-apc"
const ModuleTypeRearPortsElemTypeLcPc ModuleTypeRearPortsElemType = "lc-pc"
const ModuleTypeRearPortsElemTypeLcUpc ModuleTypeRearPortsElemType = "lc-upc"
const ModuleTypeRearPortsElemTypeLsh ModuleTypeRearPortsElemType = "lsh"
const ModuleTypeRearPortsElemTypeLshApc ModuleTypeRearPortsElemType = "lsh-apc"
const ModuleTypeRearPortsElemTypeLshPc ModuleTypeRearPortsElemType = "lsh-pc"
const ModuleTypeRearPortsElemTypeLshUpc ModuleTypeRearPortsElemType = "lsh-upc"
const ModuleTypeRearPortsElemTypeLx5 ModuleTypeRearPortsElemType = "lx5"
const ModuleTypeRearPortsElemTypeLx5Apc ModuleTypeRearPortsElemType = "lx5-apc"
const ModuleTypeRearPortsElemTypeLx5Pc ModuleTypeRearPortsElemType = "lx5-pc"
const ModuleTypeRearPortsElemTypeLx5Upc ModuleTypeRearPortsElemType = "lx5-upc"
const ModuleTypeRearPortsElemTypeMpo ModuleTypeRearPortsElemType = "mpo"
const ModuleTypeRearPortsElemTypeMrj21 ModuleTypeRearPortsElemType = "mrj21"
const ModuleTypeRearPortsElemTypeMtrj ModuleTypeRearPortsElemType = "mtrj"
const ModuleTypeRearPortsElemTypeN ModuleTypeRearPortsElemType = "n"
const ModuleTypeRearPortsElemTypeOther ModuleTypeRearPortsElemType = "other"
const ModuleTypeRearPortsElemTypeSc ModuleTypeRearPortsElemType = "sc"
const ModuleTypeRearPortsElemTypeScApc ModuleTypeRearPortsElemType = "sc-apc"
const ModuleTypeRearPortsElemTypeScPc ModuleTypeRearPortsElemType = "sc-pc"
const ModuleTypeRearPortsElemTypeScUpc ModuleTypeRearPortsElemType = "sc-upc"
const ModuleTypeRearPortsElemTypeSma905 ModuleTypeRearPortsElemType = "sma-905"
const ModuleTypeRearPortsElemTypeSma906 ModuleTypeRearPortsElemType = "sma-906"
const ModuleTypeRearPortsElemTypeSn ModuleTypeRearPortsElemType = "sn"
const ModuleTypeRearPortsElemTypeSplice ModuleTypeRearPortsElemType = "splice"
const ModuleTypeRearPortsElemTypeSt ModuleTypeRearPortsElemType = "st"
const ModuleTypeRearPortsElemTypeTera1P ModuleTypeRearPortsElemType = "tera-1p"
const ModuleTypeRearPortsElemTypeTera2P ModuleTypeRearPortsElemType = "tera-2p"
const ModuleTypeRearPortsElemTypeTera4P ModuleTypeRearPortsElemType = "tera-4p"
const ModuleTypeRearPortsElemTypeUrmP2 ModuleTypeRearPortsElemType = "urm-p2"
const ModuleTypeRearPortsElemTypeUrmP4 ModuleTypeRearPortsElemType = "urm-p4"

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypeInterfacesElemType) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ModuleTypeInterfacesElemType {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ModuleTypeInterfacesElemType, v)
	}
	*j = ModuleTypeInterfacesElemType(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypeConsolePortsElemType) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ModuleTypeConsolePortsElemType {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ModuleTypeConsolePortsElemType, v)
	}
	*j = ModuleTypeConsolePortsElemType(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypePowerOutletsElem) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if v, ok := raw["name"]; !ok || v == nil {
		return fmt.Errorf("field name in ModuleTypePowerOutletsElem: required")
	}
	if v, ok := raw["type"]; !ok || v == nil {
		return fmt.Errorf("field type in ModuleTypePowerOutletsElem: required")
	}
	type Plain ModuleTypePowerOutletsElem
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = ModuleTypePowerOutletsElem(plain)
	return nil
}

var enumValues_ModuleTypeConsoleServerPortsElemType = []interface{}{
	"de-9",
	"db-25",
	"rj-11",
	"rj-12",
	"rj-45",
	"mini-din-8",
	"usb-a",
	"usb-b",
	"usb-c",
	"usb-mini-a",
	"usb-mini-b",
	"usb-micro-a",
	"usb-micro-b",
	"usb-micro-ab",
	"other",
}
var enumValues_ModuleTypePowerPortsElemType = []interface{}{
	"iec-60320-c6",
	"iec-60320-c8",
	"iec-60320-c14",
	"iec-60320-c16",
	"iec-60320-c20",
	"iec-60320-c22",
	"iec-60309-p-n-e-4h",
	"iec-60309-p-n-e-6h",
	"iec-60309-p-n-e-9h",
	"iec-60309-2p-e-4h",
	"iec-60309-2p-e-6h",
	"iec-60309-2p-e-9h",
	"iec-60309-3p-e-4h",
	"iec-60309-3p-e-6h",
	"iec-60309-3p-e-9h",
	"iec-60309-3p-n-e-4h",
	"iec-60309-3p-n-e-6h",
	"iec-60309-3p-n-e-9h",
	"iec-60906-1",
	"nbr-14136-10a",
	"nbr-14136-20a",
	"nema-1-15p",
	"nema-5-15p",
	"nema-5-20p",
	"nema-5-30p",
	"nema-5-50p",
	"nema-6-15p",
	"nema-6-20p",
	"nema-6-30p",
	"nema-6-50p",
	"nema-10-30p",
	"nema-10-50p",
	"nema-14-20p",
	"nema-14-30p",
	"nema-14-50p",
	"nema-14-60p",
	"nema-15-15p",
	"nema-15-20p",
	"nema-15-30p",
	"nema-15-50p",
	"nema-15-60p",
	"nema-l1-15p",
	"nema-l5-15p",
	"nema-l5-20p",
	"nema-l5-30p",
	"nema-l5-50p",
	"nema-l6-15p",
	"nema-l6-20p",
	"nema-l6-30p",
	"nema-l6-50p",
	"nema-l10-30p",
	"nema-l14-20p",
	"nema-l14-30p",
	"nema-l14-50p",
	"nema-l14-60p",
	"nema-l15-20p",
	"nema-l15-30p",
	"nema-l15-50p",
	"nema-l15-60p",
	"nema-l21-20p",
	"nema-l21-30p",
	"nema-l22-30p",
	"cs6361c",
	"cs6365c",
	"cs8165c",
	"cs8265c",
	"cs8365c",
	"cs8465c",
	"ita-c",
	"ita-e",
	"ita-f",
	"ita-ef",
	"ita-g",
	"ita-h",
	"ita-i",
	"ita-j",
	"ita-k",
	"ita-l",
	"ita-m",
	"ita-n",
	"ita-o",
	"usb-a",
	"usb-b",
	"usb-c",
	"usb-mini-a",
	"usb-mini-b",
	"usb-micro-a",
	"usb-micro-b",
	"usb-micro-ab",
	"usb-3-b",
	"usb-3-micro-b",
	"dc-terminal",
	"saf-d-grid",
	"neutrik-powercon-20",
	"neutrik-powercon-32",
	"neutrik-powercon-true1",
	"neutrik-powercon-true1-top",
	"ubiquiti-smartpower",
	"hardwired",
	"other",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypeConsoleServerPortsElemType) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ModuleTypeConsoleServerPortsElemType {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ModuleTypeConsoleServerPortsElemType, v)
	}
	*j = ModuleTypeConsoleServerPortsElemType(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypeConsoleServerPortsElem) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if v, ok := raw["name"]; !ok || v == nil {
		return fmt.Errorf("field name in ModuleTypeConsoleServerPortsElem: required")
	}
	if v, ok := raw["type"]; !ok || v == nil {
		return fmt.Errorf("field type in ModuleTypeConsoleServerPortsElem: required")
	}
	type Plain ModuleTypeConsoleServerPortsElem
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = ModuleTypeConsoleServerPortsElem(plain)
	return nil
}

var enumValues_ModuleTypeFrontPortsElemType = []interface{}{
	"8p8c",
	"8p6c",
	"8p4c",
	"8p2c",
	"6p6c",
	"6p4c",
	"6p2c",
	"4p4c",
	"4p2c",
	"gg45",
	"tera-4p",
	"tera-2p",
	"tera-1p",
	"110-punch",
	"bnc",
	"f",
	"n",
	"mrj21",
	"fc",
	"lc",
	"lc-pc",
	"lc-upc",
	"lc-apc",
	"lsh",
	"lsh-pc",
	"lsh-upc",
	"lsh-apc",
	"lx5",
	"lx5-pc",
	"lx5-upc",
	"lx5-apc",
	"mpo",
	"mtrj",
	"sc",
	"sc-pc",
	"sc-upc",
	"sc-apc",
	"st",
	"cs",
	"sn",
	"sma-905",
	"sma-906",
	"urm-p2",
	"urm-p4",
	"urm-p8",
	"splice",
	"other",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypeFrontPortsElemType) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ModuleTypeFrontPortsElemType {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ModuleTypeFrontPortsElemType, v)
	}
	*j = ModuleTypeFrontPortsElemType(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypeFrontPortsElem) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if v, ok := raw["name"]; !ok || v == nil {
		return fmt.Errorf("field name in ModuleTypeFrontPortsElem: required")
	}
	if v, ok := raw["rear_port"]; !ok || v == nil {
		return fmt.Errorf("field rear_port in ModuleTypeFrontPortsElem: required")
	}
	if v, ok := raw["type"]; !ok || v == nil {
		return fmt.Errorf("field type in ModuleTypeFrontPortsElem: required")
	}
	type Plain ModuleTypeFrontPortsElem
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = ModuleTypeFrontPortsElem(plain)
	return nil
}

var enumValues_ModuleTypeInterfacesElemPoeMode = []interface{}{
	"pd",
	"pse",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypeInterfacesElemPoeMode) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ModuleTypeInterfacesElemPoeMode {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ModuleTypeInterfacesElemPoeMode, v)
	}
	*j = ModuleTypeInterfacesElemPoeMode(v)
	return nil
}

var enumValues_ModuleTypeInterfacesElemPoeType = []interface{}{
	"type1-ieee802.3af",
	"type2-ieee802.3at",
	"type3-ieee802.3bt",
	"type4-ieee802.3bt",
	"passive-24v-2pair",
	"passive-24v-4pair",
	"passive-48v-2pair",
	"passive-48v-4pair",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypeInterfacesElemPoeType) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ModuleTypeInterfacesElemPoeType {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ModuleTypeInterfacesElemPoeType, v)
	}
	*j = ModuleTypeInterfacesElemPoeType(v)
	return nil
}

var enumValues_ModuleTypeInterfacesElemType = []interface{}{
	"virtual",
	"bridge",
	"lag",
	"100base-fx",
	"100base-lfx",
	"100base-tx",
	"100base-t1",
	"1000base-t",
	"2.5gbase-t",
	"5gbase-t",
	"10gbase-t",
	"10gbase-cx4",
	"1000base-x-gbic",
	"1000base-x-sfp",
	"10gbase-x-sfpp",
	"10gbase-x-xfp",
	"10gbase-x-xenpak",
	"10gbase-x-x2",
	"25gbase-x-sfp28",
	"50gbase-x-sfp56",
	"40gbase-x-qsfpp",
	"50gbase-x-sfp28",
	"100gbase-x-cfp",
	"100gbase-x-cfp2",
	"200gbase-x-cfp2",
	"400gbase-x-cfp2",
	"100gbase-x-cfp4",
	"100gbase-x-cxp",
	"100gbase-x-cpak",
	"100gbase-x-dsfp",
	"100gbase-x-sfpdd",
	"100gbase-x-qsfp28",
	"100gbase-x-qsfpdd",
	"200gbase-x-qsfp56",
	"200gbase-x-qsfpdd",
	"400gbase-x-qsfp112",
	"400gbase-x-qsfpdd",
	"400gbase-x-osfp",
	"400gbase-x-osfp-rhs",
	"400gbase-x-cdfp",
	"400gbase-x-cfp8",
	"800gbase-x-qsfpdd",
	"800gbase-x-osfp",
	"1000base-kx",
	"10gbase-kr",
	"10gbase-kx4",
	"25gbase-kr",
	"40gbase-kr4",
	"50gbase-kr",
	"100gbase-kp4",
	"100gbase-kr2",
	"100gbase-kr4",
	"ieee802.11a",
	"ieee802.11g",
	"ieee802.11n",
	"ieee802.11ac",
	"ieee802.11ad",
	"ieee802.11ax",
	"ieee802.11ay",
	"ieee802.15.1",
	"other-wireless",
	"gsm",
	"cdma",
	"lte",
	"sonet-oc3",
	"sonet-oc12",
	"sonet-oc48",
	"sonet-oc192",
	"sonet-oc768",
	"sonet-oc1920",
	"sonet-oc3840",
	"1gfc-sfp",
	"2gfc-sfp",
	"4gfc-sfp",
	"8gfc-sfpp",
	"16gfc-sfpp",
	"32gfc-sfp28",
	"64gfc-qsfpp",
	"128gfc-qsfp28",
	"infiniband-sdr",
	"infiniband-ddr",
	"infiniband-qdr",
	"infiniband-fdr10",
	"infiniband-fdr",
	"infiniband-edr",
	"infiniband-hdr",
	"infiniband-ndr",
	"infiniband-xdr",
	"t1",
	"e1",
	"t3",
	"e3",
	"xdsl",
	"docsis",
	"gpon",
	"xg-pon",
	"xgs-pon",
	"ng-pon2",
	"epon",
	"10g-epon",
	"cisco-stackwise",
	"cisco-stackwise-plus",
	"cisco-flexstack",
	"cisco-flexstack-plus",
	"cisco-stackwise-80",
	"cisco-stackwise-160",
	"cisco-stackwise-320",
	"cisco-stackwise-480",
	"cisco-stackwise-1t",
	"juniper-vcp",
	"extreme-summitstack",
	"extreme-summitstack-128",
	"extreme-summitstack-256",
	"extreme-summitstack-512",
	"other",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypePowerOutletsElemType) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ModuleTypePowerOutletsElemType {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ModuleTypePowerOutletsElemType, v)
	}
	*j = ModuleTypePowerOutletsElemType(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypePowerPortsElemType) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ModuleTypePowerPortsElemType {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ModuleTypePowerPortsElemType, v)
	}
	*j = ModuleTypePowerPortsElemType(v)
	return nil
}

var enumValues_ModuleTypePowerOutletsElemType = []interface{}{
	"iec-60320-c5",
	"iec-60320-c7",
	"iec-60320-c13",
	"iec-60320-c15",
	"iec-60320-c19",
	"iec-60320-c21",
	"iec-60309-p-n-e-4h",
	"iec-60309-p-n-e-6h",
	"iec-60309-p-n-e-9h",
	"iec-60309-2p-e-4h",
	"iec-60309-2p-e-6h",
	"iec-60309-2p-e-9h",
	"iec-60309-3p-e-4h",
	"iec-60309-3p-e-6h",
	"iec-60309-3p-e-9h",
	"iec-60309-3p-n-e-4h",
	"iec-60309-3p-n-e-6h",
	"iec-60309-3p-n-e-9h",
	"iec-60906-1",
	"nbr-14136-10a",
	"nbr-14136-20a",
	"nema-1-15r",
	"nema-5-15r",
	"nema-5-20r",
	"nema-5-30r",
	"nema-5-50r",
	"nema-6-15r",
	"nema-6-20r",
	"nema-6-30r",
	"nema-6-50r",
	"nema-10-30r",
	"nema-10-50r",
	"nema-14-20r",
	"nema-14-30r",
	"nema-14-50r",
	"nema-14-60r",
	"nema-15-15r",
	"nema-15-20r",
	"nema-15-30r",
	"nema-15-50r",
	"nema-15-60r",
	"nema-l1-15r",
	"nema-l5-15r",
	"nema-l5-20r",
	"nema-l5-30r",
	"nema-l5-50r",
	"nema-l6-15r",
	"nema-l6-20r",
	"nema-l6-30r",
	"nema-l6-50r",
	"nema-l10-30r",
	"nema-l14-20r",
	"nema-l14-30r",
	"nema-l14-50r",
	"nema-l14-60r",
	"nema-l15-20r",
	"nema-l15-30r",
	"nema-l15-50r",
	"nema-l15-60r",
	"nema-l21-20r",
	"nema-l21-30r",
	"nema-l22-30r",
	"CS6360C",
	"CS6364C",
	"CS8164C",
	"CS8264C",
	"CS8364C",
	"CS8464C",
	"ita-e",
	"ita-f",
	"ita-g",
	"ita-h",
	"ita-i",
	"ita-j",
	"ita-k",
	"ita-l",
	"ita-m",
	"ita-n",
	"ita-o",
	"ita-multistandard",
	"usb-a",
	"usb-micro-b",
	"usb-c",
	"dc-terminal",
	"hdot-cx",
	"saf-d-grid",
	"neutrik-powercon-20a",
	"neutrik-powercon-32a",
	"neutrik-powercon-true1",
	"neutrik-powercon-true1-top",
	"ubiquiti-smartpower",
	"hardwired",
	"other",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypeConsolePortsElem) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if v, ok := raw["name"]; !ok || v == nil {
		return fmt.Errorf("field name in ModuleTypeConsolePortsElem: required")
	}
	if v, ok := raw["type"]; !ok || v == nil {
		return fmt.Errorf("field type in ModuleTypeConsolePortsElem: required")
	}
	type Plain ModuleTypeConsolePortsElem
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = ModuleTypeConsolePortsElem(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypeInterfacesElem) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if v, ok := raw["name"]; !ok || v == nil {
		return fmt.Errorf("field name in ModuleTypeInterfacesElem: required")
	}
	if v, ok := raw["type"]; !ok || v == nil {
		return fmt.Errorf("field type in ModuleTypeInterfacesElem: required")
	}
	type Plain ModuleTypeInterfacesElem
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = ModuleTypeInterfacesElem(plain)
	return nil
}

var enumValues_ModuleTypePowerOutletsElemFeedLeg = []interface{}{
	"A",
	"B",
	"C",
}
var enumValues_ModuleTypeRearPortsElemType = []interface{}{
	"8p8c",
	"8p6c",
	"8p4c",
	"8p2c",
	"6p6c",
	"6p4c",
	"6p2c",
	"4p4c",
	"4p2c",
	"gg45",
	"tera-4p",
	"tera-2p",
	"tera-1p",
	"110-punch",
	"bnc",
	"f",
	"n",
	"mrj21",
	"fc",
	"lc",
	"lc-pc",
	"lc-upc",
	"lc-apc",
	"lsh",
	"lsh-pc",
	"lsh-upc",
	"lsh-apc",
	"lx5",
	"lx5-pc",
	"lx5-upc",
	"lx5-apc",
	"mpo",
	"mtrj",
	"sc",
	"sc-pc",
	"sc-upc",
	"sc-apc",
	"st",
	"cs",
	"sn",
	"sma-905",
	"sma-906",
	"urm-p2",
	"urm-p4",
	"urm-p8",
	"splice",
	"other",
}

const ModuleTypeRearPortsElemTypeUrmP8 ModuleTypeRearPortsElemType = "urm-p8"

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypePowerPortsElem) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if v, ok := raw["name"]; !ok || v == nil {
		return fmt.Errorf("field name in ModuleTypePowerPortsElem: required")
	}
	if v, ok := raw["type"]; !ok || v == nil {
		return fmt.Errorf("field type in ModuleTypePowerPortsElem: required")
	}
	type Plain ModuleTypePowerPortsElem
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = ModuleTypePowerPortsElem(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypeRearPortsElemType) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ModuleTypeRearPortsElemType {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ModuleTypeRearPortsElemType, v)
	}
	*j = ModuleTypeRearPortsElemType(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypePowerOutletsElemFeedLeg) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ModuleTypePowerOutletsElemFeedLeg {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ModuleTypePowerOutletsElemFeedLeg, v)
	}
	*j = ModuleTypePowerOutletsElemFeedLeg(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypeRearPortsElem) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if v, ok := raw["name"]; !ok || v == nil {
		return fmt.Errorf("field name in ModuleTypeRearPortsElem: required")
	}
	if v, ok := raw["type"]; !ok || v == nil {
		return fmt.Errorf("field type in ModuleTypeRearPortsElem: required")
	}
	type Plain ModuleTypeRearPortsElem
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = ModuleTypeRearPortsElem(plain)
	return nil
}

type ModuleTypeWeightUnit string

var enumValues_ModuleTypeWeightUnit = []interface{}{
	"kg",
	"g",
	"lb",
	"oz",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleTypeWeightUnit) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ModuleTypeWeightUnit {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ModuleTypeWeightUnit, v)
	}
	*j = ModuleTypeWeightUnit(v)
	return nil
}

const ModuleTypeWeightUnitKg ModuleTypeWeightUnit = "kg"
const ModuleTypeWeightUnitG ModuleTypeWeightUnit = "g"
const ModuleTypeWeightUnitLb ModuleTypeWeightUnit = "lb"
const ModuleTypeWeightUnitOz ModuleTypeWeightUnit = "oz"

var enumValues_ModuleTypeConsolePortsElemType = []interface{}{
	"de-9",
	"db-25",
	"rj-11",
	"rj-12",
	"rj-45",
	"mini-din-8",
	"usb-a",
	"usb-b",
	"usb-c",
	"usb-mini-a",
	"usb-mini-b",
	"usb-micro-a",
	"usb-micro-b",
	"usb-micro-ab",
	"other",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ModuleType) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if v, ok := raw["manufacturer"]; !ok || v == nil {
		return fmt.Errorf("field manufacturer in ModuleType: required")
	}
	if v, ok := raw["model"]; !ok || v == nil {
		return fmt.Errorf("field model in ModuleType: required")
	}
	type Plain ModuleType
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = ModuleType(plain)
	return nil
}
