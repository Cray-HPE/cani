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

// checks if the NestedPowerPanel type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &NestedPowerPanel{}

// NestedPowerPanel Represents an object related through a ForeignKey field. On write, it accepts a primary key (PK) value or a dictionary of attributes which can be used to uniquely identify the related object. This class should be subclassed to return a full representation of the related object on read.
type NestedPowerPanel struct {
	Id                   int32  `json:"id"`
	Url                  string `json:"url"`
	Display              string `json:"display"`
	Name                 string `json:"name"`
	AdditionalProperties map[string]interface{}
}

type _NestedPowerPanel NestedPowerPanel

// NewNestedPowerPanel instantiates a new NestedPowerPanel object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewNestedPowerPanel(id int32, url string, display string, name string) *NestedPowerPanel {
	this := NestedPowerPanel{}
	this.Id = id
	this.Url = url
	this.Display = display
	this.Name = name
	return &this
}

// NewNestedPowerPanelWithDefaults instantiates a new NestedPowerPanel object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewNestedPowerPanelWithDefaults() *NestedPowerPanel {
	this := NestedPowerPanel{}
	return &this
}

// GetId returns the Id field value
func (o *NestedPowerPanel) GetId() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *NestedPowerPanel) GetIdOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *NestedPowerPanel) SetId(v int32) {
	o.Id = v
}

// GetUrl returns the Url field value
func (o *NestedPowerPanel) GetUrl() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Url
}

// GetUrlOk returns a tuple with the Url field value
// and a boolean to check if the value has been set.
func (o *NestedPowerPanel) GetUrlOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Url, true
}

// SetUrl sets field value
func (o *NestedPowerPanel) SetUrl(v string) {
	o.Url = v
}

// GetDisplay returns the Display field value
func (o *NestedPowerPanel) GetDisplay() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Display
}

// GetDisplayOk returns a tuple with the Display field value
// and a boolean to check if the value has been set.
func (o *NestedPowerPanel) GetDisplayOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Display, true
}

// SetDisplay sets field value
func (o *NestedPowerPanel) SetDisplay(v string) {
	o.Display = v
}

// GetName returns the Name field value
func (o *NestedPowerPanel) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *NestedPowerPanel) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *NestedPowerPanel) SetName(v string) {
	o.Name = v
}

func (o NestedPowerPanel) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o NestedPowerPanel) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["id"] = o.Id
	toSerialize["url"] = o.Url
	toSerialize["display"] = o.Display
	toSerialize["name"] = o.Name

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *NestedPowerPanel) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"id",
		"url",
		"display",
		"name",
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

	varNestedPowerPanel := _NestedPowerPanel{}

	err = json.Unmarshal(data, &varNestedPowerPanel)

	if err != nil {
		return err
	}

	*o = NestedPowerPanel(varNestedPowerPanel)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "id")
		delete(additionalProperties, "url")
		delete(additionalProperties, "display")
		delete(additionalProperties, "name")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableNestedPowerPanel struct {
	value *NestedPowerPanel
	isSet bool
}

func (v NullableNestedPowerPanel) Get() *NestedPowerPanel {
	return v.value
}

func (v *NullableNestedPowerPanel) Set(val *NestedPowerPanel) {
	v.value = val
	v.isSet = true
}

func (v NullableNestedPowerPanel) IsSet() bool {
	return v.isSet
}

func (v *NullableNestedPowerPanel) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableNestedPowerPanel(val *NestedPowerPanel) *NullableNestedPowerPanel {
	return &NullableNestedPowerPanel{value: val, isSet: true}
}

func (v NullableNestedPowerPanel) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableNestedPowerPanel) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
