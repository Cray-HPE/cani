/*
NetBox REST API

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: 3.7.1 (3.7)
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package netbox

import (
	"encoding/json"
)

// checks if the PatchedWritableVRFRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PatchedWritableVRFRequest{}

// PatchedWritableVRFRequest Adds support for custom fields and tags.
type PatchedWritableVRFRequest struct {
	Name *string `json:"name,omitempty"`
	// Unique route distinguisher (as defined in RFC 4364)
	Rd     NullableString `json:"rd,omitempty"`
	Tenant NullableInt32  `json:"tenant,omitempty"`
	// Prevent duplicate prefixes/IP addresses within this VRF
	EnforceUnique        *bool                  `json:"enforce_unique,omitempty"`
	Description          *string                `json:"description,omitempty"`
	Comments             *string                `json:"comments,omitempty"`
	ImportTargets        []int32                `json:"import_targets,omitempty"`
	ExportTargets        []int32                `json:"export_targets,omitempty"`
	Tags                 []NestedTagRequest     `json:"tags,omitempty"`
	CustomFields         map[string]interface{} `json:"custom_fields,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _PatchedWritableVRFRequest PatchedWritableVRFRequest

// NewPatchedWritableVRFRequest instantiates a new PatchedWritableVRFRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPatchedWritableVRFRequest() *PatchedWritableVRFRequest {
	this := PatchedWritableVRFRequest{}
	return &this
}

// NewPatchedWritableVRFRequestWithDefaults instantiates a new PatchedWritableVRFRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPatchedWritableVRFRequestWithDefaults() *PatchedWritableVRFRequest {
	this := PatchedWritableVRFRequest{}
	return &this
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *PatchedWritableVRFRequest) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedWritableVRFRequest) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *PatchedWritableVRFRequest) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *PatchedWritableVRFRequest) SetName(v string) {
	o.Name = &v
}

// GetRd returns the Rd field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *PatchedWritableVRFRequest) GetRd() string {
	if o == nil || IsNil(o.Rd.Get()) {
		var ret string
		return ret
	}
	return *o.Rd.Get()
}

// GetRdOk returns a tuple with the Rd field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *PatchedWritableVRFRequest) GetRdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return o.Rd.Get(), o.Rd.IsSet()
}

// HasRd returns a boolean if a field has been set.
func (o *PatchedWritableVRFRequest) HasRd() bool {
	if o != nil && o.Rd.IsSet() {
		return true
	}

	return false
}

// SetRd gets a reference to the given NullableString and assigns it to the Rd field.
func (o *PatchedWritableVRFRequest) SetRd(v string) {
	o.Rd.Set(&v)
}

// SetRdNil sets the value for Rd to be an explicit nil
func (o *PatchedWritableVRFRequest) SetRdNil() {
	o.Rd.Set(nil)
}

// UnsetRd ensures that no value is present for Rd, not even an explicit nil
func (o *PatchedWritableVRFRequest) UnsetRd() {
	o.Rd.Unset()
}

// GetTenant returns the Tenant field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *PatchedWritableVRFRequest) GetTenant() int32 {
	if o == nil || IsNil(o.Tenant.Get()) {
		var ret int32
		return ret
	}
	return *o.Tenant.Get()
}

// GetTenantOk returns a tuple with the Tenant field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *PatchedWritableVRFRequest) GetTenantOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return o.Tenant.Get(), o.Tenant.IsSet()
}

// HasTenant returns a boolean if a field has been set.
func (o *PatchedWritableVRFRequest) HasTenant() bool {
	if o != nil && o.Tenant.IsSet() {
		return true
	}

	return false
}

// SetTenant gets a reference to the given NullableInt32 and assigns it to the Tenant field.
func (o *PatchedWritableVRFRequest) SetTenant(v int32) {
	o.Tenant.Set(&v)
}

// SetTenantNil sets the value for Tenant to be an explicit nil
func (o *PatchedWritableVRFRequest) SetTenantNil() {
	o.Tenant.Set(nil)
}

// UnsetTenant ensures that no value is present for Tenant, not even an explicit nil
func (o *PatchedWritableVRFRequest) UnsetTenant() {
	o.Tenant.Unset()
}

// GetEnforceUnique returns the EnforceUnique field value if set, zero value otherwise.
func (o *PatchedWritableVRFRequest) GetEnforceUnique() bool {
	if o == nil || IsNil(o.EnforceUnique) {
		var ret bool
		return ret
	}
	return *o.EnforceUnique
}

// GetEnforceUniqueOk returns a tuple with the EnforceUnique field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedWritableVRFRequest) GetEnforceUniqueOk() (*bool, bool) {
	if o == nil || IsNil(o.EnforceUnique) {
		return nil, false
	}
	return o.EnforceUnique, true
}

// HasEnforceUnique returns a boolean if a field has been set.
func (o *PatchedWritableVRFRequest) HasEnforceUnique() bool {
	if o != nil && !IsNil(o.EnforceUnique) {
		return true
	}

	return false
}

// SetEnforceUnique gets a reference to the given bool and assigns it to the EnforceUnique field.
func (o *PatchedWritableVRFRequest) SetEnforceUnique(v bool) {
	o.EnforceUnique = &v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *PatchedWritableVRFRequest) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedWritableVRFRequest) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *PatchedWritableVRFRequest) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *PatchedWritableVRFRequest) SetDescription(v string) {
	o.Description = &v
}

// GetComments returns the Comments field value if set, zero value otherwise.
func (o *PatchedWritableVRFRequest) GetComments() string {
	if o == nil || IsNil(o.Comments) {
		var ret string
		return ret
	}
	return *o.Comments
}

// GetCommentsOk returns a tuple with the Comments field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedWritableVRFRequest) GetCommentsOk() (*string, bool) {
	if o == nil || IsNil(o.Comments) {
		return nil, false
	}
	return o.Comments, true
}

// HasComments returns a boolean if a field has been set.
func (o *PatchedWritableVRFRequest) HasComments() bool {
	if o != nil && !IsNil(o.Comments) {
		return true
	}

	return false
}

// SetComments gets a reference to the given string and assigns it to the Comments field.
func (o *PatchedWritableVRFRequest) SetComments(v string) {
	o.Comments = &v
}

// GetImportTargets returns the ImportTargets field value if set, zero value otherwise.
func (o *PatchedWritableVRFRequest) GetImportTargets() []int32 {
	if o == nil || IsNil(o.ImportTargets) {
		var ret []int32
		return ret
	}
	return o.ImportTargets
}

// GetImportTargetsOk returns a tuple with the ImportTargets field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedWritableVRFRequest) GetImportTargetsOk() ([]int32, bool) {
	if o == nil || IsNil(o.ImportTargets) {
		return nil, false
	}
	return o.ImportTargets, true
}

// HasImportTargets returns a boolean if a field has been set.
func (o *PatchedWritableVRFRequest) HasImportTargets() bool {
	if o != nil && !IsNil(o.ImportTargets) {
		return true
	}

	return false
}

// SetImportTargets gets a reference to the given []int32 and assigns it to the ImportTargets field.
func (o *PatchedWritableVRFRequest) SetImportTargets(v []int32) {
	o.ImportTargets = v
}

// GetExportTargets returns the ExportTargets field value if set, zero value otherwise.
func (o *PatchedWritableVRFRequest) GetExportTargets() []int32 {
	if o == nil || IsNil(o.ExportTargets) {
		var ret []int32
		return ret
	}
	return o.ExportTargets
}

// GetExportTargetsOk returns a tuple with the ExportTargets field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedWritableVRFRequest) GetExportTargetsOk() ([]int32, bool) {
	if o == nil || IsNil(o.ExportTargets) {
		return nil, false
	}
	return o.ExportTargets, true
}

// HasExportTargets returns a boolean if a field has been set.
func (o *PatchedWritableVRFRequest) HasExportTargets() bool {
	if o != nil && !IsNil(o.ExportTargets) {
		return true
	}

	return false
}

// SetExportTargets gets a reference to the given []int32 and assigns it to the ExportTargets field.
func (o *PatchedWritableVRFRequest) SetExportTargets(v []int32) {
	o.ExportTargets = v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *PatchedWritableVRFRequest) GetTags() []NestedTagRequest {
	if o == nil || IsNil(o.Tags) {
		var ret []NestedTagRequest
		return ret
	}
	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedWritableVRFRequest) GetTagsOk() ([]NestedTagRequest, bool) {
	if o == nil || IsNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *PatchedWritableVRFRequest) HasTags() bool {
	if o != nil && !IsNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given []NestedTagRequest and assigns it to the Tags field.
func (o *PatchedWritableVRFRequest) SetTags(v []NestedTagRequest) {
	o.Tags = v
}

// GetCustomFields returns the CustomFields field value if set, zero value otherwise.
func (o *PatchedWritableVRFRequest) GetCustomFields() map[string]interface{} {
	if o == nil || IsNil(o.CustomFields) {
		var ret map[string]interface{}
		return ret
	}
	return o.CustomFields
}

// GetCustomFieldsOk returns a tuple with the CustomFields field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedWritableVRFRequest) GetCustomFieldsOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.CustomFields) {
		return map[string]interface{}{}, false
	}
	return o.CustomFields, true
}

// HasCustomFields returns a boolean if a field has been set.
func (o *PatchedWritableVRFRequest) HasCustomFields() bool {
	if o != nil && !IsNil(o.CustomFields) {
		return true
	}

	return false
}

// SetCustomFields gets a reference to the given map[string]interface{} and assigns it to the CustomFields field.
func (o *PatchedWritableVRFRequest) SetCustomFields(v map[string]interface{}) {
	o.CustomFields = v
}

func (o PatchedWritableVRFRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PatchedWritableVRFRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if o.Rd.IsSet() {
		toSerialize["rd"] = o.Rd.Get()
	}
	if o.Tenant.IsSet() {
		toSerialize["tenant"] = o.Tenant.Get()
	}
	if !IsNil(o.EnforceUnique) {
		toSerialize["enforce_unique"] = o.EnforceUnique
	}
	if !IsNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	if !IsNil(o.Comments) {
		toSerialize["comments"] = o.Comments
	}
	if !IsNil(o.ImportTargets) {
		toSerialize["import_targets"] = o.ImportTargets
	}
	if !IsNil(o.ExportTargets) {
		toSerialize["export_targets"] = o.ExportTargets
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

func (o *PatchedWritableVRFRequest) UnmarshalJSON(data []byte) (err error) {
	varPatchedWritableVRFRequest := _PatchedWritableVRFRequest{}

	err = json.Unmarshal(data, &varPatchedWritableVRFRequest)

	if err != nil {
		return err
	}

	*o = PatchedWritableVRFRequest(varPatchedWritableVRFRequest)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "name")
		delete(additionalProperties, "rd")
		delete(additionalProperties, "tenant")
		delete(additionalProperties, "enforce_unique")
		delete(additionalProperties, "description")
		delete(additionalProperties, "comments")
		delete(additionalProperties, "import_targets")
		delete(additionalProperties, "export_targets")
		delete(additionalProperties, "tags")
		delete(additionalProperties, "custom_fields")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullablePatchedWritableVRFRequest struct {
	value *PatchedWritableVRFRequest
	isSet bool
}

func (v NullablePatchedWritableVRFRequest) Get() *PatchedWritableVRFRequest {
	return v.value
}

func (v *NullablePatchedWritableVRFRequest) Set(val *PatchedWritableVRFRequest) {
	v.value = val
	v.isSet = true
}

func (v NullablePatchedWritableVRFRequest) IsSet() bool {
	return v.isSet
}

func (v *NullablePatchedWritableVRFRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePatchedWritableVRFRequest(val *PatchedWritableVRFRequest) *NullablePatchedWritableVRFRequest {
	return &NullablePatchedWritableVRFRequest{value: val, isSet: true}
}

func (v NullablePatchedWritableVRFRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePatchedWritableVRFRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
