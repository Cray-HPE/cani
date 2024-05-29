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

// ConsolePortSpeedValue * `1200` - 1200 bps * `2400` - 2400 bps * `4800` - 4800 bps * `9600` - 9600 bps * `19200` - 19.2 kbps * `38400` - 38.4 kbps * `57600` - 57.6 kbps * `115200` - 115.2 kbps
type ConsolePortSpeedValue int32

// List of ConsolePort_speed_value
const (
	CONSOLEPORTSPEEDVALUE__1200   ConsolePortSpeedValue = 1200
	CONSOLEPORTSPEEDVALUE__2400   ConsolePortSpeedValue = 2400
	CONSOLEPORTSPEEDVALUE__4800   ConsolePortSpeedValue = 4800
	CONSOLEPORTSPEEDVALUE__9600   ConsolePortSpeedValue = 9600
	CONSOLEPORTSPEEDVALUE__19200  ConsolePortSpeedValue = 19200
	CONSOLEPORTSPEEDVALUE__38400  ConsolePortSpeedValue = 38400
	CONSOLEPORTSPEEDVALUE__57600  ConsolePortSpeedValue = 57600
	CONSOLEPORTSPEEDVALUE__115200 ConsolePortSpeedValue = 115200
)

// All allowed values of ConsolePortSpeedValue enum
var AllowedConsolePortSpeedValueEnumValues = []ConsolePortSpeedValue{
	1200,
	2400,
	4800,
	9600,
	19200,
	38400,
	57600,
	115200,
}

func (v *ConsolePortSpeedValue) UnmarshalJSON(src []byte) error {
	var value int32
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := ConsolePortSpeedValue(value)
	for _, existing := range AllowedConsolePortSpeedValueEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid ConsolePortSpeedValue", value)
}

// NewConsolePortSpeedValueFromValue returns a pointer to a valid ConsolePortSpeedValue
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewConsolePortSpeedValueFromValue(v int32) (*ConsolePortSpeedValue, error) {
	ev := ConsolePortSpeedValue(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for ConsolePortSpeedValue: valid values are %v", v, AllowedConsolePortSpeedValueEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v ConsolePortSpeedValue) IsValid() bool {
	for _, existing := range AllowedConsolePortSpeedValueEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to ConsolePort_speed_value value
func (v ConsolePortSpeedValue) Ptr() *ConsolePortSpeedValue {
	return &v
}

type NullableConsolePortSpeedValue struct {
	value *ConsolePortSpeedValue
	isSet bool
}

func (v NullableConsolePortSpeedValue) Get() *ConsolePortSpeedValue {
	return v.value
}

func (v *NullableConsolePortSpeedValue) Set(val *ConsolePortSpeedValue) {
	v.value = val
	v.isSet = true
}

func (v NullableConsolePortSpeedValue) IsSet() bool {
	return v.isSet
}

func (v *NullableConsolePortSpeedValue) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableConsolePortSpeedValue(val *ConsolePortSpeedValue) *NullableConsolePortSpeedValue {
	return &NullableConsolePortSpeedValue{value: val, isSet: true}
}

func (v NullableConsolePortSpeedValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableConsolePortSpeedValue) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
