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

// checks if the AvailableIP type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &AvailableIP{}

// AvailableIP Representation of an IP address which does not exist in the database.
type AvailableIP struct {
	Family               int32     `json:"family"`
	Address              string    `json:"address"`
	Vrf                  NestedVRF `json:"vrf"`
	Description          *string   `json:"description,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _AvailableIP AvailableIP

// NewAvailableIP instantiates a new AvailableIP object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewAvailableIP(family int32, address string, vrf NestedVRF) *AvailableIP {
	this := AvailableIP{}
	this.Family = family
	this.Address = address
	this.Vrf = vrf
	return &this
}

// NewAvailableIPWithDefaults instantiates a new AvailableIP object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewAvailableIPWithDefaults() *AvailableIP {
	this := AvailableIP{}
	return &this
}

// GetFamily returns the Family field value
func (o *AvailableIP) GetFamily() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Family
}

// GetFamilyOk returns a tuple with the Family field value
// and a boolean to check if the value has been set.
func (o *AvailableIP) GetFamilyOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Family, true
}

// SetFamily sets field value
func (o *AvailableIP) SetFamily(v int32) {
	o.Family = v
}

// GetAddress returns the Address field value
func (o *AvailableIP) GetAddress() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Address
}

// GetAddressOk returns a tuple with the Address field value
// and a boolean to check if the value has been set.
func (o *AvailableIP) GetAddressOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Address, true
}

// SetAddress sets field value
func (o *AvailableIP) SetAddress(v string) {
	o.Address = v
}

// GetVrf returns the Vrf field value
func (o *AvailableIP) GetVrf() NestedVRF {
	if o == nil {
		var ret NestedVRF
		return ret
	}

	return o.Vrf
}

// GetVrfOk returns a tuple with the Vrf field value
// and a boolean to check if the value has been set.
func (o *AvailableIP) GetVrfOk() (*NestedVRF, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Vrf, true
}

// SetVrf sets field value
func (o *AvailableIP) SetVrf(v NestedVRF) {
	o.Vrf = v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *AvailableIP) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AvailableIP) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *AvailableIP) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *AvailableIP) SetDescription(v string) {
	o.Description = &v
}

func (o AvailableIP) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o AvailableIP) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["family"] = o.Family
	toSerialize["address"] = o.Address
	toSerialize["vrf"] = o.Vrf
	if !IsNil(o.Description) {
		toSerialize["description"] = o.Description
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *AvailableIP) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"family",
		"address",
		"vrf",
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

	varAvailableIP := _AvailableIP{}

	err = json.Unmarshal(data, &varAvailableIP)

	if err != nil {
		return err
	}

	*o = AvailableIP(varAvailableIP)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "family")
		delete(additionalProperties, "address")
		delete(additionalProperties, "vrf")
		delete(additionalProperties, "description")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableAvailableIP struct {
	value *AvailableIP
	isSet bool
}

func (v NullableAvailableIP) Get() *AvailableIP {
	return v.value
}

func (v *NullableAvailableIP) Set(val *AvailableIP) {
	v.value = val
	v.isSet = true
}

func (v NullableAvailableIP) IsSet() bool {
	return v.isSet
}

func (v *NullableAvailableIP) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableAvailableIP(val *AvailableIP) *NullableAvailableIP {
	return &NullableAvailableIP{value: val, isSet: true}
}

func (v NullableAvailableIP) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableAvailableIP) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}