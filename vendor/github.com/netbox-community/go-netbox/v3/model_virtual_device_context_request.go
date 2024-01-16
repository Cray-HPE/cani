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

// checks if the VirtualDeviceContextRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &VirtualDeviceContextRequest{}

// VirtualDeviceContextRequest Adds support for custom fields and tags.
type VirtualDeviceContextRequest struct {
	Name   string              `json:"name"`
	Device NestedDeviceRequest `json:"device"`
	// Numeric identifier unique to the parent device
	Identifier           NullableInt32                                    `json:"identifier,omitempty"`
	Tenant               NullableNestedTenantRequest                      `json:"tenant,omitempty"`
	PrimaryIp4           NullableNestedIPAddressRequest                   `json:"primary_ip4,omitempty"`
	PrimaryIp6           NullableNestedIPAddressRequest                   `json:"primary_ip6,omitempty"`
	Status               PatchedWritableVirtualDeviceContextRequestStatus `json:"status"`
	Description          *string                                          `json:"description,omitempty"`
	Comments             *string                                          `json:"comments,omitempty"`
	Tags                 []NestedTagRequest                               `json:"tags,omitempty"`
	CustomFields         map[string]interface{}                           `json:"custom_fields,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _VirtualDeviceContextRequest VirtualDeviceContextRequest

// NewVirtualDeviceContextRequest instantiates a new VirtualDeviceContextRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewVirtualDeviceContextRequest(name string, device NestedDeviceRequest, status PatchedWritableVirtualDeviceContextRequestStatus) *VirtualDeviceContextRequest {
	this := VirtualDeviceContextRequest{}
	this.Name = name
	this.Device = device
	this.Status = status
	return &this
}

// NewVirtualDeviceContextRequestWithDefaults instantiates a new VirtualDeviceContextRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewVirtualDeviceContextRequestWithDefaults() *VirtualDeviceContextRequest {
	this := VirtualDeviceContextRequest{}
	return &this
}

// GetName returns the Name field value
func (o *VirtualDeviceContextRequest) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *VirtualDeviceContextRequest) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *VirtualDeviceContextRequest) SetName(v string) {
	o.Name = v
}

// GetDevice returns the Device field value
func (o *VirtualDeviceContextRequest) GetDevice() NestedDeviceRequest {
	if o == nil {
		var ret NestedDeviceRequest
		return ret
	}

	return o.Device
}

// GetDeviceOk returns a tuple with the Device field value
// and a boolean to check if the value has been set.
func (o *VirtualDeviceContextRequest) GetDeviceOk() (*NestedDeviceRequest, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Device, true
}

// SetDevice sets field value
func (o *VirtualDeviceContextRequest) SetDevice(v NestedDeviceRequest) {
	o.Device = v
}

// GetIdentifier returns the Identifier field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *VirtualDeviceContextRequest) GetIdentifier() int32 {
	if o == nil || IsNil(o.Identifier.Get()) {
		var ret int32
		return ret
	}
	return *o.Identifier.Get()
}

// GetIdentifierOk returns a tuple with the Identifier field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *VirtualDeviceContextRequest) GetIdentifierOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return o.Identifier.Get(), o.Identifier.IsSet()
}

// HasIdentifier returns a boolean if a field has been set.
func (o *VirtualDeviceContextRequest) HasIdentifier() bool {
	if o != nil && o.Identifier.IsSet() {
		return true
	}

	return false
}

// SetIdentifier gets a reference to the given NullableInt32 and assigns it to the Identifier field.
func (o *VirtualDeviceContextRequest) SetIdentifier(v int32) {
	o.Identifier.Set(&v)
}

// SetIdentifierNil sets the value for Identifier to be an explicit nil
func (o *VirtualDeviceContextRequest) SetIdentifierNil() {
	o.Identifier.Set(nil)
}

// UnsetIdentifier ensures that no value is present for Identifier, not even an explicit nil
func (o *VirtualDeviceContextRequest) UnsetIdentifier() {
	o.Identifier.Unset()
}

// GetTenant returns the Tenant field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *VirtualDeviceContextRequest) GetTenant() NestedTenantRequest {
	if o == nil || IsNil(o.Tenant.Get()) {
		var ret NestedTenantRequest
		return ret
	}
	return *o.Tenant.Get()
}

// GetTenantOk returns a tuple with the Tenant field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *VirtualDeviceContextRequest) GetTenantOk() (*NestedTenantRequest, bool) {
	if o == nil {
		return nil, false
	}
	return o.Tenant.Get(), o.Tenant.IsSet()
}

// HasTenant returns a boolean if a field has been set.
func (o *VirtualDeviceContextRequest) HasTenant() bool {
	if o != nil && o.Tenant.IsSet() {
		return true
	}

	return false
}

// SetTenant gets a reference to the given NullableNestedTenantRequest and assigns it to the Tenant field.
func (o *VirtualDeviceContextRequest) SetTenant(v NestedTenantRequest) {
	o.Tenant.Set(&v)
}

// SetTenantNil sets the value for Tenant to be an explicit nil
func (o *VirtualDeviceContextRequest) SetTenantNil() {
	o.Tenant.Set(nil)
}

// UnsetTenant ensures that no value is present for Tenant, not even an explicit nil
func (o *VirtualDeviceContextRequest) UnsetTenant() {
	o.Tenant.Unset()
}

// GetPrimaryIp4 returns the PrimaryIp4 field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *VirtualDeviceContextRequest) GetPrimaryIp4() NestedIPAddressRequest {
	if o == nil || IsNil(o.PrimaryIp4.Get()) {
		var ret NestedIPAddressRequest
		return ret
	}
	return *o.PrimaryIp4.Get()
}

// GetPrimaryIp4Ok returns a tuple with the PrimaryIp4 field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *VirtualDeviceContextRequest) GetPrimaryIp4Ok() (*NestedIPAddressRequest, bool) {
	if o == nil {
		return nil, false
	}
	return o.PrimaryIp4.Get(), o.PrimaryIp4.IsSet()
}

// HasPrimaryIp4 returns a boolean if a field has been set.
func (o *VirtualDeviceContextRequest) HasPrimaryIp4() bool {
	if o != nil && o.PrimaryIp4.IsSet() {
		return true
	}

	return false
}

// SetPrimaryIp4 gets a reference to the given NullableNestedIPAddressRequest and assigns it to the PrimaryIp4 field.
func (o *VirtualDeviceContextRequest) SetPrimaryIp4(v NestedIPAddressRequest) {
	o.PrimaryIp4.Set(&v)
}

// SetPrimaryIp4Nil sets the value for PrimaryIp4 to be an explicit nil
func (o *VirtualDeviceContextRequest) SetPrimaryIp4Nil() {
	o.PrimaryIp4.Set(nil)
}

// UnsetPrimaryIp4 ensures that no value is present for PrimaryIp4, not even an explicit nil
func (o *VirtualDeviceContextRequest) UnsetPrimaryIp4() {
	o.PrimaryIp4.Unset()
}

// GetPrimaryIp6 returns the PrimaryIp6 field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *VirtualDeviceContextRequest) GetPrimaryIp6() NestedIPAddressRequest {
	if o == nil || IsNil(o.PrimaryIp6.Get()) {
		var ret NestedIPAddressRequest
		return ret
	}
	return *o.PrimaryIp6.Get()
}

// GetPrimaryIp6Ok returns a tuple with the PrimaryIp6 field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *VirtualDeviceContextRequest) GetPrimaryIp6Ok() (*NestedIPAddressRequest, bool) {
	if o == nil {
		return nil, false
	}
	return o.PrimaryIp6.Get(), o.PrimaryIp6.IsSet()
}

// HasPrimaryIp6 returns a boolean if a field has been set.
func (o *VirtualDeviceContextRequest) HasPrimaryIp6() bool {
	if o != nil && o.PrimaryIp6.IsSet() {
		return true
	}

	return false
}

// SetPrimaryIp6 gets a reference to the given NullableNestedIPAddressRequest and assigns it to the PrimaryIp6 field.
func (o *VirtualDeviceContextRequest) SetPrimaryIp6(v NestedIPAddressRequest) {
	o.PrimaryIp6.Set(&v)
}

// SetPrimaryIp6Nil sets the value for PrimaryIp6 to be an explicit nil
func (o *VirtualDeviceContextRequest) SetPrimaryIp6Nil() {
	o.PrimaryIp6.Set(nil)
}

// UnsetPrimaryIp6 ensures that no value is present for PrimaryIp6, not even an explicit nil
func (o *VirtualDeviceContextRequest) UnsetPrimaryIp6() {
	o.PrimaryIp6.Unset()
}

// GetStatus returns the Status field value
func (o *VirtualDeviceContextRequest) GetStatus() PatchedWritableVirtualDeviceContextRequestStatus {
	if o == nil {
		var ret PatchedWritableVirtualDeviceContextRequestStatus
		return ret
	}

	return o.Status
}

// GetStatusOk returns a tuple with the Status field value
// and a boolean to check if the value has been set.
func (o *VirtualDeviceContextRequest) GetStatusOk() (*PatchedWritableVirtualDeviceContextRequestStatus, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Status, true
}

// SetStatus sets field value
func (o *VirtualDeviceContextRequest) SetStatus(v PatchedWritableVirtualDeviceContextRequestStatus) {
	o.Status = v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *VirtualDeviceContextRequest) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VirtualDeviceContextRequest) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *VirtualDeviceContextRequest) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *VirtualDeviceContextRequest) SetDescription(v string) {
	o.Description = &v
}

// GetComments returns the Comments field value if set, zero value otherwise.
func (o *VirtualDeviceContextRequest) GetComments() string {
	if o == nil || IsNil(o.Comments) {
		var ret string
		return ret
	}
	return *o.Comments
}

// GetCommentsOk returns a tuple with the Comments field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VirtualDeviceContextRequest) GetCommentsOk() (*string, bool) {
	if o == nil || IsNil(o.Comments) {
		return nil, false
	}
	return o.Comments, true
}

// HasComments returns a boolean if a field has been set.
func (o *VirtualDeviceContextRequest) HasComments() bool {
	if o != nil && !IsNil(o.Comments) {
		return true
	}

	return false
}

// SetComments gets a reference to the given string and assigns it to the Comments field.
func (o *VirtualDeviceContextRequest) SetComments(v string) {
	o.Comments = &v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *VirtualDeviceContextRequest) GetTags() []NestedTagRequest {
	if o == nil || IsNil(o.Tags) {
		var ret []NestedTagRequest
		return ret
	}
	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VirtualDeviceContextRequest) GetTagsOk() ([]NestedTagRequest, bool) {
	if o == nil || IsNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *VirtualDeviceContextRequest) HasTags() bool {
	if o != nil && !IsNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given []NestedTagRequest and assigns it to the Tags field.
func (o *VirtualDeviceContextRequest) SetTags(v []NestedTagRequest) {
	o.Tags = v
}

// GetCustomFields returns the CustomFields field value if set, zero value otherwise.
func (o *VirtualDeviceContextRequest) GetCustomFields() map[string]interface{} {
	if o == nil || IsNil(o.CustomFields) {
		var ret map[string]interface{}
		return ret
	}
	return o.CustomFields
}

// GetCustomFieldsOk returns a tuple with the CustomFields field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VirtualDeviceContextRequest) GetCustomFieldsOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.CustomFields) {
		return map[string]interface{}{}, false
	}
	return o.CustomFields, true
}

// HasCustomFields returns a boolean if a field has been set.
func (o *VirtualDeviceContextRequest) HasCustomFields() bool {
	if o != nil && !IsNil(o.CustomFields) {
		return true
	}

	return false
}

// SetCustomFields gets a reference to the given map[string]interface{} and assigns it to the CustomFields field.
func (o *VirtualDeviceContextRequest) SetCustomFields(v map[string]interface{}) {
	o.CustomFields = v
}

func (o VirtualDeviceContextRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o VirtualDeviceContextRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["name"] = o.Name
	toSerialize["device"] = o.Device
	if o.Identifier.IsSet() {
		toSerialize["identifier"] = o.Identifier.Get()
	}
	if o.Tenant.IsSet() {
		toSerialize["tenant"] = o.Tenant.Get()
	}
	if o.PrimaryIp4.IsSet() {
		toSerialize["primary_ip4"] = o.PrimaryIp4.Get()
	}
	if o.PrimaryIp6.IsSet() {
		toSerialize["primary_ip6"] = o.PrimaryIp6.Get()
	}
	toSerialize["status"] = o.Status
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

func (o *VirtualDeviceContextRequest) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"name",
		"device",
		"status",
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

	varVirtualDeviceContextRequest := _VirtualDeviceContextRequest{}

	err = json.Unmarshal(data, &varVirtualDeviceContextRequest)

	if err != nil {
		return err
	}

	*o = VirtualDeviceContextRequest(varVirtualDeviceContextRequest)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "name")
		delete(additionalProperties, "device")
		delete(additionalProperties, "identifier")
		delete(additionalProperties, "tenant")
		delete(additionalProperties, "primary_ip4")
		delete(additionalProperties, "primary_ip6")
		delete(additionalProperties, "status")
		delete(additionalProperties, "description")
		delete(additionalProperties, "comments")
		delete(additionalProperties, "tags")
		delete(additionalProperties, "custom_fields")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableVirtualDeviceContextRequest struct {
	value *VirtualDeviceContextRequest
	isSet bool
}

func (v NullableVirtualDeviceContextRequest) Get() *VirtualDeviceContextRequest {
	return v.value
}

func (v *NullableVirtualDeviceContextRequest) Set(val *VirtualDeviceContextRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableVirtualDeviceContextRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableVirtualDeviceContextRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableVirtualDeviceContextRequest(val *VirtualDeviceContextRequest) *NullableVirtualDeviceContextRequest {
	return &NullableVirtualDeviceContextRequest{value: val, isSet: true}
}

func (v NullableVirtualDeviceContextRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableVirtualDeviceContextRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
