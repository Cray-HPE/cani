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

// PatchedWritableVLANRequestStatus Operational status of this VLAN  * `active` - Active * `reserved` - Reserved * `deprecated` - Deprecated
type PatchedWritableVLANRequestStatus string

// List of PatchedWritableVLANRequest_status
const (
	PATCHEDWRITABLEVLANREQUESTSTATUS_ACTIVE     PatchedWritableVLANRequestStatus = "active"
	PATCHEDWRITABLEVLANREQUESTSTATUS_RESERVED   PatchedWritableVLANRequestStatus = "reserved"
	PATCHEDWRITABLEVLANREQUESTSTATUS_DEPRECATED PatchedWritableVLANRequestStatus = "deprecated"
)

// All allowed values of PatchedWritableVLANRequestStatus enum
var AllowedPatchedWritableVLANRequestStatusEnumValues = []PatchedWritableVLANRequestStatus{
	"active",
	"reserved",
	"deprecated",
}

func (v *PatchedWritableVLANRequestStatus) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := PatchedWritableVLANRequestStatus(value)
	for _, existing := range AllowedPatchedWritableVLANRequestStatusEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid PatchedWritableVLANRequestStatus", value)
}

// NewPatchedWritableVLANRequestStatusFromValue returns a pointer to a valid PatchedWritableVLANRequestStatus
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewPatchedWritableVLANRequestStatusFromValue(v string) (*PatchedWritableVLANRequestStatus, error) {
	ev := PatchedWritableVLANRequestStatus(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for PatchedWritableVLANRequestStatus: valid values are %v", v, AllowedPatchedWritableVLANRequestStatusEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v PatchedWritableVLANRequestStatus) IsValid() bool {
	for _, existing := range AllowedPatchedWritableVLANRequestStatusEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to PatchedWritableVLANRequest_status value
func (v PatchedWritableVLANRequestStatus) Ptr() *PatchedWritableVLANRequestStatus {
	return &v
}

type NullablePatchedWritableVLANRequestStatus struct {
	value *PatchedWritableVLANRequestStatus
	isSet bool
}

func (v NullablePatchedWritableVLANRequestStatus) Get() *PatchedWritableVLANRequestStatus {
	return v.value
}

func (v *NullablePatchedWritableVLANRequestStatus) Set(val *PatchedWritableVLANRequestStatus) {
	v.value = val
	v.isSet = true
}

func (v NullablePatchedWritableVLANRequestStatus) IsSet() bool {
	return v.isSet
}

func (v *NullablePatchedWritableVLANRequestStatus) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePatchedWritableVLANRequestStatus(val *PatchedWritableVLANRequestStatus) *NullablePatchedWritableVLANRequestStatus {
	return &NullablePatchedWritableVLANRequestStatus{value: val, isSet: true}
}

func (v NullablePatchedWritableVLANRequestStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePatchedWritableVLANRequestStatus) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
