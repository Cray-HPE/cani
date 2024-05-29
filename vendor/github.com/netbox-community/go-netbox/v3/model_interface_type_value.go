/*
NetBox REST API

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: 3.7.1 (3.7)
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package netbox

import (
	"encoding/json"
	"fmt"
)

// InterfaceTypeValue * `virtual` - Virtual * `bridge` - Bridge * `lag` - Link Aggregation Group (LAG) * `100base-fx` - 100BASE-FX (10/100ME FIBER) * `100base-lfx` - 100BASE-LFX (10/100ME FIBER) * `100base-tx` - 100BASE-TX (10/100ME) * `100base-t1` - 100BASE-T1 (10/100ME Single Pair) * `1000base-t` - 1000BASE-T (1GE) * `2.5gbase-t` - 2.5GBASE-T (2.5GE) * `5gbase-t` - 5GBASE-T (5GE) * `10gbase-t` - 10GBASE-T (10GE) * `10gbase-cx4` - 10GBASE-CX4 (10GE) * `1000base-x-gbic` - GBIC (1GE) * `1000base-x-sfp` - SFP (1GE) * `10gbase-x-sfpp` - SFP+ (10GE) * `10gbase-x-xfp` - XFP (10GE) * `10gbase-x-xenpak` - XENPAK (10GE) * `10gbase-x-x2` - X2 (10GE) * `25gbase-x-sfp28` - SFP28 (25GE) * `50gbase-x-sfp56` - SFP56 (50GE) * `40gbase-x-qsfpp` - QSFP+ (40GE) * `50gbase-x-sfp28` - QSFP28 (50GE) * `100gbase-x-cfp` - CFP (100GE) * `100gbase-x-cfp2` - CFP2 (100GE) * `200gbase-x-cfp2` - CFP2 (200GE) * `400gbase-x-cfp2` - CFP2 (400GE) * `100gbase-x-cfp4` - CFP4 (100GE) * `100gbase-x-cxp` - CXP (100GE) * `100gbase-x-cpak` - Cisco CPAK (100GE) * `100gbase-x-dsfp` - DSFP (100GE) * `100gbase-x-sfpdd` - SFP-DD (100GE) * `100gbase-x-qsfp28` - QSFP28 (100GE) * `100gbase-x-qsfpdd` - QSFP-DD (100GE) * `200gbase-x-qsfp56` - QSFP56 (200GE) * `200gbase-x-qsfpdd` - QSFP-DD (200GE) * `400gbase-x-qsfp112` - QSFP112 (400GE) * `400gbase-x-qsfpdd` - QSFP-DD (400GE) * `400gbase-x-osfp` - OSFP (400GE) * `400gbase-x-osfp-rhs` - OSFP-RHS (400GE) * `400gbase-x-cdfp` - CDFP (400GE) * `400gbase-x-cfp8` - CPF8 (400GE) * `800gbase-x-qsfpdd` - QSFP-DD (800GE) * `800gbase-x-osfp` - OSFP (800GE) * `1000base-kx` - 1000BASE-KX (1GE) * `10gbase-kr` - 10GBASE-KR (10GE) * `10gbase-kx4` - 10GBASE-KX4 (10GE) * `25gbase-kr` - 25GBASE-KR (25GE) * `40gbase-kr4` - 40GBASE-KR4 (40GE) * `50gbase-kr` - 50GBASE-KR (50GE) * `100gbase-kp4` - 100GBASE-KP4 (100GE) * `100gbase-kr2` - 100GBASE-KR2 (100GE) * `100gbase-kr4` - 100GBASE-KR4 (100GE) * `ieee802.11a` - IEEE 802.11a * `ieee802.11g` - IEEE 802.11b/g * `ieee802.11n` - IEEE 802.11n * `ieee802.11ac` - IEEE 802.11ac * `ieee802.11ad` - IEEE 802.11ad * `ieee802.11ax` - IEEE 802.11ax * `ieee802.11ay` - IEEE 802.11ay * `ieee802.15.1` - IEEE 802.15.1 (Bluetooth) * `other-wireless` - Other (Wireless) * `gsm` - GSM * `cdma` - CDMA * `lte` - LTE * `sonet-oc3` - OC-3/STM-1 * `sonet-oc12` - OC-12/STM-4 * `sonet-oc48` - OC-48/STM-16 * `sonet-oc192` - OC-192/STM-64 * `sonet-oc768` - OC-768/STM-256 * `sonet-oc1920` - OC-1920/STM-640 * `sonet-oc3840` - OC-3840/STM-1234 * `1gfc-sfp` - SFP (1GFC) * `2gfc-sfp` - SFP (2GFC) * `4gfc-sfp` - SFP (4GFC) * `8gfc-sfpp` - SFP+ (8GFC) * `16gfc-sfpp` - SFP+ (16GFC) * `32gfc-sfp28` - SFP28 (32GFC) * `64gfc-qsfpp` - QSFP+ (64GFC) * `128gfc-qsfp28` - QSFP28 (128GFC) * `infiniband-sdr` - SDR (2 Gbps) * `infiniband-ddr` - DDR (4 Gbps) * `infiniband-qdr` - QDR (8 Gbps) * `infiniband-fdr10` - FDR10 (10 Gbps) * `infiniband-fdr` - FDR (13.5 Gbps) * `infiniband-edr` - EDR (25 Gbps) * `infiniband-hdr` - HDR (50 Gbps) * `infiniband-ndr` - NDR (100 Gbps) * `infiniband-xdr` - XDR (250 Gbps) * `t1` - T1 (1.544 Mbps) * `e1` - E1 (2.048 Mbps) * `t3` - T3 (45 Mbps) * `e3` - E3 (34 Mbps) * `xdsl` - xDSL * `docsis` - DOCSIS * `gpon` - GPON (2.5 Gbps / 1.25 Gps) * `xg-pon` - XG-PON (10 Gbps / 2.5 Gbps) * `xgs-pon` - XGS-PON (10 Gbps) * `ng-pon2` - NG-PON2 (TWDM-PON) (4x10 Gbps) * `epon` - EPON (1 Gbps) * `10g-epon` - 10G-EPON (10 Gbps) * `cisco-stackwise` - Cisco StackWise * `cisco-stackwise-plus` - Cisco StackWise Plus * `cisco-flexstack` - Cisco FlexStack * `cisco-flexstack-plus` - Cisco FlexStack Plus * `cisco-stackwise-80` - Cisco StackWise-80 * `cisco-stackwise-160` - Cisco StackWise-160 * `cisco-stackwise-320` - Cisco StackWise-320 * `cisco-stackwise-480` - Cisco StackWise-480 * `cisco-stackwise-1t` - Cisco StackWise-1T * `juniper-vcp` - Juniper VCP * `extreme-summitstack` - Extreme SummitStack * `extreme-summitstack-128` - Extreme SummitStack-128 * `extreme-summitstack-256` - Extreme SummitStack-256 * `extreme-summitstack-512` - Extreme SummitStack-512 * `other` - Other
type InterfaceTypeValue string

// List of Interface_type_value
const (
	INTERFACETYPEVALUE_VIRTUAL                 InterfaceTypeValue = "virtual"
	INTERFACETYPEVALUE_BRIDGE                  InterfaceTypeValue = "bridge"
	INTERFACETYPEVALUE_LAG                     InterfaceTypeValue = "lag"
	INTERFACETYPEVALUE__100BASE_FX             InterfaceTypeValue = "100base-fx"
	INTERFACETYPEVALUE__100BASE_LFX            InterfaceTypeValue = "100base-lfx"
	INTERFACETYPEVALUE__100BASE_TX             InterfaceTypeValue = "100base-tx"
	INTERFACETYPEVALUE__100BASE_T1             InterfaceTypeValue = "100base-t1"
	INTERFACETYPEVALUE__1000BASE_T             InterfaceTypeValue = "1000base-t"
	INTERFACETYPEVALUE__2_5GBASE_T             InterfaceTypeValue = "2.5gbase-t"
	INTERFACETYPEVALUE__5GBASE_T               InterfaceTypeValue = "5gbase-t"
	INTERFACETYPEVALUE__10GBASE_T              InterfaceTypeValue = "10gbase-t"
	INTERFACETYPEVALUE__10GBASE_CX4            InterfaceTypeValue = "10gbase-cx4"
	INTERFACETYPEVALUE__1000BASE_X_GBIC        InterfaceTypeValue = "1000base-x-gbic"
	INTERFACETYPEVALUE__1000BASE_X_SFP         InterfaceTypeValue = "1000base-x-sfp"
	INTERFACETYPEVALUE__10GBASE_X_SFPP         InterfaceTypeValue = "10gbase-x-sfpp"
	INTERFACETYPEVALUE__10GBASE_X_XFP          InterfaceTypeValue = "10gbase-x-xfp"
	INTERFACETYPEVALUE__10GBASE_X_XENPAK       InterfaceTypeValue = "10gbase-x-xenpak"
	INTERFACETYPEVALUE__10GBASE_X_X2           InterfaceTypeValue = "10gbase-x-x2"
	INTERFACETYPEVALUE__25GBASE_X_SFP28        InterfaceTypeValue = "25gbase-x-sfp28"
	INTERFACETYPEVALUE__50GBASE_X_SFP56        InterfaceTypeValue = "50gbase-x-sfp56"
	INTERFACETYPEVALUE__40GBASE_X_QSFPP        InterfaceTypeValue = "40gbase-x-qsfpp"
	INTERFACETYPEVALUE__50GBASE_X_SFP28        InterfaceTypeValue = "50gbase-x-sfp28"
	INTERFACETYPEVALUE__100GBASE_X_CFP         InterfaceTypeValue = "100gbase-x-cfp"
	INTERFACETYPEVALUE__100GBASE_X_CFP2        InterfaceTypeValue = "100gbase-x-cfp2"
	INTERFACETYPEVALUE__200GBASE_X_CFP2        InterfaceTypeValue = "200gbase-x-cfp2"
	INTERFACETYPEVALUE__400GBASE_X_CFP2        InterfaceTypeValue = "400gbase-x-cfp2"
	INTERFACETYPEVALUE__100GBASE_X_CFP4        InterfaceTypeValue = "100gbase-x-cfp4"
	INTERFACETYPEVALUE__100GBASE_X_CXP         InterfaceTypeValue = "100gbase-x-cxp"
	INTERFACETYPEVALUE__100GBASE_X_CPAK        InterfaceTypeValue = "100gbase-x-cpak"
	INTERFACETYPEVALUE__100GBASE_X_DSFP        InterfaceTypeValue = "100gbase-x-dsfp"
	INTERFACETYPEVALUE__100GBASE_X_SFPDD       InterfaceTypeValue = "100gbase-x-sfpdd"
	INTERFACETYPEVALUE__100GBASE_X_QSFP28      InterfaceTypeValue = "100gbase-x-qsfp28"
	INTERFACETYPEVALUE__100GBASE_X_QSFPDD      InterfaceTypeValue = "100gbase-x-qsfpdd"
	INTERFACETYPEVALUE__200GBASE_X_QSFP56      InterfaceTypeValue = "200gbase-x-qsfp56"
	INTERFACETYPEVALUE__200GBASE_X_QSFPDD      InterfaceTypeValue = "200gbase-x-qsfpdd"
	INTERFACETYPEVALUE__400GBASE_X_QSFP112     InterfaceTypeValue = "400gbase-x-qsfp112"
	INTERFACETYPEVALUE__400GBASE_X_QSFPDD      InterfaceTypeValue = "400gbase-x-qsfpdd"
	INTERFACETYPEVALUE__400GBASE_X_OSFP        InterfaceTypeValue = "400gbase-x-osfp"
	INTERFACETYPEVALUE__400GBASE_X_OSFP_RHS    InterfaceTypeValue = "400gbase-x-osfp-rhs"
	INTERFACETYPEVALUE__400GBASE_X_CDFP        InterfaceTypeValue = "400gbase-x-cdfp"
	INTERFACETYPEVALUE__400GBASE_X_CFP8        InterfaceTypeValue = "400gbase-x-cfp8"
	INTERFACETYPEVALUE__800GBASE_X_QSFPDD      InterfaceTypeValue = "800gbase-x-qsfpdd"
	INTERFACETYPEVALUE__800GBASE_X_OSFP        InterfaceTypeValue = "800gbase-x-osfp"
	INTERFACETYPEVALUE__1000BASE_KX            InterfaceTypeValue = "1000base-kx"
	INTERFACETYPEVALUE__10GBASE_KR             InterfaceTypeValue = "10gbase-kr"
	INTERFACETYPEVALUE__10GBASE_KX4            InterfaceTypeValue = "10gbase-kx4"
	INTERFACETYPEVALUE__25GBASE_KR             InterfaceTypeValue = "25gbase-kr"
	INTERFACETYPEVALUE__40GBASE_KR4            InterfaceTypeValue = "40gbase-kr4"
	INTERFACETYPEVALUE__50GBASE_KR             InterfaceTypeValue = "50gbase-kr"
	INTERFACETYPEVALUE__100GBASE_KP4           InterfaceTypeValue = "100gbase-kp4"
	INTERFACETYPEVALUE__100GBASE_KR2           InterfaceTypeValue = "100gbase-kr2"
	INTERFACETYPEVALUE__100GBASE_KR4           InterfaceTypeValue = "100gbase-kr4"
	INTERFACETYPEVALUE_IEEE802_11A             InterfaceTypeValue = "ieee802.11a"
	INTERFACETYPEVALUE_IEEE802_11G             InterfaceTypeValue = "ieee802.11g"
	INTERFACETYPEVALUE_IEEE802_11N             InterfaceTypeValue = "ieee802.11n"
	INTERFACETYPEVALUE_IEEE802_11AC            InterfaceTypeValue = "ieee802.11ac"
	INTERFACETYPEVALUE_IEEE802_11AD            InterfaceTypeValue = "ieee802.11ad"
	INTERFACETYPEVALUE_IEEE802_11AX            InterfaceTypeValue = "ieee802.11ax"
	INTERFACETYPEVALUE_IEEE802_11AY            InterfaceTypeValue = "ieee802.11ay"
	INTERFACETYPEVALUE_IEEE802_15_1            InterfaceTypeValue = "ieee802.15.1"
	INTERFACETYPEVALUE_OTHER_WIRELESS          InterfaceTypeValue = "other-wireless"
	INTERFACETYPEVALUE_GSM                     InterfaceTypeValue = "gsm"
	INTERFACETYPEVALUE_CDMA                    InterfaceTypeValue = "cdma"
	INTERFACETYPEVALUE_LTE                     InterfaceTypeValue = "lte"
	INTERFACETYPEVALUE_SONET_OC3               InterfaceTypeValue = "sonet-oc3"
	INTERFACETYPEVALUE_SONET_OC12              InterfaceTypeValue = "sonet-oc12"
	INTERFACETYPEVALUE_SONET_OC48              InterfaceTypeValue = "sonet-oc48"
	INTERFACETYPEVALUE_SONET_OC192             InterfaceTypeValue = "sonet-oc192"
	INTERFACETYPEVALUE_SONET_OC768             InterfaceTypeValue = "sonet-oc768"
	INTERFACETYPEVALUE_SONET_OC1920            InterfaceTypeValue = "sonet-oc1920"
	INTERFACETYPEVALUE_SONET_OC3840            InterfaceTypeValue = "sonet-oc3840"
	INTERFACETYPEVALUE__1GFC_SFP               InterfaceTypeValue = "1gfc-sfp"
	INTERFACETYPEVALUE__2GFC_SFP               InterfaceTypeValue = "2gfc-sfp"
	INTERFACETYPEVALUE__4GFC_SFP               InterfaceTypeValue = "4gfc-sfp"
	INTERFACETYPEVALUE__8GFC_SFPP              InterfaceTypeValue = "8gfc-sfpp"
	INTERFACETYPEVALUE__16GFC_SFPP             InterfaceTypeValue = "16gfc-sfpp"
	INTERFACETYPEVALUE__32GFC_SFP28            InterfaceTypeValue = "32gfc-sfp28"
	INTERFACETYPEVALUE__64GFC_QSFPP            InterfaceTypeValue = "64gfc-qsfpp"
	INTERFACETYPEVALUE__128GFC_QSFP28          InterfaceTypeValue = "128gfc-qsfp28"
	INTERFACETYPEVALUE_INFINIBAND_SDR          InterfaceTypeValue = "infiniband-sdr"
	INTERFACETYPEVALUE_INFINIBAND_DDR          InterfaceTypeValue = "infiniband-ddr"
	INTERFACETYPEVALUE_INFINIBAND_QDR          InterfaceTypeValue = "infiniband-qdr"
	INTERFACETYPEVALUE_INFINIBAND_FDR10        InterfaceTypeValue = "infiniband-fdr10"
	INTERFACETYPEVALUE_INFINIBAND_FDR          InterfaceTypeValue = "infiniband-fdr"
	INTERFACETYPEVALUE_INFINIBAND_EDR          InterfaceTypeValue = "infiniband-edr"
	INTERFACETYPEVALUE_INFINIBAND_HDR          InterfaceTypeValue = "infiniband-hdr"
	INTERFACETYPEVALUE_INFINIBAND_NDR          InterfaceTypeValue = "infiniband-ndr"
	INTERFACETYPEVALUE_INFINIBAND_XDR          InterfaceTypeValue = "infiniband-xdr"
	INTERFACETYPEVALUE_T1                      InterfaceTypeValue = "t1"
	INTERFACETYPEVALUE_E1                      InterfaceTypeValue = "e1"
	INTERFACETYPEVALUE_T3                      InterfaceTypeValue = "t3"
	INTERFACETYPEVALUE_E3                      InterfaceTypeValue = "e3"
	INTERFACETYPEVALUE_XDSL                    InterfaceTypeValue = "xdsl"
	INTERFACETYPEVALUE_DOCSIS                  InterfaceTypeValue = "docsis"
	INTERFACETYPEVALUE_GPON                    InterfaceTypeValue = "gpon"
	INTERFACETYPEVALUE_XG_PON                  InterfaceTypeValue = "xg-pon"
	INTERFACETYPEVALUE_XGS_PON                 InterfaceTypeValue = "xgs-pon"
	INTERFACETYPEVALUE_NG_PON2                 InterfaceTypeValue = "ng-pon2"
	INTERFACETYPEVALUE_EPON                    InterfaceTypeValue = "epon"
	INTERFACETYPEVALUE__10G_EPON               InterfaceTypeValue = "10g-epon"
	INTERFACETYPEVALUE_CISCO_STACKWISE         InterfaceTypeValue = "cisco-stackwise"
	INTERFACETYPEVALUE_CISCO_STACKWISE_PLUS    InterfaceTypeValue = "cisco-stackwise-plus"
	INTERFACETYPEVALUE_CISCO_FLEXSTACK         InterfaceTypeValue = "cisco-flexstack"
	INTERFACETYPEVALUE_CISCO_FLEXSTACK_PLUS    InterfaceTypeValue = "cisco-flexstack-plus"
	INTERFACETYPEVALUE_CISCO_STACKWISE_80      InterfaceTypeValue = "cisco-stackwise-80"
	INTERFACETYPEVALUE_CISCO_STACKWISE_160     InterfaceTypeValue = "cisco-stackwise-160"
	INTERFACETYPEVALUE_CISCO_STACKWISE_320     InterfaceTypeValue = "cisco-stackwise-320"
	INTERFACETYPEVALUE_CISCO_STACKWISE_480     InterfaceTypeValue = "cisco-stackwise-480"
	INTERFACETYPEVALUE_CISCO_STACKWISE_1T      InterfaceTypeValue = "cisco-stackwise-1t"
	INTERFACETYPEVALUE_JUNIPER_VCP             InterfaceTypeValue = "juniper-vcp"
	INTERFACETYPEVALUE_EXTREME_SUMMITSTACK     InterfaceTypeValue = "extreme-summitstack"
	INTERFACETYPEVALUE_EXTREME_SUMMITSTACK_128 InterfaceTypeValue = "extreme-summitstack-128"
	INTERFACETYPEVALUE_EXTREME_SUMMITSTACK_256 InterfaceTypeValue = "extreme-summitstack-256"
	INTERFACETYPEVALUE_EXTREME_SUMMITSTACK_512 InterfaceTypeValue = "extreme-summitstack-512"
	INTERFACETYPEVALUE_OTHER                   InterfaceTypeValue = "other"
)

// All allowed values of InterfaceTypeValue enum
var AllowedInterfaceTypeValueEnumValues = []InterfaceTypeValue{
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

func (v *InterfaceTypeValue) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := InterfaceTypeValue(value)
	for _, existing := range AllowedInterfaceTypeValueEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid InterfaceTypeValue", value)
}

// NewInterfaceTypeValueFromValue returns a pointer to a valid InterfaceTypeValue
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewInterfaceTypeValueFromValue(v string) (*InterfaceTypeValue, error) {
	ev := InterfaceTypeValue(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for InterfaceTypeValue: valid values are %v", v, AllowedInterfaceTypeValueEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v InterfaceTypeValue) IsValid() bool {
	for _, existing := range AllowedInterfaceTypeValueEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Interface_type_value value
func (v InterfaceTypeValue) Ptr() *InterfaceTypeValue {
	return &v
}

type NullableInterfaceTypeValue struct {
	value *InterfaceTypeValue
	isSet bool
}

func (v NullableInterfaceTypeValue) Get() *InterfaceTypeValue {
	return v.value
}

func (v *NullableInterfaceTypeValue) Set(val *InterfaceTypeValue) {
	v.value = val
	v.isSet = true
}

func (v NullableInterfaceTypeValue) IsSet() bool {
	return v.isSet
}

func (v *NullableInterfaceTypeValue) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableInterfaceTypeValue(val *InterfaceTypeValue) *NullableInterfaceTypeValue {
	return &NullableInterfaceTypeValue{value: val, isSet: true}
}

func (v NullableInterfaceTypeValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableInterfaceTypeValue) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
