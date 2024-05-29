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

// checks if the NestedTunnelGroup type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &NestedTunnelGroup{}

// NestedTunnelGroup Represents an object related through a ForeignKey field. On write, it accepts a primary key (PK) value or a dictionary of attributes which can be used to uniquely identify the related object. This class should be subclassed to return a full representation of the related object on read.
type NestedTunnelGroup struct {
	Id                   int32  `json:"id"`
	Url                  string `json:"url"`
	Display              string `json:"display"`
	Name                 string `json:"name"`
	Slug                 string `json:"slug"`
	AdditionalProperties map[string]interface{}
}

type _NestedTunnelGroup NestedTunnelGroup

// NewNestedTunnelGroup instantiates a new NestedTunnelGroup object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewNestedTunnelGroup(id int32, url string, display string, name string, slug string) *NestedTunnelGroup {
	this := NestedTunnelGroup{}
	this.Id = id
	this.Url = url
	this.Display = display
	this.Name = name
	this.Slug = slug
	return &this
}

// NewNestedTunnelGroupWithDefaults instantiates a new NestedTunnelGroup object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewNestedTunnelGroupWithDefaults() *NestedTunnelGroup {
	this := NestedTunnelGroup{}
	return &this
}

// GetId returns the Id field value
func (o *NestedTunnelGroup) GetId() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *NestedTunnelGroup) GetIdOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *NestedTunnelGroup) SetId(v int32) {
	o.Id = v
}

// GetUrl returns the Url field value
func (o *NestedTunnelGroup) GetUrl() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Url
}

// GetUrlOk returns a tuple with the Url field value
// and a boolean to check if the value has been set.
func (o *NestedTunnelGroup) GetUrlOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Url, true
}

// SetUrl sets field value
func (o *NestedTunnelGroup) SetUrl(v string) {
	o.Url = v
}

// GetDisplay returns the Display field value
func (o *NestedTunnelGroup) GetDisplay() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Display
}

// GetDisplayOk returns a tuple with the Display field value
// and a boolean to check if the value has been set.
func (o *NestedTunnelGroup) GetDisplayOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Display, true
}

// SetDisplay sets field value
func (o *NestedTunnelGroup) SetDisplay(v string) {
	o.Display = v
}

// GetName returns the Name field value
func (o *NestedTunnelGroup) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *NestedTunnelGroup) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *NestedTunnelGroup) SetName(v string) {
	o.Name = v
}

// GetSlug returns the Slug field value
func (o *NestedTunnelGroup) GetSlug() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Slug
}

// GetSlugOk returns a tuple with the Slug field value
// and a boolean to check if the value has been set.
func (o *NestedTunnelGroup) GetSlugOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Slug, true
}

// SetSlug sets field value
func (o *NestedTunnelGroup) SetSlug(v string) {
	o.Slug = v
}

func (o NestedTunnelGroup) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o NestedTunnelGroup) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["id"] = o.Id
	toSerialize["url"] = o.Url
	toSerialize["display"] = o.Display
	toSerialize["name"] = o.Name
	toSerialize["slug"] = o.Slug

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *NestedTunnelGroup) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"id",
		"url",
		"display",
		"name",
		"slug",
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

	varNestedTunnelGroup := _NestedTunnelGroup{}

	err = json.Unmarshal(data, &varNestedTunnelGroup)

	if err != nil {
		return err
	}

	*o = NestedTunnelGroup(varNestedTunnelGroup)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "id")
		delete(additionalProperties, "url")
		delete(additionalProperties, "display")
		delete(additionalProperties, "name")
		delete(additionalProperties, "slug")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableNestedTunnelGroup struct {
	value *NestedTunnelGroup
	isSet bool
}

func (v NullableNestedTunnelGroup) Get() *NestedTunnelGroup {
	return v.value
}

func (v *NullableNestedTunnelGroup) Set(val *NestedTunnelGroup) {
	v.value = val
	v.isSet = true
}

func (v NullableNestedTunnelGroup) IsSet() bool {
	return v.isSet
}

func (v *NullableNestedTunnelGroup) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableNestedTunnelGroup(val *NestedTunnelGroup) *NullableNestedTunnelGroup {
	return &NullableNestedTunnelGroup{value: val, isSet: true}
}

func (v NullableNestedTunnelGroup) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableNestedTunnelGroup) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
