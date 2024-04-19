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

// checks if the PatchedWritableBookmarkRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PatchedWritableBookmarkRequest{}

// PatchedWritableBookmarkRequest Extends the built-in ModelSerializer to enforce calling full_clean() on a copy of the associated instance during validation. (DRF does not do this by default; see https://github.com/encode/django-rest-framework/issues/3144)
type PatchedWritableBookmarkRequest struct {
	ObjectType           *string `json:"object_type,omitempty"`
	ObjectId             *int64  `json:"object_id,omitempty"`
	User                 *int32  `json:"user,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _PatchedWritableBookmarkRequest PatchedWritableBookmarkRequest

// NewPatchedWritableBookmarkRequest instantiates a new PatchedWritableBookmarkRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPatchedWritableBookmarkRequest() *PatchedWritableBookmarkRequest {
	this := PatchedWritableBookmarkRequest{}
	return &this
}

// NewPatchedWritableBookmarkRequestWithDefaults instantiates a new PatchedWritableBookmarkRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPatchedWritableBookmarkRequestWithDefaults() *PatchedWritableBookmarkRequest {
	this := PatchedWritableBookmarkRequest{}
	return &this
}

// GetObjectType returns the ObjectType field value if set, zero value otherwise.
func (o *PatchedWritableBookmarkRequest) GetObjectType() string {
	if o == nil || IsNil(o.ObjectType) {
		var ret string
		return ret
	}
	return *o.ObjectType
}

// GetObjectTypeOk returns a tuple with the ObjectType field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedWritableBookmarkRequest) GetObjectTypeOk() (*string, bool) {
	if o == nil || IsNil(o.ObjectType) {
		return nil, false
	}
	return o.ObjectType, true
}

// HasObjectType returns a boolean if a field has been set.
func (o *PatchedWritableBookmarkRequest) HasObjectType() bool {
	if o != nil && !IsNil(o.ObjectType) {
		return true
	}

	return false
}

// SetObjectType gets a reference to the given string and assigns it to the ObjectType field.
func (o *PatchedWritableBookmarkRequest) SetObjectType(v string) {
	o.ObjectType = &v
}

// GetObjectId returns the ObjectId field value if set, zero value otherwise.
func (o *PatchedWritableBookmarkRequest) GetObjectId() int64 {
	if o == nil || IsNil(o.ObjectId) {
		var ret int64
		return ret
	}
	return *o.ObjectId
}

// GetObjectIdOk returns a tuple with the ObjectId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedWritableBookmarkRequest) GetObjectIdOk() (*int64, bool) {
	if o == nil || IsNil(o.ObjectId) {
		return nil, false
	}
	return o.ObjectId, true
}

// HasObjectId returns a boolean if a field has been set.
func (o *PatchedWritableBookmarkRequest) HasObjectId() bool {
	if o != nil && !IsNil(o.ObjectId) {
		return true
	}

	return false
}

// SetObjectId gets a reference to the given int64 and assigns it to the ObjectId field.
func (o *PatchedWritableBookmarkRequest) SetObjectId(v int64) {
	o.ObjectId = &v
}

// GetUser returns the User field value if set, zero value otherwise.
func (o *PatchedWritableBookmarkRequest) GetUser() int32 {
	if o == nil || IsNil(o.User) {
		var ret int32
		return ret
	}
	return *o.User
}

// GetUserOk returns a tuple with the User field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedWritableBookmarkRequest) GetUserOk() (*int32, bool) {
	if o == nil || IsNil(o.User) {
		return nil, false
	}
	return o.User, true
}

// HasUser returns a boolean if a field has been set.
func (o *PatchedWritableBookmarkRequest) HasUser() bool {
	if o != nil && !IsNil(o.User) {
		return true
	}

	return false
}

// SetUser gets a reference to the given int32 and assigns it to the User field.
func (o *PatchedWritableBookmarkRequest) SetUser(v int32) {
	o.User = &v
}

func (o PatchedWritableBookmarkRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PatchedWritableBookmarkRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.ObjectType) {
		toSerialize["object_type"] = o.ObjectType
	}
	if !IsNil(o.ObjectId) {
		toSerialize["object_id"] = o.ObjectId
	}
	if !IsNil(o.User) {
		toSerialize["user"] = o.User
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *PatchedWritableBookmarkRequest) UnmarshalJSON(data []byte) (err error) {
	varPatchedWritableBookmarkRequest := _PatchedWritableBookmarkRequest{}

	err = json.Unmarshal(data, &varPatchedWritableBookmarkRequest)

	if err != nil {
		return err
	}

	*o = PatchedWritableBookmarkRequest(varPatchedWritableBookmarkRequest)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "object_type")
		delete(additionalProperties, "object_id")
		delete(additionalProperties, "user")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullablePatchedWritableBookmarkRequest struct {
	value *PatchedWritableBookmarkRequest
	isSet bool
}

func (v NullablePatchedWritableBookmarkRequest) Get() *PatchedWritableBookmarkRequest {
	return v.value
}

func (v *NullablePatchedWritableBookmarkRequest) Set(val *PatchedWritableBookmarkRequest) {
	v.value = val
	v.isSet = true
}

func (v NullablePatchedWritableBookmarkRequest) IsSet() bool {
	return v.isSet
}

func (v *NullablePatchedWritableBookmarkRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePatchedWritableBookmarkRequest(val *PatchedWritableBookmarkRequest) *NullablePatchedWritableBookmarkRequest {
	return &NullablePatchedWritableBookmarkRequest{value: val, isSet: true}
}

func (v NullablePatchedWritableBookmarkRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePatchedWritableBookmarkRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
