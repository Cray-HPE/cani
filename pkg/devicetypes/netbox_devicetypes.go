package devicetypes

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type Type string

const (
	System     Type = "system"
	Cabinet    Type = "cabinet"
	CDU        Type = "cdu"
	CabinetPDU Type = "cabinet-pdu"
	CMM        Type = "cmm"
	CEC        Type = "cec"
	Rack       Type = "rack"
	Chassis    Type = "chassis"
	Blade      Type = "blade"
	Node       Type = "node"
	BMC        Type = "bmc"
	MgmtSwitch Type = "mgmt-switch"
	HSNSwitch  Type = "hsn-switch"
)

var allTypes = []Type{
	System,
	Cabinet,
	CDU,
	CabinetPDU,
	CMM,
	CEC,
	Rack,
	Chassis,
	Blade,
	Node,
	BMC,
	MgmtSwitch,
	HSNSwitch,
}

type DeviceType struct {
	CaniDeviceType
	Manufacturer    string           `yaml:"manufacturer"`
	Model           string           `yaml:"model"`
	Type            Type             `yaml:"hardware-type"`
	Slug            string           `yaml:"slug"`
	PartNumber      *string          `yaml:"part_number"`
	UHeight         *float64         `yaml:"u_height"`
	IsFullDepth     *bool            `yaml:"is_full_depth"`
	Weight          *float64         `yaml:"weight"`
	WeightUnit      *string          `yaml:"weight_unit"`
	FrontImage      bool             `yaml:"front_image"`
	RearImage       bool             `yaml:"rear_image"`
	SubDeviceRole   string           `yaml:"sub_device_role"`
	ConsolePorts    []ConsolePort    `yaml:"console-ports"`
	ModuleBays      []ModuleBay      `yaml:"module-bays"`
	PowerPorts      []PowerPort      `yaml:"power-ports"`
	DeviceBays      []DeviceBay      `yaml:"device-bays"`
	Identifications []Identification `yaml:"identifications"`
	Interfaces      []Interface      `yaml:"interfaces"`
	// PowerOutlets    []PowerOutlets   `yaml:"power-outlets"`
	// ConsoleServerPorts []ConsoleServerPort `yaml:"console-server-ports"`
}

type DeviceBay struct {
	Name    string `yaml:"name"`
	Ordinal int    `yaml:"ordinal"`
}

type Airflow string

const (
	AirflowFrontToRear Airflow = "front-to-rear"
	AirflowRearToFront Airflow = "rear-to-front"
	AirflowLeftToRight Airflow = "left-to-right"
	AirflowRightToLeft Airflow = "right-to-left"
	AirflowSideToRear  Airflow = "side-to-rear"
	AirflowPassive     Airflow = "passive"
)

type WeightUnit string

const (
	WeightUnitKiloGram WeightUnit = "kg"
	WeightUnitGram     WeightUnit = "g"
	WeightUnitPound    WeightUnit = "lb"
	WeightUnitOunce    WeightUnit = "oz"
)

type SubDeviceRole string

const (
	SubDeviceRoleParent SubDeviceRole = "parent"
	SubDeviceRoleChild  SubDeviceRole = "child"
)

type ModuleBay struct {
	// Label corresponds to the JSON schema field "label".
	Label *string `yaml:"label,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// Position corresponds to the JSON schema field "position".
	Position *string `json:"position,omitempty" yaml:"position,omitempty" mapstructure:"position,omitempty"`
}

type PowerPort struct {
	// AllocatedDraw corresponds to the JSON schema field "allocated_draw".
	AllocatedDraw *int `json:"allocated_draw,omitempty" yaml:"allocated_draw,omitempty" mapstructure:"allocated_draw,omitempty"`

	// Label corresponds to the JSON schema field "label".
	Label *string `json:"label,omitempty" yaml:"label,omitempty" mapstructure:"label,omitempty"`

	// MaximumDraw corresponds to the JSON schema field "maximum_draw".
	MaximumDraw *int `json:"maximum_draw,omitempty" yaml:"maximum_draw,omitempty" mapstructure:"maximum_draw,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// Type corresponds to the JSON schema field "type".
	Type PowerPortType `json:"type" yaml:"type" mapstructure:"type"`
}

type PowerPortType string

