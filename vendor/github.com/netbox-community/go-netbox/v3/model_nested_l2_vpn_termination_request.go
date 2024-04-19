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

// checks if the NestedL2VPNTerminationRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &NestedL2VPNTerminationRequest{}

// NestedL2VPNTerminationRequest Represents an object related through a ForeignKey field. On write, it accepts a primary key (PK) value or a dictionary of attributes which can be used to uniquely identify the related object. This class should be subclassed to return a full representation of the related object on read.
type NestedL2VPNTerminationRequest struct {
	L2vpn                NestedL2VPNRequest `json:"l2vpn"`
	AdditionalProperties map[string]interface{}
}

type _NestedL2VPNTerminationRequest NestedL2VPNTerminationRequest

// NewNestedL2VPNTerminationRequest instantiates a new NestedL2VPNTerminationRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewNestedL2VPNTerminationRequest(l2vpn NestedL2VPNRequest) *NestedL2VPNTerminationRequest {
	this := NestedL2VPNTerminationRequest{}
	this.L2vpn = l2vpn
	return &this
}

// NewNestedL2VPNTerminationRequestWithDefaults instantiates a new NestedL2VPNTerminationRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewNestedL2VPNTerminationRequestWithDefaults() *NestedL2VPNTerminationRequest {
	this := NestedL2VPNTerminationRequest{}
	return &this
}

// GetL2vpn returns the L2vpn field value
func (o *NestedL2VPNTerminationRequest) GetL2vpn() NestedL2VPNRequest {
	if o == nil {
		var ret NestedL2VPNRequest
		return ret
	}

	return o.L2vpn
}

// GetL2vpnOk returns a tuple with the L2vpn field value
// and a boolean to check if the value has been set.
func (o *NestedL2VPNTerminationRequest) GetL2vpnOk() (*NestedL2VPNRequest, bool) {
	if o == nil {
		return nil, false
	}
	return &o.L2vpn, true
}

// SetL2vpn sets field value
func (o *NestedL2VPNTerminationRequest) SetL2vpn(v NestedL2VPNRequest) {
	o.L2vpn = v
}

func (o NestedL2VPNTerminationRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o NestedL2VPNTerminationRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["l2vpn"] = o.L2vpn

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *NestedL2VPNTerminationRequest) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"l2vpn",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err
	}

	for _, requiredProperty := range requiredProperties {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varNestedL2VPNTerminationRequest := _NestedL2VPNTerminationRequest{}

	err = json.Unmarshal(data, &varNestedL2VPNTerminationRequest)

	if err != nil {
		return err
	}

	*o = NestedL2VPNTerminationRequest(varNestedL2VPNTerminationRequest)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "l2vpn")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableNestedL2VPNTerminationRequest struct {
	value *NestedL2VPNTerminationRequest
	isSet bool
}

func (v NullableNestedL2VPNTerminationRequest) Get() *NestedL2VPNTerminationRequest {
	return v.value
}

func (v *NullableNestedL2VPNTerminationRequest) Set(val *NestedL2VPNTerminationRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableNestedL2VPNTerminationRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableNestedL2VPNTerminationRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableNestedL2VPNTerminationRequest(val *NestedL2VPNTerminationRequest) *NullableNestedL2VPNTerminationRequest {
	return &NullableNestedL2VPNTerminationRequest{value: val, isSet: true}
}

func (v NullableNestedL2VPNTerminationRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableNestedL2VPNTerminationRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
