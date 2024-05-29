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

// AuthenticationCipher * `auto` - Auto * `tkip` - TKIP * `aes` - AES
type AuthenticationCipher string

// List of Authentication_cipher
const (
	AUTHENTICATIONCIPHER_AUTO  AuthenticationCipher = "auto"
	AUTHENTICATIONCIPHER_TKIP  AuthenticationCipher = "tkip"
	AUTHENTICATIONCIPHER_AES   AuthenticationCipher = "aes"
	AUTHENTICATIONCIPHER_EMPTY AuthenticationCipher = ""
)

// All allowed values of AuthenticationCipher enum
var AllowedAuthenticationCipherEnumValues = []AuthenticationCipher{
	"auto",
	"tkip",
	"aes",
	"",
}

func (v *AuthenticationCipher) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := AuthenticationCipher(value)
	for _, existing := range AllowedAuthenticationCipherEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid AuthenticationCipher", value)
}

// NewAuthenticationCipherFromValue returns a pointer to a valid AuthenticationCipher
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewAuthenticationCipherFromValue(v string) (*AuthenticationCipher, error) {
	ev := AuthenticationCipher(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for AuthenticationCipher: valid values are %v", v, AllowedAuthenticationCipherEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v AuthenticationCipher) IsValid() bool {
	for _, existing := range AllowedAuthenticationCipherEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Authentication_cipher value
func (v AuthenticationCipher) Ptr() *AuthenticationCipher {
	return &v
}

type NullableAuthenticationCipher struct {
	value *AuthenticationCipher
	isSet bool
}

func (v NullableAuthenticationCipher) Get() *AuthenticationCipher {
	return v.value
}

func (v *NullableAuthenticationCipher) Set(val *AuthenticationCipher) {
	v.value = val
	v.isSet = true
}

func (v NullableAuthenticationCipher) IsSet() bool {
	return v.isSet
}

func (v *NullableAuthenticationCipher) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableAuthenticationCipher(val *AuthenticationCipher) *NullableAuthenticationCipher {
	return &NullableAuthenticationCipher{value: val, isSet: true}
}

func (v NullableAuthenticationCipher) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableAuthenticationCipher) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