const PowerPortTypeCs6361C PowerPortType = "cs6361c"
const PowerPortTypeCs6365C PowerPortType = "cs6365c"
const PowerPortTypeCs8165C PowerPortType = "cs8165c"
const PowerPortTypeCs8265C PowerPortType = "cs8265c"
const PowerPortTypeCs8365C PowerPortType = "cs8365c"
const PowerPortTypeCs8465C PowerPortType = "cs8465c"
const PowerPortTypeDcTerminal PowerPortType = "dc-terminal"
const PowerPortTypeHardwired PowerPortType = "hardwired"
const PowerPortTypeIec603092PE4H PowerPortType = "iec-60309-2p-e-4h"
const PowerPortTypeIec603092PE6H PowerPortType = "iec-60309-2p-e-6h"
const PowerPortTypeIec603092PE9H PowerPortType = "iec-60309-2p-e-9h"
const PowerPortTypeIec603093PE4H PowerPortType = "iec-60309-3p-e-4h"
const PowerPortTypeIec603093PE6H PowerPortType = "iec-60309-3p-e-6h"
const PowerPortTypeIec603093PE9H PowerPortType = "iec-60309-3p-e-9h"
const PowerPortTypeIec603093PNE4H PowerPortType = "iec-60309-3p-n-e-4h"
const PowerPortTypeIec603093PNE6H PowerPortType = "iec-60309-3p-n-e-6h"
const PowerPortTypeIec603093PNE9H PowerPortType = "iec-60309-3p-n-e-9h"
const PowerPortTypeIec60309PNE4H PowerPortType = "iec-60309-p-n-e-4h"
const PowerPortTypeIec60309PNE6H PowerPortType = "iec-60309-p-n-e-6h"
const PowerPortTypeIec60309PNE9H PowerPortType = "iec-60309-p-n-e-9h"
const PowerPortTypeIec60320C14 PowerPortType = "iec-60320-c14"
const PowerPortTypeIec60320C16 PowerPortType = "iec-60320-c16"
const PowerPortTypeIec60320C20 PowerPortType = "iec-60320-c20"
const PowerPortTypeIec60320C22 PowerPortType = "iec-60320-c22"
const PowerPortTypeIec60320C6 PowerPortType = "iec-60320-c6"
const PowerPortTypeIec60320C8 PowerPortType = "iec-60320-c8"
const PowerPortTypeIec609061 PowerPortType = "iec-60906-1"
const PowerPortTypeItaC PowerPortType = "ita-c"
const PowerPortTypeItaE PowerPortType = "ita-e"
const PowerPortTypeItaEf PowerPortType = "ita-ef"
const PowerPortTypeItaF PowerPortType = "ita-f"
const PowerPortTypeItaG PowerPortType = "ita-g"
const PowerPortTypeItaH PowerPortType = "ita-h"
const PowerPortTypeItaI PowerPortType = "ita-i"
const PowerPortTypeItaJ PowerPortType = "ita-j"
const PowerPortTypeItaK PowerPortType = "ita-k"
const PowerPortTypeItaL PowerPortType = "ita-l"
const PowerPortTypeItaM PowerPortType = "ita-m"
const PowerPortTypeItaN PowerPortType = "ita-n"
const PowerPortTypeItaO PowerPortType = "ita-o"
const PowerPortTypeNbr1413610A PowerPortType = "nbr-14136-10a"
const PowerPortTypeNbr1413620A PowerPortType = "nbr-14136-20a"
const PowerPortTypeNema1030P PowerPortType = "nema-10-30p"
const PowerPortTypeNema1050P PowerPortType = "nema-10-50p"
const PowerPortTypeNema115P PowerPortType = "nema-1-15p"
const PowerPortTypeNema1420P PowerPortType = "nema-14-20p"
const PowerPortTypeNema1430P PowerPortType = "nema-14-30p"
const PowerPortTypeNema1450P PowerPortType = "nema-14-50p"
const PowerPortTypeNema1460P PowerPortType = "nema-14-60p"
const PowerPortTypeNema1515P PowerPortType = "nema-15-15p"
const PowerPortTypeNema1520P PowerPortType = "nema-15-20p"
const PowerPortTypeNema1530P PowerPortType = "nema-15-30p"
const PowerPortTypeNema1550P PowerPortType = "nema-15-50p"
const PowerPortTypeNema1560P PowerPortType = "nema-15-60p"
const PowerPortTypeNema515P PowerPortType = "nema-5-15p"
const PowerPortTypeNema520P PowerPortType = "nema-5-20p"
const PowerPortTypeNema530P PowerPortType = "nema-5-30p"
const PowerPortTypeNema550P PowerPortType = "nema-5-50p"
const PowerPortTypeNema615P PowerPortType = "nema-6-15p"
const PowerPortTypeNema620P PowerPortType = "nema-6-20p"
const PowerPortTypeNema630P PowerPortType = "nema-6-30p"
const PowerPortTypeNema650P PowerPortType = "nema-6-50p"
const PowerPortTypeNemaL1030P PowerPortType = "nema-l10-30p"
const PowerPortTypeNemaL115P PowerPortType = "nema-l1-15p"
const PowerPortTypeNemaL1420P PowerPortType = "nema-l14-20p"
const PowerPortTypeNemaL1430P PowerPortType = "nema-l14-30p"
const PowerPortTypeNemaL1450P PowerPortType = "nema-l14-50p"
const PowerPortTypeNemaL1460P PowerPortType = "nema-l14-60p"
const PowerPortTypeNemaL1520P PowerPortType = "nema-l15-20p"
const PowerPortTypeNemaL1530P PowerPortType = "nema-l15-30p"
const PowerPortTypeNemaL1550P PowerPortType = "nema-l15-50p"
const PowerPortTypeNemaL1560P PowerPortType = "nema-l15-60p"
const PowerPortTypeNemaL2120P PowerPortType = "nema-l21-20p"
const PowerPortTypeNemaL2130P PowerPortType = "nema-l21-30p"
const PowerPortTypeNemaL2230P PowerPortType = "nema-l22-30p"
const PowerPortTypeNemaL515P PowerPortType = "nema-l5-15p"
const PowerPortTypeNemaL520P PowerPortType = "nema-l5-20p"
const PowerPortTypeNemaL530P PowerPortType = "nema-l5-30p"
const PowerPortTypeNemaL550P PowerPortType = "nema-l5-50p"
const PowerPortTypeNemaL615P PowerPortType = "nema-l6-15p"
const PowerPortTypeNemaL620P PowerPortType = "nema-l6-20p"
const PowerPortTypeNemaL630P PowerPortType = "nema-l6-30p"
const PowerPortTypeNemaL650P PowerPortType = "nema-l6-50p"
const PowerPortTypeNeutrikPowercon20 PowerPortType = "neutrik-powercon-20"
const PowerPortTypeNeutrikPowercon32 PowerPortType = "neutrik-powercon-32"
const PowerPortTypeNeutrikPowerconTrue1 PowerPortType = "neutrik-powercon-true1"
const PowerPortTypeNeutrikPowerconTrue1Top PowerPortType = "neutrik-powercon-true1-top"
const PowerPortTypeOther PowerPortType = "other"
const PowerPortTypeSafDGrid PowerPortType = "saf-d-grid"
const PowerPortTypeUbiquitiSmartpower PowerPortType = "ubiquiti-smartpower"
const PowerPortTypeUsb3B PowerPortType = "usb-3-b"
const PowerPortTypeUsb3MicroB PowerPortType = "usb-3-micro-b"
const PowerPortTypeUsbA PowerPortType = "usb-a"
const PowerPortTypeUsbB PowerPortType = "usb-b"
const PowerPortTypeUsbC PowerPortType = "usb-c"
const PowerPortTypeUsbMicroA PowerPortType = "usb-micro-a"
const PowerPortTypeUsbMicroAb PowerPortType = "usb-micro-ab"
const PowerPortTypeUsbMicroB PowerPortType = "usb-micro-b"
const PowerPortTypeUsbMiniA PowerPortType = "usb-mini-a"
const PowerPortTypeUsbMiniB PowerPortType = "usb-mini-b"

