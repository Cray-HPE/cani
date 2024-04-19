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

// InterfacePoeModeValue * `pd` - PD * `pse` - PSE
type InterfacePoeModeValue string

// List of Interface_poe_mode_value
const (
	INTERFACEPOEMODEVALUE_PD    InterfacePoeModeValue = "pd"
	INTERFACEPOEMODEVALUE_PSE   InterfacePoeModeValue = "pse"
	INTERFACEPOEMODEVALUE_EMPTY InterfacePoeModeValue = ""
)

// All allowed values of InterfacePoeModeValue enum
var AllowedInterfacePoeModeValueEnumValues = []InterfacePoeModeValue{
	"pd",
	"pse",
	"",
}

func (v *InterfacePoeModeValue) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := InterfacePoeModeValue(value)
	for _, existing := range AllowedInterfacePoeModeValueEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid InterfacePoeModeValue", value)
}

// NewInterfacePoeModeValueFromValue returns a pointer to a valid InterfacePoeModeValue
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewInterfacePoeModeValueFromValue(v string) (*InterfacePoeModeValue, error) {
	ev := InterfacePoeModeValue(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for InterfacePoeModeValue: valid values are %v", v, AllowedInterfacePoeModeValueEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v InterfacePoeModeValue) IsValid() bool {
	for _, existing := range AllowedInterfacePoeModeValueEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Interface_poe_mode_value value
func (v InterfacePoeModeValue) Ptr() *InterfacePoeModeValue {
	return &v
}

type NullableInterfacePoeModeValue struct {
	value *InterfacePoeModeValue
	isSet bool
}

func (v NullableInterfacePoeModeValue) Get() *InterfacePoeModeValue {
	return v.value
}

func (v *NullableInterfacePoeModeValue) Set(val *InterfacePoeModeValue) {
	v.value = val
	v.isSet = true
}

func (v NullableInterfacePoeModeValue) IsSet() bool {
	return v.isSet
}

func (v *NullableInterfacePoeModeValue) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableInterfacePoeModeValue(val *InterfacePoeModeValue) *NullableInterfacePoeModeValue {
	return &NullableInterfacePoeModeValue{value: val, isSet: true}
}

func (v NullableInterfacePoeModeValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableInterfacePoeModeValue) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
