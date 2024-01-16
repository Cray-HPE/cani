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

// checks if the WirelessLinkRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &WirelessLinkRequest{}

// WirelessLinkRequest Adds support for custom fields and tags.
type WirelessLinkRequest struct {
	InterfaceA           NestedInterfaceRequest      `json:"interface_a"`
	InterfaceB           NestedInterfaceRequest      `json:"interface_b"`
	Ssid                 *string                     `json:"ssid,omitempty"`
	Status               *CableStatusValue           `json:"status,omitempty"`
	Tenant               NullableNestedTenantRequest `json:"tenant,omitempty"`
	AuthType             *WirelessLANAuthTypeValue   `json:"auth_type,omitempty"`
	AuthCipher           *WirelessLANAuthCipherValue `json:"auth_cipher,omitempty"`
	AuthPsk              *string                     `json:"auth_psk,omitempty"`
	Description          *string                     `json:"description,omitempty"`
	Comments             *string                     `json:"comments,omitempty"`
	Tags                 []NestedTagRequest          `json:"tags,omitempty"`
	CustomFields         map[string]interface{}      `json:"custom_fields,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _WirelessLinkRequest WirelessLinkRequest

// NewWirelessLinkRequest instantiates a new WirelessLinkRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewWirelessLinkRequest(interfaceA NestedInterfaceRequest, interfaceB NestedInterfaceRequest) *WirelessLinkRequest {
	this := WirelessLinkRequest{}
	this.InterfaceA = interfaceA
	this.InterfaceB = interfaceB
	return &this
}

// NewWirelessLinkRequestWithDefaults instantiates a new WirelessLinkRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewWirelessLinkRequestWithDefaults() *WirelessLinkRequest {
	this := WirelessLinkRequest{}
	return &this
}

// GetInterfaceA returns the InterfaceA field value
func (o *WirelessLinkRequest) GetInterfaceA() NestedInterfaceRequest {
	if o == nil {
		var ret NestedInterfaceRequest
		return ret
	}

	return o.InterfaceA
}

// GetInterfaceAOk returns a tuple with the InterfaceA field value
// and a boolean to check if the value has been set.
func (o *WirelessLinkRequest) GetInterfaceAOk() (*NestedInterfaceRequest, bool) {
	if o == nil {
		return nil, false
	}
	return &o.InterfaceA, true
}

// SetInterfaceA sets field value
func (o *WirelessLinkRequest) SetInterfaceA(v NestedInterfaceRequest) {
	o.InterfaceA = v
}

// GetInterfaceB returns the InterfaceB field value
func (o *WirelessLinkRequest) GetInterfaceB() NestedInterfaceRequest {
	if o == nil {
		var ret NestedInterfaceRequest
		return ret
	}

	return o.InterfaceB
}

// GetInterfaceBOk returns a tuple with the InterfaceB field value
// and a boolean to check if the value has been set.
func (o *WirelessLinkRequest) GetInterfaceBOk() (*NestedInterfaceRequest, bool) {
	if o == nil {
		return nil, false
	}
	return &o.InterfaceB, true
}

// SetInterfaceB sets field value
func (o *WirelessLinkRequest) SetInterfaceB(v NestedInterfaceRequest) {
	o.InterfaceB = v
}

// GetSsid returns the Ssid field value if set, zero value otherwise.
func (o *WirelessLinkRequest) GetSsid() string {
	if o == nil || IsNil(o.Ssid) {
		var ret string
		return ret
	}
	return *o.Ssid
}

// GetSsidOk returns a tuple with the Ssid field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WirelessLinkRequest) GetSsidOk() (*string, bool) {
	if o == nil || IsNil(o.Ssid) {
		return nil, false
	}
	return o.Ssid, true
}

// HasSsid returns a boolean if a field has been set.
func (o *WirelessLinkRequest) HasSsid() bool {
	if o != nil && !IsNil(o.Ssid) {
		return true
	}

	return false
}

// SetSsid gets a reference to the given string and assigns it to the Ssid field.
func (o *WirelessLinkRequest) SetSsid(v string) {
	o.Ssid = &v
}

// GetStatus returns the Status field value if set, zero value otherwise.
func (o *WirelessLinkRequest) GetStatus() CableStatusValue {
	if o == nil || IsNil(o.Status) {
		var ret CableStatusValue
		return ret
	}
	return *o.Status
}

// GetStatusOk returns a tuple with the Status field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WirelessLinkRequest) GetStatusOk() (*CableStatusValue, bool) {
	if o == nil || IsNil(o.Status) {
		return nil, false
	}
	return o.Status, true
}

// HasStatus returns a boolean if a field has been set.
func (o *WirelessLinkRequest) HasStatus() bool {
	if o != nil && !IsNil(o.Status) {
		return true
	}

	return false
}

// SetStatus gets a reference to the given CableStatusValue and assigns it to the Status field.
func (o *WirelessLinkRequest) SetStatus(v CableStatusValue) {
	o.Status = &v
}

// GetTenant returns the Tenant field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *WirelessLinkRequest) GetTenant() NestedTenantRequest {
	if o == nil || IsNil(o.Tenant.Get()) {
		var ret NestedTenantRequest
		return ret
	}
	return *o.Tenant.Get()
}

// GetTenantOk returns a tuple with the Tenant field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *WirelessLinkRequest) GetTenantOk() (*NestedTenantRequest, bool) {
	if o == nil {
		return nil, false
	}
	return o.Tenant.Get(), o.Tenant.IsSet()
}

// HasTenant returns a boolean if a field has been set.
func (o *WirelessLinkRequest) HasTenant() bool {
	if o != nil && o.Tenant.IsSet() {
		return true
	}

	return false
}

// SetTenant gets a reference to the given NullableNestedTenantRequest and assigns it to the Tenant field.
func (o *WirelessLinkRequest) SetTenant(v NestedTenantRequest) {
	o.Tenant.Set(&v)
}

// SetTenantNil sets the value for Tenant to be an explicit nil
func (o *WirelessLinkRequest) SetTenantNil() {
	o.Tenant.Set(nil)
}

// UnsetTenant ensures that no value is present for Tenant, not even an explicit nil
func (o *WirelessLinkRequest) UnsetTenant() {
	o.Tenant.Unset()
}

// GetAuthType returns the AuthType field value if set, zero value otherwise.
func (o *WirelessLinkRequest) GetAuthType() WirelessLANAuthTypeValue {
	if o == nil || IsNil(o.AuthType) {
		var ret WirelessLANAuthTypeValue
		return ret
	}
	return *o.AuthType
}

// GetAuthTypeOk returns a tuple with the AuthType field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WirelessLinkRequest) GetAuthTypeOk() (*WirelessLANAuthTypeValue, bool) {
	if o == nil || IsNil(o.AuthType) {
		return nil, false
	}
	return o.AuthType, true
}

// HasAuthType returns a boolean if a field has been set.
func (o *WirelessLinkRequest) HasAuthType() bool {
	if o != nil && !IsNil(o.AuthType) {
		return true
	}

	return false
}

// SetAuthType gets a reference to the given WirelessLANAuthTypeValue and assigns it to the AuthType field.
func (o *WirelessLinkRequest) SetAuthType(v WirelessLANAuthTypeValue) {
	o.AuthType = &v
}

// GetAuthCipher returns the AuthCipher field value if set, zero value otherwise.
func (o *WirelessLinkRequest) GetAuthCipher() WirelessLANAuthCipherValue {
	if o == nil || IsNil(o.AuthCipher) {
		var ret WirelessLANAuthCipherValue
		return ret
	}
	return *o.AuthCipher
}

// GetAuthCipherOk returns a tuple with the AuthCipher field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WirelessLinkRequest) GetAuthCipherOk() (*WirelessLANAuthCipherValue, bool) {
	if o == nil || IsNil(o.AuthCipher) {
		return nil, false
	}
	return o.AuthCipher, true
}

// HasAuthCipher returns a boolean if a field has been set.
func (o *WirelessLinkRequest) HasAuthCipher() bool {
	if o != nil && !IsNil(o.AuthCipher) {
		return true
	}

	return false
}

// SetAuthCipher gets a reference to the given WirelessLANAuthCipherValue and assigns it to the AuthCipher field.
func (o *WirelessLinkRequest) SetAuthCipher(v WirelessLANAuthCipherValue) {
	o.AuthCipher = &v
}

// GetAuthPsk returns the AuthPsk field value if set, zero value otherwise.
func (o *WirelessLinkRequest) GetAuthPsk() string {
	if o == nil || IsNil(o.AuthPsk) {
		var ret string
		return ret
	}
	return *o.AuthPsk
}

// GetAuthPskOk returns a tuple with the AuthPsk field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WirelessLinkRequest) GetAuthPskOk() (*string, bool) {
	if o == nil || IsNil(o.AuthPsk) {
		return nil, false
	}
	return o.AuthPsk, true
}

// HasAuthPsk returns a boolean if a field has been set.
func (o *WirelessLinkRequest) HasAuthPsk() bool {
	if o != nil && !IsNil(o.AuthPsk) {
		return true
	}

	return false
}

// SetAuthPsk gets a reference to the given string and assigns it to the AuthPsk field.
func (o *WirelessLinkRequest) SetAuthPsk(v string) {
	o.AuthPsk = &v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *WirelessLinkRequest) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WirelessLinkRequest) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *WirelessLinkRequest) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *WirelessLinkRequest) SetDescription(v string) {
	o.Description = &v
}

// GetComments returns the Comments field value if set, zero value otherwise.
func (o *WirelessLinkRequest) GetComments() string {
	if o == nil || IsNil(o.Comments) {
		var ret string
		return ret
	}
	return *o.Comments
}

// GetCommentsOk returns a tuple with the Comments field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WirelessLinkRequest) GetCommentsOk() (*string, bool) {
	if o == nil || IsNil(o.Comments) {
		return nil, false
	}
	return o.Comments, true
}

// HasComments returns a boolean if a field has been set.
func (o *WirelessLinkRequest) HasComments() bool {
	if o != nil && !IsNil(o.Comments) {
		return true
	}

	return false
}

// SetComments gets a reference to the given string and assigns it to the Comments field.
func (o *WirelessLinkRequest) SetComments(v string) {
	o.Comments = &v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *WirelessLinkRequest) GetTags() []NestedTagRequest {
	if o == nil || IsNil(o.Tags) {
		var ret []NestedTagRequest
		return ret
	}
	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WirelessLinkRequest) GetTagsOk() ([]NestedTagRequest, bool) {
	if o == nil || IsNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *WirelessLinkRequest) HasTags() bool {
	if o != nil && !IsNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given []NestedTagRequest and assigns it to the Tags field.
func (o *WirelessLinkRequest) SetTags(v []NestedTagRequest) {
	o.Tags = v
}

// GetCustomFields returns the CustomFields field value if set, zero value otherwise.
func (o *WirelessLinkRequest) GetCustomFields() map[string]interface{} {
	if o == nil || IsNil(o.CustomFields) {
		var ret map[string]interface{}
		return ret
	}
	return o.CustomFields
}

// GetCustomFieldsOk returns a tuple with the CustomFields field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *WirelessLinkRequest) GetCustomFieldsOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.CustomFields) {
		return map[string]interface{}{}, false
	}
	return o.CustomFields, true
}

// HasCustomFields returns a boolean if a field has been set.
func (o *WirelessLinkRequest) HasCustomFields() bool {
	if o != nil && !IsNil(o.CustomFields) {
		return true
	}

	return false
}

// SetCustomFields gets a reference to the given map[string]interface{} and assigns it to the CustomFields field.
func (o *WirelessLinkRequest) SetCustomFields(v map[string]interface{}) {
	o.CustomFields = v
}

func (o WirelessLinkRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o WirelessLinkRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["interface_a"] = o.InterfaceA
	toSerialize["interface_b"] = o.InterfaceB
	if !IsNil(o.Ssid) {
		toSerialize["ssid"] = o.Ssid
	}
	if !IsNil(o.Status) {
		toSerialize["status"] = o.Status
	}
	if o.Tenant.IsSet() {
		toSerialize["tenant"] = o.Tenant.Get()
	}
	if !IsNil(o.AuthType) {
		toSerialize["auth_type"] = o.AuthType
	}
	if !IsNil(o.AuthCipher) {
		toSerialize["auth_cipher"] = o.AuthCipher
	}
	if !IsNil(o.AuthPsk) {
		toSerialize["auth_psk"] = o.AuthPsk
	}
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

func (o *WirelessLinkRequest) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"interface_a",
		"interface_b",
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

	varWirelessLinkRequest := _WirelessLinkRequest{}

	err = json.Unmarshal(data, &varWirelessLinkRequest)

	if err != nil {
		return err
	}

	*o = WirelessLinkRequest(varWirelessLinkRequest)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "interface_a")
		delete(additionalProperties, "interface_b")
		delete(additionalProperties, "ssid")
		delete(additionalProperties, "status")
		delete(additionalProperties, "tenant")
		delete(additionalProperties, "auth_type")
		delete(additionalProperties, "auth_cipher")
		delete(additionalProperties, "auth_psk")
		delete(additionalProperties, "description")
		delete(additionalProperties, "comments")
		delete(additionalProperties, "tags")
		delete(additionalProperties, "custom_fields")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableWirelessLinkRequest struct {
	value *WirelessLinkRequest
	isSet bool
}

func (v NullableWirelessLinkRequest) Get() *WirelessLinkRequest {
	return v.value
}

func (v *NullableWirelessLinkRequest) Set(val *WirelessLinkRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableWirelessLinkRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableWirelessLinkRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableWirelessLinkRequest(val *WirelessLinkRequest) *NullableWirelessLinkRequest {
	return &NullableWirelessLinkRequest{value: val, isSet: true}
}

func (v NullableWirelessLinkRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableWirelessLinkRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