type Identification struct {
	Manufacturer string  `yaml:"manufacturer"`
	Model        string  `yaml:"model"`
	PartNumber   *string `yaml:"part-number"`
}

type ConsolePort struct {
	// Label corresponds to the JSON schema field "label".
	Label *string `json:"label,omitempty" yaml:"label,omitempty" mapstructure:"label,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// Poe corresponds to the JSON schema field "poe".
	Poe *bool `json:"poe,omitempty" yaml:"poe,omitempty" mapstructure:"poe,omitempty"`

	// Type corresponds to the JSON schema field "type".
	Type ConsolePortType `json:"type" yaml:"type" mapstructure:"type"`
}

type ConsolePortType string

const ConsolePortTypeDb25 ConsolePortType = "db-25"
const ConsolePortTypeDe9 ConsolePortType = "de-9"
const ConsolePortTypeMiniDin8 ConsolePortType = "mini-din-8"
const ConsolePortTypeOther ConsolePortType = "other"
const ConsolePortTypeRj11 ConsolePortType = "rj-11"
const ConsolePortTypeRj12 ConsolePortType = "rj-12"
const ConsolePortTypeRj45 ConsolePortType = "rj-45"
const ConsolePortTypeUsbA ConsolePortType = "usb-a"
const ConsolePortTypeUsbB ConsolePortType = "usb-b"
const ConsolePortTypeUsbC ConsolePortType = "usb-c"
const ConsolePortTypeUsbMicroA ConsolePortType = "usb-micro-a"
const ConsolePortTypeUsbMicroAb ConsolePortType = "usb-micro-ab"
const ConsolePortTypeUsbMicroB ConsolePortType = "usb-micro-b"
const ConsolePortTypeUsbMiniA ConsolePortType = "usb-mini-a"
const ConsolePortTypeUsbMiniB ConsolePortType = "usb-mini-b"

type Interface struct {
	// Label corresponds to the JSON schema field "label".
	Label *string `json:"label,omitempty" yaml:"label,omitempty" mapstructure:"label,omitempty"`

	// MgmtOnly corresponds to the JSON schema field "mgmt_only".
	MgmtOnly *bool `json:"mgmt_only,omitempty" yaml:"mgmt_only,omitempty" mapstructure:"mgmt_only,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// PoeMode corresponds to the JSON schema field "poe_mode".
	PoeMode *InterfacePoeElem `json:"poe_mode,omitempty" yaml:"poe_mode,omitempty" mapstructure:"poe_mode,omitempty"`

	// PoeType corresponds to the JSON schema field "poe_type".
	PoeType *InterfacesElemPoeType `json:"poe_type,omitempty" yaml:"poe_type,omitempty" mapstructure:"poe_type,omitempty"`

	// Type corresponds to the JSON schema field "type".
	Type InterfacesElemType `json:"type" yaml:"type" mapstructure:"type"`
}

