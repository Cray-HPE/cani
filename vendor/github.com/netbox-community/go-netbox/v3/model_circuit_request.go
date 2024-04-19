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

// checks if the CircuitRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CircuitRequest{}

// CircuitRequest Adds support for custom fields and tags.
type CircuitRequest struct {
	// Unique circuit ID
	Cid             string                               `json:"cid"`
	Provider        NestedProviderRequest                `json:"provider"`
	ProviderAccount NullableNestedProviderAccountRequest `json:"provider_account,omitempty"`
	Type            NestedCircuitTypeRequest             `json:"type"`
	Status          *CircuitStatusValue                  `json:"status,omitempty"`
	Tenant          NullableNestedTenantRequest          `json:"tenant,omitempty"`
	InstallDate     NullableString                       `json:"install_date,omitempty"`
	TerminationDate NullableString                       `json:"termination_date,omitempty"`
	// Committed rate
	CommitRate           NullableInt32          `json:"commit_rate,omitempty"`
	Description          *string                `json:"description,omitempty"`
	Comments             *string                `json:"comments,omitempty"`
	Tags                 []NestedTagRequest     `json:"tags,omitempty"`
	CustomFields         map[string]interface{} `json:"custom_fields,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _CircuitRequest CircuitRequest

// NewCircuitRequest instantiates a new CircuitRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCircuitRequest(cid string, provider NestedProviderRequest, type_ NestedCircuitTypeRequest) *CircuitRequest {
	this := CircuitRequest{}
	this.Cid = cid
	this.Provider = provider
	this.Type = type_
	return &this
}

// NewCircuitRequestWithDefaults instantiates a new CircuitRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCircuitRequestWithDefaults() *CircuitRequest {
	this := CircuitRequest{}
	return &this
}

// GetCid returns the Cid field value
func (o *CircuitRequest) GetCid() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Cid
}

// GetCidOk returns a tuple with the Cid field value
// and a boolean to check if the value has been set.
func (o *CircuitRequest) GetCidOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Cid, true
}

// SetCid sets field value
func (o *CircuitRequest) SetCid(v string) {
	o.Cid = v
}

// GetProvider returns the Provider field value
func (o *CircuitRequest) GetProvider() NestedProviderRequest {
	if o == nil {
		var ret NestedProviderRequest
		return ret
	}

	return o.Provider
}

// GetProviderOk returns a tuple with the Provider field value
// and a boolean to check if the value has been set.
func (o *CircuitRequest) GetProviderOk() (*NestedProviderRequest, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Provider, true
}

// SetProvider sets field value
func (o *CircuitRequest) SetProvider(v NestedProviderRequest) {
	o.Provider = v
}

// GetProviderAccount returns the ProviderAccount field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *CircuitRequest) GetProviderAccount() NestedProviderAccountRequest {
	if o == nil || IsNil(o.ProviderAccount.Get()) {
		var ret NestedProviderAccountRequest
		return ret
	}
	return *o.ProviderAccount.Get()
}

// GetProviderAccountOk returns a tuple with the ProviderAccount field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *CircuitRequest) GetProviderAccountOk() (*NestedProviderAccountRequest, bool) {
	if o == nil {
		return nil, false
	}
	return o.ProviderAccount.Get(), o.ProviderAccount.IsSet()
}

// HasProviderAccount returns a boolean if a field has been set.
func (o *CircuitRequest) HasProviderAccount() bool {
	if o != nil && o.ProviderAccount.IsSet() {
		return true
	}

	return false
}

// SetProviderAccount gets a reference to the given NullableNestedProviderAccountRequest and assigns it to the ProviderAccount field.
func (o *CircuitRequest) SetProviderAccount(v NestedProviderAccountRequest) {
	o.ProviderAccount.Set(&v)
}

// SetProviderAccountNil sets the value for ProviderAccount to be an explicit nil
func (o *CircuitRequest) SetProviderAccountNil() {
	o.ProviderAccount.Set(nil)
}

// UnsetProviderAccount ensures that no value is present for ProviderAccount, not even an explicit nil
func (o *CircuitRequest) UnsetProviderAccount() {
	o.ProviderAccount.Unset()
}

// GetType returns the Type field value
func (o *CircuitRequest) GetType() NestedCircuitTypeRequest {
	if o == nil {
		var ret NestedCircuitTypeRequest
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *CircuitRequest) GetTypeOk() (*NestedCircuitTypeRequest, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Type, true
}

// SetType sets field value
func (o *CircuitRequest) SetType(v NestedCircuitTypeRequest) {
	o.Type = v
}

// GetStatus returns the Status field value if set, zero value otherwise.
func (o *CircuitRequest) GetStatus() CircuitStatusValue {
	if o == nil || IsNil(o.Status) {
		var ret CircuitStatusValue
		return ret
	}
	return *o.Status
}

// GetStatusOk returns a tuple with the Status field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CircuitRequest) GetStatusOk() (*CircuitStatusValue, bool) {
	if o == nil || IsNil(o.Status) {
		return nil, false
	}
	return o.Status, true
}

// HasStatus returns a boolean if a field has been set.
func (o *CircuitRequest) HasStatus() bool {
	if o != nil && !IsNil(o.Status) {
		return true
	}

	return false
}

// SetStatus gets a reference to the given CircuitStatusValue and assigns it to the Status field.
func (o *CircuitRequest) SetStatus(v CircuitStatusValue) {
	o.Status = &v
}

// GetTenant returns the Tenant field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *CircuitRequest) GetTenant() NestedTenantRequest {
	if o == nil || IsNil(o.Tenant.Get()) {
		var ret NestedTenantRequest
		return ret
	}
	return *o.Tenant.Get()
}

// GetTenantOk returns a tuple with the Tenant field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *CircuitRequest) GetTenantOk() (*NestedTenantRequest, bool) {
	if o == nil {
		return nil, false
	}
	return o.Tenant.Get(), o.Tenant.IsSet()
}

// HasTenant returns a boolean if a field has been set.
func (o *CircuitRequest) HasTenant() bool {
	if o != nil && o.Tenant.IsSet() {
		return true
	}

	return false
}

// SetTenant gets a reference to the given NullableNestedTenantRequest and assigns it to the Tenant field.
func (o *CircuitRequest) SetTenant(v NestedTenantRequest) {
	o.Tenant.Set(&v)
}

// SetTenantNil sets the value for Tenant to be an explicit nil
func (o *CircuitRequest) SetTenantNil() {
	o.Tenant.Set(nil)
}

// UnsetTenant ensures that no value is present for Tenant, not even an explicit nil
func (o *CircuitRequest) UnsetTenant() {
	o.Tenant.Unset()
}

// GetInstallDate returns the InstallDate field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *CircuitRequest) GetInstallDate() string {
	if o == nil || IsNil(o.InstallDate.Get()) {
		var ret string
		return ret
	}
	return *o.InstallDate.Get()
}

// GetInstallDateOk returns a tuple with the InstallDate field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *CircuitRequest) GetInstallDateOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return o.InstallDate.Get(), o.InstallDate.IsSet()
}

// HasInstallDate returns a boolean if a field has been set.
func (o *CircuitRequest) HasInstallDate() bool {
	if o != nil && o.InstallDate.IsSet() {
		return true
	}

	return false
}

// SetInstallDate gets a reference to the given NullableString and assigns it to the InstallDate field.
func (o *CircuitRequest) SetInstallDate(v string) {
	o.InstallDate.Set(&v)
}

// SetInstallDateNil sets the value for InstallDate to be an explicit nil
func (o *CircuitRequest) SetInstallDateNil() {
	o.InstallDate.Set(nil)
}

// UnsetInstallDate ensures that no value is present for InstallDate, not even an explicit nil
func (o *CircuitRequest) UnsetInstallDate() {
	o.InstallDate.Unset()
}

// GetTerminationDate returns the TerminationDate field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *CircuitRequest) GetTerminationDate() string {
	if o == nil || IsNil(o.TerminationDate.Get()) {
		var ret string
		return ret
	}
	return *o.TerminationDate.Get()
}

// GetTerminationDateOk returns a tuple with the TerminationDate field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *CircuitRequest) GetTerminationDateOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return o.TerminationDate.Get(), o.TerminationDate.IsSet()
}

// HasTerminationDate returns a boolean if a field has been set.
func (o *CircuitRequest) HasTerminationDate() bool {
	if o != nil && o.TerminationDate.IsSet() {
		return true
	}

	return false
}

// SetTerminationDate gets a reference to the given NullableString and assigns it to the TerminationDate field.
func (o *CircuitRequest) SetTerminationDate(v string) {
	o.TerminationDate.Set(&v)
}

// SetTerminationDateNil sets the value for TerminationDate to be an explicit nil
func (o *CircuitRequest) SetTerminationDateNil() {
	o.TerminationDate.Set(nil)
}

// UnsetTerminationDate ensures that no value is present for TerminationDate, not even an explicit nil
func (o *CircuitRequest) UnsetTerminationDate() {
	o.TerminationDate.Unset()
}

// GetCommitRate returns the CommitRate field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *CircuitRequest) GetCommitRate() int32 {
	if o == nil || IsNil(o.CommitRate.Get()) {
		var ret int32
		return ret
	}
	return *o.CommitRate.Get()
}

// GetCommitRateOk returns a tuple with the CommitRate field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *CircuitRequest) GetCommitRateOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return o.CommitRate.Get(), o.CommitRate.IsSet()
}

// HasCommitRate returns a boolean if a field has been set.
func (o *CircuitRequest) HasCommitRate() bool {
	if o != nil && o.CommitRate.IsSet() {
		return true
	}

	return false
}

// SetCommitRate gets a reference to the given NullableInt32 and assigns it to the CommitRate field.
func (o *CircuitRequest) SetCommitRate(v int32) {
	o.CommitRate.Set(&v)
}

// SetCommitRateNil sets the value for CommitRate to be an explicit nil
func (o *CircuitRequest) SetCommitRateNil() {
	o.CommitRate.Set(nil)
}

// UnsetCommitRate ensures that no value is present for CommitRate, not even an explicit nil
func (o *CircuitRequest) UnsetCommitRate() {
	o.CommitRate.Unset()
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *CircuitRequest) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CircuitRequest) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *CircuitRequest) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *CircuitRequest) SetDescription(v string) {
	o.Description = &v
}

// GetComments returns the Comments field value if set, zero value otherwise.
func (o *CircuitRequest) GetComments() string {
	if o == nil || IsNil(o.Comments) {
		var ret string
		return ret
	}
	return *o.Comments
}

// GetCommentsOk returns a tuple with the Comments field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CircuitRequest) GetCommentsOk() (*string, bool) {
	if o == nil || IsNil(o.Comments) {
		return nil, false
	}
	return o.Comments, true
}

// HasComments returns a boolean if a field has been set.
func (o *CircuitRequest) HasComments() bool {
	if o != nil && !IsNil(o.Comments) {
		return true
	}

	return false
}

// SetComments gets a reference to the given string and assigns it to the Comments field.
func (o *CircuitRequest) SetComments(v string) {
	o.Comments = &v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *CircuitRequest) GetTags() []NestedTagRequest {
	if o == nil || IsNil(o.Tags) {
		var ret []NestedTagRequest
		return ret
	}
	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CircuitRequest) GetTagsOk() ([]NestedTagRequest, bool) {
	if o == nil || IsNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *CircuitRequest) HasTags() bool {
	if o != nil && !IsNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given []NestedTagRequest and assigns it to the Tags field.
func (o *CircuitRequest) SetTags(v []NestedTagRequest) {
	o.Tags = v
}

// GetCustomFields returns the CustomFields field value if set, zero value otherwise.
func (o *CircuitRequest) GetCustomFields() map[string]interface{} {
	if o == nil || IsNil(o.CustomFields) {
		var ret map[string]interface{}
		return ret
	}
	return o.CustomFields
}

// GetCustomFieldsOk returns a tuple with the CustomFields field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CircuitRequest) GetCustomFieldsOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.CustomFields) {
		return map[string]interface{}{}, false
	}
	return o.CustomFields, true
}

// HasCustomFields returns a boolean if a field has been set.
func (o *CircuitRequest) HasCustomFields() bool {
	if o != nil && !IsNil(o.CustomFields) {
		return true
	}

	return false
}

// SetCustomFields gets a reference to the given map[string]interface{} and assigns it to the CustomFields field.
func (o *CircuitRequest) SetCustomFields(v map[string]interface{}) {
	o.CustomFields = v
}

func (o CircuitRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CircuitRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["cid"] = o.Cid
	toSerialize["provider"] = o.Provider
	if o.ProviderAccount.IsSet() {
		toSerialize["provider_account"] = o.ProviderAccount.Get()
	}
	toSerialize["type"] = o.Type
	if !IsNil(o.Status) {
		toSerialize["status"] = o.Status
	}
	if o.Tenant.IsSet() {
		toSerialize["tenant"] = o.Tenant.Get()
	}
	if o.InstallDate.IsSet() {
		toSerialize["install_date"] = o.InstallDate.Get()
	}
	if o.TerminationDate.IsSet() {
		toSerialize["termination_date"] = o.TerminationDate.Get()
	}
	if o.CommitRate.IsSet() {
		toSerialize["commit_rate"] = o.CommitRate.Get()
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

func (o *CircuitRequest) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"cid",
		"provider",
		"type",
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

	varCircuitRequest := _CircuitRequest{}

	err = json.Unmarshal(data, &varCircuitRequest)

	if err != nil {
		return err
	}

	*o = CircuitRequest(varCircuitRequest)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "cid")
		delete(additionalProperties, "provider")
		delete(additionalProperties, "provider_account")
		delete(additionalProperties, "type")
		delete(additionalProperties, "status")
		delete(additionalProperties, "tenant")
		delete(additionalProperties, "install_date")
		delete(additionalProperties, "termination_date")
		delete(additionalProperties, "commit_rate")
		delete(additionalProperties, "description")
		delete(additionalProperties, "comments")
		delete(additionalProperties, "tags")
		delete(additionalProperties, "custom_fields")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableCircuitRequest struct {
	value *CircuitRequest
	isSet bool
}

func (v NullableCircuitRequest) Get() *CircuitRequest {
	return v.value
}

func (v *NullableCircuitRequest) Set(val *CircuitRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableCircuitRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableCircuitRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCircuitRequest(val *CircuitRequest) *NullableCircuitRequest {
	return &NullableCircuitRequest{value: val, isSet: true}
}

func (v NullableCircuitRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCircuitRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
