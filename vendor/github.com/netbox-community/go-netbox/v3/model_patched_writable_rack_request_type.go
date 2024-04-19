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

// PatchedWritableRackRequestType * `2-post-frame` - 2-post frame * `4-post-frame` - 4-post frame * `4-post-cabinet` - 4-post cabinet * `wall-frame` - Wall-mounted frame * `wall-frame-vertical` - Wall-mounted frame (vertical) * `wall-cabinet` - Wall-mounted cabinet * `wall-cabinet-vertical` - Wall-mounted cabinet (vertical)
type PatchedWritableRackRequestType string

// List of PatchedWritableRackRequest_type
const (
	PATCHEDWRITABLERACKREQUESTTYPE__2_POST_FRAME         PatchedWritableRackRequestType = "2-post-frame"
	PATCHEDWRITABLERACKREQUESTTYPE__4_POST_FRAME         PatchedWritableRackRequestType = "4-post-frame"
	PATCHEDWRITABLERACKREQUESTTYPE__4_POST_CABINET       PatchedWritableRackRequestType = "4-post-cabinet"
	PATCHEDWRITABLERACKREQUESTTYPE_WALL_FRAME            PatchedWritableRackRequestType = "wall-frame"
	PATCHEDWRITABLERACKREQUESTTYPE_WALL_FRAME_VERTICAL   PatchedWritableRackRequestType = "wall-frame-vertical"
	PATCHEDWRITABLERACKREQUESTTYPE_WALL_CABINET          PatchedWritableRackRequestType = "wall-cabinet"
	PATCHEDWRITABLERACKREQUESTTYPE_WALL_CABINET_VERTICAL PatchedWritableRackRequestType = "wall-cabinet-vertical"
	PATCHEDWRITABLERACKREQUESTTYPE_EMPTY                 PatchedWritableRackRequestType = ""
)

// All allowed values of PatchedWritableRackRequestType enum
var AllowedPatchedWritableRackRequestTypeEnumValues = []PatchedWritableRackRequestType{
	"2-post-frame",
	"4-post-frame",
	"4-post-cabinet",
	"wall-frame",
	"wall-frame-vertical",
	"wall-cabinet",
	"wall-cabinet-vertical",
	"",
}

func (v *PatchedWritableRackRequestType) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := PatchedWritableRackRequestType(value)
	for _, existing := range AllowedPatchedWritableRackRequestTypeEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid PatchedWritableRackRequestType", value)
}

// NewPatchedWritableRackRequestTypeFromValue returns a pointer to a valid PatchedWritableRackRequestType
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewPatchedWritableRackRequestTypeFromValue(v string) (*PatchedWritableRackRequestType, error) {
	ev := PatchedWritableRackRequestType(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for PatchedWritableRackRequestType: valid values are %v", v, AllowedPatchedWritableRackRequestTypeEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v PatchedWritableRackRequestType) IsValid() bool {
	for _, existing := range AllowedPatchedWritableRackRequestTypeEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to PatchedWritableRackRequest_type value
func (v PatchedWritableRackRequestType) Ptr() *PatchedWritableRackRequestType {
	return &v
}

type NullablePatchedWritableRackRequestType struct {
	value *PatchedWritableRackRequestType
	isSet bool
}

func (v NullablePatchedWritableRackRequestType) Get() *PatchedWritableRackRequestType {
	return v.value
}

func (v *NullablePatchedWritableRackRequestType) Set(val *PatchedWritableRackRequestType) {
	v.value = val
	v.isSet = true
}

func (v NullablePatchedWritableRackRequestType) IsSet() bool {
	return v.isSet
}

func (v *NullablePatchedWritableRackRequestType) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePatchedWritableRackRequestType(val *PatchedWritableRackRequestType) *NullablePatchedWritableRackRequestType {
	return &NullablePatchedWritableRackRequestType{value: val, isSet: true}
}

func (v NullablePatchedWritableRackRequestType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePatchedWritableRackRequestType) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