type InterfacePoeElem string

const InterfacePoeElemPd InterfacePoeElem = "pd"
const InterfacePoeElemPse InterfacePoeElem = "pse"

type InterfacesElemPoeType string

const InterfacesElemPoeTypePassive24V2Pair InterfacesElemPoeType = "passive-24v-2pair"
const InterfacesElemPoeTypePassive24V4Pair InterfacesElemPoeType = "passive-24v-4pair"
const InterfacesElemPoeTypePassive48V2Pair InterfacesElemPoeType = "passive-48v-2pair"
const InterfacesElemPoeTypePassive48V4Pair InterfacesElemPoeType = "passive-48v-4pair"
const InterfacesElemPoeTypeType1Ieee8023Af InterfacesElemPoeType = "type1-ieee802.3af"
const InterfacesElemPoeTypeType2Ieee8023At InterfacesElemPoeType = "type2-ieee802.3at"
const InterfacesElemPoeTypeType3Ieee8023Bt InterfacesElemPoeType = "type3-ieee802.3bt"
const InterfacesElemPoeTypeType4Ieee8023Bt InterfacesElemPoeType = "type4-ieee802.3bt"

type InterfacesElemType string

const InterfacesElemTypeA1000BaseKx InterfacesElemType = "1000base-kx"
const InterfacesElemTypeA1000BaseT InterfacesElemType = "1000base-t"
const InterfacesElemTypeA1000BaseXGbic InterfacesElemType = "1000base-x-gbic"
const InterfacesElemTypeA1000BaseXSfp InterfacesElemType = "1000base-x-sfp"
const InterfacesElemTypeA100BaseFx InterfacesElemType = "100base-fx"
const InterfacesElemTypeA100BaseLfx InterfacesElemType = "100base-lfx"
const InterfacesElemTypeA100BaseT1 InterfacesElemType = "100base-t1"
const InterfacesElemTypeA100BaseTx InterfacesElemType = "100base-tx"
const InterfacesElemTypeA100GbaseKp4 InterfacesElemType = "100gbase-kp4"
const InterfacesElemTypeA100GbaseKr2 InterfacesElemType = "100gbase-kr2"
const InterfacesElemTypeA100GbaseKr4 InterfacesElemType = "100gbase-kr4"
const InterfacesElemTypeA100GbaseXCfp InterfacesElemType = "100gbase-x-cfp"
const InterfacesElemTypeA100GbaseXCfp2 InterfacesElemType = "100gbase-x-cfp2"
const InterfacesElemTypeA100GbaseXCfp4 InterfacesElemType = "100gbase-x-cfp4"
const InterfacesElemTypeA100GbaseXCpak InterfacesElemType = "100gbase-x-cpak"
const InterfacesElemTypeA100GbaseXCxp InterfacesElemType = "100gbase-x-cxp"
const InterfacesElemTypeA100GbaseXDsfp InterfacesElemType = "100gbase-x-dsfp"
const InterfacesElemTypeA100GbaseXQsfp28 InterfacesElemType = "100gbase-x-qsfp28"
const InterfacesElemTypeA100GbaseXQsfpdd InterfacesElemType = "100gbase-x-qsfpdd"
const InterfacesElemTypeA100GbaseXSfpdd InterfacesElemType = "100gbase-x-sfpdd"
const InterfacesElemTypeA10GEpon InterfacesElemType = "10g-epon"
const InterfacesElemTypeA10GbaseCx4 InterfacesElemType = "10gbase-cx4"
const InterfacesElemTypeA10GbaseKr InterfacesElemType = "10gbase-kr"
const InterfacesElemTypeA10GbaseKx4 InterfacesElemType = "10gbase-kx4"
const InterfacesElemTypeA10GbaseT InterfacesElemType = "10gbase-t"
const InterfacesElemTypeA10GbaseXSfpp InterfacesElemType = "10gbase-x-sfpp"
const InterfacesElemTypeA10GbaseXX2 InterfacesElemType = "10gbase-x-x2"
const InterfacesElemTypeA10GbaseXXenpak InterfacesElemType = "10gbase-x-xenpak"
const InterfacesElemTypeA10GbaseXXfp InterfacesElemType = "10gbase-x-xfp"
const InterfacesElemTypeA128GfcQsfp28 InterfacesElemType = "128gfc-qsfp28"
const InterfacesElemTypeA16GfcSfpp InterfacesElemType = "16gfc-sfpp"
const InterfacesElemTypeA1GfcSfp InterfacesElemType = "1gfc-sfp"
const InterfacesElemTypeA200GbaseXCfp2 InterfacesElemType = "200gbase-x-cfp2"
const InterfacesElemTypeA200GbaseXQsfp56 InterfacesElemType = "200gbase-x-qsfp56"
const InterfacesElemTypeA200GbaseXQsfpdd InterfacesElemType = "200gbase-x-qsfpdd"
const InterfacesElemTypeA25GbaseKr InterfacesElemType = "25gbase-kr"
const InterfacesElemTypeA25GbaseT InterfacesElemType = "2.5gbase-t"
const InterfacesElemTypeA25GbaseXSfp28 InterfacesElemType = "25gbase-x-sfp28"
const InterfacesElemTypeA2GfcSfp InterfacesElemType = "2gfc-sfp"
const InterfacesElemTypeA32GfcSfp28 InterfacesElemType = "32gfc-sfp28"
const InterfacesElemTypeA400GbaseXCdfp InterfacesElemType = "400gbase-x-cdfp"
const InterfacesElemTypeA400GbaseXCfp2 InterfacesElemType = "400gbase-x-cfp2"
const InterfacesElemTypeA400GbaseXCfp8 InterfacesElemType = "400gbase-x-cfp8"
const InterfacesElemTypeA400GbaseXOsfp InterfacesElemType = "400gbase-x-osfp"
const InterfacesElemTypeA400GbaseXOsfpRhs InterfacesElemType = "400gbase-x-osfp-rhs"
const InterfacesElemTypeA400GbaseXQsfp112 InterfacesElemType = "400gbase-x-qsfp112"
const InterfacesElemTypeA400GbaseXQsfpdd InterfacesElemType = "400gbase-x-qsfpdd"
const InterfacesElemTypeA40GbaseKr4 InterfacesElemType = "40gbase-kr4"
const InterfacesElemTypeA40GbaseXQsfpp InterfacesElemType = "40gbase-x-qsfpp"
const InterfacesElemTypeA4GfcSfp InterfacesElemType = "4gfc-sfp"
const InterfacesElemTypeA50GbaseKr InterfacesElemType = "50gbase-kr"
const InterfacesElemTypeA50GbaseXSfp28 InterfacesElemType = "50gbase-x-sfp28"
const InterfacesElemTypeA50GbaseXSfp56 InterfacesElemType = "50gbase-x-sfp56"
const InterfacesElemTypeA5GbaseT InterfacesElemType = "5gbase-t"
const InterfacesElemTypeA64GfcQsfpp InterfacesElemType = "64gfc-qsfpp"
const InterfacesElemTypeA800GbaseXOsfp InterfacesElemType = "800gbase-x-osfp"
const InterfacesElemTypeA800GbaseXQsfpdd InterfacesElemType = "800gbase-x-qsfpdd"
const InterfacesElemTypeA8GfcSfpp InterfacesElemType = "8gfc-sfpp"
const InterfacesElemTypeBridge InterfacesElemType = "bridge"
const InterfacesElemTypeCdma InterfacesElemType = "cdma"
const InterfacesElemTypeCiscoFlexstack InterfacesElemType = "cisco-flexstack"
const InterfacesElemTypeCiscoFlexstackPlus InterfacesElemType = "cisco-flexstack-plus"
const InterfacesElemTypeCiscoStackwise InterfacesElemType = "cisco-stackwise"
const InterfacesElemTypeCiscoStackwise160 InterfacesElemType = "cisco-stackwise-160"
const InterfacesElemTypeCiscoStackwise1T InterfacesElemType = "cisco-stackwise-1t"
const InterfacesElemTypeCiscoStackwise320 InterfacesElemType = "cisco-stackwise-320"
const InterfacesElemTypeCiscoStackwise480 InterfacesElemType = "cisco-stackwise-480"
const InterfacesElemTypeCiscoStackwise80 InterfacesElemType = "cisco-stackwise-80"
const InterfacesElemTypeCiscoStackwisePlus InterfacesElemType = "cisco-stackwise-plus"
const InterfacesElemTypeDocsis InterfacesElemType = "docsis"
const InterfacesElemTypeE1 InterfacesElemType = "e1"
const InterfacesElemTypeE3 InterfacesElemType = "e3"
const InterfacesElemTypeEpon InterfacesElemType = "epon"
const InterfacesElemTypeExtremeSummitstack InterfacesElemType = "extreme-summitstack"
const InterfacesElemTypeExtremeSummitstack128 InterfacesElemType = "extreme-summitstack-128"
const InterfacesElemTypeExtremeSummitstack256 InterfacesElemType = "extreme-summitstack-256"
const InterfacesElemTypeExtremeSummitstack512 InterfacesElemType = "extreme-summitstack-512"
const InterfacesElemTypeGpon InterfacesElemType = "gpon"
const InterfacesElemTypeGsm InterfacesElemType = "gsm"
const InterfacesElemTypeIeee80211A InterfacesElemType = "ieee802.11a"
const InterfacesElemTypeIeee80211Ac InterfacesElemType = "ieee802.11ac"
const InterfacesElemTypeIeee80211Ad InterfacesElemType = "ieee802.11ad"
const InterfacesElemTypeIeee80211Ax InterfacesElemType = "ieee802.11ax"
const InterfacesElemTypeIeee80211Ay InterfacesElemType = "ieee802.11ay"
const InterfacesElemTypeIeee80211G InterfacesElemType = "ieee802.11g"
const InterfacesElemTypeIeee80211N InterfacesElemType = "ieee802.11n"
const InterfacesElemTypeIeee802151 InterfacesElemType = "ieee802.15.1"
const InterfacesElemTypeInfinibandDdr InterfacesElemType = "infiniband-ddr"
const InterfacesElemTypeInfinibandEdr InterfacesElemType = "infiniband-edr"
const InterfacesElemTypeInfinibandFdr InterfacesElemType = "infiniband-fdr"
const InterfacesElemTypeInfinibandFdr10 InterfacesElemType = "infiniband-fdr10"
const InterfacesElemTypeInfinibandHdr InterfacesElemType = "infiniband-hdr"
const InterfacesElemTypeInfinibandNdr InterfacesElemType = "infiniband-ndr"
const InterfacesElemTypeInfinibandQdr InterfacesElemType = "infiniband-qdr"
const InterfacesElemTypeInfinibandSdr InterfacesElemType = "infiniband-sdr"
const InterfacesElemTypeInfinibandXdr InterfacesElemType = "infiniband-xdr"
const InterfacesElemTypeJuniperVcp InterfacesElemType = "juniper-vcp"
const InterfacesElemTypeLag InterfacesElemType = "lag"
const InterfacesElemTypeLte InterfacesElemType = "lte"
const InterfacesElemTypeNgPon2 InterfacesElemType = "ng-pon2"
const InterfacesElemTypeOther InterfacesElemType = "other"
const InterfacesElemTypeOtherWireless InterfacesElemType = "other-wireless"
const InterfacesElemTypeSonetOc12 InterfacesElemType = "sonet-oc12"
const InterfacesElemTypeSonetOc192 InterfacesElemType = "sonet-oc192"
const InterfacesElemTypeSonetOc1920 InterfacesElemType = "sonet-oc1920"
const InterfacesElemTypeSonetOc3 InterfacesElemType = "sonet-oc3"
const InterfacesElemTypeSonetOc3840 InterfacesElemType = "sonet-oc3840"
const InterfacesElemTypeSonetOc48 InterfacesElemType = "sonet-oc48"
const InterfacesElemTypeSonetOc768 InterfacesElemType = "sonet-oc768"
const InterfacesElemTypeT1 InterfacesElemType = "t1"
const InterfacesElemTypeT3 InterfacesElemType = "t3"
const InterfacesElemTypeVirtual InterfacesElemType = "virtual"
const InterfacesElemTypeXdsl InterfacesElemType = "xdsl"
const InterfacesElemTypeXgPon InterfacesElemType = "xg-pon"
const InterfacesElemTypeXgsPon InterfacesElemType = "xgs-pon"

