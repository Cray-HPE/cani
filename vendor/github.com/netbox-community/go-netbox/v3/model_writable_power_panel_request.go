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

// checks if the WritablePowerPanelRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &WritablePowerPanelRequest{}

// WritablePowerPanelRequest Adds support for custom fields and tags.
type WritablePowerPanelRequest struct {
	Site                 int32                  `json:"site"`
	Location             NullableInt32          `json:"location,omitempty"`
	Name                 string                 `json:"name"`
	Description          *string                `json:"description,omitempty"`
	Comments             *string                `json:"comments,omitempty"`
	Tags                 []NestedTagRequest     `json:"tags,omitempty"`
	CustomFields         map[string]interface{} `json:"custom_fields,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _WritablePowerPanelRequest WritablePowerPanelRequest

// NewWritablePowerPanelRequest instantiates a new WritablePowerPanelRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewWritablePowerPanelRequest(site int32, name string) *WritablePowerPanelRequest {
	this := WritablePowerPanelRequest{}
	this.Site = site
	this.Name = name
	return &this
}

// NewWritablePowerPanelRequestWithDefaults instantiates a new WritablePowerPanelRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewWritablePowerPanelRequestWithDefaults() *WritablePowerPanelRequest {
	this := WritablePowerPanelRequest{}
	return &this
}

// GetSite returns the Site field value
func (o *WritablePowerPanelRequest) GetSite() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Site
}

// GetSiteOk returns a tuple with the Site field value
// and a boolean to check if the value has been set.
func (o *WritablePowerPanelRequest) GetSiteOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Site, true
}

// SetSite sets field value
func (o *WritablePowerPanelRequest) SetSite(v int32) {
	o.Site = v
}

// GetLocation returns the Location field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *WritablePowerPanelRequest) GetLocation() int32 {
	if o == nil || IsNil(o.Location.Get()) {
		var ret int32
		return ret
	}
	return *o.Location.Get()
}

// GetLocationOk returns a tuple with the Location field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *WritablePowerPanelRequest) GetLocationOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return o.Location.Get(), o.Location.IsSet()
}

// HasLocation returns a boolean if a field has been set.
func (o *WritablePowerPanelRequest) HasLocation() bool {
	if o != nil && o.Location.IsSet() {
		return true
	}

	return false
}

// SetLocation gets a reference to the given NullableInt32 and assigns it to the Location field.
func (o *WritablePowerPanelRequest) SetLocation(v int32) {
	o.Location.Set(&v)
}

// SetLocationNil sets the value for Location to be an explicit nil
func (o *WritablePowerPanelRequest) SetLocationNil() {
	o.Location.Set(nil)
}

// UnsetLocation ensures that no value is present for Location, not even an explicit nil
func (o *WritablePowerPanelRequest) UnsetLocation() {
	o.Location.Unset()
}

// GetName returns the Name field value
func (o *WritablePowerPanelRequest) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *WritablePowerPanelRequest) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *WritablePowerPanelRequest) SetName(v string) {
	o.Name = v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *WritablePowerPanelRequest) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WritablePowerPanelRequest) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *WritablePowerPanelRequest) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *WritablePowerPanelRequest) SetDescription(v string) {
	o.Description = &v
}

// GetComments returns the Comments field value if set, zero value otherwise.
func (o *WritablePowerPanelRequest) GetComments() string {
	if o == nil || IsNil(o.Comments) {
		var ret string
		return ret
	}
	return *o.Comments
}

// GetCommentsOk returns a tuple with the Comments field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WritablePowerPanelRequest) GetCommentsOk() (*string, bool) {
	if o == nil || IsNil(o.Comments) {
		return nil, false
	}
	return o.Comments, true
}

// HasComments returns a boolean if a field has been set.
func (o *WritablePowerPanelRequest) HasComments() bool {
	if o != nil && !IsNil(o.Comments) {
		return true
	}

	return false
}

// SetComments gets a reference to the given string and assigns it to the Comments field.
func (o *WritablePowerPanelRequest) SetComments(v string) {
	o.Comments = &v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *WritablePowerPanelRequest) GetTags() []NestedTagRequest {
	if o == nil || IsNil(o.Tags) {
		var ret []NestedTagRequest
		return ret
	}
	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WritablePowerPanelRequest) GetTagsOk() ([]NestedTagRequest, bool) {
	if o == nil || IsNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *WritablePowerPanelRequest) HasTags() bool {
	if o != nil && !IsNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given []NestedTagRequest and assigns it to the Tags field.
func (o *WritablePowerPanelRequest) SetTags(v []NestedTagRequest) {
	o.Tags = v
}

// GetCustomFields returns the CustomFields field value if set, zero value otherwise.
func (o *WritablePowerPanelRequest) GetCustomFields() map[string]interface{} {
	if o == nil || IsNil(o.CustomFields) {
		var ret map[string]interface{}
		return ret
	}
	return o.CustomFields
}

// GetCustomFieldsOk returns a tuple with the CustomFields field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WritablePowerPanelRequest) GetCustomFieldsOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.CustomFields) {
		return map[string]interface{}{}, false
	}
	return o.CustomFields, true
}

// HasCustomFields returns a boolean if a field has been set.
func (o *WritablePowerPanelRequest) HasCustomFields() bool {
	if o != nil && !IsNil(o.CustomFields) {
		return true
	}

	return false
}

// SetCustomFields gets a reference to the given map[string]interface{} and assigns it to the CustomFields field.
func (o *WritablePowerPanelRequest) SetCustomFields(v map[string]interface{}) {
	o.CustomFields = v
}

func (o WritablePowerPanelRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o WritablePowerPanelRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["site"] = o.Site
	if o.Location.IsSet() {
		toSerialize["location"] = o.Location.Get()
	}
	toSerialize["name"] = o.Name
	if !IsNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	if !IsNil(o.Comments) {
		toSerialize["comments"] = o.Comments
	}
	if !IsNil(o.Tags) {
		toSerialize["tags"] = o.Tags
	}
	if !IsNil(o.CustomFields) {
		toSerialize["custom_fields"] = o.CustomFields
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *WritablePowerPanelRequest) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"site",
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

	varWritablePowerPanelRequest := _WritablePowerPanelRequest{}

	err = json.Unmarshal(data, &varWritablePowerPanelRequest)

	if err != nil {
		return err
	}

	*o = WritablePowerPanelRequest(varWritablePowerPanelRequest)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "site")
		delete(additionalProperties, "location")
		delete(additionalProperties, "name")
		delete(additionalProperties, "description")
		delete(additionalProperties, "comments")
		delete(additionalProperties, "tags")
		delete(additionalProperties, "custom_fields")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableWritablePowerPanelRequest struct {
	value *WritablePowerPanelRequest
	isSet bool
}

func (v NullableWritablePowerPanelRequest) Get() *WritablePowerPanelRequest {
	return v.value
}

func (v *NullableWritablePowerPanelRequest) Set(val *WritablePowerPanelRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableWritablePowerPanelRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableWritablePowerPanelRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableWritablePowerPanelRequest(val *WritablePowerPanelRequest) *NullableWritablePowerPanelRequest {
	return &NullableWritablePowerPanelRequest{value: val, isSet: true}
}

func (v NullableWritablePowerPanelRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableWritablePowerPanelRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
