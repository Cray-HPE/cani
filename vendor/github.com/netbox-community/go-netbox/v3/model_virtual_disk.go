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
	"time"
)

// checks if the VirtualDisk type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &VirtualDisk{}

// VirtualDisk Adds support for custom fields and tags.
type VirtualDisk struct {
	Id                   int32                  `json:"id"`
	Url                  string                 `json:"url"`
	VirtualMachine       NestedVirtualMachine   `json:"virtual_machine"`
	Name                 string                 `json:"name"`
	Description          *string                `json:"description,omitempty"`
	Size                 int32                  `json:"size"`
	Tags                 []NestedTag            `json:"tags,omitempty"`
	CustomFields         map[string]interface{} `json:"custom_fields,omitempty"`
	Created              NullableTime           `json:"created"`
	LastUpdated          NullableTime           `json:"last_updated"`
	AdditionalProperties map[string]interface{}
}

type _VirtualDisk VirtualDisk

// NewVirtualDisk instantiates a new VirtualDisk object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewVirtualDisk(id int32, url string, virtualMachine NestedVirtualMachine, name string, size int32, created NullableTime, lastUpdated NullableTime) *VirtualDisk {
	this := VirtualDisk{}
	this.Id = id
	this.Url = url
	this.VirtualMachine = virtualMachine
	this.Name = name
	this.Size = size
	this.Created = created
	this.LastUpdated = lastUpdated
	return &this
}

// NewVirtualDiskWithDefaults instantiates a new VirtualDisk object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewVirtualDiskWithDefaults() *VirtualDisk {
	this := VirtualDisk{}
	return &this
}

// GetId returns the Id field value
func (o *VirtualDisk) GetId() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *VirtualDisk) GetIdOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *VirtualDisk) SetId(v int32) {
	o.Id = v
}

// GetUrl returns the Url field value
func (o *VirtualDisk) GetUrl() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Url
}

// GetUrlOk returns a tuple with the Url field value
// and a boolean to check if the value has been set.
func (o *VirtualDisk) GetUrlOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Url, true
}

// SetUrl sets field value
func (o *VirtualDisk) SetUrl(v string) {
	o.Url = v
}

// GetVirtualMachine returns the VirtualMachine field value
func (o *VirtualDisk) GetVirtualMachine() NestedVirtualMachine {
	if o == nil {
		var ret NestedVirtualMachine
		return ret
	}

	return o.VirtualMachine
}

// GetVirtualMachineOk returns a tuple with the VirtualMachine field value
// and a boolean to check if the value has been set.
func (o *VirtualDisk) GetVirtualMachineOk() (*NestedVirtualMachine, bool) {
	if o == nil {
		return nil, false
	}
	return &o.VirtualMachine, true
}

// SetVirtualMachine sets field value
func (o *VirtualDisk) SetVirtualMachine(v NestedVirtualMachine) {
	o.VirtualMachine = v
}

// GetName returns the Name field value
func (o *VirtualDisk) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *VirtualDisk) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *VirtualDisk) SetName(v string) {
	o.Name = v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *VirtualDisk) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VirtualDisk) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *VirtualDisk) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *VirtualDisk) SetDescription(v string) {
	o.Description = &v
}

// GetSize returns the Size field value
func (o *VirtualDisk) GetSize() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Size
}

// GetSizeOk returns a tuple with the Size field value
// and a boolean to check if the value has been set.
func (o *VirtualDisk) GetSizeOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Size, true
}

// SetSize sets field value
func (o *VirtualDisk) SetSize(v int32) {
	o.Size = v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *VirtualDisk) GetTags() []NestedTag {
	if o == nil || IsNil(o.Tags) {
		var ret []NestedTag
		return ret
	}
	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VirtualDisk) GetTagsOk() ([]NestedTag, bool) {
	if o == nil || IsNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *VirtualDisk) HasTags() bool {
	if o != nil && !IsNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given []NestedTag and assigns it to the Tags field.
func (o *VirtualDisk) SetTags(v []NestedTag) {
	o.Tags = v
}

// GetCustomFields returns the CustomFields field value if set, zero value otherwise.
func (o *VirtualDisk) GetCustomFields() map[string]interface{} {
	if o == nil || IsNil(o.CustomFields) {
		var ret map[string]interface{}
		return ret
	}
	return o.CustomFields
}

// GetCustomFieldsOk returns a tuple with the CustomFields field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VirtualDisk) GetCustomFieldsOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.CustomFields) {
		return map[string]interface{}{}, false
	}
	return o.CustomFields, true
}