// Cable Adds support for custom fields and tags.
type Cable struct {
	Id                   int32                  `json:"id"`
	Url                  string                 `json:"url"`
	DisplayUrl           *string                `json:"display_url,omitempty"`
	Display              string                 `json:"display"`
	Type                 CableType              `json:"type,omitempty"`
	ATerminations        []GenericObject        `json:"a_terminations,omitempty"`
	BTerminations        []GenericObject        `json:"b_terminations,omitempty"`
	Status               *CableStatus           `json:"status,omitempty"`
	Tenant               BriefTenant            `json:"tenant,omitempty"`
	Label                *string                `json:"label,omitempty"`
	Color                *string                `json:"color,omitempty" validate:"regexp=^[0-9a-f]{6}$"`
	Length               *float64               `json:"length,omitempty"`
	LengthUnit           CableLengthUnit        `json:"length_unit,omitempty"`
	Description          *string                `json:"description,omitempty"`
	Comments             *string                `json:"comments,omitempty"`
	Tags                 []Tag                  `json:"tags,omitempty"`
	CustomFields         map[string]interface{} `json:"custom_fields,omitempty"`
	Created              *time.Time             `json:"created,omitempty"`
	LastUpdated          *time.Time             `json:"last_updated,omitempty"`
	AdditionalProperties map[string]interface{}
}

