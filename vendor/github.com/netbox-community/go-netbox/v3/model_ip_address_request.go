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

// checks if the IPAddressRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &IPAddressRequest{}

// IPAddressRequest Adds support for custom fields and tags.
type IPAddressRequest struct {
	Address            string                         `json:"address"`
	Vrf                NullableNestedVRFRequest       `json:"vrf,omitempty"`
	Tenant             NullableNestedTenantRequest    `json:"tenant,omitempty"`
	Status             *IPAddressStatusValue          `json:"status,omitempty"`
	Role               *IPAddressRoleValue            `json:"role,omitempty"`
	AssignedObjectType NullableString                 `json:"assigned_object_type,omitempty"`
	AssignedObjectId   NullableInt64                  `json:"assigned_object_id,omitempty"`
	NatInside          NullableNestedIPAddressRequest `json:"nat_inside,omitempty"`
	// Hostname or FQDN (not case-sensitive)
	DnsName              *string                `json:"dns_name,omitempty"`
	Description          *string                `json:"description,omitempty"`
	Comments             *string                `json:"comments,omitempty"`
	Tags                 []NestedTagRequest     `json:"tags,omitempty"`
	CustomFields         map[string]interface{} `json:"custom_fields,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _IPAddressRequest IPAddressRequest

// NewIPAddressRequest instantiates a new IPAddressRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewIPAddressRequest(address string) *IPAddressRequest {
	this := IPAddressRequest{}
	this.Address = address
	return &this
}

// NewIPAddressRequestWithDefaults instantiates a new IPAddressRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewIPAddressRequestWithDefaults() *IPAddressRequest {
	this := IPAddressRequest{}
	return &this
}

// GetAddress returns the Address field value
func (o *IPAddressRequest) GetAddress() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Address
}

// GetAddressOk returns a tuple with the Address field value
// and a boolean to check if the value has been set.
func (o *IPAddressRequest) GetAddressOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Address, true
}

// SetAddress sets field value
func (o *IPAddressRequest) SetAddress(v string) {
	o.Address = v
}

// GetVrf returns the Vrf field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *IPAddressRequest) GetVrf() NestedVRFRequest {
	if o == nil || IsNil(o.Vrf.Get()) {
		var ret NestedVRFRequest
		return ret
	}
	return *o.Vrf.Get()
}

// GetVrfOk returns a tuple with the Vrf field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *IPAddressRequest) GetVrfOk() (*NestedVRFRequest, bool) {
	if o == nil {
		return nil, false
	}
	return o.Vrf.Get(), o.Vrf.IsSet()
}

// HasVrf returns a boolean if a field has been set.
func (o *IPAddressRequest) HasVrf() bool {
	if o != nil && o.Vrf.IsSet() {
		return true
	}

	return false
}

// SetVrf gets a reference to the given NullableNestedVRFRequest and assigns it to the Vrf field.
func (o *IPAddressRequest) SetVrf(v NestedVRFRequest) {
	o.Vrf.Set(&v)
}

// SetVrfNil sets the value for Vrf to be an explicit nil
func (o *IPAddressRequest) SetVrfNil() {
	o.Vrf.Set(nil)
}

// UnsetVrf ensures that no value is present for Vrf, not even an explicit nil
func (o *IPAddressRequest) UnsetVrf() {
	o.Vrf.Unset()
}

// GetTenant returns the Tenant field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *IPAddressRequest) GetTenant() NestedTenantRequest {
	if o == nil || IsNil(o.Tenant.Get()) {
		var ret NestedTenantRequest
		return ret
	}
	return *o.Tenant.Get()
}

// GetTenantOk returns a tuple with the Tenant field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *IPAddressRequest) GetTenantOk() (*NestedTenantRequest, bool) {
	if o == nil {
		return nil, false
	}
	return o.Tenant.Get(), o.Tenant.IsSet()
}

// HasTenant returns a boolean if a field has been set.
func (o *IPAddressRequest) HasTenant() bool {
	if o != nil && o.Tenant.IsSet() {
		return true
	}

	return false
}

// SetTenant gets a reference to the given NullableNestedTenantRequest and assigns it to the Tenant field.
func (o *IPAddressRequest) SetTenant(v NestedTenantRequest) {
	o.Tenant.Set(&v)
}

// SetTenantNil sets the value for Tenant to be an explicit nil
func (o *IPAddressRequest) SetTenantNil() {
	o.Tenant.Set(nil)
}

// UnsetTenant ensures that no value is present for Tenant, not even an explicit nil
func (o *IPAddressRequest) UnsetTenant() {
	o.Tenant.Unset()
}

// GetStatus returns the Status field value if set, zero value otherwise.
func (o *IPAddressRequest) GetStatus() IPAddressStatusValue {
	if o == nil || IsNil(o.Status) {
		var ret IPAddressStatusValue
		return ret
	}
	return *o.Status
}

// GetStatusOk returns a tuple with the Status field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IPAddressRequest) GetStatusOk() (*IPAddressStatusValue, bool) {
	if o == nil || IsNil(o.Status) {
		return nil, false
	}
	return o.Status, true
}

// HasStatus returns a boolean if a field has been set.
func (o *IPAddressRequest) HasStatus() bool {
	if o != nil && !IsNil(o.Status) {
		return true
	}

	return false
}

// SetStatus gets a reference to the given IPAddressStatusValue and assigns it to the Status field.
func (o *IPAddressRequest) SetStatus(v IPAddressStatusValue) {
	o.Status = &v
}

// GetRole returns the Role field value if set, zero value otherwise.
func (o *IPAddressRequest) GetRole() IPAddressRoleValue {
	if o == nil || IsNil(o.Role) {
		var ret IPAddressRoleValue
		return ret
	}
	return *o.Role
}

// GetRoleOk returns a tuple with the Role field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IPAddressRequest) GetRoleOk() (*IPAddressRoleValue, bool) {
	if o == nil || IsNil(o.Role) {
		return nil, false
	}
	return o.Role, true
}

// HasRole returns a boolean if a field has been set.
func (o *IPAddressRequest) HasRole() bool {
	if o != nil && !IsNil(o.Role) {
		return true
	}

	return false
}

// SetRole gets a reference to the given IPAddressRoleValue and assigns it to the Role field.
func (o *IPAddressRequest) SetRole(v IPAddressRoleValue) {
	o.Role = &v
}

// GetAssignedObjectType returns the AssignedObjectType field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *IPAddressRequest) GetAssignedObjectType() string {
	if o == nil || IsNil(o.AssignedObjectType.Get()) {
		var ret string
		return ret
	}
	return *o.AssignedObjectType.Get()
}

// GetAssignedObjectTypeOk returns a tuple with the AssignedObjectType field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *IPAddressRequest) GetAssignedObjectTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return o.AssignedObjectType.Get(), o.AssignedObjectType.IsSet()
}

// HasAssignedObjectType returns a boolean if a field has been set.
func (o *IPAddressRequest) HasAssignedObjectType() bool {
	if o != nil && o.AssignedObjectType.IsSet() {
		return true
	}

	return false
}

// SetAssignedObjectType gets a reference to the given NullableString and assigns it to the AssignedObjectType field.
func (o *IPAddressRequest) SetAssignedObjectType(v string) {
	o.AssignedObjectType.Set(&v)
}

// SetAssignedObjectTypeNil sets the value for AssignedObjectType to be an explicit nil
func (o *IPAddressRequest) SetAssignedObjectTypeNil() {
	o.AssignedObjectType.Set(nil)
}

// UnsetAssignedObjectType ensures that no value is present for AssignedObjectType, not even an explicit nil
func (o *IPAddressRequest) UnsetAssignedObjectType() {
	o.AssignedObjectType.Unset()
}

// GetAssignedObjectId returns the AssignedObjectId field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *IPAddressRequest) GetAssignedObjectId() int64 {
	if o == nil || IsNil(o.AssignedObjectId.Get()) {
		var ret int64
		return ret
	}
	return *o.AssignedObjectId.Get()
}

// GetAssignedObjectIdOk returns a tuple with the AssignedObjectId field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *IPAddressRequest) GetAssignedObjectIdOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return o.AssignedObjectId.Get(), o.AssignedObjectId.IsSet()
}

// HasAssignedObjectId returns a boolean if a field has been set.
func (o *IPAddressRequest) HasAssignedObjectId() bool {
	if o != nil && o.AssignedObjectId.IsSet() {
		return true
	}

	return false
}

// SetAssignedObjectId gets a reference to the given NullableInt64 and assigns it to the AssignedObjectId field.
func (o *IPAddressRequest) SetAssignedObjectId(v int64) {
	o.AssignedObjectId.Set(&v)
}

// SetAssignedObjectIdNil sets the value for AssignedObjectId to be an explicit nil
func (o *IPAddressRequest) SetAssignedObjectIdNil() {
	o.AssignedObjectId.Set(nil)
}

// UnsetAssignedObjectId ensures that no value is present for AssignedObjectId, not even an explicit nil
func (o *IPAddressRequest) UnsetAssignedObjectId() {
	o.AssignedObjectId.Unset()
}

// GetNatInside returns the NatInside field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *IPAddressRequest) GetNatInside() NestedIPAddressRequest {
	if o == nil || IsNil(o.NatInside.Get()) {
		var ret NestedIPAddressRequest
		return ret
	}
	return *o.NatInside.Get()
}

// GetNatInsideOk returns a tuple with the NatInside field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *IPAddressRequest) GetNatInsideOk() (*NestedIPAddressRequest, bool) {
	if o == nil {
		return nil, false
	}
	return o.NatInside.Get(), o.NatInside.IsSet()
}

// HasNatInside returns a boolean if a field has been set.
func (o *IPAddressRequest) HasNatInside() bool {
	if o != nil && o.NatInside.IsSet() {
		return true
	}

	return false
}

// SetNatInside gets a reference to the given NullableNestedIPAddressRequest and assigns it to the NatInside field.
func (o *IPAddressRequest) SetNatInside(v NestedIPAddressRequest) {
	o.NatInside.Set(&v)
}

// SetNatInsideNil sets the value for NatInside to be an explicit nil
func (o *IPAddressRequest) SetNatInsideNil() {
	o.NatInside.Set(nil)
}

// UnsetNatInside ensures that no value is present for NatInside, not even an explicit nil
func (o *IPAddressRequest) UnsetNatInside() {
	o.NatInside.Unset()
}

// GetDnsName returns the DnsName field value if set, zero value otherwise.
func (o *IPAddressRequest) GetDnsName() string {
	if o == nil || IsNil(o.DnsName) {
		var ret string
		return ret
	}
	return *o.DnsName
}

// GetDnsNameOk returns a tuple with the DnsName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IPAddressRequest) GetDnsNameOk() (*string, bool) {
	if o == nil || IsNil(o.DnsName) {
		return nil, false
	}
	return o.DnsName, true
}

// HasDnsName returns a boolean if a field has been set.
func (o *IPAddressRequest) HasDnsName() bool {
	if o != nil && !IsNil(o.DnsName) {
		return true
	}

	return false
}

// SetDnsName gets a reference to the given string and assigns it to the DnsName field.
func (o *IPAddressRequest) SetDnsName(v string) {
	o.DnsName = &v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *IPAddressRequest) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IPAddressRequest) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *IPAddressRequest) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *IPAddressRequest) SetDescription(v string) {
	o.Description = &v
}

// GetComments returns the Comments field value if set, zero value otherwise.
func (o *IPAddressRequest) GetComments() string {
	if o == nil || IsNil(o.Comments) {
		var ret string
		return ret
	}
	return *o.Comments
}

// GetCommentsOk returns a tuple with the Comments field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IPAddressRequest) GetCommentsOk() (*string, bool) {
	if o == nil || IsNil(o.Comments) {
		return nil, false
	}
	return o.Comments, true
}

// HasComments returns a boolean if a field has been set.
func (o *IPAddressRequest) HasComments() bool {
	if o != nil && !IsNil(o.Comments) {
		return true
	}

	return false
}

// SetComments gets a reference to the given string and assigns it to the Comments field.
func (o *IPAddressRequest) SetComments(v string) {
	o.Comments = &v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *IPAddressRequest) GetTags() []NestedTagRequest {
	if o == nil || IsNil(o.Tags) {
		var ret []NestedTagRequest
		return ret
	}
	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IPAddressRequest) GetTagsOk() ([]NestedTagRequest, bool) {
	if o == nil || IsNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *IPAddressRequest) HasTags() bool {
	if o != nil && !IsNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given []NestedTagRequest and assigns it to the Tags field.
func (o *IPAddressRequest) SetTags(v []NestedTagRequest) {
	o.Tags = v
}

// GetCustomFields returns the CustomFields field value if set, zero value otherwise.
func (o *IPAddressRequest) GetCustomFields() map[string]interface{} {
	if o == nil || IsNil(o.CustomFields) {
		var ret map[string]interface{}
		return ret
	}
	return o.CustomFields
}

// GetCustomFieldsOk returns a tuple with the CustomFields field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IPAddressRequest) GetCustomFieldsOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.CustomFields) {
		return map[string]interface{}{}, false
	}
	return o.CustomFields, true
}

// HasCustomFields returns a boolean if a field has been set.
func (o *IPAddressRequest) HasCustomFields() bool {
	if o != nil && !IsNil(o.CustomFields) {
		return true
	}

	return false
}

// SetCustomFields gets a reference to the given map[string]interface{} and assigns it to the CustomFields field.
func (o *IPAddressRequest) SetCustomFields(v map[string]interface{}) {
	o.CustomFields = v
}

func (o IPAddressRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o IPAddressRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["address"] = o.Address
	if o.Vrf.IsSet() {
		toSerialize["vrf"] = o.Vrf.Get()
	}
	if o.Tenant.IsSet() {
		toSerialize["tenant"] = o.Tenant.Get()
	}
	if !IsNil(o.Status) {
		toSerialize["status"] = o.Status
	}
	if !IsNil(o.Role) {
		toSerialize["role"] = o.Role
	}
	if o.AssignedObjectType.IsSet() {
		toSerialize["assigned_object_type"] = o.AssignedObjectType.Get()
	}
	if o.AssignedObjectId.IsSet() {
		toSerialize["assigned_object_id"] = o.AssignedObjectId.Get()
	}
	if o.NatInside.IsSet() {
		toSerialize["nat_inside"] = o.NatInside.Get()
	}
	if !IsNil(o.DnsName) {
		toSerialize["dns_name"] = o.DnsName
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

func (o *IPAddressRequest) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"address",
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

	varIPAddressRequest := _IPAddressRequest{}

	err = json.Unmarshal(data, &varIPAddressRequest)

	if err != nil {
		return err
	}

	*o = IPAddressRequest(varIPAddressRequest)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "address")
		delete(additionalProperties, "vrf")
		delete(additionalProperties, "tenant")
		delete(additionalProperties, "status")
		delete(additionalProperties, "role")
		delete(additionalProperties, "assigned_object_type")
		delete(additionalProperties, "assigned_object_id")
		delete(additionalProperties, "nat_inside")
		delete(additionalProperties, "dns_name")
		delete(additionalProperties, "description")
		delete(additionalProperties, "comments")
		delete(additionalProperties, "tags")
		delete(additionalProperties, "custom_fields")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableIPAddressRequest struct {
	value *IPAddressRequest
	isSet bool
}

func (v NullableIPAddressRequest) Get() *IPAddressRequest {
	return v.value
}

func (v *NullableIPAddressRequest) Set(val *IPAddressRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableIPAddressRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableIPAddressRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableIPAddressRequest(val *IPAddressRequest) *NullableIPAddressRequest {
	return &NullableIPAddressRequest{value: val, isSet: true}
}

func (v NullableIPAddressRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableIPAddressRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