// HasCustomFields returns a boolean if a field has been set.
func (o *VirtualDisk) HasCustomFields() bool {
	if o != nil && !IsNil(o.CustomFields) {
		return true
	}

	return false
}

// SetCustomFields gets a reference to the given map[string]interface{} and assigns it to the CustomFields field.
func (o *VirtualDisk) SetCustomFields(v map[string]interface{}) {
	o.CustomFields = v
}

// GetCreated returns the Created field value
// If the value is explicit nil, the zero value for time.Time will be returned
func (o *VirtualDisk) GetCreated() time.Time {
	if o == nil || o.Created.Get() == nil {
		var ret time.Time
		return ret
	}

	return *o.Created.Get()
}

// GetCreatedOk returns a tuple with the Created field value
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *VirtualDisk) GetCreatedOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return o.Created.Get(), o.Created.IsSet()
}

// SetCreated sets field value
func (o *VirtualDisk) SetCreated(v time.Time) {
	o.Created.Set(&v)
}

// GetLastUpdated returns the LastUpdated field value
// If the value is explicit nil, the zero value for time.Time will be returned
func (o *VirtualDisk) GetLastUpdated() time.Time {
	if o == nil || o.LastUpdated.Get() == nil {
		var ret time.Time
		return ret
	}

	return *o.LastUpdated.Get()
}

// GetLastUpdatedOk returns a tuple with the LastUpdated field value
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *VirtualDisk) GetLastUpdatedOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return o.LastUpdated.Get(), o.LastUpdated.IsSet()
}

// SetLastUpdated sets field value
func (o *VirtualDisk) SetLastUpdated(v time.Time) {
	o.LastUpdated.Set(&v)
}

func (o VirtualDisk) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o VirtualDisk) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["id"] = o.Id
	toSerialize["url"] = o.Url
	toSerialize["virtual_machine"] = o.VirtualMachine
	toSerialize["name"] = o.Name
	if !IsNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	toSerialize["size"] = o.Size
	if !IsNil(o.Tags) {
		toSerialize["tags"] = o.Tags
	}
	if !IsNil(o.CustomFields) {
		toSerialize["custom_fields"] = o.CustomFields
	}
	toSerialize["created"] = o.Created.Get()
	toSerialize["last_updated"] = o.LastUpdated.Get()

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *VirtualDisk) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"id",
		"url",
		"virtual_machine",
		"name",
		"size",
		"created",
		"last_updated",
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

	varVirtualDisk := _VirtualDisk{}

	err = json.Unmarshal(data, &varVirtualDisk)

	if err != nil {
		return err
	}

	*o = VirtualDisk(varVirtualDisk)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "id")
		delete(additionalProperties, "url")
		delete(additionalProperties, "virtual_machine")
		delete(additionalProperties, "name")
		delete(additionalProperties, "description")
		delete(additionalProperties, "size")
		delete(additionalProperties, "tags")
		delete(additionalProperties, "custom_fields")
		delete(additionalProperties, "created")
		delete(additionalProperties, "last_updated")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableVirtualDisk struct {
	value *VirtualDisk
	isSet bool
}

func (v NullableVirtualDisk) Get() *VirtualDisk {
	return v.value
}

func (v *NullableVirtualDisk) Set(val *VirtualDisk) {
	v.value = val
	v.isSet = true
}

func (v NullableVirtualDisk) IsSet() bool {
	return v.isSet
}

func (v *NullableVirtualDisk) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableVirtualDisk(val *VirtualDisk) *NullableVirtualDisk {
	return &NullableVirtualDisk{value: val, isSet: true}
}

func (v NullableVirtualDisk) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableVirtualDisk) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