type CableType string

// List of Cable_type
const (
	CABLETYPE_CAT3        CableType = "cat3"
	CABLETYPE_CAT5        CableType = "cat5"
	CABLETYPE_CAT5E       CableType = "cat5e"
	CABLETYPE_CAT6        CableType = "cat6"
	CABLETYPE_CAT6A       CableType = "cat6a"
	CABLETYPE_CAT7        CableType = "cat7"
	CABLETYPE_CAT7A       CableType = "cat7a"
	CABLETYPE_CAT8        CableType = "cat8"
	CABLETYPE_DAC_ACTIVE  CableType = "dac-active"
	CABLETYPE_DAC_PASSIVE CableType = "dac-passive"
	CABLETYPE_MRJ21_TRUNK CableType = "mrj21-trunk"
	CABLETYPE_COAXIAL     CableType = "coaxial"
	CABLETYPE_MMF         CableType = "mmf"
	CABLETYPE_MMF_OM1     CableType = "mmf-om1"
	CABLETYPE_MMF_OM2     CableType = "mmf-om2"
	CABLETYPE_MMF_OM3     CableType = "mmf-om3"
	CABLETYPE_MMF_OM4     CableType = "mmf-om4"
	CABLETYPE_MMF_OM5     CableType = "mmf-om5"
	CABLETYPE_SMF         CableType = "smf"
	CABLETYPE_SMF_OS1     CableType = "smf-os1"
	CABLETYPE_SMF_OS2     CableType = "smf-os2"
	CABLETYPE_AOC         CableType = "aoc"
	CABLETYPE_USB         CableType = "usb"
	CABLETYPE_POWER       CableType = "power"
	CABLETYPE_EMPTY       CableType = ""
)

