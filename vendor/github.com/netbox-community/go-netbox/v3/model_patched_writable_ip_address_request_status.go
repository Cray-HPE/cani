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

// PatchedWritableIPAddressRequestStatus The operational status of this IP  * `active` - Active * `reserved` - Reserved * `deprecated` - Deprecated * `dhcp` - DHCP * `slaac` - SLAAC
type PatchedWritableIPAddressRequestStatus string

// List of PatchedWritableIPAddressRequest_status
const (
	PATCHEDWRITABLEIPADDRESSREQUESTSTATUS_ACTIVE     PatchedWritableIPAddressRequestStatus = "active"
	PATCHEDWRITABLEIPADDRESSREQUESTSTATUS_RESERVED   PatchedWritableIPAddressRequestStatus = "reserved"
	PATCHEDWRITABLEIPADDRESSREQUESTSTATUS_DEPRECATED PatchedWritableIPAddressRequestStatus = "deprecated"
	PATCHEDWRITABLEIPADDRESSREQUESTSTATUS_DHCP       PatchedWritableIPAddressRequestStatus = "dhcp"
	PATCHEDWRITABLEIPADDRESSREQUESTSTATUS_SLAAC      PatchedWritableIPAddressRequestStatus = "slaac"
)

// All allowed values of PatchedWritableIPAddressRequestStatus enum
var AllowedPatchedWritableIPAddressRequestStatusEnumValues = []PatchedWritableIPAddressRequestStatus{
	"active",
	"reserved",
	"deprecated",
	"dhcp",
	"slaac",
}

func (v *PatchedWritableIPAddressRequestStatus) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := PatchedWritableIPAddressRequestStatus(value)
	for _, existing := range AllowedPatchedWritableIPAddressRequestStatusEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid PatchedWritableIPAddressRequestStatus", value)
}

// NewPatchedWritableIPAddressRequestStatusFromValue returns a pointer to a valid PatchedWritableIPAddressRequestStatus
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewPatchedWritableIPAddressRequestStatusFromValue(v string) (*PatchedWritableIPAddressRequestStatus, error) {
	ev := PatchedWritableIPAddressRequestStatus(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for PatchedWritableIPAddressRequestStatus: valid values are %v", v, AllowedPatchedWritableIPAddressRequestStatusEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v PatchedWritableIPAddressRequestStatus) IsValid() bool {
	for _, existing := range AllowedPatchedWritableIPAddressRequestStatusEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to PatchedWritableIPAddressRequest_status value
func (v PatchedWritableIPAddressRequestStatus) Ptr() *PatchedWritableIPAddressRequestStatus {
	return &v
}

type NullablePatchedWritableIPAddressRequestStatus struct {
	value *PatchedWritableIPAddressRequestStatus
	isSet bool
}

func (v NullablePatchedWritableIPAddressRequestStatus) Get() *PatchedWritableIPAddressRequestStatus {
	return v.value
}

func (v *NullablePatchedWritableIPAddressRequestStatus) Set(val *PatchedWritableIPAddressRequestStatus) {
	v.value = val
	v.isSet = true
}

func (v NullablePatchedWritableIPAddressRequestStatus) IsSet() bool {
	return v.isSet
}

func (v *NullablePatchedWritableIPAddressRequestStatus) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePatchedWritableIPAddressRequestStatus(val *PatchedWritableIPAddressRequestStatus) *NullablePatchedWritableIPAddressRequestStatus {
	return &NullablePatchedWritableIPAddressRequestStatus{value: val, isSet: true}
}

func (v NullablePatchedWritableIPAddressRequestStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePatchedWritableIPAddressRequestStatus) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
