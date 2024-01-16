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

// IKEProposalGroupLabel the model 'IKEProposalGroupLabel'
type IKEProposalGroupLabel string

// List of IKEProposal_group_label
const (
	IKEPROPOSALGROUPLABEL__1  IKEProposalGroupLabel = "Group 1"
	IKEPROPOSALGROUPLABEL__2  IKEProposalGroupLabel = "Group 2"
	IKEPROPOSALGROUPLABEL__5  IKEProposalGroupLabel = "Group 5"
	IKEPROPOSALGROUPLABEL__14 IKEProposalGroupLabel = "Group 14"
	IKEPROPOSALGROUPLABEL__15 IKEProposalGroupLabel = "Group 15"
	IKEPROPOSALGROUPLABEL__16 IKEProposalGroupLabel = "Group 16"
	IKEPROPOSALGROUPLABEL__17 IKEProposalGroupLabel = "Group 17"
	IKEPROPOSALGROUPLABEL__18 IKEProposalGroupLabel = "Group 18"
	IKEPROPOSALGROUPLABEL__19 IKEProposalGroupLabel = "Group 19"
	IKEPROPOSALGROUPLABEL__20 IKEProposalGroupLabel = "Group 20"
	IKEPROPOSALGROUPLABEL__21 IKEProposalGroupLabel = "Group 21"
	IKEPROPOSALGROUPLABEL__22 IKEProposalGroupLabel = "Group 22"
	IKEPROPOSALGROUPLABEL__23 IKEProposalGroupLabel = "Group 23"
	IKEPROPOSALGROUPLABEL__24 IKEProposalGroupLabel = "Group 24"
	IKEPROPOSALGROUPLABEL__25 IKEProposalGroupLabel = "Group 25"
	IKEPROPOSALGROUPLABEL__26 IKEProposalGroupLabel = "Group 26"
	IKEPROPOSALGROUPLABEL__27 IKEProposalGroupLabel = "Group 27"
	IKEPROPOSALGROUPLABEL__28 IKEProposalGroupLabel = "Group 28"
	IKEPROPOSALGROUPLABEL__29 IKEProposalGroupLabel = "Group 29"
	IKEPROPOSALGROUPLABEL__30 IKEProposalGroupLabel = "Group 30"
	IKEPROPOSALGROUPLABEL__31 IKEProposalGroupLabel = "Group 31"
	IKEPROPOSALGROUPLABEL__32 IKEProposalGroupLabel = "Group 32"
	IKEPROPOSALGROUPLABEL__33 IKEProposalGroupLabel = "Group 33"
	IKEPROPOSALGROUPLABEL__34 IKEProposalGroupLabel = "Group 34"
)

// All allowed values of IKEProposalGroupLabel enum
var AllowedIKEProposalGroupLabelEnumValues = []IKEProposalGroupLabel{
	"Group 1",
	"Group 2",
	"Group 5",
	"Group 14",
	"Group 15",
	"Group 16",
	"Group 17",
	"Group 18",
	"Group 19",
	"Group 20",
	"Group 21",
	"Group 22",
	"Group 23",
	"Group 24",
	"Group 25",
	"Group 26",
	"Group 27",
	"Group 28",
	"Group 29",
	"Group 30",
	"Group 31",
	"Group 32",
	"Group 33",
	"Group 34",
}

func (v *IKEProposalGroupLabel) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := IKEProposalGroupLabel(value)
	for _, existing := range AllowedIKEProposalGroupLabelEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid IKEProposalGroupLabel", value)
}

// NewIKEProposalGroupLabelFromValue returns a pointer to a valid IKEProposalGroupLabel
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewIKEProposalGroupLabelFromValue(v string) (*IKEProposalGroupLabel, error) {
	ev := IKEProposalGroupLabel(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for IKEProposalGroupLabel: valid values are %v", v, AllowedIKEProposalGroupLabelEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v IKEProposalGroupLabel) IsValid() bool {
	for _, existing := range AllowedIKEProposalGroupLabelEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to IKEProposal_group_label value
func (v IKEProposalGroupLabel) Ptr() *IKEProposalGroupLabel {
	return &v
}

type NullableIKEProposalGroupLabel struct {
	value *IKEProposalGroupLabel
	isSet bool
}

func (v NullableIKEProposalGroupLabel) Get() *IKEProposalGroupLabel {
	return v.value
}

func (v *NullableIKEProposalGroupLabel) Set(val *IKEProposalGroupLabel) {
	v.value = val
	v.isSet = true
}

func (v NullableIKEProposalGroupLabel) IsSet() bool {
	return v.isSet
}

func (v *NullableIKEProposalGroupLabel) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableIKEProposalGroupLabel(val *IKEProposalGroupLabel) *NullableIKEProposalGroupLabel {
	return &NullableIKEProposalGroupLabel{value: val, isSet: true}
}

func (v NullableIKEProposalGroupLabel) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableIKEProposalGroupLabel) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