// GenericObject Minimal representation of some generic object identified by ContentType and PK.
type GenericObject struct {
	ObjectType           string      `json:"object_type"`
	ObjectId             int32       `json:"object_id"`
	Object               interface{} `json:"object,omitempty"`
	AdditionalProperties map[string]interface{}
}

// CableStatus struct for CableStatus
type CableStatus struct {
	Value                *CableStatusValue `json:"value,omitempty"`
	Label                *CableStatusLabel `json:"label,omitempty"`
	AdditionalProperties map[string]interface{}
}

type CableStatusValue string

// List of Cable_status_value
const (
	CABLESTATUSVALUE_CONNECTED       CableStatusValue = "connected"
	CABLESTATUSVALUE_PLANNED         CableStatusValue = "planned"
	CABLESTATUSVALUE_DECOMMISSIONING CableStatusValue = "decommissioning"
)

// CableStatusLabel the model 'CableStatusLabel'
type CableStatusLabel string

// List of Cable_status_label
const (
	CABLESTATUSLABEL_CONNECTED       CableStatusLabel = "Connected"
	CABLESTATUSLABEL_PLANNED         CableStatusLabel = "Planned"
	CABLESTATUSLABEL_DECOMMISSIONING CableStatusLabel = "Decommissioning"
)

type CableLengthUnit struct {
	Value                *CableLengthUnitValue `json:"value,omitempty"`
	Label                *CableLengthUnitLabel `json:"label,omitempty"`
	AdditionalProperties map[string]interface{}
}

