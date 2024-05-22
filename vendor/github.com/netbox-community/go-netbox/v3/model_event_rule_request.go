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

// checks if the EventRuleRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &EventRuleRequest{}

// EventRuleRequest Adds support for custom fields and tags.
type EventRuleRequest struct {
	ContentTypes []string `json:"content_types"`
	Name         string   `json:"name"`
	// Triggers when a matching object is created.
	TypeCreate *bool `json:"type_create,omitempty"`
	// Triggers when a matching object is updated.
	TypeUpdate *bool `json:"type_update,omitempty"`
	// Triggers when a matching object is deleted.
	TypeDelete *bool `json:"type_delete,omitempty"`
	// Triggers when a job for a matching object is started.
	TypeJobStart *bool `json:"type_job_start,omitempty"`
	// Triggers when a job for a matching object terminates.
	TypeJobEnd *bool `json:"type_job_end,omitempty"`
	Enabled    *bool `json:"enabled,omitempty"`
	// A set of conditions which determine whether the event will be generated.
	Conditions           interface{}              `json:"conditions,omitempty"`
	ActionType           EventRuleActionTypeValue `json:"action_type"`
	ActionObjectType     string                   `json:"action_object_type"`
	ActionObjectId       NullableInt64            `json:"action_object_id,omitempty"`
	Description          *string                  `json:"description,omitempty"`
	CustomFields         map[string]interface{}   `json:"custom_fields,omitempty"`
	Tags                 []NestedTagRequest       `json:"tags,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _EventRuleRequest EventRuleRequest

// NewEventRuleRequest instantiates a new EventRuleRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewEventRuleRequest(contentTypes []string, name string, actionType EventRuleActionTypeValue, actionObjectType string) *EventRuleRequest {
	this := EventRuleRequest{}
	this.ContentTypes = contentTypes
	this.Name = name
	this.ActionType = actionType
	this.ActionObjectType = actionObjectType
	return &this
}

// NewEventRuleRequestWithDefaults instantiates a new EventRuleRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewEventRuleRequestWithDefaults() *EventRuleRequest {
	this := EventRuleRequest{}
	return &this
}

// GetContentTypes returns the ContentTypes field value
func (o *EventRuleRequest) GetContentTypes() []string {
	if o == nil {
		var ret []string
		return ret
	}

	return o.ContentTypes
}

// GetContentTypesOk returns a tuple with the ContentTypes field value
// and a boolean to check if the value has been set.
func (o *EventRuleRequest) GetContentTypesOk() ([]string, bool) {
	if o == nil {
		return nil, false
	}
	return o.ContentTypes, true
}

// SetContentTypes sets field value
func (o *EventRuleRequest) SetContentTypes(v []string) {
	o.ContentTypes = v
}

// GetName returns the Name field value
func (o *EventRuleRequest) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *EventRuleRequest) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *EventRuleRequest) SetName(v string) {
	o.Name = v
}

// GetTypeCreate returns the TypeCreate field value if set, zero value otherwise.
func (o *EventRuleRequest) GetTypeCreate() bool {
	if o == nil || IsNil(o.TypeCreate) {
		var ret bool
		return ret
	}
	return *o.TypeCreate
}

// GetTypeCreateOk returns a tuple with the TypeCreate field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EventRuleRequest) GetTypeCreateOk() (*bool, bool) {
	if o == nil || IsNil(o.TypeCreate) {
		return nil, false
	}
	return o.TypeCreate, true
}

// HasTypeCreate returns a boolean if a field has been set.
func (o *EventRuleRequest) HasTypeCreate() bool {
	if o != nil && !IsNil(o.TypeCreate) {
		return true
	}

	return false
}

// SetTypeCreate gets a reference to the given bool and assigns it to the TypeCreate field.
func (o *EventRuleRequest) SetTypeCreate(v bool) {
	o.TypeCreate = &v
}

// GetTypeUpdate returns the TypeUpdate field value if set, zero value otherwise.
func (o *EventRuleRequest) GetTypeUpdate() bool {
	if o == nil || IsNil(o.TypeUpdate) {
		var ret bool
		return ret
	}
	return *o.TypeUpdate
}

// GetTypeUpdateOk returns a tuple with the TypeUpdate field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EventRuleRequest) GetTypeUpdateOk() (*bool, bool) {
	if o == nil || IsNil(o.TypeUpdate) {
		return nil, false
	}
	return o.TypeUpdate, true
}

// HasTypeUpdate returns a boolean if a field has been set.
func (o *EventRuleRequest) HasTypeUpdate() bool {
	if o != nil && !IsNil(o.TypeUpdate) {
		return true
	}

	return false
}

// SetTypeUpdate gets a reference to the given bool and assigns it to the TypeUpdate field.
func (o *EventRuleRequest) SetTypeUpdate(v bool) {
	o.TypeUpdate = &v
}

// GetTypeDelete returns the TypeDelete field value if set, zero value otherwise.
func (o *EventRuleRequest) GetTypeDelete() bool {
	if o == nil || IsNil(o.TypeDelete) {
		var ret bool
		return ret
	}
	return *o.TypeDelete
}

// GetTypeDeleteOk returns a tuple with the TypeDelete field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EventRuleRequest) GetTypeDeleteOk() (*bool, bool) {
	if o == nil || IsNil(o.TypeDelete) {
		return nil, false
	}
	return o.TypeDelete, true
}

// HasTypeDelete returns a boolean if a field has been set.
func (o *EventRuleRequest) HasTypeDelete() bool {
	if o != nil && !IsNil(o.TypeDelete) {
		return true
	}

	return false
}

// SetTypeDelete gets a reference to the given bool and assigns it to the TypeDelete field.
func (o *EventRuleRequest) SetTypeDelete(v bool) {
	o.TypeDelete = &v
}

// GetTypeJobStart returns the TypeJobStart field value if set, zero value otherwise.
func (o *EventRuleRequest) GetTypeJobStart() bool {
	if o == nil || IsNil(o.TypeJobStart) {
		var ret bool
		return ret
	}
	return *o.TypeJobStart
}

// GetTypeJobStartOk returns a tuple with the TypeJobStart field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EventRuleRequest) GetTypeJobStartOk() (*bool, bool) {
	if o == nil || IsNil(o.TypeJobStart) {
		return nil, false
	}
	return o.TypeJobStart, true
}

// HasTypeJobStart returns a boolean if a field has been set.
func (o *EventRuleRequest) HasTypeJobStart() bool {
	if o != nil && !IsNil(o.TypeJobStart) {
		return true
	}

	return false
}

// SetTypeJobStart gets a reference to the given bool and assigns it to the TypeJobStart field.
func (o *EventRuleRequest) SetTypeJobStart(v bool) {
	o.TypeJobStart = &v
}

// GetTypeJobEnd returns the TypeJobEnd field value if set, zero value otherwise.
func (o *EventRuleRequest) GetTypeJobEnd() bool {
	if o == nil || IsNil(o.TypeJobEnd) {
		var ret bool
		return ret
	}
	return *o.TypeJobEnd
}

// GetTypeJobEndOk returns a tuple with the TypeJobEnd field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EventRuleRequest) GetTypeJobEndOk() (*bool, bool) {
	if o == nil || IsNil(o.TypeJobEnd) {
		return nil, false
	}
	return o.TypeJobEnd, true
}

// HasTypeJobEnd returns a boolean if a field has been set.
func (o *EventRuleRequest) HasTypeJobEnd() bool {
	if o != nil && !IsNil(o.TypeJobEnd) {
		return true
	}

	return false
}

// SetTypeJobEnd gets a reference to the given bool and assigns it to the TypeJobEnd field.
func (o *EventRuleRequest) SetTypeJobEnd(v bool) {
	o.TypeJobEnd = &v
}

// GetEnabled returns the Enabled field value if set, zero value otherwise.
func (o *EventRuleRequest) GetEnabled() bool {
	if o == nil || IsNil(o.Enabled) {
		var ret bool
		return ret
	}
	return *o.Enabled
}

// GetEnabledOk returns a tuple with the Enabled field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EventRuleRequest) GetEnabledOk() (*bool, bool) {
	if o == nil || IsNil(o.Enabled) {
		return nil, false
	}
	return o.Enabled, true
}

// HasEnabled returns a boolean if a field has been set.
func (o *EventRuleRequest) HasEnabled() bool {
	if o != nil && !IsNil(o.Enabled) {
		return true
	}

	return false
}

// SetEnabled gets a reference to the given bool and assigns it to the Enabled field.
func (o *EventRuleRequest) SetEnabled(v bool) {
	o.Enabled = &v
}

// GetConditions returns the Conditions field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *EventRuleRequest) GetConditions() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.Conditions
}

// GetConditionsOk returns a tuple with the Conditions field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *EventRuleRequest) GetConditionsOk() (*interface{}, bool) {
	if o == nil || IsNil(o.Conditions) {
		return nil, false
	}
	return &o.Conditions, true
}

// HasConditions returns a boolean if a field has been set.
func (o *EventRuleRequest) HasConditions() bool {
	if o != nil && IsNil(o.Conditions) {
		return true
	}

	return false
}

// SetConditions gets a reference to the given interface{} and assigns it to the Conditions field.
func (o *EventRuleRequest) SetConditions(v interface{}) {
	o.Conditions = v
}

// GetActionType returns the ActionType field value
func (o *EventRuleRequest) GetActionType() EventRuleActionTypeValue {
	if o == nil {
		var ret EventRuleActionTypeValue
		return ret
	}

	return o.ActionType
}

// GetActionTypeOk returns a tuple with the ActionType field value
// and a boolean to check if the value has been set.
func (o *EventRuleRequest) GetActionTypeOk() (*EventRuleActionTypeValue, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ActionType, true
}

// SetActionType sets field value
func (o *EventRuleRequest) SetActionType(v EventRuleActionTypeValue) {
	o.ActionType = v
}

// GetActionObjectType returns the ActionObjectType field value
func (o *EventRuleRequest) GetActionObjectType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ActionObjectType
}

// GetActionObjectTypeOk returns a tuple with the ActionObjectType field value
// and a boolean to check if the value has been set.
func (o *EventRuleRequest) GetActionObjectTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ActionObjectType, true
}

// SetActionObjectType sets field value
func (o *EventRuleRequest) SetActionObjectType(v string) {
	o.ActionObjectType = v
}

// GetActionObjectId returns the ActionObjectId field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *EventRuleRequest) GetActionObjectId() int64 {
	if o == nil || IsNil(o.ActionObjectId.Get()) {
		var ret int64
		return ret
	}
	return *o.ActionObjectId.Get()
}

// GetActionObjectIdOk returns a tuple with the ActionObjectId field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *EventRuleRequest) GetActionObjectIdOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return o.ActionObjectId.Get(), o.ActionObjectId.IsSet()
}

// HasActionObjectId returns a boolean if a field has been set.
func (o *EventRuleRequest) HasActionObjectId() bool {
	if o != nil && o.ActionObjectId.IsSet() {
		return true
	}

	return false
}

// SetActionObjectId gets a reference to the given NullableInt64 and assigns it to the ActionObjectId field.
func (o *EventRuleRequest) SetActionObjectId(v int64) {
	o.ActionObjectId.Set(&v)
}

// SetActionObjectIdNil sets the value for ActionObjectId to be an explicit nil
func (o *EventRuleRequest) SetActionObjectIdNil() {
	o.ActionObjectId.Set(nil)
}

// UnsetActionObjectId ensures that no value is present for ActionObjectId, not even an explicit nil
func (o *EventRuleRequest) UnsetActionObjectId() {
	o.ActionObjectId.Unset()
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *EventRuleRequest) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EventRuleRequest) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *EventRuleRequest) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *EventRuleRequest) SetDescription(v string) {
	o.Description = &v
}

// GetCustomFields returns the CustomFields field value if set, zero value otherwise.
func (o *EventRuleRequest) GetCustomFields() map[string]interface{} {
	if o == nil || IsNil(o.CustomFields) {
		var ret map[string]interface{}
		return ret
	}
	return o.CustomFields
}

// GetCustomFieldsOk returns a tuple with the CustomFields field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EventRuleRequest) GetCustomFieldsOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.CustomFields) {
		return map[string]interface{}{}, false
	}
	return o.CustomFields, true
}

// HasCustomFields returns a boolean if a field has been set.
func (o *EventRuleRequest) HasCustomFields() bool {
	if o != nil && !IsNil(o.CustomFields) {
		return true
	}

	return false
}

// SetCustomFields gets a reference to the given map[string]interface{} and assigns it to the CustomFields field.
func (o *EventRuleRequest) SetCustomFields(v map[string]interface{}) {
	o.CustomFields = v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *EventRuleRequest) GetTags() []NestedTagRequest {
	if o == nil || IsNil(o.Tags) {
		var ret []NestedTagRequest
		return ret
	}
	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EventRuleRequest) GetTagsOk() ([]NestedTagRequest, bool) {
	if o == nil || IsNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *EventRuleRequest) HasTags() bool {
	if o != nil && !IsNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given []NestedTagRequest and assigns it to the Tags field.
func (o *EventRuleRequest) SetTags(v []NestedTagRequest) {
	o.Tags = v
}

func (o EventRuleRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o EventRuleRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["content_types"] = o.ContentTypes
	toSerialize["name"] = o.Name
	if !IsNil(o.TypeCreate) {
		toSerialize["type_create"] = o.TypeCreate
	}
	if !IsNil(o.TypeUpdate) {
		toSerialize["type_update"] = o.TypeUpdate
	}
	if !IsNil(o.TypeDelete) {
		toSerialize["type_delete"] = o.TypeDelete
	}
	if !IsNil(o.TypeJobStart) {
		toSerialize["type_job_start"] = o.TypeJobStart
	}
	if !IsNil(o.TypeJobEnd) {
		toSerialize["type_job_end"] = o.TypeJobEnd
	}
	if !IsNil(o.Enabled) {
		toSerialize["enabled"] = o.Enabled
	}
	if o.Conditions != nil {
		toSerialize["conditions"] = o.Conditions
	}
	toSerialize["action_type"] = o.ActionType
	toSerialize["action_object_type"] = o.ActionObjectType
	if o.ActionObjectId.IsSet() {
		toSerialize["action_object_id"] = o.ActionObjectId.Get()
	}
	if !IsNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	if !IsNil(o.CustomFields) {
		toSerialize["custom_fields"] = o.CustomFields
	}
	if !IsNil(o.Tags) {
		toSerialize["tags"] = o.Tags
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *EventRuleRequest) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"content_types",
		"name",
		"action_type",
		"action_object_type",
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

	varEventRuleRequest := _EventRuleRequest{}

	err = json.Unmarshal(data, &varEventRuleRequest)

	if err != nil {
		return err
	}

	*o = EventRuleRequest(varEventRuleRequest)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "content_types")
		delete(additionalProperties, "name")
		delete(additionalProperties, "type_create")
		delete(additionalProperties, "type_update")
		delete(additionalProperties, "type_delete")
		delete(additionalProperties, "type_job_start")
		delete(additionalProperties, "type_job_end")
		delete(additionalProperties, "enabled")
		delete(additionalProperties, "conditions")
		delete(additionalProperties, "action_type")
		delete(additionalProperties, "action_object_type")
		delete(additionalProperties, "action_object_id")
		delete(additionalProperties, "description")
		delete(additionalProperties, "custom_fields")
		delete(additionalProperties, "tags")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableEventRuleRequest struct {
	value *EventRuleRequest
	isSet bool
}

func (v NullableEventRuleRequest) Get() *EventRuleRequest {
	return v.value
}

func (v *NullableEventRuleRequest) Set(val *EventRuleRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableEventRuleRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableEventRuleRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEventRuleRequest(val *EventRuleRequest) *NullableEventRuleRequest {
	return &NullableEventRuleRequest{value: val, isSet: true}
}

func (v NullableEventRuleRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEventRuleRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
