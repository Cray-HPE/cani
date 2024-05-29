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

// checks if the PatchedCustomLinkRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PatchedCustomLinkRequest{}

// PatchedCustomLinkRequest Extends the built-in ModelSerializer to enforce calling full_clean() on a copy of the associated instance during validation. (DRF does not do this by default; see https://github.com/encode/django-rest-framework/issues/3144)
type PatchedCustomLinkRequest struct {
	ContentTypes []string `json:"content_types,omitempty"`
	Name         *string  `json:"name,omitempty"`
	Enabled      *bool    `json:"enabled,omitempty"`
	// Jinja2 template code for link text
	LinkText *string `json:"link_text,omitempty"`
	// Jinja2 template code for link URL
	LinkUrl *string `json:"link_url,omitempty"`
	Weight  *int32  `json:"weight,omitempty"`
	// Links with the same group will appear as a dropdown menu
	GroupName   *string                `json:"group_name,omitempty"`
	ButtonClass *CustomLinkButtonClass `json:"button_class,omitempty"`
	// Force link to open in a new window
	NewWindow            *bool `json:"new_window,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _PatchedCustomLinkRequest PatchedCustomLinkRequest

// NewPatchedCustomLinkRequest instantiates a new PatchedCustomLinkRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPatchedCustomLinkRequest() *PatchedCustomLinkRequest {
	this := PatchedCustomLinkRequest{}
	return &this
}

// NewPatchedCustomLinkRequestWithDefaults instantiates a new PatchedCustomLinkRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPatchedCustomLinkRequestWithDefaults() *PatchedCustomLinkRequest {
	this := PatchedCustomLinkRequest{}
	return &this
}

// GetContentTypes returns the ContentTypes field value if set, zero value otherwise.
func (o *PatchedCustomLinkRequest) GetContentTypes() []string {
	if o == nil || IsNil(o.ContentTypes) {
		var ret []string
		return ret
	}
	return o.ContentTypes
}

// GetContentTypesOk returns a tuple with the ContentTypes field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedCustomLinkRequest) GetContentTypesOk() ([]string, bool) {
	if o == nil || IsNil(o.ContentTypes) {
		return nil, false
	}
	return o.ContentTypes, true
}

// HasContentTypes returns a boolean if a field has been set.
func (o *PatchedCustomLinkRequest) HasContentTypes() bool {
	if o != nil && !IsNil(o.ContentTypes) {
		return true
	}

	return false
}

// SetContentTypes gets a reference to the given []string and assigns it to the ContentTypes field.
func (o *PatchedCustomLinkRequest) SetContentTypes(v []string) {
	o.ContentTypes = v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *PatchedCustomLinkRequest) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedCustomLinkRequest) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *PatchedCustomLinkRequest) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *PatchedCustomLinkRequest) SetName(v string) {
	o.Name = &v
}

// GetEnabled returns the Enabled field value if set, zero value otherwise.
func (o *PatchedCustomLinkRequest) GetEnabled() bool {
	if o == nil || IsNil(o.Enabled) {
		var ret bool
		return ret
	}
	return *o.Enabled
}

// GetEnabledOk returns a tuple with the Enabled field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedCustomLinkRequest) GetEnabledOk() (*bool, bool) {
	if o == nil || IsNil(o.Enabled) {
		return nil, false
	}
	return o.Enabled, true
}

// HasEnabled returns a boolean if a field has been set.
func (o *PatchedCustomLinkRequest) HasEnabled() bool {
	if o != nil && !IsNil(o.Enabled) {
		return true
	}

	return false
}

// SetEnabled gets a reference to the given bool and assigns it to the Enabled field.
func (o *PatchedCustomLinkRequest) SetEnabled(v bool) {
	o.Enabled = &v
}

// GetLinkText returns the LinkText field value if set, zero value otherwise.
func (o *PatchedCustomLinkRequest) GetLinkText() string {
	if o == nil || IsNil(o.LinkText) {
		var ret string
		return ret
	}
	return *o.LinkText
}

// GetLinkTextOk returns a tuple with the LinkText field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedCustomLinkRequest) GetLinkTextOk() (*string, bool) {
	if o == nil || IsNil(o.LinkText) {
		return nil, false
	}
	return o.LinkText, true
}

// HasLinkText returns a boolean if a field has been set.
func (o *PatchedCustomLinkRequest) HasLinkText() bool {
	if o != nil && !IsNil(o.LinkText) {
		return true
	}

	return false
}

// SetLinkText gets a reference to the given string and assigns it to the LinkText field.
func (o *PatchedCustomLinkRequest) SetLinkText(v string) {
	o.LinkText = &v
}

// GetLinkUrl returns the LinkUrl field value if set, zero value otherwise.
func (o *PatchedCustomLinkRequest) GetLinkUrl() string {
	if o == nil || IsNil(o.LinkUrl) {
		var ret string
		return ret
	}
	return *o.LinkUrl
}

// GetLinkUrlOk returns a tuple with the LinkUrl field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedCustomLinkRequest) GetLinkUrlOk() (*string, bool) {
	if o == nil || IsNil(o.LinkUrl) {
		return nil, false
	}
	return o.LinkUrl, true
}

// HasLinkUrl returns a boolean if a field has been set.
func (o *PatchedCustomLinkRequest) HasLinkUrl() bool {
	if o != nil && !IsNil(o.LinkUrl) {
		return true
	}

	return false
}

// SetLinkUrl gets a reference to the given string and assigns it to the LinkUrl field.
func (o *PatchedCustomLinkRequest) SetLinkUrl(v string) {
	o.LinkUrl = &v
}

// GetWeight returns the Weight field value if set, zero value otherwise.
func (o *PatchedCustomLinkRequest) GetWeight() int32 {
	if o == nil || IsNil(o.Weight) {
		var ret int32
		return ret
	}
	return *o.Weight
}

// GetWeightOk returns a tuple with the Weight field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedCustomLinkRequest) GetWeightOk() (*int32, bool) {
	if o == nil || IsNil(o.Weight) {
		return nil, false
	}
	return o.Weight, true
}

// HasWeight returns a boolean if a field has been set.
func (o *PatchedCustomLinkRequest) HasWeight() bool {
	if o != nil && !IsNil(o.Weight) {
		return true
	}

	return false
}

// SetWeight gets a reference to the given int32 and assigns it to the Weight field.
func (o *PatchedCustomLinkRequest) SetWeight(v int32) {
	o.Weight = &v
}

// GetGroupName returns the GroupName field value if set, zero value otherwise.
func (o *PatchedCustomLinkRequest) GetGroupName() string {
	if o == nil || IsNil(o.GroupName) {
		var ret string
		return ret
	}
	return *o.GroupName
}

// GetGroupNameOk returns a tuple with the GroupName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedCustomLinkRequest) GetGroupNameOk() (*string, bool) {
	if o == nil || IsNil(o.GroupName) {
		return nil, false
	}
	return o.GroupName, true
}

// HasGroupName returns a boolean if a field has been set.
func (o *PatchedCustomLinkRequest) HasGroupName() bool {
	if o != nil && !IsNil(o.GroupName) {
		return true
	}

	return false
}

// SetGroupName gets a reference to the given string and assigns it to the GroupName field.
func (o *PatchedCustomLinkRequest) SetGroupName(v string) {
	o.GroupName = &v
}

// GetButtonClass returns the ButtonClass field value if set, zero value otherwise.
func (o *PatchedCustomLinkRequest) GetButtonClass() CustomLinkButtonClass {
	if o == nil || IsNil(o.ButtonClass) {
		var ret CustomLinkButtonClass
		return ret
	}
	return *o.ButtonClass
}

// GetButtonClassOk returns a tuple with the ButtonClass field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedCustomLinkRequest) GetButtonClassOk() (*CustomLinkButtonClass, bool) {
	if o == nil || IsNil(o.ButtonClass) {
		return nil, false
	}
	return o.ButtonClass, true
}

// HasButtonClass returns a boolean if a field has been set.
func (o *PatchedCustomLinkRequest) HasButtonClass() bool {
	if o != nil && !IsNil(o.ButtonClass) {
		return true
	}

	return false
}

// SetButtonClass gets a reference to the given CustomLinkButtonClass and assigns it to the ButtonClass field.
func (o *PatchedCustomLinkRequest) SetButtonClass(v CustomLinkButtonClass) {
	o.ButtonClass = &v
}

// GetNewWindow returns the NewWindow field value if set, zero value otherwise.
func (o *PatchedCustomLinkRequest) GetNewWindow() bool {
	if o == nil || IsNil(o.NewWindow) {
		var ret bool
		return ret
	}
	return *o.NewWindow
}

// GetNewWindowOk returns a tuple with the NewWindow field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PatchedCustomLinkRequest) GetNewWindowOk() (*bool, bool) {
	if o == nil || IsNil(o.NewWindow) {
		return nil, false
	}
	return o.NewWindow, true
}

// HasNewWindow returns a boolean if a field has been set.
func (o *PatchedCustomLinkRequest) HasNewWindow() bool {
	if o != nil && !IsNil(o.NewWindow) {
		return true
	}

	return false
}

// SetNewWindow gets a reference to the given bool and assigns it to the NewWindow field.
func (o *PatchedCustomLinkRequest) SetNewWindow(v bool) {
	o.NewWindow = &v
}

func (o PatchedCustomLinkRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PatchedCustomLinkRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.ContentTypes) {
		toSerialize["content_types"] = o.ContentTypes
	}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Enabled) {
		toSerialize["enabled"] = o.Enabled
	}
	if !IsNil(o.LinkText) {
		toSerialize["link_text"] = o.LinkText
	}
	if !IsNil(o.LinkUrl) {
		toSerialize["link_url"] = o.LinkUrl
	}
	if !IsNil(o.Weight) {
		toSerialize["weight"] = o.Weight
	}
	if !IsNil(o.GroupName) {
		toSerialize["group_name"] = o.GroupName
	}
	if !IsNil(o.ButtonClass) {
		toSerialize["button_class"] = o.ButtonClass
	}
	if !IsNil(o.NewWindow) {
		toSerialize["new_window"] = o.NewWindow
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *PatchedCustomLinkRequest) UnmarshalJSON(data []byte) (err error) {
	varPatchedCustomLinkRequest := _PatchedCustomLinkRequest{}

	err = json.Unmarshal(data, &varPatchedCustomLinkRequest)

	if err != nil {
		return err
	}

	*o = PatchedCustomLinkRequest(varPatchedCustomLinkRequest)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "content_types")
		delete(additionalProperties, "name")
		delete(additionalProperties, "enabled")
		delete(additionalProperties, "link_text")
		delete(additionalProperties, "link_url")
		delete(additionalProperties, "weight")
		delete(additionalProperties, "group_name")
		delete(additionalProperties, "button_class")
		delete(additionalProperties, "new_window")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullablePatchedCustomLinkRequest struct {
	value *PatchedCustomLinkRequest
	isSet bool
}

func (v NullablePatchedCustomLinkRequest) Get() *PatchedCustomLinkRequest {
	return v.value
}

func (v *NullablePatchedCustomLinkRequest) Set(val *PatchedCustomLinkRequest) {
	v.value = val
	v.isSet = true
}

func (v NullablePatchedCustomLinkRequest) IsSet() bool {
	return v.isSet
}

func (v *NullablePatchedCustomLinkRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePatchedCustomLinkRequest(val *PatchedCustomLinkRequest) *NullablePatchedCustomLinkRequest {
	return &NullablePatchedCustomLinkRequest{value: val, isSet: true}
}

func (v NullablePatchedCustomLinkRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePatchedCustomLinkRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