// CableLengthUnitValue * `km` - Kilometers * `m` - Meters * `cm` - Centimeters * `mi` - Miles * `ft` - Feet * `in` - Inches
type CableLengthUnitValue string

// List of Cable_length_unit_value
const (
	CABLELENGTHUNITVALUE_KM    CableLengthUnitValue = "km"
	CABLELENGTHUNITVALUE_M     CableLengthUnitValue = "m"
	CABLELENGTHUNITVALUE_CM    CableLengthUnitValue = "cm"
	CABLELENGTHUNITVALUE_MI    CableLengthUnitValue = "mi"
	CABLELENGTHUNITVALUE_FT    CableLengthUnitValue = "ft"
	CABLELENGTHUNITVALUE_IN    CableLengthUnitValue = "in"
	CABLELENGTHUNITVALUE_EMPTY CableLengthUnitValue = ""
)

// CableLengthUnitLabel the model 'CableLengthUnitLabel'
type CableLengthUnitLabel string

// List of Cable_length_unit_label
const (
	CABLELENGTHUNITLABEL_KILOMETERS  CableLengthUnitLabel = "Kilometers"
	CABLELENGTHUNITLABEL_METERS      CableLengthUnitLabel = "Meters"
	CABLELENGTHUNITLABEL_CENTIMETERS CableLengthUnitLabel = "Centimeters"
	CABLELENGTHUNITLABEL_MILES       CableLengthUnitLabel = "Miles"
	CABLELENGTHUNITLABEL_FEET        CableLengthUnitLabel = "Feet"
	CABLELENGTHUNITLABEL_INCHES      CableLengthUnitLabel = "Inches"
)

// Tag Represents an object related through a ForeignKey field
type Tag struct {
	Id                   int32   `json:"id"`
	Url                  string  `json:"url"`
	DisplayUrl           *string `json:"display_url,omitempty"`
	Display              string  `json:"display"`
	Name                 string  `json:"name"`
	Slug                 string  `json:"slug" validate:"regexp=^[-\\\\w]+$"`
	Color                *string `json:"color,omitempty" validate:"regexp=^[0-9a-f]{6}$"`
	AdditionalProperties map[string]interface{}
}

// BriefTenant Adds support for custom fields and tags.
type BriefTenant struct {
	Id                   int32   `json:"id"`
	Url                  string  `json:"url"`
	Display              string  `json:"display"`
	Name                 string  `json:"name"`
	Slug                 string  `json:"slug" validate:"regexp=^[-a-zA-Z0-9_]+$"`
	Description          *string `json:"description,omitempty"`
	AdditionalProperties map[string]interface{}
}

func unmarshalMultiple(in []byte, out *[]DeviceType) error {
	r := bytes.NewReader(in)
	decoder := yaml.NewDecoder(r)

	var documentNumber int
	for {
		var deviceType DeviceType
		if err := decoder.Decode(&deviceType); err != nil {
			// Break out of loop when more yaml documents to process
			if err != io.EOF {
				return fmt.Errorf("failed to parse document %d, error %w", documentNumber, err)
			}

			break
		}

		*out = append(*out, deviceType)
		documentNumber++
	}

	return nil
}

func (d *DeviceType) ToCaniDeviceType() *CaniDeviceType {
	return &CaniDeviceType{
		ID:             uuid.New(),
		Name:           d.generateCaniName(),
		Type:           d.Type,
		DeviceTypeSlug: d.Slug,
		Vendor:         d.Manufacturer,
		Model:          d.Model,
		Architecture:   d.Architecture,
	}
}

func (d *DeviceType) generateCaniName() string {
	// generate some random name of words
	deviceName := fmt.Sprintf("cani-device-%d", rand.Intn(1000000))

	return deviceName
}
